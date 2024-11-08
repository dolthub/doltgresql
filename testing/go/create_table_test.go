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
					ExpectedErr: `Check constraint "products_chk_al8efblh" violated`,
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
					ExpectedErr: `Check constraint "mytbl_chk_vsfp88sb" violated`,
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
					ExpectedErr: `Check constraint "mytbl_chk_vsfp88sb" violated`,
				},
				{
					Query:    "select * from mytbl;",
					Expected: []sql.Row{{1, 20}},
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
					Query:    "insert into t4 values (1, 2, 3, 4);",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from t4;",
					Expected: []sql.Row{
						{1, 2, 3, 4},
					},
				},

				{
					Query:    "create table t111 () inherits (t1, t11);",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into t111 values (1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from t111;",
					Expected: []sql.Row{
						{1},
					},
				},

				{
					Query:    "create table t1t1 (a int) inherits (t1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into t1t1 values (1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from t1t1;",
					Expected: []sql.Row{
						{1},
					},
				},
			},
		},
	})
}
