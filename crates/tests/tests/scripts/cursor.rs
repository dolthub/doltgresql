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
use harness::script::Cell::{Any, Null, Oid, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_cursors() {
    run_scripts(&[
        ScriptTest {
            name: "FETCH and MOVE directions on a SCROLL cursor",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE c SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM c;",
                    expected: Expected::Tag("FETCH 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH NEXT IN c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("3"), T("c")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH FIRST FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH LAST FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 2 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH RELATIVE 2 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH FORWARD 10 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH FORWARD 10 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD 2 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD ALL FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("3"), T("c")],
                            &[T("2"), T("b")],
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE FORWARD 3 IN c;",
                    expected: Expected::Tag("MOVE 3"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE c;",
                    expected: Expected::Tag("MOVE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE ALL IN c;",
                    expected: Expected::Tag("MOVE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH -1 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE -2 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH RELATIVE 0 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 0 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH RELATIVE 0 FROM c;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE BACKWARD ALL FROM c;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CLOSE c;",
                    expected: Expected::Tag("CLOSE CURSOR"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CLOSE c;",
                    expected: Expected::Error(Diagnostic { code: "34000", message: r#"cursor "c" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "cursors without a scroll option can move backward",
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE d CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM d;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM d;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD 1 FROM d;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM d;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                            &[T("3"), T("c")],
                            &[T("4"), T("d")],
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD ALL FROM d;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                            &[T("4"), T("d")],
                            &[T("3"), T("c")],
                            &[T("2"), T("b")],
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 5",
                    },
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
            name: "NO SCROLL cursors only move forward",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE n NO SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE BACKWARD ALL FROM n;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE 0 FROM n;",
                    expected: Expected::Tag("MOVE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 4 FROM n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM n;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "cursor can only scan forward", hint: "Declare it with SCROLL option to enable backward scan.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE n NO SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 2 FROM n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 1 FROM n;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "cursor can only scan forward", hint: "Declare it with SCROLL option to enable backward scan.", ..E }),
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
            name: "re-fetching the current row with a zero count",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE i CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 0 FROM i;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH FORWARD 0 FROM i;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 1 FROM i;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH FORWARD 0 FROM i;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD 0 FROM i;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE 0 FROM i;",
                    expected: Expected::Tag("MOVE 1"),
                    flow: Flow::Exec,
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
            name: "cursors return the rows that existed when they were declared",
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE i CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ct VALUES (6, 'f');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM i;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                            &[T("3"), T("c")],
                            &[T("4"), T("d")],
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 5",
                    },
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
            name: "cursor queries",
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE u CURSOR FOR VALUES (1), (2);",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL u;",
                    expected: Expected::Rows {
                        columns: &[Column("column1", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE w CURSOR FOR WITH x AS (SELECT 1 AS y) SELECT * FROM x;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH w;",
                    expected: Expected::Rows {
                        columns: &[Column("y", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE next CURSOR FOR SELECT 2 AS z;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH NEXT next;",
                    expected: Expected::Rows {
                        columns: &[Column("z", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DECLARE "Mixed Case" CURSOR FOR SELECT 3;"#,
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"FETCH "Mixed Case";"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH mixed;",
                    expected: Expected::Error(Diagnostic { code: "34000", message: r#"cursor "mixed" does not exist"#, ..E }),
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
            name: "cursor errors",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DECLARE c CURSOR FOR SELECT * FROM ct;",
                    expected: Expected::Error(Diagnostic { code: "25P01", message: "DECLARE CURSOR can only be used in transaction blocks", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH x;",
                    expected: Expected::Error(Diagnostic { code: "34000", message: r#"cursor "x" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE x;",
                    expected: Expected::Error(Diagnostic { code: "34000", message: r#"cursor "x" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CLOSE x;",
                    expected: Expected::Error(Diagnostic { code: "34000", message: r#"cursor "x" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CLOSE ALL;",
                    expected: Expected::Tag("CLOSE CURSOR ALL"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE a SCROLL NO SCROLL CURSOR FOR SELECT 1;",
                    expected: Expected::Error(Diagnostic { code: "42P11", message: "cannot specify both SCROLL and NO SCROLL", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE a ASENSITIVE INSENSITIVE CURSOR FOR SELECT 1;",
                    expected: Expected::Error(Diagnostic { code: "42P11", message: "cannot specify both ASENSITIVE and INSENSITIVE", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE c CURSOR FOR SELECT * FROM ct;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE c CURSOR FOR SELECT 1;",
                    expected: Expected::Error(Diagnostic { code: "42P03", message: r#"cursor "c" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE h CURSOR WITH HOLD FOR SELECT * FROM ct FOR UPDATE;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DECLARE CURSOR WITH HOLD ... FOR UPDATE is not supported", detail: "Holdable cursors must be READ ONLY.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE s SCROLL CURSOR FOR SELECT * FROM ct FOR UPDATE;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DECLARE SCROLL CURSOR ... FOR UPDATE is not supported", detail: "Scrollable cursors must be READ ONLY.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE f CURSOR FOR SELECT 1 / 0;",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH f;",
                    expected: Expected::Error(Diagnostic { code: "25P02", message: "current transaction is aborted, commands ignored until end of transaction block", ..E }),
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
            name: "pg_cursors",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE c SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE d NO SCROLL CURSOR WITHOUT HOLD FOR SELECT 1 AS x;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE e INSENSITIVE CURSOR WITH HOLD FOR SELECT * FROM ct;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, statement, is_holdable, is_binary, is_scrollable, creation_time IS NOT NULL FROM pg_cursors WHERE name <> '' ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("statement", TEXT), Column("is_holdable", BOOL), Column("is_binary", BOOL), Column("is_scrollable", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("c"), T("DECLARE c SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;"), T("f"), T("f"), T("t"), T("t")],
                            &[T("d"), T("DECLARE d NO SCROLL CURSOR WITHOUT HOLD FOR SELECT 1 AS x;"), T("f"), T("f"), T("f"), T("t")],
                            &[T("e"), T("DECLARE e INSENSITIVE CURSOR WITH HOLD FOR SELECT * FROM ct;"), T("t"), T("f"), T("t"), T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, statement, is_holdable, is_binary, is_scrollable FROM pg_cursors WHERE name IN ('c', 'd') ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("statement", TEXT), Column("is_holdable", BOOL), Column("is_binary", BOOL), Column("is_scrollable", BOOL)],
                        rows: &[
                            &[T("c"), T("DECLARE c SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;"), T("f"), T("f"), T("t")],
                            &[T("d"), T("DECLARE d NO SCROLL CURSOR WITHOUT HOLD FOR SELECT 1 AS x;"), T("f"), T("f"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_cursors WHERE name <> '';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, statement FROM pg_cursors WHERE name = '';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("statement", TEXT)],
                        rows: &[
                            &[T(""), T("SELECT name, statement FROM pg_cursors WHERE name = '';")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "WITH HOLD cursors outlive their transaction",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DECLARE h CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, is_holdable FROM pg_cursors WHERE name <> '';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("is_holdable", BOOL)],
                        rows: &[
                            &[T("h"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE h2 CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE w CURSOR FOR SELECT 1;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 1 FROM h2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, is_holdable FROM pg_cursors WHERE name <> '' ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("is_holdable", BOOL)],
                        rows: &[
                            &[T("h"), T("t")],
                            &[T("h2"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM h2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("b")],
                            &[T("3"), T("c")],
                            &[T("4"), T("d")],
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM h2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("3"), T("c")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE h3 CURSOR WITH HOLD FOR SELECT 1;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CLOSE h;",
                    expected: Expected::Tag("CLOSE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_cursors WHERE name <> '' ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("h2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 / 0;",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_cursors WHERE name <> '' ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("h2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CLOSE ALL;",
                    expected: Expected::Tag("CLOSE CURSOR ALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE e CURSOR WITH HOLD FOR SELECT 1 / 0;",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_cursors WHERE name <> '';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "cursors in the implicit transaction of a multi-statement query",
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DECLARE c CURSOR FOR SELECT * FROM ct; FETCH 2 FROM c;",
                    expected: Expected::Tag("FETCH 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_cursors WHERE name <> '';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROLLBACK TO SAVEPOINT closes cursors declared after the savepoint",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE s SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SAVEPOINT sp;",
                    expected: Expected::Tag("SAVEPOINT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE t CURSOR FOR SELECT 1;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 1 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("3"), T("c")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK TO SAVEPOINT sp;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_cursors WHERE name <> '' ORDER BY name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("s")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 1 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 1",
                    },
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
            name: "BINARY cursors",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE b BINARY CURSOR FOR SELECT 1;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, is_binary FROM pg_cursors WHERE name <> '';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("is_binary", BOOL)],
                        rows: &[
                            &[T("b"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
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
            name: "cursors read their rows as they are fetched",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE r CURSOR FOR SELECT 10 / (id - 3) FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM r;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("-5")],
                            &[T("-10")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 1 FROM r;",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
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
            name: "cursors read the data as it was when they were declared",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE za (id INT4 PRIMARY KEY, v INT4);",
                "CREATE TABLE zb (id INT4 PRIMARY KEY, a_id INT4, w TEXT);",
                "CREATE INDEX zb_a ON zb (a_id);",
                "INSERT INTO za VALUES (1, 10), (2, 20), (3, 30);",
                "INSERT INTO zb VALUES (1, 1, 'x'), (2, 2, 'y'), (3, 3, 'z');",
                "CREATE VIEW zv AS SELECT za.id, zb.w FROM za JOIN zb ON zb.a_id = za.id;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE j CURSOR FOR SELECT za.id, zb.w FROM za JOIN zb ON zb.a_id = za.id ORDER BY za.id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE s CURSOR FOR SELECT id FROM za WHERE v IN (SELECT v FROM za WHERE v > 10) ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE vw CURSOR FOR SELECT * FROM zv ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE cte CURSOR FOR WITH x AS (SELECT id, v FROM za) SELECT id, (SELECT max(v) FROM za) FROM x ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE lk CURSOR FOR SELECT w FROM zb WHERE a_id = 2;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 1 FROM j;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("w", TEXT)],
                        rows: &[
                            &[T("1"), T("x")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE zb SET w = 'changed';",
                    expected: Expected::Tag("UPDATE 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO za VALUES (4, 40);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO zb VALUES (4, 4, 'new'), (5, 2, 'dup');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM za WHERE id = 3;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM j;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("w", TEXT)],
                        rows: &[
                            &[T("2"), T("y")],
                            &[T("3"), T("z")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM vw;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("w", TEXT)],
                        rows: &[
                            &[T("1"), T("x")],
                            &[T("2"), T("y")],
                            &[T("3"), T("z")],
                        ],
                        tag: "FETCH 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM cte;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("max", INT4)],
                        rows: &[
                            &[T("1"), T("30")],
                            &[T("2"), T("30")],
                            &[T("3"), T("30")],
                        ],
                        tag: "FETCH 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM lk;",
                    expected: Expected::Rows {
                        columns: &[Column("w", TEXT)],
                        rows: &[
                            &[T("y")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD ALL FROM j;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "cursor can only scan forward", hint: "Declare it with SCROLL option to enable backward scan.", ..E }),
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
            name: "WITH HOLD cursors read their remaining rows when their transaction commits",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE h CURSOR WITH HOLD FOR SELECT 10 / (id - 3) FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 1 FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("-5")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ct VALUES (9, 'z');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_cursors WHERE name <> '';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM ct;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE w SCROLL CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM w;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM w;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM w;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("b")],
                            &[T("3"), T("c")],
                            &[T("4"), T("d")],
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD 2 FROM w;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE w2 NO SCROLL CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM w2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM w2;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "cursor can only scan forward", hint: "Declare it with SCROLL option to enable backward scan.", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "cursor queries with parameters",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE p CURSOR FOR SELECT $1::INT4 AS x;",
                    bind_vars: &[BindVar::Int(5)],
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH p;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "FETCH 1",
                    },
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
            name: "cursors over historical data",
            set_up_script: &[
                "CREATE TABLE hist (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO hist VALUES (1, 'one');",
                "SELECT dolt_commit('-Am', 'first');",
                "INSERT INTO hist VALUES (2, 'two');",
                "UPDATE hist SET v = 'ONE' WHERE id = 1;",
                "SELECT dolt_commit('-am', 'second');",
                "DELETE FROM hist WHERE id = 1;",
                "INSERT INTO hist VALUES (3, 'three');",
                "SELECT dolt_commit('-am', 'third');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE wh CURSOR WITH HOLD FOR SELECT * FROM hist AS OF 'HEAD~1' ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE h2 CURSOR FOR SELECT * FROM hist AS OF 'HEAD~2' ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE h1 SCROLL CURSOR FOR SELECT * FROM hist AS OF 'HEAD~1' ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE head CURSOR FOR SELECT * FROM hist AS OF 'HEAD' ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE j CURSOR FOR SELECT cur.id, cur.v, old.v FROM hist cur JOIN hist AS OF 'HEAD~1' old ON cur.id = old.id ORDER BY cur.id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE hh CURSOR FOR SELECT id, v FROM dolt_history_hist ORDER BY v, id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO hist VALUES (4, 'four');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "UPDATE hist SET v = 'TWO';",
                    expected: Expected::Tag("UPDATE 3"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'fourth');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM h2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("one")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH 1 FROM h1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("ONE")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM h1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("two")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH BACKWARD ALL FROM h1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("two")],
                            &[T("1"), T("ONE")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM head;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("two")],
                            &[T("3"), T("three")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM j;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("two"), T("two")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM hh;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("ONE")],
                            &[T("1"), T("one")],
                            &[T("3"), T("three")],
                            &[T("2"), T("two")],
                            &[T("2"), T("two")],
                        ],
                        tag: "FETCH 5",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM wh;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("ONE")],
                            &[T("2"), T("two")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "cursors over table functions and system tables",
            set_up_script: &[
                "CREATE TABLE tf (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO tf VALUES (1, 'a');",
                "SELECT dolt_commit('-Am', 'first');",
                "INSERT INTO tf VALUES (2, 'b');",
                "SELECT dolt_commit('-am', 'second');",
                "UPDATE tf SET v = 'c' WHERE id = 1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE lg CURSOR FOR SELECT message FROM dolt_log() LIMIT 2;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE df CURSOR FOR SELECT to_id, to_v, from_id, from_v, diff_type FROM dolt_diff('HEAD~1', 'HEAD', 'tf');",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE dw CURSOR FOR SELECT to_id, to_v, from_v, diff_type FROM dolt_diff('HEAD', 'WORKING', 'tf');",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE ds CURSOR FOR SELECT table_name, rows_modified FROM dolt_diff_stat('HEAD', 'WORKING');",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE dp CURSOR FOR SELECT table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING');",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE st CURSOR FOR SELECT table_name, staged, status FROM dolt_status;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE dd CURSOR FOR SELECT to_id, to_v, from_v, diff_type FROM dolt_diff_tf WHERE to_commit = 'WORKING';",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DECLARE gs SCROLL CURSOR FOR SELECT * FROM generate_series(1, 5);",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "UPDATE tf SET v = 'z' WHERE id = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'third');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM lg;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("second")],
                            &[T("first")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM df;",
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT4), Column("to_v", TEXT), Column("from_id", INT4), Column("from_v", TEXT), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("2"), T("b"), Null, Null, T("added")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM dw;",
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT4), Column("to_v", TEXT), Column("from_v", TEXT), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("1"), T("c"), T("a"), T("modified")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM ds;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("rows_modified", INT8)],
                        rows: &[
                            &[T("public.tf"), T("1")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM dp;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("diff_type", TEXT), Column("statement", TEXT)],
                        rows: &[
                            &[T("public.tf"), T("data"), T(r#"UPDATE "tf" SET "v"='c' WHERE "id"=1;"#)],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM st;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.tf"), T("f"), T("modified")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH ALL FROM dd;",
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT4), Column("to_v", TEXT), Column("from_v", TEXT), Column("diff_type", TEXT)],
                        rows: &[
                            &[T("1"), T("c"), T("a"), T("modified")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH 3 FROM gs;",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "FETCH 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "FETCH BACKWARD 2 FROM gs;",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("1")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "moving and rewinding a SCROLL cursor",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE s SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 7 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH RELATIVE -10 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH NEXT FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH LAST FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH NEXT FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH NEXT FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE ABSOLUTE 0 IN s;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH NEXT FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH RELATIVE 3 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD ALL FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("3"), T("c")],
                            &[T("2"), T("b")],
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE FORWARD ALL IN s;",
                    expected: Expected::Tag("MOVE 5"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD 2 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE -5 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE -6 FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH NEXT FROM s;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE 0 IN s;",
                    expected: Expected::Tag("MOVE 1"),
                    flow: Flow::Exec,
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
            name: "moving and rewinding a SCROLL WITH HOLD cursor after its transaction commits",
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE h SCROLL CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH RELATIVE -1 FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 7 FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE BACKWARD ALL IN h;",
                    expected: Expected::Tag("MOVE 4"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH NEXT FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH LAST FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE FORWARD 2 IN h;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD 1 FROM h;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CLOSE h;",
                    expected: Expected::Tag("CLOSE CURSOR"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "moving a NO SCROLL WITH HOLD cursor",
            skip: Some("Postgres-style portals that read rows as they are fetched are deferred; DECLARE still reads every row"),
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DECLARE n NO SCROLL CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE BACKWARD ALL IN n;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 2 FROM n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH RELATIVE 2 FROM n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE 0 IN n;",
                    expected: Expected::Tag("MOVE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 4 FROM n;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "cursor can only scan forward", hint: "Declare it with SCROLL option to enable backward scan.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CLOSE ALL;",
                    expected: Expected::Tag("CLOSE CURSOR ALL"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "moving through a cursor without rows",
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE e SCROLL CURSOR FOR SELECT * FROM ct WHERE id > 10;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH e;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM e;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE ALL IN e;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE -1 FROM e;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ABSOLUTE 1 FROM e;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE 0 IN e;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
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
            name: "moving a cursor without a scroll option past its last row",
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE f CURSOR FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH FORWARD 2 FROM f;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE 1 IN f;",
                    expected: Expected::Tag("MOVE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM f;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("4"), T("d")],
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM f;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "FETCH 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE FORWARD 0 IN f;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "MOVE FORWARD 3 IN f;",
                    expected: Expected::Tag("MOVE 0"),
                    flow: Flow::Exec,
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
            name: "scrolling WITH HOLD cursors after their transaction commits",
            set_up_script: &[
                "CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
                "INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE w SCROLL CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM w;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM w;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "FETCH 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH ALL FROM w;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("2"), T("b")],
                            &[T("3"), T("c")],
                            &[T("4"), T("d")],
                            &[T("5"), T("e")],
                        ],
                        tag: "FETCH 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH BACKWARD 2 FROM w;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("5"), T("e")],
                            &[T("4"), T("d")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DECLARE w2 NO SCROLL CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
                    expected: Expected::Tag("DECLARE CURSOR"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH 2 FROM w2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "FETCH 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "FETCH PRIOR FROM w2;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "cursor can only scan forward", hint: "Declare it with SCROLL option to enable backward scan.", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
