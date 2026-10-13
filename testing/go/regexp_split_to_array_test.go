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

func TestRegexpSplitToArray(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "regexp_split_to_array", Assertions: []ScriptTestAssertion{
		{Query: `SELECT regexp_split_to_array('statement_timeout=1000', '=');`, Expected: []sql.Row{{"{statement_timeout,1000}"}}},
		{Query: `SELECT (regexp_split_to_array((regexp_split_to_array('default_transaction_isolation=read committed,serializable', '='))[2], ','))[1];`, Expected: []sql.Row{{"read committed"}}},
		{Query: `SELECT regexp_split_to_array('a,b;c', '[,;]');`, Expected: []sql.Row{{"{a,b,c}"}}},
		{Query: `SELECT regexp_split_to_array('a,b,c', '(,)');`, Expected: []sql.Row{{"{a,b,c}"}}},
		{Query: `SELECT regexp_split_to_array(',a,,b,', ',');`, Expected: []sql.Row{{"{\"\",a,\"\",b,\"\"}"}}},
		{Query: `SELECT regexp_split_to_array('abc', 'xyz');`, Expected: []sql.Row{{"{abc}"}}},
		{Query: `SELECT regexp_split_to_array('', ',');`, Expected: []sql.Row{{"{\"\"}"}}},
		{Query: `SELECT regexp_split_to_array('', '');`, Expected: []sql.Row{{"{\"\"}"}}},
		{Query: `SELECT regexp_split_to_array('abc', '');`, Expected: []sql.Row{{"{a,b,c}"}}},
		{Query: `SELECT regexp_split_to_array('abc', '^|$');`, Expected: []sql.Row{{"{abc}"}}},
		{Query: `SELECT regexp_split_to_array('abc', 'b*');`, Expected: []sql.Row{{"{a,c}"}}},
		{Query: `SELECT regexp_split_to_array('a b  c', E'\\s*');`, Expected: []sql.Row{{"{a,b,c}"}}},
		{Query: `SELECT regexp_split_to_array('日本語', '');`, Expected: []sql.Row{{"{日,本,語}"}}},
		{Query: `SELECT regexp_split_to_array('aXXbxxc', 'xx', 'i');`, Expected: []sql.Row{{"{a,b,c}"}}},
		{Query: `SELECT regexp_split_to_array('aXXbxxc', 'xx', 'ic');`, Expected: []sql.Row{{"{aXXb,c}"}}},
		{Query: `SELECT regexp_split_to_array('aXXbxxc', 'xx', 'ci');`, Expected: []sql.Row{{"{a,b,c}"}}},
		{Query: `SELECT regexp_split_to_array('a.b*c', '.', 'q');`, Expected: []sql.Row{{"{a,b*c}"}}},
		{Query: `SELECT regexp_split_to_array('a.b*c', '.', 'qi');`, Expected: []sql.Row{{"{a,b*c}"}}},
		{Query: `SELECT regexp_split_to_array('a,b', ',', 'e');`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array('a+b++c', '+', 'b');`, Expected: []sql.Row{{"{a,b,\"\",c}"}}},
		{Query: `SELECT regexp_split_to_array('a,,b', E'\\(,\\)\\{2\\}', 'b');`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '.');`, Expected: []sql.Row{{"{\"\",\"\",\"\",\"\"}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '.', 'n');`, Expected: []sql.Row{{"{\"\",\"\n\",\"\"}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '.', 'p');`, Expected: []sql.Row{{"{\"\",\"\n\",\"\"}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '.', 'w');`, Expected: []sql.Row{{"{\"\",\"\",\"\",\"\"}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '.', 's');`, Expected: []sql.Row{{"{\"\",\"\",\"\",\"\"}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '^', 'n');`, Expected: []sql.Row{{"{\"a\n\",b}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '^', 'm');`, Expected: []sql.Row{{"{\"a\n\",b}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '^', 'p');`, Expected: []sql.Row{{"{\"a\nb\"}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '^', 'w');`, Expected: []sql.Row{{"{\"a\n\",b}"}}},
		{Query: `SELECT regexp_split_to_array(E'a\nb', '[^,]+', 'n');`, Expected: []sql.Row{{"{\"\",\"\n\",\"\"}"}}},
		{Query: `SELECT regexp_split_to_array('a,b;c', ' , | ; # delimiters', 'x');`, Expected: []sql.Row{{"{a,b,c}"}}},
		{Query: `SELECT regexp_split_to_array('a b#c', '[ #]', 'x');`, Expected: []sql.Row{{"{a,b,c}"}}},
		{Query: `SELECT regexp_split_to_array('a b', ' ', 'xt');`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array('a,b', '[[:space:],]', 'x');`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array('aXXb', '(?i)xx');`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array('aXXbxxc', '(?c)xx', 'i');`, Expected: []sql.Row{{"{aXXb,c}"}}},
		{Query: `SELECT regexp_split_to_array('a.b', '***=.');`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array('a,b', '***:,', 'b');`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array('abX', 'a|ab');`, Expected: []sql.Row{{"{\"\",X}"}}},
		{Query: `SELECT regexp_split_to_array('xabcy', 'a.*?c');`, Expected: []sql.Row{{"{x,y}"}}},
		{Query: `SELECT regexp_split_to_array(NULL, ',');`, Expected: []sql.Row{{nil}}},
		{Query: `SELECT regexp_split_to_array('abc', NULL);`, Expected: []sql.Row{{nil}}},
		{Query: `SELECT regexp_split_to_array('abc', ',', NULL);`, Expected: []sql.Row{{nil}}},
		{Query: `SELECT regexp_split_to_array(NULL, '[', 'g');`, Expected: []sql.Row{{nil}}},
		{Query: `SELECT regexp_split_to_array('a,b'::varchar, ','::text);`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array('abc', '[');`, ExpectedErr: "invalid regular expression", ExpectedErrCode: "2201B"},
		{Query: `SELECT regexp_split_to_array('abc', ',','g');`, ExpectedErr: "does not support the \"global\" option", ExpectedErrCode: "22023"},
		{Query: `SELECT regexp_split_to_array('abc', ',','z');`, ExpectedErr: "invalid regular expression option", ExpectedErrCode: "22023"},
		{Query: `SELECT regexp_split_to_array('abc', ',','I');`, ExpectedErr: "invalid regular expression option", ExpectedErrCode: "22023"},
		{Query: `SELECT regexp_split_to_array('abc', ',',' i');`, ExpectedErr: "invalid regular expression option", ExpectedErrCode: "22023"},
		{Query: `SELECT regexp_split_to_array('abc'::text);`, ExpectedErr: "does not exist"},
		{Query: `SELECT regexp_split_to_array(123, ','::text);`, ExpectedErr: "does not exist"},
	}}})
}

// TestRegexpSplitToArrayRows checks row-dependent patterns and flags.
func TestRegexpSplitToArrayRows(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "regexp_split_to_array rows", Assertions: []ScriptTestAssertion{
		{Query: `CREATE TABLE split_inputs (id int PRIMARY KEY, value text, delimiter text, flags text);`},
		{Query: `INSERT INTO split_inputs VALUES (1, 'aXb', 'x', 'i'), (2, 'c,d', ',', ''), (3, NULL, '[', 'g'), (4, 'text', NULL, ''), (5, 'text', ',', NULL);`},
		{Query: `SELECT id, regexp_split_to_array(value, delimiter, flags) FROM split_inputs ORDER BY id;`, Expected: []sql.Row{{int32(1), "{a,b}"}, {int32(2), "{c,d}"}, {int32(3), nil}, {int32(4), nil}, {int32(5), nil}}},
		{Query: `SELECT pg_catalog.regexp_split_to_array('a,b', ',');`, Expected: []sql.Row{{"{a,b}"}}},
		{Query: `SELECT regexp_split_to_array('a,b', ',', '', '');`, ExpectedErr: "does not exist"},
	}}})
}
