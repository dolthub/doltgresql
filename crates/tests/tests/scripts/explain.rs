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
fn test_explain() {
    run_scripts(&[
        ScriptTest {
            name: "basic explain tests",
            set_up_script: &[
                "CREATE TABLE t (i INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT * FROM T;",
                    skip: Some("the Go test expects Postgres plan text, whose costs are implementation details"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"
EXPLAIN 
(
	ANALYZE, 
	VERBOSE, 
	COSTS, 
	SETTINGS,
	BUFFERS,
	WAL,
	TIMING,
	SUMMARY,
	FORMAT TEXT
) 
	SELECT * FROM t;
"#,
                    skip: Some("the Go test expects Postgres plan text, whose costs are implementation details"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"
EXPLAIN 
(
	ANALYZE ON, 
	VERBOSE OFF, 
	COSTS TRUE, 
	SETTINGS FALSE,
	BUFFERS,
	WAL,
	TIMING,
	SUMMARY,
	FORMAT TEXT
) 
	SELECT * FROM t;
"#,
                    skip: Some("the Go test expects Postgres plan text, whose costs are implementation details"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"
EXPLAIN 
(
	NOTAVALIDOPTION
) 
	SELECT * FROM t;
"#,
                    skip: Some("the Go test expects Postgres plan text, whose costs are implementation details"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
