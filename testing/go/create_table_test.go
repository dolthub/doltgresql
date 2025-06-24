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
					// TODO: the correct error message: `new row for relation "products" violates check constraint "products_chk_al8efblh"`
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
