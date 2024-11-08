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
