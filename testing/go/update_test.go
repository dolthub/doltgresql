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

func TestUpdate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple update",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT PRIMARY KEY, b INT)",
				"INSERT INTO t1 VALUES (1, 2), (2, 3), (3, 4)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "UPDATE t1 SET b = 5 WHERE a = 2",
				},
				{
					Query: "SELECT * FROM t1 where a =  2",
					Expected: []sql.Row{
						{2, 5},
					},
				},
			},
		},
		{
			Name: "update to default",
			SetUpScript: []string{
				"create table t (i int default 10, j varchar(128) default (concat('abc', 'def')));",
				"insert into t values (100, 'a'), (200, 'b');",
				"create table t2 (i int);",
				"insert into t2 values (1), (2), (3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "update t set i = default where i = 100;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "a"},
						{200, "b"},
					},
				},
				{
					Query: "update t set j = default where i = 200;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "a"},
						{200, "abcdef"},
					},
				},
				{
					Query: "update t set i = default, j = default;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "abcdef"},
						{10, "abcdef"},
					},
				},
				{
					Query: "update t2 set i = default",
					Skip:  true, // UPDATE: non-Doltgres type found in source
				},
				{
					Query: "select * from t2",
					Skip:  true, // skipped because of above
					Expected: []sql.Row{
						{nil},
						{nil},
						{nil},
					},
				},
			},
		},
		{
			Name: "UPDATE ... RETURNING",
			SetUpScript: []string{
				"CREATE TABLE t (pk INT PRIMARY KEY, c1 TEXT);",
				"INSERT INTO t VALUES (1, 'one');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE t SET c1 = '42' RETURNING c1;",
					Expected: []sql.Row{{"42"}},
				},
				{
					// TODO: * requires extra analysis to expand columns
					Skip:     true,
					Query:    "UPDATE t SET c1 = '43' RETURNING *;",
					Expected: []sql.Row{{1, "43"}},
				},
			},
		},
		{
			// TODO: Update joins are not supported yet
			Skip: true,
			Name: "UPDATE ... RETURNING with join",
			SetUpScript: []string{
				"CREATE TABLE employees (id SERIAL PRIMARY KEY, name TEXT, department_id INT, salary NUMERIC);",
				"CREATE TABLE departments (id SERIAL PRIMARY KEY, name TEXT, bonus NUMERIC);",
				"INSERT INTO employees (name, department_id, salary) VALUES ('Alice', 1, 50000), ('Bob', 2, 60000);",
				"INSERT INTO departments (name, bonus) VALUES ('Engineering', 5000), ('Marketing', 3000);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "UPDATE employees e SET salary = salary + d.bonus FROM departments d WHERE e.department_id = d.id RETURNING e.id, e.name, e.salary;",
					Expected: []sql.Row{
						{1, "Alice", 55000},
						{2, "Bob", 63000},
					},
				},
			},
		},
	})
}
