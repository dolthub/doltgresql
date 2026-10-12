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
fn test_application_settings_execution() {
    run_scripts(&[
        ScriptTest {
            name: "built-in settings reach execution and restore at transaction boundaries",
            set_up_script: &[
                "CREATE SCHEMA settings_schema",
                "CREATE TABLE settings_schema.settings_table (v int)",
                "INSERT INTO settings_schema.settings_table VALUES (42)",
                "SET search_path = public",
                "SET row_security = on",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SAVEPOINT settings",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL search_path = settings_schema",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM settings_table",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("settings_schema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET row_security = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('row_security')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = settings_schema",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM settings_table",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM settings_table",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "settings_table" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_application_settings_formatting() {
    run_scripts(&[
        ScriptTest {
            name: "PostgreSQL boolean and read-only settings",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_setting('row_security')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET row_security = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('row_security')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET archive_mode = on",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "archive_mode" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_application_settings_wire() {
    run_wire_tests(&[
        WireTest {
            name: "TestApplicationSettingsWire",
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
                    Send::Parse { name: "", query: "SELECT current_setting('app.tenant', true)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Null]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting('app.never')", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42704", message: r#"unrecognized configuration parameter "app.never""#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET app.tenant = 'session-a'"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("session-a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET app.tenant = 42"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("42")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET app.tenant = unquoted"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("unquoted")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET app.tenant = 'session-a'"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("session-a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "read_tenant", query: "SELECT current_setting('app.tenant')", parameter_oids: &[] },
                    Send::Describe(b'S', "read_tenant"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "read_tenant", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("session-a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SHOW app.tenant", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "app.tenant", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("session-a")]),
                    Receive::CommandComplete("SHOW"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("BEGIN"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SET LOCAL app.tenant = 'local-a'"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("local-a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "read_tenant", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("local-a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SAVEPOINT request"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SAVEPOINT"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT set_config('app.tenant', 'local-b', true)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "set_config", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("local-b")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("local-b")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("ROLLBACK TO request"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("local-a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("COMMIT"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COMMIT"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("session-a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "read_tenant", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("session-a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("RESET app.tenant"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("RESET"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET app.tenant = 'session-b'"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("session-b")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "read_tenant", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("session-b")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET search_path = public"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ParameterStatus("search_path", "public"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("search_path")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("public")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("BEGIN"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT set_config('search_path', 'other', true)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "set_config", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("other")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ParameterStatus("search_path", "other"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("search_path")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("other")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SAVEPOINT settings"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SAVEPOINT"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("SET LOCAL row_security = off"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("row_security")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("off")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("ROLLBACK TO settings"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("ROLLBACK"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("row_security")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("on")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'T'),
                ]),
                Step::Send(&[
                    Send::Query("COMMIT"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("COMMIT"),
                    Receive::ParameterStatus("search_path", "public"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("search_path")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("public")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT set_config('search_path', NULL, false)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "set_config", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#""$user", public"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ParameterStatus("search_path", r#""$user", public"#),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET search_path = public"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
                    Receive::ParameterStatus("search_path", "public"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("RESET ALL"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("RESET"),
                    Receive::ParameterStatus("search_path", r#""$user", public"#),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("search_path")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#""$user", public"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("app.tenant")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT set_config('bad..name', 'x', false)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "set_config", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42602", message: r#"invalid configuration parameter name "bad..name""#, detail: "Custom parameter names must be two or more simple identifiers separated by dots.", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET app.tenant = 'session-c'"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("SET"),
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
                    Send::Parse { name: "", query: "SELECT current_setting('app.tenant', true)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting($1)", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[TEXT]),
                    Receive::RowDescription(&[Field { name: "current_setting", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[0], parameters: &[Datum::Text("search_path")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#""$user", public"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}
