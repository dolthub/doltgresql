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

func TestInt4(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_int4)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_int4,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `INSERT INTO INT4_TBL(f1) VALUES ('34.5');`,
				ErrorString: `invalid input syntax for type integer: "34.5"`,
			},
			{
				Statement:   `INSERT INTO INT4_TBL(f1) VALUES ('1000000000000');`,
				ErrorString: `value "1000000000000" is out of range for type integer`,
			},
			{
				Statement:   `INSERT INTO INT4_TBL(f1) VALUES ('asdf');`,
				ErrorString: `invalid input syntax for type integer: "asdf"`,
			},
			{
				Statement:   `INSERT INTO INT4_TBL(f1) VALUES ('     ');`,
				ErrorString: `invalid input syntax for type integer: "     "`,
			},
			{
				Statement:   `INSERT INTO INT4_TBL(f1) VALUES ('   asdf   ');`,
				ErrorString: `invalid input syntax for type integer: "   asdf   "`,
			},
			{
				Statement:   `INSERT INTO INT4_TBL(f1) VALUES ('- 1234');`,
				ErrorString: `invalid input syntax for type integer: "- 1234"`,
			},
			{
				Statement:   `INSERT INTO INT4_TBL(f1) VALUES ('123       5');`,
				ErrorString: `invalid input syntax for type integer: "123       5"`,
			},
			{
				Statement:   `INSERT INTO INT4_TBL(f1) VALUES ('');`,
				ErrorString: `invalid input syntax for type integer: ""`,
			},
			{
				Statement: `SELECT * FROM INT4_TBL;`,
				Results:   []sql.Row{{0}, {123456}, {-123456}, {2147483647}, {-2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 <> int2 '0';`,
				Results:   []sql.Row{{123456}, {-123456}, {2147483647}, {-2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 <> int4 '0';`,
				Results:   []sql.Row{{123456}, {-123456}, {2147483647}, {-2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 = int2 '0';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 = int4 '0';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 < int2 '0';`,
				Results:   []sql.Row{{-123456}, {-2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 < int4 '0';`,
				Results:   []sql.Row{{-123456}, {-2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 <= int2 '0';`,
				Results:   []sql.Row{{0}, {-123456}, {-2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 <= int4 '0';`,
				Results:   []sql.Row{{0}, {-123456}, {-2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 > int2 '0';`,
				Results:   []sql.Row{{123456}, {2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 > int4 '0';`,
				Results:   []sql.Row{{123456}, {2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 >= int2 '0';`,
				Results:   []sql.Row{{0}, {123456}, {2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE i.f1 >= int4 '0';`,
				Results:   []sql.Row{{0}, {123456}, {2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE (i.f1 % int2 '2') = int2 '1';`,
				Results:   []sql.Row{{2147483647}},
			},
			{
				Statement: `SELECT i.* FROM INT4_TBL i WHERE (i.f1 % int4 '2') = int2 '0';`,
				Results:   []sql.Row{{0}, {123456}, {-123456}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 * int2 '2' AS x FROM INT4_TBL i;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 * int2 '2' AS x FROM INT4_TBL i
WHERE abs(f1) < 1073741824;`,
				Results: []sql.Row{{0, 0}, {123456, 246912}, {-123456, -246912}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 * int4 '2' AS x FROM INT4_TBL i;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 * int4 '2' AS x FROM INT4_TBL i
WHERE abs(f1) < 1073741824;`,
				Results: []sql.Row{{0, 0}, {123456, 246912}, {-123456, -246912}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 + int2 '2' AS x FROM INT4_TBL i;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 + int2 '2' AS x FROM INT4_TBL i
WHERE f1 < 2147483646;`,
				Results: []sql.Row{{0, 2}, {123456, 123458}, {-123456, -123454}, {-2147483647, -2147483645}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 + int4 '2' AS x FROM INT4_TBL i;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 + int4 '2' AS x FROM INT4_TBL i
WHERE f1 < 2147483646;`,
				Results: []sql.Row{{0, 2}, {123456, 123458}, {-123456, -123454}, {-2147483647, -2147483645}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 - int2 '2' AS x FROM INT4_TBL i;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 - int2 '2' AS x FROM INT4_TBL i
WHERE f1 > -2147483647;`,
				Results: []sql.Row{{0, -2}, {123456, 123454}, {-123456, -123458}, {2147483647, 2147483645}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 - int4 '2' AS x FROM INT4_TBL i;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 - int4 '2' AS x FROM INT4_TBL i
WHERE f1 > -2147483647;`,
				Results: []sql.Row{{0, -2}, {123456, 123454}, {-123456, -123458}, {2147483647, 2147483645}},
			},
			{
				Statement: `SELECT i.f1, i.f1 / int2 '2' AS x FROM INT4_TBL i;`,
				Results:   []sql.Row{{0, 0}, {123456, 61728}, {-123456, -61728}, {2147483647, 1073741823}, {-2147483647, -1073741823}},
			},
			{
				Statement: `SELECT i.f1, i.f1 / int4 '2' AS x FROM INT4_TBL i;`,
				Results:   []sql.Row{{0, 0}, {123456, 61728}, {-123456, -61728}, {2147483647, 1073741823}, {-2147483647, -1073741823}},
			},
			{
				Statement: `SELECT -2+3 AS one;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT 4-2 AS two;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT 2- -1 AS three;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT 2 - -2 AS four;`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `SELECT int2 '2' * int2 '2' = int2 '16' / int2 '4' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT int4 '2' * int2 '2' = int2 '16' / int4 '4' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT int2 '2' * int4 '2' = int4 '16' / int2 '4' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT int4 '1000' < int4 '999' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 AS ten;`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `SELECT 2 + 2 / 2 AS three;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT (2 + 2) / 2 AS two;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT (-1::int4<<31)::text;`,
				Results:   []sql.Row{{-2147483648}},
			},
			{
				Statement: `SELECT ((-1::int4<<31)+1)::text;`,
				Results:   []sql.Row{{-2147483647}},
			},
			{
				Statement:   `SELECT (-2147483648)::int4 * (-1)::int4;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement:   `SELECT (-2147483648)::int4 / (-1)::int4;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT (-2147483648)::int4 % (-1)::int4;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT (-2147483648)::int4 * (-1)::int2;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement:   `SELECT (-2147483648)::int4 / (-1)::int2;`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT (-2147483648)::int4 % (-1)::int2;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT x, x::int4 AS int4_value
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
				Statement: `SELECT x, x::int4 AS int4_value
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
FROM (VALUES (0::int4, 0::int4),
             (0::int4, 6410818::int4),
             (61866666::int4, 6410818::int4),
             (-61866666::int4, 6410818::int4),
             ((-2147483648)::int4, 1::int4),
             ((-2147483648)::int4, 2147483647::int4),
             ((-2147483648)::int4, 1073741824::int4)) AS v(a, b);`,
				Results: []sql.Row{{0, 0, 0, 0, 0, 0}, {0, 6410818, 6410818, 6410818, 6410818, 6410818}, {61866666, 6410818, 1466, 1466, 1466, 1466}, {-61866666, 6410818, 1466, 1466, 1466, 1466}, {-2147483648, 1, 1, 1, 1, 1}, {-2147483648, 2147483647, 1, 1, 1, 1}, {-2147483648, 1073741824, 1073741824, 1073741824, 1073741824, 1073741824}},
			},
			{
				Statement:   `SELECT gcd((-2147483648)::int4, 0::int4); -- overflow`,
				ErrorString: `integer out of range`,
			},
			{
				Statement:   `SELECT gcd((-2147483648)::int4, (-2147483648)::int4); -- overflow`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT a, b, lcm(a, b), lcm(a, -b), lcm(b, a), lcm(-b, a)
FROM (VALUES (0::int4, 0::int4),
             (0::int4, 42::int4),
             (42::int4, 42::int4),
             (330::int4, 462::int4),
             (-330::int4, 462::int4),
             ((-2147483648)::int4, 0::int4)) AS v(a, b);`,
				Results: []sql.Row{{0, 0, 0, 0, 0, 0}, {0, 42, 0, 0, 0, 0}, {42, 42, 42, 42, 42, 42}, {330, 462, 2310, 2310, 2310, 2310}, {-330, 462, 2310, 2310, 2310, 2310}, {-2147483648, 0, 0, 0, 0, 0}},
			},
			{
				Statement:   `SELECT lcm((-2147483648)::int4, 1::int4); -- overflow`,
				ErrorString: `integer out of range`,
			},
			{
				Statement:   `SELECT lcm(2147483647::int4, 2147483646::int4); -- overflow`,
				ErrorString: `integer out of range`,
			},
		},
	})
}
