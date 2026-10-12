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
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Merge conflict detected, @autocommit transaction rolled back. @autocommit must be disabled so that merge conflicts can be resolved using the dolt_conflicts and dolt_schema_conflicts tables before manually committing the transaction. Alternatively, to commit transactions with merge conflicts, set @@dolt_allow_commit_conflicts = 1", ..E }),
                    flow: Flow::Query,
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
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Merge conflict detected, @autocommit transaction rolled back. @autocommit must be disabled so that merge conflicts can be resolved using the dolt_conflicts and dolt_schema_conflicts tables before manually committing the transaction. Alternatively, to commit transactions with merge conflicts, set @@dolt_allow_commit_conflicts = 1", ..E }),
                    flow: Flow::Query,
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
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Merge conflict detected, @autocommit transaction rolled back. @autocommit must be disabled so that merge conflicts can be resolved using the dolt_conflicts and dolt_schema_conflicts tables before manually committing the transaction. Alternatively, to commit transactions with merge conflicts, set @@dolt_allow_commit_conflicts = 1", ..E }),
                    flow: Flow::Query,
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
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_query_diff('select * from t1 as of main', 'select * from t1')",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "main""#, position: 24, ..E }),
                    flow: Flow::Query,
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
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_query_diff('select * from t1 as of main', 'select * from t1')",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "main""#, position: 24, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_query_diff('select * from t2 as of main', 'select * from t2')",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "main""#, position: 24, ..E }),
                    flow: Flow::Query,
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
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_c2 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "base_c2" does not exist"#, position: 8, ..E }),
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
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT base_c1 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "base_c1" does not exist"#, position: 8, ..E }),
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
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Merge conflict detected, @autocommit transaction rolled back. @autocommit must be disabled so that merge conflicts can be resolved using the dolt_conflicts and dolt_schema_conflicts tables before manually committing the transaction. Alternatively, to commit transactions with merge conflicts, set @@dolt_allow_commit_conflicts = 1", ..E }),
                    flow: Flow::Query,
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
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 'f_default()');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "f_default()" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 'f_trigger()');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "f_trigger()" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'other', 't_serial_pk_seq');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "t_serial_pk_seq" does not exist"#, ..E }),
                    flow: Flow::Query,
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
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    flow: Flow::Query,
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
                // Doltgres-specific: the Go server's output was wrong here, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM DOLT_QUERY_DIFF('SELECT * FROM t_simple AS OF main', 'SELECT * FROM t_simple AS OF other');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "main""#, position: 30, ..E }),
                    flow: Flow::Query,
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

#[test]
fn test_writes_after_dolt_gc() {
    run_scripts(&[
        ScriptTest {
            name: "writes after DOLT_GC reference chunks in the old generation",
            set_up_script: &[
                "CREATE TABLE t (a INT4 PRIMARY KEY);",
                "INSERT INTO t VALUES (1);",
                "SELECT DOLT_COMMIT('-A', '-m', 'first');",
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
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t2 (b INT4);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
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
fn test_create_index_after_dolt_gc() {
    run_scripts(&[
        ScriptTest {
            name: "CREATE INDEX after DOLT_GC removed the empty map",
            set_up_script: &[
                "CREATE TABLE gi (id INT4 PRIMARY KEY, v INT4);",
                "INSERT INTO gi SELECT i, i % 7 FROM generate_series(1, 50) i;",
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
                ScriptTestAssertion {
                    query: "CREATE INDEX gi_v ON gi (v);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO gi VALUES (51, 3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM gi WHERE v = 3 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("10")],
                            &[T("17")],
                            &[T("24")],
                            &[T("31")],
                            &[T("38")],
                            &[T("45")],
                            &[T("51")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_commit_author_forms() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_commit authors read as Dolt reads them",
            set_up_script: &[
                "CREATE TABLE au (id INT PRIMARY KEY);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the expected output is Dolt's behavior.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-Am', 'first', '--author', 'Alice Author <alice@example.com')::text) > 0;",
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
                    query: "SELECT committer, email, message FROM dolt_log LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("committer", TEXT), Column("email", TEXT), Column("message", TEXT)],
                        rows: &[
                            &[T("Alice Author"), T("alice@example.com"), T("first")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO au VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'second', '--author', 'Alice Author <>');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Aborting commit due to empty author email. Is your config set?", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'third', '--author', 'Alice');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Author not formatted correctly. Use 'Name <author@example.com>' format", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_table_renames_and_checkouts() {
    run_scripts(&[
        ScriptTest {
            name: "renamed tables in the status and staging",
            set_up_script: &[
                "CREATE TABLE r (id INT PRIMARY KEY, v INT);",
                "INSERT INTO r VALUES (1, 1);",
                "SELECT dolt_commit('-Am', 'base');",
                "ALTER TABLE r RENAME TO r2;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.r -> public.r2"), T("f"), T("renamed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('r');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.r -> public.r2"), T("f"), T("renamed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('r2');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.r -> public.r2"), T("t"), T("renamed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dropping the last view drops dolt_schemas",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY);",
                "CREATE VIEW vv AS SELECT id FROM t;",
                "SELECT dolt_commit('-Am', 'base');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP VIEW vv;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.dolt_schemas"), T("f"), T("deleted")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "checking out every table of a commit",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, v INT);",
                "INSERT INTO t VALUES (1, 1);",
                "SELECT dolt_commit('-Am', 'base');",
                "UPDATE t SET v = 5;",
                "SELECT dolt_add('t');",
                "UPDATE t SET v = 8;",
                "CREATE TABLE extra (id INT PRIMARY KEY);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_checkout('HEAD', '.')::text) > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.extra"), T("f"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('HEAD', 'missing');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "tablespec 'missing' did not match any table(s) known to dolt", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "blame ignores temporary tables",
            set_up_script: &[
                "CREATE TABLE b (id INT PRIMARY KEY, v TEXT);",
                "INSERT INTO b VALUES (1, 'a');",
                "SELECT dolt_commit('-Am', 'INIT');",
                "UPDATE b SET v = 'b' WHERE id = 1;",
                "SELECT dolt_commit('-Am', 'UPDATE');",
                "CREATE TEMP TABLE b (x TEXT PRIMARY KEY, y INT);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT id, message FROM dolt_blame_b;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("message", TEXT)],
                        rows: &[
                            &[T("1"), T("UPDATE")],
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
fn test_dolt_hashof_db_and_table() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_hashof_db and dolt_hashof_table",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY);",
                "SELECT dolt_commit('-Am', 'c');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_hashof_db() = dolt_hashof_db('HEAD'), dolt_hashof_db('main') = dolt_hashof_db('HEAD'), dolt_hashof_db('STAGED') = dolt_hashof_db();",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_hashof_table('t'));",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_hashof_db() = dolt_hashof_db('HEAD'), dolt_hashof_db('WORKING') = dolt_hashof_db();",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_hashof_table('nope');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "table not found: nope", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_schema_conflicts_roll_back_autocommit() {
    run_scripts(&[
        ScriptTest {
            name: "a merge with schema conflicts rolls back under autocommit",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY);",
                "INSERT INTO t VALUES (1);",
                "SELECT dolt_commit('-Am', 'ancestor');",
                "SELECT dolt_branch('feat');",
                "SELECT dolt_checkout('feat');",
                "ALTER TABLE t ADD COLUMN v TEXT;",
                "SELECT dolt_commit('-Am', 'feat_add_text');",
                "SELECT dolt_checkout('main');",
                "ALTER TABLE t ADD COLUMN v INTEGER;",
                "SELECT dolt_commit('-Am', 'main_add_int');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_merge('feat');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Merge conflict detected, @autocommit transaction rolled back. @autocommit must be disabled so that merge conflicts can be resolved using the dolt_conflicts and dolt_schema_conflicts tables before manually committing the transaction. Alternatively, to commit transactions with merge conflicts, set @@dolt_allow_commit_conflicts = 1", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT is_merging FROM dolt_merge_status;",
                    expected: Expected::Rows {
                        columns: &[Column("is_merging", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_status;",
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
fn test_dolt_procedure_null_arguments() {
    run_scripts(&[
        ScriptTest {
            name: "Dolt's procedures read NULL arguments as empty strings",
            set_up_script: &[
                "CREATE TABLE u (id INT PRIMARY KEY);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_clean(NULL);",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "failed to clean; table not found: ''", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_dolt_commit_validations() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_commit refuses what Dolt refuses",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, v INT);",
                "INSERT INTO t VALUES (1, 1);",
                "SELECT dolt_commit('-Am', 'base');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-m', '');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Aborting commit due to empty commit message.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t RENAME TO t2;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'nothing stageable');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "nothing to commit", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t -> public.t2"), T("f"), T("renamed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE parents (id INT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE children (id INT PRIMARY KEY, pid INT REFERENCES parents (id));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('children');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-m', 'child only');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "foreign key `children_pid_fkey` requires the referenced table `parents`", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-f', '-m', 'child only forced')) > 0;",
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
fn test_dolt_reset_tables() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_reset with several tables resets them all, and table names fold to lower case",
            set_up_script: &[
                "CREATE TABLE a (id INT PRIMARY KEY);",
                "CREATE TABLE b (id INT PRIMARY KEY);",
                "CREATE TABLE Camel (id INT PRIMARY KEY);",
                "SELECT dolt_commit('-Am', 'c1');",
                "INSERT INTO a VALUES (1);",
                "INSERT INTO b VALUES (1);",
                "INSERT INTO camel VALUES (1);",
                "SELECT dolt_add('-A');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_reset('a', 'b');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged FROM dolt_status ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL)],
                        rows: &[
                            &[T("public.a"), T("f")],
                            &[T("public.b"), T("f")],
                            &[T("public.camel"), T("f")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('Camel');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged FROM dolt_status ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL)],
                        rows: &[
                            &[T("public.a"), T("f")],
                            &[T("public.b"), T("f")],
                            &[T("public.camel"), T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_reset('CAMEL');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged FROM dolt_status ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL)],
                        rows: &[
                            &[T("public.a"), T("f")],
                            &[T("public.b"), T("f")],
                            &[T("public.camel"), T("f")],
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
fn test_dolt_rebase_columns() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_rebase names its result columns",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_checkout('-b', 'feat');",
                "INSERT INTO t VALUES (1);",
                "SELECT dolt_commit('-am', 'f1');",
                "SELECT dolt_checkout('main');",
                "INSERT INTO t VALUES (2);",
                "SELECT dolt_commit('-am', 'm1');",
                "SELECT dolt_checkout('feat');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_rebase('main');",
                    expected: Expected::Rows {
                        columns: &[Column("status", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("0"), T("Successfully rebased and updated refs/heads/feat")],
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
fn test_status_leaves_out_ignored_tables() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_status leaves out new tables that dolt_ignore ignores",
            set_up_script: &[
                "INSERT INTO dolt_ignore VALUES ('secret', true);",
                "CREATE TABLE secret (x INT PRIMARY KEY);",
                "CREATE TABLE kept (x INT PRIMARY KEY);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.dolt_ignore"), T("f"), T("new table")],
                            &[T("public.kept"), T("f"), T("new table")],
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
fn test_recreated_table_keeps_tags() {
    run_scripts(&[
        ScriptTest {
            name: "a table dropped and created again keeps its columns' tags, so its changes merge",
            set_up_script: &[
                "CREATE TABLE t (id INTEGER PRIMARY KEY, v INT, w INT, c INT);",
                "INSERT INTO t VALUES (1, 1, 1, 1);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_checkout('-b', 'feat');",
                "DROP TABLE t;",
                "CREATE TABLE t (id INTEGER PRIMARY KEY, v INT, w INT, c INT, CONSTRAINT ck CHECK (v > 0));",
                "INSERT INTO t VALUES (1, 1, 1, 1);",
                "SELECT dolt_commit('-Am', 'check');",
                "SELECT dolt_checkout('main');",
                "ALTER TABLE t DROP COLUMN c;",
                "SELECT dolt_commit('-Am', 'drop');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('feat');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("0"), T("merge successful")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT count(*) FROM information_schema.columns WHERE table_name = 't'), (SELECT count(*) FROM information_schema.check_constraints WHERE constraint_name = 'ck'), id, v, w FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8), Column("id", INT4), Column("v", INT4), Column("w", INT4)],
                        rows: &[
                            &[T("3"), T("1"), T("1"), T("1"), T("1")],
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
fn test_workspace_table_changes() {
    run_scripts(&[
        ScriptTest {
            name: "UPDATE of a workspace table stages and unstages rows, and DELETE discards working changes",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, v VARCHAR(20) UNIQUE, w INT);",
                "INSERT INTO t VALUES (1, 'a', 1), (2, 'b', 1), (3, 'c', 1);",
                "SELECT dolt_commit('-Am', 'base');",
                "UPDATE t SET v = 'tmp' WHERE id = 1;",
                "UPDATE t SET v = 'a' WHERE id = 2;",
                "UPDATE t SET v = 'b' WHERE id = 1;",
                "UPDATE t SET w = 5 WHERE id = 3;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "UPDATE dolt_workspace_t SET staged = true WHERE to_id <= 2;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT id, staged, diff_type, to_id, to_v, from_id, from_v FROM dolt_workspace_t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("staged", BOOL), Column("diff_type", TEXT), Column("to_id", INT4), Column("to_v", VARCHAR), Column("from_id", INT4), Column("from_v", VARCHAR)],
                        rows: &[
                            &[T("0"), T("t"), T("modified"), T("1"), T("b"), T("1"), T("a")],
                            &[T("1"), T("t"), T("modified"), T("2"), T("a"), T("2"), T("b")],
                            &[T("2"), T("f"), T("modified"), T("3"), T("c"), T("3"), T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "UPDATE dolt_workspace_t SET staged = 1 WHERE to_id = 3;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "column \"staged\" is of type boolean but expression is of type integer", hint: "You will need to rewrite or cast the expression.", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "UPDATE dolt_workspace_t SET to_w = 9 WHERE to_id = 3;",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "only update of column 'staged' is allowed", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "DELETE FROM dolt_workspace_t WHERE staged;",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "cannot delete staged rows from workspace", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "DELETE FROM dolt_workspace_t WHERE to_id = 3;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", VARCHAR), Column("w", INT4)],
                        rows: &[
                            &[T("1"), T("b"), T("1")],
                            &[T("2"), T("a"), T("1")],
                            &[T("3"), T("c"), T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "UPDATE dolt_workspace_t SET staged = false WHERE to_id = 1;",
                    expected: Expected::Error(Diagnostic { code: "23505", message: "duplicate key value violates unique constraint \"t_v_key\"", detail: "Key (v)=(a) already exists.", schema: "public", table: "t", constraint: "t_v_key", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t"), T("t"), T("modified")],
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
fn test_merge_computes_generated_columns() {
    run_scripts(&[
        ScriptTest {
            name: "a merge computes stored generated columns from the merged values",
            set_up_script: &[
                "CREATE TABLE g (id INT PRIMARY KEY, x INT, y INT, z INT GENERATED ALWAYS AS (x + y) STORED);",
                "INSERT INTO g (id, x, y) VALUES (1, 1, 1), (2, 2, NULL);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_branch('feature');",
                "UPDATE g SET x = 2 WHERE id = 1;",
                "ALTER TABLE g ADD COLUMN d INT GENERATED ALWAYS AS (x * 10) STORED;",
                "SELECT dolt_commit('-Am', 'left');",
                "SELECT dolt_checkout('feature');",
                "UPDATE g SET y = 3 WHERE id = 1;",
                "UPDATE g SET y = 5 WHERE id = 2;",
                "INSERT INTO g (id, x, y) VALUES (3, 7, 8);",
                "SELECT dolt_commit('-Am', 'right');",
                "SELECT dolt_checkout('main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('feature');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("0"), T("merge successful")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM g ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("x", INT4), Column("y", INT4), Column("z", INT4), Column("d", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3"), T("5"), T("20")],
                            &[T("2"), T("2"), T("5"), T("7"), T("20")],
                            &[T("3"), T("7"), T("8"), T("15"), T("70")],
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
fn test_dolt_tests_and_test_run() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_tests holds test queries that DOLT_TEST_RUN runs and checks",
            set_up_script: &[
                "CREATE TABLE fixture (i INT PRIMARY KEY, s TEXT, d DATE, n NUMERIC(10,2), b BOOL);",
                "INSERT INTO fixture VALUES (1, 'one', '2025-08-22', 10.4, true), (2, 'two', NULL, NULL, NULL);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_tests;",
                    expected: Expected::Rows {
                        columns: &[Column("test_name", TEXT), Column("test_group", TEXT), Column("test_query", TEXT), Column("assertion_type", TEXT), Column("assertion_comparator", TEXT), Column("assertion_value", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_test_run();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "could not find tests for argument: *", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "INSERT INTO dolt_tests VALUES ('bad', NULL, 'SELECT 1', 'row_count', '==', '1');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: "new row for relation \"dolt_tests\" violates check constraint \"assertion_type_check\"", detail: "Failing row contains (bad, null, SELECT 1, row_count, ==, 1).", schema: "public", table: "dolt_tests", constraint: "assertion_type_check", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: r#"INSERT INTO dolt_tests VALUES
						('a_rows', 'g', 'SELECT i FROM fixture', 'expected_rows', '==', '3'),
						('b_cols', 'g', 'SELECT i, s FROM fixture LIMIT 1', 'expected_columns', '==', '2'),
						('c_date', 'types', 'SELECT d FROM fixture WHERE i = 1', 'expected_single_value', '>', '2025-08-23'),
						('d_numeric', 'types', 'SELECT n FROM fixture WHERE i = 1', 'expected_single_value', '>', '10.5'),
						('e_bool', 'types', 'SELECT b FROM fixture WHERE i = 1', 'expected_single_value', '==', 'true'),
						('f_null', NULL, 'SELECT NULL', 'expected_single_value', '!=', NULL),
						('g_rows', 'shape', 'SELECT i FROM fixture', 'expected_single_value', '==', '1'),
						('h_two', 'checks', 'SELECT 1; SELECT 2', 'expected_rows', '==', '1'),
						('i_write', 'checks', 'CREATE TABLE nope (x INT)', 'expected_rows', '==', '1'),
						('j_missing', 'checks', 'SELECT * FROM absent', 'expected_rows', '==', '0');"#,
                    expected: Expected::Tag("INSERT 0 10"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT test_name, test_group_name, status, message FROM dolt_test_run();",
                    expected: Expected::Rows {
                        columns: &[Column("test_name", TEXT), Column("test_group_name", TEXT), Column("status", TEXT), Column("message", TEXT)],
                        rows: &[
                            &[T("a_rows"), T("g"), T("FAIL"), T("Assertion failed: expected_rows equal to 3, got 2")],
                            &[T("b_cols"), T("g"), T("PASS"), T("")],
                            &[T("c_date"), T("types"), T("FAIL"), T("Assertion failed: expected_single_value greater than 2025-08-23, got 2025-08-22")],
                            &[T("d_numeric"), T("types"), T("FAIL"), T("Assertion failed: expected_single_value greater than 10.5, got 10.40")],
                            &[T("e_bool"), T("types"), T("PASS"), T("")],
                            &[T("f_null"), T(""), T("FAIL"), T("Assertion failed: expected_single_value not equal to NULL, got NULL")],
                            &[T("g_rows"), T("shape"), T("FAIL"), T("expected_single_value expects exactly one cell. Received multiple rows")],
                            &[T("h_two"), T("checks"), T("FAIL"), T("Can only run exactly one query")],
                            &[T("i_write"), T("checks"), T("FAIL"), T("Cannot execute write queries")],
                            &[T("j_missing"), T("checks"), T("FAIL"), T("query error: relation \"absent\" does not exist")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT test_name, status FROM dolt_test_run('types', 'b_cols');",
                    expected: Expected::Rows {
                        columns: &[Column("test_name", TEXT), Column("status", TEXT)],
                        rows: &[
                            &[T("c_date"), T("FAIL")],
                            &[T("d_numeric"), T("FAIL")],
                            &[T("e_bool"), T("PASS")],
                            &[T("b_cols"), T("PASS")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_test_run('nothing');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "could not find tests for argument: nothing", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_schema_merge_dropped_columns() {
    run_scripts(&[
        ScriptTest {
            name: "a merge drops the other side's index of a dropped column and conflicts on its edits of one",
            set_up_script: &[
                "CREATE TABLE t (k INT PRIMARY KEY, a VARCHAR(9), b VARCHAR(9), n INT);",
                "INSERT INTO t VALUES (1, 'a1', 'b1', 11), (2, 'a2', 'b2', 22);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_branch('feat');",
                "ALTER TABLE t DROP COLUMN a;",
                "SELECT dolt_commit('-Am', 'ours');",
                "SELECT dolt_checkout('feat');",
                "CREATE INDEX ia ON t (a);",
                "UPDATE t SET b = 'edit' WHERE k = 1;",
                "SELECT dolt_commit('-Am', 'theirs');",
                "SELECT dolt_checkout('main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('feat');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("0"), T("merge successful")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("b", VARCHAR), Column("n", INT4)],
                        rows: &[
                            &[T("1"), T("edit"), T("11")],
                            &[T("2"), T("b2"), T("22")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indexname FROM pg_indexes WHERE tablename = 't' ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("indexname", NAME)],
                        rows: &[
                            &[T("t_pkey")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_branch('feat2', 'HEAD~1');",
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
                    query: "ALTER TABLE t DROP COLUMN b;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-Am', 'drop b')) > 0 AND length(dolt_checkout('feat2')::text) > 0;",
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
                    query: "UPDATE t SET b = 'edit2' WHERE k = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-Am', 'edit b')) > 0 AND length(dolt_checkout('main')::text) > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SET dolt_allow_commit_conflicts = 1;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('feat2');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("1"), T("conflicts found")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT \"table\", num_conflicts FROM dolt_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("t"), T("1")],
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
fn test_merge_column_type_changes() {
    run_scripts(&[
        ScriptTest {
            name: "a merge converts rows for type changes with implicit or assignment casts and conflicts on others",
            set_up_script: &[
                "CREATE TABLE t (k INT PRIMARY KEY, v INT, w VARCHAR(5), x TEXT);",
                "INSERT INTO t VALUES (1, 1, 'a', '1'), (2, 2, 'b', '2');",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_branch('rows');",
                "SELECT dolt_branch('explicit');",
                "SELECT dolt_branch('more_rows');",
                "ALTER TABLE t ALTER COLUMN v TYPE TEXT;",
                "ALTER TABLE t ALTER COLUMN w TYPE VARCHAR(10);",
                "SELECT dolt_commit('-Am', 'widen');",
                "SELECT dolt_checkout('rows');",
                "UPDATE t SET w = 'c', v = 7 WHERE k = 2;",
                "SELECT dolt_commit('-Am', 'rows');",
                "SELECT dolt_checkout('explicit');",
                "ALTER TABLE t ALTER COLUMN x TYPE INT USING x::int;",
                "SELECT dolt_commit('-Am', 'explicit');",
                "SELECT dolt_checkout('more_rows');",
                "UPDATE t SET x = '9' WHERE k = 1;",
                "SELECT dolt_commit('-Am', 'more rows');",
                "SELECT dolt_checkout('main');",
                "CREATE TABLE n (k INT PRIMARY KEY, y INT8);",
                "INSERT INTO n VALUES (1, 1);",
                "SELECT dolt_commit('-Am', 'n');",
                "SELECT dolt_branch('big');",
                "ALTER TABLE n ALTER COLUMN y TYPE INT4;",
                "SELECT dolt_commit('-Am', 'narrow');",
                "SELECT dolt_checkout('big');",
                "INSERT INTO n VALUES (2, 5000000000);",
                "SELECT dolt_commit('-Am', 'big');",
                "SELECT dolt_checkout('main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('rows');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("0"), T("merge successful")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k, v, w FROM t ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("v", TEXT), Column("w", VARCHAR)],
                        rows: &[
                            &[T("1"), T("1"), T("a")],
                            &[T("2"), T("7"), T("c")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('big');",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "integer out of range", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('more_rows');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T("(0,\"Switched to branch 'more_rows'\")")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SET dolt_allow_commit_conflicts = 1;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('explicit');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("1"), T("conflicts found")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt_schema_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
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
fn test_merge_new_column_edits_and_schema_conflicts() {
    run_scripts(&[
        ScriptTest {
            name: "an edit of a column the base lacks leaves a delete alone, and dolt_conflicts lists schema conflicts",
            set_up_script: &[
                "CREATE TABLE t (a INT PRIMARY KEY, b INT);",
                "INSERT INTO t VALUES (1, 1), (2, 2), (3, 3);",
                "CREATE TABLE s (id INT PRIMARY KEY, a INT);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_checkout('-b', 'feature');",
                "ALTER TABLE t ADD COLUMN d TEXT DEFAULT 'x';",
                "DELETE FROM t WHERE a = 2;",
                "ALTER TABLE s ADD COLUMN extra INT;",
                "SELECT dolt_commit('-Am', 'feature');",
                "SELECT dolt_checkout('main');",
                "ALTER TABLE t ADD COLUMN c INT DEFAULT 7;",
                "UPDATE t SET c = 8 WHERE a = 2;",
                "ALTER TABLE s ADD COLUMN extra TEXT;",
                "SELECT dolt_commit('-Am', 'main');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('feature');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("1"), T("conflicts found")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4), Column("d", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("7"), T("x")],
                            &[T("3"), T("3"), T("7"), T("x")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT \"table\", num_conflicts FROM dolt_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("s"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name FROM dolt_schema_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT)],
                        rows: &[
                            &[T("s")],
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
    ]);
}

#[test]
fn test_merge_checks_and_aborts() {
    run_scripts(&[
        ScriptTest {
            name: "a merge records check violations, keeps root object conflicts, and its abort keeps ignored tables",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, a INT, b INT, CONSTRAINT ab CHECK (a <= b));",
                "INSERT INTO t VALUES (1, 1, 5), (2, 1, 5);",
                "CREATE TABLE items (id INT PRIMARY KEY, label TEXT);",
                "INSERT INTO items VALUES (1, 'base');",
                "INSERT INTO dolt_ignore VALUES ('runtime_%', true);",
                "CREATE TABLE audit (a INT PRIMARY KEY);",
                "SELECT dolt_commit('-Am', 'base');",
                "CREATE TABLE runtime_jobs (id INT PRIMARY KEY, kind TEXT);",
                "INSERT INTO runtime_jobs VALUES (1, 'keep');",
                "SELECT dolt_checkout('-b', 'side');",
                "UPDATE t SET a = 4 WHERE id = 1;",
                "UPDATE items SET label = 'side' WHERE id = 1;",
                "CREATE FUNCTION f() RETURNS trigger AS $$ BEGIN INSERT INTO audit VALUES (NEW.id); RETURN NEW; END $$ LANGUAGE plpgsql;",
                "SELECT dolt_commit('-Am', 'side');",
                "SELECT dolt_checkout('main');",
                "UPDATE t SET b = 2 WHERE id = 1;",
                "CREATE FUNCTION f() RETURNS trigger AS $$ BEGIN INSERT INTO audit VALUES (NEW.id + 100); RETURN NEW; END $$ LANGUAGE plpgsql;",
                "SELECT dolt_commit('-Am', 'main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('side');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Committing this transaction resulted in a working set with constraint violations, transaction rolled back. This constraint violation may be the result of a previous merge or the result of transaction sequencing. Constraint violations from a merge can be resolved using the dolt_constraint_violations table before committing the transaction. To allow transactions to be committed with constraint violations from a merge or transaction sequencing set @@dolt_force_transaction_commit=1.\nConstraint violations: \nType: Check Constraint Violation,\n\tName: ab,\n\tExpression: a <= b", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SET dolt_allow_commit_conflicts = 1;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SET dolt_force_transaction_commit = 1;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('side');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("1"), T("conflicts found")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT violation_type, id, a, b FROM dolt_constraint_violations_t;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("id", INT4), Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("check constraint"), T("1"), T("4"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE runtime_jobs SET kind = 'updated' WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_merge('--abort');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_merge", RECORD)],
                        rows: &[
                            &[T("(\"\",0,0,\"merge aborted\")")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM runtime_jobs;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("kind", TEXT)],
                        rows: &[
                            &[T("1"), T("updated")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM items;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("1"), T("base")],
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
fn test_docs_agent_updates_and_aliases() {
    run_scripts(&[
        ScriptTest {
            name: "an UPDATE of the default AGENT.md stores it, and table arguments accept dolt_docs",
            set_up_script: &[],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "UPDATE dolt_docs SET doc_text = 'x' WHERE doc_name = 'README.md';",
                    expected: Expected::Tag("UPDATE 0"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_docs WHERE doc_text LIKE '# AGENT.md%';",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "UPDATE dolt_docs SET doc_text = 'edited' WHERE doc_name = 'AGENT.md';",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT doc_name, doc_text FROM dolt_docs;",
                    expected: Expected::Rows {
                        columns: &[Column("doc_name", TEXT), Column("doc_text", TEXT)],
                        rows: &[
                            &[T("AGENT.md"), T("edited")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('dolt_docs');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Rust server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("dolt.docs"), T("t"), T("new table")],
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
fn test_hashof_three_dot_log_reset_and_detached_sessions() {
    run_scripts(&[
        ScriptTest {
            name: "hashof matches names without case, logs take three-dot ranges, reset checks conflicts, and detached sessions limit tags",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, v TEXT);",
                "INSERT INTO t VALUES (1, 'a');",
                "SELECT dolt_commit('-Am', 'init');",
                "SELECT dolt_tag('v1');",
                "SELECT dolt_branch('Feat');",
                "SELECT dolt_checkout('Feat');",
                "INSERT INTO t VALUES (2, 'f');",
                "SELECT dolt_commit('-Am', 'feat');",
                "SELECT dolt_checkout('main');",
                "INSERT INTO t VALUES (3, 'm');",
                "SELECT dolt_commit('-Am', 'main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_hashof('MAIN') = dolt_hashof('main'), dolt_hashof('feat') = dolt_hashof('Feat');",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT message FROM dolt_log('main...Feat') ORDER BY message;",
                    expected: Expected::Rows {
                        columns: &[Column("message", TEXT)],
                        rows: &[
                            &[T("feat")],
                            &[T("main")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET v = 'main' WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-Am', 'main edit')) > 0 AND length(dolt_checkout('Feat')::text) > 0;",
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
                    query: "UPDATE t SET v = 'feat' WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-Am', 'feat edit')) > 0 AND length(dolt_checkout('main')::text) > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SET dolt_allow_commit_conflicts = 1;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts FROM dolt_merge('Feat');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "RESET dolt_allow_commit_conflicts;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_reset();",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "Merge conflict detected, @autocommit transaction rolled back. @autocommit must be disabled so that merge conflicts can be resolved using the dolt_conflicts and dolt_schema_conflicts tables before manually committing the transaction. Alternatively, to commit transactions with merge conflicts, set @@dolt_allow_commit_conflicts = 1", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "USE postgres/v1",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_tag('blocked');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "this operation is not supported while in a detached head state", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('v1');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "dolt does not support a detached head state. To create a branch at this tag, run: \n\tCALL DOLT_CHECKOUT('v1', '-b', <new_branch_name>)", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_gc();",
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
fn test_status_renames_foreign_keys_and_history_as_of() {
    run_scripts(&[
        ScriptTest {
            name: "status leaves a referenced table alone, table resets follow renames, and history tables resolve AS OF",
            set_up_script: &[
                "CREATE TABLE parent (id INT PRIMARY KEY);",
                "CREATE TABLE t (id INT PRIMARY KEY, v INT);",
                "SELECT dolt_commit('-Am', 'base');",
                "CREATE TABLE child (id INT PRIMARY KEY, parent_id INT REFERENCES parent (id));",
                "SELECT dolt_add('child');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status ORDER BY table_name;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.child"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'child')) > 0;",
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
                    query: "ALTER TABLE t RENAME TO t2;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('t2'), dolt_reset('t2');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8), Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t -> public.t2"), T("t"), T("renamed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_reset('t');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_reset", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, staged, status FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.t -> public.t2"), T("f"), T("renamed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_diff_summary('HEAD~1', 'HEAD', NULL);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "Invalid argument to dolt_diff_summary: NULL", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_checkout('-b', 'feature')::text) > 0;",
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
                    query: "CREATE TABLE feature_only (id INT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO feature_only VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-Am', 'feature only')) > 0 AND length(dolt_checkout('main')::text) > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT id FROM dolt_history_feature_only AS OF 'feature';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
fn test_merge_foreign_keys_and_schema_conflicts() {
    run_scripts(&[
        ScriptTest {
            name: "a merge keeps the foreign keys both sides added and refuses keys both sides changed differently",
            set_up_script: &[
                "CREATE TABLE parent (id INT PRIMARY KEY);",
                "CREATE TABLE a (id INT PRIMARY KEY, pid INT);",
                "CREATE TABLE b (id INT PRIMARY KEY, pid INT);",
                "CREATE TABLE c (id INT PRIMARY KEY, pid INT REFERENCES parent (id));",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_checkout('-b', 'feat');",
                "ALTER TABLE b ADD CONSTRAINT b_fk FOREIGN KEY (pid) REFERENCES parent (id);",
                "SELECT dolt_commit('-Am', 'feat fk');",
                "SELECT dolt_checkout('-b', 'feat2', 'main');",
                "ALTER TABLE c DROP CONSTRAINT c_pid_fkey;",
                "ALTER TABLE c ADD CONSTRAINT c_pid_fkey FOREIGN KEY (pid) REFERENCES parent (id) ON DELETE CASCADE;",
                "SELECT dolt_commit('-Am', 'cascade');",
                "SELECT dolt_checkout('main');",
                "ALTER TABLE a ADD CONSTRAINT a_fk FOREIGN KEY (pid) REFERENCES parent (id);",
                "ALTER TABLE c DROP CONSTRAINT c_pid_fkey;",
                "ALTER TABLE c ADD CONSTRAINT c_pid_fkey FOREIGN KEY (pid) REFERENCES parent (id) ON DELETE SET NULL;",
                "SELECT dolt_commit('-Am', 'main fks');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('feat');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8), Column("message", TEXT)],
                        rows: &[
                            &[T("0"), T("merge successful")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_constraint WHERE contype = 'f' ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[
                            &[T("a_fk")],
                            &[T("b_fk")],
                            &[T("c_pid_fkey")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts, message FROM dolt_merge('feat2');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "foreign key conflicts", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt_conflicts_resolve refuses schema conflicts, and dolt_verify_constraints matches tables without regard to case",
            set_up_script: &[
                "CREATE TABLE s (id INT PRIMARY KEY, a INT);",
                "CREATE TABLE \"MiXeD\" (id INT PRIMARY KEY);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_checkout('-b', 'feature');",
                "ALTER TABLE s ADD COLUMN extra INT;",
                "SELECT dolt_commit('-Am', 'feature');",
                "SELECT dolt_checkout('main');",
                "ALTER TABLE s ADD COLUMN extra TEXT;",
                "SELECT dolt_commit('-Am', 'main');",
                "INSERT INTO \"MiXeD\" VALUES (1);",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_verify_constraints('mixed');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('MIXED');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "error: the table(s) MIXED do not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts FROM dolt_merge('feature');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_conflicts_resolve('--ours', 's');",
                    expected: Expected::Error(Diagnostic {
                        code: "XX000",
                        message: "Unable to automatically resolve schema conflicts since data changes may not have been fully merged yet. To continue, abort this merge (dolt merge --abort) then apply ALTER TABLE statements to one side of this merge to get the two schemas in sync with the desired schema, then rerun the merge. To track resolution of this limitation, follow https://github.com/dolthub/dolt/issues/6616",
                        ..E
                    }),
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
            name: "a merge refuses two new foreign keys with one name on different tables",
            set_up_script: &[
                "CREATE TABLE parent (id INT PRIMARY KEY);",
                "CREATE TABLE parent2 (id INT PRIMARY KEY);",
                "CREATE TABLE child (id INT PRIMARY KEY, pid INT);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_checkout('-b', 'feat');",
                "ALTER TABLE child ADD FOREIGN KEY (pid) REFERENCES parent (id);",
                "SELECT dolt_commit('-Am', 'feat');",
                "SELECT dolt_checkout('main');",
                "ALTER TABLE child ADD FOREIGN KEY (pid) REFERENCES parent2 (id);",
                "SELECT dolt_commit('-Am', 'main');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts FROM dolt_merge('feat');",
                    expected: Expected::Error(Diagnostic {
                        code: "XX000",
                        message: "duplicate foreign key constraint name `child_pid_fkey`",
                        ..E
                    }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dolt_schema_conflicts shows a table one side dropped as <deleted>",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, v INT);",
                "CREATE TABLE u (id INT PRIMARY KEY, v INT);",
                "SELECT dolt_commit('-Am', 'base');",
                "SELECT dolt_checkout('-b', 'feature');",
                "ALTER TABLE t ADD COLUMN c INT;",
                "INSERT INTO u VALUES (1, 1);",
                "SELECT dolt_commit('-Am', 'feature');",
                "SELECT dolt_checkout('main');",
                "DROP TABLE t;",
                "DROP TABLE u;",
                "SELECT dolt_commit('-Am', 'main');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT conflicts FROM dolt_merge('feature');",
                    expected: Expected::Rows {
                        columns: &[Column("conflicts", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT table_name, our_schema, description FROM dolt_schema_conflicts ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("our_schema", TEXT), Column("description", TEXT)],
                        rows: &[
                            &[T("t"), T("<deleted>"), T("cannot merge a table deletion with schema modification")],
                            &[T("u"), T("<deleted>"), T("cannot merge a table deletion with data modification")],
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
            ],
            ..S
        },
    ]);
}

#[test]
fn test_verify_constraints_reverifies_unique_indexes() {
    run_scripts(&[
        ScriptTest {
            name: "dolt_verify_constraints records unique index violations again among changed rows or all rows",
            set_up_script: &[
                "CREATE TABLE t (pk INT PRIMARY KEY, col1 INT UNIQUE);",
                "SELECT dolt_commit('-Am', 'init');",
                "SELECT dolt_branch('branch1');",
                "INSERT INTO t VALUES (1, 1);",
                "SELECT dolt_commit('-Am', 'insert on main');",
                "SELECT dolt_checkout('branch1');",
                "INSERT INTO t VALUES (2, 1);",
                "SELECT dolt_commit('-Am', 'insert on branch1');",
                "SELECT dolt_checkout('main');",
                "SET dolt_force_transaction_commit = 1;",
                "BEGIN;",
                "SELECT dolt_merge('branch1');",
                "COMMIT;",
                "DELETE FROM dolt_constraint_violations_t;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_verify_constraints();",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT violation_type, pk, col1, violation_info::text FROM dolt_constraint_violations_t ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("pk", INT4), Column("col1", INT4), Column("violation_info", TEXT)],
                        rows: &[
                            &[T("unique index"), T("1"), T("1"), T("{\"Columns\":[\"col1\"],\"Name\":\"t_col1_key\"}")],
                            &[T("unique index"), T("2"), T("1"), T("{\"Columns\":[\"col1\"],\"Name\":\"t_col1_key\"}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM dolt_constraint_violations_t;",
                    expected: Expected::Tag("DELETE 2"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_verify_constraints('--all', '--output-only');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_constraint_violations_t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so Dolt's behavior is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_verify_constraints('--all');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_verify_constraints", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM dolt_constraint_violations_t;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("2")],
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
