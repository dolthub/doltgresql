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

func TestDoltFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "smoke test select dolt_add and dolt_commit",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select dolt_add('.')",
					Expected: []sql.Row{
						{"{0}"},
					},
				},
				{
					Query:            "select dolt_commit('-am', 'new table')",
					SkipResultsCheck: true,
				},
				{
					Query: "select count(*) from dolt_log",
					Expected: []sql.Row{
						{3}, // initial commit, CREATE DATABASE commit, CREATE TABLE commit
					},
				},
				{
					Query: "select message from dolt_log order by date desc limit 1",
					Expected: []sql.Row{
						{"new table"},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_merge",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new table');",
				"SELECT DOLT_CHECKOUT('-b', 'new-branch');",
				"CREATE TABLE t2 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new table on new branch');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT DOLT_MERGE_BASE('main', 'new-branch');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT DOLT_CHECKOUT('main');",
					Expected: []sql.Row{
						{"{0,\"Switched to branch 'main'\"}"},
					},
				},
				{
					Query: "select count(*) from dolt_log",
					Expected: []sql.Row{
						{3}, // initial commit, CREATE DATABASE commit, CREATE TABLE commit
					},
				},
				{
					Query:            "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					SkipResultsCheck: true,
				},
				{
					Query: "select count(*) from dolt_log",
					Expected: []sql.Row{
						{5}, // initial commit, CREATE DATABASE commit, CREATE TABLE t1 commit, new CREATE TABLE t2 commit, merge commit
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_reset",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt_status;",
					Expected: []sql.Row{
						{"t1", 0, "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_ADD('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt_status;",
					Expected: []sql.Row{
						{"t1", 1, "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_RESET('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt_status;",
					Expected: []sql.Row{
						{"t1", 0, "new table"},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_clean",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt_status;",
					Expected: []sql.Row{
						{"t1", 0, "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_CLEAN('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT * FROM dolt_status;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
