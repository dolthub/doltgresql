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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions"
)

const (
	authTestSuperUser = "auth_test_super"
	authTestSuperPass = "auth_test_spass"
	authTestBasicUser = "auth_test_basic"
	authTestBasicPass = "auth_test_bpass"
)

var (
	createSuperUser = fmt.Sprintf("create user if not exists '%s' with superuser password '%s';", authTestSuperUser, authTestSuperPass)
	createBasicUser = fmt.Sprintf("create user if not exists '%s' with password '%s'", authTestBasicUser, authTestBasicPass)
)

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

// TestAuthDoltProcedures tests that Dolt procedure functions apply permission checks for SUPERUSERs and basic users in
// SELECT statements. We test both CALL and SELECT to avoid regressions in [node.Call], where previous Doltgres'
// versions fell back to the node runner (on CALL an error is returned to use SELECT now). Some procedures also use [os]
// package calls like [os.TempDir] which can crash [filesys.InMemFS], so we use the local file system.
//
// Each time a new Dolt procedure is introduced in a ScriptTest, it's grouped into a set of related procedures. Each set
// is separated by a new line.
func TestAuthDoltProcedures(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			UseLocalFileSystem: true,
			Name:               "SUPERUSER authorization for CALL executing Dolt stored procedures",
			SetUpScript: []string{
				createSuperUser,
				"create table test_table (v int);",
				"insert into test_table values (1);",
				"select dolt_add('test_table');",
				"select dolt_commit('-m', 'add test table');",
			},
			Assertions: []ScriptTestAssertion{
				assertAsSuper(fmt.Sprintf("call dolt_backup('sync-url', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper(fmt.Sprintf("call dolt_backup('add', 'bak1', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_checkout('-b', 'test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_branch('new_branch');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("insert into test_table values (2);", []sql.Row{}, ""),
				assertAsSuper("call dolt_add('.');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_commit('-m', 'amend test table');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_checkout('main');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_cherry_pick('test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_clean('--dry-run');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper(fmt.Sprintf("call dolt_clone('%s', 'cloned_bak1');", fileUrl("bak1")), nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("set authtest.hash = ''", []sql.Row{}, ""),
				assertAsSuper("call dolt_commit_hash_out('authtest.hash', '-am', 'add val 3 to test table')", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_checkout('-b', 'conflict');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("update test_table set v = -1 where v = 1;", []sql.Row{}, ""),
				assertAsSuper("call dolt_commit('-am', 'amend 1 to -1');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_checkout('main');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("update test_table set v = -2 where v = 1;", []sql.Row{}, ""),
				assertAsSuper("call dolt_commit('-am', 'amend 2 to -2');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("set dolt_allow_commit_conflicts to 1;", []sql.Row{}, ""),
				assertAsSuper("call dolt_merge('conflict');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_conflicts_resolve('--theirs', 'test_table');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_count_commits('--from=main', '--to=test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_backup('remove', 'bak1');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper(fmt.Sprintf("call dolt_remote('add', 'origin', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_fetch('origin', 'main');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_undrop('cloned_bak1');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_commit('-am', 'resolve conflicts');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_update_column_tag('test_table', 'v', '123');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				// TODO(elianddb): "procedure aggregation is not yet supported" error blocks no-parameter CALLs
				skipAssertAsSuper("call dolt_purge_dropped_databases();", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_checkout('test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_rebase('-i', 'main');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_rebase('--abort');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("create table to_rm (v int);", []sql.Row{}, ""),
				assertAsSuper("call dolt_add('to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_commit('-m', 'clean state to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_rm('to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_gc('--shallow');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				// TODO(elianddb): "procedure aggregation is not yet supported" error blocks no-parameter CALLs
				skipAssertAsSuper("call dolt_thread_dump();", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_commit('-m', 'rm to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_push('origin', 'test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_pull('origin', 'test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_reset('--soft', 'HEAD~1');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_stash('push', 'to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_tag('-m', 'dolt_rm procedure', 'to_rm', 'HEAD');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsSuper("call dolt_verify_constraints('--all');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsSuper("call dolt_stats_info('--short');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				// TODO(elianddb): "procedure aggregation is not yet supported" error blocks no-parameter CALLs
				skipAssertAsSuper("call dolt_stats_wait();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsSuper("call dolt_stats_flush();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsSuper("call dolt_stats_gc();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsSuper("call dolt_stats_purge();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsSuper("call dolt_stats_restart();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsSuper("call dolt_stats_once();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
			},
		},
		{
			UseLocalFileSystem: true,
			Name:               "Basic user authentication for CALL executing Dolt stored procedures",
			SetUpScript: []string{
				createBasicUser,
				fmt.Sprintf("alter user %s createdb;", authTestBasicUser),
				createSuperUser,
				"create table test_table (v int);",
				"insert into test_table values (1);",
				"select dolt_add('test_table');",
				"select dolt_commit('-m', 'add test table');",
			},
			Assertions: []ScriptTestAssertion{
				assertAsBasic(fmt.Sprintf("call dolt_backup('sync-url', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic(fmt.Sprintf("call dolt_backup('add', 'bak1', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedureSelectOnly.Error()),

				// Grant user access to test_table before checkout to avoid merge conflict in later cherry-pick.
				grantBasic("schema public", "all"),
				grantBasic("test_table", "select", "insert", "delete", "update"),
				assertAsBasic("call dolt_checkout('-b', 'test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_branch('new_branch');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("insert into test_table values (2);", []sql.Row{}, ""),
				assertAsBasic("call dolt_add('.');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_commit('-m', 'amend test table');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_checkout('main');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_cherry_pick('test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_clean('--dry-run');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic(fmt.Sprintf("call dolt_clone('%s', 'cloned_bak1');", fileUrl("bak1")), nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("create database cloned_bak1;", []sql.Row{}, ""),

				assertAsSuper("set authtest.hash = '';", []sql.Row{}, ""),
				assertAsSuper("call dolt_commit_hash_out('authtest.hash', '-am', 'add val 3 to test table');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_checkout('-b', 'conflict');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("update test_table set v = -1 where v = 1;", []sql.Row{}, ""),
				assertAsBasic("call dolt_commit('-am', 'amend 1 to -1');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_checkout('main');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("update test_table set v = -2 where v = 1;", []sql.Row{}, ""),
				assertAsBasic("call dolt_commit('-am', 'amend 2 to -2');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("set dolt_allow_commit_conflicts to 1;", []sql.Row{}, ""),
				assertAsBasic("call dolt_merge('conflict');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_conflicts_resolve('--theirs', 'test_table');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_count_commits('--from=main', '--to=test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_backup('remove', 'bak1');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic(fmt.Sprintf("call dolt_remote('add', 'origin', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_fetch('origin', 'main');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_undrop('cloned_bak1');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_commit('-am', 'resolve conflicts');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_update_column_tag('test_table', 'v', '123');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("drop database cloned_bak1;", []sql.Row{}, ""),
				// TODO(elianddb): "procedure aggregation is not yet supported" error blocks no-parameter CALLs
				skipAssertAsBasic("call dolt_purge_dropped_databases();", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_checkout('test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_rebase('-i', 'main');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_rebase('--abort');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("create table to_rm (v int);", []sql.Row{}, ""),
				assertAsBasic("call dolt_add('to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_commit('-m', 'clean state to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_rm('to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_gc('--shallow');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				// TODO(elianddb): "procedure aggregation is not yet supported" error blocks no-parameter CALLs
				skipAssertAsBasic("call dolt_thread_dump();", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_commit('-m', 'rm to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_push('origin', 'test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_pull('origin', 'test');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_reset('--soft', 'HEAD~1');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_stash('push', 'to_rm');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_tag('-m', 'dolt_rm procedure', 'to_rm', 'HEAD');", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				assertAsBasic("call dolt_verify_constraints('--all');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				assertAsBasic("call dolt_stats_info('--short');", nil, functions.ErrDoltProcedureSelectOnly.Error()),

				// TODO(elianddb): "procedure aggregation is not yet supported" error blocks no-parameter CALLs
				skipAssertAsBasic("call dolt_stats_wait();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsBasic("call dolt_stats_flush();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsBasic("call dolt_stats_gc();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsBasic("call dolt_stats_purge();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsBasic("call dolt_stats_restart();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
				skipAssertAsBasic("call dolt_stats_once();", nil, functions.ErrDoltProcedureSelectOnly.Error()),
			},
		},
		{
			UseLocalFileSystem: true,
			Name:               "SUPERUSER authorization for SELECT executing Dolt stored procedures",
			SetUpScript: []string{
				createSuperUser,
				"create table test_table (v int);",
				"insert into test_table values (1);",
				"select dolt_add('test_table');",
				"select dolt_commit('-m', 'add test table');",
			},
			Assertions: []ScriptTestAssertion{
				assertAsSuper(fmt.Sprintf("select dolt_backup('sync-url', '%s');", fileUrl("bak1")), []sql.Row{{"{0}"}}, ""),
				assertAsSuper(fmt.Sprintf("select dolt_backup('add', 'bak1', '%s');", fileUrl("bak1")), []sql.Row{{"{0}"}}, ""),

				assertAsSuper("select dolt_checkout('-b', 'test');", []sql.Row{{"{0,\"Switched to branch 'test'\"}"}}, ""),

				assertAsSuper("select dolt_branch('new_branch');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("insert into test_table values (2);", []sql.Row{}, ""),
				assertAsSuper("select dolt_add('.');", []sql.Row{{"{0}"}}, ""),
				assertAsSuper("select length(dolt_commit('-m', 'amend test table')::text) = 34;", []sql.Row{{"t"}}, ""),

				assertAsSuper("select dolt_checkout('main');", []sql.Row{{"{0,\"Switched to branch 'main'\"}"}}, ""),
				assertAsSuper("select length(dolt_cherry_pick('test')::text);", []sql.Row{{40}}, ""),

				assertAsSuper("select dolt_clean('--dry-run');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper(fmt.Sprintf("select dolt_clone('%s', 'cloned_bak1');", fileUrl("bak1")), []sql.Row{{"{0}"}}, ""),

				assertAsSuper("set authtest.hash = '';", []sql.Row{}, ""),
				// TODO(elianddb): variadic parameter support for Dolt stored procedures functions
				skipAssertAsSuper("select dolt_commit_hash_out('authtest.hash', '-am', 'add val 3 to test table');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("select dolt_checkout('-b', 'conflict');", []sql.Row{{"{0,\"Switched to branch 'conflict'\"}"}}, ""),
				assertAsSuper("update test_table set v = -1 where v = 1;", []sql.Row{}, ""),
				assertAsSuper("select length(dolt_commit('-am', 'amend 1 to -1')::text) = 34;", []sql.Row{{"t"}}, ""),
				assertAsSuper("select dolt_checkout('main');", []sql.Row{{"{0,\"Switched to branch 'main'\"}"}}, ""),
				assertAsSuper("update test_table set v = -2 where v = 1;", []sql.Row{}, ""),
				assertAsSuper("select length(dolt_commit('-am', 'amend 2 to -2')::text) = 34;", []sql.Row{{"t"}}, ""),
				assertAsSuper("set dolt_allow_commit_conflicts to 1;", []sql.Row{}, ""),
				assertAsSuper("select dolt_merge('conflict');", []sql.Row{{"{0,1,\"conflicts found\"}"}}, ""),

				assertAsSuper("select dolt_conflicts_resolve('--theirs', 'test_table');", []sql.Row{{"{0}"}}, ""),

				// TODO(elianddb): unsupported type uint64
				skipAssertAsSuper("select dolt_count_commits('--from=main', '--to=test');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("select dolt_backup('remove', 'bak1');", []sql.Row{{"{0}"}}, ""),
				assertAsSuper(fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", fileUrl("bak1")), []sql.Row{{"{0}"}}, ""),

				assertAsSuper("select dolt_fetch('origin', 'main');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("drop database cloned_bak1", []sql.Row{}, ""),
				assertAsSuper("select dolt_undrop('cloned_bak1');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("select length(dolt_commit('-am', 'resolve conflicts')::text) = 34;", []sql.Row{{"t"}}, ""),
				// TODO(elianddb): table test_table does not exist (also tried with public.test_table)
				skipAssertAsSuper("select dolt_update_column_tag('test_table', 'v', '123');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("drop database cloned_bak1", []sql.Row{}, ""),
				assertAsSuper("select dolt_purge_dropped_databases();", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("select dolt_checkout('test');", []sql.Row{{"{0,\"Switched to branch 'test'\"}"}}, ""),
				assertAsSuper(
					"select dolt_rebase('-i', 'main');",
					[]sql.Row{{"{0,\"interactive rebase started on branch dolt_rebase_test; adjust the rebase plan in the dolt_rebase table, then continue rebasing by calling dolt_rebase('--continue')\"}"}},
					""),
				assertAsSuper("select dolt_rebase('--abort');", []sql.Row{{"{0,\"Interactive rebase aborted\"}"}}, ""),

				assertAsSuper("create table to_rm (v int);", []sql.Row{}, ""),
				assertAsSuper("select dolt_add('to_rm');", []sql.Row{{"{0}"}}, ""),
				assertAsSuper("select length(dolt_commit('-m', 'clean state to_rm')::text) = 34;", []sql.Row{{"t"}}, ""),
				assertAsSuper("select dolt_rm('to_rm');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("select dolt_gc('--shallow');", []sql.Row{{"{0}"}}, ""),

				// The paths for files, memory addresses, and number of goroutines can be different per OS.
				assertAsSuper("select instr(dolt_thread_dump()::text, 'goroutine') > 0;", []sql.Row{{"t"}}, ""),

				assertAsSuper("select length(dolt_commit('-m', 'rm to_rm')::text) = 34;", []sql.Row{{"t"}}, ""),
				assertAsSuper(
					"select dolt_push('origin', 'test');",
					[]sql.Row{{fmt.Sprintf("{0,\"To %s\n * [new branch]          test -> test\"}", fileUrl("bak1"))}},
					""),
				assertAsSuper("select dolt_pull('origin', 'test');", []sql.Row{{"{0,0,\"Everything up-to-date\"}"}}, ""),

				assertAsSuper("select dolt_reset('--soft', 'HEAD~1');", []sql.Row{{"{0}"}}, ""),
				// TODO(elianddb): unsupported type int
				skipAssertAsSuper("select dolt_stash('push', 'to_rm');", []sql.Row{{"{0}"}}, ""),

				assertAsSuper("select dolt_tag('-m', 'dolt_rm procedure', 'to_rm', 'HEAD');", []sql.Row{{"{0}"}}, ""),
				assertAsSuper("select dolt_verify_constraints('--all');", []sql.Row{{"{0}"}}, ""),

				// TODO(elianddb): provider does not implement ExtendedStatsProvider
				skipAssertAsSuper("select dolt_stats_info('--short');", []sql.Row{{"{0}"}}, ""),
				skipAssertAsSuper("select dolt_stats_wait();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsSuper("select dolt_stats_flush();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsSuper("select dolt_stats_gc();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsSuper("select dolt_stats_purge();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsSuper("select dolt_stats_restart();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsSuper("select dolt_stats_once();", []sql.Row{{"{0}"}}, ""),
			},
		},
		{
			UseLocalFileSystem: true,
			Name:               "Basic user authorization for SELECT executing Dolt stored procedures",
			SetUpScript: []string{
				createBasicUser,
				fmt.Sprintf("alter user %s createdb;", authTestBasicUser),
				createSuperUser,
				"create table test_table (v int);",
				"insert into test_table values (1);",
				"select dolt_add('test_table');",
				"select dolt_commit('-m', 'add test table');",
			},
			Assertions: []ScriptTestAssertion{
				assertAsBasic(fmt.Sprintf("select dolt_backup('sync-url', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedurePermissionDenied.Error()),
				assertAsBasic(fmt.Sprintf("select dolt_backup('add', 'bak1', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedurePermissionDenied.Error()),

				// Grant user access to test_table before checkout to avoid merge conflict in later cherry-pick.
				grantBasic("schema public", "all"),
				grantBasic("test_table", "select", "insert", "delete", "update"),
				assertAsBasic("select dolt_checkout('-b', 'test');", []sql.Row{{"{0,\"Switched to branch 'test'\"}"}}, ""),

				assertAsBasic("select dolt_branch('new_branch');", []sql.Row{{"{0}"}}, ""),

				assertAsBasic("insert into test_table values (2);", []sql.Row{}, ""),
				assertAsBasic("select dolt_add('.');", []sql.Row{{"{0}"}}, ""),
				assertAsBasic("select length(dolt_commit('-m', 'amend test table')::text) = 34;", []sql.Row{{"t"}}, ""),

				assertAsBasic("select dolt_checkout('main');", []sql.Row{{"{0,\"Switched to branch 'main'\"}"}}, ""),
				assertAsBasic("select length(dolt_cherry_pick('test')::text);", []sql.Row{{40}}, ""),

				assertAsBasic("select dolt_clean('--dry-run');", []sql.Row{{"{0}"}}, ""),

				assertAsBasic(fmt.Sprintf("select dolt_clone('%s', 'cloned_bak1');", fileUrl("bak1")), nil, functions.ErrDoltProcedurePermissionDenied.Error()),
				assertAsBasic("create database cloned_bak1;", []sql.Row{}, ""),

				assertAsBasic("set authtest.hash = '';", []sql.Row{}, ""),
				// TODO(elianddb): variadic parameter support for Dolt stored procedures
				skipAssertAsBasic("select dolt_commit_hash_out('authtest.hash', '-am', 'add val 3 to test table');", []sql.Row{{"{0}"}}, ""),

				assertAsBasic("select dolt_checkout('-b', 'conflict');", []sql.Row{{"{0,\"Switched to branch 'conflict'\"}"}}, ""),
				assertAsBasic("update test_table set v = -1 where v = 1;", []sql.Row{}, ""),
				assertAsBasic("select length(dolt_commit('-am', 'amend 1 to -1')::text) = 34;", []sql.Row{{"t"}}, ""),
				assertAsBasic("select dolt_checkout('main');", []sql.Row{{"{0,\"Switched to branch 'main'\"}"}}, ""),
				assertAsBasic("update test_table set v = -2 where v = 1;", []sql.Row{}, ""),
				assertAsBasic("select length(dolt_commit('-am', 'amend 2 to -2')::text) = 34;", []sql.Row{{"t"}}, ""),
				assertAsBasic("set dolt_allow_commit_conflicts to 1;", []sql.Row{}, ""),
				assertAsBasic("select dolt_merge('conflict');", []sql.Row{{"{0,1,\"conflicts found\"}"}}, ""),

				assertAsBasic("select dolt_conflicts_resolve('--theirs', 'test_table');", []sql.Row{{"{0}"}}, ""),

				// TODO(elianddb): unsupported type uint64
				skipAssertAsBasic("select dolt_count_commits('--from=main', '--to=test');", []sql.Row{{"{0}"}}, ""),

				assertAsBasic("select dolt_backup('remove', 'bak1');", nil, functions.ErrDoltProcedurePermissionDenied.Error()),
				assertAsBasic(fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", fileUrl("bak1")), nil, functions.ErrDoltProcedurePermissionDenied.Error()),

				assertAsBasic("select dolt_fetch('origin', 'main');", nil, functions.ErrDoltProcedurePermissionDenied.Error()),

				assertAsBasic("drop database cloned_bak1", []sql.Row{}, ""),
				assertAsBasic("select dolt_undrop('cloned_bak1');", nil, functions.ErrDoltProcedurePermissionDenied.Error()),

				assertAsBasic("select length(dolt_commit('-am', 'resolve conflicts')::text) = 34;", []sql.Row{{"t"}}, ""),
				// TODO(elianddb): table test_table does not exist (also tried with public.test_table)
				skipAssertAsBasic("select dolt_update_column_tag('test_table', 'v', '123');", []sql.Row{{"{0}"}}, ""),

				assertAsBasic("select dolt_purge_dropped_databases();", nil, functions.ErrDoltProcedurePermissionDenied.Error()),

				assertAsBasic("select dolt_checkout('test');", []sql.Row{{"{0,\"Switched to branch 'test'\"}"}}, ""),
				assertAsBasic(
					"select dolt_rebase('-i', 'main');",
					[]sql.Row{{"{0,\"interactive rebase started on branch dolt_rebase_test; adjust the rebase plan in the dolt_rebase table, then continue rebasing by calling dolt_rebase('--continue')\"}"}},
					""),
				assertAsBasic("select dolt_rebase('--abort');", []sql.Row{{"{0,\"Interactive rebase aborted\"}"}}, ""),

				assertAsBasic("create table to_rm (v int);", []sql.Row{}, ""),
				assertAsBasic("select dolt_add('to_rm');", []sql.Row{{"{0}"}}, ""),
				assertAsBasic("select length(dolt_commit('-m', 'clean state to_rm')::text) = 34;", []sql.Row{{"t"}}, ""),
				assertAsBasic("select dolt_rm('to_rm');", []sql.Row{{"{0}"}}, ""),

				assertAsBasic("select dolt_gc('--shallow');", nil, functions.ErrDoltProcedurePermissionDenied.Error()),

				assertAsBasic("select dolt_thread_dump();", nil, functions.ErrDoltProcedurePermissionDenied.Error()),

				assertAsBasic("select length(dolt_commit('-m', 'rm to_rm')::text) = 34;", []sql.Row{{"t"}}, ""),
				assertAsBasic("select dolt_push('origin', 'test');", nil, functions.ErrDoltProcedurePermissionDenied.Error()),
				assertAsBasic("select dolt_pull('origin', 'test');", nil, functions.ErrDoltProcedurePermissionDenied.Error()),

				assertAsBasic("select dolt_reset('--soft', 'HEAD~1');", []sql.Row{{"{0}"}}, ""),
				// TODO(elianddb): unsupported type int
				skipAssertAsBasic("select dolt_stash('push', 'to_rm');", []sql.Row{{"{0}"}}, ""),

				assertAsBasic("select dolt_tag('-m', 'dolt_rm procedure', 'to_rm', 'HEAD');", []sql.Row{{"{0}"}}, ""),
				assertAsBasic("select dolt_verify_constraints('--all');", []sql.Row{{"{0}"}}, ""),

				// TODO(elianddb): provider does not implement ExtendedStatsProvider
				skipAssertAsBasic("select dolt_stats_info('--short');", []sql.Row{{"{0}"}}, ""),
				skipAssertAsBasic("select dolt_stats_wait();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsBasic("select dolt_stats_flush();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsBasic("select dolt_stats_gc();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsBasic("select dolt_stats_purge();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsBasic("select dolt_stats_restart();", []sql.Row{{"{0}"}}, ""),
				skipAssertAsBasic("select dolt_stats_once();", []sql.Row{{"{0}"}}, ""),
			},
		},
	})
}

// assertAsSuper returns a ScriptTestAssertion for the given |query|, |expectedResultSet|, and/or |expectedErr| using
// authTestSuperUser.
func assertAsSuper(query string, expectedResultSet []sql.Row, expectedErr string) ScriptTestAssertion {
	return ScriptTestAssertion{
		Username:    authTestSuperUser,
		Password:    authTestSuperPass,
		Query:       query,
		Expected:    expectedResultSet,
		ExpectedErr: expectedErr,
	}
}

// skipAssertAsSuper skips the returned assertion from assertAsSuper.
func skipAssertAsSuper(query string, expectedResultSet []sql.Row, expectedErr string) ScriptTestAssertion {
	assertion := assertAsSuper(query, expectedResultSet, expectedErr)
	assertion.Skip = true
	return assertion
}

// assertAsBasic returns a ScriptTestAssertion for the given |query|, |expectedResultSet|, and/or |expectedErr| using
// authTestBasicUser.
func assertAsBasic(query string, expected []sql.Row, expectedErr string) ScriptTestAssertion {
	return ScriptTestAssertion{
		Username:    authTestBasicUser,
		Password:    authTestBasicPass,
		Query:       query,
		Expected:    expected,
		ExpectedErr: expectedErr,
	}
}

// skipAssertAsBasic skips the returned assertion from assertAsBasic.
func skipAssertAsBasic(query string, expected []sql.Row, expectedErr string) ScriptTestAssertion {
	assertion := assertAsBasic(query, expected, expectedErr)
	assertion.Skip = true
	return assertion
}

// grantBasic grants |privileges| to authTestBasicUser on given |object| (include the object type in |object| if
// applicable).
func grantBasic(object string, privileges ...string) ScriptTestAssertion {
	return ScriptTestAssertion{
		Username: "postgres",
		Password: "password",
		Query:    fmt.Sprintf("GRANT %s ON %s TO %s", strings.Join(privileges, ","), object, authTestBasicUser),
		Expected: []sql.Row{},
	}
}

// fileUrl returns a file:// URL path for a temp file.
func fileUrl(path string) string {
	path = filepath.Join(os.TempDir(), path)
	return "file://" + filepath.ToSlash(filepath.Clean(path))
}
