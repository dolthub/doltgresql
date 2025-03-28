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
	})
}
