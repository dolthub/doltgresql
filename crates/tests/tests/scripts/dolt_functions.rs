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
fn test_dolt_add() {
    run_scripts(&[
        ScriptTest {
            name: "Add all using dot",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
        ScriptTest {
            name: "Add all using -A",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('-A');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
        ScriptTest {
            name: "Add all individually",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('t_simple','t_composite','t_array','t_serial','t_default_simple','t_checked','t_fk_parent','t_fk_child','t_unique','t_generated','t_trigger','t_default_func');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('f_trigger()','f_default()');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("14")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('t_serial_pk_seq','t_trigger.trig_trigger');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
fn test_dolt_branch() {
    run_scripts(&[
        ScriptTest {
            name: "All branch options",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('-A');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('original');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_simple VALUES (4);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('-A');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('-c', 'main', 'copy');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('-c', '-f', 'original', 'forcecopy');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('main')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Already on branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('original')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'original'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('copy')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'copy'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('forcecopy')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'forcecopy'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('-d', 'forcecopy')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Cannot delete checked out branch 'forcecopy'", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('-d', 'original')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_BRANCH('-m','copy','renamedcopy')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_branch", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('original')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "error: tablespec 'original' did not match any table(s) known to dolt", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('copy')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "error: tablespec 'copy' did not match any table(s) known to dolt", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('renamedcopy')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'renamedcopy'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
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
fn test_dolt_branch_status() {
    run_scripts(&[
        ScriptTest {
            name: "Smoke test",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "SELECT DOLT_ADD('-A');",
                "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "SELECT DOLT_ADD('-A');",
                "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_BRANCH_STATUS('main', 'original');",
                    expected: Expected::Rows {
                        columns: &[Column("branch", TEXT), Column("commits_ahead", NUMERIC), Column("commits_behind", NUMERIC)],
                        rows: &[
                            &[T("original"), T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_BRANCH_STATUS('original', 'main');",
                    expected: Expected::Rows {
                        columns: &[Column("branch", TEXT), Column("commits_ahead", NUMERIC), Column("commits_behind", NUMERIC)],
                        rows: &[
                            &[T("main"), T("1"), T("0")],
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
fn test_dolt_checkout() {
    run_scripts(&[
        ScriptTest {
            name: "All checkout options",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT DOLT_ADD('-A');",
                "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT DOLT_ADD('-A');",
                "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('-b', 'checkoutbranch');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'checkoutbranch'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('checkoutbranch');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Already on branch 'checkoutbranch'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('main')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('original')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'original'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
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

#[test]
fn test_dolt_cherry_pick() {
    run_scripts(&[
        ScriptTest {
            name: "Smoke test",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT DOLT_ADD('-A');",
                "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT DOLT_ADD('-A');",
                "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('original')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT substring(DOLT_CHERRY_PICK('main')::text, 34);",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T(",0,0,0)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
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
fn test_dolt_clean() {
    run_scripts(&[
        ScriptTest {
            name: "Clean all",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CLEAN();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
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
        ScriptTest {
            name: "Clean all except one",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT DOLT_ADD('t_simple');",
                "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("15")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CLEAN();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
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
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
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
        ScriptTest {
            name: "Clean all by name",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CLEAN('t_simple','t_composite','t_array','t_serial','t_default_simple','t_checked','t_fk_parent','t_fk_child','t_unique','t_generated','t_trigger','t_default_func','f_trigger()','f_default()');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
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
        ScriptTest {
            name: "Dry run",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CLEAN('--dry-run');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
fn test_dolt_commit() {
    run_scripts(&[
        ScriptTest {
            name: "Stage all then commit",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('-A');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
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
        ScriptTest {
            name: "Commit with staging all",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
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
        ScriptTest {
            name: "Commit with staging modifications",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('t_simple','t_composite','t_array');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(DOLT_COMMIT('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("13")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_simple VALUES (4);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_composite VALUES (3, 100);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t_array VALUES (ARRAY['stu']);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(DOLT_COMMIT('-a', '-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("13")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Allow and skip empty",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_COMMIT('-m', 'should_error');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "nothing to commit", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_COMMIT('--skip-empty', '-m', 'should_error');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_commit", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(DOLT_COMMIT('--allow-empty', '-m', 'initial')::text) = 32;",
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
fn test_dolt_conflicts_resolve() {
    run_scripts(&[
        ScriptTest {
            name: "Simple conflicts resolve all using --ours",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "SELECT DOLT_COMMIT('-Am', 'initial')",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "SELECT DOLT_COMMIT('-Am', 'second row on main')",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "SELECT DOLT_COMMIT('-Am', 'conflicted row on other')",
                "SELECT DOLT_CHECKOUT('main');",
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
                    query: "SELECT DOLT_MERGE('other', '--no-ff');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_merge", RECORD)],
                        rows: &[
                            &[T(r#"("",0,1,"conflicts found")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("t_simple"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CONFLICTS_RESOLVE('--ours', '.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_conflicts_resolve", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_COMMIT('-Am', 'commit merge');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Simple conflicts resolve all using --theirs",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "SELECT DOLT_COMMIT('-Am', 'initial')",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "SELECT DOLT_COMMIT('-Am', 'second row on main')",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "SELECT DOLT_COMMIT('-Am', 'conflicted row on other')",
                "SELECT DOLT_CHECKOUT('main');",
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
                    query: "SELECT DOLT_MERGE('other', '--no-ff');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_merge", RECORD)],
                        rows: &[
                            &[T(r#"("",0,1,"conflicts found")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("t_simple"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CONFLICTS_RESOLVE('--theirs', '.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_conflicts_resolve", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_COMMIT('-Am', 'commit merge');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Simple conflicts resolve table using --ours",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "SELECT DOLT_COMMIT('-Am', 'initial')",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "SELECT DOLT_COMMIT('-Am', 'second row on main')",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "SELECT DOLT_COMMIT('-Am', 'conflicted row on other')",
                "SELECT DOLT_CHECKOUT('main');",
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
                    query: "SELECT DOLT_MERGE('other', '--no-ff');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_merge", RECORD)],
                        rows: &[
                            &[T(r#"("",0,1,"conflicts found")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("t_simple"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CONFLICTS_RESOLVE('--ours', 't_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_conflicts_resolve", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_COMMIT('-Am', 'commit merge');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Simple conflicts resolve table using --theirs",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "SELECT DOLT_COMMIT('-Am', 'initial')",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "SELECT DOLT_COMMIT('-Am', 'second row on main')",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "SELECT DOLT_COMMIT('-Am', 'conflicted row on other')",
                "SELECT DOLT_CHECKOUT('main');",
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
                    query: "SELECT DOLT_MERGE('other', '--no-ff');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_merge", RECORD)],
                        rows: &[
                            &[T(r#"("",0,1,"conflicts found")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("t_simple"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CONFLICTS_RESOLVE('--theirs', 't_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_conflicts_resolve", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_COMMIT('-Am', 'commit merge');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Resolve all using --ours",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
                "CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "INSERT INTO t_composite VALUES (1, 1, 1);",
                "INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
                "INSERT INTO t_serial VALUES (DEFAULT, 1);",
                "INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
                "INSERT INTO t_trigger VALUES (1, 1);",
                "INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "INSERT INTO t_composite VALUES (2, 2, 2);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 2);",
                "INSERT INTO t_serial VALUES (DEFAULT, 2);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "INSERT INTO t_composite VALUES (2, 2, 3);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 3);",
                "INSERT INTO t_serial VALUES (DEFAULT, 3);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE('main');",
                    flow: Flow::Query,
                    skip: Some("the Go server panics, with a stack trace that differs between runs"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Resolve all using --theirs",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
                "CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "INSERT INTO t_composite VALUES (1, 1, 1);",
                "INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
                "INSERT INTO t_serial VALUES (DEFAULT, 1);",
                "INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
                "INSERT INTO t_trigger VALUES (1, 1);",
                "INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "INSERT INTO t_composite VALUES (2, 2, 2);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 2);",
                "INSERT INTO t_serial VALUES (DEFAULT, 2);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "INSERT INTO t_composite VALUES (2, 2, 3);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 3);",
                "INSERT INTO t_serial VALUES (DEFAULT, 3);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE('main');",
                    flow: Flow::Query,
                    skip: Some("the Go server panics, with a stack trace that differs between runs"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Resolve individual items",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
                "CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "INSERT INTO t_composite VALUES (1, 1, 1);",
                "INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
                "INSERT INTO t_serial VALUES (DEFAULT, 1);",
                "INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
                "INSERT INTO t_trigger VALUES (1, 1);",
                "INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "INSERT INTO t_composite VALUES (2, 2, 2);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 2);",
                "INSERT INTO t_serial VALUES (DEFAULT, 2);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "INSERT INTO t_composite VALUES (2, 2, 3);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 3);",
                "INSERT INTO t_serial VALUES (DEFAULT, 3);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE('main');",
                    flow: Flow::Query,
                    skip: Some("the Go server panics, with a stack trace that differs between runs"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_diff() {
    run_scripts(&[
        ScriptTest {
            name: "Single commit",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("from_pk", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_pk1, from_pk2 FROM DOLT_DIFF('main', 'original', 't_composite');",
                    expected: Expected::Rows {
                        columns: &[Column("from_pk1", INT2), Column("from_pk2", INT8)],
                        rows: &[
                            &[T("3"), T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_array');",
                    expected: Expected::Rows {
                        columns: &[Column("from_pk", TEXT_ARRAY)],
                        rows: &[
                            &[T("{stu}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_serial');",
                    expected: Expected::Rows {
                        columns: &[Column("from_pk", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_v1, from_v2 FROM DOLT_DIFF('main', 'original', 't_default_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("from_v1", INT8), Column("from_v2", INT8)],
                        rows: &[
                            &[T("98"), T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_v1 FROM DOLT_DIFF('main', 'original', 't_checked');",
                    expected: Expected::Rows {
                        columns: &[Column("from_v1", NUMERIC)],
                        rows: &[
                            &[T("99")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_fk_parent');",
                    expected: Expected::Rows {
                        columns: &[Column("from_pk", INT4)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_pk FROM DOLT_DIFF('main', 'original', 't_fk_child');",
                    expected: Expected::Rows {
                        columns: &[Column("from_pk", INT4)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_v1 FROM DOLT_DIFF('main', 'original', 't_unique');",
                    expected: Expected::Rows {
                        columns: &[Column("from_v1", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_v1, from_v2 FROM DOLT_DIFF('main', 'original', 't_generated');",
                    expected: Expected::Rows {
                        columns: &[Column("from_v1", INT8), Column("from_v2", INT8)],
                        rows: &[
                            &[T("11"), T("11000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_v1 FROM DOLT_DIFF('main', 'original', 't_trigger');",
                    expected: Expected::Rows {
                        columns: &[Column("from_v1", INT8)],
                        rows: &[
                            &[T("66")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_v1, from_v2 FROM DOLT_DIFF('main', 'original', 't_default_func');",
                    expected: Expected::Rows {
                        columns: &[Column("from_v1", INT8), Column("from_v2", INT8)],
                        rows: &[
                            &[T("34"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_DIFF('main', 'original', 'f_trigger()');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "relation \"f_trigger()\" does not exist", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_DIFF('main', 'original', 'f_default()');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "relation \"f_default()\" does not exist", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_DIFF('main', 'original', 't_trigger.trig_trigger');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "relation \"t_trigger.trig_trigger\" does not exist", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_DIFF('main', 'original', 't_serial_pk_seq');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "relation \"t_serial_pk_seq\" does not exist", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt table functions WITH ORDINALITY",
            set_up_script: &[
                "CREATE TABLE t1 (pk INT4 PRIMARY KEY, v1 TEXT);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "INSERT INTO t1 VALUES (1, 'one'), (2, 'two');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pk, to_pk, ordinality FROM dolt_diff('HEAD', 'WORKING', 't1') WITH ORDINALITY;",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pk", INT4), Column("to_pk", INT4), Column("ordinality", INT8)],
                        rows: &[
                            &[T("added"), Null, T("1"), T("1")],
                            &[T("added"), Null, T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT ord, to_pk1 FROM dolt_diff('HEAD', 'WORKING', 't1') WITH ORDINALITY AS a(to_pk1, to_v11, to_commit1, to_commit_date1, from_pk1, from_v11, from_commit1, from_commit_date1, diff_type1, ord);",
                    expected: Expected::Rows {
                        columns: &[Column("ord", INT8), Column("to_pk1", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT message, ordinality FROM dolt_log() WITH ORDINALITY LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT), Column("ordinality", INT8)],
                        rows: &[
                            &[T("initial"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, ordinality FROM dolt_diff_stat('HEAD', 'WORKING') WITH ORDINALITY;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("ordinality", INT8)],
                        rows: &[
                            &[T("public.t1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_table_name, diff_type, ordinality FROM dolt_diff_summary('HEAD', 'WORKING') WITH ORDINALITY;",
                    expected: Expected::Rows {
                        columns: &[Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("ordinality", INT8)],
                        rows: &[
                            &[T("public.t1"), T("modified"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, diff_type, ordinality FROM dolt_patch('HEAD', 'WORKING') WITH ORDINALITY;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("diff_type", TEXT), Column("ordinality", INT8)],
                        rows: &[
                            &[T("public.t1"), T("data"), T("1")],
                            &[T("public.t1"), T("data"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING') WITH ORDINALITY;",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT), Column("ordinality", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "IN expressions over table function commit columns",
            set_up_script: &[
                "CREATE TABLE test (id INT PRIMARY KEY, val INT)",
                "INSERT INTO test VALUES (1, 1)",
                "SELECT dolt_commit('-Am', 'commit 1')",
                "INSERT INTO test VALUES (2, 2)",
                "SELECT dolt_commit('-Am', 'commit 2')",
                "UPDATE test SET val = 3 WHERE id = 1",
                "SELECT dolt_commit('-Am', 'commit 3')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_commit, from_commit, to_id FROM dolt_diff('HEAD~2', 'HEAD', 'test') WHERE to_commit IN ('HEAD', 'WORKING') ORDER BY to_id",
                    expected: Expected::Rows {
                        columns: &[Column("to_commit", TEXT), Column("from_commit", TEXT), Column("to_id", INT4)],
                        rows: &[
                            &[T("HEAD"), T("HEAD~2"), T("1")],
                            &[T("HEAD"), T("HEAD~2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_id FROM dolt_diff('HEAD~2', 'HEAD', 'test') WHERE to_commit IN (HASHOF('HEAD'), HASHOF('HEAD~1'))",
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT to_id FROM dolt_diff('HEAD~2', 'HEAD', 'test') WHERE to_commit NOT IN ('HEAD') ORDER BY to_id",
                    expected: Expected::Rows {
                        columns: &[Column("to_id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_log('HEAD') WHERE commit_hash IN (HASHOF('HEAD'), HASHOF('HEAD~1'))",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_log('HEAD') WHERE commit_hash IN ('not_a_hash')",
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
fn test_dolt_diff_stat() {
    run_scripts(&[
        ScriptTest {
            name: "Single commit",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_DIFF_STAT('main', 'original');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("15")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_simple")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_composite');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_composite")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_array');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_array")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_serial');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_serial")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_default_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_default_simple")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_checked');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_checked")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_fk_parent');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_fk_parent")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_fk_child');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_fk_child")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_unique');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_unique")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_generated');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_generated")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_trigger');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_trigger")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_default_func');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_default_func")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 'f_trigger()');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.f_trigger()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 'f_default()');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.f_default()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM DOLT_DIFF_STAT('main', 'original', 't_serial_pk_seq');",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("public.t_serial_pk_seq")],
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
fn test_dolt_diff_summary() {
    run_scripts(&[
        ScriptTest {
            name: "Single commit",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_DIFF_SUMMARY('main', 'original');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("15")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_simple")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_composite');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_composite")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_array');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_array")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_serial');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_serial")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_default_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_default_simple")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_checked');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_checked")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_fk_parent');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_fk_parent")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_fk_child');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_fk_child")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_unique');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_unique")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_generated');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_generated")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_trigger');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_trigger")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_default_func');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_default_func")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 'f_trigger()');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.f_trigger()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 'f_default()');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.f_default()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT from_table_name FROM DOLT_DIFF_SUMMARY('main', 'original', 't_serial_pk_seq');",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT)],
                        rows: &[
                            &[T("public.t_serial_pk_seq")],
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
fn test_dolt_function_smoke_tests() {
    run_scripts(&[
        ScriptTest {
            name: "smoke test select dolt_add and dolt_commit",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_add('.')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_commit('-am', 'new table')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.log",
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
                    query: "select message from dolt.log order by date desc limit 1",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "smoke test select dolt_merge",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
                "SELECT DOLT_COMMIT('-Am', 'new table');",
                "SELECT DOLT_CHECKOUT('-b', 'new-branch');",
                "CREATE TABLE t2 (pk int primary key);",
                "SELECT DOLT_COMMIT('-Am', 'new table on new branch');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE_BASE('main', 'new-branch');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.log",
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
                    query: "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select count(*) from dolt.log",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
        ScriptTest {
            name: "smoke test select dolt_merge dirty working set, same table",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
                "SELECT DOLT_COMMIT('-Am', 'new table');",
                "INSERT INTO t1 VALUES (1);",
                "SELECT DOLT_CHECKOUT('-b', 'new-branch');",
                "INSERT INTO t1 VALUES (2);",
                "SELECT DOLT_COMMIT('-Am', 'new row on new branch');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE_BASE('main', 'new-branch');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("modified")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"error: local changes would be stomped by merge:
	t1
 Please commit your changes before you merge."#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("modified")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "smoke test select dolt_merge dirty working set, different tables",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
                "SELECT DOLT_COMMIT('-Am', 'new table');",
                "INSERT INTO t1 VALUES (1);",
                "SELECT DOLT_CHECKOUT('-b', 'new-branch');",
                "CREATE TABLE t2 (pk int primary key);",
                "SELECT DOLT_COMMIT('-Am', 'new row on new branch');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE_BASE('main', 'new-branch');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('main');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("modified")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("modified")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "smoke test select dolt_reset",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
                "INSERT INTO t1 VALUES (1);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_ADD('t1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_RESET('t1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "smoke test select dolt_clean",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
                "INSERT INTO t1 VALUES (1);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CLEAN('t1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE t1 (pk int primary key);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CLEAN();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE TABLE t1 (pk int primary key);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CLEAN();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_clean", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "smoke test select dolt_checkout(table)",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
                "INSERT INTO t1 VALUES (1);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('t1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "smoke test select dolt diff functions and tables",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
                "INSERT INTO t1 VALUES (1);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("rows_unmodified", INT8), Column("rows_added", INT8), Column("rows_deleted", INT8), Column("rows_modified", INT8), Column("cells_added", INT8), Column("cells_deleted", INT8), Column("cells_modified", INT8), Column("old_row_count", INT8), Column("new_row_count", INT8), Column("old_cell_count", INT8), Column("new_cell_count", INT8)],
                        rows: &[
                            &[T("public.t1"), T("0"), T("1"), T("0"), T("0"), T("1"), T("0"), T("0"), T("0"), T("1"), T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("rows_unmodified", INT8), Column("rows_added", INT8), Column("rows_deleted", INT8), Column("rows_modified", INT8), Column("cells_added", INT8), Column("cells_deleted", INT8), Column("cells_modified", INT8), Column("old_row_count", INT8), Column("new_row_count", INT8), Column("old_cell_count", INT8), Column("new_cell_count", INT8)],
                        rows: &[
                            &[T("public.t1"), T("0"), T("1"), T("0"), T("0"), T("1"), T("0"), T("0"), T("0"), T("1"), T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.t1"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.t1"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pk", INT4), Column("to_pk", INT4)],
                        rows: &[
                            &[T("added"), Null, T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pk", INT4), Column("to_pk", INT4)],
                        rows: &[
                            &[T("added"), Null, T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pk, to_pk FROM dolt_commit_diff_t1 WHERE to_commit=HASHOF('main') AND from_commit='WORKING'",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pk", INT4), Column("to_pk", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.diff",
                    expected: Expected::Rows {
                        columns: &[Column("commit_hash", TEXT), Column("table_name", TEXT), Column("committer", TEXT), Column("email", TEXT), Column("date", TIMESTAMP), Column("message", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL), Column("author", TEXT), Column("author_email", TEXT), Column("author_date", TIMESTAMP)],
                        rows: &[
                            &[T("WORKING"), T("public.t1"), Null, Null, Null, Null, T("t"), T("t"), Null, Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("statement_order", NUMERIC), Column("table_name", TEXT), Column("diff_type", TEXT), Column("statement", TEXT)],
                        rows: &[
                            &[T("1"), T("public.t1"), T("schema"), T(r#"CREATE TABLE "t1" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                            &[T("2"), T("public.t1"), T("data"), T(r#"INSERT INTO "t1" ("pk") VALUES (1);"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("statement_order", NUMERIC), Column("table_name", TEXT), Column("diff_type", TEXT), Column("statement", TEXT)],
                        rows: &[
                            &[T("1"), T("public.t1"), T("schema"), T(r#"CREATE TABLE "t1" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                            &[T("2"), T("public.t1"), T("data"), T(r#"INSERT INTO "t1" ("pk") VALUES (1);"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[
                            &[T(""), T("public.t1"), T(""), T(r#"CREATE TABLE "t1" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[
                            &[T(""), T("public.t1"), T(""), T(r#"CREATE TABLE "t1" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_query_diff('select * from t1 as of main', 'select * from t1')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: t1", ..E }),
                    flow: Flow::Query,
                    skip: Some("Go's DOLT_QUERY_DIFF cannot parse AS OF in its queries; the Rust server parses the queries with Doltgres' grammar, where an unquoted revision is a syntax error"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "smoke test select dolt diff functions and tables for multiple schemas",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key);",
                "INSERT INTO t1 VALUES (1);",
                "CREATE SCHEMA testschema;",
                "CREATE TABLE testschema.t2 (pk int primary key);",
                "INSERT INTO testschema.t2 VALUES (1);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t1"), T("f"), T("new table")],
                            &[T("testschema.t2"), T("f"), T("new table")],
                            &[T("testschema"), T("f"), T("new schema")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("rows_unmodified", INT8), Column("rows_added", INT8), Column("rows_deleted", INT8), Column("rows_modified", INT8), Column("cells_added", INT8), Column("cells_deleted", INT8), Column("cells_modified", INT8), Column("old_row_count", INT8), Column("new_row_count", INT8), Column("old_cell_count", INT8), Column("new_cell_count", INT8)],
                        rows: &[
                            &[T("public.t1"), T("0"), T("1"), T("0"), T("0"), T("1"), T("0"), T("0"), T("0"), T("1"), T("0"), T("1")],
                            &[T("testschema.t2"), T("0"), T("1"), T("0"), T("0"), T("1"), T("0"), T("0"), T("0"), T("1"), T("0"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("rows_unmodified", INT8), Column("rows_added", INT8), Column("rows_deleted", INT8), Column("rows_modified", INT8), Column("cells_added", INT8), Column("cells_deleted", INT8), Column("cells_modified", INT8), Column("old_row_count", INT8), Column("new_row_count", INT8), Column("old_cell_count", INT8), Column("new_cell_count", INT8)],
                        rows: &[
                            &[T("public.t1"), T("0"), T("1"), T("0"), T("0"), T("1"), T("0"), T("0"), T("0"), T("1"), T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't2')",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("rows_unmodified", INT8), Column("rows_added", INT8), Column("rows_deleted", INT8), Column("rows_modified", INT8), Column("cells_added", INT8), Column("cells_deleted", INT8), Column("cells_modified", INT8), Column("old_row_count", INT8), Column("new_row_count", INT8), Column("old_cell_count", INT8), Column("new_cell_count", INT8)],
                        rows: &[
                            &[T("testschema.t2"), T("0"), T("1"), T("0"), T("0"), T("1"), T("0"), T("0"), T("0"), T("1"), T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.t1"), T("added"), T("t"), T("t")],
                            &[T(""), T("testschema.t2"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("public.t1"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't2')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("diff_type", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL)],
                        rows: &[
                            &[T(""), T("testschema.t2"), T("added"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pk", INT4), Column("to_pk", INT4)],
                        rows: &[
                            &[T("added"), Null, T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't2')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pk", INT4), Column("to_pk", INT4)],
                        rows: &[
                            &[T("added"), Null, T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pk, to_pk FROM dolt_commit_diff_t1 WHERE to_commit=HASHOF('main') AND from_commit='WORKING'",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pk", INT4), Column("to_pk", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.diff",
                    expected: Expected::Rows {
                        columns: &[Column("commit_hash", TEXT), Column("table_name", TEXT), Column("committer", TEXT), Column("email", TEXT), Column("date", TIMESTAMP), Column("message", TEXT), Column("data_change", BOOL), Column("schema_change", BOOL), Column("author", TEXT), Column("author_email", TEXT), Column("author_date", TIMESTAMP)],
                        rows: &[
                            &[T("WORKING"), T("public.t1"), Null, Null, Null, Null, T("t"), T("t"), Null, Null, Null],
                            &[T("WORKING"), T("testschema.t2"), Null, Null, Null, Null, T("t"), T("t"), Null, Null, Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("statement_order", NUMERIC), Column("table_name", TEXT), Column("diff_type", TEXT), Column("statement", TEXT)],
                        rows: &[
                            &[T("1"), T("public.t1"), T("schema"), T(r#"CREATE TABLE "t1" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                            &[T("2"), T("public.t1"), T("data"), T(r#"INSERT INTO "t1" ("pk") VALUES (1);"#)],
                            &[T("3"), T("testschema.t2"), T("schema"), T(r#"CREATE TABLE "t2" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                            &[T("4"), T("testschema.t2"), T("data"), T(r#"INSERT INTO "t2" ("pk") VALUES (1);"#)],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("statement_order", NUMERIC), Column("table_name", TEXT), Column("diff_type", TEXT), Column("statement", TEXT)],
                        rows: &[
                            &[T("1"), T("public.t1"), T("schema"), T(r#"CREATE TABLE "t1" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                            &[T("2"), T("public.t1"), T("data"), T(r#"INSERT INTO "t1" ("pk") VALUES (1);"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't2')",
                    expected: Expected::Rows {
                        columns: &[Column("statement_order", NUMERIC), Column("table_name", TEXT), Column("diff_type", TEXT), Column("statement", TEXT)],
                        rows: &[
                            &[T("1"), T("testschema.t2"), T("schema"), T(r#"CREATE TABLE "t2" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                            &[T("2"), T("testschema.t2"), T("data"), T(r#"INSERT INTO "t2" ("pk") VALUES (1);"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[
                            &[T(""), T("public.t1"), T(""), T(r#"CREATE TABLE "t1" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                            &[T(""), T("testschema.t2"), T(""), T(r#"CREATE TABLE "t2" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[
                            &[T(""), T("public.t1"), T(""), T(r#"CREATE TABLE "t1" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't2')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[
                            &[T(""), T("testschema.t2"), T(""), T(r#"CREATE TABLE "t2" (
  "pk" integer NOT NULL,
  PRIMARY KEY ("pk")
);"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_query_diff('select * from t1 as of main', 'select * from t1')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: t1", ..E }),
                    flow: Flow::Query,
                    skip: Some("Go's DOLT_QUERY_DIFF cannot parse AS OF in its queries; the Rust server parses the queries with Doltgres' grammar, where an unquoted revision is a syntax error"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_query_diff('select * from t2 as of main', 'select * from t2')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: t2", ..E }),
                    flow: Flow::Query,
                    skip: Some("Go's DOLT_QUERY_DIFF cannot parse AS OF in its queries; the Rust server parses the queries with Doltgres' grammar, where an unquoted revision is a syntax error"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt_patch works with JSONB columns",
            set_up_script: &[
                "CREATE TABLE repro (pk int primary key, data jsonb);",
                r#"INSERT INTO repro VALUES (1, '{"text": "hello"}');"#,
            ],
            assertions: &[
                // Doltgres-specific: Go's output, with the string literal quoted as Postgres reads it instead of with MySQL's backslash escapes.
                ScriptTestAssertion {
                    query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 'repro')",
                    expected: Expected::Rows {
                        columns: &[Column("statement_order", NUMERIC), Column("table_name", TEXT), Column("diff_type", TEXT), Column("statement", TEXT)],
                        rows: &[
                            &[T("1"), T("public.repro"), T("schema"), T(r#"CREATE TABLE "repro" (
  "pk" integer NOT NULL,
  "data" jsonb,
  PRIMARY KEY ("pk")
);"#)],
                            &[T("2"), T("public.repro"), T("data"), T(r#"INSERT INTO "repro" ("pk","data") VALUES (1,'{"text": "hello"}');"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 'repro')",
                    expected: Expected::Rows {
                        columns: &[Column("diff_type", TEXT), Column("from_pk", INT4), Column("to_pk", INT4)],
                        rows: &[
                            &[T("added"), Null, T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DOLT_PREVIEW_MERGE_CONFLICTS basic functionality",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key, c1 int);",
                "INSERT INTO t1 VALUES (1, 10), (2, 20);",
                "SELECT DOLT_COMMIT('-Am', 'initial commit');",
                "SELECT DOLT_CHECKOUT('-b', 'branch1');",
                "UPDATE t1 SET c1 = 100 WHERE pk = 1;",
                "SELECT DOLT_COMMIT('-am', 'update on branch1');",
                "SELECT DOLT_CHECKOUT('main');",
                "UPDATE t1 SET c1 = 200 WHERE pk = 1;",
                "SELECT DOLT_COMMIT('-am', 'update on main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_data_conflicts", NUMERIC), Column("num_schema_conflicts", NUMERIC)],
                        rows: &[
                            &[T("public.t1"), T("1"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_pk, base_c1, our_pk, our_c1, our_diff_type, their_pk, their_c1, their_diff_type FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("base_pk", INT4), Column("base_c1", INT4), Column("our_pk", INT4), Column("our_c1", INT4), Column("our_diff_type", TEXT), Column("their_pk", INT4), Column("their_c1", INT4), Column("their_diff_type", TEXT)],
                        rows: &[
                            &[T("1"), T("10"), T("1"), T("200"), T("modified"), T("1"), T("100"), T("modified")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DOLT_PREVIEW_MERGE_CONFLICTS with no conflicts",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key, c1 int);",
                "INSERT INTO t1 VALUES (1, 10), (2, 20);",
                "SELECT DOLT_COMMIT('-Am', 'initial commit');",
                "SELECT DOLT_CHECKOUT('-b', 'branch1');",
                "INSERT INTO t1 VALUES (3, 30);",
                "SELECT DOLT_COMMIT('-am', 'insert on branch1');",
                "SELECT DOLT_CHECKOUT('main');",
                "INSERT INTO t1 VALUES (4, 40);",
                "SELECT DOLT_COMMIT('-am', 'insert on main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_data_conflicts", NUMERIC), Column("num_schema_conflicts", NUMERIC)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_pk, base_c1, our_pk, our_c1, our_diff_type, their_pk, their_c1, their_diff_type FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("base_pk", INT4), Column("base_c1", INT4), Column("our_pk", INT4), Column("our_c1", INT4), Column("our_diff_type", TEXT), Column("their_pk", INT4), Column("their_c1", INT4), Column("their_diff_type", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple tables",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key, c1 int);",
                "CREATE TABLE t2 (pk int primary key, c1 varchar(20));",
                "INSERT INTO t1 VALUES (1, 10);",
                "INSERT INTO t2 VALUES (1, 'initial');",
                "SELECT DOLT_COMMIT('-Am', 'initial commit');",
                "SELECT DOLT_CHECKOUT('-b', 'branch1');",
                "UPDATE t1 SET c1 = 100 WHERE pk = 1;",
                "UPDATE t2 SET c1 = 'branch1' WHERE pk = 1;",
                "SELECT DOLT_COMMIT('-am', 'updates on branch1');",
                "SELECT DOLT_CHECKOUT('main');",
                "UPDATE t1 SET c1 = 200 WHERE pk = 1;",
                "UPDATE t2 SET c1 = 'main' WHERE pk = 1;",
                "SELECT DOLT_COMMIT('-am', 'updates on main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_data_conflicts", NUMERIC), Column("num_schema_conflicts", NUMERIC)],
                        rows: &[
                            &[T("public.t1"), T("1"), T("0")],
                            &[T("public.t2"), T("1"), T("0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't2')",
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
        ScriptTest {
            name: "DOLT_PREVIEW_MERGE_CONFLICTS with schema conflicts",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key, c1 int);",
                "INSERT INTO t1 VALUES (1, 10);",
                "SELECT DOLT_COMMIT('-Am', 'initial commit');",
                "SELECT DOLT_CHECKOUT('-b', 'branch1');",
                "ALTER TABLE t1 ADD COLUMN c2 varchar(50);",
                "SELECT DOLT_COMMIT('-am', 'add column on branch1');",
                "SELECT DOLT_CHECKOUT('main');",
                "ALTER TABLE t1 ADD COLUMN c2 int;",
                "SELECT DOLT_COMMIT('-am', 'add same column different type on main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_data_conflicts", NUMERIC), Column("num_schema_conflicts", NUMERIC)],
                        rows: &[
                            &[T("public.t1"), Null, T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "schema conflicts found: 1", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple schemas",
            set_up_script: &[
                "CREATE SCHEMA test_schema;",
                "CREATE TABLE t1 (pk int primary key, c1 int);",
                "CREATE TABLE test_schema.t2 (pk int primary key, c1 int);",
                "INSERT INTO t1 VALUES (1, 10);",
                "INSERT INTO test_schema.t2 VALUES (1, 20);",
                "SELECT DOLT_COMMIT('-Am', 'initial commit');",
                "SELECT DOLT_CHECKOUT('-b', 'branch1');",
                "UPDATE t1 SET c1 = 100 WHERE pk = 1;",
                "UPDATE test_schema.t2 SET c1 = 200 WHERE pk = 1;",
                "SELECT DOLT_COMMIT('-am', 'updates on branch1');",
                "SELECT DOLT_CHECKOUT('main');",
                "UPDATE t1 SET c1 = 300 WHERE pk = 1;",
                "UPDATE test_schema.t2 SET c1 = 400 WHERE pk = 1;",
                "SELECT DOLT_COMMIT('-am', 'updates on main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_data_conflicts", NUMERIC), Column("num_schema_conflicts", NUMERIC)],
                        rows: &[
                            &[T("public.t1"), T("1"), T("0")],
                            &[T("test_schema.t2"), T("1"), T("0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't2')",
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
        ScriptTest {
            name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple schemas, same name",
            set_up_script: &[
                "CREATE SCHEMA test_schema;",
                "CREATE TABLE t1 (pk int primary key, c1 int);",
                "CREATE TABLE test_schema.t1 (pk int primary key, c2 int);",
                "INSERT INTO t1 VALUES (1, 10);",
                "INSERT INTO test_schema.t1 VALUES (1, 20);",
                "SELECT DOLT_COMMIT('-Am', 'initial commit');",
                "SELECT DOLT_CHECKOUT('-b', 'branch1');",
                "UPDATE t1 SET c1 = 100 WHERE pk = 1;",
                "UPDATE test_schema.t1 SET c2 = 200 WHERE pk = 1;",
                "SELECT DOLT_COMMIT('-am', 'updates on branch1');",
                "SELECT DOLT_CHECKOUT('main');",
                "UPDATE t1 SET c1 = 300 WHERE pk = 1;",
                "UPDATE test_schema.t1 SET c2 = 400 WHERE pk = 1;",
                "SELECT DOLT_COMMIT('-am', 'updates on main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_data_conflicts", NUMERIC), Column("num_schema_conflicts", NUMERIC)],
                        rows: &[
                            &[T("public.t1"), T("1"), T("0")],
                            &[T("test_schema.t1"), T("1"), T("0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_c1 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("base_c1", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_c2 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "base_c2" could not be found in any table in scope"#, ..E }),
                    skip: Some("Go reports the missing column in go-mysql-server's words; the Rust server reports it as Postgres does"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SET search_path TO test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_c2 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Rows {
                        columns: &[Column("base_c2", INT4)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_c1 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "base_c1" could not be found in any table in scope"#, ..E }),
                    skip: Some("Go reports the missing column in go-mysql-server's words; the Rust server reports it as Postgres does"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DOLT_PREVIEW_MERGE_CONFLICTS error cases",
            set_up_script: &[
                "CREATE TABLE t1 (pk int primary key, c1 int);",
                "SELECT DOLT_COMMIT('-Am', 'initial commit');",
                "SELECT DOLT_CHECKOUT('-b', 'branch1');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('nonexistent-branch', 'main')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "branch not found: nonexistent-branch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1', 'table')",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function 'dolt_preview_merge_conflicts_summary' expected 2 arguments, 3 received", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'nonexistent-branch')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "branch not found: nonexistent-branch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('', 'main')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "left branch name cannot be empty", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', '')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "right branch name cannot be empty", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY(NULL, 'main')",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "Invalid argument to dolt_preview_merge_conflicts_summary: NULL", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', NULL)",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "Invalid argument to dolt_preview_merge_conflicts_summary: NULL", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('nonexistent-branch', 'main', 't1')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "branch not found: nonexistent-branch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'nonexistent-branch', 't1')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "branch not found: nonexistent-branch", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1')",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function 'dolt_preview_merge_conflicts' expected 3 arguments, 2 received", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1', 'extra')",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function 'dolt_preview_merge_conflicts' expected 3 arguments, 4 received", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('', 'main', 't1')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "string is not a valid branch or hash", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', '', 't1')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "string is not a valid branch or hash", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS(NULL, 'main', 't1')",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "Invalid argument to dolt_preview_merge_conflicts: NULL", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', NULL, 't1')",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "Invalid argument to dolt_preview_merge_conflicts: NULL", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', NULL)",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "Invalid argument to dolt_preview_merge_conflicts: NULL", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 'nonexistent_table')",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: public.nonexistent_table", ..E }),
                    skip: Some("Go reports the missing table in MySQL's words; the Rust server reports the missing relation as Postgres does"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', '')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "table name cannot be empty", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_gc() {
    run_scripts(&[
        ScriptTest {
            name: "Full",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_GC();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_gc", INT8)],
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
        ScriptTest {
            name: "Shallow",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_GC('--shallow');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_gc", INT8)],
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
fn test_dolt_log() {
    run_scripts(&[
        ScriptTest {
            name: "Smoke test",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_LOG('main');",
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
                    query: "SELECT COUNT(*) FROM DOLT_LOG('original');",
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
fn test_dolt_merge() {
    run_scripts(&[
        ScriptTest {
            name: "Merge without conflicts",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
                "CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "INSERT INTO t_composite VALUES (1, 1, 1);",
                "INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
                "INSERT INTO t_serial VALUES (DEFAULT, 1);",
                "INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
                "INSERT INTO t_trigger VALUES (1, 1);",
                "INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "INSERT INTO t_composite VALUES (2, 2, 2);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 2);",
                "INSERT INTO t_serial VALUES (DEFAULT, 2);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (3, 3);",
                "INSERT INTO t_composite VALUES (3, 2, 3);",
                "INSERT INTO t_array VALUES (ARRAY['dfe'], 3);",
                "INSERT INTO t_serial VALUES (4, 3);",
                "INSERT INTO t_generated (pk, v1) VALUES (3, 3);",
                "INSERT INTO t_trigger VALUES (3, 3);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT4 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (3, 3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT strpos(DOLT_MERGE('main')::text, '0,0,"merge successful"') > 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("strpos > 1", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            skip: Some("The setup changes f_default's return type with CREATE OR REPLACE FUNCTION, which Go allows and Postgres refuses with 42P13"),
            ..S
        },
        ScriptTest {
            name: "Merge with conflicts",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
                "CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "INSERT INTO t_composite VALUES (1, 1, 1);",
                "INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
                "INSERT INTO t_serial VALUES (DEFAULT, 1);",
                "INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
                "INSERT INTO t_trigger VALUES (1, 1);",
                "INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "INSERT INTO t_composite VALUES (2, 2, 2);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 2);",
                "INSERT INTO t_serial VALUES (DEFAULT, 2);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "INSERT INTO t_composite VALUES (2, 2, 3);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 3);",
                "INSERT INTO t_serial VALUES (DEFAULT, 3);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT DOLT_MERGE('main');",
                    flow: Flow::Query,
                    skip: Some("the Go server panics, with a stack trace that differs between runs"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Fast forward",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('original');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: r#"SELECT strpos(DOLT_MERGE('main')::text, '1,0,"merge successful"') > 1;"#,
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
            name: "--no-ff",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('original');",
            ],
            assertions: &[
                // Doltgres-specific: Go's output, with the column named ?column? as Postgres names it rather than after go-mysql-server's expression text.
                ScriptTestAssertion {
                    query: "SELECT strpos(DOLT_MERGE('main', '--no-ff', '-m', 'merge_commit')::text, 'merge successful') > 1;",
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
            name: "Merge with nested parentheses in default, generated, and check expressions",
            set_up_script: &[
                "CREATE TABLE t3324 (pk INT4 PRIMARY KEY, a INT4 DEFAULT (1 + 1) * 2, b INT4 GENERATED ALWAYS AS ((pk + 1) * 2) STORED, CONSTRAINT c3324 CHECK (NOT (pk = 0 OR pk + 1 = 0) AND ((pk + 1) * 2) > 3));",
                "INSERT INTO t3324 (pk) VALUES (1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t3324 (pk) VALUES (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'main')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t3324 (pk) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'other')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Go's output, with the column named ?column? as Postgres names it rather than after go-mysql-server's expression text.
                ScriptTestAssertion {
                    query: "SELECT strpos(DOLT_MERGE('main', '--no-ff', '-m', 'merge_commit')::text, 'merge successful') > 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t3324 ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("4"), T("4")],
                            &[T("2"), T("4"), T("6")],
                            &[T("3"), T("4"), T("8")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO t3324 (pk) VALUES (0);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"Check constraint "c3324" violated"#, ..E }),
                    skip: Some("Go reports the violated check constraint in go-mysql-server's words; the Rust server reports it as Postgres does"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_preview_merge_conflicts() {
    run_scripts(&[
        ScriptTest {
            name: "Preview conflicts",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
                "CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "INSERT INTO t_composite VALUES (1, 1, 1);",
                "INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
                "INSERT INTO t_serial VALUES (DEFAULT, 1);",
                "INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
                "INSERT INTO t_trigger VALUES (1, 1);",
                "INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "INSERT INTO t_composite VALUES (2, 2, 2);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 2);",
                "INSERT INTO t_serial VALUES (2, 2);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "INSERT INTO t_composite VALUES (2, 2, 3);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 3);",
                "INSERT INTO t_serial VALUES (2, 3);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_simple');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_composite');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_array');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_serial');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_generated');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_trigger');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_default_func');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 'f_default()');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: public.f_default()", ..E }),
                    flow: Flow::Query,
                    skip: Some("Go reports a root object as a missing table in MySQL's words; the Rust server reports the missing relation as Postgres does"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 'f_trigger()');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: public.f_trigger()", ..E }),
                    flow: Flow::Query,
                    skip: Some("Go reports a root object as a missing table in MySQL's words; the Rust server reports the missing relation as Postgres does"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_serial_pk_seq');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: public.t_serial_pk_seq", ..E }),
                    flow: Flow::Query,
                    skip: Some("Go reports a root object as a missing table in MySQL's words; the Rust server reports the missing relation as Postgres does"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_preview_merge_conflicts_summary() {
    run_scripts(&[
        ScriptTest {
            name: "Preview summary",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, v1 INT4, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY, v1 INT4);",
                "CREATE TABLE t_generated (pk INT8 PRIMARY KEY, v1 INT4, v2 INT8 GENERATED ALWAYS AS (pk * 1000) STORED);",
                "CREATE TABLE t_trigger (pk INT4 PRIMARY KEY, v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (pk INT4 PRIMARY KEY, v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1, 1);",
                "INSERT INTO t_composite VALUES (1, 1, 1);",
                "INSERT INTO t_array VALUES (ARRAY['abc'], 1);",
                "INSERT INTO t_serial VALUES (DEFAULT, 1);",
                "INSERT INTO t_generated (pk, v1) VALUES (1, 1);",
                "INSERT INTO t_trigger VALUES (1, 1);",
                "INSERT INTO t_default_func (pk, v2) VALUES (1, 1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (2, 2);",
                "INSERT INTO t_composite VALUES (2, 2, 2);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 2);",
                "INSERT INTO t_serial VALUES (DEFAULT, 2);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (2, 3);",
                "INSERT INTO t_composite VALUES (2, 2, 3);",
                "INSERT INTO t_array VALUES (ARRAY['def'], 3);",
                "INSERT INTO t_serial VALUES (DEFAULT, 3);",
                "INSERT INTO t_generated (pk, v1) VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 34; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger VALUES (2, 3);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 35; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (pk, v2) VALUES (2, 3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'other');",
                    flow: Flow::Query,
                    skip: Some("the Go server panics, with a stack trace that differs between runs"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_query_diff() {
    run_scripts(&[
        ScriptTest {
            name: "Smoke test",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY, v1 INT4);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('other');",
                "INSERT INTO t_simple VALUES (1, 1), (2, 1);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'next')::text) = 32;",
                "SELECT DOLT_CHECKOUT('other');",
                "INSERT INTO t_simple VALUES (1, 2), (2, 2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'next')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_QUERY_DIFF('SELECT * FROM t_simple AS OF main', 'SELECT * FROM t_simple AS OF other');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "TargetType not handled: `DB_TABLE_IDENT`", ..E }),
                    flow: Flow::Query,
                    skip: Some("Go's DOLT_QUERY_DIFF cannot parse AS OF in its queries; the Rust server parses the queries with Doltgres' grammar, where an unquoted revision is a syntax error"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_rm() {
    run_scripts(&[
        ScriptTest {
            name: "Cached only",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT DOLT_ADD('-A');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_RM('--cached', 't_simple','t_composite','t_array','t_serial','t_default_simple','t_checked','t_fk_parent','t_fk_child','t_unique','t_generated','t_trigger','t_default_func','f_trigger()','f_default()','t_serial_pk_seq','t_trigger.trig_trigger');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_rm", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
fn test_dolt_reset() {
    run_scripts(&[
        ScriptTest {
            name: "Hard",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_RESET('--hard', 'HEAD~1');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    skip: Some("Go skips this assertion until root objects are supported; untracked tables survive reset --hard, as in the Go server"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Soft",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT DOLT_ADD('-A');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_RESET('--soft', 'HEAD');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';",
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
                    query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
fn test_dolt_revert() {
    run_scripts(&[
        ScriptTest {
            name: "Revert a single commit",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT f_default();",
                    expected: Expected::Rows {
                        columns: &[Column("f_default", INT8)],
                        rows: &[
                            &[T("34")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT substring(DOLT_REVERT('HEAD')::text, 34);",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T(",0,0,0)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM t_simple;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT f_default();",
                    expected: Expected::Rows {
                        columns: &[Column("f_default", INT8)],
                        rows: &[
                            &[T("33")],
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
fn test_dolt_schema_diff() {
    run_scripts(&[
        ScriptTest {
            name: "Single commit",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "SELECT DOLT_BRANCH('original');",
                "INSERT INTO t_simple VALUES (4);",
                "INSERT INTO t_composite VALUES (3, 100);",
                "INSERT INTO t_array VALUES (ARRAY['stu']);",
                "INSERT INTO t_serial VALUES (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (98, 8);",
                "INSERT INTO t_checked VALUES (99);",
                "INSERT INTO t_fk_parent VALUES (40);",
                "INSERT INTO t_fk_child VALUES (30);",
                "INSERT INTO t_unique VALUES (10);",
                "INSERT INTO t_generated (v1) VALUES (11);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (2);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (3);",
                "ALTER TABLE t_simple RENAME COLUMN pk TO rcol;",
                "ALTER TABLE t_composite RENAME COLUMN pk1 TO rcol;",
                "ALTER TABLE t_array RENAME COLUMN pk TO rcol;",
                "ALTER TABLE t_serial RENAME COLUMN pk TO rcol;",
                "ALTER TABLE t_default_simple RENAME COLUMN v1 TO rcol;",
                "ALTER TABLE t_generated RENAME COLUMN v1 TO rcol;",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_SCHEMA_DIFF('main', 'original');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
fn test_dolt_stash() {
    run_scripts(&[
        ScriptTest {
            name: "Push and pop all untracked",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_STASH('push', 'dgstash', '--all');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stash", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
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
                    query: "SELECT DOLT_STASH('pop', 'dgstash');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stash", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
        ScriptTest {
            name: "Push and pop tracked",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
                "CREATE TABLE t_array (pk TEXT[] PRIMARY KEY);",
                "CREATE TABLE t_serial (pk SERIAL PRIMARY KEY);",
                "CREATE TABLE t_default_simple (v1 INT8 DEFAULT 22, v2 INT8);",
                "CREATE TABLE t_checked (v1 NUMERIC CHECK (v1 > 0 AND v1 <= 100));",
                "CREATE TABLE t_fk_parent (pk INT4 PRIMARY KEY);",
                "CREATE TABLE t_fk_child (pk INT4 REFERENCES t_fk_parent(pk), PRIMARY KEY(pk));",
                "CREATE TABLE t_unique (v1 INT4 UNIQUE);",
                "CREATE TABLE t_generated (v1 INT8, v2 INT8 GENERATED ALWAYS AS (v1 * 1000) STORED);",
                "CREATE TABLE t_trigger (v1 INT8);",
                "CREATE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 3; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_trigger BEFORE INSERT OR UPDATE ON t_trigger FOR EACH ROW EXECUTE FUNCTION f_trigger();",
                "CREATE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 33; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t_default_func (v1 INT8 DEFAULT f_default(), v2 INT8);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
                "INSERT INTO t_array VALUES (ARRAY['abc']), (ARRAY['def','ghi']), (ARRAY['jkl','mno','pqr']);",
                "INSERT INTO t_serial VALUES (DEFAULT), (DEFAULT), (DEFAULT);",
                "INSERT INTO t_default_simple VALUES (DEFAULT, 5), (99, 6), (DEFAULT, 7);",
                "INSERT INTO t_checked VALUES (1), (50), (100);",
                "INSERT INTO t_fk_parent VALUES (10), (20), (30);",
                "INSERT INTO t_fk_child VALUES (10), (20);",
                "INSERT INTO t_unique VALUES (7), (8), (9);",
                "INSERT INTO t_generated (v1) VALUES (1), (2), (10);",
                "CREATE OR REPLACE FUNCTION f_trigger() RETURNS TRIGGER AS $$ BEGIN NEW.v1 := NEW.v1 * 33; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_trigger (v1) VALUES (5), (10), (0);",
                "CREATE OR REPLACE FUNCTION f_default() RETURNS INT8 AS $$ BEGIN RETURN 34; END; $$ LANGUAGE plpgsql;",
                "INSERT INTO t_default_func (v2) VALUES (1), (2);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("15")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_STASH('push', 'dgstash');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_stash", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
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
                    query: "SELECT DOLT_STASH('pop', 'dgstash');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"error: Your local changes to the following tables would be overwritten by applying stash 0:
"#, message_contains: true, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM dolt_status;",
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
fn test_dolt_tag() {
    run_scripts(&[
        ScriptTest {
            name: "Smoke test",
            set_up_script: &[
                "CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
                "INSERT INTO t_simple VALUES (1), (2), (3);",
                "SELECT length(DOLT_COMMIT('-A', '-m', 'initial')::text) = 32;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_TAG('tagged_commit', 'HEAD');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_tag", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_CHECKOUT('tagged_commit');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"dolt does not support a detached head state. To create a branch at this tag, run: 
	CALL DOLT_CHECKOUT('tagged_commit', '-b', <new_branch_name>)"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_verify_constraints() {
    run_scripts(&[
        ScriptTest {
            name: "Foreign key violations from a merge",
            set_up_script: &[
                "CREATE TABLE p (id INT PRIMARY KEY);",
                "CREATE TABLE c (id INT PRIMARY KEY, pid INT REFERENCES p(id));",
                "INSERT INTO p VALUES (1), (2);",
                "SELECT DOLT_COMMIT('-Am', 'init');",
                "SELECT DOLT_CHECKOUT('-b', 'other');",
                "DELETE FROM p WHERE id = 2;",
                "SELECT DOLT_COMMIT('-am', 'del');",
                "SELECT DOLT_CHECKOUT('main');",
                "INSERT INTO c VALUES (1, 2);",
                "SELECT DOLT_COMMIT('-am', 'child');",
                "SET dolt_force_transaction_commit = on;",
                "SELECT DOLT_MERGE('other');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_VERIFY_CONSTRAINTS('p');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_VERIFY_CONSTRAINTS('c');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: the Go server returns 0 because it checks no tables without table names.
                ScriptTestAssertion {
                    query: "SELECT DOLT_VERIFY_CONSTRAINTS();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "DELETE FROM dolt_constraint_violations_c;",
                    expected: Expected::Rows {
                        columns: &[],
                        rows: &[],
                        tag: "DELETE 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_VERIFY_CONSTRAINTS('--all', 'c');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_constraint_violations;",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_violations", NUMERIC)],
                        rows: &[
                            &[T("c"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT DOLT_VERIFY_CONSTRAINTS('nope');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: nope", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
