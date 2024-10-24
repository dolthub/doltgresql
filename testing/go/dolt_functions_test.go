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
					Query: "select count(*) from dolt.log",
					Expected: []sql.Row{
						{3}, // initial commit, CREATE DATABASE commit, CREATE TABLE commit
					},
				},
				{
					Query: "select message from dolt.log order by date desc limit 1",
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
					Query: "select count(*) from dolt.log",
					Expected: []sql.Row{
						{3}, // initial commit, CREATE DATABASE commit, CREATE TABLE commit
					},
				},
				{
					Query:            "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					SkipResultsCheck: true,
				},
				{
					Query: "select count(*) from dolt.log",
					Expected: []sql.Row{
						{5}, // initial commit, CREATE DATABASE commit, CREATE TABLE t1 commit, new CREATE TABLE t2 commit, merge commit
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_merge dirty working set, same table",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new table');",
				"INSERT INTO t1 VALUES (1);",
				"SELECT DOLT_CHECKOUT('-b', 'new-branch');",
				"INSERT INTO t1 VALUES (2);",
				"SELECT DOLT_COMMIT('-Am', 'new row on new branch');",
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
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", 0, "modified"},
					},
				},
				{
					Query:       "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					ExpectedErr: "error: local changes would be stomped by merge",
				},
				{
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", 0, "modified"},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_merge dirty working set, different tables",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new table');",
				"INSERT INTO t1 VALUES (1);",
				"SELECT DOLT_CHECKOUT('-b', 'new-branch');",
				"CREATE TABLE t2 (pk int primary key);",
				"SELECT DOLT_COMMIT('-Am', 'new row on new branch');",
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
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", 0, "modified"},
					},
				},
				{
					Query:            "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", 0, "modified"},
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
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", 0, "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_ADD('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", 1, "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_RESET('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", 0, "new table"},
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
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", 0, "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_CLEAN('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT * FROM dolt.status;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE t1 (pk int primary key);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", 0, "new table"},
					},
				},
				{
					Skip:     true, // TODO: function dolt_clean() does not exist
					Query:    "SELECT DOLT_CLEAN();",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Skip:     true,
					Query:    "SELECT * FROM dolt.status;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "smoke test select dolt_checkout(table)",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", 0, "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_CHECKOUT('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT * FROM dolt.status;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "smoke test select dolt_diff functions and tables",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"public.t1", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"public.t1", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
				{
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Skip:  true, // TODO: dolt_commit_diff_* tables must be filtered to a single 'to_commit'
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_commit_diff_t1 WHERE to_commit=HASHOF('main') AND from_commit='WORKING'",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff",
					Expected: []sql.Row{
						{"WORKING", "public.t1", nil, nil, nil, nil, 1, 1},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE `t1` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO `t1` (`pk`) VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE `t1` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO `t1` (`pk`) VALUES (1);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE `t1` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE `t1` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
					},
				},
				{
					Skip:  true, // ERROR: table not found: t1
					Query: "SELECT * FROM dolt_query_diff('select * from t1 as of main', 'select * from t1')",
					Expected: []sql.Row{
						{"", "t1", "added", 1, 1},
					},
				},
			},
		},
		{
			Name: "smoke test select dolt_diff functions and tables for multiple schemas",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
				"INSERT INTO t1 VALUES (1);",
				"CREATE SCHEMA testschema;",
				"CREATE TABLE testschema.t2 (pk int primary key);",
				"INSERT INTO testschema.t2 VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", 0, "new table"},
						{"testschema.t2", 0, "new table"},
						{"testschema", 0, "new schema"},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"public.t1", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
						{"testschema.t2", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"public.t1", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"testschema.t2", 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
						{"", "testschema.t2", "added", 1, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"", "testschema.t2", "added", 1, 1},
					},
				},
				{
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_diff('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Skip:  true, // TODO: dolt_commit_diff_* tables must be filtered to a single 'to_commit'
					Query: "SELECT diff_type, from_pk, to_pk FROM dolt_commit_diff_t1 WHERE to_commit=HASHOF('main') AND from_commit='WORKING'",
					Expected: []sql.Row{
						{"added", nil, 1},
					},
				},
				{
					Query: "SELECT * FROM dolt_diff",
					Expected: []sql.Row{
						{"WORKING", "public.t1", nil, nil, nil, nil, 1, 1},
						{"WORKING", "testschema.t2", nil, nil, nil, nil, 1, 1},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE `t1` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO `t1` (`pk`) VALUES (1);"},
						{Numeric("3"), "testschema.t2", "schema", "CREATE TABLE `t2` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
						{Numeric("4"), "testschema.t2", "data", "INSERT INTO `t2` (`pk`) VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE `t1` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO `t1` (`pk`) VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{Numeric("1"), "testschema.t2", "schema", "CREATE TABLE `t2` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
						{Numeric("2"), "testschema.t2", "data", "INSERT INTO `t2` (`pk`) VALUES (1);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE `t1` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
						{"", "testschema.t2", "", "CREATE TABLE `t2` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE `t1` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"", "testschema.t2", "", "CREATE TABLE `t2` (\n  `pk` integer NOT NULL,\n  PRIMARY KEY (`pk`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;"},
					},
				},
				{
					Skip:  true, // ERROR: table not found: t1
					Query: "SELECT * FROM dolt_query_diff('select * from t1 as of main', 'select * from t1')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
				{
					Skip:  true, // ERROR: table not found: t2
					Query: "SELECT * FROM dolt_query_diff('select * from t2 as of main', 'select * from t2')",
					Expected: []sql.Row{
						{"", "public.t1", "added", 1, 1},
					},
				},
			},
		},
	})
}
