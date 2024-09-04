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

func TestJsonpath(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_jsonpath)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_jsonpath,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `select ''::jsonpath;`,
				ErrorString: `invalid input syntax for type jsonpath: ""`,
			},
			{
				Statement: `select '$'::jsonpath;`,
				Results:   []sql.Row{{`$`}},
			},
			{
				Statement: `select 'strict $'::jsonpath;`,
				Results:   []sql.Row{{`strict $`}},
			},
			{
				Statement: `select 'lax $'::jsonpath;`,
				Results:   []sql.Row{{`$`}},
			},
			{
				Statement: `select '$.a'::jsonpath;`,
				Results:   []sql.Row{{`$."a"`}},
			},
			{
				Statement: `select '$.a.v'::jsonpath;`,
				Results:   []sql.Row{{`$."a"."v"`}},
			},
			{
				Statement: `select '$.a.*'::jsonpath;`,
				Results:   []sql.Row{{`$."a".*`}},
			},
			{
				Statement: `select '$.*[*]'::jsonpath;`,
				Results:   []sql.Row{{`$.*[*]`}},
			},
			{
				Statement: `select '$.a[*]'::jsonpath;`,
				Results:   []sql.Row{{`$."a"[*]`}},
			},
			{
				Statement: `select '$.a[*][*]'::jsonpath;`,
				Results:   []sql.Row{{`$."a"[*][*]`}},
			},
			{
				Statement: `select '$[*]'::jsonpath;`,
				Results:   []sql.Row{{`$[*]`}},
			},
			{
				Statement: `select '$[0]'::jsonpath;`,
				Results:   []sql.Row{{`$[0]`}},
			},
			{
				Statement: `select '$[*][0]'::jsonpath;`,
				Results:   []sql.Row{{`$[*][0]`}},
			},
			{
				Statement: `select '$[*].a'::jsonpath;`,
				Results:   []sql.Row{{`$[*]."a"`}},
			},
			{
				Statement: `select '$[*][0].a.b'::jsonpath;`,
				Results:   []sql.Row{{`$[*][0]."a"."b"`}},
			},
			{
				Statement: `select '$.a.**.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a".**."b"`}},
			},
			{
				Statement: `select '$.a.**{2}.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a".**{2}."b"`}},
			},
			{
				Statement: `select '$.a.**{2 to 2}.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a".**{2}."b"`}},
			},
			{
				Statement: `select '$.a.**{2 to 5}.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a".**{2 to 5}."b"`}},
			},
			{
				Statement: `select '$.a.**{0 to 5}.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a".**{0 to 5}."b"`}},
			},
			{
				Statement: `select '$.a.**{5 to last}.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a".**{5 to last}."b"`}},
			},
			{
				Statement: `select '$.a.**{last}.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a".**{last}."b"`}},
			},
			{
				Statement: `select '$.a.**{last to 5}.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a".**{last to 5}."b"`}},
			},
			{
				Statement: `select '$+1'::jsonpath;`,
				Results:   []sql.Row{{`($ + 1)`}},
			},
			{
				Statement: `select '$-1'::jsonpath;`,
				Results:   []sql.Row{{`($ - 1)`}},
			},
			{
				Statement: `select '$--+1'::jsonpath;`,
				Results:   []sql.Row{{`($ - -1)`}},
			},
			{
				Statement: `select '$.a/+-1'::jsonpath;`,
				Results:   []sql.Row{{`($."a" / -1)`}},
			},
			{
				Statement: `select '1 * 2 + 4 % -3 != false'::jsonpath;`,
				Results:   []sql.Row{{`(1 * 2 + 4 % -3 != false)`}},
			},
			{
				Statement: `select '"\b\f\r\n\t\v\"\''\\"'::jsonpath;`,
				Results:   []sql.Row{{"\b\f\r\n\t\u000b\"'\\"}},
			},
			{
				Statement: `select '"\x50\u0067\u{53}\u{051}\u{00004C}"'::jsonpath;`,
				Results:   []sql.Row{{"PgSQL"}},
			},
			{
				Statement: `select '$.foo\x50\u0067\u{53}\u{051}\u{00004C}\t\"bar'::jsonpath;`,
				Results:   []sql.Row{{`$."fooPgSQL\t\"bar"`}},
			},
			{
				Statement: `select '"\z"'::jsonpath;  -- unrecognized escape is just the literal char`,
				Results:   []sql.Row{{"z"}},
			},
			{
				Statement: `select '$.g ? ($.a == 1)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?($."a" == 1)`}},
			},
			{
				Statement: `select '$.g ? (@ == 1)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@ == 1)`}},
			},
			{
				Statement: `select '$.g ? (@.a == 1)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@."a" == 1)`}},
			},
			{
				Statement: `select '$.g ? (@.a == 1 || @.a == 4)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@."a" == 1 || @."a" == 4)`}},
			},
			{
				Statement: `select '$.g ? (@.a == 1 && @.a == 4)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@."a" == 1 && @."a" == 4)`}},
			},
			{
				Statement: `select '$.g ? (@.a == 1 || @.a == 4 && @.b == 7)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@."a" == 1 || @."a" == 4 && @."b" == 7)`}},
			},
			{
				Statement: `select '$.g ? (@.a == 1 || !(@.a == 4) && @.b == 7)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@."a" == 1 || !(@."a" == 4) && @."b" == 7)`}},
			},
			{
				Statement: `select '$.g ? (@.a == 1 || !(@.x >= 123 || @.a == 4) && @.b == 7)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@."a" == 1 || !(@."x" >= 123 || @."a" == 4) && @."b" == 7)`}},
			},
			{
				Statement: `select '$.g ? (@.x >= @[*]?(@.a > "abc"))'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@."x" >= @[*]?(@."a" > "abc"))`}},
			},
			{
				Statement: `select '$.g ? ((@.x >= 123 || @.a == 4) is unknown)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?((@."x" >= 123 || @."a" == 4) is unknown)`}},
			},
			{
				Statement: `select '$.g ? (exists (@.x))'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(exists (@."x"))`}},
			},
			{
				Statement: `select '$.g ? (exists (@.x ? (@ == 14)))'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(exists (@."x"?(@ == 14)))`}},
			},
			{
				Statement: `select '$.g ? ((@.x >= 123 || @.a == 4) && exists (@.x ? (@ == 14)))'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?((@."x" >= 123 || @."a" == 4) && exists (@."x"?(@ == 14)))`}},
			},
			{
				Statement: `select '$.g ? (+@.x >= +-(+@.a + 2))'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(+@."x" >= +(-(+@."a" + 2)))`}},
			},
			{
				Statement: `select '$a'::jsonpath;`,
				Results:   []sql.Row{{`$"a"`}},
			},
			{
				Statement: `select '$a.b'::jsonpath;`,
				Results:   []sql.Row{{`$"a"."b"`}},
			},
			{
				Statement: `select '$a[*]'::jsonpath;`,
				Results:   []sql.Row{{`$"a"[*]`}},
			},
			{
				Statement: `select '$.g ? (@.zip == $zip)'::jsonpath;`,
				Results:   []sql.Row{{`$."g"?(@."zip" == $"zip")`}},
			},
			{
				Statement: `select '$.a[1,2, 3 to 16]'::jsonpath;`,
				Results:   []sql.Row{{`$."a"[1,2,3 to 16]`}},
			},
			{
				Statement: `select '$.a[$a + 1, ($b[*]) to -($[0] * 2)]'::jsonpath;`,
				Results:   []sql.Row{{`$."a"[$"a" + 1,$"b"[*] to -($[0] * 2)]`}},
			},
			{
				Statement: `select '$.a[$.a.size() - 3]'::jsonpath;`,
				Results:   []sql.Row{{`$."a"[$."a".size() - 3]`}},
			},
			{
				Statement:   `select 'last'::jsonpath;`,
				ErrorString: `LAST is allowed only in array subscripts`,
			},
			{
				Statement: `select '"last"'::jsonpath;`,
				Results:   []sql.Row{{"last"}},
			},
			{
				Statement: `select '$.last'::jsonpath;`,
				Results:   []sql.Row{{`$."last"`}},
			},
			{
				Statement:   `select '$ ? (last > 0)'::jsonpath;`,
				ErrorString: `LAST is allowed only in array subscripts`,
			},
			{
				Statement: `select '$[last]'::jsonpath;`,
				Results:   []sql.Row{{`$[last]`}},
			},
			{
				Statement: `select '$[$[0] ? (last > 0)]'::jsonpath;`,
				Results:   []sql.Row{{`$[$[0]?(last > 0)]`}},
			},
			{
				Statement: `select 'null.type()'::jsonpath;`,
				Results:   []sql.Row{{`null.type()`}},
			},
			{
				Statement:   `select '1.type()'::jsonpath;`,
				ErrorString: `trailing junk after numeric literal at or near "1.t" of jsonpath input`,
			},
			{
				Statement: `select '(1).type()'::jsonpath;`,
				Results:   []sql.Row{{`(1).type()`}},
			},
			{
				Statement: `select '1.2.type()'::jsonpath;`,
				Results:   []sql.Row{{`(1.2).type()`}},
			},
			{
				Statement: `select '"aaa".type()'::jsonpath;`,
				Results:   []sql.Row{{`"aaa".type()`}},
			},
			{
				Statement: `select 'true.type()'::jsonpath;`,
				Results:   []sql.Row{{`true.type()`}},
			},
			{
				Statement: `select '$.double().floor().ceiling().abs()'::jsonpath;`,
				Results:   []sql.Row{{`$.double().floor().ceiling().abs()`}},
			},
			{
				Statement: `select '$.keyvalue().key'::jsonpath;`,
				Results:   []sql.Row{{`$.keyvalue()."key"`}},
			},
			{
				Statement: `select '$.datetime()'::jsonpath;`,
				Results:   []sql.Row{{`$.datetime()`}},
			},
			{
				Statement: `select '$.datetime("datetime template")'::jsonpath;`,
				Results:   []sql.Row{{`$.datetime("datetime template")`}},
			},
			{
				Statement: `select '$ ? (@ starts with "abc")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ starts with "abc")`}},
			},
			{
				Statement: `select '$ ? (@ starts with $var)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ starts with $"var")`}},
			},
			{
				Statement:   `select '$ ? (@ like_regex "(invalid pattern")'::jsonpath;`,
				ErrorString: `invalid regular expression: parentheses () not balanced`,
			},
			{
				Statement: `select '$ ? (@ like_regex "pattern")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ like_regex "pattern")`}},
			},
			{
				Statement: `select '$ ? (@ like_regex "pattern" flag "")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ like_regex "pattern")`}},
			},
			{
				Statement: `select '$ ? (@ like_regex "pattern" flag "i")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ like_regex "pattern" flag "i")`}},
			},
			{
				Statement: `select '$ ? (@ like_regex "pattern" flag "is")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ like_regex "pattern" flag "is")`}},
			},
			{
				Statement: `select '$ ? (@ like_regex "pattern" flag "isim")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ like_regex "pattern" flag "ism")`}},
			},
			{
				Statement:   `select '$ ? (@ like_regex "pattern" flag "xsms")'::jsonpath;`,
				ErrorString: `XQuery "x" flag (expanded regular expressions) is not implemented`,
			},
			{
				Statement: `select '$ ? (@ like_regex "pattern" flag "q")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ like_regex "pattern" flag "q")`}},
			},
			{
				Statement: `select '$ ? (@ like_regex "pattern" flag "iq")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ like_regex "pattern" flag "iq")`}},
			},
			{
				Statement: `select '$ ? (@ like_regex "pattern" flag "smixq")'::jsonpath;`,
				Results:   []sql.Row{{`$?(@ like_regex "pattern" flag "ismxq")`}},
			},
			{
				Statement:   `select '$ ? (@ like_regex "pattern" flag "a")'::jsonpath;`,
				ErrorString: `invalid input syntax for type jsonpath`,
			},
			{
				Statement: `select '$ < 1'::jsonpath;`,
				Results:   []sql.Row{{`($ < 1)`}},
			},
			{
				Statement: `select '($ < 1) || $.a.b <= $x'::jsonpath;`,
				Results:   []sql.Row{{`($ < 1 || $."a"."b" <= $"x")`}},
			},
			{
				Statement:   `select '@ + 1'::jsonpath;`,
				ErrorString: `@ is not allowed in root expressions`,
			},
			{
				Statement: `select '($).a.b'::jsonpath;`,
				Results:   []sql.Row{{`$."a"."b"`}},
			},
			{
				Statement: `select '($.a.b).c.d'::jsonpath;`,
				Results:   []sql.Row{{`$."a"."b"."c"."d"`}},
			},
			{
				Statement: `select '($.a.b + -$.x.y).c.d'::jsonpath;`,
				Results:   []sql.Row{{`($."a"."b" + -$."x"."y")."c"."d"`}},
			},
			{
				Statement: `select '(-+$.a.b).c.d'::jsonpath;`,
				Results:   []sql.Row{{`(-(+$."a"."b"))."c"."d"`}},
			},
			{
				Statement: `select '1 + ($.a.b + 2).c.d'::jsonpath;`,
				Results:   []sql.Row{{`(1 + ($."a"."b" + 2)."c"."d")`}},
			},
			{
				Statement: `select '1 + ($.a.b > 2).c.d'::jsonpath;`,
				Results:   []sql.Row{{`(1 + ($."a"."b" > 2)."c"."d")`}},
			},
			{
				Statement: `select '($)'::jsonpath;`,
				Results:   []sql.Row{{`$`}},
			},
			{
				Statement: `select '(($))'::jsonpath;`,
				Results:   []sql.Row{{`$`}},
			},
			{
				Statement: `select '((($ + 1)).a + ((2)).b ? ((((@ > 1)) || (exists(@.c)))))'::jsonpath;`,
				Results:   []sql.Row{{`(($ + 1)."a" + (2)."b"?(@ > 1 || exists (@."c")))`}},
			},
			{
				Statement: `select '$ ? (@.a < 1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < .1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -.1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +.1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < 0.1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -0.1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +0.1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < 10.1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 10.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -10.1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -10.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +10.1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 10.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < 1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 10)`}},
			},
			{
				Statement: `select '$ ? (@.a < -1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -10)`}},
			},
			{
				Statement: `select '$ ? (@.a < +1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 10)`}},
			},
			{
				Statement: `select '$ ? (@.a < .1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -.1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +.1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < 0.1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -0.1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +0.1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < 10.1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 101)`}},
			},
			{
				Statement: `select '$ ? (@.a < -10.1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -101)`}},
			},
			{
				Statement: `select '$ ? (@.a < +10.1e1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 101)`}},
			},
			{
				Statement: `select '$ ? (@.a < 1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.1)`}},
			},
			{
				Statement: `select '$ ? (@.a < .1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < -.1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -0.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < +.1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < 0.1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < -0.1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -0.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < +0.1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 0.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < 10.1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < -10.1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -1.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < +10.1e-1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1.01)`}},
			},
			{
				Statement: `select '$ ? (@.a < 1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 10)`}},
			},
			{
				Statement: `select '$ ? (@.a < -1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -10)`}},
			},
			{
				Statement: `select '$ ? (@.a < +1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 10)`}},
			},
			{
				Statement: `select '$ ? (@.a < .1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -.1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +.1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < 0.1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < -0.1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -1)`}},
			},
			{
				Statement: `select '$ ? (@.a < +0.1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 1)`}},
			},
			{
				Statement: `select '$ ? (@.a < 10.1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 101)`}},
			},
			{
				Statement: `select '$ ? (@.a < -10.1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < -101)`}},
			},
			{
				Statement: `select '$ ? (@.a < +10.1e+1)'::jsonpath;`,
				Results:   []sql.Row{{`$?(@."a" < 101)`}},
			},
			{
				Statement: `select '0'::jsonpath;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `select '00'::jsonpath;`,
				ErrorString: `trailing junk after numeric literal at or near "00" of jsonpath input`,
			},
			{
				Statement: `select '0.0'::jsonpath;`,
				Results:   []sql.Row{{0.0}},
			},
			{
				Statement: `select '0.000'::jsonpath;`,
				Results:   []sql.Row{{0.000}},
			},
			{
				Statement: `select '0.000e1'::jsonpath;`,
				Results:   []sql.Row{{0.00}},
			},
			{
				Statement: `select '0.000e2'::jsonpath;`,
				Results:   []sql.Row{{0.0}},
			},
			{
				Statement: `select '0.000e3'::jsonpath;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select '0.0010'::jsonpath;`,
				Results:   []sql.Row{{0.0010}},
			},
			{
				Statement: `select '0.0010e-1'::jsonpath;`,
				Results:   []sql.Row{{0.00010}},
			},
			{
				Statement: `select '0.0010e+1'::jsonpath;`,
				Results:   []sql.Row{{0.010}},
			},
			{
				Statement: `select '0.0010e+2'::jsonpath;`,
				Results:   []sql.Row{{0.10}},
			},
			{
				Statement: `select '.001'::jsonpath;`,
				Results:   []sql.Row{{0.001}},
			},
			{
				Statement: `select '.001e1'::jsonpath;`,
				Results:   []sql.Row{{0.01}},
			},
			{
				Statement: `select '1.'::jsonpath;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select '1.e1'::jsonpath;`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement:   `select '1a'::jsonpath;`,
				ErrorString: `trailing junk after numeric literal at or near "1a" of jsonpath input`,
			},
			{
				Statement:   `select '1e'::jsonpath;`,
				ErrorString: `trailing junk after numeric literal at or near "1e" of jsonpath input`,
			},
			{
				Statement:   `select '1.e'::jsonpath;`,
				ErrorString: `trailing junk after numeric literal at or near "1.e" of jsonpath input`,
			},
			{
				Statement:   `select '1.2a'::jsonpath;`,
				ErrorString: `trailing junk after numeric literal at or near "1.2a" of jsonpath input`,
			},
			{
				Statement:   `select '1.2e'::jsonpath;`,
				ErrorString: `trailing junk after numeric literal at or near "1.2e" of jsonpath input`,
			},
			{
				Statement: `select '1.2.e'::jsonpath;`,
				Results:   []sql.Row{{`(1.2)."e"`}},
			},
			{
				Statement: `select '(1.2).e'::jsonpath;`,
				Results:   []sql.Row{{`(1.2)."e"`}},
			},
			{
				Statement: `select '1e3'::jsonpath;`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `select '1.e3'::jsonpath;`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `select '1.e3.e'::jsonpath;`,
				Results:   []sql.Row{{`(1000)."e"`}},
			},
			{
				Statement: `select '1.e3.e4'::jsonpath;`,
				Results:   []sql.Row{{`(1000)."e4"`}},
			},
			{
				Statement: `select '1.2e3'::jsonpath;`,
				Results:   []sql.Row{{1200}},
			},
			{
				Statement:   `select '1.2e3a'::jsonpath;`,
				ErrorString: `trailing junk after numeric literal at or near "1.2e3a" of jsonpath input`,
			},
			{
				Statement: `select '1.2.e3'::jsonpath;`,
				Results:   []sql.Row{{`(1.2)."e3"`}},
			},
			{
				Statement: `select '(1.2).e3'::jsonpath;`,
				Results:   []sql.Row{{`(1.2)."e3"`}},
			},
			{
				Statement: `select '1..e'::jsonpath;`,
				Results:   []sql.Row{{`(1)."e"`}},
			},
			{
				Statement: `select '1..e3'::jsonpath;`,
				Results:   []sql.Row{{`(1)."e3"`}},
			},
			{
				Statement: `select '(1.).e'::jsonpath;`,
				Results:   []sql.Row{{`(1)."e"`}},
			},
			{
				Statement: `select '(1.).e3'::jsonpath;`,
				Results:   []sql.Row{{`(1)."e3"`}},
			},
			{
				Statement: `select '1?(2>3)'::jsonpath;`,
				Results:   []sql.Row{{`(1)?(2 > 3)`}},
			},
		},
	})
}
