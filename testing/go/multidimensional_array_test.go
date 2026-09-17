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

func TestMultidimensionalArrays(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "constructors and literals",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ARRAY[[1,2],[3,4]];",
					Expected: []sql.Row{{"{{1,2},{3,4}}"}},
				},
				{
					Query:    "SELECT ARRAY[ARRAY[1,2],ARRAY[3,4]]::text[];",
					Expected: []sql.Row{{"{{1,2},{3,4}}"}},
				},
				{
					Query:    "SELECT '{{1,2},{3,4}}'::int[];",
					Expected: []sql.Row{{"{{1,2},{3,4}}"}},
				},
				{
					Query:    "SELECT '{ {1, 2} , {3,4} }'::int[];",
					Expected: []sql.Row{{"{{1,2},{3,4}}"}},
				},
				{
					Query:    "SELECT '{{{1},{2}},{{3},{4}}}'::int[];",
					Expected: []sql.Row{{"{{{1},{2}},{{3},{4}}}"}},
				},
				{
					Query:    `SELECT '{{"a b",c},{d,NULL}}'::text[];`,
					Expected: []sql.Row{{`{{"a b",c},{d,NULL}}`}},
				},
				{
					Query:    "SELECT ARRAY[['a b','c'],['d',NULL]];",
					Expected: []sql.Row{{`{{"a b",c},{d,NULL}}`}},
				},
				{
					Query:    "SELECT ARRAY[ARRAY[]::int[]];",
					Expected: []sql.Row{{"{}"}},
				},
				{
					Query:       "SELECT ARRAY[[1,2],[3]];",
					ExpectedErr: "multidimensional arrays must have array expressions with matching dimensions",
				},
				{
					Query:       "SELECT ARRAY[NULL::int[], ARRAY[1]];",
					ExpectedErr: "multidimensional arrays must have array expressions with matching dimensions",
				},
				{
					Query:       "SELECT '{{1,2},{3}}'::int[];",
					ExpectedErr: `malformed array literal: "{{1,2},{3}}"`,
				},
				{
					Query:       "SELECT '{1,{2}}'::int[];",
					ExpectedErr: `malformed array literal: "{1,{2}}"`,
				},
				{
					Query:       "SELECT '{{}}'::int[];",
					ExpectedErr: `malformed array literal: "{{}}"`,
				},
				{
					Query:       "SELECT '{a,}'::text[];",
					ExpectedErr: `malformed array literal: "{a,}"`,
				},
				{
					Query:    "SELECT pg_typeof(ARRAY[[1,2]]);",
					Expected: []sql.Row{{"integer[]"}},
				},
				{
					Query:    "SELECT ARRAY[[1.5,2],[3,4]]::numeric(3,1)[];",
					Expected: []sql.Row{{"{{1.5,2.0},{3.0,4.0}}"}},
				},
				{
					Query:    "SELECT ARRAY[[true,false]];",
					Expected: []sql.Row{{"{{t,f}}"}},
				},
			},
		},
		{
			Name: "table storage",
			SetUpScript: []string{
				"CREATE TABLE t (pk INT PRIMARY KEY, v INT[][], w TEXT[3][4]);",
				"INSERT INTO t VALUES (1, ARRAY[[1,2],[3,4]], ARRAY[['a','b']]), (2, ARRAY[5,6], '{{{x}}}'), (3, '{}', NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM t ORDER BY pk;",
					Expected: []sql.Row{
						{1, "{{1,2},{3,4}}", "{{a,b}}"},
						{2, "{5,6}", "{{{x}}}"},
						{3, "{}", nil},
					},
				},
				{
					Query: "UPDATE t SET v = ARRAY[[[1,2,3]]] WHERE pk = 2;",
				},
				{
					Query: "SELECT pk, v, array_ndims(v), array_dims(v) FROM t ORDER BY pk;",
					Expected: []sql.Row{
						{1, "{{1,2},{3,4}}", 2, "[1:2][1:2]"},
						{2, "{{{1,2,3}}}", 3, "[1:1][1:1][1:3]"},
						{3, "{}", nil, nil},
					},
				},
				{
					Query:    "SELECT pk FROM t WHERE v = ARRAY[[1,2],[3,4]];",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT pk FROM t WHERE v[2][1] = 3;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT pk FROM t ORDER BY v;",
					Expected: []sql.Row{{3}, {2}, {1}},
				},
				{
					Query:    "SELECT pk FROM t WHERE 3 = ANY(v);",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:       "INSERT INTO t VALUES (4, '{{1,2},{3}}', NULL);",
					ExpectedErr: `malformed array literal: "{{1,2},{3}}"`,
				},
			},
		},
		{
			Name: "subscripts, functions, and operators",
			SetUpScript: []string{
				"CREATE TABLE agg (pk INT PRIMARY KEY, v INT[]);",
				"INSERT INTO agg VALUES (1, ARRAY[1,2]), (2, ARRAY[3,4]), (3, ARRAY[5]), (4, NULL), (5, '{}');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT (ARRAY[[1,2],[3,4]])[1][2], (ARRAY[[1,2],[3,4]])[2][1];",
					Expected: []sql.Row{{2, 3}},
				},
				{
					Query:    "SELECT (ARRAY[[1,2],[3,4]])[1], (ARRAY[[1,2],[3,4]])[1][2][3], (ARRAY[[1,2],[3,4]])[3][1], (ARRAY[1,2])[1][1];",
					Expected: []sql.Row{{nil, nil, nil, nil}},
				},
				{
					Query:    "SELECT array_length(ARRAY[[1,2,3],[4,5,6]], 1), array_length(ARRAY[[1,2,3],[4,5,6]], 2), array_length(ARRAY[[1,2,3],[4,5,6]], 3), array_length(ARRAY[[1,2,3],[4,5,6]], 0);",
					Expected: []sql.Row{{2, 3, nil, nil}},
				},
				{
					Query:    "SELECT array_upper(ARRAY[[1,2,3],[4,5,6]], 2), array_upper(ARRAY[1,2,3,4], 2);",
					Expected: []sql.Row{{3, nil}},
				},
				{
					Query:    "SELECT array_ndims(ARRAY[[1,2],[3,4]]), array_ndims(ARRAY[1]), array_ndims('{}'::int[]);",
					Expected: []sql.Row{{2, 1, nil}},
				},
				{
					Query:    "SELECT array_dims(ARRAY[[1,2,3],[4,5,6]]), array_dims(ARRAY[1]), array_dims('{}'::int[]);",
					Expected: []sql.Row{{"[1:2][1:3]", "[1:1]", nil}},
				},
				{
					Query:    "SELECT generate_subscripts(ARRAY[[1,2,3],[4,5,6]], 2);",
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    "SELECT generate_subscripts(ARRAY[[1,2,3],[4,5,6]], 3);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT unnest(ARRAY[[1,2],[3,4]]);",
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					Query:    "SELECT 3 = ANY(ARRAY[[1,2],[3,4]]), 5 = ANY(ARRAY[[1,2],[3,4]]), 3 = ALL(ARRAY[[3,3],[3,3]]);",
					Expected: []sql.Row{{"t", "f", "t"}},
				},
				{
					Query:    "SELECT array_to_string(ARRAY[[1,2],[3,4]], ',');",
					Expected: []sql.Row{{"1,2,3,4"}},
				},
				{
					Query:    "SELECT array_cat(ARRAY[[1,2],[3,4]], ARRAY[5,6]), array_cat(ARRAY[5,6], ARRAY[[1,2],[3,4]]), array_cat(ARRAY[[1,2],[3,4]], ARRAY[[5,6]]), array_cat(ARRAY[[1,2],[3,4]], '{}'::int[]);",
					Expected: []sql.Row{{"{{1,2},{3,4},{5,6}}", "{{5,6},{1,2},{3,4}}", "{{1,2},{3,4},{5,6}}", "{{1,2},{3,4}}"}},
				},
				{
					Query:    "SELECT ARRAY[[1,2],[3,4]] || ARRAY[5,6];",
					Expected: []sql.Row{{"{{1,2},{3,4},{5,6}}"}},
				},
				{
					Query:       "SELECT array_cat(ARRAY[[1,2],[3,4]], ARRAY[5,6,7]);",
					ExpectedErr: "cannot concatenate incompatible arrays",
				},
				{
					Query:       "SELECT array_append(ARRAY[[1,2],[3,4]], 5);",
					ExpectedErr: "argument must be empty or one-dimensional array",
				},
				{
					Query:       "SELECT array_prepend(5, ARRAY[[1,2],[3,4]]);",
					ExpectedErr: "argument must be empty or one-dimensional array",
				},
				{
					Query:       "SELECT array_position(ARRAY[[1,2],[3,4]], 3);",
					ExpectedErr: "searching for elements in multidimensional arrays is not supported",
				},
				{
					Query:    "SELECT ARRAY[[1,2],[3,4]] = ARRAY[1,2,3,4], ARRAY[[1,2],[3,4]] > ARRAY[1,2,3,4], ARRAY[[1,2],[3,4]] < ARRAY[[1,2,3,4]], ARRAY[[1,2],[3,4]] = ARRAY[[1,2],[3,4]];",
					Expected: []sql.Row{{"f", "t", "f", "t"}},
				},
				{
					Query:    "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk <= 2;",
					Expected: []sql.Row{{"{{1,2},{3,4}}"}},
				},
				{
					Query:       "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk <= 3;",
					ExpectedErr: "cannot accumulate arrays of different dimensionality",
				},
				{
					Query:       "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk IN (1, 4);",
					ExpectedErr: "cannot accumulate null arrays",
				},
				{
					Query:       "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk IN (1, 5);",
					ExpectedErr: "cannot accumulate arrays of different dimensionality",
				},
				{
					Query:       "SELECT array_agg(v ORDER BY pk DESC) FROM agg WHERE pk IN (1, 5);",
					ExpectedErr: "cannot accumulate empty arrays",
				},
				{
					Query:       "SELECT ARRAY(SELECT v FROM agg);",
					ExpectedErr: "cannot accumulate arrays of different dimensionality",
				},
			},
		},
		{
			Name: "arrays of vectors are one-dimensional",
			SetUpScript: []string{
				"CREATE TABLE ov (pk INT PRIMARY KEY, v oidvector[]);",
				`INSERT INTO ov VALUES (1, '{"1 2","3"}');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM ov;",
					Expected: []sql.Row{{1, `{"1 2",3}`}},
				},
				{
					Query:    `SELECT pk FROM ov WHERE v = '{"1 2","3"}';`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT ('{"1 2","3"}'::oidvector[])[1], array_ndims('{"1 2","3"}'::oidvector[]);`,
					Expected: []sql.Row{{"1 2", 1}},
				},
			},
		},
	})
}
