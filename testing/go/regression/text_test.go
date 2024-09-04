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

func TestText(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_text)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_text,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT text 'this is a text string' = text 'this is a text string' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT text 'this is a text string' = text 'this is a text strin' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT * FROM TEXT_TBL;`,
				Results:   []sql.Row{{`doh!`}, {`hi de ho neighbor`}},
			},
			{
				Statement:   `select length(42);`,
				ErrorString: `function length(integer) does not exist`,
			},
			{
				Statement: `select 'four: '::text || 2+2;`,
				Results:   []sql.Row{{`four: 4`}},
			},
			{
				Statement: `select 'four: ' || 2+2;`,
				Results:   []sql.Row{{`four: 4`}},
			},
			{
				Statement:   `select 3 || 4.0;`,
				ErrorString: `operator does not exist: integer || numeric`,
			},
			{
				Statement: `/*
 * various string functions
 */
select concat('one');`,
				Results: []sql.Row{{`one`}},
			},
			{
				Statement: `select concat(1,2,3,'hello',true, false, to_date('20100309','YYYYMMDD'));`,
				Results:   []sql.Row{{`123hellotf03-09-2010`}},
			},
			{
				Statement: `select concat_ws('#','one');`,
				Results:   []sql.Row{{`one`}},
			},
			{
				Statement: `select concat_ws('#',1,2,3,'hello',true, false, to_date('20100309','YYYYMMDD'));`,
				Results:   []sql.Row{{`1#2#3#hello#t#f#03-09-2010`}},
			},
			{
				Statement: `select concat_ws(',',10,20,null,30);`,
				Results:   []sql.Row{{`10,20,30`}},
			},
			{
				Statement: `select concat_ws('',10,20,null,30);`,
				Results:   []sql.Row{{102030}},
			},
			{
				Statement: `select concat_ws(NULL,10,20,null,30) is null;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select reverse('abcde');`,
				Results:   []sql.Row{{`edcba`}},
			},
			{
				Statement: `select i, left('ahoj', i), right('ahoj', i) from generate_series(-5, 5) t(i) order by i;`,
				Results:   []sql.Row{{-5, ``, ``}, {-4, ``, ``}, {-3, `a`, `j`}, {-2, `ah`, `oj`}, {-1, `aho`, `hoj`}, {0, ``, ``}, {1, `a`, `j`}, {2, `ah`, `oj`}, {3, `aho`, `hoj`}, {4, `ahoj`, `ahoj`}, {5, `ahoj`, `ahoj`}},
			},
			{
				Statement: `select quote_literal('');`,
				Results:   []sql.Row{{`''`}},
			},
			{
				Statement: `select quote_literal('abc''');`,
				Results:   []sql.Row{{`'abc'''`}},
			},
			{
				Statement: `select quote_literal(e'\\');`,
				Results:   []sql.Row{{`E'\\'`}},
			},
			{
				Statement: `select concat(variadic array[1,2,3]);`,
				Results:   []sql.Row{{123}},
			},
			{
				Statement: `select concat_ws(',', variadic array[1,2,3]);`,
				Results:   []sql.Row{{`1,2,3`}},
			},
			{
				Statement: `select concat_ws(',', variadic NULL::int[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select concat(variadic NULL::int[]) is NULL;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select concat(variadic '{}'::int[]) = '';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select concat_ws(',', variadic 10);`,
				ErrorString: `VARIADIC argument must be an array`,
			},
			{
				Statement: `/*
 * format
 */
select format(NULL);`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `select format('Hello');`,
				Results:   []sql.Row{{`Hello`}},
			},
			{
				Statement: `select format('Hello %s', 'World');`,
				Results:   []sql.Row{{`Hello World`}},
			},
			{
				Statement: `select format('Hello %%');`,
				Results:   []sql.Row{{`Hello %`}},
			},
			{
				Statement: `select format('Hello %%%%');`,
				Results:   []sql.Row{{`Hello %%`}},
			},
			{
				Statement:   `select format('Hello %s %s', 'World');`,
				ErrorString: `too few arguments for format()`,
			},
			{
				Statement:   `select format('Hello %s');`,
				ErrorString: `too few arguments for format()`,
			},
			{
				Statement:   `select format('Hello %x', 20);`,
				ErrorString: `unrecognized format() type specifier "x"`,
			},
			{
				Statement: `select format('INSERT INTO %I VALUES(%L,%L)', 'mytab', 10, 'Hello');`,
				Results:   []sql.Row{{`INSERT INTO mytab VALUES('10','Hello')`}},
			},
			{
				Statement: `select format('%s%s%s','Hello', NULL,'World');`,
				Results:   []sql.Row{{`HelloWorld`}},
			},
			{
				Statement: `select format('INSERT INTO %I VALUES(%L,%L)', 'mytab', 10, NULL);`,
				Results:   []sql.Row{{`INSERT INTO mytab VALUES('10',NULL)`}},
			},
			{
				Statement: `select format('INSERT INTO %I VALUES(%L,%L)', 'mytab', NULL, 'Hello');`,
				Results:   []sql.Row{{`INSERT INTO mytab VALUES(NULL,'Hello')`}},
			},
			{
				Statement:   `select format('INSERT INTO %I VALUES(%L,%L)', NULL, 10, 'Hello');`,
				ErrorString: `null values cannot be formatted as an SQL identifier`,
			},
			{
				Statement: `select format('%1$s %3$s', 1, 2, 3);`,
				Results:   []sql.Row{{`1 3`}},
			},
			{
				Statement: `select format('%1$s %12$s', 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12);`,
				Results:   []sql.Row{{`1 12`}},
			},
			{
				Statement:   `select format('%1$s %4$s', 1, 2, 3);`,
				ErrorString: `too few arguments for format()`,
			},
			{
				Statement:   `select format('%1$s %13$s', 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12);`,
				ErrorString: `too few arguments for format()`,
			},
			{
				Statement:   `select format('%0$s', 'Hello');`,
				ErrorString: `format specifies argument 0, but arguments are numbered from 1`,
			},
			{
				Statement:   `select format('%*0$s', 'Hello');`,
				ErrorString: `format specifies argument 0, but arguments are numbered from 1`,
			},
			{
				Statement:   `select format('%1$', 1);`,
				ErrorString: `unterminated format() type specifier`,
			},
			{
				Statement:   `select format('%1$1', 1);`,
				ErrorString: `unterminated format() type specifier`,
			},
			{
				Statement: `select format('Hello %s %1$s %s', 'World', 'Hello again');`,
				Results:   []sql.Row{{`Hello World World Hello again`}},
			},
			{
				Statement: `select format('Hello %s %s, %2$s %2$s', 'World', 'Hello again');`,
				Results:   []sql.Row{{`Hello World Hello again, Hello again Hello again`}},
			},
			{
				Statement: `select format('%s, %s', variadic array['Hello','World']);`,
				Results:   []sql.Row{{`Hello, World`}},
			},
			{
				Statement: `select format('%s, %s', variadic array[1, 2]);`,
				Results:   []sql.Row{{`1, 2`}},
			},
			{
				Statement: `select format('%s, %s', variadic array[true, false]);`,
				Results:   []sql.Row{{`t, f`}},
			},
			{
				Statement: `select format('%s, %s', variadic array[true, false]::text[]);`,
				Results:   []sql.Row{{`true, false`}},
			},
			{
				Statement: `select format('%2$s, %1$s', variadic array['first', 'second']);`,
				Results:   []sql.Row{{`second, first`}},
			},
			{
				Statement: `select format('%2$s, %1$s', variadic array[1, 2]);`,
				Results:   []sql.Row{{`2, 1`}},
			},
			{
				Statement: `select format('Hello', variadic NULL::int[]);`,
				Results:   []sql.Row{{`Hello`}},
			},
			{
				Statement: `select format(string_agg('%s',','), variadic array_agg(i))
from generate_series(1,200) g(i);`,
				Results: []sql.Row{{`1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100,101,102,103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119,120,121,122,123,124,125,126,127,128,129,130,131,132,133,134,135,136,137,138,139,140,141,142,143,144,145,146,147,148,149,150,151,152,153,154,155,156,157,158,159,160,161,162,163,164,165,166,167,168,169,170,171,172,173,174,175,176,177,178,179,180,181,182,183,184,185,186,187,188,189,190,191,192,193,194,195,196,197,198,199,200`}},
			},
			{
				Statement: `select format('>>%10s<<', 'Hello');`,
				Results:   []sql.Row{{`>>     Hello<<`}},
			},
			{
				Statement: `select format('>>%10s<<', NULL);`,
				Results:   []sql.Row{{`>>          <<`}},
			},
			{
				Statement: `select format('>>%10s<<', '');`,
				Results:   []sql.Row{{`>>          <<`}},
			},
			{
				Statement: `select format('>>%-10s<<', '');`,
				Results:   []sql.Row{{`>>          <<`}},
			},
			{
				Statement: `select format('>>%-10s<<', 'Hello');`,
				Results:   []sql.Row{{`>>Hello     <<`}},
			},
			{
				Statement: `select format('>>%-10s<<', NULL);`,
				Results:   []sql.Row{{`>>          <<`}},
			},
			{
				Statement: `select format('>>%1$10s<<', 'Hello');`,
				Results:   []sql.Row{{`>>     Hello<<`}},
			},
			{
				Statement: `select format('>>%1$-10I<<', 'Hello');`,
				Results:   []sql.Row{{`>>"Hello"   <<`}},
			},
			{
				Statement: `select format('>>%2$*1$L<<', 10, 'Hello');`,
				Results:   []sql.Row{{`>>   'Hello'<<`}},
			},
			{
				Statement: `select format('>>%2$*1$L<<', 10, NULL);`,
				Results:   []sql.Row{{`>>      NULL<<`}},
			},
			{
				Statement: `select format('>>%2$*1$L<<', -10, NULL);`,
				Results:   []sql.Row{{`>>NULL      <<`}},
			},
			{
				Statement: `select format('>>%*s<<', 10, 'Hello');`,
				Results:   []sql.Row{{`>>     Hello<<`}},
			},
			{
				Statement: `select format('>>%*1$s<<', 10, 'Hello');`,
				Results:   []sql.Row{{`>>     Hello<<`}},
			},
			{
				Statement: `select format('>>%-s<<', 'Hello');`,
				Results:   []sql.Row{{`>>Hello<<`}},
			},
			{
				Statement: `select format('>>%10L<<', NULL);`,
				Results:   []sql.Row{{`>>      NULL<<`}},
			},
			{
				Statement: `select format('>>%2$*1$L<<', NULL, 'Hello');`,
				Results:   []sql.Row{{`>>'Hello'<<`}},
			},
			{
				Statement: `select format('>>%2$*1$L<<', 0, 'Hello');`,
				Results:   []sql.Row{{`>>'Hello'<<`}},
			},
		},
	})
}
