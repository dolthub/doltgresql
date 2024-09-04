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

func TestInt2(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_int2)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_int2,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `INSERT INTO INT2_TBL(f1) VALUES ('34.5');`,
				ErrorString: `invalid input syntax for type smallint: "34.5"`,
			},
			{
				Statement:   `INSERT INTO INT2_TBL(f1) VALUES ('100000');`,
				ErrorString: `value "100000" is out of range for type smallint`,
			},
			{
				Statement:   `INSERT INTO INT2_TBL(f1) VALUES ('asdf');`,
				ErrorString: `invalid input syntax for type smallint: "asdf"`,
			},
			{
				Statement:   `INSERT INTO INT2_TBL(f1) VALUES ('    ');`,
				ErrorString: `invalid input syntax for type smallint: "    "`,
			},
			{
				Statement:   `INSERT INTO INT2_TBL(f1) VALUES ('- 1234');`,
				ErrorString: `invalid input syntax for type smallint: "- 1234"`,
			},
			{
				Statement:   `INSERT INTO INT2_TBL(f1) VALUES ('4 444');`,
				ErrorString: `invalid input syntax for type smallint: "4 444"`,
			},
			{
				Statement:   `INSERT INTO INT2_TBL(f1) VALUES ('123 dt');`,
				ErrorString: `invalid input syntax for type smallint: "123 dt"`,
			},
			{
				Statement:   `INSERT INTO INT2_TBL(f1) VALUES ('');`,
				ErrorString: `invalid input syntax for type smallint: ""`,
			},
			{
				Statement: `SELECT * FROM INT2_TBL;`,
				Results:   []sql.Row{{0}, {1234}, {-1234}, {32767}, {-32767}},
			},
			{
				Statement:   `SELECT * FROM INT2_TBL AS f(a, b);`,
				ErrorString: `table "f" has 1 columns available but 2 columns specified`,
			},
			{
				Statement:   `SELECT * FROM (TABLE int2_tbl) AS s (a, b);`,
				ErrorString: `table "s" has 1 columns available but 2 columns specified`,
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 <> int2 '0';`,
				Results:   []sql.Row{{1234}, {-1234}, {32767}, {-32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 <> int4 '0';`,
				Results:   []sql.Row{{1234}, {-1234}, {32767}, {-32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 = int2 '0';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 = int4 '0';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 < int2 '0';`,
				Results:   []sql.Row{{-1234}, {-32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 < int4 '0';`,
				Results:   []sql.Row{{-1234}, {-32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 <= int2 '0';`,
				Results:   []sql.Row{{0}, {-1234}, {-32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 <= int4 '0';`,
				Results:   []sql.Row{{0}, {-1234}, {-32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 > int2 '0';`,
				Results:   []sql.Row{{1234}, {32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 > int4 '0';`,
				Results:   []sql.Row{{1234}, {32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 >= int2 '0';`,
				Results:   []sql.Row{{0}, {1234}, {32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE i.f1 >= int4 '0';`,
				Results:   []sql.Row{{0}, {1234}, {32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE (i.f1 % int2 '2') = int2 '1';`,
				Results:   []sql.Row{{32767}},
			},
			{
				Statement: `SELECT i.* FROM INT2_TBL i WHERE (i.f1 % int4 '2') = int2 '0';`,
				Results:   []sql.Row{{0}, {1234}, {-1234}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 * int2 '2' AS x FROM INT2_TBL i;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 * int2 '2' AS x FROM INT2_TBL i
WHERE abs(f1) < 16384;`,
				Results: []sql.Row{{0, 0}, {1234, 2468}, {-1234, -2468}},
			},
			{
				Statement: `SELECT i.f1, i.f1 * int4 '2' AS x FROM INT2_TBL i;`,
				Results:   []sql.Row{{0, 0}, {1234, 2468}, {-1234, -2468}, {32767, 65534}, {-32767, -65534}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 + int2 '2' AS x FROM INT2_TBL i;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 + int2 '2' AS x FROM INT2_TBL i
WHERE f1 < 32766;`,
				Results: []sql.Row{{0, 2}, {1234, 1236}, {-1234, -1232}, {-32767, -32765}},
			},
			{
				Statement: `SELECT i.f1, i.f1 + int4 '2' AS x FROM INT2_TBL i;`,
				Results:   []sql.Row{{0, 2}, {1234, 1236}, {-1234, -1232}, {32767, 32769}, {-32767, -32765}},
			},
			{
				Statement:   `SELECT i.f1, i.f1 - int2 '2' AS x FROM INT2_TBL i;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT i.f1, i.f1 - int2 '2' AS x FROM INT2_TBL i
WHERE f1 > -32767;`,
				Results: []sql.Row{{0, -2}, {1234, 1232}, {-1234, -1236}, {32767, 32765}},
			},
			{
				Statement: `SELECT i.f1, i.f1 - int4 '2' AS x FROM INT2_TBL i;`,
				Results:   []sql.Row{{0, -2}, {1234, 1232}, {-1234, -1236}, {32767, 32765}, {-32767, -32769}},
			},
			{
				Statement: `SELECT i.f1, i.f1 / int2 '2' AS x FROM INT2_TBL i;`,
				Results:   []sql.Row{{0, 0}, {1234, 617}, {-1234, -617}, {32767, 16383}, {-32767, -16383}},
			},
			{
				Statement: `SELECT i.f1, i.f1 / int4 '2' AS x FROM INT2_TBL i;`,
				Results:   []sql.Row{{0, 0}, {1234, 617}, {-1234, -617}, {32767, 16383}, {-32767, -16383}},
			},
			{
				Statement: `SELECT (-1::int2<<15)::text;`,
				Results:   []sql.Row{{-32768}},
			},
			{
				Statement: `SELECT ((-1::int2<<15)+1::int2)::text;`,
				Results:   []sql.Row{{-32767}},
			},
			{
				Statement:   `SELECT (-32768)::int2 * (-1)::int2;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement:   `SELECT (-32768)::int2 / (-1)::int2;`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `SELECT (-32768)::int2 % (-1)::int2;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT x, x::int2 AS int2_value
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
				Statement: `SELECT x, x::int2 AS int2_value
FROM (VALUES (-2.5::numeric),
             (-1.5::numeric),
             (-0.5::numeric),
             (0.0::numeric),
             (0.5::numeric),
             (1.5::numeric),
             (2.5::numeric)) t(x);`,
				Results: []sql.Row{{-2.5, -3}, {-1.5, -2}, {-0.5, -1}, {0.0, 0}, {0.5, 1}, {1.5, 2}, {2.5, 3}},
			},
		},
	})
}
