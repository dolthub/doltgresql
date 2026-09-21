// Copyright 2024 Dolthub, Inc.
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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

func TestCreateTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			// https://github.com/dolthub/doltgresql/issues/2580
			Name: "create table with UTF8 identifiers",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE foo😏(data🍆 TEXT);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE INDEX idx🍤 ON foo😏(data🍆);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `Insert into foo😏 (data🍆) VALUES ('foo');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT data🍆 FROM foo😏;`,
					Expected: []sql.Row{{"foo"}},
				},
			},
		},
		{
			Name: "create table with primary key",
			Assertions: []ScriptTestAssertion{
				{
					// TODO: we don't currently have a way to check for warnings in these tests, but this query was incorrectly
					//  producing a warning. Would be nice to assert no warnings on most queries.
					Query: "create table employees (" +
						"    id int8," +
						"    last_name text," +
						"    first_name text," +
						"    primary key(id));",
				},
				{
					Query: "insert into employees (id, last_name, first_name) values (1, 'Doe', 'John');",
				},
				{
					Query: "select * from employees;",
					Expected: []sql.Row{
						{1, "Doe", "John"},
					},
				},
				{
					// Test that the PK constraint shows up in the information schema
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'employees'::regclass AND contype = 'p';",
					Expected: []sql.Row{{"employees_pkey"}},
				},
				{
					Query:    "ALTER TABLE employees DROP CONSTRAINT employees_pkey;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			// TODO: We don't currently support storing a custom name for a primary key constraint.
			Skip: true,
			Name: "create table with primary key, using custom constraint name",
			SetUpScript: []string{
				"CREATE TABLE users (id SERIAL, name TEXT, CONSTRAINT users_primary_key PRIMARY KEY (id));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'users'::regclass AND contype = 'p';",
					Expected: []sql.Row{{"users_primary_key"}},
				},
				{
					Query:    "ALTER TABLE users DROP CONSTRAINT users_primary_key;",
					Expected: []sql.Row{{types.NewOkResult(0)}},
				},
			},
		},
		{
			Name: "Create table with column default expression using function",
			Assertions: []ScriptTestAssertion{
				{
					// Test with a function in the column default expression
					Query:    "create table t1 (pk int primary key, c1 TEXT default length('Hello World!'));",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into t1(pk) values (1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from t1;",
					Expected: []sql.Row{{1, "12"}},
				},
			},
		},
		{
			Name: "Create table with table check constraint",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE products (name text, price numeric, discounted_price numeric, CHECK (price > discounted_price));`,
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into products values ('apple', 1.20, 0.80);",
					Expected: []sql.Row{},
				},
				{
					// TODO: the correct error message: `new row for relation "products" violates check constraint "products_chk_rqcthh8j"`
					Query:       "insert into products values ('peach', 1.20, 1.80);",
					ExpectedErr: `Check constraint "products_chk_`,
				},
				{
					Query:    "select * from products;",
					Expected: []sql.Row{{"apple", Numeric("1.20"), Numeric("0.80")}},
				},
			},
		},
		{
			Name: "Create table with column check constraint",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "create table mytbl (pk int, v1 int constraint v1constraint check (v1 < 100));",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into mytbl values (1, 20);",
					Expected: []sql.Row{},
				},
				{
					Query:       "insert into mytbl values (2, 200);",
					ExpectedErr: `Check constraint "v1constraint" violated`,
				},
				{
					Query:    "select * from mytbl;",
					Expected: []sql.Row{{1, 20}},
				},
			},
		},
		{
			Name: "check constraint with a function",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TABLE mytbl (a text CHECK (length(a) > 2) PRIMARY KEY, b text);",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into mytbl values ('abc', 'def');",
					Expected: []sql.Row{},
				},
				{
					Query:       "insert into mytbl values ('de', 'abc');",
					ExpectedErr: `Check constraint "mytbl_chk_`,
				},
				{
					Query:    "select * from mytbl;",
					Expected: []sql.Row{{"abc", "def"}},
				},
			},
		},
		{
			Skip: true, // TODO: vitess does not support multiple check constraint on a single column
			Name: "Create table with multiple check constraints on a single column",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "create table mytbl (pk int, v1 int constraint v1constraint check (v1 < 100) check (v1 > 10));",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into mytbl values (1, 20);",
					Expected: []sql.Row{},
				},
				{
					Query:       "insert into mytbl values (2, 200);",
					ExpectedErr: `Check constraint "v1constraint" violated`,
				},
				{
					Query:       "insert into mytbl values (3, 5);",
					ExpectedErr: `Check constraint "mytbl_chk_`,
				},
				{
					Query:    "select * from mytbl;",
					Expected: []sql.Row{{1, 20}},
				},
			},
		},
		{
			Name: "Create table with a check constraints on a single column and a table check constraint",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "create table mytbl (pk int, v1 int constraint v1constraint check (v1 < 100), check (v1 > 10));",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into mytbl values (1, 20);",
					Expected: []sql.Row{},
				},
				{
					Query:       "insert into mytbl values (2, 200);",
					ExpectedErr: `Check constraint "v1constraint" violated`,
				},
				{
					Query:       "insert into mytbl values (3, 5);",
					ExpectedErr: `Check constraint "mytbl_chk_`,
				},
				{
					Query:    "select * from mytbl;",
					Expected: []sql.Row{{1, 20}},
				},
			},
		},
		{
			Name: "create table with generated column",
			SetUpScript: []string{
				"create table t1 (a int primary key, b int, c int generated always as (a + b) stored);",
				"insert into t1 (a, b) values (1, 2);",
				"create table t2 (a int primary key, b int, c int generated always as (b * 10) stored);",
				"insert into t2 (a, b) values (1, 2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select * from t1;",
					Expected: []sql.Row{{1, 2, 3}},
				},
				{
					Query:    "select * from t2;",
					Expected: []sql.Row{{1, 2, 20}},
				},
			},
		},
		{
			Name: "create table with function in generated column",
			SetUpScript: []string{
				"create table t1 (a varchar(10) primary key, b varchar(10), c varchar(20) generated always as (concat(a,b)) stored);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "insert into t1 (a, b) values ('foo', 'bar');",
				},
				{
					Query:    "select * from t1;",
					Expected: []sql.Row{{"foo", "bar", "foobar"}},
				},
			},
		},
		{
			Name: "generated column with complex expression",
			SetUpScript: []string{
				`create table t1 (a varchar(10) primary key,
				b varchar(20) generated always as 
				    ((
				        ("substring"(TRIM(BOTH FROM a), '([^ ]+)$'::text) || ' '::text)
				          || "substring"(TRIM(BOTH FROM a), '^([^ ]+)'::text)
				    )) stored
				);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "insert into t1 (a) values (' foo ');",
				},
				{
					Query:    "select * from t1;",
					Expected: []sql.Row{{" foo ", "foo foo"}},
				},
			},
		},
		{
			Name: "generated column with reference to another column",
			SetUpScript: []string{
				`create table t1 (
    			a varchar(10) primary key,
    			b varchar(20),
				  b_not_null bool generated always as ((b is not null)) stored
				);`,
				"insert into t1 (a, b) values ('foo', 'bar');",
				"insert into t1 (a) values ('foo2');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from t1 order by a;",
					Expected: []sql.Row{
						{"foo", "bar", "t"},
						{"foo2", nil, "f"},
					},
				},
			},
		},
		{
			Name: "generated column with space in column name",
			SetUpScript: []string{
				`create table t1 (
    			a varchar(10) primary key,
    			"b 2" varchar(20),
				  b_not_null bool generated always as (("b 2" is not null)) stored
				);`,
				`insert into t1 (a, "b 2") values ('foo', 'bar');`,
				"insert into t1 (a) values ('foo2');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from t1 order by a;",
					Expected: []sql.Row{
						{"foo", "bar", "t"},
						{"foo2", nil, "f"},
					},
				},
			},
		},
		{
			Name: "primary key GENERATED ALWAYS AS IDENTITY",
			SetUpScript: []string{
				`create table t1 (
    			a BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
				  b varchar(100)
				);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "insert into t1 (b) values ('foo') returning a;",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query:       "insert into t1 (a, b) values (2, 'foo') returning a;",
					ExpectedErr: "The value specified for generated column \"a\" in table \"t1\" is not allowed",
				},
			},
		},
		{
			Name: "create table with default value",
			SetUpScript: []string{
				"create table t1 (a varchar(10) primary key, b varchar(10) default (concat('foo', 'bar')));",
				"insert into t1 (a) values ('abc');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select * from t1;",
					Expected: []sql.Row{{"abc", "foobar"}},
				},
			},
		},
		{
			Name: "create table with collation",
			SetUpScript: []string{
				`CREATE TABLE collate_test1 (
    a int,
        b text COLLATE "en-x-icu" NOT NULL
        )`,
				"insert into collate_test1 (a, b) values (1, 'foo');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select * from collate_test1;",
					Expected: []sql.Row{{1, "foo"}},
				},
			},
		},
		{
			Name: "inline comments",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE TABLE inline_comments (
	a int,
	b int,
	c int, -- comment on end of line
	CONSTRAINT check_b CHECK (b IS NULL OR b = 'a'),
	CONSTRAINT check_a CHECK (a IS NOT NULL AND a = 7)
);`,
				},
				{
					Query: `CREATE TABLE block_comments (
	a int,
	b /* block comment */ /* one more thing */ int, -- comment on end of line
	c int, -- comment on end of line /* block comment */
	CONSTRAINT check_b CHECK (b IS NULL OR b = 'a'),
	CONSTRAINT check_a CHECK (a IS NOT NULL AND a = 7)
);`,
				},
			},
		},
		{
			Name: "create temporary table with serial column",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TEMP TABLE temp (id serial primary key)",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "table with check constraint with ANY expression",
			SetUpScript: []string{
				`CREATE TABLE location (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(100),
    CONSTRAINT location_type_check CHECK (((type)::text = ANY ((ARRAY['Внутренни'::character varying, 'Покупатель'::character varying, 'Поставщик'::character varying])::text[])))
);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "insert into location values (1, 'Склад Москва', 'Внутренни'), (2, 'Склад Спб', null);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM location;",
					Expected: []sql.Row{
						{1, "Склад Москва", "Внутренни"},
						{2, "Склад Спб", nil},
					},
				},
			},
		},
		{
			Name: "Table names must be unique across all relation types",
			SetUpScript: []string{
				"CREATE TABLE existing_tbl (pk int PRIMARY KEY, v1 int);",
				"CREATE SEQUENCE seq1;",
				"CREATE VIEW view1 AS SELECT pk FROM existing_tbl;",
				"CREATE INDEX idx1 ON existing_tbl (v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "CREATE TABLE existing_tbl (c1 int);",
					ExpectedErr: `relation "existing_tbl" already exists`,
				},
				{
					Query:    "CREATE TABLE IF NOT EXISTS existing_tbl (c1 int);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE TABLE seq1 (c1 int);",
					ExpectedErr: `relation "seq1" already exists`,
				},
				{
					Query:    "CREATE TABLE IF NOT EXISTS seq1 (c1 int);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE TABLE view1 (c1 int);",
					ExpectedErr: `relation "view1" already exists`,
				},
				{
					Query:    "CREATE TABLE IF NOT EXISTS view1 (c1 int);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE TABLE idx1 (c1 int);",
					ExpectedErr: `relation "idx1" already exists`,
				},
				{
					Query:    "CREATE TABLE IF NOT EXISTS idx1 (c1 int);",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "nested parentheses are kept in default and check expressions",
			SetUpScript: []string{
				"CREATE TABLE t3324 (a INT PRIMARY KEY, b INT DEFAULT (1 + 1) * 2, c INT DEFAULT 2 * (3 + 1) + 1, d INT DEFAULT -(1 + 1), e INT GENERATED ALWAYS AS ((a + 1) * 2) STORED, f INT DEFAULT (((1 + 2)) * ((3))), g INT DEFAULT 10 - (4 - 1), h INT DEFAULT (2 + 3) % 4, i BOOLEAN DEFAULT (NOT (1 = 1 AND 2 = 2)), j INT DEFAULT abs(1 - 3) * 2, k TEXT DEFAULT ('a' || 'b') || 'c', l INT DEFAULT (1 + 2)::INT * 2, m INT DEFAULT -(-1), n BOOLEAN DEFAULT ((1 IS NULL) IS NULL), CONSTRAINT chk3324 CHECK (((a + 1) * 2) > 3), CONSTRAINT chk3324b CHECK (NOT (a = 0 OR a + 1 = 0) AND a - (a - 1) = 1));",
				"INSERT INTO t3324 (a) VALUES (1);",
				"ALTER TABLE t3324 ADD COLUMN o INT DEFAULT (1 + 1) * 2;",
				"ALTER TABLE t3324 ALTER COLUMN o SET DEFAULT 2 * (1 + 1) + 1;",
				"ALTER TABLE t3324 ADD CONSTRAINT chk3324c CHECK ((a * 2) - 1 > 0);",
				"INSERT INTO t3324 (a) VALUES (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM t3324 ORDER BY a;",
					Expected: []sql.Row{{1, 4, 9, -2, 4, 9, 7, 1, "f", 4, "abc", 6, 1, "f", 4}, {2, 4, 9, -2, 6, 9, 7, 1, "f", 4, "abc", 6, 1, "f", 5}},
				},
				{
					Query:       "INSERT INTO t3324 (a) VALUES (0);",
					ExpectedErr: "violated",
				},
			},
		},
		{
			Name: "nested parentheses in generated and check expressions survive ALTER TABLE",
			SetUpScript: []string{
				"CREATE TABLE t3324b (a INT NOT NULL, b INT DEFAULT (1 + 1) * 2, c INT GENERATED ALWAYS AS ((a + 1) * 2) STORED, d INT GENERATED ALWAYS AS (2 * (a + 1) - (a - 1)) STORED, CHECK ((a + 1) * 2 > 3));",
				"INSERT INTO t3324b (a) VALUES (1);",
				"ALTER TABLE t3324b ADD PRIMARY KEY (a);",
				"INSERT INTO t3324b (a) VALUES (2);",
				"ALTER TABLE t3324b ADD COLUMN e INT DEFAULT (3 + 4) * 5;",
				"INSERT INTO t3324b (a) VALUES (3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO t3324b (a) VALUES (0);",
					ExpectedErr: "violated",
				},
				{
					Query:    "SELECT * FROM t3324b ORDER BY a;",
					Expected: []sql.Row{{1, 4, 4, 4, 35}, {2, 4, 6, 5, 35}, {3, 4, 8, 6, 35}},
				},
			},
		},
		{
			Name: "nested parentheses are kept on the right side and under unary minus in generated expressions",
			SetUpScript: []string{
				"CREATE TABLE t3324c (a INT, b INT GENERATED ALWAYS AS (-(a + 1)) STORED, c INT GENERATED ALWAYS AS (2 * (a + 1)) STORED, d INT GENERATED ALWAYS AS (a - (1 - 2)) STORED, e INT DEFAULT ((1 + 2) * 3));",
				"INSERT INTO t3324c (a) VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM t3324c;",
					Expected: []sql.Row{{1, -2, 4, 2, 9}},
				},
			},
		},
		{
			Name: "nested parentheses are kept around LIKE, IN, subscripts, and double negation",
			SetUpScript: []string{
				"CREATE TABLE t3324d (a TEXT, b BOOLEAN DEFAULT (('abc' LIKE 'a%') IS NOT NULL), c BOOLEAN DEFAULT ((1 + 1) IN (2, 3)), d INT DEFAULT ((ARRAY[1] || ARRAY[2])[1]), e INT DEFAULT -(-1), f INT DEFAULT (- (- 2)), CHECK ((a || 'x') LIKE 'a%'));",
				"INSERT INTO t3324d (a) VALUES ('a');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO t3324d (a) VALUES ('b');",
					ExpectedErr: "violated",
				},
				{
					Query:    "SELECT * FROM t3324d;",
					Expected: []sql.Row{{"a", "t", "t", 1, 1, 2}},
				},
			},
		},
		{
			Name: "nested parentheses are kept in check constraints using NOT, AND, OR, BETWEEN, CAST, and LIKE",
			SetUpScript: []string{
				"CREATE TABLE tc3324 (a INT, b INT, CONSTRAINT c1 CHECK (NOT (a = 0 OR a + 1 = 0) AND a - (a - 1) = 1), CONSTRAINT c2 CHECK ((a BETWEEN 1 AND 10) OR (b IS NULL)), CONSTRAINT c3 CHECK (((a + 1) * 2) > 3), CONSTRAINT c4 CHECK (NOT ((a + b) > 100)), CONSTRAINT c5 CHECK (CAST(a + 1 AS INT) > 0), CONSTRAINT c6 CHECK ((a || '') NOT LIKE 'x%'));",
				"INSERT INTO tc3324 VALUES (1, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO tc3324 VALUES (0, 1);",
					ExpectedErr: "c1",
				},
				{
					Query:       "INSERT INTO tc3324 VALUES (50, 60);",
					ExpectedErr: "c2",
				},
				{
					Query:       "INSERT INTO tc3324 VALUES (-1, 5);",
					ExpectedErr: "c1",
				},
				{
					Query:       "INSERT INTO tc3324 VALUES (5, 96);",
					ExpectedErr: "c4",
				},
				{
					Query:    "INSERT INTO tc3324 VALUES (2, 3);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM tc3324 ORDER BY a;",
					Expected: []sql.Row{{1, nil}, {2, 3}},
				},
			},
		},
		{
			Name: "generated column and default with nested parentheses match the equivalent SELECT expressions",
			SetUpScript: []string{
				"CREATE TABLE tx3324 (a INT, b INT GENERATED ALWAYS AS ((a + 1) * 2) STORED, c INT DEFAULT ((1 + 2) * 3));",
				"INSERT INTO tx3324 (a) VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT a, b, (a + 1) * 2 AS expected_b, c, (1 + 2) * 3 AS expected_c FROM tx3324;",
					Expected: []sql.Row{{1, 4, 4, 9, 9}},
				},
			},
		},
		{
			Name: "named column constraints",
			SetUpScript: []string{
				"CREATE TABLE t3332_issue (id INTEGER CONSTRAINT id_required NOT NULL, v TEXT);",
				"CREATE TABLE t3332 (id INT CONSTRAINT id_nn NOT NULL, u INT CONSTRAINT u_uni UNIQUE, d INT CONSTRAINT d_def DEFAULT 5, n INT CONSTRAINT n_null NULL, PRIMARY KEY (id));",
				"ALTER TABLE t3332 ADD COLUMN w INT CONSTRAINT w_nn NOT NULL DEFAULT 1 CONSTRAINT w_uni UNIQUE;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT table_name FROM information_schema.tables WHERE table_name = 't3332_issue';",
					Expected: []sql.Row{{"t3332_issue"}},
				},
				{
					Query:    "INSERT INTO t3332 (id, u) VALUES (1, 1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t3332 (id, u, w) VALUES (2, 2, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM t3332 ORDER BY id;",
					Expected: []sql.Row{{1, 1, 5, nil, 1}, {2, 2, 5, nil, 2}},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE tablename = 't3332' ORDER BY indexname;",
					Expected: []sql.Row{{"t3332_pkey"}, {"u_uni"}, {"w_uni"}},
				},
				{
					Query:    "SELECT conname, contype FROM pg_constraint WHERE conrelid = 't3332'::regclass ORDER BY conname;",
					Expected: []sql.Row{{"t3332_pkey", "p"}, {"u_uni", "u"}, {"w_uni", "u"}},
				},
				{
					Query:       "INSERT INTO t3332 (id, u, w) VALUES (3, 1, 3);",
					ExpectedErr: "duplicate unique key",
				},
				{
					Query:       "INSERT INTO t3332 (id, u, w) VALUES (3, 3, 2);",
					ExpectedErr: "duplicate unique key",
				},
				{
					Query:       "INSERT INTO t3332 (id, u, w) VALUES (NULL, 4, 4);",
					ExpectedErr: "non-nullable",
				},
			},
		},
	})
}

func TestCreateTableInherit(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Create table with inheritance",
			SetUpScript: []string{
				"create table t1 (a int);",
				"create table t2 (b int);",
				"create table t3 (c int);",
				"create table t11 (a int);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "create table t4 (d int) inherits (t1, t2, t3);",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into t4(a, b, c, d) values (1, 2, 3, 4);",
					Expected: []sql.Row{},
				},
				{
					Query: "select * from t4;",
					Expected: []sql.Row{
						{1, 2, 3, 4},
					},
				},
				{
					Query:    "create table t111 () inherits (t1, t11);",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into t111(a) values (1);",
					Expected: []sql.Row{},
				},
				{
					Query: "select * from t111;",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query:    "create table t1t1 (a int) inherits (t1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into t1t1(a) values (1);",
					Expected: []sql.Row{},
				},
				{
					Query: "select * from t1t1;",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query:    "create table TT1t1 (A int) inherits (t1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into TT1t1(a) values (1);",
					Expected: []sql.Row{},
				},
				{
					Query: "select * from TT1t1;",
					Expected: []sql.Row{
						{1},
					},
				},
			},
		},
	})
}
