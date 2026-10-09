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
fn test_sql_state_codes() {
    run_scripts(&[
        ScriptTest {
            name: "class 22 data exception",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1/0",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 5 % 0",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::int4",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "abc""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (2000000000)::int4 + (2000000000)::int4",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "integer out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 9223372036854775807::int8 + 1",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (-9223372036854775807)::int8 - 2",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3037000500::int8 * 3037000500::int8",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT acos(2.0)",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "input is out of range", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "class 23 integrity constraint violation",
            set_up_script: &[
                "CREATE TABLE constraints_t (id INT PRIMARY KEY, u INT UNIQUE, n INT NOT NULL, c INT CHECK (c > 0))",
                "INSERT INTO constraints_t VALUES (1, 1, 1, 1)",
                "CREATE TABLE fk_parent (id INT PRIMARY KEY)",
                "CREATE TABLE fk_child (id INT PRIMARY KEY, pid INT REFERENCES fk_parent(id))",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO constraints_t VALUES (1, 2, 2, 2)",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "constraints_t_pkey""#, detail: "Key (id)=(1) already exists.", schema: "public", table: "constraints_t", constraint: "constraints_t_pkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO constraints_t VALUES (2, 1, 2, 2)",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "constraints_t_u_key""#, detail: "Key (u)=(1) already exists.", schema: "public", table: "constraints_t", constraint: "constraints_t_u_key", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO constraints_t VALUES (2, 2, NULL, 2)",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "n" of relation "constraints_t" violates not-null constraint"#, detail: "Failing row contains (2, 2, null, 2).", schema: "public", table: "constraints_t", column: "n", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO constraints_t VALUES (2, 2, 2, -1)",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "constraints_t" violates check constraint "constraints_t_c_check""#, detail: "Failing row contains (2, 2, 2, -1).", schema: "public", table: "constraints_t", constraint: "constraints_t_c_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO fk_child VALUES (1, 99)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "fk_child" violates foreign key constraint "fk_child_pid_fkey""#, detail: r#"Key (pid)=(99) is not present in table "fk_parent"."#, schema: "public", table: "fk_child", constraint: "fk_child_pid_fkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "class 42 undefined and duplicate objects",
            set_up_script: &[
                "CREATE TABLE existing_t (id INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELEC 1",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "SELEC""#, position: 1, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM no_such_table",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "no_such_table" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT no_such_col FROM existing_t",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "no_such_col" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT no_such_function(1)",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function no_such_function(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE existing_t (id INT PRIMARY KEY)",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "existing_t" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1::no_such_type",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "no_such_type" does not exist"#, position: 11, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "class 3D undefined database",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM no_such_db.public.tbl",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "no_such_db.public.tbl""#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "class 21 cardinality violation",
            set_up_script: &[
                "CREATE TABLE two_rows (id INT PRIMARY KEY)",
                "INSERT INTO two_rows VALUES (1), (2)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (SELECT id FROM two_rows)",
                    expected: Expected::Error(Diagnostic { code: "21000", message: "more than one row returned by a subquery used as an expression", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
