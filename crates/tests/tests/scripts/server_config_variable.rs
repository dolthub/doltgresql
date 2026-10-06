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
fn test_server_config_variable_statement() {
    run_wire_tests(&[
        WireTest {
            name: "TestServerConfigVariableStatement",
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
                    Send::Parse { name: "", query: "SHOW dolt_skip_replication_errors", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42704", message: r#"unrecognized configuration parameter "dolt_skip_replication_errors""#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "26000", message: "unnamed prepared statement does not exist", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting('dolt_skip_replication_errors')", parameter_oids: &[] },
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
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42704", message: r#"unrecognized configuration parameter "dolt_skip_replication_errors""#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SHOW DOLT_SKIP_REPLICATION_ERRORS", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42704", message: r#"unrecognized configuration parameter "dolt_skip_replication_errors""#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "26000", message: "unnamed prepared statement does not exist", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SHOW port", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "port", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("{PORT}")]),
                    Receive::CommandComplete("SHOW"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SET port TO '5432'"),
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "55P02", message: r#"parameter "port" cannot be changed without restarting the server"#, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "SELECT current_setting('port')", parameter_oids: &[] },
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
                    Receive::DataRow(&[Datum::Text("{PORT}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}
