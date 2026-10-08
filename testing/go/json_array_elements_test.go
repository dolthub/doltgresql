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
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TestJsonArrayElements covers the SELECT-list and FROM-clause forms requested in issue #3498.
func TestJsonArrayElements(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "call forms and result metadata",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT pg_catalog.json_array_elements('[1, 2, 3]'::json);`,
					Expected:         []sql.Row{{"1"}, {"2"}, {"3"}},
					ExpectedColNames: []string{"json_array_elements"},
					ExpectedColTypes: []id.Type{pgtypes.Json.ID},
				},
				{
					Query:            `SELECT * FROM pg_catalog.json_array_elements('[1, 2, 3]'::json);`,
					Expected:         []sql.Row{{"1"}, {"2"}, {"3"}},
					ExpectedColNames: []string{"value"},
					ExpectedColTypes: []id.Type{pgtypes.Json.ID},
				},
				{
					Query:    `SELECT json_array_elements('[4,5]');`,
					Expected: []sql.Row{{"4"}, {"5"}},
				},
				{
					Query:    `SELECT value FROM json_array_elements('[4,5]');`,
					Expected: []sql.Row{{"4"}, {"5"}},
				},
				{
					Query:            `SELECT json_array_elements('[1]'::json) AS elem;`,
					Expected:         []sql.Row{{"1"}},
					ExpectedColNames: []string{"elem"},
					ExpectedColTypes: []id.Type{pgtypes.Json.ID},
				},
				{
					Query:    `SELECT e.value FROM pg_catalog.json_array_elements('[1]'::json) AS e;`,
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:            `SELECT e FROM pg_catalog.json_array_elements('[1]'::json) AS e;`,
					Expected:         []sql.Row{{"1"}},
					ExpectedColNames: []string{"e"},
					ExpectedColTypes: []id.Type{pgtypes.Json.ID},
				},
				{
					Query:    `SELECT e->>'name' FROM json_array_elements('[{"name":"first"}]'::json) AS e;`,
					Expected: []sql.Row{{"first"}},
				},
				{
					Query: `SELECT elem->>'name' FROM
						(SELECT e AS elem FROM json_array_elements('[{"name":"first"}]'::json) AS e) AS expanded;`,
					Expected: []sql.Row{{"first"}},
				},
				{
					Query:            `SELECT e.elem FROM json_array_elements('[1]'::json) AS e(elem);`,
					Expected:         []sql.Row{{"1"}},
					ExpectedColNames: []string{"elem"},
					ExpectedColTypes: []id.Type{pgtypes.Json.ID},
				},
				{
					Query:    `SELECT pg_typeof(value) FROM json_array_elements('[1,null,{}]'::json);`,
					Expected: []sql.Row{{"json"}, {"json"}, {"json"}},
				},
			},
		},
		{
			Name: "top-level elements preserve JSON values",
			Assertions: []ScriptTestAssertion{
				{
					// pgx decodes JSON null as nil; the SQL assertions below distinguish it from SQL NULL.
					Query: `SELECT json_array_elements('[1,-2,1.5,true,false,"hello","",null,[2,false],{"a":1}]'::json);`,
					Expected: []sql.Row{
						{"1"}, {"-2"}, {"1.5"}, {"true"}, {"false"}, {`"hello"`}, {`""`}, {nil}, {`[2,false]`}, {`{"a":1}`},
					},
				},
				{
					Query: `SELECT * FROM json_array_elements('[1,-2,1.5,true,false,"hello","",null,[2,false],{"a":1}]'::json);`,
					Expected: []sql.Row{
						{"1"}, {"-2"}, {"1.5"}, {"true"}, {"false"}, {`"hello"`}, {`""`}, {nil}, {`[2,false]`}, {`{"a":1}`},
					},
				},
				{
					// Ordinality checks the original order and preserves duplicate and nested elements.
					Query: `SELECT value::text, ordinality FROM json_array_elements('[3,1,3,[],{},null]'::json)
						WITH ORDINALITY ORDER BY ordinality;`,
					Expected: []sql.Row{{"3", int64(1)}, {"1", int64(2)}, {"3", int64(3)}, {"[]", int64(4)}, {"{}", int64(5)}, {"null", int64(6)}},
				},
				{
					Query:    `SELECT json_array_elements('["a\"b","line\nbreak","雪"]'::json);`,
					Expected: []sql.Row{{`"a\"b"`}, {`"line\nbreak"`}, {`"雪"`}},
				},
				{
					// A JSON null is one non-NULL JSON value, rather than an empty set or SQL NULL.
					Query:    `SELECT value::text, value IS NULL, json_typeof(value) FROM json_array_elements('[null]'::json);`,
					Expected: []sql.Row{{"null", "f", "null"}},
				},
				{
					Query: `SELECT elem::text, elem IS NULL FROM
						(SELECT json_array_elements('[null]'::json) AS elem) AS expanded;`,
					Expected: []sql.Row{{"null", "f"}},
				},
				{
					Query:    `SELECT value->>'name' FROM json_array_elements('[{"name":"first"},{"name":"second"}]'::json);`,
					Expected: []sql.Row{{"first"}, {"second"}},
				},
			},
		},
		{
			Name: "empty arrays and SQL NULL produce no rows",
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT json_array_elements('[]'::json);`, Expected: []sql.Row{}},
				{Query: `SELECT * FROM json_array_elements('[]'::json);`, Expected: []sql.Row{}},
				{Query: `SELECT json_array_elements(NULL::json);`, Expected: []sql.Row{}},
				{Query: `SELECT * FROM json_array_elements(NULL::json);`, Expected: []sql.Row{}},
				{Query: `SELECT 42, json_array_elements('[]'::json);`, Expected: []sql.Row{}},
				{Query: `SELECT 42, json_array_elements(NULL::json);`, Expected: []sql.Row{}},
			},
		},
		{
			Name: "multiple SELECT-list calls expand together",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT json_array_elements('[1,2]'::json) AS a,
						json_array_elements('[3]'::json) AS b;`,
					Expected: []sql.Row{{"1", "3"}, {"2", nil}},
				},
			},
		},
	})
}

func TestJsonArrayElementsErrors(t *testing.T) {
	var scripts []ScriptTest
	for _, input := range []struct {
		name, json, message string
	}{
		{"object", `{"a":1}`, "cannot call json_array_elements on a non-array"},
		{"number", `1`, "cannot call json_array_elements on a scalar"},
		{"string", `"hello"`, "cannot call json_array_elements on a scalar"},
		{"boolean", `true`, "cannot call json_array_elements on a scalar"},
		{"JSON null", `null`, "cannot call json_array_elements on a scalar"},
	} {
		scripts = append(scripts, ScriptTest{
			Name: input.name,
			Assertions: []ScriptTestAssertion{
				{
					Query:           fmt.Sprintf(`SELECT pg_catalog.json_array_elements('%s'::json);`, input.json),
					ExpectedErr:     input.message,
					ExpectedErrCode: "22023",
				},
				{
					Query:           fmt.Sprintf(`SELECT * FROM pg_catalog.json_array_elements('%s'::json);`, input.json),
					ExpectedErr:     input.message,
					ExpectedErrCode: "22023",
				},
			},
		})
	}
	RunScripts(t, scripts)
}

func TestJsonArrayElementsStoredValues(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "correlated expansion of stored arrays",
			SetUpScript: []string{
				`CREATE TABLE arrays (id INT PRIMARY KEY, doc JSON);`,
				`INSERT INTO arrays VALUES (1,'[2,1]'), (2,'[3]'), (3,'[]'), (4,NULL), (5,'[null]');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT id, elem::text FROM
						(SELECT id, json_array_elements(doc) AS elem FROM arrays) AS expanded
						ORDER BY id, elem::text;`,
					Expected: []sql.Row{{1, "1"}, {1, "2"}, {2, "3"}, {5, "null"}},
				},
				{
					Query: `SELECT a.id, e.value::text, e.ordinality FROM arrays a,
						pg_catalog.json_array_elements(a.doc) WITH ORDINALITY AS e ORDER BY a.id, e.ordinality;`,
					Expected: []sql.Row{{1, "2", int64(1)}, {1, "1", int64(2)}, {2, "3", int64(1)}, {5, "null", int64(1)}},
				},
				{
					Query: `SELECT a.id, e.value::text FROM arrays a
						CROSS JOIN LATERAL json_array_elements(a.doc) AS e ORDER BY a.id, e.value::text;`,
					Expected: []sql.Row{{1, "1"}, {1, "2"}, {2, "3"}, {5, "null"}},
				},
				{
					Query: `SELECT a.id, e.value::text, e.value IS NULL FROM arrays a
						LEFT JOIN LATERAL json_array_elements(a.doc) AS e ON true ORDER BY a.id, e.value::text;`,
					Expected: []sql.Row{{1, "1", "f"}, {1, "2", "f"}, {2, "3", "f"}, {3, nil, "t"}, {4, nil, "t"}, {5, "null", "f"}},
				},
			},
		},
		{
			Name: "large stored JSON document",
			SetUpScript: []string{
				`CREATE TABLE bigarray (doc JSON);`,
				`INSERT INTO bigarray VALUES ('` + makeLargeJSONArray(80) + `'::json);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Read every element of a document that exceeds the inline storage threshold.
					Query: `SELECT count(*), min((e.value->>'id')::int), max((e.value->>'id')::int)
						FROM bigarray b, json_array_elements(b.doc) AS e;`,
					Expected: []sql.Row{{int64(80), 0, 79}},
				},
				{
					Query: `SELECT elem->>'label', elem->'payload' FROM
						(SELECT json_array_elements(doc) AS elem FROM bigarray) AS expanded
						WHERE (elem->>'id')::int = 79;`,
					Expected: []sql.Row{{"row_0079", `["a","b","c","d","e"]`}},
				},
			},
		},
	})
}

// TestJsonArrayElementsPostgREST exercises the target_entries CTE from issue #3498.
func TestJsonArrayElementsPostgREST(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "PostgREST view target-list expansion",
			SetUpScript: []string{
				`CREATE TABLE transform_json (view_id INT PRIMARY KEY, view_schema_id INT, view_schema TEXT, view_name TEXT, view_definition JSON);`,
				`INSERT INTO transform_json VALUES
					(1,2200,'public','view_one','[{"targetList":[{"resno":1,"resname":"id"},{"resno":2,"resname":"name"}]}]'),
					(2,2200,'public','view_two','[{"targetList":[{"resno":1,"resname":"other"}]}]'),
					(3,2200,'public','view_empty','[{"targetList":[]}]'),
					(4,2200,'public','view_missing','[{}]'),
					(5,2200,'public','view_null',NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `WITH target_entries AS (
						SELECT view_id, view_schema_id, view_schema, view_name,
							pg_catalog.json_array_elements(view_definition->0->'targetList') AS entry
						FROM transform_json
					)
					SELECT view_id, view_schema_id, view_schema, view_name, entry->>'resno', entry->>'resname'
					FROM target_entries ORDER BY view_id, entry->>'resno';`,
					Expected: []sql.Row{
						{1, 2200, "public", "view_one", "1", "id"},
						{1, 2200, "public", "view_one", "2", "name"},
						{2, 2200, "public", "view_two", "1", "other"},
					},
				},
			},
		},
	})
}
