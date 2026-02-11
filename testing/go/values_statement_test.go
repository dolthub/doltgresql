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
				Query:    `SELECT SUM(n) FROM (VALUES(1),(2.01),(3)) v(n);`,
				Expected: []sql.Row{{Numeric("6.01")}},
			},
			{
				// Exact repro from issue #1648: integer first, explicit cast to numeric
				Query:    `SELECT SUM(n::numeric) FROM (VALUES(1),(2.01),(3)) v(n);`,
				Expected: []sql.Row{{Numeric("6.01")}},
			},
			{
				// Exact repro from issue #1648: decimal first, explicit cast to numeric
				Query:    `SELECT SUM(n::numeric) FROM (VALUES(1.01),(2),(3)) v(n);`,
				Expected: []sql.Row{{Numeric("6.01")}},
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
					{"a", Numeric("4")},
					{"b", Numeric("7.0")},
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
				// TODO: pre-existing bug: arithmetic in subquery over VALUES is not applied (returns original values)
				Skip:  true,
				Query: `SELECT * FROM (SELECT n * 2 AS doubled FROM (VALUES(1),(2.5),(3)) v(n)) sub;`,
				Expected: []sql.Row{
					{Numeric("2")},
					{Numeric("5.0")},
					{Numeric("6")},
				},
			},
			{
				// VALUES with LIMIT inside subquery
				// TODO: pre-existing bug: LIMIT inside subquery over VALUES is ignored (returns all rows)
				Skip:  true,
				Query: `SELECT * FROM (SELECT * FROM (VALUES(1),(2.5),(3),(4.5)) v(n) LIMIT 2) sub;`,
				Expected: []sql.Row{
					{Numeric("1")},
					{Numeric("2.5")},
				},
			},
			{
				// VALUES with ORDER BY inside subquery
				// TODO: pre-existing bug - ORDER BY inside subquery over VALUES is ignored
				Skip:  true,
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
	{
		Name:        "VALUES with NULL values",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// NULL mixed with integers - should resolve to integer, NULL stays NULL
				Query: `SELECT * FROM (VALUES(1),(NULL),(3)) v(n);`,
				Expected: []sql.Row{
					{int32(1)},
					{nil},
					{int32(3)},
				},
			},
			{
				// NULL mixed with decimals - should resolve to numeric
				Query: `SELECT * FROM (VALUES(1.5),(NULL),(3.5)) v(n);`,
				Expected: []sql.Row{
					{Numeric("1.5")},
					{nil},
					{Numeric("3.5")},
				},
			},
			{
				// NULL mixed with int and decimal - should resolve to numeric
				Query: `SELECT * FROM (VALUES(1),(NULL),(2.5)) v(n);`,
				Expected: []sql.Row{
					{Numeric("1")},
					{nil},
					{Numeric("2.5")},
				},
			},
			{
				// All NULLs - should resolve to text (PostgreSQL behavior)
				Query: `SELECT * FROM (VALUES(NULL),(NULL)) v(n);`,
				Expected: []sql.Row{
					{nil},
					{nil},
				},
			},
		},
	},
	{
		Name:        "VALUES type mismatch errors",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// Integer and unknown('text'): FindCommonType resolves to int4 (the non-unknown type),
				// then the I/O cast from 'text' to int4 fails at execution time. This matches PostgreSQL behavior:
				// psql returns "invalid input syntax for type integer: "text""
				Query:       `SELECT * FROM (VALUES(1),('text'),(3)) v(n);`,
				ExpectedErr: "invalid input syntax for type int4",
			},
			{
				// Boolean and integer cannot be matched
				Query:       `SELECT * FROM (VALUES(true),(1),(false)) v(n);`,
				ExpectedErr: "cannot be matched",
			},
		},
	},
	{
		Name:        "VALUES with all unknown types (string literals)",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// All string literals should resolve to text
				Query: `SELECT * FROM (VALUES('a'),('b'),('c')) v(n);`,
				Expected: []sql.Row{
					{"a"},
					{"b"},
					{"c"},
				},
			},
			{
				// String literals with operations
				Query: `SELECT n || '!' FROM (VALUES('hello'),('world')) v(n);`,
				Expected: []sql.Row{
					{"hello!"},
					{"world!"},
				},
			},
		},
	},
	{
		Name:        "VALUES with array types",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// Integer arrays: doltgresql returns arrays in text format over the wire
				Query: `SELECT * FROM (VALUES(ARRAY[1,2]),(ARRAY[3,4])) v(arr);`,
				Expected: []sql.Row{
					{"{1,2}"},
					{"{3,4}"},
				},
			},
			{
				// Text arrays: doltgresql returns arrays in text format over the wire
				Query: `SELECT * FROM (VALUES(ARRAY['a','b']),(ARRAY['c','d'])) v(arr);`,
				Expected: []sql.Row{
					{"{a,b}"},
					{"{c,d}"},
				},
			},
		},
	},
	{
		Name:        "VALUES with all same type multi-row (no casts needed)",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// All integers
				Query: `SELECT * FROM (VALUES(1),(2),(3)) v(n);`,
				Expected: []sql.Row{
					{int32(1)},
					{int32(2)},
					{int32(3)},
				},
			},
			{
				// All decimals
				Query: `SELECT * FROM (VALUES(1.5),(2.5),(3.5)) v(n);`,
				Expected: []sql.Row{
					{Numeric("1.5")},
					{Numeric("2.5")},
					{Numeric("3.5")},
				},
			},
			{
				// All text
				Query: `SELECT * FROM (VALUES('x'),('y'),('z')) v(n);`,
				Expected: []sql.Row{
					{"x"},
					{"y"},
					{"z"},
				},
			},
		},
	},
	{
		Name:        "VALUES with multi-column partial cast",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// Only first column needs cast
				Query: `SELECT * FROM (VALUES(1, 'a'),(2.5, 'b'),(3, 'c')) v(num, str);`,
				Expected: []sql.Row{
					{Numeric("1"), "a"},
					{Numeric("2.5"), "b"},
					{Numeric("3"), "c"},
				},
			},
			{
				// Only second column needs cast
				Query: `SELECT * FROM (VALUES(1, 10),(2, 20.5),(3, 30)) v(a, b);`,
				Expected: []sql.Row{
					{int32(1), Numeric("10")},
					{int32(2), Numeric("20.5")},
					{int32(3), Numeric("30")},
				},
			},
		},
	},
	{
		Name:        "VALUES in CTE (WITH clause)",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// Mixed types via CTE
				Query: `WITH nums AS (SELECT * FROM (VALUES(1),(2.5),(3)) v(n)) SELECT * FROM nums;`,
				Expected: []sql.Row{
					{Numeric("1")},
					{Numeric("2.5")},
					{Numeric("3")},
				},
			},
			{
				// SUM over CTE
				Query:    `WITH nums AS (SELECT * FROM (VALUES(1),(2.5),(3)) v(n)) SELECT SUM(n) FROM nums;`,
				Expected: []sql.Row{{Numeric("6.5")}},
			},
		},
	},
	{
		Name:        "VALUES with JOIN",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT a.n, b.label FROM (VALUES(1),(2),(3)) a(n) JOIN (VALUES(1, 'one'),(2, 'two'),(3, 'three')) b(id, label) ON a.n = b.id;`,
				Expected: []sql.Row{
					{int32(1), "one"},
					{int32(2), "two"},
					{int32(3), "three"},
				},
			},
			{
				// Mixed types in one of the joined VALUES
				Query: `SELECT a.n, b.label FROM (VALUES(1),(2.5),(3)) a(n) JOIN (VALUES(1, 'one'),(3, 'three')) b(id, label) ON a.n = b.id;`,
				Expected: []sql.Row{
					{Numeric("1"), "one"},
					{Numeric("3"), "three"},
				},
			},
		},
	},
	{
		Name:        "VALUES with same-type booleans",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				// All booleans, returned as "t"/"f" over the wire
				Query: `SELECT * FROM (VALUES(true),(false),(true)) v(b);`,
				Expected: []sql.Row{
					{"t"},
					{"f"},
					{"t"},
				},
			},
			{
				// Boolean WHERE filter
				Query: `SELECT * FROM (VALUES(true),(false),(true),(false)) v(b) WHERE b = true;`,
				Expected: []sql.Row{
					{"t"},
					{"t"},
				},
			},
		},
	},
}
