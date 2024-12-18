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
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// TestAuthQuick is modeled after the QuickPrivilegeTest in GMS, so please refer to the documentation there:
// https://github.com/dolthub/go-mysql-server/blob/main/enginetest/queries/priv_auth_queries.go
func TestAuthQuick(t *testing.T) {
	// Statements that are run before every test (the state that all tests start with):
	// CREATE USER tester PASSWORD 'password';
	// CREATE SCHEMA mysch;
	// CREATE SCHEMA othersch;
	// CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);
	// CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);
	// CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);
	// CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);
	// INSERT INTO mysch.test VALUES (0, 0), (1, 1);
	// INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);
	// INSERT INTO othersch.test VALUES (1, 1), (2, 2);
	// INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);
	type QuickPrivilegeTest struct {
		Focus       bool
		Queries     []string
		Expected    []sql.Row
		ExpectedErr string
	}
	tests := []QuickPrivilegeTest{
		{
			Queries: []string{
				"GRANT SELECT ON ALL TABLES IN SCHEMA mysch TO tester;",
				"SELECT * FROM mysch.test;",
			},
			Expected: []sql.Row{{0, 0}, {1, 1}},
		},
		{
			Queries: []string{
				"GRANT SELECT ON ALL TABLES IN SCHEMA mysch TO tester;",
				"SELECT * FROM mysch.test2;",
			},
			Expected: []sql.Row{{0, 1}, {1, 2}},
		},
		{
			Queries: []string{
				"GRANT SELECT ON mysch.test TO tester;",
				"SELECT * FROM mysch.test;",
			},
			Expected: []sql.Row{{0, 0}, {1, 1}},
		},
		{
			Queries: []string{
				"GRANT SELECT ON mysch.test TO tester;",
				"SELECT * FROM mysch.test2;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT SELECT ON ALL TABLES IN SCHEMA othersch TO tester;",
				"SELECT * FROM mysch.test;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT SELECT ON othersch.test TO tester;",
				"SELECT * FROM mysch.test;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT SELECT ON othersch.test TO tester;",
				"SELECT * FROM mysch.test;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"CREATE SCHEMA newsch;",
			},
			ExpectedErr: "permission denied for database",
		},
		{
			Queries: []string{
				"GRANT CREATE ON DATABASE postgres TO tester;",
				"CREATE SCHEMA newsch;",
			},
		},
		{ // This isn't supported yet, but it is supposed to fail since tester is not an owner
			Queries: []string{
				"GRANT CREATE ON DATABASE postgres TO tester;",
				"CREATE SCHEMA newsch;",
				"DROP SCHEMA newsch;",
			},
			ExpectedErr: "not yet supported",
		},
		{
			Queries: []string{
				"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
			},
			ExpectedErr: "permission denied for schema",
		},
		{
			Queries: []string{
				"GRANT CREATE ON SCHEMA mysch TO tester;",
				"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
			},
		},
		{
			Queries: []string{
				"CREATE ROLE new_role;",
			},
			ExpectedErr: "does not have permission",
		},
		{
			Queries: []string{
				"ALTER ROLE tester CREATEROLE;",
				"CREATE ROLE new_role;",
			},
		},
		{
			Queries: []string{
				"CREATE USER new_user;",
			},
			ExpectedErr: "does not have permission",
		},
		{
			Queries: []string{
				"ALTER ROLE tester SUPERUSER;",
				"CREATE USER new_user;",
			},
		},
		{
			Queries: []string{
				"CREATE USER new_user;",
				"DROP USER new_user;",
			},
			ExpectedErr: "does not have permission",
		},
		{
			Queries: []string{
				"CREATE USER new_user;",
				"ALTER ROLE tester CREATEROLE;",
				"DROP USER new_user;",
			},
		},
		{
			Queries: []string{
				"CREATE USER new_user SUPERUSER;",
				"ALTER ROLE tester CREATEROLE;",
				"DROP USER new_user;",
			},
			ExpectedErr: "does not have permission",
		},
		{
			Queries: []string{
				"CREATE USER new_user SUPERUSER;",
				"ALTER ROLE tester SUPERUSER;",
				"DROP USER new_user;",
			},
		},
		{
			Queries: []string{
				"DELETE FROM mysch.test WHERE pk >= 0;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT DELETE ON ALL TABLES IN SCHEMA mysch TO tester;",
				"DELETE FROM mysch.test WHERE pk >= 0;",
			},
		},
		{
			Queries: []string{
				"GRANT DELETE ON mysch.test TO tester;",
				"DELETE FROM mysch.test WHERE pk >= 0;",
			},
		},
		{
			Queries: []string{
				"CREATE USER tester2;",
				"GRANT DELETE ON ALL TABLES IN SCHEMA mysch TO tester2;",
				"GRANT tester2 TO tester;",
				"DELETE FROM mysch.test WHERE pk >= 0;",
			},
		},
		{
			Queries: []string{
				"SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT SELECT ON mysch.test TO tester;",
				"SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT SELECT ON mysch.test2 TO tester;",
				"SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT SELECT ON mysch.test TO tester;",
				"GRANT SELECT ON mysch.test2 TO tester;",
				"SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
			},
			Expected: []sql.Row{{0, 0, 0, 1}, {1, 1, 1, 2}},
		},
		{
			Queries: []string{
				"CREATE USER tester2;",
				"GRANT SELECT ON mysch.test2 TO tester2;",
				"GRANT tester2 TO tester;",
				"SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"CREATE USER tester2;",
				"GRANT SELECT ON mysch.test TO tester2;",
				"GRANT SELECT ON mysch.test2 TO tester2;",
				"GRANT tester2 TO tester;",
				"SELECT * FROM mysch.test JOIN mysch.test2 ON test.pk = test2.pk;",
			},
			Expected: []sql.Row{{0, 0, 0, 1}, {1, 1, 1, 2}},
		},
		{
			Queries: []string{
				"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
				"DROP TABLE mysch.new_table;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA mysch TO tester;",
				"REVOKE DROP ON ALL TABLES IN SCHEMA mysch FROM tester;",
				"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
				"DROP TABLE mysch.new_table;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
				"GRANT DROP ON mysch.new_table TO tester;",
				"DROP TABLE mysch.new_table;",
			},
		},
		{
			Queries: []string{
				"CREATE TABLE mysch.new_table (pk BIGINT PRIMARY KEY);",
				"GRANT postgres TO tester;",
				"DROP TABLE mysch.new_table;",
			},
		},
		{
			Queries: []string{
				"CREATE ROLE new_role;",
				"DROP ROLE new_role;",
			},
			ExpectedErr: "does not have permission",
		},
		{
			Queries: []string{
				"ALTER ROLE tester CREATEROLE;",
				"CREATE ROLE new_role;",
				"DROP ROLE new_role;",
			},
		},
		{
			Queries: []string{
				"INSERT INTO mysch.test VALUES (9, 9);",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT INSERT ON ALL TABLES IN SCHEMA mysch TO tester;",
				"INSERT INTO mysch.test VALUES (9, 9);",
			},
		},
		{
			Queries: []string{
				"GRANT INSERT ON mysch.test TO tester;",
				"INSERT INTO mysch.test VALUES (9, 9);",
			},
		},
		{
			Queries: []string{
				"UPDATE mysch.test SET v1 = 0;",
			},
			ExpectedErr: "permission denied for table",
		},
		{
			Queries: []string{
				"GRANT UPDATE ON ALL TABLES IN SCHEMA mysch TO tester;",
				"UPDATE mysch.test SET v1 = 0;",
			},
		},
		{
			Queries: []string{
				"GRANT UPDATE ON mysch.test TO tester;",
				"UPDATE mysch.test SET v1 = 0;",
			},
		},
	}
	// Here we'll convert each quick test into a standard test
	scriptTests := make([]ScriptTest, len(tests))
	for testIdx, test := range tests {
		scriptTests[testIdx] = ScriptTest{
			Name:     strings.Join(test.Queries, "\n > "),
			Database: "",
			SetUpScript: []string{
				"CREATE USER tester PASSWORD 'password';",
				"CREATE SCHEMA mysch;",
				"CREATE SCHEMA othersch;",
				"CREATE TABLE mysch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"CREATE TABLE mysch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"CREATE TABLE othersch.test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"CREATE TABLE othersch.test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO mysch.test VALUES (0, 0), (1, 1);",
				"INSERT INTO mysch.test2 VALUES (0, 1), (1, 2);",
				"INSERT INTO othersch.test VALUES (1, 1), (2, 2);",
				"INSERT INTO othersch.test2 VALUES (1, 1), (2, 2);",
			},
			Assertions: make([]ScriptTestAssertion, len(test.Queries)),
			Focus:      test.Focus,
		}
		for queryIdx := 0; queryIdx < len(test.Queries)-1; queryIdx++ {
			scriptTests[testIdx].Assertions[queryIdx] = ScriptTestAssertion{
				Query:            test.Queries[queryIdx],
				SkipResultsCheck: true,
				Username:         "postgres",
				Password:         "password",
			}
		}
		scriptTests[testIdx].Assertions[len(test.Queries)-1] = ScriptTestAssertion{
			Query:       test.Queries[len(test.Queries)-1],
			Expected:    test.Expected,
			ExpectedErr: test.ExpectedErr,
			Username:    "tester",
			Password:    "password",
		}
	}
	RunScripts(t, scriptTests)
}
