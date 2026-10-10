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
fn test_set_local() {
    run_scripts(&[
        ScriptTest {
            name: "SET LOCAL reverts on COMMIT",
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET LOCAL reverts on ROLLBACK",
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET LOCAL reverts to the session value, not the default",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET enable_hashjoin = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = on",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin = on",
                    expected: Expected::Tag("SET"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET LOCAL reverts when a failed transaction is rolled back",
            set_up_script: &[
                "CREATE TABLE test (a INT PRIMARY KEY)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT no_such_column FROM test",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "no_such_column" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Error(Diagnostic { code: "25P02", message: "current transaction is aborted, commands ignored until end of transaction block", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET LOCAL outside a transaction block has no lasting effect",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = off",
                    expected: Expected::Tag("SET"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET LOCAL can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET LOCAL with savepoints does not abort the transaction",
            set_up_script: &[
                "CREATE TABLE test (a INT PRIMARY KEY)",
                "INSERT INTO test VALUES (1)",
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
                    query: "SET LOCAL enable_hashjoin = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_mergejoin = on",
                    expected: Expected::Tag("SET"),
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
                    query: "ROLLBACK TO settings",
                    expected: Expected::Tag("ROLLBACK"),
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
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET after SET LOCAL persists after COMMIT",
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin = off",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin = on",
                    expected: Expected::Tag("SET"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SET LOCAL on an unknown parameter errors",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET LOCAL no_such_parameter = on",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "no_such_parameter""#, ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "SET LOCAL can only be used in transaction blocks", ..E }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set_config with is_local reverts on COMMIT",
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('enable_seqscan', 'off', true)",
                    expected: Expected::Rows {
                        columns: &[Column("set_config", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_seqscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_seqscan", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_seqscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_seqscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_set_local_and_order_rules() {
    run_scripts(&[
        ScriptTest {
            name: "SET after SET LOCAL",
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = off;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin = off;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin;",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin = on;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = off;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin = off;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin;",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin = off;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL enable_hashjoin = on;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin;",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin;",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL work_mem = '1MB';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET work_mem = '2MB';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL work_mem = '3MB';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW work_mem;",
                    expected: Expected::Rows {
                        columns: &[Column("work_mem", TEXT)],
                        rows: &[
                            &[T("3MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW work_mem;",
                    expected: Expected::Rows {
                        columns: &[Column("work_mem", TEXT)],
                        rows: &[
                            &[T("2MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb index order",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE jorder (id int primary key, val jsonb);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX jorder_val ON jorder (val);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO jorder VALUES (1,'null'), (2,'[]'), (3,'{}'), (4,'"a"'), (5,'1'), (6,'true'), (7,'[1]'), (8,'false'), (9,'{"a":1}'), (10, '10'), (11, '9');"#,
                    expected: Expected::Tag("INSERT 0 11"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM jorder WHERE val < '2' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("val", JSONB)],
                        rows: &[
                            &[T("null")],
                            &[T("[]")],
                            &[T(r#""a""#)],
                            &[T("1")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT val FROM jorder WHERE val > '"a"' ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("val", JSONB)],
                        rows: &[
                            &[T("{}")],
                            &[T("1")],
                            &[T("true")],
                            &[T("[1]")],
                            &[T("false")],
                            &[T(r#"{"a": 1}"#)],
                            &[T("10")],
                            &[T("9")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM jorder WHERE val BETWEEN '1' AND '10' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("val", JSONB)],
                        rows: &[
                            &[T("1")],
                            &[T("10")],
                            &[T("9")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM jorder WHERE val = '10' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("val", JSONB)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT val FROM jorder ORDER BY val;",
                    expected: Expected::Rows {
                        columns: &[Column("val", JSONB)],
                        rows: &[
                            &[T("[]")],
                            &[T("null")],
                            &[T(r#""a""#)],
                            &[T("1")],
                            &[T("9")],
                            &[T("10")],
                            &[T("false")],
                            &[T("true")],
                            &[T("[1]")],
                            &[T("{}")],
                            &[T(r#"{"a": 1}"#)],
                        ],
                        tag: "SELECT 11",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "References to missing databases",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM no_such_db.public.tbl;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "no_such_db.public.tbl""#, position: 15, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO no_such_db.public.tbl VALUES (1);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cross-database references are not implemented: "no_such_db.public.tbl""#, position: 13, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
