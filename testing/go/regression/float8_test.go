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

func TestFloat8(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_float8)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_float8,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TEMP TABLE FLOAT8_TBL(f1 float8);`,
			},
			{
				Statement: `INSERT INTO FLOAT8_TBL(f1) VALUES ('    0.0   ');`,
			},
			{
				Statement: `INSERT INTO FLOAT8_TBL(f1) VALUES ('1004.30  ');`,
			},
			{
				Statement: `INSERT INTO FLOAT8_TBL(f1) VALUES ('   -34.84');`,
			},
			{
				Statement: `INSERT INTO FLOAT8_TBL(f1) VALUES ('1.2345678901234e+200');`,
			},
			{
				Statement: `INSERT INTO FLOAT8_TBL(f1) VALUES ('1.2345678901234e-200');`,
			},
			{
				Statement:   `SELECT '10e400'::float8;`,
				ErrorString: `"10e400" is out of range for type double precision`,
			},
			{
				Statement:   `SELECT '-10e400'::float8;`,
				ErrorString: `"-10e400" is out of range for type double precision`,
			},
			{
				Statement:   `SELECT '10e-400'::float8;`,
				ErrorString: `"10e-400" is out of range for type double precision`,
			},
			{
				Statement:   `SELECT '-10e-400'::float8;`,
				ErrorString: `"-10e-400" is out of range for type double precision`,
			},
			{
				Statement: `SELECT float8send('2.2250738585072014E-308'::float8);`,
				Results:   []sql.Row{{`\x0010000000000000`}},
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('');`,
				ErrorString: `invalid input syntax for type double precision: ""`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('     ');`,
				ErrorString: `invalid input syntax for type double precision: "     "`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('xyz');`,
				ErrorString: `invalid input syntax for type double precision: "xyz"`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('5.0.0');`,
				ErrorString: `invalid input syntax for type double precision: "5.0.0"`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('5 . 0');`,
				ErrorString: `invalid input syntax for type double precision: "5 . 0"`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('5.   0');`,
				ErrorString: `invalid input syntax for type double precision: "5.   0"`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('    - 3');`,
				ErrorString: `invalid input syntax for type double precision: "    - 3"`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('123           5');`,
				ErrorString: `invalid input syntax for type double precision: "123           5"`,
			},
			{
				Statement: `SELECT 'NaN'::float8;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT 'nan'::float8;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT '   NAN  '::float8;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT 'infinity'::float8;`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT '          -INFINiTY   '::float8;`,
				Results:   []sql.Row{{`-Infinity`}},
			},
			{
				Statement:   `SELECT 'N A N'::float8;`,
				ErrorString: `invalid input syntax for type double precision: "N A N"`,
			},
			{
				Statement:   `SELECT 'NaN x'::float8;`,
				ErrorString: `invalid input syntax for type double precision: "NaN x"`,
			},
			{
				Statement:   `SELECT ' INFINITY    x'::float8;`,
				ErrorString: `invalid input syntax for type double precision: " INFINITY    x"`,
			},
			{
				Statement: `SELECT 'Infinity'::float8 + 100.0;`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT 'Infinity'::float8 / 'Infinity'::float8;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT '42'::float8 / 'Infinity'::float8;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT 'nan'::float8 / 'nan'::float8;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT 'nan'::float8 / '0'::float8;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT 'nan'::numeric::float8;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT * FROM FLOAT8_TBL;`,
				Results:   []sql.Row{{0}, {1004.3}, {-34.84}, {1.2345678901234e+200}, {1.2345678901234e-200}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT8_TBL f WHERE f.f1 <> '1004.3';`,
				Results:   []sql.Row{{0}, {-34.84}, {1.2345678901234e+200}, {1.2345678901234e-200}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT8_TBL f WHERE f.f1 = '1004.3';`,
				Results:   []sql.Row{{1004.3}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT8_TBL f WHERE '1004.3' > f.f1;`,
				Results:   []sql.Row{{0}, {-34.84}, {1.2345678901234e-200}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT8_TBL f WHERE  f.f1 < '1004.3';`,
				Results:   []sql.Row{{0}, {-34.84}, {1.2345678901234e-200}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT8_TBL f WHERE '1004.3' >= f.f1;`,
				Results:   []sql.Row{{0}, {1004.3}, {-34.84}, {1.2345678901234e-200}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT8_TBL f WHERE  f.f1 <= '1004.3';`,
				Results:   []sql.Row{{0}, {1004.3}, {-34.84}, {1.2345678901234e-200}},
			},
			{
				Statement: `SELECT f.f1, f.f1 * '-10' AS x
   FROM FLOAT8_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, -10043}, {1.2345678901234e+200, -1.2345678901234e+201}, {1.2345678901234e-200, -1.2345678901234e-199}},
			},
			{
				Statement: `SELECT f.f1, f.f1 + '-10' AS x
   FROM FLOAT8_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, 994.3}, {1.2345678901234e+200, 1.2345678901234e+200}, {1.2345678901234e-200, -10}},
			},
			{
				Statement: `SELECT f.f1, f.f1 / '-10' AS x
   FROM FLOAT8_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, -100.42999999999999}, {1.2345678901234e+200, -1.2345678901234e+199}, {1.2345678901234e-200, -1.2345678901234e-201}},
			},
			{
				Statement: `SELECT f.f1, f.f1 - '-10' AS x
   FROM FLOAT8_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, 1014.3}, {1.2345678901234e+200, 1.2345678901234e+200}, {1.2345678901234e-200, 10}},
			},
			{
				Statement: `SELECT f.f1 ^ '2.0' AS square_f1
   FROM FLOAT8_TBL f where f.f1 = '1004.3';`,
				Results: []sql.Row{{1008618.4899999999}},
			},
			{
				Statement: `SELECT f.f1, @f.f1 AS abs_f1
   FROM FLOAT8_TBL f;`,
				Results: []sql.Row{{0, 0}, {1004.3, 1004.3}, {-34.84, 34.84}, {1.2345678901234e+200, 1.2345678901234e+200}, {1.2345678901234e-200, 1.2345678901234e-200}},
			},
			{
				Statement: `SELECT f.f1, trunc(f.f1) AS trunc_f1
   FROM FLOAT8_TBL f;`,
				Results: []sql.Row{{0, 0}, {1004.3, 1004}, {-34.84, -34}, {1.2345678901234e+200, 1.2345678901234e+200}, {1.2345678901234e-200, 0}},
			},
			{
				Statement: `SELECT f.f1, round(f.f1) AS round_f1
   FROM FLOAT8_TBL f;`,
				Results: []sql.Row{{0, 0}, {1004.3, 1004}, {-34.84, -35}, {1.2345678901234e+200, 1.2345678901234e+200}, {1.2345678901234e-200, 0}},
			},
			{
				Statement: `select ceil(f1) as ceil_f1 from float8_tbl f;`,
				Results:   []sql.Row{{0}, {1005}, {-34}, {1.2345678901234e+200}, {1}},
			},
			{
				Statement: `select ceiling(f1) as ceiling_f1 from float8_tbl f;`,
				Results:   []sql.Row{{0}, {1005}, {-34}, {1.2345678901234e+200}, {1}},
			},
			{
				Statement: `select floor(f1) as floor_f1 from float8_tbl f;`,
				Results:   []sql.Row{{0}, {1004}, {-35}, {1.2345678901234e+200}, {0}},
			},
			{
				Statement: `select sign(f1) as sign_f1 from float8_tbl f;`,
				Results:   []sql.Row{{0}, {1}, {-1}, {1}, {1}},
			},
			{
				Statement: `SET extra_float_digits = 0;`,
			},
			{
				Statement: `SELECT sqrt(float8 '64') AS eight;`,
				Results:   []sql.Row{{8}},
			},
			{
				Statement: `SELECT |/ float8 '64' AS eight;`,
				Results:   []sql.Row{{8}},
			},
			{
				Statement: `SELECT f.f1, |/f.f1 AS sqrt_f1
   FROM FLOAT8_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, 31.6906926399535}, {1.2345678901234e+200, 1.11111110611109e+100}, {1.2345678901234e-200, 1.11111110611109e-100}},
			},
			{
				Statement: `SELECT power(float8 '144', float8 '0.5');`,
				Results:   []sql.Row{{12}},
			},
			{
				Statement: `SELECT power(float8 'NaN', float8 '0.5');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT power(float8 '144', float8 'NaN');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT power(float8 'NaN', float8 'NaN');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT power(float8 '-1', float8 'NaN');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT power(float8 '1', float8 'NaN');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT power(float8 'NaN', float8 '0');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT power(float8 'inf', float8 '0');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT power(float8 '-inf', float8 '0');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT power(float8 '0', float8 'inf');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT power(float8 '0', float8 '-inf');`,
				ErrorString: `zero raised to a negative power is undefined`,
			},
			{
				Statement: `SELECT power(float8 '1', float8 'inf');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT power(float8 '1', float8 '-inf');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT power(float8 '-1', float8 'inf');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT power(float8 '-1', float8 '-inf');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT power(float8 '0.1', float8 'inf');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT power(float8 '-0.1', float8 'inf');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT power(float8 '1.1', float8 'inf');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT power(float8 '-1.1', float8 'inf');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT power(float8 '0.1', float8 '-inf');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT power(float8 '-0.1', float8 '-inf');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT power(float8 '1.1', float8 '-inf');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT power(float8 '-1.1', float8 '-inf');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT power(float8 'inf', float8 '-2');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT power(float8 'inf', float8 '2');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT power(float8 'inf', float8 'inf');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT power(float8 'inf', float8 '-inf');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT power(float8 '-inf', float8 '-2') = '0';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT power(float8 '-inf', float8 '-3');`,
				Results:   []sql.Row{{-0}},
			},
			{
				Statement: `SELECT power(float8 '-inf', float8 '2');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT power(float8 '-inf', float8 '3');`,
				Results:   []sql.Row{{`-Infinity`}},
			},
			{
				Statement:   `SELECT power(float8 '-inf', float8 '3.5');`,
				ErrorString: `a negative number raised to a non-integer power yields a complex result`,
			},
			{
				Statement: `SELECT power(float8 '-inf', float8 'inf');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT power(float8 '-inf', float8 '-inf');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT f.f1, exp(ln(f.f1)) AS exp_ln_f1
   FROM FLOAT8_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, 1004.3}, {1.2345678901234e+200, 1.23456789012338e+200}, {1.2345678901234e-200, 1.23456789012339e-200}},
			},
			{
				Statement: `SELECT exp('inf'::float8), exp('-inf'::float8), exp('nan'::float8);`,
				Results:   []sql.Row{{`Infinity`, 0, `NaN`}},
			},
			{
				Statement: `SELECT ||/ float8 '27' AS three;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT f.f1, ||/f.f1 AS cbrt_f1 FROM FLOAT8_TBL f;`,
				Results:   []sql.Row{{0, 0}, {1004.3, 10.014312837827}, {-34.84, -3.26607421344208}, {1.2345678901234e+200, 4.97933859234765e+66}, {1.2345678901234e-200, 2.3112042409018e-67}},
			},
			{
				Statement: `SELECT * FROM FLOAT8_TBL;`,
				Results:   []sql.Row{{0}, {1004.3}, {-34.84}, {1.2345678901234e+200}, {1.2345678901234e-200}},
			},
			{
				Statement: `UPDATE FLOAT8_TBL
   SET f1 = FLOAT8_TBL.f1 * '-1'
   WHERE FLOAT8_TBL.f1 > '0.0';`,
			},
			{
				Statement:   `SELECT f.f1 * '1e200' from FLOAT8_TBL f;`,
				ErrorString: `value out of range: overflow`,
			},
			{
				Statement:   `SELECT f.f1 ^ '1e200' from FLOAT8_TBL f;`,
				ErrorString: `value out of range: overflow`,
			},
			{
				Statement: `SELECT 0 ^ 0 + 0 ^ 1 + 0 ^ 0.0 + 0 ^ 0.5;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement:   `SELECT ln(f.f1) from FLOAT8_TBL f where f.f1 = '0.0' ;`,
				ErrorString: `cannot take logarithm of zero`,
			},
			{
				Statement:   `SELECT ln(f.f1) from FLOAT8_TBL f where f.f1 < '0.0' ;`,
				ErrorString: `cannot take logarithm of a negative number`,
			},
			{
				Statement:   `SELECT exp(f.f1) from FLOAT8_TBL f;`,
				ErrorString: `value out of range: underflow`,
			},
			{
				Statement:   `SELECT f.f1 / '0.0' from FLOAT8_TBL f;`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `SELECT * FROM FLOAT8_TBL;`,
				Results:   []sql.Row{{0}, {-34.84}, {-1004.3}, {-1.2345678901234e+200}, {-1.2345678901234e-200}},
			},
			{
				Statement: `SELECT sinh(float8 '1');`,
				Results:   []sql.Row{{1.1752011936438}},
			},
			{
				Statement: `SELECT cosh(float8 '1');`,
				Results:   []sql.Row{{1.54308063481524}},
			},
			{
				Statement: `SELECT tanh(float8 '1');`,
				Results:   []sql.Row{{0.761594155955765}},
			},
			{
				Statement: `SELECT asinh(float8 '1');`,
				Results:   []sql.Row{{0.881373587019543}},
			},
			{
				Statement: `SELECT acosh(float8 '2');`,
				Results:   []sql.Row{{1.31695789692482}},
			},
			{
				Statement: `SELECT atanh(float8 '0.5');`,
				Results:   []sql.Row{{0.549306144334055}},
			},
			{
				Statement: `SELECT sinh(float8 'infinity');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT sinh(float8 '-infinity');`,
				Results:   []sql.Row{{`-Infinity`}},
			},
			{
				Statement: `SELECT sinh(float8 'nan');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT cosh(float8 'infinity');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT cosh(float8 '-infinity');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT cosh(float8 'nan');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT tanh(float8 'infinity');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT tanh(float8 '-infinity');`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT tanh(float8 'nan');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT asinh(float8 'infinity');`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT asinh(float8 '-infinity');`,
				Results:   []sql.Row{{`-Infinity`}},
			},
			{
				Statement: `SELECT asinh(float8 'nan');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement:   `SELECT acosh(float8 '-infinity');`,
				ErrorString: `input is out of range`,
			},
			{
				Statement: `SELECT acosh(float8 'nan');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement:   `SELECT atanh(float8 'infinity');`,
				ErrorString: `input is out of range`,
			},
			{
				Statement:   `SELECT atanh(float8 '-infinity');`,
				ErrorString: `input is out of range`,
			},
			{
				Statement: `SELECT atanh(float8 'nan');`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `RESET extra_float_digits;`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('10e400');`,
				ErrorString: `"10e400" is out of range for type double precision`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('-10e400');`,
				ErrorString: `"-10e400" is out of range for type double precision`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('10e-400');`,
				ErrorString: `"10e-400" is out of range for type double precision`,
			},
			{
				Statement:   `INSERT INTO FLOAT8_TBL(f1) VALUES ('-10e-400');`,
				ErrorString: `"-10e-400" is out of range for type double precision`,
			},
			{
				Statement: `DROP TABLE FLOAT8_TBL;`,
			},
			{
				Statement: `SELECT * FROM FLOAT8_TBL;`,
				Results:   []sql.Row{{0}, {-34.84}, {-1004.3}, {-1.2345678901234e+200}, {-1.2345678901234e-200}},
			},
			{
				Statement: `SELECT '32767.4'::float8::int2;`,
				Results:   []sql.Row{{32767}},
			},
			{
				Statement:   `SELECT '32767.6'::float8::int2;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT '-32768.4'::float8::int2;`,
				Results:   []sql.Row{{-32768}},
			},
			{
				Statement:   `SELECT '-32768.6'::float8::int2;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT '2147483647.4'::float8::int4;`,
				Results:   []sql.Row{{2147483647}},
			},
			{
				Statement:   `SELECT '2147483647.6'::float8::int4;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT '-2147483648.4'::float8::int4;`,
				Results:   []sql.Row{{-2147483648}},
			},
			{
				Statement:   `SELECT '-2147483648.6'::float8::int4;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT '9223372036854773760'::float8::int8;`,
				Results:   []sql.Row{{9223372036854773760}},
			},
			{
				Statement:   `SELECT '9223372036854775807'::float8::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT '-9223372036854775808.5'::float8::int8;`,
				Results:   []sql.Row{{-9223372036854775808}},
			},
			{
				Statement:   `SELECT '-9223372036854780000'::float8::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT x,
       sind(x),
       sind(x) IN (-1,-0.5,0,0.5,1) AS sind_exact
FROM (VALUES (0), (30), (90), (150), (180),
      (210), (270), (330), (360)) AS t(x);`,
				Results: []sql.Row{{0, 0, true}, {30, 0.5, true}, {90, 1, true}, {150, 0.5, true}, {180, 0, true}, {210, -0.5, true}, {270, -1, true}, {330, -0.5, true}, {360, 0, true}},
			},
			{
				Statement: `SELECT x,
       cosd(x),
       cosd(x) IN (-1,-0.5,0,0.5,1) AS cosd_exact
FROM (VALUES (0), (60), (90), (120), (180),
      (240), (270), (300), (360)) AS t(x);`,
				Results: []sql.Row{{0, 1, true}, {60, 0.5, true}, {90, 0, true}, {120, -0.5, true}, {180, -1, true}, {240, -0.5, true}, {270, 0, true}, {300, 0.5, true}, {360, 1, true}},
			},
			{
				Statement: `SELECT x,
       tand(x),
       tand(x) IN ('-Infinity'::float8,-1,0,
                   1,'Infinity'::float8) AS tand_exact,
       cotd(x),
       cotd(x) IN ('-Infinity'::float8,-1,0,
                   1,'Infinity'::float8) AS cotd_exact
FROM (VALUES (0), (45), (90), (135), (180),
      (225), (270), (315), (360)) AS t(x);`,
				Results: []sql.Row{{0, 0, true, `Infinity`, true}, {45, 1, true, 1, true}, {90, `Infinity`, true, 0, true}, {135, -1, true, -1, true}, {180, 0, true, `-Infinity`, true}, {225, 1, true, 1, true}, {270, `-Infinity`, true, 0, true}, {315, -1, true, -1, true}, {360, 0, true, `Infinity`, true}},
			},
			{
				Statement: `SELECT x,
       asind(x),
       asind(x) IN (-90,-30,0,30,90) AS asind_exact,
       acosd(x),
       acosd(x) IN (0,60,90,120,180) AS acosd_exact
FROM (VALUES (-1), (-0.5), (0), (0.5), (1)) AS t(x);`,
				Results: []sql.Row{{-1, -90, true, 180, true}, {-0.5, -30, true, 120, true}, {0, 0, true, 90, true}, {0.5, 30, true, 60, true}, {1, 90, true, 0, true}},
			},
			{
				Statement: `SELECT x,
       atand(x),
       atand(x) IN (-90,-45,0,45,90) AS atand_exact
FROM (VALUES ('-Infinity'::float8), (-1), (0), (1),
      ('Infinity'::float8)) AS t(x);`,
				Results: []sql.Row{{`-Infinity`, -90, true}, {-1, -45, true}, {0, 0, true}, {1, 45, true}, {`Infinity`, 90, true}},
			},
			{
				Statement: `SELECT x, y,
       atan2d(y, x),
       atan2d(y, x) IN (-90,0,90,180) AS atan2d_exact
FROM (SELECT 10*cosd(a), 10*sind(a)
      FROM generate_series(0, 360, 90) AS t(a)) AS t(x,y);`,
				Results: []sql.Row{{10, 0, 0, true}, {0, 10, 90, true}, {-10, 0, 180, true}, {0, -10, -90, true}, {10, 0, 0, true}},
			},
			{
				Statement: `create type xfloat8;`,
			},
			{
				Statement: `create function xfloat8in(cstring) returns xfloat8 immutable strict
  language internal as 'int8in';`,
			},
			{
				Statement: `create function xfloat8out(xfloat8) returns cstring immutable strict
  language internal as 'int8out';`,
			},
			{
				Statement: `create type xfloat8 (input = xfloat8in, output = xfloat8out, like = float8);`,
			},
			{
				Statement: `create cast (xfloat8 as float8) without function;`,
			},
			{
				Statement: `create cast (float8 as xfloat8) without function;`,
			},
			{
				Statement: `create cast (xfloat8 as bigint) without function;`,
			},
			{
				Statement: `create cast (bigint as xfloat8) without function;`,
			},
			{
				Statement: `with testdata(bits) as (values
  -- small subnormals
  (x'0000000000000001'),
  (x'0000000000000002'), (x'0000000000000003'),
  (x'0000000000001000'), (x'0000000100000000'),
  (x'0000010000000000'), (x'0000010100000000'),
  (x'0000400000000000'), (x'0000400100000000'),
  (x'0000800000000000'), (x'0000800000000001'),
  -- these values taken from upstream testsuite
  (x'00000000000f4240'),
  (x'00000000016e3600'),
  (x'0000008cdcdea440'),
  -- borderline between subnormal and normal
  (x'000ffffffffffff0'), (x'000ffffffffffff1'),
  (x'000ffffffffffffe'), (x'000fffffffffffff'))
select float8send(flt) as ibits,
       flt
  from (select bits::bigint::xfloat8::float8 as flt
          from testdata
	offset 0) s;`,
				Results: []sql.Row{{`\x0000000000000001`, 5e-324}, {`\x0000000000000002`, 1e-323}, {`\x0000000000000003`, 1.5e-323}, {`\x0000000000001000`, 2.0237e-320}, {`\x0000000100000000`, 2.121995791e-314}, {`\x0000010000000000`, 5.43230922487e-312}, {`\x0000010100000000`, 5.45352918278e-312}, {`\x0000400000000000`, 3.4766779039175e-310}, {`\x0000400100000000`, 3.4768901034966e-310}, {`\x0000800000000000`, 6.953355807835e-310}, {`\x0000800000000001`, 6.95335580783505e-310}, {`\x00000000000f4240`, 4.940656e-318}, {`\x00000000016e3600`, 1.18575755e-316}, {`\x0000008cdcdea440`, 2.989102097996e-312}, {`\x000ffffffffffff0`, 2.2250738585071935e-308}, {`\x000ffffffffffff1`, 2.225073858507194e-308}, {`\x000ffffffffffffe`, 2.2250738585072004e-308}, {`\x000fffffffffffff`, 2.225073858507201e-308}},
			},
			{
				Statement: `with testdata(bits) as (values
  (x'0000000000000000'),
  -- smallest normal values
  (x'0010000000000000'), (x'0010000000000001'),
  (x'0010000000000002'), (x'0018000000000000'),
  --
  (x'3ddb7cdfd9d7bdba'), (x'3ddb7cdfd9d7bdbb'), (x'3ddb7cdfd9d7bdbc'),
  (x'3e112e0be826d694'), (x'3e112e0be826d695'), (x'3e112e0be826d696'),
  (x'3e45798ee2308c39'), (x'3e45798ee2308c3a'), (x'3e45798ee2308c3b'),
  (x'3e7ad7f29abcaf47'), (x'3e7ad7f29abcaf48'), (x'3e7ad7f29abcaf49'),
  (x'3eb0c6f7a0b5ed8c'), (x'3eb0c6f7a0b5ed8d'), (x'3eb0c6f7a0b5ed8e'),
  (x'3ee4f8b588e368ef'), (x'3ee4f8b588e368f0'), (x'3ee4f8b588e368f1'),
  (x'3f1a36e2eb1c432c'), (x'3f1a36e2eb1c432d'), (x'3f1a36e2eb1c432e'),
  (x'3f50624dd2f1a9fb'), (x'3f50624dd2f1a9fc'), (x'3f50624dd2f1a9fd'),
  (x'3f847ae147ae147a'), (x'3f847ae147ae147b'), (x'3f847ae147ae147c'),
  (x'3fb9999999999999'), (x'3fb999999999999a'), (x'3fb999999999999b'),
  -- values very close to 1
  (x'3feffffffffffff0'), (x'3feffffffffffff1'), (x'3feffffffffffff2'),
  (x'3feffffffffffff3'), (x'3feffffffffffff4'), (x'3feffffffffffff5'),
  (x'3feffffffffffff6'), (x'3feffffffffffff7'), (x'3feffffffffffff8'),
  (x'3feffffffffffff9'), (x'3feffffffffffffa'), (x'3feffffffffffffb'),
  (x'3feffffffffffffc'), (x'3feffffffffffffd'), (x'3feffffffffffffe'),
  (x'3fefffffffffffff'),
  (x'3ff0000000000000'),
  (x'3ff0000000000001'), (x'3ff0000000000002'), (x'3ff0000000000003'),
  (x'3ff0000000000004'), (x'3ff0000000000005'), (x'3ff0000000000006'),
  (x'3ff0000000000007'), (x'3ff0000000000008'), (x'3ff0000000000009'),
  --
  (x'3ff921fb54442d18'),
  (x'4005bf0a8b14576a'),
  (x'400921fb54442d18'),
  --
  (x'4023ffffffffffff'), (x'4024000000000000'), (x'4024000000000001'),
  (x'4058ffffffffffff'), (x'4059000000000000'), (x'4059000000000001'),
  (x'408f3fffffffffff'), (x'408f400000000000'), (x'408f400000000001'),
  (x'40c387ffffffffff'), (x'40c3880000000000'), (x'40c3880000000001'),
  (x'40f869ffffffffff'), (x'40f86a0000000000'), (x'40f86a0000000001'),
  (x'412e847fffffffff'), (x'412e848000000000'), (x'412e848000000001'),
  (x'416312cfffffffff'), (x'416312d000000000'), (x'416312d000000001'),
  (x'4197d783ffffffff'), (x'4197d78400000000'), (x'4197d78400000001'),
  (x'41cdcd64ffffffff'), (x'41cdcd6500000000'), (x'41cdcd6500000001'),
  (x'4202a05f1fffffff'), (x'4202a05f20000000'), (x'4202a05f20000001'),
  (x'42374876e7ffffff'), (x'42374876e8000000'), (x'42374876e8000001'),
  (x'426d1a94a1ffffff'), (x'426d1a94a2000000'), (x'426d1a94a2000001'),
  (x'42a2309ce53fffff'), (x'42a2309ce5400000'), (x'42a2309ce5400001'),
  (x'42d6bcc41e8fffff'), (x'42d6bcc41e900000'), (x'42d6bcc41e900001'),
  (x'430c6bf52633ffff'), (x'430c6bf526340000'), (x'430c6bf526340001'),
  (x'4341c37937e07fff'), (x'4341c37937e08000'), (x'4341c37937e08001'),
  (x'4376345785d89fff'), (x'4376345785d8a000'), (x'4376345785d8a001'),
  (x'43abc16d674ec7ff'), (x'43abc16d674ec800'), (x'43abc16d674ec801'),
  (x'43e158e460913cff'), (x'43e158e460913d00'), (x'43e158e460913d01'),
  (x'4415af1d78b58c3f'), (x'4415af1d78b58c40'), (x'4415af1d78b58c41'),
  (x'444b1ae4d6e2ef4f'), (x'444b1ae4d6e2ef50'), (x'444b1ae4d6e2ef51'),
  (x'4480f0cf064dd591'), (x'4480f0cf064dd592'), (x'4480f0cf064dd593'),
  (x'44b52d02c7e14af5'), (x'44b52d02c7e14af6'), (x'44b52d02c7e14af7'),
  (x'44ea784379d99db3'), (x'44ea784379d99db4'), (x'44ea784379d99db5'),
  (x'45208b2a2c280290'), (x'45208b2a2c280291'), (x'45208b2a2c280292'),
  --
  (x'7feffffffffffffe'), (x'7fefffffffffffff'),
  -- round to even tests (+ve)
  (x'4350000000000002'),
  (x'4350000000002e06'),
  (x'4352000000000003'),
  (x'4352000000000004'),
  (x'4358000000000003'),
  (x'4358000000000004'),
  (x'435f000000000020'),
  -- round to even tests (-ve)
  (x'c350000000000002'),
  (x'c350000000002e06'),
  (x'c352000000000003'),
  (x'c352000000000004'),
  (x'c358000000000003'),
  (x'c358000000000004'),
  (x'c35f000000000020'),
  -- exercise fixed-point memmoves
  (x'42dc12218377de66'),
  (x'42a674e79c5fe51f'),
  (x'4271f71fb04cb74c'),
  (x'423cbe991a145879'),
  (x'4206fee0e1a9e061'),
  (x'41d26580b487e6b4'),
  (x'419d6f34540ca453'),
  (x'41678c29dcd6e9dc'),
  (x'4132d687e3df217d'),
  (x'40fe240c9fcb68c8'),
  (x'40c81cd6e63c53d3'),
  (x'40934a4584fd0fdc'),
  (x'405edd3c07fb4c93'),
  (x'4028b0fcd32f7076'),
  (x'3ff3c0ca428c59f8'),
  -- these cases come from the upstream's testsuite
  -- LotsOfTrailingZeros)
  (x'3e60000000000000'),
  -- Regression
  (x'c352bd2668e077c4'),
  (x'434018601510c000'),
  (x'43d055dc36f24000'),
  (x'43e052961c6f8000'),
  (x'3ff3c0ca2a5b1d5d'),
  -- LooksLikePow5
  (x'4830f0cf064dd592'),
  (x'4840f0cf064dd592'),
  (x'4850f0cf064dd592'),
  -- OutputLength
  (x'3ff3333333333333'),
  (x'3ff3ae147ae147ae'),
  (x'3ff3be76c8b43958'),
  (x'3ff3c083126e978d'),
  (x'3ff3c0c1fc8f3238'),
  (x'3ff3c0c9539b8887'),
  (x'3ff3c0ca2a5b1d5d'),
  (x'3ff3c0ca4283de1b'),
  (x'3ff3c0ca43db770a'),
  (x'3ff3c0ca428abd53'),
  (x'3ff3c0ca428c1d2b'),
  (x'3ff3c0ca428c51f2'),
  (x'3ff3c0ca428c58fc'),
  (x'3ff3c0ca428c59dd'),
  (x'3ff3c0ca428c59f8'),
  (x'3ff3c0ca428c59fb'),
  -- 32-bit chunking
  (x'40112e0be8047a7d'),
  (x'40112e0be815a889'),
  (x'40112e0be826d695'),
  (x'40112e0be83804a1'),
  (x'40112e0be84932ad'),
  -- MinMaxShift
  (x'0040000000000000'),
  (x'007fffffffffffff'),
  (x'0290000000000000'),
  (x'029fffffffffffff'),
  (x'4350000000000000'),
  (x'435fffffffffffff'),
  (x'1330000000000000'),
  (x'133fffffffffffff'),
  (x'3a6fa7161a4d6e0c')
)
select float8send(flt) as ibits,
       flt,
       flt::text::float8 as r_flt,
       float8send(flt::text::float8) as obits,
       float8send(flt::text::float8) = float8send(flt) as correct
  from (select bits::bigint::xfloat8::float8 as flt
          from testdata
	offset 0) s;`,
				Results: []sql.Row{{`\x0000000000000000`, 0, 0, `\x0000000000000000`, true}, {`\x0010000000000000`, 2.2250738585072014e-308, 2.2250738585072014e-308, `\x0010000000000000`, true}, {`\x0010000000000001`, 2.225073858507202e-308, 2.225073858507202e-308, `\x0010000000000001`, true}, {`\x0010000000000002`, 2.2250738585072024e-308, 2.2250738585072024e-308, `\x0010000000000002`, true}, {`\x0018000000000000`, 3.337610787760802e-308, 3.337610787760802e-308, `\x0018000000000000`, true}, {`\x3ddb7cdfd9d7bdba`, 9.999999999999999e-11, 9.999999999999999e-11, `\x3ddb7cdfd9d7bdba`, true}, {`\x3ddb7cdfd9d7bdbb`, 1e-10, 1e-10, `\x3ddb7cdfd9d7bdbb`, true}, {`\x3ddb7cdfd9d7bdbc`, 1.0000000000000002e-10, 1.0000000000000002e-10, `\x3ddb7cdfd9d7bdbc`, true}, {`\x3e112e0be826d694`, 9.999999999999999e-10, 9.999999999999999e-10, `\x3e112e0be826d694`, true}, {`\x3e112e0be826d695`, 1e-09, 1e-09, `\x3e112e0be826d695`, true}, {`\x3e112e0be826d696`, 1.0000000000000003e-09, 1.0000000000000003e-09, `\x3e112e0be826d696`, true}, {`\x3e45798ee2308c39`, 9.999999999999999e-09, 9.999999999999999e-09, `\x3e45798ee2308c39`, true}, {`\x3e45798ee2308c3a`, 1e-08, 1e-08, `\x3e45798ee2308c3a`, true}, {`\x3e45798ee2308c3b`, 1.0000000000000002e-08, 1.0000000000000002e-08, `\x3e45798ee2308c3b`, true}, {`\x3e7ad7f29abcaf47`, 9.999999999999998e-08, 9.999999999999998e-08, `\x3e7ad7f29abcaf47`, true}, {`\x3e7ad7f29abcaf48`, 1e-07, 1e-07, `\x3e7ad7f29abcaf48`, true}, {`\x3e7ad7f29abcaf49`, 1.0000000000000001e-07, 1.0000000000000001e-07, `\x3e7ad7f29abcaf49`, true}, {`\x3eb0c6f7a0b5ed8c`, 9.999999999999997e-07, 9.999999999999997e-07, `\x3eb0c6f7a0b5ed8c`, true}, {`\x3eb0c6f7a0b5ed8d`, 1e-06, 1e-06, `\x3eb0c6f7a0b5ed8d`, true}, {`\x3eb0c6f7a0b5ed8e`, 1.0000000000000002e-06, 1.0000000000000002e-06, `\x3eb0c6f7a0b5ed8e`, true}, {`\x3ee4f8b588e368ef`, 9.999999999999997e-06, 9.999999999999997e-06, `\x3ee4f8b588e368ef`, true}, {`\x3ee4f8b588e368f0`, 9.999999999999999e-06, 9.999999999999999e-06, `\x3ee4f8b588e368f0`, true}, {`\x3ee4f8b588e368f1`, 1e-05, 1e-05, `\x3ee4f8b588e368f1`, true}, {`\x3f1a36e2eb1c432c`, 9.999999999999999e-05, 9.999999999999999e-05, `\x3f1a36e2eb1c432c`, true}, {`\x3f1a36e2eb1c432d`, 0.0001, 0.0001, `\x3f1a36e2eb1c432d`, true}, {`\x3f1a36e2eb1c432e`, 0.00010000000000000002, 0.00010000000000000002, `\x3f1a36e2eb1c432e`, true}, {`\x3f50624dd2f1a9fb`, 0.0009999999999999998, 0.0009999999999999998, `\x3f50624dd2f1a9fb`, true}, {`\x3f50624dd2f1a9fc`, 0.001, 0.001, `\x3f50624dd2f1a9fc`, true}, {`\x3f50624dd2f1a9fd`, 0.0010000000000000002, 0.0010000000000000002, `\x3f50624dd2f1a9fd`, true}, {`\x3f847ae147ae147a`, 0.009999999999999998, 0.009999999999999998, `\x3f847ae147ae147a`, true}, {`\x3f847ae147ae147b`, 0.01, 0.01, `\x3f847ae147ae147b`, true}, {`\x3f847ae147ae147c`, 0.010000000000000002, 0.010000000000000002, `\x3f847ae147ae147c`, true}, {`\x3fb9999999999999`, 0.09999999999999999, 0.09999999999999999, `\x3fb9999999999999`, true}, {`\x3fb999999999999a`, 0.1, 0.1, `\x3fb999999999999a`, true}, {`\x3fb999999999999b`, 0.10000000000000002, 0.10000000000000002, `\x3fb999999999999b`, true}, {`\x3feffffffffffff0`, 0.9999999999999982, 0.9999999999999982, `\x3feffffffffffff0`, true}, {`\x3feffffffffffff1`, 0.9999999999999983, 0.9999999999999983, `\x3feffffffffffff1`, true}, {`\x3feffffffffffff2`, 0.9999999999999984, 0.9999999999999984, `\x3feffffffffffff2`, true}, {`\x3feffffffffffff3`, 0.9999999999999986, 0.9999999999999986, `\x3feffffffffffff3`, true}, {`\x3feffffffffffff4`, 0.9999999999999987, 0.9999999999999987, `\x3feffffffffffff4`, true}, {`\x3feffffffffffff5`, 0.9999999999999988, 0.9999999999999988, `\x3feffffffffffff5`, true}, {`\x3feffffffffffff6`, 0.9999999999999989, 0.9999999999999989, `\x3feffffffffffff6`, true}, {`\x3feffffffffffff7`, 0.999999999999999, 0.999999999999999, `\x3feffffffffffff7`, true}, {`\x3feffffffffffff8`, 0.9999999999999991, 0.9999999999999991, `\x3feffffffffffff8`, true}, {`\x3feffffffffffff9`, 0.9999999999999992, 0.9999999999999992, `\x3feffffffffffff9`, true}, {`\x3feffffffffffffa`, 0.9999999999999993, 0.9999999999999993, `\x3feffffffffffffa`, true}, {`\x3feffffffffffffb`, 0.9999999999999994, 0.9999999999999994, `\x3feffffffffffffb`, true}, {`\x3feffffffffffffc`, 0.9999999999999996, 0.9999999999999996, `\x3feffffffffffffc`, true}, {`\x3feffffffffffffd`, 0.9999999999999997, 0.9999999999999997, `\x3feffffffffffffd`, true}, {`\x3feffffffffffffe`, 0.9999999999999998, 0.9999999999999998, `\x3feffffffffffffe`, true}, {`\x3fefffffffffffff`, 0.9999999999999999, 0.9999999999999999, `\x3fefffffffffffff`, true}, {`\x3ff0000000000000`, 1, 1, `\x3ff0000000000000`, true}, {`\x3ff0000000000001`, 1.0000000000000002, 1.0000000000000002, `\x3ff0000000000001`, true}, {`\x3ff0000000000002`, 1.0000000000000004, 1.0000000000000004, `\x3ff0000000000002`, true}, {`\x3ff0000000000003`, 1.0000000000000007, 1.0000000000000007, `\x3ff0000000000003`, true}, {`\x3ff0000000000004`, 1.0000000000000009, 1.0000000000000009, `\x3ff0000000000004`, true}, {`\x3ff0000000000005`, 1.000000000000001, 1.000000000000001, `\x3ff0000000000005`, true}, {`\x3ff0000000000006`, 1.0000000000000013, 1.0000000000000013, `\x3ff0000000000006`, true}, {`\x3ff0000000000007`, 1.0000000000000016, 1.0000000000000016, `\x3ff0000000000007`, true}, {`\x3ff0000000000008`, 1.0000000000000018, 1.0000000000000018, `\x3ff0000000000008`, true}, {`\x3ff0000000000009`, 1.000000000000002, 1.000000000000002, `\x3ff0000000000009`, true}, {`\x3ff921fb54442d18`, 1.5707963267948966, 1.5707963267948966, `\x3ff921fb54442d18`, true}, {`\x4005bf0a8b14576a`, 2.7182818284590455, 2.7182818284590455, `\x4005bf0a8b14576a`, true}, {`\x400921fb54442d18`, 3.141592653589793, 3.141592653589793, `\x400921fb54442d18`, true}, {`\x4023ffffffffffff`, 9.999999999999998, 9.999999999999998, `\x4023ffffffffffff`, true}, {`\x4024000000000000`, 10, 10, `\x4024000000000000`, true}, {`\x4024000000000001`, 10.000000000000002, 10.000000000000002, `\x4024000000000001`, true}, {`\x4058ffffffffffff`, 99.99999999999999, 99.99999999999999, `\x4058ffffffffffff`, true}, {`\x4059000000000000`, 100, 100, `\x4059000000000000`, true}, {`\x4059000000000001`, 100.00000000000001, 100.00000000000001, `\x4059000000000001`, true}, {`\x408f3fffffffffff`, 999.9999999999999, 999.9999999999999, `\x408f3fffffffffff`, true}, {`\x408f400000000000`, 1000, 1000, `\x408f400000000000`, true}, {`\x408f400000000001`, 1000.0000000000001, 1000.0000000000001, `\x408f400000000001`, true}, {`\x40c387ffffffffff`, 9999.999999999998, 9999.999999999998, `\x40c387ffffffffff`, true}, {`\x40c3880000000000`, 10000, 10000, `\x40c3880000000000`, true}, {`\x40c3880000000001`, 10000.000000000002, 10000.000000000002, `\x40c3880000000001`, true}, {`\x40f869ffffffffff`, 99999.99999999999, 99999.99999999999, `\x40f869ffffffffff`, true}, {`\x40f86a0000000000`, 100000, 100000, `\x40f86a0000000000`, true}, {`\x40f86a0000000001`, 100000.00000000001, 100000.00000000001, `\x40f86a0000000001`, true}, {`\x412e847fffffffff`, 999999.9999999999, 999999.9999999999, `\x412e847fffffffff`, true}, {`\x412e848000000000`, 1000000, 1000000, `\x412e848000000000`, true}, {`\x412e848000000001`, 1000000.0000000001, 1000000.0000000001, `\x412e848000000001`, true}, {`\x416312cfffffffff`, 9999999.999999998, 9999999.999999998, `\x416312cfffffffff`, true}, {`\x416312d000000000`, 10000000, 10000000, `\x416312d000000000`, true}, {`\x416312d000000001`, 10000000.000000002, 10000000.000000002, `\x416312d000000001`, true}, {`\x4197d783ffffffff`, 99999999.99999999, 99999999.99999999, `\x4197d783ffffffff`, true}, {`\x4197d78400000000`, 100000000, 100000000, `\x4197d78400000000`, true}, {`\x4197d78400000001`, 100000000.00000001, 100000000.00000001, `\x4197d78400000001`, true}, {`\x41cdcd64ffffffff`, 999999999.9999999, 999999999.9999999, `\x41cdcd64ffffffff`, true}, {`\x41cdcd6500000000`, 1000000000, 1000000000, `\x41cdcd6500000000`, true}, {`\x41cdcd6500000001`, 1000000000.0000001, 1000000000.0000001, `\x41cdcd6500000001`, true}, {`\x4202a05f1fffffff`, 9999999999.999998, 9999999999.999998, `\x4202a05f1fffffff`, true}, {`\x4202a05f20000000`, 10000000000, 10000000000, `\x4202a05f20000000`, true}, {`\x4202a05f20000001`, 10000000000.000002, 10000000000.000002, `\x4202a05f20000001`, true}, {`\x42374876e7ffffff`, 99999999999.99998, 99999999999.99998, `\x42374876e7ffffff`, true}, {`\x42374876e8000000`, 100000000000, 100000000000, `\x42374876e8000000`, true}, {`\x42374876e8000001`, 100000000000.00002, 100000000000.00002, `\x42374876e8000001`, true}, {`\x426d1a94a1ffffff`, 999999999999.9999, 999999999999.9999, `\x426d1a94a1ffffff`, true}, {`\x426d1a94a2000000`, 1000000000000, 1000000000000, `\x426d1a94a2000000`, true}, {`\x426d1a94a2000001`, 1000000000000.0001, 1000000000000.0001, `\x426d1a94a2000001`, true}, {`\x42a2309ce53fffff`, 9999999999999.998, 9999999999999.998, `\x42a2309ce53fffff`, true}, {`\x42a2309ce5400000`, 10000000000000, 10000000000000, `\x42a2309ce5400000`, true}, {`\x42a2309ce5400001`, 10000000000000.002, 10000000000000.002, `\x42a2309ce5400001`, true}, {`\x42d6bcc41e8fffff`, 99999999999999.98, 99999999999999.98, `\x42d6bcc41e8fffff`, true}, {`\x42d6bcc41e900000`, 100000000000000, 100000000000000, `\x42d6bcc41e900000`, true}, {`\x42d6bcc41e900001`, 100000000000000.02, 100000000000000.02, `\x42d6bcc41e900001`, true}, {`\x430c6bf52633ffff`, 999999999999999.9, 999999999999999.9, `\x430c6bf52633ffff`, true}, {`\x430c6bf526340000`, 1e+15, 1e+15, `\x430c6bf526340000`, true}, {`\x430c6bf526340001`, 1.0000000000000001e+15, 1.0000000000000001e+15, `\x430c6bf526340001`, true}, {`\x4341c37937e07fff`, 9.999999999999998e+15, 9.999999999999998e+15, `\x4341c37937e07fff`, true}, {`\x4341c37937e08000`, 1e+16, 1e+16, `\x4341c37937e08000`, true}, {`\x4341c37937e08001`, 1.0000000000000002e+16, 1.0000000000000002e+16, `\x4341c37937e08001`, true}, {`\x4376345785d89fff`, 9.999999999999998e+16, 9.999999999999998e+16, `\x4376345785d89fff`, true}, {`\x4376345785d8a000`, 1e+17, 1e+17, `\x4376345785d8a000`, true}, {`\x4376345785d8a001`, 1.0000000000000002e+17, 1.0000000000000002e+17, `\x4376345785d8a001`, true}, {`\x43abc16d674ec7ff`, 9.999999999999999e+17, 9.999999999999999e+17, `\x43abc16d674ec7ff`, true}, {`\x43abc16d674ec800`, 1e+18, 1e+18, `\x43abc16d674ec800`, true}, {`\x43abc16d674ec801`, 1.0000000000000001e+18, 1.0000000000000001e+18, `\x43abc16d674ec801`, true}, {`\x43e158e460913cff`, 9.999999999999998e+18, 9.999999999999998e+18, `\x43e158e460913cff`, true}, {`\x43e158e460913d00`, 1e+19, 1e+19, `\x43e158e460913d00`, true}, {`\x43e158e460913d01`, 1.0000000000000002e+19, 1.0000000000000002e+19, `\x43e158e460913d01`, true}, {`\x4415af1d78b58c3f`, 9.999999999999998e+19, 9.999999999999998e+19, `\x4415af1d78b58c3f`, true}, {`\x4415af1d78b58c40`, 1e+20, 1e+20, `\x4415af1d78b58c40`, true}, {`\x4415af1d78b58c41`, 1.0000000000000002e+20, 1.0000000000000002e+20, `\x4415af1d78b58c41`, true}, {`\x444b1ae4d6e2ef4f`, 9.999999999999999e+20, 9.999999999999999e+20, `\x444b1ae4d6e2ef4f`, true}, {`\x444b1ae4d6e2ef50`, 1e+21, 1e+21, `\x444b1ae4d6e2ef50`, true}, {`\x444b1ae4d6e2ef51`, 1.0000000000000001e+21, 1.0000000000000001e+21, `\x444b1ae4d6e2ef51`, true}, {`\x4480f0cf064dd591`, 9.999999999999998e+21, 9.999999999999998e+21, `\x4480f0cf064dd591`, true}, {`\x4480f0cf064dd592`, 1e+22, 1e+22, `\x4480f0cf064dd592`, true}, {`\x4480f0cf064dd593`, 1.0000000000000002e+22, 1.0000000000000002e+22, `\x4480f0cf064dd593`, true}, {`\x44b52d02c7e14af5`, 9.999999999999997e+22, 9.999999999999997e+22, `\x44b52d02c7e14af5`, true}, {`\x44b52d02c7e14af6`, 9.999999999999999e+22, 9.999999999999999e+22, `\x44b52d02c7e14af6`, true}, {`\x44b52d02c7e14af7`, 1.0000000000000001e+23, 1.0000000000000001e+23, `\x44b52d02c7e14af7`, true}, {`\x44ea784379d99db3`, 9.999999999999998e+23, 9.999999999999998e+23, `\x44ea784379d99db3`, true}, {`\x44ea784379d99db4`, 1e+24, 1e+24, `\x44ea784379d99db4`, true}, {`\x44ea784379d99db5`, 1.0000000000000001e+24, 1.0000000000000001e+24, `\x44ea784379d99db5`, true}, {`\x45208b2a2c280290`, 9.999999999999999e+24, 9.999999999999999e+24, `\x45208b2a2c280290`, true}, {`\x45208b2a2c280291`, 1e+25, 1e+25, `\x45208b2a2c280291`, true}, {`\x45208b2a2c280292`, 1.0000000000000003e+25, 1.0000000000000003e+25, `\x45208b2a2c280292`, true}, {`\x7feffffffffffffe`, 1.7976931348623155e+308, 1.7976931348623155e+308, `\x7feffffffffffffe`, true}, {`\x7fefffffffffffff`, 1.7976931348623157e+308, 1.7976931348623157e+308, `\x7fefffffffffffff`, true}, {`\x4350000000000002`, 1.8014398509481992e+16, 1.8014398509481992e+16, `\x4350000000000002`, true}, {`\x4350000000002e06`, 1.8014398509529112e+16, 1.8014398509529112e+16, `\x4350000000002e06`, true}, {`\x4352000000000003`, 2.0266198323167244e+16, 2.0266198323167244e+16, `\x4352000000000003`, true}, {`\x4352000000000004`, 2.0266198323167248e+16, 2.0266198323167248e+16, `\x4352000000000004`, true}, {`\x4358000000000003`, 2.7021597764222988e+16, 2.7021597764222988e+16, `\x4358000000000003`, true}, {`\x4358000000000004`, 2.7021597764222992e+16, 2.7021597764222992e+16, `\x4358000000000004`, true}, {`\x435f000000000020`, 3.4902897112121472e+16, 3.4902897112121472e+16, `\x435f000000000020`, true}, {`\xc350000000000002`, -1.8014398509481992e+16, -1.8014398509481992e+16, `\xc350000000000002`, true}, {`\xc350000000002e06`, -1.8014398509529112e+16, -1.8014398509529112e+16, `\xc350000000002e06`, true}, {`\xc352000000000003`, -2.0266198323167244e+16, -2.0266198323167244e+16, `\xc352000000000003`, true}, {`\xc352000000000004`, -2.0266198323167248e+16, -2.0266198323167248e+16, `\xc352000000000004`, true}, {`\xc358000000000003`, -2.7021597764222988e+16, -2.7021597764222988e+16, `\xc358000000000003`, true}, {`\xc358000000000004`, -2.7021597764222992e+16, -2.7021597764222992e+16, `\xc358000000000004`, true}, {`\xc35f000000000020`, -3.4902897112121472e+16, -3.4902897112121472e+16, `\xc35f000000000020`, true}, {`\x42dc12218377de66`, 123456789012345.6, 123456789012345.6, `\x42dc12218377de66`, true}, {`\x42a674e79c5fe51f`, 12345678901234.56, 12345678901234.56, `\x42a674e79c5fe51f`, true}, {`\x4271f71fb04cb74c`, 1234567890123.456, 1234567890123.456, `\x4271f71fb04cb74c`, true}, {`\x423cbe991a145879`, 123456789012.3456, 123456789012.3456, `\x423cbe991a145879`, true}, {`\x4206fee0e1a9e061`, 12345678901.23456, 12345678901.23456, `\x4206fee0e1a9e061`, true}, {`\x41d26580b487e6b4`, 1234567890.123456, 1234567890.123456, `\x41d26580b487e6b4`, true}, {`\x419d6f34540ca453`, 123456789.0123456, 123456789.0123456, `\x419d6f34540ca453`, true}, {`\x41678c29dcd6e9dc`, 12345678.90123456, 12345678.90123456, `\x41678c29dcd6e9dc`, true}, {`\x4132d687e3df217d`, 1234567.890123456, 1234567.890123456, `\x4132d687e3df217d`, true}, {`\x40fe240c9fcb68c8`, 123456.7890123456, 123456.7890123456, `\x40fe240c9fcb68c8`, true}, {`\x40c81cd6e63c53d3`, 12345.67890123456, 12345.67890123456, `\x40c81cd6e63c53d3`, true}, {`\x40934a4584fd0fdc`, 1234.567890123456, 1234.567890123456, `\x40934a4584fd0fdc`, true}, {`\x405edd3c07fb4c93`, 123.4567890123456, 123.4567890123456, `\x405edd3c07fb4c93`, true}, {`\x4028b0fcd32f7076`, 12.34567890123456, 12.34567890123456, `\x4028b0fcd32f7076`, true}, {`\x3ff3c0ca428c59f8`, 1.234567890123456, 1.234567890123456, `\x3ff3c0ca428c59f8`, true}, {`\x3e60000000000000`, 2.9802322387695312e-08, 2.9802322387695312e-08, `\x3e60000000000000`, true}, {`\xc352bd2668e077c4`, -2.1098088986959632e+16, -2.1098088986959632e+16, `\xc352bd2668e077c4`, true}, {`\x434018601510c000`, 9.0608011534336e+15, 9.0608011534336e+15, `\x434018601510c000`, true}, {`\x43d055dc36f24000`, 4.708356024711512e+18, 4.708356024711512e+18, `\x43d055dc36f24000`, true}, {`\x43e052961c6f8000`, 9.409340012568248e+18, 9.409340012568248e+18, `\x43e052961c6f8000`, true}, {`\x3ff3c0ca2a5b1d5d`, 1.2345678, 1.2345678, `\x3ff3c0ca2a5b1d5d`, true}, {`\x4830f0cf064dd592`, 5.764607523034235e+39, 5.764607523034235e+39, `\x4830f0cf064dd592`, true}, {`\x4840f0cf064dd592`, 1.152921504606847e+40, 1.152921504606847e+40, `\x4840f0cf064dd592`, true}, {`\x4850f0cf064dd592`, 2.305843009213694e+40, 2.305843009213694e+40, `\x4850f0cf064dd592`, true}, {`\x3ff3333333333333`, 1.2, 1.2, `\x3ff3333333333333`, true}, {`\x3ff3ae147ae147ae`, 1.23, 1.23, `\x3ff3ae147ae147ae`, true}, {`\x3ff3be76c8b43958`, 1.234, 1.234, `\x3ff3be76c8b43958`, true}, {`\x3ff3c083126e978d`, 1.2345, 1.2345, `\x3ff3c083126e978d`, true}, {`\x3ff3c0c1fc8f3238`, 1.23456, 1.23456, `\x3ff3c0c1fc8f3238`, true}, {`\x3ff3c0c9539b8887`, 1.234567, 1.234567, `\x3ff3c0c9539b8887`, true}, {`\x3ff3c0ca2a5b1d5d`, 1.2345678, 1.2345678, `\x3ff3c0ca2a5b1d5d`, true}, {`\x3ff3c0ca4283de1b`, 1.23456789, 1.23456789, `\x3ff3c0ca4283de1b`, true}, {`\x3ff3c0ca43db770a`, 1.234567895, 1.234567895, `\x3ff3c0ca43db770a`, true}, {`\x3ff3c0ca428abd53`, 1.2345678901, 1.2345678901, `\x3ff3c0ca428abd53`, true}, {`\x3ff3c0ca428c1d2b`, 1.23456789012, 1.23456789012, `\x3ff3c0ca428c1d2b`, true}, {`\x3ff3c0ca428c51f2`, 1.234567890123, 1.234567890123, `\x3ff3c0ca428c51f2`, true}, {`\x3ff3c0ca428c58fc`, 1.2345678901234, 1.2345678901234, `\x3ff3c0ca428c58fc`, true}, {`\x3ff3c0ca428c59dd`, 1.23456789012345, 1.23456789012345, `\x3ff3c0ca428c59dd`, true}, {`\x3ff3c0ca428c59f8`, 1.234567890123456, 1.234567890123456, `\x3ff3c0ca428c59f8`, true}, {`\x3ff3c0ca428c59fb`, 1.2345678901234567, 1.2345678901234567, `\x3ff3c0ca428c59fb`, true}, {`\x40112e0be8047a7d`, 4.294967294, 4.294967294, `\x40112e0be8047a7d`, true}, {`\x40112e0be815a889`, 4.294967295, 4.294967295, `\x40112e0be815a889`, true}, {`\x40112e0be826d695`, 4.294967296, 4.294967296, `\x40112e0be826d695`, true}, {`\x40112e0be83804a1`, 4.294967297, 4.294967297, `\x40112e0be83804a1`, true}, {`\x40112e0be84932ad`, 4.294967298, 4.294967298, `\x40112e0be84932ad`, true}, {`\x0040000000000000`, 1.7800590868057611e-307, 1.7800590868057611e-307, `\x0040000000000000`, true}, {`\x007fffffffffffff`, 2.8480945388892175e-306, 2.8480945388892175e-306, `\x007fffffffffffff`, true}, {`\x0290000000000000`, 2.446494580089078e-296, 2.446494580089078e-296, `\x0290000000000000`, true}, {`\x029fffffffffffff`, 4.8929891601781557e-296, 4.8929891601781557e-296, `\x029fffffffffffff`, true}, {`\x4350000000000000`, 1.8014398509481984e+16, 1.8014398509481984e+16, `\x4350000000000000`, true}, {`\x435fffffffffffff`, 3.6028797018963964e+16, 3.6028797018963964e+16, `\x435fffffffffffff`, true}, {`\x1330000000000000`, 2.900835519859558e-216, 2.900835519859558e-216, `\x1330000000000000`, true}, {`\x133fffffffffffff`, 5.801671039719115e-216, 5.801671039719115e-216, `\x133fffffffffffff`, true}, {`\x3a6fa7161a4d6e0c`, 3.196104012172126e-27, 3.196104012172126e-27, `\x3a6fa7161a4d6e0c`, true}},
			},
			{
				Statement: `drop type xfloat8 cascade;`,
			},
			{
				Statement: `DETAIL:  drop cascades to function xfloat8in(cstring)
drop cascades to function xfloat8out(xfloat8)
drop cascades to cast from xfloat8 to double precision
drop cascades to cast from double precision to xfloat8
drop cascades to cast from xfloat8 to bigint
drop cascades to cast from bigint to xfloat8`,
			},
		},
	})
}
