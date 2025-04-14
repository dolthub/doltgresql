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

func TestIn(t *testing.T) {
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
				{
					Query:    `SELECT 4 IN (null, 1, 2, 3);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT 4 IN (null, 1, 2, 3, 4);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT NULL IN (null, 1, 2, 3);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT 4 IN (1, 2, 3, null::int4);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT 4 IN (1, 2, 3);`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 4 IN (1, 2, 3, 4);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT concat('a', 'b') in ('a', 'b', 'ab');`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT concat('a', 'b') in ('a', 'b');`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT concat('a', 'b') in ('a', NULL, 'b');`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT concat('a', NULL) in ('a', 'b', 'ab');`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT concat('a', NULL) in ('a', NULL);`,
					Expected: []sql.Row{{nil}},
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
			Query:    `SELECT * FROM test2 WHERE test_id = %s(SELECT * FROM test WHERE id = 2);`,
			Expected: []sql.Row{{int32(3), int32(2), "baz"}},
		},
		{
			Query:    `SELECT * FROM test2 WHERE test_id = %s(SELECT * FROM test WHERE id = 10);`,
			Expected: []sql.Row{},
		},
		{
			Query:    `SELECT * FROM test2 WHERE test_id = %s(SELECT * FROM test WHERE id > 1) AND txt = 'baz';`,
			Expected: []sql.Row{{int32(3), int32(2), "baz"}},
		},
		{
			Query: `SELECT * FROM test2 WHERE test_id > %s(SELECT * FROM test);`,
			Expected: []sql.Row{
				{int32(2), int32(10), "bar"},
				{int32(3), int32(2), "baz"},
			},
		},
		{
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

func TestBinaryLogic(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "AND",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 1 = 1 AND 2 = 2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT (1 = 1 AND 2 = 2) AND (false);`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT (1 > 1 AND 2 = 2);`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT (1 = 1 AND 2 = 2) AND (false);`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT (1 = 1 AND 2 = 2) AND (true);`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "OR",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 1 = 1 OR 2 = 2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT (1 = 1 AND 2 = 2) OR (false);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT (1 > 1 OR 2 = 2);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT (1 > 1 OR 2 > 2);`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT (1 > 1 OR 2 > 2) OR (true);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT (1 = 1 AND 2 = 2) OR (true);`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
	})
}

func TestSubscript(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "array literal",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT ARRAY[1, 2, 3][1];`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT (ARRAY[1, 2, 3])[3];`,
					Expected: []sql.Row{{3}},
				},
				{
					Query:    `SELECT (ARRAY[1, 2, 3])[1+1];`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT ARRAY[1, 2, 3][0];`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT ARRAY[1, 2, 3][4];`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT ARRAY[1, 2, 3][null];`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT ARRAY[1, 2, 3][1:3];`,
					ExpectedErr: "not yet supported",
				},
				{
					Query:       `SELECT ARRAY[1, 2, 3]['abc'];`,
					ExpectedErr: "integer: unhandled type: string",
				},
			},
		},
	})
}
