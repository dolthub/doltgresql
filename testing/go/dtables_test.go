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

func TestDoltSystemTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "dolt_branches",
			SetUpScript: []string{
				"CREATE SCHEMA test;",
				"SET SEARCH_PATH TO test;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SHOW search_path;",
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    "SELECT name, latest_commit_message FROM dolt_branches;",
					Expected: []sql.Row{{"main", "Initialize data repository"}},
				},
				{
					Query:    "SELECT dolt_branch('new_branch');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT name, latest_commit_message FROM dolt_branches;",
					Expected: []sql.Row{
						{"main", "Initialize data repository"},
						{"new_branch", "Initialize data repository"},
					},
				},
				{
					Query:    "SET SEARCH_PATH TO doltgres;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SHOW search_path;",
					Expected: []sql.Row{{"doltgres"}},
				},
				{
					Skip:     true, // TODO: This currently returns the `test` schema branch too
					Query:    "SELECT name, latest_commit_message FROM dolt_branches;",
					Expected: []sql.Row{{"main", "Initialize data repository"}},
				},
			},
		},
	})
}
