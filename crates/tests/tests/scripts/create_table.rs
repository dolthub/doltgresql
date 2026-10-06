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
fn test_create_table() {
    run_scripts(&[
        ScriptTest {
            name: "create table with UTF8 identifiers",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE foo😏(data🍆 TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx🍤 ON foo😏(data🍆);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "Insert into foo😏 (data🍆) VALUES ('foo');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT data🍆 FROM foo😏;",
                    expected: Expected::Rows {
                        columns: &[Column("data🍆", TEXT)],
                        rows: &[
                            &[T("foo")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create table with primary key",
            assertions: &[
                ScriptTestAssertion {
                    query: "create table employees (    id int8,    last_name text,    first_name text,    primary key(id));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into employees (id, last_name, first_name) values (1, 'Doe', 'John');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from employees;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("last_name", TEXT), Column("first_name", TEXT)],
                        rows: &[
                            &[T("1"), T("Doe"), T("John")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_constraint WHERE conrelid = 'employees'::regclass AND contype = 'p';",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[
                            &[T("employees_pkey")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE employees DROP CONSTRAINT employees_pkey;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create table with primary key, using custom constraint name",
            set_up_script: &[
                "CREATE TABLE users (id SERIAL, name TEXT, CONSTRAINT users_primary_key PRIMARY KEY (id));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_constraint WHERE conrelid = 'users'::regclass AND contype = 'p';",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[
                            &[T("users_primary_key")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE users DROP CONSTRAINT users_primary_key;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Create table with column default expression using function",
            assertions: &[
                ScriptTestAssertion {
                    query: "create table t1 (pk int primary key, c1 TEXT default length('Hello World!'));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t1(pk) values (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT)],
                        rows: &[
                            &[T("1"), T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Create table with table check constraint",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE products (name text, price numeric, discounted_price numeric, CHECK (price > discounted_price));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into products values ('apple', 1.20, 0.80);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into products values ('peach', 1.20, 1.80);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "products" violates check constraint "products_check""#, detail: "Failing row contains (peach, 1.20, 1.80).", schema: "public", table: "products", constraint: "products_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from products;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT), Column("price", NUMERIC), Column("discounted_price", NUMERIC)],
                        rows: &[
                            &[T("apple"), T("1.20"), T("0.80")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Create table with column check constraint",
            assertions: &[
                ScriptTestAssertion {
                    query: "create table mytbl (pk int, v1 int constraint v1constraint check (v1 < 100));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values (1, 20);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values (2, 200);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "mytbl" violates check constraint "v1constraint""#, detail: "Failing row contains (2, 200).", schema: "public", table: "mytbl", constraint: "v1constraint", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from mytbl;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "check constraint with a function",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE mytbl (a text CHECK (length(a) > 2) PRIMARY KEY, b text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values ('abc', 'def');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values ('de', 'abc');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "mytbl" violates check constraint "mytbl_a_check""#, detail: "Failing row contains (de, abc).", schema: "public", table: "mytbl", constraint: "mytbl_a_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from mytbl;",
                    expected: Expected::Rows {
                        columns: &[Column("a", TEXT), Column("b", TEXT)],
                        rows: &[
                            &[T("abc"), T("def")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Create table with multiple check constraints on a single column",
            assertions: &[
                ScriptTestAssertion {
                    query: "create table mytbl (pk int, v1 int constraint v1constraint check (v1 < 100) check (v1 > 10));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values (1, 20);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values (2, 200);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "mytbl" violates check constraint "v1constraint""#, detail: "Failing row contains (2, 200).", schema: "public", table: "mytbl", constraint: "v1constraint", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values (3, 5);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "mytbl" violates check constraint "mytbl_v1_check""#, detail: "Failing row contains (3, 5).", schema: "public", table: "mytbl", constraint: "mytbl_v1_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from mytbl;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Create table with a check constraints on a single column and a table check constraint",
            assertions: &[
                ScriptTestAssertion {
                    query: "create table mytbl (pk int, v1 int constraint v1constraint check (v1 < 100), check (v1 > 10));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values (1, 20);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values (2, 200);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "mytbl" violates check constraint "v1constraint""#, detail: "Failing row contains (2, 200).", schema: "public", table: "mytbl", constraint: "v1constraint", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into mytbl values (3, 5);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "mytbl" violates check constraint "mytbl_v1_check""#, detail: "Failing row contains (3, 5).", schema: "public", table: "mytbl", constraint: "mytbl_v1_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from mytbl;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", INT4)],
                        rows: &[
                            &[T("1"), T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create table with generated column",
            set_up_script: &[
                "create table t1 (a int primary key, b int, c int generated always as (a + b) stored);",
                "insert into t1 (a, b) values (1, 2);",
                "create table t2 (a int primary key, b int, c int generated always as (b * 10) stored);",
                "insert into t2 (a, b) values (1, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: concat is not immutable in Postgres, so it cannot define a generated column; || is.
        ScriptTest {
            name: "create table with function in generated column",
            set_up_script: &[
                "create table t1 (a varchar(10) primary key, b varchar(10), c varchar(20) generated always as (a || b) stored);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "insert into t1 (a, b) values ('foo', 'bar');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", VARCHAR), Column("b", VARCHAR), Column("c", VARCHAR)],
                        rows: &[
                            &[T("foo"), T("bar"), T("foobar")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated column with complex expression",
            set_up_script: &[
                r#"create table t1 (a varchar(10) primary key,
				b varchar(20) generated always as 
				    ((
				        ("substring"(TRIM(BOTH FROM a), '([^ ]+)$'::text) || ' '::text)
				          || "substring"(TRIM(BOTH FROM a), '^([^ ]+)'::text)
				    )) stored
				);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "insert into t1 (a) values (' foo ');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", VARCHAR), Column("b", VARCHAR)],
                        rows: &[
                            &[T(" foo "), T("foo foo")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated column with reference to another column",
            set_up_script: &[
                r#"create table t1 (
    			a varchar(10) primary key,
    			b varchar(20),
				  b_not_null bool generated always as ((b is not null)) stored
				);"#,
                "insert into t1 (a, b) values ('foo', 'bar');",
                "insert into t1 (a) values ('foo2');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from t1 order by a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", VARCHAR), Column("b", VARCHAR), Column("b_not_null", BOOL)],
                        rows: &[
                            &[T("foo"), T("bar"), T("t")],
                            &[T("foo2"), Null, T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated column with space in column name",
            set_up_script: &[
                r#"create table t1 (
    			a varchar(10) primary key,
    			"b 2" varchar(20),
				  b_not_null bool generated always as (("b 2" is not null)) stored
				);"#,
                r#"insert into t1 (a, "b 2") values ('foo', 'bar');"#,
                "insert into t1 (a) values ('foo2');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from t1 order by a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", VARCHAR), Column("b 2", VARCHAR), Column("b_not_null", BOOL)],
                        rows: &[
                            &[T("foo"), T("bar"), T("t")],
                            &[T("foo2"), Null, T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "primary key GENERATED ALWAYS AS IDENTITY",
            set_up_script: &[
                r#"create table t1 (
    			a BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
				  b varchar(100)
				);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "insert into t1 (b) values ('foo') returning a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "INSERT 0 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t1 (a, b) values (2, 'foo') returning a;",
                    expected: Expected::Error(Diagnostic { code: "428C9", message: r#"cannot insert a non-DEFAULT value into column "a""#, detail: r#"Column "a" is an identity column defined as GENERATED ALWAYS."#, hint: "Use OVERRIDING SYSTEM VALUE to override.", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create table with default value",
            set_up_script: &[
                "create table t1 (a varchar(10) primary key, b varchar(10) default (concat('foo', 'bar')));",
                "insert into t1 (a) values ('abc');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", VARCHAR), Column("b", VARCHAR)],
                        rows: &[
                            &[T("abc"), T("foobar")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create table with collation",
            set_up_script: &[
                r#"CREATE TABLE collate_test1 (
    a int,
        b text COLLATE "en-x-icu" NOT NULL
        )"#,
                "insert into collate_test1 (a, b) values (1, 'foo');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from collate_test1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("foo")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "inline comments",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE TABLE inline_comments (
	a int,
	b int,
	c int, -- comment on end of line
	CONSTRAINT check_b CHECK (b IS NULL OR b = 'a'),
	CONSTRAINT check_a CHECK (a IS NOT NULL AND a = 7)
);"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 126, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE TABLE block_comments (
	a int,
	b /* block comment */ /* one more thing */ int, -- comment on end of line
	c int, -- comment on end of line /* block comment */
	CONSTRAINT check_b CHECK (b IS NULL OR b = 'a'),
	CONSTRAINT check_a CHECK (a IS NOT NULL AND a = 7)
);"#,
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 212, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create temporary table with serial column",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TEMP TABLE temp (id serial primary key)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table with check constraint with ANY expression",
            set_up_script: &[
                r#"CREATE TABLE location (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(100),
    CONSTRAINT location_type_check CHECK (((type)::text = ANY ((ARRAY['Внутренни'::character varying, 'Покупатель'::character varying, 'Поставщик'::character varying])::text[])))
);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "insert into location values (1, 'Склад Москва', 'Внутренни'), (2, 'Склад Спб', null);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM location;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", VARCHAR), Column("type", VARCHAR)],
                        rows: &[
                            &[T("1"), T("Склад Москва"), T("Внутренни")],
                            &[T("2"), T("Склад Спб"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Table names must be unique across all relation types",
            set_up_script: &[
                "CREATE TABLE existing_tbl (pk int PRIMARY KEY, v1 int);",
                "CREATE SEQUENCE seq1;",
                "CREATE VIEW view1 AS SELECT pk FROM existing_tbl;",
                "CREATE INDEX idx1 ON existing_tbl (v1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE existing_tbl (c1 int);",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "existing_tbl" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE IF NOT EXISTS existing_tbl (c1 int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "existing_tbl" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE seq1 (c1 int);",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "seq1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE IF NOT EXISTS seq1 (c1 int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "seq1" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE view1 (c1 int);",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "view1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE IF NOT EXISTS view1 (c1 int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "view1" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE idx1 (c1 int);",
                    expected: Expected::Error(Diagnostic { code: "42P07", message: r#"relation "idx1" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE IF NOT EXISTS idx1 (c1 int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "idx1" already exists, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nested parentheses are kept in default and check expressions",
            set_up_script: &[
                "CREATE TABLE t3324 (a INT PRIMARY KEY, b INT DEFAULT (1 + 1) * 2, c INT DEFAULT 2 * (3 + 1) + 1, d INT DEFAULT -(1 + 1), e INT GENERATED ALWAYS AS ((a + 1) * 2) STORED, f INT DEFAULT (((1 + 2)) * ((3))), g INT DEFAULT 10 - (4 - 1), h INT DEFAULT (2 + 3) % 4, i BOOLEAN DEFAULT (NOT (1 = 1 AND 2 = 2)), j INT DEFAULT abs(1 - 3) * 2, k TEXT DEFAULT ('a' || 'b') || 'c', l INT DEFAULT (1 + 2)::INT * 2, m INT DEFAULT -(-1), n BOOLEAN DEFAULT ((1 IS NULL) IS NULL), CONSTRAINT chk3324 CHECK (((a + 1) * 2) > 3), CONSTRAINT chk3324b CHECK (NOT (a = 0 OR a + 1 = 0) AND a - (a - 1) = 1));",
                "INSERT INTO t3324 (a) VALUES (1);",
                "ALTER TABLE t3324 ADD COLUMN o INT DEFAULT (1 + 1) * 2;",
                "ALTER TABLE t3324 ALTER COLUMN o SET DEFAULT 2 * (1 + 1) + 1;",
                "ALTER TABLE t3324 ADD CONSTRAINT chk3324c CHECK ((a * 2) - 1 > 0);",
                "INSERT INTO t3324 (a) VALUES (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t3324 ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4), Column("d", INT4), Column("e", INT4), Column("f", INT4), Column("g", INT4), Column("h", INT4), Column("i", BOOL), Column("j", INT4), Column("k", TEXT), Column("l", INT4), Column("m", INT4), Column("n", BOOL), Column("o", INT4)],
                        rows: &[
                            &[T("1"), T("4"), T("9"), T("-2"), T("4"), T("9"), T("7"), T("1"), T("f"), T("4"), T("abc"), T("6"), T("1"), T("f"), T("4")],
                            &[T("2"), T("4"), T("9"), T("-2"), T("6"), T("9"), T("7"), T("1"), T("f"), T("4"), T("abc"), T("6"), T("1"), T("f"), T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3324 (a) VALUES (0);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t3324" violates check constraint "chk3324""#, detail: "Failing row contains (0, 4, 9, -2, 2, 9, 7, 1, f, 4, abc, 6, 1, f, 5).", schema: "public", table: "t3324", constraint: "chk3324", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nested parentheses in generated and check expressions survive ALTER TABLE",
            set_up_script: &[
                "CREATE TABLE t3324b (a INT NOT NULL, b INT DEFAULT (1 + 1) * 2, c INT GENERATED ALWAYS AS ((a + 1) * 2) STORED, d INT GENERATED ALWAYS AS (2 * (a + 1) - (a - 1)) STORED, CHECK ((a + 1) * 2 > 3));",
                "INSERT INTO t3324b (a) VALUES (1);",
                "ALTER TABLE t3324b ADD PRIMARY KEY (a);",
                "INSERT INTO t3324b (a) VALUES (2);",
                "ALTER TABLE t3324b ADD COLUMN e INT DEFAULT (3 + 4) * 5;",
                "INSERT INTO t3324b (a) VALUES (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t3324b (a) VALUES (0);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t3324b" violates check constraint "t3324b_a_check""#, detail: "Failing row contains (0, 4, 2, 3, 35).", schema: "public", table: "t3324b", constraint: "t3324b_a_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3324b ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4), Column("d", INT4), Column("e", INT4)],
                        rows: &[
                            &[T("1"), T("4"), T("4"), T("4"), T("35")],
                            &[T("2"), T("4"), T("6"), T("5"), T("35")],
                            &[T("3"), T("4"), T("8"), T("6"), T("35")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nested parentheses are kept on the right side and under unary minus in generated expressions",
            set_up_script: &[
                "CREATE TABLE t3324c (a INT, b INT GENERATED ALWAYS AS (-(a + 1)) STORED, c INT GENERATED ALWAYS AS (2 * (a + 1)) STORED, d INT GENERATED ALWAYS AS (a - (1 - 2)) STORED, e INT DEFAULT ((1 + 2) * 3));",
                "INSERT INTO t3324c (a) VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t3324c;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4), Column("d", INT4), Column("e", INT4)],
                        rows: &[
                            &[T("1"), T("-2"), T("4"), T("2"), T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nested parentheses are kept around LIKE, IN, subscripts, and double negation",
            set_up_script: &[
                "CREATE TABLE t3324d (a TEXT, b BOOLEAN DEFAULT (('abc' LIKE 'a%') IS NOT NULL), c BOOLEAN DEFAULT ((1 + 1) IN (2, 3)), d INT DEFAULT ((ARRAY[1] || ARRAY[2])[1]), e INT DEFAULT -(-1), f INT DEFAULT (- (- 2)), CHECK ((a || 'x') LIKE 'a%'));",
                "INSERT INTO t3324d (a) VALUES ('a');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t3324d (a) VALUES ('b');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t3324d" violates check constraint "t3324d_a_check""#, detail: "Failing row contains (b, t, t, 1, 1, 2).", schema: "public", table: "t3324d", constraint: "t3324d_a_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3324d;",
                    expected: Expected::Rows {
                        columns: &[Column("a", TEXT), Column("b", BOOL), Column("c", BOOL), Column("d", INT4), Column("e", INT4), Column("f", INT4)],
                        rows: &[
                            &[T("a"), T("t"), T("t"), T("1"), T("1"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nested parentheses are kept in check constraints using NOT, AND, OR, BETWEEN, CAST, and LIKE",
            set_up_script: &[
                "CREATE TABLE tc3324 (a INT, b INT, CONSTRAINT c1 CHECK (NOT (a = 0 OR a + 1 = 0) AND a - (a - 1) = 1), CONSTRAINT c2 CHECK ((a BETWEEN 1 AND 10) OR (b IS NULL)), CONSTRAINT c3 CHECK (((a + 1) * 2) > 3), CONSTRAINT c4 CHECK (NOT ((a + b) > 100)), CONSTRAINT c5 CHECK (CAST(a + 1 AS INT) > 0), CONSTRAINT c6 CHECK ((a || '') NOT LIKE 'x%'));",
                "INSERT INTO tc3324 VALUES (1, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO tc3324 VALUES (0, 1);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "tc3324" violates check constraint "c1""#, detail: "Failing row contains (0, 1).", schema: "public", table: "tc3324", constraint: "c1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tc3324 VALUES (50, 60);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "tc3324" violates check constraint "c2""#, detail: "Failing row contains (50, 60).", schema: "public", table: "tc3324", constraint: "c2", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tc3324 VALUES (-1, 5);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "tc3324" violates check constraint "c1""#, detail: "Failing row contains (-1, 5).", schema: "public", table: "tc3324", constraint: "c1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tc3324 VALUES (5, 96);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "tc3324" violates check constraint "c4""#, detail: "Failing row contains (5, 96).", schema: "public", table: "tc3324", constraint: "c4", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO tc3324 VALUES (2, 3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tc3324 ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), Null],
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
            name: "generated column and default with nested parentheses match the equivalent SELECT expressions",
            set_up_script: &[
                "CREATE TABLE tx3324 (a INT, b INT GENERATED ALWAYS AS ((a + 1) * 2) STORED, c INT DEFAULT ((1 + 2) * 3));",
                "INSERT INTO tx3324 (a) VALUES (1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT a, b, (a + 1) * 2 AS expected_b, c, (1 + 2) * 3 AS expected_c FROM tx3324;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("expected_b", INT4), Column("c", INT4), Column("expected_c", INT4)],
                        rows: &[
                            &[T("1"), T("4"), T("4"), T("9"), T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "named column constraints",
            set_up_script: &[
                "CREATE TABLE t3332_issue (id INTEGER CONSTRAINT id_required NOT NULL, v TEXT);",
                "CREATE TABLE t3332 (id INT CONSTRAINT id_nn NOT NULL, u INT CONSTRAINT u_uni UNIQUE, d INT CONSTRAINT d_def DEFAULT 5, n INT CONSTRAINT n_null NULL, PRIMARY KEY (id));",
                "ALTER TABLE t3332 ADD COLUMN w INT CONSTRAINT w_nn NOT NULL DEFAULT 1 CONSTRAINT w_uni UNIQUE;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT table_name FROM information_schema.tables WHERE table_name = 't3332_issue';",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", NAME)],
                        rows: &[
                            &[T("t3332_issue")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3332 (id, u) VALUES (1, 1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3332 (id, u, w) VALUES (2, 2, 2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3332 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4), Column("d", INT4), Column("n", INT4), Column("w", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("5"), Null, T("1")],
                            &[T("2"), T("2"), T("5"), Null, T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indexname FROM pg_indexes WHERE tablename = 't3332' ORDER BY indexname;",
                    expected: Expected::Rows {
                        columns: &[Column("indexname", NAME)],
                        rows: &[
                            &[T("t3332_pkey")],
                            &[T("u_uni")],
                            &[T("w_uni")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, contype FROM pg_constraint WHERE conrelid = 't3332'::regclass ORDER BY conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("contype", CHAR)],
                        rows: &[
                            &[T("t3332_pkey"), T("p")],
                            &[T("u_uni"), T("u")],
                            &[T("w_uni"), T("u")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3332 (id, u, w) VALUES (3, 1, 3);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "u_uni""#, detail: "Key (u)=(1) already exists.", schema: "public", table: "t3332", constraint: "u_uni", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3332 (id, u, w) VALUES (3, 3, 2);",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "w_uni""#, detail: "Key (w)=(2) already exists.", schema: "public", table: "t3332", constraint: "w_uni", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t3332 (id, u, w) VALUES (NULL, 4, 4);",
                    expected: Expected::Error(Diagnostic { code: "23502", message: r#"null value in column "id" of relation "t3332" violates not-null constraint"#, detail: "Failing row contains (null, 4, 5, null, 4).", schema: "public", table: "t3332", column: "id", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DEFERRABLE constraints",
            set_up_script: &[
                "CREATE TABLE p (id INTEGER PRIMARY KEY);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE a1 (x INTEGER REFERENCES p(id) DEFERRABLE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a2 (x INTEGER REFERENCES p(id) INITIALLY DEFERRED);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a3 (x INTEGER REFERENCES p(id) DEFERRABLE INITIALLY IMMEDIATE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a4 (x INTEGER REFERENCES p(id) NOT DEFERRABLE INITIALLY IMMEDIATE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a5 (x INTEGER REFERENCES p(id) NOT DEFERRABLE INITIALLY DEFERRED);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "constraint declared INITIALLY DEFERRED must be DEFERRABLE", position: 60, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a6 (x INTEGER UNIQUE DEFERRABLE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a7 (x INTEGER PRIMARY KEY DEFERRABLE INITIALLY IMMEDIATE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a8 (x INTEGER UNIQUE NOT DEFERRABLE INITIALLY IMMEDIATE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a9 (x INTEGER CHECK (x > 0) DEFERRABLE);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "misplaced DEFERRABLE clause", position: 42, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a10 (x INTEGER NOT NULL DEFERRABLE);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "misplaced DEFERRABLE clause", position: 38, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a11 (x INTEGER DEFAULT 1 DEFERRABLE);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "misplaced DEFERRABLE clause", position: 39, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a12 (x INTEGER NULL NOT DEFERRABLE);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "misplaced NOT DEFERRABLE clause", position: 34, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a13 (x INTEGER CHECK (x > 0) INITIALLY IMMEDIATE);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "misplaced INITIALLY IMMEDIATE clause", position: 43, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a14 (x INTEGER CHECK (x > 0) INITIALLY DEFERRED);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "misplaced INITIALLY DEFERRED clause", position: 43, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a15 (x INTEGER GENERATED ALWAYS AS (1) STORED DEFERRABLE);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "misplaced DEFERRABLE clause", position: 60, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b1 (x INTEGER, FOREIGN KEY (x) REFERENCES p(id) DEFERRABLE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b2 (x INTEGER, CONSTRAINT b2u UNIQUE (x) INITIALLY DEFERRED);",
                    expected: Expected::Tag("CREATE TABLE"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b3 (x INTEGER, PRIMARY KEY (x) NOT DEFERRABLE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b4 (x INTEGER, FOREIGN KEY (x) REFERENCES p(id) INITIALLY IMMEDIATE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b5 (x INTEGER, UNIQUE (x) NOT DEFERRABLE INITIALLY DEFERRED);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "constraint declared INITIALLY DEFERRED must be DEFERRABLE", position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b6 (x INTEGER, CHECK (x > 0) DEFERRABLE);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "CHECK constraints cannot be marked DEFERRABLE", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b7 (x INTEGER, CHECK (x > 0) INITIALLY DEFERRED);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "CHECK constraints cannot be marked DEFERRABLE", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b8 (x INTEGER, CHECK (x > 0) NOT DEFERRABLE INITIALLY DEFERRED);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "constraint declared INITIALLY DEFERRED must be DEFERRABLE", position: 58, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b9 (x INTEGER, CHECK (x > 0) NOT DEFERRABLE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b10 (x INTEGER, CHECK (x > 0) INITIALLY IMMEDIATE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DEFERRABLE constraints are stored",
            set_up_script: &[
                "CREATE TABLE p (id INTEGER PRIMARY KEY);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE a1 (x INTEGER REFERENCES p(id) DEFERRABLE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a2 (x INTEGER REFERENCES p(id) INITIALLY DEFERRED);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a3 (x INTEGER REFERENCES p(id) DEFERRABLE INITIALLY IMMEDIATE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a6 (x INTEGER UNIQUE DEFERRABLE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE a7 (x INTEGER PRIMARY KEY DEFERRABLE INITIALLY IMMEDIATE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b1 (x INTEGER, FOREIGN KEY (x) REFERENCES p(id) DEFERRABLE);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE b2 (x INTEGER, CONSTRAINT b2u UNIQUE (x) INITIALLY DEFERRED);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname, condeferrable, condeferred FROM pg_constraint WHERE connamespace = 'public'::regnamespace AND condeferrable ORDER BY conname;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("condeferrable", BOOL), Column("condeferred", BOOL)],
                        rows: &[
                            &[T("a1_x_fkey"), T("t"), T("f")],
                            &[T("a2_x_fkey"), T("t"), T("t")],
                            &[T("a3_x_fkey"), T("t"), T("f")],
                            &[T("a6_x_key"), T("t"), T("f")],
                            &[T("a7_pkey"), T("t"), T("f")],
                            &[T("b1_x_fkey"), T("t"), T("f")],
                            &[T("b2u"), T("t"), T("t")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_create_table_inherit() {
    run_scripts(&[
        ScriptTest {
            name: "Create table with inheritance",
            set_up_script: &[
                "create table t1 (a int);",
                "create table t2 (b int);",
                "create table t3 (c int);",
                "create table t11 (a int);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "create table t4 (d int) inherits (t1, t2, t3);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t4(a, b, c, d) values (1, 2, 3, 4);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t4;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4), Column("d", INT4)],
                        rows: &[
                            &[T("1"), T("2"), T("3"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "create table t111 () inherits (t1, t11);",
                    expected: Expected::Tag("CREATE TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"merging multiple inherited definitions of column "a""#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t111(a) values (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t111;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "create table t1t1 (a int) inherits (t1);",
                    expected: Expected::Tag("CREATE TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"merging column "a" with inherited definition"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into t1t1(a) values (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t1t1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "create table TT1t1 (A int) inherits (t1);",
                    expected: Expected::Tag("CREATE TABLE"),
                    notices: &[Diagnostic { code: "00000", message: r#"merging column "a" with inherited definition"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into TT1t1(a) values (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from TT1t1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
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
