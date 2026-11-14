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

// TestTablePrivilegeFunctions covers the has_table_privilege and has_any_column_privilege functions.
// These are used by Metabase's Postgres driver during its schema sync (sync-tables).
func TestTablePrivilegeFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "has_table_privilege",
			SetUpScript: []string{
				"CREATE TABLE privilege_test (pk INT PRIMARY KEY);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT has_table_privilege('privilege_test', 'SELECT');",
					Expected:         []sql.Row{{"t"}},
					ExpectedColNames: []string{"has_table_privilege"},
				},
				{
					Query:            "SELECT has_table_privilege('privilege_test', 'DELETE');",
					Expected:         []sql.Row{{"t"}},
					ExpectedColNames: []string{"has_table_privilege"},
				},
				{
					Query:    "SELECT has_table_privilege(current_user, 'privilege_test', 'UPDATE');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT has_table_privilege(current_user, 'privilege_test'::regclass, 'INSERT');",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "has_any_column_privilege",
			SetUpScript: []string{
				"CREATE TABLE privilege_test (pk INT PRIMARY KEY);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT has_any_column_privilege('privilege_test', 'SELECT');",
					Expected:         []sql.Row{{"t"}},
					ExpectedColNames: []string{"has_any_column_privilege"},
				},
				{
					Query:    "SELECT has_any_column_privilege(current_user, 'privilege_test', 'UPDATE');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT has_any_column_privilege(current_user, 'privilege_test', 'INSERT');",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
	})
}
