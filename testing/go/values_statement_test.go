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
	{
		Name:        "VALUES with GROUP BY",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// GROUP BY on mixed type VALUES - tests that GetField types are updated correctly
				Query: `SELECT n, COUNT(*) FROM (VALUES(1),(2.5),(1),(3.5),(2.5)) v(n) GROUP BY n ORDER BY n;`,
				Expected: []sql.Row{
					{Numeric("1"), int64(2)},
					{Numeric("2.5"), int64(2)},
					{Numeric("3.5"), int64(1)},
				},
			},
			{
				// SUM with GROUP BY
				Query: `SELECT category, SUM(amount) FROM (VALUES('a', 1),('b', 2.5),('a', 3),('b', 4.5)) v(category, amount) GROUP BY category ORDER BY category;`,
				Expected: []sql.Row{
					{"a", 4.0},
					{"b", 7.0},
				},
			},
		},
	},
	{
		Name:        "VALUES with DISTINCT",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// DISTINCT on mixed type VALUES
				Query: `SELECT DISTINCT n FROM (VALUES(1),(2.5),(1),(2.5),(3)) v(n) ORDER BY n;`,
				Expected: []sql.Row{
					{Numeric("1")},
					{Numeric("2.5")},
					{Numeric("3")},
				},
			},
		},
	},
	{
		Name:        "VALUES with LIMIT and OFFSET",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// LIMIT on mixed type VALUES
				Query: `SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) LIMIT 3;`,
				Expected: []sql.Row{
					{Numeric("1")},
					{Numeric("2.5")},
					{Numeric("3")},
				},
			},
			{
				// LIMIT with OFFSET
				Query: `SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) LIMIT 2 OFFSET 2;`,
				Expected: []sql.Row{
					{Numeric("3")},
					{Numeric("4.5")},
				},
			},
		},
	},
	{
		Name:        "VALUES with ORDER BY",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// ORDER BY on mixed type VALUES - ascending
				Query: `SELECT * FROM (VALUES(3),(1.5),(2),(4.5)) v(n) ORDER BY n;`,
				Expected: []sql.Row{
					{Numeric("1.5")},
					{Numeric("2")},
					{Numeric("3")},
					{Numeric("4.5")},
				},
			},
			{
				// ORDER BY descending
				Query: `SELECT * FROM (VALUES(3),(1.5),(2),(4.5)) v(n) ORDER BY n DESC;`,
				Expected: []sql.Row{
					{Numeric("4.5")},
					{Numeric("3")},
					{Numeric("2")},
					{Numeric("1.5")},
				},
			},
		},
	},
	{
		Name:        "VALUES in subquery",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// VALUES as subquery in FROM clause
				Query: `SELECT * FROM (SELECT n * 2 AS doubled FROM (VALUES(1),(2.5),(3)) v(n)) sub;`,
				Expected: []sql.Row{
					{Numeric("2")},
					{Numeric("5.0")},
					{Numeric("6")},
				},
			},
			{
				// VALUES with LIMIT inside subquery
				Query: `SELECT * FROM (SELECT * FROM (VALUES(1),(2.5),(3),(4.5)) v(n) LIMIT 2) sub;`,
				Expected: []sql.Row{
					{Numeric("1")},
					{Numeric("2.5")},
				},
			},
			{
				// VALUES with ORDER BY inside subquery
				Query: `SELECT * FROM (SELECT * FROM (VALUES(3),(1.5),(2)) v(n) ORDER BY n) sub;`,
				Expected: []sql.Row{
					{Numeric("1.5")},
					{Numeric("2")},
					{Numeric("3")},
				},
			},
		},
	},
	{
		Name:        "VALUES with WHERE clause (Filter node)",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// Filter on mixed type VALUES
				Query: `SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) WHERE n > 2;`,
				Expected: []sql.Row{
					{Numeric("2.5")},
					{Numeric("3")},
					{Numeric("4.5")},
					{Numeric("5")},
				},
			},
			{
				// Filter with multiple conditions
				Query: `SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) WHERE n > 1 AND n < 4.5;`,
				Expected: []sql.Row{
					{Numeric("2.5")},
					{Numeric("3")},
				},
			},
		},
	},
	{
		Name:        "VALUES with aggregate functions",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// AVG on mixed types
				Query:    `SELECT AVG(n) FROM (VALUES(1),(2),(3),(4)) v(n);`,
				Expected: []sql.Row{{2.5}},
			},
			{
				// MIN/MAX on mixed types
				Query: `SELECT MIN(n), MAX(n) FROM (VALUES(1),(2.5),(3),(0.5)) v(n);`,
				Expected: []sql.Row{
					{Numeric("0.5"), Numeric("3")},
				},
			},
		},
	},
	{
		Name:        "VALUES combined operations",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// GROUP BY + ORDER BY + LIMIT
				Query: `SELECT n, COUNT(*) as cnt FROM (VALUES(1),(2.5),(1),(2.5),(3),(1)) v(n) GROUP BY n ORDER BY cnt DESC LIMIT 2;`,
				Expected: []sql.Row{
					{Numeric("1"), int64(3)},
					{Numeric("2.5"), int64(2)},
				},
			},
			{
				// DISTINCT + ORDER BY + LIMIT
				Query: `SELECT DISTINCT n FROM (VALUES(1),(2.5),(1),(3),(2.5),(4)) v(n) ORDER BY n DESC LIMIT 3;`,
				Expected: []sql.Row{
					{Numeric("4")},
					{Numeric("3")},
					{Numeric("2.5")},
				},
			},
			{
				// WHERE + ORDER BY + LIMIT
				Query: `SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) WHERE n > 1 ORDER BY n DESC LIMIT 2;`,
				Expected: []sql.Row{
					{Numeric("5")},
					{Numeric("4.5")},
				},
			},
		},
	},
	{
		Name:        "VALUES with single row (no type unification needed)",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// Single row should pass through unchanged
				Query: `SELECT * FROM (VALUES(42)) v(n);`,
				Expected: []sql.Row{
					{int32(42)},
				},
			},
			{
				// Single row with decimal
				Query: `SELECT * FROM (VALUES(3.14)) v(n);`,
				Expected: []sql.Row{
					{Numeric("3.14")},
				},
			},
		},
	},
}
