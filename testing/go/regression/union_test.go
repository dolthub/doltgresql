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

func TestUnion(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_union)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_union,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT 1 AS two UNION SELECT 2 ORDER BY 1;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `SELECT 1 AS one UNION SELECT 1 ORDER BY 1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT 1 AS two UNION ALL SELECT 2;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `SELECT 1 AS two UNION ALL SELECT 1;`,
				Results:   []sql.Row{{1}, {1}},
			},
			{
				Statement: `SELECT 1 AS three UNION SELECT 2 UNION SELECT 3 ORDER BY 1;`,
				Results:   []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement: `SELECT 1 AS two UNION SELECT 2 UNION SELECT 2 ORDER BY 1;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `SELECT 1 AS three UNION SELECT 2 UNION ALL SELECT 2 ORDER BY 1;`,
				Results:   []sql.Row{{1}, {2}, {2}},
			},
			{
				Statement: `SELECT 1.1 AS two UNION SELECT 2.2 ORDER BY 1;`,
				Results:   []sql.Row{{1.1}, {2.2}},
			},
			{
				Statement: `SELECT 1.1 AS two UNION SELECT 2 ORDER BY 1;`,
				Results:   []sql.Row{{1.1}, {2}},
			},
			{
				Statement: `SELECT 1 AS two UNION SELECT 2.2 ORDER BY 1;`,
				Results:   []sql.Row{{1}, {2.2}},
			},
			{
				Statement: `SELECT 1 AS one UNION SELECT 1.0::float8 ORDER BY 1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT 1.1 AS two UNION ALL SELECT 2 ORDER BY 1;`,
				Results:   []sql.Row{{1.1}, {2}},
			},
			{
				Statement: `SELECT 1.0::float8 AS two UNION ALL SELECT 1 ORDER BY 1;`,
				Results:   []sql.Row{{1}, {1}},
			},
			{
				Statement: `SELECT 1.1 AS three UNION SELECT 2 UNION SELECT 3 ORDER BY 1;`,
				Results:   []sql.Row{{1.1}, {2}, {3}},
			},
			{
				Statement: `SELECT 1.1::float8 AS two UNION SELECT 2 UNION SELECT 2.0::float8 ORDER BY 1;`,
				Results:   []sql.Row{{1.1}, {2}},
			},
			{
				Statement: `SELECT 1.1 AS three UNION SELECT 2 UNION ALL SELECT 2 ORDER BY 1;`,
				Results:   []sql.Row{{1.1}, {2}, {2}},
			},
			{
				Statement: `SELECT 1.1 AS two UNION (SELECT 2 UNION ALL SELECT 2) ORDER BY 1;`,
				Results:   []sql.Row{{1.1}, {2}},
			},
			{
				Statement: `SELECT f1 AS five FROM FLOAT8_TBL
UNION
SELECT f1 FROM FLOAT8_TBL
ORDER BY 1;`,
				Results: []sql.Row{{-1.2345678901234e+200}, {-1004.3}, {-34.84}, {-1.2345678901234e-200}, {0}},
			},
			{
				Statement: `SELECT f1 AS ten FROM FLOAT8_TBL
UNION ALL
SELECT f1 FROM FLOAT8_TBL;`,
				Results: []sql.Row{{0}, {-34.84}, {-1004.3}, {-1.2345678901234e+200}, {-1.2345678901234e-200}, {0}, {-34.84}, {-1004.3}, {-1.2345678901234e+200}, {-1.2345678901234e-200}},
			},
			{
				Statement: `SELECT f1 AS nine FROM FLOAT8_TBL
UNION
SELECT f1 FROM INT4_TBL
ORDER BY 1;`,
				Results: []sql.Row{{-1.2345678901234e+200}, {-2147483647}, {-123456}, {-1004.3}, {-34.84}, {-1.2345678901234e-200}, {0}, {123456}, {2147483647}},
			},
			{
				Statement: `SELECT f1 AS ten FROM FLOAT8_TBL
UNION ALL
SELECT f1 FROM INT4_TBL;`,
				Results: []sql.Row{{0}, {-34.84}, {-1004.3}, {-1.2345678901234e+200}, {-1.2345678901234e-200}, {0}, {123456}, {-123456}, {2147483647}, {-2147483647}},
			},
			{
				Statement: `SELECT f1 AS five FROM FLOAT8_TBL
  WHERE f1 BETWEEN -1e6 AND 1e6
UNION
SELECT f1 FROM INT4_TBL
  WHERE f1 BETWEEN 0 AND 1000000
ORDER BY 1;`,
				Results: []sql.Row{{-1004.3}, {-34.84}, {-1.2345678901234e-200}, {0}, {123456}},
			},
			{
				Statement: `SELECT CAST(f1 AS char(4)) AS three FROM VARCHAR_TBL
UNION
SELECT f1 FROM CHAR_TBL
ORDER BY 1;`,
				Results: []sql.Row{{`a`}, {`ab`}, {`abcd`}},
			},
			{
				Statement: `SELECT f1 AS three FROM VARCHAR_TBL
UNION
SELECT CAST(f1 AS varchar) FROM CHAR_TBL
ORDER BY 1;`,
				Results: []sql.Row{{`a`}, {`ab`}, {`abcd`}},
			},
			{
				Statement: `SELECT f1 AS eight FROM VARCHAR_TBL
UNION ALL
SELECT f1 FROM CHAR_TBL;`,
				Results: []sql.Row{{`a`}, {`ab`}, {`abcd`}, {`abcd`}, {`a`}, {`ab`}, {`abcd`}, {`abcd`}},
			},
			{
				Statement: `SELECT f1 AS five FROM TEXT_TBL
UNION
SELECT f1 FROM VARCHAR_TBL
UNION
SELECT TRIM(TRAILING FROM f1) FROM CHAR_TBL
ORDER BY 1;`,
				Results: []sql.Row{{`a`}, {`ab`}, {`abcd`}, {`doh!`}, {`hi de ho neighbor`}},
			},
			{
				Statement: `SELECT q2 FROM int8_tbl INTERSECT SELECT q1 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{123}, {4567890123456789}},
			},
			{
				Statement: `SELECT q2 FROM int8_tbl INTERSECT ALL SELECT q1 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{123}, {4567890123456789}, {4567890123456789}},
			},
			{
				Statement: `SELECT q2 FROM int8_tbl EXCEPT SELECT q1 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{-4567890123456789}, {456}},
			},
			{
				Statement: `SELECT q2 FROM int8_tbl EXCEPT ALL SELECT q1 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{-4567890123456789}, {456}},
			},
			{
				Statement: `SELECT q2 FROM int8_tbl EXCEPT ALL SELECT DISTINCT q1 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{-4567890123456789}, {456}, {4567890123456789}},
			},
			{
				Statement: `SELECT q1 FROM int8_tbl EXCEPT SELECT q2 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT q1 FROM int8_tbl EXCEPT ALL SELECT q2 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{123}, {4567890123456789}},
			},
			{
				Statement: `SELECT q1 FROM int8_tbl EXCEPT ALL SELECT DISTINCT q2 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{123}, {4567890123456789}, {4567890123456789}},
			},
			{
				Statement:   `SELECT q1 FROM int8_tbl EXCEPT ALL SELECT q1 FROM int8_tbl FOR NO KEY UPDATE;`,
				ErrorString: `FOR NO KEY UPDATE is not allowed with UNION/INTERSECT/EXCEPT`,
			},
			{
				Statement: `(SELECT 1,2,3 UNION SELECT 4,5,6) INTERSECT SELECT 4,5,6;`,
				Results:   []sql.Row{{4, 5, 6}},
			},
			{
				Statement: `(SELECT 1,2,3 UNION SELECT 4,5,6 ORDER BY 1,2) INTERSECT SELECT 4,5,6;`,
				Results:   []sql.Row{{4, 5, 6}},
			},
			{
				Statement: `(SELECT 1,2,3 UNION SELECT 4,5,6) EXCEPT SELECT 4,5,6;`,
				Results:   []sql.Row{{1, 2, 3}},
			},
			{
				Statement: `(SELECT 1,2,3 UNION SELECT 4,5,6 ORDER BY 1,2) EXCEPT SELECT 4,5,6;`,
				Results:   []sql.Row{{1, 2, 3}},
			},
			{
				Statement: `set enable_hashagg to on;`,
			},
			{
				Statement: `explain (costs off)
select count(*) from
  ( select unique1 from tenk1 union select fivethous from tenk1 ) ss;`,
				Results: []sql.Row{{`Aggregate`}, {`->  HashAggregate`}, {`Group Key: tenk1.unique1`}, {`->  Append`}, {`->  Index Only Scan using tenk1_unique1 on tenk1`}, {`->  Seq Scan on tenk1 tenk1_1`}},
			},
			{
				Statement: `select count(*) from
  ( select unique1 from tenk1 union select fivethous from tenk1 ) ss;`,
				Results: []sql.Row{{10000}},
			},
			{
				Statement: `explain (costs off)
select count(*) from
  ( select unique1 from tenk1 intersect select fivethous from tenk1 ) ss;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Subquery Scan on ss`}, {`->  HashSetOp Intersect`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Seq Scan on tenk1`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 tenk1_1`}},
			},
			{
				Statement: `select count(*) from
  ( select unique1 from tenk1 intersect select fivethous from tenk1 ) ss;`,
				Results: []sql.Row{{5000}},
			},
			{
				Statement: `explain (costs off)
select unique1 from tenk1 except select unique2 from tenk1 where unique2 != 10;`,
				Results: []sql.Row{{`HashSetOp Except`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Index Only Scan using tenk1_unique1 on tenk1`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Index Only Scan using tenk1_unique2 on tenk1 tenk1_1`}, {`Filter: (unique2 <> 10)`}},
			},
			{
				Statement: `select unique1 from tenk1 except select unique2 from tenk1 where unique2 != 10;`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `set enable_hashagg to off;`,
			},
			{
				Statement: `explain (costs off)
select count(*) from
  ( select unique1 from tenk1 union select fivethous from tenk1 ) ss;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Unique`}, {`->  Sort`}, {`Sort Key: tenk1.unique1`}, {`->  Append`}, {`->  Index Only Scan using tenk1_unique1 on tenk1`}, {`->  Seq Scan on tenk1 tenk1_1`}},
			},
			{
				Statement: `select count(*) from
  ( select unique1 from tenk1 union select fivethous from tenk1 ) ss;`,
				Results: []sql.Row{{10000}},
			},
			{
				Statement: `explain (costs off)
select count(*) from
  ( select unique1 from tenk1 intersect select fivethous from tenk1 ) ss;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Subquery Scan on ss`}, {`->  SetOp Intersect`}, {`->  Sort`}, {`Sort Key: "*SELECT* 2".fivethous`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Seq Scan on tenk1`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 tenk1_1`}},
			},
			{
				Statement: `select count(*) from
  ( select unique1 from tenk1 intersect select fivethous from tenk1 ) ss;`,
				Results: []sql.Row{{5000}},
			},
			{
				Statement: `explain (costs off)
select unique1 from tenk1 except select unique2 from tenk1 where unique2 != 10;`,
				Results: []sql.Row{{`SetOp Except`}, {`->  Sort`}, {`Sort Key: "*SELECT* 1".unique1`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Index Only Scan using tenk1_unique1 on tenk1`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Index Only Scan using tenk1_unique2 on tenk1 tenk1_1`}, {`Filter: (unique2 <> 10)`}},
			},
			{
				Statement: `select unique1 from tenk1 except select unique2 from tenk1 where unique2 != 10;`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `reset enable_hashagg;`,
			},
			{
				Statement: `set enable_hashagg to on;`,
			},
			{
				Statement: `explain (costs off)
select x from (values (100::money), (200::money)) _(x) union select x from (values (100::money), (300::money)) _(x);`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `set enable_hashagg to off;`,
			},
			{
				Statement: `explain (costs off)
select x from (values (100::money), (200::money)) _(x) union select x from (values (100::money), (300::money)) _(x);`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `reset enable_hashagg;`,
			},
			{
				Statement: `set enable_hashagg to on;`,
			},
			{
				Statement: `explain (costs off)
select x from (values (array[1, 2]), (array[1, 3])) _(x) union select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results: []sql.Row{{`HashAggregate`}, {`Group Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (array[1, 2]), (array[1, 3])) _(x) union select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results:   []sql.Row{{`{1,4}`}, {`{1,2}`}, {`{1,3}`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (array[1, 2]), (array[1, 3])) _(x) intersect select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results: []sql.Row{{`HashSetOp Intersect`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Values Scan on "*VALUES*"`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (array[1, 2]), (array[1, 3])) _(x) intersect select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results:   []sql.Row{{`{1,2}`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (array[1, 2]), (array[1, 3])) _(x) except select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results: []sql.Row{{`HashSetOp Except`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Values Scan on "*VALUES*"`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (array[1, 2]), (array[1, 3])) _(x) except select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results:   []sql.Row{{`{1,3}`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (array[100::money]), (array[200::money])) _(x) union select x from (values (array[100::money]), (array[300::money])) _(x);`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (array[100::money]), (array[200::money])) _(x) union select x from (values (array[100::money]), (array[300::money])) _(x);`,
				Results:   []sql.Row{{`{$100.00}`}, {`{$200.00}`}, {`{$300.00}`}},
			},
			{
				Statement: `set enable_hashagg to off;`,
			},
			{
				Statement: `explain (costs off)
select x from (values (array[1, 2]), (array[1, 3])) _(x) union select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (array[1, 2]), (array[1, 3])) _(x) union select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results:   []sql.Row{{`{1,2}`}, {`{1,3}`}, {`{1,4}`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (array[1, 2]), (array[1, 3])) _(x) intersect select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results: []sql.Row{{`SetOp Intersect`}, {`->  Sort`}, {`Sort Key: "*SELECT* 1".x`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Values Scan on "*VALUES*"`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (array[1, 2]), (array[1, 3])) _(x) intersect select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results:   []sql.Row{{`{1,2}`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (array[1, 2]), (array[1, 3])) _(x) except select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results: []sql.Row{{`SetOp Except`}, {`->  Sort`}, {`Sort Key: "*SELECT* 1".x`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Values Scan on "*VALUES*"`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (array[1, 2]), (array[1, 3])) _(x) except select x from (values (array[1, 2]), (array[1, 4])) _(x);`,
				Results:   []sql.Row{{`{1,3}`}},
			},
			{
				Statement: `reset enable_hashagg;`,
			},
			{
				Statement: `set enable_hashagg to on;`,
			},
			{
				Statement: `explain (costs off)
select x from (values (row(1, 2)), (row(1, 3))) _(x) union select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (row(1, 2)), (row(1, 3))) _(x) union select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results:   []sql.Row{{`(1,2)`}, {`(1,3)`}, {`(1,4)`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (row(1, 2)), (row(1, 3))) _(x) intersect select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results: []sql.Row{{`SetOp Intersect`}, {`->  Sort`}, {`Sort Key: "*SELECT* 1".x`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Values Scan on "*VALUES*"`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (row(1, 2)), (row(1, 3))) _(x) intersect select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results:   []sql.Row{{`(1,2)`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (row(1, 2)), (row(1, 3))) _(x) except select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results: []sql.Row{{`SetOp Except`}, {`->  Sort`}, {`Sort Key: "*SELECT* 1".x`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Values Scan on "*VALUES*"`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (row(1, 2)), (row(1, 3))) _(x) except select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results:   []sql.Row{{`(1,3)`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (row(100::money)), (row(200::money))) _(x) union select x from (values (row(100::money)), (row(300::money))) _(x);`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (row(100::money)), (row(200::money))) _(x) union select x from (values (row(100::money)), (row(300::money))) _(x);`,
				Results:   []sql.Row{{`($100.00)`}, {`($200.00)`}, {`($300.00)`}},
			},
			{
				Statement: `create type ct1 as (f1 money);`,
			},
			{
				Statement: `explain (costs off)
select x from (values (row(100::money)::ct1), (row(200::money)::ct1)) _(x) union select x from (values (row(100::money)::ct1), (row(300::money)::ct1)) _(x);`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (row(100::money)::ct1), (row(200::money)::ct1)) _(x) union select x from (values (row(100::money)::ct1), (row(300::money)::ct1)) _(x);`,
				Results:   []sql.Row{{`($100.00)`}, {`($200.00)`}, {`($300.00)`}},
			},
			{
				Statement: `drop type ct1;`,
			},
			{
				Statement: `set enable_hashagg to off;`,
			},
			{
				Statement: `explain (costs off)
select x from (values (row(1, 2)), (row(1, 3))) _(x) union select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: "*VALUES*".column1`}, {`->  Append`}, {`->  Values Scan on "*VALUES*"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (row(1, 2)), (row(1, 3))) _(x) union select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results:   []sql.Row{{`(1,2)`}, {`(1,3)`}, {`(1,4)`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (row(1, 2)), (row(1, 3))) _(x) intersect select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results: []sql.Row{{`SetOp Intersect`}, {`->  Sort`}, {`Sort Key: "*SELECT* 1".x`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Values Scan on "*VALUES*"`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (row(1, 2)), (row(1, 3))) _(x) intersect select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results:   []sql.Row{{`(1,2)`}},
			},
			{
				Statement: `explain (costs off)
select x from (values (row(1, 2)), (row(1, 3))) _(x) except select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results: []sql.Row{{`SetOp Except`}, {`->  Sort`}, {`Sort Key: "*SELECT* 1".x`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Values Scan on "*VALUES*"`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Values Scan on "*VALUES*_1"`}},
			},
			{
				Statement: `select x from (values (row(1, 2)), (row(1, 3))) _(x) except select x from (values (row(1, 2)), (row(1, 4))) _(x);`,
				Results:   []sql.Row{{`(1,3)`}},
			},
			{
				Statement: `reset enable_hashagg;`,
			},
			{
				Statement: `SELECT f1 FROM float8_tbl INTERSECT SELECT f1 FROM int4_tbl ORDER BY 1;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT f1 FROM float8_tbl EXCEPT SELECT f1 FROM int4_tbl ORDER BY 1;`,
				Results:   []sql.Row{{-1.2345678901234e+200}, {-1004.3}, {-34.84}, {-1.2345678901234e-200}},
			},
			{
				Statement: `SELECT q1 FROM int8_tbl INTERSECT SELECT q2 FROM int8_tbl UNION ALL SELECT q2 FROM int8_tbl  ORDER BY 1;`,
				Results:   []sql.Row{{-4567890123456789}, {123}, {123}, {456}, {4567890123456789}, {4567890123456789}, {4567890123456789}},
			},
			{
				Statement: `SELECT q1 FROM int8_tbl INTERSECT (((SELECT q2 FROM int8_tbl UNION ALL SELECT q2 FROM int8_tbl))) ORDER BY 1;`,
				Results:   []sql.Row{{123}, {4567890123456789}},
			},
			{
				Statement: `(((SELECT q1 FROM int8_tbl INTERSECT SELECT q2 FROM int8_tbl ORDER BY 1))) UNION ALL SELECT q2 FROM int8_tbl;`,
				Results:   []sql.Row{{123}, {4567890123456789}, {456}, {4567890123456789}, {123}, {4567890123456789}, {-4567890123456789}},
			},
			{
				Statement: `SELECT q1 FROM int8_tbl UNION ALL SELECT q2 FROM int8_tbl EXCEPT SELECT q1 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{-4567890123456789}, {456}},
			},
			{
				Statement: `SELECT q1 FROM int8_tbl UNION ALL (((SELECT q2 FROM int8_tbl EXCEPT SELECT q1 FROM int8_tbl ORDER BY 1)));`,
				Results:   []sql.Row{{123}, {123}, {4567890123456789}, {4567890123456789}, {4567890123456789}, {-4567890123456789}, {456}},
			},
			{
				Statement: `(((SELECT q1 FROM int8_tbl UNION ALL SELECT q2 FROM int8_tbl))) EXCEPT SELECT q1 FROM int8_tbl ORDER BY 1;`,
				Results:   []sql.Row{{-4567890123456789}, {456}},
			},
			{
				Statement: `SELECT q1,q2 FROM int8_tbl EXCEPT SELECT q2,q1 FROM int8_tbl
ORDER BY q2,q1;`,
				Results: []sql.Row{{4567890123456789, -4567890123456789}, {123, 456}},
			},
			{
				Statement:   `SELECT q1 FROM int8_tbl EXCEPT SELECT q2 FROM int8_tbl ORDER BY q2 LIMIT 1;`,
				ErrorString: `column "q2" does not exist`,
			},
			{
				Statement: `SELECT q1 FROM int8_tbl EXCEPT (((SELECT q2 FROM int8_tbl ORDER BY q2 LIMIT 1))) ORDER BY 1;`,
				Results:   []sql.Row{{123}, {4567890123456789}},
			},
			{
				Statement: `(((((select * from int8_tbl)))));`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `select union select;`,
			},
			{
				Statement: `(1 row)
select intersect select;`,
			},
			{
				Statement: `(1 row)
select except select;`,
			},
			{
				Statement: `(0 rows)
set enable_hashagg = true;`,
			},
			{
				Statement: `set enable_sort = false;`,
			},
			{
				Statement: `explain (costs off)
select from generate_series(1,5) union select from generate_series(1,3);`,
				Results: []sql.Row{{`HashAggregate`}, {`->  Append`}, {`->  Function Scan on generate_series`}, {`->  Function Scan on generate_series generate_series_1`}},
			},
			{
				Statement: `explain (costs off)
select from generate_series(1,5) intersect select from generate_series(1,3);`,
				Results: []sql.Row{{`HashSetOp Intersect`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Function Scan on generate_series`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Function Scan on generate_series generate_series_1`}},
			},
			{
				Statement: `select from generate_series(1,5) union select from generate_series(1,3);`,
			},
			{
				Statement: `(1 row)
select from generate_series(1,5) union all select from generate_series(1,3);`,
			},
			{
				Statement: `(8 rows)
select from generate_series(1,5) intersect select from generate_series(1,3);`,
			},
			{
				Statement: `(1 row)
select from generate_series(1,5) intersect all select from generate_series(1,3);`,
			},
			{
				Statement: `(3 rows)
select from generate_series(1,5) except select from generate_series(1,3);`,
			},
			{
				Statement: `(0 rows)
select from generate_series(1,5) except all select from generate_series(1,3);`,
			},
			{
				Statement: `(2 rows)
set enable_hashagg = false;`,
			},
			{
				Statement: `set enable_sort = true;`,
			},
			{
				Statement: `explain (costs off)
select from generate_series(1,5) union select from generate_series(1,3);`,
				Results: []sql.Row{{`Unique`}, {`->  Append`}, {`->  Function Scan on generate_series`}, {`->  Function Scan on generate_series generate_series_1`}},
			},
			{
				Statement: `explain (costs off)
select from generate_series(1,5) intersect select from generate_series(1,3);`,
				Results: []sql.Row{{`SetOp Intersect`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Function Scan on generate_series`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Function Scan on generate_series generate_series_1`}},
			},
			{
				Statement: `select from generate_series(1,5) union select from generate_series(1,3);`,
			},
			{
				Statement: `(1 row)
select from generate_series(1,5) union all select from generate_series(1,3);`,
			},
			{
				Statement: `(8 rows)
select from generate_series(1,5) intersect select from generate_series(1,3);`,
			},
			{
				Statement: `(1 row)
select from generate_series(1,5) intersect all select from generate_series(1,3);`,
			},
			{
				Statement: `(3 rows)
select from generate_series(1,5) except select from generate_series(1,3);`,
			},
			{
				Statement: `(0 rows)
select from generate_series(1,5) except all select from generate_series(1,3);`,
			},
			{
				Statement: `(2 rows)
reset enable_hashagg;`,
			},
			{
				Statement: `reset enable_sort;`,
			},
			{
				Statement: `SELECT a.f1 FROM (SELECT 'test' AS f1 FROM varchar_tbl) a
UNION
SELECT b.f1 FROM (SELECT f1 FROM varchar_tbl) b
ORDER BY 1;`,
				Results: []sql.Row{{`a`}, {`ab`}, {`abcd`}, {`test`}},
			},
			{
				Statement:   `SELECT '3.4'::numeric UNION SELECT 'foo';`,
				ErrorString: `invalid input syntax for type numeric: "foo"`,
			},
			{
				Statement: `CREATE TEMP TABLE t1 (a text, b text);`,
			},
			{
				Statement: `CREATE INDEX t1_ab_idx on t1 ((a || b));`,
			},
			{
				Statement: `CREATE TEMP TABLE t2 (ab text primary key);`,
			},
			{
				Statement: `INSERT INTO t1 VALUES ('a', 'b'), ('x', 'y');`,
			},
			{
				Statement: `INSERT INTO t2 VALUES ('ab'), ('xy');`,
			},
			{
				Statement: `set enable_seqscan = off;`,
			},
			{
				Statement: `set enable_indexscan = on;`,
			},
			{
				Statement: `set enable_bitmapscan = off;`,
			},
			{
				Statement: `explain (costs off)
 SELECT * FROM
 (SELECT a || b AS ab FROM t1
  UNION ALL
  SELECT * FROM t2) t
 WHERE ab = 'ab';`,
				Results: []sql.Row{{`Append`}, {`->  Index Scan using t1_ab_idx on t1`}, {`Index Cond: ((a || b) = 'ab'::text)`}, {`->  Index Only Scan using t2_pkey on t2`}, {`Index Cond: (ab = 'ab'::text)`}},
			},
			{
				Statement: `explain (costs off)
 SELECT * FROM
 (SELECT a || b AS ab FROM t1
  UNION
  SELECT * FROM t2) t
 WHERE ab = 'ab';`,
				Results: []sql.Row{{`HashAggregate`}, {`Group Key: ((t1.a || t1.b))`}, {`->  Append`}, {`->  Index Scan using t1_ab_idx on t1`}, {`Index Cond: ((a || b) = 'ab'::text)`}, {`->  Index Only Scan using t2_pkey on t2`}, {`Index Cond: (ab = 'ab'::text)`}},
			},
			{
				Statement: `CREATE TEMP TABLE t1c (b text, a text);`,
			},
			{
				Statement: `ALTER TABLE t1c INHERIT t1;`,
			},
			{
				Statement: `CREATE TEMP TABLE t2c (primary key (ab)) INHERITS (t2);`,
			},
			{
				Statement: `INSERT INTO t1c VALUES ('v', 'w'), ('c', 'd'), ('m', 'n'), ('e', 'f');`,
			},
			{
				Statement: `INSERT INTO t2c VALUES ('vw'), ('cd'), ('mn'), ('ef');`,
			},
			{
				Statement: `CREATE INDEX t1c_ab_idx on t1c ((a || b));`,
			},
			{
				Statement: `set enable_seqscan = on;`,
			},
			{
				Statement: `set enable_indexonlyscan = off;`,
			},
			{
				Statement: `explain (costs off)
  SELECT * FROM
  (SELECT a || b AS ab FROM t1
   UNION ALL
   SELECT ab FROM t2) t
  ORDER BY 1 LIMIT 8;`,
				Results: []sql.Row{{`Limit`}, {`->  Merge Append`}, {`Sort Key: ((t1.a || t1.b))`}, {`->  Index Scan using t1_ab_idx on t1`}, {`->  Index Scan using t1c_ab_idx on t1c t1_1`}, {`->  Index Scan using t2_pkey on t2`}, {`->  Index Scan using t2c_pkey on t2c t2_1`}},
			},
			{
				Statement: `  SELECT * FROM
  (SELECT a || b AS ab FROM t1
   UNION ALL
   SELECT ab FROM t2) t
  ORDER BY 1 LIMIT 8;`,
				Results: []sql.Row{{`ab`}, {`ab`}, {`cd`}, {`dc`}, {`ef`}, {`fe`}, {`mn`}, {`nm`}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_indexscan;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `create table events (event_id int primary key);`,
			},
			{
				Statement: `create table other_events (event_id int primary key);`,
			},
			{
				Statement: `create table events_child () inherits (events);`,
			},
			{
				Statement: `explain (costs off)
select event_id
 from (select event_id from events
       union all
       select event_id from other_events) ss
 order by event_id;`,
				Results: []sql.Row{{`Merge Append`}, {`Sort Key: events.event_id`}, {`->  Index Scan using events_pkey on events`}, {`->  Sort`}, {`Sort Key: events_1.event_id`}, {`->  Seq Scan on events_child events_1`}, {`->  Index Scan using other_events_pkey on other_events`}},
			},
			{
				Statement: `drop table events_child, events, other_events;`,
			},
			{
				Statement: `reset enable_indexonlyscan;`,
			},
			{
				Statement: `explain (costs off)
 SELECT * FROM
  (SELECT 1 AS t, * FROM tenk1 a
   UNION ALL
   SELECT 2 AS t, * FROM tenk1 b) c
 WHERE t = 2;`,
				Results: []sql.Row{{`Seq Scan on tenk1 b`}},
			},
			{
				Statement: `explain (costs off)
SELECT * FROM
  (SELECT 1 AS t, 2 AS x
   UNION
   SELECT 2 AS t, 4 AS x) ss
WHERE x < 4
ORDER BY x;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: (2)`}, {`->  Unique`}, {`->  Sort`}, {`Sort Key: (1), (2)`}, {`->  Append`}, {`->  Result`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `SELECT * FROM
  (SELECT 1 AS t, 2 AS x
   UNION
   SELECT 2 AS t, 4 AS x) ss
WHERE x < 4
ORDER BY x;`,
				Results: []sql.Row{{1, 2}},
			},
			{
				Statement: `explain (costs off)
SELECT * FROM
  (SELECT 1 AS t, generate_series(1,10) AS x
   UNION
   SELECT 2 AS t, 4 AS x) ss
WHERE x < 4
ORDER BY x;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: ss.x`}, {`->  Subquery Scan on ss`}, {`Filter: (ss.x < 4)`}, {`->  HashAggregate`}, {`Group Key: (1), (generate_series(1, 10))`}, {`->  Append`}, {`->  ProjectSet`}, {`->  Result`}, {`->  Result`}},
			},
			{
				Statement: `SELECT * FROM
  (SELECT 1 AS t, generate_series(1,10) AS x
   UNION
   SELECT 2 AS t, 4 AS x) ss
WHERE x < 4
ORDER BY x;`,
				Results: []sql.Row{{1, 1}, {1, 2}, {1, 3}},
			},
			{
				Statement: `explain (costs off)
SELECT * FROM
  (SELECT 1 AS t, (random()*3)::int AS x
   UNION
   SELECT 2 AS t, 4 AS x) ss
WHERE x > 3
ORDER BY x;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: ss.x`}, {`->  Subquery Scan on ss`}, {`Filter: (ss.x > 3)`}, {`->  Unique`}, {`->  Sort`}, {`Sort Key: (1), (((random() * '3'::double precision))::integer)`}, {`->  Append`}, {`->  Result`}, {`->  Result`}},
			},
			{
				Statement: `SELECT * FROM
  (SELECT 1 AS t, (random()*3)::int AS x
   UNION
   SELECT 2 AS t, 4 AS x) ss
WHERE x > 3
ORDER BY x;`,
				Results: []sql.Row{{2, 4}},
			},
			{
				Statement: `explain (costs off)
select distinct q1 from
  (select distinct * from int8_tbl i81
   union all
   select distinct * from int8_tbl i82) ss
where q2 = q2;`,
				Results: []sql.Row{{`Unique`}, {`->  Merge Append`}, {`Sort Key: "*SELECT* 1".q1`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Unique`}, {`->  Sort`}, {`Sort Key: i81.q1, i81.q2`}, {`->  Seq Scan on int8_tbl i81`}, {`Filter: (q2 IS NOT NULL)`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Unique`}, {`->  Sort`}, {`Sort Key: i82.q1, i82.q2`}, {`->  Seq Scan on int8_tbl i82`}, {`Filter: (q2 IS NOT NULL)`}},
			},
			{
				Statement: `select distinct q1 from
  (select distinct * from int8_tbl i81
   union all
   select distinct * from int8_tbl i82) ss
where q2 = q2;`,
				Results: []sql.Row{{123}, {4567890123456789}},
			},
			{
				Statement: `explain (costs off)
select distinct q1 from
  (select distinct * from int8_tbl i81
   union all
   select distinct * from int8_tbl i82) ss
where -q1 = q2;`,
				Results: []sql.Row{{`Unique`}, {`->  Merge Append`}, {`Sort Key: "*SELECT* 1".q1`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Unique`}, {`->  Sort`}, {`Sort Key: i81.q1, i81.q2`}, {`->  Seq Scan on int8_tbl i81`}, {`Filter: ((- q1) = q2)`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Unique`}, {`->  Sort`}, {`Sort Key: i82.q1, i82.q2`}, {`->  Seq Scan on int8_tbl i82`}, {`Filter: ((- q1) = q2)`}},
			},
			{
				Statement: `select distinct q1 from
  (select distinct * from int8_tbl i81
   union all
   select distinct * from int8_tbl i82) ss
where -q1 = q2;`,
				Results: []sql.Row{{4567890123456789}},
			},
			{
				Statement: `create function expensivefunc(int) returns int
language plpgsql immutable strict cost 10000
as $$begin return $1; end$$;`,
			},
			{
				Statement: `create temp table t3 as select generate_series(-1000,1000) as x;`,
			},
			{
				Statement: `create index t3i on t3 (expensivefunc(x));`,
			},
			{
				Statement: `analyze t3;`,
			},
			{
				Statement: `explain (costs off)
select * from
  (select * from t3 a union all select * from t3 b) ss
  join int4_tbl on f1 = expensivefunc(x);`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Seq Scan on int4_tbl`}, {`->  Append`}, {`->  Index Scan using t3i on t3 a`}, {`Index Cond: (expensivefunc(x) = int4_tbl.f1)`}, {`->  Index Scan using t3i on t3 b`}, {`Index Cond: (expensivefunc(x) = int4_tbl.f1)`}},
			},
			{
				Statement: `select * from
  (select * from t3 a union all select * from t3 b) ss
  join int4_tbl on f1 = expensivefunc(x);`,
				Results: []sql.Row{{0, 0}, {0, 0}},
			},
			{
				Statement: `drop table t3;`,
			},
			{
				Statement: `drop function expensivefunc(int);`,
			},
			{
				Statement: `explain (costs off)
select * from
  (select *, 0 as x from int8_tbl a
   union all
   select *, 1 as x from int8_tbl b) ss
where (x = 0) or (q1 >= q2 and q1 <= q2);`,
				Results: []sql.Row{{`Append`}, {`->  Seq Scan on int8_tbl a`}, {`->  Seq Scan on int8_tbl b`}, {`Filter: ((q1 >= q2) AND (q1 <= q2))`}},
			},
			{
				Statement: `select * from
  (select *, 0 as x from int8_tbl a
   union all
   select *, 1 as x from int8_tbl b) ss
where (x = 0) or (q1 >= q2 and q1 <= q2);`,
				Results: []sql.Row{{123, 456, 0}, {123, 4567890123456789, 0}, {4567890123456789, 123, 0}, {4567890123456789, 4567890123456789, 0}, {4567890123456789, -4567890123456789, 0}, {4567890123456789, 4567890123456789, 1}},
			},
		},
	})
}
