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

// TestImplicitLateralJoin tests that a set-returning function in the FROM list may reference columns of
// tables listed earlier in the same FROM clause, which is an implicit LATERAL join in Postgres.
// https://github.com/dolthub/doltgresql/issues/3112
func TestImplicitLateralJoin(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Issue #3112: implicit lateral join for set-returning functions",
			SetUpScript: []string{
				"CREATE TABLE bug15 (a integer, b integer);",
				"CREATE INDEX bug15_ab ON bug15 (a, b);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// The exact repro from the issue: unnest over a column of a preceding FROM item
					Query:    "SELECT k.u FROM pg_index i, unnest(i.indkey) AS k(u) WHERE i.indexrelid = 'bug15_ab'::regclass;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					// The same query with an explicit LATERAL keyword
					Query:    "SELECT k.u FROM pg_index i, LATERAL unnest(i.indkey) AS k(u) WHERE i.indexrelid = 'bug15_ab'::regclass;",
					Expected: []sql.Row{{1}, {2}},
				},
			},
		},
		{
			Name: "implicit lateral unnest over user table",
			SetUpScript: []string{
				"CREATE TABLE t1 (id integer, arr integer[]);",
				"INSERT INTO t1 VALUES (1, ARRAY[10, 20]), (2, ARRAY[30]);",
				"CREATE TABLE t2 (id integer, name text);",
				"INSERT INTO t2 VALUES (1, 'one'), (2, 'two');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT id, u FROM t1, unnest(t1.arr) AS x(u) ORDER BY id, u;",
					Expected: []sql.Row{{1, 10}, {1, 20}, {2, 30}},
				},
				{
					// Explicit LATERAL keyword (a noise word for functions in FROM)
					Query:    "SELECT id, u FROM t1, LATERAL unnest(t1.arr) AS x(u) ORDER BY id, u;",
					Expected: []sql.Row{{1, 10}, {1, 20}, {2, 30}},
				},
				{
					// Unqualified column reference in the function's arguments
					Query:    "SELECT id, u FROM t1, unnest(arr) AS x(u) ORDER BY id, u;",
					Expected: []sql.Row{{1, 10}, {1, 20}, {2, 30}},
				},
				{
					// Qualified reference to the function's output column
					Query:    "SELECT id, x.u FROM t1, unnest(t1.arr) AS x(u) ORDER BY id, x.u;",
					Expected: []sql.Row{{1, 10}, {1, 20}, {2, 30}},
				},
				{
					Query: "SELECT * FROM t1, unnest(t1.arr) AS x(u) ORDER BY id, u;",
					Expected: []sql.Row{
						{1, "{10,20}", 10},
						{1, "{10,20}", 20},
						{2, "{30}", 30},
					},
				},
				{
					// Without an alias, the function's name is the implicit alias and column name
					Query:    "SELECT id, unnest FROM t1, unnest(t1.arr) ORDER BY 1, 2;",
					Expected: []sql.Row{{1, 10}, {1, 20}, {2, 30}},
				},
				{
					// Multiple tables before the set-returning function
					Query:    "SELECT t2.name, u FROM t1, t2, unnest(t1.arr) AS x(u) WHERE t1.id = t2.id ORDER BY u;",
					Expected: []sql.Row{{"one", 10}, {"one", 20}, {"two", 30}},
				},
				{
					// Explicit CROSS JOIN with a set-returning function is also an implicit lateral join
					Query:    "SELECT id, u FROM t1 CROSS JOIN unnest(t1.arr) AS x(u) ORDER BY id, u;",
					Expected: []sql.Row{{1, 10}, {1, 20}, {2, 30}},
				},
				{
					// Explicit INNER JOIN with a set-returning function is also an implicit lateral join
					Query:    "SELECT id, u FROM t1 INNER JOIN unnest(t1.arr) AS x(u) ON true ORDER BY id, u;",
					Expected: []sql.Row{{1, 10}, {1, 20}, {2, 30}},
				},
				{
					// A set-returning function other than unnest
					Query:    "SELECT id, n FROM t1, generate_series(1, t1.id) AS g(n) ORDER BY id, n;",
					Expected: []sql.Row{{1, 1}, {2, 1}, {2, 2}},
				},
				{
					// WITH ORDINALITY over an implicit lateral function
					Query: "SELECT id, u, ord FROM t1, unnest(t1.arr) WITH ORDINALITY AS x(u, ord) ORDER BY 1, 2;",
					Expected: []sql.Row{
						{1, 10, 1},
						{1, 20, 2},
						{2, 30, 1},
					},
				},
				{
					// WITH ORDINALITY over an explicit lateral function
					Query: "SELECT id, u, ord FROM t1, LATERAL unnest(t1.arr) WITH ORDINALITY AS x(u, ord) ORDER BY 1, 2;",
					Expected: []sql.Row{
						{1, 10, 1},
						{1, 20, 2},
						{2, 30, 1},
					},
				},
				{
					// Set-returning functions without references to preceding FROM items still work
					Query:    "SELECT id, u FROM t1, unnest(ARRAY[7]) AS x(u) ORDER BY id;",
					Expected: []sql.Row{{1, 7}, {2, 7}},
				},
				{
					// Subqueries are NOT implicitly lateral: they require the explicit LATERAL keyword
					Query:       "SELECT * FROM t1, (SELECT t1.id) s;",
					ExpectedErr: "table not found: t1",
				},
			},
		},
		{
			Name: "left lateral join null-extends rows when the function returns no rows",
			SetUpScript: []string{
				"CREATE TABLE t3 (id integer, vals integer[]);",
				"INSERT INTO t3 VALUES (1, ARRAY[]::integer[]), (2, ARRAY[10, 20]), (3, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT t3.id, u FROM t3 LEFT JOIN LATERAL unnest(t3.vals) AS x(u) ON true ORDER BY t3.id, u NULLS FIRST;",
					Expected: []sql.Row{
						{1, nil},
						{2, 10},
						{2, 20},
						{3, nil},
					},
				},
				{
					// A non-trivial join condition must also null-extend non-matching rows
					Query: "SELECT t3.id, u FROM t3 LEFT JOIN LATERAL unnest(t3.vals) AS x(u) ON u > 10 ORDER BY t3.id, u NULLS FIRST;",
					Expected: []sql.Row{
						{1, nil},
						{2, 20},
						{3, nil},
					},
				},
			},
		},
	})
}
