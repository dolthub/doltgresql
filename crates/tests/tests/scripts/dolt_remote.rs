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
fn test_dolt_remote() {
    run_scripts(&[
        ScriptTest {
            name: "push and clone preserve table data and commit history",
            set_up_script: &[
                "create table items (id int primary key, label text not null);",
                "insert into items values (1, 'apple'), (2, 'banana');",
                "select dolt_commit('-Am', 'first commit: add apple and banana');",
                "insert into items values (3, 'cherry');",
                "select dolt_commit('-Am', 'second commit: add cherry');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_items');",
                "USE cloned_items",
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
                    query: "select name from dolt.remotes;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
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
fn test_dolt_remote_2() {
    run_scripts(&[
        ScriptTest {
            name: "standalone sequence state is preserved across push and clone",
            set_up_script: &[
                "create sequence counter start 1 increment 5;",
                "select nextval('counter');",
                "select nextval('counter');",
                "select nextval('counter');",
                "select dolt_commit('-Am', 'seed counter');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_seq');",
                "drop sequence counter",
                "USE cloned_seq",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' output for the query.
                ScriptTestAssertion {
                    query: "select sequence_name, sequence_catalog from information_schema.sequences where sequence_schema = 'public';",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_name", NAME), Column("sequence_catalog", NAME)],
                        rows: &[
                            &[T("counter"), T("cloned_seq")],
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
fn test_dolt_remote_3() {
    run_scripts(&[
        ScriptTest {
            name: "serial column's owned sequence state is preserved across push and clone",
            set_up_script: &[
                "create table widgets (id serial primary key, name text);",
                "insert into widgets (name) values ('a'), ('b'), ('c');",
                "select dolt_commit('-Am', 'seed widgets');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_widgets');",
                "drop table widgets",
                "USE cloned_widgets",
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
fn test_dolt_remote_4() {
    run_scripts(&[
        ScriptTest {
            name: "identity column's owned sequence state is preserved across push and clone",
            set_up_script: &[
                "create table gadgets (id bigint generated by default as identity primary key, name text);",
                "insert into gadgets (name) values ('a'), ('b'), ('c');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'seed gadgets');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_gadgets');",
                "drop table gadgets",
                "USE cloned_gadgets",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, name from gadgets order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("name", TEXT)],
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
                    query: "insert into gadgets (name) values ('d') returning id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8)],
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
fn test_dolt_remote_5() {
    run_scripts(&[
        ScriptTest {
            name: "sequence in a non-public schema is preserved across push and clone",
            set_up_script: &[
                "create schema myschema;",
                "create sequence myschema.seq2 start 100 increment 10;",
                "select nextval('myschema.seq2');",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'seed myschema.seq2');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_schema');",
                "USE cloned_schema",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' output for the query.
                ScriptTestAssertion {
                    query: "select sequence_schema, sequence_name from information_schema.sequences where sequence_name = 'seq2';",
                    expected: Expected::Rows {
                        columns: &[Column("sequence_schema", NAME), Column("sequence_name", NAME)],
                        rows: &[
                            &[T("myschema"), T("seq2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select nextval('myschema.seq2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("110")],
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
fn test_dolt_remote_6() {
    run_scripts(&[
        ScriptTest {
            name: "incremental push and pull keep sequence and table state in sync",
            set_up_script: &[
                "create sequence counter start 1 increment 5;",
                "create table orders (id int primary key default nextval('counter'), item text);",
                "select dolt_commit('-Am', 'initial schema');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'replica');",
                "USE postgres",
                "insert into orders (item) values ('widget');",
                "select dolt_commit('-Am', 'first order');",
                "select dolt_push('origin', 'main');",
                "USE replica",
                "select dolt_pull('origin');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, item from orders order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("item", TEXT)],
                        rows: &[
                            &[T("1"), T("widget")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT schemaname, sequencename, start_value, min_value, max_value, increment_by, cycle, cache_size, last_value FROM pg_sequences;",
                    expected: Expected::Rows {
                        columns: &[Column("schemaname", NAME), Column("sequencename", NAME), Column("start_value", INT8), Column("min_value", INT8), Column("max_value", INT8), Column("increment_by", INT8), Column("cycle", BOOL), Column("cache_size", INT8), Column("last_value", INT8)],
                        rows: &[
                            &[T("public"), T("counter"), T("1"), T("1"), T("9223372036854775807"), T("5"), T("f"), T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: the Go server returns 1 again after the pull, reusing the value the source took, while the pulled sequence's next value is 6.
                ScriptTestAssertion {
                    query: "select nextval('counter');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("6")],
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
fn test_dolt_remote_7() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_remotes reflects add and remove",
            set_up_script: &[
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote_a}');",
                "select dolt_remote('add', 'other', 'file://{NEWDIR:remote_b}');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_remotes order by name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                            &[T("other")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_remote('remove', 'other');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_remote", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_remotes;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_remote('remove', 'other');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "error: unknown remote: 'other'", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_8() {
    run_scripts(&[
        ScriptTest {
            name: "fetch updates remote-tracking branches without touching the working branch; pull merges",
            set_up_script: &[
                "create table events (id int primary key, name text);",
                "insert into events values (1, 'first');",
                "select dolt_commit('-Am', 'first event');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'subscriber');",
                "USE postgres",
                "insert into events values (2, 'second');",
                "select dolt_commit('-Am', 'second event');",
                "select dolt_push('origin', 'main');",
                "USE subscriber",
                "select dolt_fetch('origin', 'main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, name from events order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("first")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_remote_branches;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("remotes/origin/main")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_pull('origin');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_pull", RECORD)],
                        rows: &[
                            &[T(r#"(1,0,"merge successful")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
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
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_9() {
    run_scripts(&[
        ScriptTest {
            name: "pull fails when the working set has uncommitted changes",
            set_up_script: &[
                "create table logs (id int primary key, msg text);",
                "insert into logs values (1, 'a');",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'dirty_clone');",
                "USE postgres",
                "insert into logs values (2, 'b');",
                "select dolt_commit('-Am', 'second');",
                "select dolt_push('origin', 'main');",
                "USE dirty_clone",
                "insert into logs values (3, 'uncommitted');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_pull('origin');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "cannot merge with uncommitted changes", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_10() {
    run_scripts(&[
        ScriptTest {
            name: "push without a configured remote returns an error",
            set_up_script: &[
                "create table t (id int primary key);",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'seed');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_push('origin', 'main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"fatal: remote 'origin' not found.
Please make sure the remote exists."#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_11() {
    run_scripts(&[
        ScriptTest {
            name: "pull without a configured remote returns an error",
            set_up_script: &[
                "create table t (id int primary key);",
                "select dolt_add('.');",
                "select dolt_commit('-m', 'seed');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_pull('origin');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"fatal: remote 'origin' not found.
Please make sure the remote exists."#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_12() {
    run_scripts(&[
        ScriptTest {
            name: "custom enum type is preserved across push and clone",
            set_up_script: &[
                "create type mood as enum ('sad', 'ok', 'happy');",
                "create table moods (id int primary key, m mood);",
                "insert into moods values (1, 'happy'), (2, 'sad');",
                "select dolt_commit('-Am', 'seed moods');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_moods');",
                "drop table moods;",
                "drop type mood;",
                "USE cloned_moods",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, m::text from moods order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("m", TEXT)],
                        rows: &[
                            &[T("1"), T("happy")],
                            &[T("2"), T("sad")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into moods values (3, 'ok') returning m::text;",
                    expected: Expected::Rows {
                        columns: &[Column("m", TEXT)],
                        rows: &[
                            &[T("ok")],
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
fn test_dolt_remote_13() {
    run_scripts(&[
        ScriptTest {
            name: "user-defined function is preserved across push and clone",
            set_up_script: &[
                "create function double_it(x int) returns int as $$ begin return x * 2; end; $$ language plpgsql;",
                "select dolt_commit('-Am', 'seed function');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_func');",
                "drop function double_it(int);",
                "USE cloned_func",
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
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_14() {
    run_scripts(&[
        ScriptTest {
            name: "pushing and cloning a non-default branch carries its own sequence and table state",
            set_up_script: &[
                "create sequence counter start 1 increment 1;",
                "create table gadgets (id int primary key default nextval('counter'), name text);",
                "insert into gadgets (name) values ('a'), ('b'), ('c');",
                "select dolt_commit('-Am', 'seed on main');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_checkout('-b', 'feature');",
                "insert into gadgets (name) values ('feature-only');",
                "select dolt_commit('-Am', 'feature work');",
                "select dolt_push('origin', 'feature');",
                "select dolt_clone('--branch', 'feature', 'file://{NEWDIR:remote}', 'cloned_feature');",
                "drop table gadgets;",
                "drop sequence counter;",
                "USE cloned_feature",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from gadgets order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("c")],
                            &[T("feature-only")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into gadgets (name) values ('next') returning id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("5")],
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
fn test_dolt_remote_15() {
    run_scripts(&[
        ScriptTest {
            name: "non-fast-forward push is rejected, then succeeds with --force",
            set_up_script: &[
                "create table t (id int primary key, note text);",
                "insert into t values (1, 'source');",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'clone1');",
                "insert into t values (2, 'source-diverged');",
                "select dolt_commit('-Am', 'source diverges');",
                "select dolt_push('origin', 'main');",
                "USE clone1",
                "insert into t values (3, 'clone-diverged');",
                "select dolt_commit('-Am', 'clone diverges');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_push('origin', 'main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"To file://{TEMPDIR}/remote
 ! [rejected]            main -> main (non-fast-forward)
error: failed to push some refs to 'file://{TEMPDIR}/remote'
hint: Updates were rejected because the tip of your current branch is behind
hint: its remote counterpart. Integrate the remote changes (e.g.
hint: 'dolt pull ...') before pushing again.
"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_push('-f', 'origin', 'main');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_16() {
    run_scripts(&[
        ScriptTest {
            name: "pushing a branch delete removes it from the remote",
            set_up_script: &[
                "create table t (id int primary key);",
                "insert into t values (1);",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_checkout('-b', 'doomed');",
                "insert into t values (2);",
                "select dolt_commit('-Am', 'doomed work');",
                "select dolt_push('origin', 'doomed');",
                "select dolt_clone('file://{NEWDIR:remote}', 'observer');",
                "USE observer",
                "select dolt_fetch('origin', 'doomed');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_remote_branches order by name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("remotes/origin/doomed")],
                            &[T("remotes/origin/main")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE postgres",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_push('origin', ':doomed');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "USE observer",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_fetch('--prune', 'origin');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_remote_branches order by name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("remotes/origin/main")],
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
fn test_dolt_remote_17() {
    run_scripts(&[
        ScriptTest {
            name: "tags are pushed and fetched along with commits",
            set_up_script: &[
                "create table t (id int primary key);",
                "insert into t values (1);",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_tag('v1.0', '-m', 'first release');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_push('origin', 'v1.0');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_tags');",
                "USE cloned_tags",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select tag_name from dolt_tags;",
                    expected: Expected::Rows {
                        columns: &[Column("tag_name", TEXT)],
                        rows: &[
                            &[T("v1.0")],
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
fn test_dolt_remote_18() {
    run_scripts(&[
        ScriptTest {
            name: "fetch supports an explicit refspec into a custom tracking ref",
            set_up_script: &[
                "create table t (id int primary key);",
                "insert into t values (1);",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'observer');",
                "USE observer",
                "select dolt_fetch('origin', 'refs/heads/main:refs/remotes/custom/main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_remote_branches where name = 'remotes/custom/main';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("remotes/custom/main")],
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
fn test_dolt_remote_19() {
    run_scripts(&[
        ScriptTest {
            name: "push --all pushes every local branch",
            set_up_script: &[
                "create table t (id int primary key);",
                "insert into t values (1);",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_checkout('-b', 'branch_a');",
                "select dolt_checkout('main');",
                "select dolt_checkout('-b', 'branch_b');",
                "select dolt_checkout('main');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('--all', 'origin');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_all');",
                "USE cloned_all",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select name from dolt_remote_branches order by name;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("remotes/origin/branch_a")],
                            &[T("remotes/origin/branch_b")],
                            &[T("remotes/origin/main")],
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

#[test]
fn test_dolt_remote_20() {
    run_scripts(&[
        ScriptTest {
            name: "adding a remote with a name that already exists returns an error",
            set_up_script: &[
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote_a}');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote_b}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "remote already exists", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_21() {
    run_scripts(&[
        ScriptTest {
            name: "adding a remote with an invalid name returns an error",
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_remote('add', 'bad name', 'file://{NEWDIR:remote}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "remote name invalid", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_22() {
    run_scripts(&[
        ScriptTest {
            name: "fetching an invalid ref spec returns an error",
            set_up_script: &[
                "create table t (id int primary key);",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_fetch('origin', 'garbage');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "fetch failed: invalid ref spec: 'garbage'", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_23() {
    run_scripts(&[
        ScriptTest {
            name: "pull auto-merges non-conflicting divergent history",
            set_up_script: &[
                "create table t (id int primary key, v text);",
                "insert into t values (1, 'a'), (2, 'b');",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'clone1');",
                "USE postgres",
                "insert into t values (3, 'source-row');",
                "select dolt_commit('-Am', 'source adds row 3');",
                "select dolt_push('origin', 'main');",
                "USE clone1",
                "insert into t values (4, 'clone-row');",
                "select dolt_commit('-Am', 'clone adds row 4');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_pull('origin');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_pull", RECORD)],
                        rows: &[
                            &[T(r#"(0,0,"merge successful")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select id, v from t order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                            &[T("3"), T("source-row")],
                            &[T("4"), T("clone-row")],
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
fn test_dolt_remote_24() {
    run_scripts(&[
        ScriptTest {
            name: "pull surfaces real merge conflicts, resolvable via dolt_conflicts_resolve",
            set_up_script: &[
                "create table t (id int primary key, v text);",
                "insert into t values (1, 'a'), (2, 'b');",
                "select dolt_commit('-Am', 'seed');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'clone1');",
                "USE postgres",
                "update t set v = 'source-edit' where id = 1;",
                "select dolt_commit('-Am', 'source edits row 1');",
                "select dolt_push('origin', 'main');",
                "USE clone1",
                "update t set v = 'clone-edit' where id = 1;",
                "select dolt_commit('-Am', 'clone edits row 1');",
                "set dolt_allow_commit_conflicts to 1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_pull('origin');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_pull", RECORD)],
                        rows: &[
                            &[T(r#"(0,1,"merge has unresolved conflicts or constraint violations")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select base_v, our_v, their_v from dolt_conflicts_t;",
                    expected: Expected::Rows {
                        columns: &[Column("base_v", TEXT), Column("our_v", TEXT), Column("their_v", TEXT)],
                        rows: &[
                            &[T("a"), T("clone-edit"), T("source-edit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_conflicts_resolve('--ours', 't');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt_conflicts;",
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
                    query: "select id, v from t order by id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("clone-edit")],
                            &[T("2"), T("b")],
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
fn test_dolt_remote_25() {
    run_scripts(&[
        ScriptTest {
            name: "domain type and its check constraint are preserved across push and clone",
            set_up_script: &[
                "create domain pos_int as integer check (value > 0);",
                "create table measurements (id int primary key, val pos_int);",
                "insert into measurements values (1, 42);",
                "select dolt_commit('-Am', 'seed measurements');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_domain');",
                "drop table measurements;",
                "drop domain pos_int;",
                "USE cloned_domain",
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
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' output for the query.
                ScriptTestAssertion {
                    query: "insert into measurements values (2, -1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"value for domain pos_int violates check constraint "pos_int_check""#, schema: "public", data_type: "pos_int", constraint: "pos_int_check", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_remote_26() {
    run_scripts(&[
        ScriptTest {
            name: "user-defined composite type is preserved across push and clone",
            set_up_script: &[
                "create type point_t as (x float8, y float8);",
                "create function distance(p point_t) returns float8 as $$ begin return sqrt((p).x * (p).x + (p).y * (p).y); end; $$ language plpgsql;",
                "select dolt_commit('-Am', 'seed composite type');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_composite');",
                "drop function distance(point_t);",
                "drop type point_t;",
                "USE cloned_composite",
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
fn test_dolt_remote_27() {
    run_scripts(&[
        ScriptTest {
            name: "stored procedure is preserved across push and clone",
            set_up_script: &[
                "create table job_log (id int primary key, status text);",
                "insert into job_log values (1, 'pending');",
                "create procedure mark_done(job_id int) as $$ begin update job_log set status = 'done' where id = job_id; end; $$ language plpgsql;",
                "select dolt_commit('-Am', 'seed procedure');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_proc');",
                "drop procedure mark_done(int);",
                "drop table job_log;",
                "USE cloned_proc",
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
fn test_dolt_remote_28() {
    run_scripts(&[
        ScriptTest {
            name: "trigger is preserved and fires after clone",
            set_up_script: &[
                "create table readings (id int primary key, val int);",
                "create function clamp_val() returns trigger as $$ begin if NEW.val > 100 then NEW.val := 100; end if; return NEW; end; $$ language plpgsql;",
                "create trigger clamp_trigger before insert on readings for each row execute function clamp_val();",
                "select dolt_commit('-Am', 'seed trigger');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_trigger');",
                "drop trigger clamp_trigger on readings;",
                "drop function clamp_val();",
                "drop table readings;",
                "USE cloned_trigger",
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
fn test_dolt_remote_29() {
    run_scripts(&[
        ScriptTest {
            name: "extension is preserved across push and clone",
            set_up_script: &[
                r#"create extension "uuid-ossp";"#,
                "select dolt_commit('-Am', 'seed extension');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_ext');",
                "USE cloned_ext",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' output for the query.
                ScriptTestAssertion {
                    query: "select extname, extversion from pg_catalog.pg_extension;",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME), Column("extversion", TEXT)],
                        rows: &[
                            &[T("plpgsql"), T("1.0")],
                            &[T("uuid-ossp"), T("1.1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' output for the query.
                ScriptTestAssertion {
                    query: "select length(uuid_generate_v4()::text) = 36;",
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
    ]);
}

#[test]
fn test_dolt_remote_30() {
    run_scripts(&[
        ScriptTest {
            name: "named schema contents are preserved across push and clone",
            set_up_script: &[
                "create schema inventory;",
                "create table inventory.products (id int primary key, name text);",
                "insert into inventory.products values (1, 'widget'), (2, 'gadget');",
                "select dolt_commit('-Am', 'add inventory schema');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_schema_contents');",
                "drop table inventory.products;",
                "drop schema inventory;",
                "USE cloned_schema_contents",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' output for the query.
                ScriptTestAssertion {
                    query: "select schema_name from information_schema.schemata where schema_name = 'inventory';",
                    expected: Expected::Rows {
                        columns: &[Column("schema_name", NAME)],
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
fn test_dolt_remote_31() {
    run_scripts(&[
        ScriptTest {
            name: "custom cast is preserved across push and clone",
            set_up_script: &[
                "CREATE TABLE cast_src (v text);",
                "CREATE TABLE cast_dst (v text, tag text);",
                "CREATE FUNCTION cast_src_to_dst(src cast_src) RETURNS cast_dst AS $$ SELECT ROW((src).v, 'casted')::cast_dst $$ LANGUAGE SQL;",
                "CREATE FUNCTION cast_verify(dst cast_dst) RETURNS text AS $$ SELECT (dst).v || ':' || (dst).tag $$ LANGUAGE SQL;",
                "CREATE CAST (cast_src AS cast_dst) WITH FUNCTION cast_src_to_dst(cast_src);",
                "select dolt_commit('-Am', 'seed cast');",
                "select dolt_remote('add', 'origin', 'file://{NEWDIR:remote}');",
                "select dolt_push('origin', 'main');",
                "select dolt_clone('file://{NEWDIR:remote}', 'cloned_cast');",
                "drop cast (cast_src as cast_dst);",
                "drop function cast_src_to_dst(cast_src);",
                "drop function cast_verify(cast_dst);",
                "drop table cast_src;",
                "drop table cast_dst;",
                "USE cloned_cast",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: r#"SELECT c.castcontext::text, c.castmethod::text
					FROM pg_cast c
					JOIN pg_type src ON src.oid = c.castsource
					JOIN pg_type dst ON dst.oid = c.casttarget
					WHERE src.typname = 'cast_src' AND dst.typname = 'cast_dst';"#,
                    expected: Expected::Rows {
                        columns: &[Column("castcontext", TEXT), Column("castmethod", TEXT)],
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
