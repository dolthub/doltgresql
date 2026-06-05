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
