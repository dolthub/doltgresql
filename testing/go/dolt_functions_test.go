// Copyright 2023 Dolthub, Inc.
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

func TestDoltFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "smoke test select dolt_add and dolt_commit",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk int primary key);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select dolt_add('.')",
					Expected: []sql.Row{
						{"{0}"},
					},
				},
				{
					Query:            "select dolt_commit('-am', 'initial commit')",
					SkipResultsCheck: true,
				},
				{
					Query: "select count(*) from dolt_log",
					Expected: []sql.Row{
						{2},
					},
				},
				{
					Query: "select message from dolt_log order by date desc limit 1",
					Expected: []sql.Row{
						{"initial commit"},
					},
				},
			},
		},
	})
}
