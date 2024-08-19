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
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestExpressions(t *testing.T) {
	RunScriptsWithoutNormalization(t, []ScriptTest{
		anyTests("ANY"),
		anyTests("SOME"),
		{
			Name: "IN",
			SetUpScript: []string{
				`CREATE TABLE test (id INT);`,
				`INSERT INTO test VALUES (1), (3), (2);`,

				`CREATE TABLE test2 (id INT, test_id INT, txt text);`,
				`INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM test WHERE id IN (2, 3, 4, 5);`,
					Expected: []sql.Row{{int32(3)}, {int32(2)}},
				},
				{
					Query:    `SELECT * FROM test WHERE id IN (4, 3, 2, 1, 0);`,
					Expected: []sql.Row{{int32(1)}, {int32(3)}, {int32(2)}},
				},
				{
					Query:    `SELECT * FROM test2 WHERE test_id IN (SELECT * FROM test WHERE id = 2);`,
					Expected: []sql.Row{{int32(3), int32(2), "baz"}},
				},
				{
					Query: `SELECT * FROM test2 WHERE test_id IN(SELECT * FROM test WHERE id > 0);`,
					Expected: []sql.Row{
						{int32(1), int32(1), "foo"},
						{int32(3), int32(2), "baz"},
					},
				},
			},
		},
	})
}

func anyTests(name string) ScriptTest {
	tests := []ScriptTestAssertion{
		{
			Query:    `SELECT 3 = %s (ARRAY[1, 2, 3, 4, 5]);`,
			Expected: []sql.Row{{"t"}},
		},
		{
			Query:    `SELECT 3 = %s (ARRAY[1, 2, 4, 5]);`,
			Expected: []sql.Row{{"f"}},
		},
		{
			Query:    `SELECT 'a' = %s (ARRAY['c', 'a', 't']);`,
			Expected: []sql.Row{{"t"}},
		},
		{
			Query:    `SELECT 'a' = %s (ARRAY['c', 'at', 't']);`,
			Expected: []sql.Row{{"f"}},
		},
		{
			Query:    `SELECT 3 = %s (ARRAY[1.0, 2.1, 3.0, 5]);`,
			Expected: []sql.Row{{"t"}},
		},
		{
			Query:    `SELECT 6 > %s (ARRAY[1, 2, 3, 4, 5]);`,
			Expected: []sql.Row{{"t"}},
		},
		{
			Query:    `SELECT 6 < %s (ARRAY[1, 2, 3, 4, 5]);`,
			Expected: []sql.Row{{"f"}},
		},
		{
			Query:    `SELECT 6 <= %s (ARRAY[1, 2, 3, 4, 5]);`,
			Expected: []sql.Row{{"f"}},
		},
		{
			Query:    `SELECT 6 >= %s (ARRAY[1, 2, 3, 6, 5]);`,
			Expected: []sql.Row{{"t"}},
		},
		{
			Query:    `SELECT * FROM test WHERE id = %s(ARRAY[2, 3, 4, 5]);`,
			Expected: []sql.Row{{int32(3)}, {int32(2)}},
		},
		{
			Query:    `SELECT * FROM test WHERE id = %s(ARRAY[4, 3, 2, 1, 0]);`,
			Expected: []sql.Row{{int32(1)}, {int32(3)}, {int32(2)}},
		},
		{
			Query:    `SELECT * FROM test WHERE id = %s(ARRAY[4, 5, 6]);`,
			Expected: []sql.Row{},
		},
		{
			Query: `SELECT id FROM test3 WHERE 4 = %s(carr);`,
			Expected: []sql.Row{
				{int32(2)},
			},
		},
		{
			Skip:     true,
			Query:    `SELECT * FROM test2 WHERE test_id = %s(SELECT * FROM test WHERE id = 2);`,
			Expected: []sql.Row{{int32(3), int32(2), "baz"}},
		},
		{
			Skip:     true,
			Query:    `SELECT * FROM test2 WHERE test_id = %s(SELECT * FROM test WHERE id = 10);`,
			Expected: []sql.Row{},
		},
		{
			Skip:     true,
			Query:    `SELECT * FROM test2 WHERE test_id = %s(SELECT * FROM test WHERE id > 1) AND txt = 'baz';`,
			Expected: []sql.Row{{int32(3), int32(2), "baz"}},
		},
		{
			Skip:  true, // TODO: Panics in EvalMultiple when >1 row matches
			Query: `SELECT * FROM test2 WHERE test_id > %s(SELECT * FROM test);`,
			Expected: []sql.Row{
				{int32(2), int32(10), "bar"},
				{int32(3), int32(2), "baz"},
			},
		},
		{
			Skip:  true, // TODO: Panics in EvalMultiple when >1 row matches
			Query: `SELECT * FROM test2 WHERE test_id = %s(SELECT * FROM test WHERE id > 0);`,
			Expected: []sql.Row{
				{int32(1), int32(1), "foo"},
				{int32(3), int32(2), "baz"},
			},
		},
		{
			Query: `SELECT "ns"."nspname" AS "table_schema",
       "t"."relname" AS "table_name",
       "cnst"."conname" AS "constraint_name",
       pg_get_constraintdef("cnst"."oid") AS "expression",
       CASE "cnst"."contype" 
           WHEN 'p' THEN 'PRIMARY'
           WHEN 'u' THEN 'UNIQUE'
           WHEN 'c' THEN 'CHECK'
           WHEN 'x' THEN 'EXCLUDE'
           END AS "constraint_type", 
    "a"."attname" AS "column_name" 
FROM "pg_catalog"."pg_constraint" "cnst" 
    INNER JOIN "pg_catalog"."pg_class" "t" ON "t"."oid" = "cnst"."conrelid"
    INNER JOIN "pg_catalog"."pg_namespace" "ns" ON "ns"."oid" = "cnst"."connamespace"
    LEFT JOIN "pg_catalog"."pg_attribute" "a" ON "a"."attrelid" = "cnst"."conrelid" AND "a"."attnum" = %s ("cnst"."conkey")
WHERE "t"."relkind" IN ('r', 'p') AND (("ns"."nspname" = 'public' AND "t"."relname" = 'test2'));`,
			Expected: []sql.Row{
				{"public", "test2", "test2_pkey", "PRIMARY KEY (id)", "PRIMARY", "id"},
			},
		},
	}

	formattedTests := make([]ScriptTestAssertion, len(tests))
	for i, test := range tests {
		formattedTests[i] = ScriptTestAssertion{
			Query:       fmt.Sprintf(test.Query, name),
			Skip:        test.Skip,
			Expected:    test.Expected,
			ExpectedErr: test.ExpectedErr,
		}
	}

	return ScriptTest{
		Name: name,
		SetUpScript: []string{
			`CREATE TABLE test (id INT);`,
			`INSERT INTO test VALUES (1), (3), (2);`,

			`CREATE TABLE test2 (id INT PRIMARY KEY, test_id INT, txt text);`,
			`INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');`,

			`CREATE TABLE test3 (id INT PRIMARY KEY, carr smallint[]);`,
			`INSERT INTO test3 VALUES (1, ARRAY[1, 2, 3]), (2, ARRAY[4, 5, 6]);`,
		},
		Assertions: formattedTests,
	}
}

func TestSubqueries(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Subselect",
			SetUpScript: []string{
				`CREATE TABLE test (id INT);`,
				`INSERT INTO test VALUES (1), (3), (2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM test WHERE id = (SELECT 2);`,
					Expected: []sql.Row{
						{int32(2)},
					},
				},
				{
					Query: `SELECT *, (SELECT id from test where id = 2) FROM test order by id;`,
					Expected: []sql.Row{
						{1, 2},
						{2, 2},
						{3, 2},
					},
				},
				{
					Query: `SELECT *, (SELECT id from test t2 where t2.id = test.id) FROM test order by id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
						{3, 3},
					},
				},
			},
		},
		{
			Name: "IN",
			SetUpScript: []string{
				`CREATE TABLE test (id INT);`,
				`INSERT INTO test VALUES (1), (3), (2);`,

				`CREATE TABLE test2 (id INT, test_id INT, txt text);`,
				`INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM test WHERE id IN (SELECT * FROM test WHERE id = 2);`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT * FROM test WHERE id IN (SELECT id FROM test WHERE id = 3);`,
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    `SELECT * FROM test WHERE id IN (SELECT * FROM test WHERE id > 0);`,
					Expected: []sql.Row{{int32(1)}, {int32(3)}, {int32(2)}},
				},
				{
					Query:    `SELECT * FROM test2 WHERE test_id IN (SELECT * FROM test WHERE id = 2);`,
					Expected: []sql.Row{{int32(3), int32(2), "baz"}},
				},
				{
					Query: `SELECT * FROM test2 WHERE test_id IN (SELECT * FROM test WHERE id > 0);`,
					Expected: []sql.Row{
						{int32(1), int32(1), "foo"},
						{int32(3), int32(2), "baz"},
					},
				},
				{
					Query: `SELECT id FROM test2 WHERE (2, 10) IN (SELECT id, test_id FROM test2 WHERE id > 0);`,
					Skip:  true, // won't pass until we have a doltgres tuple type to match against for equality funcs
					Expected: []sql.Row{
						{1}, {2}, {3},
					},
				},
				{
					Query: `SELECT id FROM test2 WHERE (id, test_id) IN (SELECT id, test_id FROM test2 WHERE id > 0);`,
					Skip:  true, // won't pass until we have a doltgres tuple type to match against for equality funcs
					Expected: []sql.Row{
						{2},
					},
				},
			},
		},
	})
}
