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
fn test_auth_quick() {
    run_scripts(&[
        ScriptTest {
            name: r#"GRANT SELECT ON ALL TABLES IN SCHEMA mysch TO tester;
 > SELECT * FROM mysch.test;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON ALL TABLES IN SCHEMA mysch TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON ALL TABLES IN SCHEMA mysch TO tester;
 > SELECT * FROM mysch.test2;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON ALL TABLES IN SCHEMA mysch TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test2;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON mysch.test TO tester;
 > SELECT * FROM mysch.test;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON mysch.test TO tester;
 > SELECT * FROM mysch.test2;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test2;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON ALL TABLES IN SCHEMA othersch TO tester;
 > SELECT * FROM mysch.test;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON ALL TABLES IN SCHEMA othersch TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON othersch.test TO tester;
 > SELECT * FROM mysch.test;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON othersch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON othersch.test TO tester;
 > SELECT * FROM mysch.test;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON othersch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE SCHEMA newsch;",
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newsch;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for database postgres", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT CREATE ON DATABASE postgres TO tester;
 > CREATE SCHEMA newsch;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT CREATE ON DATABASE postgres TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newsch;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT CREATE ON DATABASE postgres TO tester;
 > CREATE SCHEMA newsch;
 > DROP SCHEMA newsch;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT CREATE ON DATABASE postgres TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA newsch;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SCHEMA newsch;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "must be owner of schema newsch", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 14, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT CREATE ON SCHEMA mysch TO tester;
 > CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT CREATE ON SCHEMA mysch TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE ROLE new_role;",
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE ROLE new_role;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to create role", detail: "Only roles with the CREATEROLE attribute may create roles.", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"ALTER ROLE tester CREATEROLE;
 > CREATE ROLE new_role;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER ROLE tester CREATEROLE;",
                    expected: Expected::Tag("ALTER ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE new_role;",
                    expected: Expected::Tag("CREATE ROLE"),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE USER new_user;",
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER new_user;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to create role", detail: "Only roles with the CREATEROLE attribute may create roles.", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"ALTER ROLE tester SUPERUSER;
 > CREATE USER new_user;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER ROLE tester SUPERUSER;",
                    expected: Expected::Tag("ALTER ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER new_user;",
                    expected: Expected::Tag("CREATE ROLE"),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE USER new_user;
 > DROP USER new_user;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER new_user;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER new_user;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to drop role", detail: "Only roles with the CREATEROLE attribute and the ADMIN option on the target roles may drop roles.", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE USER new_user;
 > ALTER ROLE tester CREATEROLE;
 > DROP USER new_user;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER new_user;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER ROLE tester CREATEROLE;",
                    expected: Expected::Tag("ALTER ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER new_user;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to drop role", detail: r#"Only roles with the CREATEROLE attribute and the ADMIN option on role "new_user" may drop this role."#, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE USER new_user SUPERUSER;
 > ALTER ROLE tester CREATEROLE;
 > DROP USER new_user;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER new_user SUPERUSER;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER ROLE tester CREATEROLE;",
                    expected: Expected::Tag("ALTER ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER new_user;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to drop role", detail: "Only roles with the SUPERUSER attribute may drop roles with the SUPERUSER attribute.", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE USER new_user SUPERUSER;
 > ALTER ROLE tester SUPERUSER;
 > DROP USER new_user;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER new_user SUPERUSER;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER ROLE tester SUPERUSER;",
                    expected: Expected::Tag("ALTER ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER new_user;",
                    expected: Expected::Tag("DROP ROLE"),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DELETE FROM mysch.test WHERE pk >= 0;",
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DELETE FROM mysch.test WHERE pk >= 0;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 13, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT DELETE ON ALL TABLES IN SCHEMA mysch TO tester;
 > DELETE FROM mysch.test WHERE pk >= 0;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT DELETE ON ALL TABLES IN SCHEMA mysch TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM mysch.test WHERE pk >= 0;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 13, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT DELETE ON mysch.test TO tester;
 > DELETE FROM mysch.test WHERE pk >= 0;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT DELETE ON mysch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM mysch.test WHERE pk >= 0;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 13, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE USER tester2;
 > GRANT DELETE ON ALL TABLES IN SCHEMA mysch TO tester2;
 > GRANT tester2 TO tester;
 > DELETE FROM mysch.test WHERE pk >= 0;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER tester2;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT DELETE ON ALL TABLES IN SCHEMA mysch TO tester2;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT tester2 TO tester;",
                    expected: Expected::Tag("GRANT ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM mysch.test WHERE pk >= 0;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 13, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON mysch.test TO tester;
 > SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON mysch.test2 TO tester;
 > SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test2 TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT SELECT ON mysch.test TO tester;
 > GRANT SELECT ON mysch.test2 TO tester;
 > SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test2 TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE USER tester2;
 > GRANT SELECT ON mysch.test2 TO tester2;
 > GRANT tester2 TO tester;
 > SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER tester2;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test2 TO tester2;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT tester2 TO tester;",
                    expected: Expected::Tag("GRANT ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE USER tester2;
 > GRANT SELECT ON mysch.test TO tester2;
 > GRANT SELECT ON mysch.test2 TO tester2;
 > GRANT tester2 TO tester;
 > SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER tester2;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test TO tester2;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT SELECT ON mysch.test2 TO tester2;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT tester2 TO tester;",
                    expected: Expected::Tag("GRANT ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 15, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);
 > DROP TABLE mysch.new_table;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE mysch.new_table;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA mysch TO tester;
 > REVOKE DROP ON ALL TABLES IN SCHEMA mysch FROM tester;
 > CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);
 > DROP TABLE mysch.new_table;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA mysch TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "REVOKE DROP ON ALL TABLES IN SCHEMA mysch FROM tester;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"unrecognized privilege type "drop""#, ..E }),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE mysch.new_table;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);
 > GRANT DROP ON mysch.new_table TO tester;
 > DROP TABLE mysch.new_table;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT DROP ON mysch.new_table TO tester;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"unrecognized privilege type "drop""#, ..E }),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE mysch.new_table;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);
 > GRANT postgres TO tester;
 > DROP TABLE mysch.new_table;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT postgres TO tester;",
                    expected: Expected::Tag("GRANT ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE mysch.new_table;",
                    expected: Expected::Tag("DROP TABLE"),
                    flow: Flow::Exec,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"CREATE ROLE new_role;
 > DROP ROLE new_role;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE ROLE new_role;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE new_role;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to drop role", detail: "Only roles with the CREATEROLE attribute and the ADMIN option on the target roles may drop roles.", ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"ALTER ROLE tester CREATEROLE;
 > CREATE ROLE new_role;
 > DROP ROLE new_role;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER ROLE tester CREATEROLE;",
                    expected: Expected::Tag("ALTER ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE new_role;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE new_role;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to drop role", detail: r#"Only roles with the CREATEROLE attribute and the ADMIN option on role "new_role" may drop this role."#, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "INSERT INTO mysch.test VALUES (9, 9);",
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO mysch.test VALUES (9, 9);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 13, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT INSERT ON ALL TABLES IN SCHEMA mysch TO tester;
 > INSERT INTO mysch.test VALUES (9, 9);"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT INSERT ON ALL TABLES IN SCHEMA mysch TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mysch.test VALUES (9, 9);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 13, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT INSERT ON mysch.test TO tester;
 > INSERT INTO mysch.test VALUES (9, 9);"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT INSERT ON mysch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mysch.test VALUES (9, 9);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 13, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE mysch.test SET v1 = 0;",
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE mysch.test SET v1 = 0;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 8, ..E }),
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT UPDATE ON ALL TABLES IN SCHEMA mysch TO tester;
 > UPDATE mysch.test SET v1 = 0;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT UPDATE ON ALL TABLES IN SCHEMA mysch TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE mysch.test SET v1 = 0;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 8, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"GRANT UPDATE ON mysch.test TO tester;
 > UPDATE mysch.test SET v1 = 0;"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA mysch;",
                "CREATE SCHEMA othersch;",
                "CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
                "INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
                "INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
                "INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
                "INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "GRANT UPDATE ON mysch.test TO tester;",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE mysch.test SET v1 = 0;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema mysch", position: 8, ..E }),
                    flow: Flow::Query,
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
    ]);
}
