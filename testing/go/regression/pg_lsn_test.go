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

func TestPgLsn(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_pg_lsn)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_pg_lsn,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE PG_LSN_TBL (f1 pg_lsn);`,
			},
			{
				Statement: `INSERT INTO PG_LSN_TBL VALUES ('0/0');`,
			},
			{
				Statement: `INSERT INTO PG_LSN_TBL VALUES ('FFFFFFFF/FFFFFFFF');`,
			},
			{
				Statement:   `INSERT INTO PG_LSN_TBL VALUES ('G/0');`,
				ErrorString: `invalid input syntax for type pg_lsn: "G/0"`,
			},
			{
				Statement:   `INSERT INTO PG_LSN_TBL VALUES ('-1/0');`,
				ErrorString: `invalid input syntax for type pg_lsn: "-1/0"`,
			},
			{
				Statement:   `INSERT INTO PG_LSN_TBL VALUES (' 0/12345678');`,
				ErrorString: `invalid input syntax for type pg_lsn: " 0/12345678"`,
			},
			{
				Statement:   `INSERT INTO PG_LSN_TBL VALUES ('ABCD/');`,
				ErrorString: `invalid input syntax for type pg_lsn: "ABCD/"`,
			},
			{
				Statement:   `INSERT INTO PG_LSN_TBL VALUES ('/ABCD');`,
				ErrorString: `invalid input syntax for type pg_lsn: "/ABCD"`,
			},
			{
				Statement: `SELECT MIN(f1), MAX(f1) FROM PG_LSN_TBL;`,
				Results:   []sql.Row{{`0/0`, `FFFFFFFF/FFFFFFFF`}},
			},
			{
				Statement: `DROP TABLE PG_LSN_TBL;`,
			},
			{
				Statement: `SELECT '0/16AE7F8' = '0/16AE7F8'::pg_lsn;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '0/16AE7F8'::pg_lsn != '0/16AE7F7';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '0/16AE7F7' < '0/16AE7F8'::pg_lsn;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '0/16AE7F8' > pg_lsn '0/16AE7F7';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT '0/16AE7F7'::pg_lsn - '0/16AE7F8'::pg_lsn;`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT '0/16AE7F8'::pg_lsn - '0/16AE7F7'::pg_lsn;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT '0/16AE7F7'::pg_lsn + 16::numeric;`,
				Results:   []sql.Row{{`0/16AE807`}},
			},
			{
				Statement: `SELECT 16::numeric + '0/16AE7F7'::pg_lsn;`,
				Results:   []sql.Row{{`0/16AE807`}},
			},
			{
				Statement: `SELECT '0/16AE7F7'::pg_lsn - 16::numeric;`,
				Results:   []sql.Row{{`0/16AE7E7`}},
			},
			{
				Statement: `SELECT 'FFFFFFFF/FFFFFFFE'::pg_lsn + 1::numeric;`,
				Results:   []sql.Row{{`FFFFFFFF/FFFFFFFF`}},
			},
			{
				Statement:   `SELECT 'FFFFFFFF/FFFFFFFE'::pg_lsn + 2::numeric; -- out of range error`,
				ErrorString: `pg_lsn out of range`,
			},
			{
				Statement: `SELECT '0/1'::pg_lsn - 1::numeric;`,
				Results:   []sql.Row{{`0/0`}},
			},
			{
				Statement:   `SELECT '0/1'::pg_lsn - 2::numeric; -- out of range error`,
				ErrorString: `pg_lsn out of range`,
			},
			{
				Statement: `SELECT '0/0'::pg_lsn + ('FFFFFFFF/FFFFFFFF'::pg_lsn - '0/0'::pg_lsn);`,
				Results:   []sql.Row{{`FFFFFFFF/FFFFFFFF`}},
			},
			{
				Statement: `SELECT 'FFFFFFFF/FFFFFFFF'::pg_lsn - ('FFFFFFFF/FFFFFFFF'::pg_lsn - '0/0'::pg_lsn);`,
				Results:   []sql.Row{{`0/0`}},
			},
			{
				Statement:   `SELECT '0/16AE7F7'::pg_lsn + 'NaN'::numeric;`,
				ErrorString: `cannot add NaN to pg_lsn`,
			},
			{
				Statement:   `SELECT '0/16AE7F7'::pg_lsn - 'NaN'::numeric;`,
				ErrorString: `cannot subtract NaN from pg_lsn`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT DISTINCT (i || '/' || j)::pg_lsn f
  FROM generate_series(1, 10) i,
       generate_series(1, 10) j,
       generate_series(1, 5) k
  WHERE i <= 10 AND j > 0 AND j <= 10
  ORDER BY f;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: (((((i.i)::text || '/'::text) || (j.j)::text))::pg_lsn)`}, {`->  HashAggregate`}, {`Group Key: ((((i.i)::text || '/'::text) || (j.j)::text))::pg_lsn`}, {`->  Nested Loop`}, {`->  Function Scan on generate_series k`}, {`->  Materialize`}, {`->  Nested Loop`}, {`->  Function Scan on generate_series j`}, {`Filter: ((j > 0) AND (j <= 10))`}, {`->  Function Scan on generate_series i`}, {`Filter: (i <= 10)`}},
			},
			{
				Statement: `SELECT DISTINCT (i || '/' || j)::pg_lsn f
  FROM generate_series(1, 10) i,
       generate_series(1, 10) j,
       generate_series(1, 5) k
  WHERE i <= 10 AND j > 0 AND j <= 10
  ORDER BY f;`,
				Results: []sql.Row{{`1/1`}, {`1/2`}, {`1/3`}, {`1/4`}, {`1/5`}, {`1/6`}, {`1/7`}, {`1/8`}, {`1/9`}, {`1/10`}, {`2/1`}, {`2/2`}, {`2/3`}, {`2/4`}, {`2/5`}, {`2/6`}, {`2/7`}, {`2/8`}, {`2/9`}, {`2/10`}, {`3/1`}, {`3/2`}, {`3/3`}, {`3/4`}, {`3/5`}, {`3/6`}, {`3/7`}, {`3/8`}, {`3/9`}, {`3/10`}, {`4/1`}, {`4/2`}, {`4/3`}, {`4/4`}, {`4/5`}, {`4/6`}, {`4/7`}, {`4/8`}, {`4/9`}, {`4/10`}, {`5/1`}, {`5/2`}, {`5/3`}, {`5/4`}, {`5/5`}, {`5/6`}, {`5/7`}, {`5/8`}, {`5/9`}, {`5/10`}, {`6/1`}, {`6/2`}, {`6/3`}, {`6/4`}, {`6/5`}, {`6/6`}, {`6/7`}, {`6/8`}, {`6/9`}, {`6/10`}, {`7/1`}, {`7/2`}, {`7/3`}, {`7/4`}, {`7/5`}, {`7/6`}, {`7/7`}, {`7/8`}, {`7/9`}, {`7/10`}, {`8/1`}, {`8/2`}, {`8/3`}, {`8/4`}, {`8/5`}, {`8/6`}, {`8/7`}, {`8/8`}, {`8/9`}, {`8/10`}, {`9/1`}, {`9/2`}, {`9/3`}, {`9/4`}, {`9/5`}, {`9/6`}, {`9/7`}, {`9/8`}, {`9/9`}, {`9/10`}, {`10/1`}, {`10/2`}, {`10/3`}, {`10/4`}, {`10/5`}, {`10/6`}, {`10/7`}, {`10/8`}, {`10/9`}, {`10/10`}},
			},
		},
	})
}
