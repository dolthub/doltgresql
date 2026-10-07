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
fn test_getting_started_guide() {
    run_scripts(&[
        ScriptTest {
            name: "Doltgres Getting Started Guide",
            set_up_script: &[
                r#"create table employees (
    id int8,
    last_name text,
    first_name text,
    primary key(id));"#,
                r#"create table teams (
    id int8,
    team_name text,
    primary key(id));"#,
                r#"create table employees_teams(
    team_id int8,
    employee_id int8,
    primary key(team_id, employee_id),
    foreign key (team_id) references teams(id),
    foreign key (employee_id) references employees(id));"#,
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.employees"), T("f"), T("new table")],
                            &[T("public.employees_teams"), T("f"), T("new table")],
                            &[T("public.teams"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('teams', 'employees', 'employees_teams');",
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
                    query: "select * from dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.employees"), T("t"), T("new table")],
                            &[T("public.employees_teams"), T("t"), T("new table")],
                            &[T("public.teams"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-m', 'Created initial schema')::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_reset('--hard');",
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
                    query: "select * from dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.log;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into employees values (0, 'Sehn', 'Tim'), (1, 'Hendriks', 'Brian'), (2, 'Son','Aaron'), (3, 'Fitzgerald', 'Brian');",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into teams values (0, 'Engineering'), (1, 'Sales');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into employees_teams(employee_id, team_id) values (0,0), (1,0), (2,0), (0,1), (3,1);",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select first_name, last_name, team_name from employees
							join employees_teams on (employees.id=employees_teams.employee_id)
							join teams on (teams.id=employees_teams.team_id)
							where team_name='Engineering';"#,
                    expected: Expected::Rows {
                        columns: &[Column("first_name", TEXT), Column("last_name", TEXT), Column("team_name", TEXT)],
                        rows: &[
                            &[T("Tim"), T("Sehn"), T("Engineering")],
                            &[T("Brian"), T("Hendriks"), T("Engineering")],
                            &[T("Aaron"), T("Son"), T("Engineering")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from employees_teams where employee_id='0' and team_id='1';",
                    expected: Expected::Rows {
                        columns: &[Column("team_id", INT8), Column("employee_id", INT8)],
                        rows: &[
                            &[T("1"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from dolt.status order by table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.employees"), T("f"), T("modified")],
                            &[T("public.employees_teams"), T("f"), T("modified")],
                            &[T("public.teams"), T("f"), T("modified")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select to_last_name, to_first_name, to_id, to_commit, from_last_name, from_first_name, from_id, diff_type from dolt_diff_employees;",
                    expected: Expected::Rows {
                        columns: &[Column("to_last_name", TEXT), Column("to_first_name", TEXT), Column("to_id", INT8), Column("to_commit", TEXT), Column("from_last_name", TEXT), Column("from_first_name", TEXT), Column("from_id", INT8), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("Sehn"), T("Tim"), T("0"), T("WORKING"), Null, Null, Null, T("added")],
                            &[T("Hendriks"), T("Brian"), T("1"), T("WORKING"), Null, Null, Null, T("added")],
                            &[T("Son"), T("Aaron"), T("2"), T("WORKING"), Null, Null, Null, T("added")],
                            &[T("Fitzgerald"), T("Brian"), T("3"), T("WORKING"), Null, Null, Null, T("added")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-am', 'Populated tables with data')::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from dolt.status order by table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt.log;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("Populated tables with data")],
                            &[T("Created initial schema")],
                            &[T("CREATE DATABASE")],
                            &[T("Initialize data repository")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select table_name, message, data_change, schema_change from dolt.diff order by date desc, table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("message", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T("public.employees"), T("Populated tables with data"), T("t"), T("f")],
                            &[T("public.employees_teams"), T("Populated tables with data"), T("t"), T("f")],
                            &[T("public.teams"), T("Populated tables with data"), T("t"), T("f")],
                            &[T("public.employees"), T("Created initial schema"), T("f"), T("t")],
                            &[T("public.employees_teams"), T("Created initial schema"), T("f"), T("t")],
                            &[T("public.teams"), T("Created initial schema"), T("f"), T("t")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "drop table employees_teams;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "select count(*) from employees_teams;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "relation \"employees_teams\" does not exist", position: 22, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_reset('--hard');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from employees_teams;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('-b','modifications');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'modifications'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "update employees SET first_name='Timothy' where first_name='Tim';",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert INTO employees (id, first_name, last_name) values (4,'Daylon', 'Wilkins');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into employees_teams(team_id, employee_id) values (0,4);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "delete from employees_teams where employee_id=0 and team_id=1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-am', 'Modifications on a branch')::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name, latest_commit_message from dolt.branches;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("latest_commit_message", TEXT)],
                        rows: &[
                            &[T("main"), T("Populated tables with data")],
                            &[T("modifications"), T("Modifications on a branch")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select active_branch();",
                    expected: Expected::Rows {
                        columns: &[Column("active_branch", TEXT)],
                        rows: &[
                            &[T("main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from employees;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("last_name", TEXT), Column("first_name", TEXT)],
                        rows: &[
                            &[T("0"), T("Sehn"), T("Tim")],
                            &[T("1"), T("Hendriks"), T("Brian")],
                            &[T("2"), T("Son"), T("Aaron")],
                            &[T("3"), T("Fitzgerald"), T("Brian")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from employees as of 'modifications';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("last_name", TEXT), Column("first_name", TEXT)],
                        rows: &[
                            &[T("0"), T("Sehn"), T("Timothy")],
                            &[T("1"), T("Hendriks"), T("Brian")],
                            &[T("2"), T("Son"), T("Aaron")],
                            &[T("3"), T("Fitzgerald"), T("Brian")],
                            &[T("4"), T("Wilkins"), T("Daylon")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"select to_id, to_last_name, to_first_name, to_commit, 
							from_id, from_last_name, from_first_name, from_commit, diff_type
							from dolt_diff('main', 'modifications', 'employees');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT8), Column("to_last_name", TEXT), Column("to_first_name", TEXT), Column("to_commit", TEXT), Column("from_id", INT8), Column("from_last_name", TEXT), Column("from_first_name", TEXT), Column("from_commit", TEXT), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("0"), T("Sehn"), T("Timothy"), T("modifications"), T("0"), T("Sehn"), T("Tim"), T("main"), T("modified")],
                            &[T("4"), T("Wilkins"), T("Daylon"), T("modifications"), Null, Null, Null, T("main"), T("added")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('-b', 'schema_changes');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'schema_changes'")"#)],
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
