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

func TestJsonbJsonpath(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_jsonb_jsonpath)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_jsonb_jsonpath,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `select jsonb '{"a": 12}' @? '$';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": 12}' @? '1';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": 12}' @? '$.a.b';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"a": 12}' @? '$.b';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"a": 12}' @? '$.a + 2';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": 12}' @? '$.b + 2';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '{"a": {"a": 12}}' @? '$.a.a';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"a": 12}}' @? '$.*.a';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"b": {"a": 12}}' @? '$.*.a';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"b": {"a": 12}}' @? '$.*.b';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"b": {"a": 12}}' @? 'strict $.*.b';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '{}' @? '$.*';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"a": 1}' @? '$.*';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? 'lax $.**{1}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? 'lax $.**{2}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? 'lax $.**{3}';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '[]' @? '$[*]';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '[1]' @? '$[*]';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1]' @? '$[1]';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '[1]' @? 'strict $[1]';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `select jsonb_path_query('[1]', 'strict $[1]');`,
				ErrorString: `jsonpath array subscript is out of bounds`,
			},
			{
				Statement: `select jsonb_path_query('[1]', 'strict $[1]', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb '[1]' @? 'lax $[10000000000000000]';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '[1]' @? 'strict $[10000000000000000]';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `select jsonb_path_query('[1]', 'lax $[10000000000000000]');`,
				ErrorString: `jsonpath array subscript is out of integer range`,
			},
			{
				Statement:   `select jsonb_path_query('[1]', 'strict $[10000000000000000]');`,
				ErrorString: `jsonpath array subscript is out of integer range`,
			},
			{
				Statement: `select jsonb '[1]' @? '$[0]';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1]' @? '$[0.3]';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1]' @? '$[0.5]';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1]' @? '$[0.9]';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1]' @? '$[1.2]';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '[1]' @? 'strict $[1.2]';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '{"a": [1,2,3], "b": [3,4,5]}' @? '$ ? (@.a[*] >  @.b[*])';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"a": [1,2,3], "b": [3,4,5]}' @? '$ ? (@.a[*] >= @.b[*])';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": [1,2,3], "b": [3,4,"5"]}' @? '$ ? (@.a[*] >= @.b[*])';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": [1,2,3], "b": [3,4,"5"]}' @? 'strict $ ? (@.a[*] >= @.b[*])';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"a": [1,2,3], "b": [3,4,null]}' @? '$ ? (@.a[*] >= @.b[*])';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '1' @? '$ ? ((@ == "1") is unknown)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '1' @? '$ ? ((@ == 1) is unknown)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '[{"a": 1}, {"a": 2}]' @? '$[0 to 1] ? (@.a > 1)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_exists('[{"a": 1}, {"a": 2}, 3]', 'lax $[*].a', silent => false);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_exists('[{"a": 1}, {"a": 2}, 3]', 'lax $[*].a', silent => true);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select jsonb_path_exists('[{"a": 1}, {"a": 2}, 3]', 'strict $[*].a', silent => false);`,
				ErrorString: `jsonpath member accessor can only be applied to an object`,
			},
			{
				Statement: `select jsonb_path_exists('[{"a": 1}, {"a": 2}, 3]', 'strict $[*].a', silent => true);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb_path_query('1', 'lax $.a');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select jsonb_path_query('1', 'strict $.a');`,
				ErrorString: `jsonpath member accessor can only be applied to an object`,
			},
			{
				Statement:   `select jsonb_path_query('1', 'strict $.*');`,
				ErrorString: `jsonpath wildcard member accessor can only be applied to an object`,
			},
			{
				Statement: `select jsonb_path_query('1', 'strict $.a', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('1', 'strict $.*', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[]', 'lax $.a');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select jsonb_path_query('[]', 'strict $.a');`,
				ErrorString: `jsonpath member accessor can only be applied to an object`,
			},
			{
				Statement: `select jsonb_path_query('[]', 'strict $.a', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{}', 'lax $.a');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select jsonb_path_query('{}', 'strict $.a');`,
				ErrorString: `JSON object does not contain key "a"`,
			},
			{
				Statement: `select jsonb_path_query('{}', 'strict $.a', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select jsonb_path_query('1', 'strict $[1]');`,
				ErrorString: `jsonpath array accessor can only be applied to an array`,
			},
			{
				Statement:   `select jsonb_path_query('1', 'strict $[*]');`,
				ErrorString: `jsonpath wildcard array accessor can only be applied to an array`,
			},
			{
				Statement:   `select jsonb_path_query('[]', 'strict $[1]');`,
				ErrorString: `jsonpath array subscript is out of bounds`,
			},
			{
				Statement:   `select jsonb_path_query('[]', 'strict $["a"]');`,
				ErrorString: `jsonpath array subscript is not a single numeric value`,
			},
			{
				Statement: `select jsonb_path_query('1', 'strict $[1]', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('1', 'strict $[*]', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[]', 'strict $[1]', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[]', 'strict $["a"]', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"a": 12, "b": {"a": 13}}', '$.a');`,
				Results:   []sql.Row{{12}},
			},
			{
				Statement: `select jsonb_path_query('{"a": 12, "b": {"a": 13}}', '$.b');`,
				Results:   []sql.Row{{`{"a": 13}`}},
			},
			{
				Statement: `select jsonb_path_query('{"a": 12, "b": {"a": 13}}', '$.*');`,
				Results:   []sql.Row{{12}, {`{"a": 13}`}},
			},
			{
				Statement: `select jsonb_path_query('{"a": 12, "b": {"a": 13}}', 'lax $.*.a');`,
				Results:   []sql.Row{{13}},
			},
			{
				Statement: `select jsonb_path_query('[12, {"a": 13}, {"b": 14}]', 'lax $[*].a');`,
				Results:   []sql.Row{{13}},
			},
			{
				Statement: `select jsonb_path_query('[12, {"a": 13}, {"b": 14}]', 'lax $[*].*');`,
				Results:   []sql.Row{{13}, {14}},
			},
			{
				Statement: `select jsonb_path_query('[12, {"a": 13}, {"b": 14}]', 'lax $[0].a');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[12, {"a": 13}, {"b": 14}]', 'lax $[1].a');`,
				Results:   []sql.Row{{13}},
			},
			{
				Statement: `select jsonb_path_query('[12, {"a": 13}, {"b": 14}]', 'lax $[2].a');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[12, {"a": 13}, {"b": 14}]', 'lax $[0,1].a');`,
				Results:   []sql.Row{{13}},
			},
			{
				Statement: `select jsonb_path_query('[12, {"a": 13}, {"b": 14}]', 'lax $[0 to 10].a');`,
				Results:   []sql.Row{{13}},
			},
			{
				Statement:   `select jsonb_path_query('[12, {"a": 13}, {"b": 14}]', 'lax $[0 to 10 / 0].a');`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `select jsonb_path_query('[12, {"a": 13}, {"b": 14}, "ccc", true]', '$[2.5 - 1 to $.size() - 2]');`,
				Results:   []sql.Row{{`{"a": 13}`}, {`{"b": 14}`}, {"ccc"}},
			},
			{
				Statement: `select jsonb_path_query('1', 'lax $[0]');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('1', 'lax $[*]');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('[1]', 'lax $[0]');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('[1]', 'lax $[*]');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('[1,2,3]', 'lax $[*]');`,
				Results:   []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement:   `select jsonb_path_query('[1,2,3]', 'strict $[*].a');`,
				ErrorString: `jsonpath member accessor can only be applied to an object`,
			},
			{
				Statement: `select jsonb_path_query('[1,2,3]', 'strict $[*].a', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[]', '$[last]');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[]', '$[last ? (exists(last))]');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select jsonb_path_query('[]', 'strict $[last]');`,
				ErrorString: `jsonpath array subscript is out of bounds`,
			},
			{
				Statement: `select jsonb_path_query('[]', 'strict $[last]', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[1]', '$[last]');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('[1,2,3]', '$[last]');`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select jsonb_path_query('[1,2,3]', '$[last - 1]');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select jsonb_path_query('[1,2,3]', '$[last ? (@.type() == "number")]');`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement:   `select jsonb_path_query('[1,2,3]', '$[last ? (@.type() == "string")]');`,
				ErrorString: `jsonpath array subscript is not a single numeric value`,
			},
			{
				Statement: `select jsonb_path_query('[1,2,3]', '$[last ? (@.type() == "string")]', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from jsonb_path_query('{"a": 10}', '$');`,
				Results:   []sql.Row{{`{"a": 10}`}},
			},
			{
				Statement:   `select * from jsonb_path_query('{"a": 10}', '$ ? (@.a < $value)');`,
				ErrorString: `could not find jsonpath variable "value"`,
			},
			{
				Statement:   `select * from jsonb_path_query('{"a": 10}', '$ ? (@.a < $value)', '1');`,
				ErrorString: `"vars" argument is not an object`,
			},
			{
				Statement:   `select * from jsonb_path_query('{"a": 10}', '$ ? (@.a < $value)', '[{"value" : 13}]');`,
				ErrorString: `"vars" argument is not an object`,
			},
			{
				Statement: `select * from jsonb_path_query('{"a": 10}', '$ ? (@.a < $value)', '{"value" : 13}');`,
				Results:   []sql.Row{{`{"a": 10}`}},
			},
			{
				Statement: `select * from jsonb_path_query('{"a": 10}', '$ ? (@.a < $value)', '{"value" : 8}');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from jsonb_path_query('{"a": 10}', '$.a ? (@ < $value)', '{"value" : 13}');`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `select * from jsonb_path_query('[10,11,12,13,14,15]', '$[*] ? (@ < $value)', '{"value" : 13}');`,
				Results:   []sql.Row{{10}, {11}, {12}},
			},
			{
				Statement: `select * from jsonb_path_query('[10,11,12,13,14,15]', '$[0,1] ? (@ < $x.value)', '{"x": {"value" : 13}}');`,
				Results:   []sql.Row{{10}, {11}},
			},
			{
				Statement: `select * from jsonb_path_query('[10,11,12,13,14,15]', '$[0 to 2] ? (@ < $value)', '{"value" : 15}');`,
				Results:   []sql.Row{{10}, {11}, {12}},
			},
			{
				Statement: `select * from jsonb_path_query('[1,"1",2,"2",null]', '$[*] ? (@ == "1")');`,
				Results:   []sql.Row{{"1"}},
			},
			{
				Statement: `select * from jsonb_path_query('[1,"1",2,"2",null]', '$[*] ? (@ == $value)', '{"value" : "1"}');`,
				Results:   []sql.Row{{"1"}},
			},
			{
				Statement: `select * from jsonb_path_query('[1,"1",2,"2",null]', '$[*] ? (@ == $value)', '{"value" : null}');`,
				Results:   []sql.Row{{`null`}},
			},
			{
				Statement: `select * from jsonb_path_query('[1, "2", null]', '$[*] ? (@ != null)');`,
				Results:   []sql.Row{{1}, {"2"}},
			},
			{
				Statement: `select * from jsonb_path_query('[1, "2", null]', '$[*] ? (@ == null)');`,
				Results:   []sql.Row{{`null`}},
			},
			{
				Statement: `select * from jsonb_path_query('{}', '$ ? (@ == @)');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from jsonb_path_query('[]', 'strict $ ? (@ == @)');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**');`,
				Results:   []sql.Row{{`{"a": {"b": 1}}`}, {`{"b": 1}`}, {1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{0}');`,
				Results:   []sql.Row{{`{"a": {"b": 1}}`}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{0 to last}');`,
				Results:   []sql.Row{{`{"a": {"b": 1}}`}, {`{"b": 1}`}, {1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{1}');`,
				Results:   []sql.Row{{`{"b": 1}`}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{1 to last}');`,
				Results:   []sql.Row{{`{"b": 1}`}, {1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{2}');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{2 to last}');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{3 to last}');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{last}');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{0}.b ? (@ > 0)');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{1}.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{0 to last}.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{1 to last}.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"b": 1}}', 'lax $.**{1 to 2}.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"c": {"b": 1}}}', 'lax $.**.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"c": {"b": 1}}}', 'lax $.**{0}.b ? (@ > 0)');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"c": {"b": 1}}}', 'lax $.**{1}.b ? (@ > 0)');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"c": {"b": 1}}}', 'lax $.**{0 to last}.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"c": {"b": 1}}}', 'lax $.**{1 to last}.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"c": {"b": 1}}}', 'lax $.**{1 to 2}.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb_path_query('{"a": {"c": {"b": 1}}}', 'lax $.**{2 to 3}.b ? (@ > 0)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? '$.**.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? '$.**{0}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? '$.**{1}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? '$.**{0 to last}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? '$.**{1 to last}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"b": 1}}' @? '$.**{1 to 2}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"c": {"b": 1}}}' @? '$.**.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"c": {"b": 1}}}' @? '$.**{0}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"a": {"c": {"b": 1}}}' @? '$.**{1}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"a": {"c": {"b": 1}}}' @? '$.**{0 to last}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"c": {"b": 1}}}' @? '$.**{1 to last}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"c": {"b": 1}}}' @? '$.**{1 to 2}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": {"c": {"b": 1}}}' @? '$.**{2 to 3}.b ? ( @ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_query('{"g": {"x": 2}}', '$.g ? (exists (@.x))');`,
				Results:   []sql.Row{{`{"x": 2}`}},
			},
			{
				Statement: `select jsonb_path_query('{"g": {"x": 2}}', '$.g ? (exists (@.y))');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"g": {"x": 2}}', '$.g ? (exists (@.x ? (@ >= 2) ))');`,
				Results:   []sql.Row{{`{"x": 2}`}},
			},
			{
				Statement: `select jsonb_path_query('{"g": [{"x": 2}, {"y": 3}]}', 'lax $.g ? (exists (@.x))');`,
				Results:   []sql.Row{{`{"x": 2}`}},
			},
			{
				Statement: `select jsonb_path_query('{"g": [{"x": 2}, {"y": 3}]}', 'lax $.g ? (exists (@.x + "3"))');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"g": [{"x": 2}, {"y": 3}]}', 'lax $.g ? ((exists (@.x + "3")) is unknown)');`,
				Results:   []sql.Row{{`{"x": 2}`}, {`{"y": 3}`}},
			},
			{
				Statement: `select jsonb_path_query('{"g": [{"x": 2}, {"y": 3}]}', 'strict $.g[*] ? (exists (@.x))');`,
				Results:   []sql.Row{{`{"x": 2}`}},
			},
			{
				Statement: `select jsonb_path_query('{"g": [{"x": 2}, {"y": 3}]}', 'strict $.g[*] ? ((exists (@.x)) is unknown)');`,
				Results:   []sql.Row{{`{"y": 3}`}},
			},
			{
				Statement: `select jsonb_path_query('{"g": [{"x": 2}, {"y": 3}]}', 'strict $.g ? (exists (@[*].x))');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"g": [{"x": 2}, {"y": 3}]}', 'strict $.g ? ((exists (@[*].x)) is unknown)');`,
				Results:   []sql.Row{{`[{"x": 2}, {"y": 3}]`}},
			},
			{
				Statement: `select
	x, y,
	jsonb_path_query(
		'[true, false, null]',
		'$[*] ? (@ == true  &&  ($x == true && $y == true) ||
				 @ == false && !($x == true && $y == true) ||
				 @ == null  &&  ($x == true && $y == true) is unknown)',
		jsonb_build_object('x', x, 'y', y)
	) as "x && y"
from
	(values (jsonb 'true'), ('false'), ('"null"')) x(x),
	(values (jsonb 'true'), ('false'), ('"null"')) y(y);`,
				Results: []sql.Row{{`true`, `true`, `true`}, {`true`, `false`, `false`}, {`true`, "null", `null`}, {`false`, `true`, `false`}, {`false`, `false`, `false`}, {`false`, "null", `false`}, {"null", `true`, `null`}, {"null", `false`, `false`}, {"null", "null", `null`}},
			},
			{
				Statement: `select
	x, y,
	jsonb_path_query(
		'[true, false, null]',
		'$[*] ? (@ == true  &&  ($x == true || $y == true) ||
				 @ == false && !($x == true || $y == true) ||
				 @ == null  &&  ($x == true || $y == true) is unknown)',
		jsonb_build_object('x', x, 'y', y)
	) as "x || y"
from
	(values (jsonb 'true'), ('false'), ('"null"')) x(x),
	(values (jsonb 'true'), ('false'), ('"null"')) y(y);`,
				Results: []sql.Row{{`true`, `true`, `true`}, {`true`, `false`, `true`}, {`true`, "null", `true`}, {`false`, `true`, `true`}, {`false`, `false`, `false`}, {`false`, "null", `null`}, {"null", `true`, `true`}, {"null", `false`, `null`}, {"null", "null", `null`}},
			},
			{
				Statement: `select jsonb '{"a": 1, "b":1}' @? '$ ? (@.a == @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": 1, "b":1}}' @? '$ ? (@.a == @.b)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '{"c": {"a": 1, "b":1}}' @? '$.c ? (@.a == @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": 1, "b":1}}' @? '$.c ? ($.c.a == @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": 1, "b":1}}' @? '$.* ? (@.a == @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": 1, "b":1}' @? '$.** ? (@.a == @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": 1, "b":1}}' @? '$.** ? (@.a == @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_query('{"c": {"a": 2, "b":1}}', '$.** ? (@.a == 1 + 1)');`,
				Results:   []sql.Row{{`{"a": 2, "b": 1}`}},
			},
			{
				Statement: `select jsonb_path_query('{"c": {"a": 2, "b":1}}', '$.** ? (@.a == (1 + 1))');`,
				Results:   []sql.Row{{`{"a": 2, "b": 1}`}},
			},
			{
				Statement: `select jsonb_path_query('{"c": {"a": 2, "b":1}}', '$.** ? (@.a == @.b + 1)');`,
				Results:   []sql.Row{{`{"a": 2, "b": 1}`}},
			},
			{
				Statement: `select jsonb_path_query('{"c": {"a": 2, "b":1}}', '$.** ? (@.a == (@.b + 1))');`,
				Results:   []sql.Row{{`{"a": 2, "b": 1}`}},
			},
			{
				Statement: `select jsonb '{"c": {"a": -1, "b":1}}' @? '$.** ? (@.a == - 1)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": -1, "b":1}}' @? '$.** ? (@.a == -1)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": -1, "b":1}}' @? '$.** ? (@.a == -@.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": -1, "b":1}}' @? '$.** ? (@.a == - @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": 0, "b":1}}' @? '$.** ? (@.a == 1 - @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": 2, "b":1}}' @? '$.** ? (@.a == 1 - - @.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"c": {"a": 0, "b":1}}' @? '$.** ? (@.a == 1 - +@.b)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1,2,3]' @? '$ ? (+@[*] > +2)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1,2,3]' @? '$ ? (+@[*] > +3)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '[1,2,3]' @? '$ ? (-@[*] < -2)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1,2,3]' @? '$ ? (-@[*] < -3)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '1' @? '$ ? ($ > 0)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_query('[1,2,0,3]', '$[*] ? (2 / @ > 0)');`,
				Results:   []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement: `select jsonb_path_query('[1,2,0,3]', '$[*] ? ((2 / @ > 0) is unknown)');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `select jsonb_path_query('0', '1 / $');`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select jsonb_path_query('0', '1 / $ + 2');`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select jsonb_path_query('0', '-(3 + 1 % $)');`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select jsonb_path_query('1', '$ + "2"');`,
				ErrorString: `right operand of jsonpath operator + is not a single numeric value`,
			},
			{
				Statement:   `select jsonb_path_query('[1, 2]', '3 * $');`,
				ErrorString: `right operand of jsonpath operator * is not a single numeric value`,
			},
			{
				Statement:   `select jsonb_path_query('"a"', '-$');`,
				ErrorString: `operand of unary jsonpath operator - is not a numeric value`,
			},
			{
				Statement:   `select jsonb_path_query('[1,"2",3]', '+$');`,
				ErrorString: `operand of unary jsonpath operator + is not a numeric value`,
			},
			{
				Statement: `select jsonb_path_query('1', '$ + "2"', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[1, 2]', '3 * $', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('"a"', '-$', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[1,"2",3]', '+$', silent => true);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select jsonb '["1",2,0,3]' @? '-$[*]';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '[1,"2",0,3]' @? '-$[*]';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '["1",2,0,3]' @? 'strict -$[*]';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '[1,"2",0,3]' @? 'strict -$[*]';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb_path_query('{"a": [2]}', 'lax $.a * 3');`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `select jsonb_path_query('{"a": [2]}', 'lax $.a + 3');`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `select jsonb_path_query('{"a": [2, 3, 4]}', 'lax -$.a');`,
				Results:   []sql.Row{{-2}, {-3}, {-4}},
			},
			{
				Statement:   `select jsonb_path_query('{"a": [1, 2]}', 'lax $.a * 3');`,
				ErrorString: `left operand of jsonpath operator * is not a single numeric value`,
			},
			{
				Statement: `select jsonb_path_query('{"a": [1, 2]}', 'lax $.a * 3', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('2', '$ > 1');`,
				Results:   []sql.Row{{`true`}},
			},
			{
				Statement: `select jsonb_path_query('2', '$ <= 1');`,
				Results:   []sql.Row{{`false`}},
			},
			{
				Statement: `select jsonb_path_query('2', '$ == "2"');`,
				Results:   []sql.Row{{`null`}},
			},
			{
				Statement: `select jsonb '2' @? '$ == "2"';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '2' @@ '$ > 1';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '2' @@ '$ <= 1';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb '2' @@ '$ == "2"';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '2' @@ '1';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '{}' @@ '$';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '[]' @@ '$';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '[1,2,3]' @@ '$[*]';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb '[]' @@ '$[*]';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb_path_match('[[1, true], [2, false]]', 'strict $[*] ? (@[0] > $x) [1]', '{"x": 1}');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select jsonb_path_match('[[1, true], [2, false]]', 'strict $[*] ? (@[0] < $x) [1]', '{"x": 2}');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_match('[{"a": 1}, {"a": 2}, 3]', 'lax exists($[*].a)', silent => false);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_match('[{"a": 1}, {"a": 2}, 3]', 'lax exists($[*].a)', silent => true);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_match('[{"a": 1}, {"a": 2}, 3]', 'strict exists($[*].a)', silent => false);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb_path_match('[{"a": 1}, {"a": 2}, 3]', 'strict exists($[*].a)', silent => true);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select jsonb_path_query('[null,1,true,"a",[],{}]', '$.type()');`,
				Results:   []sql.Row{{"array"}},
			},
			{
				Statement: `select jsonb_path_query('[null,1,true,"a",[],{}]', 'lax $.type()');`,
				Results:   []sql.Row{{"array"}},
			},
			{
				Statement: `select jsonb_path_query('[null,1,true,"a",[],{}]', '$[*].type()');`,
				Results:   []sql.Row{{"null"}, {"number"}, {"boolean"}, {"string"}, {"array"}, {"object"}},
			},
			{
				Statement: `select jsonb_path_query('null', 'null.type()');`,
				Results:   []sql.Row{{"null"}},
			},
			{
				Statement: `select jsonb_path_query('null', 'true.type()');`,
				Results:   []sql.Row{{"boolean"}},
			},
			{
				Statement: `select jsonb_path_query('null', '(123).type()');`,
				Results:   []sql.Row{{"number"}},
			},
			{
				Statement: `select jsonb_path_query('null', '"123".type()');`,
				Results:   []sql.Row{{"string"}},
			},
			{
				Statement: `select jsonb_path_query('{"a": 2}', '($.a - 5).abs() + 10');`,
				Results:   []sql.Row{{13}},
			},
			{
				Statement: `select jsonb_path_query('{"a": 2.5}', '-($.a * $.a).floor() % 4.3');`,
				Results:   []sql.Row{{-1.7}},
			},
			{
				Statement: `select jsonb_path_query('[1, 2, 3]', '($[*] > 2) ? (@ == true)');`,
				Results:   []sql.Row{{`true`}},
			},
			{
				Statement: `select jsonb_path_query('[1, 2, 3]', '($[*] > 3).type()');`,
				Results:   []sql.Row{{"boolean"}},
			},
			{
				Statement: `select jsonb_path_query('[1, 2, 3]', '($[*].a > 3).type()');`,
				Results:   []sql.Row{{"boolean"}},
			},
			{
				Statement: `select jsonb_path_query('[1, 2, 3]', 'strict ($[*].a > 3).type()');`,
				Results:   []sql.Row{{"null"}},
			},
			{
				Statement:   `select jsonb_path_query('[1,null,true,"11",[],[1],[1,2,3],{},{"a":1,"b":2}]', 'strict $[*].size()');`,
				ErrorString: `jsonpath item method .size() can only be applied to an array`,
			},
			{
				Statement: `select jsonb_path_query('[1,null,true,"11",[],[1],[1,2,3],{},{"a":1,"b":2}]', 'strict $[*].size()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[1,null,true,"11",[],[1],[1,2,3],{},{"a":1,"b":2}]', 'lax $[*].size()');`,
				Results:   []sql.Row{{1}, {1}, {1}, {1}, {0}, {1}, {3}, {1}, {1}},
			},
			{
				Statement: `select jsonb_path_query('[0, 1, -2, -3.4, 5.6]', '$[*].abs()');`,
				Results:   []sql.Row{{0}, {1}, {2}, {3.4}, {5.6}},
			},
			{
				Statement: `select jsonb_path_query('[0, 1, -2, -3.4, 5.6]', '$[*].floor()');`,
				Results:   []sql.Row{{0}, {1}, {-2}, {-4}, {5}},
			},
			{
				Statement: `select jsonb_path_query('[0, 1, -2, -3.4, 5.6]', '$[*].ceiling()');`,
				Results:   []sql.Row{{0}, {1}, {-2}, {-3}, {6}},
			},
			{
				Statement: `select jsonb_path_query('[0, 1, -2, -3.4, 5.6]', '$[*].ceiling().abs()');`,
				Results:   []sql.Row{{0}, {1}, {2}, {3}, {6}},
			},
			{
				Statement: `select jsonb_path_query('[0, 1, -2, -3.4, 5.6]', '$[*].ceiling().abs().type()');`,
				Results:   []sql.Row{{"number"}, {"number"}, {"number"}, {"number"}, {"number"}},
			},
			{
				Statement:   `select jsonb_path_query('[{},1]', '$[*].keyvalue()');`,
				ErrorString: `jsonpath item method .keyvalue() can only be applied to an object`,
			},
			{
				Statement: `select jsonb_path_query('[{},1]', '$[*].keyvalue()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{}', '$.keyvalue()');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{"a": 1, "b": [1, 2], "c": {"a": "bbb"}}', '$.keyvalue()');`,
				Results:   []sql.Row{{`{"id": 0, "key": "a", "value": 1}`}, {`{"id": 0, "key": "b", "value": [1, 2]}`}, {`{"id": 0, "key": "c", "value": {"a": "bbb"}}`}},
			},
			{
				Statement: `select jsonb_path_query('[{"a": 1, "b": [1, 2]}, {"c": {"a": "bbb"}}]', '$[*].keyvalue()');`,
				Results:   []sql.Row{{`{"id": 12, "key": "a", "value": 1}`}, {`{"id": 12, "key": "b", "value": [1, 2]}`}, {`{"id": 72, "key": "c", "value": {"a": "bbb"}}`}},
			},
			{
				Statement:   `select jsonb_path_query('[{"a": 1, "b": [1, 2]}, {"c": {"a": "bbb"}}]', 'strict $.keyvalue()');`,
				ErrorString: `jsonpath item method .keyvalue() can only be applied to an object`,
			},
			{
				Statement: `select jsonb_path_query('[{"a": 1, "b": [1, 2]}, {"c": {"a": "bbb"}}]', 'lax $.keyvalue()');`,
				Results:   []sql.Row{{`{"id": 12, "key": "a", "value": 1}`}, {`{"id": 12, "key": "b", "value": [1, 2]}`}, {`{"id": 72, "key": "c", "value": {"a": "bbb"}}`}},
			},
			{
				Statement:   `select jsonb_path_query('[{"a": 1, "b": [1, 2]}, {"c": {"a": "bbb"}}]', 'strict $.keyvalue().a');`,
				ErrorString: `jsonpath item method .keyvalue() can only be applied to an object`,
			},
			{
				Statement: `select jsonb '{"a": 1, "b": [1, 2]}' @? 'lax $.keyvalue()';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb '{"a": 1, "b": [1, 2]}' @? 'lax $.keyvalue().key';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select jsonb_path_query('null', '$.double()');`,
				ErrorString: `jsonpath item method .double() can only be applied to a string or numeric value`,
			},
			{
				Statement:   `select jsonb_path_query('true', '$.double()');`,
				ErrorString: `jsonpath item method .double() can only be applied to a string or numeric value`,
			},
			{
				Statement: `select jsonb_path_query('null', '$.double()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('true', '$.double()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[]', '$.double()');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select jsonb_path_query('[]', 'strict $.double()');`,
				ErrorString: `jsonpath item method .double() can only be applied to a string or numeric value`,
			},
			{
				Statement:   `select jsonb_path_query('{}', '$.double()');`,
				ErrorString: `jsonpath item method .double() can only be applied to a string or numeric value`,
			},
			{
				Statement: `select jsonb_path_query('[]', 'strict $.double()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('{}', '$.double()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('1.23', '$.double()');`,
				Results:   []sql.Row{{1.23}},
			},
			{
				Statement: `select jsonb_path_query('"1.23"', '$.double()');`,
				Results:   []sql.Row{{1.23}},
			},
			{
				Statement:   `select jsonb_path_query('"1.23aaa"', '$.double()');`,
				ErrorString: `string argument of jsonpath item method .double() is not a valid representation of a double precision number`,
			},
			{
				Statement:   `select jsonb_path_query('1e1000', '$.double()');`,
				ErrorString: `numeric argument of jsonpath item method .double() is out of range for type double precision`,
			},
			{
				Statement:   `select jsonb_path_query('"nan"', '$.double()');`,
				ErrorString: `string argument of jsonpath item method .double() is not a valid representation of a double precision number`,
			},
			{
				Statement:   `select jsonb_path_query('"NaN"', '$.double()');`,
				ErrorString: `string argument of jsonpath item method .double() is not a valid representation of a double precision number`,
			},
			{
				Statement:   `select jsonb_path_query('"inf"', '$.double()');`,
				ErrorString: `string argument of jsonpath item method .double() is not a valid representation of a double precision number`,
			},
			{
				Statement:   `select jsonb_path_query('"-inf"', '$.double()');`,
				ErrorString: `string argument of jsonpath item method .double() is not a valid representation of a double precision number`,
			},
			{
				Statement: `select jsonb_path_query('"inf"', '$.double()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('"-inf"', '$.double()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select jsonb_path_query('{}', '$.abs()');`,
				ErrorString: `jsonpath item method .abs() can only be applied to a numeric value`,
			},
			{
				Statement:   `select jsonb_path_query('true', '$.floor()');`,
				ErrorString: `jsonpath item method .floor() can only be applied to a numeric value`,
			},
			{
				Statement:   `select jsonb_path_query('"1.2"', '$.ceiling()');`,
				ErrorString: `jsonpath item method .ceiling() can only be applied to a numeric value`,
			},
			{
				Statement: `select jsonb_path_query('{}', '$.abs()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('true', '$.floor()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('"1.2"', '$.ceiling()', silent => true);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('["", "a", "abc", "abcabc"]', '$[*] ? (@ starts with "abc")');`,
				Results:   []sql.Row{{"abc"}, {"abcabc"}},
			},
			{
				Statement: `select jsonb_path_query('["", "a", "abc", "abcabc"]', 'strict $ ? (@[*] starts with "abc")');`,
				Results:   []sql.Row{{`["", "a", "abc", "abcabc"]`}},
			},
			{
				Statement: `select jsonb_path_query('["", "a", "abd", "abdabc"]', 'strict $ ? (@[*] starts with "abc")');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('["abc", "abcabc", null, 1]', 'strict $ ? (@[*] starts with "abc")');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('["abc", "abcabc", null, 1]', 'strict $ ? ((@[*] starts with "abc") is unknown)');`,
				Results:   []sql.Row{{`["abc", "abcabc", null, 1]`}},
			},
			{
				Statement: `select jsonb_path_query('[[null, 1, "abc", "abcabc"]]', 'lax $ ? (@[*] starts with "abc")');`,
				Results:   []sql.Row{{`[null, 1, "abc", "abcabc"]`}},
			},
			{
				Statement: `select jsonb_path_query('[[null, 1, "abd", "abdabc"]]', 'lax $ ? ((@[*] starts with "abc") is unknown)');`,
				Results:   []sql.Row{{`[null, 1, "abd", "abdabc"]`}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "abd", "abdabc"]', 'lax $[*] ? ((@ starts with "abc") is unknown)');`,
				Results:   []sql.Row{{`null`}, {1}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "abc", "abd", "aBdC", "abdacb", "babc", "adc\nabc", "ab\nadc"]', 'lax $[*] ? (@ like_regex "^ab.*c")');`,
				Results:   []sql.Row{{"abc"}, {"abdacb"}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "abc", "abd", "aBdC", "abdacb", "babc", "adc\nabc", "ab\nadc"]', 'lax $[*] ? (@ like_regex "^ab.*c" flag "i")');`,
				Results:   []sql.Row{{"abc"}, {"aBdC"}, {"abdacb"}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "abc", "abd", "aBdC", "abdacb", "babc", "adc\nabc", "ab\nadc"]', 'lax $[*] ? (@ like_regex "^ab.*c" flag "m")');`,
				Results:   []sql.Row{{"abc"}, {"abdacb"}, {"adc\nabc"}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "abc", "abd", "aBdC", "abdacb", "babc", "adc\nabc", "ab\nadc"]', 'lax $[*] ? (@ like_regex "^ab.*c" flag "s")');`,
				Results:   []sql.Row{{"abc"}, {"abdacb"}, {"ab\nadc"}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "a\b", "a\\b", "^a\\b$"]', 'lax $[*] ? (@ like_regex "a\\b" flag "q")');`,
				Results:   []sql.Row{{"a\\b"}, {"^a\\b$"}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "a\b", "a\\b", "^a\\b$"]', 'lax $[*] ? (@ like_regex "a\\b" flag "")');`,
				Results:   []sql.Row{{"a\b"}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "a\b", "a\\b", "^a\\b$"]', 'lax $[*] ? (@ like_regex "^a\\b$" flag "q")');`,
				Results:   []sql.Row{{"^a\\b$"}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "a\b", "a\\b", "^a\\b$"]', 'lax $[*] ? (@ like_regex "^a\\B$" flag "q")');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "a\b", "a\\b", "^a\\b$"]', 'lax $[*] ? (@ like_regex "^a\\B$" flag "iq")');`,
				Results:   []sql.Row{{"^a\\b$"}},
			},
			{
				Statement: `select jsonb_path_query('[null, 1, "a\b", "a\\b", "^a\\b$"]', 'lax $[*] ? (@ like_regex "^a\\b$" flag "")');`,
				Results:   []sql.Row{{"a\b"}},
			},
			{
				Statement:   `select jsonb_path_query('null', '$.datetime()');`,
				ErrorString: `jsonpath item method .datetime() can only be applied to a string`,
			},
			{
				Statement:   `select jsonb_path_query('true', '$.datetime()');`,
				ErrorString: `jsonpath item method .datetime() can only be applied to a string`,
			},
			{
				Statement:   `select jsonb_path_query('1', '$.datetime()');`,
				ErrorString: `jsonpath item method .datetime() can only be applied to a string`,
			},
			{
				Statement: `select jsonb_path_query('[]', '$.datetime()');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select jsonb_path_query('[]', 'strict $.datetime()');`,
				ErrorString: `jsonpath item method .datetime() can only be applied to a string`,
			},
			{
				Statement:   `select jsonb_path_query('{}', '$.datetime()');`,
				ErrorString: `jsonpath item method .datetime() can only be applied to a string`,
			},
			{
				Statement:   `select jsonb_path_query('"bogus"', '$.datetime()');`,
				ErrorString: `datetime format is not recognized: "bogus"`,
			},
			{
				Statement:   `select jsonb_path_query('"12:34"', '$.datetime("aaa")');`,
				ErrorString: `invalid datetime format separator: "a"`,
			},
			{
				Statement:   `select jsonb_path_query('"aaaa"', '$.datetime("HH24")');`,
				ErrorString: `invalid value "aa" for "HH24"`,
			},
			{
				Statement: `select jsonb '"10-03-2017"' @? '$.datetime("dd-mm-yyyy")';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017"', '$.datetime("dd-mm-yyyy")');`,
				Results:   []sql.Row{{"2017-03-10"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017"', '$.datetime("dd-mm-yyyy").type()');`,
				Results:   []sql.Row{{"date"}},
			},
			{
				Statement:   `select jsonb_path_query('"10-03-2017 12:34"', '$.datetime("dd-mm-yyyy")');`,
				ErrorString: `trailing characters remain in input string after datetime format`,
			},
			{
				Statement:   `select jsonb_path_query('"10-03-2017 12:34"', '$.datetime("dd-mm-yyyy").type()');`,
				ErrorString: `trailing characters remain in input string after datetime format`,
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34"', '       $.datetime("dd-mm-yyyy HH24:MI").type()');`,
				Results:   []sql.Row{{"timestamp without time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 +05:20"', '$.datetime("dd-mm-yyyy HH24:MI TZH:TZM").type()');`,
				Results:   []sql.Row{{"timestamp with time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34:56"', '$.datetime("HH24:MI:SS").type()');`,
				Results:   []sql.Row{{"time without time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34:56 +05:20"', '$.datetime("HH24:MI:SS TZH:TZM").type()');`,
				Results:   []sql.Row{{"time with time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017T12:34:56"', '$.datetime("dd-mm-yyyy\"T\"HH24:MI:SS")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:56"}},
			},
			{
				Statement:   `select jsonb_path_query('"10-03-2017t12:34:56"', '$.datetime("dd-mm-yyyy\"T\"HH24:MI:SS")');`,
				ErrorString: `unmatched format character "T"`,
			},
			{
				Statement:   `select jsonb_path_query('"10-03-2017 12:34:56"', '$.datetime("dd-mm-yyyy\"T\"HH24:MI:SS")');`,
				ErrorString: `unmatched format character "T"`,
			},
			{
				Statement: `set time zone '+00';`,
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34"', '$.datetime("dd-mm-yyyy HH24:MI")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00"}},
			},
			{
				Statement:   `select jsonb_path_query('"10-03-2017 12:34"', '$.datetime("dd-mm-yyyy HH24:MI TZH")');`,
				ErrorString: `input string is too short for datetime format`,
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 +05"', '$.datetime("dd-mm-yyyy HH24:MI TZH")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00+05:00"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 -05"', '$.datetime("dd-mm-yyyy HH24:MI TZH")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00-05:00"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 +05:20"', '$.datetime("dd-mm-yyyy HH24:MI TZH:TZM")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00+05:20"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 -05:20"', '$.datetime("dd-mm-yyyy HH24:MI TZH:TZM")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00-05:20"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34"', '$.datetime("HH24:MI")');`,
				Results:   []sql.Row{{"12:34:00"}},
			},
			{
				Statement:   `select jsonb_path_query('"12:34"', '$.datetime("HH24:MI TZH")');`,
				ErrorString: `input string is too short for datetime format`,
			},
			{
				Statement: `select jsonb_path_query('"12:34 +05"', '$.datetime("HH24:MI TZH")');`,
				Results:   []sql.Row{{"12:34:00+05:00"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34 -05"', '$.datetime("HH24:MI TZH")');`,
				Results:   []sql.Row{{"12:34:00-05:00"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34 +05:20"', '$.datetime("HH24:MI TZH:TZM")');`,
				Results:   []sql.Row{{"12:34:00+05:20"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34 -05:20"', '$.datetime("HH24:MI TZH:TZM")');`,
				Results:   []sql.Row{{"12:34:00-05:20"}},
			},
			{
				Statement: `set time zone '+10';`,
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34"', '$.datetime("dd-mm-yyyy HH24:MI")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00"}},
			},
			{
				Statement:   `select jsonb_path_query('"10-03-2017 12:34"', '$.datetime("dd-mm-yyyy HH24:MI TZH")');`,
				ErrorString: `input string is too short for datetime format`,
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 +05"', '$.datetime("dd-mm-yyyy HH24:MI TZH")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00+05:00"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 -05"', '$.datetime("dd-mm-yyyy HH24:MI TZH")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00-05:00"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 +05:20"', '$.datetime("dd-mm-yyyy HH24:MI TZH:TZM")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00+05:20"}},
			},
			{
				Statement: `select jsonb_path_query('"10-03-2017 12:34 -05:20"', '$.datetime("dd-mm-yyyy HH24:MI TZH:TZM")');`,
				Results:   []sql.Row{{"2017-03-10T12:34:00-05:20"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34"', '$.datetime("HH24:MI")');`,
				Results:   []sql.Row{{"12:34:00"}},
			},
			{
				Statement:   `select jsonb_path_query('"12:34"', '$.datetime("HH24:MI TZH")');`,
				ErrorString: `input string is too short for datetime format`,
			},
			{
				Statement: `select jsonb_path_query('"12:34 +05"', '$.datetime("HH24:MI TZH")');`,
				Results:   []sql.Row{{"12:34:00+05:00"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34 -05"', '$.datetime("HH24:MI TZH")');`,
				Results:   []sql.Row{{"12:34:00-05:00"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34 +05:20"', '$.datetime("HH24:MI TZH:TZM")');`,
				Results:   []sql.Row{{"12:34:00+05:20"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34 -05:20"', '$.datetime("HH24:MI TZH:TZM")');`,
				Results:   []sql.Row{{"12:34:00-05:20"}},
			},
			{
				Statement: `set time zone default;`,
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10"', '$.datetime().type()');`,
				Results:   []sql.Row{{"date"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10"', '$.datetime()');`,
				Results:   []sql.Row{{"2017-03-10"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10 12:34:56"', '$.datetime().type()');`,
				Results:   []sql.Row{{"timestamp without time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10 12:34:56"', '$.datetime()');`,
				Results:   []sql.Row{{"2017-03-10T12:34:56"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10 12:34:56+3"', '$.datetime().type()');`,
				Results:   []sql.Row{{"timestamp with time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10 12:34:56+3"', '$.datetime()');`,
				Results:   []sql.Row{{"2017-03-10T12:34:56+03:00"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10 12:34:56+3:10"', '$.datetime().type()');`,
				Results:   []sql.Row{{"timestamp with time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10 12:34:56+3:10"', '$.datetime()');`,
				Results:   []sql.Row{{"2017-03-10T12:34:56+03:10"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10T12:34:56+3:10"', '$.datetime()');`,
				Results:   []sql.Row{{"2017-03-10T12:34:56+03:10"}},
			},
			{
				Statement:   `select jsonb_path_query('"2017-03-10t12:34:56+3:10"', '$.datetime()');`,
				ErrorString: `datetime format is not recognized: "2017-03-10t12:34:56+3:10"`,
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10 12:34:56.789+3:10"', '$.datetime()');`,
				Results:   []sql.Row{{"2017-03-10T12:34:56.789+03:10"}},
			},
			{
				Statement: `select jsonb_path_query('"2017-03-10T12:34:56.789+3:10"', '$.datetime()');`,
				Results:   []sql.Row{{"2017-03-10T12:34:56.789+03:10"}},
			},
			{
				Statement:   `select jsonb_path_query('"2017-03-10t12:34:56.789+3:10"', '$.datetime()');`,
				ErrorString: `datetime format is not recognized: "2017-03-10t12:34:56.789+3:10"`,
			},
			{
				Statement: `select jsonb_path_query('"12:34:56"', '$.datetime().type()');`,
				Results:   []sql.Row{{"time without time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34:56"', '$.datetime()');`,
				Results:   []sql.Row{{"12:34:56"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34:56+3"', '$.datetime().type()');`,
				Results:   []sql.Row{{"time with time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34:56+3"', '$.datetime()');`,
				Results:   []sql.Row{{"12:34:56+03:00"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34:56+3:10"', '$.datetime().type()');`,
				Results:   []sql.Row{{"time with time zone"}},
			},
			{
				Statement: `select jsonb_path_query('"12:34:56+3:10"', '$.datetime()');`,
				Results:   []sql.Row{{"12:34:56+03:10"}},
			},
			{
				Statement: `set time zone '+00';`,
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10", "2017-03-11", "2017-03-09", "12:34:56", "01:02:03+04", "2017-03-10 00:00:00", "2017-03-10 12:34:56", "2017-03-10 01:02:03+04", "2017-03-10 03:00:00+03"]',
	'$[*].datetime() ? (@ == "10.03.2017".datetime("dd.mm.yyyy"))');`,
				ErrorString: `cannot convert value from date to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10", "2017-03-11", "2017-03-09", "12:34:56", "01:02:03+04", "2017-03-10 00:00:00", "2017-03-10 12:34:56", "2017-03-10 01:02:03+04", "2017-03-10 03:00:00+03"]',
	'$[*].datetime() ? (@ >= "10.03.2017".datetime("dd.mm.yyyy"))');`,
				ErrorString: `cannot convert value from date to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10", "2017-03-11", "2017-03-09", "12:34:56", "01:02:03+04", "2017-03-10 00:00:00", "2017-03-10 12:34:56", "2017-03-10 01:02:03+04", "2017-03-10 03:00:00+03"]',
	'$[*].datetime() ? (@ <  "10.03.2017".datetime("dd.mm.yyyy"))');`,
				ErrorString: `cannot convert value from date to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10", "2017-03-11", "2017-03-09", "12:34:56", "01:02:03+04", "2017-03-10 00:00:00", "2017-03-10 12:34:56", "2017-03-10 01:02:03+04", "2017-03-10 03:00:00+03"]',
	'$[*].datetime() ? (@ == "10.03.2017".datetime("dd.mm.yyyy"))');`,
				Results: []sql.Row{{"2017-03-10"}, {"2017-03-10T00:00:00"}, {"2017-03-10T03:00:00+03:00"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10", "2017-03-11", "2017-03-09", "12:34:56", "01:02:03+04", "2017-03-10 00:00:00", "2017-03-10 12:34:56", "2017-03-10 01:02:03+04", "2017-03-10 03:00:00+03"]',
	'$[*].datetime() ? (@ >= "10.03.2017".datetime("dd.mm.yyyy"))');`,
				Results: []sql.Row{{"2017-03-10"}, {"2017-03-11"}, {"2017-03-10T00:00:00"}, {"2017-03-10T12:34:56"}, {"2017-03-10T03:00:00+03:00"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10", "2017-03-11", "2017-03-09", "12:34:56", "01:02:03+04", "2017-03-10 00:00:00", "2017-03-10 12:34:56", "2017-03-10 01:02:03+04", "2017-03-10 03:00:00+03"]',
	'$[*].datetime() ? (@ <  "10.03.2017".datetime("dd.mm.yyyy"))');`,
				Results: []sql.Row{{"2017-03-09"}, {"2017-03-10T01:02:03+04:00"}},
			},
			{
				Statement: `select jsonb_path_query(
	'["12:34:00", "12:35:00", "12:36:00", "12:35:00+00", "12:35:00+01", "13:35:00+01", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00+01"]',
	'$[*].datetime() ? (@ == "12:35".datetime("HH24:MI"))');`,
				ErrorString: `cannot convert value from time to timetz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["12:34:00", "12:35:00", "12:36:00", "12:35:00+00", "12:35:00+01", "13:35:00+01", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00+01"]',
	'$[*].datetime() ? (@ >= "12:35".datetime("HH24:MI"))');`,
				ErrorString: `cannot convert value from time to timetz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["12:34:00", "12:35:00", "12:36:00", "12:35:00+00", "12:35:00+01", "13:35:00+01", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00+01"]',
	'$[*].datetime() ? (@ <  "12:35".datetime("HH24:MI"))');`,
				ErrorString: `cannot convert value from time to timetz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["12:34:00", "12:35:00", "12:36:00", "12:35:00+00", "12:35:00+01", "13:35:00+01", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00+01"]',
	'$[*].datetime() ? (@ == "12:35".datetime("HH24:MI"))');`,
				Results: []sql.Row{{"12:35:00"}, {"12:35:00+00:00"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["12:34:00", "12:35:00", "12:36:00", "12:35:00+00", "12:35:00+01", "13:35:00+01", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00+01"]',
	'$[*].datetime() ? (@ >= "12:35".datetime("HH24:MI"))');`,
				Results: []sql.Row{{"12:35:00"}, {"12:36:00"}, {"12:35:00+00:00"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["12:34:00", "12:35:00", "12:36:00", "12:35:00+00", "12:35:00+01", "13:35:00+01", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00+01"]',
	'$[*].datetime() ? (@ <  "12:35".datetime("HH24:MI"))');`,
				Results: []sql.Row{{"12:34:00"}, {"12:35:00+01:00"}, {"13:35:00+01:00"}},
			},
			{
				Statement: `select jsonb_path_query(
	'["12:34:00+01", "12:35:00+01", "12:36:00+01", "12:35:00+02", "12:35:00-02", "10:35:00", "11:35:00", "12:35:00", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00 +1"]',
	'$[*].datetime() ? (@ == "12:35 +1".datetime("HH24:MI TZH"))');`,
				ErrorString: `cannot convert value from time to timetz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["12:34:00+01", "12:35:00+01", "12:36:00+01", "12:35:00+02", "12:35:00-02", "10:35:00", "11:35:00", "12:35:00", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00 +1"]',
	'$[*].datetime() ? (@ >= "12:35 +1".datetime("HH24:MI TZH"))');`,
				ErrorString: `cannot convert value from time to timetz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["12:34:00+01", "12:35:00+01", "12:36:00+01", "12:35:00+02", "12:35:00-02", "10:35:00", "11:35:00", "12:35:00", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00 +1"]',
	'$[*].datetime() ? (@ <  "12:35 +1".datetime("HH24:MI TZH"))');`,
				ErrorString: `cannot convert value from time to timetz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["12:34:00+01", "12:35:00+01", "12:36:00+01", "12:35:00+02", "12:35:00-02", "10:35:00", "11:35:00", "12:35:00", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00 +1"]',
	'$[*].datetime() ? (@ == "12:35 +1".datetime("HH24:MI TZH"))');`,
				Results: []sql.Row{{"12:35:00+01:00"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["12:34:00+01", "12:35:00+01", "12:36:00+01", "12:35:00+02", "12:35:00-02", "10:35:00", "11:35:00", "12:35:00", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00 +1"]',
	'$[*].datetime() ? (@ >= "12:35 +1".datetime("HH24:MI TZH"))');`,
				Results: []sql.Row{{"12:35:00+01:00"}, {"12:36:00+01:00"}, {"12:35:00-02:00"}, {"11:35:00"}, {"12:35:00"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["12:34:00+01", "12:35:00+01", "12:36:00+01", "12:35:00+02", "12:35:00-02", "10:35:00", "11:35:00", "12:35:00", "2017-03-10", "2017-03-10 12:35:00", "2017-03-10 12:35:00 +1"]',
	'$[*].datetime() ? (@ <  "12:35 +1".datetime("HH24:MI TZH"))');`,
				Results: []sql.Row{{"12:34:00+01:00"}, {"12:35:00+02:00"}, {"10:35:00"}},
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10 12:34:00", "2017-03-10 12:35:00", "2017-03-10 12:36:00", "2017-03-10 12:35:00+01", "2017-03-10 13:35:00+01", "2017-03-10 12:35:00-01", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ == "10.03.2017 12:35".datetime("dd.mm.yyyy HH24:MI"))');`,
				ErrorString: `cannot convert value from timestamp to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10 12:34:00", "2017-03-10 12:35:00", "2017-03-10 12:36:00", "2017-03-10 12:35:00+01", "2017-03-10 13:35:00+01", "2017-03-10 12:35:00-01", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ >= "10.03.2017 12:35".datetime("dd.mm.yyyy HH24:MI"))');`,
				ErrorString: `cannot convert value from timestamp to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10 12:34:00", "2017-03-10 12:35:00", "2017-03-10 12:36:00", "2017-03-10 12:35:00+01", "2017-03-10 13:35:00+01", "2017-03-10 12:35:00-01", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ < "10.03.2017 12:35".datetime("dd.mm.yyyy HH24:MI"))');`,
				ErrorString: `cannot convert value from timestamp to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10 12:34:00", "2017-03-10 12:35:00", "2017-03-10 12:36:00", "2017-03-10 12:35:00+01", "2017-03-10 13:35:00+01", "2017-03-10 12:35:00-01", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ == "10.03.2017 12:35".datetime("dd.mm.yyyy HH24:MI"))');`,
				Results: []sql.Row{{"2017-03-10T12:35:00"}, {"2017-03-10T13:35:00+01:00"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10 12:34:00", "2017-03-10 12:35:00", "2017-03-10 12:36:00", "2017-03-10 12:35:00+01", "2017-03-10 13:35:00+01", "2017-03-10 12:35:00-01", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ >= "10.03.2017 12:35".datetime("dd.mm.yyyy HH24:MI"))');`,
				Results: []sql.Row{{"2017-03-10T12:35:00"}, {"2017-03-10T12:36:00"}, {"2017-03-10T13:35:00+01:00"}, {"2017-03-10T12:35:00-01:00"}, {"2017-03-11"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10 12:34:00", "2017-03-10 12:35:00", "2017-03-10 12:36:00", "2017-03-10 12:35:00+01", "2017-03-10 13:35:00+01", "2017-03-10 12:35:00-01", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ < "10.03.2017 12:35".datetime("dd.mm.yyyy HH24:MI"))');`,
				Results: []sql.Row{{"2017-03-10T12:34:00"}, {"2017-03-10T12:35:00+01:00"}, {"2017-03-10"}},
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10 12:34:00+01", "2017-03-10 12:35:00+01", "2017-03-10 12:36:00+01", "2017-03-10 12:35:00+02", "2017-03-10 12:35:00-02", "2017-03-10 10:35:00", "2017-03-10 11:35:00", "2017-03-10 12:35:00", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ == "10.03.2017 12:35 +1".datetime("dd.mm.yyyy HH24:MI TZH"))');`,
				ErrorString: `cannot convert value from timestamp to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10 12:34:00+01", "2017-03-10 12:35:00+01", "2017-03-10 12:36:00+01", "2017-03-10 12:35:00+02", "2017-03-10 12:35:00-02", "2017-03-10 10:35:00", "2017-03-10 11:35:00", "2017-03-10 12:35:00", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ >= "10.03.2017 12:35 +1".datetime("dd.mm.yyyy HH24:MI TZH"))');`,
				ErrorString: `cannot convert value from timestamp to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query(
	'["2017-03-10 12:34:00+01", "2017-03-10 12:35:00+01", "2017-03-10 12:36:00+01", "2017-03-10 12:35:00+02", "2017-03-10 12:35:00-02", "2017-03-10 10:35:00", "2017-03-10 11:35:00", "2017-03-10 12:35:00", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ < "10.03.2017 12:35 +1".datetime("dd.mm.yyyy HH24:MI TZH"))');`,
				ErrorString: `cannot convert value from timestamp to timestamptz without time zone usage`,
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10 12:34:00+01", "2017-03-10 12:35:00+01", "2017-03-10 12:36:00+01", "2017-03-10 12:35:00+02", "2017-03-10 12:35:00-02", "2017-03-10 10:35:00", "2017-03-10 11:35:00", "2017-03-10 12:35:00", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ == "10.03.2017 12:35 +1".datetime("dd.mm.yyyy HH24:MI TZH"))');`,
				Results: []sql.Row{{"2017-03-10T12:35:00+01:00"}, {"2017-03-10T11:35:00"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10 12:34:00+01", "2017-03-10 12:35:00+01", "2017-03-10 12:36:00+01", "2017-03-10 12:35:00+02", "2017-03-10 12:35:00-02", "2017-03-10 10:35:00", "2017-03-10 11:35:00", "2017-03-10 12:35:00", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ >= "10.03.2017 12:35 +1".datetime("dd.mm.yyyy HH24:MI TZH"))');`,
				Results: []sql.Row{{"2017-03-10T12:35:00+01:00"}, {"2017-03-10T12:36:00+01:00"}, {"2017-03-10T12:35:00-02:00"}, {"2017-03-10T11:35:00"}, {"2017-03-10T12:35:00"}, {"2017-03-11"}},
			},
			{
				Statement: `select jsonb_path_query_tz(
	'["2017-03-10 12:34:00+01", "2017-03-10 12:35:00+01", "2017-03-10 12:36:00+01", "2017-03-10 12:35:00+02", "2017-03-10 12:35:00-02", "2017-03-10 10:35:00", "2017-03-10 11:35:00", "2017-03-10 12:35:00", "2017-03-10", "2017-03-11", "12:34:56", "12:34:56+01"]',
	'$[*].datetime() ? (@ < "10.03.2017 12:35 +1".datetime("dd.mm.yyyy HH24:MI TZH"))');`,
				Results: []sql.Row{{"2017-03-10T12:34:00+01:00"}, {"2017-03-10T12:35:00+02:00"}, {"2017-03-10T10:35:00"}, {"2017-03-10"}},
			},
			{
				Statement: `select jsonb_path_query('"1000000-01-01"', '$.datetime() > "2020-01-01 12:00:00".datetime()'::jsonpath);`,
				Results:   []sql.Row{{`true`}},
			},
			{
				Statement: `set time zone default;`,
			},
			{
				Statement: `SELECT jsonb_path_query('[{"a": 1}, {"a": 2}]', '$[*]');`,
				Results:   []sql.Row{{`{"a": 1}`}, {`{"a": 2}`}},
			},
			{
				Statement: `SELECT jsonb_path_query('[{"a": 1}, {"a": 2}]', '$[*] ? (@.a > 10)');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT jsonb_path_query('[{"a": 1}]', '$undefined_var');`,
				ErrorString: `could not find jsonpath variable "undefined_var"`,
			},
			{
				Statement: `SELECT jsonb_path_query('[{"a": 1}]', 'false');`,
				Results:   []sql.Row{{`false`}},
			},
			{
				Statement:   `SELECT jsonb_path_query_array('[{"a": 1}, {"a": 2}, {}]', 'strict $[*].a');`,
				ErrorString: `JSON object does not contain key "a"`,
			},
			{
				Statement: `SELECT jsonb_path_query_array('[{"a": 1}, {"a": 2}]', '$[*].a');`,
				Results:   []sql.Row{{`[1, 2]`}},
			},
			{
				Statement: `SELECT jsonb_path_query_array('[{"a": 1}, {"a": 2}]', '$[*].a ? (@ == 1)');`,
				Results:   []sql.Row{{`[1]`}},
			},
			{
				Statement: `SELECT jsonb_path_query_array('[{"a": 1}, {"a": 2}]', '$[*].a ? (@ > 10)');`,
				Results:   []sql.Row{{`[]`}},
			},
			{
				Statement: `SELECT jsonb_path_query_array('[{"a": 1}, {"a": 2}, {"a": 3}, {"a": 5}]', '$[*].a ? (@ > $min && @ < $max)', vars => '{"min": 1, "max": 4}');`,
				Results:   []sql.Row{{`[2, 3]`}},
			},
			{
				Statement: `SELECT jsonb_path_query_array('[{"a": 1}, {"a": 2}, {"a": 3}, {"a": 5}]', '$[*].a ? (@ > $min && @ < $max)', vars => '{"min": 3, "max": 4}');`,
				Results:   []sql.Row{{`[]`}},
			},
			{
				Statement:   `SELECT jsonb_path_query_first('[{"a": 1}, {"a": 2}, {}]', 'strict $[*].a');`,
				ErrorString: `JSON object does not contain key "a"`,
			},
			{
				Statement: `SELECT jsonb_path_query_first('[{"a": 1}, {"a": 2}, {}]', 'strict $[*].a', silent => true);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT jsonb_path_query_first('[{"a": 1}, {"a": 2}]', '$[*].a');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT jsonb_path_query_first('[{"a": 1}, {"a": 2}]', '$[*].a ? (@ == 1)');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT jsonb_path_query_first('[{"a": 1}, {"a": 2}]', '$[*].a ? (@ > 10)');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT jsonb_path_query_first('[{"a": 1}, {"a": 2}, {"a": 3}, {"a": 5}]', '$[*].a ? (@ > $min && @ < $max)', vars => '{"min": 1, "max": 4}');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT jsonb_path_query_first('[{"a": 1}, {"a": 2}, {"a": 3}, {"a": 5}]', '$[*].a ? (@ > $min && @ < $max)', vars => '{"min": 3, "max": 4}');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT jsonb_path_query_first('[{"a": 1}]', '$undefined_var');`,
				ErrorString: `could not find jsonpath variable "undefined_var"`,
			},
			{
				Statement: `SELECT jsonb_path_query_first('[{"a": 1}]', 'false');`,
				Results:   []sql.Row{{`false`}},
			},
			{
				Statement: `SELECT jsonb '[{"a": 1}, {"a": 2}]' @? '$[*].a ? (@ > 1)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT jsonb '[{"a": 1}, {"a": 2}]' @? '$[*] ? (@.a > 2)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT jsonb_path_exists('[{"a": 1}, {"a": 2}]', '$[*].a ? (@ > 1)');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT jsonb_path_exists('[{"a": 1}, {"a": 2}, {"a": 3}, {"a": 5}]', '$[*] ? (@.a > $min && @.a < $max)', vars => '{"min": 1, "max": 4}');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT jsonb_path_exists('[{"a": 1}, {"a": 2}, {"a": 3}, {"a": 5}]', '$[*] ? (@.a > $min && @.a < $max)', vars => '{"min": 3, "max": 4}');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `SELECT jsonb_path_exists('[{"a": 1}]', '$undefined_var');`,
				ErrorString: `could not find jsonpath variable "undefined_var"`,
			},
			{
				Statement: `SELECT jsonb_path_exists('[{"a": 1}]', 'false');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT jsonb_path_match('true', '$', silent => false);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT jsonb_path_match('false', '$', silent => false);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT jsonb_path_match('null', '$', silent => false);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT jsonb_path_match('1', '$', silent => true);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT jsonb_path_match('1', '$', silent => false);`,
				ErrorString: `single boolean result is expected`,
			},
			{
				Statement:   `SELECT jsonb_path_match('"a"', '$', silent => false);`,
				ErrorString: `single boolean result is expected`,
			},
			{
				Statement:   `SELECT jsonb_path_match('{}', '$', silent => false);`,
				ErrorString: `single boolean result is expected`,
			},
			{
				Statement:   `SELECT jsonb_path_match('[true]', '$', silent => false);`,
				ErrorString: `single boolean result is expected`,
			},
			{
				Statement:   `SELECT jsonb_path_match('{}', 'lax $.a', silent => false);`,
				ErrorString: `single boolean result is expected`,
			},
			{
				Statement:   `SELECT jsonb_path_match('{}', 'strict $.a', silent => false);`,
				ErrorString: `JSON object does not contain key "a"`,
			},
			{
				Statement: `SELECT jsonb_path_match('{}', 'strict $.a', silent => true);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT jsonb_path_match('[true, true]', '$[*]', silent => false);`,
				ErrorString: `single boolean result is expected`,
			},
			{
				Statement: `SELECT jsonb '[{"a": 1}, {"a": 2}]' @@ '$[*].a > 1';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT jsonb '[{"a": 1}, {"a": 2}]' @@ '$[*].a > 2';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT jsonb_path_match('[{"a": 1}, {"a": 2}]', '$[*].a > 1');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `SELECT jsonb_path_match('[{"a": 1}]', '$undefined_var');`,
				ErrorString: `could not find jsonpath variable "undefined_var"`,
			},
			{
				Statement: `SELECT jsonb_path_match('[{"a": 1}]', 'false');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `WITH str(j, num) AS
(
	SELECT jsonb_build_object('s', s), num
	FROM unnest('{"", "a", "ab", "abc", "abcd", "b", "A", "AB", "ABC", "ABc", "ABcD", "B"}'::text[]) WITH ORDINALITY AS a(s, num)
)
SELECT
	s1.j, s2.j,
	jsonb_path_query_first(s1.j, '$.s < $s', vars => s2.j) lt,
	jsonb_path_query_first(s1.j, '$.s <= $s', vars => s2.j) le,
	jsonb_path_query_first(s1.j, '$.s == $s', vars => s2.j) eq,
	jsonb_path_query_first(s1.j, '$.s >= $s', vars => s2.j) ge,
	jsonb_path_query_first(s1.j, '$.s > $s', vars => s2.j) gt
FROM str s1, str s2
ORDER BY s1.num, s2.num;`,
				Results: []sql.Row{{`{"s": ""}`, `{"s": ""}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": ""}`, `{"s": "a"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "ab"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "A"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "AB"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "ABC"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "ABc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "ABcD"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": ""}`, `{"s": "B"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "a"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "a"}`, `{"s": "a"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "a"}`, `{"s": "ab"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "a"}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "a"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "a"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "a"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "a"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "a"}`, `{"s": "ABC"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "a"}`, `{"s": "ABc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "a"}`, `{"s": "ABcD"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "a"}`, `{"s": "B"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ab"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ab"}`, `{"s": "a"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ab"}`, `{"s": "ab"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "ab"}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ab"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ab"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ab"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ab"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ab"}`, `{"s": "ABC"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ab"}`, `{"s": "ABc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ab"}`, `{"s": "ABcD"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ab"}`, `{"s": "B"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": "a"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": "ab"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": "abc"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "abc"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "abc"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "abc"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": "ABC"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": "ABc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": "ABcD"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abc"}`, `{"s": "B"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "a"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "ab"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "abc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "abcd"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "abcd"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "abcd"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "ABC"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "ABc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "ABcD"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "abcd"}`, `{"s": "B"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "a"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "ab"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "abc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "abcd"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "b"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "b"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "ABC"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "ABc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "ABcD"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "b"}`, `{"s": "B"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "A"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "A"}`, `{"s": "a"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "ab"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "A"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "A"}`, `{"s": "AB"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "ABC"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "ABc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "ABcD"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "A"}`, `{"s": "B"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "AB"}`, `{"s": "a"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": "ab"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "AB"}`, `{"s": "AB"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "AB"}`, `{"s": "ABC"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": "ABc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": "ABcD"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "AB"}`, `{"s": "B"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABC"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABC"}`, `{"s": "a"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABC"}`, `{"s": "ab"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABC"}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABC"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABC"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABC"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABC"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABC"}`, `{"s": "ABC"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "ABC"}`, `{"s": "ABc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABC"}`, `{"s": "ABcD"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABC"}`, `{"s": "B"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABc"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABc"}`, `{"s": "a"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABc"}`, `{"s": "ab"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABc"}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABc"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABc"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABc"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABc"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABc"}`, `{"s": "ABC"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABc"}`, `{"s": "ABc"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "ABc"}`, `{"s": "ABcD"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABc"}`, `{"s": "B"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABcD"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABcD"}`, `{"s": "a"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABcD"}`, `{"s": "ab"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABcD"}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABcD"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABcD"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "ABcD"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABcD"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABcD"}`, `{"s": "ABC"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABcD"}`, `{"s": "ABc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "ABcD"}`, `{"s": "ABcD"}`, `false`, `true`, `true`, `true`, `false`}, {`{"s": "ABcD"}`, `{"s": "B"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "B"}`, `{"s": ""}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "B"}`, `{"s": "a"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "B"}`, `{"s": "ab"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "B"}`, `{"s": "abc"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "B"}`, `{"s": "abcd"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "B"}`, `{"s": "b"}`, `true`, `true`, `false`, `false`, `false`}, {`{"s": "B"}`, `{"s": "A"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "B"}`, `{"s": "AB"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "B"}`, `{"s": "ABC"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "B"}`, `{"s": "ABc"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "B"}`, `{"s": "ABcD"}`, `false`, `false`, `false`, `true`, `true`}, {`{"s": "B"}`, `{"s": "B"}`, `false`, `true`, `true`, `true`, `false`}},
			},
		},
	})
}
