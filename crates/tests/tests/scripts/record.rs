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
fn test_records() {
    run_scripts(&[
        ScriptTest {
            name: "Record cannot be used as column type",
            set_up_script: &[
                "CREATE TABLE t2 (pk INT PRIMARY KEY, c1 VARCHAR(100));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t (pk INT PRIMARY KEY, r RECORD);",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"column "r" has pseudo-type record"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t2 ADD COLUMN c2 RECORD;",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"column "c2" has pseudo-type record"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t2 ALTER COLUMN c1 TYPE RECORD;",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"column "c1" has pseudo-type record"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DOMAIN my_domain AS record;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#""record" is not a valid base type for a domain"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE my_seq AS record;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "sequence type must be smallint, integer, or bigint", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE outer_type AS (id int, payload record);",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"column "payload" has pseudo-type record"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Casting to record",
            assertions: &[
                ScriptTestAssertion {
                    query: "select row(1, 1)::record;",
                    expected: Expected::Rows {
                        columns: &[Column("row", RECORD)],
                        rows: &[
                            &[T("(1,1)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() wrapping table rows",
            set_up_script: &[
                "create table users (name text, location text, age int);",
                "insert into users values ('jason', 'SEA', 42), ('max', 'SFO', 31);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select row(p) from users p;",
                    expected: Expected::Rows {
                        columns: &[Column("row", RECORD)],
                        rows: &[
                            &[T(r#"("(jason,SEA,42)")"#)],
                            &[T(r#"("(max,SFO,31)")"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select row(p.*, 42) from users p;",
                    expected: Expected::Rows {
                        columns: &[Column("row", RECORD)],
                        rows: &[
                            &[T("(jason,SEA,42,42)")],
                            &[T("(max,SFO,31,42)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (u).location FROM users u;",
                    expected: Expected::Rows {
                        columns: &[Column("location", TEXT)],
                        rows: &[
                            &[T("SEA")],
                            &[T("SFO")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() wrapping values",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2, 3) as myRow;",
                    expected: Expected::Rows {
                        columns: &[Column("myrow", RECORD)],
                        rows: &[
                            &[T("(1,2,3)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (4, 5, 6) as myRow;",
                    expected: Expected::Rows {
                        columns: &[Column("myrow", RECORD)],
                        rows: &[
                            &[T("(4,5,6)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (NULL, 'foo', NULL) as myRow;",
                    expected: Expected::Rows {
                        columns: &[Column("myrow", RECORD)],
                        rows: &[
                            &[T("(,foo,)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (NULL, (1 > 0), 'baz') as myRow;",
                    expected: Expected::Rows {
                        columns: &[Column("myrow", RECORD)],
                        rows: &[
                            &[T("(,t,baz)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() equality and comparison",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 'x') = ROW(1, 'x');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 'x') = ROW(1, 'y');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL) = ROW(1, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) < ROW(1, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) < ROW(2, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(2, 2) < ROW(2, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(2, 2, 1) < ROW(2, NULL, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) < ROW(NULL, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL, NULL, NULL) < ROW(NULL, NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) <= ROW(1, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) <= ROW(1, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL) <= ROW(1, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(2, 1) > ROW(1, 999);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(2, 1) > ROW(1, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(2, 1) >= ROW(1, 999);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(2, 1) >= ROW(2, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL, 1) >= ROW(2, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) != ROW(3, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) != ROW(NULL, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL, 4) != ROW(NULL, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL) IS NOT DISTINCT FROM ROW(1, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, '2') = ROW(1, 2::TEXT);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 1) = ROW(1, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 1) != ROW(1, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 1) < ROW(1, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 1) <= ROW(1, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 1) > ROW(1, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 1) = ROW(1, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2, null, 4) = ROW(null, 2, 3, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2, null, 4) != ROW(null, 2, 3, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() use inserting and selecting composite rows",
            set_up_script: &[
                "CREATE TYPE user_info AS (id INT, name TEXT, email TEXT);",
                "CREATE TABLE accounts (info user_info);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO accounts VALUES (ROW(1, 'alice', 'a@example.com'));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT info FROM accounts;",
                    expected: Expected::Rows {
                        columns: &[Column("info", USER_DEFINED)],
                        rows: &[
                            &[T("(1,alice,a@example.com)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (a.info).name FROM accounts a;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("alice")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() use in WHERE clause",
            set_up_script: &[
                "create table users (id int primary key, name text, email text);",
                "insert into users values (1, 'John', 'j@a.com'), (2, 'Joe', 'joe@joe.com');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM users WHERE ROW(id, name, email) = ROW(1, 'John', 'j@a.com');",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("email", TEXT)],
                        rows: &[
                            &[T("1"), T("John"), T("j@a.com")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM users WHERE ROW(id, name) IS NOT DISTINCT FROM ROW(2, 'Jane');",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("email", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() casting and type inference",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 'a')::record;",
                    expected: Expected::Rows {
                        columns: &[Column("row", RECORD)],
                        rows: &[
                            &[T("(1,a)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = ROW(1, 'two');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "two""#, position: 27, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = ROW(1, '2');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() error cases and edge conditions",
            set_up_script: &[
                "create table users (id int primary key, name text, email text);",
                "insert into users values (1, 'John', 'j@a.com'), (2, 'Joe', 'joe@joe.com');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = ROW(1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = ROW(1, 2, 3);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) < ROW(1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) <= ROW(1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) > ROW(1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) >= ROW(1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) != ROW(1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT NULL::record IS NULL",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL) IS NULL",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL, NULL, NULL) IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL, 42, NULL) IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(42) IS NULL",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL) IS NOT NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL, NULL) IS NOT NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL, 1) IS NOT NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 1) IS NOT NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(42) IS NOT NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(id, name), COUNT(*) FROM users GROUP BY ROW(id, name);",
                    expected: Expected::Rows {
                        columns: &[Column("row", RECORD), Column("count", INT8)],
                        rows: &[
                            &[T("(1,John)"), T("1")],
                            &[T("(2,Joe)"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() nesting",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(ROW(1, 'x'), true);",
                    expected: Expected::Rows {
                        columns: &[Column("row", RECORD)],
                        rows: &[
                            &[T(r#"("(1,x)",t)"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() NULL handling depends on the comparison context",
            set_up_script: &[
                "CREATE TYPE ct AS (a INT4, b INT4);",
                "CREATE TABLE ctt (id INT4, c ct);",
                "INSERT INTO ctt VALUES (1, ROW(1, NULL)), (2, ROW(1, 2)), (3, ROW(NULL, NULL)), (4, NULL);",
                "CREATE TABLE rf (id INT4 PRIMARY KEY, a INT4, b INT4);",
                "CREATE INDEX rfi ON rf (a, b);",
                "INSERT INTO rf VALUES (1, 1, NULL), (2, 1, 2), (3, 2, 1);",
                "CREATE TABLE ck (a INT4, b INT4, CHECK ((a, b) < (5, 5)));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL::INT4) = ROW(NULL::INT4), ROW(NULL::INT4) = ANY(ARRAY[ROW(NULL::INT4)]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[Null, T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL::INT4) = ANY(ARRAY[ROW(1, NULL::INT4)]), ROW(1, NULL::INT4) <> ALL(ARRAY[ROW(1, NULL::INT4)]), ROW(1, NULL::INT4) < ANY(ARRAY[ROW(1, 2)]), ROW(1, 2) < ANY(ARRAY[ROW(1, NULL::INT4)]), ROW(1, NULL::INT4) > ANY(ARRAY[ROW(1, 2)]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("f"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL::INT4) IS DISTINCT FROM ROW(1, NULL::INT4), ROW(1, NULL::INT4) = ALL(ARRAY[ROW(1, NULL::INT4)]), ROW(1, NULL::INT4) <= ANY(ARRAY[ROW(1, NULL::INT4)]), ROW(NULL::INT4, 1) < ANY(ARRAY[ROW(1, 1)]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("t"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(ROW(NULL::INT4)) = ROW(ROW(NULL::INT4)), ARRAY[ROW(NULL::INT4)] = ARRAY[ROW(NULL::INT4)];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT record_eq(ROW(NULL::INT4), ROW(NULL::INT4)), record_lt(ROW(NULL::INT4), ROW(1)), record_gt(ROW(NULL::INT4), ROW(1)), record_ne(ROW(1, NULL::INT4), ROW(1, NULL::INT4));",
                    expected: Expected::Rows {
                        columns: &[Column("record_eq", BOOL), Column("record_lt", BOOL), Column("record_gt", BOOL), Column("record_ne", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL)::ct = ROW(1, NULL)::ct, (ROW(1, NULL::INT4)) = (ROW(1, NULL::INT4)), (1, NULL::INT4) = (1, NULL::INT4), ROW(1, NULL)::ct = ROW(1, NULL::INT4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), Null, Null, T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL::INT4) < ROW(1, NULL::INT4), ROW(NULL::INT4, 1) <= ROW(NULL::INT4, 1), ROW(1, 2, NULL::INT4) >= ROW(1, 1, NULL::INT4), ROW(1, 2, 3) > ROW(1, 2, NULL::INT4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[Null, Null, T("t"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL::INT4) IN (ROW(1, NULL::INT4), ROW(2, 3)), ROW(1, NULL::INT4) NOT IN (ROW(1, NULL::INT4), ROW(2, 3)), ROW(1, NULL::INT4) IN (ROW(2, NULL::INT4), ROW(2, 3));",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[Null, Null, T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) IN (ROW(1, 2)), ROW(1, 2) NOT IN (ROW(1, 3), ROW(2, 2));",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) IN (ROW(1, 2), ROW(1));",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c = c, c = ROW(1, NULL)::ct, c < ROW(1, 3)::ct, c > ROW(1, 3)::ct, c <> ROW(1, NULL)::ct, c >= ROW(NULL, NULL)::ct FROM ctt ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("t"), T("t"), T("f"), T("t"), T("f"), T("f")],
                            &[T("2"), T("t"), T("f"), T("t"), T("f"), T("t"), T("f")],
                            &[T("3"), T("t"), T("f"), T("f"), T("t"), T("t"), T("t")],
                            &[T("4"), Null, Null, Null, Null, Null, Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, ROW(1, NULL::INT4) = c, c = ROW(1, NULL::INT4), c IN (ROW(1, NULL::INT4)), ROW(1, NULL::INT4) IN (c) FROM ctt ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("t"), T("t"), T("t"), T("t")],
                            &[T("2"), T("f"), T("f"), T("f"), T("f")],
                            &[T("3"), T("f"), T("f"), T("f"), T("f")],
                            &[T("4"), Null, Null, Null, Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ctt WHERE c = ANY(ARRAY[ROW(1, NULL)::ct, ROW(NULL, NULL)::ct]) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ctt WHERE c IN (ROW(1, NULL)::ct, ROW(NULL, NULL)::ct) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM ctt WHERE c IN (SELECT c FROM ctt WHERE id = 1) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE ROW(ROW(a, b)) = ROW(ROW(1, NULL::INT4)) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE ROW(a, b) = ROW(1, 2) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE (a, b) < (1, 3) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE (a, b) >= (1, 2) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE (a, b) > (1, 1) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE (a, b) <= (2, 0) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE (a, b) <> (1, 2) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE (a, b) IN ((1, 2), (2, 1)) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE (a, b) NOT IN ((1, 2), (5, 5)) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW() = ROW();",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "cannot compare rows of zero length", position: 14, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ck VALUES (1, 9);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ck VALUES (5, 6);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "ck" violates check constraint "ck_check""#, detail: "Failing row contains (5, 6).", schema: "public", table: "ck", constraint: "ck_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ck;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() compared to subqueries",
            set_up_script: &[
                "CREATE TABLE sq (x INT4, y INT4);",
                "INSERT INTO sq VALUES (1, 2), (1, NULL), (3, 4);",
                "CREATE TABLE rf (id INT4 PRIMARY KEY, a INT4, b INT4);",
                "CREATE INDEX rfi ON rf (a, b);",
                "INSERT INTO rf VALUES (1, 1, NULL), (2, 1, 2), (3, 2, 1);",
                "CREATE TABLE ck (a INT4, b INT4, CHECK (ROW(a, b) IS DISTINCT FROM ROW(1, 1)));",
                "CREATE SEQUENCE rseq;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL::INT4) IN (SELECT 1, NULL::INT4), ROW(1, 2) IN (SELECT 1, 2), ROW(1, 2) IN (SELECT 1, 3), ROW(1, 2) NOT IN (SELECT 1, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[Null, T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL::INT4) = ANY(SELECT 1, NULL::INT4), ROW(1, 2) = ANY(SELECT 1, 2), ROW(1, 2) < ANY(SELECT 1, 3), ROW(1, 2) < ALL(SELECT 1, 1), ROW(1, 2) <> ALL(SELECT 1, NULL::INT4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[Null, T("t"), T("t"), T("f"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL::INT4) = (SELECT 1, NULL::INT4), ROW(1, 2) = (SELECT 1, 2), ROW(1, 2) < (SELECT 1, 3), ROW(1, 2) <> (SELECT 1, 2), ROW(1, 2) >= (SELECT 1, NULL::INT4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[Null, T("t"), T("t"), T("f"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) IN (SELECT x, y FROM sq), ROW(1, 5) IN (SELECT x, y FROM sq), ROW(9, 9) IN (SELECT x, y FROM sq), ROW(1, 2) = ANY(SELECT x, y FROM sq), ROW(1, 5) < ALL(SELECT x, y FROM sq), ROW(0, 0) < ALL(SELECT x, y FROM sq);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), Null, T("f"), T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = (SELECT x, y FROM sq WHERE x = 3), ROW(1, 2) = (SELECT x, y FROM sq WHERE x = 99), ROW(1, 2) IN (SELECT x, y FROM sq WHERE x = 99), ROW(1, 2) = ALL(SELECT x, y FROM sq WHERE x = 99);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), Null, T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x, y, ROW(x, y) IN (SELECT 1, 2), ROW(x, y) < (SELECT 2, 0) FROM sq ORDER BY x, y;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("y", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("2"), T("t"), T("t")],
                            &[T("1"), Null, Null, T("t")],
                            &[T("3"), T("4"), T("f"), T("f")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x FROM sq WHERE ROW(x, y) IN (SELECT 1, y FROM sq WHERE y IS NOT NULL) ORDER BY x;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1) IN (SELECT 1), ROW(NULL::INT4) = ANY(SELECT NULL::INT4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = (SELECT x, y FROM sq);",
                    expected: Expected::Error(Diagnostic { code: "21000", message: "more than one row returned by a subquery used as an expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) IN (SELECT x FROM sq);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "subquery has too few columns", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = ANY(SELECT 1, 2, 3);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "subquery has too many columns", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = (SELECT 1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "subquery has too few columns", position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL::INT4) = (SELECT ROW(NULL::INT4));",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "operator does not exist: integer = record", hint: "No operator matches the given name and argument types. You might need to add explicit type casts.", position: 24, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE ROW(ROW(a, b)) < ROW(ROW(1, 3)) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, (a, b) < (1, 3), (a, b) >= (1, 2) FROM rf ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), Null, Null],
                            &[T("2"), T("t"), T("t")],
                            &[T("3"), T("f"), T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(nextval('rseq'), 1) < ROW(100, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('rseq');",
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
                    query: "INSERT INTO ck VALUES (1, 2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ck VALUES (1, 1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "ck" violates check constraint "ck_check""#, detail: "Failing row contains (1, 1).", schema: "public", table: "ck", constraint: "ck_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ck;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o.id FROM rf o WHERE ROW(o.a, o.b) = (SELECT i.a, i.b FROM rf i WHERE i.id = o.id) ORDER BY o.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o.id FROM rf o WHERE ROW(o.a, o.b) >= (SELECT i.a, i.b FROM rf i WHERE i.id = o.id) ORDER BY o.id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ROW() comparisons use indexes",
            set_up_script: &[
                "SET enable_seqscan = off;",
                "CREATE TABLE rf (id INT4 PRIMARY KEY, a INT4, b INT4);",
                "CREATE INDEX rfi ON rf (a, b);",
                "INSERT INTO rf VALUES (1, 1, NULL), (2, 1, 2), (3, 2, 1);",
                "CREATE TABLE sq (x INT4, y INT4);",
                "INSERT INTO sq VALUES (1, 2), (1, NULL), (3, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM rf WHERE (a, b) = (1, 2);",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{[1, 1], [2, 2]}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM rf WHERE (a, b) <> (1, 2);",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{[NULL, NULL], (NULL, 2)}, {[NULL, NULL], (2, ∞)}, {(NULL, 1), [NULL, ∞)}, {[1, 1], (NULL, 2)}, {[1, 1], (2, ∞)}, {(1, ∞), [NULL, ∞)}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM rf WHERE (a, b) < (1, 3);",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{(NULL, 1), [NULL, ∞)}, {[1, 1], (NULL, 3)}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM rf WHERE (a, b) <= (2, 0);",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{(NULL, 2), [NULL, ∞)}, {[2, 2], (NULL, 0]}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM rf WHERE (a, b) > (1, 1);",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{[1, 1], (1, ∞)}, {(1, ∞), [NULL, ∞)}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM rf WHERE (a, b) >= (1, 2);",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{[1, 1], [2, ∞)}, {(1, ∞), [NULL, ∞)}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM rf WHERE (a, b) IN ((1, 2), (2, 1));",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{[1, 1], [2, 2]}, {[2, 2], [1, 1]}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT id FROM (SELECT id FROM rf WHERE (a, b) < (1, 3)) s;",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{(NULL, 1), [NULL, ∞)}, {[1, 1], (NULL, 3)}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN WITH c AS (SELECT id FROM rf WHERE (a, b) > (1, 1)) SELECT id FROM c;",
                    expected: Expected::Plan(&[PlanFact::IndexScan { table: "rf", columns: &["a", "b"], ranges: "[{[1, 1], (1, ∞)}, {(1, ∞), [NULL, ∞)}]" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "EXPLAIN SELECT sq.x, sq.y, rf.id FROM sq JOIN rf ON (rf.a, rf.b) > (sq.x, sq.y);",
                    expected: Expected::Plan(&[PlanFact::Join { kind: "InnerJoin", left: "rf", right: "sq" }, PlanFact::FullScan { table: "rf" }, PlanFact::FullScan { table: "sq" }]),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM (SELECT id FROM rf WHERE (a, b) < (1, 3)) s ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH c AS (SELECT id FROM rf WHERE (a, b) > (1, 1)) SELECT id FROM c ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sq.x, sq.y, rf.id FROM sq JOIN rf ON (rf.a, rf.b) > (sq.x, sq.y) ORDER BY sq.x, sq.y, rf.id;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("y", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                            &[T("1"), Null, T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rf WHERE id IN (SELECT id FROM rf WHERE (a, b) >= (1, 2)) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_row_comparison_rules() {
    run_scripts(&[
        ScriptTest {
            name: "row constructor comparisons",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = ROW(1, 2), ROW(1, NULL) = ROW(1, 1), ROW(1, NULL) = ROW(2, 1), ROW(1, NULL) <> ROW(2, 1), ROW(NULL, 4) <> ROW(NULL, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), Null, T("f"), T("t"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) < ROW(1, 3), ROW(2, 2) < ROW(2, NULL), ROW(1, 2) < ROW(NULL, 3), ROW(1, NULL) <= ROW(1, 2), ROW(NULL, 1) >= ROW(2, 1), (1, 2) >= (1, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), Null, Null, Null, Null, T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = ROW(1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) < ROW(1, 2, 3);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW() = ROW();",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "cannot compare rows of zero length", position: 14, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) IS DISTINCT FROM ROW(1, NULL), ROW(1, NULL) IS NOT DISTINCT FROM ROW(1, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(NULL) IS NULL, ROW(NULL, NULL) IS NULL, ROW(NULL, 1) IS NULL, ROW(NULL, 1) IS NOT NULL, ROW(1, 2) IS NOT NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("f"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL::INT4) IN (ROW(1, NULL::INT4), ROW(2, 3)), ROW(1, 2) IN (ROW(1, 2), ROW(3, 4)), ROW(1, 2) NOT IN (ROW(5, 6));",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[Null, T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) IN (ROW(1, 2), ROW(1));",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unequal number of entries in row expressions", position: 18, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT record_eq(ROW(NULL::INT4), ROW(NULL::INT4)), record_lt(ROW(NULL::INT4), ROW(1)), record_gt(ROW(NULL::INT4), ROW(1));",
                    expected: Expected::Rows {
                        columns: &[Column("record_eq", BOOL), Column("record_lt", BOOL), Column("record_gt", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "rows compared to subqueries and composite values",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE sq (x INT, y INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO sq VALUES (1, 2), (1, NULL), (3, 4);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) IN (SELECT x, y FROM sq), ROW(3, 5) IN (SELECT x, y FROM sq), ROW(1, 2) = ANY (SELECT x, y FROM sq), ROW(1, 3) < ALL (SELECT x, y FROM sq WHERE y IS NOT NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = (SELECT 1, 2), ROW(1, 2) < (SELECT 1, 3), ROW(1, NULL::INT4) = (SELECT 1, NULL::INT4), ROW(1, 2) = (SELECT x, y FROM sq WHERE x = 99);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) = (SELECT x, y FROM sq);",
                    expected: Expected::Error(Diagnostic { code: "21000", message: "more than one row returned by a subquery used as an expression", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, 2) IN (SELECT x FROM sq);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "subquery has too few columns", position: 18, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x, y, (x, y) IN (SELECT 1, 2), (x, y) >= (1, 3) FROM sq ORDER BY x, y;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("y", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("2"), T("t"), T("f")],
                            &[T("1"), Null, Null, Null],
                            &[T("3"), T("4"), T("f"), T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TYPE ct AS (a INT, b INT);",
                    expected: Expected::Tag("CREATE TYPE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL)::ct = ROW(1, NULL)::ct, ROW(1, NULL)::ct = ROW(1, NULL::INT4), (ROW(1, NULL)::ct) IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("f")],
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
fn test_row_comparisons() {
    run_scripts(&[
        ScriptTest {
            name: "Row comparisons by ordering operators",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ROW(1, NULL) < ROW(2, 1), ROW(1, NULL) < ROW(1, 2), ROW(NULL, 1) > ROW(0, 2), ROW(1, 2, 3) <= ROW(1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), Null, Null, T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE rcx (id INT PRIMARY KEY, a INT, b INT, c INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX rcx_ab ON rcx (a, b);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rcx SELECT g, g / 100, g % 100, g % 7 FROM generate_series(1, 20000) g;",
                    expected: Expected::Tag("INSERT 0 20000"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ANALYZE rcx;",
                    expected: Expected::Tag("ANALYZE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM rcx WHERE (a, b) > (199, 97) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("19998")],
                            &[T("19999")],
                            &[T("20000")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM rcx WHERE (a, c) >= (199, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM rcx WHERE (199, 97) < (a, b);",
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
                    query: "SELECT id FROM rcx WHERE (a, b) <= (0, 3) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
