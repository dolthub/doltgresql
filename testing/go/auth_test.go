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
	"os"
	"path/filepath"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// fileUrl returns a file:// URL path for a temp file.
func fileUrl(path string) string {
	path = filepath.Join(os.TempDir(), path)
	return "file://" + filepath.ToSlash(filepath.Clean(path))
}

func TestAuthTests(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: `Simple CREATE USER and DROP USER`,
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT 1;`,
					Username:    `user1`,
					Password:    `hello`,
					ExpectedErr: `authentication failed`,
				},
				{
					Query:    `CREATE USER user1 PASSWORD 'hello';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 2;`,
					Username: `user1`,
					Password: `hello`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `DROP USER user1;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT 3;`,
					Username:    `user1`,
					Password:    `hello`,
					ExpectedErr: `authentication failed`,
				},
			},
		},
		{
			Name: `ALTER PASSWORD`,
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE USER user1 PASSWORD 'something';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 1;`,
					Username: `user1`,
					Password: `something`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `ALTER USER user1 PASSWORD 'another_thing';`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT 2;`,
					Username:    `user1`,
					Password:    `something`,
					ExpectedErr: `authentication failed`,
				},
				{
					Query:    `SELECT 3;`,
					Username: `user1`,
					Password: `another_thing`,
					Expected: []sql.Row{{3}},
				},
				{ // No password will work, the user is effectively unable to be accessed with password-based auth
					Query:    `ALTER USER user1 WITH PASSWORD NULL;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT 4;`,
					Username:    `user1`,
					Password:    `something`,
					ExpectedErr: `authentication failed`,
				},
				{
					Query:       `SELECT 5;`,
					Username:    `user1`,
					Password:    `another_thing`,
					ExpectedErr: `authentication failed`,
				},
				{
					Query:       `SELECT 6;`,
					Username:    `user1`,
					Password:    ``, // Even the empty password won't work
					ExpectedErr: `authentication failed`,
				},
				{
					Query:    `ALTER USER user1 PASSWORD 'different484';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 7;`,
					Username: `user1`,
					Password: `different484`,
					Expected: []sql.Row{{7}},
				},
			},
		},
		{
			Name: `ALTER LOGIN`,
			Assertions: []ScriptTestAssertion{
				{ // By default, roles cannot be logged into
					Query:    `CREATE ROLE user1 PASSWORD 'pass1';`,
					Expected: []sql.Row{},
				},
				{ // Users can be logged into by default, this is the only difference between roles and users
					Query:    `CREATE USER user2 PASSWORD 'pass2';`,
					Expected: []sql.Row{},
				},
				{ // A role with LOGIN defined is exactly equivalent to a default user
					Query:    `CREATE ROLE user3 PASSWORD 'pass3' LOGIN;`,
					Expected: []sql.Row{},
				},
				{ // A user with NOLOGIN defined is exactly equivalent to a default role
					Query:    `CREATE USER user4 PASSWORD 'pass4' NOLOGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT 1;`,
					Username:    `user1`,
					Password:    `pass1`,
					ExpectedErr: `authentication failed`,
				},
				{
					Query:    `SELECT 2;`,
					Username: `user2`,
					Password: `pass2`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT 3;`,
					Username: `user3`,
					Password: `pass3`,
					Expected: []sql.Row{{3}},
				},
				{
					Query:       `SELECT 4;`,
					Username:    `user4`,
					Password:    `pass4`,
					ExpectedErr: `authentication failed`,
				},
				{ // We'll flip LOGIN/NOLOGIN statuses
					Query:    `ALTER USER user1 WITH LOGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `ALTER USER user2 WITH NOLOGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 5;`,
					Username: `user1`,
					Password: `pass1`,
					Expected: []sql.Row{{5}},
				},
				{
					Query:       `SELECT 6;`,
					Username:    `user2`,
					Password:    `pass2`,
					ExpectedErr: `authentication failed`,
				},
			},
		},
		{
			Name: `CREATE USER IF NOT EXISTS`,
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT 1;`,
					Username:    `user1`,
					Password:    `hello`,
					ExpectedErr: `authentication failed`,
				},
				{
					Query:    `CREATE USER user1 PASSWORD 'hello1';`,
					Expected: []sql.Row{},
				},
				{
					Query:       `CREATE USER user1 PASSWORD 'hello2';`,
					ExpectedErr: `already exists`,
				},
				{
					Query:    `CREATE USER IF NOT EXISTS user1 PASSWORD 'hello3';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 2;`,
					Username: `user1`,
					Password: `hello1`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `CREATE ROLE IF NOT EXISTS user2 PASSWORD 'hi1' LOGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `CREATE ROLE user2 PASSWORD 'hi2' LOGIN;`,
					ExpectedErr: `already exists`,
				},
				{
					Query:    `CREATE ROLE IF NOT EXISTS user2 PASSWORD 'hi3' LOGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 3;`,
					Username: `user2`,
					Password: `hi1`,
					Expected: []sql.Row{{3}},
				},
			},
		},
		{
			Name: `DROP USER IF EXISTS`,
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE USER user1 PASSWORD 'hello1';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE USER user2 PASSWORD 'hello2';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 1;`,
					Username: `user1`,
					Password: `hello1`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT 2;`,
					Username: `user2`,
					Password: `hello2`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `DROP USER user1;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `DROP USER user1;`,
					ExpectedErr: `does not exist`,
				},
				{
					Query:    `DROP USER IF EXISTS user1;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT 3;`,
					Username:    `user1`,
					Password:    `hello1`,
					ExpectedErr: `authentication failed`,
				},
				{
					Query:    `DROP ROLE IF EXISTS user2;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `DROP ROLE user2;`,
					ExpectedErr: `does not exist`,
				},
				{
					Query:    `DROP ROLE IF EXISTS user2;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT 4;`,
					Username:    `user2`,
					Password:    `hello2`,
					ExpectedErr: `authentication failed`,
				},
			},
		},
		{
			Name: `DROP USER with multiple users`,
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE USER user1 PASSWORD 'hello1';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE USER user2 PASSWORD 'hello2';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE USER user3 PASSWORD 'hello3';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 1;`,
					Username: `user1`,
					Password: `hello1`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT 2;`,
					Username: `user2`,
					Password: `hello2`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT 3;`,
					Username: `user3`,
					Password: `hello3`,
					Expected: []sql.Row{{3}},
				},
				{
					Query:    `DROP USER user1, user3;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT 4;`,
					Username:    `user1`,
					Password:    `hello1`,
					ExpectedErr: `authentication failed`,
				},
				{
					Query:    `SELECT 5;`,
					Username: `user2`,
					Password: `hello2`,
					Expected: []sql.Row{{5}},
				},
				{
					Query:       `SELECT 6;`,
					Username:    `user3`,
					Password:    `hello3`,
					ExpectedErr: `authentication failed`,
				},
			},
		},
		{
			Name: `GRANT/REVOKE SELECT Privilege`,
			SetUpScript: []string{
				`CREATE USER user1 PASSWORD 'a';`,
				`CREATE USER user2 PASSWORD 'b';`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user2;`,
				`GRANT ALL PRIVILEGES ON test TO user1 WITH GRANT OPTION;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE test (pk INT4 PRIMARY KEY);`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO test VALUES (1), (5), (6);`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM test ORDER BY pk`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{{1}, {5}, {6}},
				},
				{
					Query:       `SELECT * FROM test ORDER BY pk`,
					Username:    `user2`,
					Password:    `b`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `GRANT SELECT ON test TO user2;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM test ORDER BY pk`,
					Username: `user2`,
					Password: `b`,
					Expected: []sql.Row{{1}, {5}, {6}},
				},
				{
					Query:    `REVOKE SELECT ON test FROM user2;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test ORDER BY pk`,
					Username:    `user2`,
					Password:    `b`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `GRANT SELECT ON test TO PUBLIC;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM test ORDER BY pk`,
					Username: `user2`,
					Password: `b`,
					Expected: []sql.Row{{1}, {5}, {6}},
				},
			},
		},
		{
			Name: `INSERT, UPDATE, DELETE Privileges`,
			SetUpScript: []string{
				`CREATE TABLE test (pk INT4 PRIMARY KEY);`,
				`INSERT INTO test VALUES (1), (6), (7);`,
				`CREATE USER user1 PASSWORD 'a';`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT * FROM test ORDER BY pk;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:       `WITH cte AS (SELECT * FROM test ORDER BY pk) SELECT * FROM cte;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:       `INSERT INTO test VALUES (10);`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:       `UPDATE test SET pk=pk+20;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:       `DELETE FROM test WHERE pk > 3;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `GRANT SELECT, INSERT, UPDATE, DELETE ON test TO user1;`,
					Username: `postgres`,
					Password: `password`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM test ORDER BY pk;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{{1}, {6}, {7}},
				},
				{
					Skip:     true, // CTEs are seen as different tables
					Query:    `WITH cte AS (SELECT * FROM test ORDER BY pk) SELECT * FROM cte;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{{1}, {6}, {7}},
				},
				{
					Query:    `INSERT INTO test VALUES (10);`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `UPDATE test SET pk=pk+20;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DELETE FROM test WHERE pk = 21;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM test ORDER BY pk;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{{26}, {27}, {30}},
				},
				{
					Query:    `REVOKE SELECT, INSERT, UPDATE, DELETE ON test FROM user1;`,
					Username: `postgres`,
					Password: `password`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test ORDER BY pk;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:       `INSERT INTO test VALUES (100);`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:       `UPDATE test SET pk=pk+200;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:       `DELETE FROM test WHERE pk > 3;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
			},
		},
	})
}

func TestDoltStoredProceduresAuth(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Super user access to Dolt procedures",
			SetUpScript: []string{
				"CREATE USER super_user WITH SUPERUSER PASSWORD 'test456';",
				"CREATE TABLE test_table (pk INT PRIMARY KEY);",
				"INSERT INTO test_table VALUES (1);",
				"CALL dolt_add('test_table');",
				"CALL dolt_commit('-m', 'initial commit');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CALL dolt_add('.');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "CALL dolt_backup('add', 'backup1', '" + fileUrl("test_backup") + "');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_branch();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_checkout('main');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "CALL dolt_cherry_pick('HEAD~1', '--allow-empty');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_clean();",
					Skip:  true,
				},
				{
					Query:       "CALL dolt_clone('" + fileUrl("test_repo") + "');",
					Username:    "super_user",
					Password:    "test456",
					ExpectedErr: "The system cannot find the file specified",
				},
				{
					Query:            "CALL dolt_commit('-m', 'test commit', '--allow-empty');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_commit_hash_out();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_conflicts_resolve('--ours', 'test_table');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_count_commits();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_fetch();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true,
				},
				{
					Query:       "CALL dolt_undrop('db1');",
					Username:    "super_user",
					Password:    "test456",
					ExpectedErr: "no database named 'db1' found to undrop", // Auth passed, but logic failed
				},
				{
					Query:            "CALL dolt_update_column_tag();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true,
				},
				{
					Query:            "CALL dolt_purge_dropped_databases();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true,
				},
				{
					Query: "CALL dolt_rebase();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_rm('test_table');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "CALL dolt_gc();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true,
				},
				{
					Query:            "CALL dolt_thread_dump();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true,
				},
				{
					Query:            "CALL dolt_merge('main');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "CALL dolt_remote('add', 'test_remote', '" + fileUrl("remote") + "');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "CALL dolt_pull();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true,
				},
				{
					Query:            "CALL dolt_push();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true,
				},
				{
					Query: "CALL dolt_reset();",
					Skip:  true,
				},
				{
					Query:       "CALL dolt_revert('HEAD');",
					Username:    "super_user",
					Password:    "test456",
					ExpectedErr: "You must commit any changes",
				},
				{
					Query: "CALL dolt_stash();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_tag('v1.0', '-m', 'tag message');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_verify_constraints();",
					Skip:  true,
				},
				{Query: "CALL dolt_stats_restart();", Skip: true},
				{Query: "CALL dolt_stats_stop();", Skip: true},
				{Query: "CALL dolt_stats_info();", Skip: true},
				{Query: "CALL dolt_stats_purge();", Skip: true},
				{Query: "CALL dolt_stats_wait();", Skip: true},
				{Query: "CALL dolt_stats_flush();", Skip: true},
				{Query: "CALL dolt_stats_once();", Skip: true},
				{Query: "CALL dolt_stats_gc();", Skip: true},
				{Query: "CALL dolt_stats_timers();", Skip: true},
			},
		},
		{
			Name: "Regular user access to Dolt procedures",
			SetUpScript: []string{
				"CREATE USER regular_user PASSWORD 'test123';",
				"CREATE TABLE test_table (pk INT PRIMARY KEY);",
				"INSERT INTO test_table VALUES (1);",
				"CALL dolt_add('test_table');",
				"CALL dolt_commit('-m', 'initial commit');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CALL dolt_add('.');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:       "CALL dolt_backup('add', 'backup1', '" + fileUrl("test_backup_reg") + "');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_backup",
				},
				{
					Query: "CALL dolt_branch();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_checkout('main');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:            "CALL dolt_cherry_pick('HEAD~1', '--allow-empty');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_clean();",
					Skip:  true,
				},
				{
					Query:       "CALL dolt_clone('" + fileUrl("test_repo_reg") + "');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_clone",
				},
				{
					Query:            "CALL dolt_commit('-m', 'test commit', '--allow-empty');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_commit_hash_out();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_conflicts_resolve('--ours', 'test_table');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_count_commits();",
					Skip:  true,
				},
				{
					Query:       "CALL dolt_fetch();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_fetch",
					Skip:        true,
				},
				{
					Query:       "CALL dolt_undrop('db1');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_undrop",
				},
				{
					Query:       "CALL dolt_update_column_tag();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_update_column_tag",
					Skip:        true,
				},
				{
					Query:       "CALL dolt_purge_dropped_databases();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_purge_dropped_databases",
					Skip:        true,
				},
				{
					Query: "CALL dolt_rebase();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_rm('test_table');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:       "CALL dolt_gc();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_gc",
					Skip:        true,
				},
				{
					Query:       "CALL dolt_thread_dump();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_thread_dump",
					Skip:        true,
				},
				{
					Query:            "CALL dolt_merge('main');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:       "CALL dolt_remote('add', 'test_remote', '" + fileUrl("remote_reg") + "');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_remote",
				},
				{
					Query:       "CALL dolt_pull();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_pull",
					Skip:        true,
				},
				{
					Query:       "CALL dolt_push();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_push",
					Skip:        true,
				},
				{
					Query: "CALL dolt_reset();",
					Skip:  true,
				},
				{
					Query:       "CALL dolt_revert('HEAD');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "You must commit any changes",
				},
				{
					Query: "CALL dolt_stash();",
					Skip:  true,
				},
				{
					Query:            "CALL dolt_tag('v1.0', '-m', 'tag message');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query: "CALL dolt_verify_constraints();",
					Skip:  true,
				},
				{Query: "CALL dolt_stats_restart();", Skip: true},
				{Query: "CALL dolt_stats_stop();", Skip: true},
				{Query: "CALL dolt_stats_info();", Skip: true},
				{Query: "CALL dolt_stats_purge();", Skip: true},
				{Query: "CALL dolt_stats_wait();", Skip: true},
				{Query: "CALL dolt_stats_flush();", Skip: true},
				{Query: "CALL dolt_stats_once();", Skip: true},
				{Query: "CALL dolt_stats_gc();", Skip: true},
				{Query: "CALL dolt_stats_timers();", Skip: true},
			},
		},
		{
			Name: "Super user access to Dolt procedures (SELECT)",
			SetUpScript: []string{
				"CREATE USER super_user WITH SUPERUSER PASSWORD 'test456';",
				"CREATE TABLE test_table (pk INT PRIMARY KEY);",
				"INSERT INTO test_table VALUES (1);",
				"SELECT dolt_add('test_table');",
				"SELECT dolt_commit('-m', 'initial commit');",
				// Create a backup to generate the remote repo
				"SELECT dolt_backup('add', 'temp_backup', '" + fileUrl("origin_repo") + "');",
				"SELECT dolt_backup('remove', 'temp_backup')",
				//"SELECT dolt_backup('sync', 'temp_backup');",
				//"SELECT dolt_backup('remove', 'temp_backup');", // Remove backup to avoid conflict
				// Add the generated repo as a remote
				"SELECT dolt_remote('add', 'origin', '" + fileUrl("origin_repo") + "');",
				// Setup for undrop
				"CREATE DATABASE db_to_drop;",
				"DROP DATABASE db_to_drop;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT dolt_add('.');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_backup('add', 'backup1_sel', '" + fileUrl("test_backup_sel") + "');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_branch();",
					ExpectedErr: "invalid usage",
				},
				{
					Query:            "SELECT dolt_checkout('main');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_cherry_pick('HEAD~1', '--allow-empty');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_clean();",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_clone('" + fileUrl("origin_repo") + "', 'cloned_db');",
					Username:    "super_user",
					Password:    "test456",
					ExpectedErr: "failed to init repo",
				},
				{
					Query:            "SELECT dolt_commit('-m', 'test commit', '--allow-empty');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_commit_hash_out();",
					ExpectedErr: "Call with too few input arguments",
				},
				{
					Query:            "SELECT dolt_conflicts_resolve('--ours', 'test_table');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_count_commits();",
					ExpectedErr: "missing from ref",
				},
				{
					Query:            "SELECT dolt_fetch('origin');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true, // memory filesystem incompatibility on Windows
				},
				{
					Query:            "SELECT dolt_undrop('db_to_drop');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_update_column_tag();",
					Username:    "super_user",
					Password:    "test456",
					ExpectedErr: "incorrect number of arguments",
				},
				{
					Query:       "SELECT dolt_purge_dropped_databases();",
					Username:    "super_user",
					Password:    "test456",
					ExpectedErr: "unable to check user privileges",
				},
				{
					Query:       "SELECT dolt_rebase();",
					ExpectedErr: "not enough args",
				},
				{
					Query:            "SELECT dolt_rm('test_table');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_thread_dump();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_merge('main');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_remote('add', 'test_remote', '" + fileUrl("remote_sel") + "');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_pull('origin', 'main');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
					Skip:             true,
				},
				{
					Query:       "SELECT dolt_push('origin', 'main');",
					Username:    "super_user",
					Password:    "test456",
					ExpectedErr: "unknown push error",
				},
				{
					Query:            "SELECT dolt_reset();",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_revert('HEAD');",
					Username:    "super_user",
					Password:    "test456",
					ExpectedErr: "You must commit any changes",
				},
				{
					Query:       "SELECT dolt_stash();",
					ExpectedErr: "invalid arguments",
				},
				{
					Query:            "SELECT dolt_tag('v1.0', '-m', 'tag message');",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_verify_constraints();",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_stats_restart();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_stop();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_info();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_purge();",
					ExpectedErr: "stats not persisted, cannot purge",
				},
				{
					Query:       "SELECT dolt_stats_wait();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_flush();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_once();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_gc();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_timers();",
					ExpectedErr: "expected timer arguments",
				},
				{
					Query:            "SELECT dolt_gc();",
					Username:         "super_user",
					Password:         "test456",
					SkipResultsCheck: true,
				},
			},
		},
		{
			Name: "Regular user access to Dolt procedures (SELECT)",
			SetUpScript: []string{
				"CREATE USER regular_user PASSWORD 'test123';",
				"CREATE TABLE test_table (pk INT PRIMARY KEY);",
				"INSERT INTO test_table VALUES (1);",
				"SELECT dolt_add('test_table');",
				"SELECT dolt_commit('-m', 'initial commit');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT dolt_add('.');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_backup('add', 'backup1_sel_reg', '" + fileUrl("test_backup_sel_reg") + "');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_backup",
				},
				{
					Query:       "SELECT dolt_branch();",
					ExpectedErr: "invalid usage",
				},
				{
					Query:            "SELECT dolt_checkout('main');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_cherry_pick('HEAD~1', '--allow-empty');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_clean();",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_clone('" + fileUrl("test_repo_sel_reg") + "');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_clone",
				},
				{
					Query:            "SELECT dolt_commit('-m', 'test commit', '--allow-empty');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_commit_hash_out();",
					ExpectedErr: "Call with too few input arguments",
				},
				{
					Query:            "SELECT dolt_conflicts_resolve('--ours', 'test_table');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_count_commits();",
					ExpectedErr: "missing from ref",
				},
				{
					Query:       "SELECT dolt_fetch();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_fetch",
				},
				{
					Query:       "SELECT dolt_undrop('db1');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_undrop",
				},
				{
					Query:       "SELECT dolt_update_column_tag();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_update_column_tag",
				},
				{
					Query:       "SELECT dolt_purge_dropped_databases();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_purge_dropped_databases",
				},
				{
					Query:       "SELECT dolt_rebase();",
					ExpectedErr: "not enough args",
				},
				{
					Query:            "SELECT dolt_rm('test_table');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_gc();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_gc",
				},
				{
					Query:       "SELECT dolt_thread_dump();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_thread_dump",
				},
				{
					Query:            "SELECT dolt_merge('main');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_remote('add', 'test_remote', '" + fileUrl("remote_sel_reg") + "');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_remote",
				},
				{
					Query:       "SELECT dolt_pull();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_pull",
				},
				{
					Query:       "SELECT dolt_push();",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "permission denied for procedure dolt_push",
				},
				{
					Query:            "SELECT dolt_reset();",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_revert('HEAD');",
					Username:    "regular_user",
					Password:    "test123",
					ExpectedErr: "You must commit any changes",
				},
				{
					Query:       "SELECT dolt_stash();",
					ExpectedErr: "invalid arguments",
				},
				{
					Query:            "SELECT dolt_tag('v1.0', '-m', 'tag message');",
					Username:         "regular_user",
					Password:         "test123",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT dolt_verify_constraints();",
					SkipResultsCheck: true,
				},
				{
					Query:       "SELECT dolt_stats_restart();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_stop();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_info();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_purge();",
					ExpectedErr: "stats not persisted, cannot purge",
				},
				{
					Query:       "SELECT dolt_stats_wait();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_flush();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_once();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_gc();",
					ExpectedErr: "provider does not implement ExtendedStatsProvider",
				},
				{
					Query:       "SELECT dolt_stats_timers();",
					ExpectedErr: "expected timer arguments",
				},
			},
		},
	})
}
