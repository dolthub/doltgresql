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
fn test_auth_dolt_procedures() {
    run_scripts(&[
        ScriptTest {
            name: "SUPERUSER authorization for CALL executing Dolt stored procedures",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "create table test_table (v int);",
                "insert into test_table values (1);",
                "select dolt_add('test_table');",
                "select dolt_commit('-m', 'add test table');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_backup('sync-url', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_backup('add', 'bak1', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('-b', 'test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_branch('new_branch');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into test_table values (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_add('.');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-m', 'amend test table');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_cherry_pick('test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_clean('--dry-run');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_clone('file://{TEMPDIR}/bak1', 'cloned_bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set authtest.hash = ''",
                    expected: Expected::Tag("SET"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit_hash_out('authtest.hash', '-am', 'add val 3 to test table')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('-b', 'conflict');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update test_table set v = -1 where v = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-am', 'amend 1 to -1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update test_table set v = -2 where v = 1;",
                    expected: Expected::Tag("UPDATE 0"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-am', 'amend 2 to -2');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set dolt_allow_commit_conflicts to 1;",
                    expected: Expected::Tag("SET"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_merge('conflict');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_conflicts_resolve('--theirs', 'test_table');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_count_commits('--from=main', '--to=test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_backup('remove', 'bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_remote('add', 'origin', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_fetch('origin', 'main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_undrop('cloned_bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-am', 'resolve conflicts');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_update_column_tag('test_table', 'v', '123');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_purge_dropped_databases();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_rebase('-i', 'main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_rebase('--abort');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "create table to_rm (v int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_add('to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-m', 'clean state to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_rm('to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_gc('--shallow');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_thread_dump();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-m', 'rm to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_push('origin', 'test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_pull('origin', 'test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_reset('--soft', 'HEAD~1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stash('push', 'to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_tag('-m', 'dolt_rm procedure', 'to_rm', 'HEAD');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_verify_constraints('--all');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_info('--short');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_wait();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_flush();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_gc();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_purge();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_restart();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_once();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Basic user authentication for CALL executing Dolt stored procedures",
            set_up_script: &[
                "create user if not exists 'auth_test_basic' with password 'auth_test_bpass'",
                "alter user auth_test_basic createdb;",
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "create table test_table (v int);",
                "insert into test_table values (1);",
                "select dolt_add('test_table');",
                "select dolt_commit('-m', 'add test table');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_backup('sync-url', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_backup('add', 'bak1', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT all ON schema public TO auth_test_basic",
                    expected: Expected::Tag("GRANT"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT select,insert,delete,update ON test_table TO auth_test_basic",
                    expected: Expected::Tag("GRANT"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('-b', 'test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_branch('new_branch');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into test_table values (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_add('.');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-m', 'amend test table');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_cherry_pick('test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_clean('--dry-run');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_clone('file://{TEMPDIR}/bak1', 'cloned_bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "create database cloned_bak1;",
                    expected: Expected::Tag("CREATE DATABASE"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set authtest.hash = '';",
                    expected: Expected::Tag("SET"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit_hash_out('authtest.hash', '-am', 'add val 3 to test table');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('-b', 'conflict');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update test_table set v = -1 where v = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-am', 'amend 1 to -1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update test_table set v = -2 where v = 1;",
                    expected: Expected::Tag("UPDATE 0"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-am', 'amend 2 to -2');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set dolt_allow_commit_conflicts to 1;",
                    expected: Expected::Tag("SET"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_merge('conflict');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_conflicts_resolve('--theirs', 'test_table');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_count_commits('--from=main', '--to=test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_backup('remove', 'bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_remote('add', 'origin', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_fetch('origin', 'main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_undrop('cloned_bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-am', 'resolve conflicts');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_update_column_tag('test_table', 'v', '123');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "drop database cloned_bak1;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "must be owner of database cloned_bak1", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "drop database cloned_bak1;",
                    expected: Expected::Tag("DROP DATABASE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_purge_dropped_databases();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_checkout('test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_rebase('-i', 'main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_rebase('--abort');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "create table to_rm (v int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_add('to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-m', 'clean state to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_rm('to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_gc('--shallow');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_thread_dump();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_commit('-m', 'rm to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_push('origin', 'test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_pull('origin', 'test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_reset('--soft', 'HEAD~1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stash('push', 'to_rm');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_tag('-m', 'dolt_rm procedure', 'to_rm', 'HEAD');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_verify_constraints('--all');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_info('--short');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_wait();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_flush();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_gc();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_purge();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_restart();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "call dolt_stats_once();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Dolt stored procedure may only be invoked using SELECT", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SUPERUSER authorization for SELECT executing Dolt stored procedures",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "create table test_table (v int);",
                "insert into test_table values (1);",
                "select dolt_add('test_table');",
                "select dolt_commit('-m', 'add test table');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('sync-url', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('add', 'bak1', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('-b', 'test');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'test'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_branch('new_branch');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into test_table values (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-m', 'amend test table')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
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
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_cherry_pick('test')::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_clean('--dry-run');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_clone('file://{TEMPDIR}/bak1', 'cloned_bak1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clone", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set authtest.hash = '';",
                    expected: Expected::Tag("SET"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                ScriptTestAssertion {
                    query: "select dolt_commit_hash_out('authtest.hash', '-am', 'add val 3 to test table');",
                    flow: Flow::Query,
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    skip: Some("the Go server panics, with a stack trace that differs between runs"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('-b', 'conflict');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'conflict'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update test_table set v = -1 where v = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-am', 'amend 1 to -1')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
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
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update test_table set v = -2 where v = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-am', 'amend 2 to -2')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set dolt_allow_commit_conflicts to 1;",
                    expected: Expected::Tag("SET"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_merge('conflict');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_merge", RECORD)],
                        rows: &[
                            &[T(r#"("",0,1,"conflicts found")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_conflicts_resolve('--theirs', 'test_table');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_conflicts_resolve", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_count_commits('--from=main', '--to=test');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_count_commits", RECORD)],
                        rows: &[
                            &[T("(2,1)")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('remove', 'bak1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('add', 'bak2', 'file://{TEMPDIR}/bak2');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('sync', 'bak2');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('restore', 'file://{TEMPDIR}/bak2', 'restored_db');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "drop database restored_db;",
                    expected: Expected::Tag("DROP DATABASE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('remove', 'bak2');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_backup", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_remote('add', 'origin', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_remote", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_fetch('origin', 'main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_fetch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "drop database cloned_bak1",
                    expected: Expected::Tag("DROP DATABASE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_undrop('cloned_bak1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_undrop", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-am', 'resolve conflicts')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_update_column_tag('test_table', 'v', '123');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "table test_table does not exist", ..E }),
                    flow: Flow::Query,
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "drop database cloned_bak1",
                    expected: Expected::Tag("DROP DATABASE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_purge_dropped_databases();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_purge_dropped_databases", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('test');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'test'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rebase('-i', 'main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rebase", RECORD)],
                        rows: &[
                            &[T(r#"(0,"interactive rebase started on branch dolt_rebase_test; adjust the rebase plan in the dolt_rebase table, then continue rebasing by calling dolt_rebase('--continue')")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rebase('--abort');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rebase", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Interactive rebase aborted")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "create table to_rm (v int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('to_rm');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-m', 'clean state to_rm')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rm('to_rm');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rm", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_gc('--shallow');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_gc", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select instr(dolt_thread_dump()::text, 'goroutine') > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("instr > 0", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-m', 'rm to_rm')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_push('origin', 'test');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_push", RECORD)],
                        rows: &[
                            &[T(r#"(0,"To file://{TEMPDIR}/bak1
 * [new branch]          test -> test")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_pull('origin', 'test');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_pull", RECORD)],
                        rows: &[
                            &[T(r#"(0,0,"Everything up-to-date")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_reset('--soft', 'HEAD~1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stash('push', 'to_rm');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stash", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_tag('-m', 'dolt_rm procedure', 'to_rm', 'HEAD');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_tag", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_verify_constraints('--all');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_info('--short');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_info", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_wait();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_wait", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_flush();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_flush", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_gc();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_gc", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_purge();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_purge", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_restart();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_restart", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_once();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_once", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Basic user authorization for SELECT executing Dolt stored procedures",
            set_up_script: &[
                "create user if not exists 'auth_test_basic' with password 'auth_test_bpass'",
                "alter user auth_test_basic createdb;",
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "create table test_table (v int);",
                "insert into test_table values (1);",
                "select dolt_add('test_table');",
                "select dolt_commit('-m', 'add test table');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('sync-url', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('add', 'bak1', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT all ON schema public TO auth_test_basic",
                    expected: Expected::Tag("GRANT"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT select,insert,delete,update ON test_table TO auth_test_basic",
                    expected: Expected::Tag("GRANT"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('-b', 'test');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'test'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_branch('new_branch');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into test_table values (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-m', 'amend test table')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
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
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_cherry_pick('test')::text);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_clean('--dry-run');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_clone('file://{TEMPDIR}/bak1', 'cloned_bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "create database cloned_bak1;",
                    expected: Expected::Tag("CREATE DATABASE"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set authtest.hash = '';",
                    expected: Expected::Tag("SET"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                ScriptTestAssertion {
                    query: "select dolt_commit_hash_out('authtest.hash', '-am', 'add val 3 to test table');",
                    flow: Flow::Query,
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    skip: Some("the Go server panics, with a stack trace that differs between runs"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('-b', 'conflict');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'conflict'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update test_table set v = -1 where v = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-am', 'amend 1 to -1')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
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
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "update test_table set v = -2 where v = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-am', 'amend 2 to -2')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set dolt_allow_commit_conflicts to 1;",
                    expected: Expected::Tag("SET"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_merge('conflict');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_merge", RECORD)],
                        rows: &[
                            &[T(r#"("",0,1,"conflicts found")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_conflicts_resolve('--theirs', 'test_table');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_conflicts_resolve", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_count_commits('--from=main', '--to=test');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_count_commits", RECORD)],
                        rows: &[
                            &[T("(2,1)")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('remove', 'bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('sync', 'bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_backup('restore', 'file://{TEMPDIR}/bak1', 'restored_db');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_remote('add', 'origin', 'file://{TEMPDIR}/bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_fetch('origin', 'main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "drop database cloned_bak1",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "must be owner of database cloned_bak1", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "drop database cloned_bak1",
                    expected: Expected::Tag("DROP DATABASE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_undrop('cloned_bak1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-am', 'resolve conflicts')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_update_column_tag('test_table', 'v', '123');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    flow: Flow::Query,
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_purge_dropped_databases();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_checkout('test');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'test'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rebase('-i', 'main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rebase", RECORD)],
                        rows: &[
                            &[T(r#"(0,"interactive rebase started on branch dolt_rebase_test; adjust the rebase plan in the dolt_rebase table, then continue rebasing by calling dolt_rebase('--continue')")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rebase('--abort');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rebase", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Interactive rebase aborted")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "create table to_rm (v int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('to_rm');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-m', 'clean state to_rm')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_rm('to_rm');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rm", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_gc('--shallow');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_thread_dump();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select length(dolt_commit('-m', 'rm to_rm')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_push('origin', 'test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_pull('origin', 'test');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "permission denied for Dolt procedure", ..E }),
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_reset('--soft', 'HEAD~1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stash('push', 'to_rm');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stash", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_tag('-m', 'dolt_rm procedure', 'to_rm', 'HEAD');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_tag", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_verify_constraints('--all');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_info('--short');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_info", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_wait();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_wait", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_flush();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_flush", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_gc();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_gc", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_purge();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_purge", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_restart();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_restart", TEXT)],
                        rows: &[
                            &[T("Ok")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_stats_once();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stats_once", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_basic",
                    password: "auth_test_bpass",
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_auth_tests() {
    run_scripts(&[
        ScriptTest {
            name: "Simple CREATE USER and DROP USER",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    password: "hello",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER user1 PASSWORD 'hello';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "hello",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER user1;",
                    expected: Expected::Tag("DROP ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    password: "hello",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER PASSWORD",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER user1 PASSWORD 'something';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "something",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER USER user1 PASSWORD 'another_thing';",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    password: "something",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "another_thing",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER USER user1 WITH PASSWORD NULL;",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    password: "something",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 5;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    password: "another_thing",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER USER user1 PASSWORD 'different484';",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 7;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "different484",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER LOGIN",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE ROLE user1 PASSWORD 'pass1';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER user2 PASSWORD 'pass2';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE user3 PASSWORD 'pass3' LOGIN;",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER user4 PASSWORD 'pass4' NOLOGIN;",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28000", message: r#"role "user1" is not permitted to log in"#, ..E }),
                    username: "user1",
                    password: "pass1",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user2",
                    password: "pass2",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user3",
                    password: "pass3",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28000", message: r#"role "user4" is not permitted to log in"#, ..E }),
                    username: "user4",
                    password: "pass4",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER USER user1 WITH LOGIN;",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER USER user2 WITH NOLOGIN;",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "pass1",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28000", message: r#"role "user2" is not permitted to log in"#, ..E }),
                    username: "user2",
                    password: "pass2",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE USER IF NOT EXISTS",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 1;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    password: "hello",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER user1 PASSWORD 'hello1';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER user1 PASSWORD 'hello2';",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"role "user1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER IF NOT EXISTS user1 PASSWORD 'hello3';",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "NOT""#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "hello1",
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE IF NOT EXISTS user2 PASSWORD 'hi1' LOGIN;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "NOT""#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE user2 PASSWORD 'hi2' LOGIN;",
                    expected: Expected::Tag("CREATE ROLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE IF NOT EXISTS user2 PASSWORD 'hi3' LOGIN;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "NOT""#, position: 16, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user2""#, ..E }),
                    flow: Flow::Query,
                    username: "user2",
                    password: "hi1",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP USER IF EXISTS",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER user1 PASSWORD 'hello1';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER user2 PASSWORD 'hello2';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "hello1",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user2",
                    password: "hello2",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER user1;",
                    expected: Expected::Tag("DROP ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER user1;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "user1" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER IF EXISTS user1;",
                    expected: Expected::Tag("DROP ROLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"role "user1" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    password: "hello1",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE IF EXISTS user2;",
                    expected: Expected::Tag("DROP ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE user2;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "user2" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE IF EXISTS user2;",
                    expected: Expected::Tag("DROP ROLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"role "user2" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user2""#, ..E }),
                    username: "user2",
                    password: "hello2",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP USER with multiple users",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE USER user1 PASSWORD 'hello1';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER user2 PASSWORD 'hello2';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE USER user3 PASSWORD 'hello3';",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "hello1",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user2",
                    password: "hello2",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user3",
                    password: "hello3",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP USER user1, user3;",
                    expected: Expected::Tag("DROP ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 4;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user1""#, ..E }),
                    username: "user1",
                    password: "hello1",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user2",
                    password: "hello2",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 6;",
                    expected: Expected::Error(Diagnostic { severity: "FATAL", code: "28P01", message: r#"password authentication failed for user "user3""#, ..E }),
                    username: "user3",
                    password: "hello3",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP ROLE removes all inherited privileges",
            set_up_script: &[
                "CREATE TABLE drop_role_table (v integer);",
                "INSERT INTO drop_role_table VALUES (1);",
                "CREATE SEQUENCE drop_role_sequence;",
                "CREATE FUNCTION drop_role_routine() RETURNS integer AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE ROLE dropped_group;",
                "CREATE USER surviving_member PASSWORD 'password';",
                "GRANT CREATE ON DATABASE postgres TO dropped_group;",
                "GRANT CREATE ON SCHEMA public TO dropped_group;",
                "GRANT SELECT ON drop_role_table TO dropped_group;",
                "GRANT USAGE ON SEQUENCE drop_role_sequence TO dropped_group;",
                "GRANT EXECUTE ON FUNCTION drop_role_routine() TO dropped_group;",
                "GRANT dropped_group TO surviving_member;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE drop_role_schema_table (v integer);",
                    expected: Expected::Tag("CREATE TABLE"),
                    username: "surviving_member",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM drop_role_table;",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "surviving_member",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('drop_role_sequence');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "surviving_member",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT drop_role_routine();",
                    expected: Expected::Rows {
                        columns: &[Column("drop_role_routine", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "surviving_member",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE dropped_group;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: r#"role "dropped_group" cannot be dropped because some objects depend on it"#, detail: r#"privileges for database postgres
privileges for schema public
privileges for table drop_role_table
privileges for sequence drop_role_sequence
privileges for function drop_role_routine()"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE drop_role_denied_schema_table (v integer);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    username: "surviving_member",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM drop_role_table;",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    username: "surviving_member",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('drop_role_sequence');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    username: "surviving_member",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT drop_role_routine();",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    username: "surviving_member",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE DATABASE authorization",
            set_up_script: &[
                "CREATE ROLE demo LOGIN PASSWORD 'password';",
                "GRANT ALL PRIVILEGES ON DATABASE postgres TO demo;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT rolsuper, rolcreatedb, rolcreaterole FROM pg_roles WHERE rolname = 'demo';",
                    expected: Expected::Rows {
                        columns: &[Column("rolsuper", BOOL), Column("rolcreatedb", BOOL), Column("rolcreaterole", BOOL)],
                        rows: &[
                            &[T("f"), T("f"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DATABASE made_by_demo;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to create database", ..E }),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT datname FROM pg_database WHERE datname = 'made_by_demo';",
                    expected: Expected::Rows {
                        columns: &[Column("datname", NAME)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER ROLE demo CREATEDB;",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DATABASE made_by_demo;",
                    expected: Expected::Tag("CREATE DATABASE"),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER ROLE demo NOCREATEDB;",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DATABASE denied_after_revoke;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied to create database", ..E }),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE made_by_demo;",
                    expected: Expected::Tag("DROP DATABASE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER ROLE demo SUPERUSER;",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE DATABASE made_by_super;",
                    expected: Expected::Tag("CREATE DATABASE"),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE made_by_super;",
                    expected: Expected::Tag("DROP DATABASE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP DATABASE authorization",
            set_up_script: &[
                "CREATE DATABASE victim;",
                "CREATE ROLE demo LOGIN PASSWORD 'password';",
                "GRANT ALL PRIVILEGES ON DATABASE victim TO demo;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP DATABASE victim;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "must be owner of database victim", ..E }),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE IF EXISTS victim;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "must be owner of database victim", ..E }),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE IF EXISTS missing_database;",
                    expected: Expected::Tag("DROP DATABASE"),
                    notices: &[Diagnostic { code: "00000", message: r#"database "missing_database" does not exist, skipping"#, ..N }],
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE missing_database;",
                    expected: Expected::Error(Diagnostic { code: "3D000", message: r#"database "missing_database" does not exist"#, ..E }),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER ROLE demo CREATEDB;",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE victim;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "must be owner of database victim", ..E }),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT datname FROM pg_database WHERE datname = 'victim';",
                    expected: Expected::Rows {
                        columns: &[Column("datname", NAME)],
                        rows: &[
                            &[T("victim")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER ROLE demo SUPERUSER;",
                    expected: Expected::Tag("ALTER ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP DATABASE victim;",
                    expected: Expected::Tag("DROP DATABASE"),
                    username: "demo",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT datname FROM pg_database WHERE datname = 'victim';",
                    expected: Expected::Rows {
                        columns: &[Column("datname", NAME)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "GRANT/REVOKE SELECT Privilege",
            set_up_script: &[
                "CREATE USER user1 PASSWORD 'a';",
                "CREATE USER user2 PASSWORD 'b';",
                "GRANT ALL PRIVILEGES ON SCHEMA public TO user1;",
                "GRANT ALL PRIVILEGES ON SCHEMA public TO user2;",
                "GRANT ALL PRIVILEGES ON test TO user1 WITH GRANT OPTION;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE test (pk INT4 PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1), (5), (6);",
                    expected: Expected::Tag("INSERT 0 3"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user2",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT SELECT ON test TO user2;",
                    expected: Expected::Tag("GRANT"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "user2",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "REVOKE SELECT ON test FROM user2;",
                    expected: Expected::Tag("REVOKE"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user2",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT SELECT ON test TO PUBLIC;",
                    expected: Expected::Tag("GRANT"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "user2",
                    password: "b",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "INSERT, UPDATE, DELETE Privileges",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY);",
                "INSERT INTO test VALUES (1), (6), (7);",
                "CREATE USER user1 PASSWORD 'a';",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH cte AS (SELECT * FROM test ORDER BY pk) SELECT * FROM cte;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (10);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET pk=pk+20;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk > 3;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT SELECT, INSERT, UPDATE, DELETE ON test TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("6")],
                            &[T("7")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH cte AS (SELECT * FROM test ORDER BY pk) SELECT * FROM cte;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("6")],
                            &[T("7")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (10);",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET pk=pk+20;",
                    expected: Expected::Tag("UPDATE 4"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = 21;",
                    expected: Expected::Tag("DELETE 1"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("26")],
                            &[T("27")],
                            &[T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "REVOKE SELECT, INSERT, UPDATE, DELETE ON test FROM user1;",
                    expected: Expected::Tag("REVOKE"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (100);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET pk=pk+200;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk > 3;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table test", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CREATE privilege ON SEQUENCE and ROUTINE",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "CREATE USER user1 PASSWORD 'a';",
                "CREATE USER user2 PASSWORD 'b';",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION testfunc3() RETURNS int AS $$ BEGIN RETURN 3; END; $$ LANGUAGE plpgsql",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE interpreted_example_3(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('3' || input); END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE genre_id_seq_by_3 AS integer START WITH 1 INCREMENT BY 2 NO MINVALUE NO MAXVALUE CACHE 1;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT CREATE ON SCHEMA public TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION testfunc3() RETURNS int AS $$ BEGIN RETURN 3; END; $$ LANGUAGE plpgsql",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE interpreted_example_3(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('3' || input); END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE genre_id_seq_by_3 AS integer START WITH 1 INCREMENT BY 2 NO MINVALUE NO MAXVALUE CACHE 1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_3');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    flow: Flow::Query,
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "REVOKE USAGE ON SEQUENCE genre_id_seq_by_3 FROM user1;",
                    expected: Expected::Tag("REVOKE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_3');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    flow: Flow::Query,
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE genre_id_seq_by_3 OWNER TO auth_test_user;",
                    expected: Expected::Tag("ALTER SEQUENCE"),
                    notices: &[Diagnostic { severity: "WARNING", message: "OWNER TO is unsupported and ignored", ..E }],
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_3');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "owner of a relation has all privileges granted by default and cannot be revoked unless ownership is altered",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "CREATE USER user1 PASSWORD 'a';",
                "CREATE USER user2 PASSWORD 'b';",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE mytable (pk int);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT CREATE ON SCHEMA public TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE mytable (pk int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * from mytable;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table mytable", ..E }),
                    flow: Flow::Query,
                    username: "user1",
                    password: "a",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "GRANT/REVOKE USAGE ON SCHEMA",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "CREATE USER user1 PASSWORD 'a';",
                "CREATE SEQUENCE genre_id_seq_by_3 AS integer START WITH 1 INCREMENT BY 2 NO MINVALUE NO MAXVALUE CACHE 1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_3');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT USAGE ON SCHEMA public TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_3');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "REVOKE USAGE ON SCHEMA public FROM user1;",
                    expected: Expected::Tag("REVOKE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_3');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "privileges ON FUNCTION",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "CREATE USER user1 PASSWORD 'a';",
                "CREATE FUNCTION testfunc1() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql",
                "CREATE FUNCTION testfunc2() RETURNS int AS $$ BEGIN RETURN 2; END; $$ LANGUAGE plpgsql",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT testfunc1();",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for routine testfunc1", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT ALL ON FUNCTION public.testfunc1() TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT testfunc1();",
                    expected: Expected::Rows {
                        columns: &[Column("testfunc1", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT testfunc2();",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for routine testfunc2", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "REVOKE ALL ON FUNCTION testfunc1() FROM user1;",
                    expected: Expected::Tag("REVOKE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT testfunc1();",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for routine testfunc1", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT UPPER('hello');",
                    expected: Expected::Rows {
                        columns: &[Column("upper", TEXT)],
                        rows: &[
                            &[T("HELLO")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Already on branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "privileges ON PROCEDURE",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "CREATE USER user1 PASSWORD 'a';",
                "CREATE TABLE test (v1 TEXT);",
                "CREATE PROCEDURE public.interpreted_example_1(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('1' || input); END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE interpreted_example_3(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('3' || input); END; $$ LANGUAGE plpgsql;",
                "GRANT ALL PRIVILEGES ON test TO user1 WITH GRANT OPTION;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CALL interpreted_example_1('12');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for routine interpreted_example_1", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT ALL ON PROCEDURE public.interpreted_example_1(input TEXT) TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CALL interpreted_example_1('22');",
                    expected: Expected::Tag("CALL"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", TEXT)],
                        rows: &[
                            &[T("122")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CALL interpreted_example_3('32');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for routine interpreted_example_3", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", TEXT)],
                        rows: &[
                            &[T("122")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "REVOKE ALL ON PROCEDURE public.interpreted_example_1(input TEXT) FROM user1;",
                    expected: Expected::Tag("REVOKE"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CALL interpreted_example_1('42');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for routine interpreted_example_1", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", TEXT)],
                        rows: &[
                            &[T("122")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT/USAGE privileges ON SEQUENCE",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "CREATE USER user1 PASSWORD 'a';",
                "CREATE USER user2 PASSWORD 'b';",
                "CREATE SEQUENCE genre_id_seq_by_2 AS integer START WITH 1 INCREMENT BY 2 NO MINVALUE NO MAXVALUE CACHE 1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_2');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT USAGE ON SEQUENCE public.genre_id_seq_by_2 TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_2');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT SELECT ON SEQUENCE public.genre_id_seq_by_2 TO user2;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_2');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user2",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT is_called FROM genre_id_seq_by_2;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: genre_id_seq_by_2", ..E }),
                    flow: Flow::Query,
                    username: "user2",
                    password: "b",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "user can create table with sequence that they don't have privileges for but they cannot use it'",
            set_up_script: &[
                "create user if not exists 'auth_test_super' with superuser password 'auth_test_spass';",
                "CREATE USER user1 PASSWORD 'a';",
                "CREATE SEQUENCE genre_id_seq_by_2 AS integer START WITH 1 INCREMENT BY 2 NO MINVALUE NO MAXVALUE CACHE 1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT nextval('genre_id_seq_by_2');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT CREATE ON SCHEMA public TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "create table test_by_user1 (pk int, v1 INTEGER DEFAULT nextval('genre_id_seq_by_2'));",
                    expected: Expected::Tag("CREATE TABLE"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into test_by_user1(pk) values (3);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for schema public", ..E }),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test_by_user1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "GRANT USAGE ON SEQUENCE public.genre_id_seq_by_2 TO user1;",
                    expected: Expected::Tag("GRANT"),
                    username: "auth_test_super",
                    password: "auth_test_spass",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "insert into test_by_user1(pk) values (3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "user1",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test_by_user1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("3"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "user1",
                    password: "a",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Branch Control: Namespace entries block",
            set_up_script: &[
                "DELETE FROM dolt_branch_control WHERE user = '%';",
                "INSERT INTO dolt_branch_control VALUES ('%', '%', 'postgres', '%', 'admin');",
                "CREATE USER testuser PASSWORD 'a';",
                "GRANT ALL PRIVILEGES ON SCHEMA public TO testuser;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('otherbranch1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "testuser",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_branch_namespace_control VALUES ('%', 'other%', 'postgres', '%');",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('otherbranch2');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "`testuser`@`127.0.0.1` cannot create a branch named `otherbranch2`", ..E }),
                    username: "testuser",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_branch_namespace_control VALUES ('%', 'other%', 'testuser', '%');",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('otherbranch2');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "testuser",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_branch_namespace_control VALUES ('%', 'otherbranch%', 'postgres', '%');",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('otherbranch3');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "`testuser`@`127.0.0.1` cannot create a branch named `otherbranch3`", ..E }),
                    username: "testuser",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('other3');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "testuser",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_branch_namespace_control VALUES ('%', 'otherbranch%', 'testuser', '%');",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "postgres",
                    password: "password",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('otherbranch3');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "testuser",
                    password: "a",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Require admin to modify tables",
            set_up_script: &[
                r#"DELETE FROM dolt_branch_control WHERE "user" = '%';"#,
                "INSERT INTO dolt_branch_control VALUES ('%', '%', 'postgres', '%', 'admin');",
                "CREATE USER a PASSWORD 'a';",
                "CREATE USER b PASSWORD 'b';",
                "GRANT ALL PRIVILEGES ON SCHEMA public TO a;",
                "GRANT ALL PRIVILEGES ON SCHEMA public TO b;",
                "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO a;",
                "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO b;",
                "INSERT INTO dolt_branch_control VALUES ('%', 'other', 'a', '%', 'write'), ('%', 'prefix%', 'a', '%', 'admin')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM dolt_branch_control WHERE "branch" = 'other';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`a`@`127.0.0.1` cannot delete the row ["%", "other", "a", "%"]"#, ..E }),
                    username: "a",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM dolt_branch_control WHERE "branch" = 'other';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`b`@`127.0.0.1` cannot delete the row ["%", "other", "a", "%"]"#, ..E }),
                    username: "b",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM dolt_branch_control WHERE "branch" = 'prefix%';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`b`@`127.0.0.1` cannot delete the row ["%", "prefix%", "a", "%"]"#, ..E }),
                    username: "b",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_branch_control VALUES ('%', 'prefix1%', 'b', '%', 'write');",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "a",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM dolt_branch_control WHERE "branch" = 'prefix1%';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`b`@`127.0.0.1` cannot delete the row ["%", "prefix1%", "b", "%"]"#, ..E }),
                    username: "b",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE dolt_branch_control SET "branch" = 'other1' WHERE "user" = 'b' AND "branch" = 'prefix1%';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`a`@`127.0.0.1` cannot update the row ["%", "prefix1%", "b", "%"] to the new branch expression ["%", "other1"]"#, ..E }),
                    username: "a",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE dolt_branch_control SET "permissions" = 'admin' WHERE "user" = 'b' AND "branch" = 'prefix1%';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`b`@`127.0.0.1` cannot update the row ["%", "prefix1%", "b", "%"]"#, ..E }),
                    username: "b",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE dolt_branch_control SET "permissions" = 'admin' WHERE "user" = 'b' AND "branch" = 'prefix1%';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    username: "a",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM dolt_branch_control WHERE "branch" = 'prefix1%';"#,
                    expected: Expected::Tag("DELETE 1"),
                    username: "b",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_branch_control VALUES ('%', 'prefix1%', 'b', '%', 'admin');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`b`@`127.0.0.1` cannot add the row ["%", "prefix1%", "b", "%", "admin"]"#, ..E }),
                    username: "b",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_branch_namespace_control VALUES ('%', 'prefix___', 'a', '%');",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "a",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_branch_namespace_control VALUES ('%', 'prefix', 'b', '%');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`b`@`127.0.0.1` cannot add the row ["%", "prefix", "b", "%"]"#, ..E }),
                    username: "b",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE dolt_branch_namespace_control SET "branch" = 'prefix%';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    username: "a",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE dolt_branch_namespace_control SET "branch" = 'other';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`a`@`127.0.0.1` cannot update the row ["%", "prefix%", "a", "%"] to the new branch expression ["%", "other"]"#, ..E }),
                    username: "a",
                    password: "a",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM dolt_branch_namespace_control WHERE "branch" = 'prefix%';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`b`@`127.0.0.1` cannot delete the row ["%", "prefix%", "a", "%"]"#, ..E }),
                    username: "b",
                    password: "b",
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE dolt_branch_namespace_control SET "branch" = 'anything';"#,
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"`b`@`127.0.0.1` cannot update the row ["%", "prefix%", "a", "%"]"#, ..E }),
                    username: "b",
                    password: "b",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Built-in routines are executable without an explicit grant",
            set_up_script: &[
                "CREATE TABLE t3327 (x INT);",
                "INSERT INTO t3327 VALUES (1), (2), (3);",
                "CREATE ROLE reader LOGIN PASSWORD 'password';",
                "GRANT SELECT ON t3327 TO reader;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM t3327;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*), SUM(x), MAX(x) FROM t3327;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("sum", INT8), Column("max", INT4)],
                        rows: &[
                            &[T("3"), T("6"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x, ROW_NUMBER() OVER (ORDER BY x) FROM t3327 ORDER BY x;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("row_number", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_catalog.lower('A'), lower('A');",
                    expected: Expected::Rows {
                        columns: &[Column("lower", TEXT), Column("lower", TEXT)],
                        rows: &[
                            &[T("a"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CTE names are not checked as tables",
            set_up_script: &[
                "CREATE TABLE edges (src TEXT, dst TEXT, project_id TEXT);",
                "INSERT INTO edges VALUES ('a', 'b', 'p'), ('b', 'c', 'p');",
                "CREATE TABLE secret (x INT);",
                "INSERT INTO secret VALUES (1);",
                "CREATE TABLE target (src TEXT, dst TEXT);",
                "CREATE ROLE reader LOGIN PASSWORD 'password';",
                "GRANT SELECT ON edges TO reader;",
                "GRANT SELECT, INSERT, UPDATE, DELETE ON target TO reader;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "WITH traversal AS (SELECT src, dst FROM edges) SELECT * FROM traversal ORDER BY src;",
                    expected: Expected::Rows {
                        columns: &[Column("src", TEXT), Column("dst", TEXT)],
                        rows: &[
                            &[T("a"), T("b")],
                            &[T("b"), T("c")],
                        ],
                        tag: "SELECT 2",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH RECURSIVE r(n) AS (SELECT 1 UNION ALL SELECT n + 1 FROM r WHERE n < 3) SELECT * FROM r;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH RECURSIVE walk(node) AS (SELECT 'a'::TEXT UNION SELECT e.dst FROM edges e JOIN walk w ON e.src = w.node) SELECT * FROM walk ORDER BY node;",
                    expected: Expected::Rows {
                        columns: &[Column("node", TEXT)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                            &[T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH s AS (SELECT * FROM secret) SELECT * FROM s;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table secret", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH secret AS (SELECT 5 AS x) SELECT * FROM secret;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH t AS (SELECT src FROM edges) SELECT * FROM edges WHERE src IN (SELECT src FROM t) ORDER BY src;",
                    expected: Expected::Rows {
                        columns: &[Column("src", TEXT), Column("dst", TEXT), Column("project_id", TEXT)],
                        rows: &[
                            &[T("a"), T("b"), T("p")],
                            &[T("b"), T("c"), T("p")],
                        ],
                        tag: "SELECT 2",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH t AS (SELECT src, dst FROM edges) SELECT a.src FROM t a JOIN t b ON a.dst = b.src;",
                    expected: Expected::Rows {
                        columns: &[Column("src", TEXT)],
                        rows: &[
                            &[T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH a AS (SELECT src FROM edges), b AS (SELECT * FROM a) SELECT COUNT(*) FROM b;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH t AS (SELECT src, dst FROM edges) INSERT INTO target SELECT * FROM t;",
                    expected: Expected::Tag("INSERT 0 2"),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO target WITH t AS (SELECT src, dst FROM edges) SELECT * FROM t;",
                    expected: Expected::Tag("INSERT 0 2"),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH t AS (SELECT src FROM edges) UPDATE target SET dst = 'z' WHERE src IN (SELECT src FROM t);",
                    expected: Expected::Tag("UPDATE 4"),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM target ORDER BY src;",
                    expected: Expected::Rows {
                        columns: &[Column("src", TEXT), Column("dst", TEXT)],
                        rows: &[
                            &[T("a"), T("z")],
                            &[T("a"), T("z")],
                            &[T("b"), T("z")],
                            &[T("b"), T("z")],
                        ],
                        tag: "SELECT 4",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "WITH t AS (SELECT src FROM edges) DELETE FROM target WHERE src IN (SELECT src FROM t);",
                    expected: Expected::Tag("DELETE 4"),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM target;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Reading as a role with only table privileges",
            set_up_script: &[
                "CREATE TABLE edges (src TEXT, dst TEXT);",
                "INSERT INTO edges VALUES ('a', 'b');",
                "CREATE TABLE secret (x INT);",
                "INSERT INTO secret VALUES (1);",
                "CREATE VIEW v_edges AS SELECT src, dst FROM edges;",
                "CREATE VIEW v_secret AS SELECT x FROM secret;",
                "CREATE ROLE reader LOGIN PASSWORD 'password';",
                "GRANT SELECT ON edges TO reader;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) > 0 FROM pg_catalog.pg_class;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) > 0 FROM information_schema.tables;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) > 0 FROM pg_tables;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM v_edges;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view v_edges", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM missing_table;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "missing_table" does not exist"#, position: 15, ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT rolname FROM pg_authid;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table pg_authid", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) > 0 FROM pg_roles;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE secret SET x = 2;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table secret", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM secret;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table secret", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.secret;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table secret", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT SELECT ON v_edges, v_secret TO reader;",
                    expected: Expected::Tag("GRANT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM v_edges;",
                    expected: Expected::Rows {
                        columns: &[Column("src", TEXT), Column("dst", TEXT)],
                        rows: &[
                            &[T("a"), T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM v_secret;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO missing_table VALUES (1);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "missing_table" does not exist"#, position: 13, ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO v_edges VALUES ('x', 'y');",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view v_edges", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "TRUNCATE missing_table;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "missing_table" does not exist"#, ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "TRUNCATE secret;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table secret", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE missing_table;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"table "missing_table" does not exist"#, ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE secret;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "must be owner of table secret", ..E }),
                    username: "reader",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Views are read with the privileges of their owner",
            set_up_script: &[
                "CREATE SCHEMA reporting;",
                "CREATE TABLE reporting.orders (id INT);",
                "INSERT INTO reporting.orders VALUES (1);",
                "CREATE VIEW reporting.order_view AS SELECT id FROM reporting.orders;",
                "CREATE TABLE items (id INT);",
                "INSERT INTO items VALUES (1);",
                "CREATE VIEW items_view AS SELECT id FROM items;",
                "CREATE ROLE analyst LOGIN PASSWORD 'password';",
                "GRANT USAGE ON SCHEMA reporting TO analyst;",
                "GRANT SELECT ON reporting.order_view TO analyst;",
                "CREATE ROLE analyst2 LOGIN PASSWORD 'password';",
                "GRANT USAGE ON SCHEMA reporting TO analyst2;",
                "GRANT SELECT ON reporting.order_view, reporting.orders TO analyst2;",
                "CREATE ROLE writer LOGIN PASSWORD 'password';",
                "GRANT SELECT ON items_view TO writer;",
                "CREATE ROLE writer2 LOGIN PASSWORD 'password';",
                "GRANT SELECT ON items_view, items TO writer2;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM reporting.order_view;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "analyst",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM reporting.order_view;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "analyst2",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM items_view;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "writer",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items_view VALUES (2);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view items_view", ..E }),
                    username: "writer",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE items_view SET id = 3;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view items_view", ..E }),
                    username: "writer",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM items_view WHERE id = 1;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view items_view", ..E }),
                    username: "writer",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items_view VALUES (2);",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view items_view", ..E }),
                    username: "writer2",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE items_view SET id = 3;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view items_view", ..E }),
                    username: "writer2",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM items_view WHERE id = 1;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for view items_view", ..E }),
                    username: "writer2",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Upserts check INSERT and UPDATE on the resolved table",
            set_up_script: &[
                "CREATE TABLE counts (id INT PRIMARY KEY, n INT);",
                "INSERT INTO counts VALUES (1, 1);",
                "CREATE ROLE inserter LOGIN PASSWORD 'password';",
                "GRANT SELECT, INSERT ON counts TO inserter;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO counts VALUES (1, 2) ON CONFLICT (id) DO UPDATE SET n = excluded.n;",
                    expected: Expected::Error(Diagnostic { code: "42501", message: "permission denied for table counts", ..E }),
                    username: "inserter",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO missing_table VALUES (1) ON CONFLICT (id) DO UPDATE SET id = 1;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "missing_table" does not exist"#, position: 13, ..E }),
                    username: "inserter",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT UPDATE ON counts TO inserter;",
                    expected: Expected::Tag("GRANT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO counts VALUES (1, 2) ON CONFLICT (id) DO UPDATE SET n = excluded.n;",
                    expected: Expected::Tag("INSERT 0 1"),
                    username: "inserter",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM counts;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "inserter",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
    ]);
}
