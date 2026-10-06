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
fn test_dolt_backup() {
    run_scripts(&[
        ScriptTest {
            name: "sync-url restore preserves data and commit history",
            set_up_script: &[
                "create table items (id int primary key, label text not null);",
                "insert into items values (1, 'apple'), (2, 'banana');",
                "select dolt_commit('-Am', 'first commit: add apple and banana');",
                "insert into items values (3, 'cherry');",
                "select dolt_commit('-Am', 'second commit: add cherry');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_syncurl');",
                "USE restored_syncurl",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, label from items order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("apple")],
                            &[T("2"), T("banana")],
                            &[T("3"), T("cherry")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.commits;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt.commits order by date;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("Initialize data repository")],
                            &[T("CREATE DATABASE")],
                            &[T("first commit: add apple and banana")],
                            &[T("second commit: add cherry")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select column_name, is_nullable from information_schema.columns where table_name = 'items' and table_schema = 'public' order by column_name;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", VARCHAR), Column("is_nullable", VARCHAR)],
                        rows: &[
                            &[T("id"), T("NO")],
                            &[T("label"), T("NO")],
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
fn test_dolt_backup_2() {
    run_scripts(&[
        ScriptTest {
            name: "named backup sync restore preserves data and commit history",
            set_up_script: &[
                "create table orders (id int primary key, item text, qty int);",
                "insert into orders values (1, 'widget', 10);",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'initial order');",
                "insert into orders values (2, 'gadget', 5);",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'second order');",
                "select dolt_backup('add', 'named_bak', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('sync', 'named_bak');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_named');",
                "USE restored_named",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, item, qty from orders order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("item", TEXT), Column("qty", INT4)],
                        rows: &[
                            &[T("1"), T("widget"), T("10")],
                            &[T("2"), T("gadget"), T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.commits;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt.commits order by date;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("Initialize data repository")],
                            &[T("CREATE DATABASE")],
                            &[T("initial order")],
                            &[T("second order")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_3() {
    run_scripts(&[
        ScriptTest {
            name: "dolt.dolt_backups table reflects add and remove",
            set_up_script: &[
                "select dolt_backup('add', 'backup_a', 'file://{TEMPDIR}/backup_a');",
                "select dolt_backup('add', 'backup_b', 'file://{TEMPDIR}/backup_b');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt.dolt_backups order by name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("backup_a")],
                            &[T("backup_b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('remove', 'backup_a');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt.dolt_backups order by name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("backup_b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('remove', 'backup_b');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.dolt_backups;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
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
fn test_dolt_backup_4() {
    run_scripts(&[
        ScriptTest {
            name: "restore --force overwrites an existing database and data is correct",
            set_up_script: &[
                "create table t (id int primary key, v text);",
                "insert into t values (1, 'original');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'seed');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'db_to_overwrite');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'db_to_overwrite');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "database 'db_to_overwrite' already exists, use '--force' to overwrite", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('restore', '--force', 'file://{TEMPDIR}/backup', 'db_to_overwrite');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE db_to_overwrite;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, v from t order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("original")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.commits;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE postgres;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "drop database db_to_overwrite;",
                    expected: Expected::Tag("DROP DATABASE"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_5() {
    run_scripts(&[
        ScriptTest {
            name: "multi-branch backup and restore preserves all branches",
            set_up_script: &[
                "create table products (id int primary key, name text);",
                "insert into products values (1, 'alpha');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'add alpha on main');",
                "select dolt_branch('feature');",
                "select dolt_checkout('feature');",
                "insert into products values (2, 'beta');",
                "select dolt_commit('-Am', 'add beta on feature');",
                "select dolt_checkout('main');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_branches');",
                "USE restored_branches",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, name from products order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("alpha")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt.branches order by name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("feature")],
                            &[T("main")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('feature');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, name from products order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("alpha")],
                            &[T("2"), T("beta")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select message from dolt.commits order by date desc limit 1;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("add beta on feature")],
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
fn test_dolt_backup_6() {
    run_scripts(&[
        ScriptTest {
            name: "working set state is preserved in backup",
            set_up_script: &[
                "create table logs (id int primary key, msg text);",
                "insert into logs values (1, 'committed');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'first commit');",
                "insert into logs values (2, 'staged only');",
                "select dolt_add('.');",
                "insert into logs values (3, 'unstaged');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_workingset');",
                "USE restored_workingset",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, msg from logs order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("msg", TEXT)],
                        rows: &[
                            &[T("1"), T("committed")],
                            &[T("2"), T("staged only")],
                            &[T("3"), T("unstaged")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.commits;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("3")],
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
fn test_dolt_backup_7() {
    run_scripts(&[
        ScriptTest {
            name: "incremental sync captures new commits added after the first sync",
            set_up_script: &[
                "select dolt_backup('add', 'incr_bak', 'file://{TEMPDIR}/backup');",
                "create table events (id int primary key, name text);",
                "insert into events values (1, 'first');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'first event');",
                "select dolt_backup('sync', 'incr_bak');",
                "insert into events values (2, 'second');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'second event');",
                "select dolt_backup('sync', 'incr_bak');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_incr');",
                "select dolt_backup('remove', 'incr_bak');",
                "USE restored_incr",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, name from events order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("first")],
                            &[T("2"), T("second")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.commits;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("4")],
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
fn test_dolt_backup_8() {
    run_scripts(&[
        ScriptTest {
            name: "sync nonexistent named backup returns error",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('sync', 'does_not_exist');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "backup 'does_not_exist' not found", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_9() {
    run_scripts(&[
        ScriptTest {
            name: "remove nonexistent backup returns error",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('remove', 'no_such_backup');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "backup 'no_such_backup' not found", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_10() {
    run_scripts(&[
        ScriptTest {
            name: "restore from nonexistent URL returns error",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('restore', 'file:///nonexistent/doltgres/backup/path', 'new_db');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "failed to create directory '/nonexistent/doltgres/backup/path': mkdir /nonexistent: read-only file system", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_11() {
    run_scripts(&[
        ScriptTest {
            name: "adding a backup with a duplicate name returns error",
            set_up_script: &[
                "select dolt_backup('add', 'mybackup', 'file://{TEMPDIR}/backup_a');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('add', 'mybackup', 'file://{TEMPDIR}/backup_b');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "backup 'mybackup' already exists", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('remove', 'mybackup');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
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
fn test_dolt_backup_12() {
    run_scripts(&[
        ScriptTest {
            name: "unknown dolt_backup operation returns error",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('typo', 'name', 'url');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "unrecognized dolt_backup parameter 'typo'", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_13() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_backup with no args returns usage error",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "use 'dolt_backups' table to list backups", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_14() {
    run_scripts(&[
        ScriptTest {
            name: "empty database backup and restore",
            set_up_script: &[
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_empty');",
                "USE restored_empty",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from information_schema.tables where table_schema = 'public';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) >= 2 from dolt.commits;",
                    expected: Expected::Rows {
                        columns: &[Column("count >= 2", BOOL)],
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
    ]);
}

#[test]
fn test_dolt_backup_15() {
    run_scripts(&[
        ScriptTest {
            name: "standalone sequence state is preserved across backup and restore",
            set_up_script: &[
                "create sequence counter start 1 increment 5;",
                "select nextval('counter');",
                "select nextval('counter');",
                "select nextval('counter');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_seq');",
                "USE restored_seq",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select sequence_name from information_schema.sequences where sequence_schema = 'public';",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_name", VARCHAR)],
                        rows: &[
                            &[T("counter")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select nextval('counter');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("16")],
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
fn test_dolt_backup_16() {
    run_scripts(&[
        ScriptTest {
            name: "serial column sequence state is preserved across backup and restore",
            set_up_script: &[
                "create table widgets (id serial primary key, name text);",
                "insert into widgets (name) values ('a'), ('b'), ('c');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'seed widgets');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_serial');",
                "USE restored_serial",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, name from widgets order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                            &[T("3"), T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into widgets (name) values ('d') returning id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_17() {
    run_scripts(&[
        ScriptTest {
            name: "views are preserved across backup and restore",
            set_up_script: &[
                "create table employees (id int primary key, dept text, salary int);",
                "insert into employees values (1, 'eng', 100000), (2, 'eng', 120000), (3, 'ops', 90000);",
                "create view eng_employees as select id, salary from employees where dept = 'eng';",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'add employees and view');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_views');",
                "USE restored_views",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select table_name from information_schema.views where table_schema = 'public';",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", VARCHAR)],
                        rows: &[
                            &[T("eng_employees")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, salary from eng_employees order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("salary", INT4)],
                        rows: &[
                            &[T("1"), T("100000")],
                            &[T("2"), T("120000")],
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
fn test_dolt_backup_18() {
    run_scripts(&[
        ScriptTest {
            name: "foreign key constraints are preserved and enforced after restore",
            set_up_script: &[
                "create table categories (id int primary key, name text);",
                "create table items (id int primary key, cat_id int references categories(id), label text);",
                "insert into categories values (1, 'fruit'), (2, 'veg');",
                "insert into items values (1, 1, 'apple'), (2, 2, 'carrot');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'seed fk data');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_fk');",
                "USE restored_fk",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, label from items order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("apple")],
                            &[T("2"), T("carrot")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into items values (3, 99, 'mystery');",
                    expected: Expected::Error(Diagnostic { code: "23503", message: "cannot add or update a child row - Foreign key violation on fk: `items_cat_id_fkey`, table: `items`, referenced table: `categories`, key: `[99]`", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from information_schema.referential_constraints where constraint_schema = 'public';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
fn test_dolt_backup_19() {
    run_scripts(&[
        ScriptTest {
            name: "secondary index is preserved and used after restore",
            set_up_script: &[
                "create table products (id int primary key, sku text not null, price int);",
                "create index idx_sku on products (sku);",
                "insert into products values (1, 'AAA', 10), (2, 'BBB', 20), (3, 'CCC', 30);",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'seed products with index');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_idx');",
                "USE restored_idx",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select indexname from pg_indexes where tablename = 'products' and indexname = 'idx_sku';",
                    expected: Expected::Rows {
                        columns: &[Column("indexname", NAME)],
                        rows: &[
                            &[T("idx_sku")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, sku, price from products order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("sku", TEXT), Column("price", INT4)],
                        rows: &[
                            &[T("1"), T("AAA"), T("10")],
                            &[T("2"), T("BBB"), T("20")],
                            &[T("3"), T("CCC"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, price from products where sku = 'BBB';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("price", INT4)],
                        rows: &[
                            &[T("2"), T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into products values (4, 'AAA', 99);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id from products where sku = 'AAA' order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("4")],
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
fn test_dolt_backup_20() {
    run_scripts(&[
        ScriptTest {
            name: "custom enum type is preserved across backup and restore",
            set_up_script: &[
                "create type my_mood as enum ('happy', 'ok', 'sad');",
                "create table user_states (id int primary key, name text, mood my_mood);",
                "insert into user_states values (1, 'alice', 'happy'), (2, 'bob', 'sad');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_types');",
                "USE restored_types",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select typname from pg_type where typname = 'my_mood';",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME)],
                        rows: &[
                            &[T("my_mood")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select 'ok'::my_mood;",
                    expected: Expected::Rows {
                        columns: &[Column("my_mood", USER_DEFINED)],
                        rows: &[
                            &[T("ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, name, mood::text from user_states order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("mood", TEXT)],
                        rows: &[
                            &[T("1"), T("alice"), T("happy")],
                            &[T("2"), T("bob"), T("sad")],
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
fn test_dolt_backup_21() {
    run_scripts(&[
        ScriptTest {
            name: "domain type is preserved across backup and restore",
            set_up_script: &[
                "create domain pos_int as integer check (value > 0);",
                "create table measurements (id int primary key, val pos_int);",
                "insert into measurements values (1, 42);",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_domain');",
                "USE restored_domain",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select typname, typtype from pg_type where typname = 'pos_int';",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME), Column("typtype", CHAR)],
                        rows: &[
                            &[T("pos_int"), T("d")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select val from measurements where id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("val", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into measurements values (2, -1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"Check constraint "pos_int_check" violated"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_backup_22() {
    run_scripts(&[
        ScriptTest {
            name: "user-defined composite type is preserved across backup and restore",
            set_up_script: &[
                "create type point_t as (x float8, y float8);",
                r#"create function distance(p point_t) returns float8 as $$
			 begin return sqrt((p).x * (p).x + (p).y * (p).y); end;
			 $$ language plpgsql;"#,
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_composite');",
                "USE restored_composite",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select typname, typtype from pg_type where typname = 'point_t';",
                    expected: Expected::Rows {
                        columns: &[Column("typname", NAME), Column("typtype", CHAR)],
                        rows: &[
                            &[T("point_t"), T("c")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select (row(3.0, 4.0)::point_t).x;",
                    expected: Expected::Rows {
                        columns: &[Column("((3.0, 4.0)::point_t).x", FLOAT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select distance(row(3.0, 4.0)::point_t);",
                    expected: Expected::Rows {
                        columns: &[Column("distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
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
fn test_dolt_backup_23() {
    run_scripts(&[
        ScriptTest {
            name: "user-defined function is preserved across backup and restore",
            set_up_script: &[
                r#"create function double_it(n int) returns int as $$
			 begin return n * 2; end;
			 $$ language plpgsql;"#,
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_funcs');",
                "USE restored_funcs",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select double_it(21);",
                    expected: Expected::Rows {
                        columns: &[Column("double_it", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select double_it(0);",
                    expected: Expected::Rows {
                        columns: &[Column("double_it", INT4)],
                        rows: &[
                            &[T("0")],
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
fn test_dolt_backup_24() {
    run_scripts(&[
        ScriptTest {
            name: "stored procedure is preserved across backup and restore",
            set_up_script: &[
                "create table job_log (id int primary key, status text);",
                "insert into job_log values (1, 'pending');",
                r#"create procedure mark_done(job_id int) as $$
			 begin update job_log set status = 'done' where id = job_id; end;
			 $$ language plpgsql;"#,
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_procs');",
                "USE restored_procs",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call mark_done(1);",
                    expected: Expected::Tag("CALL"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select status from job_log where id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("status", TEXT)],
                        rows: &[
                            &[T("done")],
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
fn test_dolt_backup_25() {
    run_scripts(&[
        ScriptTest {
            name: "trigger is preserved and fires after restore",
            set_up_script: &[
                "create table readings (id int primary key, val int);",
                r#"create function clamp_val() returns trigger as $$
			 begin
			   if NEW.val > 100 then NEW.val := 100; end if;
			   return NEW;
			 end;
			 $$ language plpgsql;"#,
                "create trigger clamp_trigger before insert on readings for each row execute function clamp_val();",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_triggers');",
                "USE restored_triggers",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into readings values (1, 200);",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select val from readings where id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("val", INT4)],
                        rows: &[
                            &[T("100")],
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
fn test_dolt_backup_26() {
    run_scripts(&[
        ScriptTest {
            name: "extension is preserved across backup and restore",
            set_up_script: &[
                r#"create extension "uuid-ossp";"#,
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_ext');",
                "USE restored_ext",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select extname, extversion from pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME), Column("extversion", TEXT)],
                        rows: &[
                            &[T("uuid-ossp"), T("1.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(uuid_generate_v4()::text) = 36;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 36", BOOL)],
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
    ]);
}

#[test]
fn test_dolt_backup_27() {
    run_scripts(&[
        ScriptTest {
            name: "named schema contents are preserved across backup and restore",
            set_up_script: &[
                "create schema inventory;",
                "create table inventory.products (id int primary key, name text);",
                "insert into inventory.products values (1, 'widget'), (2, 'gadget');",
                "select dolt_commit('-Am', 'add inventory schema');",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_schema');",
                "USE restored_schema",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select schema_name from information_schema.schemata where schema_name = 'inventory';",
                    expected: Expected::Rows {
                        columns: &[Column("schema_name", VARCHAR)],
                        rows: &[
                            &[T("inventory")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, name from inventory.products order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("widget")],
                            &[T("2"), T("gadget")],
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
fn test_dolt_backup_28() {
    run_scripts(&[
        ScriptTest {
            name: "custom cast is preserved across backup and restore",
            set_up_script: &[
                "CREATE TABLE cast_src (v text);",
                "CREATE TABLE cast_dst (v text, tag text);",
                r#"CREATE FUNCTION cast_src_to_dst(src cast_src) RETURNS cast_dst AS $$
			 SELECT ROW((src).v, 'casted')::cast_dst
			 $$ LANGUAGE SQL;"#,
                r#"CREATE FUNCTION cast_verify(dst cast_dst) RETURNS text AS $$
			 SELECT (dst).v || ':' || (dst).tag
			 $$ LANGUAGE SQL;"#,
                "CREATE CAST (cast_src AS cast_dst) WITH FUNCTION cast_src_to_dst(cast_src);",
                "select dolt_backup('sync-url', 'file://{TEMPDIR}/backup');",
                "select dolt_backup('restore', 'file://{TEMPDIR}/backup', 'restored_casts');",
                "USE restored_casts",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT c.castcontext::text, c.castmethod::text
					FROM pg_cast c
					JOIN pg_type src ON src.oid = c.castsource
					JOIN pg_type dst ON dst.oid = c.casttarget
					WHERE src.typname = 'cast_src' AND dst.typname = 'cast_dst';"#,
                    expected: Expected::Rows {
                        columns: &[Column("c.castcontext", TEXT), Column("c.castmethod", TEXT)],
                        rows: &[
                            &[T("e"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT cast_verify((ROW('hello')::cast_src)::cast_dst);",
                    expected: Expected::Rows {
                        columns: &[Column("cast_verify", TEXT)],
                        rows: &[
                            &[T("hello:casted")],
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
