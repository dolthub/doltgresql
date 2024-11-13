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
					Query:    `SELECT name FROM dolt.branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:    `SELECT name FROM dolt_branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:    `SELECT branches.name FROM dolt.branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
					Query:    `SELECT dolt.branches.name FROM dolt.branches`,
					Expected: []sql.Row{{"main"}},
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
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
					Query:    `SELECT dolt.column_diff.commit_hash FROM dolt.column_diff`,
					Expected: []sql.Row{},
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
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
					Expected: []sql.Row{{"test", Numeric("1")}},
				},
				{
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
					Expected: []sql.Row{{"test", Numeric("2")}},
				},
				{
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
					Query: `SELECT violation_type, pk, col1 FROM dolt_constraint_violations_test`,
					Expected: []sql.Row{
						{"unique index", 1, 1},
						{"unique index", 2, 1},
					},
				},
				{
					Query: `SELECT violation_type, pk, col1 FROM public.dolt_constraint_violations_test`,
					Expected: []sql.Row{
						{"unique index", 1, 1},
						{"unique index", 2, 1},
					},
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
					Query:    `SELECT table_name FROM dolt_diff`,
					Expected: []sql.Row{{"public.test"}},
				},
				{
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT is_merging FROM dolt_merge_status`,
					Expected: []sql.Row{{0}},
				},
				{
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
					Query:    `SELECT dolt.merge_status.is_merging FROM dolt.merge_status`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT dolt_merge_status.is_merging FROM dolt_merge_status`,
					Expected: []sql.Row{{0}},
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
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT is_merging FROM merge_status`,
					Expected: []sql.Row{{0}},
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
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
			Name: "dolt docs",
			SetUpScript: []string{
				"INSERT INTO dolt.docs values ('README.md', 'testing')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM dolt.docs`,
					Expected: []sql.Row{
						{"README.md", "testing"},
					},
				},
				{
					Query: `SELECT * FROM dolt_docs`,
					Expected: []sql.Row{
						{"README.md", "testing"},
					},
				},
				{
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
					Query:    `SELECT dolt.docs.doc_name FROM dolt.docs`,
					Expected: []sql.Row{{"README.md"}},
				},
				{
					Skip:     true, // TODO: table not found: dolt_docs
					Query:    `SELECT dolt_docs.doc_name FROM dolt_docs`,
					Expected: []sql.Row{{"README.md"}},
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
					Query:    `SELECT doc_name FROM dolt.docs`,
					Expected: []sql.Row{{"README.md"}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT doc_name FROM docs`,
					Expected: []sql.Row{{"README.md"}},
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
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
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
					Expected: []sql.Row{{Numeric("0"), 0, nil, 10}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test`,
					Expected: []sql.Row{{Numeric("0"), 0, nil, 10}},
				},
				{
					Query:    `SELECT dolt_workspace_test.id FROM public.dolt_workspace_test`,
					Expected: []sql.Row{{Numeric("0")}},
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
					Expected: []sql.Row{{Numeric("0"), 1, nil, 11}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test_sch`,
					Expected: []sql.Row{{Numeric("0"), 1, nil, 11}},
				},
				{
					Query:    `SELECT * FROM dolt_workspace_test`,
					Expected: []sql.Row{}, // dolt_workspace empty for unknown table
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test`,
					Expected: []sql.Row{{Numeric("0"), 0, nil, 10}},
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
					Expected: []sql.Row{{Numeric("0"), 0, nil, 12}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test`,
					Expected: []sql.Row{{Numeric("0"), 0, nil, 12}},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM public.dolt_workspace_test`,
					Expected: []sql.Row{{Numeric("0"), 0, nil, 10}},
				},
				{
					Query:    "SET search_path = 'newschema,public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, staged, from_id, to_id FROM dolt_workspace_test`,
					Expected: []sql.Row{{Numeric("0"), 0, nil, 12}},
				},
			},
		},
	})
}
