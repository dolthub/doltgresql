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
fn test_begin_isolation_level() {
    run_scripts(&[
        ScriptTest {
            name: "BEGIN with any isolation level clause is accepted as a no-op",
            set_up_script: &[
                "CREATE TABLE test (a INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN TRANSACTION ISOLATION LEVEL READ COMMITTED",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN ISOLATION LEVEL READ UNCOMMITTED",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
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
                    query: "BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN ISOLATION LEVEL REPEATABLE READ, READ WRITE",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ READ ONLY",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "A duplicate BEGIN does not change the active transaction's characteristics",
            set_up_script: &[
                "CREATE TABLE test (a INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN ISOLATION LEVEL SERIALIZABLE, READ WRITE",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN ISOLATION LEVEL REPEATABLE READ, READ ONLY",
                    expected: Expected::Tag("BEGIN"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25001", message: "there is already a transaction in progress", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1)",
                    expected: Expected::Error(Diagnostic { code: "25006", message: "cannot execute INSERT in a read-only transaction", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROLLBACK clears the in-transaction flag so a following BEGIN is honored",
            set_up_script: &[
                "CREATE TABLE test_rollback (a INT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN READ WRITE",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN READ ONLY",
                    expected: Expected::Tag("BEGIN"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25001", message: "there is already a transaction in progress", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN READ ONLY",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test_rollback VALUES (1)",
                    expected: Expected::Error(Diagnostic { code: "25006", message: "cannot execute INSERT in a read-only transaction", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_discard() {
    run_scripts(&[
        ScriptTest {
            name: "Test discard",
            set_up_script: &[
                "CREATE temporary TABLE test (a INT)",
                "insert into test values (1)",
                "SET search_path = pg_catalog",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("pg_catalog")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD ALL",
                    expected: Expected::Tag("DISCARD ALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Test discard errors",
            set_up_script: &[
                "CREATE temporary TABLE test (a INT)",
                "insert into test values (1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DISCARD SEQUENCES",
                    expected: Expected::Tag("DISCARD SEQUENCES"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Test discard in transaction",
            set_up_script: &[
                "CREATE temporary TABLE test (a INT)",
                "insert into test values (1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD ALL",
                    expected: Expected::Error(Diagnostic { code: "25001", message: "DISCARD ALL cannot run inside a transaction block", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_rollback() {
    run_scripts(&[
        ScriptTest {
            name: "Test rollback transaction",
            set_up_script: &[
                "BEGIN",
                "CREATE temporary TABLE test (a INT)",
                "insert into test values (1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
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
                    query: "select * from test",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "test" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "create temp table test (b int)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_session_state_after_query_error() {
    run_scripts(&[
        ScriptTest {
            name: "Test failed query does not pin the session to a stale root",
            set_up_script: &[
                "CREATE TABLE test (a INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM doesnotexist",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "doesnotexist" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Test failed root object lookup does not pin the session to a stale root",
            set_up_script: &[
                "CREATE TABLE test (a INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT doesnotexist()",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function doesnotexist() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE seq",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('seq')",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Test failed query inside a transaction aborts the transaction",
            set_up_script: &[
                "CREATE TABLE test (a INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "START TRANSACTION",
                    expected: Expected::Tag("START TRANSACTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM doesnotexist",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "doesnotexist" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test",
                    expected: Expected::Error(Diagnostic { code: "25P02", message: "current transaction is aborted, commands ignored until end of transaction block", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
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

#[test]
fn test_read_only_transaction_rules() {
    run_scripts(&[
        ScriptTest {
            name: "Read-only transactions",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE test (a INT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN ISOLATION LEVEL SERIALIZABLE, READ WRITE;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN ISOLATION LEVEL REPEATABLE READ, READ ONLY;",
                    expected: Expected::Tag("BEGIN"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25001", message: "there is already a transaction in progress", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_read_only;",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_read_only", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_isolation;",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_isolation", TEXT)],
                        rows: &[
                            &[T("repeatable read")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1);",
                    expected: Expected::Error(Diagnostic { code: "25006", message: "cannot execute INSERT in a read-only transaction", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_read_only;",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_read_only", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN READ ONLY;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t2 (a int);",
                    expected: Expected::Error(Diagnostic { code: "25006", message: "cannot execute CREATE TABLE in a read-only transaction", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_transaction_read_only = on;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (2);",
                    expected: Expected::Error(Diagnostic { code: "25006", message: "cannot execute INSERT in a read-only transaction", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_read_only;",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_read_only", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test;",
                    expected: Expected::Error(Diagnostic { code: "25006", message: "cannot execute DELETE in a read-only transaction", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN READ WRITE;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_transaction_read_only = off;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "START TRANSACTION READ ONLY;",
                    expected: Expected::Tag("START TRANSACTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET a = 4;",
                    expected: Expected::Error(Diagnostic { code: "25006", message: "cannot execute UPDATE in a read-only transaction", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_discard_and_hidden_setting_rules() {
    run_scripts(&[
        ScriptTest {
            name: "DISCARD forms",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE ds;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('ds');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD SEQUENCES;",
                    expected: Expected::Tag("DISCARD SEQUENCES"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT currval('ds');",
                    expected: Expected::Error(Diagnostic { code: "55000", message: r#"currval of sequence "ds" is not yet defined in this session"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lastval();",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "lastval is not yet defined in this session", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD PLANS;",
                    expected: Expected::Tag("DISCARD PLANS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD TEMP;",
                    expected: Expected::Tag("DISCARD TEMP"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD TEMPORARY;",
                    expected: Expected::Tag("DISCARD TEMP"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('ds');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD ALL;",
                    expected: Expected::Tag("DISCARD ALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lastval();",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "lastval is not yet defined in this session", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD PLANS;",
                    expected: Expected::Tag("DISCARD PLANS"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD SEQUENCES;",
                    expected: Expected::Tag("DISCARD SEQUENCES"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DISCARD ALL;",
                    expected: Expected::Error(Diagnostic { code: "25001", message: "DISCARD ALL cannot run inside a transaction block", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Hidden default_with_oids",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET default_with_oids = false;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_with_oids;",
                    expected: Expected::Rows {
                        columns: &[Column("default_with_oids", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_with_oids = true;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "tables declared WITH OIDS are not supported", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, setting, category, short_desc, vartype, context FROM pg_settings WHERE name = 'default_with_oids';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("setting", TEXT), Column("category", TEXT), Column("short_desc", TEXT), Column("vartype", TEXT), Column("context", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET default_with_oids;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
