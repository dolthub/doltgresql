// Copyright 2026 Dolthub, Inc.
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

// TestPostgresDoltTables restates the system table scenarios from
// TestUserSpaceDoltTables whose original assertions depend on Doltgres-specific
// details: the extra CREATE DATABASE bootstrap commit, running version-control
// commands inside an explicit transaction, and MySQL DESCRIBE output. The
// expectations here hold on both engines.
func TestPostgresDoltTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "commits and ancestors relative to the bootstrap history",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY)",
				"SELECT dolt_commit('-Am', 'first')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT count(*) >= 2 FROM dolt.commits",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT message FROM dolt_commits LIMIT 1",
					Expected: []sql.Row{{"first"}},
				},
				{
					Query:    "SELECT count(*) FROM dolt.commits WHERE message = 'Initialize data repository'",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT (SELECT count(*) FROM dolt.commit_ancestors) = (SELECT count(*) FROM dolt.commits)",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT count(*) FROM dolt.commit_ancestors WHERE parent_hash IS NULL",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT parent_index FROM dolt_commit_ancestors WHERE commit_hash = dolt_hashof('HEAD')",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT parent_hash = dolt_hashof('HEAD~1') FROM dolt.commit_ancestors WHERE commit_hash = dolt_hashof('HEAD')",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT (SELECT count(*) FROM dolt.log) = (SELECT count(*) FROM dolt_log)",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT count(*) >= 2 FROM dolt.log",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "status describes staged and unstaged changes",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY)",
				"SELECT dolt_commit('-Am', 'first')",
				"CREATE SCHEMA s2",
				"CREATE TABLE s2.u (id INT PRIMARY KEY)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM dolt.status ORDER BY table_name",
					Expected: []sql.Row{{"s2", "f", "new schema"}, {"s2.u", "f", "new table"}},
				},
				{
					Query:    "SELECT dolt_add('.')",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "INSERT INTO t VALUES (1)",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM dolt_status WHERE staged = true ORDER BY table_name",
					Expected: []sql.Row{{"s2", "t", "new schema"}, {"s2.u", "t", "new table"}},
				},
				{
					Query:    "SELECT * FROM dolt_status WHERE staged = false",
					Expected: []sql.Row{{"public.t", "f", "modified"}},
				},
				{
					Query:    "SELECT commit_hash, table_name, data_change, schema_change FROM dolt.diff WHERE commit_hash IN ('STAGED', 'WORKING') ORDER BY 1, 2",
					Expected: []sql.Row{{"STAGED", "s2.u", "f", "t"}, {"WORKING", "public.t", "t", "f"}},
				},
			},
		},
		{
			Name: "branches report dirtiness and latest commits",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY)",
				"SELECT dolt_commit('-Am', 'first', '--author', 'Author <author@example.com>')",
				"SELECT dolt_branch('other')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT name, dirty, latest_commit_message, latest_author, latest_author_email FROM dolt.branches ORDER BY name",
					Expected: []sql.Row{{"main", "f", "first", "Author", "author@example.com"}, {"other", "f", "first", "Author", "author@example.com"}},
				},
				{
					Query:    "SELECT hash = dolt_hashof('main') FROM dolt_branches WHERE name = 'other'",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "INSERT INTO t VALUES (1)",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT name, dirty FROM dolt.branches ORDER BY name",
					Expected: []sql.Row{{"main", "t"}, {"other", "f"}},
				},
			},
		},
		{
			Name: "conflicts and merge status without an enclosing transaction",
			SetUpScript: []string{
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
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT is_merging FROM dolt.merge_status",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT conflicts FROM dolt_merge('b2')",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT * FROM dolt.conflicts",
					Expected: []sql.Row{{"test", Numeric("1")}},
				},
				{
					Query:    "SELECT dolt_conflicts.table FROM dolt_conflicts",
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    "SELECT is_merging, source, source_commit = dolt_hashof('b2'), target, unmerged_tables FROM dolt.merge_status",
					Expected: []sql.Row{{"t", "b2", "t", "refs/heads/main", "test"}},
				},
				{
					Query:    "SELECT * FROM dolt_status",
					Expected: []sql.Row{{"public.test", "f", "conflict"}},
				},
				{
					Query:    "SELECT dolt_conflicts_resolve('--ours', 'test')",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT * FROM dolt.conflicts",
					Expected: []sql.Row{},
				},
				{
					Query:            "SELECT dolt_commit('-m', 'merged')",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT is_merging FROM dolt_merge_status",
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "docs keep the default agent document",
			SetUpScript: []string{
				"INSERT INTO dolt.docs VALUES ('README.md', 'testing')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT doc_name FROM dolt.docs ORDER BY 1",
					Expected: []sql.Row{{"AGENT.md"}, {"README.md"}},
				},
				{
					Query:    "SELECT doc_text FROM dolt_docs WHERE doc_name = 'README.md'",
					Expected: []sql.Row{{"testing"}},
				},
				{
					Query:    "SELECT doc_text LIKE '# AGENT.md%' FROM public.dolt_docs WHERE doc_name = 'AGENT.md'",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "DELETE FROM dolt_docs WHERE doc_name = 'README.md'",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT doc_name FROM dolt.docs",
					Expected: []sql.Row{{"AGENT.md"}},
				},
			},
		},
	})
}
