// Copyright 2023 Dolthub, Inc.
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

func TestSmokeTests(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Simple statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test VALUES (1, 1), (2, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test2 VALUES (3, 3), (4, 4);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{3, 3},
						{4, 4},
					},
				},
				{
					Query: "SELECT * FROM test ORDER BY 1 LIMIT 1 OFFSET 1;",
					Expected: []sql.Row{
						{2, 2},
					},
				},
			},
		},
		{
			Name: "Dolt Getting Started example", /* https://docs.dolthub.com/introduction/getting-started/database */
			SetUpScript: []string{
				"create table employees (id int, last_name varchar(255), first_name varchar(255), primary key(id));",
				"create table teams (id int, team_name varchar(255), primary key(id));",
				"create table employees_teams(team_id int, employee_id int, primary key(team_id, employee_id), foreign key (team_id) references teams(id), foreign key (employee_id) references employees(id));",
				"call dolt_add('teams', 'employees', 'employees_teams');",
				"call dolt_commit('-m', 'Created initial schema');",
				"insert into employees values (0, 'Sehn', 'Tim'), (1, 'Hendriks', 'Brian'), (2, 'Son','Aaron'), (3, 'Fitzgerald', 'Brian');",
				"insert into teams values (0, 'Engineering'), (1, 'Sales');",
				"insert into employees_teams(employee_id, team_id) values (0,0), (1,0), (2,0), (0,1), (3,1);",
				"call dolt_commit('-am', 'Populated tables with data');",
				"call dolt_checkout('-b','modifications');",
				"update employees SET first_name='Timothy' where first_name='Tim';",
				"insert INTO employees (id, first_name, last_name) values (4,'Daylon', 'Wilkins');",
				"insert into employees_teams(team_id, employee_id) values (0,4);",
				"delete from employees_teams where employee_id=0 and team_id=1;",
				"call dolt_commit('-am', 'Modifications on a branch')",
				"call dolt_checkout('modifications');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select to_last_name, to_first_name, to_id, to_commit, from_last_name, from_first_name," +
						"from_id, from_commit, diff_type from dolt_diff('main', 'modifications', 'employees');",
					Expected: []sql.Row{
						{"Sehn", "Timothy", 0, "modifications", "Sehn", "Tim", 0, "main", "modified"},
						{"Wilkins", "Daylon", 4, "modifications", nil, nil, nil, "main", "added"},
					},
				},
			},
		},
		{
			Name: "Boolean results",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT 1 IN (2);",
					Expected: []sql.Row{
						{0},
					},
				},
				{
					Query: "SELECT 2 IN (2);",
					Expected: []sql.Row{
						{1},
					},
				},
			},
		},
		{
			Name: "Commit and diff across branches",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO test VALUES (1, 1), (2, 2);",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'initial commit');",
				"CALL DOLT_BRANCH('other');",
				"UPDATE test SET v1 = 3;",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'commit main');",
				"CALL DOLT_CHECKOUT('other');",
				"UPDATE test SET v1 = 4 WHERE pk = 2;",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'commit other');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CALL DOLT_CHECKOUT('main');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 3},
						{2, 3},
					},
				},
				{
					Query:            "CALL DOLT_CHECKOUT('other');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 4},
					},
				},
				{
					Query: "SELECT from_pk, to_pk, from_v1, to_v1 FROM dolt_diff_test;",
					Expected: []sql.Row{
						{2, 2, 2, 4},
						{nil, 1, nil, 1},
						{nil, 2, nil, 2},
					},
				},
			},
		},
		{
			Name: "ARRAY expression",
			SetUpScript: []string{
				"CREATE TABLE test1 (id INTEGER primary key, v1 BOOLEAN);",
				"INSERT INTO test1 VALUES (1, 'true'), (2, 'false');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT ARRAY[v1]::boolean[] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t}"},
						{"{f}"},
					},
				},
				{
					Query: "SELECT ARRAY[v1] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t}"},
						{"{f}"},
					},
				},
				{
					Query: "SELECT ARRAY[v1, true, v1] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t,t,t}"},
						{"{f,t,f}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::float8, 2::numeric];",
					Expected: []sql.Row{
						{"{1,2}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::float8, NULL];",
					Expected: []sql.Row{
						{"{1,NULL}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::int2, 2::int4, 3::int8]::varchar[];",
					Expected: []sql.Row{
						{"{1,2,3}"},
					},
				},
				{
					Query:       "SELECT ARRAY[1::int8]::int;",
					ExpectedErr: "abc",
				},
				{
					Query:       "SELECT ARRAY[1::int8, 2::varchar];",
					ExpectedErr: "abc",
				},
			},
		},
		{
			Name: "Empty statement",
			Assertions: []ScriptTestAssertion{
				{
					Query:    ";",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Unsupported MySQL statements",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SHOW CREATE TABLE;",
					ExpectedErr: "abc",
				},
			},
		},
	})
}
