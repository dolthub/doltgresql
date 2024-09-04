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

func TestMultirangetypes(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_multirangetypes)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_multirangetypes,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup, RegressionFileName_rangetypes},
		Statements: []RegressionFileStatement{
			{
				Statement:   `select ''::textmultirange;`,
				ErrorString: `malformed multirange literal: ""`,
			},
			{
				Statement:   `select '{,}'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{,}"`,
			},
			{
				Statement:   `select '{(,)}.'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{(,)}."`,
			},
			{
				Statement:   `select '{[a,c),}'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{[a,c),}"`,
			},
			{
				Statement:   `select '{,[a,c)}'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{,[a,c)}"`,
			},
			{
				Statement:   `select '{-[a,z)}'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{-[a,z)}"`,
			},
			{
				Statement:   `select '{[a,z) - }'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{[a,z) - }"`,
			},
			{
				Statement:   `select '{(",a)}'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{(",a)}"`,
			},
			{
				Statement:   `select '{(,,a)}'::textmultirange;`,
				ErrorString: `malformed range literal: "(,,a)"`,
			},
			{
				Statement:   `select '{(),a)}'::textmultirange;`,
				ErrorString: `malformed range literal: "()"`,
			},
			{
				Statement:   `select '{(a,))}'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{(a,))}"`,
			},
			{
				Statement:   `select '{(],a)}'::textmultirange;`,
				ErrorString: `malformed range literal: "(]"`,
			},
			{
				Statement:   `select '{(a,])}'::textmultirange;`,
				ErrorString: `malformed multirange literal: "{(a,])}"`,
			},
			{
				Statement:   `select '{[z,a]}'::textmultirange;`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `select '{}'::textmultirange;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select '  {}  '::textmultirange;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select ' { empty, empty }  '::textmultirange;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select ' {( " a " " a ", " z " " z " )  }'::textmultirange;`,
				Results:   []sql.Row{{`{("  a   a ","  z   z  ")}`}},
			},
			{
				Statement: `select textrange('\\\\', repeat('a', 200))::textmultirange;`,
				Results:   []sql.Row{{`{["\\\\\\\\",aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa)}`}},
			},
			{
				Statement: `select '{(,z)}'::textmultirange;`,
				Results:   []sql.Row{{`{(,z)}`}},
			},
			{
				Statement: `select '{(a,)}'::textmultirange;`,
				Results:   []sql.Row{{`{(a,)}`}},
			},
			{
				Statement: `select '{[,z]}'::textmultirange;`,
				Results:   []sql.Row{{`{(,z]}`}},
			},
			{
				Statement: `select '{[a,]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,)}`}},
			},
			{
				Statement: `select '{(,)}'::textmultirange;`,
				Results:   []sql.Row{{`{(,)}`}},
			},
			{
				Statement: `select '{[ , ]}'::textmultirange;`,
				Results:   []sql.Row{{`{[" "," "]}`}},
			},
			{
				Statement: `select '{["",""]}'::textmultirange;`,
				Results:   []sql.Row{{`{["",""]}`}},
			},
			{
				Statement: `select '{[",",","]}'::textmultirange;`,
				Results:   []sql.Row{{`{[",",","]}`}},
			},
			{
				Statement: `select '{["\\","\\"]}'::textmultirange;`,
				Results:   []sql.Row{{`{["\\","\\"]}`}},
			},
			{
				Statement: `select '{["""","\""]}'::textmultirange;`,
				Results:   []sql.Row{{`{["""",""""]}`}},
			},
			{
				Statement: `select '{(\\,a)}'::textmultirange;`,
				Results:   []sql.Row{{`{("\\",a)}`}},
			},
			{
				Statement: `select '{((,z)}'::textmultirange;`,
				Results:   []sql.Row{{`{("(",z)}`}},
			},
			{
				Statement: `select '{([,z)}'::textmultirange;`,
				Results:   []sql.Row{{`{("[",z)}`}},
			},
			{
				Statement: `select '{(!,()}'::textmultirange;`,
				Results:   []sql.Row{{`{(!,"(")}`}},
			},
			{
				Statement: `select '{(!,[)}'::textmultirange;`,
				Results:   []sql.Row{{`{(!,"[")}`}},
			},
			{
				Statement: `select '{[a,a]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,a]}`}},
			},
			{
				Statement: `select '{[a,a],[a,b]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,b]}`}},
			},
			{
				Statement: `select '{[a,b), [b,e]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,e]}`}},
			},
			{
				Statement: `select '{[a,d), [b,f]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,f]}`}},
			},
			{
				Statement: `select '{[a,a],[b,b]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,a],[b,b]}`}},
			},
			{
				Statement: `select '{[a,a], [b,b]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,a],[b,b]}`}},
			},
			{
				Statement: `select '{[1,2], [3,4]}'::int4multirange;`,
				Results:   []sql.Row{{`{[1,5)}`}},
			},
			{
				Statement: `select '{[a,a], [b,b], [c,c]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,a],[b,b],[c,c]}`}},
			},
			{
				Statement: `select '{[a,d], [b,e]}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,e]}`}},
			},
			{
				Statement: `select '{[a,d), [d,e)}'::textmultirange;`,
				Results:   []sql.Row{{`{[a,e)}`}},
			},
			{
				Statement: `select '{[a,a)}'::textmultirange;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select '{(a,a]}'::textmultirange;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select '{(a,a)}'::textmultirange;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `---
select textmultirange();`,
				Results: []sql.Row{{`{}`}},
			},
			{
				Statement: `select textmultirange(textrange('a', 'c'));`,
				Results:   []sql.Row{{`{[a,c)}`}},
			},
			{
				Statement: `select textmultirange(textrange('a', 'c'), textrange('f', 'g'));`,
				Results:   []sql.Row{{`{[a,c),[f,g)}`}},
			},
			{
				Statement: `select textmultirange(textrange('\\\\', repeat('a', 200)), textrange('c', 'd'));`,
				Results:   []sql.Row{{`{["\\\\\\\\",aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa),[c,d)}`}},
			},
			{
				Statement: `select 'empty'::int4range::int4multirange;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select int4range(1, 3)::int4multirange;`,
				Results:   []sql.Row{{`{[1,3)}`}},
			},
			{
				Statement: `select int4range(1, null)::int4multirange;`,
				Results:   []sql.Row{{`{[1,)}`}},
			},
			{
				Statement: `select int4range(null, null)::int4multirange;`,
				Results:   []sql.Row{{`{(,)}`}},
			},
			{
				Statement: `select 'empty'::textrange::textmultirange;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select textrange('a', 'c')::textmultirange;`,
				Results:   []sql.Row{{`{[a,c)}`}},
			},
			{
				Statement: `select textrange('a', null)::textmultirange;`,
				Results:   []sql.Row{{`{[a,)}`}},
			},
			{
				Statement: `select textrange(null, null)::textmultirange;`,
				Results:   []sql.Row{{`{(,)}`}},
			},
			{
				Statement: `select unnest(int4multirange(int4range('5', '6'), int4range('1', '2')));`,
				Results:   []sql.Row{{`[1,2)`}, {`[5,6)`}},
			},
			{
				Statement: `select unnest(textmultirange(textrange('a', 'b'), textrange('d', 'e')));`,
				Results:   []sql.Row{{`[a,b)`}, {`[d,e)`}},
			},
			{
				Statement: `select unnest(textmultirange(textrange('\\\\', repeat('a', 200)), textrange('c', 'd')));`,
				Results:   []sql.Row{{`["\\\\\\\\",aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa)`}, {`[c,d)`}},
			},
			{
				Statement: `CREATE TABLE nummultirange_test (nmr NUMMULTIRANGE);`,
			},
			{
				Statement: `CREATE INDEX nummultirange_test_btree ON nummultirange_test(nmr);`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES('{}');`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES('{[,)}');`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES('{[3,]}');`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES('{[,), [3,]}');`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES('{[, 5)}');`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES(nummultirange());`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES(nummultirange(variadic '{}'::numrange[]));`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES(nummultirange(numrange(1.1, 2.2)));`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES('{empty}');`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES(nummultirange(numrange(1.7, 1.7, '[]'), numrange(1.7, 1.9)));`,
			},
			{
				Statement: `INSERT INTO nummultirange_test VALUES(nummultirange(numrange(1.7, 1.7, '[]'), numrange(1.9, 2.1)));`,
			},
			{
				Statement: `SELECT nmr, isempty(nmr), lower(nmr), upper(nmr) FROM nummultirange_test ORDER BY nmr;`,
				Results:   []sql.Row{{`{}`, true, ``, ``}, {`{}`, true, ``, ``}, {`{}`, true, ``, ``}, {`{}`, true, ``, ``}, {`{(,5)}`, false, ``, 5}, {`{(,)}`, false, ``, ``}, {`{(,)}`, false, ``, ``}, {`{[1.1,2.2)}`, false, 1.1, 2.2}, {`{[1.7,1.7],[1.9,2.1)}`, false, 1.7, 2.1}, {`{[1.7,1.9)}`, false, 1.7, 1.9}, {`{[3,)}`, false, 3, ``}},
			},
			{
				Statement: `SELECT nmr, lower_inc(nmr), lower_inf(nmr), upper_inc(nmr), upper_inf(nmr) FROM nummultirange_test ORDER BY nmr;`,
				Results:   []sql.Row{{`{}`, false, false, false, false}, {`{}`, false, false, false, false}, {`{}`, false, false, false, false}, {`{}`, false, false, false, false}, {`{(,5)}`, false, true, false, false}, {`{(,)}`, false, true, false, true}, {`{(,)}`, false, true, false, true}, {`{[1.1,2.2)}`, true, false, false, false}, {`{[1.7,1.7],[1.9,2.1)}`, true, false, false, false}, {`{[1.7,1.9)}`, true, false, false, false}, {`{[3,)}`, true, false, false, true}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr = '{}';`,
				Results:   []sql.Row{{`{}`}, {`{}`}, {`{}`}, {`{}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr = '{(,5)}';`,
				Results:   []sql.Row{{`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr = '{[3,)}';`,
				Results:   []sql.Row{{`{[3,)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr = '{[1.7,1.7]}';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr = '{[1.7,1.7],[1.9,2.1)}';`,
				Results:   []sql.Row{{`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr < '{}';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr < '{[-1000.0, -1000.0]}';`,
				Results:   []sql.Row{{`{}`}, {`{(,)}`}, {`{(,)}`}, {`{(,5)}`}, {`{}`}, {`{}`}, {`{}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr < '{[0.0, 1.0]}';`,
				Results:   []sql.Row{{`{}`}, {`{(,)}`}, {`{(,)}`}, {`{(,5)}`}, {`{}`}, {`{}`}, {`{}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr < '{[1000.0, 1001.0]}';`,
				Results:   []sql.Row{{`{}`}, {`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}, {`{}`}, {`{}`}, {`{[1.1,2.2)}`}, {`{}`}, {`{[1.7,1.9)}`}, {`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr <= '{}';`,
				Results:   []sql.Row{{`{}`}, {`{}`}, {`{}`}, {`{}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr <= '{[3,)}';`,
				Results:   []sql.Row{{`{}`}, {`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}, {`{}`}, {`{}`}, {`{[1.1,2.2)}`}, {`{}`}, {`{[1.7,1.9)}`}, {`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr >= '{}';`,
				Results:   []sql.Row{{`{}`}, {`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}, {`{}`}, {`{}`}, {`{[1.1,2.2)}`}, {`{}`}, {`{[1.7,1.9)}`}, {`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr >= '{[3,)}';`,
				Results:   []sql.Row{{`{[3,)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr > '{}';`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}, {`{[1.1,2.2)}`}, {`{[1.7,1.9)}`}, {`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr > '{[-1000.0, -1000.0]}';`,
				Results:   []sql.Row{{`{[3,)}`}, {`{[1.1,2.2)}`}, {`{[1.7,1.9)}`}, {`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr > '{[0.0, 1.0]}';`,
				Results:   []sql.Row{{`{[3,)}`}, {`{[1.1,2.2)}`}, {`{[1.7,1.9)}`}, {`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr > '{[1000.0, 1001.0]}';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr <> '{}';`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}, {`{[1.1,2.2)}`}, {`{[1.7,1.9)}`}, {`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr <> '{(,5)}';`,
				Results:   []sql.Row{{`{}`}, {`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{}`}, {`{}`}, {`{[1.1,2.2)}`}, {`{}`}, {`{[1.7,1.9)}`}, {`{[1.7,1.7],[1.9,2.1)}`}},
			},
			{
				Statement:   `select nummultirange(numrange(2.0, 1.0));`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `select nummultirange(numrange(5.0, 6.0), numrange(1.0, 2.0));`,
				Results:   []sql.Row{{`{[1.0,2.0),[5.0,6.0)}`}},
			},
			{
				Statement: `analyze nummultirange_test;`,
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE range_overlaps_multirange(numrange(4.0, 4.2), nmr);`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE numrange(4.0, 4.2) && nmr;`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE multirange_overlaps_range(nmr, numrange(4.0, 4.2));`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr && numrange(4.0, 4.2);`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE multirange_overlaps_multirange(nmr, nummultirange(numrange(4.0, 4.2), numrange(6.0, 7.0)));`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr && nummultirange(numrange(4.0, 4.2), numrange(6.0, 7.0));`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr && nummultirange(numrange(6.0, 7.0));`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr && nummultirange(numrange(6.0, 7.0), numrange(8.0, 9.0));`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE multirange_contains_elem(nmr, 4.0);`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr @> 4.0;`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE multirange_contains_range(nmr, numrange(4.0, 4.2));`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr @> numrange(4.0, 4.2);`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE multirange_contains_multirange(nmr, '{[4.0,4.2), [6.0, 8.0)}');`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE nmr @> '{[4.0,4.2), [6.0, 8.0)}'::nummultirange;`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE elem_contained_by_multirange(4.0, nmr);`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE 4.0 <@ nmr;`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE range_contained_by_multirange(numrange(4.0, 4.2), nmr);`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE numrange(4.0, 4.2) <@ nmr;`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}, {`{(,5)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE multirange_contained_by_multirange('{[4.0,4.2), [6.0, 8.0)}', nmr);`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}},
			},
			{
				Statement: `SELECT * FROM nummultirange_test WHERE '{[4.0,4.2), [6.0, 8.0)}'::nummultirange <@ nmr;`,
				Results:   []sql.Row{{`{(,)}`}, {`{[3,)}`}, {`{(,)}`}},
			},
			{
				Statement: `SELECT 'empty'::numrange && nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'empty'::numrange && nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() && 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) && 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() && nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() && nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) && nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) && nummultirange(numrange(1,2), numrange(7,8));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(7,8)) && nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) && nummultirange(numrange(1,2), numrange(3.5,8));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(3.5,8)) && numrange(3,4);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(3.5,8)) && nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select '{(10,20),(30,40),(50,60)}'::nummultirange && '(42,92)'::numrange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange() @> nummultirange();`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange() @> 'empty'::numrange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(null,null)) @> numrange(1,2);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(null,null)) @> numrange(null,2);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(null,null)) @> numrange(2,null);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(null,5)) @> numrange(null,3);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(null,5)) @> numrange(null,8);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(5,null)) @> numrange(8,null);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(5,null)) @> numrange(3,null);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5)) @> numrange(8,9);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5)) @> numrange(3,9);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5)) @> numrange(1,4);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5)) @> numrange(1,5);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(-4,-2), numrange(1,5)) @> numrange(1,5);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5), numrange(8,9)) @> numrange(1,5);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5), numrange(8,9)) @> numrange(6,7);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5), numrange(6,9)) @> numrange(6,7);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,5)}'::nummultirange @> '{[1,5)}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[-4,-2), [1,5)}'::nummultirange @> '{[1,5)}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,5), [8,9)}'::nummultirange @> '{[1,5)}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,5), [8,9)}'::nummultirange @> '{[6,7)}';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT '{[1,5), [6,9)}'::nummultirange @> '{[6,7)}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select '{(10,20),(30,40),(50,60)}'::nummultirange @> '(52,56)'::numrange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(null,null) @> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(null,null) @> nummultirange(numrange(null,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(null,null) @> nummultirange(numrange(2,null));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(null,5) @> nummultirange(numrange(null,3));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(null,5) @> nummultirange(numrange(null,8));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(5,null) @> nummultirange(numrange(8,null));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(5,null) @> nummultirange(numrange(3,null));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(1,5) @> nummultirange(numrange(8,9));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(1,5) @> nummultirange(numrange(3,9));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(1,5) @> nummultirange(numrange(1,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,5) @> nummultirange(numrange(1,5));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,9) @> nummultirange(numrange(-4,-2), numrange(1,5));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(1,9) @> nummultirange(numrange(1,5), numrange(8,9));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,9) @> nummultirange(numrange(1,5), numrange(6,9));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,9) @> nummultirange(numrange(1,5), numrange(6,10));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT '{[1,9)}' @> '{[1,5)}'::nummultirange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,9)}' @> '{[-4,-2), [1,5)}'::nummultirange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT '{[1,9)}' @> '{[1,5), [8,9)}'::nummultirange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,9)}' @> '{[1,5), [6,9)}'::nummultirange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,9)}' @> '{[1,5), [6,10)}'::nummultirange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() <@ nummultirange();`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT 'empty'::numrange <@ nummultirange();`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,2) <@ nummultirange(numrange(null,null));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(null,2) <@ nummultirange(numrange(null,null));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(2,null) <@ nummultirange(numrange(null,null));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(null,3) <@ nummultirange(numrange(null,5));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(null,8) <@ nummultirange(numrange(null,5));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(8,null) <@ nummultirange(numrange(5,null));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(3,null) <@ nummultirange(numrange(5,null));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(8,9) <@ nummultirange(numrange(1,5));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(3,9) <@ nummultirange(numrange(1,5));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(1,4) <@ nummultirange(numrange(1,5));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,5) <@ nummultirange(numrange(1,5));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,5) <@ nummultirange(numrange(-4,-2), numrange(1,5));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,5) <@ nummultirange(numrange(1,5), numrange(8,9));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(6,7) <@ nummultirange(numrange(1,5), numrange(8,9));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(6,7) <@ nummultirange(numrange(1,5), numrange(6,9));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,5)}' <@ '{[1,5)}'::nummultirange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,5)}' <@ '{[-4,-2), [1,5)}'::nummultirange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,5)}' <@ '{[1,5), [8,9)}'::nummultirange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[6,7)}' <@ '{[1,5), [8,9)}'::nummultirange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT '{[6,7)}' <@ '{[1,5), [6,9)}'::nummultirange;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) <@ numrange(null,null);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(null,2)) <@ numrange(null,null);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(2,null)) <@ numrange(null,null);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(null,3)) <@ numrange(null,5);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(null,8)) <@ numrange(null,5);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(8,null)) <@ numrange(5,null);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,null)) <@ numrange(5,null);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(8,9)) <@ numrange(1,5);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,9)) <@ numrange(1,5);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,4)) <@ numrange(1,5);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5)) <@ numrange(1,5);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(-4,-2), numrange(1,5)) <@ numrange(1,9);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5), numrange(8,9)) <@ numrange(1,9);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5), numrange(6,9)) <@ numrange(1,9);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,5), numrange(6,10)) <@ numrange(1,9);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT '{[1,5)}'::nummultirange <@ '{[1,9)}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[-4,-2), [1,5)}'::nummultirange <@ '{[1,9)}';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT '{[1,5), [8,9)}'::nummultirange <@ '{[1,9)}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,5), [6,9)}'::nummultirange <@ '{[1,9)}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '{[1,5), [6,10)}'::nummultirange <@ '{[1,9)}';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'empty'::numrange &< nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'empty'::numrange &< nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() &< 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) &< 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() &< nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) &< nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() &< nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(6,7) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(1,2) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,4) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,6) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(3.5,6) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(6,7)) &< numrange(3,4);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) &< numrange(3,4);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,4)) &< numrange(3,4);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,6)) &< numrange(3,4);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3.5,6)) &< numrange(3,4);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(6,7)) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,4)) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,6)) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3.5,6)) &< nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() &> 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) &> 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'empty'::numrange &> nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'empty'::numrange &> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() &> nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() &> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) &> nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> numrange(6,7);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> numrange(1,2);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> numrange(1,4);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> numrange(1,6);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> numrange(3.5,6);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(3,4) &> nummultirange(numrange(6,7));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(3,4) &> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(3,4) &> nummultirange(numrange(1,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(3,4) &> nummultirange(numrange(1,6));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(3,4) &> nummultirange(numrange(3.5,6));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> nummultirange(numrange(6,7));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> nummultirange(numrange(1,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> nummultirange(numrange(1,6));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(3,4)) &> nummultirange(numrange(3.5,6));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'empty'::numrange -|- nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'empty'::numrange -|- nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() -|- 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) -|- 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() -|- nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) -|- nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() -|- nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT numrange(1,2) -|- nummultirange(numrange(2,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT numrange(1,2) -|- nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) -|- numrange(2,4);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) -|- numrange(3,4);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) -|- nummultirange(numrange(2,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) -|- nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(5,6)) -|- nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(5,6)) -|- nummultirange(numrange(6,7));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(5,6)) -|- nummultirange(numrange(8,9));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) -|- nummultirange(numrange(2,4), numrange(6,7));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'empty'::numrange << nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(1,2) << nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(1,2) << nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(1,2) << nummultirange(numrange(0,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(1,2) << nummultirange(numrange(0,4), numrange(7,8));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange() << 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange() << numrange(1,2);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(3,4)) << numrange(3,6);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(0,2)) << numrange(3,6);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select nummultirange(numrange(0,2), numrange(7,8)) << numrange(3,6);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(-4,-2), numrange(0,2)) << numrange(3,6);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select nummultirange() << nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange() << nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(1,2)) << nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(1,2)) << nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(1,2)) << nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select nummultirange(numrange(1,2)) << nummultirange(numrange(3,4), numrange(7,8));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select nummultirange(numrange(1,2), numrange(4,5)) << nummultirange(numrange(3,4), numrange(7,8));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange() >> 'empty'::numrange;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange() >> numrange(1,2);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(3,4)) >> numrange(1,2);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select nummultirange(numrange(0,4)) >> numrange(1,2);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(0,4), numrange(7,8)) >> numrange(1,2);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'empty'::numrange >> nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(1,2) >> nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(3,6) >> nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(3,6) >> nummultirange(numrange(0,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select numrange(3,6) >> nummultirange(numrange(0,2), numrange(7,8));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select numrange(3,6) >> nummultirange(numrange(-4,-2), numrange(0,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select nummultirange() >> nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(1,2)) >> nummultirange();`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange() >> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(1,2)) >> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select nummultirange(numrange(3,4)) >> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select nummultirange(numrange(3,4), numrange(7,8)) >> nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select nummultirange(numrange(3,4), numrange(7,8)) >> nummultirange(numrange(1,2), numrange(4,5));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT nummultirange() + nummultirange();`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT nummultirange() + nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) + nummultirange();`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) + nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) + nummultirange(numrange(2,4));`,
				Results:   []sql.Row{{`{[1,4)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) + nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{`{[1,2),[3,4)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) + nummultirange(numrange(2,4));`,
				Results:   []sql.Row{{`{[1,5)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) + nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{`{[1,2),[3,5)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) + nummultirange(numrange(0,9));`,
				Results:   []sql.Row{{`{[0,9)}`}},
			},
			{
				Statement: `SELECT range_merge(nummultirange());`,
				Results:   []sql.Row{{`empty`}},
			},
			{
				Statement: `SELECT range_merge(nummultirange(numrange(1,2)));`,
				Results:   []sql.Row{{`[1,2)`}},
			},
			{
				Statement: `SELECT range_merge(nummultirange(numrange(1,2), numrange(7,8)));`,
				Results:   []sql.Row{{`[1,8)`}},
			},
			{
				Statement: `SELECT nummultirange() - nummultirange();`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT nummultirange() - nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) - nummultirange();`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(3,4)) - nummultirange();`,
				Results:   []sql.Row{{`{[1,2),[3,4)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) - nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) - nummultirange(numrange(2,4));`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) - nummultirange(numrange(3,4));`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,4)) - nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{`{[2,4)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,4)) - nummultirange(numrange(2,3));`,
				Results:   []sql.Row{{`{[1,2),[3,4)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,4)) - nummultirange(numrange(0,8));`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,4)) - nummultirange(numrange(0,2));`,
				Results:   []sql.Row{{`{[2,4)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,8)) - nummultirange(numrange(0,2), numrange(3,4));`,
				Results:   []sql.Row{{`{[2,3),[4,8)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,8)) - nummultirange(numrange(2,3), numrange(5,null));`,
				Results:   []sql.Row{{`{[1,2),[3,5)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) - nummultirange(numrange(-2,0));`,
				Results:   []sql.Row{{`{[1,2),[4,5)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) - nummultirange(numrange(2,4));`,
				Results:   []sql.Row{{`{[1,2),[4,5)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) - nummultirange(numrange(3,5));`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) - nummultirange(numrange(0,9));`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,3), numrange(4,5)) - nummultirange(numrange(2,9));`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) - nummultirange(numrange(8,9));`,
				Results:   []sql.Row{{`{[1,2),[4,5)}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2), numrange(4,5)) - nummultirange(numrange(-2,0), numrange(8,9));`,
				Results:   []sql.Row{{`{[1,2),[4,5)}`}},
			},
			{
				Statement: `SELECT nummultirange() * nummultirange();`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT nummultirange() * nummultirange(numrange(1,2));`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT nummultirange(numrange(1,2)) * nummultirange();`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT '{[1,3)}'::nummultirange * '{[1,5)}'::nummultirange;`,
				Results:   []sql.Row{{`{[1,3)}`}},
			},
			{
				Statement: `SELECT '{[1,3)}'::nummultirange * '{[0,5)}'::nummultirange;`,
				Results:   []sql.Row{{`{[1,3)}`}},
			},
			{
				Statement: `SELECT '{[1,3)}'::nummultirange * '{[0,2)}'::nummultirange;`,
				Results:   []sql.Row{{`{[1,2)}`}},
			},
			{
				Statement: `SELECT '{[1,3)}'::nummultirange * '{[2,5)}'::nummultirange;`,
				Results:   []sql.Row{{`{[2,3)}`}},
			},
			{
				Statement: `SELECT '{[1,4)}'::nummultirange * '{[2,3)}'::nummultirange;`,
				Results:   []sql.Row{{`{[2,3)}`}},
			},
			{
				Statement: `SELECT '{[1,4)}'::nummultirange * '{[0,2), [3,5)}'::nummultirange;`,
				Results:   []sql.Row{{`{[1,2),[3,4)}`}},
			},
			{
				Statement: `SELECT '{[1,4), [7,10)}'::nummultirange * '{[0,8), [9,12)}'::nummultirange;`,
				Results:   []sql.Row{{`{[1,4),[7,8),[9,10)}`}},
			},
			{
				Statement: `SELECT '{[1,4), [7,10)}'::nummultirange * '{[9,12)}'::nummultirange;`,
				Results:   []sql.Row{{`{[9,10)}`}},
			},
			{
				Statement: `SELECT '{[1,4), [7,10)}'::nummultirange * '{[-5,-4), [5,6), [9,12)}'::nummultirange;`,
				Results:   []sql.Row{{`{[9,10)}`}},
			},
			{
				Statement: `SELECT '{[1,4), [7,10)}'::nummultirange * '{[0,2), [3,8), [9,12)}'::nummultirange;`,
				Results:   []sql.Row{{`{[1,2),[3,4),[7,8),[9,10)}`}},
			},
			{
				Statement: `SELECT '{[1,4), [7,10)}'::nummultirange * '{[0,2), [3,8), [9,12)}'::nummultirange;`,
				Results:   []sql.Row{{`{[1,2),[3,4),[7,8),[9,10)}`}},
			},
			{
				Statement: `create table test_multirange_gist(mr int4multirange);`,
			},
			{
				Statement: `insert into test_multirange_gist select int4multirange(int4range(g, g+10),int4range(g+20, g+30),int4range(g+40, g+50)) from generate_series(1,2000) g;`,
			},
			{
				Statement: `insert into test_multirange_gist select '{}'::int4multirange from generate_series(1,500) g;`,
			},
			{
				Statement: `insert into test_multirange_gist select int4multirange(int4range(g, g+10000)) from generate_series(1,1000) g;`,
			},
			{
				Statement: `insert into test_multirange_gist select int4multirange(int4range(NULL, g*10, '(]'), int4range(g*10, g*20, '(]')) from generate_series(1,100) g;`,
			},
			{
				Statement: `insert into test_multirange_gist select int4multirange(int4range(g*10, g*20, '(]'), int4range(g*20, NULL, '(]')) from generate_series(1,100) g;`,
			},
			{
				Statement: `create index test_mulrirange_gist_idx on test_multirange_gist using gist (mr);`,
			},
			{
				Statement: `analyze test_multirange_gist;`,
			},
			{
				Statement: `SET enable_seqscan    = t;`,
			},
			{
				Statement: `SET enable_indexscan  = f;`,
			},
			{
				Statement: `SET enable_bitmapscan = f;`,
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr = '{}'::int4multirange;`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> 'empty'::int4range;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr && 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr <@ 'empty'::int4range;`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr << 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr >> 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &< 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &> 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr -|- 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> '{}'::int4multirange;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> '{}'::int4multirange;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr && '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr <@ '{}'::int4multirange;`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr << '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr >> '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &< '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &> '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr -|- '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr = int4multirange(int4range(10,20), int4range(30,40), int4range(50,60));`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> 10;`,
				Results:   []sql.Row{{120}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> int4range(10,20);`,
				Results:   []sql.Row{{111}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr && int4range(10,20);`,
				Results:   []sql.Row{{139}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr <@ int4range(10,50);`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr << int4range(100,500);`,
				Results:   []sql.Row{{54}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr >> int4range(100,500);`,
				Results:   []sql.Row{{2053}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &< int4range(100,500);`,
				Results:   []sql.Row{{474}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &> int4range(100,500);`,
				Results:   []sql.Row{{2893}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr -|- int4range(100,500);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> '{}'::int4multirange;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> int4multirange(int4range(10,20), int4range(30,40));`,
				Results:   []sql.Row{{110}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr && '{(10,20),(30,40),(50,60)}'::int4multirange;`,
				Results:   []sql.Row{{218}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr <@ '{(10,30),(40,60),(70,90)}'::int4multirange;`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr << int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{54}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr >> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{2053}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &< int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{474}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{2893}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr -|- int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SET enable_seqscan    = f;`,
			},
			{
				Statement: `SET enable_indexscan  = t;`,
			},
			{
				Statement: `SET enable_bitmapscan = f;`,
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr = '{}'::int4multirange;`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> 'empty'::int4range;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr && 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr <@ 'empty'::int4range;`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr << 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr >> 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &< 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &> 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr -|- 'empty'::int4range;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> '{}'::int4multirange;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> '{}'::int4multirange;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr && '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr <@ '{}'::int4multirange;`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr << '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr >> '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &< '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &> '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr -|- '{}'::int4multirange;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> 'empty'::int4range;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr = int4multirange(int4range(10,20), int4range(30,40), int4range(50,60));`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> 10;`,
				Results:   []sql.Row{{120}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> int4range(10,20);`,
				Results:   []sql.Row{{111}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr && int4range(10,20);`,
				Results:   []sql.Row{{139}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr <@ int4range(10,50);`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr << int4range(100,500);`,
				Results:   []sql.Row{{54}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr >> int4range(100,500);`,
				Results:   []sql.Row{{2053}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &< int4range(100,500);`,
				Results:   []sql.Row{{474}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &> int4range(100,500);`,
				Results:   []sql.Row{{2893}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr -|- int4range(100,500);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> '{}'::int4multirange;`,
				Results:   []sql.Row{{3700}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr @> int4multirange(int4range(10,20), int4range(30,40));`,
				Results:   []sql.Row{{110}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr && '{(10,20),(30,40),(50,60)}'::int4multirange;`,
				Results:   []sql.Row{{218}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr <@ '{(10,30),(40,60),(70,90)}'::int4multirange;`,
				Results:   []sql.Row{{500}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr << int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{54}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr >> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{2053}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &< int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{474}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr &> int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{2893}},
			},
			{
				Statement: `select count(*) from test_multirange_gist where mr -|- int4multirange(int4range(100,200), int4range(400,500));`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `drop table test_multirange_gist;`,
			},
			{
				Statement: `create table reservations ( room_id integer not null, booked_during daterange );`,
			},
			{
				Statement: `insert into reservations values
(1, daterange('2018-07-01', '2018-07-07')),
(1, daterange('2018-07-07', '2018-07-14')),
(1, daterange('2018-07-20', '2018-07-22')),
(2, daterange('2018-07-01', '2018-07-03')),
(3, NULL),
(4, NULL),
(4, NULL),
(5, NULL),
(5, daterange('2018-07-01', '2018-07-03')),
(6, daterange('2018-07-01', '2018-07-07')),
(6, daterange('2018-07-05', '2018-07-10')),
(7, daterange('2018-07-01', '2018-07-07')),
(7, daterange('2018-07-07', '2018-07-14')),
(8, 'empty'::daterange)
;`,
			},
			{
				Statement: `SELECT   room_id, range_agg(booked_during)
FROM     reservations
GROUP BY room_id
ORDER BY room_id;`,
				Results: []sql.Row{{1, `{[07-01-2018,07-14-2018),[07-20-2018,07-22-2018)}`}, {2, `{[07-01-2018,07-03-2018)}`}, {3, ``}, {4, ``}, {5, `{[07-01-2018,07-03-2018)}`}, {6, `{[07-01-2018,07-10-2018)}`}, {7, `{[07-01-2018,07-14-2018)}`}, {8, `{}`}},
			},
			{
				Statement: `SELECT  range_agg(r)
FROM    (VALUES
          ('[a,c]'::textrange),
          ('[b,b]'::textrange),
          ('[c,f]'::textrange),
          ('[g,h)'::textrange),
          ('[h,j)'::textrange)
        ) t(r);`,
				Results: []sql.Row{{`{[a,f],[g,j)}`}},
			},
			{
				Statement: `select range_agg(nmr) from nummultirange_test;`,
				Results:   []sql.Row{{`{(,)}`}},
			},
			{
				Statement: `select range_agg(nmr) from nummultirange_test where false;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select range_agg(null::nummultirange) from nummultirange_test;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select range_agg(nmr) from (values ('{}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select range_agg(nmr) from (values ('{}'::nummultirange), ('{}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select range_agg(nmr) from (values ('{[1,2]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[1,2]}`}},
			},
			{
				Statement: `select range_agg(nmr) from (values ('{[1,2], [5,6]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[1,2],[5,6]}`}},
			},
			{
				Statement: `select range_agg(nmr) from (values ('{[1,2], [2,3]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[1,3]}`}},
			},
			{
				Statement: `select range_agg(nmr) from (values ('{[1,2]}'::nummultirange), ('{[5,6]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[1,2],[5,6]}`}},
			},
			{
				Statement: `select range_agg(nmr) from (values ('{[1,2]}'::nummultirange), ('{[2,3]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[1,3]}`}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from nummultirange_test;`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from nummultirange_test where false;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select range_intersect_agg(null::nummultirange) from nummultirange_test;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from (values ('{[1,3]}'::nummultirange), ('{[6,12]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from (values ('{[1,6]}'::nummultirange), ('{[3,12]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[3,6]}`}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from (values ('{[1,6], [10,12]}'::nummultirange), ('{[4,14]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[4,6],[10,12]}`}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from (values ('{}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from (values ('{[1,2]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[1,2]}`}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from (values ('{[1,6], [10,12]}'::nummultirange)) t(nmr);`,
				Results:   []sql.Row{{`{[1,6],[10,12]}`}},
			},
			{
				Statement: `select range_intersect_agg(nmr) from nummultirange_test where nmr @> 4.0;`,
				Results:   []sql.Row{{`{[3,5)}`}},
			},
			{
				Statement: `create table nummultirange_test2(nmr nummultirange);`,
			},
			{
				Statement: `create index nummultirange_test2_hash_idx on nummultirange_test2 using hash (nmr);`,
			},
			{
				Statement: `INSERT INTO nummultirange_test2 VALUES('{[, 5)}');`,
			},
			{
				Statement: `INSERT INTO nummultirange_test2 VALUES(nummultirange(numrange(1.1, 2.2)));`,
			},
			{
				Statement: `INSERT INTO nummultirange_test2 VALUES(nummultirange(numrange(1.1, 2.2)));`,
			},
			{
				Statement: `INSERT INTO nummultirange_test2 VALUES(nummultirange(numrange(1.1, 2.2,'()')));`,
			},
			{
				Statement: `INSERT INTO nummultirange_test2 VALUES('{}');`,
			},
			{
				Statement: `select * from nummultirange_test2 where nmr = '{}';`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select * from nummultirange_test2 where nmr = nummultirange(numrange(1.1, 2.2));`,
				Results:   []sql.Row{{`{[1.1,2.2)}`}, {`{[1.1,2.2)}`}},
			},
			{
				Statement: `select * from nummultirange_test2 where nmr = nummultirange(numrange(1.1, 2.3));`,
				Results:   []sql.Row{},
			},
			{
				Statement: `set enable_nestloop=t;`,
			},
			{
				Statement: `set enable_hashjoin=f;`,
			},
			{
				Statement: `set enable_mergejoin=f;`,
			},
			{
				Statement: `select * from nummultirange_test natural join nummultirange_test2 order by nmr;`,
				Results:   []sql.Row{{`{}`}, {`{}`}, {`{}`}, {`{}`}, {`{(,5)}`}, {`{[1.1,2.2)}`}, {`{[1.1,2.2)}`}},
			},
			{
				Statement: `set enable_nestloop=f;`,
			},
			{
				Statement: `set enable_hashjoin=t;`,
			},
			{
				Statement: `set enable_mergejoin=f;`,
			},
			{
				Statement: `select * from nummultirange_test natural join nummultirange_test2 order by nmr;`,
				Results:   []sql.Row{{`{}`}, {`{}`}, {`{}`}, {`{}`}, {`{(,5)}`}, {`{[1.1,2.2)}`}, {`{[1.1,2.2)}`}},
			},
			{
				Statement: `set enable_nestloop=f;`,
			},
			{
				Statement: `set enable_hashjoin=f;`,
			},
			{
				Statement: `set enable_mergejoin=t;`,
			},
			{
				Statement: `select * from nummultirange_test natural join nummultirange_test2 order by nmr;`,
				Results:   []sql.Row{{`{}`}, {`{}`}, {`{}`}, {`{}`}, {`{(,5)}`}, {`{[1.1,2.2)}`}, {`{[1.1,2.2)}`}},
			},
			{
				Statement: `set enable_nestloop to default;`,
			},
			{
				Statement: `set enable_hashjoin to default;`,
			},
			{
				Statement: `set enable_mergejoin to default;`,
			},
			{
				Statement: `DROP TABLE nummultirange_test2;`,
			},
			{
				Statement: `select '{[123.001, 5.e9)}'::float8multirange @> 888.882::float8;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create table float8multirange_test(f8mr float8multirange, i int);`,
			},
			{
				Statement: `insert into float8multirange_test values(float8multirange(float8range(-100.00007, '1.111113e9')), 42);`,
			},
			{
				Statement: `select * from float8multirange_test;`,
				Results:   []sql.Row{{`{[-100.00007,1111113000)}`, 42}},
			},
			{
				Statement: `drop table float8multirange_test;`,
			},
			{
				Statement: `create domain mydomain as int4;`,
			},
			{
				Statement: `create type mydomainrange as range(subtype=mydomain);`,
			},
			{
				Statement: `select '{[4,50)}'::mydomainmultirange @> 7::mydomain;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `drop domain mydomain cascade;`,
			},
			{
				Statement: `create domain restrictedmultirange as int4multirange check (upper(value) < 10);`,
			},
			{
				Statement: `select '{[4,5)}'::restrictedmultirange @> 7;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `select '{[4,50)}'::restrictedmultirange @> 7; -- should fail`,
				ErrorString: `value for domain restrictedmultirange violates check constraint "restrictedmultirange_check"`,
			},
			{
				Statement: `drop domain restrictedmultirange;`,
			},
			{
				Statement: `---
---
create type intr as range(subtype=int);`,
			},
			{
				Statement: `select intr_multirange(intr(1,10));`,
				Results:   []sql.Row{{`{[1,10)}`}},
			},
			{
				Statement: `drop type intr;`,
			},
			{
				Statement: `create type intmultirange as (x int, y int);`,
			},
			{
				Statement:   `create type intrange as range(subtype=int); -- should fail`,
				ErrorString: `type "intmultirange" already exists`,
			},
			{
				Statement: `drop type intmultirange;`,
			},
			{
				Statement: `create type intr_multirange as (x int, y int);`,
			},
			{
				Statement:   `create type intr as range(subtype=int); -- should fail`,
				ErrorString: `type "intr_multirange" already exists`,
			},
			{
				Statement: `drop type intr_multirange;`,
			},
			{
				Statement:   `create type textrange1 as range(subtype=text, multirange_type_name=int, collation="C");`,
				ErrorString: `type "int4" already exists`,
			},
			{
				Statement: `create type textrange1 as range(subtype=text, multirange_type_name=multirange_of_text, collation="C");`,
			},
			{
				Statement: `create type textrange2 as range(subtype=text, multirange_type_name=_textrange1, collation="C");`,
			},
			{
				Statement:   `select multirange_of_text(textrange2('a','Z'));  -- should fail`,
				ErrorString: `function multirange_of_text(textrange2) does not exist`,
			},
			{
				Statement:   `select multirange_of_text(textrange1('a','Z')) @> 'b'::text;`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `select unnest(multirange_of_text(textrange1('a','b'), textrange1('d','e')));`,
				Results:   []sql.Row{{`[a,b)`}, {`[d,e)`}},
			},
			{
				Statement: `select _textrange1(textrange2('a','z')) @> 'b'::text;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `drop type textrange1;`,
			},
			{
				Statement: `drop type textrange2;`,
			},
			{
				Statement: `create function anyarray_anymultirange_func(a anyarray, r anymultirange)
  returns anyelement as 'select $1[1] + lower($2);' language sql;`,
			},
			{
				Statement: `select anyarray_anymultirange_func(ARRAY[1,2], int4multirange(int4range(10,20)));`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement:   `select anyarray_anymultirange_func(ARRAY[1,2], nummultirange(numrange(10,20)));`,
				ErrorString: `function anyarray_anymultirange_func(integer[], nummultirange) does not exist`,
			},
			{
				Statement: `drop function anyarray_anymultirange_func(anyarray, anymultirange);`,
			},
			{
				Statement: `create function bogus_func(anyelement)
  returns anymultirange as 'select int4multirange(int4range(1,10))' language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function bogus_func(int)
  returns anymultirange as 'select int4multirange(int4range(1,10))' language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function range_add_bounds(anymultirange)
  returns anyelement as 'select lower($1) + upper($1)' language sql;`,
			},
			{
				Statement: `select range_add_bounds(int4multirange(int4range(1, 17)));`,
				Results:   []sql.Row{{18}},
			},
			{
				Statement: `select range_add_bounds(nummultirange(numrange(1.0001, 123.123)));`,
				Results:   []sql.Row{{124.1231}},
			},
			{
				Statement: `create function multirangetypes_sql(q anymultirange, b anyarray, out c anyelement)
  as $$ select upper($1) + $2[1] $$
  language sql;`,
			},
			{
				Statement: `select multirangetypes_sql(int4multirange(int4range(1,10)), ARRAY[2,20]);`,
				Results:   []sql.Row{{12}},
			},
			{
				Statement:   `select multirangetypes_sql(nummultirange(numrange(1,10)), ARRAY[2,20]);  -- match failure`,
				ErrorString: `function multirangetypes_sql(nummultirange, integer[]) does not exist`,
			},
			{
				Statement: `create function anycompatiblearray_anycompatiblemultirange_func(a anycompatiblearray, mr anycompatiblemultirange)
  returns anycompatible as 'select $1[1] + lower($2);' language sql;`,
			},
			{
				Statement: `select anycompatiblearray_anycompatiblemultirange_func(ARRAY[1,2], multirange(int4range(10,20)));`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `select anycompatiblearray_anycompatiblemultirange_func(ARRAY[1,2], multirange(numrange(10,20)));`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement:   `select anycompatiblearray_anycompatiblemultirange_func(ARRAY[1.1,2], multirange(int4range(10,20)));`,
				ErrorString: `function anycompatiblearray_anycompatiblemultirange_func(numeric[], int4multirange) does not exist`,
			},
			{
				Statement: `drop function anycompatiblearray_anycompatiblemultirange_func(anycompatiblearray, anycompatiblemultirange);`,
			},
			{
				Statement: `create function anycompatiblerange_anycompatiblemultirange_func(r anycompatiblerange, mr anycompatiblemultirange)
  returns anycompatible as 'select lower($1) + lower($2);' language sql;`,
			},
			{
				Statement: `select anycompatiblerange_anycompatiblemultirange_func(int4range(1,2), multirange(int4range(10,20)));`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement:   `select anycompatiblerange_anycompatiblemultirange_func(numrange(1,2), multirange(int4range(10,20)));`,
				ErrorString: `function anycompatiblerange_anycompatiblemultirange_func(numrange, int4multirange) does not exist`,
			},
			{
				Statement: `drop function anycompatiblerange_anycompatiblemultirange_func(anycompatiblerange, anycompatiblemultirange);`,
			},
			{
				Statement: `create function bogus_func(anycompatible)
  returns anycompatiblerange as 'select int4range(1,10)' language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `select ARRAY[nummultirange(numrange(1.1, 1.2)), nummultirange(numrange(12.3, 155.5))];`,
				Results:   []sql.Row{{`{"{[1.1,1.2)}","{[12.3,155.5)}"}`}},
			},
			{
				Statement: `create table i8mr_array (f1 int, f2 int8multirange[]);`,
			},
			{
				Statement: `insert into i8mr_array values (42, array[int8multirange(int8range(1,10)), int8multirange(int8range(2,20))]);`,
			},
			{
				Statement: `select * from i8mr_array;`,
				Results:   []sql.Row{{42, `{"{[1,10)}","{[2,20)}"}`}},
			},
			{
				Statement: `drop table i8mr_array;`,
			},
			{
				Statement: `select arraymultirange(arrayrange(ARRAY[1,2], ARRAY[2,1]));`,
				Results:   []sql.Row{{`{["{1,2}","{2,1}")}`}},
			},
			{
				Statement:   `select arraymultirange(arrayrange(ARRAY[2,1], ARRAY[1,2]));  -- fail`,
				ErrorString: `range lower bound must be less than or equal to range upper bound`,
			},
			{
				Statement: `select array[1,1] <@ arraymultirange(arrayrange(array[1,2], array[2,1]));`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select array[1,3] <@ arraymultirange(arrayrange(array[1,2], array[2,1]));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create type two_ints as (a int, b int);`,
			},
			{
				Statement: `create type two_ints_range as range (subtype = two_ints);`,
			},
			{
				Statement: `select *, row_to_json(upper(t)) as u from
  (values (two_ints_multirange(two_ints_range(row(1,2), row(3,4)))),
          (two_ints_multirange(two_ints_range(row(5,6), row(7,8))))) v(t);`,
				Results: []sql.Row{{`{["(1,2)","(3,4)")}`, `{"a":3,"b":4}`}, {`{["(5,6)","(7,8)")}`, `{"a":7,"b":8}`}},
			},
			{
				Statement: `drop type two_ints cascade;`,
			},
			{
				Statement: `set enable_sort = off;  -- try to make it pick a hash setop implementation`,
			},
			{
				Statement: `select '{(2,5)}'::cashmultirange except select '{(5,6)}'::cashmultirange;`,
				Results:   []sql.Row{{`{($2.00,$5.00)}`}},
			},
			{
				Statement: `reset enable_sort;`,
			},
			{
				Statement: `create function mr_outparam_succeed(i anymultirange, out r anymultirange, out t text)
  as $$ select $1, 'foo'::text $$ language sql;`,
			},
			{
				Statement: `select * from mr_outparam_succeed(int4multirange(int4range(1,2)));`,
				Results:   []sql.Row{{`{[1,2)}`, `foo`}},
			},
			{
				Statement: `create function mr_outparam_succeed2(i anymultirange, out r anyarray, out t text)
  as $$ select ARRAY[upper($1)], 'foo'::text $$ language sql;`,
			},
			{
				Statement: `select * from mr_outparam_succeed2(int4multirange(int4range(1,2)));`,
				Results:   []sql.Row{{`{2}`, `foo`}},
			},
			{
				Statement: `create function mr_outparam_succeed3(i anymultirange, out r anyrange, out t text)
  as $$ select range_merge($1), 'foo'::text $$ language sql;`,
			},
			{
				Statement: `select * from mr_outparam_succeed3(int4multirange(int4range(1,2)));`,
				Results:   []sql.Row{{`[1,2)`, `foo`}},
			},
			{
				Statement: `create function mr_outparam_succeed4(i anyrange, out r anymultirange, out t text)
  as $$ select multirange($1), 'foo'::text $$ language sql;`,
			},
			{
				Statement: `select * from mr_outparam_succeed4(int4range(1,2));`,
				Results:   []sql.Row{{`{[1,2)}`, `foo`}},
			},
			{
				Statement: `create function mr_inoutparam_succeed(out i anyelement, inout r anymultirange)
  as $$ select upper($1), $1 $$ language sql;`,
			},
			{
				Statement: `select * from mr_inoutparam_succeed(int4multirange(int4range(1,2)));`,
				Results:   []sql.Row{{2, `{[1,2)}`}},
			},
			{
				Statement: `create function mr_table_succeed(i anyelement, r anymultirange) returns table(i anyelement, r anymultirange)
  as $$ select $1, $2 $$ language sql;`,
			},
			{
				Statement: `select * from mr_table_succeed(123, int4multirange(int4range(1,11)));`,
				Results:   []sql.Row{{123, `{[1,11)}`}},
			},
			{
				Statement: `create function mr_polymorphic(i anyrange) returns anymultirange
  as $$ begin return multirange($1); end; $$ language plpgsql;`,
			},
			{
				Statement: `select mr_polymorphic(int4range(1, 4));`,
				Results:   []sql.Row{{`{[1,4)}`}},
			},
			{
				Statement: `create function mr_outparam_fail(i anyelement, out r anymultirange, out t text)
  as $$ select '[1,10]', 'foo' $$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function mr_inoutparam_fail(inout i anyelement, out r anymultirange)
  as $$ select $1, '[1,10]' $$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function mr_table_fail(i anyelement) returns table(i anyelement, r anymultirange)
  as $$ select $1, '[1,10]' $$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
		},
	})
}
