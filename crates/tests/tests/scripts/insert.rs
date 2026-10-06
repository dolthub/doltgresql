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
fn test_insert() {
    run_scripts(&[
        ScriptTest {
            name: "simple insert",
            set_up_script: &[
                "CREATE TABLE mytable (id INT PRIMARY KEY, name TEXT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mytable order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("hello")],
                            &[T("2"), T("world")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "keyless insert",
            set_up_script: &[
                "CREATE TABLE mytable (id INT, name TEXT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mytable order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("hello")],
                            &[T("2"), T("world")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "on conflict clause",
            set_up_script: &[
                "CREATE TABLE mytable (id INT primary key, name TEXT)",
                "create table t2 (id int primary key, c1 text, c2 text)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (ID, naME) VALUES (1, 'world') ON CONFLICT (id) DO UPDATE SET name = 'world'",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (ID, naME) VALUES (2, 'hello') ON CONFLICT (id) DO UPDATE SET name = 'conflict'",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (ID, naME) VALUES (1, 'not inserted') ON CONFLICT (id) DO NOTHING",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mytable order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("world")],
                            &[T("2"), T("conflict")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mytable (ID, naME) VALUES (1, 'hello') ON CONFLICT (id) DO UPDATE set name = concat('new', name)",
                    expected: Expected::Error(Diagnostic { code: "42702", message: r#"column reference "name" is ambiguous"#, position: 104, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM mytable order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT)],
                        rows: &[
                            &[T("1"), T("world")],
                            &[T("2"), T("conflict")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 (id, c1, c2) VALUES (1, 'hello', 'world'), (2, 'world', 'hello')",
                    expected: Expected::Tag("INSERT 0 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 (id, c1, c2) VALUES (1, 'hello', 'world') ON CONFLICT (id) DO UPDATE SET c1 = 'conflict', c2 = c1",
                    expected: Expected::Error(Diagnostic { code: "42702", message: r#"column reference "c1" is ambiguous"#, position: 111, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 (id, c1, c2) VALUES (2, 'hello', 'world') ON CONFLICT (id) DO UPDATE SET c2 = c1",
                    expected: Expected::Error(Diagnostic { code: "42702", message: r#"column reference "c1" is ambiguous"#, position: 94, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2 order by id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c1", TEXT), Column("c2", TEXT)],
                        rows: &[
                            &[T("1"), T("hello"), T("world")],
                            &[T("2"), T("world"), T("hello")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO t2 (id, c1, c2) 
VALUES ($1, $2, $3)
ON CONFLICT (id) do update set c1 = $4"#,
                    bind_vars: &[BindVar::Int(1), BindVar::Str("x"), BindVar::Str("y"), BindVar::Str("no conflict expected")],
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "conditional on conflict update",
            set_up_script: &[
                "CREATE TABLE conditional_upsert (id INT PRIMARY KEY, version INT, note TEXT)",
                "INSERT INTO conditional_upsert VALUES (1, 5, 'original'), (2, 1, 'second')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert VALUES (1, 6, 'advanced') ON CONFLICT (id) DO UPDATE SET version = 6, note = 'advanced' WHERE conditional_upsert.version < 6",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert VALUES (1, 6, 'advanced') ON CONFLICT (id) DO UPDATE SET version = 6, note = 'advanced' WHERE conditional_upsert.version <= 6",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert VALUES (1, 4, 'stale') ON CONFLICT (id) DO UPDATE SET version = 4, note = 'stale' WHERE conditional_upsert.version < 4",
                    expected: Expected::Tag("INSERT 0 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert VALUES (1, 7, 'one'), (2, 0, 'two'), (3, 3, 'three') ON CONFLICT (id) DO UPDATE SET version = 7, note = 'updated' WHERE conditional_upsert.version <= 5",
                    expected: Expected::Tag("INSERT 0 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert VALUES (2, 8, 'null predicate') ON CONFLICT (id) DO UPDATE SET version = 8 WHERE NULL",
                    expected: Expected::Tag("INSERT 0 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert VALUES (2, $1, 'bound') ON CONFLICT (id) DO UPDATE SET version = $1, note = 'bound' WHERE conditional_upsert.version < $1",
                    bind_vars: &[BindVar::Int(8)],
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert VALUES (1, 9, 'proposed') ON CONFLICT (id) DO UPDATE SET version = excluded.version, note = excluded.note WHERE conditional_upsert.version < excluded.version RETURNING id, version, note",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("version", INT4), Column("note", TEXT)],
                        rows: &[
                            &[T("1"), T("9"), T("proposed")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert VALUES (1, 10, 'casted') ON CONFLICT (id) DO UPDATE SET version = excluded.version::BIGINT, note = excluded.note WHERE conditional_upsert.version < excluded.version RETURNING id, version",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("version", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conditional_upsert SELECT 1, 11::BIGINT, 'selected' ON CONFLICT (id) DO UPDATE SET version = excluded.version, note = excluded.note WHERE conditional_upsert.version < excluded.version RETURNING id, version, note",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("version", INT4), Column("note", TEXT)],
                        rows: &[
                            &[T("1"), T("11"), T("selected")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM conditional_upsert ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("version", INT4), Column("note", TEXT)],
                        rows: &[
                            &[T("1"), T("11"), T("selected")],
                            &[T("2"), T("8"), T("bound")],
                            &[T("3"), T("3"), T("three")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "on conflict update returning with check constraint",
            set_up_script: &[
                "CREATE TABLE checked_upsert (id INT PRIMARY KEY, a TEXT CHECK (a <> ''), b TEXT, c TEXT)",
                "INSERT INTO checked_upsert VALUES (1, 'x', 'y', 'z')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO checked_upsert VALUES (1, 'x', 'y', 'z') ON CONFLICT (id) DO UPDATE SET a = 'n1', b = 'n2' RETURNING id, a, b",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", TEXT), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("n1"), T("n2")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO checked_upsert VALUES (2, 'x', 'y', 'z') ON CONFLICT (id) DO UPDATE SET a = 'n1', b = 'n2' RETURNING id, a, b",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", TEXT), Column("b", TEXT)],
                        rows: &[
                            &[T("2"), T("x"), T("y")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "on conflict do nothing only ignores uniqueness conflicts",
            set_up_script: &[
                "CREATE TABLE conflict_parent (id INT PRIMARY KEY)",
                "CREATE TABLE conflict_child (id INT PRIMARY KEY, parent_id INT NOT NULL REFERENCES conflict_parent(id), positive INT CHECK (positive > 0))",
                "CREATE TABLE self_referencing_child (id INT PRIMARY KEY, parent_id INT REFERENCES self_referencing_child(id))",
                "CREATE TABLE secondary_unique_child (id INT PRIMARY KEY, unique_value INT UNIQUE, parent_id INT REFERENCES conflict_parent(id))",
                "CREATE TABLE conflict_arbiter (id INT PRIMARY KEY, unique_value INT UNIQUE, a INT, b INT, UNIQUE (a, b))",
                "CREATE TABLE invalid_conflict_target (id INT PRIMARY KEY, non_unique INT)",
                "CREATE TABLE crossed_conflict (id INT PRIMARY KEY, unique_value INT UNIQUE)",
                "CREATE TABLE crossed_keyless_conflict (a INT UNIQUE, b INT UNIQUE)",
                "CREATE TABLE crossed_partial_conflict (a INT, b INT UNIQUE)",
                "CREATE UNIQUE INDEX a_partial ON crossed_partial_conflict (a) WHERE b > 0",
                "CREATE UNIQUE INDEX z_full ON crossed_partial_conflict (a)",
                "INSERT INTO conflict_parent VALUES (1)",
                "INSERT INTO conflict_child VALUES (1, 1, 1)",
                "INSERT INTO secondary_unique_child VALUES (1, 10, 1)",
                "INSERT INTO conflict_arbiter VALUES (1, 10, 20, 30)",
                "INSERT INTO crossed_conflict VALUES (1, 10), (2, 20)",
                "INSERT INTO crossed_keyless_conflict VALUES (1, 10), (2, 20)",
                "INSERT INTO crossed_partial_conflict VALUES (1, 10), (2, 20)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_child VALUES (2, 999, 1) ON CONFLICT DO NOTHING",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "conflict_child" violates foreign key constraint "conflict_child_parent_id_fkey""#, detail: r#"Key (parent_id)=(999) is not present in table "conflict_parent"."#, schema: "public", table: "conflict_child", constraint: "conflict_child_parent_id_fkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_child VALUES (2, 1, -1) ON CONFLICT DO NOTHING",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "conflict_child" violates check constraint "conflict_child_positive_check""#, detail: "Failing row contains (2, 1, -1).", schema: "public", table: "conflict_child", constraint: "conflict_child_positive_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_child VALUES (1, 999, 1) ON CONFLICT DO NOTHING",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO secondary_unique_child VALUES (2, 10, 999) ON CONFLICT DO NOTHING",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM secondary_unique_child ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("unique_value", INT4), Column("parent_id", INT4)],
                        rows: &[
                            &[T("1"), T("10"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_arbiter VALUES (1, 11, 21, 31) ON CONFLICT (id) DO NOTHING",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_arbiter VALUES (2, 10, 21, 31) ON CONFLICT (id) DO NOTHING",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "conflict_arbiter_unique_value_key""#, detail: "Key (unique_value)=(10) already exists.", schema: "public", table: "conflict_arbiter", constraint: "conflict_arbiter_unique_value_key", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_arbiter VALUES (2, 11, 20, 30) ON CONFLICT (a, b) DO NOTHING",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_arbiter VALUES (2, 10, 21, 31) ON CONFLICT DO NOTHING",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM conflict_arbiter ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("unique_value", INT4), Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("10"), T("20"), T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO invalid_conflict_target VALUES (1, 10) ON CONFLICT (non_unique) DO NOTHING",
                    expected: Expected::Error(Diagnostic { code: "42P10", message: "there is no unique or exclusion constraint matching the ON CONFLICT specification", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO invalid_conflict_target VALUES (1, 10) ON CONFLICT (missing) DO NOTHING",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "missing" does not exist"#, position: 64, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM invalid_conflict_target",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("non_unique", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO crossed_conflict VALUES (1, 20) ON CONFLICT (id) DO NOTHING",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO crossed_conflict VALUES (1, 20) ON CONFLICT (unique_value) DO NOTHING",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO crossed_conflict VALUES (3, 30), (3, 10) ON CONFLICT (id) DO NOTHING",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM crossed_conflict ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("unique_value", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("20")],
                            &[T("3"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO crossed_keyless_conflict VALUES (3, 30), (3, 10) ON CONFLICT (a) DO NOTHING",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM crossed_keyless_conflict ORDER BY a",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("20")],
                            &[T("3"), T("30")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO crossed_partial_conflict VALUES (3, -1), (3, 10) ON CONFLICT (a) DO NOTHING",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM crossed_partial_conflict ORDER BY a",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("20")],
                            &[T("3"), T("-1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO self_referencing_child VALUES (1, 1) ON CONFLICT DO NOTHING",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM self_referencing_child",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("parent_id", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_child VALUES (1, 1, 1), (2, 1, 1) ON CONFLICT DO NOTHING RETURNING id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM conflict_child ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("parent_id", INT4), Column("positive", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("1")],
                            &[T("2"), T("1"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO conflict_child VALUES (2, 1, 1), (3, 999, 1) ON CONFLICT DO NOTHING",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "conflict_child" violates foreign key constraint "conflict_child_parent_id_fkey""#, detail: r#"Key (parent_id)=(999) is not present in table "conflict_parent"."#, schema: "public", table: "conflict_child", constraint: "conflict_child_parent_id_fkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM conflict_child ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("parent_id", INT4), Column("positive", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("1")],
                            &[T("2"), T("1"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "null and unspecified default values",
            set_up_script: &[
                "CREATE TABLE t (i INT DEFAULT NULL, j INT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (default, default)",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("j", INT4)],
                        rows: &[
                            &[Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "default values compatibility",
            set_up_script: &[
                "CREATE TABLE ordinary_defaults (a INT DEFAULT 1, b INT DEFAULT 2)",
                "CREATE TABLE generated_defaults (a INT DEFAULT 1, b INT GENERATED ALWAYS AS (a + 1) STORED)",
                "CREATE TABLE identity_defaults (id BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY, v INT DEFAULT 5)",
                "CREATE TEMP TABLE serial_defaults (small_id SMALLSERIAL, id SERIAL PRIMARY KEY, big_id BIGSERIAL, v INT DEFAULT 5)",
                "CREATE TABLE rejected_empty_rows (a INT DEFAULT 1, b INT DEFAULT 2)",
                "CREATE TABLE required_defaults (a INT NOT NULL, b INT DEFAULT 2)",
                "INSERT INTO required_defaults VALUES (9, 9)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO ordinary_defaults DEFAULT VALUES",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ordinary_defaults",
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
                    query: "INSERT INTO generated_defaults DEFAULT VALUES",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generated_defaults",
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
                    query: "INSERT INTO identity_defaults DEFAULT VALUES",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO identity_defaults VALUES (DEFAULT, DEFAULT), (DEFAULT, DEFAULT)",
                    expected: Expected::Tag("INSERT 0 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM identity_defaults ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("5")],
                            &[T("2"), T("5")],
                            &[T("3"), T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO serial_defaults DEFAULT VALUES",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO serial_defaults DEFAULT VALUES",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM serial_defaults ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("small_id", INT2), Column("id", INT4), Column("big_id", INT8), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("1"), T("5")],
                            &[T("2"), T("2"), T("2"), T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rejected_empty_rows VALUES ()",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near ")""#, position: 41, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rejected_empty_rows VALUES (), ()",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near ")""#, position: 41, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rejected_empty_rows () VALUES ()",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near ")""#, position: 34, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rejected_empty_rows (a) VALUES ()",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near ")""#, position: 45, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM rejected_empty_rows",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ordinary_defaults VALUES (DEFAULT, DEFAULT), (3, DEFAULT)",
                    expected: Expected::Tag("INSERT 0 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM ordinary_defaults ORDER BY a",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("1"), T("2")],
                            &[T("3"), T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO generated_defaults VALUES (DEFAULT, DEFAULT), (3, DEFAULT)",
                    expected: Expected::Tag("INSERT 0 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generated_defaults ORDER BY a",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("1"), T("2")],
                            &[T("3"), T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO required_defaults VALUES (DEFAULT, DEFAULT), (5, DEFAULT)",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "a" of relation "required_defaults" violates not-null constraint"#, detail: "Failing row contains (null, 2).", schema: "public", table: "required_defaults", column: "a", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM required_defaults",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("9"), T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "types",
            set_up_script: &[
                "create table child (i2 int2, i4 int4, i8 int8, f float, d double precision, v varchar, vl varchar(100), t text, j json, ts timestamp);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"insert into child values (1, 2, 3, 4.5, 6.7, 'hello', 'world', 'text', '{"a": 1}', '2021-01-01 00:00:00');"#,
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from child;",
                    expected: Expected::Rows {
                        columns: &[Column("i2", INT2), Column("i4", INT4), Column("i8", INT8), Column("f", FLOAT8), Column("d", FLOAT8), Column("v", VARCHAR), Column("vl", VARCHAR), Column("t", TEXT), Column("j", JSON), Column("ts", TIMESTAMP)],
                        rows: &[
                            &[T("1"), T("2"), T("3"), T("4.5"), T("6.7"), T("hello"), T("world"), T("text"), T(r#"{"a": 1}"#), T("2021-01-01 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "insert returning",
            set_up_script: &[
                "CREATE TABLE t (i serial, j INT)",
                "CREATE TABLE u (u uuid DEFAULT 'ac1f3e2d-1e4b-4d3e-8b1f-2b7f1e7f0e3d', j INT)",
                "CREATE TABLE s (v1 varchar DEFAULT 'hello', v2 varchar DEFAULT 'world')",
                "CREATE SCHEMA ts",
                "CREATE TABLE ts.t (i serial, j INT)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "INSERT 0 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i+3",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("7")],
                            &[T("8")],
                            &[T("9")],
                        ],
                        tag: "INSERT 0 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i+j, j-3*i",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("12"), T("-16")],
                            &[T("14"), T("-18")],
                            &[T("16"), T("-20")],
                        ],
                        tag: "INSERT 0 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO u (j) VALUES (5), (6), (7) RETURNING u",
                    expected: Expected::Rows {
                        columns: &[Column("u", UUID)],
                        rows: &[
                            &[T("ac1f3e2d-1e4b-4d3e-8b1f-2b7f1e7f0e3d")],
                            &[T("ac1f3e2d-1e4b-4d3e-8b1f-2b7f1e7f0e3d")],
                            &[T("ac1f3e2d-1e4b-4d3e-8b1f-2b7f1e7f0e3d")],
                        ],
                        tag: "INSERT 0 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO s (v2) VALUES (' a') RETURNING concat(v1, v2)",
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT)],
                        rows: &[
                            &[T("hello a")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO s (v1) VALUES ('sup ') RETURNING concat(v1, v2)",
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT)],
                        rows: &[
                            &[T("sup world")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO s (v2, v1) VALUES ('def', 'abc'), ('xyz', 'uvw') RETURNING concat(v1, v2), concat(v2, v1), 100",
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT), Column("concat", TEXT), Column("?column?", INT4)],
                        rows: &[
                            &[T("abcdef"), T("defabc"), T("100")],
                            &[T("uvwxyz"), T("xyzuvw"), T("100")],
                        ],
                        tag: "INSERT 0 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i, doesnotexist",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "doesnotexist" does not exist"#, position: 53, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i, doesnotexist(j)",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function doesnotexist(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 53, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.t (j) VALUES (8) RETURNING t.j",
                    expected: Expected::Rows {
                        columns: &[Column("j", INT4)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.t (j) VALUES (9) RETURNING public.t.j",
                    expected: Expected::Rows {
                        columns: &[Column("j", INT4)],
                        rows: &[
                            &[T("9")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ts.t (j) VALUES (10) RETURNING ts.t.j",
                    expected: Expected::Rows {
                        columns: &[Column("j", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.t (j) VALUES ($1) RETURNING j;",
                    bind_vars: &[BindVar::Int(11)],
                    expected: Expected::Rows {
                        columns: &[Column("j", INT4)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.t (j) VALUES ($1) RETURNING t.j;",
                    bind_vars: &[BindVar::Int(12)],
                    expected: Expected::Rows {
                        columns: &[Column("j", INT4)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.t (j) VALUES ($1) RETURNING public.t.j;",
                    bind_vars: &[BindVar::Int(13)],
                    expected: Expected::Rows {
                        columns: &[Column("j", INT4)],
                        rows: &[
                            &[T("13")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "insert iso8601 timestamptz literal",
            set_up_script: &[
                "CREATE TABLE django_migrations (id serial primary key, app varchar, name varchar, applied timestamptz)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-24T19:21:59.690479+00:00'::timestamptz) RETURNING "django_migrations"."id""#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "insert on conflict do nothing returning",
            set_up_script: &[
                "CREATE TABLE t4 (k INT PRIMARY KEY, v TEXT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t4 VALUES (1, 'a') ON CONFLICT DO NOTHING RETURNING k, v;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t4 VALUES (1, 'b'), (2, 'c') ON CONFLICT DO NOTHING RETURNING k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t4 VALUES (1, 'b') ON CONFLICT (k) DO NOTHING RETURNING *;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("v", TEXT)],
                        rows: &[],
                        tag: "INSERT 0 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t4 ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("v", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("c")],
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
