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

func TestPartitionAggregate(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_partition_aggregate)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_partition_aggregate,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET enable_partitionwise_aggregate TO true;`,
			},
			{
				Statement: `SET enable_partitionwise_join TO true;`,
			},
			{
				Statement: `SET max_parallel_workers_per_gather TO 0;`,
			},
			{
				Statement: `SET enable_incremental_sort TO off;`,
			},
			{
				Statement: `CREATE TABLE pagg_tab (a int, b int, c text, d int) PARTITION BY LIST(c);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_p1 PARTITION OF pagg_tab FOR VALUES IN ('0000', '0001', '0002', '0003', '0004');`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_p2 PARTITION OF pagg_tab FOR VALUES IN ('0005', '0006', '0007', '0008');`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_p3 PARTITION OF pagg_tab FOR VALUES IN ('0009', '0010', '0011');`,
			},
			{
				Statement: `INSERT INTO pagg_tab SELECT i % 20, i % 30, to_char(i % 12, 'FM0000'), i % 30 FROM generate_series(0, 2999) i;`,
			},
			{
				Statement: `ANALYZE pagg_tab;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT c, sum(a), avg(b), count(*), min(a), max(b) FROM pagg_tab GROUP BY c HAVING avg(d) < 15 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab.c, (sum(pagg_tab.a)), (avg(pagg_tab.b))`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab.c`}, {`Filter: (avg(pagg_tab.d) < '15'::numeric)`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_1.c`}, {`Filter: (avg(pagg_tab_1.d) < '15'::numeric)`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_2.c`}, {`Filter: (avg(pagg_tab_2.d) < '15'::numeric)`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `SELECT c, sum(a), avg(b), count(*), min(a), max(b) FROM pagg_tab GROUP BY c HAVING avg(d) < 15 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{"0000", 2000, 12.0000000000000000, 250, 0, 24}, {"0001", 2250, 13.0000000000000000, 250, 1, 25}, {"0002", 2500, 14.0000000000000000, 250, 2, 26}, {"0006", 2500, 12.0000000000000000, 250, 2, 24}, {"0007", 2750, 13.0000000000000000, 250, 3, 25}, {"0008", 2000, 14.0000000000000000, 250, 0, 26}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), avg(b), count(*), min(a), max(b) FROM pagg_tab GROUP BY a HAVING avg(d) < 15 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab.a, (sum(pagg_tab.b)), (avg(pagg_tab.b))`}, {`->  Finalize HashAggregate`}, {`Group Key: pagg_tab.a`}, {`Filter: (avg(pagg_tab.d) < '15'::numeric)`}, {`->  Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab.a`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_1.a`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_2.a`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `SELECT a, sum(b), avg(b), count(*), min(a), max(b) FROM pagg_tab GROUP BY a HAVING avg(d) < 15 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 1500, 10.0000000000000000, 150, 0, 20}, {1, 1650, 11.0000000000000000, 150, 1, 21}, {2, 1800, 12.0000000000000000, 150, 2, 22}, {3, 1950, 13.0000000000000000, 150, 3, 23}, {4, 2100, 14.0000000000000000, 150, 4, 24}, {10, 1500, 10.0000000000000000, 150, 10, 20}, {11, 1650, 11.0000000000000000, 150, 11, 21}, {12, 1800, 12.0000000000000000, 150, 12, 22}, {13, 1950, 13.0000000000000000, 150, 13, 23}, {14, 2100, 14.0000000000000000, 150, 14, 24}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, c, count(*) FROM pagg_tab GROUP BY a, c;`,
				Results: []sql.Row{{`Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab.a, pagg_tab.c`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_1.a, pagg_tab_1.c`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_2.a, pagg_tab_2.c`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, c, count(*) FROM pagg_tab GROUP BY c, a;`,
				Results: []sql.Row{{`Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab.c, pagg_tab.a`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_1.c, pagg_tab_1.a`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_2.c, pagg_tab_2.a`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT c, a, count(*) FROM pagg_tab GROUP BY a, c;`,
				Results: []sql.Row{{`Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab.a, pagg_tab.c`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_1.a, pagg_tab_1.c`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_2.a, pagg_tab_2.c`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT c, sum(a) FROM pagg_tab WHERE 1 = 2 GROUP BY c;`,
				Results: []sql.Row{{`HashAggregate`}, {`Group Key: c`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `SELECT c, sum(a) FROM pagg_tab WHERE 1 = 2 GROUP BY c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT c, sum(a) FROM pagg_tab WHERE c = 'x' GROUP BY c;`,
				Results: []sql.Row{{`GroupAggregate`}, {`Group Key: c`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `SELECT c, sum(a) FROM pagg_tab WHERE c = 'x' GROUP BY c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SET enable_hashagg TO false;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT c, sum(a), avg(b), count(*) FROM pagg_tab GROUP BY 1 HAVING avg(d) < 15 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab.c, (sum(pagg_tab.a)), (avg(pagg_tab.b))`}, {`->  Append`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab.c`}, {`Filter: (avg(pagg_tab.d) < '15'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab.c`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_1.c`}, {`Filter: (avg(pagg_tab_1.d) < '15'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_1.c`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_2.c`}, {`Filter: (avg(pagg_tab_2.d) < '15'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_2.c`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `SELECT c, sum(a), avg(b), count(*) FROM pagg_tab GROUP BY 1 HAVING avg(d) < 15 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{"0000", 2000, 12.0000000000000000, 250}, {"0001", 2250, 13.0000000000000000, 250}, {"0002", 2500, 14.0000000000000000, 250}, {"0006", 2500, 12.0000000000000000, 250}, {"0007", 2750, 13.0000000000000000, 250}, {"0008", 2000, 14.0000000000000000, 250}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), avg(b), count(*) FROM pagg_tab GROUP BY 1 HAVING avg(d) < 15 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab.a, (sum(pagg_tab.b)), (avg(pagg_tab.b))`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab.a`}, {`Filter: (avg(pagg_tab.d) < '15'::numeric)`}, {`->  Merge Append`}, {`Sort Key: pagg_tab.a`}, {`->  Partial GroupAggregate`}, {`Group Key: pagg_tab.a`}, {`->  Sort`}, {`Sort Key: pagg_tab.a`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  Partial GroupAggregate`}, {`Group Key: pagg_tab_1.a`}, {`->  Sort`}, {`Sort Key: pagg_tab_1.a`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  Partial GroupAggregate`}, {`Group Key: pagg_tab_2.a`}, {`->  Sort`}, {`Sort Key: pagg_tab_2.a`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `SELECT a, sum(b), avg(b), count(*) FROM pagg_tab GROUP BY 1 HAVING avg(d) < 15 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 1500, 10.0000000000000000, 150}, {1, 1650, 11.0000000000000000, 150}, {2, 1800, 12.0000000000000000, 150}, {3, 1950, 13.0000000000000000, 150}, {4, 2100, 14.0000000000000000, 150}, {10, 1500, 10.0000000000000000, 150}, {11, 1650, 11.0000000000000000, 150}, {12, 1800, 12.0000000000000000, 150}, {13, 1950, 13.0000000000000000, 150}, {14, 2100, 14.0000000000000000, 150}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT c FROM pagg_tab GROUP BY c ORDER BY 1;`,
				Results: []sql.Row{{`Merge Append`}, {`Sort Key: pagg_tab.c`}, {`->  Group`}, {`Group Key: pagg_tab.c`}, {`->  Sort`}, {`Sort Key: pagg_tab.c`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  Group`}, {`Group Key: pagg_tab_1.c`}, {`->  Sort`}, {`Sort Key: pagg_tab_1.c`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  Group`}, {`Group Key: pagg_tab_2.c`}, {`->  Sort`}, {`Sort Key: pagg_tab_2.c`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `SELECT c FROM pagg_tab GROUP BY c ORDER BY 1;`,
				Results:   []sql.Row{{"0000"}, {"0001"}, {"0002"}, {"0003"}, {"0004"}, {"0005"}, {"0006"}, {"0007"}, {"0008"}, {"0009"}, {"0010"}, {"0011"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a FROM pagg_tab WHERE a < 3 GROUP BY a ORDER BY 1;`,
				Results: []sql.Row{{`Group`}, {`Group Key: pagg_tab.a`}, {`->  Merge Append`}, {`Sort Key: pagg_tab.a`}, {`->  Group`}, {`Group Key: pagg_tab.a`}, {`->  Sort`}, {`Sort Key: pagg_tab.a`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`Filter: (a < 3)`}, {`->  Group`}, {`Group Key: pagg_tab_1.a`}, {`->  Sort`}, {`Sort Key: pagg_tab_1.a`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`Filter: (a < 3)`}, {`->  Group`}, {`Group Key: pagg_tab_2.a`}, {`->  Sort`}, {`Sort Key: pagg_tab_2.a`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}, {`Filter: (a < 3)`}},
			},
			{
				Statement: `SELECT a FROM pagg_tab WHERE a < 3 GROUP BY a ORDER BY 1;`,
				Results:   []sql.Row{{0}, {1}, {2}},
			},
			{
				Statement: `RESET enable_hashagg;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT c, sum(a) FROM pagg_tab GROUP BY rollup(c) ORDER BY 1, 2;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab.c, (sum(pagg_tab.a))`}, {`->  MixedAggregate`}, {`Hash Key: pagg_tab.c`}, {`Group Key: ()`}, {`->  Append`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab_1`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_2`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT c, sum(b order by a) FROM pagg_tab GROUP BY c ORDER BY 1, 2;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab.c, (sum(pagg_tab.b ORDER BY pagg_tab.a))`}, {`->  Append`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab.c`}, {`->  Sort`}, {`Sort Key: pagg_tab.c`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_1.c`}, {`->  Sort`}, {`Sort Key: pagg_tab_1.c`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_1`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_2.c`}, {`->  Sort`}, {`Sort Key: pagg_tab_2.c`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_2`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b order by a) FROM pagg_tab GROUP BY a ORDER BY 1, 2;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab.a, (sum(pagg_tab.b ORDER BY pagg_tab.a))`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab.a`}, {`->  Sort`}, {`Sort Key: pagg_tab.a`}, {`->  Append`}, {`->  Seq Scan on pagg_tab_p1 pagg_tab_1`}, {`->  Seq Scan on pagg_tab_p2 pagg_tab_2`}, {`->  Seq Scan on pagg_tab_p3 pagg_tab_3`}},
			},
			{
				Statement: `CREATE TABLE pagg_tab1(x int, y int) PARTITION BY RANGE(x);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab1_p1 PARTITION OF pagg_tab1 FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab1_p2 PARTITION OF pagg_tab1 FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab1_p3 PARTITION OF pagg_tab1 FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab2(x int, y int) PARTITION BY RANGE(y);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab2_p1 PARTITION OF pagg_tab2 FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab2_p2 PARTITION OF pagg_tab2 FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab2_p3 PARTITION OF pagg_tab2 FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `INSERT INTO pagg_tab1 SELECT i % 30, i % 20 FROM generate_series(0, 299, 2) i;`,
			},
			{
				Statement: `INSERT INTO pagg_tab2 SELECT i % 20, i % 30 FROM generate_series(0, 299, 3) i;`,
			},
			{
				Statement: `ANALYZE pagg_tab1;`,
			},
			{
				Statement: `ANALYZE pagg_tab2;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.x, sum(t1.y), count(*) FROM pagg_tab1 t1, pagg_tab2 t2 WHERE t1.x = t2.y GROUP BY t1.x ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.x, (sum(t1.y)), (count(*))`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: t1.x`}, {`->  Hash Join`}, {`Hash Cond: (t1.x = t2.y)`}, {`->  Seq Scan on pagg_tab1_p1 t1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p1 t2`}, {`->  HashAggregate`}, {`Group Key: t1_1.x`}, {`->  Hash Join`}, {`Hash Cond: (t1_1.x = t2_1.y)`}, {`->  Seq Scan on pagg_tab1_p2 t1_1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p2 t2_1`}, {`->  HashAggregate`}, {`Group Key: t1_2.x`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.y = t1_2.x)`}, {`->  Seq Scan on pagg_tab2_p3 t2_2`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab1_p3 t1_2`}},
			},
			{
				Statement: `SELECT t1.x, sum(t1.y), count(*) FROM pagg_tab1 t1, pagg_tab2 t2 WHERE t1.x = t2.y GROUP BY t1.x ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 500, 100}, {6, 1100, 100}, {12, 700, 100}, {18, 1300, 100}, {24, 900, 100}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.x, sum(t1.y), count(t1) FROM pagg_tab1 t1, pagg_tab2 t2 WHERE t1.x = t2.y GROUP BY t1.x ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.x, (sum(t1.y)), (count(((t1.*)::pagg_tab1)))`}, {`->  HashAggregate`}, {`Group Key: t1.x`}, {`->  Hash Join`}, {`Hash Cond: (t1.x = t2.y)`}, {`->  Append`}, {`->  Seq Scan on pagg_tab1_p1 t1_1`}, {`->  Seq Scan on pagg_tab1_p2 t1_2`}, {`->  Seq Scan on pagg_tab1_p3 t1_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on pagg_tab2_p1 t2_1`}, {`->  Seq Scan on pagg_tab2_p2 t2_2`}, {`->  Seq Scan on pagg_tab2_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.x, sum(t1.y), count(t1) FROM pagg_tab1 t1, pagg_tab2 t2 WHERE t1.x = t2.y GROUP BY t1.x ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 500, 100}, {6, 1100, 100}, {12, 700, 100}, {18, 1300, 100}, {24, 900, 100}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t2.y, sum(t1.y), count(*) FROM pagg_tab1 t1, pagg_tab2 t2 WHERE t1.x = t2.y GROUP BY t2.y ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t2.y, (sum(t1.y)), (count(*))`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: t2.y`}, {`->  Hash Join`}, {`Hash Cond: (t1.x = t2.y)`}, {`->  Seq Scan on pagg_tab1_p1 t1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p1 t2`}, {`->  HashAggregate`}, {`Group Key: t2_1.y`}, {`->  Hash Join`}, {`Hash Cond: (t1_1.x = t2_1.y)`}, {`->  Seq Scan on pagg_tab1_p2 t1_1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p2 t2_1`}, {`->  HashAggregate`}, {`Group Key: t2_2.y`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.y = t1_2.x)`}, {`->  Seq Scan on pagg_tab2_p3 t2_2`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab1_p3 t1_2`}},
			},
			{
				Statement: `SET enable_hashagg TO false;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.y, sum(t1.x), count(*) FROM pagg_tab1 t1, pagg_tab2 t2 WHERE t1.x = t2.y GROUP BY t1.y HAVING avg(t1.x) > 10 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.y, (sum(t1.x)), (count(*))`}, {`->  Finalize GroupAggregate`}, {`Group Key: t1.y`}, {`Filter: (avg(t1.x) > '10'::numeric)`}, {`->  Merge Append`}, {`Sort Key: t1.y`}, {`->  Partial GroupAggregate`}, {`Group Key: t1.y`}, {`->  Sort`}, {`Sort Key: t1.y`}, {`->  Hash Join`}, {`Hash Cond: (t1.x = t2.y)`}, {`->  Seq Scan on pagg_tab1_p1 t1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p1 t2`}, {`->  Partial GroupAggregate`}, {`Group Key: t1_1.y`}, {`->  Sort`}, {`Sort Key: t1_1.y`}, {`->  Hash Join`}, {`Hash Cond: (t1_1.x = t2_1.y)`}, {`->  Seq Scan on pagg_tab1_p2 t1_1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p2 t2_1`}, {`->  Partial GroupAggregate`}, {`Group Key: t1_2.y`}, {`->  Sort`}, {`Sort Key: t1_2.y`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.y = t1_2.x)`}, {`->  Seq Scan on pagg_tab2_p3 t2_2`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab1_p3 t1_2`}},
			},
			{
				Statement: `SELECT t1.y, sum(t1.x), count(*) FROM pagg_tab1 t1, pagg_tab2 t2 WHERE t1.x = t2.y GROUP BY t1.y HAVING avg(t1.x) > 10 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{2, 600, 50}, {4, 1200, 50}, {8, 900, 50}, {12, 600, 50}, {14, 1200, 50}, {18, 900, 50}},
			},
			{
				Statement: `RESET enable_hashagg;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT b.y, sum(a.y) FROM pagg_tab1 a LEFT JOIN pagg_tab2 b ON a.x = b.y GROUP BY b.y ORDER BY 1 NULLS LAST;`,
				Results: []sql.Row{{`Finalize GroupAggregate`}, {`Group Key: b.y`}, {`->  Sort`}, {`Sort Key: b.y`}, {`->  Append`}, {`->  Partial HashAggregate`}, {`Group Key: b.y`}, {`->  Hash Left Join`}, {`Hash Cond: (a.x = b.y)`}, {`->  Seq Scan on pagg_tab1_p1 a`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p1 b`}, {`->  Partial HashAggregate`}, {`Group Key: b_1.y`}, {`->  Hash Left Join`}, {`Hash Cond: (a_1.x = b_1.y)`}, {`->  Seq Scan on pagg_tab1_p2 a_1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p2 b_1`}, {`->  Partial HashAggregate`}, {`Group Key: b_2.y`}, {`->  Hash Right Join`}, {`Hash Cond: (b_2.y = a_2.x)`}, {`->  Seq Scan on pagg_tab2_p3 b_2`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab1_p3 a_2`}},
			},
			{
				Statement: `SELECT b.y, sum(a.y) FROM pagg_tab1 a LEFT JOIN pagg_tab2 b ON a.x = b.y GROUP BY b.y ORDER BY 1 NULLS LAST;`,
				Results:   []sql.Row{{0, 500}, {6, 1100}, {12, 700}, {18, 1300}, {24, 900}, {``, 900}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT b.y, sum(a.y) FROM pagg_tab1 a RIGHT JOIN pagg_tab2 b ON a.x = b.y GROUP BY b.y ORDER BY 1 NULLS LAST;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: b.y`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: b.y`}, {`->  Hash Right Join`}, {`Hash Cond: (a.x = b.y)`}, {`->  Seq Scan on pagg_tab1_p1 a`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p1 b`}, {`->  HashAggregate`}, {`Group Key: b_1.y`}, {`->  Hash Right Join`}, {`Hash Cond: (a_1.x = b_1.y)`}, {`->  Seq Scan on pagg_tab1_p2 a_1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p2 b_1`}, {`->  HashAggregate`}, {`Group Key: b_2.y`}, {`->  Hash Left Join`}, {`Hash Cond: (b_2.y = a_2.x)`}, {`->  Seq Scan on pagg_tab2_p3 b_2`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab1_p3 a_2`}},
			},
			{
				Statement: `SELECT b.y, sum(a.y) FROM pagg_tab1 a RIGHT JOIN pagg_tab2 b ON a.x = b.y GROUP BY b.y ORDER BY 1 NULLS LAST;`,
				Results:   []sql.Row{{0, 500}, {3, ``}, {6, 1100}, {9, ``}, {12, 700}, {15, ``}, {18, 1300}, {21, ``}, {24, 900}, {27, ``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a.x, sum(b.x) FROM pagg_tab1 a FULL OUTER JOIN pagg_tab2 b ON a.x = b.y GROUP BY a.x ORDER BY 1 NULLS LAST;`,
				Results: []sql.Row{{`Finalize GroupAggregate`}, {`Group Key: a.x`}, {`->  Sort`}, {`Sort Key: a.x`}, {`->  Append`}, {`->  Partial HashAggregate`}, {`Group Key: a.x`}, {`->  Hash Full Join`}, {`Hash Cond: (a.x = b.y)`}, {`->  Seq Scan on pagg_tab1_p1 a`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p1 b`}, {`->  Partial HashAggregate`}, {`Group Key: a_1.x`}, {`->  Hash Full Join`}, {`Hash Cond: (a_1.x = b_1.y)`}, {`->  Seq Scan on pagg_tab1_p2 a_1`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab2_p2 b_1`}, {`->  Partial HashAggregate`}, {`Group Key: a_2.x`}, {`->  Hash Full Join`}, {`Hash Cond: (b_2.y = a_2.x)`}, {`->  Seq Scan on pagg_tab2_p3 b_2`}, {`->  Hash`}, {`->  Seq Scan on pagg_tab1_p3 a_2`}},
			},
			{
				Statement: `SELECT a.x, sum(b.x) FROM pagg_tab1 a FULL OUTER JOIN pagg_tab2 b ON a.x = b.y GROUP BY a.x ORDER BY 1 NULLS LAST;`,
				Results:   []sql.Row{{0, 500}, {2, ``}, {4, ``}, {6, 1100}, {8, ``}, {10, ``}, {12, 700}, {14, ``}, {16, ``}, {18, 1300}, {20, ``}, {22, ``}, {24, 900}, {26, ``}, {28, ``}, {``, 500}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a.x, b.y, count(*) FROM (SELECT * FROM pagg_tab1 WHERE x < 20) a LEFT JOIN (SELECT * FROM pagg_tab2 WHERE y > 10) b ON a.x = b.y WHERE a.x > 5 or b.y < 20  GROUP BY a.x, b.y ORDER BY 1, 2;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab1.x, pagg_tab2.y`}, {`->  HashAggregate`}, {`Group Key: pagg_tab1.x, pagg_tab2.y`}, {`->  Hash Left Join`}, {`Hash Cond: (pagg_tab1.x = pagg_tab2.y)`}, {`Filter: ((pagg_tab1.x > 5) OR (pagg_tab2.y < 20))`}, {`->  Append`}, {`->  Seq Scan on pagg_tab1_p1 pagg_tab1_1`}, {`Filter: (x < 20)`}, {`->  Seq Scan on pagg_tab1_p2 pagg_tab1_2`}, {`Filter: (x < 20)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on pagg_tab2_p2 pagg_tab2_1`}, {`Filter: (y > 10)`}, {`->  Seq Scan on pagg_tab2_p3 pagg_tab2_2`}, {`Filter: (y > 10)`}},
			},
			{
				Statement: `SELECT a.x, b.y, count(*) FROM (SELECT * FROM pagg_tab1 WHERE x < 20) a LEFT JOIN (SELECT * FROM pagg_tab2 WHERE y > 10) b ON a.x = b.y WHERE a.x > 5 or b.y < 20  GROUP BY a.x, b.y ORDER BY 1, 2;`,
				Results:   []sql.Row{{6, ``, 10}, {8, ``, 10}, {10, ``, 10}, {12, 12, 100}, {14, ``, 10}, {16, ``, 10}, {18, 18, 100}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a.x, b.y, count(*) FROM (SELECT * FROM pagg_tab1 WHERE x < 20) a FULL JOIN (SELECT * FROM pagg_tab2 WHERE y > 10) b ON a.x = b.y WHERE a.x > 5 or b.y < 20  GROUP BY a.x, b.y ORDER BY 1, 2;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab1.x, pagg_tab2.y`}, {`->  HashAggregate`}, {`Group Key: pagg_tab1.x, pagg_tab2.y`}, {`->  Hash Full Join`}, {`Hash Cond: (pagg_tab1.x = pagg_tab2.y)`}, {`Filter: ((pagg_tab1.x > 5) OR (pagg_tab2.y < 20))`}, {`->  Append`}, {`->  Seq Scan on pagg_tab1_p1 pagg_tab1_1`}, {`Filter: (x < 20)`}, {`->  Seq Scan on pagg_tab1_p2 pagg_tab1_2`}, {`Filter: (x < 20)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on pagg_tab2_p2 pagg_tab2_1`}, {`Filter: (y > 10)`}, {`->  Seq Scan on pagg_tab2_p3 pagg_tab2_2`}, {`Filter: (y > 10)`}},
			},
			{
				Statement: `SELECT a.x, b.y, count(*) FROM (SELECT * FROM pagg_tab1 WHERE x < 20) a FULL JOIN (SELECT * FROM pagg_tab2 WHERE y > 10) b ON a.x = b.y WHERE a.x > 5 or b.y < 20 GROUP BY a.x, b.y ORDER BY 1, 2;`,
				Results:   []sql.Row{{6, ``, 10}, {8, ``, 10}, {10, ``, 10}, {12, 12, 100}, {14, ``, 10}, {16, ``, 10}, {18, 18, 100}, {``, 15, 10}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a.x, a.y, count(*) FROM (SELECT * FROM pagg_tab1 WHERE x = 1 AND x = 2) a LEFT JOIN pagg_tab2 b ON a.x = b.y GROUP BY a.x, a.y ORDER BY 1, 2;`,
				Results: []sql.Row{{`GroupAggregate`}, {`Group Key: pagg_tab1.x, pagg_tab1.y`}, {`->  Sort`}, {`Sort Key: pagg_tab1.y`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `SELECT a.x, a.y, count(*) FROM (SELECT * FROM pagg_tab1 WHERE x = 1 AND x = 2) a LEFT JOIN pagg_tab2 b ON a.x = b.y GROUP BY a.x, a.y ORDER BY 1, 2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `CREATE TABLE pagg_tab_m (a int, b int, c int) PARTITION BY RANGE(a, ((a+b)/2));`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_m_p1 PARTITION OF pagg_tab_m FOR VALUES FROM (0, 0) TO (12, 12);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_m_p2 PARTITION OF pagg_tab_m FOR VALUES FROM (12, 12) TO (22, 22);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_m_p3 PARTITION OF pagg_tab_m FOR VALUES FROM (22, 22) TO (30, 30);`,
			},
			{
				Statement: `INSERT INTO pagg_tab_m SELECT i % 30, i % 40, i % 50 FROM generate_series(0, 2999) i;`,
			},
			{
				Statement: `ANALYZE pagg_tab_m;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), avg(c), count(*) FROM pagg_tab_m GROUP BY a HAVING avg(c) < 22 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_m.a, (sum(pagg_tab_m.b)), (avg(pagg_tab_m.c))`}, {`->  Finalize HashAggregate`}, {`Group Key: pagg_tab_m.a`}, {`Filter: (avg(pagg_tab_m.c) < '22'::numeric)`}, {`->  Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_m.a`}, {`->  Seq Scan on pagg_tab_m_p1 pagg_tab_m`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_m_1.a`}, {`->  Seq Scan on pagg_tab_m_p2 pagg_tab_m_1`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_m_2.a`}, {`->  Seq Scan on pagg_tab_m_p3 pagg_tab_m_2`}},
			},
			{
				Statement: `SELECT a, sum(b), avg(c), count(*) FROM pagg_tab_m GROUP BY a HAVING avg(c) < 22 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 1500, 20.0000000000000000, 100}, {1, 1600, 21.0000000000000000, 100}, {10, 1500, 20.0000000000000000, 100}, {11, 1600, 21.0000000000000000, 100}, {20, 1500, 20.0000000000000000, 100}, {21, 1600, 21.0000000000000000, 100}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), avg(c), count(*) FROM pagg_tab_m GROUP BY a, (a+b)/2 HAVING sum(b) < 50 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_m.a, (sum(pagg_tab_m.b)), (avg(pagg_tab_m.c))`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_m.a, ((pagg_tab_m.a + pagg_tab_m.b) / 2)`}, {`Filter: (sum(pagg_tab_m.b) < 50)`}, {`->  Seq Scan on pagg_tab_m_p1 pagg_tab_m`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_m_1.a, ((pagg_tab_m_1.a + pagg_tab_m_1.b) / 2)`}, {`Filter: (sum(pagg_tab_m_1.b) < 50)`}, {`->  Seq Scan on pagg_tab_m_p2 pagg_tab_m_1`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_m_2.a, ((pagg_tab_m_2.a + pagg_tab_m_2.b) / 2)`}, {`Filter: (sum(pagg_tab_m_2.b) < 50)`}, {`->  Seq Scan on pagg_tab_m_p3 pagg_tab_m_2`}},
			},
			{
				Statement: `SELECT a, sum(b), avg(c), count(*) FROM pagg_tab_m GROUP BY a, (a+b)/2 HAVING sum(b) < 50 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 0, 20.0000000000000000, 25}, {1, 25, 21.0000000000000000, 25}, {10, 0, 20.0000000000000000, 25}, {11, 25, 21.0000000000000000, 25}, {20, 0, 20.0000000000000000, 25}, {21, 25, 21.0000000000000000, 25}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, c, sum(b), avg(c), count(*) FROM pagg_tab_m GROUP BY (a+b)/2, 2, 1 HAVING sum(b) = 50 AND avg(c) > 25 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_m.a, pagg_tab_m.c, (sum(pagg_tab_m.b))`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: ((pagg_tab_m.a + pagg_tab_m.b) / 2), pagg_tab_m.c, pagg_tab_m.a`}, {`Filter: ((sum(pagg_tab_m.b) = 50) AND (avg(pagg_tab_m.c) > '25'::numeric))`}, {`->  Seq Scan on pagg_tab_m_p1 pagg_tab_m`}, {`->  HashAggregate`}, {`Group Key: ((pagg_tab_m_1.a + pagg_tab_m_1.b) / 2), pagg_tab_m_1.c, pagg_tab_m_1.a`}, {`Filter: ((sum(pagg_tab_m_1.b) = 50) AND (avg(pagg_tab_m_1.c) > '25'::numeric))`}, {`->  Seq Scan on pagg_tab_m_p2 pagg_tab_m_1`}, {`->  HashAggregate`}, {`Group Key: ((pagg_tab_m_2.a + pagg_tab_m_2.b) / 2), pagg_tab_m_2.c, pagg_tab_m_2.a`}, {`Filter: ((sum(pagg_tab_m_2.b) = 50) AND (avg(pagg_tab_m_2.c) > '25'::numeric))`}, {`->  Seq Scan on pagg_tab_m_p3 pagg_tab_m_2`}},
			},
			{
				Statement: `SELECT a, c, sum(b), avg(c), count(*) FROM pagg_tab_m GROUP BY (a+b)/2, 2, 1 HAVING sum(b) = 50 AND avg(c) > 25 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 30, 50, 30.0000000000000000, 5}, {0, 40, 50, 40.0000000000000000, 5}, {10, 30, 50, 30.0000000000000000, 5}, {10, 40, 50, 40.0000000000000000, 5}, {20, 30, 50, 30.0000000000000000, 5}, {20, 40, 50, 40.0000000000000000, 5}},
			},
			{
				Statement: `CREATE TABLE pagg_tab_ml (a int, b int, c text) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_ml_p1 PARTITION OF pagg_tab_ml FOR VALUES FROM (0) TO (12);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_ml_p2 PARTITION OF pagg_tab_ml FOR VALUES FROM (12) TO (20) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_ml_p2_s1 PARTITION OF pagg_tab_ml_p2 FOR VALUES IN ('0000', '0001', '0002');`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_ml_p2_s2 PARTITION OF pagg_tab_ml_p2 FOR VALUES IN ('0003');`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_ml_p3(b int, c text, a int) PARTITION BY RANGE (b);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_ml_p3_s1(c text, a int, b int);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_ml_p3_s2 PARTITION OF pagg_tab_ml_p3 FOR VALUES FROM (7) TO (10);`,
			},
			{
				Statement: `ALTER TABLE pagg_tab_ml_p3 ATTACH PARTITION pagg_tab_ml_p3_s1 FOR VALUES FROM (0) TO (7);`,
			},
			{
				Statement: `ALTER TABLE pagg_tab_ml ATTACH PARTITION pagg_tab_ml_p3 FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `INSERT INTO pagg_tab_ml SELECT i % 30, i % 10, to_char(i % 4, 'FM0000') FROM generate_series(0, 29999) i;`,
			},
			{
				Statement: `ANALYZE pagg_tab_ml;`,
			},
			{
				Statement: `SET max_parallel_workers_per_gather TO 2;`,
			},
			{
				Statement: `SET parallel_setup_cost = 0;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), array_agg(distinct c), count(*) FROM pagg_tab_ml GROUP BY a HAVING avg(b) < 3 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_ml.a, (sum(pagg_tab_ml.b)), (array_agg(DISTINCT pagg_tab_ml.c))`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Parallel Append`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_ml.a`}, {`Filter: (avg(pagg_tab_ml.b) < '3'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml.a`}, {`->  Seq Scan on pagg_tab_ml_p1 pagg_tab_ml`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_ml_5.a`}, {`Filter: (avg(pagg_tab_ml_5.b) < '3'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml_5.a`}, {`->  Append`}, {`->  Seq Scan on pagg_tab_ml_p3_s1 pagg_tab_ml_5`}, {`->  Seq Scan on pagg_tab_ml_p3_s2 pagg_tab_ml_6`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_ml_2.a`}, {`Filter: (avg(pagg_tab_ml_2.b) < '3'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml_2.a`}, {`->  Append`}, {`->  Seq Scan on pagg_tab_ml_p2_s1 pagg_tab_ml_2`}, {`->  Seq Scan on pagg_tab_ml_p2_s2 pagg_tab_ml_3`}},
			},
			{
				Statement: `SELECT a, sum(b), array_agg(distinct c), count(*) FROM pagg_tab_ml GROUP BY a HAVING avg(b) < 3 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 0, `{0000,0002}`, 1000}, {1, 1000, `{0001,0003}`, 1000}, {2, 2000, `{0000,0002}`, 1000}, {10, 0, `{0000,0002}`, 1000}, {11, 1000, `{0001,0003}`, 1000}, {12, 2000, `{0000,0002}`, 1000}, {20, 0, `{0000,0002}`, 1000}, {21, 1000, `{0001,0003}`, 1000}, {22, 2000, `{0000,0002}`, 1000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), array_agg(distinct c), count(*) FROM pagg_tab_ml GROUP BY a HAVING avg(b) < 3;`,
				Results: []sql.Row{{`Gather`}, {`Workers Planned: 2`}, {`->  Parallel Append`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_ml.a`}, {`Filter: (avg(pagg_tab_ml.b) < '3'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml.a`}, {`->  Seq Scan on pagg_tab_ml_p1 pagg_tab_ml`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_ml_5.a`}, {`Filter: (avg(pagg_tab_ml_5.b) < '3'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml_5.a`}, {`->  Append`}, {`->  Seq Scan on pagg_tab_ml_p3_s1 pagg_tab_ml_5`}, {`->  Seq Scan on pagg_tab_ml_p3_s2 pagg_tab_ml_6`}, {`->  GroupAggregate`}, {`Group Key: pagg_tab_ml_2.a`}, {`Filter: (avg(pagg_tab_ml_2.b) < '3'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml_2.a`}, {`->  Append`}, {`->  Seq Scan on pagg_tab_ml_p2_s1 pagg_tab_ml_2`}, {`->  Seq Scan on pagg_tab_ml_p2_s2 pagg_tab_ml_3`}},
			},
			{
				Statement: `RESET parallel_setup_cost;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), count(*) FROM pagg_tab_ml GROUP BY a HAVING avg(b) < 3 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_ml.a, (sum(pagg_tab_ml.b)), (count(*))`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml.a`}, {`Filter: (avg(pagg_tab_ml.b) < '3'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p1 pagg_tab_ml`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_ml_2.a`}, {`Filter: (avg(pagg_tab_ml_2.b) < '3'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml_2.a`}, {`->  Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_2.a`}, {`->  Seq Scan on pagg_tab_ml_p2_s1 pagg_tab_ml_2`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_3.a`}, {`->  Seq Scan on pagg_tab_ml_p2_s2 pagg_tab_ml_3`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_ml_5.a`}, {`Filter: (avg(pagg_tab_ml_5.b) < '3'::numeric)`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml_5.a`}, {`->  Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_5.a`}, {`->  Seq Scan on pagg_tab_ml_p3_s1 pagg_tab_ml_5`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_6.a`}, {`->  Seq Scan on pagg_tab_ml_p3_s2 pagg_tab_ml_6`}},
			},
			{
				Statement: `SELECT a, sum(b), count(*) FROM pagg_tab_ml GROUP BY a HAVING avg(b) < 3 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 0, 1000}, {1, 1000, 1000}, {2, 2000, 1000}, {10, 0, 1000}, {11, 1000, 1000}, {12, 2000, 1000}, {20, 0, 1000}, {21, 1000, 1000}, {22, 2000, 1000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT b, sum(a), count(*) FROM pagg_tab_ml GROUP BY b ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_ml.b, (sum(pagg_tab_ml.a)), (count(*))`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_ml.b`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml.b`}, {`->  Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml.b`}, {`->  Seq Scan on pagg_tab_ml_p1 pagg_tab_ml`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_1.b`}, {`->  Seq Scan on pagg_tab_ml_p2_s1 pagg_tab_ml_1`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_2.b`}, {`->  Seq Scan on pagg_tab_ml_p2_s2 pagg_tab_ml_2`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_3.b`}, {`->  Seq Scan on pagg_tab_ml_p3_s1 pagg_tab_ml_3`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_4.b`}, {`->  Seq Scan on pagg_tab_ml_p3_s2 pagg_tab_ml_4`}},
			},
			{
				Statement: `SELECT b, sum(a), count(*) FROM pagg_tab_ml GROUP BY b HAVING avg(a) < 15 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 30000, 3000}, {1, 33000, 3000}, {2, 36000, 3000}, {3, 39000, 3000}, {4, 42000, 3000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), count(*) FROM pagg_tab_ml GROUP BY a, b, c HAVING avg(b) > 7 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_ml.a, (sum(pagg_tab_ml.b)), (count(*))`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml.a, pagg_tab_ml.b, pagg_tab_ml.c`}, {`Filter: (avg(pagg_tab_ml.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p1 pagg_tab_ml`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml_1.a, pagg_tab_ml_1.b, pagg_tab_ml_1.c`}, {`Filter: (avg(pagg_tab_ml_1.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p2_s1 pagg_tab_ml_1`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml_2.a, pagg_tab_ml_2.b, pagg_tab_ml_2.c`}, {`Filter: (avg(pagg_tab_ml_2.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p2_s2 pagg_tab_ml_2`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml_3.a, pagg_tab_ml_3.b, pagg_tab_ml_3.c`}, {`Filter: (avg(pagg_tab_ml_3.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p3_s1 pagg_tab_ml_3`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml_4.a, pagg_tab_ml_4.b, pagg_tab_ml_4.c`}, {`Filter: (avg(pagg_tab_ml_4.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p3_s2 pagg_tab_ml_4`}},
			},
			{
				Statement: `SELECT a, sum(b), count(*) FROM pagg_tab_ml GROUP BY a, b, c HAVING avg(b) > 7 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{8, 4000, 500}, {8, 4000, 500}, {9, 4500, 500}, {9, 4500, 500}, {18, 4000, 500}, {18, 4000, 500}, {19, 4500, 500}, {19, 4500, 500}, {28, 4000, 500}, {28, 4000, 500}, {29, 4500, 500}, {29, 4500, 500}},
			},
			{
				Statement: `SET min_parallel_table_scan_size TO '8kB';`,
			},
			{
				Statement: `SET parallel_setup_cost TO 0;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), count(*) FROM pagg_tab_ml GROUP BY a HAVING avg(b) < 3 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_ml.a, (sum(pagg_tab_ml.b)), (count(*))`}, {`->  Append`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_ml.a`}, {`Filter: (avg(pagg_tab_ml.b) < '3'::numeric)`}, {`->  Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml.a`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml.a`}, {`->  Parallel Seq Scan on pagg_tab_ml_p1 pagg_tab_ml`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_ml_2.a`}, {`Filter: (avg(pagg_tab_ml_2.b) < '3'::numeric)`}, {`->  Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml_2.a`}, {`->  Parallel Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_2.a`}, {`->  Parallel Seq Scan on pagg_tab_ml_p2_s1 pagg_tab_ml_2`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_3.a`}, {`->  Parallel Seq Scan on pagg_tab_ml_p2_s2 pagg_tab_ml_3`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_ml_5.a`}, {`Filter: (avg(pagg_tab_ml_5.b) < '3'::numeric)`}, {`->  Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml_5.a`}, {`->  Parallel Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_5.a`}, {`->  Parallel Seq Scan on pagg_tab_ml_p3_s1 pagg_tab_ml_5`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_6.a`}, {`->  Parallel Seq Scan on pagg_tab_ml_p3_s2 pagg_tab_ml_6`}},
			},
			{
				Statement: `SELECT a, sum(b), count(*) FROM pagg_tab_ml GROUP BY a HAVING avg(b) < 3 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 0, 1000}, {1, 1000, 1000}, {2, 2000, 1000}, {10, 0, 1000}, {11, 1000, 1000}, {12, 2000, 1000}, {20, 0, 1000}, {21, 1000, 1000}, {22, 2000, 1000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT b, sum(a), count(*) FROM pagg_tab_ml GROUP BY b ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_ml.b, (sum(pagg_tab_ml.a)), (count(*))`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_ml.b`}, {`->  Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml.b`}, {`->  Parallel Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml.b`}, {`->  Parallel Seq Scan on pagg_tab_ml_p1 pagg_tab_ml`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_3.b`}, {`->  Parallel Seq Scan on pagg_tab_ml_p3_s1 pagg_tab_ml_3`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_1.b`}, {`->  Parallel Seq Scan on pagg_tab_ml_p2_s1 pagg_tab_ml_1`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_4.b`}, {`->  Parallel Seq Scan on pagg_tab_ml_p3_s2 pagg_tab_ml_4`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_ml_2.b`}, {`->  Parallel Seq Scan on pagg_tab_ml_p2_s2 pagg_tab_ml_2`}},
			},
			{
				Statement: `SELECT b, sum(a), count(*) FROM pagg_tab_ml GROUP BY b HAVING avg(a) < 15 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 30000, 3000}, {1, 33000, 3000}, {2, 36000, 3000}, {3, 39000, 3000}, {4, 42000, 3000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, sum(b), count(*) FROM pagg_tab_ml GROUP BY a, b, c HAVING avg(b) > 7 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_ml.a, (sum(pagg_tab_ml.b)), (count(*))`}, {`->  Parallel Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml.a, pagg_tab_ml.b, pagg_tab_ml.c`}, {`Filter: (avg(pagg_tab_ml.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p1 pagg_tab_ml`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml_3.a, pagg_tab_ml_3.b, pagg_tab_ml_3.c`}, {`Filter: (avg(pagg_tab_ml_3.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p3_s1 pagg_tab_ml_3`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml_1.a, pagg_tab_ml_1.b, pagg_tab_ml_1.c`}, {`Filter: (avg(pagg_tab_ml_1.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p2_s1 pagg_tab_ml_1`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml_4.a, pagg_tab_ml_4.b, pagg_tab_ml_4.c`}, {`Filter: (avg(pagg_tab_ml_4.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p3_s2 pagg_tab_ml_4`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_ml_2.a, pagg_tab_ml_2.b, pagg_tab_ml_2.c`}, {`Filter: (avg(pagg_tab_ml_2.b) > '7'::numeric)`}, {`->  Seq Scan on pagg_tab_ml_p2_s2 pagg_tab_ml_2`}},
			},
			{
				Statement: `SELECT a, sum(b), count(*) FROM pagg_tab_ml GROUP BY a, b, c HAVING avg(b) > 7 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{8, 4000, 500}, {8, 4000, 500}, {9, 4500, 500}, {9, 4500, 500}, {18, 4000, 500}, {18, 4000, 500}, {19, 4500, 500}, {19, 4500, 500}, {28, 4000, 500}, {28, 4000, 500}, {29, 4500, 500}, {29, 4500, 500}},
			},
			{
				Statement: `SET parallel_setup_cost TO 10;`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_para(x int, y int) PARTITION BY RANGE(x);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_para_p1 PARTITION OF pagg_tab_para FOR VALUES FROM (0) TO (12);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_para_p2 PARTITION OF pagg_tab_para FOR VALUES FROM (12) TO (22);`,
			},
			{
				Statement: `CREATE TABLE pagg_tab_para_p3 PARTITION OF pagg_tab_para FOR VALUES FROM (22) TO (30);`,
			},
			{
				Statement: `INSERT INTO pagg_tab_para SELECT i % 30, i % 20 FROM generate_series(0, 29999) i;`,
			},
			{
				Statement: `ANALYZE pagg_tab_para;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT x, sum(y), avg(y), count(*) FROM pagg_tab_para GROUP BY x HAVING avg(y) < 7 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_para.x, (sum(pagg_tab_para.y)), (avg(pagg_tab_para.y))`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_para.x`}, {`Filter: (avg(pagg_tab_para.y) < '7'::numeric)`}, {`->  Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_para.x`}, {`->  Parallel Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_para.x`}, {`->  Parallel Seq Scan on pagg_tab_para_p1 pagg_tab_para`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_para_1.x`}, {`->  Parallel Seq Scan on pagg_tab_para_p2 pagg_tab_para_1`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_para_2.x`}, {`->  Parallel Seq Scan on pagg_tab_para_p3 pagg_tab_para_2`}},
			},
			{
				Statement: `SELECT x, sum(y), avg(y), count(*) FROM pagg_tab_para GROUP BY x HAVING avg(y) < 7 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 5000, 5.0000000000000000, 1000}, {1, 6000, 6.0000000000000000, 1000}, {10, 5000, 5.0000000000000000, 1000}, {11, 6000, 6.0000000000000000, 1000}, {20, 5000, 5.0000000000000000, 1000}, {21, 6000, 6.0000000000000000, 1000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT y, sum(x), avg(x), count(*) FROM pagg_tab_para GROUP BY y HAVING avg(x) < 12 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_para.y, (sum(pagg_tab_para.x)), (avg(pagg_tab_para.x))`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_para.y`}, {`Filter: (avg(pagg_tab_para.x) < '12'::numeric)`}, {`->  Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_para.y`}, {`->  Parallel Append`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_para.y`}, {`->  Parallel Seq Scan on pagg_tab_para_p1 pagg_tab_para`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_para_1.y`}, {`->  Parallel Seq Scan on pagg_tab_para_p2 pagg_tab_para_1`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_para_2.y`}, {`->  Parallel Seq Scan on pagg_tab_para_p3 pagg_tab_para_2`}},
			},
			{
				Statement: `SELECT y, sum(x), avg(x), count(*) FROM pagg_tab_para GROUP BY y HAVING avg(x) < 12 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 15000, 10.0000000000000000, 1500}, {1, 16500, 11.0000000000000000, 1500}, {10, 15000, 10.0000000000000000, 1500}, {11, 16500, 11.0000000000000000, 1500}},
			},
			{
				Statement: `ALTER TABLE pagg_tab_para_p1 SET (parallel_workers = 0);`,
			},
			{
				Statement: `ALTER TABLE pagg_tab_para_p3 SET (parallel_workers = 0);`,
			},
			{
				Statement: `ANALYZE pagg_tab_para;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT x, sum(y), avg(y), sum(x+y), count(*) FROM pagg_tab_para GROUP BY x HAVING avg(y) < 7 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_para.x, (sum(pagg_tab_para.y)), (avg(pagg_tab_para.y))`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_para.x`}, {`Filter: (avg(pagg_tab_para.y) < '7'::numeric)`}, {`->  Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_para.x`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_para.x`}, {`->  Parallel Append`}, {`->  Seq Scan on pagg_tab_para_p1 pagg_tab_para_1`}, {`->  Seq Scan on pagg_tab_para_p3 pagg_tab_para_3`}, {`->  Parallel Seq Scan on pagg_tab_para_p2 pagg_tab_para_2`}},
			},
			{
				Statement: `SELECT x, sum(y), avg(y), sum(x+y), count(*) FROM pagg_tab_para GROUP BY x HAVING avg(y) < 7 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 5000, 5.0000000000000000, 5000, 1000}, {1, 6000, 6.0000000000000000, 7000, 1000}, {10, 5000, 5.0000000000000000, 15000, 1000}, {11, 6000, 6.0000000000000000, 17000, 1000}, {20, 5000, 5.0000000000000000, 25000, 1000}, {21, 6000, 6.0000000000000000, 27000, 1000}},
			},
			{
				Statement: `ALTER TABLE pagg_tab_para_p2 SET (parallel_workers = 0);`,
			},
			{
				Statement: `ANALYZE pagg_tab_para;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT x, sum(y), avg(y), sum(x+y), count(*) FROM pagg_tab_para GROUP BY x HAVING avg(y) < 7 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_para.x, (sum(pagg_tab_para.y)), (avg(pagg_tab_para.y))`}, {`->  Finalize GroupAggregate`}, {`Group Key: pagg_tab_para.x`}, {`Filter: (avg(pagg_tab_para.y) < '7'::numeric)`}, {`->  Gather Merge`}, {`Workers Planned: 2`}, {`->  Sort`}, {`Sort Key: pagg_tab_para.x`}, {`->  Partial HashAggregate`}, {`Group Key: pagg_tab_para.x`}, {`->  Parallel Append`}, {`->  Seq Scan on pagg_tab_para_p1 pagg_tab_para_1`}, {`->  Seq Scan on pagg_tab_para_p2 pagg_tab_para_2`}, {`->  Seq Scan on pagg_tab_para_p3 pagg_tab_para_3`}},
			},
			{
				Statement: `SELECT x, sum(y), avg(y), sum(x+y), count(*) FROM pagg_tab_para GROUP BY x HAVING avg(y) < 7 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 5000, 5.0000000000000000, 5000, 1000}, {1, 6000, 6.0000000000000000, 7000, 1000}, {10, 5000, 5.0000000000000000, 15000, 1000}, {11, 6000, 6.0000000000000000, 17000, 1000}, {20, 5000, 5.0000000000000000, 25000, 1000}, {21, 6000, 6.0000000000000000, 27000, 1000}},
			},
			{
				Statement: `RESET min_parallel_table_scan_size;`,
			},
			{
				Statement: `RESET parallel_setup_cost;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT x, sum(y), avg(y), count(*) FROM pagg_tab_para GROUP BY x HAVING avg(y) < 7 ORDER BY 1, 2, 3;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: pagg_tab_para.x, (sum(pagg_tab_para.y)), (avg(pagg_tab_para.y))`}, {`->  Append`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_para.x`}, {`Filter: (avg(pagg_tab_para.y) < '7'::numeric)`}, {`->  Seq Scan on pagg_tab_para_p1 pagg_tab_para`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_para_1.x`}, {`Filter: (avg(pagg_tab_para_1.y) < '7'::numeric)`}, {`->  Seq Scan on pagg_tab_para_p2 pagg_tab_para_1`}, {`->  HashAggregate`}, {`Group Key: pagg_tab_para_2.x`}, {`Filter: (avg(pagg_tab_para_2.y) < '7'::numeric)`}, {`->  Seq Scan on pagg_tab_para_p3 pagg_tab_para_2`}},
			},
			{
				Statement: `SELECT x, sum(y), avg(y), count(*) FROM pagg_tab_para GROUP BY x HAVING avg(y) < 7 ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{0, 5000, 5.0000000000000000, 1000}, {1, 6000, 6.0000000000000000, 1000}, {10, 5000, 5.0000000000000000, 1000}, {11, 6000, 6.0000000000000000, 1000}, {20, 5000, 5.0000000000000000, 1000}, {21, 6000, 6.0000000000000000, 1000}},
			},
		},
	})
}
