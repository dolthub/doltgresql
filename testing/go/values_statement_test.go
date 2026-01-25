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
	{
		Name:        "VALUES with mixed int and decimal - issue 1648",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// Integer first, then decimal - should resolve to numeric
				Query: `SELECT * FROM (VALUES(1),(2.01),(3)) v(n);`,
				Expected: []sql.Row{
					{Numeric("1")},
					{Numeric("2.01")},
					{Numeric("3")},
				},
			},
			{
				// Decimal first, then integers - should resolve to numeric
				Query: `SELECT * FROM (VALUES(1.01),(2),(3)) v(n);`,
				Expected: []sql.Row{
					{Numeric("1.01")},
					{Numeric("2")},
					{Numeric("3")},
				},
			},
			{
				// SUM should work directly now that VALUES has correct type
				// Note: SUM returns float64 (double precision) for numeric input
				Query:    `SELECT SUM(n) FROM (VALUES(1),(2.01),(3)) v(n);`,
				Expected: []sql.Row{{6.01}},
			},
		},
	},
	{
		Name:        "VALUES with multiple columns mixed types",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT * FROM (VALUES(1, 'a'), (2.5, 'b')) v(num, str);`,
				Expected: []sql.Row{
					{Numeric("1"), "a"},
					{Numeric("2.5"), "b"},
				},
			},
		},
	},
}
