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
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// jsonbLexicalOrder is the 40 test values in ascending comparison order as determined
// by CompareJSON (type precedence: null < number < string < object < array < boolean).
// Objects are ordered by: shared-key value comparison, then key-count, then least
// non-shared key lexicographically.  Arrays are ordered element-by-element; shorter
// wins when all shared positions are equal.
var jsonbLexicalOrder = []string{
	`null`,
	`-1`,
	`0`,
	`1`,
	`2`,
	`3.14`,
	`42`,
	`100`,
	`9999`,
	`"a"`,
	`"ab"`,
	`"abc"`,
	`"b"`,
	`"foo"`,
	`"hello"`,
	`"hello world"`,
	`"longer string value"`,
	`"z"`,
	`{}`,
	`{"a":1}`,
	`{"a":{"b":{"c":1}}}`,
	`{"aa":1}`,
	`{"b":2}`,
	`{"x":{"y":1}}`,
	`{"z":null}`,
	`{"a":1,"b":2}`,
	`{"name":"test","value":42}`,
	`{"a":1,"b":2,"c":3}`,
	`[]`,
	`[null]`,
	`[1]`,
	`[1,2]`,
	`[1,2,3]`,
	`["a"]`,
	`["a","b","c"]`,
	`[[1,2],[3,4]]`,
	`[false]`,
	`[true]`,
	`false`,
	`true`,
}

// jsonbExpectedOutput is the JSONB-normalized text representation of each value in
// jsonbLexicalOrder.  Keys in objects are sorted by length then lexicographically;
// objects and arrays use a space after each ':' and ','.
var jsonbExpectedOutput = []string{
	`null`,
	`-1`,
	`0`,
	`1`,
	`2`,
	`3.14`,
	`42`,
	`100`,
	`9999`,
	`"a"`,
	`"ab"`,
	`"abc"`,
	`"b"`,
	`"foo"`,
	`"hello"`,
	`"hello world"`,
	`"longer string value"`,
	`"z"`,
	`{}`,
	`{"a": 1}`,
	`{"a": {"b": {"c": 1}}}`,
	`{"aa": 1}`,
	`{"b": 2}`,
	`{"x": {"y": 1}}`,
	`{"z": null}`,
	`{"a": 1, "b": 2}`,
	`{"name": "test", "value": 42}`,
	`{"a": 1, "b": 2, "c": 3}`,
	`[]`,
	`[null]`,
	`[1]`,
	`[1, 2]`,
	`[1, 2, 3]`,
	`["a"]`,
	`["a", "b", "c"]`,
	`[[1, 2], [3, 4]]`,
	`[false]`,
	`[true]`,
	`false`,
	`true`,
}

// TestJsonBPairwiseLessThan walks jsonbLexicalOrder and asserts that every
// consecutive pair satisfies the < operator: SELECT 'a'::jsonb < 'b'::jsonb = t.
// It also inserts all values into an indexed JSONB column and verifies that
// ORDER BY returns them in exactly the same order.
func TestJsonBPairwiseLessThan(t *testing.T) {
	// 39 pairwise < assertions
	assertions := make([]ScriptTestAssertion, 0, len(jsonbLexicalOrder))
	for i := 0; i < len(jsonbLexicalOrder)-1; i++ {
		a, b := jsonbLexicalOrder[i], jsonbLexicalOrder[i+1]
		assertions = append(assertions, ScriptTestAssertion{
			Query:    `SELECT '` + a + `'::jsonb < '` + b + `'::jsonb`,
			Expected: []sql.Row{{"t"}},
		})
	}

	// ORDER BY assertion: index scan must return rows in the same order
	orderByExpected := make([]sql.Row, len(jsonbExpectedOutput))
	for i, v := range jsonbExpectedOutput {
		orderByExpected[i] = sql.Row{v}
	}
	assertions = append(assertions, ScriptTestAssertion{
		Query:    `SELECT val FROM jorder ORDER BY val`,
		Expected: orderByExpected,
	})

	// Build the VALUES list for the INSERT
	vals := make([]string, len(jsonbLexicalOrder))
	for i, v := range jsonbLexicalOrder {
		vals[i] = `('` + v + `')`
	}

	RunScriptsWithoutNormalization(t, []ScriptTest{
		{
			Name: "JSONB pairwise less-than along lexical order",
			SetUpScript: []string{
				`CREATE TABLE jorder (val JSONB NOT NULL)`,
				`CREATE INDEX jorder_val_idx ON jorder (val)`,
				`INSERT INTO jorder (val) VALUES ` + strings.Join(vals, ", "),
			},
			Assertions: assertions,
		},
	})
}

// TestJsonBIndexSortConsistency verifies that, for an indexed JSONB column, the sort
// order produced by ORDER BY is consistent with the < and > comparison operators.
//
// Contract: for every pair of elements (a, b) stored in the table,
//   - if a < b (operator), then a must appear before b in ORDER BY
//   - if a > b (operator), then a must appear after  b in ORDER BY
//
// The test uses a variety of JSON documents — nulls, booleans, numbers, strings,
// arrays, and objects at different nesting depths and key counts — to maximise
// the chance of exposing any mismatch between the index byte-encoding sort order
// and the semantic comparison order used by < / >.
func TestJsonBIndexSortConsistency(t *testing.T) {
	RunScriptsWithoutNormalization(t, []ScriptTest{
		{
			Name: "JSONB index sort order matches < and > operators",
			SetUpScript: []string{
				`CREATE TABLE jtest (id SERIAL PRIMARY KEY, val JSONB NOT NULL)`,
				`CREATE INDEX jtest_val_idx ON jtest (val)`,
				// Diverse documents: null, booleans, numbers (various magnitudes /
				// signs / decimals), strings (short to long), arrays (empty,
				// scalars, nested), objects (empty, single key, multi-key, nested).
				`INSERT INTO jtest (val) VALUES
					('null'),
					('false'),
					('true'),
					('-1'),
					('0'),
					('1'),
					('2'),
					('3.14'),
					('42'),
					('100'),
					('9999'),
					('"a"'),
					('"b"'),
					('"z"'),
					('"ab"'),
					('"abc"'),
					('"foo"'),
					('"hello"'),
					('"hello world"'),
					('"longer string value"'),
					('[]'),
					('[1]'),
					('[1,2]'),
					('[1,2,3]'),
					('["a"]'),
					('[null]'),
					('[false]'),
					('[true]'),
					('["a","b","c"]'),
					('[[1,2],[3,4]]'),
					('{}'),
					('{"a":1}'),
					('{"b":2}'),
					('{"aa":1}'),
					('{"a":1,"b":2}'),
					('{"a":1,"b":2,"c":3}'),
					('{"x":{"y":1}}'),
					('{"name":"test","value":42}'),
					('{"a":{"b":{"c":1}}}'),
					('{"z":null}')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Transitivity check: if a < b and b < c then a < c must hold.
					// A non-zero count reveals a comparison function that is not a
					// valid total order, which makes any index ordering undefined.
					Query: `SELECT COUNT(*) FROM jtest a
					        JOIN jtest b ON a.val < b.val
					        JOIN jtest c ON b.val < c.val
					        WHERE NOT (a.val < c.val)`,
					Expected: []sql.Row{{int64(0)}},
				},
				{
					// Converse transitivity via >: if a > b and b > c then a > c.
					Query: `SELECT COUNT(*) FROM jtest a
					        JOIN jtest b ON a.val > b.val
					        JOIN jtest c ON b.val > c.val
					        WHERE NOT (a.val > c.val)`,
					Expected: []sql.Row{{int64(0)}},
				},
				{
					// Adjacent-pair check using LAG: for every consecutive pair
					// (prev, curr) in the ORDER BY sequence, prev < curr must hold.
					// This catches any local inversion in the sort order.
					Query: `SELECT COUNT(*) FROM (
					          SELECT val,
					                 LAG(val) OVER (ORDER BY val) AS prev_val
					          FROM jtest
					        ) t
					        WHERE prev_val IS NOT NULL AND NOT (prev_val < val)`,
					Expected: []sql.Row{{int64(0)}},
				},
				{
					// LEAD version of the same check: curr < next must hold for
					// every consecutive pair.
					Query: `SELECT COUNT(*) FROM (
					          SELECT val,
					                 LEAD(val) OVER (ORDER BY val) AS next_val
					          FROM jtest
					        ) t
					        WHERE next_val IS NOT NULL AND NOT (val < next_val)`,
					Expected: []sql.Row{{int64(0)}},
				},
				{
					// Antisymmetry: if a < b then NOT (a > b).
					// A non-zero count indicates the comparison function is not
					// antisymmetric, which would make index ordering undefined.
					Query: `SELECT COUNT(*) FROM jtest a
					        JOIN jtest b ON a.val < b.val
					        WHERE a.val > b.val`,
					Expected: []sql.Row{{int64(0)}},
				},
				{
					// Totality / strict total order: for every unequal pair,
					// exactly one of a < b or b < a must hold.
					// If neither holds for a pair of unequal values the relation
					// is not a total order and the index cannot sort it correctly.
					Query: `SELECT COUNT(*) FROM jtest a
					        JOIN jtest b ON a.id < b.id
					        WHERE a.val <> b.val
					          AND NOT (a.val < b.val)
					          AND NOT (b.val < a.val)`,
					Expected: []sql.Row{{int64(0)}},
				},
			},
		},
	})
}
