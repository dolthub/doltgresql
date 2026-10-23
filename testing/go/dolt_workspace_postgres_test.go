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

// TestPostgresWorkspaceCommands restates the add, branch, checkout, clean,
// reset, rm, revert, cherry-pick and log scenarios of dolt_functions_test.go
// over plain tables. The originals mix array primary keys, serial columns,
// foreign keys, generated columns, triggers and routines, which the
// PostgreSQL fork's storage does not support yet, and compare bootstrap
// commit counts. Hashes are not compared: the engines use different formats.
func TestPostgresWorkspaceCommands(t *testing.T) {
	tables := []string{
		"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
		"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
		"CREATE TABLE t_values (pk INT4 PRIMARY KEY, v1 INT8, v2 TEXT);",
		"CREATE TABLE t_keyless (v1 INT4);",
		"INSERT INTO t_simple VALUES (1), (2), (3);",
		"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
		"INSERT INTO t_values VALUES (1, 10, 'a'), (2, 20, 'b');",
		"INSERT INTO t_keyless VALUES (7), (8), (9);",
	}
	twoCommits := append(append([]string{}, tables...),
		"SELECT DOLT_COMMIT('-A', '-m', 'initial');",
		"SELECT DOLT_BRANCH('original');",
		"INSERT INTO t_simple VALUES (4);",
		"INSERT INTO t_composite VALUES (3, 100);",
		"UPDATE t_values SET v1 = 11 WHERE pk = 1;",
		"INSERT INTO t_keyless VALUES (10);",
		"SELECT DOLT_COMMIT('-A', '-m', 'changes');",
	)
	RunScripts(t, []ScriptTest{
		{
			Name:        "add all and individually",
			SetUpScript: tables,
			Assertions: []ScriptTestAssertion{
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';", Expected: []sql.Row{{4}}},
				{Query: "SELECT DOLT_ADD('t_simple', 't_composite');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';", Expected: []sql.Row{{2}}},
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';", Expected: []sql.Row{{2}}},
				{Query: "SELECT DOLT_RESET('t_simple');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT table_name FROM dolt_status WHERE staged = 't';", Expected: []sql.Row{{"public.t_composite"}}},
				{Query: "SELECT DOLT_ADD('.');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';", Expected: []sql.Row{{4}}},
				{Query: "SELECT DOLT_RESET();", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';", Expected: []sql.Row{{0}}},
				{Query: "SELECT DOLT_ADD('-A');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';", Expected: []sql.Row{{4}}},
				{Query: "SELECT DOLT_RM('--cached', 't_keyless');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT staged FROM dolt_status WHERE table_name = 'public.t_keyless';", Expected: []sql.Row{{"f"}}},
			},
		},
		{
			Name:        "branch options",
			SetUpScript: twoCommits,
			Assertions: []ScriptTestAssertion{
				{Query: "SELECT DOLT_BRANCH('-c', 'main', 'copy');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT DOLT_BRANCH('-c', '-f', 'original', 'forcecopy');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}, {4}}},
				{Query: "SELECT DOLT_CHECKOUT('main');", Expected: []sql.Row{{[]any{int64(0), "Already on branch 'main'"}}}},
				{Query: "SELECT DOLT_CHECKOUT('original');", Expected: []sql.Row{{[]any{int64(0), "Switched to branch 'original'"}}}},
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}}},
				{Query: "SELECT DOLT_CHECKOUT('copy');", Expected: []sql.Row{{[]any{int64(0), "Switched to branch 'copy'"}}}},
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}, {4}}},
				{Query: "SELECT DOLT_CHECKOUT('forcecopy');", Expected: []sql.Row{{[]any{int64(0), "Switched to branch 'forcecopy'"}}}},
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}}},
				{Query: "SELECT DOLT_BRANCH('-d', 'forcecopy');", ExpectedErr: "Cannot delete checked out branch 'forcecopy'"},
				{Query: "SELECT DOLT_BRANCH('-d', 'original');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT DOLT_BRANCH('-m', 'copy', 'renamedcopy');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT DOLT_CHECKOUT('original');", ExpectedErr: "'original' did not match any table"},
				{Query: "SELECT DOLT_CHECKOUT('copy');", ExpectedErr: "'copy' did not match any table"},
				{Query: "SELECT DOLT_CHECKOUT('renamedcopy');", Expected: []sql.Row{{[]any{int64(0), "Switched to branch 'renamedcopy'"}}}},
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}, {4}}},
				{Query: "SELECT * FROM DOLT_BRANCH_STATUS('main', 'renamedcopy');", Expected: []sql.Row{{"renamedcopy", Numeric("0"), Numeric("0")}}},
				{Query: "SELECT COUNT(*) FROM DOLT_LOG('main') WHERE message IN ('initial', 'changes');", Expected: []sql.Row{{2}}},
				{Query: "SELECT (SELECT COUNT(*) FROM DOLT_LOG('main')) - (SELECT COUNT(*) FROM DOLT_LOG('main~1'));", Expected: []sql.Row{{1}}},
			},
		},
		{
			Name:        "checkout tables and clean",
			SetUpScript: twoCommits,
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_values SET v2 = 'changed' WHERE pk = 2;", Expected: []sql.Row{}},
				{Query: "CREATE TABLE untracked (pk INT4 PRIMARY KEY);", Expected: []sql.Row{}},
				{Query: "SELECT table_name, staged, status FROM dolt_status ORDER BY table_name;", Expected: []sql.Row{{"public.t_values", "f", "modified"}, {"public.untracked", "f", "new table"}}},
				{Query: "SELECT DOLT_CHECKOUT('t_values');", Expected: []sql.Row{{[]any{int64(0), ""}}}},
				{Query: "SELECT v2 FROM t_values WHERE pk = 2;", Expected: []sql.Row{{"b"}}},
				{Query: "SELECT DOLT_CLEAN('--dry-run');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT COUNT(*) FROM dolt_status;", Expected: []sql.Row{{1}}},
				{Query: "SELECT DOLT_CLEAN();", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT COUNT(*) FROM dolt_status;", Expected: []sql.Row{{0}}},
				{Query: "SELECT DOLT_CHECKOUT('nothing');", ExpectedErr: "'nothing' did not match any table"},
			},
		},
		{
			Name:        "reset hard and soft",
			SetUpScript: twoCommits,
			Assertions: []ScriptTestAssertion{
				{Query: "INSERT INTO t_simple VALUES (5);", Expected: []sql.Row{}},
				{Query: "SELECT DOLT_ADD('-A');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';", Expected: []sql.Row{{1}}},
				{Query: "SELECT DOLT_RESET('--soft', 'HEAD');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 'f';", Expected: []sql.Row{{0}}},
				{Query: "SELECT COUNT(*) FROM dolt_status WHERE staged = 't';", Expected: []sql.Row{{1}}},
				{Query: "SELECT DOLT_RESET('--hard', 'HEAD~1');", Expected: []sql.Row{{int64(0)}}},
				{Query: "SELECT * FROM dolt_status;", Expected: []sql.Row{}},
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}}},
				{Query: "SELECT message FROM DOLT_LOG() LIMIT 1;", Expected: []sql.Row{{"initial"}}},
				{Query: "SELECT DOLT_RESET('--hard', '--soft');", ExpectedErr: "mutually exclusive"},
			},
		},
		{
			Name:        "revert and cherry-pick",
			SetUpScript: twoCommits,
			Assertions: []ScriptTestAssertion{
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}, {4}}},
				{Query: "SELECT data_conflicts, schema_conflicts, constraint_violations FROM DOLT_REVERT('HEAD');", Expected: []sql.Row{{int64(0), int64(0), int64(0)}}},
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}}},
				{Query: "SELECT v1 FROM t_values WHERE pk = 1;", Expected: []sql.Row{{10}}},
				{Query: "SELECT message FROM DOLT_LOG() LIMIT 1;", Expected: []sql.Row{{"Revert \"changes\""}}},
				{Query: "SELECT DOLT_CHECKOUT('original');", Expected: []sql.Row{{[]any{int64(0), "Switched to branch 'original'"}}}},
				{Query: "SELECT data_conflicts, schema_conflicts, constraint_violations FROM DOLT_CHERRY_PICK('main~1');", Expected: []sql.Row{{int64(0), int64(0), int64(0)}}},
				{Query: "SELECT * FROM t_simple ORDER BY pk;", Expected: []sql.Row{{1}, {2}, {3}, {4}}},
				{Query: "SELECT message FROM DOLT_LOG() LIMIT 1;", Expected: []sql.Row{{"changes"}}},
			},
		},
	})
}
