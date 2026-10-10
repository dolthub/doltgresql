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
fn test_identity_expressions() {
    run_scripts(&[
        ScriptTest {
            name: "login identity expressions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_user, current_role, user, session_user, current_user()",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "(""#, position: 68, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE identity_reader LOGIN PASSWORD 'identity_password'",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, current_role, user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("current_role", NAME), Column("user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("identity_reader"), T("identity_reader"), T("identity_reader"), T("identity_reader")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "identity_reader",
                    password: "identity_password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE identity_reader",
                    expected: Expected::Tag("DROP ROLE"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_identity_prepared_and_independent_connections() {
    run_scripts(&[
        ScriptTest {
            name: "identity prepared and independent connections",
            set_up_script: &[
                "CREATE ROLE identity_reader LOGIN PASSWORD 'identity_password'",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Tag(""),
                    username: "identity_reader",
                    password: "identity_password",
                    client: "reader",
                    prepare: "identity_query",
                    ..A
                },
                ScriptTestAssertion {
                    query: "identity_query",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("identity_reader"), T("identity_reader")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "identity_reader",
                    password: "identity_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "identity_query",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("identity_reader"), T("identity_reader")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "identity_reader",
                    password: "identity_password",
                    client: "reader",
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_session_authorization_role_drop_and_reuse() {
    run_scripts(&[
        ScriptTest {
            name: "session authorization role drop and reuse",
            set_up_script: &[
                "CREATE ROLE session_dropped",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_dropped",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT session_user",
                    expected: Expected::Rows {
                        columns: &[Column("session_user", NAME)],
                        rows: &[
                            &[T("session_dropped")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE session_dropped",
                    expected: Expected::Tag("DROP ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    client: "other",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE session_dropped",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    client: "other",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT session_user",
                    expected: Expected::Error(Diagnostic { code: "42704", message: "invalid role OID: 16384", ..E }),
                    flow: Flow::Query,
                    client: "main",
                    skip: Some("the error names the role's Postgres OID, which Doltgres numbers differently"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET SESSION AUTHORIZATION",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT session_user",
                    expected: Expected::Rows {
                        columns: &[Column("session_user", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_dropped",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD ALL",
                    expected: Expected::Tag("DISCARD ALL"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT session_user",
                    expected: Expected::Rows {
                        columns: &[Column("session_user", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_session_authorization_wire_and_reset() {
    run_scripts(&[
        ScriptTest {
            name: "session authorization wire and reset",
            set_up_script: &[
                "CREATE ROLE session_login LOGIN PASSWORD 'session_password'",
                "CREATE ROLE session_target",
                "CREATE ROLE session_other",
                "GRANT session_target TO session_login",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SET ROLE session_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_target",
                    expected: Expected::Error(Diagnostic { code: "42501", message: r#"permission denied to set session authorization "session_target""#, ..E }),
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_target"), T("session_login"), T("session_target")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_login",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_login"), T("session_login"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_other",
                    expected: Expected::Error(Diagnostic { code: "42501", message: r#"permission denied to set session authorization "session_other""#, ..E }),
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_missing",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"role "session_missing" does not exist"#, ..E }),
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET SESSION AUTHORIZATION",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE session_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_other",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_other"), T("session_other"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE session_target",
                    expected: Expected::Error(Diagnostic { code: "42501", message: r#"permission denied to set role "session_target""#, ..E }),
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET SESSION AUTHORIZATION",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("postgres"), T("postgres"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE session_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_other",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_other"), T("session_other"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_target"), T("postgres"), T("session_target")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET ROLE",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SAVEPOINT auth_sp",
                    expected: Expected::Tag("SAVEPOINT"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL SESSION AUTHORIZATION session_other",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_other"), T("session_other"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK TO SAVEPOINT auth_sp",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_target"), T("session_target"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_target"), T("session_target"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION DEFAULT",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("postgres"), T("postgres"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_target"), T("session_target"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('session_authorization', 'default', false)",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"role "default" does not exist"#, ..E }),
                    flow: Flow::Query,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_target"), T("session_target"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET session_authorization = session_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_target"), T("session_target"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD ALL",
                    expected: Expected::Tag("DISCARD ALL"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("postgres"), T("postgres"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW session_authorization",
                    expected: Expected::Rows {
                        columns: &[Column("session_authorization", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SHOW",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL SESSION AUTHORIZATION session_other",
                    expected: Expected::Tag("SET"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET LOCAL can only be used in transaction blocks", ..E }],
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("postgres"), T("postgres"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SESSION AUTHORIZATION session_other",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user, current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME), Column("current_setting", TEXT)],
                        rows: &[
                            &[T("session_other"), T("session_other"), T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET SESSION AUTHORIZATION",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT session_user",
                    expected: Expected::Rows {
                        columns: &[Column("session_user", NAME)],
                        rows: &[
                            &[T("session_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "session_login",
                    password: "session_password",
                    client: "login",
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_set_role_reset_after_concurrent_drop() {
    run_scripts(&[
        ScriptTest {
            name: "set role reset after concurrent drop",
            set_up_script: &[
                "CREATE ROLE role_deleted_selected",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SET ROLE role_deleted_selected",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE role_deleted_selected",
                    expected: Expected::Tag("DROP ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    client: "other",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE role_deleted_selected",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    client: "other",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user",
                    expected: Expected::Error(Diagnostic { code: "42704", message: "invalid role OID: 16384", ..E }),
                    flow: Flow::Query,
                    client: "main",
                    skip: Some("the error names the role's Postgres OID, which Doltgres numbers differently"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET ROLE",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_deleted_selected",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE role_deleted_selected",
                    expected: Expected::Tag("DROP ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    client: "other",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE role_deleted_selected",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    username: "postgres",
                    password: "password",
                    client: "other",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user",
                    expected: Expected::Error(Diagnostic { code: "42704", message: "invalid role OID: 16385", ..E }),
                    flow: Flow::Query,
                    client: "main",
                    skip: Some("the error names the role's Postgres OID, which Doltgres numbers differently"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE NONE",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_set_role_wire_and_transaction_scopes() {
    run_scripts(&[
        ScriptTest {
            name: "set role wire and transaction scopes",
            set_up_script: &[
                "CREATE ROLE role_login LOGIN PASSWORD 'role_password' NOINHERIT",
                "CREATE ROLE role_middle",
                "CREATE ROLE role_target",
                "CREATE ROLE role_unrelated",
                "CREATE SCHEMA role_login",
                "CREATE SCHEMA role_target",
                "GRANT USAGE ON SCHEMA role_login, role_target TO PUBLIC",
                "GRANT role_target TO role_middle",
                "GRANT role_middle TO role_login",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_login"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema()",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Tag(""),
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    prepare: "role_identity",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema()",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("role_target")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW role",
                    expected: Expected::Rows {
                        columns: &[Column("role", TEXT)],
                        rows: &[
                            &[T("role_target")],
                        ],
                        tag: "SHOW",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE TO role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema()",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("role_target")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW role",
                    expected: Expected::Rows {
                        columns: &[Column("role", TEXT)],
                        rows: &[
                            &[T("role_target")],
                        ],
                        tag: "SHOW",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE = role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema()",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("role_target")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW role",
                    expected: Expected::Rows {
                        columns: &[Column("role", TEXT)],
                        rows: &[
                            &[T("role_target")],
                        ],
                        tag: "SHOW",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "role_identity",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE role_forbidden",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to create role", detail: "Only roles with the CREATEROLE attribute may create roles.", ..E }),
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_unrelated",
                    expected: Expected::Error(Diagnostic { code: "42501", message: r#"permission denied to set role "role_unrelated""#, ..E }),
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_missing",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"role "role_missing" does not exist"#, ..E }),
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE NONE",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_login"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_middle",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET ROLE",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_login"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('role', 'role_target', false)",
                    expected: Expected::Rows {
                        columns: &[Column("set_config", TEXT)],
                        rows: &[
                            &[T("role_target")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('role', 'role_unrelated', false)",
                    expected: Expected::Error(Diagnostic { code: "42501", message: r#"permission denied to set role "role_unrelated""#, ..E }),
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE DEFAULT",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "DEFAULT""#, position: 10, ..E }),
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_login"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL ROLE role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SAVEPOINT role_sp",
                    expected: Expected::Tag("SAVEPOINT"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL ROLE role_middle",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_middle"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK TO SAVEPOINT role_sp",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_login"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_login"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('role', 'role_target', true)",
                    expected: Expected::Rows {
                        columns: &[Column("set_config", TEXT)],
                        rows: &[
                            &[T("role_target")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_login"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "REVOKE role_target FROM role_middle",
                    expected: Expected::Tag("REVOKE ROLE"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_target",
                    expected: Expected::Error(Diagnostic { code: "42501", message: r#"permission denied to set role "role_target""#, ..E }),
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_target"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE NONE",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user, session_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME), Column("session_user", NAME)],
                        rows: &[
                            &[T("role_login"), T("role_login")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "role_login",
                    password: "role_password",
                    client: "reader",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_user",
                    expected: Expected::Rows {
                        columns: &[Column("current_user", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE role_protected (id INT PRIMARY KEY)",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO role_protected VALUES (7)",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM role_protected",
                    expected: Expected::Tag(""),
                    client: "main",
                    prepare: "protected_query",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "protected_query",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table role_protected", ..E }),
                    flow: Flow::Query,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET ROLE",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT SELECT ON role_protected TO role_target",
                    expected: Expected::Tag("GRANT"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM role_protected",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "protected_query",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET ROLE",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "REVOKE SELECT ON role_protected FROM role_target",
                    expected: Expected::Tag("REVOKE"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE role_target",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "protected_query",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table role_protected", ..E }),
                    flow: Flow::Query,
                    client: "main",
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET ROLE",
                    expected: Expected::Tag("RESET"),
                    flow: Flow::Exec,
                    client: "main",
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_settings_rules() {
    run_scripts(&[
        ScriptTest {
            name: "SET LOCAL outside a transaction block warns",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET LOCAL work_mem = '1MB';",
                    expected: Expected::Tag("SET"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET LOCAL can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW work_mem;",
                    expected: Expected::Rows {
                        columns: &[Column("work_mem", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL ROLE postgres;",
                    expected: Expected::Tag("SET"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET LOCAL can only be used in transaction blocks", ..E }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RESET ALL keeps placeholder parameters empty",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET app.tenant = 'a';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET ALL;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('app.tenant');",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('bad..name', 'x', false);",
                    expected: Expected::Error(Diagnostic { code: "42602", message: r#"invalid configuration parameter name "bad..name""#, detail: "Custom parameter names must be two or more simple identifiers separated by dots.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('a.1b', 'x', false);",
                    expected: Expected::Error(Diagnostic { code: "42602", message: r#"invalid configuration parameter name "a.1b""#, detail: "Custom parameter names must be two or more simple identifiers separated by dots.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('a.b$1', 'x', false);",
                    expected: Expected::Rows {
                        columns: &[Column("set_config", TEXT)],
                        rows: &[
                            &[T("x")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "the user schema follows SET ROLE",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE ROLE schema_role;",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA schema_role;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT USAGE ON SCHEMA schema_role TO schema_role;",
                    expected: Expected::Tag("GRANT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path = "$user", public;"#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ROLE schema_role;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("schema_role")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET ROLE;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("public")],
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
