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
fn test_foreign_keys() {
    run_scripts(&[
        ScriptTest {
            name: "simple foreign key",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY, b int)",
                "CREATE TABLE child (a INT PRIMARY KEY, b INT, FOREIGN KEY (b) REFERENCES parent(a))",
                "INSERT INTO parent VALUES (1, 1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (1, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 2)",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "child_pkey""#, detail: "Key (a)=(2) already exists.", schema: "public", table: "child", constraint: "child_pkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "named constraint",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY, b int)",
                "CREATE TABLE child (a INT PRIMARY KEY, b INT)",
                "INSERT INTO parent VALUES (1, 1)",
                "ALTER TABLE child ADD CONSTRAINT fk123 FOREIGN KEY (b) REFERENCES parent(a)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (1, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 2)",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "child_pkey""#, detail: "Key (a)=(2) already exists.", schema: "public", table: "child", constraint: "child_pkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "unnamed constraint",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY, b int)",
                "CREATE TABLE child (a INT PRIMARY KEY, b INT)",
                "INSERT INTO parent VALUES (1, 1)",
                "ALTER TABLE child ADD FOREIGN KEY (b) REFERENCES parent(a)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (1, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 2)",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "child_pkey""#, detail: "Key (a)=(2) already exists.", schema: "public", table: "child", constraint: "child_pkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "inline column REFERENCES enforces a foreign key constraint",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY)",
                "CREATE TABLE child (id INT PRIMARY KEY, pid INT REFERENCES parent(a))",
                "INSERT INTO parent VALUES (1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (1, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 999)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_pid_fkey""#, detail: r#"Key (pid)=(999) is not present in table "parent"."#, schema: "public", table: "child", constraint: "child_pid_fkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "inline column REFERENCES enforces constraints across a chain of related tables",
            set_up_script: &[
                "CREATE TABLE regions (id SERIAL UNIQUE NOT NULL, code VARCHAR(4) UNIQUE NOT NULL, name TEXT UNIQUE NOT NULL)",
                "CREATE TABLE departments (id SERIAL UNIQUE NOT NULL, code VARCHAR(4) UNIQUE NOT NULL, region VARCHAR(4) NOT NULL REFERENCES regions(code), name TEXT UNIQUE NOT NULL)",
                "CREATE TABLE towns (id SERIAL UNIQUE NOT NULL, code VARCHAR(10) NOT NULL, name TEXT NOT NULL, department VARCHAR(4) NOT NULL REFERENCES departments(code), UNIQUE (code, department))",
                "INSERT INTO regions VALUES (1, '01', 'Region1')",
                "INSERT INTO departments VALUES (1, 'D01', '01', 'Department1')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO towns VALUES (1, 'T01', 'Town1', 'D01')",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO towns VALUES (2, 'T02', 'Town2', 'NOPE')",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "towns" violates foreign key constraint "towns_department_fkey""#, detail: r#"Key (department)=(NOPE) is not present in table "departments"."#, schema: "public", table: "towns", constraint: "towns_department_fkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "text foreign key",
            set_up_script: &[
                "CREATE TABLE parent (a text PRIMARY KEY, b int)",
                "CREATE TABLE child (a INT PRIMARY KEY, b text, FOREIGN KEY (b) REFERENCES parent(a))",
                "INSERT INTO parent VALUES ('a', 1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (1, 'a')",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 'a')",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (3, 'b')",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_b_fkey""#, detail: r#"Key (b)=(b) is not present in table "parent"."#, schema: "public", table: "child", constraint: "child_b_fkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "type compatibility",
            set_up_script: &[
                "create table parent (i2 int2, i4 int4, i8 int8, f float, d double precision, v varchar, vl varchar(100), t text, j json, ts timestamp);",
                "alter table parent add constraint u1 unique (i2);",
                "alter table parent add constraint u2 unique (i4);",
                "alter table parent add constraint u3 unique (i8);",
                "alter table parent add constraint u4 unique (d);",
                "alter table parent add constraint u5 unique (f);",
                "alter table parent add constraint u6 unique (v);",
                "alter table parent add constraint u7 unique (vl);",
                "alter table parent add constraint u8 unique (t);",
                "alter table parent add constraint u9 unique (ts);",
                "create table child (i2 int2, i4 int4, i8 int8, f float, d double precision, v varchar, vl varchar(100), t text, j json, ts timestamp);",
                r#"insert into parent values (1, 1, 1, 1.0, 1.0, 'a', 'a', 'a', '{"a": 1}', '2021-01-01 00:00:00');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2i2 foreign key (i2) references parent(i2)",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2i4 foreign key (i2) references parent(i4)",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2i8 foreign key (i2) references parent(i8);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2f foreign key (i2) references parent(f);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2d foreign key (i2) references parent(d);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2v foreign key (i2) references parent(v);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fi2v" cannot be implemented"#, detail: r#"Key columns "i2" of the referencing table and "v" of the referenced table are of incompatible types: smallint and character varying."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2vl foreign key (i2) references parent(vl);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fi2vl" cannot be implemented"#, detail: r#"Key columns "i2" of the referencing table and "vl" of the referenced table are of incompatible types: smallint and character varying."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2t foreign key (i2) references parent(t);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fi2t" cannot be implemented"#, detail: r#"Key columns "i2" of the referencing table and "t" of the referenced table are of incompatible types: smallint and text."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi2ts foreign key (i2) references parent(ts);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fi2ts" cannot be implemented"#, detail: r#"Key columns "i2" of the referencing table and "ts" of the referenced table are of incompatible types: smallint and timestamp without time zone."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi4i2 foreign key (i4) references parent(i2);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi4i4 foreign key (i4) references parent(i4);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi4i8 foreign key (i4) references parent(i8);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi4f foreign key (i4) references parent(f);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi8i2 foreign key (i8) references parent(i2);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi8i4 foreign key (i8) references parent(i4);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi8d foreign key (i8) references parent(d);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fi8t foreign key (i8) references parent(t);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fi8t" cannot be implemented"#, detail: r#"Key columns "i8" of the referencing table and "t" of the referenced table are of incompatible types: bigint and text."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ffi2 foreign key (f) references parent(i2);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "ffi2" cannot be implemented"#, detail: r#"Key columns "f" of the referencing table and "i2" of the referenced table are of incompatible types: double precision and smallint."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ffi4 foreign key (f) references parent(i4);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "ffi4" cannot be implemented"#, detail: r#"Key columns "f" of the referencing table and "i4" of the referenced table are of incompatible types: double precision and integer."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ffi8 foreign key (f) references parent(i8);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "ffi8" cannot be implemented"#, detail: r#"Key columns "f" of the referencing table and "i8" of the referenced table are of incompatible types: double precision and bigint."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ffd foreign key (f) references parent(d);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fdf foreign key (d) references parent(f);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fft foreign key (f) references parent(t);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fft" cannot be implemented"#, detail: r#"Key columns "f" of the referencing table and "t" of the referenced table are of incompatible types: double precision and text."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ffv foreign key (f) references parent(v);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "ffv" cannot be implemented"#, detail: r#"Key columns "f" of the referencing table and "v" of the referenced table are of incompatible types: double precision and character varying."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvv foreign key (v) references parent(v);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvvl foreign key (v) references parent(vl);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvi8 foreign key (v) references parent(i8);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fvi8" cannot be implemented"#, detail: r#"Key columns "v" of the referencing table and "i8" of the referenced table are of incompatible types: character varying and bigint."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvf foreign key (v) references parent(f);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fvf" cannot be implemented"#, detail: r#"Key columns "v" of the referencing table and "f" of the referenced table are of incompatible types: character varying and double precision."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvts foreign key (v) references parent(ts);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fvts" cannot be implemented"#, detail: r#"Key columns "v" of the referencing table and "ts" of the referenced table are of incompatible types: character varying and timestamp without time zone."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvj foreign key (v) references parent(j);",
                    expected: Expected::Error(Diagnostic { code: "42830", message: r#"there is no unique constraint matching given keys for referenced table "parent""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvt foreign key (v) references parent(t);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvllv foreign key (vl) references parent(vl);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvlv foreign key (vl) references parent(v);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fvlt foreign key (vl) references parent(t);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ftt foreign key (t) references parent(t);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ftv foreign key (t) references parent(v);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ftvl foreign key (t) references parent(vl);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint fti8 foreign key (t) references parent(i8);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "fti8" cannot be implemented"#, detail: r#"Key columns "t" of the referencing table and "i8" of the referenced table are of incompatible types: text and bigint."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ftsts foreign key (ts) references parent(ts);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ftst foreign key (ts) references parent(t);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "ftst" cannot be implemented"#, detail: r#"Key columns "ts" of the referencing table and "t" of the referenced table are of incompatible types: timestamp without time zone and text."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "alter table child add constraint ftsi8 foreign key (ts) references parent(i8);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "ftsi8" cannot be implemented"#, detail: r#"Key columns "ts" of the referencing table and "i8" of the referenced table are of incompatible types: timestamp without time zone and bigint."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"insert into child values (1, 1, 1, 1.0, 1.0, 'a', 'a', 'a', '{"a": 1}', '2021-01-01 00:00:00');"#,
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"insert into child values (1, 2, 1, 1.0, 1.0, 'a', 'a', 'a', '{"a": 1}', '2021-01-01 00:00:00');"#,
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "fi4i2""#, detail: r#"Key (i4)=(2) is not present in table "parent"."#, schema: "public", table: "child", constraint: "fi4i2", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"insert into child values (1, 1, 1, 2.0, 1.0, 'a', 'a', 'a', '{"a": 1}', '2021-01-01 00:00:00');"#,
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "ffd""#, detail: r#"Key (f)=(2) is not present in table "parent"."#, schema: "public", table: "child", constraint: "ffd", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"insert into child values (1, 1, 1, 1.0, 1.0, 'a', 'a', 'b', '{"a": 1}', '2021-01-01 00:00:00');"#,
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "ftt""#, detail: r#"Key (t)=(b) is not present in table "parent"."#, schema: "public", table: "child", constraint: "ftt", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"insert into child values (1, 1, 1, 1.0, 1.0, 'a', 'a', 'a', '{"a": 1}', '2021-01-01 00:00:01');"#,
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "ftsts""#, detail: r#"Key (ts)=(2021-01-01 00:00:01) is not present in table "parent"."#, schema: "public", table: "child", constraint: "ftsts", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "type conversion: text to varchar",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY, b varchar(100))",
                "CREATE TABLE child (c INT PRIMARY KEY, d text)",
                "INSERT INTO parent VALUES (1, 'abc'), (2, 'def')",
                "alter table parent add constraint ub unique (b)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "alter table child add constraint fk foreign key (d) references parent(b)",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (1, 'abc')",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (2, 'xyz')",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "fk""#, detail: r#"Key (d)=(xyz) is not present in table "parent"."#, schema: "public", table: "child", constraint: "fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from parent where b = 'def'",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from parent where b = 'abc'",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "parent" violates foreign key constraint "fk" on table "child""#, detail: r#"Key (b)=(abc) is still referenced from table "child"."#, schema: "public", table: "child", constraint: "fk", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "type conversion: integer to double",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY, b double precision)",
                "CREATE TABLE child (c INT PRIMARY KEY, d int)",
                "INSERT INTO parent VALUES (1, 1), (3, 3)",
                "alter table parent add constraint ub unique (b)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "alter table child add constraint fk foreign key (d) references parent(b)",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from parent where b = 1.0",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (1, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (2, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (2, 2)",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "child_pkey""#, detail: "Key (c)=(2) already exists.", schema: "public", table: "child", constraint: "child_pkey", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from parent where b = 3.0",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from parent where b = 1.0",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "parent" violates foreign key constraint "fk" on table "child""#, detail: r#"Key (b)=(1) is still referenced from table "child"."#, schema: "public", table: "child", constraint: "fk", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "type conversion: value out of bounds, child larger",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY, b int2)",
                "CREATE TABLE child (c INT PRIMARY KEY, d int8)",
                "INSERT INTO parent VALUES (1, 1), (3, 3)",
                "alter table parent add constraint ub unique (b)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "alter table child add constraint fk foreign key (d) references parent(b)",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (1, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (2, 2)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "fk""#, detail: r#"Key (d)=(2) is not present in table "parent"."#, schema: "public", table: "child", constraint: "fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (2, 65536)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "fk""#, detail: r#"Key (d)=(65536) is not present in table "parent"."#, schema: "public", table: "child", constraint: "fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from parent where b = 3",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from parent where b = 1",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "parent" violates foreign key constraint "fk" on table "child""#, detail: r#"Key (b)=(1) is still referenced from table "child"."#, schema: "public", table: "child", constraint: "fk", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "type conversion: value out of bound, parent larger",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY, b int8)",
                "CREATE TABLE child (c INT PRIMARY KEY, d int2)",
                "INSERT INTO parent VALUES (1, 1), (65536, 65536)",
                "alter table parent add constraint ub unique (b)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "alter table child add constraint fk foreign key (d) references parent(b)",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (1, 1)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (2, 2)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "fk""#, detail: r#"Key (d)=(2) is not present in table "parent"."#, schema: "public", table: "child", constraint: "fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into child values (2, 65536)",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "smallint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from parent where b = 65536",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "delete from parent where b = 1",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "parent" violates foreign key constraint "fk" on table "child""#, detail: r#"Key (b)=(1) is still referenced from table "child"."#, schema: "public", table: "child", constraint: "fk", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key with dolt_add, dolt_commit",
            set_up_script: &[
                r#"create table test (pk int, "value" int, primary key(pk));"#,
                "CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk))",
                "INSERT INTO test VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_ADD('.')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.test"), T("t"), T("new table")],
                            &[T("public.test_info"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'new tables')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'test_info')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test_info VALUES (2, 'two', 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO test_info VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "test_info" violates foreign key constraint "test_info_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "test"."#, schema: "public", table: "test_info", constraint: "test_info_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test_info",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("2"), T("two"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key with explicit schema",
            set_up_script: &[
                r#"create table parent (pk int, "value" int, primary key(pk));"#,
                "CREATE TABLE child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references public.parent(pk))",
                "INSERT INTO parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_ADD('.')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.child"), T("t"), T("new table")],
                            &[T("public.parent"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'new tables')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 'two', 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "public", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM child",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("2"), T("two"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key in another schema with search path",
            set_up_script: &[
                "create schema parent",
                "create schema child",
                "create schema fake",
                "select dolt_commit('-Am', 'create schemas')",
                "set search_path to parent, child",
                "create table parent.parent (pk int, val int, primary key(pk));",
                "create table fake.parent (pk int, val int, primary key(pk));",
                "CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent(pk))",
                "INSERT INTO parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_ADD('.')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("child.child"), T("t"), T("new table")],
                            &[T("fake.parent"), T("t"), T("new table")],
                            &[T("parent.parent"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'new tables')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 'two', 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM child.child",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("2"), T("two"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key in another schema with search path, parent table not on search path",
            set_up_script: &[
                "create schema parent",
                "create schema child",
                "create schema fake",
                "select dolt_commit('-Am', 'create schemas')",
                "set search_path to child, fake",
                "create table parent.parent (pk int, val int, primary key(pk));",
                "create table fake.parent (pk int, val int, primary key(pk));",
                "CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent.parent(pk))",
                "INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_ADD('.')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("child.child"), T("t"), T("new table")],
                            &[T("fake.parent"), T("t"), T("new table")],
                            &[T("parent.parent"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'new tables')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (2, 'two', 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM child.child",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("2"), T("two"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key in another schema, no search path",
            set_up_script: &[
                "create schema parent",
                "create schema child",
                "create schema fake",
                "select dolt_commit('-Am', 'create schemas')",
                "create table parent.parent (pk int, val int, primary key(pk));",
                "create table fake.parent (pk int, val int, primary key(pk));",
                "CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent.parent(pk))",
                "INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_ADD('.')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("child.child"), T("t"), T("new table")],
                            &[T("fake.parent"), T("t"), T("new table")],
                            &[T("parent.parent"), T("t"), T("new table")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_commit('-am', 'new tables')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt.status",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
                    expected: Expected::Rows {
                        columns: &[Column("from_table_name", TEXT), Column("to_table_name", TEXT), Column("from_create_statement", TEXT), Column("to_create_statement", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (2, 'two', 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM child.child",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("2"), T("two"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "add foreign key in another schema on search path",
            set_up_script: &[
                "create schema parent",
                "create schema child",
                "create schema fake",
                "select dolt_commit('-Am', 'create schemas')",
                "set search_path to child, parent",
                "create table parent.parent (pk int, val int, primary key(pk));",
                "create table fake.parent (pk int, val int, primary key(pk));",
                "CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
                "INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_COMMIT('-Am', 'new tables')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (2, 'two', 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ALTER TABLE child ADD FOREIGN KEY (test_pk) REFERENCES parent(pk)",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM child",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("2"), T("two"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "add foreign key in another schema, parent table not on search path",
            set_up_script: &[
                "create schema parent",
                "create schema child",
                "create schema fake",
                "select dolt_commit('-Am', 'create schemas')",
                "set search_path to child, fake",
                "create table parent.parent (pk int, val int, primary key(pk));",
                "create table fake.parent (pk int, val int, primary key(pk));",
                "CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
                "INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_COMMIT('-Am', 'new tables')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (2, 'two', 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ALTER TABLE child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM child",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("2"), T("two"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "add foreign key in another schema, no search path",
            set_up_script: &[
                "create schema parent",
                "create schema child",
                "create schema fake",
                "select dolt_commit('-Am', 'create schemas')",
                "create table parent.parent (pk int, val int, primary key(pk));",
                "create table fake.parent (pk int, val int, primary key(pk));",
                "CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
                "INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_COMMIT('-Am', 'new tables')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (2, 'two', 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "ALTER TABLE child.child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM child.child",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("info", VARCHAR), Column("test_pk", INT4)],
                        rows: &[
                            &[T("2"), T("two"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop foreign key in another schema, on search path",
            set_up_script: &[
                "create schema parent",
                "create schema child",
                "create schema fake",
                "select dolt_commit('-Am', 'create schemas')",
                "set search_path to child, parent",
                "create table parent.parent (pk int, val int, primary key(pk));",
                "create table fake.parent (pk int, val int, primary key(pk));",
                "CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
                "INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_COMMIT('-Am', 'new tables')",
                "INSERT INTO child.child VALUES (2, 'two', 2)",
                "ALTER TABLE child.child ADD CONSTRAINT fk1 FOREIGN KEY (test_pk) REFERENCES parent(pk)",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "fk1""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "fk1", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "alter table child DROP constraint fk1;",
                    expected: Expected::Tag("ALTER TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (3, 'three', 3)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "drop foreign key in another schema, no search path",
            set_up_script: &[
                "create schema parent",
                "create schema child",
                "create schema fake",
                "select dolt_commit('-Am', 'create schemas')",
                "create table parent.parent (pk int, val int, primary key(pk));",
                "create table fake.parent (pk int, val int, primary key(pk));",
                "CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
                "INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
                "SELECT DOLT_COMMIT('-Am', 'new tables')",
                "INSERT INTO child.child VALUES (2, 'two', 2)",
                "ALTER TABLE child.child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "alter table child.child DROP constraint child_ibfk_1",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"constraint "child_ibfk_1" of relation "child" does not exist"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO child.child VALUES (3, 'three', 3)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "child" violates foreign key constraint "child_test_pk_fkey""#, detail: r#"Key (test_pk)=(3) is not present in table "parent"."#, schema: "child", table: "child", constraint: "child_test_pk_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key default naming",
            set_up_script: &[
                "CREATE TABLE webhooks (id varchar not null, id2 int8, primary key (id));",
                "CREATE UNIQUE INDEX idx1 on webhooks(id, id2);",
                "CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, webhook_id2_fk int8, foreign key (webhook_id_fk) references webhooks(id), foreign key (webhook_id_fk, webhook_id2_fk) references webhooks(id, id2), primary key (id));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_name", NAME)],
                        rows: &[
                            &[T("t33_webhook_id_fk_fkey")],
                            &[T("t33_webhook_id_fk_webhook_id2_fk_fkey")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key default naming, name collision ",
            set_up_script: &[
                "CREATE TABLE parent (id varchar not null primary key);",
                "CREATE TABLE child (id varchar primary key, constraint t33_webhook_id_fk_fkey foreign key (id) references parent(id));",
                "CREATE TABLE webhooks (id varchar not null, id2 int8, primary key (id));",
                "CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, foreign key (webhook_id_fk) references webhooks(id), primary key (id));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_name", NAME)],
                        rows: &[
                            &[T("t33_webhook_id_fk_fkey1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey1;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key default naming, in column definition",
            set_up_script: &[
                "CREATE TABLE webhooks (id varchar not null, primary key (id));",
                "CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null references webhooks(id), primary key (id));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_name", NAME)],
                        rows: &[
                            &[T("t33_webhook_id_fk_fkey")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key custom naming",
            set_up_script: &[
                "CREATE TABLE webhooks (id VARCHAR NOT NULL, PRIMARY KEY (id));",
                "CREATE TABLE t33 (id VARCHAR NOT NULL, webhook_id_fk VARCHAR NOT NULL, CONSTRAINT foo1 FOREIGN KEY (webhook_id_fk) REFERENCES webhooks(id), PRIMARY KEY (id));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass AND contype = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_name", NAME)],
                        rows: &[
                            &[T("foo1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t33 DROP CONSTRAINT foo1;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign key default naming, added through alter table",
            set_up_script: &[
                "CREATE TABLE webhooks (id varchar not null, primary key (id));",
                "CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, primary key (id));",
                "ALTER TABLE t33 ADD FOREIGN KEY (webhook_id_fk) REFERENCES webhooks(id);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_name", NAME)],
                        rows: &[
                            &[T("t33_webhook_id_fk_fkey")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ON DELETE ... SET DEFAULT",
            set_up_script: &[
                "CREATE TABLE public.hn_stories (title text NOT NULL, website_url text);",
                "CREATE TABLE public.websites (url text primary key, title text);",
                "INSERT into public.websites VALUES ('http://www.dolthub.com', 'foo1'), ('http://www.google.com', 'foo2');",
                "INSERT into public.hn_stories VALUES ('test1', 'http://www.dolthub.com'), ('test2', 'http://www.google.com');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"ALTER TABLE ONLY public.hn_stories
				ADD CONSTRAINT hn_stories_website_url_fkey FOREIGN KEY (website_url) REFERENCES public.websites(url) ON UPDATE CASCADE ON DELETE SET DEFAULT;"#,
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM public.websites WHERE title = 'foo1';",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.hn_stories where title = 'test1';",
                    expected: Expected::Rows {
                        columns: &[Column("title", TEXT), Column("website_url", TEXT)],
                        rows: &[
                            &[T("test1"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE hn_stories ALTER COLUMN website_url SET DEFAULT (title);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "cannot use column reference in DEFAULT expression", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM public.websites WHERE title = 'foo2';",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.hn_stories where title = 'test2';",
                    expected: Expected::Rows {
                        columns: &[Column("title", TEXT), Column("website_url", TEXT)],
                        rows: &[
                            &[T("test2"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "ON UPDATE ... SET DEFAULT",
            set_up_script: &[
                "CREATE TABLE public.hn_stories (title text NOT NULL, website_url text);",
                "CREATE TABLE public.websites (url text primary key, title text);",
                "INSERT into public.websites VALUES ('http://www.dolthub.com', 'foo1'), ('http://www.google.com', 'foo2');",
                "INSERT into public.hn_stories VALUES ('test1', 'http://www.dolthub.com'), ('test2', 'http://www.google.com');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"ALTER TABLE ONLY public.hn_stories
				ADD CONSTRAINT hn_stories_website_url_fkey FOREIGN KEY (website_url) REFERENCES public.websites(url) ON UPDATE SET DEFAULT;"#,
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE public.websites SET url = 'http://fake.com' WHERE title = 'foo1';",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.hn_stories where title = 'test1';",
                    expected: Expected::Rows {
                        columns: &[Column("title", TEXT), Column("website_url", TEXT)],
                        rows: &[
                            &[T("test1"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE hn_stories ALTER COLUMN website_url SET DEFAULT (title);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "cannot use column reference in DEFAULT expression", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE public.websites SET url = 'http://doltdb.com' WHERE title = 'foo2';",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM public.hn_stories where title = 'test2';",
                    expected: Expected::Rows {
                        columns: &[Column("title", TEXT), Column("website_url", TEXT)],
                        rows: &[
                            &[T("test2"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "merging",
            set_up_script: &[
                r#"CREATE TABLE "evaluation_job_config" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"project_id" varchar(256) NOT NULL,
	"job_filters" jsonb,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "evaluation_job_config_tenant_id_project_id_id_pk" PRIMARY KEY("tenant_id","project_id","id")
);"#,
                r#"CREATE TABLE "evaluation_job_config_evaluator_relations" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"project_id" varchar(256) NOT NULL,
	"evaluation_job_config_id" text NOT NULL,
	"evaluator_id" text NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "eval_job_cfg_evaluator_rel_pk" PRIMARY KEY("tenant_id","project_id","id")
);"#,
                r#"CREATE TABLE "agent" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"project_id" varchar(256) NOT NULL,
	"name" varchar(256) NOT NULL,
	"description" text,
	"default_sub_agent_id" varchar(256),
	"context_config_id" varchar(256),
	"models" jsonb,
	"status_updates" jsonb,
	"prompt" text,
	"stop_when" jsonb,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "agent_tenant_id_project_id_id_pk" PRIMARY KEY("tenant_id","project_id","id")
);"#,
                r#"CREATE TABLE "projects" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"name" varchar(256) NOT NULL,
	"description" text,
	"models" jsonb,
	"stop_when" jsonb,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "projects_tenant_id_id_pk" PRIMARY KEY("tenant_id","id")
);"#,
                r#"ALTER TABLE "evaluation_job_config" ADD CONSTRAINT "evaluation_job_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "public"."projects"("tenant_id","id") ON DELETE cascade ON UPDATE no action;"#,
                r#"ALTER TABLE "evaluation_job_config_evaluator_relations" ADD CONSTRAINT "eval_job_cfg_evaluator_rel_job_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_job_config_id") REFERENCES "public"."evaluation_job_config"("tenant_id","project_id","id") ON DELETE cascade ON UPDATE no action;"#,
                r#"INSERT INTO projects VALUES ('tenant1', 'project1', 'Project One', 'First project', '{"model": "gpt-4"}', '{"condition": "complete"}', now(), now());"#,
                r#"INSERT INTO evaluation_job_config VALUES ('tenant1', 'jobconfig1', 'project1', '{"filter": "all"}', now(), now());"#,
                "INSERT INTO evaluation_job_config_evaluator_relations VALUES ('tenant1', 'rel1', 'project1', 'jobconfig1', 'evaluator1', now(), now());",
                r#"INSERT INTO agent VALUES ('tenant1', 'agent1', 'project1', 'Agent One', 'First agent', null, null, '{"model": "gpt-4"}', '{}', 'You are an agent.', '{}', now(), now());"#,
                "SELECT DOLT_COMMIT('-Am', 'initial tables')",
                "SELECT DOLT_BRANCH('feature')",
                r#"CREATE TABLE "triggers" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"project_id" varchar(256) NOT NULL,
	"agent_id" varchar(256) NOT NULL,
	"name" varchar(256) NOT NULL,
	"description" text,
	"enabled" boolean DEFAULT true NOT NULL,
	"input_schema" jsonb,
	"output_transform" jsonb,
	"message_template" text NOT NULL,
	"authentication" jsonb,
	"signing_secret" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "triggers_tenant_id_project_id_agent_id_id_pk" PRIMARY KEY("tenant_id","project_id","agent_id","id")
);"#,
                r#"ALTER TABLE "triggers" ADD CONSTRAINT "triggers_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "public"."agent"("tenant_id","project_id","id") ON DELETE cascade ON UPDATE no action;"#,
                "select DOLT_COMMIT('-Am', 'add triggers table')",
                "select dolt_checkout('feature')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"insert into agent VALUES ('tenant1', 'agent2', 'project1', 'Agent Two', 'Second agent', null, null, '{"model": "gpt-4"}', '{}', 'You are another agent.', '{}', now(), now());"#,
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_commit('-Am', 'add second agent')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 1;",
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
            name: "merge with constraint violations",
            set_up_script: &[
                "CREATE TABLE parent (a INT PRIMARY KEY, b INT UNIQUE);",
                "CREATE TABLE child (c INT PRIMARY KEY, d INT);",
                "alter table child add constraint fk foreign key (d) references parent(b);",
                "INSERT INTO parent VALUES (1, 1), (2, 2), (3, 3);",
                "INSERT INTO child VALUES (1, 1), (2, 2);",
                "SELECT DOLT_COMMIT('-Am', 'initial commit')",
                "SELECT DOLT_BRANCH('feature')",
                "insert into child VALUES (3, 3);",
                "SELECT DOLT_COMMIT('-Am', 'new child')",
                "select dolt_checkout('feature')",
                "delete from parent where b = 3;",
                "SELECT DOLT_COMMIT('-Am', 'delete from parent')",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_merge('main')",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: r#"Committing this transaction resulted in a working set with constraint violations, transaction rolled back. This constraint violation may be the result of a previous merge or the result of transaction sequencing. Constraint violations from a merge can be resolved using the dolt_constraint_violations table before committing the transaction. To allow transactions to be committed with constraint violations from a merge or transaction sequencing set @@dolt_force_transaction_commit=1.
Constraint violations: 
Type: Foreign Key Constraint Violation
	ForeignKey: fk,
	Table: child,
	ReferencedTable: parent,
	Index: fk,
	ReferencedIndex: parent_b_key"#, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "set dolt_force_transaction_commit = 1;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select dolt_merge('main')",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select * from dolt_constraint_violations order by 1",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_violations", NUMERIC)],
                        rows: &[
                            &[T("child"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select violation_type, c, d, violation_info from dolt_constraint_violations_child order by 1",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("c", INT4), Column("d", INT4), Column("violation_info", JSON)],
                        rows: &[
                            &[T("foreign key"), T("3"), T("3"), T(r#"{"Columns":["d"],"ForeignKey":"fk","Index":"fk","OnDelete":"RESTRICT","OnUpdate":"RESTRICT","ReferencedColumns":["b"],"ReferencedIndex":"parent_b_key","ReferencedTable":"parent","Table":"child"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "foreign keys in drizzle migration and merge",
            set_up_script: &[
                r#"CREATE SCHEMA "drizzle";"#,
                r#"CREATE SEQUENCE drizzle."__drizzle_migrations_id_seq" AS int4;"#,
                r#"CREATE TABLE "__drizzle_migrations" (
  "id" integer NOT NULL DEFAULT (nextval('drizzle.__drizzle_migrations_id_seq')),
  "hash" text NOT NULL,
  "created_at" bigint,
  PRIMARY KEY ("id")
);"#,
                r#"INSERT INTO "__drizzle_migrations" ("hash","created_at") VALUES ('d3445cf0eaeb405a6b4b9c8386188aece144d40ba89b9616175ca0f69229cc51',1767821157311);"#,
                r#"CREATE TABLE "projects" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "models" jsonb,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","id")
);"#,
                r#"CREATE TABLE "agent" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "default_sub_agent_id" varchar(256),
  "context_config_id" varchar(256),
  "models" jsonb,
  "status_updates" jsonb,
  "prompt" text,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "agent_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "artifact_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "props" jsonb,
  "render" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "artifact_components_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "context_configs" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "headers_schema" jsonb,
  "context_variables" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "context_configs_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "credential_references" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "type" varchar(256) NOT NULL,
  "credential_store_id" varchar(256) NOT NULL,
  "retrieval_params" jsonb,
  "tool_id" varchar(256),
  "user_id" varchar(256),
  "created_by" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "credential_references_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE UNIQUE INDEX "credential_references_id_unique" ON "credential_references" ("id");"#,
                r#"CREATE UNIQUE INDEX "credential_references_tool_user_unique" ON "credential_references" ("tool_id", "user_id");"#,
                r#"CREATE TABLE "data_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "props" jsonb,
  "render" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "data_components_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "dataset" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "dataset_item" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "dataset_id" text NOT NULL,
  "input" jsonb NOT NULL,
  "expected_output" jsonb,
  "simulation_agent" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_item_dataset_fk" FOREIGN KEY ("tenant_id","project_id","dataset_id") REFERENCES "dataset" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "dataset_run_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "dataset_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_run_config_dataset_fk" FOREIGN KEY ("tenant_id","project_id","dataset_id") REFERENCES "dataset" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "dataset_run_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "dataset_run_config_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "dataset_run_config_id" text NOT NULL,
  "agent_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_run_config_agent_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "dataset_run_config_agent_relations_dataset_run_config_fk" FOREIGN KEY ("tenant_id","project_id","dataset_run_config_id") REFERENCES "dataset_run_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "evaluation_job_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "job_filters" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_job_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "evaluator" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "prompt" text NOT NULL,
  "schema" jsonb NOT NULL,
  "model" jsonb NOT NULL,
  "pass_criteria" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluator_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "evaluation_job_config_evaluator_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_job_config_id" text NOT NULL,
  "evaluator_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_job_cfg_evaluator_rel_evaluator_fk" FOREIGN KEY ("tenant_id","project_id","evaluator_id") REFERENCES "evaluator" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_job_cfg_evaluator_rel_job_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_job_config_id") REFERENCES "evaluation_job_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "evaluation_run_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "is_active" boolean NOT NULL DEFAULT 'true',
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_run_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "evaluation_suite_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "filters" jsonb,
  "sample_rate" double precision,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_suite_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "evaluation_run_config_evaluation_suite_config_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_run_config_id" text NOT NULL,
  "evaluation_suite_config_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_run_cfg_eval_suite_rel_run_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_run_config_id") REFERENCES "evaluation_run_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_run_cfg_eval_suite_rel_suite_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_suite_config_id") REFERENCES "evaluation_suite_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "evaluation_suite_config_evaluator_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_suite_config_id" text NOT NULL,
  "evaluator_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_suite_cfg_evaluator_rel_evaluator_fk" FOREIGN KEY ("tenant_id","project_id","evaluator_id") REFERENCES "evaluator" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_suite_cfg_evaluator_rel_suite_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_suite_config_id") REFERENCES "evaluation_suite_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "external_agents" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "base_url" text NOT NULL,
  "credential_reference_id" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "external_agents_credential_reference_fk" FOREIGN KEY ("credential_reference_id") REFERENCES "credential_references" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
  CONSTRAINT "external_agents_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "functions" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "input_schema" jsonb,
  "execute_code" text NOT NULL,
  "dependencies" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "functions_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "function_tools" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "function_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "function_tools_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "function_tools_function_fk" FOREIGN KEY ("tenant_id","project_id","function_id") REFERENCES "functions" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "sub_agents" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "prompt" text,
  "conversation_history_config" jsonb DEFAULT '{"mode":"full","limit":50,"maxOutputTokens":4000,"includeInternal":false,"messageTypes":["chat","tool-result"]}'::JSONB,
  "models" jsonb,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agents_agents_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "tools" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "config" jsonb NOT NULL,
  "credential_reference_id" varchar(256),
  "credential_scope" varchar(50) NOT NULL DEFAULT 'project',
  "headers" jsonb,
  "image_url" text,
  "capabilities" jsonb,
  "last_error" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "tools_credential_reference_fk" FOREIGN KEY ("credential_reference_id") REFERENCES "credential_references" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
  CONSTRAINT "tools_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "sub_agent_artifact_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "artifact_component_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","sub_agent_id","id"),
  CONSTRAINT "sub_agent_artifact_components_artifact_component_fk" FOREIGN KEY ("tenant_id","project_id","artifact_component_id") REFERENCES "artifact_components" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_artifact_components_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "sub_agent_data_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "data_component_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "sub_agent_data_components_data_component_fk" FOREIGN KEY ("tenant_id","project_id","data_component_id") REFERENCES "data_components" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_data_components_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "sub_agent_external_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "external_agent_id" varchar(256) NOT NULL,
  "headers" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_external_agent_relations_external_agent_fk" FOREIGN KEY ("tenant_id","project_id","external_agent_id") REFERENCES "external_agents" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_external_agent_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "sub_agent_function_tool_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "function_tool_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_function_tool_relations_function_tool_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","function_tool_id") REFERENCES "function_tools" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_function_tool_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "sub_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "source_sub_agent_id" varchar(256) NOT NULL,
  "target_sub_agent_id" varchar(256),
  "relation_type" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "sub_agent_team_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "target_agent_id" varchar(256) NOT NULL,
  "headers" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_team_agent_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_team_agent_relations_target_agent_fk" FOREIGN KEY ("tenant_id","project_id","target_agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                r#"CREATE TABLE "sub_agent_tool_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "tool_id" varchar(256) NOT NULL,
  "selected_tools" jsonb,
  "headers" jsonb,
  "tool_policies" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_tool_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_tool_relations_tool_fk" FOREIGN KEY ("tenant_id","project_id","tool_id") REFERENCES "tools" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                "SELECT DOLT_COMMIT('-Am', 'Applied database migrations');",
                "SELECT DOLT_BRANCH('default_my-weather-project_main');",
                r#"INSERT INTO "__drizzle_migrations" ("hash","created_at") VALUES ('634b9140001f10d551fe0d81ca19050f3cc8af8da1ab6c9b6e93d99f33e5fc84',1768766675586);"#,
                r#"CREATE TABLE "triggers" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "enabled" boolean NOT NULL DEFAULT 'true',
  "input_schema" jsonb,
  "output_transform" jsonb,
  "message_template" text NOT NULL,
  "authentication" jsonb,
  "signing_secret" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "triggers_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);"#,
                "SELECT DOLT_COMMIT('-Am', 'Applied database migrations');",
                "SELECT DOLT_CHECKOUT('default_my-weather-project_main');",
                r#"INSERT INTO "projects" ("tenant_id","id","name","description","models","stop_when","created_at","updated_at") VALUES ('default','my-weather-project','Weather Project','Project containing sample agent framework using ','{"base": {"model": "openai/gpt-4o-mini"}}',NULL,'2026-01-22 16:19:32.74','2026-01-22 16:19:32.74');"#,
                r#"INSERT INTO "agent" ("tenant_id","id","project_id","name","description","default_sub_agent_id","context_config_id","models","status_updates","prompt","stop_when","created_at","updated_at") VALUES ('default','weather-agent','my-weather-project','Weather agent',NULL,'weather-assistant',NULL,NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.782','2026-01-22 16:19:32.862');"#,
                r#"INSERT INTO "data_components" ("tenant_id","id","project_id","name","description","props","render","created_at","updated_at") VALUES ('default','weather-forecast','my-weather-project','WeatherForecast','A hourly forecast for the weather at a given location','{"type": "object", "required": ["forecast"], "properties": {"forecast": {"type": "array", "items": {"type": "object", "required": ["time", "temperature", "code"], "properties": {"code": {"type": "number", "description": "Weather code at given time"}, "time": {"type": "string", "description": "The time of current item E.g. 12PM, 1PM"}, "temperature": {"type": "number", "description": "The temperature at given time in Farenheit"}}, "additionalProperties": false}, "description": "The hourly forecast for the weather at a given location"}}, "additionalProperties": false}',NULL,'2026-01-22 16:19:32.773665','2026-01-22 16:19:32.773665');"#,
                r#"INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','geocoder-agent','my-weather-project','weather-agent','Geocoder agent','Specialized agent for converting addresses and location names into geographic coordinates. This agent handles all location-related queries and provides accurate latitude/longitude data for weather lookups.','You are a geocoding specialist that converts addresses, place names, and location descriptions
 into precise geographic coordinates. You help users find the exact location they''re asking about
 and provide the coordinates needed for weather forecasting.

 When users provide:
 - Street addresses
 - City names
 - Landmarks
 - Postal codes
 - General location descriptions

 You should use your geocoding tools to find the most accurate coordinates and provide clear
 information about the location found.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.848333','2026-01-22 16:19:32.848333');"#,
                r#"INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','weather-assistant','my-weather-project','weather-agent','Weather assistant','Main weather assistant that coordinates between geocoding and forecasting services to provide comprehensive weather information. This assistant handles user queries and delegates tasks to specialized sub-agents as needed.','You are a helpful weather assistant that provides comprehensive weather information
 for any location worldwide. You coordinate with specialized agents to:

 1. Convert location names/addresses to coordinates (via geocoder)
 2. Retrieve detailed weather forecasts (via weather forecaster)
 3. Present weather information in a clear, user-friendly format

 When users ask about weather:
 - If they provide a location name or address, delegate to the geocoder first
 - Once you have coordinates, delegate to the weather forecaster
 - Present the final weather information in an organized, easy-to-understand format
 - Include relevant details like temperature, conditions, precipitation, wind, etc.
 - Provide helpful context and recommendations when appropriate

 You have access to weather forecast data components that can enhance your responses
 with structured weather information.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.851804','2026-01-22 16:19:32.851804');"#,
                r#"INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','weather-forecaster','my-weather-project','weather-agent','Weather forecaster','Specialized agent for retrieving detailed weather forecasts and current conditions. This agent focuses on providing accurate, up-to-date weather information using geographic coordinates.','You are a weather forecasting specialist that provides detailed weather information
 including current conditions, forecasts, and weather-related insights.

 You work with precise geographic coordinates to deliver:
 - Current weather conditions
 - Short-term and long-term forecasts
 - Temperature, humidity, wind, and precipitation data
 - Weather alerts and advisories
 - Seasonal and climate information

 Always provide clear, actionable weather information that helps users plan their activities.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.844618','2026-01-22 16:19:32.844618');"#,
                r#"INSERT INTO "tools" ("tenant_id","id","project_id","name","description","config","credential_reference_id","credential_scope","headers","image_url","capabilities","last_error","created_at","updated_at") VALUES ('default','fUI2riwrBVJ6MepT8rjx0','my-weather-project','Forecast weather',NULL,'{"mcp": {"server": {"url": "https://weather-mcp-hazel.vercel.app/mcp"}}, "type": "mcp"}',NULL,'project',NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.748','2026-01-22 16:19:32.748');"#,
                r#"INSERT INTO "tools" ("tenant_id","id","project_id","name","description","config","credential_reference_id","credential_scope","headers","image_url","capabilities","last_error","created_at","updated_at") VALUES ('default','fdxgfv9HL7SXlfynPx8hf','my-weather-project','Geocode address',NULL,'{"mcp": {"server": {"url": "https://weather-mcp-hazel.vercel.app/mcp"}}, "type": "mcp"}',NULL,'project',NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.75','2026-01-22 16:19:32.75');"#,
                r#"INSERT INTO "sub_agent_relations" ("tenant_id","id","project_id","agent_id","source_sub_agent_id","target_sub_agent_id","relation_type","created_at","updated_at") VALUES ('default','0y59hwkkyzml4dq4t1sx8','my-weather-project','weather-agent','weather-assistant','weather-forecaster','delegate','2026-01-22 16:19:32.92219','2026-01-22 16:19:32.92219');"#,
                r#"INSERT INTO "sub_agent_relations" ("tenant_id","id","project_id","agent_id","source_sub_agent_id","target_sub_agent_id","relation_type","created_at","updated_at") VALUES ('default','7ye45uc4j5442ihgqwn6d','my-weather-project','weather-agent','weather-assistant','geocoder-agent','delegate','2026-01-22 16:19:32.925527','2026-01-22 16:19:32.925527');"#,
                r#"INSERT INTO "sub_agent_data_components" ("tenant_id","id","project_id","agent_id","sub_agent_id","data_component_id","created_at") VALUES ('default','689yd78rj16p9880bndfo','my-weather-project','weather-agent','weather-assistant','weather-forecast','2026-01-22 16:19:32.907332');"#,
                r#"INSERT INTO "sub_agent_tool_relations" ("tenant_id","id","project_id","agent_id","sub_agent_id","tool_id","selected_tools","headers","tool_policies","created_at","updated_at") VALUES ('default','4kws0lm8bqi1mkzwbvmz4','my-weather-project','weather-agent','weather-forecaster','fUI2riwrBVJ6MepT8rjx0',NULL,NULL,NULL,'2026-01-22 16:19:32.888','2026-01-22 16:19:32.888');"#,
                r#"INSERT INTO "sub_agent_tool_relations" ("tenant_id","id","project_id","agent_id","sub_agent_id","tool_id","selected_tools","headers","tool_policies","created_at","updated_at") VALUES ('default','ttz1a9tnso0sxim79iphr','my-weather-project','weather-agent','geocoder-agent','fdxgfv9HL7SXlfynPx8hf',NULL,NULL,NULL,'2026-01-22 16:19:32.889','2026-01-22 16:19:32.889');"#,
                "SELECT DOLT_COMMIT('-Am', '//Update /manage/tenants/default/project-full/my-weather-project via API');",
                r#"UPDATE "tools" SET "updated_at"='2026-01-22 16:19:50.912' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';"#,
                r#"UPDATE "tools" SET "updated_at"='2026-01-22 16:19:50.967' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';"#,
                "SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');",
                r#"INSERT INTO "evaluator" ("tenant_id","id","project_id","name","description","prompt","schema","model","pass_criteria","created_at","updated_at") VALUES ('default','ubqho5lsm6h7bd3ra8loz','my-weather-project','test','test','test','{"type": "object", "required": ["test"], "properties": {"test": {"type": "string", "description": "test"}}, "additionalProperties": false}','{"model": "anthropic/claude-opus-4-5"}',NULL,'2026-01-22 16:20:07.188','2026-01-22 16:20:07.188');"#,
                "SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluators via API');",
                r#"UPDATE "tools" SET "updated_at"='2026-01-22 16:20:11.438' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';"#,
                r#"UPDATE "tools" SET "updated_at"='2026-01-22 16:20:11.448' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';"#,
                "SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');",
                r#"UPDATE "tools" SET "updated_at"='2026-01-22 16:20:17.821' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';"#,
                r#"UPDATE "tools" SET "updated_at"='2026-01-22 16:20:18.082' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';"#,
                "SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');",
                r#"INSERT INTO "evaluation_job_config" ("tenant_id","id","project_id","job_filters","created_at","updated_at") VALUES ('default','tj06kzjt8ltlyixgfzeao','my-weather-project','{"dateRange": {"endDate": "2026-01-23T04:59:59.999Z", "startDate": "2026-01-21T05:00:00.000Z"}}','2026-01-22 16:20:55.774','2026-01-22 16:20:55.774');"#,
                r#"INSERT INTO "evaluation_job_config_evaluator_relations" ("tenant_id","id","project_id","evaluation_job_config_id","evaluator_id","created_at","updated_at") VALUES ('default','5qk0w692h5ij1sxtohdua','my-weather-project','tj06kzjt8ltlyixgfzeao','ubqho5lsm6h7bd3ra8loz','2026-01-22 16:20:55.781','2026-01-22 16:20:55.781');"#,
                "SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-job-configs via API');",
                r#"INSERT INTO "evaluation_suite_config" ("tenant_id","id","project_id","filters","sample_rate","created_at","updated_at") VALUES ('default','j5gvgluqzwzhjhycrsnpf','my-weather-project','{"agentIds": ["weather-agent"]}',NULL,'2026-01-22 16:21:19.974','2026-01-22 16:21:19.974');"#,
                r#"INSERT INTO "evaluation_suite_config_evaluator_relations" ("tenant_id","id","project_id","evaluation_suite_config_id","evaluator_id","created_at","updated_at") VALUES ('default','tz51dzynx71gits265e9d','my-weather-project','j5gvgluqzwzhjhycrsnpf','ubqho5lsm6h7bd3ra8loz','2026-01-22 16:21:19.982','2026-01-22 16:21:19.982');"#,
                "SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-suite-configs via API');",
                r#"INSERT INTO "evaluation_run_config" ("tenant_id","id","project_id","name","description","is_active","created_at","updated_at") VALUES ('default','74pgwrprmea2o7e6avbh7','my-weather-project','test','test',true,'2026-01-22 16:21:20.104','2026-01-22 16:21:20.104');"#,
                r#"INSERT INTO "evaluation_run_config_evaluation_suite_config_relations" ("tenant_id","id","project_id","evaluation_run_config_id","evaluation_suite_config_id","created_at","updated_at") VALUES ('default','plb31qfzw9803g6hbjhef','my-weather-project','74pgwrprmea2o7e6avbh7','j5gvgluqzwzhjhycrsnpf','2026-01-22 16:21:20.111','2026-01-22 16:21:20.111');"#,
                "SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-run-configs via API');",
                r#"UPDATE "tools" SET "updated_at"='2026-01-22 16:21:23.521' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';"#,
                r#"UPDATE "tools" SET "updated_at"='2026-01-22 16:21:23.771' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';"#,
                "SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
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
            name: "Merge foreign keys across types, text -> varchar",
            set_up_script: &[
                r#"CREATE TABLE table1 (
            table1_col1 VARCHAR(256),
            table1_col2 VARCHAR(256),
            table1_col3 VARCHAR(256),
            PRIMARY KEY (table1_col1, table1_col3, table1_col2)
        );"#,
                r#"CREATE TABLE table2 (
            table2_col1 VARCHAR(256),
            table2_col2 VARCHAR(256),
            table2_col3 VARCHAR(256),
            table2_col4 TEXT,
            PRIMARY KEY (table2_col1, table2_col3, table2_col2),
            CONSTRAINT table2_fk FOREIGN KEY (table2_col1, table2_col3, table2_col4) REFERENCES table1 (table1_col1, table1_col3, table1_col2) ON DELETE CASCADE ON UPDATE NO ACTION
        );"#,
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "INSERT INTO table1 (table1_col1, table1_col2, table1_col3) VALUES ('abc','def','ghi');",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','jkl','ghi','def');",
                "SELECT DOLT_COMMIT('-Am', '4');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
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
            name: "Merge foreign keys across types, varchar -> text",
            set_up_script: &[
                r#"CREATE TABLE table1 (
            table1_col1 VARCHAR(256),
            table1_col2 text,
            table1_col3 text,
            PRIMARY KEY (table1_col1, table1_col3, table1_col2)
        );"#,
                r#"CREATE TABLE table2 (
            table2_col1 VARCHAR(256),
            table2_col2 VARCHAR(256),
            table2_col3 VARCHAR(256),
            table2_col4 VARCHAR(256),
            PRIMARY KEY (table2_col1, table2_col3, table2_col2),
            CONSTRAINT table2_fk FOREIGN KEY (table2_col1, table2_col3, table2_col4) REFERENCES table1 (table1_col1, table1_col3, table1_col2) ON DELETE CASCADE ON UPDATE NO ACTION
        );"#,
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "INSERT INTO table1 (table1_col1, table1_col2, table1_col3) VALUES ('abc','def','ghi');",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','jkl','ghi','def');",
                "SELECT DOLT_COMMIT('-Am', '4');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
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
            name: "Merge foreign keys across types, varchar -> text, violation",
            set_up_script: &[
                "CREATE TABLE parent (a TEXT PRIMARY KEY);",
                "CREATE TABLE child (b INT PRIMARY KEY, c varchar(255), CONSTRAINT fk FOREIGN KEY (c) REFERENCES parent(a) ON DELETE CASCADE ON UPDATE NO ACTION);",
                "INSERT INTO parent (a) VALUES ('abc'), ('def');",
                "INSERT INTO child (b, c) VALUES (1, 'abc'), (2, 'def');",
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "DELETE FROM child WHERE b=1;",
                "DELETE FROM parent WHERE a='abc';",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "INSERT INTO child (b, c) VALUES (3, 'abc');",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "set dolt_force_transaction_commit=1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select violation_type, b, c from dolt_constraint_violations_child;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("b", INT4), Column("c", VARCHAR)],
                        rows: &[
                            &[T("foreign key"), T("3"), T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Merge foreign keys across types, varchar -> text 2 column key, violation",
            set_up_script: &[
                r#"CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );"#,
                "CREATE INDEX idx_parent_on_a_b ON parent (a, b);",
                r#"CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d),
            CONSTRAINT child_fk FOREIGN KEY (d, f) REFERENCES parent (a, b) ON DELETE CASCADE ON UPDATE NO ACTION
        );"#,
                "INSERT INTO parent VALUES ('abc','def', 'xyz');",
                "INSERT INTO child VALUES ('abc','123','def');",
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "INSERT INTO child VALUES ('abc','www','def');",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "delete from child where e='def';",
                "delete from parent where a='abc';",
                "SELECT DOLT_COMMIT('-Am', '4');",
                "set dolt_force_transaction_commit=1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select violation_type, d, e, f from dolt_constraint_violations_child;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("d", VARCHAR), Column("e", VARCHAR), Column("f", VARCHAR)],
                        rows: &[
                            &[T("foreign key"), T("abc"), T("www"), T("def")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Merge foreign keys across types, varchar -> text 2 column key, no violation",
            set_up_script: &[
                r#"CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );"#,
                "CREATE INDEX idx_parent_on_a_b ON parent (a, b);",
                r#"CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d),
            CONSTRAINT child_fk FOREIGN KEY (d, f) REFERENCES parent (a, b) ON DELETE CASCADE ON UPDATE NO ACTION
        );"#,
                "INSERT INTO parent VALUES ('abc','def', 'xyz');",
                "INSERT INTO child VALUES ('abc','123','def');",
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "INSERT INTO child VALUES ('abc','www','def');",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "INSERT INTO child VALUES ('abc','xyz','def');",
                "SELECT DOLT_COMMIT('-Am', '4');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
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
                    query: "select violation_type, d, e, f from dolt_constraint_violations_child;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("d", VARCHAR), Column("e", VARCHAR), Column("f", VARCHAR)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Merge foreign keys across types, varchar -> text 2 column key, parent primary index, violation",
            set_up_script: &[
                r#"CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );"#,
                r#"CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d),
            CONSTRAINT child_fk FOREIGN KEY (f, d) REFERENCES parent (b, a) ON DELETE CASCADE ON UPDATE NO ACTION
        );"#,
                "INSERT INTO parent VALUES ('abc','def', 'xyz');",
                "INSERT INTO child VALUES ('abc','123','def');",
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "INSERT INTO child VALUES ('abc','www','def');",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "delete from child where e='def';",
                "delete from parent where a='abc';",
                "SELECT DOLT_COMMIT('-Am', '4');",
                "set dolt_force_transaction_commit=1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select violation_type, d, e, f from dolt_constraint_violations_child;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("d", VARCHAR), Column("e", VARCHAR), Column("f", VARCHAR)],
                        rows: &[
                            &[T("foreign key"), T("abc"), T("www"), T("def")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Merge foreign keys across types, varchar -> text 2 column key, parent primary index, no violation",
            set_up_script: &[
                r#"CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );"#,
                r#"CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d),
            CONSTRAINT child_fk FOREIGN KEY (f, d) REFERENCES parent (b, a) ON DELETE CASCADE ON UPDATE NO ACTION
        );"#,
                "INSERT INTO parent VALUES ('abc','def', 'xyz');",
                "INSERT INTO child VALUES ('abc','123','def');",
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "INSERT INTO child VALUES ('abc','www','def');",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "INSERT INTO child VALUES ('abc','xyz','def');",
                "SELECT DOLT_COMMIT('-Am', '4');",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
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
                    query: "select violation_type, d, e, f from dolt_constraint_violations_child;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("d", VARCHAR), Column("e", VARCHAR), Column("f", VARCHAR)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Merge foreign keys across types, varchar -> text 2 column key, child primary index, violation",
            set_up_script: &[
                r#"CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );"#,
                r#"CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d)
        );"#,
                "INSERT INTO parent VALUES ('abc','def', 'xyz');",
                "INSERT INTO child VALUES ('abc','123', 'def');",
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "ALTER TABLE child ADD CONSTRAINT child_fk FOREIGN KEY (f, d) REFERENCES parent (b, a);",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "INSERT INTO child VALUES ('abc','def','xxx');",
                "SELECT DOLT_COMMIT('-Am', '4');",
                "set dolt_force_transaction_commit=1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select violation_type, d, e, f from dolt_constraint_violations_child;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("d", VARCHAR), Column("e", VARCHAR), Column("f", VARCHAR)],
                        rows: &[
                            &[T("foreign key"), T("abc"), T("def"), T("xxx")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Merge foreign keys across types, varchar -> text 2 column key, child secondary index, violation",
            set_up_script: &[
                r#"CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );"#,
                r#"CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            g VARCHAR(256),
            PRIMARY KEY (e, d)
        );"#,
                "INSERT INTO parent VALUES ('abc','def', 'xyz');",
                "INSERT INTO child VALUES ('abc','123', 'abc', 'def');",
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "ALTER TABLE child ADD CONSTRAINT child_fk FOREIGN KEY (g, f) REFERENCES parent (b, a);",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "INSERT INTO child VALUES ('xyz','123', 'def', 'abc');",
                "SELECT DOLT_COMMIT('-Am', '4');",
                "set dolt_force_transaction_commit=1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select violation_type, d, e, f, g from dolt_constraint_violations_child;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("d", VARCHAR), Column("e", VARCHAR), Column("f", VARCHAR), Column("g", VARCHAR)],
                        rows: &[
                            &[T("foreign key"), T("xyz"), T("123"), T("def"), T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Merge foreign keys across types, varchar -> text 3 column key, violation",
            set_up_script: &[
                r#"CREATE TABLE table1 (
            table1_col1 VARCHAR(256),
            table1_col2 text,
            table1_col3 text,
            PRIMARY KEY (table1_col1, table1_col3, table1_col2)
        );"#,
                r#"CREATE TABLE table2 (
            table2_col1 VARCHAR(256),
            table2_col2 VARCHAR(256),
            table2_col3 VARCHAR(256),
            table2_col4 VARCHAR(256),
            PRIMARY KEY (table2_col1, table2_col3, table2_col2),
            CONSTRAINT table2_fk FOREIGN KEY (table2_col1, table2_col3, table2_col4) REFERENCES table1 (table1_col1, table1_col3, table1_col2) ON DELETE CASCADE ON UPDATE NO ACTION
        );"#,
                "INSERT INTO table1 (table1_col1, table1_col2, table1_col3) VALUES ('abc','def','ghi');",
                "INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','jkl','ghi','def');",
                "SELECT DOLT_COMMIT('-Am', '1');",
                "SELECT DOLT_BRANCH('other_branch');",
                "INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','xyz','ghi','def');",
                "SELECT DOLT_COMMIT('-Am', '2');",
                "CREATE TABLE table3 (table3_col1 VARCHAR(256));",
                "SELECT DOLT_COMMIT('-Am', '3');",
                "SELECT DOLT_CHECKOUT('other_branch');",
                "delete from table2 where table2_col2='jkl';",
                "delete from table1 where table1_col2='def';",
                "SELECT DOLT_COMMIT('-Am', '4');",
                "set dolt_force_transaction_commit=1;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "select violation_type, table2_col1, table2_col2, table2_col3, table2_col4 from dolt_constraint_violations_table2;",
                    expected: Expected::Rows {
                        columns: &[Column("violation_type", VARCHAR), Column("table2_col1", VARCHAR), Column("table2_col2", VARCHAR), Column("table2_col3", VARCHAR), Column("table2_col4", VARCHAR)],
                        rows: &[
                            &[T("foreign key"), T("abc"), T("xyz"), T("ghi"), T("def")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Self-referential foreign key with schema-qualified column reference",
            set_up_script: &[
                "CREATE SCHEMA myschema",
                "CREATE TABLE myschema.t (a INT PRIMARY KEY, b INT REFERENCES myschema.t (a))",
                "INSERT INTO myschema.t VALUES (1, NULL), (2, 1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO myschema.t VALUES (3, 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO myschema.t VALUES (4, 99)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "t" violates foreign key constraint "t_b_fkey""#, detail: r#"Key (b)=(99) is not present in table "t"."#, schema: "myschema", table: "t", constraint: "t_b_fkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Self-referential foreign key with schema-qualified table constraint",
            set_up_script: &[
                "CREATE TABLE public.t (a INT PRIMARY KEY, b INT, FOREIGN KEY (b) REFERENCES public.t (a))",
                "INSERT INTO public.t VALUES (1, NULL), (2, 1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (3, 2)",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (4, 99)",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "t" violates foreign key constraint "t_b_fkey""#, detail: r#"Key (b)=(99) is not present in table "t"."#, schema: "public", table: "t", constraint: "t_b_fkey", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dropping a column that removes the index backing a foreign key",
            set_up_script: &[
                "CREATE TABLE fkbug_parent (id INT NOT NULL PRIMARY KEY);",
                "CREATE TABLE fkbug_child (id INT NOT NULL PRIMARY KEY, a_id INT NOT NULL, b_id INT NOT NULL);",
                "ALTER TABLE fkbug_child ADD CONSTRAINT fkbug_child_a_b_uniq UNIQUE (a_id, b_id);",
                "ALTER TABLE fkbug_child ADD CONSTRAINT fkbug_child_a_id_fk FOREIGN KEY (a_id) REFERENCES fkbug_parent (id);",
                "INSERT INTO fkbug_parent VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE fkbug_child DROP COLUMN b_id;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('-A');",
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
                    query: "SELECT dolt_commit('-am', 'probe');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_constraint WHERE conrelid = 'fkbug_child'::regclass AND contype = 'f';",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[
                            &[T("fkbug_child_a_id_fk")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO fkbug_child VALUES (1, 1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "INSERT INTO fkbug_child VALUES (2, 2);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "fkbug_child" violates foreign key constraint "fkbug_child_a_id_fk""#, detail: r#"Key (a_id)=(2) is not present in table "fkbug_parent"."#, schema: "public", table: "fkbug_child", constraint: "fkbug_child_a_id_fk", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM fkbug_child;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a_id", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP COLUMN drops dependent foreign keys",
            set_up_script: &[
                "CREATE TABLE fkbug2_parent (id INT NOT NULL PRIMARY KEY, u INT UNIQUE);",
                "CREATE TABLE fkbug2_child (id INT NOT NULL PRIMARY KEY, p_id INT NULL, q_id INT);",
                "CREATE INDEX fkbug2_child_p_id_idx ON fkbug2_child (p_id);",
                "ALTER TABLE fkbug2_child ADD CONSTRAINT fkbug2_child_p_id_fk FOREIGN KEY (p_id) REFERENCES fkbug2_parent (id);",
                "ALTER TABLE fkbug2_child ADD CONSTRAINT fkbug2_child_q_id_fk FOREIGN KEY (q_id) REFERENCES fkbug2_parent (u);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TABLE fkbug2_parent DROP COLUMN u;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop column u of table fkbug2_parent because other objects depend on it", detail: "constraint fkbug2_child_q_id_fk on table fkbug2_child depends on column u of table fkbug2_parent", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE fkbug2_parent DROP COLUMN u CASCADE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to constraint fkbug2_child_q_id_fk on table fkbug2_child", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_constraint WHERE conrelid = 'fkbug2_child'::regclass AND contype = 'f' ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[
                            &[T("fkbug2_child_p_id_fk")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE fkbug2_child DROP COLUMN p_id CASCADE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_constraint WHERE conrelid = 'fkbug2_child'::regclass AND contype = 'f' ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO fkbug2_child VALUES (1, 99);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM fkbug2_child;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("q_id", INT4)],
                        rows: &[
                            &[T("1"), T("99")],
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
fn test_foreign_key_rules() {
    run_scripts(&[
        ScriptTest {
            name: "referential actions on single-column keys",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE p (id INT PRIMARY KEY, code TEXT UNIQUE, v INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE c (cid INT PRIMARY KEY, pid INT REFERENCES p, pcode TEXT, FOREIGN KEY (pcode) REFERENCES p (code) ON DELETE CASCADE ON UPDATE CASCADE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO p VALUES (1, 'a', 10), (2, 'b', 20), (3, 'c', 30);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO c VALUES (1, 1, 'a'), (2, 2, 'b');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO c VALUES (3, 5, NULL);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "c" violates foreign key constraint "c_pid_fkey""#, detail: r#"Key (pid)=(5) is not present in table "p"."#, schema: "public", table: "c", constraint: "c_pid_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO c VALUES (3, NULL, 'z');",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "c" violates foreign key constraint "c_pcode_fkey""#, detail: r#"Key (pcode)=(z) is not present in table "p"."#, schema: "public", table: "c", constraint: "c_pcode_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO c VALUES (3, NULL, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM p WHERE id = 1;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "p" violates foreign key constraint "c_pid_fkey" on table "c""#, detail: r#"Key (id)=(1) is still referenced from table "c"."#, schema: "public", table: "c", constraint: "c_pid_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE p SET id = 10 WHERE id = 2;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "p" violates foreign key constraint "c_pid_fkey" on table "c""#, detail: r#"Key (id)=(2) is still referenced from table "c"."#, schema: "public", table: "c", constraint: "c_pid_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM p WHERE id = 3;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE p SET code = 'bb' WHERE code = 'b';",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM c ORDER BY cid;",
                    expected: Expected::Rows {
                        columns: &[Column("cid", INT4), Column("pid", INT4), Column("pcode", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("a")],
                            &[T("2"), T("2"), T("bb")],
                            &[T("3"), Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM p WHERE code = 'bb';",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "p" violates foreign key constraint "c_pid_fkey" on table "c""#, detail: r#"Key (id)=(2) is still referenced from table "c"."#, schema: "public", table: "c", constraint: "c_pid_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM c ORDER BY cid;",
                    expected: Expected::Rows {
                        columns: &[Column("cid", INT4), Column("pid", INT4), Column("pcode", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("a")],
                            &[T("2"), T("2"), T("bb")],
                            &[T("3"), Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE d (x INT REFERENCES p (v));",
                    expected: Expected::Error(Diagnostic { code: "42830", message: r#"there is no unique constraint matching given keys for referenced table "p""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE d (x TEXT REFERENCES p);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"foreign key constraint "d_x_fkey" cannot be implemented"#, detail: r#"Key columns "x" of the referencing table and "id" of the referenced table are of incompatible types: text and integer."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE d (x INT, y INT, FOREIGN KEY (x, y) REFERENCES p);",
                    expected: Expected::Error(Diagnostic { code: "42830", message: "number of referencing and referenced columns for foreign key disagree", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE p;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table p because other objects depend on it", detail: r#"constraint c_pid_fkey on table c depends on table p
constraint c_pcode_fkey on table c depends on table p"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "TRUNCATE p;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "cannot truncate a table referenced in a foreign key constraint", detail: r#"Table "c" references "p"."#, hint: r#"Truncate table "c" at the same time, or use TRUNCATE ... CASCADE."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE p CASCADE;",
                    expected: Expected::Tag("DROP TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to 2 other objects", detail: r#"drop cascades to constraint c_pid_fkey on table c
drop cascades to constraint c_pcode_fkey on table c"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO c VALUES (9, 99, 'q');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "self-referencing keys, NOT VALID, and SET DEFAULT",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE s (id INT PRIMARY KEY, parent INT REFERENCES s ON DELETE SET NULL);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO s VALUES (1, NULL), (2, 1), (3, 2);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM s WHERE id = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM s ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("parent", INT4)],
                        rows: &[
                            &[T("2"), Null],
                            &[T("3"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO s VALUES (4, 9);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "s" violates foreign key constraint "s_parent_fkey""#, detail: r#"Key (parent)=(9) is not present in table "s"."#, schema: "public", table: "s", constraint: "s_parent_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE e (id INT PRIMARY KEY, pid INT DEFAULT 1);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO e VALUES (1, 1), (2, 7);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE e ADD CONSTRAINT e_fk FOREIGN KEY (pid) REFERENCES s;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "e" violates foreign key constraint "e_fk""#, detail: r#"Key (pid)=(1) is not present in table "s"."#, schema: "public", table: "e", constraint: "e_fk", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE e ADD CONSTRAINT e_fk FOREIGN KEY (pid) REFERENCES s NOT VALID;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO e VALUES (3, 8);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "e" violates foreign key constraint "e_fk""#, detail: r#"Key (pid)=(8) is not present in table "s"."#, schema: "public", table: "e", constraint: "e_fk", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE e DROP CONSTRAINT e_fk;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO e VALUES (3, 8);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM e;",
                    expected: Expected::Tag("DELETE 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO s VALUES (1, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE e ADD FOREIGN KEY (pid) REFERENCES s ON DELETE SET DEFAULT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO e VALUES (1, 2), (2, 3);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM s WHERE id = 2;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM e ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("pid", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM s ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("parent", INT4)],
                        rows: &[
                            &[T("1"), Null],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE s DROP CONSTRAINT s_pkey;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop constraint s_pkey on table s because other objects depend on it", detail: r#"constraint s_parent_fkey on table s depends on index s_pkey
constraint e_pid_fkey on table e depends on index s_pkey"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE s DROP CONSTRAINT s_pkey CASCADE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to 2 other objects", detail: r#"drop cascades to constraint s_parent_fkey on table s
drop cascades to constraint e_pid_fkey on table e"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO e VALUES (9, 99);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "MATCH FULL, RESTRICT, and renames",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE m (a INT, b INT, UNIQUE (a, b));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE mc (a INT, b INT, FOREIGN KEY (a, b) REFERENCES m (a, b) MATCH FULL);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mc VALUES (1, NULL);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "mc" violates foreign key constraint "mc_a_b_fkey""#, detail: "MATCH FULL does not allow mixing of null and nonnull key values.", schema: "public", table: "mc", constraint: "mc_a_b_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO mc VALUES (NULL, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE r (id INT PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE rc (rid INT REFERENCES r ON DELETE RESTRICT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO r VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rc VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM r;",
                    expected: Expected::Error(Diagnostic { code: "23001", message: r#"update or delete on table "r" violates RESTRICT setting of foreign key constraint "rc_rid_fkey" on table "rc""#, detail: r#"Key (id)=(1) is referenced from table "rc"."#, schema: "public", table: "rc", constraint: "rc_rid_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE rc RENAME COLUMN rid TO r_id;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE r RENAME TO r2;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM r2;",
                    expected: Expected::Error(Diagnostic { code: "23001", message: r#"update or delete on table "r2" violates RESTRICT setting of foreign key constraint "rc_rid_fkey" on table "rc""#, detail: r#"Key (id)=(1) is referenced from table "rc"."#, schema: "public", table: "rc", constraint: "rc_rid_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rc VALUES (5);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "rc" violates foreign key constraint "rc_rid_fkey""#, detail: r#"Key (r_id)=(5) is not present in table "r2"."#, schema: "public", table: "rc", constraint: "rc_rid_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE rc RENAME CONSTRAINT rc_rid_fkey TO rc_fk;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO rc VALUES (5);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "rc" violates foreign key constraint "rc_fk""#, detail: r#"Key (r_id)=(5) is not present in table "r2"."#, schema: "public", table: "rc", constraint: "rc_fk", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE r2;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop table r2 because other objects depend on it", detail: "constraint rc_fk on table rc depends on table r2", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE rc, r2;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dropping indexes that foreign keys use",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE q (id INT PRIMARY KEY, u INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE UNIQUE INDEX q_u ON q (u);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE qc (a INT, b INT REFERENCES q (u));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX qc_b ON qc (b);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP INDEX qc_b;",
                    expected: Expected::Tag("DROP INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP INDEX q_u;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop index q_u because other objects depend on it", detail: "constraint qc_b_fkey on table qc depends on index q_u", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP INDEX q_u CASCADE;",
                    expected: Expected::Tag("DROP INDEX"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to constraint qc_b_fkey on table qc", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO qc VALUES (1, 77);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "keys of types that convert to each other",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE tp (i2 INT2 PRIMARY KEY);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE tc (i4 INT4 REFERENCES tp);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tp VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tc VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tc VALUES (65536);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "tc" violates foreign key constraint "tc_i4_fkey""#, detail: r#"Key (i4)=(65536) is not present in table "tp"."#, schema: "public", table: "tc", constraint: "tc_i4_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM tp;",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"update or delete on table "tp" violates foreign key constraint "tc_i4_fkey" on table "tc""#, detail: r#"Key (i2)=(1) is still referenced from table "tc"."#, schema: "public", table: "tc", constraint: "tc_i4_fkey", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dropping columns that foreign keys use",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE p (id INT PRIMARY KEY, u INT UNIQUE, a INT, b INT, UNIQUE (a, b));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE c (id INT PRIMARY KEY, pid INT REFERENCES p (id), uid INT REFERENCES p (u), x INT, y INT, FOREIGN KEY (x, y) REFERENCES p (a, b));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE p DROP COLUMN a;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop column a of table p because other objects depend on it", detail: "constraint c_x_y_fkey on table c depends on column a of table p", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE p DROP COLUMN id;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop column id of table p because other objects depend on it", detail: "constraint c_pid_fkey on table c depends on column id of table p", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE p DROP COLUMN u CASCADE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to constraint c_uid_fkey on table c", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c DROP COLUMN x;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO c VALUES (2, NULL, 5, 7);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE s (id INT PRIMARY KEY, parent INT REFERENCES s (id));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE s DROP COLUMN id;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop column id of table s because other objects depend on it", detail: "constraint s_parent_fkey on table s depends on column id of table s", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE s DROP COLUMN parent;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE s DROP COLUMN id;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO s VALUES (1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "INSERT has more expressions than target columns", position: 23, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM s;",
                    expected: Expected::Tag("SELECT 0"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "dropping a column that several foreign keys use",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE p (id INT PRIMARY KEY, u INT UNIQUE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE c (id INT PRIMARY KEY, uid INT REFERENCES p (u));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE c2 (id INT PRIMARY KEY, uid INT CONSTRAINT c2_fk REFERENCES p (u));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE p DROP COLUMN u;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop column u of table p because other objects depend on it", detail: r#"constraint c_uid_fkey on table c depends on column u of table p
constraint c2_fk on table c2 depends on column u of table p"#, hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE p DROP COLUMN u CASCADE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to 2 other objects", detail: r#"drop cascades to constraint c_uid_fkey on table c
drop cascades to constraint c2_fk on table c2"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "several foreign keys that a row violates",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE p (a INT UNIQUE, b INT UNIQUE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE c (a INT, b INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT fk_b FOREIGN KEY (b) REFERENCES p (b);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE c ADD CONSTRAINT fk_a FOREIGN KEY (a) REFERENCES p (a);",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO c VALUES (1, 1);",
                    expected: Expected::Error(Diagnostic { code: "23503", message: r#"insert or update on table "c" violates foreign key constraint "fk_b""#, detail: r#"Key (b)=(1) is not present in table "p"."#, schema: "public", table: "c", constraint: "fk_b", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
