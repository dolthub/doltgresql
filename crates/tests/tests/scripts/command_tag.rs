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
fn test_command_tag() {
    run_scripts(&[
        ScriptTest {
            name: "set",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET extra_float_digits = 3",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "show",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW extra_float_digits",
                    expected: Expected::Tag("SHOW"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create database",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE DATABASE mydb",
                    expected: Expected::Tag("CREATE DATABASE"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "insert",
            set_up_script: &[
                "CREATE TABLE table0 (id int, name text)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO table0 VALUES (1,'Dolt'), (2,'Doltgres'), (3,'DoltHub')",
                    expected: Expected::Tag("INSERT 0 3"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM table0 order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("Dolt")],
                            &[T("2"), T("Doltgres")],
                            &[T("3"), T("DoltHub")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM table0",
                    expected: Expected::Tag("SELECT 3"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "update",
            set_up_script: &[
                "CREATE TABLE table0 (id int, name text)",
                "INSERT INTO table0 VALUES (1,'Dolt'), (2,'Doltgres'), (3,'DoltHub')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE table0 SET id = 4 WHERE name = 'Doltgres'",
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM table0 order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("Dolt")],
                            &[T("3"), T("DoltHub")],
                            &[T("4"), T("Doltgres")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM table0 WHERE name <> 'Dolt'",
                    expected: Expected::Tag("SELECT 2"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "delete",
            set_up_script: &[
                "CREATE TABLE table0 (id int, name text)",
                "INSERT INTO table0 VALUES (1,'Dolt'), (2,'Doltgres'), (3,'DoltHub')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DELETE FROM table0",
                    expected: Expected::Tag("DELETE 3"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM table0 order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM table0",
                    expected: Expected::Tag("SELECT 0"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
