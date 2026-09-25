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
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// makeLargeJSONObject builds a JSONB object literal with the given number of
// keys named k_0000…k_NNNN, each mapped to a small nested object. With 100
// keys the serialized form is roughly 8 KB, which is comfortably above the
// 4 KB threshold that triggers out-of-band storage in Dolt's indexed JSON
// document representation.
func makeLargeJSONObject(numKeys int) string {
	var b strings.Builder
	b.WriteByte('{')
	for i := 0; i < numKeys; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b,
			`"k_%04d":{"name":"value_%04d","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":%d}`,
			i, i, i)
	}
	b.WriteByte('}')
	return b.String()
}

// makeLargeJSONArray builds a JSONB array literal with the given number of
// element objects, each labeled row_0000…row_NNNN. With 80 elements the
// serialized form is roughly 5 KB.
func makeLargeJSONArray(numElems int) string {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < numElems; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b,
			`{"id":%d,"label":"row_%04d","payload":["a","b","c","d","e"]}`,
			i, i)
	}
	b.WriteByte(']')
	return b.String()
}

// makeLargeJSONObjectWithNumericKeys builds a JSONB object large enough to be
// stored as an indexed document, whose "nums" key maps to a nested object with
// numeric string keys. It exercises the extract-path fast path's fallback: a
// numeric path element is first guessed as an array index, which an object
// rejects, so resolution must fall back to treating it as an object key.
func makeLargeJSONObjectWithNumericKeys() string {
	padding := makeLargeJSONObject(100)
	// Splice a numeric-keyed sub-object onto the front of the padding object,
	// dropping the padding's leading '{'.
	return `{"nums":{"0":"zero","1":"one","2":"two"},` + padding[1:]
}

// TestJsonObjectField exercises the `->` operator with a text right-hand side
// against both jsonb and json values (jsonb_object_field / json_object_field),
// plus the `->>` text-returning variants. The optimization path uses
// types.LookupJSONValue against SearchableJSON wrappers, but the semantics
// must match for non-object inputs and special keys as well.
func TestJsonObjectField(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb_object_field returns object value",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '{"a":1,"b":"two"}'::jsonb -> 'a';`,
					Expected: []sql.Row{{`1`}},
				},
				{
					Query:    `SELECT '{"a":1,"b":"two"}'::jsonb -> 'b';`,
					Expected: []sql.Row{{`"two"`}},
				},
				{
					Query:    `SELECT '{"a":1,"b":"two"}'::jsonb -> null;`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT '{"nested":{"x":[1,2,3]}}'::jsonb -> 'nested';`,
					Expected: []sql.Row{{`{"x": [1, 2, 3]}`}},
				},
				{
					// Missing key returns SQL NULL.
					Query:    `SELECT '{"a":1}'::jsonb -> 'missing';`,
					Expected: []sql.Row{{nil}},
				},
				{
					// `->` with a text key on an array returns SQL NULL.
					Query:    `SELECT '[1,2,3]'::jsonb -> 'a';`,
					Expected: []sql.Row{{nil}},
				},
				{
					// `->` with a text key on a scalar returns SQL NULL.
					Query:    `SELECT '42'::jsonb -> 'a';`,
					Expected: []sql.Row{{nil}},
				},
				{
					// Key with a literal dot in it: the optimized lookup
					// builds a quoted-key MySQL JSON path, so the dot must
					// not be treated as a path separator.
					Query:    `SELECT '{"a.b":1, "a":{"b":2}}'::jsonb -> 'a.b';`,
					Expected: []sql.Row{{`1`}},
				},
				{
					// Key containing a literal double-quote, which must be
					// escaped in the constructed MySQL JSON path.
					Query:    `SELECT '{"a\"b":7}'::jsonb -> 'a"b';`,
					Expected: []sql.Row{{`7`}},
				},
			},
		},
		{
			Name: "json_object_field returns object value",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '{"a":1,"b":"two"}'::json -> 'a';`,
					Expected: []sql.Row{{`1`}},
				},
				{
					Query:    `SELECT '{"a":1,"b":"two"}'::json -> 'b';`,
					Expected: []sql.Row{{`"two"`}},
				},
				{
					Query:    `SELECT '{"a":1}'::json -> 'missing';`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT '[1,2,3]'::json -> 'a';`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "jsonb_object_field_text returns object value as text",
			Assertions: []ScriptTestAssertion{
				{
					// `->>` on a string value returns the raw string (no
					// surrounding quotes).
					Query:    `SELECT '{"a":1,"b":"two"}'::jsonb ->> 'b';`,
					Expected: []sql.Row{{`two`}},
				},
				{
					// Numeric value is rendered as its JSON text.
					Query:    `SELECT '{"a":42}'::jsonb ->> 'a';`,
					Expected: []sql.Row{{`42`}},
				},
				{
					// Nested object is rendered as the JSON object text.
					Query:    `SELECT '{"a":{"b":1}}'::jsonb ->> 'a';`,
					Expected: []sql.Row{{`{"b": 1}`}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb ->> 'missing';`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
	})
}

// TestJsonArrayElement exercises the `->` operator with an integer right-hand
// side (jsonb_array_element / json_array_element) and the `->>` text variant.
// The optimized path uses $[N] lookups; negative indices fall back to a
// materialized walk to resolve the absolute index.
func TestJsonArrayElement(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb_array_element returns array element",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '[10,20,30]'::jsonb -> 0;`,
					Expected: []sql.Row{{`10`}},
				},
				{
					Query:    `SELECT '[10,20,30]'::jsonb -> 2;`,
					Expected: []sql.Row{{`30`}},
				},
				{
					// Out-of-range positive index returns SQL NULL.
					Query:    `SELECT '[10,20,30]'::jsonb -> 5;`,
					Expected: []sql.Row{{nil}},
				},
				{
					// Negative indices count from the end.
					Query:    `SELECT '[10,20,30]'::jsonb -> -1;`,
					Expected: []sql.Row{{`30`}},
				},
				{
					Query:    `SELECT '[10,20,30]'::jsonb -> -3;`,
					Expected: []sql.Row{{`10`}},
				},
				{
					// Out-of-range negative index returns SQL NULL.
					Query:    `SELECT '[10,20,30]'::jsonb -> -5;`,
					Expected: []sql.Row{{nil}},
				},
				{
					// Indexing a non-array returns SQL NULL.
					Query:    `SELECT '{"a":1}'::jsonb -> 0;`,
					Expected: []sql.Row{{nil}},
				},
				{
					// Indexing a scalar returns SQL NULL.
					Query:    `SELECT '42'::jsonb -> 0;`,
					Expected: []sql.Row{{nil}},
				},
				{
					// Nested object element survives the lookup with full
					// structure intact.
					Query:    `SELECT '[{"a":1},{"b":2}]'::jsonb -> 1;`,
					Expected: []sql.Row{{`{"b": 2}`}},
				},
			},
		},
		{
			Name: "json_array_element returns array element",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '[10,20,30]'::json -> 1;`,
					Expected: []sql.Row{{`20`}},
				},
				{
					Query:    `SELECT '[10,20,30]'::json -> -1;`,
					Expected: []sql.Row{{`30`}},
				},
				{
					Query:    `SELECT '[10,20,30]'::json -> 99;`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "jsonb_array_element_text returns text representation",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '["alpha","beta"]'::jsonb ->> 0;`,
					Expected: []sql.Row{{`alpha`}},
				},
				{
					Query:    `SELECT '[10,20,30]'::jsonb ->> -1;`,
					Expected: []sql.Row{{`30`}},
				},
				{
					Query:    `SELECT '[10,20,30]'::jsonb ->> 99;`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
	})
}

// TestJsonExtractPath exercises the `#>` operator (jsonb_extract_path /
// json_extract_path) and the text-returning `#>>` variant. The path is a
// text array; each element selects a key on an object or an integer index
// on an array at the current location.
func TestJsonExtractPath(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb_extract_path follows mixed key/index paths",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '{"a":{"b":{"c":1}}}'::jsonb #> '{a,b,c}';`,
					Expected: []sql.Row{{`1`}},
				},
				{
					Query:    `SELECT '{"a":[10,20,30]}'::jsonb #> '{a,1}';`,
					Expected: []sql.Row{{`20`}},
				},
				{
					Query:    `SELECT '{"a":[10,20,30]}'::jsonb #> '{a,-1}';`,
					Expected: []sql.Row{{`30`}},
				},
				{
					Query:    `SELECT '{"a":[10,20]}'::jsonb #> '{a,not-an-int}';`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT '{"a":{"b":1}}'::jsonb #> '{a,missing,c}';`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb #> '{a,b}';`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "jsonb_extract_path_text renders the leaf as text",
			Assertions: []ScriptTestAssertion{
				{
					// String leaf returns the raw string.
					Query:    `SELECT '{"a":{"b":"hello"}}'::jsonb #>> '{a,b}';`,
					Expected: []sql.Row{{`hello`}},
				},
				{
					// Object leaf returns the JSON text of the object.
					Query:    `SELECT '{"a":{"b":{"c":1}}}'::jsonb #>> '{a,b}';`,
					Expected: []sql.Row{{`{"c": 1}`}},
				},
				{
					Query:    `SELECT '{"a":[1,2,3]}'::jsonb #>> '{a,2}';`,
					Expected: []sql.Row{{`3`}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb #>> '{missing}';`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "json_extract_path follows mixed key/index paths",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '{"a":{"b":[10,20]}}'::json #> '{a,b,0}';`,
					Expected: []sql.Row{{`10`}},
				},
				{
					Query:    `SELECT '{"a":1}'::json #> '{missing}';`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "jsonb_extract_path with multi-element text-array paths",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '{"a":{"b":42}}'::jsonb #> ARRAY['a','b'];`,
					Expected: []sql.Row{{`42`}},
				},
				{
					// A deeper path mixing object keys and an array index.
					Query:    `SELECT '{"a":{"b":{"c":[10,20]}}}'::jsonb #> ARRAY['a','b','c','1'];`,
					Expected: []sql.Row{{`20`}},
				},
				{
					Query:    `SELECT '{"a":{"b":42}}'::jsonb #>> ARRAY['a','b'];`,
					Expected: []sql.Row{{`42`}},
				},
				{
					// The string 'NULL' is an ordinary key, distinct from a SQL
					// NULL element: both the ARRAY['NULL'] form and the quoted
					// '{"NULL"}' literal select the key named "NULL".
					Query:    `SELECT '{"NULL":7}'::jsonb #> ARRAY['NULL'];`,
					Expected: []sql.Row{{`7`}},
				},
				{
					Query:    `SELECT '{"NULL":7}'::jsonb #> '{"NULL"}';`,
					Expected: []sql.Row{{`7`}},
				},
			},
		},
		{
			Name: "jsonb_extract_path returns NULL for NULL path elements",
			Assertions: []ScriptTestAssertion{
				{
					// NULL as the trailing element.
					Query:    `SELECT '{"a":{"b":42}}'::jsonb #> ARRAY['a',NULL];`,
					Expected: []sql.Row{{nil}},
				},
				{
					// NULL as the leading element.
					Query:    `SELECT '{"a":{"b":42}}'::jsonb #> ARRAY[NULL,'b'];`,
					Expected: []sql.Row{{nil}},
				},
				{
					// NULL element in the middle of an otherwise valid path.
					Query:    `SELECT '{"a":{"b":42}}'::jsonb #> ARRAY['a',NULL,'b'];`,
					Expected: []sql.Row{{nil}},
				},
				{
					// Unquoted NULL in the '{...}' literal is a SQL NULL element.
					Query:    `SELECT '{"a":{"b":42}}'::jsonb #> '{a,NULL,b}';`,
					Expected: []sql.Row{{nil}},
				},
				{
					// A single unquoted NULL element, even when a key named
					// "NULL" exists, still yields NULL.
					Query:    `SELECT '{"NULL":7}'::jsonb #> '{NULL}';`,
					Expected: []sql.Row{{nil}},
				},
				{
					// The text-returning #>> variant behaves the same way.
					Query:    `SELECT '{"a":{"b":42}}'::jsonb #>> ARRAY['a',NULL];`,
					Expected: []sql.Row{{nil}},
				},
				{
					// A NULL array operand (vs. a NULL element) is NULL via the
					// function being strict.
					Query:    `SELECT '{"a":{"b":42}}'::jsonb #> NULL::text[];`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			// The json (non-binary) variants resolve through json_extract_path /
			// json_extract_path_text and must match the jsonb behavior above.
			Name: "json_extract_path with text-array paths and NULL elements",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '{"a":{"b":42}}'::json #> ARRAY['a','b'];`,
					Expected: []sql.Row{{`42`}},
				},
				{
					Query:    `SELECT '{"a":{"b":42}}'::json #>> ARRAY['a','b'];`,
					Expected: []sql.Row{{`42`}},
				},
				{
					Query:    `SELECT '{"a":{"b":42}}'::json #> ARRAY['a',NULL];`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT '{"a":{"b":42}}'::json #>> ARRAY[NULL,'b'];`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
	})
}

// TestJsonExists exercises the `?`, `?|`, and `?&` operators
// (jsonb_exists / jsonb_exists_any / jsonb_exists_all). For object operands
// the optimized path tests for the key via types.LookupJSONValue; for arrays
// and scalars the existing materialized check is used.
func TestJsonExists(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb_exists (?) tests key/element presence",
			Assertions: []ScriptTestAssertion{
				{
					// Object: key exists.
					Query:    `SELECT '{"a":1,"b":2}'::jsonb ? 'a';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// Object: missing key.
					Query:    `SELECT '{"a":1,"b":2}'::jsonb ? 'z';`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// Object: key whose value is JSON null still counts as
					// existing.
					Query:    `SELECT '{"a":null}'::jsonb ? 'a';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// Array: text matches a string element.
					Query:    `SELECT '["alpha","beta","gamma"]'::jsonb ? 'beta';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// Array: text does not match any element.
					Query:    `SELECT '["alpha","beta"]'::jsonb ? 'gamma';`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// Array: matching only on string elements, not numbers.
					Query:    `SELECT '[1,2,3]'::jsonb ? '1';`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// Scalar string equality.
					Query:    `SELECT '"hello"'::jsonb ? 'hello';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '"hello"'::jsonb ? 'world';`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// Non-string scalar never matches.
					Query:    `SELECT '42'::jsonb ? '42';`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "jsonb_exists_any (?|) tests presence of any key",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '{"a":1,"b":2}'::jsonb ?| ARRAY['x','b'];`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1,"b":2}'::jsonb ?| ARRAY['x','y'];`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '["a","b","c"]'::jsonb ?| ARRAY['x','b'];`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '["a","b","c"]'::jsonb ?| ARRAY['x','y'];`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "jsonb_exists_all (?&) tests presence of all keys",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '{"a":1,"b":2,"c":3}'::jsonb ?& ARRAY['a','b'];`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1,"b":2}'::jsonb ?& ARRAY['a','missing'];`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '["a","b","c"]'::jsonb ?& ARRAY['a','b'];`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '["a","b"]'::jsonb ?& ARRAY['a','missing'];`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
	})
}

// TestJsonLargeDocumentAccess exercises the same operators against JSONB
// values that are stored in a table column. Documents that exceed ~4 KB are
// stored as out-of-band IndexedJsonDocument values by Dolt's storage layer,
// which implements the SearchableJSON and ComparableJSON interfaces; this
// test ensures the optimized lookup paths in jsonb_object_field,
// jsonb_array_element, jsonb_extract_path, and jsonb_exists* still produce
// correct results when fed through the indexed representation.
func TestJsonLargeDocumentAccess(t *testing.T) {
	largeObj := makeLargeJSONObject(100) // ~8 KB
	largeArr := makeLargeJSONArray(80)   // ~5 KB

	RunScripts(t, []ScriptTest{
		{
			Name: "JSONB operators on large stored object (>4 KB)",
			SetUpScript: []string{
				`CREATE TABLE bigobj (id INT PRIMARY KEY, doc JSONB)`,
				`INSERT INTO bigobj (id, doc) VALUES (1, '` + largeObj + `'::jsonb)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Sanity check: the stored document is larger than 4 KB,
					// which exercises the indexed JSON document path.
					Query:    `SELECT length(doc::text) > 4096 FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// jsonb_object_field on a stored indexed document.
					Query:    `SELECT doc -> 'k_0037' ->> 'name' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{`value_0037`}},
				},
				{
					// First key at the start of the document.
					Query:    `SELECT doc -> 'k_0000' ->> 'name' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{`value_0000`}},
				},
				{
					// Last key at the end of the document.
					Query:    `SELECT doc -> 'k_0099' ->> 'name' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{`value_0099`}},
				},
				{
					// Missing key returns SQL NULL.
					Query:    `SELECT doc -> 'no_such_key' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{nil}},
				},
				{
					// Numeric value via ->>.
					Query:    `SELECT doc -> 'k_0042' ->> 'n' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{`42`}},
				},
				{
					// jsonb_extract_path through several levels of an indexed
					// document, ending at an array element.
					Query:    `SELECT doc #>> '{k_0010, tags, 2}' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{`tag-c`}},
				},
				{
					// jsonb_extract_path with a negative index hits the
					// negative-index fallback path inside extractOneJsonPathStep.
					Query:    `SELECT doc #>> '{k_0050, tags, -1}' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{`tag-e`}},
				},
				{
					// Missing intermediate path returns SQL NULL.
					Query:    `SELECT doc #> '{k_0001, missing}' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{nil}},
				},
				{
					// jsonb_exists on a stored indexed document.
					Query:    `SELECT doc ? 'k_0017' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT doc ? 'no_such_key' FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// jsonb_exists_any with a mix of present and missing keys.
					Query:    `SELECT doc ?| ARRAY['no_such_key', 'k_0005'] FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT doc ?| ARRAY['nope_1', 'nope_2'] FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// jsonb_exists_all where every key is present.
					Query:    `SELECT doc ?& ARRAY['k_0001', 'k_0099'] FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT doc ?& ARRAY['k_0001', 'no_such_key'] FROM bigobj WHERE id = 1;`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "JSONB operators on large stored array (>4 KB)",
			SetUpScript: []string{
				`CREATE TABLE bigarr (id INT PRIMARY KEY, doc JSONB)`,
				`INSERT INTO bigarr (id, doc) VALUES (1, '` + largeArr + `'::jsonb)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT length(doc::text) > 4096 FROM bigarr WHERE id = 1;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// Positive index hits the SearchableJSON fast path.
					Query:    `SELECT doc -> 17 ->> 'label' FROM bigarr WHERE id = 1;`,
					Expected: []sql.Row{{`row_0017`}},
				},
				{
					// First element.
					Query:    `SELECT doc -> 0 ->> 'label' FROM bigarr WHERE id = 1;`,
					Expected: []sql.Row{{`row_0000`}},
				},
				{
					// Last element via positive index.
					Query:    `SELECT doc -> 79 ->> 'label' FROM bigarr WHERE id = 1;`,
					Expected: []sql.Row{{`row_0079`}},
				},
				{
					// Negative index hits the materialized fallback path,
					// which must agree with the optimized path on the answer.
					Query:    `SELECT doc -> -1 ->> 'label' FROM bigarr WHERE id = 1;`,
					Expected: []sql.Row{{`row_0079`}},
				},
				{
					// Out-of-range positive index.
					Query:    `SELECT doc -> 1000 FROM bigarr WHERE id = 1;`,
					Expected: []sql.Row{{nil}},
				},
				{
					// jsonb_extract_path on an array followed by an object key.
					Query:    `SELECT doc #>> '{42, label}' FROM bigarr WHERE id = 1;`,
					Expected: []sql.Row{{`row_0042`}},
				},
				{
					// jsonb_extract_path into a nested array element.
					Query:    `SELECT doc #>> '{42, payload, 3}' FROM bigarr WHERE id = 1;`,
					Expected: []sql.Row{{`d`}},
				},
			},
		},
		{
			Name: "jsonb_extract_path on large stored object with numeric keys (>4 KB)",
			SetUpScript: []string{
				`CREATE TABLE numkeys (id INT PRIMARY KEY, doc JSONB)`,
				`INSERT INTO numkeys (id, doc) VALUES (1, '` + makeLargeJSONObjectWithNumericKeys() + `'::jsonb)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT length(doc::text) > 4096 FROM numkeys WHERE id = 1;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// A numeric path element on an object is first guessed as an
					// array index ([0]), which the indexed lookup rejects, so it
					// must fall back to the object key "0".
					Query:    `SELECT doc #>> '{nums, 0}' FROM numkeys WHERE id = 1;`,
					Expected: []sql.Row{{`zero`}},
				},
				{
					Query:    `SELECT doc #>> '{nums, 2}' FROM numkeys WHERE id = 1;`,
					Expected: []sql.Row{{`two`}},
				},
				{
					// Missing numeric key returns SQL NULL after the fallback.
					Query:    `SELECT doc #> '{nums, 5}' FROM numkeys WHERE id = 1;`,
					Expected: []sql.Row{{nil}},
				},
				{
					// A genuine object key + array index path through the same
					// large document still resolves on the single-lookup path.
					Query:    `SELECT doc #>> '{k_0001, tags, 0}' FROM numkeys WHERE id = 1;`,
					Expected: []sql.Row{{`tag-a`}},
				},
			},
		},
	})
}

// TestJsonbNumericCasts exercises the jsonb → numeric type casts in
// server/cast/jsonb.go. The integer casts must round half-to-even (matching
// Postgres' numeric → integer rules) and return an out-of-range error when
// the rounded value doesn't fit in the destination type. The float casts
// must reject values too large to represent as a finite value in the
// destination floating-point type. The non-numeric jsonb cases (object,
// array, string, boolean, null) must each error with a type-specific
// message.
func TestJsonbNumericCasts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb -> int2: rounding, boundaries, and out-of-range",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '12345'::jsonb::int2;`,
					Expected: []sql.Row{{int16(12345)}},
				},
				{
					Query:    `SELECT '-12345'::jsonb::int2;`,
					Expected: []sql.Row{{int16(-12345)}},
				},
				{
					// Half-to-even rounding: 0.4 always rounds down.
					Query:    `SELECT '12345.4'::jsonb::int2;`,
					Expected: []sql.Row{{int16(12345)}},
				},
				{
					// 12345.5 → 12346 (round half to even, 12346 is even).
					Query:    `SELECT '12345.5'::jsonb::int2;`,
					Expected: []sql.Row{{int16(12346)}},
				},
				{
					// 12346.5 → 12346 (round half to even, 12346 is even).
					Query:    `SELECT '12346.5'::jsonb::int2;`,
					Expected: []sql.Row{{int16(12346)}},
				},
				{
					// Boundary values that fit exactly.
					Query:    `SELECT '32767'::jsonb::int2;`,
					Expected: []sql.Row{{int16(32767)}},
				},
				{
					Query:    `SELECT '-32768'::jsonb::int2;`,
					Expected: []sql.Row{{int16(-32768)}},
				},
				{
					// Fractional value that rounds down into range.
					Query:    `SELECT '32767.4'::jsonb::int2;`,
					Expected: []sql.Row{{int16(32767)}},
				},
				{
					// One past the upper bound.
					Query:       `SELECT '32768'::jsonb::int2;`,
					ExpectedErr: "smallint out of range",
				},
				{
					// 32767.5 rounds to 32768, which is out of range.
					Query:       `SELECT '32767.5'::jsonb::int2;`,
					ExpectedErr: "smallint out of range",
				},
				{
					Query:       `SELECT '-32769'::jsonb::int2;`,
					ExpectedErr: "smallint out of range",
				},
				{
					// Values far outside the int16 range still produce a
					// clean out-of-range error rather than an int64 overflow.
					Query:       `SELECT '1e20'::jsonb::int2;`,
					ExpectedErr: "smallint out of range",
				},
			},
		},
		{
			Name: "jsonb -> int4: rounding, boundaries, and out-of-range",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '0'::jsonb::int4;`,
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    `SELECT '2147483647'::jsonb::int4;`,
					Expected: []sql.Row{{int32(2147483647)}},
				},
				{
					Query:    `SELECT '-2147483648'::jsonb::int4;`,
					Expected: []sql.Row{{int32(-2147483648)}},
				},
				{
					// Fractional that rounds down into range.
					Query:    `SELECT '2147483647.4'::jsonb::int4;`,
					Expected: []sql.Row{{int32(2147483647)}},
				},
				{
					Query:       `SELECT '2147483648'::jsonb::int4;`,
					ExpectedErr: "integer out of range",
				},
				{
					Query:       `SELECT '-2147483649'::jsonb::int4;`,
					ExpectedErr: "integer out of range",
				},
				{
					Query:       `SELECT '1e20'::jsonb::int4;`,
					ExpectedErr: "integer out of range",
				},
			},
		},
		{
			Name: "jsonb -> int8: rounding, boundaries, and out-of-range",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '0'::jsonb::int8;`,
					Expected: []sql.Row{{int64(0)}},
				},
				{
					// 2^53 - 1: the largest integer that survives the
					// round-trip through float64 that the jsonb parser
					// currently performs on input.
					Query:    `SELECT '9007199254740991'::jsonb::int8;`,
					Expected: []sql.Row{{int64(9007199254740991)}},
				},
				{
					Query:    `SELECT '-9007199254740991'::jsonb::int8;`,
					Expected: []sql.Row{{int64(-9007199254740991)}},
				},
				{
					// Large value that doesn't fit in int64 must error
					// rather than silently truncating.
					Query:       `SELECT '1e20'::jsonb::int8;`,
					ExpectedErr: "bigint out of range",
				},
				{
					Query:       `SELECT '-1e20'::jsonb::int8;`,
					ExpectedErr: "bigint out of range",
				},
			},
		},
		{
			Name: "jsonb -> float4: out-of-range",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '0'::jsonb::float4;`,
					Expected: []sql.Row{{float32(0)}},
				},
				{
					Query:    `SELECT '1.5'::jsonb::float4;`,
					Expected: []sql.Row{{float32(1.5)}},
				},
				{
					// Just inside float32 max (~3.4028235e38).
					Query:    `SELECT '3.4e38'::jsonb::float4;`,
					Expected: []sql.Row{{float32(3.4e38)}},
				},
				{
					// Just outside float32 max.
					Query:       `SELECT '3.5e38'::jsonb::float4;`,
					ExpectedErr: "out of range",
				},
				{
					Query:       `SELECT '-3.5e38'::jsonb::float4;`,
					ExpectedErr: "out of range",
				},
				{
					Query:       `SELECT '1e40'::jsonb::float4;`,
					ExpectedErr: "out of range",
				},
			},
		},
		{
			Name: "jsonb -> float8 round-trips finite values",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '0'::jsonb::float8;`,
					Expected: []sql.Row{{float64(0)}},
				},
				{
					Query:    `SELECT '1.5'::jsonb::float8;`,
					Expected: []sql.Row{{float64(1.5)}},
				},
				{
					// Larger value that still fits in float64.
					Query:    `SELECT '1e300'::jsonb::float8;`,
					Expected: []sql.Row{{float64(1e300)}},
				},
				// Out-of-range float8 values can't be tested via a jsonb
				// literal: the jsonb parser itself rejects '1e400' because
				// it cannot be represented in the float64 used for input
				// parsing.
			},
		},
		{
			Name: "jsonb -> numeric: preserves precision",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '12345'::jsonb::numeric;`,
					Expected: []sql.Row{{Numeric("12345")}},
				},
				{
					Query:    `SELECT '12345.67'::jsonb::numeric;`,
					Expected: []sql.Row{{Numeric("12345.67")}},
				},
				{
					Query:    `SELECT '-12345.67'::jsonb::numeric;`,
					Expected: []sql.Row{{Numeric("-12345.67")}},
				},
			},
		},
		{
			Name: "jsonb non-numeric values reject numeric casts",
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT '{}'::jsonb::int4;`,
					ExpectedErr: "cannot cast jsonb object",
				},
				{
					Query:       `SELECT '[]'::jsonb::int4;`,
					ExpectedErr: "cannot cast jsonb array",
				},
				{
					Query:       `SELECT '"42"'::jsonb::int4;`,
					ExpectedErr: "cannot cast jsonb string",
				},
				{
					Query:       `SELECT 'true'::jsonb::int4;`,
					ExpectedErr: "cannot cast jsonb boolean",
				},
				{
					Query:       `SELECT 'null'::jsonb::int4;`,
					ExpectedErr: "cannot cast jsonb null",
				},
				{
					Query:       `SELECT '{}'::jsonb::float4;`,
					ExpectedErr: "cannot cast jsonb object",
				},
				{
					Query:       `SELECT '[]'::jsonb::numeric;`,
					ExpectedErr: "cannot cast jsonb array",
				},
			},
		},
	})
}

// TestJsonTypeof exercises json_typeof and jsonb_typeof, which report the
// shape of a JSON value. These show up most often inside CHECK constraints,
// which is covered at the end of the test.
func TestJsonTypeof(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb_typeof over every JSON type",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT jsonb_typeof('{}'::jsonb);`,
					Expected: []sql.Row{{"object"}},
				},
				{
					Query:    `SELECT jsonb_typeof('{"a":1}'::jsonb);`,
					Expected: []sql.Row{{"object"}},
				},
				{
					Query:    `SELECT jsonb_typeof('[]'::jsonb);`,
					Expected: []sql.Row{{"array"}},
				},
				{
					Query:    `SELECT jsonb_typeof('[1,2,3]'::jsonb);`,
					Expected: []sql.Row{{"array"}},
				},
				{
					Query:    `SELECT jsonb_typeof('"str"'::jsonb);`,
					Expected: []sql.Row{{"string"}},
				},
				{
					Query:    `SELECT jsonb_typeof('1'::jsonb), jsonb_typeof('-1.5'::jsonb), jsonb_typeof('1e3'::jsonb);`,
					Expected: []sql.Row{{"number", "number", "number"}},
				},
				{
					Query:    `SELECT jsonb_typeof('true'::jsonb), jsonb_typeof('false'::jsonb);`,
					Expected: []sql.Row{{"boolean", "boolean"}},
				},
				{
					// The JSON value `null` reports as "null"...
					Query:    `SELECT jsonb_typeof('null'::jsonb);`,
					Expected: []sql.Row{{"null"}},
				},
				{
					// ...while a SQL NULL input yields a SQL NULL result.
					Query:    `SELECT jsonb_typeof(null::jsonb);`,
					Expected: []sql.Row{{nil}},
				},
				{
					// Only the top level is inspected, not the nested values.
					Query:    `SELECT jsonb_typeof('{"a":[1,2]}'::jsonb);`,
					Expected: []sql.Row{{"object"}},
				},
				{
					Query:    `SELECT pg_typeof(jsonb_typeof('{}'::jsonb));`,
					Expected: []sql.Row{{"text"}},
				},
			},
		},
		{
			Name: "json_typeof over every JSON type",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT json_typeof('{"a":1}'::json);`,
					Expected: []sql.Row{{"object"}},
				},
				{
					Query:    `SELECT json_typeof('[1]'::json);`,
					Expected: []sql.Row{{"array"}},
				},
				{
					Query:    `SELECT json_typeof('"str"'::json);`,
					Expected: []sql.Row{{"string"}},
				},
				{
					Query:    `SELECT json_typeof('42'::json);`,
					Expected: []sql.Row{{"number"}},
				},
				{
					Query:    `SELECT json_typeof('true'::json);`,
					Expected: []sql.Row{{"boolean"}},
				},
				{
					Query:    `SELECT json_typeof('null'::json);`,
					Expected: []sql.Row{{"null"}},
				},
				{
					Query:    `SELECT json_typeof(null::json);`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "jsonb_typeof pins the shape of a column in a CHECK constraint",
			SetUpScript: []string{
				`CREATE TABLE public.t (
				    ext jsonb,
				    evidence jsonb,
				    CONSTRAINT t_ext_object_check CHECK ((jsonb_typeof(ext) = 'object'::text)),
				    CONSTRAINT t_evidence_check   CHECK (((jsonb_typeof(evidence) = 'array'::text)
				                                          AND (jsonb_array_length(evidence) >= 1)))
				);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO t VALUES ('{"a":1}', '[1]');`,
					Expected: []sql.Row{},
				},
				{
					Query:       `INSERT INTO t VALUES ('[1]', '[1]');`,
					ExpectedErr: `"t_ext_object_check" violated`,
				},
				{
					// The array is the right shape but empty, so the length
					// half of the constraint rejects it.
					Query:       `INSERT INTO t VALUES ('{"a":1}', '[]');`,
					ExpectedErr: `"t_evidence_check" violated`,
				},
				{
					Query:    `SELECT jsonb_typeof(ext), jsonb_array_length(evidence) FROM t;`,
					Expected: []sql.Row{{"object", int32(1)}},
				},
			},
		},
	})
}

// TestJsonArrayLength exercises json_array_length and jsonb_array_length,
// including the two distinct errors Postgres raises for non-array inputs.
func TestJsonArrayLength(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb_array_length",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT jsonb_array_length('[]'::jsonb);`,
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    `SELECT jsonb_array_length('[1,2,3]'::jsonb);`,
					Expected: []sql.Row{{int32(3)}},
				},
				{
					// Only the top level is counted; nested arrays count as
					// one element each.
					Query:    `SELECT jsonb_array_length('[[1,2],[3,4],[5]]'::jsonb);`,
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    `SELECT jsonb_array_length('[null,null]'::jsonb);`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT jsonb_array_length(null::jsonb);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT pg_typeof(jsonb_array_length('[]'::jsonb));`,
					Expected: []sql.Row{{"integer"}},
				},
				{
					// Postgres words the object and scalar cases differently.
					Query:           `SELECT jsonb_array_length('{}'::jsonb);`,
					ExpectedErr:     "cannot get array length of a non-array",
					ExpectedErrCode: "22023",
				},
				{
					Query:           `SELECT jsonb_array_length('1'::jsonb);`,
					ExpectedErr:     "cannot get array length of a scalar",
					ExpectedErrCode: "22023",
				},
				{
					Query:       `SELECT jsonb_array_length('"str"'::jsonb);`,
					ExpectedErr: "cannot get array length of a scalar",
				},
				{
					Query:       `SELECT jsonb_array_length('null'::jsonb);`,
					ExpectedErr: "cannot get array length of a scalar",
				},
			},
		},
		{
			Name: "json_array_length",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT json_array_length('[1,2]'::json);`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT json_array_length('[]'::json);`,
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    `SELECT json_array_length(null::json);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:       `SELECT json_array_length('{"a":1}'::json);`,
					ExpectedErr: "cannot get array length of a non-array",
				},
				{
					Query:       `SELECT json_array_length('true'::json);`,
					ExpectedErr: "cannot get array length of a scalar",
				},
			},
		},
	})
}

// TestJsonObjectKeys exercises json_object_keys and jsonb_object_keys, which
// are set-returning functions over the top-level keys of a JSON object.
func TestJsonObjectKeys(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb_object_keys",
			Assertions: []ScriptTestAssertion{
				{
					// Keys come back in jsonb order: shortest first, then
					// bytewise, matching the order the document prints in.
					Query:    `SELECT jsonb_object_keys('{"b":1,"aa":2,"c":3}'::jsonb);`,
					Expected: []sql.Row{{"b"}, {"c"}, {"aa"}},
				},
				{
					Query:            `SELECT jsonb_object_keys('{"a":1}'::jsonb);`,
					Expected:         []sql.Row{{"a"}},
					ExpectedColNames: []string{"jsonb_object_keys"},
				},
				{
					// An empty object produces no rows at all.
					Query:    `SELECT jsonb_object_keys('{}'::jsonb);`,
					Expected: []sql.Row{},
				},
				{
					// Only top-level keys are returned, not nested ones.
					Query:    `SELECT jsonb_object_keys('{"a":{"nested":1},"b":2}'::jsonb);`,
					Expected: []sql.Row{{"a"}, {"b"}},
				},
				{
					// Usable in the FROM clause with an alias.
					Query:    `SELECT k FROM jsonb_object_keys('{"b":1,"aa":2}'::jsonb) AS k;`,
					Expected: []sql.Row{{"b"}, {"aa"}},
				},
				{
					Query:           `SELECT jsonb_object_keys('[1,2]'::jsonb);`,
					ExpectedErr:     "cannot call jsonb_object_keys on an array",
					ExpectedErrCode: "22023",
				},
				{
					Query:           `SELECT jsonb_object_keys('42'::jsonb);`,
					ExpectedErr:     "cannot call jsonb_object_keys on a scalar",
					ExpectedErrCode: "22023",
				},
			},
		},
		{
			Name: "json_object_keys",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT json_object_keys('{"b":1,"aa":2,"c":3}'::json);`,
					Expected: []sql.Row{{"b"}, {"c"}, {"aa"}},
				},
				{
					Query:    `SELECT json_object_keys('{}'::json);`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT json_object_keys('["a"]'::json);`,
					ExpectedErr: "cannot call json_object_keys on an array",
				},
				{
					Query:       `SELECT json_object_keys('"str"'::json);`,
					ExpectedErr: "cannot call json_object_keys on a scalar",
				},
			},
		},
	})
}

// TestJsonStripNulls exercises json_strip_nulls and jsonb_strip_nulls, which
// recursively drop object fields whose value is JSON null.
func TestJsonStripNulls(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "jsonb_strip_nulls",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT jsonb_strip_nulls('{"a":1,"b":null}'::jsonb);`,
					Expected: []sql.Row{{`{"a": 1}`}},
				},
				{
					// Nulls are stripped at every level of nesting.
					Query:    `SELECT jsonb_strip_nulls('{"a":1,"b":null,"c":{"d":null,"e":2}}'::jsonb);`,
					Expected: []sql.Row{{`{"a": 1, "c": {"e": 2}}`}},
				},
				{
					// Null array elements are kept, as are objects nested
					// inside arrays, whose null fields are still stripped.
					Query:    `SELECT jsonb_strip_nulls('{"a":[1,null,{"b":null,"c":3}]}'::jsonb);`,
					Expected: []sql.Row{{`{"a": [1, null, {"c": 3}]}`}},
				},
				{
					// Stripping every field leaves an empty object, not null.
					Query:    `SELECT jsonb_strip_nulls('{"a":null}'::jsonb);`,
					Expected: []sql.Row{{`{}`}},
				},
				{
					Query:    `SELECT jsonb_strip_nulls('[1,null,2]'::jsonb);`,
					Expected: []sql.Row{{`[1, null, 2]`}},
				},
				{
					Query:    `SELECT jsonb_strip_nulls('{}'::jsonb);`,
					Expected: []sql.Row{{`{}`}},
				},
				{
					Query:    `SELECT jsonb_strip_nulls(null::jsonb);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT pg_typeof(jsonb_strip_nulls('{}'::jsonb));`,
					Expected: []sql.Row{{"jsonb"}},
				},
			},
		},
		{
			Name: "json_strip_nulls",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT json_strip_nulls('{"a":null,"b":1}'::json);`,
					Expected: []sql.Row{{`{"b": 1}`}},
				},
				{
					Query:    `SELECT json_strip_nulls('{"a":{"b":null}}'::json);`,
					Expected: []sql.Row{{`{"a": {}}`}},
				},
				{
					Query:    `SELECT json_strip_nulls(null::json);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT pg_typeof(json_strip_nulls('{}'::json));`,
					Expected: []sql.Row{{"json"}},
				},
			},
		},
	})
}

// TestToJsonb exercises to_jsonb, the jsonb counterpart of to_json. Unlike
// to_json, which hands back the rendered text as-is, to_jsonb produces a
// parsed document, so object keys come back deduplicated and reordered.
func TestToJsonb(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "to_jsonb over scalar types",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT to_jsonb(1), to_jsonb(1.5::numeric), to_jsonb(true), to_jsonb('a'::text);`,
					Expected: []sql.Row{{`1`, `1.5`, `true`, `"a"`}},
				},
				{
					// Text is quoted and escaped rather than parsed as JSON.
					Query:    `SELECT to_jsonb('{"not":"parsed"}'::text);`,
					Expected: []sql.Row{{`"{\"not\":\"parsed\"}"`}},
				},
				{
					// Dates and timestamps render as ISO 8601 strings.
					Query:    `SELECT to_jsonb('2020-01-01'::date), jsonb_typeof(to_jsonb('2020-01-01'::date));`,
					Expected: []sql.Row{{`"2020-01-01"`, "string"}},
				},
				{
					Query:    `SELECT to_jsonb(null::int4);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT pg_typeof(to_jsonb(1));`,
					Expected: []sql.Row{{"jsonb"}},
				},
			},
		},
		{
			Name: "to_jsonb over composite values",
			SetUpScript: []string{
				`CREATE TABLE t (id int4, name text)`,
				`INSERT INTO t VALUES (1, 'one')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT to_jsonb(ARRAY[1,2,3]);`,
					Expected: []sql.Row{{`[1, 2, 3]`}},
				},
				{
					Query:    `SELECT to_jsonb(ARRAY['a','b']::text[]);`,
					Expected: []sql.Row{{`["a", "b"]`}},
				},
				{
					// A row value becomes an object keyed by column name,
					// which is how to_jsonb(NEW)/to_jsonb(OLD) is used in
					// trigger bodies.
					Query:    `SELECT to_jsonb(t) FROM t;`,
					Expected: []sql.Row{{`{"id": 1, "name": "one"}`}},
				},
				{
					// Unlike to_json, to_jsonb reorders object keys into
					// jsonb order (shortest first, then bytewise).
					Query:    `SELECT to_jsonb('{"bbb":1,"a":2}'::json), to_json('{"bbb":1,"a":2}'::json);`,
					Expected: []sql.Row{{`{"a": 2, "bbb": 1}`, `{"bbb":1,"a":2}`}},
				},
				{
					Query:    `SELECT to_jsonb('{"a":1}'::jsonb);`,
					Expected: []sql.Row{{`{"a": 1}`}},
				},
			},
		},
	})
}

// TestJsonInspectionStoredValues runs the inspection functions against JSONB
// values read out of a table column rather than written as literals. Documents
// over ~4 KB are held as out-of-band IndexedJsonDocument values by Dolt's
// storage layer, a different sql.JSONWrapper implementation than the one a
// literal produces.
func TestJsonInspectionStoredValues(t *testing.T) {
	largeObj := makeLargeJSONObject(100) // ~8 KB
	largeArr := makeLargeJSONArray(80)   // ~5 KB

	RunScripts(t, []ScriptTest{
		{
			Name: "inspection functions on large stored documents (>4 KB)",
			SetUpScript: []string{
				`CREATE TABLE bigdoc (id INT PRIMARY KEY, doc JSONB)`,
				`INSERT INTO bigdoc (id, doc) VALUES (1, '` + largeObj + `'::jsonb)`,
				`INSERT INTO bigdoc (id, doc) VALUES (2, '` + largeArr + `'::jsonb)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Sanity check: both documents exercise the indexed path.
					Query:    `SELECT length(doc::text) > 4096 FROM bigdoc ORDER BY id;`,
					Expected: []sql.Row{{"t"}, {"t"}},
				},
				{
					Query:    `SELECT jsonb_typeof(doc) FROM bigdoc ORDER BY id;`,
					Expected: []sql.Row{{"object"}, {"array"}},
				},
				{
					Query:    `SELECT jsonb_array_length(doc) FROM bigdoc WHERE id = 2;`,
					Expected: []sql.Row{{int32(80)}},
				},
				{
					Query:    `SELECT count(*) FROM jsonb_object_keys((SELECT doc FROM bigdoc WHERE id = 1)) AS k;`,
					Expected: []sql.Row{{int64(100)}},
				},
				{
					// Nothing to strip, so the document round-trips.
					Query:    `SELECT jsonb_strip_nulls(doc) = doc FROM bigdoc ORDER BY id;`,
					Expected: []sql.Row{{"t"}, {"t"}},
				},
			},
		},
		{
			Name: "inspection functions on a NULL column value",
			SetUpScript: []string{
				`CREATE TABLE nulldoc (id INT PRIMARY KEY, doc JSONB, j JSON)`,
				`INSERT INTO nulldoc VALUES (1, NULL, NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// All of these are strict, so a NULL column value yields
					// a NULL result rather than an error.
					Query: `SELECT jsonb_typeof(doc), json_typeof(j), jsonb_array_length(doc),
					               jsonb_strip_nulls(doc), to_jsonb(doc) FROM nulldoc;`,
					Expected: []sql.Row{{nil, nil, nil, nil, nil}},
				},
				{
					// A strict set-returning function over NULL yields no rows.
					Query:    `SELECT jsonb_object_keys(doc) FROM nulldoc;`,
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestJsonTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "JSON_TABLE",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc JSONB);",
				`INSERT INTO docs VALUES (1, '[{"a":1},{"a":2}]'), (2, '[{"a":3}]'), (3, NULL), (4, '{"a":4}');`,
				"CREATE TABLE texts (id INT PRIMARY KEY, doc TEXT);",
				`INSERT INTO texts VALUES (1, '{"a":"x"}');`,
				"CREATE VIEW v AS SELECT docs.id, jt.a FROM docs, JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]', '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::json, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::text, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1]'::varchar, '$[*]' COLUMNS (a INT PATH '$')) AS jt;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1]'::text FORMAT JSON, '$[*]' COLUMNS (a INT PATH '$')) AS jt;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE(1, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;`,
					ExpectedErr:     "cannot cast type integer to jsonb",
					ExpectedErrCode: "42846",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::bytea, '$[*]' COLUMNS (a INT PATH '$')) AS jt;`,
					ExpectedErr:     "cannot cast type bytea to jsonb",
					ExpectedErrCode: "42846",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb FORMAT JSON ENCODING UTF8, '$[*]' COLUMNS (a INT PATH '$')) AS jt;`,
					ExpectedErr:     "JSON ENCODING clause is only allowed for bytea input type",
					ExpectedErrCode: "42804",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('not json'::text, '$[*]' COLUMNS (a INT)) AS jt;`,
					ExpectedErr:     "invalid input syntax for type json",
					ExpectedErrCode: "22P02",
				},
				{
					Query:            `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (a INT)) AS jt(x);`,
					Expected:         []sql.Row{{1}, {2}},
					ExpectedColNames: []string{"x"},
				},
				{
					Query:    `SELECT json_table.a FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (a INT));`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE(NULL::jsonb, '$[*]' COLUMNS (a INT)) AS jt;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE(NULL, '$[*]' COLUMNS (a INT)) AS jt;`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT docs.id, jt.* FROM docs, JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt ORDER BY 1, 2;",
					Expected: []sql.Row{{1, 1}, {1, 2}, {2, 3}, {4, 4}},
				},
				{
					Query:    "SELECT docs.id, jt.* FROM docs CROSS JOIN LATERAL JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt ORDER BY 1, 2;",
					Expected: []sql.Row{{1, 1}, {1, 2}, {2, 3}, {4, 4}},
				},
				{
					Query:    "SELECT docs.id, jt.* FROM docs LEFT JOIN LATERAL JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt ON TRUE ORDER BY 1, 2;",
					Expected: []sql.Row{{1, 1}, {1, 2}, {2, 3}, {3, nil}, {4, 4}},
				},
				{
					Query:    "SELECT docs.id, jt.* FROM docs, JSON_TABLE(docs.doc, '$[*] ? (@.a > $m)' PASSING docs.id AS m COLUMNS (a INT PATH '$.a' DEFAULT -1 ON EMPTY)) AS jt ORDER BY 1, 2;",
					Expected: []sql.Row{{1, 2}, {2, 3}},
				},
				{
					Query:    "SELECT * FROM v ORDER BY 1, 2;",
					Expected: []sql.Row{{1, 1}, {1, 2}, {2, 3}, {4, 4}},
				},
				{
					Query:    "SELECT jt.a FROM texts, JSON_TABLE(texts.doc, '$' COLUMNS (a TEXT)) AS jt;",
					Expected: []sql.Row{{"x"}},
				},
				{
					Query:    "SELECT count(*), sum(jt.a) FROM docs, JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT)) AS jt WHERE jt.a > 1;",
					Expected: []sql.Row{{3, 9}},
				},
			},
		},
		{
			Name: "JSON_TABLE paths and PASSING",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*] ? (@.a > $x)' PASSING 1 AS x COLUMNS (a INT)) AS jt;`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' AS p COLUMNS (a INT)) AS jt;`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1,2,3]'::jsonb, '$[*] ? (@ >= $a && @ < $b)' PASSING 2 AS a, 3.5::numeric AS b COLUMNS (v INT PATH '$')) AS jt;`,
					Expected: []sql.Row{{2}, {3}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('["a","b"]'::jsonb, '$[*] ? (@ == $s)' PASSING 'b' AS s, 'b'::varchar AS t COLUMNS (v TEXT PATH '$', w TEXT PATH '$t')) AS jt;`,
					Expected: []sql.Row{{"b", "b"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[true,false]'::jsonb, '$[*] ? (@ == $s)' PASSING TRUE AS s COLUMNS (v TEXT PATH '$')) AS jt;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"k":1},2]'::jsonb, '$[*] ? (@ == $s.k)' PASSING '{"k":2}'::jsonb AS s COLUMNS (v TEXT PATH '$')) AS jt;`,
					Expected: []sql.Row{{"2"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1,2]'::jsonb, '$' PASSING '2020-01-01'::date AS d, '2020-01-01 10:00'::timestamp AS ts, '10:00'::time AS t COLUMNS (v TEXT PATH '$d', w JSONB PATH '$d', x TEXT PATH '$ts', y JSONB PATH '$ts', z TEXT PATH '$t')) AS jt;`,
					Expected: []sql.Row{{"2020-01-01", `"2020-01-01"`, "2020-01-01 10:00:00", `"2020-01-01T10:00:00"`, "10:00:00"}},
				},
				{
					Query:    `SELECT v, w::text FROM JSON_TABLE('[1,2]'::jsonb, '$' PASSING NULL AS d, NULL::int AS e COLUMNS (v TEXT PATH '$d', w JSONB PATH '$e')) AS jt;`,
					Expected: []sql.Row{{nil, "null"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1,2]'::jsonb, '$' PASSING 1.5::float8 AS d, 2::int8 AS e, 3::int2 AS f, 4.5::float4 AS g COLUMNS (v TEXT PATH '$d', w TEXT PATH '$e', x TEXT PATH '$f', y TEXT PATH '$g')) AS jt;`,
					Expected: []sql.Row{{"1.5", "2", "3", "4.5"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$' PASSING ARRAY[1] AS d, ROW(1,2) AS e COLUMNS (v JSONB PATH '$d' ERROR ON ERROR, w JSONB PATH '$e' ERROR ON ERROR)) AS jt;`,
					Expected: []sql.Row{{"[1]", `{"f1": 1, "f2": 2}`}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1,2]'::jsonb, '$' PASSING 1 AS d, 2 AS d, 3 AS "D" COLUMNS (v TEXT PATH '$d', w TEXT PATH '$D')) AS jt;`,
					Expected: []sql.Row{{"1", "3"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1,2]'::jsonb, '$' PASSING '[1]' FORMAT JSON AS d, '[2]'::text FORMAT JSON AS e COLUMNS (v TEXT PATH '$d[0]', w TEXT PATH '$e[0]')) AS jt;`,
					Expected: []sql.Row{{"1", "2"}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1,2]'::jsonb, '$' PASSING 1 AS D COLUMNS (v TEXT PATH '$D')) AS jt;`,
					ExpectedErr:     `could not find jsonpath variable "D"`,
					ExpectedErrCode: "42704",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*] ? ($x > 0)' COLUMNS (e INT PATH '$')) AS jt;`,
					ExpectedErr:     `could not find jsonpath variable "x"`,
					ExpectedErrCode: "42704",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('["a","b"]'::jsonb, '$[*] ? (@ == $s)' PASSING 'b'::name AS s COLUMNS (v TEXT PATH '$')) AS jt;`,
					ExpectedErr:     "could not convert value of type name to jsonpath",
					ExpectedErrCode: "22023",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (a INT, a TEXT)) AS jt;`,
					ExpectedErr:     "duplicate JSON_TABLE column or path name: a",
					ExpectedErrCode: "42712",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' AS a COLUMNS (a INT)) AS jt;`,
					ExpectedErr:     "duplicate JSON_TABLE column or path name: a",
					ExpectedErrCode: "42712",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (n FOR ORDINALITY, m FOR ORDINALITY)) AS jt;`,
					ExpectedErr:     "only one FOR ORDINALITY column is allowed",
					ExpectedErrCode: "42601",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*' COLUMNS (a INT)) AS jt;`,
					ExpectedErr:     "syntax error at end of jsonpath input",
					ExpectedErrCode: "42601",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '' COLUMNS (a INT)) AS jt;`,
					ExpectedErr:     `invalid input syntax for type jsonpath: ""`,
					ExpectedErrCode: "22P02",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$' || '[*]' COLUMNS (a INT)) AS jt;`,
					ExpectedErr:     "only string constants are supported in JSON_TABLE path specification",
					ExpectedErrCode: "0A000",
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('{"a":1}'::jsonb, 'strict $.b' COLUMNS (a INT)) AS jt;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('{"a":1}'::jsonb, 'strict $.b' COLUMNS (a INT) EMPTY ARRAY ON ERROR) AS jt;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('{"a":1}'::jsonb, 'strict $.b' COLUMNS (a INT) ERROR ON ERROR) AS jt;`,
					ExpectedErr:     `JSON object does not contain key "b"`,
					ExpectedErrCode: "2203A",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('{"a":1}'::jsonb, 'strict $.b' COLUMNS (a INT) NULL ON ERROR) AS jt;`,
					ExpectedErr:     "invalid ON ERROR behavior",
					ExpectedErrCode: "42601",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*] / 0' COLUMNS (e INT PATH '$') ERROR ON ERROR) AS jt;`,
					ExpectedErr:     "division by zero",
					ExpectedErrCode: "22012",
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a') ERROR ON ERROR) AS jt;`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "JSON_TABLE columns",
			SetUpScript: []string{
				"CREATE TYPE pair AS (x INT, y TEXT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (n FOR ORDINALITY, a INT)) AS jt;`,
					Expected: []sql.Row{{1, 1}, {2, 2}},
				},
				{
					Query:    `SELECT a, b::text, n FROM JSON_TABLE('[1,"x",true,null,1.5]'::jsonb, '$[*]' COLUMNS (a TEXT PATH '$', b JSONB PATH '$', n FOR ORDINALITY)) AS jt;`,
					Expected: []sql.Row{{"1", "1", 1}, {"x", `"x"`, 2}, {"t", "true", 3}, {nil, "null", 4}, {"1.5", "1.5", 5}},
				},
				{
					Query: `SELECT i, t, n, b, j::text, js::text FROM JSON_TABLE('[{"a":1.5},{"a":"12"},{"a":true},{"a":[1]},{"a":{"b":1}},{"a":null},{}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a', t TEXT PATH '$.a', n NUMERIC PATH '$.a', b BOOL PATH '$.a', j JSONB PATH '$.a', js JSON PATH '$.a')) AS jt;`,
					Expected: []sql.Row{
						{nil, "1.5", Numeric("1.5"), nil, "1.5", "1.5"},
						{12, "12", Numeric("12"), nil, `"12"`, `"12"`},
						{nil, "t", nil, "t", "true", "true"},
						{nil, nil, nil, nil, "[1]", "[1]"},
						{nil, nil, nil, nil, `{"b": 1}`, `{"b": 1}`},
						{nil, nil, nil, nil, "null", "null"},
						{nil, nil, nil, nil, nil, nil},
					},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('{"a":{"b":1}}'::jsonb, '$' COLUMNS (a TEXT PATH '$.a', b TEXT FORMAT JSON PATH '$.a', c JSON PATH '$.a', d TEXT PATH '$."a"."b"', "A" INT PATH '$.a.b')) AS jt;`,
					Expected: []sql.Row{{nil, `{"b": 1}`, `{"b": 1}`, "1", 1}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('{"A":1, "a":2, "a b":3, "q\"x":4}'::jsonb, '$' COLUMNS ("A" INT, a INT, "a b" INT, "q""x" INT)) AS jt;`,
					Expected: []sql.Row{{1, 2, 3, 4}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1.5}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "invalid input syntax for type",
					ExpectedErrCode: "22P02",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":[1]}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     `JSON path expression for column "i" must return single scalar item`,
					ExpectedErrCode: "2203F",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a[*]' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     `JSON path expression for column "i" must return single scalar item`,
					ExpectedErrCode: "22034",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (i JSONB PATH '$.a[*]' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     `JSON path expression for column "i" must return single item when no wrapper is requested`,
					ExpectedErrCode: "22034",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' ERROR ON EMPTY)) AS jt;`,
					ExpectedErr:     `no SQL/JSON item found for specified path of column "i"`,
					ExpectedErrCode: "22035",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{}]'::jsonb, '$[*]' COLUMNS (i INT PATH 'strict $.a' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     `JSON object does not contain key "a"`,
					ExpectedErrCode: "2203A",
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{}]'::jsonb, '$[*]' COLUMNS (i INT PATH 'strict $.a' ERROR ON EMPTY, j INT PATH '$.a' DEFAULT 5 ON EMPTY)) AS jt;`,
					Expected: []sql.Row{{nil, 5}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' DEFAULT 5 ON EMPTY DEFAULT 6 ON ERROR, j INT PATH '$.a' DEFAULT '7' ON ERROR, k INT PATH '$.a' DEFAULT 7.5 ON ERROR, l INT PATH '$.a' DEFAULT 1 + 1 ON ERROR, m INT PATH '$.a' DEFAULT length('ab') ON ERROR)) AS jt;`,
					Expected: []sql.Row{{6, 7, 8, 2, 2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i JSONB PATH '$.b' DEFAULT '{"q":1}' ON EMPTY)) AS jt;`,
					Expected: []sql.Row{{`{"q": 1}`}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' DEFAULT 'q' ON ERROR)) AS jt;`,
					ExpectedErr:     "invalid input syntax for type",
					ExpectedErrCode: "22P02",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' DEFAULT (SELECT 1) ON ERROR)) AS jt;`,
					ExpectedErr:     "can only specify a constant, non-aggregate function, or operator expression for DEFAULT",
					ExpectedErrCode: "42804",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' EMPTY ON ERROR)) AS jt;`,
					ExpectedErr:     `invalid ON ERROR behavior for column "i"`,
					ExpectedErrCode: "42601",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' EMPTY ARRAY ON EMPTY)) AS jt;`,
					ExpectedErr:     `invalid ON EMPTY behavior for column "i"`,
					ExpectedErrCode: "42601",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (q JSONB PATH '$.a[*]' TRUE ON ERROR)) AS jt;`,
					ExpectedErr:     `invalid ON ERROR behavior for column "q"`,
					ExpectedErrCode: "42601",
				},
				{
					Query: `SELECT * FROM JSON_TABLE('[{"a":[1,2]},{"a":3},{"a":"s"},{"a":[]},{"a":{"k":1}},{}]'::jsonb, '$[*]' COLUMNS (w1 JSONB PATH '$.a[*]' WITH WRAPPER, w2 JSONB PATH '$.a[*]' WITH CONDITIONAL WRAPPER, w3 JSONB PATH '$.a' WITH CONDITIONAL WRAPPER, w4 JSONB PATH '$.a' WITH UNCONDITIONAL ARRAY WRAPPER, w5 JSONB PATH '$.a' WITHOUT WRAPPER, q1 TEXT PATH '$.a' OMIT QUOTES, q2 TEXT PATH '$.a' KEEP QUOTES, q3 TEXT PATH '$.a' OMIT QUOTES ON SCALAR STRING, q4 JSONB PATH '$.a' OMIT QUOTES, f1 TEXT FORMAT JSON PATH '$.a', f2 TEXT PATH '$.a' WITH WRAPPER, f3 INT PATH '$.a' KEEP QUOTES, f4 INT PATH '$.a' OMIT QUOTES)) AS jt;`,
					Expected: []sql.Row{
						{"[1, 2]", "[1, 2]", "[1, 2]", "[[1, 2]]", "[1, 2]", "[1, 2]", "[1, 2]", "[1, 2]", "[1, 2]", "[1, 2]", "[[1, 2]]", nil, nil},
						{"[3]", "3", "3", "[3]", "3", "3", "3", "3", "3", "3", "[3]", 3, 3},
						{`["s"]`, `"s"`, `"s"`, `["s"]`, `"s"`, "s", `"s"`, "s", nil, `"s"`, `["s"]`, nil, nil},
						{nil, nil, "[]", "[[]]", "[]", "[]", "[]", "[]", "[]", "[]", "[[]]", nil, nil},
						{`[{"k": 1}]`, `{"k": 1}`, `{"k": 1}`, `[{"k": 1}]`, `{"k": 1}`, `{"k": 1}`, `{"k": 1}`, `{"k": 1}`, `{"k": 1}`, `{"k": 1}`, `[{"k": 1}]`, nil, nil},
						{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
					},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (q JSONB PATH '$.a[*]' EMPTY ON ERROR, r JSONB PATH '$.a[*]' EMPTY OBJECT ON ERROR, s JSONB PATH '$.b' EMPTY ARRAY ON EMPTY, t TEXT FORMAT JSON PATH '$.b' EMPTY OBJECT ON EMPTY, u JSONB PATH '$.a[*]' DEFAULT '"z"' ON ERROR)) AS jt;`,
					Expected: []sql.Row{{"[]", "{}", "[]", "{}", `"z"`}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (q JSONB PATH '$.b' WITH WRAPPER, r JSONB PATH '$.b' WITH WRAPPER ERROR ON EMPTY)) AS jt;`,
					ExpectedErr:     `no SQL/JSON item found for specified path of column "r"`,
					ExpectedErrCode: "22035",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":"s"}]'::jsonb, '$[*]' COLUMNS (q JSONB PATH '$.a' OMIT QUOTES ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "invalid input syntax for type json",
					ExpectedErrCode: "22P02",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (q TEXT PATH '$.a' WITH WRAPPER OMIT QUOTES)) AS jt;`,
					ExpectedErr:     "SQL/JSON QUOTES behavior must not be specified when WITH WRAPPER is used",
					ExpectedErrCode: "42601",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a INT FORMAT JSON)) AS jt;`,
					ExpectedErr:     "cannot use JSON format with non-string output types",
					ExpectedErrCode: "0A000",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a JSON FORMAT JSON ENCODING UTF8 PATH '$.a')) AS jt;`,
					ExpectedErr:     "cannot set JSON encoding for non-bytea output types",
					ExpectedErrCode: "0A000",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a BYTEA FORMAT JSON ENCODING UTF16)) AS jt;`,
					ExpectedErr:     "unsupported JSON encoding",
					ExpectedErrCode: "0A000",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a BYTEA FORMAT JSON ENCODING FOO)) AS jt;`,
					ExpectedErr:     "unrecognized JSON encoding: foo",
					ExpectedErrCode: "22023",
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a INT PATH '$.a' WITHOUT ARRAY WRAPPER, b INT PATH '$.a' WITH CONDITIONAL ARRAY WRAPPER, c INT PATH '$.a' WITH ARRAY WRAPPER, d TEXT PATH '$.a' WITHOUT WRAPPER OMIT QUOTES, e TEXT PATH 'lax $.a' DEFAULT 'x' ON EMPTY ERROR ON ERROR)) AS jt;`,
					Expected: []sql.Row{{1, 1, nil, "1", "1"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"d":"2020-01-02"}]'::jsonb, '$[*]' COLUMNS (d DATE PATH '$.d', t TEXT PATH '$.d.datetime()', j JSONB PATH '$.d.datetime()')) AS jt;`,
					Expected: []sql.Row{{"2020-01-02", "2020-01-02", `"2020-01-02"`}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":[[1,2],[3,4]]},{"a":[["a"]]},{"a":[1,[2]]},{"a":[null,"3"]},{"a":[]}]'::jsonb, '$[*]' COLUMNS (a INT[], t TEXT[] PATH '$.a')) AS jt;`,
					Expected: []sql.Row{{"{{1,2},{3,4}}", "{{1,2},{3,4}}"}, {nil, "{{a}}"}, {nil, "{1,[2]}"}, {"{NULL,3}", "{NULL,3}"}, {"{}", "{}"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":"12"},{"a":"{1,2}"}]'::jsonb, '$[*]' COLUMNS (a INT[] PATH '$.a' OMIT QUOTES)) AS jt;`,
					Expected: []sql.Row{{nil}, {"{1,2}"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":{"x":1,"y":"q"}},{"a":{"x":"z"}},{"a":3}]'::jsonb, '$[*]' COLUMNS (a pair)) AS jt;`,
					Expected: []sql.Row{{"(1,q)"}, {nil}, {nil}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{}]'::jsonb, '$[*]' COLUMNS (e BOOL EXISTS PATH '$.a', i INT EXISTS PATH '$.a', t TEXT EXISTS PATH '$.a', j JSONB EXISTS PATH '$.a', a BOOL EXISTS)) AS jt;`,
					Expected: []sql.Row{{"t", 1, "true", "true", "t"}, {"f", 0, "false", "false", "f"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1},{}]'::jsonb, '$[*]' COLUMNS (e BOOL EXISTS PATH 'strict $.a', f BOOL EXISTS PATH 'strict $.a' TRUE ON ERROR, g BOOL EXISTS PATH 'strict $.a' UNKNOWN ON ERROR, h BOOL EXISTS PATH 'strict $.a' FALSE ON ERROR)) AS jt;`,
					Expected: []sql.Row{{"t", "t", "t", "t"}, {"f", "t", nil, "f"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT EXISTS PATH '$.a' TRUE ON ERROR, u INT EXISTS PATH 'strict $.a' UNKNOWN ON ERROR, x INT EXISTS PATH 'strict $.a' TRUE ON ERROR, y BOOL EXISTS PATH '$ ? (@ / 0 > 1)' ERROR ON ERROR)) AS jt;`,
					Expected: []sql.Row{{0, nil, 1, "f"}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1},{}]'::jsonb, '$[*]' COLUMNS (e BOOL EXISTS PATH 'strict $.a' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     `JSON object does not contain key "a"`,
					ExpectedErrCode: "2203A",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1},{}]'::jsonb, '$[*]' COLUMNS (e DATE EXISTS PATH '$.a')) AS jt;`,
					ExpectedErr:     "could not coerce ON ERROR expression (FALSE) to the RETURNING type",
					ExpectedErrCode: "42804",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e BOOL EXISTS PATH '$.a' NULL ON ERROR)) AS jt;`,
					ExpectedErr:     `invalid ON ERROR behavior for column "e"`,
					ExpectedErrCode: "42601",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e BOOL EXISTS PATH '$.a' EMPTY ON ERROR)) AS jt;`,
					ExpectedErr:     `invalid ON ERROR behavior for column "e"`,
					ExpectedErrCode: "42601",
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH '$ / 0')) AS jt;`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH '$ / 0' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "division by zero",
					ExpectedErrCode: "22012",
				},
			},
		},
		{
			Name: "JSON_TABLE NESTED PATH",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM JSON_TABLE('[{"a":1,"b":[10,20],"c":["x"]},{"a":2,"b":[],"c":[]},{"a":3,"b":[30],"c":["y","z"]}]'::jsonb, '$[*]' COLUMNS (n FOR ORDINALITY, a INT, NESTED PATH '$.b[*]' COLUMNS (bn FOR ORDINALITY, b INT PATH '$'), NESTED '$.c[*]' AS cp COLUMNS (cn FOR ORDINALITY, c TEXT PATH '$'))) AS jt;`,
					Expected: []sql.Row{
						{1, 1, 1, 10, nil, nil},
						{1, 1, 2, 20, nil, nil},
						{1, 1, nil, nil, 1, "x"},
						{2, 2, nil, nil, nil, nil},
						{3, 3, 1, 30, nil, nil},
						{3, 3, nil, nil, 1, "y"},
						{3, 3, nil, nil, 2, "z"},
					},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1,"b":[{"x":[1,2]},{"x":[3]}]}]'::jsonb, '$[*]' COLUMNS (a INT, NESTED PATH '$.b[*]' COLUMNS (bn FOR ORDINALITY, NESTED PATH '$.x[*]' COLUMNS (xn FOR ORDINALITY, x INT PATH '$')))) AS jt;`,
					Expected: []sql.Row{{1, 1, 1, 1}, {1, 1, 2, 2}, {1, 2, 1, 3}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1,"b":[1]}]'::jsonb, '$[*]' COLUMNS (NESTED PATH '$.b[*]' COLUMNS (b INT PATH '$'))) AS jt;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1,"b":1}]'::jsonb, '$[*]' COLUMNS (a INT, NESTED PATH 'strict $.q' COLUMNS (b INT PATH '$'))) AS jt;`,
					Expected: []sql.Row{{1, nil}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1,"b":1}]'::jsonb, '$[*]' COLUMNS (a INT, NESTED PATH 'strict $.q' COLUMNS (b INT PATH '$')) ERROR ON ERROR) AS jt;`,
					ExpectedErr:     `JSON object does not contain key "q"`,
					ExpectedErrCode: "2203A",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1,"b":[1]}]'::jsonb, '$[*]' COLUMNS (a INT, NESTED PATH '$.b[*]' COLUMNS (a INT PATH '$'))) AS jt;`,
					ExpectedErr:     "duplicate JSON_TABLE column or path name: a",
					ExpectedErrCode: "42712",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":1,"b":[1]}]'::jsonb, '$[*]' AS p COLUMNS (a INT, NESTED PATH '$.b[*]' AS p COLUMNS (b INT PATH '$'))) AS jt;`,
					ExpectedErr:     "duplicate JSON_TABLE column or path name: p",
					ExpectedErrCode: "42712",
				},
			},
		},
		{
			Name: "JSON_TABLE composition",
			SetUpScript: []string{
				"CREATE TABLE orders (id INT PRIMARY KEY, doc JSONB);",
				`INSERT INTO orders VALUES (1, '{"customer":"ann","items":[{"sku":"a","qty":2,"tags":["x","y"]},{"sku":"b","qty":1,"tags":[]}]}'), (2, '{"customer":"bob","items":[{"sku":"c","qty":5,"tags":["z"]}]}'), (3, '{"customer":"cy","items":[]}');`,
				"CREATE VIEW order_items AS SELECT o.id, jt.* FROM orders o, JSON_TABLE(o.doc, '$' AS root PASSING 1 AS minqty COLUMNS (customer TEXT, NESTED PATH '$.items[*] ? (@.qty >= $minqty)' AS items COLUMNS (n FOR ORDINALITY, sku TEXT, qty INT, big BOOL EXISTS PATH '$ ? (@.qty > 1)', NESTED PATH '$.tags[*]' COLUMNS (tag TEXT PATH '$')))) AS jt;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM order_items ORDER BY id, n, tag;",
					Expected: []sql.Row{
						{1, "ann", 1, "a", 2, "t", "x"},
						{1, "ann", 1, "a", 2, "t", "y"},
						{1, "ann", 2, "b", 1, "f", nil},
						{2, "bob", 1, "c", 5, "t", "z"},
						{3, "cy", nil, nil, nil, nil, nil},
					},
				},
				{
					Query:    "SELECT o.id, i.sku, t.tag FROM orders o, JSON_TABLE(o.doc, '$.items[*]' COLUMNS (sku TEXT, tags JSONB)) AS i, JSON_TABLE(i.tags, '$[*]' COLUMNS (tag TEXT PATH '$')) AS t ORDER BY 1, 2, 3;",
					Expected: []sql.Row{{1, "a", "x"}, {1, "a", "y"}, {2, "c", "z"}},
				},
				{
					Query:    "SELECT jt.customer, count(*) FROM orders, JSON_TABLE(orders.doc, '$' COLUMNS (customer TEXT, NESTED PATH '$.items[*]' COLUMNS (sku TEXT))) AS jt WHERE jt.sku IS NOT NULL GROUP BY jt.customer ORDER BY 1;",
					Expected: []sql.Row{{"ann", 2}, {"bob", 1}},
				},
				{
					Query:    "WITH items AS (SELECT jt.* FROM orders, JSON_TABLE(orders.doc, '$.items[*]' COLUMNS (sku TEXT, qty INT)) AS jt) SELECT sum(qty) FROM items;",
					Expected: []sql.Row{{8}},
				},
				{
					Query:    "SELECT id FROM orders WHERE EXISTS (SELECT 1 FROM JSON_TABLE(orders.doc, '$.items[*]' COLUMNS (qty INT)) AS jt WHERE jt.qty > 4);",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT (SELECT max(qty) FROM JSON_TABLE(orders.doc, '$.items[*]' COLUMNS (qty INT)) AS jt) AS m FROM orders ORDER BY id;",
					Expected: []sql.Row{{2}, {5}, {nil}},
				},
				{
					Query:    "SELECT * FROM JSON_TABLE($1::jsonb, '$[*] ? (@.a > $x)' PASSING $2::int AS x COLUMNS (a INT)) AS jt;",
					BindVars: []any{`[{"a":1},{"a":2},{"a":3}]`, 1},
					Expected: []sql.Row{{2}, {3}},
				},
				{
					Query:    "SELECT o.id, jt.qty FROM orders o, JSON_TABLE(o.doc, '$.items[*] ? (@.qty >= $m)' PASSING $1::int AS m COLUMNS (qty INT)) AS jt ORDER BY 1, 2;",
					BindVars: []any{2},
					Expected: []sql.Row{{1, 2}, {2, 5}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a INT)) AS jt WHERE a = 1 ORDER BY a DESC LIMIT 1;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":1}]', '$[*]' COLUMNS (a INT)) AS jt JOIN JSON_TABLE('[{"b":1}]', '$[*]' COLUMNS (b INT)) AS jt2 ON jt.a = jt2.b;`,
					Expected: []sql.Row{{1, 1}},
				},
				{
					Query:    `SELECT pg_typeof(a), pg_typeof(b), pg_typeof(c), pg_typeof(n) FROM JSON_TABLE('[{"a":1}]', '$[*]' COLUMNS (a INT, b JSONB PATH '$.a', c TEXT EXISTS PATH '$.a', n FOR ORDINALITY)) AS jt;`,
					Expected: []sql.Row{{"integer", "jsonb", "text", "integer"}},
				},
			},
		},
		{
			Name: "JSON_TABLE column types and nested behaviors",
			SetUpScript: []string{
				"CREATE TYPE pair AS (x INT, y TEXT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"n":"1.234","v":"abc","ts":"2020-01-02 03:04:05","d":"2020-01-02"}]'::jsonb, '$[*]' COLUMNS (n NUMERIC(5,2), v VARCHAR(5), ts TIMESTAMP, d DATE, i2 SMALLINT PATH '$.n' DEFAULT 7 ON ERROR, f FLOAT8 PATH '$.n')) AS jt;`,
					Expected: []sql.Row{{Numeric("1.23"), "abc", "2020-01-02 03:04:05", "2020-01-02", 7, 1.234}},
				},
				{
					Query:    `SELECT a, b::text, c::text FROM JSON_TABLE('[1]'::bytea FORMAT JSON ENCODING UTF8, '$[*]' COLUMNS (a INT PATH '$', b BYTEA FORMAT JSON PATH '$', c BYTEA PATH '$')) AS jt;`,
					Expected: []sql.Row{{1, `\x31`, `\x31`}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('{"é":"ü"}'::jsonb, '$' COLUMNS ("é" TEXT)) AS jt;`,
					Expected: []sql.Row{{"ü"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"k":{"a":1}},{"k":2}]'::jsonb, '$[*]' COLUMNS (kv TEXT PATH '$.k.keyvalue().key', sz INT PATH '$.k.size()', t TEXT PATH '$.k.type()', dbl FLOAT8 PATH '$.k.double()')) AS jt;`,
					Expected: []sql.Row{{"a", 1, "object", nil}, {nil, 1, "number", 2.0}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $.a' DEFAULT -1 ON ERROR, f INT PATH '$.a' DEFAULT -2 ON EMPTY)) AS jt;`,
					Expected: []sql.Row{{-1, -2}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":{"x":1,"y":"q"}}]'::jsonb, '$[*]' COLUMNS (a pair[] PATH '$[*].a' WITH WRAPPER, b TEXT PATH '$.a.y')) AS jt;`,
					Expected: []sql.Row{{`{"(1,q)"}`, "q"}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":[1,2],"b":[3]}]'::jsonb, '$[*]' COLUMNS (NESTED PATH '$.a[*]' COLUMNS (a INT PATH '$', ae BOOL EXISTS PATH '$ ? (@ > 1)', aj JSONB PATH '$' WITH WRAPPER), NESTED PATH '$.b[*]' COLUMNS (b INT PATH '$' DEFAULT 0 ON ERROR))) AS jt;`,
					Expected: []sql.Row{{1, "f", "[1]", nil}, {2, "t", "[2]", nil}, {nil, nil, nil, 3}},
				},
				{
					Query:    `SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' PASSING 2 AS lim COLUMNS (NESTED PATH '$.a[*] ? (@ < $lim)' COLUMNS (a INT PATH '$'))) AS jt;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (NESTED PATH '$.a[*]' COLUMNS (a INT PATH 'strict $.x' ERROR ON ERROR))) AS jt;`,
					ExpectedErr:     "jsonpath member accessor can only be applied to an object",
					ExpectedErrCode: "2203A",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":{"k":1}}]'::jsonb, '$[*]' COLUMNS (a INT[] ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "expected JSON array",
					ExpectedErrCode: "22P02",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":[[1],2]}]'::jsonb, '$[*]' COLUMNS (a INT[] ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "expected JSON array",
					ExpectedErrCode: "22P02",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":3}]'::jsonb, '$[*]' COLUMNS (a pair ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "cannot call populate_composite on a scalar",
					ExpectedErrCode: "22023",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[{"a":[1]}]'::jsonb, '$[*]' COLUMNS (a pair ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "cannot call populate_composite on an array",
					ExpectedErrCode: "22023",
				},
			},
		},
		{
			Name: "JSON_TABLE path error codes",
			Assertions: []ScriptTestAssertion{
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $.a' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath member accessor can only be applied to an object",
					ExpectedErrCode: "2203A",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $.*' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath wildcard member accessor can only be applied to an object",
					ExpectedErrCode: "2203C",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $[0]' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath array accessor can only be applied to an array",
					ExpectedErrCode: "22039",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[[1]]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $[5]' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath array subscript is out of bounds",
					ExpectedErrCode: "22033",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('["a"]'::jsonb, '$[*]' COLUMNS (e INT PATH '-$' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "operand of unary jsonpath operator - is not a numeric value",
					ExpectedErrCode: "2203B",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('["a"]'::jsonb, '$[*]' COLUMNS (e INT PATH '$ + 1' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "left operand of jsonpath operator + is not a single numeric value",
					ExpectedErrCode: "22038",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('["a"]'::jsonb, '$[*]' COLUMNS (e INT PATH '$.abs()' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath item method .abs() can only be applied to a numeric value",
					ExpectedErrCode: "22036",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $.size()' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath item method .size() can only be applied to an array",
					ExpectedErrCode: "22039",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e JSONB PATH '$.keyvalue()' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath item method .keyvalue() can only be applied to an object",
					ExpectedErrCode: "2203C",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e TEXT PATH '$.datetime()' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath item method .datetime() can only be applied to a string",
					ExpectedErrCode: "22031",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('["x"]'::jsonb, '$[*]' COLUMNS (e TEXT PATH '$.datetime()' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     `datetime format is not recognized: "x"`,
					ExpectedErrCode: "22031",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('["x"]'::jsonb, '$[*]' COLUMNS (e FLOAT8 PATH '$.double()' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     `argument "x" of jsonpath item method .double() is invalid for type double precision`,
					ExpectedErrCode: "22036",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[[]]'::jsonb, '$[*]' COLUMNS (e FLOAT8 PATH 'strict $.double()' ERROR ON ERROR)) AS jt;`,
					ExpectedErr:     "jsonpath item method .double() can only be applied to a string or numeric value",
					ExpectedErrCode: "22036",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, 'strict $.a' COLUMNS (e INT PATH '$') ERROR ON ERROR) AS jt;`,
					ExpectedErr:     "jsonpath member accessor can only be applied to an object",
					ExpectedErrCode: "2203A",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '$$' COLUMNS (a INT)) AS jt;`,
					ExpectedErr:     "jsonpath input",
					ExpectedErrCode: "42601",
				},
				{
					Query:           `SELECT * FROM JSON_TABLE('[1]'::jsonb, '@' COLUMNS (a INT)) AS jt;`,
					ExpectedErr:     "@ is not allowed in root expressions",
					ExpectedErrCode: "42601",
				},
			},
		},
	})
}
