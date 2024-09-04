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

func TestFloat4(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_float4)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_float4,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE FLOAT4_TBL (f1  float4);`,
			},
			{
				Statement: `INSERT INTO FLOAT4_TBL(f1) VALUES ('    0.0');`,
			},
			{
				Statement: `INSERT INTO FLOAT4_TBL(f1) VALUES ('1004.30   ');`,
			},
			{
				Statement: `INSERT INTO FLOAT4_TBL(f1) VALUES ('     -34.84    ');`,
			},
			{
				Statement: `INSERT INTO FLOAT4_TBL(f1) VALUES ('1.2345678901234e+20');`,
			},
			{
				Statement: `INSERT INTO FLOAT4_TBL(f1) VALUES ('1.2345678901234e-20');`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('10e70');`,
				ErrorString: `"10e70" is out of range for type real`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('-10e70');`,
				ErrorString: `"-10e70" is out of range for type real`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('10e-70');`,
				ErrorString: `"10e-70" is out of range for type real`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('-10e-70');`,
				ErrorString: `"-10e-70" is out of range for type real`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('10e70'::float8);`,
				ErrorString: `value out of range: overflow`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('-10e70'::float8);`,
				ErrorString: `value out of range: overflow`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('10e-70'::float8);`,
				ErrorString: `value out of range: underflow`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('-10e-70'::float8);`,
				ErrorString: `value out of range: underflow`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('10e400');`,
				ErrorString: `"10e400" is out of range for type real`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('-10e400');`,
				ErrorString: `"-10e400" is out of range for type real`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('10e-400');`,
				ErrorString: `"10e-400" is out of range for type real`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('-10e-400');`,
				ErrorString: `"-10e-400" is out of range for type real`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('');`,
				ErrorString: `invalid input syntax for type real: ""`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('       ');`,
				ErrorString: `invalid input syntax for type real: "       "`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('xyz');`,
				ErrorString: `invalid input syntax for type real: "xyz"`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('5.0.0');`,
				ErrorString: `invalid input syntax for type real: "5.0.0"`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('5 . 0');`,
				ErrorString: `invalid input syntax for type real: "5 . 0"`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('5.   0');`,
				ErrorString: `invalid input syntax for type real: "5.   0"`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('     - 3.0');`,
				ErrorString: `invalid input syntax for type real: "     - 3.0"`,
			},
			{
				Statement:   `INSERT INTO FLOAT4_TBL(f1) VALUES ('123            5');`,
				ErrorString: `invalid input syntax for type real: "123            5"`,
			},
			{
				Statement: `SELECT 'NaN'::float4;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT 'nan'::float4;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT '   NAN  '::float4;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT 'infinity'::float4;`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT '          -INFINiTY   '::float4;`,
				Results:   []sql.Row{{`-Infinity`}},
			},
			{
				Statement:   `SELECT 'N A N'::float4;`,
				ErrorString: `invalid input syntax for type real: "N A N"`,
			},
			{
				Statement:   `SELECT 'NaN x'::float4;`,
				ErrorString: `invalid input syntax for type real: "NaN x"`,
			},
			{
				Statement:   `SELECT ' INFINITY    x'::float4;`,
				ErrorString: `invalid input syntax for type real: " INFINITY    x"`,
			},
			{
				Statement: `SELECT 'Infinity'::float4 + 100.0;`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT 'Infinity'::float4 / 'Infinity'::float4;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT '42'::float4 / 'Infinity'::float4;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT 'nan'::float4 / 'nan'::float4;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT 'nan'::float4 / '0'::float4;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT 'nan'::numeric::float4;`,
				Results:   []sql.Row{{`NaN`}},
			},
			{
				Statement: `SELECT * FROM FLOAT4_TBL;`,
				Results:   []sql.Row{{0}, {1004.3}, {-34.84}, {1.2345679e+20}, {1.2345679e-20}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT4_TBL f WHERE f.f1 <> '1004.3';`,
				Results:   []sql.Row{{0}, {-34.84}, {1.2345679e+20}, {1.2345679e-20}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT4_TBL f WHERE f.f1 = '1004.3';`,
				Results:   []sql.Row{{1004.3}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT4_TBL f WHERE '1004.3' > f.f1;`,
				Results:   []sql.Row{{0}, {-34.84}, {1.2345679e-20}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT4_TBL f WHERE  f.f1 < '1004.3';`,
				Results:   []sql.Row{{0}, {-34.84}, {1.2345679e-20}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT4_TBL f WHERE '1004.3' >= f.f1;`,
				Results:   []sql.Row{{0}, {1004.3}, {-34.84}, {1.2345679e-20}},
			},
			{
				Statement: `SELECT f.* FROM FLOAT4_TBL f WHERE  f.f1 <= '1004.3';`,
				Results:   []sql.Row{{0}, {1004.3}, {-34.84}, {1.2345679e-20}},
			},
			{
				Statement: `SELECT f.f1, f.f1 * '-10' AS x FROM FLOAT4_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, -10043}, {1.2345679e+20, -1.2345678e+21}, {1.2345679e-20, -1.2345678e-19}},
			},
			{
				Statement: `SELECT f.f1, f.f1 + '-10' AS x FROM FLOAT4_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, 994.3}, {1.2345679e+20, 1.2345679e+20}, {1.2345679e-20, -10}},
			},
			{
				Statement: `SELECT f.f1, f.f1 / '-10' AS x FROM FLOAT4_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, -100.43}, {1.2345679e+20, -1.2345679e+19}, {1.2345679e-20, -1.2345679e-21}},
			},
			{
				Statement: `SELECT f.f1, f.f1 - '-10' AS x FROM FLOAT4_TBL f
   WHERE f.f1 > '0.0';`,
				Results: []sql.Row{{1004.3, 1014.3}, {1.2345679e+20, 1.2345679e+20}, {1.2345679e-20, 10}},
			},
			{
				Statement:   `SELECT f.f1 / '0.0' from FLOAT4_TBL f;`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `SELECT * FROM FLOAT4_TBL;`,
				Results:   []sql.Row{{0}, {1004.3}, {-34.84}, {1.2345679e+20}, {1.2345679e-20}},
			},
			{
				Statement: `SELECT f.f1, @f.f1 AS abs_f1 FROM FLOAT4_TBL f;`,
				Results:   []sql.Row{{0, 0}, {1004.3, 1004.3}, {-34.84, 34.84}, {1.2345679e+20, 1.2345679e+20}, {1.2345679e-20, 1.2345679e-20}},
			},
			{
				Statement: `UPDATE FLOAT4_TBL
   SET f1 = FLOAT4_TBL.f1 * '-1'
   WHERE FLOAT4_TBL.f1 > '0.0';`,
			},
			{
				Statement: `SELECT * FROM FLOAT4_TBL;`,
				Results:   []sql.Row{{0}, {-34.84}, {-1004.3}, {-1.2345679e+20}, {-1.2345679e-20}},
			},
			{
				Statement: `SELECT '32767.4'::float4::int2;`,
				Results:   []sql.Row{{32767}},
			},
			{
				Statement:   `SELECT '32767.6'::float4::int2;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT '-32768.4'::float4::int2;`,
				Results:   []sql.Row{{-32768}},
			},
			{
				Statement:   `SELECT '-32768.6'::float4::int2;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT '2147483520'::float4::int4;`,
				Results:   []sql.Row{{2147483520}},
			},
			{
				Statement:   `SELECT '2147483647'::float4::int4;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT '-2147483648.5'::float4::int4;`,
				Results:   []sql.Row{{-2147483648}},
			},
			{
				Statement:   `SELECT '-2147483900'::float4::int4;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT '9223369837831520256'::float4::int8;`,
				Results:   []sql.Row{{9223369837831520256}},
			},
			{
				Statement:   `SELECT '9223372036854775807'::float4::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT '-9223372036854775808.5'::float4::int8;`,
				Results:   []sql.Row{{-9223372036854775808}},
			},
			{
				Statement:   `SELECT '-9223380000000000000'::float4::int8;`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT float4send('5e-20'::float4);`,
				Results:   []sql.Row{{`\x1f6c1e4a`}},
			},
			{
				Statement: `SELECT float4send('67e14'::float4);`,
				Results:   []sql.Row{{`\x59be6cea`}},
			},
			{
				Statement: `SELECT float4send('985e15'::float4);`,
				Results:   []sql.Row{{`\x5d5ab6c4`}},
			},
			{
				Statement: `SELECT float4send('55895e-16'::float4);`,
				Results:   []sql.Row{{`\x2cc4a9bd`}},
			},
			{
				Statement: `SELECT float4send('7038531e-32'::float4);`,
				Results:   []sql.Row{{`\x15ae43fd`}},
			},
			{
				Statement: `SELECT float4send('702990899e-20'::float4);`,
				Results:   []sql.Row{{`\x2cf757ca`}},
			},
			{
				Statement: `SELECT float4send('3e-23'::float4);`,
				Results:   []sql.Row{{`\x1a111234`}},
			},
			{
				Statement: `SELECT float4send('57e18'::float4);`,
				Results:   []sql.Row{{`\x6045c22c`}},
			},
			{
				Statement: `SELECT float4send('789e-35'::float4);`,
				Results:   []sql.Row{{`\x0a23de70`}},
			},
			{
				Statement: `SELECT float4send('2539e-18'::float4);`,
				Results:   []sql.Row{{`\x2736f449`}},
			},
			{
				Statement: `SELECT float4send('76173e28'::float4);`,
				Results:   []sql.Row{{`\x7616398a`}},
			},
			{
				Statement: `SELECT float4send('887745e-11'::float4);`,
				Results:   []sql.Row{{`\x3714f05c`}},
			},
			{
				Statement: `SELECT float4send('5382571e-37'::float4);`,
				Results:   []sql.Row{{`\x0d2eaca7`}},
			},
			{
				Statement: `SELECT float4send('82381273e-35'::float4);`,
				Results:   []sql.Row{{`\x128289d1`}},
			},
			{
				Statement: `SELECT float4send('750486563e-38'::float4);`,
				Results:   []sql.Row{{`\x0f18377e`}},
			},
			{
				Statement: `SELECT float4send('1.17549435e-38'::float4);`,
				Results:   []sql.Row{{`\x00800000`}},
			},
			{
				Statement: `SELECT float4send('1.1754944e-38'::float4);`,
				Results:   []sql.Row{{`\x00800000`}},
			},
			{
				Statement: `create type xfloat4;`,
			},
			{
				Statement: `create function xfloat4in(cstring) returns xfloat4 immutable strict
  language internal as 'int4in';`,
			},
			{
				Statement: `create function xfloat4out(xfloat4) returns cstring immutable strict
  language internal as 'int4out';`,
			},
			{
				Statement: `create type xfloat4 (input = xfloat4in, output = xfloat4out, like = float4);`,
			},
			{
				Statement: `create cast (xfloat4 as float4) without function;`,
			},
			{
				Statement: `create cast (float4 as xfloat4) without function;`,
			},
			{
				Statement: `create cast (xfloat4 as integer) without function;`,
			},
			{
				Statement: `create cast (integer as xfloat4) without function;`,
			},
			{
				Statement: `with testdata(bits) as (values
  -- small subnormals
  (x'00000001'),
  (x'00000002'), (x'00000003'),
  (x'00000010'), (x'00000011'), (x'00000100'), (x'00000101'),
  (x'00004000'), (x'00004001'), (x'00080000'), (x'00080001'),
  -- stress values
  (x'0053c4f4'),  -- 7693e-42
  (x'006c85c4'),  -- 996622e-44
  (x'0041ca76'),  -- 60419369e-46
  (x'004b7678'),  -- 6930161142e-48
  -- taken from upstream testsuite
  (x'00000007'),
  (x'00424fe2'),
  -- borderline between subnormal and normal
  (x'007ffff0'), (x'007ffff1'), (x'007ffffe'), (x'007fffff'))
select float4send(flt) as ibits,
       flt
  from (select bits::integer::xfloat4::float4 as flt
          from testdata
	offset 0) s;`,
				Results: []sql.Row{{`\x00000001`, 1e-45}, {`\x00000002`, 3e-45}, {`\x00000003`, 4e-45}, {`\x00000010`, 2.2e-44}, {`\x00000011`, 2.4e-44}, {`\x00000100`, 3.59e-43}, {`\x00000101`, 3.6e-43}, {`\x00004000`, 2.2959e-41}, {`\x00004001`, 2.296e-41}, {`\x00080000`, 7.34684e-40}, {`\x00080001`, 7.34685e-40}, {`\x0053c4f4`, 7.693e-39}, {`\x006c85c4`, 9.96622e-39}, {`\x0041ca76`, 6.041937e-39}, {`\x004b7678`, 6.930161e-39}, {`\x00000007`, 1e-44}, {`\x00424fe2`, 6.0898e-39}, {`\x007ffff0`, 1.1754921e-38}, {`\x007ffff1`, 1.1754922e-38}, {`\x007ffffe`, 1.1754941e-38}, {`\x007fffff`, 1.1754942e-38}},
			},
			{
				Statement: `with testdata(bits) as (values
  (x'00000000'),
  -- smallest normal values
  (x'00800000'), (x'00800001'), (x'00800004'), (x'00800005'),
  (x'00800006'),
  -- small normal values chosen for short vs. long output
  (x'008002f1'), (x'008002f2'), (x'008002f3'),
  (x'00800e17'), (x'00800e18'), (x'00800e19'),
  -- assorted values (random mantissae)
  (x'01000001'), (x'01102843'), (x'01a52c98'),
  (x'0219c229'), (x'02e4464d'), (x'037343c1'), (x'03a91b36'),
  (x'047ada65'), (x'0496fe87'), (x'0550844f'), (x'05999da3'),
  (x'060ea5e2'), (x'06e63c45'), (x'07f1e548'), (x'0fc5282b'),
  (x'1f850283'), (x'2874a9d6'),
  -- values around 5e-08
  (x'3356bf94'), (x'3356bf95'), (x'3356bf96'),
  -- around 1e-07
  (x'33d6bf94'), (x'33d6bf95'), (x'33d6bf96'),
  -- around 3e-07 .. 1e-04
  (x'34a10faf'), (x'34a10fb0'), (x'34a10fb1'),
  (x'350637bc'), (x'350637bd'), (x'350637be'),
  (x'35719786'), (x'35719787'), (x'35719788'),
  (x'358637bc'), (x'358637bd'), (x'358637be'),
  (x'36a7c5ab'), (x'36a7c5ac'), (x'36a7c5ad'),
  (x'3727c5ab'), (x'3727c5ac'), (x'3727c5ad'),
  -- format crossover at 1e-04
  (x'38d1b714'), (x'38d1b715'), (x'38d1b716'),
  (x'38d1b717'), (x'38d1b718'), (x'38d1b719'),
  (x'38d1b71a'), (x'38d1b71b'), (x'38d1b71c'),
  (x'38d1b71d'),
  --
  (x'38dffffe'), (x'38dfffff'), (x'38e00000'),
  (x'38efffff'), (x'38f00000'), (x'38f00001'),
  (x'3a83126e'), (x'3a83126f'), (x'3a831270'),
  (x'3c23d709'), (x'3c23d70a'), (x'3c23d70b'),
  (x'3dcccccc'), (x'3dcccccd'), (x'3dccccce'),
  -- chosen to need 9 digits for 3dcccd70
  (x'3dcccd6f'), (x'3dcccd70'), (x'3dcccd71'),
  --
  (x'3effffff'), (x'3f000000'), (x'3f000001'),
  (x'3f333332'), (x'3f333333'), (x'3f333334'),
  -- approach 1.0 with increasing numbers of 9s
  (x'3f666665'), (x'3f666666'), (x'3f666667'),
  (x'3f7d70a3'), (x'3f7d70a4'), (x'3f7d70a5'),
  (x'3f7fbe76'), (x'3f7fbe77'), (x'3f7fbe78'),
  (x'3f7ff971'), (x'3f7ff972'), (x'3f7ff973'),
  (x'3f7fff57'), (x'3f7fff58'), (x'3f7fff59'),
  (x'3f7fffee'), (x'3f7fffef'),
  -- values very close to 1
  (x'3f7ffff0'), (x'3f7ffff1'), (x'3f7ffff2'),
  (x'3f7ffff3'), (x'3f7ffff4'), (x'3f7ffff5'),
  (x'3f7ffff6'), (x'3f7ffff7'), (x'3f7ffff8'),
  (x'3f7ffff9'), (x'3f7ffffa'), (x'3f7ffffb'),
  (x'3f7ffffc'), (x'3f7ffffd'), (x'3f7ffffe'),
  (x'3f7fffff'),
  (x'3f800000'),
  (x'3f800001'), (x'3f800002'), (x'3f800003'),
  (x'3f800004'), (x'3f800005'), (x'3f800006'),
  (x'3f800007'), (x'3f800008'), (x'3f800009'),
  -- values 1 to 1.1
  (x'3f80000f'), (x'3f800010'), (x'3f800011'),
  (x'3f800012'), (x'3f800013'), (x'3f800014'),
  (x'3f800017'), (x'3f800018'), (x'3f800019'),
  (x'3f80001a'), (x'3f80001b'), (x'3f80001c'),
  (x'3f800029'), (x'3f80002a'), (x'3f80002b'),
  (x'3f800053'), (x'3f800054'), (x'3f800055'),
  (x'3f800346'), (x'3f800347'), (x'3f800348'),
  (x'3f8020c4'), (x'3f8020c5'), (x'3f8020c6'),
  (x'3f8147ad'), (x'3f8147ae'), (x'3f8147af'),
  (x'3f8ccccc'), (x'3f8ccccd'), (x'3f8cccce'),
  --
  (x'3fc90fdb'), -- pi/2
  (x'402df854'), -- e
  (x'40490fdb'), -- pi
  --
  (x'409fffff'), (x'40a00000'), (x'40a00001'),
  (x'40afffff'), (x'40b00000'), (x'40b00001'),
  (x'411fffff'), (x'41200000'), (x'41200001'),
  (x'42c7ffff'), (x'42c80000'), (x'42c80001'),
  (x'4479ffff'), (x'447a0000'), (x'447a0001'),
  (x'461c3fff'), (x'461c4000'), (x'461c4001'),
  (x'47c34fff'), (x'47c35000'), (x'47c35001'),
  (x'497423ff'), (x'49742400'), (x'49742401'),
  (x'4b18967f'), (x'4b189680'), (x'4b189681'),
  (x'4cbebc1f'), (x'4cbebc20'), (x'4cbebc21'),
  (x'4e6e6b27'), (x'4e6e6b28'), (x'4e6e6b29'),
  (x'501502f8'), (x'501502f9'), (x'501502fa'),
  (x'51ba43b6'), (x'51ba43b7'), (x'51ba43b8'),
  -- stress values
  (x'1f6c1e4a'),  -- 5e-20
  (x'59be6cea'),  -- 67e14
  (x'5d5ab6c4'),  -- 985e15
  (x'2cc4a9bd'),  -- 55895e-16
  (x'15ae43fd'),  -- 7038531e-32
  (x'2cf757ca'),  -- 702990899e-20
  (x'665ba998'),  -- 25933168707e13
  (x'743c3324'),  -- 596428896559e20
  -- exercise fixed-point memmoves
  (x'47f1205a'),
  (x'4640e6ae'),
  (x'449a5225'),
  (x'42f6e9d5'),
  (x'414587dd'),
  (x'3f9e064b'),
  -- these cases come from the upstream's testsuite
  -- BoundaryRoundEven
  (x'4c000004'),
  (x'50061c46'),
  (x'510006a8'),
  -- ExactValueRoundEven
  (x'48951f84'),
  (x'45fd1840'),
  -- LotsOfTrailingZeros
  (x'39800000'),
  (x'3b200000'),
  (x'3b900000'),
  (x'3bd00000'),
  -- Regression
  (x'63800000'),
  (x'4b000000'),
  (x'4b800000'),
  (x'4c000001'),
  (x'4c800b0d'),
  (x'00d24584'),
  (x'00d90b88'),
  (x'45803f34'),
  (x'4f9f24f7'),
  (x'3a8722c3'),
  (x'5c800041'),
  (x'15ae43fd'),
  (x'5d4cccfb'),
  (x'4c800001'),
  (x'57800ed8'),
  (x'5f000000'),
  (x'700000f0'),
  (x'5f23e9ac'),
  (x'5e9502f9'),
  (x'5e8012b1'),
  (x'3c000028'),
  (x'60cde861'),
  (x'03aa2a50'),
  (x'43480000'),
  (x'4c000000'),
  -- LooksLikePow5
  (x'5D1502F9'),
  (x'5D9502F9'),
  (x'5E1502F9'),
  -- OutputLength
  (x'3f99999a'),
  (x'3f9d70a4'),
  (x'3f9df3b6'),
  (x'3f9e0419'),
  (x'3f9e0610'),
  (x'3f9e064b'),
  (x'3f9e0651'),
  (x'03d20cfe')
)
select float4send(flt) as ibits,
       flt,
       flt::text::float4 as r_flt,
       float4send(flt::text::float4) as obits,
       float4send(flt::text::float4) = float4send(flt) as correct
  from (select bits::integer::xfloat4::float4 as flt
          from testdata
	offset 0) s;`,
				Results: []sql.Row{{`\x00000000`, 0, 0, `\x00000000`, true}, {`\x00800000`, 1.1754944e-38, 1.1754944e-38, `\x00800000`, true}, {`\x00800001`, 1.1754945e-38, 1.1754945e-38, `\x00800001`, true}, {`\x00800004`, 1.1754949e-38, 1.1754949e-38, `\x00800004`, true}, {`\x00800005`, 1.175495e-38, 1.175495e-38, `\x00800005`, true}, {`\x00800006`, 1.1754952e-38, 1.1754952e-38, `\x00800006`, true}, {`\x008002f1`, 1.1755999e-38, 1.1755999e-38, `\x008002f1`, true}, {`\x008002f2`, 1.1756e-38, 1.1756e-38, `\x008002f2`, true}, {`\x008002f3`, 1.1756001e-38, 1.1756001e-38, `\x008002f3`, true}, {`\x00800e17`, 1.1759998e-38, 1.1759998e-38, `\x00800e17`, true}, {`\x00800e18`, 1.176e-38, 1.176e-38, `\x00800e18`, true}, {`\x00800e19`, 1.1760001e-38, 1.1760001e-38, `\x00800e19`, true}, {`\x01000001`, 2.350989e-38, 2.350989e-38, `\x01000001`, true}, {`\x01102843`, 2.647751e-38, 2.647751e-38, `\x01102843`, true}, {`\x01a52c98`, 6.0675416e-38, 6.0675416e-38, `\x01a52c98`, true}, {`\x0219c229`, 1.1296386e-37, 1.1296386e-37, `\x0219c229`, true}, {`\x02e4464d`, 3.354194e-37, 3.354194e-37, `\x02e4464d`, true}, {`\x037343c1`, 7.148906e-37, 7.148906e-37, `\x037343c1`, true}, {`\x03a91b36`, 9.939175e-37, 9.939175e-37, `\x03a91b36`, true}, {`\x047ada65`, 2.948764e-36, 2.948764e-36, `\x047ada65`, true}, {`\x0496fe87`, 3.5498577e-36, 3.5498577e-36, `\x0496fe87`, true}, {`\x0550844f`, 9.804414e-36, 9.804414e-36, `\x0550844f`, true}, {`\x05999da3`, 1.4445957e-35, 1.4445957e-35, `\x05999da3`, true}, {`\x060ea5e2`, 2.6829103e-35, 2.6829103e-35, `\x060ea5e2`, true}, {`\x06e63c45`, 8.660494e-35, 8.660494e-35, `\x06e63c45`, true}, {`\x07f1e548`, 3.639641e-34, 3.639641e-34, `\x07f1e548`, true}, {`\x0fc5282b`, 1.9441172e-29, 1.9441172e-29, `\x0fc5282b`, true}, {`\x1f850283`, 5.6331846e-20, 5.6331846e-20, `\x1f850283`, true}, {`\x2874a9d6`, 1.3581548e-14, 1.3581548e-14, `\x2874a9d6`, true}, {`\x3356bf94`, 4.9999997e-08, 4.9999997e-08, `\x3356bf94`, true}, {`\x3356bf95`, 5e-08, 5e-08, `\x3356bf95`, true}, {`\x3356bf96`, 5.0000004e-08, 5.0000004e-08, `\x3356bf96`, true}, {`\x33d6bf94`, 9.9999994e-08, 9.9999994e-08, `\x33d6bf94`, true}, {`\x33d6bf95`, 1e-07, 1e-07, `\x33d6bf95`, true}, {`\x33d6bf96`, 1.0000001e-07, 1.0000001e-07, `\x33d6bf96`, true}, {`\x34a10faf`, 2.9999998e-07, 2.9999998e-07, `\x34a10faf`, true}, {`\x34a10fb0`, 3e-07, 3e-07, `\x34a10fb0`, true}, {`\x34a10fb1`, 3.0000004e-07, 3.0000004e-07, `\x34a10fb1`, true}, {`\x350637bc`, 4.9999994e-07, 4.9999994e-07, `\x350637bc`, true}, {`\x350637bd`, 5e-07, 5e-07, `\x350637bd`, true}, {`\x350637be`, 5.0000006e-07, 5.0000006e-07, `\x350637be`, true}, {`\x35719786`, 8.999999e-07, 8.999999e-07, `\x35719786`, true}, {`\x35719787`, 9e-07, 9e-07, `\x35719787`, true}, {`\x35719788`, 9.0000003e-07, 9.0000003e-07, `\x35719788`, true}, {`\x358637bc`, 9.999999e-07, 9.999999e-07, `\x358637bc`, true}, {`\x358637bd`, 1e-06, 1e-06, `\x358637bd`, true}, {`\x358637be`, 1.0000001e-06, 1.0000001e-06, `\x358637be`, true}, {`\x36a7c5ab`, 4.9999994e-06, 4.9999994e-06, `\x36a7c5ab`, true}, {`\x36a7c5ac`, 5e-06, 5e-06, `\x36a7c5ac`, true}, {`\x36a7c5ad`, 5.0000003e-06, 5.0000003e-06, `\x36a7c5ad`, true}, {`\x3727c5ab`, 9.999999e-06, 9.999999e-06, `\x3727c5ab`, true}, {`\x3727c5ac`, 1e-05, 1e-05, `\x3727c5ac`, true}, {`\x3727c5ad`, 1.0000001e-05, 1.0000001e-05, `\x3727c5ad`, true}, {`\x38d1b714`, 9.9999976e-05, 9.9999976e-05, `\x38d1b714`, true}, {`\x38d1b715`, 9.999998e-05, 9.999998e-05, `\x38d1b715`, true}, {`\x38d1b716`, 9.999999e-05, 9.999999e-05, `\x38d1b716`, true}, {`\x38d1b717`, 0.0001, 0.0001, `\x38d1b717`, true}, {`\x38d1b718`, 0.000100000005, 0.000100000005, `\x38d1b718`, true}, {`\x38d1b719`, 0.00010000001, 0.00010000001, `\x38d1b719`, true}, {`\x38d1b71a`, 0.00010000002, 0.00010000002, `\x38d1b71a`, true}, {`\x38d1b71b`, 0.00010000003, 0.00010000003, `\x38d1b71b`, true}, {`\x38d1b71c`, 0.000100000034, 0.000100000034, `\x38d1b71c`, true}, {`\x38d1b71d`, 0.00010000004, 0.00010000004, `\x38d1b71d`, true}, {`\x38dffffe`, 0.00010681151, 0.00010681151, `\x38dffffe`, true}, {`\x38dfffff`, 0.000106811516, 0.000106811516, `\x38dfffff`, true}, {`\x38e00000`, 0.00010681152, 0.00010681152, `\x38e00000`, true}, {`\x38efffff`, 0.00011444091, 0.00011444091, `\x38efffff`, true}, {`\x38f00000`, 0.00011444092, 0.00011444092, `\x38f00000`, true}, {`\x38f00001`, 0.000114440925, 0.000114440925, `\x38f00001`, true}, {`\x3a83126e`, 0.0009999999, 0.0009999999, `\x3a83126e`, true}, {`\x3a83126f`, 0.001, 0.001, `\x3a83126f`, true}, {`\x3a831270`, 0.0010000002, 0.0010000002, `\x3a831270`, true}, {`\x3c23d709`, 0.009999999, 0.009999999, `\x3c23d709`, true}, {`\x3c23d70a`, 0.01, 0.01, `\x3c23d70a`, true}, {`\x3c23d70b`, 0.010000001, 0.010000001, `\x3c23d70b`, true}, {`\x3dcccccc`, 0.099999994, 0.099999994, `\x3dcccccc`, true}, {`\x3dcccccd`, 0.1, 0.1, `\x3dcccccd`, true}, {`\x3dccccce`, 0.10000001, 0.10000001, `\x3dccccce`, true}, {`\x3dcccd6f`, 0.10000121, 0.10000121, `\x3dcccd6f`, true}, {`\x3dcccd70`, 0.100001216, 0.100001216, `\x3dcccd70`, true}, {`\x3dcccd71`, 0.10000122, 0.10000122, `\x3dcccd71`, true}, {`\x3effffff`, 0.49999997, 0.49999997, `\x3effffff`, true}, {`\x3f000000`, 0.5, 0.5, `\x3f000000`, true}, {`\x3f000001`, 0.50000006, 0.50000006, `\x3f000001`, true}, {`\x3f333332`, 0.6999999, 0.6999999, `\x3f333332`, true}, {`\x3f333333`, 0.7, 0.7, `\x3f333333`, true}, {`\x3f333334`, 0.70000005, 0.70000005, `\x3f333334`, true}, {`\x3f666665`, 0.8999999, 0.8999999, `\x3f666665`, true}, {`\x3f666666`, 0.9, 0.9, `\x3f666666`, true}, {`\x3f666667`, 0.90000004, 0.90000004, `\x3f666667`, true}, {`\x3f7d70a3`, 0.98999995, 0.98999995, `\x3f7d70a3`, true}, {`\x3f7d70a4`, 0.99, 0.99, `\x3f7d70a4`, true}, {`\x3f7d70a5`, 0.99000007, 0.99000007, `\x3f7d70a5`, true}, {`\x3f7fbe76`, 0.99899995, 0.99899995, `\x3f7fbe76`, true}, {`\x3f7fbe77`, 0.999, 0.999, `\x3f7fbe77`, true}, {`\x3f7fbe78`, 0.9990001, 0.9990001, `\x3f7fbe78`, true}, {`\x3f7ff971`, 0.9998999, 0.9998999, `\x3f7ff971`, true}, {`\x3f7ff972`, 0.9999, 0.9999, `\x3f7ff972`, true}, {`\x3f7ff973`, 0.99990004, 0.99990004, `\x3f7ff973`, true}, {`\x3f7fff57`, 0.9999899, 0.9999899, `\x3f7fff57`, true}, {`\x3f7fff58`, 0.99999, 0.99999, `\x3f7fff58`, true}, {`\x3f7fff59`, 0.99999005, 0.99999005, `\x3f7fff59`, true}, {`\x3f7fffee`, 0.9999989, 0.9999989, `\x3f7fffee`, true}, {`\x3f7fffef`, 0.999999, 0.999999, `\x3f7fffef`, true}, {`\x3f7ffff0`, 0.99999905, 0.99999905, `\x3f7ffff0`, true}, {`\x3f7ffff1`, 0.9999991, 0.9999991, `\x3f7ffff1`, true}, {`\x3f7ffff2`, 0.99999917, 0.99999917, `\x3f7ffff2`, true}, {`\x3f7ffff3`, 0.9999992, 0.9999992, `\x3f7ffff3`, true}, {`\x3f7ffff4`, 0.9999993, 0.9999993, `\x3f7ffff4`, true}, {`\x3f7ffff5`, 0.99999934, 0.99999934, `\x3f7ffff5`, true}, {`\x3f7ffff6`, 0.9999994, 0.9999994, `\x3f7ffff6`, true}, {`\x3f7ffff7`, 0.99999946, 0.99999946, `\x3f7ffff7`, true}, {`\x3f7ffff8`, 0.9999995, 0.9999995, `\x3f7ffff8`, true}, {`\x3f7ffff9`, 0.9999996, 0.9999996, `\x3f7ffff9`, true}, {`\x3f7ffffa`, 0.99999964, 0.99999964, `\x3f7ffffa`, true}, {`\x3f7ffffb`, 0.9999997, 0.9999997, `\x3f7ffffb`, true}, {`\x3f7ffffc`, 0.99999976, 0.99999976, `\x3f7ffffc`, true}, {`\x3f7ffffd`, 0.9999998, 0.9999998, `\x3f7ffffd`, true}, {`\x3f7ffffe`, 0.9999999, 0.9999999, `\x3f7ffffe`, true}, {`\x3f7fffff`, 0.99999994, 0.99999994, `\x3f7fffff`, true}, {`\x3f800000`, 1, 1, `\x3f800000`, true}, {`\x3f800001`, 1.0000001, 1.0000001, `\x3f800001`, true}, {`\x3f800002`, 1.0000002, 1.0000002, `\x3f800002`, true}, {`\x3f800003`, 1.0000004, 1.0000004, `\x3f800003`, true}, {`\x3f800004`, 1.0000005, 1.0000005, `\x3f800004`, true}, {`\x3f800005`, 1.0000006, 1.0000006, `\x3f800005`, true}, {`\x3f800006`, 1.0000007, 1.0000007, `\x3f800006`, true}, {`\x3f800007`, 1.0000008, 1.0000008, `\x3f800007`, true}, {`\x3f800008`, 1.000001, 1.000001, `\x3f800008`, true}, {`\x3f800009`, 1.0000011, 1.0000011, `\x3f800009`, true}, {`\x3f80000f`, 1.0000018, 1.0000018, `\x3f80000f`, true}, {`\x3f800010`, 1.0000019, 1.0000019, `\x3f800010`, true}, {`\x3f800011`, 1.000002, 1.000002, `\x3f800011`, true}, {`\x3f800012`, 1.0000021, 1.0000021, `\x3f800012`, true}, {`\x3f800013`, 1.0000023, 1.0000023, `\x3f800013`, true}, {`\x3f800014`, 1.0000024, 1.0000024, `\x3f800014`, true}, {`\x3f800017`, 1.0000027, 1.0000027, `\x3f800017`, true}, {`\x3f800018`, 1.0000029, 1.0000029, `\x3f800018`, true}, {`\x3f800019`, 1.000003, 1.000003, `\x3f800019`, true}, {`\x3f80001a`, 1.0000031, 1.0000031, `\x3f80001a`, true}, {`\x3f80001b`, 1.0000032, 1.0000032, `\x3f80001b`, true}, {`\x3f80001c`, 1.0000033, 1.0000033, `\x3f80001c`, true}, {`\x3f800029`, 1.0000049, 1.0000049, `\x3f800029`, true}, {`\x3f80002a`, 1.000005, 1.000005, `\x3f80002a`, true}, {`\x3f80002b`, 1.0000051, 1.0000051, `\x3f80002b`, true}, {`\x3f800053`, 1.0000099, 1.0000099, `\x3f800053`, true}, {`\x3f800054`, 1.00001, 1.00001, `\x3f800054`, true}, {`\x3f800055`, 1.0000101, 1.0000101, `\x3f800055`, true}, {`\x3f800346`, 1.0000999, 1.0000999, `\x3f800346`, true}, {`\x3f800347`, 1.0001, 1.0001, `\x3f800347`, true}, {`\x3f800348`, 1.0001001, 1.0001001, `\x3f800348`, true}, {`\x3f8020c4`, 1.0009999, 1.0009999, `\x3f8020c4`, true}, {`\x3f8020c5`, 1.001, 1.001, `\x3f8020c5`, true}, {`\x3f8020c6`, 1.0010002, 1.0010002, `\x3f8020c6`, true}, {`\x3f8147ad`, 1.0099999, 1.0099999, `\x3f8147ad`, true}, {`\x3f8147ae`, 1.01, 1.01, `\x3f8147ae`, true}, {`\x3f8147af`, 1.0100001, 1.0100001, `\x3f8147af`, true}, {`\x3f8ccccc`, 1.0999999, 1.0999999, `\x3f8ccccc`, true}, {`\x3f8ccccd`, 1.1, 1.1, `\x3f8ccccd`, true}, {`\x3f8cccce`, 1.1000001, 1.1000001, `\x3f8cccce`, true}, {`\x3fc90fdb`, 1.5707964, 1.5707964, `\x3fc90fdb`, true}, {`\x402df854`, 2.7182817, 2.7182817, `\x402df854`, true}, {`\x40490fdb`, 3.1415927, 3.1415927, `\x40490fdb`, true}, {`\x409fffff`, 4.9999995, 4.9999995, `\x409fffff`, true}, {`\x40a00000`, 5, 5, `\x40a00000`, true}, {`\x40a00001`, 5.0000005, 5.0000005, `\x40a00001`, true}, {`\x40afffff`, 5.4999995, 5.4999995, `\x40afffff`, true}, {`\x40b00000`, 5.5, 5.5, `\x40b00000`, true}, {`\x40b00001`, 5.5000005, 5.5000005, `\x40b00001`, true}, {`\x411fffff`, 9.999999, 9.999999, `\x411fffff`, true}, {`\x41200000`, 10, 10, `\x41200000`, true}, {`\x41200001`, 10.000001, 10.000001, `\x41200001`, true}, {`\x42c7ffff`, 99.99999, 99.99999, `\x42c7ffff`, true}, {`\x42c80000`, 100, 100, `\x42c80000`, true}, {`\x42c80001`, 100.00001, 100.00001, `\x42c80001`, true}, {`\x4479ffff`, 999.99994, 999.99994, `\x4479ffff`, true}, {`\x447a0000`, 1000, 1000, `\x447a0000`, true}, {`\x447a0001`, 1000.00006, 1000.00006, `\x447a0001`, true}, {`\x461c3fff`, 9999.999, 9999.999, `\x461c3fff`, true}, {`\x461c4000`, 10000, 10000, `\x461c4000`, true}, {`\x461c4001`, 10000.001, 10000.001, `\x461c4001`, true}, {`\x47c34fff`, 99999.99, 99999.99, `\x47c34fff`, true}, {`\x47c35000`, 100000, 100000, `\x47c35000`, true}, {`\x47c35001`, 100000.01, 100000.01, `\x47c35001`, true}, {`\x497423ff`, 999999.94, 999999.94, `\x497423ff`, true}, {`\x49742400`, 1e+06, 1e+06, `\x49742400`, true}, {`\x49742401`, 1.00000006e+06, 1.00000006e+06, `\x49742401`, true}, {`\x4b18967f`, 9.999999e+06, 9.999999e+06, `\x4b18967f`, true}, {`\x4b189680`, 1e+07, 1e+07, `\x4b189680`, true}, {`\x4b189681`, 1.0000001e+07, 1.0000001e+07, `\x4b189681`, true}, {`\x4cbebc1f`, 9.999999e+07, 9.999999e+07, `\x4cbebc1f`, true}, {`\x4cbebc20`, 1e+08, 1e+08, `\x4cbebc20`, true}, {`\x4cbebc21`, 1.0000001e+08, 1.0000001e+08, `\x4cbebc21`, true}, {`\x4e6e6b27`, 9.9999994e+08, 9.9999994e+08, `\x4e6e6b27`, true}, {`\x4e6e6b28`, 1e+09, 1e+09, `\x4e6e6b28`, true}, {`\x4e6e6b29`, 1.00000006e+09, 1.00000006e+09, `\x4e6e6b29`, true}, {`\x501502f8`, 9.999999e+09, 9.999999e+09, `\x501502f8`, true}, {`\x501502f9`, 1e+10, 1e+10, `\x501502f9`, true}, {`\x501502fa`, 1.0000001e+10, 1.0000001e+10, `\x501502fa`, true}, {`\x51ba43b6`, 9.999999e+10, 9.999999e+10, `\x51ba43b6`, true}, {`\x51ba43b7`, 1e+11, 1e+11, `\x51ba43b7`, true}, {`\x51ba43b8`, 1.0000001e+11, 1.0000001e+11, `\x51ba43b8`, true}, {`\x1f6c1e4a`, 5e-20, 5e-20, `\x1f6c1e4a`, true}, {`\x59be6cea`, 6.7e+15, 6.7e+15, `\x59be6cea`, true}, {`\x5d5ab6c4`, 9.85e+17, 9.85e+17, `\x5d5ab6c4`, true}, {`\x2cc4a9bd`, 5.5895e-12, 5.5895e-12, `\x2cc4a9bd`, true}, {`\x15ae43fd`, 7.038531e-26, 7.038531e-26, `\x15ae43fd`, true}, {`\x2cf757ca`, 7.0299088e-12, 7.0299088e-12, `\x2cf757ca`, true}, {`\x665ba998`, 2.5933168e+23, 2.5933168e+23, `\x665ba998`, true}, {`\x743c3324`, 5.9642887e+31, 5.9642887e+31, `\x743c3324`, true}, {`\x47f1205a`, 123456.7, 123456.7, `\x47f1205a`, true}, {`\x4640e6ae`, 12345.67, 12345.67, `\x4640e6ae`, true}, {`\x449a5225`, 1234.567, 1234.567, `\x449a5225`, true}, {`\x42f6e9d5`, 123.4567, 123.4567, `\x42f6e9d5`, true}, {`\x414587dd`, 12.34567, 12.34567, `\x414587dd`, true}, {`\x3f9e064b`, 1.234567, 1.234567, `\x3f9e064b`, true}, {`\x4c000004`, 3.3554448e+07, 3.3554448e+07, `\x4c000004`, true}, {`\x50061c46`, 8.999999e+09, 8.999999e+09, `\x50061c46`, true}, {`\x510006a8`, 3.4366718e+10, 3.4366718e+10, `\x510006a8`, true}, {`\x48951f84`, 305404.12, 305404.12, `\x48951f84`, true}, {`\x45fd1840`, 8099.0312, 8099.0312, `\x45fd1840`, true}, {`\x39800000`, 0.00024414062, 0.00024414062, `\x39800000`, true}, {`\x3b200000`, 0.0024414062, 0.0024414062, `\x3b200000`, true}, {`\x3b900000`, 0.0043945312, 0.0043945312, `\x3b900000`, true}, {`\x3bd00000`, 0.0063476562, 0.0063476562, `\x3bd00000`, true}, {`\x63800000`, 4.7223665e+21, 4.7223665e+21, `\x63800000`, true}, {`\x4b000000`, 8.388608e+06, 8.388608e+06, `\x4b000000`, true}, {`\x4b800000`, 1.6777216e+07, 1.6777216e+07, `\x4b800000`, true}, {`\x4c000001`, 3.3554436e+07, 3.3554436e+07, `\x4c000001`, true}, {`\x4c800b0d`, 6.7131496e+07, 6.7131496e+07, `\x4c800b0d`, true}, {`\x00d24584`, 1.9310392e-38, 1.9310392e-38, `\x00d24584`, true}, {`\x00d90b88`, 1.993244e-38, 1.993244e-38, `\x00d90b88`, true}, {`\x45803f34`, 4103.9004, 4103.9004, `\x45803f34`, true}, {`\x4f9f24f7`, 5.3399997e+09, 5.3399997e+09, `\x4f9f24f7`, true}, {`\x3a8722c3`, 0.0010310042, 0.0010310042, `\x3a8722c3`, true}, {`\x5c800041`, 2.882326e+17, 2.882326e+17, `\x5c800041`, true}, {`\x15ae43fd`, 7.038531e-26, 7.038531e-26, `\x15ae43fd`, true}, {`\x5d4cccfb`, 9.223404e+17, 9.223404e+17, `\x5d4cccfb`, true}, {`\x4c800001`, 6.710887e+07, 6.710887e+07, `\x4c800001`, true}, {`\x57800ed8`, 2.816025e+14, 2.816025e+14, `\x57800ed8`, true}, {`\x5f000000`, 9.223372e+18, 9.223372e+18, `\x5f000000`, true}, {`\x700000f0`, 1.5846086e+29, 1.5846086e+29, `\x700000f0`, true}, {`\x5f23e9ac`, 1.1811161e+19, 1.1811161e+19, `\x5f23e9ac`, true}, {`\x5e9502f9`, 5.368709e+18, 5.368709e+18, `\x5e9502f9`, true}, {`\x5e8012b1`, 4.6143166e+18, 4.6143166e+18, `\x5e8012b1`, true}, {`\x3c000028`, 0.007812537, 0.007812537, `\x3c000028`, true}, {`\x60cde861`, 1.18697725e+20, 1.18697725e+20, `\x60cde861`, true}, {`\x03aa2a50`, 1.00014165e-36, 1.00014165e-36, `\x03aa2a50`, true}, {`\x43480000`, 200, 200, `\x43480000`, true}, {`\x4c000000`, 3.3554432e+07, 3.3554432e+07, `\x4c000000`, true}, {`\x5d1502f9`, 6.7108864e+17, 6.7108864e+17, `\x5d1502f9`, true}, {`\x5d9502f9`, 1.3421773e+18, 1.3421773e+18, `\x5d9502f9`, true}, {`\x5e1502f9`, 2.6843546e+18, 2.6843546e+18, `\x5e1502f9`, true}, {`\x3f99999a`, 1.2, 1.2, `\x3f99999a`, true}, {`\x3f9d70a4`, 1.23, 1.23, `\x3f9d70a4`, true}, {`\x3f9df3b6`, 1.234, 1.234, `\x3f9df3b6`, true}, {`\x3f9e0419`, 1.2345, 1.2345, `\x3f9e0419`, true}, {`\x3f9e0610`, 1.23456, 1.23456, `\x3f9e0610`, true}, {`\x3f9e064b`, 1.234567, 1.234567, `\x3f9e064b`, true}, {`\x3f9e0651`, 1.2345678, 1.2345678, `\x3f9e0651`, true}, {`\x03d20cfe`, 1.23456735e-36, 1.23456735e-36, `\x03d20cfe`, true}},
			},
			{
				Statement: `drop type xfloat4 cascade;`,
			},
			{
				Statement: `DETAIL:  drop cascades to function xfloat4in(cstring)
drop cascades to function xfloat4out(xfloat4)
drop cascades to cast from xfloat4 to real
drop cascades to cast from real to xfloat4
drop cascades to cast from xfloat4 to integer
drop cascades to cast from integer to xfloat4`,
			},
		},
	})
}
