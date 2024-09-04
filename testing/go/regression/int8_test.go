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

func TestInt8(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_int8)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_int8,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `INSERT INTO INT8_TBL(q1) VALUES ('      ');`,
				ErrorString: `invalid input syntax for type bigint: "      "`,
			},
			{
				Statement:   `INSERT INTO INT8_TBL(q1) VALUES ('xxx');`,
				ErrorString: `invalid input syntax for type bigint: "xxx"`,
			},
			{
				Statement:   `INSERT INTO INT8_TBL(q1) VALUES ('3908203590239580293850293850329485');`,
				ErrorString: `value "3908203590239580293850293850329485" is out of range for type bigint`,
			},
			{
				Statement:   `INSERT INTO INT8_TBL(q1) VALUES ('-1204982019841029840928340329840934');`,
				ErrorString: `value "-1204982019841029840928340329840934" is out of range for type bigint`,
			},
			{
				Statement:   `INSERT INTO INT8_TBL(q1) VALUES ('- 123');`,
				ErrorString: `invalid input syntax for type bigint: "- 123"`,
			},
			{
				Statement:   `INSERT INTO INT8_TBL(q1) VALUES ('  345     5');`,
				ErrorString: `invalid input syntax for type bigint: "  345     5"`,
			},
			{
				Statement:   `INSERT INTO INT8_TBL(q1) VALUES ('');`,
				ErrorString: `invalid input syntax for type bigint: ""`,
			},
			{
				Statement: `SELECT * FROM INT8_TBL;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 = 4567890123456789;`,
				Results:   []sql.Row{{123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 <> 4567890123456789;`,
				Results:   []sql.Row{{123, 456}, {4567890123456789, 123}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 < 4567890123456789;`,
				Results:   []sql.Row{{123, 456}, {4567890123456789, 123}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 > 4567890123456789;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 <= 4567890123456789;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 >= 4567890123456789;`,
				Results:   []sql.Row{{123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 = 456;`,
				Results:   []sql.Row{{123, 456}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 <> 456;`,
				Results:   []sql.Row{{123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 < 456;`,
				Results:   []sql.Row{{4567890123456789, 123}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 > 456;`,
				Results:   []sql.Row{{123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 <= 456;`,
				Results:   []sql.Row{{123, 456}, {4567890123456789, 123}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 >= 456;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE 123 = q1;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE 123 <> q1;`,
				Results:   []sql.Row{{4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE 123 < q1;`,
				Results:   []sql.Row{{4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE 123 > q1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE 123 <= q1;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE 123 >= q1;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 = '456'::int2;`,
				Results:   []sql.Row{{123, 456}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 <> '456'::int2;`,
				Results:   []sql.Row{{123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 < '456'::int2;`,
				Results:   []sql.Row{{4567890123456789, 123}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 > '456'::int2;`,
				Results:   []sql.Row{{123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 <= '456'::int2;`,
				Results:   []sql.Row{{123, 456}, {4567890123456789, 123}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE q2 >= '456'::int2;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE '123'::int2 = q1;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE '123'::int2 <> q1;`,
				Results:   []sql.Row{{4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE '123'::int2 < q1;`,
				Results:   []sql.Row{{4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE '123'::int2 > q1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE '123'::int2 <= q1;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT * FROM INT8_TBL WHERE '123'::int2 >= q1;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `SELECT q1 AS plus, -q1 AS minus FROM INT8_TBL;`,
				Results:   []sql.Row{{123, -123}, {123, -123}, {4567890123456789, -4567890123456789}, {4567890123456789, -4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `SELECT q1, q2, q1 + q2 AS plus FROM INT8_TBL;`,
				Results:   []sql.Row{{123, 456, 579}, {123, 4567890123456789, 4567890123456912}, {4567890123456789, 123, 4567890123456912}, {4567890123456789, 4567890123456789, 9135780246913578}, {4567890123456789, -4567890123456789, 0}},
			},
			{
				Statement: `SELECT q1, q2, q1 - q2 AS minus FROM INT8_TBL;`,
				Results:   []sql.Row{{123, 456, -333}, {123, 4567890123456789, -4567890123456666}, {4567890123456789, 123, 4567890123456666}, {4567890123456789, 4567890123456789, 0}, {4567890123456789, -4567890123456789, 9135780246913578}},
			},
			{
				Statement:   `SELECT q1, q2, q1 * q2 AS multiply FROM INT8_TBL;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT q1, q2, q1 * q2 AS multiply FROM INT8_TBL
 WHERE q1 < 1000 or (q2 > 0 and q2 < 1000);`,
				Results: []sql.Row{{123, 456, 56088}, {123, 4567890123456789, 561850485185185047}, {4567890123456789, 123, 561850485185185047}},
			},
			{
				Statement: `SELECT q1, q2, q1 / q2 AS divide, q1 % q2 AS mod FROM INT8_TBL;`,
				Results:   []sql.Row{{123, 456, 0, 123}, {123, 4567890123456789, 0, 123}, {4567890123456789, 123, 37137318076884, 57}, {4567890123456789, 4567890123456789, 1, 0}, {4567890123456789, -4567890123456789, -1, 0}},
			},
			{
				Statement: `SELECT q1, float8(q1) FROM INT8_TBL;`,
				Results:   []sql.Row{{123, 123}, {123, 123}, {4567890123456789, 4.567890123456789e+15}, {4567890123456789, 4.567890123456789e+15}, {4567890123456789, 4.567890123456789e+15}},
			},
			{
				Statement: `SELECT q2, float8(q2) FROM INT8_TBL;`,
				Results:   []sql.Row{{456, 456}, {4567890123456789, 4.567890123456789e+15}, {123, 123}, {4567890123456789, 4.567890123456789e+15}, {-4567890123456789, -4.567890123456789e+15}},
			},
			{
				Statement: `SELECT 37 + q1 AS plus4 FROM INT8_TBL;`,
				Results:   []sql.Row{{160}, {160}, {4567890123456826}, {4567890123456826}, {4567890123456826}},
			},
			{
				Statement: `SELECT 37 - q1 AS minus4 FROM INT8_TBL;`,
				Results:   []sql.Row{{-86}, {-86}, {-4567890123456752}, {-4567890123456752}, {-4567890123456752}},
			},
			{
				Statement: `SELECT 2 * q1 AS "twice int4" FROM INT8_TBL;`,
				Results:   []sql.Row{{246}, {246}, {9135780246913578}, {9135780246913578}, {9135780246913578}},
			},
			{
				Statement: `SELECT q1 * 2 AS "twice int4" FROM INT8_TBL;`,
				Results:   []sql.Row{{246}, {246}, {9135780246913578}, {9135780246913578}, {9135780246913578}},
			},
			{
				Statement: `SELECT q1 + 42::int4 AS "8plus4", q1 - 42::int4 AS "8minus4", q1 * 42::int4 AS "8mul4", q1 / 42::int4 AS "8div4" FROM INT8_TBL;`,
				Results:   []sql.Row{{165, 81, 5166, 2}, {165, 81, 5166, 2}, {4567890123456831, 4567890123456747, 191851385185185138, 108759288653733}, {4567890123456831, 4567890123456747, 191851385185185138, 108759288653733}, {4567890123456831, 4567890123456747, 191851385185185138, 108759288653733}},
			},
			{
				Statement: `SELECT 246::int4 + q1 AS "4plus8", 246::int4 - q1 AS "4minus8", 246::int4 * q1 AS "4mul8", 246::int4 / q1 AS "4div8" FROM INT8_TBL;`,
				Results:   []sql.Row{{369, 123, 30258, 2}, {369, 123, 30258, 2}, {4567890123457035, -4567890123456543, 1123700970370370094, 0}, {4567890123457035, -4567890123456543, 1123700970370370094, 0}, {4567890123457035, -4567890123456543, 1123700970370370094, 0}},
			},
			{
				Statement: `SELECT q1 + 42::int2 AS "8plus2", q1 - 42::int2 AS "8minus2", q1 * 42::int2 AS "8mul2", q1 / 42::int2 AS "8div2" FROM INT8_TBL;`,
				Results:   []sql.Row{{165, 81, 5166, 2}, {165, 81, 5166, 2}, {4567890123456831, 4567890123456747, 191851385185185138, 108759288653733}, {4567890123456831, 4567890123456747, 191851385185185138, 108759288653733}, {4567890123456831, 4567890123456747, 191851385185185138, 108759288653733}},
			},
			{
				Statement: `SELECT 246::int2 + q1 AS "2plus8", 246::int2 - q1 AS "2minus8", 246::int2 * q1 AS "2mul8", 246::int2 / q1 AS "2div8" FROM INT8_TBL;`,
				Results:   []sql.Row{{369, 123, 30258, 2}, {369, 123, 30258, 2}, {4567890123457035, -4567890123456543, 1123700970370370094, 0}, {4567890123457035, -4567890123456543, 1123700970370370094, 0}, {4567890123457035, -4567890123456543, 1123700970370370094, 0}},
			},
			{
				Statement: `SELECT q2, abs(q2) FROM INT8_TBL;`,
				Results:   []sql.Row{{456, 456}, {4567890123456789, 4567890123456789}, {123, 123}, {4567890123456789, 4567890123456789}, {-4567890123456789, 4567890123456789}},
			},
			{
				Statement: `SELECT min(q1), min(q2) FROM INT8_TBL;`,
				Results:   []sql.Row{{123, -4567890123456789}},
			},
			{
				Statement: `SELECT max(q1), max(q2) FROM INT8_TBL;`,
				Results:   []sql.Row{{4567890123456789, 4567890123456789}},
			},
			{
				Statement: `SELECT to_char(q1, '9G999G999G999G999G999'), to_char(q2, '9,999,999,999,999,999')
	FROM INT8_TBL;`,
				Results: []sql.Row{{123, 456}, {123, `4,567,890,123,456,789`}, {`4,567,890,123,456,789`, 123}, {`4,567,890,123,456,789`, `4,567,890,123,456,789`}, {`4,567,890,123,456,789`, `-4,567,890,123,456,789`}},
			},
			{
				Statement: `SELECT to_char(q1, '9G999G999G999G999G999D999G999'), to_char(q2, '9,999,999,999,999,999.999,999')
	FROM INT8_TBL;`,
				Results: []sql.Row{{`123.000,000`, `456.000,000`}, {`123.000,000`, `4,567,890,123,456,789.000,000`}, {`4,567,890,123,456,789.000,000`, `123.000,000`}, {`4,567,890,123,456,789.000,000`, `4,567,890,123,456,789.000,000`}, {`4,567,890,123,456,789.000,000`, `-4,567,890,123,456,789.000,000`}},
			},
			{
				Statement: `SELECT to_char( (q1 * -1), '9999999999999999PR'), to_char( (q2 * -1), '9999999999999999.999PR')
	FROM INT8_TBL;`,
				Results: []sql.Row{{`<123>`, `<456.000>`}, {`<123>`, `<4567890123456789.000>`}, {`<4567890123456789>`, `<123.000>`}, {`<4567890123456789>`, `<4567890123456789.000>`}, {`<4567890123456789>`, 4567890123456789.000}},
			},
			{
				Statement: `SELECT to_char( (q1 * -1), '9999999999999999S'), to_char( (q2 * -1), 'S9999999999999999')
	FROM INT8_TBL;`,
				Results: []sql.Row{{`123-`, -456}, {`123-`, -4567890123456789}, {`4567890123456789-`, -123}, {`4567890123456789-`, -4567890123456789}, {`4567890123456789-`, +4567890123456789}},
			},
			{
				Statement: `SELECT to_char(q2, 'MI9999999999999999')     FROM INT8_TBL;`,
				Results:   []sql.Row{{456}, {4567890123456789}, {123}, {4567890123456789}, {-4567890123456789}},
			},
			{
				Statement: `SELECT to_char(q2, 'FMS9999999999999999')    FROM INT8_TBL;`,
				Results:   []sql.Row{{+456}, {+4567890123456789}, {+123}, {+4567890123456789}, {-4567890123456789}},
			},
			{
				Statement: `SELECT to_char(q2, 'FM9999999999999999THPR') FROM INT8_TBL;`,
				Results:   []sql.Row{{`456TH`}, {`4567890123456789TH`}, {`123RD`}, {`4567890123456789TH`}, {`<4567890123456789>`}},
			},
			{
				Statement: `SELECT to_char(q2, 'SG9999999999999999th')   FROM INT8_TBL;`,
				Results:   []sql.Row{{`+             456th`}, {`+4567890123456789th`}, {`+             123rd`}, {`+4567890123456789th`}, {-4567890123456789}},
			},
			{
				Statement: `SELECT to_char(q2, '0999999999999999')       FROM INT8_TBL;`,
				Results:   []sql.Row{{0000000000000456}, {4567890123456789}, {0000000000000123}, {4567890123456789}, {-4567890123456789}},
			},
			{
				Statement: `SELECT to_char(q2, 'S0999999999999999')      FROM INT8_TBL;`,
				Results:   []sql.Row{{+0000000000000456}, {+4567890123456789}, {+0000000000000123}, {+4567890123456789}, {-4567890123456789}},
			},
			{
				Statement: `SELECT to_char(q2, 'FM0999999999999999')     FROM INT8_TBL;`,
				Results:   []sql.Row{{0000000000000456}, {4567890123456789}, {0000000000000123}, {4567890123456789}, {-4567890123456789}},
			},
			{
				Statement: `SELECT to_char(q2, 'FM9999999999999999.000') FROM INT8_TBL;`,
				Results:   []sql.Row{{456.000}, {4567890123456789.000}, {123.000}, {4567890123456789.000}, {-4567890123456789.000}},
			},
			{
				Statement: `SELECT to_char(q2, 'L9999999999999999.000')  FROM INT8_TBL;`,
				Results:   []sql.Row{{456.000}, {4567890123456789.000}, {123.000}, {4567890123456789.000}, {-4567890123456789.000}},
			},
			{
				Statement: `SELECT to_char(q2, 'FM9999999999999999.999') FROM INT8_TBL;`,
				Results:   []sql.Row{{456.}, {4567890123456789.}, {123.}, {4567890123456789.}, {-4567890123456789.}},
			},
			{
				Statement: `SELECT to_char(q2, 'S 9 9 9 9 9 9 9 9 9 9 9 9 9 9 9 9 . 9 9 9') FROM INT8_TBL;`,
				Results:   []sql.Row{{`+4 5 6 . 0 0 0`}, {`+4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 . 0 0 0`}, {`+1 2 3 . 0 0 0`}, {`+4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 . 0 0 0`}, {`-4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 . 0 0 0`}},
			},
			{
				Statement: `SELECT to_char(q2, E'99999 "text" 9999 "9999" 999 "\\"text between quote marks\\"" 9999') FROM INT8_TBL;`,
				Results:   []sql.Row{{`text      9999     "text between quote marks"   456`}, {`45678 text 9012 9999 345 "text between quote marks" 6789`}, {`text      9999     "text between quote marks"   123`}, {`45678 text 9012 9999 345 "text between quote marks" 6789`}, {`-45678 text 9012 9999 345 "text between quote marks" 6789`}},
			},
			{
				Statement: `SELECT to_char(q2, '999999SG9999999999')     FROM INT8_TBL;`,
				Results:   []sql.Row{{`+       456`}, {`456789+0123456789`}, {`+       123`}, {`456789+0123456789`}, {`456789-0123456789`}},
			},
			{
				Statement: `select '-9223372036854775808'::int8;`,
				Results:   []sql.Row{{-9223372036854775808}},
			},
			{
				Statement:   `select '-9223372036854775809'::int8;`,
				ErrorString: `value "-9223372036854775809" is out of range for type bigint`,
			},
			{
				Statement: `select '9223372036854775807'::int8;`,
				Results:   []sql.Row{{9223372036854775807}},
			},
			{
				Statement:   `select '9223372036854775808'::int8;`,
				ErrorString: `value "9223372036854775808" is out of range for type bigint`,
			},
			{
				Statement: `select -('-9223372036854775807'::int8);`,
				Results:   []sql.Row{{9223372036854775807}},
			},
			{
				Statement:   `select -('-9223372036854775808'::int8);`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 + '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '-9223372036854775800'::int8 + '-9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 - '-9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '-9223372036854775800'::int8 - '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 * '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 / '0'::int8;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 % '0'::int8;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select abs('-9223372036854775808'::int8);`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 + '100'::int4;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '-9223372036854775800'::int8 - '100'::int4;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 * '100'::int4;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '100'::int4 + '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '-100'::int4 - '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '100'::int4 * '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 + '100'::int2;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '-9223372036854775800'::int8 - '100'::int2;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '9223372036854775800'::int8 * '100'::int2;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '-9223372036854775808'::int8 / '0'::int2;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `select '100'::int2 + '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '-100'::int2 - '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '100'::int2 * '9223372036854775800'::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `select '100'::int2 / '0'::int8;`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `SELECT CAST(q1 AS int4) FROM int8_tbl WHERE q2 = 456;`,
				Results:   []sql.Row{{123}},
			},
			{
				Statement:   `SELECT CAST(q1 AS int4) FROM int8_tbl WHERE q2 <> 456;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT CAST(q1 AS int2) FROM int8_tbl WHERE q2 = 456;`,
				Results:   []sql.Row{{123}},
			},
			{
				Statement:   `SELECT CAST(q1 AS int2) FROM int8_tbl WHERE q2 <> 456;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT CAST('42'::int2 AS int8), CAST('-37'::int2 AS int8);`,
				Results:   []sql.Row{{42, -37}},
			},
			{
				Statement: `SELECT CAST(q1 AS float4), CAST(q2 AS float8) FROM INT8_TBL;`,
				Results:   []sql.Row{{123, 456}, {123, 4.567890123456789e+15}, {4.56789e+15, 123}, {4.56789e+15, 4.567890123456789e+15}, {4.56789e+15, -4.567890123456789e+15}},
			},
			{
				Statement: `SELECT CAST('36854775807.0'::float4 AS int8);`,
				Results:   []sql.Row{{36854775808}},
			},
			{
				Statement:   `SELECT CAST('922337203685477580700.0'::float8 AS int8);`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `SELECT CAST(q1 AS oid) FROM INT8_TBL;`,
				ErrorString: `OID out of range`,
			},
			{
				Statement: `SELECT oid::int8 FROM pg_class WHERE relname = 'pg_class';`,
				Results:   []sql.Row{{1259}},
			},
			{
				Statement: `SELECT q1, q2, q1 & q2 AS "and", q1 | q2 AS "or", q1 # q2 AS "xor", ~q1 AS "not" FROM INT8_TBL;`,
				Results:   []sql.Row{{123, 456, 72, 507, 435, -124}, {123, 4567890123456789, 17, 4567890123456895, 4567890123456878, -124}, {4567890123456789, 123, 17, 4567890123456895, 4567890123456878, -4567890123456790}, {4567890123456789, 4567890123456789, 4567890123456789, 4567890123456789, 0, -4567890123456790}, {4567890123456789, -4567890123456789, 1, -1, -2, -4567890123456790}},
			},
			{
				Statement: `SELECT q1, q1 << 2 AS "shl", q1 >> 3 AS "shr" FROM INT8_TBL;`,
				Results:   []sql.Row{{123, 492, 15}, {123, 492, 15}, {4567890123456789, 18271560493827156, 570986265432098}, {4567890123456789, 18271560493827156, 570986265432098}, {4567890123456789, 18271560493827156, 570986265432098}},
			},
			{
				Statement: `SELECT * FROM generate_series('+4567890123456789'::int8, '+4567890123456799'::int8);`,
				Results:   []sql.Row{{4567890123456789}, {4567890123456790}, {4567890123456791}, {4567890123456792}, {4567890123456793}, {4567890123456794}, {4567890123456795}, {4567890123456796}, {4567890123456797}, {4567890123456798}, {4567890123456799}},
			},
			{
				Statement:   `SELECT * FROM generate_series('+4567890123456789'::int8, '+4567890123456799'::int8, 0);`,
				ErrorString: `step size cannot equal zero`,
			},
			{
				Statement: `SELECT * FROM generate_series('+4567890123456789'::int8, '+4567890123456799'::int8, 2);`,
				Results:   []sql.Row{{4567890123456789}, {4567890123456791}, {4567890123456793}, {4567890123456795}, {4567890123456797}, {4567890123456799}},
			},
			{
				Statement: `SELECT (-1::int8<<63)::text;`,
				Results:   []sql.Row{{-9223372036854775808}},
			},
			{
				Statement: `SELECT ((-1::int8<<63)+1)::text;`,
				Results:   []sql.Row{{-9223372036854775807}},
			},
			{
				Statement:   `SELECT (-9223372036854775808)::int8 * (-1)::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `SELECT (-9223372036854775808)::int8 / (-1)::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT (-9223372036854775808)::int8 % (-1)::int8;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT (-9223372036854775808)::int8 * (-1)::int4;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `SELECT (-9223372036854775808)::int8 / (-1)::int4;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT (-9223372036854775808)::int8 % (-1)::int4;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT (-9223372036854775808)::int8 * (-1)::int2;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `SELECT (-9223372036854775808)::int8 / (-1)::int2;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT (-9223372036854775808)::int8 % (-1)::int2;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT x, x::int8 AS int8_value
FROM (VALUES (-2.5::float8),
             (-1.5::float8),
             (-0.5::float8),
             (0.0::float8),
             (0.5::float8),
             (1.5::float8),
             (2.5::float8)) t(x);`,
				Results: []sql.Row{{-2.5, -2}, {-1.5, -2}, {-0.5, 0}, {0, 0}, {0.5, 0}, {1.5, 2}, {2.5, 2}},
			},
			{
				Statement: `SELECT x, x::int8 AS int8_value
FROM (VALUES (-2.5::numeric),
             (-1.5::numeric),
             (-0.5::numeric),
             (0.0::numeric),
             (0.5::numeric),
             (1.5::numeric),
             (2.5::numeric)) t(x);`,
				Results: []sql.Row{{-2.5, -3}, {-1.5, -2}, {-0.5, -1}, {0.0, 0}, {0.5, 1}, {1.5, 2}, {2.5, 3}},
			},
			{
				Statement: `SELECT a, b, gcd(a, b), gcd(a, -b), gcd(b, a), gcd(-b, a)
FROM (VALUES (0::int8, 0::int8),
             (0::int8, 29893644334::int8),
             (288484263558::int8, 29893644334::int8),
             (-288484263558::int8, 29893644334::int8),
             ((-9223372036854775808)::int8, 1::int8),
             ((-9223372036854775808)::int8, 9223372036854775807::int8),
             ((-9223372036854775808)::int8, 4611686018427387904::int8)) AS v(a, b);`,
				Results: []sql.Row{{0, 0, 0, 0, 0, 0}, {0, 29893644334, 29893644334, 29893644334, 29893644334, 29893644334}, {288484263558, 29893644334, 6835958, 6835958, 6835958, 6835958}, {-288484263558, 29893644334, 6835958, 6835958, 6835958, 6835958}, {-9223372036854775808, 1, 1, 1, 1, 1}, {-9223372036854775808, 9223372036854775807, 1, 1, 1, 1}, {-9223372036854775808, 4611686018427387904, 4611686018427387904, 4611686018427387904, 4611686018427387904, 4611686018427387904}},
			},
			{
				Statement:   `SELECT gcd((-9223372036854775808)::int8, 0::int8); -- overflow`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `SELECT gcd((-9223372036854775808)::int8, (-9223372036854775808)::int8); -- overflow`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT a, b, lcm(a, b), lcm(a, -b), lcm(b, a), lcm(-b, a)
FROM (VALUES (0::int8, 0::int8),
             (0::int8, 29893644334::int8),
             (29893644334::int8, 29893644334::int8),
             (288484263558::int8, 29893644334::int8),
             (-288484263558::int8, 29893644334::int8),
             ((-9223372036854775808)::int8, 0::int8)) AS v(a, b);`,
				Results: []sql.Row{{0, 0, 0, 0, 0, 0}, {0, 29893644334, 0, 0, 0, 0}, {29893644334, 29893644334, 29893644334, 29893644334, 29893644334, 29893644334}, {288484263558, 29893644334, 1261541684539134, 1261541684539134, 1261541684539134, 1261541684539134}, {-288484263558, 29893644334, 1261541684539134, 1261541684539134, 1261541684539134, 1261541684539134}, {-9223372036854775808, 0, 0, 0, 0, 0}},
			},
			{
				Statement:   `SELECT lcm((-9223372036854775808)::int8, 1::int8); -- overflow`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `SELECT lcm(9223372036854775807::int8, 9223372036854775806::int8); -- overflow`,
				ErrorString: `bigint out of range`,
			},
		},
	})
}
