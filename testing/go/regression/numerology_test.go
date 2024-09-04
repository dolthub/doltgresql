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

func TestNumerology(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_numerology)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_numerology,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `SELECT 123abc;`,
				ErrorString: `trailing junk after numeric literal at or near "123a"`,
			},
			{
				Statement:   `SELECT 0x0o;`,
				ErrorString: `trailing junk after numeric literal at or near "0x"`,
			},
			{
				Statement:   `SELECT 1_2_3;`,
				ErrorString: `trailing junk after numeric literal at or near "1_"`,
			},
			{
				Statement:   `SELECT 0.a;`,
				ErrorString: `trailing junk after numeric literal at or near "0.a"`,
			},
			{
				Statement:   `SELECT 0.0a;`,
				ErrorString: `trailing junk after numeric literal at or near "0.0a"`,
			},
			{
				Statement:   `SELECT .0a;`,
				ErrorString: `trailing junk after numeric literal at or near ".0a"`,
			},
			{
				Statement:   `SELECT 0.0e1a;`,
				ErrorString: `trailing junk after numeric literal at or near "0.0e1a"`,
			},
			{
				Statement:   `SELECT 0.0e;`,
				ErrorString: `trailing junk after numeric literal at or near "0.0e"`,
			},
			{
				Statement:   `SELECT 0.0e+a;`,
				ErrorString: `trailing junk after numeric literal at or near "0.0e+"`,
			},
			{
				Statement:   `PREPARE p1 AS SELECT $1a;`,
				ErrorString: `trailing junk after parameter at or near "$1a"`,
			},
			{
				Statement: `CREATE TABLE TEMP_FLOAT (f1 FLOAT8);`,
			},
			{
				Statement: `INSERT INTO TEMP_FLOAT (f1)
  SELECT float8(f1) FROM INT4_TBL;`,
			},
			{
				Statement: `INSERT INTO TEMP_FLOAT (f1)
  SELECT float8(f1) FROM INT2_TBL;`,
			},
			{
				Statement: `SELECT f1 FROM TEMP_FLOAT
  ORDER BY f1;`,
				Results: []sql.Row{{-2147483647}, {-123456}, {-32767}, {-1234}, {0}, {0}, {1234}, {32767}, {123456}, {2147483647}},
			},
			{
				Statement: `CREATE TABLE TEMP_INT4 (f1 INT4);`,
			},
			{
				Statement: `INSERT INTO TEMP_INT4 (f1)
  SELECT int4(f1) FROM FLOAT8_TBL
  WHERE (f1 > -2147483647) AND (f1 < 2147483647);`,
			},
			{
				Statement: `INSERT INTO TEMP_INT4 (f1)
  SELECT int4(f1) FROM INT2_TBL;`,
			},
			{
				Statement: `SELECT f1 FROM TEMP_INT4
  ORDER BY f1;`,
				Results: []sql.Row{{-32767}, {-1234}, {-1004}, {-35}, {0}, {0}, {0}, {1234}, {32767}},
			},
			{
				Statement: `CREATE TABLE TEMP_INT2 (f1 INT2);`,
			},
			{
				Statement: `INSERT INTO TEMP_INT2 (f1)
  SELECT int2(f1) FROM FLOAT8_TBL
  WHERE (f1 >= -32767) AND (f1 <= 32767);`,
			},
			{
				Statement: `INSERT INTO TEMP_INT2 (f1)
  SELECT int2(f1) FROM INT4_TBL
  WHERE (f1 >= -32767) AND (f1 <= 32767);`,
			},
			{
				Statement: `SELECT f1 FROM TEMP_INT2
  ORDER BY f1;`,
				Results: []sql.Row{{-1004}, {-35}, {0}, {0}, {0}},
			},
			{
				Statement: `CREATE TABLE TEMP_GROUP (f1 INT4, f2 INT4, f3 FLOAT8);`,
			},
			{
				Statement: `INSERT INTO TEMP_GROUP
  SELECT 1, (- i.f1), (- f.f1)
  FROM INT4_TBL i, FLOAT8_TBL f;`,
			},
			{
				Statement: `INSERT INTO TEMP_GROUP
  SELECT 2, i.f1, f.f1
  FROM INT4_TBL i, FLOAT8_TBL f;`,
			},
			{
				Statement: `SELECT DISTINCT f1 AS two FROM TEMP_GROUP ORDER BY 1;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `SELECT f1 AS two, max(f3) AS max_float, min(f3) as min_float
  FROM TEMP_GROUP
  GROUP BY f1
  ORDER BY two, max_float, min_float;`,
				Results: []sql.Row{{1, 1.2345678901234e+200, -0}, {2, 0, -1.2345678901234e+200}},
			},
			{
				Statement: `SELECT f1 AS two, max(f3) AS max_float, min(f3) AS min_float
  FROM TEMP_GROUP
  GROUP BY two
  ORDER BY two, max_float, min_float;`,
				Results: []sql.Row{{1, 1.2345678901234e+200, -0}, {2, 0, -1.2345678901234e+200}},
			},
			{
				Statement: `SELECT f1 AS two, (max(f3) + 1) AS max_plus_1, (min(f3) - 1) AS min_minus_1
  FROM TEMP_GROUP
  GROUP BY f1
  ORDER BY two, min_minus_1;`,
				Results: []sql.Row{{1, 1.2345678901234e+200, -1}, {2, 1, -1.2345678901234e+200}},
			},
			{
				Statement: `SELECT f1 AS two,
       max(f2) + min(f2) AS max_plus_min,
       min(f3) - 1 AS min_minus_1
  FROM TEMP_GROUP
  GROUP BY f1
  ORDER BY two, min_minus_1;`,
				Results: []sql.Row{{1, 0, -1}, {2, 0, -1.2345678901234e+200}},
			},
			{
				Statement: `DROP TABLE TEMP_INT2;`,
			},
			{
				Statement: `DROP TABLE TEMP_INT4;`,
			},
			{
				Statement: `DROP TABLE TEMP_FLOAT;`,
			},
			{
				Statement: `DROP TABLE TEMP_GROUP;`,
			},
		},
	})
}
