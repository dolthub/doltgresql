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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestJsonEncoding(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_json_encoding)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_json_encoding,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT getdatabaseencoding() NOT IN ('UTF8', 'SQL_ASCII')
       AS skip_test \gset
\if :skip_test
\quit
\endif
SELECT getdatabaseencoding();           -- just to label the results files`,
				Results: []sql.Row{{`UTF8`}},
			},
			{
				Statement:   `SELECT '"\u"'::json;			-- ERROR, incomplete escape`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: "\u"
SELECT '"\u00"'::json;			-- ERROR, incomplete escape`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: "\u00"
SELECT '"\u000g"'::json;		-- ERROR, g is not a hex digit`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: "\u000g...
SELECT '"\u0000"'::json;		-- OK, legal escape`,
				Results: []sql.Row{{"\u0000"}},
			},
			{
				Statement: `SELECT '"\uaBcD"'::json;		-- OK, uppercase and lower case both OK`,
				Results:   []sql.Row{{"\uaBcD"}},
			},
			{
				Statement: `select json '{ "a":  "\ud83d\ude04\ud83d\udc36" }' -> 'a' as correct_in_utf8;`,
				Results:   []sql.Row{{`\ud83d\ude04\ud83d\udc36`}},
				Skip:      true,
			},
			{
				Statement:   `select json '{ "a":  "\ud83d\ud83d" }' -> 'a'; -- 2 high surrogates in a row`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "\ud83d\ud83d...
select json '{ "a":  "\ude04\ud83d" }' -> 'a'; -- surrogates in wrong order`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "\ude04...
select json '{ "a":  "\ud83dX" }' -> 'a'; -- orphan high surrogate`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "\ud83dX...
select json '{ "a":  "\ude04X" }' -> 'a'; -- orphan low surrogate`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "\ude04...
select json '{ "a":  "the Copyright \u00a9 sign" }' as correct_in_utf8;`,
				Results: []sql.Row{{`{ "a":  "the Copyright \u00a9 sign" }`}},
			},
			{
				Statement: `select json '{ "a":  "dollar \u0024 character" }' as correct_everywhere;`,
				Results:   []sql.Row{{`{ "a":  "dollar \u0024 character" }`}},
			},
			{
				Statement: `select json '{ "a":  "dollar \\u0024 character" }' as not_an_escape;`,
				Results:   []sql.Row{{`{ "a":  "dollar \\u0024 character" }`}},
			},
			{
				Statement: `select json '{ "a":  "null \u0000 escape" }' as not_unescaped;`,
				Results:   []sql.Row{{`{ "a":  "null \u0000 escape" }`}},
			},
			{
				Statement: `select json '{ "a":  "null \\u0000 escape" }' as not_an_escape;`,
				Results:   []sql.Row{{`{ "a":  "null \\u0000 escape" }`}},
			},
			{
				Statement: `select json '{ "a":  "the Copyright \u00a9 sign" }' ->> 'a' as correct_in_utf8;`,
				Results:   []sql.Row{{`the Copyright © sign`}},
			},
			{
				Statement: `select json '{ "a":  "dollar \u0024 character" }' ->> 'a' as correct_everywhere;`,
				Results:   []sql.Row{{`dollar $ character`}},
			},
			{
				Statement: `select json '{ "a":  "dollar \\u0024 character" }' ->> 'a' as not_an_escape;`,
				Results:   []sql.Row{{`dollar \u0024 character`}},
			},
			{
				Statement:   `select json '{ "a":  "null \u0000 escape" }' ->> 'a' as fails;`,
				ErrorString: `unsupported Unicode escape sequence`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "null \u0000...
select json '{ "a":  "null \\u0000 escape" }' ->> 'a' as not_an_escape;`,
				Results: []sql.Row{{`null \u0000 escape`}},
			},
			{
				Statement:   `SELECT '"\u"'::jsonb;			-- ERROR, incomplete escape`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: "\u"
SELECT '"\u00"'::jsonb;			-- ERROR, incomplete escape`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: "\u00"
SELECT '"\u000g"'::jsonb;		-- ERROR, g is not a hex digit`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: "\u000g...
SELECT '"\u0045"'::jsonb;		-- OK, legal escape`,
				Results: []sql.Row{{"E"}},
			},
			{
				Statement:   `SELECT '"\u0000"'::jsonb;		-- ERROR, we don't support U+0000`,
				ErrorString: `unsupported Unicode escape sequence`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: "\u0000...
SELECT octet_length('"\uaBcD"'::jsonb::text); -- OK, uppercase and lower case both OK`,
				Results: []sql.Row{{5}},
			},
			{
				Statement: `SELECT octet_length((jsonb '{ "a":  "\ud83d\ude04\ud83d\udc36" }' -> 'a')::text) AS correct_in_utf8;`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement:   `SELECT jsonb '{ "a":  "\ud83d\ud83d" }' -> 'a'; -- 2 high surrogates in a row`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "\ud83d\ud83d...
SELECT jsonb '{ "a":  "\ude04\ud83d" }' -> 'a'; -- surrogates in wrong order`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "\ude04...
SELECT jsonb '{ "a":  "\ud83dX" }' -> 'a'; -- orphan high surrogate`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "\ud83dX...
SELECT jsonb '{ "a":  "\ude04X" }' -> 'a'; -- orphan low surrogate`,
				ErrorString: `invalid input syntax for type json`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "\ude04...
SELECT jsonb '{ "a":  "the Copyright \u00a9 sign" }' as correct_in_utf8;`,
				Results: []sql.Row{{`{"a": "the Copyright © sign"}`}},
			},
			{
				Statement: `SELECT jsonb '{ "a":  "dollar \u0024 character" }' as correct_everywhere;`,
				Results:   []sql.Row{{`{"a": "dollar $ character"}`}},
			},
			{
				Statement: `SELECT jsonb '{ "a":  "dollar \\u0024 character" }' as not_an_escape;`,
				Results:   []sql.Row{{`{"a": "dollar \\u0024 character"}`}},
			},
			{
				Statement:   `SELECT jsonb '{ "a":  "null \u0000 escape" }' as fails;`,
				ErrorString: `unsupported Unicode escape sequence`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "null \u0000...
SELECT jsonb '{ "a":  "null \\u0000 escape" }' as not_an_escape;`,
				Results: []sql.Row{{`{"a": "null \\u0000 escape"}`}},
			},
			{
				Statement: `SELECT jsonb '{ "a":  "the Copyright \u00a9 sign" }' ->> 'a' as correct_in_utf8;`,
				Results:   []sql.Row{{`the Copyright © sign`}},
			},
			{
				Statement: `SELECT jsonb '{ "a":  "dollar \u0024 character" }' ->> 'a' as correct_everywhere;`,
				Results:   []sql.Row{{`dollar $ character`}},
			},
			{
				Statement: `SELECT jsonb '{ "a":  "dollar \\u0024 character" }' ->> 'a' as not_an_escape;`,
				Results:   []sql.Row{{`dollar \u0024 character`}},
			},
			{
				Statement:   `SELECT jsonb '{ "a":  "null \u0000 escape" }' ->> 'a' as fails;`,
				ErrorString: `unsupported Unicode escape sequence`,
			},
			{
				Statement: `CONTEXT:  JSON data, line 1: { "a":  "null \u0000...
SELECT jsonb '{ "a":  "null \\u0000 escape" }' ->> 'a' as not_an_escape;`,
				Results: []sql.Row{{`null \u0000 escape`}},
			},
		},
	})
}
