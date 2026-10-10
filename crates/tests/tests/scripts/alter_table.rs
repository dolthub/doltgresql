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
fn test_alter_table() {
    run_scripts(&[
        ScriptTest {
            name: "Add Foreign Key Constraint",
            set_up_script: &[
                "create table child (pk int primary key, c1 int);",
                "insert into child values (1,1), (2,2), (3,3);",
                "create index idx_child_c1 on child (pk, c1);",
                "create table parent (pk int primary key, c1 int, c2 int);",
                "insert into parent values (1, 1, 10);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE parent ADD FOREIGN KEY (c1) REFERENCES child (pk) ON DELETE CASCADE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO parent VALUES (10, 10, 10);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "parent" violates foreign key constraint "parent_c1_fkey""#, detail: r#"Key (c1)=(10) is not present in table "child"."#, schema: "public", table: "parent", constraint: "parent_c1_fkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE parent ADD FOREIGN KEY (c2) REFERENCES child (pk);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "parent" violates foreign key constraint "parent_c2_fkey""#, detail: r#"Key (c2)=(10) is not present in table "child"."#, schema: "public", table: "parent", constraint: "parent_c2_fkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE parent ADD FOREIGN KEY (c1, c2) REFERENCES child (pk, c1);",
                    expected: Expected::Error(Diagnostic { code: "42830", message: r#"there is no unique constraint matching given keys for referenced table "child""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE parent ADD FOREIGN KEY (c1, c2) REFERENCES child (pk, c1) MATCH PARTIAL;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "MATCH PARTIAL not yet implemented", position: 71, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Add Unique Constraint",
            set_up_script: &[
                "create table t1 (pk int primary key, c1 int);",
                "insert into t1 values (1,1);",
                "create table t2 (pk int primary key, c1 int);",
                "insert into t2 values (1,1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE UNIQUE INDEX ON t1(c1);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (2, 1);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t1_c1_idx""#, detail: "Key (c1)=(1) already exists.", schema: "public", table: "t1", constraint: "t1_c1_idx", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t2 ADD CONSTRAINT uniq1 UNIQUE (c1);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (2, 1);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "uniq1""#, detail: "Key (c1)=(1) already exists.", schema: "public", table: "t2", constraint: "uniq1", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Add Check Constraint",
            set_up_script: &[
                "create table t1 (pk int primary key, c1 int);",
                "insert into t1 values (1,1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 > 100);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"check constraint "constraint1" of relation "t1" is violated by some row"#, schema: "public", table: "t1", constraint: "constraint1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 < 100);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (2, 2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (3, 101);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t1" violates check constraint "constraint1""#, detail: "Failing row contains (3, 101).", schema: "public", table: "t1", constraint: "constraint1", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Add Check Constraint with IN tuple",
            set_up_script: &[
                "create table t1 (pk int primary key, c1 int);",
                "insert into t1 values (1,1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 in (100));",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"check constraint "constraint1" of relation "t1" is violated by some row"#, schema: "public", table: "t1", constraint: "constraint1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 in (1,2));",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (2, 2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (3, 101);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t1" violates check constraint "constraint1""#, detail: "Failing row contains (3, 101).", schema: "public", table: "t1", constraint: "constraint1", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Add Check Constraint and another constraint in same statement",
            set_up_script: &[
                "create table t1 (pk int, c1 int);",
                "insert into t1 values (1,1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: " ALTER TABLE t1 ADD CONSTRAINT check_a CHECK (c1 IN (1)), ALTER c1 SET NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (2, 2);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t1" violates check constraint "check_a""#, detail: "Failing row contains (2, 2).", schema: "public", table: "t1", constraint: "check_a", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (1, NULL);",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "c1" of relation "t1" violates not-null constraint"#, detail: "Failing row contains (1, null).", schema: "public", table: "t1", column: "c1", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Drop Constraint",
            set_up_script: &[
                "create table t1 (pk int primary key, c1 int);",
                "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 > 100);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 DROP CONSTRAINT constraint1;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (1, 1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 DROP CONSTRAINT doesnotexist;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"constraint "doesnotexist" of relation "t1" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 DROP CONSTRAINT IF EXISTS doesnotexist;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"constraint "doesnotexist" of relation "t1" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Add Primary Key",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b INT);",
                "CREATE TABLE test2 (a INT, b INT, c INT);",
                "CREATE TABLE pkTable1 (a INT PRIMARY KEY);",
                "CREATE TABLE duplicateRows (a INT, b INT);",
                "INSERT INTO duplicateRows VALUES (1, 2), (1, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ADD PRIMARY KEY (a);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT into test1 values (1, 2), (1, 3);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "test1_pkey""#, detail: "Key (a)=(1) already exists.", schema: "public", table: "test1", constraint: "test1_pkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test2 ADD PRIMARY KEY (a, b);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT into test2 values (1, 2, 3), (1, 2, 4);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "test2_pkey""#, detail: "Key (a, b)=(1, 2) already exists.", schema: "public", table: "test2", constraint: "test2_pkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE pkTable1 ADD PRIMARY KEY (a);",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"multiple primary keys for table "pktable1" are not allowed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE duplicateRows ADD PRIMARY KEY (a);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"could not create unique index "duplicaterows_pkey""#, detail: "Key (a)=(1) is duplicated.", schema: "public", table: "duplicaterows", constraint: "duplicaterows_pkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE IF EXISTS doesNotExist ADD PRIMARY KEY (a, b);",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"relation "doesnotexist" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Add Primary Key on text column",
            set_up_script: &[
                "CREATE TABLE test1 (a text, b INT);",
                "insert into test1 values ('a', 1), ('b', 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ADD PRIMARY KEY (a);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT into test1 values ('a', 3);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "test1_pkey""#, detail: "Key (a)=(a) already exists.", schema: "public", table: "test1", constraint: "test1_pkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from test1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", TEXT), Column("b", INT4)],
                        rows: &[
                            &[T("a"), T("1")],
                            &[T("b"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: Postgres defines gen_random_uuid in pg_catalog, not public.
        ScriptTest {
            name: "Add primary key with generated column",
            set_up_script: &[
                r#"CREATE TABLE t1 (
      id uuid DEFAULT gen_random_uuid() NOT NULL,
      data jsonb,
      has_data boolean GENERATED ALWAYS AS ((data IS NOT NULL)) STORED
  );"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: " ALTER TABLE ONLY t1 ADD CONSTRAINT pk PRIMARY KEY (id);",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t1 (id, data) values (default, '{}');",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "Select has_data from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("has_data", BOOL)],
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
            name: "Add Column",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b INT);",
                "INSERT INTO test1 VALUES (1, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ADD COLUMN c INT NOT NULL DEFAULT 42;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from test1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ADD COLUMN l non_existing_type;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "non_existing_type" does not exist"#, position: 32, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ADD COLUMN m xid;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Add column with inline check constraint",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ADD COLUMN c INT NOT NULL DEFAULT 42 CONSTRAINT chk1 CHECK (c > 0);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 VALUES (2, 2, -2);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "test1" violates check constraint "chk1""#, detail: "Failing row contains (2, 2, -2).", schema: "public", table: "test1", constraint: "chk1", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Add column with inline check constraint to table with existing data",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b INT);",
                "INSERT INTO test1 VALUES (1, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ADD COLUMN c INT NOT NULL DEFAULT 42 CONSTRAINT chk1 CHECK (c > 0);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 VALUES (2, 2, -2);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "test1" violates check constraint "chk1""#, detail: "Failing row contains (2, 2, -2).", schema: "public", table: "test1", constraint: "chk1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ADD COLUMN c2 INT CONSTRAINT chk2 CHECK (c2 IS NOT NULL);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"check constraint "chk2" of relation "test1" is violated by some row"#, schema: "public", table: "test1", constraint: "chk2", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Drop Column",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b INT, c INT, d INT);",
                "INSERT INTO test1 VALUES (1, 2, 3, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 DROP COLUMN c;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from test1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("d", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 DROP COLUMN d;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from test1;",
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
                    query: "ALTER TABLE test1 DROP COLUMN IF EXISTS zzz;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"column "zzz" of relation "test1" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE IF EXISTS doesNotExist DROP COLUMN z;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"relation "doesnotexist" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Rename Column",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b INT, c INT, d INT);",
                "INSERT INTO test1 VALUES (1, 2, 3, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 RENAME COLUMN c to jjj;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from test1 where jjj=3;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("jjj", INT4), Column("d", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Set Column Default",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b INT DEFAULT 42, c INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ALTER COLUMN c SET DEFAULT 43;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 (a) VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("42"), T("43")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ALTER COLUMN b DROP DEFAULT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 (a) VALUES (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test1 where a = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("2"), Null, T("43")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ALTER COLUMN c SET DEFAULT length('hello world');",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 (a) VALUES (3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test1 where a = 3;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("3"), Null, T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Set Column Nullability",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ALTER COLUMN b SET NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 VALUES (1, NULL);",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "b" of relation "test1" violates not-null constraint"#, detail: "Failing row contains (1, null).", schema: "public", table: "test1", column: "b", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ALTER COLUMN b DROP NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 VALUES (2, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test1 where a = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ALTER COLUMN b SET NOT NULL;",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"column "b" of relation "test1" contains null values"#, schema: "public", table: "test1", column: "b", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Alter Column Type",
            set_up_script: &[
                "CREATE TABLE test1 (a INT, b smallint);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test1 VALUES (1, 32769);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "smallint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ALTER COLUMN b TYPE INT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test1 VALUES (1, 32769);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("32769")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test1 ALTER COLUMN b TYPE smallint;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "smallint out of range", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Alter Column Type with USING clause",
            set_up_script: &[
                "CREATE TABLE t1 (id INT PRIMARY KEY, c TEXT);",
                "INSERT INTO t1 VALUES (1, '100'), (2, '-42'), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE integer USING c::integer;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c, pg_typeof(c) FROM t1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT4), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("1"), T("100"), T("integer")],
                            &[T("2"), T("-42"), T("integer")],
                            &[T("3"), Null, T("integer")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c FROM t1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("100")],
                            &[T("2"), T("-42")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (4, 'abc');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "abc""#, position: 27, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (4, 999);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE text USING 'val: ' || (c * 2);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c FROM t1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", TEXT)],
                        rows: &[
                            &[T("1"), T("val: 200")],
                            &[T("2"), T("val: -84")],
                            &[T("3"), Null],
                            &[T("4"), T("val: 1998")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Alter Column Type with USING expression form",
            set_up_script: &[
                "CREATE TABLE t1 (id INT PRIMARY KEY, c TEXT NOT NULL);",
                "INSERT INTO t1 VALUES (1, '1'), (2, '25');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE integer USING (c || '0')::integer;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c FROM t1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("250")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE text USING ('id-' || id || ': ' || c);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c FROM t1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", TEXT)],
                        rows: &[
                            &[T("1"), T("id-1: 10")],
                            &[T("2"), T("id-2: 250")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Alter Column Type with USING in Django migration style",
            set_up_script: &[
                r#"CREATE TABLE "app_event" ("id" integer PRIMARY KEY, "created" text);"#,
                r#"INSERT INTO "app_event" VALUES (1, '2024-01-15 10:30:00+00'), (2, '2025-06-01 08:00:00+00');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"ALTER TABLE "app_event" ALTER COLUMN "created" TYPE timestamp with time zone USING "created"::timestamp with time zone;"#,
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM app_event WHERE created = '2024-01-15 10:30:00+00'::timestamptz;",
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
                    query: "SELECT count(*) FROM app_event WHERE created > '2024-12-31 00:00:00+00'::timestamptz;",
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
            name: "Alter Column Type with USING error cases",
            set_up_script: &[
                "CREATE TABLE t1 (id INT PRIMARY KEY, c TEXT);",
                "INSERT INTO t1 VALUES (1, '100'), (2, 'abc');",
                "CREATE TABLE t2 (id INT PRIMARY KEY, c TEXT NOT NULL);",
                "INSERT INTO t2 VALUES (1, '');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE integer USING c::integer;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "abc""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c FROM t1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", TEXT)],
                        rows: &[
                            &[T("1"), T("100")],
                            &[T("2"), T("abc")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t2 ALTER COLUMN c TYPE integer USING NULLIF(c, '')::integer;",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"column "c" of relation "t2" contains null values"#, schema: "public", table: "t2", column: "c", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE integer USING c::integer, ALTER COLUMN id TYPE bigint;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "abc""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE integer USING d::integer;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "d" does not exist"#, position: 50, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE integer USING other.c::integer;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"missing FROM-clause entry for table "other""#, position: 50, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c TYPE varchar(20) USING t1.c || '!';",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c FROM t1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", VARCHAR)],
                        rows: &[
                            &[T("1"), T("100!")],
                            &[T("2"), T("abc!")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE doesnotexist ALTER COLUMN c TYPE integer USING c::integer;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "doesnotexist" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE IF EXISTS doesnotexist ALTER COLUMN c TYPE integer USING c::integer;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"relation "doesnotexist" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE fkparent (id TEXT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE fkchild (id INT PRIMARY KEY, p TEXT REFERENCES fkparent(id));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE fkchild ALTER COLUMN p TYPE integer USING p::integer;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fkchild_p_fkey" cannot be implemented"#, detail: r#"Key columns "p" of the referencing table and "id" of the referenced table are of incompatible types: integer and text."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE fkparent ALTER COLUMN id TYPE integer USING id::integer;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fkchild_p_fkey" cannot be implemented"#, detail: r#"Key columns "p" of the referencing table and "id" of the referenced table are of incompatible types: text and integer."#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Alter Column Type with USING failure leaves the table intact",
            set_up_script: &[
                "CREATE TABLE t1 (id INT PRIMARY KEY, value TEXT);",
                "CREATE INDEX t1_value_idx ON t1 (value);",
                "INSERT INTO t1 VALUES (1, '1'), (2, '2'), (3, '3'), (4, '4'), (5, '5'), (6, '6'), (7, '7'), (8, '8'), (9, '9'), (10, '10');",
                "INSERT INTO t1 SELECT id + 10, (id + 10)::text FROM t1;",
                "INSERT INTO t1 SELECT id + 20, (id + 20)::text FROM t1 WHERE id <= 20;",
                "INSERT INTO t1 SELECT id + 40, (id + 40)::text FROM t1 WHERE id <= 40;",
                "INSERT INTO t1 SELECT id + 80, (id + 80)::text FROM t1 WHERE id <= 20;",
                "INSERT INTO t1 VALUES (101, 'not-a-number');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN value TYPE integer USING value::integer;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "not-a-number""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(value) FROM t1 WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("text")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t1 WHERE value = '42';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t1 WHERE value = 'not-a-number';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (102, 'still-text');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("102")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("102")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(value) FROM t1 WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("text")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN value TYPE integer USING value::integer;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "not-a-number""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("102")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t1 WHERE value = '42';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(value) FROM t1 WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("text")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Alter Column Type with USING on a schema-qualified table",
            set_up_script: &[
                "CREATE SCHEMA s1;",
                "CREATE TABLE s1.t (id INT PRIMARY KEY, c TEXT);",
                "INSERT INTO s1.t VALUES (1, '7'), (2, '8');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE s1.t ALTER COLUMN c TYPE integer USING c::integer;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c FROM s1.t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("7")],
                            &[T("2"), T("8")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Alter Column Type with USING on keys and indexes",
            set_up_script: &[
                "CREATE TABLE t1 (id TEXT PRIMARY KEY, c INT);",
                "INSERT INTO t1 VALUES ('3', 30), ('1', 10), ('2', 20);",
                "CREATE TABLE t2 (id INT PRIMARY KEY, c TEXT);",
                "CREATE INDEX t2_c_idx ON t2 (c);",
                "INSERT INTO t2 VALUES (1, '100'), (2, '200');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN id TYPE integer USING id::integer;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, c FROM t1 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("c", INT4)],
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
                    query: "INSERT INTO t1 VALUES (1, 11);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "t1_pkey""#, detail: "Key (id)=(1) already exists.", schema: "public", table: "t1", constraint: "t1_pkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t2 ALTER COLUMN c TYPE integer USING c::integer;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t2 WHERE c = 200;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
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
        ScriptTest {
            name: "ALTER COLUMN resolves column default expressions",
            set_up_script: &[
                "CREATE TABLE t1 (id VARCHAR PRIMARY KEY, c1 TIMESTAMP DEFAULT CURRENT_TIMESTAMP);",
                "CREATE TABLE t2 (id VARCHAR PRIMARY KEY, c1 VARCHAR(100) DEFAULT concat('f', 'oo'));",
                "CREATE TABLE t3 (id VARCHAR PRIMARY KEY, c1 VARCHAR(20) NOT NULL DEFAULT CONCAT('f', 'oo'));",
                "CREATE TABLE t4 (id VARCHAR PRIMARY KEY, c1 VARCHAR(100) DEFAULT CONCAT('f', 'oo'));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 ALTER COLUMN c1 SET NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t2 ALTER COLUMN c1 TYPE VARCHAR(50);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t3 ALTER COLUMN c1 DROP NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t4 RENAME COLUMN c1 TO ccc1;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE ADD COLUMN with inline FK constraint",
            set_up_script: &[
                "create table t (v varchar(100));",
                "create table parent (id int primary key);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t ADD COLUMN c1 int REFERENCES parent(id);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname AS constraint_name, pg_get_constraintdef(oid) AS constraint_definition FROM pg_constraint WHERE conrelid = 't'::regclass AND contype='f';",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_name", NAME), Column("constraint_definition", TEXT)],
                        rows: &[
                            &[T("t_c1_fkey"), T("FOREIGN KEY (c1) REFERENCES parent(id)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('abc', 123);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "t" violates foreign key constraint "t_c1_fkey""#, detail: r#"Key (c1)=(123) is not present in table "parent"."#, schema: "public", table: "t", constraint: "t_c1_fkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Rename table",
            set_up_script: &[
                "create schema s1",
                "create schema s2",
                "CREATE TABLE t1 (a INT, b INT);",
                "INSERT INTO t1 VALUES (1, 2);",
                "CREATE TABLE t2 (c INT, d INT);",
                "INSERT INTO t2 VALUES (3, 4);",
                "create table s1.t1 (e INT, f INT);",
                "INSERT INTO s1.t1 VALUES (5, 6);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE doesnotexist RENAME TO t3;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "doesnotexist" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 RENAME TO t3;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3;",
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
                    query: "SELECT * FROM public.t3;",
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
                    query: "SELECT * FROM t1;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "t1" does not exist"#, position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t3 RENAME TO t2;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "t2" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE s1.t1 RENAME TO t4;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM s1.t4;",
                    expected: Expected::Rows {
                        columns: &[Column("e", INT4), Column("f", INT4)],
                        rows: &[
                            &[T("5"), T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Rename table with a foreign key",
            set_up_script: &[
                "CREATE TABLE bug8_parent (id integer PRIMARY KEY);",
                "CREATE TABLE bug8_child (id integer PRIMARY KEY, parent_id integer);",
                "ALTER TABLE bug8_child ADD CONSTRAINT bug8_fk FOREIGN KEY (parent_id) REFERENCES bug8_parent(id);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE bug8_child RENAME TO bug8_child2;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO bug8_parent VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO bug8_child2 VALUES (1, 1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO bug8_child2 VALUES (2, 2);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "bug8_child2" violates foreign key constraint "bug8_fk""#, detail: r#"Key (parent_id)=(2) is not present in table "bug8_parent"."#, schema: "public", table: "bug8_child2", constraint: "bug8_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE bug8_parent RENAME TO bug8_parent2;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO bug8_child2 VALUES (3, 3);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "bug8_child2" violates foreign key constraint "bug8_fk""#, detail: r#"Key (parent_id)=(3) is not present in table "bug8_parent2"."#, schema: "public", table: "bug8_child2", constraint: "bug8_fk", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Rename table must not collide with other relation types",
            set_up_script: &[
                "CREATE TABLE src_tbl (pk int PRIMARY KEY, v1 int);",
                "CREATE TABLE other_tbl (pk int PRIMARY KEY, v1 int);",
                "CREATE SEQUENCE seq1;",
                "CREATE VIEW view1 AS SELECT pk FROM src_tbl;",
                "CREATE INDEX idx1 ON src_tbl (v1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE src_tbl RENAME TO other_tbl;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "other_tbl" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE src_tbl RENAME TO seq1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "seq1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE src_tbl RENAME TO view1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "view1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE src_tbl RENAME TO idx1;",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "idx1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE src_tbl RENAME TO new_name;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM new_name;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter table owner",
            set_up_script: &[
                "CREATE TABLE t1 (a INT, b INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 OWNER TO new_owner;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "new_owner" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter table add primary key with timestamp column default values",
            set_up_script: &[
                r#"CREATE TABLE t1 (
    id int NOT NULL,
    uid uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);"#,
                "INSERT INTO t1 (id, uid) VALUES (1, '00000000-0000-0000-0000-000000000001');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE ONLY public.t1 ADD CONSTRAINT t1_pkey PRIMARY KEY (id);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select created_at is not null from t1 where id = 1;",
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
                    query: "select updated_at is not null from t1 where id = 1;",
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
                    query: "select created_at = updated_at from t1 where id = 1;",
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
            name: "alter table add primary key with uuid column default values",
            set_up_script: &[
                r#"CREATE TABLE t1 (
    id int NOT NULL,
    uid uuid default gen_random_uuid() NOT NULL
);"#,
                "INSERT INTO t1 (id) VALUES (1);",
                "INSERT INTO t1 (id) VALUES (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE ONLY public.t1 ADD CONSTRAINT t1_pkey PRIMARY KEY (id);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select uid is not null from t1 where id = 1;",
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
                    query: "select uid is not null from t1 where id = 2;",
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
                    query: "select (select uid from t1 where id = 2) = (select uid from t1 where id = 1);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter table drop primary key",
            set_up_script: &[
                "CREATE TABLE t1 (id int PRIMARY KEY);",
                "INSERT INTO t1 (id) VALUES (1), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t1 DROP CONSTRAINT t1_pkey;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1 VALUES (1), (2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE RENAME with table types",
            set_up_script: &[
                "CREATE TABLE t1a (a INT4, b VARCHAR(3));",
                "CREATE TABLE t1b (a VARCHAR(3), b INT4);",
                "CREATE TABLE t2 (id INT4, t1a t1a, t1b t1b);",
                "INSERT INTO t2 VALUES (1, ROW(1, 'abc'), ROW('abc', 1));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,abc)"), T("(abc,1)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a RENAME TO t1x;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (2, ROW(2, 'def'), ROW('def', 2));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,abc)"), T("(abc,1)")],
                            &[T("2"), T("(2,def)"), T("(def,2)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1x RENAME TO t1y;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (3, ROW(4, 'ghi'), ROW('kjl', 5));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(1,abc)"), T("(abc,1)")],
                            &[T("2"), T("(2,def)"), T("(def,2)")],
                            &[T("3"), T("(4,ghi)"), T("(kjl,5)")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE RENAME COLUMN with table types",
            set_up_script: &[
                "CREATE TABLE t1a (a INT4, b VARCHAR(3));",
                "CREATE TABLE t1b (a VARCHAR(3), b INT4);",
                "CREATE TABLE t2 (id INT4, t1a t1a, t1b t1b);",
                "INSERT INTO t2 VALUES (1, ROW(2, 'abc'), ROW('def', 3));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a RENAME COLUMN a TO x;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (t1a).a FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "a" not found in data type t1a"#, position: 9, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (t1a).x, (t1a).@1 FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "@""#, position: 23, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1b RENAME COLUMN b TO bb;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1b RENAME COLUMN a TO aa;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (4, ROW(5, 'ghi'), ROW('jkl', 6));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (t1b).aa, (t1b).@1, (t1b).bb, (t1b).@2 FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "@""#, position: 24, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE SET DEFAULT with table types",
            set_up_script: &[
                "CREATE TABLE t1a (a INT4, b VARCHAR(3));",
                "CREATE TABLE t1b (a VARCHAR(3), b INT4);",
                "CREATE TABLE t2 (id INT4, t1a t1a, t1b t1b);",
                "INSERT INTO t2 VALUES (1, ROW(2, 'abc'), ROW('def', 3));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ALTER COLUMN a SET DEFAULT 55;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1b ALTER COLUMN b SET DEFAULT 77, ALTER COLUMN a SET DEFAULT 'hi';",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (4, ROW(5, 'ghi'), ROW('kjl', 6));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                            &[T("4"), T("(5,ghi)"), T("(kjl,6)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE DROP DEFAULT with table types",
            set_up_script: &[
                "CREATE TABLE t1a (a INT4 DEFAULT 55, b VARCHAR(3) DEFAULT 'hi');",
                "CREATE TABLE t1b (a VARCHAR(5) DEFAULT 'hello', b INT4 DEFAULT 77);",
                "CREATE TABLE t2 (id INT4, t1a t1a, t1b t1b);",
                "INSERT INTO t2 VALUES (1, ROW(2, 'abc'), ROW('def', 3));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ALTER COLUMN a DROP DEFAULT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ALTER COLUMN b DROP DEFAULT, ALTER COLUMN a DROP DEFAULT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (4, ROW(5, 'ghi'), ROW('kjl', 6));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                            &[T("4"), T("(5,ghi)"), T("(kjl,6)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE SET DATA TYPE with table types",
            set_up_script: &[
                "CREATE TABLE t1a (a INT4, b VARCHAR(3));",
                "CREATE TABLE t1b (a VARCHAR(3), b INT4);",
                "CREATE TABLE t2 (id INT4, t1a t1a, t1b t1b);",
                "INSERT INTO t2 VALUES (1, ROW(2, 'abc'), ROW('def', 3));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ALTER COLUMN a SET DATA TYPE INT8;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cannot alter table "t1a" because column "t2.t1a" uses its row type"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ALTER COLUMN a SET DATA TYPE INT4;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"cannot alter table "t1a" because column "t2.t1a" uses its row type"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t2 DROP COLUMN t1a;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ALTER COLUMN a SET DATA TYPE INT8, ALTER COLUMN b SET DATA TYPE TEXT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(def,3)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE SET/DROP NOT NULL with table types",
            set_up_script: &[
                "CREATE TABLE t1a (a INT4, b VARCHAR(3));",
                "CREATE TABLE t1b (a VARCHAR(3), b INT4);",
                "CREATE TABLE t2 (id INT4, t1a t1a, t1b t1b);",
                "INSERT INTO t2 VALUES (1, ROW(2, 'abc'), ROW('def', 3));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1a ALTER COLUMN a SET NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1b ALTER COLUMN b SET NOT NULL, ALTER COLUMN a SET NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t1a VALUES (NULL, 'hi');",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "a" of relation "t1a" violates not-null constraint"#, detail: "Failing row contains (null, hi).", schema: "public", table: "t1a", column: "a", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (4, ROW(NULL, 'ghi'), ROW(NULL, 6));",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                            &[T("4"), T("(,ghi)"), T("(,6)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t1b ALTER COLUMN b DROP NOT NULL, ALTER COLUMN a DROP NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("t1a", USER_DEFINED), Column("t1b", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(2,abc)"), T("(def,3)")],
                            &[T("4"), T("(,ghi)"), T("(,6)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE RENAME on view",
            set_up_script: &[
                r#"CREATE TABLE tenk1 (
	unique1		int4,
	unique2		int4,
	two			int4,
	four		int4,
	ten			int4,
	twenty		int4,
	hundred		int4,
	thousand	int4,
	twothousand	int4,
	fivethous	int4,
	tenthous	int4,
	odd			int4,
	even		int4,
	stringu1	name,
	stringu2	name,
	string4		name);"#,
                "CREATE VIEW attmp_view (unique1) AS SELECT unique1 FROM tenk1;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp_view RENAME TO attmp_view_new;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE IF EXISTS",
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE IF EXISTS t1a ALTER COLUMN a SET NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"relation "t1a" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE with schema defined that is not the current schema",
            set_up_script: &[
                "CREATE SCHEMA grassroots;",
                r#"CREATE TYPE grassroots.user_role AS ENUM (
					'ADMIN',
					'USER'
				);"#,
                r#"CREATE TABLE grassroots.users (
					id uuid DEFAULT gen_random_uuid() NOT NULL,
					email text NOT NULL,
					password_hash text NOT NULL,
					first_name text,
					last_name text,
					role grassroots.user_role DEFAULT 'USER'::grassroots.user_role NOT NULL
				);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE ONLY grassroots.users ADD CONSTRAINT users_email_key UNIQUE (email);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "setting foreign key to SET NULL on NOT NULL defined column",
            set_up_script: &[
                r#"CREATE TABLE public.products (
		   product_id integer NOT NULL,
		   product_name character varying(100) NOT NULL,
		   category_id integer NOT NULL,
		   price numeric(10,2) NOT NULL,
		   description text
		);"#,
                r#"INSERT INTO public.products VALUES
		                               (13, 'Smartphone', 1, 599.99, 'Latest model with advanced features'),
		                               (14, 'Laptop', 1, 999.99, 'High performance laptop with 16GB RAM'),
		                               (18, 'Novel', 2, 19.99, 'Bestselling fiction novel');"#,
                r#"CREATE TABLE public.categories (
		   category_id integer NOT NULL,
		   category_name character varying(50) NOT NULL
		);"#,
                "INSERT INTO public.categories VALUES (1, 'Electronics'), (2, 'Books'), (3, 'Clothing');",
                r#"ALTER TABLE ONLY public.products
		   ADD CONSTRAINT products_pkey PRIMARY KEY (product_id);"#,
                r#"ALTER TABLE ONLY public.categories
		   ADD CONSTRAINT categories_pkey PRIMARY KEY (category_id);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"ALTER TABLE ONLY public.products
		   ADD CONSTRAINT fk_category_id FOREIGN KEY (category_id) REFERENCES public.categories(category_id) ON UPDATE SET NULL ON DELETE SET NULL;"#,
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM categories WHERE category_id = 1;",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "category_id" of relation "products" violates not-null constraint"#, detail: "Failing row contains (13, Smartphone, null, 599.99, Latest model with advanced features).", schema: "public", table: "products", column: "category_id", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "use double quoted column name for SET DEFAULT",
            set_up_script: &[
                r#" CREATE TABLE pages_links (
		"idRefferer" bigint NOT NULL,
		"idDestination" bigint NOT NULL
	);"#,
                r#"CREATE SEQUENCE pages_links_iddestination_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;"#,
                r#"ALTER SEQUENCE pages_links_iddestination_seq OWNED BY pages_links."idDestination";"#,
                r#"CREATE SEQUENCE pages_links_idrefferer_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;"#,
                r#"ALTER SEQUENCE pages_links_idrefferer_seq OWNED BY pages_links."idRefferer";"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"ALTER TABLE ONLY pages_links ALTER COLUMN "idRefferer" SET DEFAULT nextval('pages_links_idrefferer_seq'::regclass);"#,
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key names can be the same but should be on different tables",
            set_up_script: &[
                r#" CREATE TABLE public.boards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    title character varying(255),
    project_id uuid
);"#,
                r#"CREATE TABLE public.project_assignments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    role character varying(20) DEFAULT 'viewer'::character varying,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL
);"#,
                r#"CREATE TABLE public.projects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(100) NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    owner_id uuid
);"#,
                r#"ALTER TABLE ONLY public.boards
    ADD CONSTRAINT boards_pkey PRIMARY KEY (id);"#,
                r#"ALTER TABLE ONLY public.project_assignments
    ADD CONSTRAINT project_assignments_pkey PRIMARY KEY (id);"#,
                r#"ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);"#,
                r#"ALTER TABLE ONLY public.project_assignments
    ADD CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"ALTER TABLE ONLY public.boards
    ADD CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;"#,
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE with NOT VALID clauses on foreign key constraint",
            set_up_script: &[
                "CREATE TABLE attmp2 (a int primary key);",
                "CREATE TABLE attmp3 (a int, b int);",
                "INSERT INTO attmp2 values (1),(2),(3),(4);",
                "INSERT INTO attmp3 values (1,10),(1,20),(3, 22),(5,50);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 add constraint attmpconstr foreign key(c) references attmp2(a);",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "c" referenced in foreign key constraint does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 add constraint attmpconstr foreign key(a) references attmp2(b);",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "b" referenced in foreign key constraint does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 add constraint attmpconstr foreign key (a) references attmp2(a);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "attmp3" violates foreign key constraint "attmpconstr""#, detail: r#"Key (a)=(5) is not present in table "attmp2"."#, schema: "public", table: "attmp3", constraint: "attmpconstr", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 add constraint attmpconstr foreign key (a) references attmp2 NOT VALID;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 add constraint attmpconstr foreign key (a) references attmp2 (a) NOT VALID;",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"constraint "attmpconstr" for relation "attmp3" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 validate constraint attmpconstr;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "attmp3" violates foreign key constraint "attmpconstr""#, detail: r#"Key (a)=(5) is not present in table "attmp2"."#, schema: "public", table: "attmp3", constraint: "attmpconstr", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM attmp3 where a=5;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 validate constraint attmpconstr;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 validate constraint attmpconstr;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO attmp3 VALUES (6, 5);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "attmp3" violates foreign key constraint "attmpconstr""#, detail: r#"Key (a)=(6) is not present in table "attmp2"."#, schema: "public", table: "attmp3", constraint: "attmpconstr", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO attmp2 VALUES (6);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO attmp3 VALUES (6, 5);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE attmp3 SET a=7 where b=22;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "attmp3" violates foreign key constraint "attmpconstr""#, detail: r#"Key (a)=(7) is not present in table "attmp2"."#, schema: "public", table: "attmp3", constraint: "attmpconstr", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE attmp3 SET a=2 where b=22;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT 
    con.conname AS constraint_name,
    cl_child.relname AS child_table,
    (SELECT array_agg(attname) FROM pg_attribute WHERE attrelid = con.conrelid AND attnum = ANY(con.conkey)) AS child_columns,
    cl_parent.relname AS parent_table,
    (SELECT array_agg(attname) FROM pg_attribute WHERE attrelid = con.confrelid AND attnum = ANY(con.confkey)) AS parent_columns
FROM pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class cl_child ON con.conrelid = cl_child.oid
JOIN pg_catalog.pg_class cl_parent ON con.confrelid = cl_parent.oid
WHERE con.contype = 'f';"#,
                    expected: Expected::Rows {
                        columns: &[Column("constraint_name", NAME), Column("child_table", NAME), Column("child_columns", NAME_ARRAY), Column("parent_table", NAME), Column("parent_columns", NAME_ARRAY)],
                        rows: &[
                            &[T("attmpconstr"), T("attmp3"), T("{a}"), T("attmp2"), T("{a}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE with NOT VALID clauses on check constraint",
            set_up_script: &[
                "CREATE TABLE attmp3 (a int, b int);",
                "INSERT INTO attmp3 values (1,10),(1,20),(3,22);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 ADD CONSTRAINT b_greater_than_ten CHECK (b > 10);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"check constraint "b_greater_than_ten" of relation "attmp3" is violated by some row"#, schema: "public", table: "attmp3", constraint: "b_greater_than_ten", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 ADD CONSTRAINT b_greater_than_ten CHECK (b > 10) NOT VALID;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 VALIDATE CONSTRAINT b_greater_than_ten;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"check constraint "b_greater_than_ten" of relation "attmp3" is violated by some row"#, schema: "public", table: "attmp3", constraint: "b_greater_than_ten", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM attmp3 WHERE NOT b > 10;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 VALIDATE CONSTRAINT b_greater_than_ten;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE attmp3 VALIDATE CONSTRAINT b_greater_than_ten;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO attmp3 VALUES (5, 9);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "attmp3" violates check constraint "b_greater_than_ten""#, detail: "Failing row contains (5, 9).", schema: "public", table: "attmp3", constraint: "b_greater_than_ten", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO attmp3 VALUES (6, 11);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE attmp3 SET b=7 where a=3;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "attmp3" violates check constraint "b_greater_than_ten""#, detail: "Failing row contains (3, 7).", schema: "public", table: "attmp3", constraint: "b_greater_than_ten", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE attmp3 SET b=77 where a=3;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT                                                              
    ns.nspname AS schema_name,
    cl.relname AS table_name,
    con.conname AS constraint_name,
    pg_get_constraintdef(con.oid) AS constraint_definition
FROM pg_catalog.pg_constraint con
JOIN pg_catalog.pg_class cl ON con.conrelid = cl.oid
JOIN pg_catalog.pg_namespace ns ON cl.relnamespace = ns.oid
WHERE con.contype = 'c'
ORDER BY schema_name, table_name;"#,
                    expected: Expected::Rows {
                        columns: &[Column("schema_name", NAME), Column("table_name", NAME), Column("constraint_name", NAME), Column("constraint_definition", TEXT)],
                        rows: &[
                            &[T("public"), T("attmp3"), T("b_greater_than_ten"), T("CHECK ((b > 10))")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALIDATE CONSTRAINT with the table's schema not on the search path",
            set_up_script: &[
                "CREATE TABLE public.vc (a int, b int);",
                "INSERT INTO public.vc VALUES (1, 10);",
                "ALTER TABLE public.vc ADD CONSTRAINT b_gt_ten CHECK (b > 10) NOT VALID;",
                "SELECT pg_catalog.set_config('search_path', '', false);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE public.vc VALIDATE CONSTRAINT b_gt_ten;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"check constraint "b_gt_ten" of relation "vc" is violated by some row"#, schema: "public", table: "vc", constraint: "b_gt_ten", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "VALIDATE CONSTRAINT with a same-named table earlier on the search path",
            set_up_script: &[
                "CREATE SCHEMA s2;",
                "CREATE TABLE public.vc (a int, b int);",
                "INSERT INTO public.vc VALUES (11, 5);",
                "ALTER TABLE public.vc ADD CONSTRAINT b_gt_ten CHECK (b > 10) NOT VALID;",
                "CREATE TABLE s2.vc (b int, a int);",
                "SET search_path TO s2, public;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE public.vc VALIDATE CONSTRAINT b_gt_ten;",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"check constraint "b_gt_ten" of relation "vc" is violated by some row"#, schema: "public", table: "vc", constraint: "b_gt_ten", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "duplicate key handling after ADD COLUMN with an expression index",
            set_up_script: &[
                "CREATE TABLE expression_index_alter (id int PRIMARY KEY, name text UNIQUE);",
                "CREATE INDEX expression_index_alter_lower_name ON expression_index_alter ((lower(name)));",
                "ALTER TABLE expression_index_alter ADD COLUMN extra timestamptz;",
                "INSERT INTO expression_index_alter (id, name) VALUES (1, 'v1');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO expression_index_alter (id, name) VALUES (1, 'duplicate');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "expression_index_alter_pkey""#, detail: "Key (id)=(1) already exists.", schema: "public", table: "expression_index_alter", constraint: "expression_index_alter_pkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO expression_index_alter (id, name) VALUES (2, 'v1');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "expression_index_alter_name_key""#, detail: "Key (name)=(v1) already exists.", schema: "public", table: "expression_index_alter", constraint: "expression_index_alter_name_key", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO expression_index_alter (id, name) VALUES (1, 'v2') ON CONFLICT (id) DO UPDATE SET name = 'v2';",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO expression_index_alter (id, name) VALUES (2, 'v2');",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "expression_index_alter_name_key""#, detail: "Key (name)=(v2) already exists.", schema: "public", table: "expression_index_alter", constraint: "expression_index_alter_name_key", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO expression_index_alter (id, name) VALUES (1, 'ignored') ON CONFLICT DO NOTHING;",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO expression_index_alter (id, name) VALUES (1, 'ignored') ON CONFLICT (id) DO NOTHING;",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, name, extra FROM expression_index_alter;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("extra", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1"), T("v2"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ALTER TABLE with MATCH FULL on foreign key",
            set_up_script: &[
                "CREATE TABLE parent_table (parent_id INT, sub_id INT, name TEXT, PRIMARY KEY (parent_id, sub_id));",
                "CREATE TABLE child_table (child_id INT PRIMARY KEY, parent_id INT, sub_id INT, description TEXT);",
                "INSERT INTO parent_table (parent_id, sub_id, name) VALUES (1, 10, 'Parent Alpha'), (2, 20, 'Parent Beta'), (2, 40, 'Parent 1');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE child_table add constraint constr FOREIGN KEY (parent_id, sub_id) REFERENCES parent_table(parent_id, sub_id) MATCH FULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_table (child_id, parent_id, sub_id, description) VALUES (101, 1, 10, 'Valid reference');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_table (child_id, parent_id, sub_id, description) VALUES (102, NULL, NULL, 'Completely unlinked row');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_table (child_id, parent_id, sub_id, description) VALUES (105, 2, 30, NULL);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child_table" violates foreign key constraint "constr""#, detail: r#"Key (parent_id, sub_id)=(2, 30) is not present in table "parent_table"."#, schema: "public", table: "child_table", constraint: "constr", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_table (child_id, parent_id, sub_id, description) VALUES (105, 2, 20, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child_table (child_id, parent_id, sub_id, description) VALUES (103, 1, NULL, 'Partial null mix');",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child_table" violates foreign key constraint "constr""#, detail: "MATCH FULL does not allow mixing of null and nonnull key values.", schema: "public", table: "child_table", constraint: "constr", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE child_table SET sub_id = NULL where parent_id = 2;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child_table" violates foreign key constraint "constr""#, detail: "MATCH FULL does not allow mixing of null and nonnull key values.", schema: "public", table: "child_table", constraint: "constr", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE child_table SET sub_id = 40 where parent_id = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ADD COLUMN IF NOT EXISTS",
            set_up_script: &[
                "CREATE TABLE t7 (a INT);",
                "INSERT INTO t7 VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t7 ADD COLUMN a INT;",
                    expected: Expected::Error(Diagnostic { code: "42701", message: r#"column "a" of relation "t7" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t7 ADD COLUMN IF NOT EXISTS a INT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "42701", message: r#"column "a" of relation "t7" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t7 ADD COLUMN IF NOT EXISTS b INT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t7 ADD COLUMN IF NOT EXISTS a TEXT DEFAULT 'x' UNIQUE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "42701", message: r#"column "a" of relation "t7" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t7 ADD IF NOT EXISTS c INT, ADD COLUMN IF NOT EXISTS a INT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "42701", message: r#"column "a" of relation "t7" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t7;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 't7' ORDER BY ordinal_position;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("data_type", VARCHAR)],
                        rows: &[
                            &[T("a"), T("integer")],
                            &[T("b"), T("integer")],
                            &[T("c"), T("integer")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indexname FROM pg_indexes WHERE tablename = 't7';",
                    expected: Expected::Rows {
                        columns: &[Column("indexname", NAME)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated column with COALESCE keeps working after ADD PRIMARY KEY",
            set_up_script: &[
                "CREATE TABLE t (a int NOT NULL, b int GENERATED ALWAYS AS (COALESCE(a + 1, 0)) STORED);",
                "INSERT INTO t (a) VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t ADD PRIMARY KEY (a);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t (a) VALUES (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
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
            name: "generated columns with function calls keep working after other alterations",
            set_up_script: &[
                "CREATE TABLE t1 (a INT PRIMARY KEY, b INT GENERATED ALWAYS AS (a + 1) STORED);",
                "INSERT INTO t1 (a) VALUES (1);",
                "ALTER TABLE t1 ADD COLUMN c INT DEFAULT 0;",
                "CREATE TABLE t2 (a INT PRIMARY KEY, b TEXT GENERATED ALWAYS AS (upper(a::text)) STORED);",
                "INSERT INTO t2 (a) VALUES (1);",
                "ALTER TABLE t2 ADD COLUMN c INT DEFAULT 0;",
                "CREATE TABLE t3 (a INT NOT NULL, b INT GENERATED ALWAYS AS (COALESCE(abs(a), 0)) STORED, c INT);",
                "INSERT INTO t3 (a) VALUES (1);",
                "ALTER TABLE t3 ADD UNIQUE (a);",
                "INSERT INTO t3 (a) VALUES (2);",
                "ALTER TABLE t3 DROP COLUMN c;",
                "INSERT INTO t3 (a) VALUES (3);",
                "UPDATE t3 SET a = 4 WHERE a = 3;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t1 (a) VALUES (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t1 ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("0")],
                            &[T("2"), T("3"), T("0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 (a) VALUES (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2 ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("0")],
                            &[T("2"), T("2"), T("0")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3 (a) VALUES (5);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3 ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("4"), T("4")],
                            &[T("5"), T("5")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DEFERRABLE constraints",
            set_up_script: &[
                "CREATE TABLE p (id INTEGER PRIMARY KEY);",
                "CREATE TABLE c (x INTEGER);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT c_fk FOREIGN KEY (x) REFERENCES p(id) DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT c_u UNIQUE (x) DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT c_pk PRIMARY KEY (x) INITIALLY DEFERRED;",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT c_chk CHECK (x > 0) DEFERRABLE;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "CHECK constraints cannot be marked DEFERRABLE", position: 50, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD COLUMN y INTEGER REFERENCES p(id) DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD COLUMN z INTEGER UNIQUE DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD COLUMN w INTEGER CHECK (w > 0) DEFERRABLE;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "misplaced DEFERRABLE clause", position: 50, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT c_fk2 FOREIGN KEY (x) REFERENCES p(id) NOT DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DEFERRABLE constraints are stored",
            set_up_script: &[
                "CREATE TABLE p (id INTEGER PRIMARY KEY);",
                "CREATE TABLE c (x INTEGER);",
                "CREATE TABLE c2 (x INTEGER);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT c_fk FOREIGN KEY (x) REFERENCES p(id) DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT c_u UNIQUE (x) DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD COLUMN y INTEGER REFERENCES p(id) DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD COLUMN z INTEGER UNIQUE DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c2 ADD CONSTRAINT c2_pk PRIMARY KEY (x) INITIALLY DEFERRED;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT c_fk2 FOREIGN KEY (x) REFERENCES p(id) NOT DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ALTER CONSTRAINT c_fk2 DEFERRABLE INITIALLY DEFERRED;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, condeferrable, condeferred FROM pg_constraint WHERE conname = 'c_fk2';",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("condeferrable", BOOL), Column("condeferred", BOOL)],
                        rows: &[
                            &[T("c_fk2"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ALTER CONSTRAINT c_fk2 NOT DEFERRABLE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, condeferrable, condeferred FROM pg_constraint WHERE conname = 'c_fk2';",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("condeferrable", BOOL), Column("condeferred", BOOL)],
                        rows: &[
                            &[T("c_fk2"), T("f"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ALTER CONSTRAINT c_u DEFERRABLE;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#"constraint "c_u" of relation "c" is not a foreign key constraint"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, condeferrable, condeferred FROM pg_constraint WHERE connamespace = 'public'::regnamespace AND condeferrable ORDER BY conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("condeferrable", BOOL), Column("condeferred", BOOL)],
                        rows: &[
                            &[T("c2_pk"), T("t"), T("t")],
                            &[T("c_fk"), T("t"), T("f")],
                            &[T("c_u"), T("t"), T("f")],
                            &[T("c_y_fkey"), T("t"), T("f")],
                            &[T("c_z_key"), T("t"), T("f")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "column defaults with user-defined functions keep working after table rewrites",
            set_up_script: &[
                "CREATE FUNCTION f() RETURNS INT LANGUAGE SQL AS $$ SELECT 1 $$;",
                "CREATE TABLE t (id INT, a INT DEFAULT f());",
                "INSERT INTO t (id) VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE t ALTER COLUMN a SET NOT NULL;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t ALTER COLUMN a TYPE BIGINT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t ADD COLUMN b INT DEFAULT 2;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t DROP COLUMN b;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t ADD PRIMARY KEY (id);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t (id) VALUES (2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "NOT NULL constraint names",
            set_up_script: &[
                "CREATE TABLE nn1 (id INT PRIMARY KEY, a INT NOT NULL);",
                "CREATE TABLE nn2 (id INT CONSTRAINT keep NOT NULL, PRIMARY KEY (id));",
                "CREATE TABLE nn3 (LIKE nn1);",
                "ALTER TABLE nn1 RENAME COLUMN a TO b;",
                "ALTER TABLE nn1 RENAME TO nn4;",
                "CREATE TABLE nn1 (x INT);",
                "ALTER TABLE nn1 ADD CONSTRAINT xx NOT NULL x;",
                "ALTER TABLE nn1 ALTER x DROP NOT NULL;",
                "ALTER TABLE nn1 ALTER x SET NOT NULL;",
                "CREATE TABLE nn5 (a INT NOT NULL, b INT, c INT PRIMARY KEY);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT conrelid::regclass, conname, contype FROM pg_constraint WHERE conrelid IN ('nn1'::regclass, 'nn2'::regclass, 'nn3'::regclass, 'nn4'::regclass) ORDER BY conrelid::regclass::text, conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conrelid", REGCLASS), Column("conname", NAME), Column("contype", CHAR)],
                        rows: &[
                            &[T("nn1"), T("nn1_x_not_null"), T("n")],
                            &[T("nn2"), T("keep"), T("n")],
                            &[T("nn2"), T("nn2_pkey"), T("p")],
                            &[T("nn3"), T("nn1_a_not_null"), T("n")],
                            &[T("nn3"), T("nn1_id_not_null"), T("n")],
                            &[T("nn4"), T("nn1_a_not_null"), T("n")],
                            &[T("nn4"), T("nn1_id_not_null"), T("n")],
                            &[T("nn4"), T("nn1_pkey"), T("p")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE nn5 ADD CONSTRAINT other NOT NULL a;",
                    expected: Expected::Error(Diagnostic { code: "55000", message: r#"cannot create not-null constraint "other" on column "a" of table "nn5""#, detail: r#"A not-null constraint named "nn5_a_not_null" already exists for this column."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE nn5 ADD CONSTRAINT bnn NOT NULL b;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE nn5 RENAME CONSTRAINT bnn TO bnn2;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_constraint WHERE conrelid = 'nn5'::regclass AND contype = 'n' ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[
                            &[T("bnn2")],
                            &[T("nn5_a_not_null")],
                            &[T("nn5_c_not_null")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE nn5 DROP CONSTRAINT nn5_c_not_null;",
                    expected: Expected::Error(Diagnostic { code: "42P16", message: r#"column "c" is in a primary key"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE nn5 DROP CONSTRAINT bnn2;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT attname, attnotnull FROM pg_attribute WHERE attrelid = 'nn5'::regclass AND attnum > 0 ORDER BY attnum;",
                    expected: Expected::Rows {
                        columns: &[Column("attname", NAME), Column("attnotnull", BOOL)],
                        rows: &[
                            &[T("a"), T("t")],
                            &[T("b"), T("f")],
                            &[T("c"), T("t")],
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
