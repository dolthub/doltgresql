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
						{"public.t1", "f", "modified"},
					},
				},
				{
					Query:       "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					ExpectedErr: "error: local changes would be stomped by merge",
				},
				{
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", "f", "modified"},
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
						{"public.t1", "f", "modified"},
					},
				},
				{
					Query:            "SELECT DOLT_MERGE('new-branch', '--no-ff', '-m', 'merge new-branch into main');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM dolt.status",
					Expected: []sql.Row{
						{"public.t1", "f", "modified"},
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
						{"public.t1", "f", "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_ADD('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "t", "new table"},
					},
				},
				{
					Query:    "SELECT DOLT_RESET('t1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.t1", "f", "new table"},
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
						{"public.t1", "f", "new table"},
					},
				},
				{
					Skip:     true, // TODO: function dolt_clean() does not exist
					Query:    "SELECT DOLT_CLEAN();",
					Expected: []sql.Row{{"{0}"}},
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
						{"public.t1", "f", "new table"},
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
						{"public.t1", "f", "new table"},
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
			Name: "smoke test select dolt diff functions and tables",
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
					Query: "SELECT * FROM dolt.diff",
					Expected: []sql.Row{
						{"WORKING", "public.t1", nil, nil, nil, nil, "t", "t"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO \"t1\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO \"t1\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
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
			Name: "smoke test select dolt diff functions and tables for multiple schemas",
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
						{"public.t1", "f", "new table"},
						{"testschema.t2", "f", "new table"},
						{"testschema", "f", "new schema"},
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
					Query: "SELECT * FROM dolt.diff",
					Expected: []sql.Row{
						{"WORKING", "public.t1", nil, nil, nil, nil, "t", "t"},
						{"WORKING", "testschema.t2", nil, nil, nil, nil, "t", "t"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO \"t1\" (\"pk\") VALUES (1);"},
						{Numeric("3"), "testschema.t2", "schema", "CREATE TABLE \"t2\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("4"), "testschema.t2", "data", "INSERT INTO \"t2\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{Numeric("1"), "public.t1", "schema", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "public.t1", "data", "INSERT INTO \"t1\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT statement_order, table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{Numeric("1"), "testschema.t2", "schema", "CREATE TABLE \"t2\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{Numeric("2"), "testschema.t2", "data", "INSERT INTO \"t2\" (\"pk\") VALUES (1);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
						{"", "testschema.t2", "", "CREATE TABLE \"t2\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't1')",
					Expected: []sql.Row{
						{"", "public.t1", "", "CREATE TABLE \"t1\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
					},
				},
				{
					Query: "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 't2')",
					Expected: []sql.Row{
						{"", "testschema.t2", "", "CREATE TABLE \"t2\" (\n  \"pk\" integer NOT NULL,\n  PRIMARY KEY (\"pk\")\n);"},
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
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS basic functionality",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"INSERT INTO t1 VALUES (1, 10), (2, 20);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"UPDATE t1 SET c1 = 100 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'update on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"UPDATE t1 SET c1 = 200 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'update on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
					Expected: []sql.Row{
						{"public.t1", Numeric("1"), Numeric("0")},
					},
				},
				{
					Query: "SELECT base_pk, base_c1, our_pk, our_c1, our_diff_type, their_pk, their_c1, their_diff_type FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{1, 10, 1, 200, "modified", 1, 100, "modified"},
					},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with no conflicts",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"INSERT INTO t1 VALUES (1, 10), (2, 20);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"INSERT INTO t1 VALUES (3, 30);",
				"SELECT DOLT_COMMIT('-am', 'insert on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"INSERT INTO t1 VALUES (4, 40);",
				"SELECT DOLT_COMMIT('-am', 'insert on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT base_pk, base_c1, our_pk, our_c1, our_diff_type, their_pk, their_c1, their_diff_type FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple tables",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"CREATE TABLE t2 (pk int primary key, c1 varchar(20));",
				"INSERT INTO t1 VALUES (1, 10);",
				"INSERT INTO t2 VALUES (1, 'initial');",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"UPDATE t1 SET c1 = 100 WHERE pk = 1;",
				"UPDATE t2 SET c1 = 'branch1' WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"UPDATE t1 SET c1 = 200 WHERE pk = 1;",
				"UPDATE t2 SET c1 = 'main' WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
					Expected: []sql.Row{
						{"public.t1", Numeric("1"), Numeric("0")},
						{"public.t2", Numeric("1"), Numeric("0")},
					},
				},
				{
					Query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't2')",
					Expected: []sql.Row{
						{1},
					},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with schema conflicts",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"INSERT INTO t1 VALUES (1, 10);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"ALTER TABLE t1 ADD COLUMN c2 varchar(50);",
				"SELECT DOLT_COMMIT('-am', 'add column on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"ALTER TABLE t1 ADD COLUMN c2 int;",
				"SELECT DOLT_COMMIT('-am', 'add same column different type on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1')",
					Expected: []sql.Row{
						{"public.t1", nil, Numeric("1")},
					},
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					ExpectedErr: "schema conflicts found: 1",
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple schemas",
			SetUpScript: []string{
				"CREATE SCHEMA test_schema;",
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"CREATE TABLE test_schema.t2 (pk int primary key, c1 int);",
				"INSERT INTO t1 VALUES (1, 10);",
				"INSERT INTO test_schema.t2 VALUES (1, 20);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"UPDATE t1 SET c1 = 100 WHERE pk = 1;",
				"UPDATE test_schema.t2 SET c1 = 200 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"UPDATE t1 SET c1 = 300 WHERE pk = 1;",
				"UPDATE test_schema.t2 SET c1 = 400 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
					Expected: []sql.Row{
						{"public.t1", Numeric("1"), Numeric("0")},
						{"test_schema.t2", Numeric("1"), Numeric("0")},
					},
				},
				{
					Query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query: "SELECT COUNT(*) FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't2')",
					Expected: []sql.Row{
						{1},
					},
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS with multiple schemas, same name",
			SetUpScript: []string{
				"CREATE SCHEMA test_schema;",
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"CREATE TABLE test_schema.t1 (pk int primary key, c2 int);",
				"INSERT INTO t1 VALUES (1, 10);",
				"INSERT INTO test_schema.t1 VALUES (1, 20);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
				"UPDATE t1 SET c1 = 100 WHERE pk = 1;",
				"UPDATE test_schema.t1 SET c2 = 200 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on branch1');",
				"SELECT DOLT_CHECKOUT('main');",
				"UPDATE t1 SET c1 = 300 WHERE pk = 1;",
				"UPDATE test_schema.t1 SET c2 = 400 WHERE pk = 1;",
				"SELECT DOLT_COMMIT('-am', 'updates on main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1') ORDER BY 'table'",
					Expected: []sql.Row{
						{"public.t1", Numeric("1"), Numeric("0")},
						{"test_schema.t1", Numeric("1"), Numeric("0")},
					},
				},
				{
					Query: "SELECT base_c1 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{10},
					},
				},
				{
					Query:       "SELECT base_c2 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					ExpectedErr: "column \"base_c2\" could not be found in any table in scope",
				},
				{
					Query:    "SET search_path TO test_schema;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT base_c2 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					Expected: []sql.Row{
						{20},
					},
				},
				{
					Query:       "SELECT base_c1 FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1')",
					ExpectedErr: "column \"base_c1\" could not be found in any table in scope",
				},
			},
		},
		{
			Name: "DOLT_PREVIEW_MERGE_CONFLICTS error cases",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key, c1 int);",
				"SELECT DOLT_COMMIT('-Am', 'initial commit');",
				"SELECT DOLT_CHECKOUT('-b', 'branch1');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('nonexistent-branch', 'main')",
					ExpectedErr: "branch not found: nonexistent-branch",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'branch1', 'table')",
					ExpectedErr: "function 'dolt_preview_merge_conflicts_summary' expected 2 arguments, 3 received",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', 'nonexistent-branch')",
					ExpectedErr: "branch not found: nonexistent-branch",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('', 'main')",
					ExpectedErr: "branch name cannot be empty",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', '')",
					ExpectedErr: "branch name cannot be empty",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY(NULL, 'main')",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts_summary: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS_SUMMARY('main', NULL)",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts_summary: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('nonexistent-branch', 'main', 't1')",
					ExpectedErr: "branch not found: nonexistent-branch",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'nonexistent-branch', 't1')",
					ExpectedErr: "branch not found: nonexistent-branch",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1')",
					ExpectedErr: "function 'dolt_preview_merge_conflicts' expected 3 arguments, 2 received",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 't1', 'extra')",
					ExpectedErr: "function 'dolt_preview_merge_conflicts' expected 3 arguments, 4 received",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('', 'main', 't1')",
					ExpectedErr: "string is not a valid branch or hash",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', '', 't1')",
					ExpectedErr: "string is not a valid branch or hash",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS(NULL, 'main', 't1')",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', NULL, 't1')",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', NULL)",
					ExpectedErr: "Invalid argument to dolt_preview_merge_conflicts: NULL",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', 'nonexistent_table')",
					ExpectedErr: "table not found: public.nonexistent_table",
				},
				{
					Query:       "SELECT * FROM DOLT_PREVIEW_MERGE_CONFLICTS('main', 'branch1', '')",
					ExpectedErr: "table name cannot be empty",
				},
			},
		},
	})
}
