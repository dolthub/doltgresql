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

// TestAllQuantifier verifies that `expr op ALL (array)` and `expr op ALL (subquery)` are supported, matching
// PostgreSQL's semantics: https://github.com/dolthub/doltgresql/issues/3013
func TestAllQuantifier(t *testing.T) {
	RunScripts(t, AllQuantifierTests)
}

var AllQuantifierTests = []ScriptTest{
	{
		Name: "ALL quantifier over arrays",
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT 1 = ALL(ARRAY[1, 1, 1]);",
				Expected: []sql.Row{
					{"t"},
				},
			},
			{
				Query: "SELECT 1 = ALL(ARRAY[1, 2, 1]);",
				Expected: []sql.Row{
					{"f"},
				},
			},
			{
				Query: "SELECT 1 = ALL(ARRAY[]::int[]);",
				Expected: []sql.Row{
					{"t"},
				},
			},
			{
				Query: "SELECT 1 = ALL(ARRAY[1, NULL]::int[]);",
				Expected: []sql.Row{
					{nil},
				},
			},
			{
				Query: "SELECT 1 = ALL(ARRAY[2, NULL]::int[]);",
				Expected: []sql.Row{
					{"f"},
				},
			},
			{
				Query: "SELECT 5 != ALL(ARRAY[1, 2, 3]);",
				Expected: []sql.Row{
					{"t"},
				},
			},
			{
				Query: "SELECT 2 != ALL(ARRAY[1, 2, 3]);",
				Expected: []sql.Row{
					{"f"},
				},
			},
		},
	},
	{
		Name: "ALL quantifier over subqueries",
		SetUpScript: []string{
			"create table t (i int primary key);",
			"insert into t values (1), (2), (3);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT 0 < ALL(SELECT i FROM t);",
				Expected: []sql.Row{
					{"t"},
				},
			},
			{
				Query: "SELECT 2 < ALL(SELECT i FROM t);",
				Expected: []sql.Row{
					{"f"},
				},
			},
			{
				Query: "SELECT 1 = ALL(SELECT i FROM t WHERE i = 4);",
				Expected: []sql.Row{
					{"t"},
				},
			},
		},
	},
	{
		// A NULL comparing against a nonempty array/subquery is unknown for every element, so the overall result is
		// NULL; against an empty array/subquery, ALL is vacuously true without ever evaluating a comparison.
		Name: "ALL quantifier with a NULL left-hand operand",
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT NULL = ALL(ARRAY[1, 2]::int[]);",
				Expected: []sql.Row{
					{nil},
				},
			},
			{
				Query: "SELECT NULL = ALL(ARRAY[]::int[]);",
				Expected: []sql.Row{
					{"t"},
				},
			},
			{
				Query: "SELECT NULL != ALL(ARRAY[]::int[]);",
				Expected: []sql.Row{
					{"t"},
				},
			},
		},
	},
	{
		// A subquery result set containing a NULL row should behave the same as an array literal containing NULL:
		// a false comparison short-circuits false regardless of the NULL, but absent any false, a NULL forces NULL.
		Name: "ALL quantifier over a subquery containing NULL rows",
		SetUpScript: []string{
			"create table t3 (id int primary key, v int);",
			"insert into t3 values (1, 1), (2, NULL), (3, 2);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT 1 = ALL(SELECT v FROM t3);",
				Expected: []sql.Row{
					{"f"},
				},
			},
			{
				Query: "SELECT 1 = ALL(SELECT v FROM t3 WHERE id IN (1, 2));",
				Expected: []sql.Row{
					{nil},
				},
			},
			{
				Query: "SELECT 3 != ALL(SELECT v FROM t3 WHERE id = 1);",
				Expected: []sql.Row{
					{"t"},
				},
			},
		},
	},
	{
		// ALL should filter rows the same way any other boolean expression does, including combining with NOT, and
		// should work per-row against a column (not just a single evaluated-once literal).
		Name: "ALL quantifier filtering table rows",
		SetUpScript: []string{
			"create table t2 (id int primary key, vals int[]);",
			"insert into t2 values (1, ARRAY[1,2,3]), (2, ARRAY[4,5,6]), (3, ARRAY[1,7,8]), (4, ARRAY[9,9,9]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT id FROM t2 WHERE 9 = ALL(vals) ORDER BY id;",
				Expected: []sql.Row{
					{4},
				},
			},
			{
				Query: "SELECT id FROM t2 WHERE 1 != ALL(vals) ORDER BY id;",
				Expected: []sql.Row{
					{2},
					{4},
				},
			},
			{
				Query: "SELECT id FROM t2 WHERE NOT (1 != ALL(vals)) ORDER BY id;",
				Expected: []sql.Row{
					{1},
					{3},
				},
			},
		},
	},
	{
		// The right-hand array of ALL/ANY/SOME can be an untyped bind variable; its type is inferred from the
		// left-hand expression (see the BindVar handling in AnyExpr.WithChildren).
		Name: "ALL quantifier with an untyped bind variable array",
		SetUpScript: []string{
			"create table t4 (id int primary key);",
			"insert into t4 values (1), (2), (3);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT id FROM t4 WHERE id != ALL($1) ORDER BY id;",
				BindVars: []any{[]int32{1, 2}},
				Expected: []sql.Row{
					{3},
				},
			},
		},
	},
	{
		Name: "!= works with all three quantifiers",
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT 1 != ANY(ARRAY[1, 2, 3]);",
				Expected: []sql.Row{
					{"t"},
				},
			},
			{
				Query: "SELECT 1 != SOME(ARRAY[1, 1, 1]);",
				Expected: []sql.Row{
					{"f"},
				},
			},
			{
				Query: "SELECT 1 != ALL(ARRAY[1, 1, 1]);",
				Expected: []sql.Row{
					{"f"},
				},
			},
		},
	},
	{
		// The standard cycle-check idiom used by recursive CTEs, as reported in
		// https://github.com/dolthub/doltgresql/issues/3013
		Name: "ALL quantifier in a recursive CTE cycle check",
		Assertions: []ScriptTestAssertion{
			{
				Query: `WITH RECURSIVE r(id, path) AS (
					SELECT 'a'::text, ARRAY['a'::text]
					UNION ALL
					SELECT r.id || 'x', r.path || (r.id || 'x')
					FROM r
					WHERE length(r.id) < 3 AND (r.id || 'x') != ALL(r.path)
				)
				SELECT count(*) FROM r;`,
				Expected: []sql.Row{
					{3},
				},
			},
		},
	},
	{
		Name: "any expression with array in string format",
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT 'i' = ANY('{information_schema, something}');`,
				Expected: []sql.Row{
					{"f"},
				},
			},
			{
				Query: `SELECT 'somedb' = ANY('{information_schema, somedb}');`,
				Expected: []sql.Row{
					{"t"},
				},
			},
			{
				Query: `SELECT 'somedb' = SOME('{information_schema, somedb}');`,
				Expected: []sql.Row{
					{"t"},
				},
			},
			{
				Query: `SELECT 'somedb' = ALL('{information_schema, somedb}');`,
				Expected: []sql.Row{
					{"f"},
				},
			},
			{
				Query: `SELECT nsp.nspname AS schema_name,
       (nsp.nspname = 'pg_catalog'
        AND EXISTS (SELECT 1
                    FROM   pg_catalog.pg_class
                    WHERE  relname = 'pg_class'
                           AND relnamespace = nsp.oid LIMIT 1))
       OR (nsp.nspname = 'pgagent'
           AND EXISTS (SELECT 1
                       FROM   pg_catalog.pg_class
                       WHERE  relname = 'pga_job'
                              AND relnamespace = nsp.oid LIMIT 1))
       OR (nsp.nspname = 'information_schema'
           AND EXISTS (SELECT 1
                       FROM   pg_catalog.pg_class
                       WHERE  relname = 'tables'
                              AND relnamespace = nsp.oid LIMIT 1)) AS is_catalog,
       CASE
         WHEN nsp.nspname = ANY('{information_schema}')
         THEN FALSE
         ELSE TRUE
       END AS db_support
FROM   pg_catalog.pg_namespace nsp
WHERE  nsp.oid = 2200::OID;`,
				ExpectedColNames: []string{"schema_name", "is_catalog", "db_support"},
				Expected:         []sql.Row{{"public", "f", "t"}},
			},
			{
				Query:    `select oid from pg_class join pg_index on pg_class.oid = ANY(indclass);`, // oid = ANY(oidvector)
				Expected: []sql.Row{},
			},
		},
	},
}
