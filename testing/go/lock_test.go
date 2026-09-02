// Copyright 2025 Dolthub, Inc.
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

// TestLocks tests the advisory lock functions, such as pg_try_advisory_lock and pg_advisory_unlock.
func TestAdvisoryLocks(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "basic lock tests",
			SetUpScript: []string{
				`CREATE USER user1 PASSWORD 'password';`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT pg_advisory_lock(1)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_try_advisory_lock(2)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// When a different session tries to acquire the same lock, it fails.
					Username: "user1",
					Password: "password",
					Query:    `SELECT pg_try_advisory_lock(1)`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// When a different session tries to acquire the same lock, it fails.
					Username: "user1",
					Password: "password",
					Query:    `SELECT pg_try_advisory_lock(2)`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT pg_advisory_unlock(1)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_advisory_unlock(2)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_advisory_unlock(3)`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "advisory locks are reentrant",
			SetUpScript: []string{
				`CREATE USER user1 PASSWORD 'password';`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT pg_advisory_lock(10)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// The same session may acquire a lock it already holds.
					Query:    `SELECT pg_advisory_lock(10)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_try_advisory_lock(10)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `SELECT pg_try_advisory_lock(10)`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT pg_advisory_unlock(10)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_advisory_unlock(10)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// The lock was acquired three times, so two unlocks leave it held and
					// other sessions still cannot acquire it.
					Username: "user1",
					Password: "password",
					Query:    `SELECT pg_try_advisory_lock(10)`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT pg_advisory_unlock(10)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// After the final unlock, other sessions may acquire the lock.
					Username: "user1",
					Password: "password",
					Query:    `SELECT pg_try_advisory_lock(10)`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
	})

	RunTransactionTests(t, []ScriptTest{
		{
			Name: "transaction advisory locks",
			SetUpScript: []string{
				`CREATE TABLE lock_commit_test (pk INT PRIMARY KEY)`,
				`SELECT DOLT_BRANCH('lock-test-branch')`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `/* client A */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client A */ SELECT pg_advisory_xact_lock(20)`, Expected: []sql.Row{{nil}}},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(20)`, Expected: []sql.Row{{"f"}}},
				{Query: `/* client A */ COMMIT`, ExpectedTag: "COMMIT"},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(20)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_advisory_unlock(20)`, Expected: []sql.Row{{"t"}}},

				{Query: `/* client A */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client A */ SELECT pg_try_advisory_xact_lock(21)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(21)`, Expected: []sql.Row{{"f"}}},
				{Query: `/* client A */ ROLLBACK`, ExpectedTag: "ROLLBACK"},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(21)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_advisory_unlock(21)`, Expected: []sql.Row{{"t"}}},

				// The blocking form remains in flight until client A releases the lock.
				{Query: `/* client A */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client A */ SELECT pg_advisory_xact_lock(22)`, Expected: []sql.Row{{nil}}},
				{Query: `/* client B */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client B */ SELECT pg_advisory_xact_lock(22)`, ExpectedBlocking: true},
				{Query: `/* client A */ COMMIT`, ExpectedTag: "COMMIT"},
				{Query: `/* client B */ COMMIT`, ExpectedTag: "COMMIT"},
				{Query: `/* client C */ SELECT pg_try_advisory_lock(22)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client C */ SELECT pg_advisory_unlock(22)`, Expected: []sql.Row{{"t"}}},

				// Autocommit also ends the transaction and releases its lock.
				{Query: `/* client A */ SELECT pg_try_advisory_xact_lock(23)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(23)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_advisory_unlock(23)`, Expected: []sql.Row{{"t"}}},

				// DOLT_COMMIT clears the transaction from inside a stored function.
				{Query: `/* client A */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client A */ SELECT pg_advisory_xact_lock(24)`, Expected: []sql.Row{{nil}}},
				{Query: `/* client A */ INSERT INTO lock_commit_test VALUES (1)`, ExpectedTag: "INSERT 0 1"},
				{Query: `/* client A */ SELECT DOLT_COMMIT('-Am', 'lock lifecycle test')`, SkipResultsCheck: true},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(24)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_advisory_unlock(24)`, Expected: []sql.Row{{"t"}}},

				// Clearing branch-dependent caches must not discard lock callbacks.
				{Query: `/* client A */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client A */ SELECT pg_advisory_xact_lock(25)`, Expected: []sql.Row{{nil}}},
				{Query: `/* client A */ SELECT DOLT_CHECKOUT('lock-test-branch')`, SkipResultsCheck: true},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(25)`, Expected: []sql.Row{{"f"}}},
				{Query: `/* client A */ COMMIT`, ExpectedTag: "COMMIT"},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(25)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_advisory_unlock(25)`, Expected: []sql.Row{{"t"}}},
			},
		},
		{
			Name: "transaction locks release across session reuse",
			Assertions: []ScriptTestAssertion{
				{Query: `/* client A */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client A */ SELECT pg_advisory_xact_lock(30)`, Expected: []sql.Row{{nil}}},
				{Query: `/* client A */ COMMIT`, ExpectedTag: "COMMIT"},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(30)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_advisory_unlock(30)`, Expected: []sql.Row{{"t"}}},

				// Reuse A for another transaction and verify rollback independently.
				{Query: `/* client A */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client A */ SELECT pg_advisory_xact_lock(31)`, Expected: []sql.Row{{nil}}},
				{Query: `/* client A */ ROLLBACK`, ExpectedTag: "ROLLBACK"},
				{Query: `/* client B */ SELECT pg_try_advisory_lock(31)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client B */ SELECT pg_advisory_unlock(31)`, Expected: []sql.Row{{"t"}}},

				// A waiting client must acquire the lock after the owner commits.
				{Query: `/* client A */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client A */ SELECT pg_advisory_xact_lock(32)`, Expected: []sql.Row{{nil}}},
				{Query: `/* client B */ BEGIN`, ExpectedTag: "BEGIN"},
				{Query: `/* client B */ SELECT pg_advisory_xact_lock(32)`, ExpectedBlocking: true},
				{Query: `/* client A */ COMMIT`, ExpectedTag: "COMMIT"},
				{Query: `/* client B */ COMMIT`, ExpectedTag: "COMMIT"},
				{Query: `/* client C */ SELECT pg_try_advisory_lock(32)`, Expected: []sql.Row{{"t"}}},
				{Query: `/* client C */ SELECT pg_advisory_unlock(32)`, Expected: []sql.Row{{"t"}}},
			},
		},
	})
}
