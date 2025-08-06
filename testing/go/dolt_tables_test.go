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

func TestUserSpaceDoltTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "dolt branches",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT name FROM dolt.branches`,
					ExpectedColNames: []string{"name"},
					Expected:         []sql.Row{{"main"}},
				},
				{
					Query:            `SELECT name FROM dolt_branches`,
					ExpectedColNames: []string{"name"},
					Expected:         []sql.Row{{"main"}},
				},
				{
					Query:            `SELECT name FROM public.dolt_branches`,
					ExpectedColNames: []string{"name"},
					Expected:         []sql.Row{{"main"}},
				},
				{
					Query:            `SELECT branches.name FROM dolt.branches`,
					ExpectedColNames: []string{"name"},
					Expected:         []sql.Row{{"main"}},
				},
				{
					Query:            `SELECT dolt.branches.name FROM dolt.branches`,
					ExpectedColNames: []string{"name"},
					Expected:         []sql.Row{{"main"}},
				},
				{
					Query:    `SELECT dolt_branches.name FROM dolt_branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:       `SELECT * FROM public.branches`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM branches`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE branches (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO branches VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT name FROM dolt.branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:       `CREATE SCHEMA dolt`,
					ExpectedErr: "schema exists",
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:    `SELECT * FROM public.branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM BRANCHES`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT "dolt_branches"."name" FROM "dolt_branches" WHERE "dolt_branches"."name" IN ('main') ORDER BY "dolt_branches"."name" DESC LIMIT 21;`,
					Expected: []sql.Row{{"main"}},
				},
			},
		},
		{
			Skip: true, // TODO: dolt blame will not work until the first query (with clause) works
			Name: "dolt blame with tablename",
			SetUpScript: []string{
				"CREATE TABLE test (id INT PRIMARY KEY)",
				"INSERT INTO test VALUES (1)",
				"SELECT dolt_commit('-Am', 'test commit', '--author', 'John Doe <johndoe@example.com>')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `WITH sorted_diffs_by_pk
									AS (SELECT
													"to_id",
													to_commit,
													to_commit_date,
													diff_type,
													ROW_NUMBER() OVER (
															PARTITION BY coalesce("to_id", "from_id")
															ORDER BY coalesce(to_commit_date, from_commit_date) DESC
													) row_num
											FROM "dolt_diff_test"
										)
									SELECT
											sd."to_id" AS "id",
											dl.committer,
											dl.email,
											dl.message
									FROM
											sorted_diffs_by_pk as sd,
											dolt_log as dl
									WHERE
											dl.commit_hash = sd.to_commit
											and sd.row_num = 1
											and sd.diff_type <> 'removed'
									ORDER BY
													sd."to_id" ASC;`,
					Expected: []sql.Row{{1, "John Doe", "johndoe@example.com", "test commit"}},
				},
				{
					Query:    `SELECT id, committer FROM dolt_blame_test`,
					Expected: []sql.Row{{10, "John Doe"}},
				},
				{
					Query:    `SELECT id, committer FROM public.dolt_blame_test`,
					Expected: []sql.Row{{10, "John Doe"}},
				},
				{
					Query:    `SELECT dolt_blame_test.id FROM public.dolt_blame_test`,
					Expected: []sql.Row{{10}},
				},
				{
					Query:       `SELECT * FROM other.dolt_blame_test`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_sch (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test_sch VALUES (11)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_commit('-Am', 'add test_sch')`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id FROM newschema.dolt_blame_test_sch`,
					Expected: []sql.Row{{11}},
				},
				{
					Query:    `SELECT id, committer FROM public.dolt_blame_test`,
					Expected: []sql.Row{{10, "John Doe"}},
				},
			},
		},
		{
			Name: "dolt column diff",
			SetUpScript: []string{
				"CREATE TABLE test (id INT PRIMARY KEY)",
				"SELECT dolt_commit('-Am', 'test commit')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT table_name, column_name FROM dolt.column_diff`,
					Expected: []sql.Row{{"public.test", "id"}},
				},
				{
					Query:    `SELECT table_name, column_name FROM dolt_column_diff`,
					Expected: []sql.Row{{"public.test", "id"}},
				},
				{
					Query:    `SELECT dolt.column_diff.table_name FROM dolt.column_diff`,
					Expected: []sql.Row{{"public.test"}},
				},
				{
					Query:    `SELECT dolt_column_diff.table_name, dolt_column_diff.column_name FROM dolt_column_diff`,
					Expected: []sql.Row{{"public.test", "id"}},
				},
				{
					Query:       `SELECT * FROM public.column_diff`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM column_diff`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE column_diff (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO column_diff VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM column_diff`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT table_name, column_name FROM dolt.column_diff WHERE table_name = 'public.test'`,
					Expected: []sql.Row{{"public.test", "id"}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT table_name, column_name FROM column_diff WHERE table_name = 'public.test'`,
					Expected: []sql.Row{{"public.test", "id"}},
				},
				{
					Query:    `SELECT * FROM public.column_diff`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM column_diff`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM column_diff`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM COLUMN_DIFF`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt commit ancestors",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT count(*) FROM dolt.commit_ancestors`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT count(*) FROM dolt_commit_ancestors`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT dolt.commit_ancestors.parent_index FROM dolt.commit_ancestors`,
					Expected: []sql.Row{{0}, {0}},
				},
				{
					Query:    `SELECT dolt_commit_ancestors.parent_index FROM dolt_commit_ancestors`,
					Expected: []sql.Row{{0}, {0}},
				},
				{
					Query:       `SELECT * FROM public.commit_ancestors`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM commit_ancestors`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE commit_ancestors (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO commit_ancestors VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM commit_ancestors`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT count(*) FROM dolt.commit_ancestors`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT count(*) FROM commit_ancestors`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT * FROM public.commit_ancestors`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM commit_ancestors`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM commit_ancestors`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM COMMIT_ANCESTORS`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Skip: true, // TODO: dolt_commit_diff_* tables must be filtered to a single 'to_commit'
			Name: "dolt commit diff with tablename",
			SetUpScript: []string{
				"CREATE TABLE test (id INT PRIMARY KEY)",
				"INSERT INTO test VALUES (10)",
				"SELECT dolt_commit('-Am', 'test commit 1')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT from_id, to_id, diff_type FROM dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM public.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
				{
					Query:    `SELECT dolt_commit_diff_test.to_id FROM public.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{10}},
				},
				{
					Query:       `SELECT * FROM other.dolt_commit_diff_test`,
					ExpectedErr: "database schema not found",
				},
				{
					Query:       `SELECT * FROM public.dolt_commit_diff_none`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_sch (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test_sch VALUES (11)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test_sch')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM newschema.dolt_commit_diff_test_sch WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 11, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM dolt_commit_diff_test_sch WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 11, "added"}},
				},
				{
					Query:       `SELECT from_id, to_id, diff_type FROM dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `SELECT to_id, diff_type FROM public.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^2') AND to_commit=HASHOF('HEAD^1')`,
					Expected: []sql.Row{{11, "added"}},
				},
				{
					Query:       `SELECT to_id FROM public.dolt_commit_diff_test_sch WHERE from_commit=HASHOF('HEAD^2') AND to_commit=HASHOF('HEAD^1')`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT to_id, diff_type FROM newschema.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD^1') AND to_commit=HASHOF('HEAD')`,
					ExpectedErr: "table not found",
				},
				{
					// Same name as table in public schema
					Query:    `CREATE TABLE test (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test VALUES (12)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM newschema.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD~1') AND to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 12, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD~1') AND to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 12, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM public.dolt_commit_diff_test WHERE from_commit=HASHOF('HEAD~3') AND to_commit=HASHOF('HEAD~2')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
			},
		},
		{
			Name: "dolt commits",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT count(*) FROM dolt.commits`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT count(*) FROM dolt_commits`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT dolt.commits.message FROM dolt.commits`,
					Expected: []sql.Row{{"CREATE DATABASE"}, {"Initialize data repository"}},
				},
				{
					Query:    `SELECT dolt_commits.message FROM dolt_commits`,
					Expected: []sql.Row{{"CREATE DATABASE"}, {"Initialize data repository"}},
				},
				{
					Query:       `SELECT * FROM public.commits`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM commits`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE commits (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO commits VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM commits`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT count(*) FROM dolt.commits`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT count(*) FROM commits`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT * FROM public.commits`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM commits`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM commits`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM COMMITS`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt conflicts",
			SetUpScript: []string{
				"START TRANSACTION",
				"CREATE TABLE test (id INT PRIMARY KEY, col1 TEXT)",
				"SELECT dolt_commit('-Am', 'first commit')",
				"SELECT dolt_branch('b1')",
				"SELECT dolt_checkout('-b', 'b2')",
				"INSERT INTO test VALUES (1, 'a')",
				"SELECT dolt_commit('-Am', 'commit b2')",
				"SELECT dolt_checkout('b1')",
				"INSERT INTO test VALUES (1, 'b')",
				"SELECT dolt_commit('-Am', 'commit b1')",
				"SELECT dolt_checkout('main')",
				"SELECT dolt_merge('b1')",
				"SELECT dolt_merge('b2')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM dolt.conflicts`,
					Expected: []sql.Row{{"test", Numeric("1")}},
				},
				{
					Query:    `SELECT * FROM dolt_conflicts`,
					Expected: []sql.Row{{"test", 1}},
				},
				{
					Query:    `SELECT dolt.conflicts.table FROM dolt.conflicts`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    `SELECT dolt_conflicts.table FROM dolt_conflicts`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:       `SELECT * FROM public.conflicts`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM conflicts`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE conflicts (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO conflicts VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM conflicts`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM dolt.conflicts`,
					Expected: []sql.Row{{"test", Numeric("1")}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM conflicts`,
					Expected: []sql.Row{{"test", Numeric("1")}},
				},
				{
					Query:    `SELECT * FROM public.conflicts`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM conflicts`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM conflicts`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM CONFLICTS`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt conflicts with tablename",
			SetUpScript: []string{
				"START TRANSACTION",
				"CREATE TABLE test (id INT PRIMARY KEY, col1 TEXT)",
				"SELECT dolt_commit('-Am', 'first commit')",
				"SELECT dolt_branch('b1')",
				"SELECT dolt_checkout('-b', 'b2')",
				"INSERT INTO test VALUES (1, 'a')",
				"SELECT dolt_commit('-Am', 'commit b2')",
				"SELECT dolt_checkout('b1')",
				"INSERT INTO test VALUES (1, 'b')",
				"SELECT dolt_commit('-Am', 'commit b1')",
				"SELECT dolt_checkout('main')",
				"SELECT dolt_merge('b1')",
				"SELECT dolt_merge('b2')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT base_id, base_col1, our_id, our_col1, their_id, their_col1 FROM dolt_conflicts_test`,
					Expected: []sql.Row{{nil, nil, 1, "b", 1, "a"}},
				},
				{
					Query:    `SELECT our_col1, their_col1 FROM public.dolt_conflicts_test`,
					Expected: []sql.Row{{"b", "a"}},
				},
				{
					Query:    `SELECT dolt_conflicts_test.their_col1 FROM public.dolt_conflicts_test`,
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:       `SELECT * FROM other.dolt_conflicts_test`,
					ExpectedErr: "database schema not found",
				},
				{
					Query:       `SELECT * FROM public.dolt_conflicts_none`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `DELETE FROM public.dolt_conflicts_test`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT base_id, base_col1, our_id, our_col1, their_id, their_col1 FROM dolt_conflicts_test`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_sch (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test_sch VALUES (11)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test_sch')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT * FROM newschema.dolt_conflicts_test_sch`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_conflicts_test_sch`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM dolt_conflicts_test`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `SELECT * FROM public.dolt_conflicts_test`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT id FROM public.dolt_conflicts_test_sch`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM newschema.dolt_conflicts_test`,
					ExpectedErr: "table not found",
				},
				{
					// Same name as table in public schema
					Query:    `CREATE TABLE test (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test VALUES (12)`,
					Expected: []sql.Row{},
				},
				// TODO: Create conflict to test correct table
				{
					Query:            `SELECT dolt_commit('-Am', 'add test')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT * FROM newschema.dolt_conflicts_test`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_conflicts_test`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM public.dolt_conflicts_test`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "dolt constraint violations",
			SetUpScript: []string{
				"CREATE TABLE otherTable (pk int primary key);",
				"CREATE TABLE test (pk int primary key, col1 int unique);",
				"SELECT dolt_commit('-Am', 'initial commit');",
				"SELECT dolt_branch('branch1');",
				"INSERT INTO test (pk, col1) VALUES (1, 1);",
				"SELECT dolt_commit('-am', 'insert on main');",
				"SELECT dolt_checkout('branch1');",
				"INSERT INTO test (pk, col1) VALUES (2, 1);",
				"SELECT dolt_commit('-am', 'insert on branch1');",
				"START TRANSACTION",
				"SELECT dolt_merge('main', '--squash')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM dolt.constraint_violations`,
					Expected: []sql.Row{{"test", Numeric("2")}},
				},
				{
					Query:    `SELECT * FROM dolt_constraint_violations`,
					Expected: []sql.Row{{"test", 2}},
				},
				{
					Query:    `SELECT dolt.constraint_violations.table FROM dolt.constraint_violations`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    `SELECT dolt_constraint_violations.table FROM dolt_constraint_violations`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:       `SELECT * FROM public.constraint_violations`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM constraint_violations`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE constraint_violations (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO constraint_violations VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM constraint_violations`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM dolt.constraint_violations`,
					Expected: []sql.Row{{"test", Numeric("2")}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM constraint_violations`,
					Expected: []sql.Row{{"test", Numeric("2")}},
				},
				{
					Query:    `SELECT * FROM public.constraint_violations`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM constraint_violations`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM constraint_violations`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM CONSTRAINT_VIOLATIONS`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt constraint violations with tablename",
			SetUpScript: []string{
				"CREATE TABLE otherTable (pk int primary key);",
				"CREATE TABLE test (pk int primary key, col1 int unique);",
				"SELECT dolt_commit('-Am', 'initial commit');",
				"SELECT dolt_branch('branch1');",
				"INSERT INTO test (pk, col1) VALUES (1, 1);",
				"SELECT dolt_commit('-am', 'insert on main');",
				"SELECT dolt_checkout('branch1');",
				"INSERT INTO test (pk, col1) VALUES (2, 1);",
				"SELECT dolt_commit('-am', 'insert on branch1');",
				"START TRANSACTION",
				"SELECT dolt_merge('main', '--squash')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT violation_type, pk, col1, violation_info FROM dolt_constraint_violations_test`,
					Expected: []sql.Row{
						{"unique index", 1, 1, `{"Columns": ["col1"], "Name": "col1"}`},
						{"unique index", 2, 1, `{"Columns": ["col1"], "Name": "col1"}`},
					},
				},
				{
					Query: `SELECT violation_type, pk, col1, violation_info FROM public.dolt_constraint_violations_test`,
					Expected: []sql.Row{
						{"unique index", 1, 1, `{"Columns": ["col1"], "Name": "col1"}`},
						{"unique index", 2, 1, `{"Columns": ["col1"], "Name": "col1"}`},
					},
				},
				{
					Query:    `SELECT * FROM public.dolt_constraint_violations_test WHERE violation_type = 'foreign key'`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_constraint_violations_test.violation_type FROM public.dolt_constraint_violations_test`,
					Expected: []sql.Row{{"unique index"}, {"unique index"}},
				},
				{
					Query:       `SELECT * FROM other.dolt_constraint_violations_test`,
					ExpectedErr: "database schema not found",
				},
				{
					Query:       `SELECT * FROM public.dolt_constraint_violations_none`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `DELETE FROM public.dolt_constraint_violations_test`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_constraint_violations_test`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_sch (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test_sch VALUES (11)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test_sch')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT * FROM newschema.dolt_constraint_violations_test_sch`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_constraint_violations_test_sch`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM dolt_constraint_violations_test`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `SELECT * FROM public.dolt_constraint_violations_test`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT id FROM public.dolt_constraint_violations_test_sch`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM newschema.dolt_constraint_violations_test`,
					ExpectedErr: "table not found",
				},
				{
					// Same name as table in public schema
					Query:    `CREATE TABLE test (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test VALUES (12)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test')`,
					SkipResultsCheck: true,
				},
				// TODO: Create constraint violation to test correct table
				{
					Query:    `SELECT * FROM newschema.dolt_constraint_violations_test`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_constraint_violations_test`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM public.dolt_constraint_violations_test`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "dolt docs",
			SetUpScript: []string{
				"INSERT INTO dolt.docs values ('README.md', 'testing')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM dolt.docs`,
					Expected: []sql.Row{
						{"README.md", "testing"},
						{"AGENT.md", "# AGENT.md - Dolt Database Operations Guide\n\nThis file provides guidance for AI agents working with Dolt databases to maximize productivity and follow best practices.\n\n## Quick Start\n\nDolt is \"Git for Data\" - a SQL database with version control capabilities. All Git commands have Dolt equivalents:\n- `git add` → `dolt add`  \n- `git commit` → `dolt commit`\n- `git branch` → `dolt branch`\n- `git merge` → `dolt merge`\n- `git diff` → `dolt diff`\n\nFor help and documentation on commands, you can run `dolt --help` and `dolt <command> --help`.\n\n## Essential Dolt CLI Commands\n\n### Repository Operations\n```bash\n# Initialize new database\ndolt init\n\n# Clone existing database\ndolt clone <remote-url>\n\n# Show current status\ndolt status\n\n# View commit history\ndolt log\n```\n\n### Branch Management\n```bash\n# List branches\ndolt branch\n\n# Create new branch\ndolt branch <branch-name>\n\n# Switch branches\ndolt checkout <branch-name>\n\n# Create and switch to new branch\ndolt checkout -b <branch-name>\n```\n\n### Data Operations\n```bash\n# Stage changes\ndolt add <table-name>\ndolt add .  # stage all changes\n\n# Commit changes\ndolt commit -m \"commit message\"\n\n# View differences\ndolt diff\ndolt diff <table-name>\ndolt diff <branch1> <branch2>\n\n# Merge branches\ndolt merge <branch-name>\n```\n\n## Starting and Connecting to Dolt SQL Server\n\n### Start SQL Server\n```bash\n# Start server on default port (3306)\ndolt sql-server\n\n# Start on specific port\ndolt sql-server --port=3307\n\n# Start with specific host\ndolt sql-server --host=0.0.0.0 --port=3307\n\n# Start in background\ndolt sql-server --port=3307 &\n```\n\n### Connecting to SQL Server\n```bash\n# Connect with dolt sql command\ndolt sql\n\n# Connect with mysql client\nmysql -h 127.0.0.1 -P 3306 -u root\n\n# Connect with specific database\nmysql -h 127.0.0.1 -P 3306 -u root -D <database-name>\n```\n\n## Dolt CI Testing\n\n### Prerequisites\n- Requires Dolt v1.43.14 or later\n- Must initialize CI capabilities: `dolt ci init`\n- Workflows defined in YAML files\n\n### Available CI Commands\n```bash\n# Initialize CI capabilities\ndolt ci init\n\n# List available workflows\ndolt ci ls\n\n# View workflow details\ndolt ci view <workflow-name>\n\n# View specific job in workflow\ndolt ci view <workflow-name> <job-name>\n\n# Run workflow locally\ndolt ci run <workflow-name>\n```\n\n### Creating CI Workflows\n\n#### 1. Create Saved Queries First\nBefore creating workflows, save your validation queries:\n\n```bash\n# Save queries using CLI\ndolt sql --save \"show_tables\" -q \"SHOW TABLES;\"\ndolt sql --save \"user_count_check\" -q \"SELECT COUNT(*) as user_count FROM users;\"\ndolt sql --save \"valid_emails\" -q \"SELECT COUNT(*) FROM users WHERE email NOT LIKE '%@%';\"\n```\n\nOr insert directly into the query catalog:\n```sql\nINSERT INTO dolt_query_catalog VALUES \n('show_tables', 1, 'show_tables', 'SHOW TABLES;', 'Table existence check'),\n('user_count_check', 2, 'user_count_check', 'SELECT COUNT(*) as user_count FROM users;', 'User count validation'),\n('valid_emails', 3, 'valid_emails', 'SELECT COUNT(*) FROM users WHERE email NOT LIKE \"%@%\";', 'Email format check');\n```\n\n#### 2. Create Workflow YAML File\nCreate a workflow file (e.g., `data-validation.yaml`) in your current directory:\n\n```yaml\nname: data validation workflow\non:\n  push:\n    branches:\n      - master\n      - main\njobs:\n  - name: validate schema\n    steps:\n      - name: check required tables exist\n        saved_query_name: show_tables\n        expected_rows: \">= 3\"\n      \n      - name: validate user data\n        saved_query_name: user_count_check\n        expected_columns: \"== 1\"\n        expected_rows: \"> 0\"\n  \n  - name: data integrity checks\n    steps:\n      - name: check email format\n        saved_query_name: valid_emails\n        expected_rows: \"== 0\"  # No invalid emails\n```\n\n#### 3. Workflow Structure Reference\n\n**Required Fields:**\n- `name`: Unique workflow identifier\n- `on`: Trigger configuration (currently only `push` supported)\n- `jobs`: Array of job definitions\n\n**Job Structure:**\n- `name`: Job identifier\n- `steps`: Array of step definitions\n\n**Step Structure:**\n- `name`: Step description\n- `saved_query_name`: Reference to saved query\n- `expected_rows`: Optional row count validation (operators: `==`, `>`, `<`, `>=`, `<=`)\n- `expected_columns`: Optional column count validation\n\n**Trigger Options:**\n```yaml\non:\n  push:\n    branches:\n      - master\n      - main\n      - feature/*\n```\n\n### Advanced CI Examples\n\n#### Schema Validation Workflow\n```yaml\nname: schema validation\non:\n  push:\n    branches: [\"*\"]\njobs:\n  - name: table structure\n    steps:\n      - name: users table has required columns\n        saved_query_name: describe_users\n        expected_rows: \"== 5\"\n      \n      - name: products table exists\n        saved_query_name: check_products_table\n        expected_rows: \"> 0\"\n```\n\n#### Data Quality Workflow\n```yaml\nname: data quality checks\non:\n  push:\n    branches:\n      - production\njobs:\n  - name: referential integrity\n    steps:\n      - name: no orphaned orders\n        saved_query_name: orphaned_orders_check\n        expected_rows: \"== 0\"\n      \n      - name: valid price ranges\n        saved_query_name: price_validation\n        expected_rows: \"== 0\"\n  \n  - name: business rules\n    steps:\n      - name: active users have orders\n        saved_query_name: active_users_orders\n        expected_rows: \"> 0\"\n```\n\n### Managing Saved Queries for CI\n\n```bash\n# List all saved queries\ndolt sql --list-saved\n# or\ndolt sql -l\n```\n\n```sql\n-- View saved queries via SQL\nSELECT * FROM dolt_query_catalog;\n\n-- Create queries by inserting into catalog\nINSERT INTO dolt_query_catalog VALUES \n('table_row_counts', 4, 'table_row_counts', \n 'SELECT table_name, table_rows FROM information_schema.tables WHERE table_schema = database();', \n 'Count rows in all tables');\n\n-- Delete saved query\nDELETE FROM dolt_query_catalog WHERE id = 'old_query_name';\n```\n\n### Best Practices for CI\n\n1. **Create Comprehensive Validation Queries**\n   - Test data integrity constraints\n   - Validate business rules\n   - Check schema requirements\n   - Verify data relationships\n\n2. **Use Descriptive Names**\n   - Clear workflow names\n   - Meaningful job descriptions\n   - Descriptive step names\n\n3. **Test Locally First**\n   ```bash\n   dolt ci run <workflow-name>\n   ```\n\n4. **Version Control Your Workflows**\n   - Commit workflow files to repository\n   - Track changes to CI configuration\n   - Use branches for CI development\n\n## System Tables for Version Control\n\nDolt exposes version control operations through system tables accessible via SQL:\n\n### Core System Tables\n```sql\n-- View commit history\nSELECT * FROM dolt_log;\n\n-- Check current status\nSELECT * FROM dolt_status;\n\n-- View branch information\nSELECT * FROM dolt_branches;\n\n-- See table diffs\nSELECT * FROM dolt_diff_<table_name>;\n\n-- View schema changes\nSELECT * FROM dolt_schema_diff;\n\n-- Check conflicts during merge\nSELECT * FROM dolt_conflicts_<table_name>;\n\n-- View commit metadata\nSELECT * FROM dolt_commits;\n```\n\n### Version Control Operations via SQL\n\nWhen working in SQL sessions, you can execute version control operations using stored procedures:\n\n```sql\n-- Stage and commit changes\nCALL dolt_add('.');\nCALL dolt_commit('-m', 'commit message');\n\n-- Branch operations\nCALL dolt_branch('<branch_name>');\nCALL dolt_checkout('<branch_name>');\nCALL dolt_merge('<branch_name>');\n```\n\n**Note:** Use CLI commands (`dolt add`, `dolt commit`, etc.) for most operations. SQL procedures are useful when already in a SQL session.\n\n### Advanced System Tables\n```sql\n-- View remotes\nSELECT * FROM dolt_remotes;\n\n-- Check merge conflicts\nSELECT * FROM dolt_conflicts;\n\n-- View statistics\nSELECT * FROM dolt_statistics;\n\n-- See ignored tables\nSELECT * FROM dolt_ignore;\n```\n\n## CLI vs SQL Approach\n\n**Prefer CLI commands for:**\n- Version control operations (add, commit, branch, merge)\n- Repository management (init, clone, push, pull)\n- Conflict resolution\n- Status checking and history viewing\n\n**Use SQL for:**\n- Data queries and analysis\n- Complex data transformations\n- Examining system tables (dolt_log, dolt_status, etc.)\n- When already in an active SQL session\n\n## Best Practices for Agents\n\n### 1. Always Work on Feature Branches\n```bash\n# Create feature branch before making changes\ndolt checkout -b feature/agent-changes\n\n# Make changes on feature branch\ndolt sql -q \"INSERT INTO users VALUES (1, 'Alice');\"\n\n# Stage and commit\ndolt add .\ndolt commit -m \"Add new user Alice\"\n\n# Switch back to main to merge\ndolt checkout main\ndolt merge feature/agent-changes\n```\n\n### 2. Use SQL for Data Operations, CLI for Version Control\n```bash\n# Use dolt sql for data changes\ndolt sql -q \"INSERT INTO users VALUES (1, 'Alice');\"\ndolt sql -q \"UPDATE products SET price = price * 1.1 WHERE category = 'electronics';\"\n\n# Check status and commit using CLI\ndolt status\ndolt add .\ndolt commit -m \"Update user and product data\"\n```\n\n### 3. Validate Changes with System Tables\n```sql\n-- Before major operations, check current state\nSELECT * FROM dolt_status;\nSELECT * FROM dolt_branches;\n\n-- After changes, verify with diffs\nSELECT * FROM dolt_diff_users;\nSELECT * FROM dolt_schema_diff;\n```\n\n### 4. Use CI for Data Validation\nCreate workflows to validate:\n- Data integrity after changes\n- Schema compatibility\n- Business rule compliance\n- Cross-table relationships\n\n### 5. Handle Conflicts Gracefully\n```bash\n# Check for conflicts using CLI\ndolt conflicts cat <table_name>\ndolt conflicts resolve --ours <table_name>\ndolt conflicts resolve --theirs <table_name>\n\n# Or use SQL to examine conflicts\ndolt sql -q \"SELECT * FROM dolt_conflicts_<table_name>;\"\n```\n\n## Common Workflow Examples\n\n### Data Migration Workflow\n```bash\n# Create migration branch\ndolt checkout -b migration/update-schema\n\n# Apply schema changes via SQL\ndolt sql -q \"ALTER TABLE users ADD COLUMN email VARCHAR(255);\"\n\n# Create CI validation query\ndolt sql --save \"schema_check\" -q \"DESCRIBE users;\"\n\n# Define a CI workflow\ndolt ci import schema-validation.yaml\n\n# Test with CI\ndolt ci run schema-validation\n\n# Stage and commit\ndolt add .\ndolt commit -m \"Add email column to users table\"\n\n# Merge back\ndolt checkout main\ndolt merge migration/update-schema\n```\n\n### Data Analysis Workflow\n```bash\n# Create analysis branch\ndolt checkout -b analysis/user-behavior\n\n# Create analysis tables via SQL\ndolt sql -q \"CREATE TABLE user_metrics AS \n            SELECT user_id, COUNT(*) as actions \n            FROM user_actions \n            GROUP BY user_id;\"\n\n# Stage and commit using CLI\ndolt add user_metrics\ndolt commit -m \"Add user behavior analysis\"\n```\n\n## Integration with External Tools\n\n### Database Clients\nMost MySQL clients work with Dolt:\n- MySQL Workbench\n- phpMyAdmin  \n- DataGrip\n- DBeaver\n\n### Backup and Sync\n```bash\n# Push to remote\ndolt push origin main\n\n# Pull changes\ndolt pull origin main\n\n# Clone for backup\ndolt clone <remote-url> backup-location\n```\n\nThis guide enables agents to leverage Dolt's unique version control capabilities while maintaining data integrity and following collaborative development practices."},
					},
				},
				{
					Query: `SELECT dolt.docs.doc_name FROM dolt.docs`,
					Expected: []sql.Row{
						{"README.md"},
						{"AGENT.md"},
					},
				},
				{
					Query:       `SELECT * FROM public.docs`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM docs`,
					ExpectedErr: "table not found",
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING')`,
					Expected: []sql.Row{
						{"", "dolt.docs", "added", 1, 1},
					},
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'docs')`,
					Expected: []sql.Row{
						{"", "dolt.docs", "added", 1, 1},
					},
				},
				{
					Skip:  true, // TODO: we should support this
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_docs')`,
					Expected: []sql.Row{
						{"", "dolt_docs", "added", 1, 1},
					},
				},
				{
					Skip:  true, // TODO: we should support this or a --schema flag
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt.docs')`,
					Expected: []sql.Row{
						{"", "dolt.docs", "added", 1, 1},
					},
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'docs')`,
					Expected: []sql.Row{
						{"", "dolt.docs", "added", 1, 1},
					},
				},
				{
					Query: `SELECT diff_type, from_doc_name, to_doc_name FROM dolt_diff('main', 'WORKING', 'docs')`,
					Expected: []sql.Row{
						{"added", nil, "README.md"},
					},
				},
				{
					Query: `SELECT diff_type, from_doc_name, to_doc_name FROM dolt_diff('main', 'WORKING', 'docs')`,
					Expected: []sql.Row{
						{"added", nil, "README.md"},
					},
				},
				{
					Query:    `CREATE TABLE docs (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO docs VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM docs`,
					Expected: []sql.Row{{1}},
				},
				{
					Query: `SELECT doc_name FROM dolt.docs`,
					Expected: []sql.Row{
						{"README.md"},
						{"AGENT.md"},
					},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT doc_name FROM docs`,
					Expected: []sql.Row{
						{"README.md"},
						{"AGENT.md"},
					},
				},
				{
					Query:    `SELECT * FROM public.docs`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM docs`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM docs`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM DOCS`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `DELETE FROM dolt.docs WHERE doc_name = 'README.md'`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM dolt.docs`,
					Expected: []sql.Row{
						{"AGENT.md", "# AGENT.md - Dolt Database Operations Guide\n\nThis file provides guidance for AI agents working with Dolt databases to maximize productivity and follow best practices.\n\n## Quick Start\n\nDolt is \"Git for Data\" - a SQL database with version control capabilities. All Git commands have Dolt equivalents:\n- `git add` → `dolt add`  \n- `git commit` → `dolt commit`\n- `git branch` → `dolt branch`\n- `git merge` → `dolt merge`\n- `git diff` → `dolt diff`\n\nFor help and documentation on commands, you can run `dolt --help` and `dolt <command> --help`.\n\n## Essential Dolt CLI Commands\n\n### Repository Operations\n```bash\n# Initialize new database\ndolt init\n\n# Clone existing database\ndolt clone <remote-url>\n\n# Show current status\ndolt status\n\n# View commit history\ndolt log\n```\n\n### Branch Management\n```bash\n# List branches\ndolt branch\n\n# Create new branch\ndolt branch <branch-name>\n\n# Switch branches\ndolt checkout <branch-name>\n\n# Create and switch to new branch\ndolt checkout -b <branch-name>\n```\n\n### Data Operations\n```bash\n# Stage changes\ndolt add <table-name>\ndolt add .  # stage all changes\n\n# Commit changes\ndolt commit -m \"commit message\"\n\n# View differences\ndolt diff\ndolt diff <table-name>\ndolt diff <branch1> <branch2>\n\n# Merge branches\ndolt merge <branch-name>\n```\n\n## Starting and Connecting to Dolt SQL Server\n\n### Start SQL Server\n```bash\n# Start server on default port (3306)\ndolt sql-server\n\n# Start on specific port\ndolt sql-server --port=3307\n\n# Start with specific host\ndolt sql-server --host=0.0.0.0 --port=3307\n\n# Start in background\ndolt sql-server --port=3307 &\n```\n\n### Connecting to SQL Server\n```bash\n# Connect with dolt sql command\ndolt sql\n\n# Connect with mysql client\nmysql -h 127.0.0.1 -P 3306 -u root\n\n# Connect with specific database\nmysql -h 127.0.0.1 -P 3306 -u root -D <database-name>\n```\n\n## Dolt CI Testing\n\n### Prerequisites\n- Requires Dolt v1.43.14 or later\n- Must initialize CI capabilities: `dolt ci init`\n- Workflows defined in YAML files\n\n### Available CI Commands\n```bash\n# Initialize CI capabilities\ndolt ci init\n\n# List available workflows\ndolt ci ls\n\n# View workflow details\ndolt ci view <workflow-name>\n\n# View specific job in workflow\ndolt ci view <workflow-name> <job-name>\n\n# Run workflow locally\ndolt ci run <workflow-name>\n```\n\n### Creating CI Workflows\n\n#### 1. Create Saved Queries First\nBefore creating workflows, save your validation queries:\n\n```bash\n# Save queries using CLI\ndolt sql --save \"show_tables\" -q \"SHOW TABLES;\"\ndolt sql --save \"user_count_check\" -q \"SELECT COUNT(*) as user_count FROM users;\"\ndolt sql --save \"valid_emails\" -q \"SELECT COUNT(*) FROM users WHERE email NOT LIKE '%@%';\"\n```\n\nOr insert directly into the query catalog:\n```sql\nINSERT INTO dolt_query_catalog VALUES \n('show_tables', 1, 'show_tables', 'SHOW TABLES;', 'Table existence check'),\n('user_count_check', 2, 'user_count_check', 'SELECT COUNT(*) as user_count FROM users;', 'User count validation'),\n('valid_emails', 3, 'valid_emails', 'SELECT COUNT(*) FROM users WHERE email NOT LIKE \"%@%\";', 'Email format check');\n```\n\n#### 2. Create Workflow YAML File\nCreate a workflow file (e.g., `data-validation.yaml`) in your current directory:\n\n```yaml\nname: data validation workflow\non:\n  push:\n    branches:\n      - master\n      - main\njobs:\n  - name: validate schema\n    steps:\n      - name: check required tables exist\n        saved_query_name: show_tables\n        expected_rows: \">= 3\"\n      \n      - name: validate user data\n        saved_query_name: user_count_check\n        expected_columns: \"== 1\"\n        expected_rows: \"> 0\"\n  \n  - name: data integrity checks\n    steps:\n      - name: check email format\n        saved_query_name: valid_emails\n        expected_rows: \"== 0\"  # No invalid emails\n```\n\n#### 3. Workflow Structure Reference\n\n**Required Fields:**\n- `name`: Unique workflow identifier\n- `on`: Trigger configuration (currently only `push` supported)\n- `jobs`: Array of job definitions\n\n**Job Structure:**\n- `name`: Job identifier\n- `steps`: Array of step definitions\n\n**Step Structure:**\n- `name`: Step description\n- `saved_query_name`: Reference to saved query\n- `expected_rows`: Optional row count validation (operators: `==`, `>`, `<`, `>=`, `<=`)\n- `expected_columns`: Optional column count validation\n\n**Trigger Options:**\n```yaml\non:\n  push:\n    branches:\n      - master\n      - main\n      - feature/*\n```\n\n### Advanced CI Examples\n\n#### Schema Validation Workflow\n```yaml\nname: schema validation\non:\n  push:\n    branches: [\"*\"]\njobs:\n  - name: table structure\n    steps:\n      - name: users table has required columns\n        saved_query_name: describe_users\n        expected_rows: \"== 5\"\n      \n      - name: products table exists\n        saved_query_name: check_products_table\n        expected_rows: \"> 0\"\n```\n\n#### Data Quality Workflow\n```yaml\nname: data quality checks\non:\n  push:\n    branches:\n      - production\njobs:\n  - name: referential integrity\n    steps:\n      - name: no orphaned orders\n        saved_query_name: orphaned_orders_check\n        expected_rows: \"== 0\"\n      \n      - name: valid price ranges\n        saved_query_name: price_validation\n        expected_rows: \"== 0\"\n  \n  - name: business rules\n    steps:\n      - name: active users have orders\n        saved_query_name: active_users_orders\n        expected_rows: \"> 0\"\n```\n\n### Managing Saved Queries for CI\n\n```bash\n# List all saved queries\ndolt sql --list-saved\n# or\ndolt sql -l\n```\n\n```sql\n-- View saved queries via SQL\nSELECT * FROM dolt_query_catalog;\n\n-- Create queries by inserting into catalog\nINSERT INTO dolt_query_catalog VALUES \n('table_row_counts', 4, 'table_row_counts', \n 'SELECT table_name, table_rows FROM information_schema.tables WHERE table_schema = database();', \n 'Count rows in all tables');\n\n-- Delete saved query\nDELETE FROM dolt_query_catalog WHERE id = 'old_query_name';\n```\n\n### Best Practices for CI\n\n1. **Create Comprehensive Validation Queries**\n   - Test data integrity constraints\n   - Validate business rules\n   - Check schema requirements\n   - Verify data relationships\n\n2. **Use Descriptive Names**\n   - Clear workflow names\n   - Meaningful job descriptions\n   - Descriptive step names\n\n3. **Test Locally First**\n   ```bash\n   dolt ci run <workflow-name>\n   ```\n\n4. **Version Control Your Workflows**\n   - Commit workflow files to repository\n   - Track changes to CI configuration\n   - Use branches for CI development\n\n## System Tables for Version Control\n\nDolt exposes version control operations through system tables accessible via SQL:\n\n### Core System Tables\n```sql\n-- View commit history\nSELECT * FROM dolt_log;\n\n-- Check current status\nSELECT * FROM dolt_status;\n\n-- View branch information\nSELECT * FROM dolt_branches;\n\n-- See table diffs\nSELECT * FROM dolt_diff_<table_name>;\n\n-- View schema changes\nSELECT * FROM dolt_schema_diff;\n\n-- Check conflicts during merge\nSELECT * FROM dolt_conflicts_<table_name>;\n\n-- View commit metadata\nSELECT * FROM dolt_commits;\n```\n\n### Version Control Operations via SQL\n\nWhen working in SQL sessions, you can execute version control operations using stored procedures:\n\n```sql\n-- Stage and commit changes\nCALL dolt_add('.');\nCALL dolt_commit('-m', 'commit message');\n\n-- Branch operations\nCALL dolt_branch('<branch_name>');\nCALL dolt_checkout('<branch_name>');\nCALL dolt_merge('<branch_name>');\n```\n\n**Note:** Use CLI commands (`dolt add`, `dolt commit`, etc.) for most operations. SQL procedures are useful when already in a SQL session.\n\n### Advanced System Tables\n```sql\n-- View remotes\nSELECT * FROM dolt_remotes;\n\n-- Check merge conflicts\nSELECT * FROM dolt_conflicts;\n\n-- View statistics\nSELECT * FROM dolt_statistics;\n\n-- See ignored tables\nSELECT * FROM dolt_ignore;\n```\n\n## CLI vs SQL Approach\n\n**Prefer CLI commands for:**\n- Version control operations (add, commit, branch, merge)\n- Repository management (init, clone, push, pull)\n- Conflict resolution\n- Status checking and history viewing\n\n**Use SQL for:**\n- Data queries and analysis\n- Complex data transformations\n- Examining system tables (dolt_log, dolt_status, etc.)\n- When already in an active SQL session\n\n## Best Practices for Agents\n\n### 1. Always Work on Feature Branches\n```bash\n# Create feature branch before making changes\ndolt checkout -b feature/agent-changes\n\n# Make changes on feature branch\ndolt sql -q \"INSERT INTO users VALUES (1, 'Alice');\"\n\n# Stage and commit\ndolt add .\ndolt commit -m \"Add new user Alice\"\n\n# Switch back to main to merge\ndolt checkout main\ndolt merge feature/agent-changes\n```\n\n### 2. Use SQL for Data Operations, CLI for Version Control\n```bash\n# Use dolt sql for data changes\ndolt sql -q \"INSERT INTO users VALUES (1, 'Alice');\"\ndolt sql -q \"UPDATE products SET price = price * 1.1 WHERE category = 'electronics';\"\n\n# Check status and commit using CLI\ndolt status\ndolt add .\ndolt commit -m \"Update user and product data\"\n```\n\n### 3. Validate Changes with System Tables\n```sql\n-- Before major operations, check current state\nSELECT * FROM dolt_status;\nSELECT * FROM dolt_branches;\n\n-- After changes, verify with diffs\nSELECT * FROM dolt_diff_users;\nSELECT * FROM dolt_schema_diff;\n```\n\n### 4. Use CI for Data Validation\nCreate workflows to validate:\n- Data integrity after changes\n- Schema compatibility\n- Business rule compliance\n- Cross-table relationships\n\n### 5. Handle Conflicts Gracefully\n```bash\n# Check for conflicts using CLI\ndolt conflicts cat <table_name>\ndolt conflicts resolve --ours <table_name>\ndolt conflicts resolve --theirs <table_name>\n\n# Or use SQL to examine conflicts\ndolt sql -q \"SELECT * FROM dolt_conflicts_<table_name>;\"\n```\n\n## Common Workflow Examples\n\n### Data Migration Workflow\n```bash\n# Create migration branch\ndolt checkout -b migration/update-schema\n\n# Apply schema changes via SQL\ndolt sql -q \"ALTER TABLE users ADD COLUMN email VARCHAR(255);\"\n\n# Create CI validation query\ndolt sql --save \"schema_check\" -q \"DESCRIBE users;\"\n\n# Define a CI workflow\ndolt ci import schema-validation.yaml\n\n# Test with CI\ndolt ci run schema-validation\n\n# Stage and commit\ndolt add .\ndolt commit -m \"Add email column to users table\"\n\n# Merge back\ndolt checkout main\ndolt merge migration/update-schema\n```\n\n### Data Analysis Workflow\n```bash\n# Create analysis branch\ndolt checkout -b analysis/user-behavior\n\n# Create analysis tables via SQL\ndolt sql -q \"CREATE TABLE user_metrics AS \n            SELECT user_id, COUNT(*) as actions \n            FROM user_actions \n            GROUP BY user_id;\"\n\n# Stage and commit using CLI\ndolt add user_metrics\ndolt commit -m \"Add user behavior analysis\"\n```\n\n## Integration with External Tools\n\n### Database Clients\nMost MySQL clients work with Dolt:\n- MySQL Workbench\n- phpMyAdmin  \n- DataGrip\n- DBeaver\n\n### Backup and Sync\n```bash\n# Push to remote\ndolt push origin main\n\n# Pull changes\ndolt pull origin main\n\n# Clone for backup\ndolt clone <remote-url> backup-location\n```\n\nThis guide enables agents to leverage Dolt's unique version control capabilities while maintaining data integrity and following collaborative development practices."},
					},
				},
				{
					Query:    `DELETE FROM dolt_docs WHERE doc_name = 'README.md'`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "dolt diff",
			SetUpScript: []string{
				"CREATE TABLE test (id INT PRIMARY KEY)",
				"SELECT dolt_commit('-Am', 'test commit')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT table_name FROM dolt.diff`,
					Expected: []sql.Row{{"public.test"}},
				},
				{
					Query:    `SELECT table_name, committer, email, message, data_change, schema_change FROM dolt.diff`,
					Expected: []sql.Row{{"public.test", "postgres", "postgres@127.0.0.1", "test commit", "f", "t"}},
				},
				{
					Query:    `SELECT table_name, data_change, schema_change FROM dolt.diff WHERE data_change=false`,
					Expected: []sql.Row{{"public.test", "f", "t"}},
				},
				{
					Query:    `SELECT table_name, data_change, schema_change FROM dolt.diff WHERE schema_change=false`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT table_name FROM dolt_diff`,
					Expected: []sql.Row{{"public.test"}},
				},
				{
					Query:    `SELECT dolt.diff.table_name FROM dolt.diff`,
					Expected: []sql.Row{{"public.test"}},
				},
				{
					Query:    `SELECT dolt_diff.table_name FROM dolt_diff`,
					Expected: []sql.Row{{"public.test"}},
				},
				{
					Query:       `SELECT * FROM public.diff`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM diff`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE diff (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO diff VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM diff`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT table_name FROM dolt.diff WHERE table_name = 'public.test'`,
					Expected: []sql.Row{{"public.test"}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT table_name FROM diff WHERE table_name = 'public.test'`,
					Expected: []sql.Row{{"public.test"}},
				},
				{
					Query:    `SELECT * FROM public.diff`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM diff`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM diff`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM DIFF`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt diff with tablename",
			SetUpScript: []string{
				"CREATE TABLE test (id INT PRIMARY KEY)",
				"INSERT INTO test VALUES (10)",
				"SELECT dolt_commit('-Am', 'test commit 1')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM doLt_DIff_tEst WHERE to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM public.dolt_diff_test WHERE to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM public.doLt_DIff_tEst WHERE to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
				{
					Query:    `SELECT dolt_diff_test.to_id FROM public.dolt_diff_test WHERE to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{10}},
				},
				{
					Query:       `SELECT * FROM other.dolt_diff_test`,
					ExpectedErr: "database schema not found",
				},
				{
					Query:       `SELECT * FROM public.dolt_diff_none`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM dolt_diff_none`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_sch (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test_sch VALUES (11)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test_sch')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM newschema.dolt_diff_test_sch WHERE  to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 11, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM dolt_diff_test_sch WHERE to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 11, "added"}},
				},
				{
					Query:       `SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD')`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM public.dolt_diff_test WHERE to_commit=HASHOF('HEAD^1')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
				{
					Query:       `SELECT to_id FROM public.dolt_diff_test_sch WHERE to_commit=HASHOF('HEAD^1')`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT to_id FROM newschema.dolt_diff_test WHERE to_commit=HASHOF('HEAD')`,
					ExpectedErr: "table not found",
				},
				{
					// Same name as table in public schema
					Query:    `CREATE TABLE test (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test VALUES (12)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM newschema.dolt_diff_test WHERE  to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 12, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 12, "added"}},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM public.dolt_diff_test WHERE to_commit=HASHOF('HEAD~2')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
				{
					Query:    "SET search_path = 'newschema,public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD')`,
					Expected: []sql.Row{{nil, 12, "added"}},
				},
				{
					Query:    "SET search_path = 'public,newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT from_id, to_id, diff_type FROM dolt_diff_test WHERE to_commit=HASHOF('HEAD~2')`,
					Expected: []sql.Row{{nil, 10, "added"}},
				},
			},
		},
		{
			Name: "dolt history with tablename",
			SetUpScript: []string{
				"CREATE TABLE test (id INT PRIMARY KEY)",
				"INSERT INTO test VALUES (10)",
				"SELECT dolt_commit('-Am', 'test commit', '--author', 'John Doe <johndoe@example.com>')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT id, committer FROM dolt_history_test`,
					Expected: []sql.Row{{10, "John Doe"}},
				},
				{
					Query:    `SELECT id, committer FROM public.dolt_history_test`,
					Expected: []sql.Row{{10, "John Doe"}},
				},
				{
					Query:    `SELECT dolt_history_test.id FROM public.dolt_history_test`,
					Expected: []sql.Row{{10}},
				},
				{
					Query:       `SELECT * FROM other.dolt_history_test`,
					ExpectedErr: "database schema not found",
				},
				{
					Query:       `SELECT * FROM public.dolt_history_none`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_sch (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test_sch VALUES (11)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test_sch', '--author', 'Another Doe <adoe@example.com>')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT id, committer FROM newschema.dolt_history_test_sch`,
					Expected: []sql.Row{{11, "Another Doe"}},
				},
				{
					Query:    `SELECT id, committer FROM dolt_history_test_sch`,
					Expected: []sql.Row{{11, "Another Doe"}},
				},
				{
					Query:       `SELECT id, committer FROM dolt_history_test`,
					ExpectedErr: "table not found",
				},
				{
					Skip:     true, // TODO: Returning rows for both tables
					Query:    `SELECT id, committer FROM public.dolt_history_test`,
					Expected: []sql.Row{{10, "John Doe"}},
				},
				{
					Query:       `SELECT id FROM public.dolt_history_test_sch`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT id, committer FROM newschema.dolt_history_test`,
					ExpectedErr: "table not found",
				},
				{
					// Same name as table in public schema
					Query:    `CREATE TABLE test (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test VALUES (12)`,
					Expected: []sql.Row{},
				},
				{
					Query:            `SELECT dolt_commit('-Am', 'add test')`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT id, committer FROM newschema.dolt_history_test`,
					Expected: []sql.Row{{12, "postgres"}},
				},
				{
					Query:    `SELECT id, committer FROM dolt_history_test`,
					Expected: []sql.Row{{12, "postgres"}},
				},
				{
					Skip:     true, // TODO: Returning rows for all tables
					Query:    `SELECT id, committer FROM public.dolt_history_test`,
					Expected: []sql.Row{{10, "John Doe"}},
				},
				{
					Query:    "SET search_path = 'newschema,public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, committer FROM dolt_history_test`,
					Expected: []sql.Row{{12, "postgres"}},
				},
			},
		},
		{
			Name:        "dolt ignore",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM dolt_ignore`,
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO dolt_ignore VALUES ('generated_*', true), ('generated_exception', false)",
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM dolt_ignore`,
					Expected: []sql.Row{
						{"generated_*", "t"},
						{"generated_exception", "f"},
					},
				},
				{
					Query: `SELECT * FROM dolt_ignore WHERE ignored=false`,
					Expected: []sql.Row{
						{"generated_exception", "f"},
					},
				},
				{
					Query: `SELECT * FROM public.dolt_ignore`,
					Expected: []sql.Row{
						{"generated_*", "t"},
						{"generated_exception", "f"},
					},
				},
				{
					Query: `SELECT dolt_ignore.pattern FROM public.dolt_ignore`,
					Expected: []sql.Row{
						{"generated_*"},
						{"generated_exception"},
					},
				},
				{
					Query:       `SELECT name FROM other.dolt_ignore`,
					ExpectedErr: "database schema not found",
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING')`,
					Expected: []sql.Row{
						{"", "public.dolt_ignore", "added", 1, 1},
					},
				},
				{
					Query: `SELECT diff_type, from_pattern, to_pattern FROM dolt_diff('main', 'WORKING', 'dolt_ignore')`,
					Expected: []sql.Row{
						{"added", nil, "generated_*"},
						{"added", nil, "generated_exception"},
					},
				},
				{
					Query:    "CREATE TABLE foo (pk int);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE generated_foo (pk int);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE generated_exception (pk int);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT dolt_add('-A');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt_status;",
					Expected: []sql.Row{
						{"public.dolt_ignore", "t", "new table"},
						{"public.foo", "t", "new table"},
						{"public.generated_exception", "t", "new table"},
						{"public.generated_foo", "f", "new table"},
					},
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO newschema.dolt_ignore VALUES ('test_*', true)",
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM dolt_ignore`,
					Expected: []sql.Row{
						{"test_*", "t"},
					},
				},
				{
					// Should ignore generated_expected table in newschema but not in public
					Query:    "INSERT INTO dolt_ignore VALUES ('generated_exception', true)",
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM dolt_ignore`,
					Expected: []sql.Row{
						{"generated_exception", "t"},
						{"test_*", "t"},
					},
				},
				{
					Query: `SELECT * FROM newschema.dolt_ignore`,
					Expected: []sql.Row{
						{"generated_exception", "t"},
						{"test_*", "t"},
					},
				},
				{
					Query: `SELECT * FROM public.dolt_ignore`,
					Expected: []sql.Row{
						{"generated_*", "t"},
						{"generated_exception", "f"},
					},
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_ignore')`,
					Expected: []sql.Row{
						{"", "newschema.dolt_ignore", "added", 1, 1},
					},
				},
				{
					Query: `SELECT pattern FROM public.dolt_ignore`,
					Expected: []sql.Row{
						{"generated_*"},
						{"generated_exception"},
					},
				},
				{
					Query:    "CREATE TABLE foo (pk int);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE test_foo (pk int);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE generated_foo (pk int);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE generated_exception (pk int);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT dolt_add('-A');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt_status ORDER BY table_name;",
					Expected: []sql.Row{
						{"newschema", "t", "new schema"},
						{"newschema.dolt_ignore", "t", "new table"},
						{"newschema.foo", "t", "new table"},
						{"newschema.generated_exception", "f", "new table"},
						{"newschema.generated_foo", "t", "new table"},
						{"newschema.test_foo", "f", "new table"},
						{"public.dolt_ignore", "t", "new table"},
						{"public.foo", "t", "new table"},
						{"public.generated_exception", "t", "new table"},
						{"public.generated_foo", "f", "new table"},
					},
				},
			},
		},
		{
			Name: "dolt log",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT count(*) FROM dolt.log`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT count(*) FROM dolt_log`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:       `SELECT * FROM public.log`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM log`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE log (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO log VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM log`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT count(*) FROM dolt.log`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT count(*) FROM log`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT * FROM public.log`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM log`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt merge status",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT is_merging FROM dolt.merge_status`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT is_merging FROM dolt.merge_status WHERE is_merging=true`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT is_merging FROM dolt_merge_status`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT dolt.merge_status.is_merging FROM dolt.merge_status`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT dolt_merge_status.is_merging FROM dolt_merge_status`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:       `SELECT * FROM public.merge_status`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM merge_status`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE merge_status (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO merge_status VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM merge_status`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT is_merging FROM dolt.merge_status`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT is_merging FROM merge_status`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT * FROM public.merge_status`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM merge_status`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM merge_status`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM MERGE_STATUS`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		// TODO: turn on statistics
		// {
		//	Name: "dolt statistics",
		//	SetUpScript: []string{
		//		"CREATE TABLE horses (id int primary key, name varchar(10));",
		//		"CREATE INDEX horses_name_idx ON horses(name);",
		//		"insert into horses select x, 'Steve' from (with recursive inputs(x) as (select 1 union select x+1 from inputs where x < 1000) select * from inputs) dt;",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    `ANALYZE horses;`,
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query: `SELECT database_name, table_name, index_name, row_count, distinct_count, columns, upper_bound, upper_bound_cnt FROM dolt_statistics ORDER BY index_name, row_count`,
		//			Expected: []sql.Row{
		//				{"postgres", "horses", "horses_name_idx", 10, 1, "name", "Steve", 10},
		//				{"postgres", "horses", "horses_name_idx", 167, 1, "name", "Steve", 167},
		//				{"postgres", "horses", "horses_name_idx", 197, 1, "name", "Steve", 197},
		//				{"postgres", "horses", "horses_name_idx", 306, 1, "name", "Steve", 306},
		//				{"postgres", "horses", "horses_name_idx", 320, 1, "name", "Steve", 320},
		//				{"postgres", "horses", "primary", 46, 46, "id", "1000", 1},
		//				{"postgres", "horses", "primary", 203, 203, "id", "954", 1},
		//				{"postgres", "horses", "primary", 347, 347, "id", "347", 1},
		//				{"postgres", "horses", "primary", 404, 404, "id", "751", 1},
		//			},
		//		},
		//		{
		//			Query:    `SELECT count(*) FROM dolt_statistics`,
		//			Expected: []sql.Row{{9}},
		//		},
		//		{
		//			Query:    `SELECT count(*) FROM public.dolt_statistics`,
		//			Expected: []sql.Row{{9}},
		//		},
		//		{
		//			Query:    `SELECT dolt_statistics.index_name FROM public.dolt_statistics GROUP BY index_name ORDER BY index_name`,
		//			Expected: []sql.Row{{"horses_name_idx"}, {"primary"}},
		//		},
		//		{
		//			Query:       `SELECT name FROM other.dolt_statistics`,
		//			ExpectedErr: "database schema not found",
		//		},
		//		{
		//			Query:    `CREATE SCHEMA newschema`,
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "SET search_path = 'newschema'",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    `SELECT count(*) FROM dolt_statistics`,
		//			Expected: []sql.Row{{0}},
		//		},
		//		{
		//			Query:    "CREATE TABLE horses2 (id int primary key, name varchar(10));",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "CREATE INDEX horses2_name_idx ON horses2(name);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "insert into horses2 select x, 'Steve' from (with recursive inputs(x) as (select 1 union select x+1 from inputs where x < 1000) select * from inputs) dt;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    `ANALYZE horses2;`,
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Skip:     true, // http://github.com/dolthub/doltgresql/issues/1352
		//			Query:    `SELECT dolt_statistics.index_name FROM dolt_statistics GROUP BY index_name ORDER BY index_name`,
		//			Expected: []sql.Row{{"horses2_name_idx"}, {"primary"}},
		//		},
		//		{
		//			Skip:     true, // http://github.com/dolthub/doltgresql/issues/1352
		//			Query:    `SELECT dolt_statistics.index_name FROM newschema.dolt_statistics GROUP BY index_name ORDER BY index_name`,
		//			Expected: []sql.Row{{"horses2_name_idx"}, {"primary"}},
		//		},
		//		{
		//			Query:    `SELECT dolt_statistics.index_name FROM public.dolt_statistics GROUP BY index_name ORDER BY index_name`,
		//			Expected: []sql.Row{{"horses_name_idx"}, {"primary"}},
		//		},
		//		// Same table name, different schema
		//		{
		//			Query:    "CREATE TABLE horses (id int primary key, name varchar(10));",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "CREATE INDEX horses3_name_idx ON horses(name);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "insert into horses select x, 'Steve' from (with recursive inputs(x) as (select 1 union select x+1 from inputs where x < 1000) select * from inputs) dt;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    `ANALYZE horses;`,
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query: `SELECT table_name, index_name FROM dolt_statistics GROUP BY table_name, index_name ORDER BY table_name, index_name`,
		//			Skip:  true, // TODO: seems to be flaky on CI, works locally no matter how many times it's run
		//			Expected: []sql.Row{
		//				{"horses", "horses3_name_idx"},
		//				{"horses", "primary"},
		//				{"horses2", "horses2_name_idx"},
		//				{"horses2", "primary"},
		//			},
		//		},
		//		{
		//			Query: `SELECT table_name, index_name FROM newschema.dolt_statistics GROUP BY table_name, index_name ORDER BY table_name, index_name`,
		//			Skip:  true, // TODO: seems to be flaky on CI, works locally no matter how many times it's run
		//			Expected: []sql.Row{
		//				{"horses", "horses3_name_idx"},
		//				{"horses", "primary"},
		//				{"horses2", "horses2_name_idx"},
		//				{"horses2", "primary"},
		//			},
		//		},
		//		{
		//			Query:    `SELECT table_name, index_name FROM public.dolt_statistics GROUP BY index_name ORDER BY index_name`,
		//			Expected: []sql.Row{{"horses", "horses_name_idx"}, {"horses", "primary"}},
		//		},
		//	},
		// },
		{
			Name: "dolt status",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM dolt.status`,
					Expected: []sql.Row{{"public.t", "f", "new table"}},
				},
				{
					Query:    `SELECT * FROM dolt_status`,
					Expected: []sql.Row{{"public.t", "f", "new table"}},
				},
				{
					Query: `DESCRIBE dolt."status"`,
					Expected: []sql.Row{
						{"table_name", "text", "NO", "PRI", nil, ""},
						{"staged", "boolean", "NO", "PRI", nil, ""},
						{"status", "text", "NO", "PRI", nil, ""},
					},
				},
				{
					Skip:  true, // TODO: ERROR: at or near "status": syntax error
					Query: `DESCRIBE dolt.status`,
					Expected: []sql.Row{
						{"table_name", "text", "NO", "PRI", nil, ""},
						{"staged", "boolean", "NO", "PRI", nil, ""},
						{"status", "text", "NO", "PRI", nil, ""},
					},
				},
				{
					Query: `DESCRIBE dolt_status`,
					Expected: []sql.Row{
						{"table_name", "text", "NO", "PRI", nil, ""},
						{"staged", "boolean", "NO", "PRI", nil, ""},
						{"status", "text", "NO", "PRI", nil, ""},
					},
				},
				{
					Query:    `SELECT * FROM dolt.status WHERE staged=true`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt.status.table_name FROM dolt.status`,
					Expected: []sql.Row{{"public.t"}},
				},
				{
					Query:    `SELECT dolt_status.table_name FROM dolt_status`,
					Expected: []sql.Row{{"public.t"}},
				},
				{
					Query:       `SELECT * FROM public.status`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM status`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE status (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO status VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM status`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT table_name FROM dolt.status`,
					Expected: []sql.Row{{"public.status"}, {"public.t"}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT table_name FROM status`,
					Expected: []sql.Row{{"public.status"}, {"public.t"}},
				},
				{
					Query:    `SELECT * FROM public.status`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM status`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM status`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM STATUS`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt tags",
			SetUpScript: []string{
				"SELECT dolt_tag('v1')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT tag_name FROM dolt.tags`,
					Expected: []sql.Row{{"v1"}},
				},
				{
					Query:    `SELECT tag_name FROM dolt_tags`,
					Expected: []sql.Row{{"v1"}},
				},
				{
					Query:    `SELECT dolt.tags.tag_name FROM dolt.tags`,
					Expected: []sql.Row{{"v1"}},
				},
				{
					Query:    `SELECT dolt_tags.tag_name FROM dolt_tags`,
					Expected: []sql.Row{{"v1"}},
				},
				{
					Query:       `SELECT * FROM public.tags`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM tags`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE tags (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO tags VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM tags`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT tag_name FROM dolt.tags`,
					Expected: []sql.Row{{"v1"}},
				},
				{
					Query:       `CREATE SCHEMA dolt`,
					ExpectedErr: "schema exists",
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT tag_name FROM tags`,
					Expected: []sql.Row{{"v1"}},
				},
				{
					Query:    `SELECT * FROM public.tags`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM tags`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM tags`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM TAGS`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name:        "dolt procedures",
			SetUpScript: []string{
				// TODO: Create procedure when supported
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM dolt_procedures`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM public.dolt_procedures`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_procedures.name FROM public.dolt_procedures`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT name FROM other.dolt_procedures`,
					ExpectedErr: "database schema not found",
				},
				// TODO: Add diff tests when create procedure works
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM newschema.dolt_procedures`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM dolt_procedures`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM public.dolt_procedures`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema,public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM dolt_procedures`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "dolt rebase",
			SetUpScript: []string{
				// create a simple table
				"create table t (pk int primary key);",
				"select dolt_commit('-Am', 'creating table t');",

				// create a new branch that we'll add more commits to later
				"select dolt_branch('branch1');",

				// create another commit on the main branch, right after where branch1 branched off
				"insert into t values (0);",
				"select dolt_commit('-am', 'inserting row 0');",

				// switch to branch1 and create three more commits that each insert one row
				"select dolt_checkout('branch1');",
				"insert into t values (1);",
				"select dolt_commit('-am', 'inserting row 1');",
				"insert into t values (2);",
				"select dolt_commit('-am', 'inserting row 2');",
				"insert into t values (3);",
				"select dolt_commit('-am', 'inserting row 3');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select message from dolt_log;",
					Expected: []sql.Row{
						{"inserting row 3"},
						{"inserting row 2"},
						{"inserting row 1"},
						{"creating table t"},
						{"CREATE DATABASE"},
						{"Initialize data repository"},
					},
				},
				{
					Query:    `select dolt_rebase('-i', 'main');`,
					Expected: []sql.Row{{"{0,\"interactive rebase started on branch dolt_rebase_branch1; adjust the rebase plan in the dolt_rebase table, then continue rebasing by calling dolt_rebase('--continue')\"}"}},
				},
				{
					Query: "select rebase_order, action, commit_message from dolt_rebase order by rebase_order;",
					Expected: []sql.Row{
						{float64(1), "pick", "inserting row 1"},
						{float64(2), "pick", "inserting row 2"},
						{float64(3), "pick", "inserting row 3"},
					},
				},
				{
					Query: "select rebase_order, action, commit_message from dolt.rebase order by rebase_order;",
					Expected: []sql.Row{
						{float64(1), "pick", "inserting row 1"},
						{float64(2), "pick", "inserting row 2"},
						{float64(3), "pick", "inserting row 3"},
					},
				},
				{
					Query: "select rebase.commit_message from dolt.rebase order by rebase_order;",
					Expected: []sql.Row{
						{"inserting row 1"},
						{"inserting row 2"},
						{"inserting row 3"},
					},
				},
				{
					Skip:  true, // TODO: table not found: dolt_rebase
					Query: "select dolt_rebase.commit_message from dolt_rebase order by rebase_order;",
					Expected: []sql.Row{
						{"inserting row 1"},
						{"inserting row 2"},
						{"inserting row 3"},
					},
				},
				{
					Query:       `SELECT * FROM public.rebase`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM rebase`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE rebase (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO rebase VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM rebase`,
					Expected: []sql.Row{{1}},
				},
				{
					Query: `SELECT commit_message FROM dolt.rebase`,
					Expected: []sql.Row{
						{"inserting row 1"},
						{"inserting row 2"},
						{"inserting row 3"},
					},
				},
				{
					Query:       `CREATE SCHEMA dolt`,
					ExpectedErr: "schema exists",
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT commit_message FROM rebase`,
					Expected: []sql.Row{
						{"inserting row 1"},
						{"inserting row 2"},
						{"inserting row 3"}},
				},
				{
					Query:    `SELECT * FROM public.rebase`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM rebase`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM rebase`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM REBASE`,
					Expected: []sql.Row{{1}},
				},
				{
					// Remove created table so we can continue with the rebase
					Query:    `DROP TABLE public.rebase;`,
					Expected: []sql.Row{},
				},
				{
					Query:    "update dolt.rebase set action='reword', commit_message='insert rows' where rebase_order=1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "update dolt.rebase set action='drop' where rebase_order=2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "update dolt_rebase set action='fixup' where rebase_order=3;",
					Expected: []sql.Row{},
				},
				{
					Query: "select rebase_order, action, commit_message from dolt_rebase order by rebase_order;",
					Expected: []sql.Row{
						{float64(1), "reword", "insert rows"},
						{float64(2), "drop", "inserting row 2"},
						{float64(3), "fixup", "inserting row 3"},
					},
				},
				{
					Query: "select rebase_order, action, commit_message from dolt.rebase order by rebase_order;",
					Expected: []sql.Row{
						{float64(1), "reword", "insert rows"},
						{float64(2), "drop", "inserting row 2"},
						{float64(3), "fixup", "inserting row 3"},
					},
				},
				{
					Query:    "select dolt_rebase('--continue');",
					Expected: []sql.Row{{"{0,\"Successfully rebased and updated refs/heads/branch1\"}"}},
				},
				{
					Query: "select message from dolt_log;",
					Expected: []sql.Row{
						{"insert rows"},
						{"inserting row 0"},
						{"creating table t"},
						{"CREATE DATABASE"},
						{"Initialize data repository"},
					},
				},
				{
					Query:       "select * from dolt_rebase;",
					ExpectedErr: "table not found: dolt_rebase",
				},
				{
					Query:       "select * from dolt.rebase;",
					ExpectedErr: "table not found: rebase",
				},
			},
		},
		{
			Name: "dolt remote branches",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT name FROM dolt.remote_branches`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM dolt_remote_branches`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt.remote_branches.name FROM dolt.remote_branches`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_remote_branches.name FROM dolt_remote_branches`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM public.remote_branches`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM remote_branches`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE remote_branches (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO remote_branches VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM remote_branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT name FROM dolt.remote_branches`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM remote_branches`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM public.remote_branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM remote_branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM remote_branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM REMOTE_BRANCHES`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt remotes",
			SetUpScript: []string{
				"SELECT dolt_remote('add', 'origin', 'https://doltremoteapi.dolthub.com/dolthub/test')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT name FROM dolt.remotes`,
					Expected: []sql.Row{{"origin"}},
				},
				{
					Query:    `SELECT name FROM dolt_remotes`,
					Expected: []sql.Row{{"origin"}},
				},
				{
					Query:    `SELECT dolt.remotes.name FROM dolt.remotes`,
					Expected: []sql.Row{{"origin"}},
				},
				{
					Query:    `SELECT dolt_remotes.name FROM dolt_remotes`,
					Expected: []sql.Row{{"origin"}},
				},
				{
					Query:       `SELECT * FROM public.remotes`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM remotes`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE remotes (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO remotes VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM remotes`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT name FROM dolt.remotes`,
					Expected: []sql.Row{{"origin"}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM remotes`,
					Expected: []sql.Row{{"origin"}},
				},
				{
					Query:    `SELECT * FROM public.remotes`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM remotes`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM remotes`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM REMOTES`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt schema conflicts",
			SetUpScript: []string{
				"CREATE TABLE test (pk int primary key, c0 varchar(20))",
				"SELECT dolt_commit('-Am', 'added table t')",
				"SELECT dolt_checkout('-b', 'other')",
				"ALTER TABLE test ALTER COLUMN c0 TYPE int",
				"SELECT dolt_commit('-am', 'altered t on branch other')",
				"SELECT dolt_checkout('main')",
				"ALTER TABLE test ALTER COLUMN c0 TYPE date",
				"SELECT dolt_commit('-am', 'altered t on branch main')",
				"START TRANSACTION",
				"SELECT dolt_merge('other')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT table_name FROM dolt.schema_conflicts`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    `SELECT table_name FROM dolt_schema_conflicts`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    `SELECT dolt.schema_conflicts.table_name FROM dolt.schema_conflicts`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    `SELECT dolt_schema_conflicts.table_name FROM dolt_schema_conflicts`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:       `SELECT * FROM public.schema_conflicts`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM schema_conflicts`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE schema_conflicts (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO schema_conflicts VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM schema_conflicts`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT table_name FROM dolt.schema_conflicts`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT table_name FROM schema_conflicts`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    `SELECT * FROM public.schema_conflicts`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM schema_conflicts`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public,dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM schema_conflicts`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM SCHEMA_CONFLICTS`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt schemas",
			SetUpScript: []string{
				"create view myView as select 2 + 2",
				// TODO: Add more tests when triggers and events work in doltgres
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM dolt_schemas`,
					Expected: []sql.Row{
						{
							"view",
							"myview",
							"create view myView as select 2 + 2",
							"{\"CreatedAt\":0}",
							"NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES",
						},
					},
				},
				{
					Query: `SELECT * FROM public.dolt_schemas`,
					Expected: []sql.Row{
						{
							"view",
							"myview",
							"create view myView as select 2 + 2",
							"{\"CreatedAt\":0}",
							"NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES",
						},
					},
				},
				{
					Query:    `SELECT dolt_schemas.name FROM public.dolt_schemas`,
					Expected: []sql.Row{{"myview"}},
				},
				{
					Query:    `SELECT * FROM public.myview`,
					Expected: []sql.Row{{4}},
				},
				{
					Query:       `SELECT name FROM other.dolt_schemas`,
					ExpectedErr: "database schema not found",
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING')`,
					Expected: []sql.Row{
						{"", "public.dolt_schemas", "added", 1, 1},
					},
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_schemas')`,
					Expected: []sql.Row{
						{"", "public.dolt_schemas", "added", 1, 1},
					},
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_schemas')`,
					Expected: []sql.Row{
						{"", "public.dolt_schemas", "added", 1, 1},
					},
				},
				{
					Query: `SELECT diff_type, from_name, to_name FROM dolt_diff('main', 'WORKING', 'dolt_schemas')`,
					Expected: []sql.Row{
						{"added", nil, "myview"},
					},
				},
				{
					Query: `SELECT diff_type, from_name, to_name FROM dolt_diff('main', 'WORKING', 'dolt_schemas')`,
					Expected: []sql.Row{
						{"added", nil, "myview"},
					},
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM myview`,
					ExpectedErr: "table not found: myview",
				},
				{
					Query:    `SELECT * FROM public.myview`,
					Expected: []sql.Row{{4}},
				},
				{
					Query:    `CREATE VIEW testView AS SELECT 1 + 1`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM newschema.dolt_schemas`,
					Expected: []sql.Row{
						{
							"view",
							"testview",
							"CREATE VIEW testView AS SELECT 1 + 1",
							"{\"CreatedAt\":0}",
							"NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES",
						},
					},
				},
				{
					Query:    `SELECT name FROM dolt_schemas`,
					Expected: []sql.Row{{"testview"}},
				},
				{
					Query: "SELECT table_schema, table_name FROM information_schema.views",
					Expected: []sql.Row{
						{"newschema", "testview"},
						{"public", "myview"},
					},
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_schemas')`,
					Expected: []sql.Row{
						{"", "newschema.dolt_schemas", "added", 1, 1},
					},
				},
				{
					Skip:  true, // TODO: Should be able to specify schema
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'public.dolt_schemas')`,
					Expected: []sql.Row{
						{"", "public.dolt_schemas", "added", 1, 1},
					},
				},
				{
					Query:    `SELECT name FROM public.dolt_schemas`,
					Expected: []sql.Row{{"myview"}},
				},
				{
					Query:       "DROP VIEW myView",
					ExpectedErr: "the view postgres.myview does not exist",
				},
				{
					Query:    "DROP VIEW public.myView",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM public.dolt_schemas`,
					Expected: []sql.Row{},
				},
				{
					Query:    "create view public.myNewView as select 3 + 3",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM public.dolt_schemas`,
					Expected: []sql.Row{{"mynewview"}},
				},
				{
					Query:    `SELECT name FROM dolt_schemas`,
					Expected: []sql.Row{{"testview"}},
				},
				{
					Query:    "SET search_path = 'newschema,public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM dolt_schemas`,
					Expected: []sql.Row{{"testview"}},
				},
				{
					Query: `SELECT * FROM dolt_diff_summary('main', 'WORKING', 'dolt_schemas')`,
					Expected: []sql.Row{
						{"", "newschema.dolt_schemas", "added", 1, 1},
					},
				},
				// Test same view name on different schemas
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE VIEW testView AS SELECT 4 + 4`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT name, fragment FROM dolt_schemas`,
					Expected: []sql.Row{
						{"mynewview", "create view public.myNewView as select 3 + 3"},
						{"testview", "CREATE VIEW testView AS SELECT 4 + 4"},
					},
				},
				{
					Query:    `SELECT name, fragment FROM newschema.dolt_schemas`,
					Expected: []sql.Row{{"testview", "CREATE VIEW testView AS SELECT 1 + 1"}},
				},
				{
					Query: `SELECT name, fragment FROM dolt_schemas`,
					Expected: []sql.Row{
						{"mynewview", "create view public.myNewView as select 3 + 3"},
						{"testview", "CREATE VIEW testView AS SELECT 4 + 4"},
					},
				},
				{
					Query:    "DROP VIEW IF EXISTS noexist.testView",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP VIEW IF EXISTS newschema.testView",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM newschema.dolt_schemas`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM dolt_schemas`,
					Expected: []sql.Row{{"mynewview"}, {"testview"}},
				},
			},
		},
		{
			Name: "dolt workspace with tablename",
			SetUpScript: []string{
				"CREATE TABLE test (id INT PRIMARY KEY)",
				"INSERT INTO test VALUES (10)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test`,
					Expected: []sql.Row{{0, "f", nil, 10}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test`,
					Expected: []sql.Row{{0, "f", nil, 10}},
				},
				{
					Query:    `SELECT dolt_workspace_test.id FROM public.dolt_workspace_test`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:       `SELECT * FROM other.dolt_workspace_test`,
					ExpectedErr: "database schema not found",
				},
				{
					Query:    `SELECT * FROM public.dolt_workspace_none`,
					Expected: []sql.Row{}, // dolt_workspace empty for unknown table
				},
				{
					Query:    `CREATE SCHEMA newschema`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SET search_path = 'newschema'",
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_sch (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test_sch VALUES (11)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('test_sch')`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM newschema.dolt_workspace_test_sch`,
					Expected: []sql.Row{{0, "t", nil, 11}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test_sch`,
					Expected: []sql.Row{{0, "t", nil, 11}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test_sch WHERE staged=true`,
					Expected: []sql.Row{{0, "t", nil, 11}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test_sch WHERE staged=false`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_workspace_test`,
					Expected: []sql.Row{}, // dolt_workspace empty for unknown table
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test`,
					Expected: []sql.Row{{0, "f", nil, 10}},
				},
				{
					Query:    `SELECT * FROM public.dolt_workspace_test_sch`,
					Expected: []sql.Row{}, // dolt_workspace empty for unknown table
				},
				{
					Query:    `SELECT * FROM newschema.dolt_workspace_test`,
					Expected: []sql.Row{}, // dolt_workspace empty for unknown table
				},
				{
					// Same name as table in public schema
					Query:    `CREATE TABLE test (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test VALUES (12)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM newschema.dolt_workspace_test`,
					Expected: []sql.Row{{0, "f", nil, 12}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test`,
					Expected: []sql.Row{{0, "f", nil, 12}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test`,
					Expected: []sql.Row{{0, "f", nil, 10}},
				},
				{
					Query:    "SET search_path = 'newschema,public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test`,
					Expected: []sql.Row{{0, "f", nil, 12}},
				},
			},
		},
	})
}
