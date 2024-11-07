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

func TestValuesStatement(t *testing.T) {
	RunScripts(t, ValuesStatementTests)
}

var ValuesStatementTests = []ScriptTest{
	{
		Name:        "basic values statements",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT * FROM (VALUES (1), (2), (3)) sqa;`,
				Expected: []sql.Row{
					{1},
					{2},
					{3},
				},
			},
			{
				Query: `SELECT * FROM (VALUES (1, 2), (3, 4)) sqa;`,
				Expected: []sql.Row{
					{1, 2},
					{3, 4},
				},
			},
			{
				Query: `SELECT i * 10, j * 100 FROM (VALUES (1, 2), (3, 4)) sqa(i, j);`,
				Expected: []sql.Row{
					{10, 200},
					{30, 400},
				},
			},
		},
	},
}
