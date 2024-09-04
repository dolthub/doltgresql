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

func TestJsonpathEncoding1(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_jsonpath_encoding_1)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_jsonpath_encoding_1,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT getdatabaseencoding() NOT IN ('UTF8', 'SQL_ASCII')
       AS skip_test \gset
\if :skip_test
\quit
\endif
SELECT getdatabaseencoding();           -- just to label the results files`,
				Results: []sql.Row{{`SQL_ASCII`}},
			},
			{
				Statement:   `SELECT '"\u"'::jsonpath;		-- ERROR, incomplete escape`,
				ErrorString: `invalid unicode sequence at or near "\u" of jsonpath input`,
			},
			{
				Statement:   `SELECT '"\u00"'::jsonpath;		-- ERROR, incomplete escape`,
				ErrorString: `invalid unicode sequence at or near "\u00" of jsonpath input`,
			},
			{
				Statement:   `SELECT '"\u000g"'::jsonpath;	-- ERROR, g is not a hex digit`,
				ErrorString: `invalid unicode sequence at or near "\u000" of jsonpath input`,
			},
			{
				Statement:   `SELECT '"\u0000"'::jsonpath;	-- OK, legal escape`,
				ErrorString: `unsupported Unicode escape sequence`,
			},
			{
				Statement:   `SELECT '"\uaBcD"'::jsonpath;	-- OK, uppercase and lower case both OK`,
				ErrorString: `conversion between UTF8 and SQL_ASCII is not supported`,
			},
			{
				Statement:   `select '"\ud83d\ude04\ud83d\udc36"'::jsonpath as correct_in_utf8;`,
				ErrorString: `conversion between UTF8 and SQL_ASCII is not supported`,
			},
			{
				Statement:   `select '"\ud83d\ud83d"'::jsonpath; -- 2 high surrogates in a row`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement:   `select '"\ude04\ud83d"'::jsonpath; -- surrogates in wrong order`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement:   `select '"\ud83dX"'::jsonpath; -- orphan high surrogate`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement:   `select '"\ude04X"'::jsonpath; -- orphan low surrogate`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement:   `select '"the Copyright \u00a9 sign"'::jsonpath as correct_in_utf8;`,
				ErrorString: `conversion between UTF8 and SQL_ASCII is not supported`,
			},
			{
				Statement: `select '"dollar \u0024 character"'::jsonpath as correct_everywhere;`,
				Results:   []sql.Row{{"dollar $ character"}},
			},
			{
				Statement: `select '"dollar \\u0024 character"'::jsonpath as not_an_escape;`,
				Results:   []sql.Row{{"dollar \\u0024 character"}},
			},
			{
				Statement:   `select '"null \u0000 escape"'::jsonpath as not_unescaped;`,
				ErrorString: `unsupported Unicode escape sequence`,
			},
			{
				Statement: `select '"null \\u0000 escape"'::jsonpath as not_an_escape;`,
				Results:   []sql.Row{{"null \\u0000 escape"}},
			},
			{
				Statement:   `SELECT '$."\u"'::jsonpath;		-- ERROR, incomplete escape`,
				ErrorString: `invalid unicode sequence at or near "\u" of jsonpath input`,
			},
			{
				Statement:   `SELECT '$."\u00"'::jsonpath;	-- ERROR, incomplete escape`,
				ErrorString: `invalid unicode sequence at or near "\u00" of jsonpath input`,
			},
			{
				Statement:   `SELECT '$."\u000g"'::jsonpath;	-- ERROR, g is not a hex digit`,
				ErrorString: `invalid unicode sequence at or near "\u000" of jsonpath input`,
			},
			{
				Statement:   `SELECT '$."\u0000"'::jsonpath;	-- OK, legal escape`,
				ErrorString: `unsupported Unicode escape sequence`,
			},
			{
				Statement:   `SELECT '$."\uaBcD"'::jsonpath;	-- OK, uppercase and lower case both OK`,
				ErrorString: `conversion between UTF8 and SQL_ASCII is not supported`,
			},
			{
				Statement:   `select '$."\ud83d\ude04\ud83d\udc36"'::jsonpath as correct_in_utf8;`,
				ErrorString: `conversion between UTF8 and SQL_ASCII is not supported`,
			},
			{
				Statement:   `select '$."\ud83d\ud83d"'::jsonpath; -- 2 high surrogates in a row`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement:   `select '$."\ude04\ud83d"'::jsonpath; -- surrogates in wrong order`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement:   `select '$."\ud83dX"'::jsonpath; -- orphan high surrogate`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement:   `select '$."\ude04X"'::jsonpath; -- orphan low surrogate`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement:   `select '$."the Copyright \u00a9 sign"'::jsonpath as correct_in_utf8;`,
				ErrorString: `conversion between UTF8 and SQL_ASCII is not supported`,
			},
			{
				Statement: `select '$."dollar \u0024 character"'::jsonpath as correct_everywhere;`,
				Results:   []sql.Row{{`$."dollar $ character"`}},
			},
			{
				Statement: `select '$."dollar \\u0024 character"'::jsonpath as not_an_escape;`,
				Results:   []sql.Row{{`$."dollar \\u0024 character"`}},
			},
			{
				Statement:   `select '$."null \u0000 escape"'::jsonpath as not_unescaped;`,
				ErrorString: `unsupported Unicode escape sequence`,
			},
			{
				Statement: `select '$."null \\u0000 escape"'::jsonpath as not_an_escape;`,
				Results:   []sql.Row{{`$."null \\u0000 escape"`}},
			},
		},
	})
}
