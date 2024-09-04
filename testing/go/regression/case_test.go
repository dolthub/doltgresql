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

func TestCase(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_case)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_case,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE CASE_TBL (
  i integer,
  f double precision
);`,
			},
			{
				Statement: `CREATE TABLE CASE2_TBL (
  i integer,
  j integer
);`,
			},
			{
				Statement: `INSERT INTO CASE_TBL VALUES (1, 10.1);`,
			},
			{
				Statement: `INSERT INTO CASE_TBL VALUES (2, 20.2);`,
			},
			{
				Statement: `INSERT INTO CASE_TBL VALUES (3, -30.3);`,
			},
			{
				Statement: `INSERT INTO CASE_TBL VALUES (4, NULL);`,
			},
			{
				Statement: `INSERT INTO CASE2_TBL VALUES (1, -1);`,
			},
			{
				Statement: `INSERT INTO CASE2_TBL VALUES (2, -2);`,
			},
			{
				Statement: `INSERT INTO CASE2_TBL VALUES (3, -3);`,
			},
			{
				Statement: `INSERT INTO CASE2_TBL VALUES (2, -4);`,
			},
			{
				Statement: `INSERT INTO CASE2_TBL VALUES (1, NULL);`,
			},
			{
				Statement: `INSERT INTO CASE2_TBL VALUES (NULL, -6);`,
			},
			{
				Statement: `SELECT '3' AS "One",
  CASE
    WHEN 1 < 2 THEN 3
  END AS "Simple WHEN";`,
				Results: []sql.Row{{3, 3}},
			},
			{
				Statement: `SELECT '<NULL>' AS "One",
  CASE
    WHEN 1 > 2 THEN 3
  END AS "Simple default";`,
				Results: []sql.Row{{`<NULL>`, ``}},
			},
			{
				Statement: `SELECT '3' AS "One",
  CASE
    WHEN 1 < 2 THEN 3
    ELSE 4
  END AS "Simple ELSE";`,
				Results: []sql.Row{{3, 3}},
			},
			{
				Statement: `SELECT '4' AS "One",
  CASE
    WHEN 1 > 2 THEN 3
    ELSE 4
  END AS "ELSE default";`,
				Results: []sql.Row{{4, 4}},
			},
			{
				Statement: `SELECT '6' AS "One",
  CASE
    WHEN 1 > 2 THEN 3
    WHEN 4 < 5 THEN 6
    ELSE 7
  END AS "Two WHEN with default";`,
				Results: []sql.Row{{6, 6}},
			},
			{
				Statement: `SELECT '7' AS "None",
   CASE WHEN random() < 0 THEN 1
   END AS "NULL on no matches";`,
				Results: []sql.Row{{7, ``}},
			},
			{
				Statement: `SELECT CASE WHEN 1=0 THEN 1/0 WHEN 1=1 THEN 1 ELSE 2/0 END;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT CASE 1 WHEN 0 THEN 1/0 WHEN 1 THEN 1 ELSE 2/0 END;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT CASE WHEN i > 100 THEN 1/0 ELSE 0 END FROM case_tbl;`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `SELECT CASE 'a' WHEN 'a' THEN 1 ELSE 2 END;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT
  CASE
    WHEN i >= 3 THEN i
  END AS ">= 3 or Null"
  FROM CASE_TBL;`,
				Results: []sql.Row{{``}, {``}, {3}, {4}},
			},
			{
				Statement: `SELECT
  CASE WHEN i >= 3 THEN (i + i)
       ELSE i
  END AS "Simplest Math"
  FROM CASE_TBL;`,
				Results: []sql.Row{{1}, {2}, {6}, {8}},
			},
			{
				Statement: `SELECT i AS "Value",
  CASE WHEN (i < 0) THEN 'small'
       WHEN (i = 0) THEN 'zero'
       WHEN (i = 1) THEN 'one'
       WHEN (i = 2) THEN 'two'
       ELSE 'big'
  END AS "Category"
  FROM CASE_TBL;`,
				Results: []sql.Row{{1, `one`}, {2, `two`}, {3, `big`}, {4, `big`}},
			},
			{
				Statement: `SELECT
  CASE WHEN ((i < 0) or (i < 0)) THEN 'small'
       WHEN ((i = 0) or (i = 0)) THEN 'zero'
       WHEN ((i = 1) or (i = 1)) THEN 'one'
       WHEN ((i = 2) or (i = 2)) THEN 'two'
       ELSE 'big'
  END AS "Category"
  FROM CASE_TBL;`,
				Results: []sql.Row{{`one`}, {`two`}, {`big`}, {`big`}},
			},
			{
				Statement: `SELECT * FROM CASE_TBL WHERE COALESCE(f,i) = 4;`,
				Results:   []sql.Row{{4, ``}},
			},
			{
				Statement: `SELECT * FROM CASE_TBL WHERE NULLIF(f,i) = 2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT COALESCE(a.f, b.i, b.j)
  FROM CASE_TBL a, CASE2_TBL b;`,
				Results: []sql.Row{{10.1}, {20.2}, {-30.3}, {1}, {10.1}, {20.2}, {-30.3}, {2}, {10.1}, {20.2}, {-30.3}, {3}, {10.1}, {20.2}, {-30.3}, {2}, {10.1}, {20.2}, {-30.3}, {1}, {10.1}, {20.2}, {-30.3}, {-6}},
			},
			{
				Statement: `SELECT *
  FROM CASE_TBL a, CASE2_TBL b
  WHERE COALESCE(a.f, b.i, b.j) = 2;`,
				Results: []sql.Row{{4, ``, 2, -2}, {4, ``, 2, -4}},
			},
			{
				Statement: `SELECT NULLIF(a.i,b.i) AS "NULLIF(a.i,b.i)",
  NULLIF(b.i, 4) AS "NULLIF(b.i,4)"
  FROM CASE_TBL a, CASE2_TBL b;`,
				Results: []sql.Row{{``, 1}, {2, 1}, {3, 1}, {4, 1}, {1, 2}, {``, 2}, {3, 2}, {4, 2}, {1, 3}, {2, 3}, {``, 3}, {4, 3}, {1, 2}, {``, 2}, {3, 2}, {4, 2}, {``, 1}, {2, 1}, {3, 1}, {4, 1}, {1, ``}, {2, ``}, {3, ``}, {4, ``}},
			},
			{
				Statement: `SELECT *
  FROM CASE_TBL a, CASE2_TBL b
  WHERE COALESCE(f,b.i) = 2;`,
				Results: []sql.Row{{4, ``, 2, -2}, {4, ``, 2, -4}},
			},
			{
				Statement: `explain (costs off)
SELECT * FROM CASE_TBL WHERE NULLIF(1, 2) = 2;`,
				Results: []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off)
SELECT * FROM CASE_TBL WHERE NULLIF(1, 1) IS NOT NULL;`,
				Results: []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off)
SELECT * FROM CASE_TBL WHERE NULLIF(1, null) = 2;`,
				Results: []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `UPDATE CASE_TBL
  SET i = CASE WHEN i >= 3 THEN (- i)
                ELSE (2 * i) END;`,
			},
			{
				Statement: `SELECT * FROM CASE_TBL;`,
				Results:   []sql.Row{{2, 10.1}, {4, 20.2}, {-3, -30.3}, {-4, ``}},
			},
			{
				Statement: `UPDATE CASE_TBL
  SET i = CASE WHEN i >= 2 THEN (2 * i)
                ELSE (3 * i) END;`,
			},
			{
				Statement: `SELECT * FROM CASE_TBL;`,
				Results:   []sql.Row{{4, 10.1}, {8, 20.2}, {-9, -30.3}, {-12, ``}},
			},
			{
				Statement: `UPDATE CASE_TBL
  SET i = CASE WHEN b.i >= 2 THEN (2 * j)
                ELSE (3 * j) END
  FROM CASE2_TBL b
  WHERE j = -CASE_TBL.i;`,
			},
			{
				Statement: `SELECT * FROM CASE_TBL;`,
				Results:   []sql.Row{{8, 20.2}, {-9, -30.3}, {-12, ``}, {-8, 10.1}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE FUNCTION vol(text) returns text as
  'begin return $1; end' language plpgsql volatile;`,
			},
			{
				Statement: `SELECT CASE
  (CASE vol('bar')
    WHEN 'foo' THEN 'it was foo!'
    WHEN vol(null) THEN 'null input'
    WHEN 'bar' THEN 'it was bar!' END
  )
  WHEN 'it was foo!' THEN 'foo recognized'
  WHEN 'it was bar!' THEN 'bar recognized'
  ELSE 'unrecognized' END;`,
				Results: []sql.Row{{`bar recognized`}},
			},
			{
				Statement: `CREATE DOMAIN foodomain AS text;`,
			},
			{
				Statement: `CREATE FUNCTION volfoo(text) returns foodomain as
  'begin return $1::foodomain; end' language plpgsql volatile;`,
			},
			{
				Statement: `CREATE FUNCTION inline_eq(foodomain, foodomain) returns boolean as
  'SELECT CASE $2::text WHEN $1::text THEN true ELSE false END' language sql;`,
			},
			{
				Statement: `CREATE OPERATOR = (procedure = inline_eq,
                   leftarg = foodomain, rightarg = foodomain);`,
			},
			{
				Statement: `SELECT CASE volfoo('bar') WHEN 'foo'::foodomain THEN 'is foo' ELSE 'is not foo' END;`,
				Results:   []sql.Row{{`is not foo`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE DOMAIN arrdomain AS int[];`,
			},
			{
				Statement: `CREATE FUNCTION make_ad(int,int) returns arrdomain as
  'declare x arrdomain;`,
			},
			{
				Statement: `   begin
     x := array[$1,$2];`,
			},
			{
				Statement: `     return x;`,
			},
			{
				Statement: `   end' language plpgsql volatile;`,
			},
			{
				Statement: `CREATE FUNCTION ad_eq(arrdomain, arrdomain) returns boolean as
  'begin return array_eq($1, $2); end' language plpgsql;`,
			},
			{
				Statement: `CREATE OPERATOR = (procedure = ad_eq,
                   leftarg = arrdomain, rightarg = arrdomain);`,
			},
			{
				Statement: `SELECT CASE make_ad(1,2)
  WHEN array[2,4]::arrdomain THEN 'wrong'
  WHEN array[2,5]::arrdomain THEN 'still wrong'
  WHEN array[1,2]::arrdomain THEN 'right'
  END;`,
				Results: []sql.Row{{`right`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TYPE casetestenum AS ENUM ('e', 'f', 'g');`,
			},
			{
				Statement: `SELECT
  CASE 'foo'::text
    WHEN 'foo' THEN ARRAY['a', 'b', 'c', 'd'] || enum_range(NULL::casetestenum)::text[]
    ELSE ARRAY['x', 'y']
    END;`,
				Results: []sql.Row{{`{a,b,c,d,e,f,g}`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE CASE_TBL;`,
			},
			{
				Statement: `DROP TABLE CASE2_TBL;`,
			},
		},
	})
}
