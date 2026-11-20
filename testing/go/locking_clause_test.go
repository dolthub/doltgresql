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
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/servercfg"
)

func TestLockingClauses(t *testing.T) {
	permissiveConfig, err := servercfg.ConfigFromYamlData([]byte("behavior:\n  permit_unsupported_locking_statements: true\n"))
	require.NoError(t, err)

	lockingAssertions := []ScriptTestAssertion{
		{
			Query:    "SELECT 1 FOR UPDATE",
			Expected: []sql.Row{{1}},
		},
		{
			Query:    "SELECT * FROM locking_test ORDER BY pk FOR UPDATE",
			Expected: []sql.Row{{1}, {2}},
		},
		{
			Query:    "SELECT * FROM locking_test ORDER BY pk FOR NO KEY UPDATE NOWAIT",
			Expected: []sql.Row{{1}, {2}},
		},
		{
			Query:    "SELECT * FROM locking_test ORDER BY pk FOR SHARE SKIP LOCKED",
			Expected: []sql.Row{{1}, {2}},
		},
		{
			Query:    "SELECT * FROM locking_test ORDER BY pk FOR KEY SHARE OF locking_test",
			Expected: []sql.Row{{1}, {2}},
		},
	}
	setup := []string{
		"CREATE TABLE locking_test (pk INT PRIMARY KEY)",
		"INSERT INTO locking_test VALUES (1), (2)",
	}

	RunScripts(t, []ScriptTest{
		{
			Name:        "locking clauses are accepted",
			SetUpScript: setup,
			Assertions:  lockingAssertions,
		},
		{
			// The setting dates from when locking clauses had no effect, and could only be refused or ignored.
			Name:         "locking clauses are accepted with permit_unsupported_locking_statements",
			ServerConfig: permissiveConfig,
			SetUpScript:  setup,
			Assertions:   lockingAssertions,
		},
	})
}
