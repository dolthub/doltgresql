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

// TestJsonbSet verifies PostgreSQL-compatible JSONB mutation and numeric behavior.
func TestJsonbSet(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "json preserves exact numbers outside jsonb numeric range",
			SetUpScript: []string{
				`CREATE TABLE json_number_text_test (id INT PRIMARY KEY, value JSON);`,
				`INSERT INTO json_number_text_test VALUES (1, '{"n":123456789012345678901234567890.123456789}'), (2, '{"n":1e3000000000}');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT ('123456789012345678901234567890.123456789'::json)::text, ('1e131072'::json)::text, ('1e3000000000'::json)::text;`,
					Expected: []sql.Row{{`123456789012345678901234567890.123456789`, `1e131072`, `1e3000000000`}},
				},
				{
					Query:    `SELECT value ->> 'n' FROM json_number_text_test ORDER BY id;`,
					Expected: []sql.Row{{`123456789012345678901234567890.123456789`}, {`1e3000000000`}},
				},
			},
		},
		{
			Name: "jsonb_set objects and paths",
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT '1e308'::jsonb IS NOT NULL, '1e-16383'::jsonb IS NOT NULL, '0e1000000000'::jsonb IS NOT NULL;`,
					ExpectedRaw: [][][]byte{{{1}, {1}, {1}}},
				},
				{
					Query:           `SELECT '1e131072'::jsonb;`,
					ExpectedErr:     "value overflows numeric format",
					ExpectedErrCode: "22003",
				},
				{
					Query:           `SELECT '{"n":1e-16384}'::jsonb;`,
					ExpectedErr:     "value overflows numeric format",
					ExpectedErrCode: "22003",
				},
				{
					Query:    `SELECT jsonb_set('{"a":1,"b":2}'::jsonb, '{a}', '9'::jsonb);`,
					Expected: []sql.Row{{`{"a": 9, "b": 2}`}},
				},
				{
					Query:    `SELECT jsonb_set('{"a":1}'::jsonb, '{b}', '2'::jsonb);`,
					Expected: []sql.Row{{`{"a": 1, "b": 2}`}},
				},
				{
					Query:    `SELECT jsonb_set('{"a":{"b":1}}'::jsonb, '{a,b}', '2'::jsonb);`,
					Expected: []sql.Row{{`{"a": {"b": 2}}`}},
				},
				{
					Query:    `SELECT jsonb_set('{"a":1}'::jsonb, '{b}', '2'::jsonb, true);`,
					Expected: []sql.Row{{`{"a": 1, "b": 2}`}},
				},
				{
					Query:    `SELECT jsonb_set('{"a":1}'::jsonb, '{b}', '2'::jsonb, false);`,
					Expected: []sql.Row{{`{"a": 1}`}},
				},
				{
					Query:    `SELECT jsonb_set('{"a":1}'::jsonb, '{b,c}', '2'::jsonb, true);`,
					Expected: []sql.Row{{`{"a": 1}`}},
				},
				{
					Query:    `SELECT jsonb_set('{"a":1}'::jsonb, '{a,b}', '2'::jsonb);`,
					Expected: []sql.Row{{`{"a": 1}`}},
				},
				{
					Query: `WITH input(doc) AS (VALUES ('{"a":{"b":1},"c":3}'::jsonb))
						SELECT doc, jsonb_set(doc, '{a,b}', '2'::jsonb), doc FROM input;`,
					Expected: []sql.Row{{
						`{"a": {"b": 1}, "c": 3}`,
						`{"a": {"b": 2}, "c": 3}`,
						`{"a": {"b": 1}, "c": 3}`,
					}},
				},
			},
		},
		{
			Name: "jsonb_set arrays",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, '{1}', '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 9, 2]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, '{-1}', '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 1, 9]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, '{9}', '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 1, 2, 9]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, '{-9}', '9'::jsonb);`,
					Expected: []sql.Row{{`[9, 0, 1, 2]`}},
				},
				{
					Query:    `SELECT jsonb_set('[{"a":1}]'::jsonb, '{0,a}', '2'::jsonb);`,
					Expected: []sql.Row{{`[{"a": 2}]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, '{3}', '9'::jsonb, false);`,
					Expected: []sql.Row{{`[0, 1, 2]`}},
				},
				{
					Query:    `SELECT jsonb_set('[]'::jsonb, '{0}', '1'::jsonb);`,
					Expected: []sql.Row{{`[1]`}},
				},
				{
					Query:    `SELECT jsonb_set('[]'::jsonb, '{-1}', '1'::jsonb);`,
					Expected: []sql.Row{{`[1]`}},
				},
			},
		},
		{
			Name: "jsonb_set array index whitespace",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, ARRAY[' 1'], '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 9, 2]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, ARRAY[E'\t1'], '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 9, 2]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, ARRAY[E'\n1'], '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 9, 2]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, ARRAY[E'\r1'], '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 9, 2]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, ARRAY[E'\f1'], '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 9, 2]`}},
				},
				{
					Query:    `SELECT jsonb_set('[0,1,2]'::jsonb, ARRAY[chr(11) || '1'], '9'::jsonb);`,
					Expected: []sql.Row{{`[0, 9, 2]`}},
				},
				{
					Query:       `SELECT jsonb_set('[0,1,2]'::jsonb, ARRAY['1 '], '9'::jsonb);`,
					ExpectedErr: `path element at position 1 is not an integer: "1 "`,
				},
				{
					// The path starts with U+00A0 NO-BREAK SPACE, which is not C-locale whitespace.
					Query:       `SELECT jsonb_set('[0,1,2]'::jsonb, ARRAY[' 1'], '9'::jsonb);`,
					ExpectedErr: "path element at position 1 is not an integer",
				},
			},
		},
		{
			Name: "jsonb_set empty, null, and scalar inputs",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT jsonb_set('{"a":1}'::jsonb, '{}'::text[], '2'::jsonb);`,
					Expected: []sql.Row{{`{"a": 1}`}},
				},
				{
					Query:    `SELECT jsonb_set('{}'::jsonb, '{a}', '1'::jsonb);`,
					Expected: []sql.Row{{`{"a": 1}`}},
				},
				{
					Query:    `SELECT jsonb_set('{"a":1}'::jsonb, '{a}', 'null'::jsonb);`,
					Expected: []sql.Row{{`{"a": null}`}},
				},
				{
					Query:    `SELECT jsonb_set(NULL::jsonb, '{a}', '1'::jsonb);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT jsonb_set('{}'::jsonb, NULL::text[], '1'::jsonb);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT jsonb_set('{}'::jsonb, '{a}', NULL::jsonb);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT jsonb_set('{}'::jsonb, '{a}', '1'::jsonb, NULL::boolean);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:       `SELECT jsonb_set('{}'::jsonb, ARRAY[NULL]::text[], '1'::jsonb);`,
					ExpectedErr: "path element at position 1 is null",
				},
				{
					Query:       `SELECT jsonb_set('1'::jsonb, '{a}', '2'::jsonb);`,
					ExpectedErr: "cannot set path in scalar",
				},
				{
					Query:       `SELECT jsonb_set('null'::jsonb, '{a}', '2'::jsonb);`,
					ExpectedErr: "cannot set path in scalar",
				},
				{
					Query:       `SELECT jsonb_set('1'::jsonb, '{}'::text[], '2'::jsonb);`,
					ExpectedErr: "cannot set path in scalar",
				},
			},
		},
		{
			Name: "jsonb_set validation and canonical output",
			SetUpScript: []string{
				`CREATE TABLE jsonb_precision_test (id INT PRIMARY KEY, value JSONB);`,
				`INSERT INTO jsonb_precision_test VALUES (1, '{"n":0}');`,
				`INSERT INTO jsonb_precision_test VALUES (2, '{"value":123456789012345678901234567890.123456789}');`,
				`UPDATE jsonb_precision_test SET value = jsonb_set(value, '{other}', 'true'::jsonb) WHERE id = 2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT '123456789012345678901234567890'::jsonb, '1234567890.12345678901234567890'::jsonb;`,
					ExpectedRaw: [][][]byte{{
						[]byte(`123456789012345678901234567890`),
						[]byte(`1234567890.12345678901234567890`),
					}},
				},
				{
					Query: `SELECT jsonb_set(value, '{n}', '1234567890.12345678901234567890'::jsonb) FROM jsonb_precision_test WHERE id = 1;`,
					ExpectedRaw: [][][]byte{{
						[]byte(`{"n": 1234567890.12345678901234567890}`),
					}},
				},
				{
					Query: `SELECT value FROM jsonb_precision_test WHERE id = 2;`,
					ExpectedRaw: [][][]byte{{
						[]byte(`{"other": true, "value": 123456789012345678901234567890.123456789}`),
					}},
				},
				{
					Query: `SELECT value ->> 'value', value #>> '{value}', (value -> 'value')::numeric = 123456789012345678901234567890.123456789::numeric, value -> 'value' = '123456789012345678901234567890.123456789'::jsonb, value -> 'value' < '123456789012345678901234567890.123456790'::jsonb FROM jsonb_precision_test WHERE id = 2;`,
					ExpectedRaw: [][][]byte{{
						[]byte(`123456789012345678901234567890.123456789`),
						[]byte(`123456789012345678901234567890.123456789`),
						{1},
						{1},
						{1},
					}},
				},
				{
					Query: `SELECT value || '{"third": 123456789012345678901234567890.123456788}'::jsonb FROM jsonb_precision_test WHERE id = 2;`,
					ExpectedRaw: [][][]byte{{
						[]byte(`{"other": true, "third": 123456789012345678901234567890.123456788, "value": 123456789012345678901234567890.123456789}`),
					}},
				},
				{
					Query:       `SELECT '1.0'::jsonb = '1'::jsonb, '9007199254740992.1'::jsonb = '9007199254740992.2'::jsonb, '9007199254740992.1'::jsonb < '9007199254740992.2'::jsonb;`,
					ExpectedRaw: [][][]byte{{{1}, {0}, {1}}},
				},
				{
					Query: `SELECT jsonb_set('{"n":0}'::jsonb, '{n}', '123456789012345678901234567890'::jsonb), jsonb_set('{"n":0}'::jsonb, '{n}', '1234567890.12345678901234567890'::jsonb);`,
					ExpectedRaw: [][][]byte{{
						[]byte(`{"n": 123456789012345678901234567890}`),
						[]byte(`{"n": 1234567890.12345678901234567890}`),
					}},
				},
				{
					Query:       `SELECT jsonb_set('[]'::jsonb, '{x}', '1'::jsonb);`,
					ExpectedErr: `path element at position 1 is not an integer: "x"`,
				},
				{
					Query:       `SELECT jsonb_set('[0]'::jsonb, '{999999999999999999999999}', '1'::jsonb);`,
					ExpectedErr: `path element at position 1 is not an integer: "999999999999999999999999"`,
				},
				{
					Query:       `SELECT jsonb_set('[0]'::jsonb, '{2147483648}', '1'::jsonb);`,
					ExpectedErr: `path element at position 1 is not an integer: "2147483648"`,
				},
				{
					Query:    `SELECT jsonb_set('{"a.b":{"q\"uote":1}}'::jsonb, ARRAY['a.b','q"uote'], '"x\\y"'::jsonb);`,
					Expected: []sql.Row{{`{"a.b": {"q\"uote": "x\\y"}}`}},
				},
				{
					Query:    `SELECT jsonb_set('{"b":0,"a":1,"a":2}'::jsonb, '{c}', '{"z":1,"z":2,"a":3}'::jsonb);`,
					Expected: []sql.Row{{`{"a": 2, "b": 0, "c": {"a": 3, "z": 2}}`}},
				},
			},
		},
	})
}
