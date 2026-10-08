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

func TestSetLocal(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "SET LOCAL reverts on COMMIT",
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN",
				},
				{
					Query: "SET LOCAL enable_hashjoin = off",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{0}},
				},
				{
					Query: "COMMIT",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "SET LOCAL reverts on ROLLBACK",
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN",
				},
				{
					Query: "SET LOCAL enable_hashjoin = off",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{0}},
				},
				{
					Query: "ROLLBACK",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "SET LOCAL reverts to the session value, not the default",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SET enable_hashjoin = off",
				},
				{
					Query: "BEGIN",
				},
				{
					Query: "SET LOCAL enable_hashjoin = on",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{1}},
				},
				{
					Query: "COMMIT",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{0}},
				},
				{
					Query: "SET enable_hashjoin = on",
				},
			},
		},
		{
			Name: "SET LOCAL reverts when a failed transaction is rolled back",
			SetUpScript: []string{
				`CREATE TABLE test (a INT PRIMARY KEY)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN",
				},
				{
					Query: "SET LOCAL enable_hashjoin = off",
				},
				{
					Query:       "SELECT no_such_column FROM test",
					ExpectedErr: "could not be found",
				},
				{
					Query:       "SHOW enable_hashjoin",
					ExpectedErr: "current transaction is aborted",
				},
				{
					Query: "ROLLBACK",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "SET LOCAL outside a transaction block has no lasting effect",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SET LOCAL enable_hashjoin = off",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "SET LOCAL with savepoints does not abort the transaction",
			SetUpScript: []string{
				`CREATE TABLE test (a INT PRIMARY KEY)`,
				`INSERT INTO test VALUES (1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN",
				},
				{
					Query: "SAVEPOINT settings",
				},
				{
					Query: "SET LOCAL enable_hashjoin = off",
				},
				{
					Query: "SET LOCAL enable_mergejoin = on",
				},
				{
					Query:    "SELECT * FROM test",
					Expected: []sql.Row{{1}},
				},
				{
					Query: "ROLLBACK TO settings",
				},
				{
					Query:    "SELECT * FROM test",
					Expected: []sql.Row{{1}},
				},
				{
					Query: "COMMIT",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "SET after SET LOCAL persists after COMMIT",
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN",
				},
				{
					Query: "SET LOCAL enable_hashjoin = off",
				},
				{
					Query: "SET enable_hashjoin = off",
				},
				{
					Query: "COMMIT",
				},
				{
					Query:    "SHOW enable_hashjoin",
					Expected: []sql.Row{{0}},
				},
				{
					Query: "SET enable_hashjoin = on",
				},
			},
		},
		{
			Name: "SET LOCAL on an unknown parameter errors",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SET LOCAL no_such_parameter = on",
					ExpectedErr: "unrecognized configuration parameter",
				},
			},
		},
		{
			Name: "set_config with is_local reverts on COMMIT",
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN",
				},
				{
					Query:    "SELECT set_config('enable_seqscan', 'off', true)",
					Expected: []sql.Row{{"off"}},
				},
				{
					Query:    "SHOW enable_seqscan",
					Expected: []sql.Row{{0}},
				},
				{
					Query: "COMMIT",
				},
				{
					Query:    "SHOW enable_seqscan",
					Expected: []sql.Row{{1}},
				},
			},
		},
	})
}
