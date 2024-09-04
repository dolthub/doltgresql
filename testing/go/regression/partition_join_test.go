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

func TestPartitionJoin(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_partition_join)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_partition_join,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET enable_partitionwise_join to true;`,
			},
			{
				Statement: `CREATE TABLE prt1 (a int, b int, c varchar) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE prt1_p1 PARTITION OF prt1 FOR VALUES FROM (0) TO (250);`,
			},
			{
				Statement: `CREATE TABLE prt1_p3 PARTITION OF prt1 FOR VALUES FROM (500) TO (600);`,
			},
			{
				Statement: `CREATE TABLE prt1_p2 PARTITION OF prt1 FOR VALUES FROM (250) TO (500);`,
			},
			{
				Statement: `INSERT INTO prt1 SELECT i, i % 25, to_char(i, 'FM0000') FROM generate_series(0, 599) i WHERE i % 2 = 0;`,
			},
			{
				Statement: `CREATE INDEX iprt1_p1_a on prt1_p1(a);`,
			},
			{
				Statement: `CREATE INDEX iprt1_p2_a on prt1_p2(a);`,
			},
			{
				Statement: `CREATE INDEX iprt1_p3_a on prt1_p3(a);`,
			},
			{
				Statement: `ANALYZE prt1;`,
			},
			{
				Statement: `CREATE TABLE prt2 (a int, b int, c varchar) PARTITION BY RANGE(b);`,
			},
			{
				Statement: `CREATE TABLE prt2_p1 PARTITION OF prt2 FOR VALUES FROM (0) TO (250);`,
			},
			{
				Statement: `CREATE TABLE prt2_p2 PARTITION OF prt2 FOR VALUES FROM (250) TO (500);`,
			},
			{
				Statement: `CREATE TABLE prt2_p3 PARTITION OF prt2 FOR VALUES FROM (500) TO (600);`,
			},
			{
				Statement: `INSERT INTO prt2 SELECT i % 25, i, to_char(i, 'FM0000') FROM generate_series(0, 599) i WHERE i % 3 = 0;`,
			},
			{
				Statement: `CREATE INDEX iprt2_p1_b on prt2_p1(b);`,
			},
			{
				Statement: `CREATE INDEX iprt2_p2_b on prt2_p2(b);`,
			},
			{
				Statement: `CREATE INDEX iprt2_p3_b on prt2_p3(b);`,
			},
			{
				Statement: `ANALYZE prt2;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1, prt2 t2 WHERE t1.a = t2.b AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1, prt2 t2 WHERE t1.a = t2.b AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000"}, {150, "0150", 150, "0150"}, {300, "0300", 300, "0300"}, {450, "0450", 450, "0450"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1, t2 FROM prt1 t1 LEFT JOIN prt2 t2 ON t1.a = t2.b WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b`}, {`->  Hash Right Join`}, {`Hash Cond: (t2.b = t1.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1, t2 FROM prt1 t1 LEFT JOIN prt2 t2 ON t1.a = t2.b WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{`(0,0,0000)`, `(0,0,0000)`}, {`(50,0,0050)`, ``}, {`(100,0,0100)`, ``}, {`(150,0,0150)`, `(0,150,0150)`}, {`(200,0,0200)`, ``}, {`(250,0,0250)`, ``}, {`(300,0,0300)`, `(0,300,0300)`}, {`(350,0,0350)`, ``}, {`(400,0,0400)`, ``}, {`(450,0,0450)`, `(0,450,0450)`}, {`(500,0,0500)`, ``}, {`(550,0,0550)`, ``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1 RIGHT JOIN prt2 t2 ON t1.a = t2.b WHERE t2.a = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: (t1_1.a = t2_1.b)`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`Filter: (a = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (t1_2.a = t2_2.b)`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`Filter: (a = 0)`}, {`->  Nested Loop Left Join`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`Filter: (a = 0)`}, {`->  Index Scan using iprt1_p3_a on prt1_p3 t1_3`}, {`Index Cond: (a = t2_3.b)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1 RIGHT JOIN prt2 t2 ON t1.a = t2.b WHERE t2.a = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000"}, {150, "0150", 150, "0150"}, {300, "0300", 300, "0300"}, {450, "0450", 450, "0450"}, {``, ``, 75, 0075}, {``, ``, 225, "0225"}, {``, ``, 375, "0375"}, {``, ``, 525, "0525"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT 50 phv, * FROM prt1 WHERE prt1.b = 0) t1 FULL JOIN (SELECT 75 phv, * FROM prt2 WHERE prt2.a = 0) t2 ON (t1.a = t2.b) WHERE t1.phv = t1.a OR t2.phv = t2.b ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1.a, prt2.b`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_1.a = prt2_1.b)`}, {`Filter: (((50) = prt1_1.a) OR ((75) = prt2_1.b))`}, {`->  Seq Scan on prt1_p1 prt1_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 prt2_1`}, {`Filter: (a = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_2.a = prt2_2.b)`}, {`Filter: (((50) = prt1_2.a) OR ((75) = prt2_2.b))`}, {`->  Seq Scan on prt1_p2 prt1_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 prt2_2`}, {`Filter: (a = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_3.a = prt2_3.b)`}, {`Filter: (((50) = prt1_3.a) OR ((75) = prt2_3.b))`}, {`->  Seq Scan on prt1_p3 prt1_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 prt2_3`}, {`Filter: (a = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT 50 phv, * FROM prt1 WHERE prt1.b = 0) t1 FULL JOIN (SELECT 75 phv, * FROM prt2 WHERE prt2.a = 0) t2 ON (t1.a = t2.b) WHERE t1.phv = t1.a OR t2.phv = t2.b ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{50, 0050, ``, ``}, {``, ``, 75, 0075}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1, prt2 t2 WHERE t1.a = t2.b AND t1.a < 450 AND t2.b > 250 AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Join`}, {`Hash Cond: (t2.b = t1.a)`}, {`->  Seq Scan on prt2_p2 t2`}, {`Filter: (b > 250)`}, {`->  Hash`}, {`->  Seq Scan on prt1_p2 t1`}, {`Filter: ((a < 450) AND (b = 0))`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1, prt2 t2 WHERE t1.a = t2.b AND t1.a < 450 AND t2.b > 250 AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{300, "0300", 300, "0300"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1 WHERE a < 450) t1 LEFT JOIN (SELECT * FROM prt2 WHERE b > 250) t2 ON t1.a = t2.b WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1.a, prt2.b`}, {`->  Hash Right Join`}, {`Hash Cond: (prt2.b = prt1.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_p2 prt2_1`}, {`Filter: (b > 250)`}, {`->  Seq Scan on prt2_p3 prt2_2`}, {`Filter: (b > 250)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 prt1_1`}, {`Filter: ((a < 450) AND (b = 0))`}, {`->  Seq Scan on prt1_p2 prt1_2`}, {`Filter: ((a < 450) AND (b = 0))`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1 WHERE a < 450) t1 LEFT JOIN (SELECT * FROM prt2 WHERE b > 250) t2 ON t1.a = t2.b WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", ``, ``}, {50, 0050, ``, ``}, {100, "0100", ``, ``}, {150, "0150", ``, ``}, {200, "0200", ``, ``}, {250, "0250", ``, ``}, {300, "0300", 300, "0300"}, {350, "0350", ``, ``}, {400, 0400, ``, ``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1 WHERE a < 450) t1 FULL JOIN (SELECT * FROM prt2 WHERE b > 250) t2 ON t1.a = t2.b WHERE t1.b = 0 OR t2.a = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1.a, prt2.b`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1.a = prt2.b)`}, {`Filter: ((prt1.b = 0) OR (prt2.a = 0))`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 prt1_1`}, {`Filter: (a < 450)`}, {`->  Seq Scan on prt1_p2 prt1_2`}, {`Filter: (a < 450)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt2_p2 prt2_1`}, {`Filter: (b > 250)`}, {`->  Seq Scan on prt2_p3 prt2_2`}, {`Filter: (b > 250)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1 WHERE a < 450) t1 FULL JOIN (SELECT * FROM prt2 WHERE b > 250) t2 ON t1.a = t2.b WHERE t1.b = 0 OR t2.a = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", ``, ``}, {50, 0050, ``, ``}, {100, "0100", ``, ``}, {150, "0150", ``, ``}, {200, "0200", ``, ``}, {250, "0250", ``, ``}, {300, "0300", 300, "0300"}, {350, "0350", ``, ``}, {400, 0400, ``, ``}, {``, ``, 375, "0375"}, {``, ``, 450, "0450"}, {``, ``, 525, "0525"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1 t1 WHERE t1.a IN (SELECT t2.b FROM prt2 t2 WHERE t2.a = 0) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_1.a = t2_1.b)`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`Filter: (a = 0)`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_2.a = t2_2.b)`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`Filter: (a = 0)`}, {`->  Nested Loop Semi Join`}, {`Join Filter: (t1_3.a = t2_3.b)`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Materialize`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`Filter: (a = 0)`}},
			},
			{
				Statement: `SELECT t1.* FROM prt1 t1 WHERE t1.a IN (SELECT t2.b FROM prt2 t2 WHERE t2.a = 0) AND t1.b = 0 ORDER BY t1.a;`,
				Results:   []sql.Row{{0, 0, "0000"}, {150, 0, "0150"}, {300, 0, "0300"}, {450, 0, "0450"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT sum(t1.a), avg(t1.a), sum(t1.b), avg(t1.b) FROM prt1 t1 WHERE NOT EXISTS (SELECT 1 FROM prt2 t2 WHERE t1.a = t2.b);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Append`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_1.a = t2_1.b)`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_2.a = t2_2.b)`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_3.a = t2_3.b)`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 t2_3`}},
			},
			{
				Statement: `SELECT sum(t1.a), avg(t1.a), sum(t1.b), avg(t1.b) FROM prt1 t1 WHERE NOT EXISTS (SELECT 1 FROM prt2 t2 WHERE t1.a = t2.b);`,
				Results:   []sql.Row{{6, "0000", 300.000000000000, "0000", 2400, 12.000000000000, "0000"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM prt1 t1 LEFT JOIN LATERAL
			  (SELECT t2.a AS t2a, t3.a AS t3a, least(t1.a,t2.a,t3.b) FROM prt1 t2 JOIN prt2 t3 ON (t2.a = t3.b)) ss
			  ON t1.a = ss.t2a WHERE t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop Left Join`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Nested Loop`}, {`->  Index Only Scan using iprt1_p1_a on prt1_p1 t2_1`}, {`Index Cond: (a = t1_1.a)`}, {`->  Index Scan using iprt2_p1_b on prt2_p1 t3_1`}, {`Index Cond: (b = t2_1.a)`}, {`->  Nested Loop Left Join`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Nested Loop`}, {`->  Index Only Scan using iprt1_p2_a on prt1_p2 t2_2`}, {`Index Cond: (a = t1_2.a)`}, {`->  Index Scan using iprt2_p2_b on prt2_p2 t3_2`}, {`Index Cond: (b = t2_2.a)`}, {`->  Nested Loop Left Join`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Nested Loop`}, {`->  Index Only Scan using iprt1_p3_a on prt1_p3 t2_3`}, {`Index Cond: (a = t1_3.a)`}, {`->  Index Scan using iprt2_p3_b on prt2_p3 t3_3`}, {`Index Cond: (b = t2_3.a)`}},
			},
			{
				Statement: `SELECT * FROM prt1 t1 LEFT JOIN LATERAL
			  (SELECT t2.a AS t2a, t3.a AS t3a, least(t1.a,t2.a,t3.b) FROM prt1 t2 JOIN prt2 t3 ON (t2.a = t3.b)) ss
			  ON t1.a = ss.t2a WHERE t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{0, 0, "0000", 0, 0, 0}, {50, 0, 0050, ``, ``, ``}, {100, 0, "0100", ``, ``, ``}, {150, 0, "0150", 150, 0, 150}, {200, 0, "0200", ``, ``, ``}, {250, 0, "0250", ``, ``, ``}, {300, 0, "0300", 300, 0, 300}, {350, 0, "0350", ``, ``, ``}, {400, 0, 0400, ``, ``, ``}, {450, 0, "0450", 450, 0, 450}, {500, 0, "0500", ``, ``, ``}, {550, 0, "0550", ``, ``, ``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, ss.t2a, ss.t2c FROM prt1 t1 LEFT JOIN LATERAL
			  (SELECT t2.a AS t2a, t3.a AS t3a, t2.b t2b, t2.c t2c, least(t1.a,t2.a,t3.b) FROM prt1 t2 JOIN prt2 t3 ON (t2.a = t3.b)) ss
			  ON t1.c = ss.t2c WHERE (t1.b + coalesce(ss.t2b, 0)) = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Left Join`}, {`Hash Cond: ((t1.c)::text = (t2.c)::text)`}, {`Filter: ((t1.b + COALESCE(t2.b, 0)) = 0)`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`->  Hash`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.a = t3_1.b)`}, {`->  Seq Scan on prt1_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 t3_1`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.a = t3_2.b)`}, {`->  Seq Scan on prt1_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 t3_2`}, {`->  Hash Join`}, {`Hash Cond: (t2_3.a = t3_3.b)`}, {`->  Seq Scan on prt1_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 t3_3`}},
			},
			{
				Statement: `SELECT t1.a, ss.t2a, ss.t2c FROM prt1 t1 LEFT JOIN LATERAL
			  (SELECT t2.a AS t2a, t3.a AS t3a, t2.b t2b, t2.c t2c, least(t1.a,t2.a,t3.a) FROM prt1 t2 JOIN prt2 t3 ON (t2.a = t3.b)) ss
			  ON t1.c = ss.t2c WHERE (t1.b + coalesce(ss.t2b, 0)) = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{0, 0, "0000"}, {50, ``, ``}, {100, ``, ``}, {150, 150, "0150"}, {200, ``, ``}, {250, ``, ``}, {300, 300, "0300"}, {350, ``, ``}, {400, ``, ``}, {450, 450, "0450"}, {500, ``, ``}, {550, ``, ``}},
			},
			{
				Statement: `SET enable_partitionwise_aggregate TO true;`,
			},
			{
				Statement: `SET enable_hashjoin TO false;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT a, b FROM prt1 FULL JOIN prt2 p2(b,a,c) USING(a,b)
  WHERE a BETWEEN 490 AND 510
  GROUP BY 1, 2 ORDER BY 1, 2;`,
				Results: []sql.Row{{`Group`}, {`Group Key: (COALESCE(prt1.a, p2.a)), (COALESCE(prt1.b, p2.b))`}, {`->  Merge Append`}, {`Sort Key: (COALESCE(prt1.a, p2.a)), (COALESCE(prt1.b, p2.b))`}, {`->  Group`}, {`Group Key: (COALESCE(prt1.a, p2.a)), (COALESCE(prt1.b, p2.b))`}, {`->  Sort`}, {`Sort Key: (COALESCE(prt1.a, p2.a)), (COALESCE(prt1.b, p2.b))`}, {`->  Merge Full Join`}, {`Merge Cond: ((prt1.a = p2.a) AND (prt1.b = p2.b))`}, {`Filter: ((COALESCE(prt1.a, p2.a) >= 490) AND (COALESCE(prt1.a, p2.a) <= 510))`}, {`->  Sort`}, {`Sort Key: prt1.a, prt1.b`}, {`->  Seq Scan on prt1_p1 prt1`}, {`->  Sort`}, {`Sort Key: p2.a, p2.b`}, {`->  Seq Scan on prt2_p1 p2`}, {`->  Group`}, {`Group Key: (COALESCE(prt1_1.a, p2_1.a)), (COALESCE(prt1_1.b, p2_1.b))`}, {`->  Sort`}, {`Sort Key: (COALESCE(prt1_1.a, p2_1.a)), (COALESCE(prt1_1.b, p2_1.b))`}, {`->  Merge Full Join`}, {`Merge Cond: ((prt1_1.a = p2_1.a) AND (prt1_1.b = p2_1.b))`}, {`Filter: ((COALESCE(prt1_1.a, p2_1.a) >= 490) AND (COALESCE(prt1_1.a, p2_1.a) <= 510))`}, {`->  Sort`}, {`Sort Key: prt1_1.a, prt1_1.b`}, {`->  Seq Scan on prt1_p2 prt1_1`}, {`->  Sort`}, {`Sort Key: p2_1.a, p2_1.b`}, {`->  Seq Scan on prt2_p2 p2_1`}, {`->  Group`}, {`Group Key: (COALESCE(prt1_2.a, p2_2.a)), (COALESCE(prt1_2.b, p2_2.b))`}, {`->  Sort`}, {`Sort Key: (COALESCE(prt1_2.a, p2_2.a)), (COALESCE(prt1_2.b, p2_2.b))`}, {`->  Merge Full Join`}, {`Merge Cond: ((prt1_2.a = p2_2.a) AND (prt1_2.b = p2_2.b))`}, {`Filter: ((COALESCE(prt1_2.a, p2_2.a) >= 490) AND (COALESCE(prt1_2.a, p2_2.a) <= 510))`}, {`->  Sort`}, {`Sort Key: prt1_2.a, prt1_2.b`}, {`->  Seq Scan on prt1_p3 prt1_2`}, {`->  Sort`}, {`Sort Key: p2_2.a, p2_2.b`}, {`->  Seq Scan on prt2_p3 p2_2`}},
			},
			{
				Statement: `SELECT a, b FROM prt1 FULL JOIN prt2 p2(b,a,c) USING(a,b)
  WHERE a BETWEEN 490 AND 510
  GROUP BY 1, 2 ORDER BY 1, 2;`,
				Results: []sql.Row{{490, 15}, {492, 17}, {494, 19}, {495, 20}, {496, 21}, {498, 23}, {500, 0}, {501, 1}, {502, 2}, {504, 4}, {506, 6}, {507, 7}, {508, 8}, {510, 10}},
			},
			{
				Statement: `RESET enable_partitionwise_aggregate;`,
			},
			{
				Statement: `RESET enable_hashjoin;`,
			},
			{
				Statement: `CREATE TABLE prt1_e (a int, b int, c int) PARTITION BY RANGE(((a + b)/2));`,
			},
			{
				Statement: `CREATE TABLE prt1_e_p1 PARTITION OF prt1_e FOR VALUES FROM (0) TO (250);`,
			},
			{
				Statement: `CREATE TABLE prt1_e_p2 PARTITION OF prt1_e FOR VALUES FROM (250) TO (500);`,
			},
			{
				Statement: `CREATE TABLE prt1_e_p3 PARTITION OF prt1_e FOR VALUES FROM (500) TO (600);`,
			},
			{
				Statement: `INSERT INTO prt1_e SELECT i, i, i % 25 FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `CREATE INDEX iprt1_e_p1_ab2 on prt1_e_p1(((a+b)/2));`,
			},
			{
				Statement: `CREATE INDEX iprt1_e_p2_ab2 on prt1_e_p2(((a+b)/2));`,
			},
			{
				Statement: `CREATE INDEX iprt1_e_p3_ab2 on prt1_e_p3(((a+b)/2));`,
			},
			{
				Statement: `ANALYZE prt1_e;`,
			},
			{
				Statement: `CREATE TABLE prt2_e (a int, b int, c int) PARTITION BY RANGE(((b + a)/2));`,
			},
			{
				Statement: `CREATE TABLE prt2_e_p1 PARTITION OF prt2_e FOR VALUES FROM (0) TO (250);`,
			},
			{
				Statement: `CREATE TABLE prt2_e_p2 PARTITION OF prt2_e FOR VALUES FROM (250) TO (500);`,
			},
			{
				Statement: `CREATE TABLE prt2_e_p3 PARTITION OF prt2_e FOR VALUES FROM (500) TO (600);`,
			},
			{
				Statement: `INSERT INTO prt2_e SELECT i, i, i % 25 FROM generate_series(0, 599, 3) i;`,
			},
			{
				Statement: `ANALYZE prt2_e;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_e t1, prt2_e t2 WHERE (t1.a + t1.b)/2 = (t2.b + t2.a)/2 AND t1.c = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (((t2_1.b + t2_1.a) / 2) = ((t1_1.a + t1_1.b) / 2))`}, {`->  Seq Scan on prt2_e_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p1 t1_1`}, {`Filter: (c = 0)`}, {`->  Hash Join`}, {`Hash Cond: (((t2_2.b + t2_2.a) / 2) = ((t1_2.a + t1_2.b) / 2))`}, {`->  Seq Scan on prt2_e_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p2 t1_2`}, {`Filter: (c = 0)`}, {`->  Hash Join`}, {`Hash Cond: (((t2_3.b + t2_3.a) / 2) = ((t1_3.a + t1_3.b) / 2))`}, {`->  Seq Scan on prt2_e_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p3 t1_3`}, {`Filter: (c = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_e t1, prt2_e t2 WHERE (t1.a + t1.b)/2 = (t2.b + t2.a)/2 AND t1.c = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, 0, 0, 0}, {150, 0, 150, 0}, {300, 0, 300, 0}, {450, 0, 450, 0}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c, t3.a + t3.b, t3.c FROM prt1 t1, prt2 t2, prt1_e t3 WHERE t1.a = t2.b AND t1.a = (t3.a + t3.b)/2 AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop`}, {`Join Filter: (t1_1.a = ((t3_1.a + t3_1.b) / 2))`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Index Scan using iprt1_e_p1_ab2 on prt1_e_p1 t3_1`}, {`Index Cond: (((a + b) / 2) = t2_1.b)`}, {`->  Nested Loop`}, {`Join Filter: (t1_2.a = ((t3_2.a + t3_2.b) / 2))`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Index Scan using iprt1_e_p2_ab2 on prt1_e_p2 t3_2`}, {`Index Cond: (((a + b) / 2) = t2_2.b)`}, {`->  Nested Loop`}, {`Join Filter: (t1_3.a = ((t3_3.a + t3_3.b) / 2))`}, {`->  Hash Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Index Scan using iprt1_e_p3_ab2 on prt1_e_p3 t3_3`}, {`Index Cond: (((a + b) / 2) = t2_3.b)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c, t3.a + t3.b, t3.c FROM prt1 t1, prt2 t2, prt1_e t3 WHERE t1.a = t2.b AND t1.a = (t3.a + t3.b)/2 AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000", 0, 0}, {150, "0150", 150, "0150", 300, 0}, {300, "0300", 300, "0300", 600, 0}, {450, "0450", 450, "0450", 900, 0}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c, t3.a + t3.b, t3.c FROM (prt1 t1 LEFT JOIN prt2 t2 ON t1.a = t2.b) LEFT JOIN prt1_e t3 ON (t1.a = (t3.a + t3.b)/2) WHERE t1.b = 0 ORDER BY t1.a, t2.b, t3.a + t3.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b, ((t3.a + t3.b))`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: (((t3_1.a + t3_1.b) / 2) = t1_1.a)`}, {`->  Seq Scan on prt1_e_p1 t3_1`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (((t3_2.a + t3_2.b) / 2) = t1_2.a)`}, {`->  Seq Scan on prt1_e_p2 t3_2`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (((t3_3.a + t3_3.b) / 2) = t1_3.a)`}, {`->  Seq Scan on prt1_e_p3 t3_3`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c, t3.a + t3.b, t3.c FROM (prt1 t1 LEFT JOIN prt2 t2 ON t1.a = t2.b) LEFT JOIN prt1_e t3 ON (t1.a = (t3.a + t3.b)/2) WHERE t1.b = 0 ORDER BY t1.a, t2.b, t3.a + t3.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000", 0, 0}, {50, 0050, ``, ``, 100, 0}, {100, "0100", ``, ``, 200, 0}, {150, "0150", 150, "0150", 300, 0}, {200, "0200", ``, ``, 400, 0}, {250, "0250", ``, ``, 500, 0}, {300, "0300", 300, "0300", 600, 0}, {350, "0350", ``, ``, 700, 0}, {400, 0400, ``, ``, 800, 0}, {450, "0450", 450, "0450", 900, 0}, {500, "0500", ``, ``, 1000, 0}, {550, "0550", ``, ``, 1100, 0}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c, t3.a + t3.b, t3.c FROM (prt1 t1 LEFT JOIN prt2 t2 ON t1.a = t2.b) RIGHT JOIN prt1_e t3 ON (t1.a = (t3.a + t3.b)/2) WHERE t3.c = 0 ORDER BY t1.a, t2.b, t3.a + t3.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b, ((t3.a + t3.b))`}, {`->  Append`}, {`->  Nested Loop Left Join`}, {`->  Hash Right Join`}, {`Hash Cond: (t1_1.a = ((t3_1.a + t3_1.b) / 2))`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p1 t3_1`}, {`Filter: (c = 0)`}, {`->  Index Scan using iprt2_p1_b on prt2_p1 t2_1`}, {`Index Cond: (b = t1_1.a)`}, {`->  Nested Loop Left Join`}, {`->  Hash Right Join`}, {`Hash Cond: (t1_2.a = ((t3_2.a + t3_2.b) / 2))`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p2 t3_2`}, {`Filter: (c = 0)`}, {`->  Index Scan using iprt2_p2_b on prt2_p2 t2_2`}, {`Index Cond: (b = t1_2.a)`}, {`->  Nested Loop Left Join`}, {`->  Hash Right Join`}, {`Hash Cond: (t1_3.a = ((t3_3.a + t3_3.b) / 2))`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p3 t3_3`}, {`Filter: (c = 0)`}, {`->  Index Scan using iprt2_p3_b on prt2_p3 t2_3`}, {`Index Cond: (b = t1_3.a)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c, t3.a + t3.b, t3.c FROM (prt1 t1 LEFT JOIN prt2 t2 ON t1.a = t2.b) RIGHT JOIN prt1_e t3 ON (t1.a = (t3.a + t3.b)/2) WHERE t3.c = 0 ORDER BY t1.a, t2.b, t3.a + t3.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000", 0, 0}, {50, 0050, ``, ``, 100, 0}, {100, "0100", ``, ``, 200, 0}, {150, "0150", 150, "0150", 300, 0}, {200, "0200", ``, ``, 400, 0}, {250, "0250", ``, ``, 500, 0}, {300, "0300", 300, "0300", 600, 0}, {350, "0350", ``, ``, 700, 0}, {400, 0400, ``, ``, 800, 0}, {450, "0450", 450, "0450", 900, 0}, {500, "0500", ``, ``, 1000, 0}, {550, "0550", ``, ``, 1100, 0}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT COUNT(*) FROM prt1 FULL JOIN prt2 p2(b,a,c) USING(a,b) FULL JOIN prt2 p3(b,a,c) USING (a, b)
  WHERE a BETWEEN 490 AND 510;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(prt1_1.a, p2_1.a) = p3_1.a) AND (COALESCE(prt1_1.b, p2_1.b) = p3_1.b))`}, {`Filter: ((COALESCE(COALESCE(prt1_1.a, p2_1.a), p3_1.a) >= 490) AND (COALESCE(COALESCE(prt1_1.a, p2_1.a), p3_1.a) <= 510))`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_1.a = p2_1.a) AND (prt1_1.b = p2_1.b))`}, {`->  Seq Scan on prt1_p1 prt1_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 p2_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 p3_1`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(prt1_2.a, p2_2.a) = p3_2.a) AND (COALESCE(prt1_2.b, p2_2.b) = p3_2.b))`}, {`Filter: ((COALESCE(COALESCE(prt1_2.a, p2_2.a), p3_2.a) >= 490) AND (COALESCE(COALESCE(prt1_2.a, p2_2.a), p3_2.a) <= 510))`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_2.a = p2_2.a) AND (prt1_2.b = p2_2.b))`}, {`->  Seq Scan on prt1_p2 prt1_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 p2_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 p3_2`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(prt1_3.a, p2_3.a) = p3_3.a) AND (COALESCE(prt1_3.b, p2_3.b) = p3_3.b))`}, {`Filter: ((COALESCE(COALESCE(prt1_3.a, p2_3.a), p3_3.a) >= 490) AND (COALESCE(COALESCE(prt1_3.a, p2_3.a), p3_3.a) <= 510))`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_3.a = p2_3.a) AND (prt1_3.b = p2_3.b))`}, {`->  Seq Scan on prt1_p3 prt1_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 p2_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 p3_3`}},
			},
			{
				Statement: `SELECT COUNT(*) FROM prt1 FULL JOIN prt2 p2(b,a,c) USING(a,b) FULL JOIN prt2 p3(b,a,c) USING (a, b)
  WHERE a BETWEEN 490 AND 510;`,
				Results: []sql.Row{{14}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT COUNT(*) FROM prt1 FULL JOIN prt2 p2(b,a,c) USING(a,b) FULL JOIN prt2 p3(b,a,c) USING (a, b) FULL JOIN prt1 p4 (a,b,c) USING (a, b)
  WHERE a BETWEEN 490 AND 510;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(COALESCE(prt1_1.a, p2_1.a), p3_1.a) = p4_1.a) AND (COALESCE(COALESCE(prt1_1.b, p2_1.b), p3_1.b) = p4_1.b))`}, {`Filter: ((COALESCE(COALESCE(COALESCE(prt1_1.a, p2_1.a), p3_1.a), p4_1.a) >= 490) AND (COALESCE(COALESCE(COALESCE(prt1_1.a, p2_1.a), p3_1.a), p4_1.a) <= 510))`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(prt1_1.a, p2_1.a) = p3_1.a) AND (COALESCE(prt1_1.b, p2_1.b) = p3_1.b))`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_1.a = p2_1.a) AND (prt1_1.b = p2_1.b))`}, {`->  Seq Scan on prt1_p1 prt1_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 p2_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 p3_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_p1 p4_1`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(COALESCE(prt1_2.a, p2_2.a), p3_2.a) = p4_2.a) AND (COALESCE(COALESCE(prt1_2.b, p2_2.b), p3_2.b) = p4_2.b))`}, {`Filter: ((COALESCE(COALESCE(COALESCE(prt1_2.a, p2_2.a), p3_2.a), p4_2.a) >= 490) AND (COALESCE(COALESCE(COALESCE(prt1_2.a, p2_2.a), p3_2.a), p4_2.a) <= 510))`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(prt1_2.a, p2_2.a) = p3_2.a) AND (COALESCE(prt1_2.b, p2_2.b) = p3_2.b))`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_2.a = p2_2.a) AND (prt1_2.b = p2_2.b))`}, {`->  Seq Scan on prt1_p2 prt1_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 p2_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 p3_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_p2 p4_2`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(COALESCE(prt1_3.a, p2_3.a), p3_3.a) = p4_3.a) AND (COALESCE(COALESCE(prt1_3.b, p2_3.b), p3_3.b) = p4_3.b))`}, {`Filter: ((COALESCE(COALESCE(COALESCE(prt1_3.a, p2_3.a), p3_3.a), p4_3.a) >= 490) AND (COALESCE(COALESCE(COALESCE(prt1_3.a, p2_3.a), p3_3.a), p4_3.a) <= 510))`}, {`->  Hash Full Join`}, {`Hash Cond: ((COALESCE(prt1_3.a, p2_3.a) = p3_3.a) AND (COALESCE(prt1_3.b, p2_3.b) = p3_3.b))`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_3.a = p2_3.a) AND (prt1_3.b = p2_3.b))`}, {`->  Seq Scan on prt1_p3 prt1_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 p2_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 p3_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_p3 p4_3`}},
			},
			{
				Statement: `SELECT COUNT(*) FROM prt1 FULL JOIN prt2 p2(b,a,c) USING(a,b) FULL JOIN prt2 p3(b,a,c) USING (a, b) FULL JOIN prt1 p4 (a,b,c) USING (a, b)
  WHERE a BETWEEN 490 AND 510;`,
				Results: []sql.Row{{14}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.phv, t2.b, t2.phv, t3.a + t3.b, t3.phv FROM ((SELECT 50 phv, * FROM prt1 WHERE prt1.b = 0) t1 FULL JOIN (SELECT 75 phv, * FROM prt2 WHERE prt2.a = 0) t2 ON (t1.a = t2.b)) FULL JOIN (SELECT 50 phv, * FROM prt1_e WHERE prt1_e.c = 0) t3 ON (t1.a = (t3.a + t3.b)/2) WHERE t1.a = t1.phv OR t2.b = t2.phv OR (t3.a + t3.b)/2 = t3.phv ORDER BY t1.a, t2.b, t3.a + t3.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1.a, prt2.b, ((prt1_e.a + prt1_e.b))`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_1.a = ((prt1_e_1.a + prt1_e_1.b) / 2))`}, {`Filter: ((prt1_1.a = (50)) OR (prt2_1.b = (75)) OR (((prt1_e_1.a + prt1_e_1.b) / 2) = (50)))`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_1.a = prt2_1.b)`}, {`->  Seq Scan on prt1_p1 prt1_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 prt2_1`}, {`Filter: (a = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p1 prt1_e_1`}, {`Filter: (c = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_2.a = ((prt1_e_2.a + prt1_e_2.b) / 2))`}, {`Filter: ((prt1_2.a = (50)) OR (prt2_2.b = (75)) OR (((prt1_e_2.a + prt1_e_2.b) / 2) = (50)))`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_2.a = prt2_2.b)`}, {`->  Seq Scan on prt1_p2 prt1_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 prt2_2`}, {`Filter: (a = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p2 prt1_e_2`}, {`Filter: (c = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_3.a = ((prt1_e_3.a + prt1_e_3.b) / 2))`}, {`Filter: ((prt1_3.a = (50)) OR (prt2_3.b = (75)) OR (((prt1_e_3.a + prt1_e_3.b) / 2) = (50)))`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_3.a = prt2_3.b)`}, {`->  Seq Scan on prt1_p3 prt1_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 prt2_3`}, {`Filter: (a = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p3 prt1_e_3`}, {`Filter: (c = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.phv, t2.b, t2.phv, t3.a + t3.b, t3.phv FROM ((SELECT 50 phv, * FROM prt1 WHERE prt1.b = 0) t1 FULL JOIN (SELECT 75 phv, * FROM prt2 WHERE prt2.a = 0) t2 ON (t1.a = t2.b)) FULL JOIN (SELECT 50 phv, * FROM prt1_e WHERE prt1_e.c = 0) t3 ON (t1.a = (t3.a + t3.b)/2) WHERE t1.a = t1.phv OR t2.b = t2.phv OR (t3.a + t3.b)/2 = t3.phv ORDER BY t1.a, t2.b, t3.a + t3.b;`,
				Results:   []sql.Row{{50, 50, ``, ``, 100, 50}, {``, ``, 75, 75, ``, ``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1 t1 WHERE t1.a IN (SELECT t1.b FROM prt2 t1, prt1_e t2 WHERE t1.a = 0 AND t1.b = (t2.a + t2.b)/2) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop`}, {`Join Filter: (t1_2.a = t1_5.b)`}, {`->  HashAggregate`}, {`Group Key: t1_5.b`}, {`->  Hash Join`}, {`Hash Cond: (((t2_1.a + t2_1.b) / 2) = t1_5.b)`}, {`->  Seq Scan on prt1_e_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 t1_5`}, {`Filter: (a = 0)`}, {`->  Index Scan using iprt1_p1_a on prt1_p1 t1_2`}, {`Index Cond: (a = ((t2_1.a + t2_1.b) / 2))`}, {`Filter: (b = 0)`}, {`->  Nested Loop`}, {`Join Filter: (t1_3.a = t1_6.b)`}, {`->  HashAggregate`}, {`Group Key: t1_6.b`}, {`->  Hash Join`}, {`Hash Cond: (((t2_2.a + t2_2.b) / 2) = t1_6.b)`}, {`->  Seq Scan on prt1_e_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 t1_6`}, {`Filter: (a = 0)`}, {`->  Index Scan using iprt1_p2_a on prt1_p2 t1_3`}, {`Index Cond: (a = ((t2_2.a + t2_2.b) / 2))`}, {`Filter: (b = 0)`}, {`->  Nested Loop`}, {`Join Filter: (t1_4.a = t1_7.b)`}, {`->  HashAggregate`}, {`Group Key: t1_7.b`}, {`->  Nested Loop`}, {`->  Seq Scan on prt2_p3 t1_7`}, {`Filter: (a = 0)`}, {`->  Index Scan using iprt1_e_p3_ab2 on prt1_e_p3 t2_3`}, {`Index Cond: (((a + b) / 2) = t1_7.b)`}, {`->  Index Scan using iprt1_p3_a on prt1_p3 t1_4`}, {`Index Cond: (a = ((t2_3.a + t2_3.b) / 2))`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.* FROM prt1 t1 WHERE t1.a IN (SELECT t1.b FROM prt2 t1, prt1_e t2 WHERE t1.a = 0 AND t1.b = (t2.a + t2.b)/2) AND t1.b = 0 ORDER BY t1.a;`,
				Results:   []sql.Row{{0, 0, "0000"}, {150, 0, "0150"}, {300, 0, "0300"}, {450, 0, "0450"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1 t1 WHERE t1.a IN (SELECT t1.b FROM prt2 t1 WHERE t1.b IN (SELECT (t1.a + t1.b)/2 FROM prt1_e t1 WHERE t1.c = 0)) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop`}, {`->  HashAggregate`}, {`Group Key: t1_6.b`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_6.b = ((t1_9.a + t1_9.b) / 2))`}, {`->  Seq Scan on prt2_p1 t1_6`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p1 t1_9`}, {`Filter: (c = 0)`}, {`->  Index Scan using iprt1_p1_a on prt1_p1 t1_3`}, {`Index Cond: (a = t1_6.b)`}, {`Filter: (b = 0)`}, {`->  Nested Loop`}, {`->  HashAggregate`}, {`Group Key: t1_7.b`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_7.b = ((t1_10.a + t1_10.b) / 2))`}, {`->  Seq Scan on prt2_p2 t1_7`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p2 t1_10`}, {`Filter: (c = 0)`}, {`->  Index Scan using iprt1_p2_a on prt1_p2 t1_4`}, {`Index Cond: (a = t1_7.b)`}, {`Filter: (b = 0)`}, {`->  Nested Loop`}, {`->  HashAggregate`}, {`Group Key: t1_8.b`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_8.b = ((t1_11.a + t1_11.b) / 2))`}, {`->  Seq Scan on prt2_p3 t1_8`}, {`->  Hash`}, {`->  Seq Scan on prt1_e_p3 t1_11`}, {`Filter: (c = 0)`}, {`->  Index Scan using iprt1_p3_a on prt1_p3 t1_5`}, {`Index Cond: (a = t1_8.b)`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.* FROM prt1 t1 WHERE t1.a IN (SELECT t1.b FROM prt2 t1 WHERE t1.b IN (SELECT (t1.a + t1.b)/2 FROM prt1_e t1 WHERE t1.c = 0)) AND t1.b = 0 ORDER BY t1.a;`,
				Results:   []sql.Row{{0, 0, "0000"}, {150, 0, "0150"}, {300, 0, "0300"}, {450, 0, "0450"}},
			},
			{
				Statement: `SET enable_hashjoin TO off;`,
			},
			{
				Statement: `SET enable_nestloop TO off;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1 t1 WHERE t1.a IN (SELECT t1.b FROM prt2 t1 WHERE t1.b IN (SELECT (t1.a + t1.b)/2 FROM prt1_e t1 WHERE t1.c = 0)) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Merge Append`}, {`Sort Key: t1.a`}, {`->  Merge Semi Join`}, {`Merge Cond: (t1_3.a = t1_6.b)`}, {`->  Sort`}, {`Sort Key: t1_3.a`}, {`->  Seq Scan on prt1_p1 t1_3`}, {`Filter: (b = 0)`}, {`->  Merge Semi Join`}, {`Merge Cond: (t1_6.b = (((t1_9.a + t1_9.b) / 2)))`}, {`->  Sort`}, {`Sort Key: t1_6.b`}, {`->  Seq Scan on prt2_p1 t1_6`}, {`->  Sort`}, {`Sort Key: (((t1_9.a + t1_9.b) / 2))`}, {`->  Seq Scan on prt1_e_p1 t1_9`}, {`Filter: (c = 0)`}, {`->  Merge Semi Join`}, {`Merge Cond: (t1_4.a = t1_7.b)`}, {`->  Sort`}, {`Sort Key: t1_4.a`}, {`->  Seq Scan on prt1_p2 t1_4`}, {`Filter: (b = 0)`}, {`->  Merge Semi Join`}, {`Merge Cond: (t1_7.b = (((t1_10.a + t1_10.b) / 2)))`}, {`->  Sort`}, {`Sort Key: t1_7.b`}, {`->  Seq Scan on prt2_p2 t1_7`}, {`->  Sort`}, {`Sort Key: (((t1_10.a + t1_10.b) / 2))`}, {`->  Seq Scan on prt1_e_p2 t1_10`}, {`Filter: (c = 0)`}, {`->  Merge Semi Join`}, {`Merge Cond: (t1_5.a = t1_8.b)`}, {`->  Sort`}, {`Sort Key: t1_5.a`}, {`->  Seq Scan on prt1_p3 t1_5`}, {`Filter: (b = 0)`}, {`->  Merge Semi Join`}, {`Merge Cond: (t1_8.b = (((t1_11.a + t1_11.b) / 2)))`}, {`->  Sort`}, {`Sort Key: t1_8.b`}, {`->  Seq Scan on prt2_p3 t1_8`}, {`->  Sort`}, {`Sort Key: (((t1_11.a + t1_11.b) / 2))`}, {`->  Seq Scan on prt1_e_p3 t1_11`}, {`Filter: (c = 0)`}},
			},
			{
				Statement: `SELECT t1.* FROM prt1 t1 WHERE t1.a IN (SELECT t1.b FROM prt2 t1 WHERE t1.b IN (SELECT (t1.a + t1.b)/2 FROM prt1_e t1 WHERE t1.c = 0)) AND t1.b = 0 ORDER BY t1.a;`,
				Results:   []sql.Row{{0, 0, "0000"}, {150, 0, "0150"}, {300, 0, "0300"}, {450, 0, "0450"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c, t3.a + t3.b, t3.c FROM (prt1 t1 LEFT JOIN prt2 t2 ON t1.a = t2.b) RIGHT JOIN prt1_e t3 ON (t1.a = (t3.a + t3.b)/2) WHERE t3.c = 0 ORDER BY t1.a, t2.b, t3.a + t3.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b, ((t3.a + t3.b))`}, {`->  Append`}, {`->  Merge Left Join`}, {`Merge Cond: (t1_1.a = t2_1.b)`}, {`->  Sort`}, {`Sort Key: t1_1.a`}, {`->  Merge Left Join`}, {`Merge Cond: ((((t3_1.a + t3_1.b) / 2)) = t1_1.a)`}, {`->  Sort`}, {`Sort Key: (((t3_1.a + t3_1.b) / 2))`}, {`->  Seq Scan on prt1_e_p1 t3_1`}, {`Filter: (c = 0)`}, {`->  Sort`}, {`Sort Key: t1_1.a`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Sort`}, {`Sort Key: t2_1.b`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Merge Left Join`}, {`Merge Cond: (t1_2.a = t2_2.b)`}, {`->  Sort`}, {`Sort Key: t1_2.a`}, {`->  Merge Left Join`}, {`Merge Cond: ((((t3_2.a + t3_2.b) / 2)) = t1_2.a)`}, {`->  Sort`}, {`Sort Key: (((t3_2.a + t3_2.b) / 2))`}, {`->  Seq Scan on prt1_e_p2 t3_2`}, {`Filter: (c = 0)`}, {`->  Sort`}, {`Sort Key: t1_2.a`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Sort`}, {`Sort Key: t2_2.b`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Merge Left Join`}, {`Merge Cond: (t1_3.a = t2_3.b)`}, {`->  Sort`}, {`Sort Key: t1_3.a`}, {`->  Merge Left Join`}, {`Merge Cond: ((((t3_3.a + t3_3.b) / 2)) = t1_3.a)`}, {`->  Sort`}, {`Sort Key: (((t3_3.a + t3_3.b) / 2))`}, {`->  Seq Scan on prt1_e_p3 t3_3`}, {`Filter: (c = 0)`}, {`->  Sort`}, {`Sort Key: t1_3.a`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`->  Sort`}, {`Sort Key: t2_3.b`}, {`->  Seq Scan on prt2_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c, t3.a + t3.b, t3.c FROM (prt1 t1 LEFT JOIN prt2 t2 ON t1.a = t2.b) RIGHT JOIN prt1_e t3 ON (t1.a = (t3.a + t3.b)/2) WHERE t3.c = 0 ORDER BY t1.a, t2.b, t3.a + t3.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000", 0, 0}, {50, 0050, ``, ``, 100, 0}, {100, "0100", ``, ``, 200, 0}, {150, "0150", 150, "0150", 300, 0}, {200, "0200", ``, ``, 400, 0}, {250, "0250", ``, ``, 500, 0}, {300, "0300", 300, "0300", 600, 0}, {350, "0350", ``, ``, 700, 0}, {400, 0400, ``, ``, 800, 0}, {450, "0450", 450, "0450", 900, 0}, {500, "0500", ``, ``, 1000, 0}, {550, "0550", ``, ``, 1100, 0}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t2.b FROM (SELECT * FROM prt1 WHERE a < 450) t1 LEFT JOIN (SELECT * FROM prt2 WHERE b > 250) t2 ON t1.a = t2.b WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1.a, prt2.b`}, {`->  Merge Left Join`}, {`Merge Cond: (prt1.a = prt2.b)`}, {`->  Sort`}, {`Sort Key: prt1.a`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 prt1_1`}, {`Filter: ((a < 450) AND (b = 0))`}, {`->  Seq Scan on prt1_p2 prt1_2`}, {`Filter: ((a < 450) AND (b = 0))`}, {`->  Sort`}, {`Sort Key: prt2.b`}, {`->  Append`}, {`->  Seq Scan on prt2_p2 prt2_1`}, {`Filter: (b > 250)`}, {`->  Seq Scan on prt2_p3 prt2_2`}, {`Filter: (b > 250)`}},
			},
			{
				Statement: `SELECT t1.a, t2.b FROM (SELECT * FROM prt1 WHERE a < 450) t1 LEFT JOIN (SELECT * FROM prt2 WHERE b > 250) t2 ON t1.a = t2.b WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, ``}, {50, ``}, {100, ``}, {150, ``}, {200, ``}, {250, ``}, {300, 300}, {350, ``}, {400, ``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t2.b FROM prt1 t1, prt2 t2 WHERE t1::text = t2::text AND t1.a = t2.b ORDER BY t1.a;`,
				Results: []sql.Row{{`Merge Join`}, {`Merge Cond: ((t1.a = t2.b) AND (((((t1.*)::prt1))::text) = ((((t2.*)::prt2))::text)))`}, {`->  Sort`}, {`Sort Key: t1.a, ((((t1.*)::prt1))::text)`}, {`->  Result`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`->  Sort`}, {`Sort Key: t2.b, ((((t2.*)::prt2))::text)`}, {`->  Result`}, {`->  Append`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Seq Scan on prt2_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.a, t2.b FROM prt1 t1, prt2 t2 WHERE t1::text = t2::text AND t1.a = t2.b ORDER BY t1.a;`,
				Results:   []sql.Row{{0, 0}, {6, 6}, {12, 12}, {18, 18}, {24, 24}},
			},
			{
				Statement: `RESET enable_hashjoin;`,
			},
			{
				Statement: `RESET enable_nestloop;`,
			},
			{
				Statement: `CREATE TABLE prt1_m (a int, b int, c int) PARTITION BY RANGE(a, ((a + b)/2));`,
			},
			{
				Statement: `CREATE TABLE prt1_m_p1 PARTITION OF prt1_m FOR VALUES FROM (0, 0) TO (250, 250);`,
			},
			{
				Statement: `CREATE TABLE prt1_m_p2 PARTITION OF prt1_m FOR VALUES FROM (250, 250) TO (500, 500);`,
			},
			{
				Statement: `CREATE TABLE prt1_m_p3 PARTITION OF prt1_m FOR VALUES FROM (500, 500) TO (600, 600);`,
			},
			{
				Statement: `INSERT INTO prt1_m SELECT i, i, i % 25 FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `ANALYZE prt1_m;`,
			},
			{
				Statement: `CREATE TABLE prt2_m (a int, b int, c int) PARTITION BY RANGE(((b + a)/2), b);`,
			},
			{
				Statement: `CREATE TABLE prt2_m_p1 PARTITION OF prt2_m FOR VALUES FROM (0, 0) TO (250, 250);`,
			},
			{
				Statement: `CREATE TABLE prt2_m_p2 PARTITION OF prt2_m FOR VALUES FROM (250, 250) TO (500, 500);`,
			},
			{
				Statement: `CREATE TABLE prt2_m_p3 PARTITION OF prt2_m FOR VALUES FROM (500, 500) TO (600, 600);`,
			},
			{
				Statement: `INSERT INTO prt2_m SELECT i, i, i % 25 FROM generate_series(0, 599, 3) i;`,
			},
			{
				Statement: `ANALYZE prt2_m;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1_m WHERE prt1_m.c = 0) t1 FULL JOIN (SELECT * FROM prt2_m WHERE prt2_m.c = 0) t2 ON (t1.a = (t2.b + t2.a)/2 AND t2.b = (t1.a + t1.b)/2) ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1_m.a, prt2_m.b`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_m_1.a = ((prt2_m_1.b + prt2_m_1.a) / 2)) AND (((prt1_m_1.a + prt1_m_1.b) / 2) = prt2_m_1.b))`}, {`->  Seq Scan on prt1_m_p1 prt1_m_1`}, {`Filter: (c = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_m_p1 prt2_m_1`}, {`Filter: (c = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_m_2.a = ((prt2_m_2.b + prt2_m_2.a) / 2)) AND (((prt1_m_2.a + prt1_m_2.b) / 2) = prt2_m_2.b))`}, {`->  Seq Scan on prt1_m_p2 prt1_m_2`}, {`Filter: (c = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_m_p2 prt2_m_2`}, {`Filter: (c = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_m_3.a = ((prt2_m_3.b + prt2_m_3.a) / 2)) AND (((prt1_m_3.a + prt1_m_3.b) / 2) = prt2_m_3.b))`}, {`->  Seq Scan on prt1_m_p3 prt1_m_3`}, {`Filter: (c = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_m_p3 prt2_m_3`}, {`Filter: (c = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1_m WHERE prt1_m.c = 0) t1 FULL JOIN (SELECT * FROM prt2_m WHERE prt2_m.c = 0) t2 ON (t1.a = (t2.b + t2.a)/2 AND t2.b = (t1.a + t1.b)/2) ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, 0, 0, 0}, {50, 0, ``, ``}, {100, 0, ``, ``}, {150, 0, 150, 0}, {200, 0, ``, ``}, {250, 0, ``, ``}, {300, 0, 300, 0}, {350, 0, ``, ``}, {400, 0, ``, ``}, {450, 0, 450, 0}, {500, 0, ``, ``}, {550, 0, ``, ``}, {``, ``, 75, 0}, {``, ``, 225, 0}, {``, ``, 375, 0}, {``, ``, 525, 0}},
			},
			{
				Statement: `CREATE TABLE plt1 (a int, b int, c text) PARTITION BY LIST(c);`,
			},
			{
				Statement: `CREATE TABLE plt1_p1 PARTITION OF plt1 FOR VALUES IN ('0000', '0003', '0004', '0010');`,
			},
			{
				Statement: `CREATE TABLE plt1_p2 PARTITION OF plt1 FOR VALUES IN ('0001', '0005', '0002', '0009');`,
			},
			{
				Statement: `CREATE TABLE plt1_p3 PARTITION OF plt1 FOR VALUES IN ('0006', '0007', '0008', '0011');`,
			},
			{
				Statement: `INSERT INTO plt1 SELECT i, i, to_char(i/50, 'FM0000') FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `ANALYZE plt1;`,
			},
			{
				Statement: `CREATE TABLE plt2 (a int, b int, c text) PARTITION BY LIST(c);`,
			},
			{
				Statement: `CREATE TABLE plt2_p1 PARTITION OF plt2 FOR VALUES IN ('0000', '0003', '0004', '0010');`,
			},
			{
				Statement: `CREATE TABLE plt2_p2 PARTITION OF plt2 FOR VALUES IN ('0001', '0005', '0002', '0009');`,
			},
			{
				Statement: `CREATE TABLE plt2_p3 PARTITION OF plt2 FOR VALUES IN ('0006', '0007', '0008', '0011');`,
			},
			{
				Statement: `INSERT INTO plt2 SELECT i, i, to_char(i/50, 'FM0000') FROM generate_series(0, 599, 3) i;`,
			},
			{
				Statement: `ANALYZE plt2;`,
			},
			{
				Statement: `CREATE TABLE plt1_e (a int, b int, c text) PARTITION BY LIST(ltrim(c, 'A'));`,
			},
			{
				Statement: `CREATE TABLE plt1_e_p1 PARTITION OF plt1_e FOR VALUES IN ('0000', '0003', '0004', '0010');`,
			},
			{
				Statement: `CREATE TABLE plt1_e_p2 PARTITION OF plt1_e FOR VALUES IN ('0001', '0005', '0002', '0009');`,
			},
			{
				Statement: `CREATE TABLE plt1_e_p3 PARTITION OF plt1_e FOR VALUES IN ('0006', '0007', '0008', '0011');`,
			},
			{
				Statement: `INSERT INTO plt1_e SELECT i, i, 'A' || to_char(i/50, 'FM0000') FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `ANALYZE plt1_e;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT avg(t1.a), avg(t2.b), avg(t3.a + t3.b), t1.c, t2.c, t3.c FROM plt1 t1, plt2 t2, plt1_e t3 WHERE t1.b = t2.b AND t1.c = t2.c AND ltrim(t3.c, 'A') = t1.c GROUP BY t1.c, t2.c, t3.c ORDER BY t1.c, t2.c, t3.c;`,
				Results: []sql.Row{{`GroupAggregate`}, {`Group Key: t1.c, t2.c, t3.c`}, {`->  Sort`}, {`Sort Key: t1.c, t3.c`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t1_1.c = ltrim(t3_1.c, 'A'::text))`}, {`->  Hash Join`}, {`Hash Cond: ((t1_1.b = t2_1.b) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on plt1_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on plt2_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_e_p1 t3_1`}, {`->  Hash Join`}, {`Hash Cond: (t1_2.c = ltrim(t3_2.c, 'A'::text))`}, {`->  Hash Join`}, {`Hash Cond: ((t1_2.b = t2_2.b) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on plt1_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on plt2_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_e_p2 t3_2`}, {`->  Hash Join`}, {`Hash Cond: (t1_3.c = ltrim(t3_3.c, 'A'::text))`}, {`->  Hash Join`}, {`Hash Cond: ((t1_3.b = t2_3.b) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on plt1_p3 t1_3`}, {`->  Hash`}, {`->  Seq Scan on plt2_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_e_p3 t3_3`}},
			},
			{
				Statement: `SELECT avg(t1.a), avg(t2.b), avg(t3.a + t3.b), t1.c, t2.c, t3.c FROM plt1 t1, plt2 t2, plt1_e t3 WHERE t1.b = t2.b AND t1.c = t2.c AND ltrim(t3.c, 'A') = t1.c GROUP BY t1.c, t2.c, t3.c ORDER BY t1.c, t2.c, t3.c;`,
				Results:   []sql.Row{{24.0000000000000000, 24.0000000000000000, 48.0000000000000000, "0000", "0000", `A0000`}, {75.0000000000000000, 75.0000000000000000, 148.0000000000000000, "0001", "0001", `A0001`}, {123.0000000000000000, 123.0000000000000000, 248.0000000000000000, "0002", "0002", `A0002`}, {174.0000000000000000, 174.0000000000000000, 348.0000000000000000, "0003", "0003", `A0003`}, {225.0000000000000000, 225.0000000000000000, 448.0000000000000000, "0004", "0004", `A0004`}, {273.0000000000000000, 273.0000000000000000, 548.0000000000000000, "0005", "0005", `A0005`}, {324.0000000000000000, 324.0000000000000000, 648.0000000000000000, "0006", "0006", `A0006`}, {375.0000000000000000, 375.0000000000000000, 748.0000000000000000, "0007", "0007", `A0007`}, {423.0000000000000000, 423.0000000000000000, 848.0000000000000000, "0008", "0008", `A0008`}, {474.0000000000000000, 474.0000000000000000, 948.0000000000000000, "0009", "0009", `A0009`}, {525.0000000000000000, 525.0000000000000000, 1048.0000000000000000, "0010", "0010", `A0010`}, {573.0000000000000000, 573.0000000000000000, 1148.0000000000000000, "0011", "0011", `A0011`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1, prt2 t2 WHERE t1.a = t2.b AND t1.a = 1 AND t1.a = 2;`,
				Results: []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1 WHERE a = 1 AND a = 2) t1 LEFT JOIN prt2 t2 ON t1.a = t2.b;`,
				Results: []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1 WHERE a = 1 AND a = 2) t1 RIGHT JOIN prt2 t2 ON t1.a = t2.b, prt1 t3 WHERE t2.b = t3.a;`,
				Results: []sql.Row{{`Hash Left Join`}, {`Hash Cond: (t2.b = a)`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t3_1.a = t2_1.b)`}, {`->  Seq Scan on prt1_p1 t3_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Hash Join`}, {`Hash Cond: (t3_2.a = t2_2.b)`}, {`->  Seq Scan on prt1_p2 t3_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Hash Join`}, {`Hash Cond: (t3_3.a = t2_3.b)`}, {`->  Seq Scan on prt1_p3 t3_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`->  Hash`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1 WHERE a = 1 AND a = 2) t1 FULL JOIN prt2 t2 ON t1.a = t2.b WHERE t2.a = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: a, t2.b`}, {`->  Hash Left Join`}, {`Hash Cond: (t2.b = a)`}, {`->  Append`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`Filter: (a = 0)`}, {`->  Hash`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `CREATE TABLE pht1 (a int, b int, c text) PARTITION BY HASH(c);`,
			},
			{
				Statement: `CREATE TABLE pht1_p1 PARTITION OF pht1 FOR VALUES WITH (MODULUS 3, REMAINDER 0);`,
			},
			{
				Statement: `CREATE TABLE pht1_p2 PARTITION OF pht1 FOR VALUES WITH (MODULUS 3, REMAINDER 1);`,
			},
			{
				Statement: `CREATE TABLE pht1_p3 PARTITION OF pht1 FOR VALUES WITH (MODULUS 3, REMAINDER 2);`,
			},
			{
				Statement: `INSERT INTO pht1 SELECT i, i, to_char(i/50, 'FM0000') FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `ANALYZE pht1;`,
			},
			{
				Statement: `CREATE TABLE pht2 (a int, b int, c text) PARTITION BY HASH(c);`,
			},
			{
				Statement: `CREATE TABLE pht2_p1 PARTITION OF pht2 FOR VALUES WITH (MODULUS 3, REMAINDER 0);`,
			},
			{
				Statement: `CREATE TABLE pht2_p2 PARTITION OF pht2 FOR VALUES WITH (MODULUS 3, REMAINDER 1);`,
			},
			{
				Statement: `CREATE TABLE pht2_p3 PARTITION OF pht2 FOR VALUES WITH (MODULUS 3, REMAINDER 2);`,
			},
			{
				Statement: `INSERT INTO pht2 SELECT i, i, to_char(i/50, 'FM0000') FROM generate_series(0, 599, 3) i;`,
			},
			{
				Statement: `ANALYZE pht2;`,
			},
			{
				Statement: `CREATE TABLE pht1_e (a int, b int, c text) PARTITION BY HASH(ltrim(c, 'A'));`,
			},
			{
				Statement: `CREATE TABLE pht1_e_p1 PARTITION OF pht1_e FOR VALUES WITH (MODULUS 3, REMAINDER 0);`,
			},
			{
				Statement: `CREATE TABLE pht1_e_p2 PARTITION OF pht1_e FOR VALUES WITH (MODULUS 3, REMAINDER 1);`,
			},
			{
				Statement: `CREATE TABLE pht1_e_p3 PARTITION OF pht1_e FOR VALUES WITH (MODULUS 3, REMAINDER 2);`,
			},
			{
				Statement: `INSERT INTO pht1_e SELECT i, i, 'A' || to_char(i/50, 'FM0000') FROM generate_series(0, 299, 2) i;`,
			},
			{
				Statement: `ANALYZE pht1_e;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT avg(t1.a), avg(t2.b), avg(t3.a + t3.b), t1.c, t2.c, t3.c FROM pht1 t1, pht2 t2, pht1_e t3 WHERE t1.b = t2.b AND t1.c = t2.c AND ltrim(t3.c, 'A') = t1.c GROUP BY t1.c, t2.c, t3.c ORDER BY t1.c, t2.c, t3.c;`,
				Results: []sql.Row{{`GroupAggregate`}, {`Group Key: t1.c, t2.c, t3.c`}, {`->  Sort`}, {`Sort Key: t1.c, t3.c`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t1_1.c = ltrim(t3_1.c, 'A'::text))`}, {`->  Hash Join`}, {`Hash Cond: ((t1_1.b = t2_1.b) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on pht1_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on pht2_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on pht1_e_p1 t3_1`}, {`->  Hash Join`}, {`Hash Cond: (t1_2.c = ltrim(t3_2.c, 'A'::text))`}, {`->  Hash Join`}, {`Hash Cond: ((t1_2.b = t2_2.b) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on pht1_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on pht2_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on pht1_e_p2 t3_2`}, {`->  Hash Join`}, {`Hash Cond: (t1_3.c = ltrim(t3_3.c, 'A'::text))`}, {`->  Hash Join`}, {`Hash Cond: ((t1_3.b = t2_3.b) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on pht1_p3 t1_3`}, {`->  Hash`}, {`->  Seq Scan on pht2_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on pht1_e_p3 t3_3`}},
			},
			{
				Statement: `SELECT avg(t1.a), avg(t2.b), avg(t3.a + t3.b), t1.c, t2.c, t3.c FROM pht1 t1, pht2 t2, pht1_e t3 WHERE t1.b = t2.b AND t1.c = t2.c AND ltrim(t3.c, 'A') = t1.c GROUP BY t1.c, t2.c, t3.c ORDER BY t1.c, t2.c, t3.c;`,
				Results:   []sql.Row{{24.0000000000000000, 24.0000000000000000, 48.0000000000000000, "0000", "0000", `A0000`}, {75.0000000000000000, 75.0000000000000000, 148.0000000000000000, "0001", "0001", `A0001`}, {123.0000000000000000, 123.0000000000000000, 248.0000000000000000, "0002", "0002", `A0002`}, {174.0000000000000000, 174.0000000000000000, 348.0000000000000000, "0003", "0003", `A0003`}, {225.0000000000000000, 225.0000000000000000, 448.0000000000000000, "0004", "0004", `A0004`}, {273.0000000000000000, 273.0000000000000000, 548.0000000000000000, "0005", "0005", `A0005`}},
			},
			{
				Statement: `ALTER TABLE prt1 DETACH PARTITION prt1_p3;`,
			},
			{
				Statement: `ALTER TABLE prt1 ATTACH PARTITION prt1_p3 DEFAULT;`,
			},
			{
				Statement: `ANALYZE prt1;`,
			},
			{
				Statement: `ALTER TABLE prt2 DETACH PARTITION prt2_p3;`,
			},
			{
				Statement: `ALTER TABLE prt2 ATTACH PARTITION prt2_p3 DEFAULT;`,
			},
			{
				Statement: `ANALYZE prt2;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1, prt2 t2 WHERE t1.a = t2.b AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Seq Scan on prt2_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `ALTER TABLE plt1 DETACH PARTITION plt1_p3;`,
			},
			{
				Statement: `ALTER TABLE plt1 ATTACH PARTITION plt1_p3 DEFAULT;`,
			},
			{
				Statement: `ANALYZE plt1;`,
			},
			{
				Statement: `ALTER TABLE plt2 DETACH PARTITION plt2_p3;`,
			},
			{
				Statement: `ALTER TABLE plt2 ATTACH PARTITION plt2_p3 DEFAULT;`,
			},
			{
				Statement: `ANALYZE plt2;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT avg(t1.a), avg(t2.b), t1.c, t2.c FROM plt1 t1 RIGHT JOIN plt2 t2 ON t1.c = t2.c WHERE t1.a % 25 = 0 GROUP BY t1.c, t2.c ORDER BY t1.c, t2.c;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.c`}, {`->  HashAggregate`}, {`Group Key: t1.c, t2.c`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.c = t1_1.c)`}, {`->  Seq Scan on plt2_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_p1 t1_1`}, {`Filter: ((a % 25) = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.c = t1_2.c)`}, {`->  Seq Scan on plt2_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_p2 t1_2`}, {`Filter: ((a % 25) = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_3.c = t1_3.c)`}, {`->  Seq Scan on plt2_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_p3 t1_3`}, {`Filter: ((a % 25) = 0)`}},
			},
			{
				Statement: `CREATE TABLE prt1_l (a int, b int, c varchar) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE prt1_l_p1 PARTITION OF prt1_l FOR VALUES FROM (0) TO (250);`,
			},
			{
				Statement: `CREATE TABLE prt1_l_p2 PARTITION OF prt1_l FOR VALUES FROM (250) TO (500) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE prt1_l_p2_p1 PARTITION OF prt1_l_p2 FOR VALUES IN ('0000', '0001');`,
			},
			{
				Statement: `CREATE TABLE prt1_l_p2_p2 PARTITION OF prt1_l_p2 FOR VALUES IN ('0002', '0003');`,
			},
			{
				Statement: `CREATE TABLE prt1_l_p3 PARTITION OF prt1_l FOR VALUES FROM (500) TO (600) PARTITION BY RANGE (b);`,
			},
			{
				Statement: `CREATE TABLE prt1_l_p3_p1 PARTITION OF prt1_l_p3 FOR VALUES FROM (0) TO (13);`,
			},
			{
				Statement: `CREATE TABLE prt1_l_p3_p2 PARTITION OF prt1_l_p3 FOR VALUES FROM (13) TO (25);`,
			},
			{
				Statement: `INSERT INTO prt1_l SELECT i, i % 25, to_char(i % 4, 'FM0000') FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `ANALYZE prt1_l;`,
			},
			{
				Statement: `CREATE TABLE prt2_l (a int, b int, c varchar) PARTITION BY RANGE(b);`,
			},
			{
				Statement: `CREATE TABLE prt2_l_p1 PARTITION OF prt2_l FOR VALUES FROM (0) TO (250);`,
			},
			{
				Statement: `CREATE TABLE prt2_l_p2 PARTITION OF prt2_l FOR VALUES FROM (250) TO (500) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE prt2_l_p2_p1 PARTITION OF prt2_l_p2 FOR VALUES IN ('0000', '0001');`,
			},
			{
				Statement: `CREATE TABLE prt2_l_p2_p2 PARTITION OF prt2_l_p2 FOR VALUES IN ('0002', '0003');`,
			},
			{
				Statement: `CREATE TABLE prt2_l_p3 PARTITION OF prt2_l FOR VALUES FROM (500) TO (600) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE prt2_l_p3_p1 PARTITION OF prt2_l_p3 FOR VALUES FROM (0) TO (13);`,
			},
			{
				Statement: `CREATE TABLE prt2_l_p3_p2 PARTITION OF prt2_l_p3 FOR VALUES FROM (13) TO (25);`,
			},
			{
				Statement: `INSERT INTO prt2_l SELECT i % 25, i, to_char(i % 4, 'FM0000') FROM generate_series(0, 599, 3) i;`,
			},
			{
				Statement: `ANALYZE prt2_l;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_l t1, prt2_l t2 WHERE t1.a = t2.b AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_l_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_l_p2_p1 t2_3`}, {`->  Seq Scan on prt2_l_p2_p2 t2_4`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_l_p2_p1 t1_3`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_l_p2_p2 t1_4`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_6.b = t1_5.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_l_p3_p1 t2_6`}, {`->  Seq Scan on prt2_l_p3_p2 t2_7`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p3_p1 t1_5`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_l t1, prt2_l t2 WHERE t1.a = t2.b AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000"}, {150, "0002", 150, "0002"}, {300, "0000", 300, "0000"}, {450, "0002", 450, "0002"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_l t1 LEFT JOIN prt2_l t2 ON t1.a = t2.b AND t1.c = t2.c WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_1.b = t1_1.a) AND ((t2_1.c)::text = (t1_1.c)::text))`}, {`->  Seq Scan on prt2_l_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_2.b = t1_2.a) AND ((t2_2.c)::text = (t1_2.c)::text))`}, {`->  Seq Scan on prt2_l_p2_p1 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p2_p1 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_3.b = t1_3.a) AND ((t2_3.c)::text = (t1_3.c)::text))`}, {`->  Seq Scan on prt2_l_p2_p2 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p2_p2 t1_3`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_5.b = t1_4.a) AND ((t2_5.c)::text = (t1_4.c)::text))`}, {`->  Append`}, {`->  Seq Scan on prt2_l_p3_p1 t2_5`}, {`->  Seq Scan on prt2_l_p3_p2 t2_6`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p3_p1 t1_4`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_l t1 LEFT JOIN prt2_l t2 ON t1.a = t2.b AND t1.c = t2.c WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000"}, {50, "0002", ``, ``}, {100, "0000", ``, ``}, {150, "0002", 150, "0002"}, {200, "0000", ``, ``}, {250, "0002", ``, ``}, {300, "0000", 300, "0000"}, {350, "0002", ``, ``}, {400, "0000", ``, ``}, {450, "0002", 450, "0002"}, {500, "0000", ``, ``}, {550, "0002", ``, ``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_l t1 RIGHT JOIN prt2_l t2 ON t1.a = t2.b AND t1.c = t2.c WHERE t2.a = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: ((t1_1.a = t2_1.b) AND ((t1_1.c)::text = (t2_1.c)::text))`}, {`->  Seq Scan on prt1_l_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_l_p1 t2_1`}, {`Filter: (a = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t1_2.a = t2_2.b) AND ((t1_2.c)::text = (t2_2.c)::text))`}, {`->  Seq Scan on prt1_l_p2_p1 t1_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_l_p2_p1 t2_2`}, {`Filter: (a = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t1_3.a = t2_3.b) AND ((t1_3.c)::text = (t2_3.c)::text))`}, {`->  Seq Scan on prt1_l_p2_p2 t1_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_l_p2_p2 t2_3`}, {`Filter: (a = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t1_5.a = t2_4.b) AND ((t1_5.c)::text = (t2_4.c)::text))`}, {`->  Append`}, {`->  Seq Scan on prt1_l_p3_p1 t1_5`}, {`->  Seq Scan on prt1_l_p3_p2 t1_6`}, {`->  Hash`}, {`->  Seq Scan on prt2_l_p3_p1 t2_4`}, {`Filter: (a = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_l t1 RIGHT JOIN prt2_l t2 ON t1.a = t2.b AND t1.c = t2.c WHERE t2.a = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000"}, {150, "0002", 150, "0002"}, {300, "0000", 300, "0000"}, {450, "0002", 450, "0002"}, {``, ``, 75, "0003"}, {``, ``, 225, "0001"}, {``, ``, 375, "0003"}, {``, ``, 525, "0001"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1_l WHERE prt1_l.b = 0) t1 FULL JOIN (SELECT * FROM prt2_l WHERE prt2_l.a = 0) t2 ON (t1.a = t2.b AND t1.c = t2.c) ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1_l.a, prt2_l.b`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_l_1.a = prt2_l_1.b) AND ((prt1_l_1.c)::text = (prt2_l_1.c)::text))`}, {`->  Seq Scan on prt1_l_p1 prt1_l_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_l_p1 prt2_l_1`}, {`Filter: (a = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_l_2.a = prt2_l_2.b) AND ((prt1_l_2.c)::text = (prt2_l_2.c)::text))`}, {`->  Seq Scan on prt1_l_p2_p1 prt1_l_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_l_p2_p1 prt2_l_2`}, {`Filter: (a = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_l_3.a = prt2_l_3.b) AND ((prt1_l_3.c)::text = (prt2_l_3.c)::text))`}, {`->  Seq Scan on prt1_l_p2_p2 prt1_l_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_l_p2_p2 prt2_l_3`}, {`Filter: (a = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: ((prt1_l_4.a = prt2_l_4.b) AND ((prt1_l_4.c)::text = (prt2_l_4.c)::text))`}, {`->  Seq Scan on prt1_l_p3_p1 prt1_l_4`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_l_p3_p1 prt2_l_4`}, {`Filter: (a = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1_l WHERE prt1_l.b = 0) t1 FULL JOIN (SELECT * FROM prt2_l WHERE prt2_l.a = 0) t2 ON (t1.a = t2.b AND t1.c = t2.c) ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{0, "0000", 0, "0000"}, {50, "0002", ``, ``}, {100, "0000", ``, ``}, {150, "0002", 150, "0002"}, {200, "0000", ``, ``}, {250, "0002", ``, ``}, {300, "0000", 300, "0000"}, {350, "0002", ``, ``}, {400, "0000", ``, ``}, {450, "0002", 450, "0002"}, {500, "0000", ``, ``}, {550, "0002", ``, ``}, {``, ``, 75, "0003"}, {``, ``, 225, "0001"}, {``, ``, 375, "0003"}, {``, ``, 525, "0001"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM prt1_l t1 LEFT JOIN LATERAL
			  (SELECT t2.a AS t2a, t2.c AS t2c, t2.b AS t2b, t3.b AS t3b, least(t1.a,t2.a,t3.b) FROM prt1_l t2 JOIN prt2_l t3 ON (t2.a = t3.b AND t2.c = t3.c)) ss
			  ON t1.a = ss.t2a AND t1.c = ss.t2c WHERE t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop Left Join`}, {`->  Seq Scan on prt1_l_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: ((t3_1.b = t2_1.a) AND ((t3_1.c)::text = (t2_1.c)::text))`}, {`->  Seq Scan on prt2_l_p1 t3_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p1 t2_1`}, {`Filter: ((t1_1.a = a) AND ((t1_1.c)::text = (c)::text))`}, {`->  Nested Loop Left Join`}, {`->  Seq Scan on prt1_l_p2_p1 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: ((t3_2.b = t2_2.a) AND ((t3_2.c)::text = (t2_2.c)::text))`}, {`->  Seq Scan on prt2_l_p2_p1 t3_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p2_p1 t2_2`}, {`Filter: ((t1_2.a = a) AND ((t1_2.c)::text = (c)::text))`}, {`->  Nested Loop Left Join`}, {`->  Seq Scan on prt1_l_p2_p2 t1_3`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: ((t3_3.b = t2_3.a) AND ((t3_3.c)::text = (t2_3.c)::text))`}, {`->  Seq Scan on prt2_l_p2_p2 t3_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_l_p2_p2 t2_3`}, {`Filter: ((t1_3.a = a) AND ((t1_3.c)::text = (c)::text))`}, {`->  Nested Loop Left Join`}, {`->  Seq Scan on prt1_l_p3_p1 t1_4`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: ((t3_5.b = t2_5.a) AND ((t3_5.c)::text = (t2_5.c)::text))`}, {`->  Append`}, {`->  Seq Scan on prt2_l_p3_p1 t3_5`}, {`->  Seq Scan on prt2_l_p3_p2 t3_6`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_l_p3_p1 t2_5`}, {`Filter: ((t1_4.a = a) AND ((t1_4.c)::text = (c)::text))`}, {`->  Seq Scan on prt1_l_p3_p2 t2_6`}, {`Filter: ((t1_4.a = a) AND ((t1_4.c)::text = (c)::text))`}},
			},
			{
				Statement: `SELECT * FROM prt1_l t1 LEFT JOIN LATERAL
			  (SELECT t2.a AS t2a, t2.c AS t2c, t2.b AS t2b, t3.b AS t3b, least(t1.a,t2.a,t3.b) FROM prt1_l t2 JOIN prt2_l t3 ON (t2.a = t3.b AND t2.c = t3.c)) ss
			  ON t1.a = ss.t2a AND t1.c = ss.t2c WHERE t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{0, 0, "0000", 0, "0000", 0, 0, 0}, {50, 0, "0002", ``, ``, ``, ``, ``}, {100, 0, "0000", ``, ``, ``, ``, ``}, {150, 0, "0002", 150, "0002", 0, 150, 150}, {200, 0, "0000", ``, ``, ``, ``, ``}, {250, 0, "0002", ``, ``, ``, ``, ``}, {300, 0, "0000", 300, "0000", 0, 300, 300}, {350, 0, "0002", ``, ``, ``, ``, ``}, {400, 0, "0000", ``, ``, ``, ``, ``}, {450, 0, "0002", 450, "0002", 0, 450, 450}, {500, 0, "0000", ``, ``, ``, ``, ``}, {550, 0, "0002", ``, ``, ``, ``, ``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT * FROM prt1_l WHERE a = 1 AND a = 2) t1 RIGHT JOIN prt2_l t2 ON t1.a = t2.b AND t1.b = t2.a AND t1.c = t2.c;`,
				Results: []sql.Row{{`Hash Left Join`}, {`Hash Cond: ((t2.b = a) AND (t2.a = b) AND ((t2.c)::text = (c)::text))`}, {`->  Append`}, {`->  Seq Scan on prt2_l_p1 t2_1`}, {`->  Seq Scan on prt2_l_p2_p1 t2_2`}, {`->  Seq Scan on prt2_l_p2_p2 t2_3`}, {`->  Seq Scan on prt2_l_p3_p1 t2_4`}, {`->  Seq Scan on prt2_l_p3_p2 t2_5`}, {`->  Hash`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
DELETE FROM prt1_l
WHERE EXISTS (
  SELECT 1
    FROM int4_tbl,
         LATERAL (SELECT int4_tbl.f1 FROM int8_tbl LIMIT 2) ss
    WHERE prt1_l.c IS NULL);`,
				Results: []sql.Row{{`Delete on prt1_l`}, {`Delete on prt1_l_p1 prt1_l_1`}, {`Delete on prt1_l_p3_p1 prt1_l_2`}, {`Delete on prt1_l_p3_p2 prt1_l_3`}, {`->  Nested Loop Semi Join`}, {`->  Append`}, {`->  Seq Scan on prt1_l_p1 prt1_l_1`}, {`Filter: (c IS NULL)`}, {`->  Seq Scan on prt1_l_p3_p1 prt1_l_2`}, {`Filter: (c IS NULL)`}, {`->  Seq Scan on prt1_l_p3_p2 prt1_l_3`}, {`Filter: (c IS NULL)`}, {`->  Materialize`}, {`->  Nested Loop`}, {`->  Seq Scan on int4_tbl`}, {`->  Subquery Scan on ss`}, {`->  Limit`}, {`->  Seq Scan on int8_tbl`}},
			},
			{
				Statement: `CREATE TABLE prt1_n (a int, b int, c varchar) PARTITION BY RANGE(c);`,
			},
			{
				Statement: `CREATE TABLE prt1_n_p1 PARTITION OF prt1_n FOR VALUES FROM ('0000') TO ('0250');`,
			},
			{
				Statement: `CREATE TABLE prt1_n_p2 PARTITION OF prt1_n FOR VALUES FROM ('0250') TO ('0500');`,
			},
			{
				Statement: `INSERT INTO prt1_n SELECT i, i, to_char(i, 'FM0000') FROM generate_series(0, 499, 2) i;`,
			},
			{
				Statement: `ANALYZE prt1_n;`,
			},
			{
				Statement: `CREATE TABLE prt2_n (a int, b int, c text) PARTITION BY LIST(c);`,
			},
			{
				Statement: `CREATE TABLE prt2_n_p1 PARTITION OF prt2_n FOR VALUES IN ('0000', '0003', '0004', '0010', '0006', '0007');`,
			},
			{
				Statement: `CREATE TABLE prt2_n_p2 PARTITION OF prt2_n FOR VALUES IN ('0001', '0005', '0002', '0009', '0008', '0011');`,
			},
			{
				Statement: `INSERT INTO prt2_n SELECT i, i, to_char(i/50, 'FM0000') FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `ANALYZE prt2_n;`,
			},
			{
				Statement: `CREATE TABLE prt3_n (a int, b int, c text) PARTITION BY LIST(c);`,
			},
			{
				Statement: `CREATE TABLE prt3_n_p1 PARTITION OF prt3_n FOR VALUES IN ('0000', '0004', '0006', '0007');`,
			},
			{
				Statement: `CREATE TABLE prt3_n_p2 PARTITION OF prt3_n FOR VALUES IN ('0001', '0002', '0008', '0010');`,
			},
			{
				Statement: `CREATE TABLE prt3_n_p3 PARTITION OF prt3_n FOR VALUES IN ('0003', '0005', '0009', '0011');`,
			},
			{
				Statement: `INSERT INTO prt2_n SELECT i, i, to_char(i/50, 'FM0000') FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `ANALYZE prt3_n;`,
			},
			{
				Statement: `CREATE TABLE prt4_n (a int, b int, c text) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE prt4_n_p1 PARTITION OF prt4_n FOR VALUES FROM (0) TO (300);`,
			},
			{
				Statement: `CREATE TABLE prt4_n_p2 PARTITION OF prt4_n FOR VALUES FROM (300) TO (500);`,
			},
			{
				Statement: `CREATE TABLE prt4_n_p3 PARTITION OF prt4_n FOR VALUES FROM (500) TO (600);`,
			},
			{
				Statement: `INSERT INTO prt4_n SELECT i, i, to_char(i, 'FM0000') FROM generate_series(0, 599, 2) i;`,
			},
			{
				Statement: `ANALYZE prt4_n;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1, prt4_n t2 WHERE t1.a = t2.a;`,
				Results: []sql.Row{{`Hash Join`}, {`Hash Cond: (t1.a = t2.a)`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt4_n_p1 t2_1`}, {`->  Seq Scan on prt4_n_p2 t2_2`}, {`->  Seq Scan on prt4_n_p3 t2_3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1, prt4_n t2, prt2 t3 WHERE t1.a = t2.a and t1.a = t3.b;`,
				Results: []sql.Row{{`Hash Join`}, {`Hash Cond: (t2.a = t1.a)`}, {`->  Append`}, {`->  Seq Scan on prt4_n_p1 t2_1`}, {`->  Seq Scan on prt4_n_p2 t2_2`}, {`->  Seq Scan on prt4_n_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t1_1.a = t3_1.b)`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on prt2_p1 t3_1`}, {`->  Hash Join`}, {`Hash Cond: (t1_2.a = t3_2.b)`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_p2 t3_2`}, {`->  Hash Join`}, {`Hash Cond: (t1_3.a = t3_3.b)`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_p3 t3_3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1 t1 LEFT JOIN prt2 t2 ON (t1.a < t2.b);`,
				Results: []sql.Row{{`Nested Loop Left Join`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 t1_1`}, {`->  Seq Scan on prt1_p2 t1_2`}, {`->  Seq Scan on prt1_p3 t1_3`}, {`->  Append`}, {`->  Index Scan using iprt2_p1_b on prt2_p1 t2_1`}, {`Index Cond: (b > t1.a)`}, {`->  Index Scan using iprt2_p2_b on prt2_p2 t2_2`}, {`Index Cond: (b > t1.a)`}, {`->  Index Scan using iprt2_p3_b on prt2_p3 t2_3`}, {`Index Cond: (b > t1.a)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_m t1, prt2_m t2 WHERE t1.a = (t2.b + t2.a)/2;`,
				Results: []sql.Row{{`Hash Join`}, {`Hash Cond: (((t2.b + t2.a) / 2) = t1.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_m_p1 t2_1`}, {`->  Seq Scan on prt2_m_p2 t2_2`}, {`->  Seq Scan on prt2_m_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_m_p1 t1_1`}, {`->  Seq Scan on prt1_m_p2 t1_2`}, {`->  Seq Scan on prt1_m_p3 t1_3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_m t1 LEFT JOIN prt2_m t2 ON t1.a = t2.b;`,
				Results: []sql.Row{{`Hash Left Join`}, {`Hash Cond: (t1.a = t2.b)`}, {`->  Append`}, {`->  Seq Scan on prt1_m_p1 t1_1`}, {`->  Seq Scan on prt1_m_p2 t1_2`}, {`->  Seq Scan on prt1_m_p3 t1_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt2_m_p1 t2_1`}, {`->  Seq Scan on prt2_m_p2 t2_2`}, {`->  Seq Scan on prt2_m_p3 t2_3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_m t1 LEFT JOIN prt2_m t2 ON t1.c = t2.c;`,
				Results: []sql.Row{{`Hash Left Join`}, {`Hash Cond: (t1.c = t2.c)`}, {`->  Append`}, {`->  Seq Scan on prt1_m_p1 t1_1`}, {`->  Seq Scan on prt1_m_p2 t1_2`}, {`->  Seq Scan on prt1_m_p3 t1_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt2_m_p1 t2_1`}, {`->  Seq Scan on prt2_m_p2 t2_2`}, {`->  Seq Scan on prt2_m_p3 t2_3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_n t1 LEFT JOIN prt2_n t2 ON (t1.c = t2.c);`,
				Results: []sql.Row{{`Hash Right Join`}, {`Hash Cond: (t2.c = (t1.c)::text)`}, {`->  Append`}, {`->  Seq Scan on prt2_n_p1 t2_1`}, {`->  Seq Scan on prt2_n_p2 t2_2`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_n_p1 t1_1`}, {`->  Seq Scan on prt1_n_p2 t1_2`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_n t1 JOIN prt2_n t2 ON (t1.c = t2.c) JOIN plt1 t3 ON (t1.c = t3.c);`,
				Results: []sql.Row{{`Hash Join`}, {`Hash Cond: (t2.c = (t1.c)::text)`}, {`->  Append`}, {`->  Seq Scan on prt2_n_p1 t2_1`}, {`->  Seq Scan on prt2_n_p2 t2_2`}, {`->  Hash`}, {`->  Hash Join`}, {`Hash Cond: (t3.c = (t1.c)::text)`}, {`->  Append`}, {`->  Seq Scan on plt1_p1 t3_1`}, {`->  Seq Scan on plt1_p2 t3_2`}, {`->  Seq Scan on plt1_p3 t3_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_n_p1 t1_1`}, {`->  Seq Scan on prt1_n_p2 t1_2`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_n t1 FULL JOIN prt1 t2 ON (t1.c = t2.c);`,
				Results: []sql.Row{{`Hash Full Join`}, {`Hash Cond: ((t2.c)::text = (t1.c)::text)`}, {`->  Append`}, {`->  Seq Scan on prt1_p1 t2_1`}, {`->  Seq Scan on prt1_p2 t2_2`}, {`->  Seq Scan on prt1_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_n_p1 t1_1`}, {`->  Seq Scan on prt1_n_p2 t1_2`}},
			},
			{
				Statement: `create temp table prtx1 (a integer, b integer, c integer)
  partition by range (a);`,
			},
			{
				Statement: `create temp table prtx1_1 partition of prtx1 for values from (1) to (11);`,
			},
			{
				Statement: `create temp table prtx1_2 partition of prtx1 for values from (11) to (21);`,
			},
			{
				Statement: `create temp table prtx1_3 partition of prtx1 for values from (21) to (31);`,
			},
			{
				Statement: `create temp table prtx2 (a integer, b integer, c integer)
  partition by range (a);`,
			},
			{
				Statement: `create temp table prtx2_1 partition of prtx2 for values from (1) to (11);`,
			},
			{
				Statement: `create temp table prtx2_2 partition of prtx2 for values from (11) to (21);`,
			},
			{
				Statement: `create temp table prtx2_3 partition of prtx2 for values from (21) to (31);`,
			},
			{
				Statement: `insert into prtx1 select 1 + i%30, i, i
  from generate_series(1,1000) i;`,
			},
			{
				Statement: `insert into prtx2 select 1 + i%30, i, i
  from generate_series(1,500) i, generate_series(1,10) j;`,
			},
			{
				Statement: `create index on prtx2 (b);`,
			},
			{
				Statement: `create index on prtx2 (c);`,
			},
			{
				Statement: `analyze prtx1;`,
			},
			{
				Statement: `analyze prtx2;`,
			},
			{
				Statement: `explain (costs off)
select * from prtx1
where not exists (select 1 from prtx2
                  where prtx2.a=prtx1.a and prtx2.b=prtx1.b and prtx2.c=123)
  and a<20 and c=120;`,
				Results: []sql.Row{{`Append`}, {`->  Nested Loop Anti Join`}, {`->  Seq Scan on prtx1_1`}, {`Filter: ((a < 20) AND (c = 120))`}, {`->  Bitmap Heap Scan on prtx2_1`}, {`Recheck Cond: ((b = prtx1_1.b) AND (c = 123))`}, {`Filter: (a = prtx1_1.a)`}, {`->  BitmapAnd`}, {`->  Bitmap Index Scan on prtx2_1_b_idx`}, {`Index Cond: (b = prtx1_1.b)`}, {`->  Bitmap Index Scan on prtx2_1_c_idx`}, {`Index Cond: (c = 123)`}, {`->  Nested Loop Anti Join`}, {`->  Seq Scan on prtx1_2`}, {`Filter: ((a < 20) AND (c = 120))`}, {`->  Bitmap Heap Scan on prtx2_2`}, {`Recheck Cond: ((b = prtx1_2.b) AND (c = 123))`}, {`Filter: (a = prtx1_2.a)`}, {`->  BitmapAnd`}, {`->  Bitmap Index Scan on prtx2_2_b_idx`}, {`Index Cond: (b = prtx1_2.b)`}, {`->  Bitmap Index Scan on prtx2_2_c_idx`}, {`Index Cond: (c = 123)`}},
			},
			{
				Statement: `select * from prtx1
where not exists (select 1 from prtx2
                  where prtx2.a=prtx1.a and prtx2.b=prtx1.b and prtx2.c=123)
  and a<20 and c=120;`,
				Results: []sql.Row{{1, 120, 120}},
			},
			{
				Statement: `explain (costs off)
select * from prtx1
where not exists (select 1 from prtx2
                  where prtx2.a=prtx1.a and (prtx2.b=prtx1.b+1 or prtx2.c=99))
  and a<20 and c=91;`,
				Results: []sql.Row{{`Append`}, {`->  Nested Loop Anti Join`}, {`->  Seq Scan on prtx1_1`}, {`Filter: ((a < 20) AND (c = 91))`}, {`->  Bitmap Heap Scan on prtx2_1`}, {`Recheck Cond: ((b = (prtx1_1.b + 1)) OR (c = 99))`}, {`Filter: (a = prtx1_1.a)`}, {`->  BitmapOr`}, {`->  Bitmap Index Scan on prtx2_1_b_idx`}, {`Index Cond: (b = (prtx1_1.b + 1))`}, {`->  Bitmap Index Scan on prtx2_1_c_idx`}, {`Index Cond: (c = 99)`}, {`->  Nested Loop Anti Join`}, {`->  Seq Scan on prtx1_2`}, {`Filter: ((a < 20) AND (c = 91))`}, {`->  Bitmap Heap Scan on prtx2_2`}, {`Recheck Cond: ((b = (prtx1_2.b + 1)) OR (c = 99))`}, {`Filter: (a = prtx1_2.a)`}, {`->  BitmapOr`}, {`->  Bitmap Index Scan on prtx2_2_b_idx`}, {`Index Cond: (b = (prtx1_2.b + 1))`}, {`->  Bitmap Index Scan on prtx2_2_c_idx`}, {`Index Cond: (c = 99)`}},
			},
			{
				Statement: `select * from prtx1
where not exists (select 1 from prtx2
                  where prtx2.a=prtx1.a and (prtx2.b=prtx1.b+1 or prtx2.c=99))
  and a<20 and c=91;`,
				Results: []sql.Row{{2, 91, 91}},
			},
			{
				Statement: `CREATE TABLE prt1_adv (a int, b int, c varchar) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE prt1_adv_p1 PARTITION OF prt1_adv FOR VALUES FROM (100) TO (200);`,
			},
			{
				Statement: `CREATE TABLE prt1_adv_p2 PARTITION OF prt1_adv FOR VALUES FROM (200) TO (300);`,
			},
			{
				Statement: `CREATE TABLE prt1_adv_p3 PARTITION OF prt1_adv FOR VALUES FROM (300) TO (400);`,
			},
			{
				Statement: `CREATE INDEX prt1_adv_a_idx ON prt1_adv (a);`,
			},
			{
				Statement: `INSERT INTO prt1_adv SELECT i, i % 25, to_char(i, 'FM0000') FROM generate_series(100, 399) i;`,
			},
			{
				Statement: `ANALYZE prt1_adv;`,
			},
			{
				Statement: `CREATE TABLE prt2_adv (a int, b int, c varchar) PARTITION BY RANGE (b);`,
			},
			{
				Statement: `CREATE TABLE prt2_adv_p1 PARTITION OF prt2_adv FOR VALUES FROM (100) TO (150);`,
			},
			{
				Statement: `CREATE TABLE prt2_adv_p2 PARTITION OF prt2_adv FOR VALUES FROM (200) TO (300);`,
			},
			{
				Statement: `CREATE TABLE prt2_adv_p3 PARTITION OF prt2_adv FOR VALUES FROM (350) TO (500);`,
			},
			{
				Statement: `CREATE INDEX prt2_adv_b_idx ON prt2_adv (b);`,
			},
			{
				Statement: `INSERT INTO prt2_adv_p1 SELECT i % 25, i, to_char(i, 'FM0000') FROM generate_series(100, 149) i;`,
			},
			{
				Statement: `INSERT INTO prt2_adv_p2 SELECT i % 25, i, to_char(i, 'FM0000') FROM generate_series(200, 299) i;`,
			},
			{
				Statement: `INSERT INTO prt2_adv_p3 SELECT i % 25, i, to_char(i, 'FM0000') FROM generate_series(350, 499) i;`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{100, "0100", 100, "0100"}, {125, "0125", 125, "0125"}, {200, "0200", 200, "0200"}, {225, "0225", 225, "0225"}, {250, "0250", 250, "0250"}, {275, "0275", 275, "0275"}, {350, "0350", 350, "0350"}, {375, "0375", 375, "0375"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1_adv t1 WHERE EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_1.a = t2_1.b)`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_2.a = t2_2.b)`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_3.a = t2_3.b)`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM prt1_adv t1 WHERE EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results:   []sql.Row{{100, 0, "0100"}, {125, 0, "0125"}, {200, 0, "0200"}, {225, 0, "0225"}, {250, 0, "0250"}, {275, 0, "0275"}, {350, 0, "0350"}, {375, 0, "0375"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 LEFT JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 LEFT JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{100, "0100", 100, "0100"}, {125, "0125", 125, "0125"}, {150, "0150", ``, ``}, {175, 0175, ``, ``}, {200, "0200", 200, "0200"}, {225, "0225", 225, "0225"}, {250, "0250", 250, "0250"}, {275, "0275", 275, "0275"}, {300, "0300", ``, ``}, {325, "0325", ``, ``}, {350, "0350", 350, "0350"}, {375, "0375", 375, "0375"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_1.a = t2_1.b)`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_2.a = t2_2.b)`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_3.a = t2_3.b)`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM prt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results:   []sql.Row{{150, 0, "0150"}, {175, 0, 0175}, {300, 0, "0300"}, {325, 0, "0325"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT 175 phv, * FROM prt1_adv WHERE prt1_adv.b = 0) t1 FULL JOIN (SELECT 425 phv, * FROM prt2_adv WHERE prt2_adv.a = 0) t2 ON (t1.a = t2.b) WHERE t1.phv = t1.a OR t2.phv = t2.b ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1_adv.a, prt2_adv.b`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_adv_1.a = prt2_adv_1.b)`}, {`Filter: (((175) = prt1_adv_1.a) OR ((425) = prt2_adv_1.b))`}, {`->  Seq Scan on prt1_adv_p1 prt1_adv_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p1 prt2_adv_1`}, {`Filter: (a = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: (prt1_adv_2.a = prt2_adv_2.b)`}, {`Filter: (((175) = prt1_adv_2.a) OR ((425) = prt2_adv_2.b))`}, {`->  Seq Scan on prt1_adv_p2 prt1_adv_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p2 prt2_adv_2`}, {`Filter: (a = 0)`}, {`->  Hash Full Join`}, {`Hash Cond: (prt2_adv_3.b = prt1_adv_3.a)`}, {`Filter: (((175) = prt1_adv_3.a) OR ((425) = prt2_adv_3.b))`}, {`->  Seq Scan on prt2_adv_p3 prt2_adv_3`}, {`Filter: (a = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p3 prt1_adv_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT 175 phv, * FROM prt1_adv WHERE prt1_adv.b = 0) t1 FULL JOIN (SELECT 425 phv, * FROM prt2_adv WHERE prt2_adv.a = 0) t2 ON (t1.a = t2.b) WHERE t1.phv = t1.a OR t2.phv = t2.b ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{175, 0175, ``, ``}, {``, ``, 425, "0425"}},
			},
			{
				Statement: `CREATE TABLE prt2_adv_extra PARTITION OF prt2_adv FOR VALUES FROM (500) TO (MAXVALUE);`,
			},
			{
				Statement: `INSERT INTO prt2_adv SELECT i % 25, i, to_char(i, 'FM0000') FROM generate_series(500, 599) i;`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{100, "0100", 100, "0100"}, {125, "0125", 125, "0125"}, {200, "0200", 200, "0200"}, {225, "0225", 225, "0225"}, {250, "0250", 250, "0250"}, {275, "0275", 275, "0275"}, {350, "0350", 350, "0350"}, {375, "0375", 375, "0375"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1_adv t1 WHERE EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_1.a = t2_1.b)`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_2.a = t2_2.b)`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1_3.a = t2_3.b)`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM prt1_adv t1 WHERE EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results:   []sql.Row{{100, 0, "0100"}, {125, 0, "0125"}, {200, 0, "0200"}, {225, 0, "0225"}, {250, 0, "0250"}, {275, 0, "0275"}, {350, 0, "0350"}, {375, 0, "0375"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 LEFT JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_3.b = t1_3.a)`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 LEFT JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{100, "0100", 100, "0100"}, {125, "0125", 125, "0125"}, {150, "0150", ``, ``}, {175, 0175, ``, ``}, {200, "0200", 200, "0200"}, {225, "0225", 225, "0225"}, {250, "0250", 250, "0250"}, {275, "0275", 275, "0275"}, {300, "0300", ``, ``}, {325, "0325", ``, ``}, {350, "0350", 350, "0350"}, {375, "0375", 375, "0375"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.b, t1.c, t2.a, t2.c FROM prt2_adv t1 LEFT JOIN prt1_adv t2 ON (t1.b = t2.a) WHERE t1.a = 0 ORDER BY t1.b, t2.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.b, t2.a`}, {`->  Hash Right Join`}, {`Hash Cond: (t2.a = t1.b)`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p1 t2_1`}, {`->  Seq Scan on prt1_adv_p2 t2_2`}, {`->  Seq Scan on prt1_adv_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 t1_1`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p2 t1_2`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p3 t1_3`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_extra t1_4`}, {`Filter: (a = 0)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_1.a = t2_1.b)`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_2.a = t2_2.b)`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1_3.a = t2_3.b)`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM prt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results:   []sql.Row{{150, 0, "0150"}, {175, 0, 0175}, {300, 0, "0300"}, {325, 0, "0325"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt2_adv t1 WHERE NOT EXISTS (SELECT 1 FROM prt1_adv t2 WHERE t1.b = t2.a) AND t1.a = 0 ORDER BY t1.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.b`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1.b = t2.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 t1_1`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p2 t1_2`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p3 t1_3`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_extra t1_4`}, {`Filter: (a = 0)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p1 t2_1`}, {`->  Seq Scan on prt1_adv_p2 t2_2`}, {`->  Seq Scan on prt1_adv_p3 t2_3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT 175 phv, * FROM prt1_adv WHERE prt1_adv.b = 0) t1 FULL JOIN (SELECT 425 phv, * FROM prt2_adv WHERE prt2_adv.a = 0) t2 ON (t1.a = t2.b) WHERE t1.phv = t1.a OR t2.phv = t2.b ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1_adv.a, prt2_adv.b`}, {`->  Hash Full Join`}, {`Hash Cond: (prt2_adv.b = prt1_adv.a)`}, {`Filter: (((175) = prt1_adv.a) OR ((425) = prt2_adv.b))`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 prt2_adv_1`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p2 prt2_adv_2`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p3 prt2_adv_3`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_extra prt2_adv_4`}, {`Filter: (a = 0)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p1 prt1_adv_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p2 prt1_adv_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p3 prt1_adv_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.b, t1.c, t2.a, t2.c, t3.a, t3.c FROM prt2_adv t1 LEFT JOIN prt1_adv t2 ON (t1.b = t2.a) INNER JOIN prt1_adv t3 ON (t1.b = t3.a) WHERE t1.a = 0 ORDER BY t1.b, t2.a, t3.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.b, t2.a`}, {`->  Append`}, {`->  Nested Loop Left Join`}, {`->  Nested Loop`}, {`->  Seq Scan on prt2_adv_p1 t1_1`}, {`Filter: (a = 0)`}, {`->  Index Scan using prt1_adv_p1_a_idx on prt1_adv_p1 t3_1`}, {`Index Cond: (a = t1_1.b)`}, {`->  Index Scan using prt1_adv_p1_a_idx on prt1_adv_p1 t2_1`}, {`Index Cond: (a = t1_1.b)`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_2.a = t1_2.b)`}, {`->  Seq Scan on prt1_adv_p2 t2_2`}, {`->  Hash`}, {`->  Hash Join`}, {`Hash Cond: (t3_2.a = t1_2.b)`}, {`->  Seq Scan on prt1_adv_p2 t3_2`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p2 t1_2`}, {`Filter: (a = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_3.a = t1_3.b)`}, {`->  Seq Scan on prt1_adv_p3 t2_3`}, {`->  Hash`}, {`->  Hash Join`}, {`Hash Cond: (t3_3.a = t1_3.b)`}, {`->  Seq Scan on prt1_adv_p3 t3_3`}, {`->  Hash`}, {`->  Seq Scan on prt2_adv_p3 t1_3`}, {`Filter: (a = 0)`}},
			},
			{
				Statement: `SELECT t1.b, t1.c, t2.a, t2.c, t3.a, t3.c FROM prt2_adv t1 LEFT JOIN prt1_adv t2 ON (t1.b = t2.a) INNER JOIN prt1_adv t3 ON (t1.b = t3.a) WHERE t1.a = 0 ORDER BY t1.b, t2.a, t3.a;`,
				Results:   []sql.Row{{100, "0100", 100, "0100", 100, "0100"}, {125, "0125", 125, "0125", 125, "0125"}, {200, "0200", 200, "0200", 200, "0200"}, {225, "0225", 225, "0225", 225, "0225"}, {250, "0250", 250, "0250", 250, "0250"}, {275, "0275", 275, "0275", 275, "0275"}, {350, "0350", 350, "0350", 350, "0350"}, {375, "0375", 375, "0375", 375, "0375"}},
			},
			{
				Statement: `DROP TABLE prt2_adv_extra;`,
			},
			{
				Statement: `ALTER TABLE prt2_adv DETACH PARTITION prt2_adv_p3;`,
			},
			{
				Statement: `CREATE TABLE prt2_adv_p3_1 PARTITION OF prt2_adv FOR VALUES FROM (350) TO (375);`,
			},
			{
				Statement: `CREATE TABLE prt2_adv_p3_2 PARTITION OF prt2_adv FOR VALUES FROM (375) TO (500);`,
			},
			{
				Statement: `INSERT INTO prt2_adv SELECT i % 25, i, to_char(i, 'FM0000') FROM generate_series(350, 499) i;`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Join`}, {`Hash Cond: (t2.b = t1.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Seq Scan on prt2_adv_p3_1 t2_3`}, {`->  Seq Scan on prt2_adv_p3_2 t2_4`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1_adv t1 WHERE EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Semi Join`}, {`Hash Cond: (t1.a = t2.b)`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Seq Scan on prt2_adv_p3_1 t2_3`}, {`->  Seq Scan on prt2_adv_p3_2 t2_4`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 LEFT JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b`}, {`->  Hash Right Join`}, {`Hash Cond: (t2.b = t1.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Seq Scan on prt2_adv_p3_1 t2_3`}, {`->  Seq Scan on prt2_adv_p3_2 t2_4`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM prt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM prt2_adv t2 WHERE t1.a = t2.b) AND t1.b = 0 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Anti Join`}, {`Hash Cond: (t1.a = t2.b)`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p3 t1_3`}, {`Filter: (b = 0)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Seq Scan on prt2_adv_p3_1 t2_3`}, {`->  Seq Scan on prt2_adv_p3_2 t2_4`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM (SELECT 175 phv, * FROM prt1_adv WHERE prt1_adv.b = 0) t1 FULL JOIN (SELECT 425 phv, * FROM prt2_adv WHERE prt2_adv.a = 0) t2 ON (t1.a = t2.b) WHERE t1.phv = t1.a OR t2.phv = t2.b ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: prt1_adv.a, prt2_adv.b`}, {`->  Hash Full Join`}, {`Hash Cond: (prt2_adv.b = prt1_adv.a)`}, {`Filter: (((175) = prt1_adv.a) OR ((425) = prt2_adv.b))`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 prt2_adv_1`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p2 prt2_adv_2`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p3_1 prt2_adv_3`}, {`Filter: (a = 0)`}, {`->  Seq Scan on prt2_adv_p3_2 prt2_adv_4`}, {`Filter: (a = 0)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p1 prt1_adv_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p2 prt1_adv_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p3 prt1_adv_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `DROP TABLE prt2_adv_p3_1;`,
			},
			{
				Statement: `DROP TABLE prt2_adv_p3_2;`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `ALTER TABLE prt1_adv DETACH PARTITION prt1_adv_p1;`,
			},
			{
				Statement: `ALTER TABLE prt1_adv ATTACH PARTITION prt1_adv_p1 DEFAULT;`,
			},
			{
				Statement: `ALTER TABLE prt1_adv DETACH PARTITION prt1_adv_p3;`,
			},
			{
				Statement: `ANALYZE prt1_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_2.a)`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p1 t1_2`}, {`Filter: (b = 0)`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.b = t1_1.a)`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p2 t1_1`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{100, "0100", 100, "0100"}, {125, "0125", 125, "0125"}, {200, "0200", 200, "0200"}, {225, "0225", 225, "0225"}, {250, "0250", 250, "0250"}, {275, "0275", 275, "0275"}},
			},
			{
				Statement: `ALTER TABLE prt1_adv ATTACH PARTITION prt1_adv_p3 FOR VALUES FROM (300) TO (400);`,
			},
			{
				Statement: `ANALYZE prt1_adv;`,
			},
			{
				Statement: `ALTER TABLE prt2_adv ATTACH PARTITION prt2_adv_p3 FOR VALUES FROM (350) TO (500);`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Join`}, {`Hash Cond: (t2.b = t1.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p2 t1_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p3 t1_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p1 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `ALTER TABLE prt2_adv DETACH PARTITION prt2_adv_p3;`,
			},
			{
				Statement: `ALTER TABLE prt2_adv ATTACH PARTITION prt2_adv_p3 DEFAULT;`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Join`}, {`Hash Cond: (t2.b = t1.a)`}, {`->  Append`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Seq Scan on prt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on prt1_adv_p2 t1_1`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p3 t1_2`}, {`Filter: (b = 0)`}, {`->  Seq Scan on prt1_adv_p1 t1_3`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `DROP TABLE prt1_adv_p3;`,
			},
			{
				Statement: `ANALYZE prt1_adv;`,
			},
			{
				Statement: `DROP TABLE prt2_adv_p3;`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `CREATE TABLE prt3_adv (a int, b int, c varchar) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE prt3_adv_p1 PARTITION OF prt3_adv FOR VALUES FROM (200) TO (300);`,
			},
			{
				Statement: `CREATE TABLE prt3_adv_p2 PARTITION OF prt3_adv FOR VALUES FROM (300) TO (400);`,
			},
			{
				Statement: `CREATE INDEX prt3_adv_a_idx ON prt3_adv (a);`,
			},
			{
				Statement: `INSERT INTO prt3_adv SELECT i, i % 25, to_char(i, 'FM0000') FROM generate_series(200, 399) i;`,
			},
			{
				Statement: `ANALYZE prt3_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c, t3.a, t3.c FROM prt1_adv t1 LEFT JOIN prt2_adv t2 ON (t1.a = t2.b) LEFT JOIN prt3_adv t3 ON (t1.a = t3.a) WHERE t1.b = 0 ORDER BY t1.a, t2.b, t3.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.b, t3.a`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: (t3_1.a = t1_1.a)`}, {`->  Seq Scan on prt3_adv_p1 t3_1`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_2.b = t1_1.a)`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p2 t1_1`}, {`Filter: (b = 0)`}, {`->  Hash Right Join`}, {`Hash Cond: (t3_2.a = t1_2.a)`}, {`->  Seq Scan on prt3_adv_p2 t3_2`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: (t2_1.b = t1_2.a)`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p1 t1_2`}, {`Filter: (b = 0)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c, t3.a, t3.c FROM prt1_adv t1 LEFT JOIN prt2_adv t2 ON (t1.a = t2.b) LEFT JOIN prt3_adv t3 ON (t1.a = t3.a) WHERE t1.b = 0 ORDER BY t1.a, t2.b, t3.a;`,
				Results:   []sql.Row{{100, "0100", 100, "0100", ``, ``}, {125, "0125", 125, "0125", ``, ``}, {150, "0150", ``, ``, ``, ``}, {175, 0175, ``, ``, ``, ``}, {200, "0200", 200, "0200", 200, "0200"}, {225, "0225", 225, "0225", 225, "0225"}, {250, "0250", 250, "0250", 250, "0250"}, {275, "0275", 275, "0275", 275, "0275"}},
			},
			{
				Statement: `DROP TABLE prt1_adv;`,
			},
			{
				Statement: `DROP TABLE prt2_adv;`,
			},
			{
				Statement: `DROP TABLE prt3_adv;`,
			},
			{
				Statement: `CREATE TABLE prt1_adv (a int, b int, c varchar) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE prt1_adv_p1 PARTITION OF prt1_adv FOR VALUES FROM (100) TO (200);`,
			},
			{
				Statement: `CREATE TABLE prt1_adv_p2 PARTITION OF prt1_adv FOR VALUES FROM (200) TO (300);`,
			},
			{
				Statement: `CREATE TABLE prt1_adv_p3 PARTITION OF prt1_adv FOR VALUES FROM (300) TO (400);`,
			},
			{
				Statement: `CREATE INDEX prt1_adv_a_idx ON prt1_adv (a);`,
			},
			{
				Statement: `INSERT INTO prt1_adv SELECT i, i % 25, to_char(i, 'FM0000') FROM generate_series(100, 399) i;`,
			},
			{
				Statement: `ANALYZE prt1_adv;`,
			},
			{
				Statement: `CREATE TABLE prt2_adv (a int, b int, c varchar) PARTITION BY RANGE (b);`,
			},
			{
				Statement: `CREATE TABLE prt2_adv_p1 PARTITION OF prt2_adv FOR VALUES FROM (100) TO (200);`,
			},
			{
				Statement: `CREATE TABLE prt2_adv_p2 PARTITION OF prt2_adv FOR VALUES FROM (200) TO (400);`,
			},
			{
				Statement: `CREATE INDEX prt2_adv_b_idx ON prt2_adv (b);`,
			},
			{
				Statement: `INSERT INTO prt2_adv SELECT i % 25, i, to_char(i, 'FM0000') FROM generate_series(100, 399) i;`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.a < 300 AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: ((a < 300) AND (b = 0))`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: ((a < 300) AND (b = 0))`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.a < 300 AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{100, "0100", 100, "0100"}, {125, "0125", 125, "0125"}, {150, "0150", 150, "0150"}, {175, 0175, 175, 0175}, {200, "0200", 200, "0200"}, {225, "0225", 225, "0225"}, {250, "0250", 250, "0250"}, {275, "0275", 275, "0275"}},
			},
			{
				Statement: `DROP TABLE prt1_adv_p3;`,
			},
			{
				Statement: `CREATE TABLE prt1_adv_default PARTITION OF prt1_adv DEFAULT;`,
			},
			{
				Statement: `ANALYZE prt1_adv;`,
			},
			{
				Statement: `CREATE TABLE prt2_adv_default PARTITION OF prt2_adv DEFAULT;`,
			},
			{
				Statement: `ANALYZE prt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.a >= 100 AND t1.a < 300 AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: (t2_1.b = t1_1.a)`}, {`->  Seq Scan on prt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p1 t1_1`}, {`Filter: ((a >= 100) AND (a < 300) AND (b = 0))`}, {`->  Hash Join`}, {`Hash Cond: (t2_2.b = t1_2.a)`}, {`->  Seq Scan on prt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on prt1_adv_p2 t1_2`}, {`Filter: ((a >= 100) AND (a < 300) AND (b = 0))`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.b, t2.c FROM prt1_adv t1 INNER JOIN prt2_adv t2 ON (t1.a = t2.b) WHERE t1.a >= 100 AND t1.a < 300 AND t1.b = 0 ORDER BY t1.a, t2.b;`,
				Results:   []sql.Row{{100, "0100", 100, "0100"}, {125, "0125", 125, "0125"}, {150, "0150", 150, "0150"}, {175, 0175, 175, 0175}, {200, "0200", 200, "0200"}, {225, "0225", 225, "0225"}, {250, "0250", 250, "0250"}, {275, "0275", 275, "0275"}},
			},
			{
				Statement: `DROP TABLE prt1_adv;`,
			},
			{
				Statement: `DROP TABLE prt2_adv;`,
			},
			{
				Statement: `CREATE TABLE plt1_adv (a int, b int, c text) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p1 PARTITION OF plt1_adv FOR VALUES IN ('0001', '0003');`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p2 PARTITION OF plt1_adv FOR VALUES IN ('0004', '0006');`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p3 PARTITION OF plt1_adv FOR VALUES IN ('0008', '0009');`,
			},
			{
				Statement: `INSERT INTO plt1_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (1, 3, 4, 6, 8, 9);`,
			},
			{
				Statement: `ANALYZE plt1_adv;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv (a int, b int, c text) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p1 PARTITION OF plt2_adv FOR VALUES IN ('0002', '0003');`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p2 PARTITION OF plt2_adv FOR VALUES IN ('0004', '0006');`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p3 PARTITION OF plt2_adv FOR VALUES IN ('0007', '0009');`,
			},
			{
				Statement: `INSERT INTO plt2_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (2, 3, 4, 6, 7, 9);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt1_adv t1 WHERE EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop Semi Join`}, {`Join Filter: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Nested Loop Semi Join`}, {`Join Filter: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Nested Loop Semi Join`}, {`Join Filter: ((t1_3.a = t2_3.a) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM plt1_adv t1 WHERE EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, 3, "0003"}, {4, 4, "0004"}, {6, 6, "0006"}, {9, 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{1, "0001", ``, ``}, {3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {8, "0008", ``, ``}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop Anti Join`}, {`Join Filter: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Nested Loop Anti Join`}, {`Join Filter: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Nested Loop Anti Join`}, {`Join Filter: ((t1_3.a = t2_3.a) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM plt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{1, 1, "0001"}, {8, 8, "0008"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.a`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`Filter: ((COALESCE(t1_1.b, 0) < 10) AND (COALESCE(t2_1.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`Filter: ((COALESCE(t1_2.b, 0) < 10) AND (COALESCE(t2_2.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_3.a = t2_3.a) AND (t1_3.c = t2_3.c))`}, {`Filter: ((COALESCE(t1_3.b, 0) < 10) AND (COALESCE(t2_3.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results:   []sql.Row{{1, "0001", ``, ``}, {3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {8, "0008", ``, ``}, {9, "0009", 9, "0009"}, {``, ``, 2, "0002"}, {``, ``, 7, "0007"}},
			},
			{
				Statement: `CREATE TABLE plt2_adv_extra PARTITION OF plt2_adv FOR VALUES IN ('0000');`,
			},
			{
				Statement: `INSERT INTO plt2_adv_extra VALUES (0, 0, '0000');`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt1_adv t1 WHERE EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop Semi Join`}, {`Join Filter: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Nested Loop Semi Join`}, {`Join Filter: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Nested Loop Semi Join`}, {`Join Filter: ((t1_3.a = t2_3.a) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM plt1_adv t1 WHERE EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, 3, "0003"}, {4, 4, "0004"}, {6, 6, "0006"}, {9, 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{1, "0001", ``, ``}, {3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {8, "0008", ``, ``}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt2_adv t1 LEFT JOIN plt1_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t2_1`}, {`->  Seq Scan on plt1_adv_p2 t2_2`}, {`->  Seq Scan on plt1_adv_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_extra t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p1 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p2 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p3 t1_4`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Nested Loop Anti Join`}, {`Join Filter: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Nested Loop Anti Join`}, {`Join Filter: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Nested Loop Anti Join`}, {`Join Filter: ((t1_3.a = t2_3.a) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM plt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{1, 1, "0001"}, {8, 8, "0008"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt2_adv t1 WHERE NOT EXISTS (SELECT 1 FROM plt1_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Anti Join`}, {`Hash Cond: ((t1.a = t2.a) AND (t1.c = t2.c))`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_extra t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p1 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p2 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p3 t1_4`}, {`Filter: (b < 10)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t2_1`}, {`->  Seq Scan on plt1_adv_p2 t2_2`}, {`->  Seq Scan on plt1_adv_p3 t2_3`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.a`}, {`->  Hash Full Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`Filter: ((COALESCE(t1.b, 0) < 10) AND (COALESCE(t2.b, 0) < 10))`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_extra t2_1`}, {`->  Seq Scan on plt2_adv_p1 t2_2`}, {`->  Seq Scan on plt2_adv_p2 t2_3`}, {`->  Seq Scan on plt2_adv_p3 t2_4`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}},
			},
			{
				Statement: `DROP TABLE plt2_adv_extra;`,
			},
			{
				Statement: `ALTER TABLE plt2_adv DETACH PARTITION plt2_adv_p2;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p2_1 PARTITION OF plt2_adv FOR VALUES IN ('0004');`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p2_2 PARTITION OF plt2_adv FOR VALUES IN ('0006');`,
			},
			{
				Statement: `INSERT INTO plt2_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (4, 6);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2_1 t2_2`}, {`->  Seq Scan on plt2_adv_p2_2 t2_3`}, {`->  Seq Scan on plt2_adv_p3 t2_4`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt1_adv t1 WHERE EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Semi Join`}, {`Hash Cond: ((t1.a = t2.a) AND (t1.c = t2.c))`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2_1 t2_2`}, {`->  Seq Scan on plt2_adv_p2_2 t2_3`}, {`->  Seq Scan on plt2_adv_p3 t2_4`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2_1 t2_2`}, {`->  Seq Scan on plt2_adv_p2_2 t2_3`}, {`->  Seq Scan on plt2_adv_p3 t2_4`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Anti Join`}, {`Hash Cond: ((t1.a = t2.a) AND (t1.c = t2.c))`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2_1 t2_2`}, {`->  Seq Scan on plt2_adv_p2_2 t2_3`}, {`->  Seq Scan on plt2_adv_p3 t2_4`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.a`}, {`->  Hash Full Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`Filter: ((COALESCE(t1.b, 0) < 10) AND (COALESCE(t2.b, 0) < 10))`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2_1 t2_2`}, {`->  Seq Scan on plt2_adv_p2_2 t2_3`}, {`->  Seq Scan on plt2_adv_p3 t2_4`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}},
			},
			{
				Statement: `DROP TABLE plt2_adv_p2_1;`,
			},
			{
				Statement: `DROP TABLE plt2_adv_p2_2;`,
			},
			{
				Statement: `ALTER TABLE plt2_adv ATTACH PARTITION plt2_adv_p2 FOR VALUES IN ('0004', '0006');`,
			},
			{
				Statement: `ALTER TABLE plt1_adv DETACH PARTITION plt1_adv_p1;`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p1_null PARTITION OF plt1_adv FOR VALUES IN (NULL, '0001', '0003');`,
			},
			{
				Statement: `INSERT INTO plt1_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (1, 3);`,
			},
			{
				Statement: `INSERT INTO plt1_adv VALUES (-1, -1, NULL);`,
			},
			{
				Statement: `ANALYZE plt1_adv;`,
			},
			{
				Statement: `ALTER TABLE plt2_adv DETACH PARTITION plt2_adv_p3;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p3_null PARTITION OF plt2_adv FOR VALUES IN (NULL, '0007', '0009');`,
			},
			{
				Statement: `INSERT INTO plt2_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (7, 9);`,
			},
			{
				Statement: `INSERT INTO plt2_adv VALUES (-1, -1, NULL);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1_null t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3_null t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt1_adv t1 WHERE EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Semi Join`}, {`Hash Cond: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on plt1_adv_p1_null t1_1`}, {`Filter: (b < 10)`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Nested Loop Semi Join`}, {`Join Filter: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Nested Loop Semi Join`}, {`Join Filter: ((t1_3.a = t2_3.a) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p3_null t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM plt1_adv t1 WHERE EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, 3, "0003"}, {4, 4, "0004"}, {6, 6, "0006"}, {9, 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1_null t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3_null t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{-1, ``, ``, ``}, {1, "0001", ``, ``}, {3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {8, "0008", ``, ``}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.* FROM plt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Anti Join`}, {`Hash Cond: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on plt1_adv_p1_null t1_1`}, {`Filter: (b < 10)`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Nested Loop Anti Join`}, {`Join Filter: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Nested Loop Anti Join`}, {`Join Filter: ((t1_3.a = t2_3.a) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_p3_null t2_3`}},
			},
			{
				Statement: `SELECT t1.* FROM plt1_adv t1 WHERE NOT EXISTS (SELECT 1 FROM plt2_adv t2 WHERE t1.a = t2.a AND t1.c = t2.c) AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{-1, -1, ``}, {1, 1, "0001"}, {8, 8, "0008"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.a`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`Filter: ((COALESCE(t1_1.b, 0) < 10) AND (COALESCE(t2_1.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_p1_null t1_1`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`Filter: ((COALESCE(t1_2.b, 0) < 10) AND (COALESCE(t2_2.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash Full Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`Filter: ((COALESCE(t1_3.b, 0) < 10) AND (COALESCE(t2_3.b, 0) < 10))`}, {`->  Seq Scan on plt2_adv_p3_null t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results:   []sql.Row{{-1, ``, ``, ``}, {1, "0001", ``, ``}, {3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {8, "0008", ``, ``}, {9, "0009", 9, "0009"}, {``, ``, -1, ``}, {``, ``, 2, "0002"}, {``, ``, 7, "0007"}},
			},
			{
				Statement: `DROP TABLE plt1_adv_p1_null;`,
			},
			{
				Statement: `ALTER TABLE plt1_adv ATTACH PARTITION plt1_adv_p1 FOR VALUES IN ('0001', '0003');`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_extra PARTITION OF plt1_adv FOR VALUES IN (NULL);`,
			},
			{
				Statement: `INSERT INTO plt1_adv VALUES (-1, -1, NULL);`,
			},
			{
				Statement: `ANALYZE plt1_adv;`,
			},
			{
				Statement: `DROP TABLE plt2_adv_p3_null;`,
			},
			{
				Statement: `ALTER TABLE plt2_adv ATTACH PARTITION plt2_adv_p3 FOR VALUES IN ('0007', '0009');`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_extra t1_4`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.a`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1.a = t2.a) AND (t1.c = t2.c))`}, {`Filter: ((COALESCE(t1.b, 0) < 10) AND (COALESCE(t2.b, 0) < 10))`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`->  Seq Scan on plt1_adv_extra t1_4`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}},
			},
			{
				Statement: `CREATE TABLE plt2_adv_extra PARTITION OF plt2_adv FOR VALUES IN (NULL);`,
			},
			{
				Statement: `INSERT INTO plt2_adv VALUES (-1, -1, NULL);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Nested Loop Left Join`}, {`Join Filter: ((t1_4.a = t2_4.a) AND (t1_4.c = t2_4.c))`}, {`->  Seq Scan on plt1_adv_extra t1_4`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_extra t2_4`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{-1, ``, ``, ``}, {1, "0001", ``, ``}, {3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {8, "0008", ``, ``}, {9, "0009", 9, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t2.a`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_1.a = t2_1.a) AND (t1_1.c = t2_1.c))`}, {`Filter: ((COALESCE(t1_1.b, 0) < 10) AND (COALESCE(t2_1.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`Filter: ((COALESCE(t1_2.b, 0) < 10) AND (COALESCE(t2_2.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_3.a = t2_3.a) AND (t1_3.c = t2_3.c))`}, {`Filter: ((COALESCE(t1_3.b, 0) < 10) AND (COALESCE(t2_3.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash Full Join`}, {`Hash Cond: ((t1_4.a = t2_4.a) AND (t1_4.c = t2_4.c))`}, {`Filter: ((COALESCE(t1_4.b, 0) < 10) AND (COALESCE(t2_4.b, 0) < 10))`}, {`->  Seq Scan on plt1_adv_extra t1_4`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_extra t2_4`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 FULL JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE coalesce(t1.b, 0) < 10 AND coalesce(t2.b, 0) < 10 ORDER BY t1.a, t2.a;`,
				Results:   []sql.Row{{-1, ``, ``, ``}, {1, "0001", ``, ``}, {3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}, {8, "0008", ``, ``}, {9, "0009", 9, "0009"}, {``, ``, -1, ``}, {``, ``, 2, "0002"}, {``, ``, 7, "0007"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c, t3.a, t3.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) LEFT JOIN plt1_adv t3 ON (t1.a = t3.a AND t1.c = t3.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: ((t3_1.a = t1_1.a) AND (t3_1.c = t1_1.c))`}, {`->  Seq Scan on plt1_adv_p1 t3_1`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t3_2.a = t1_2.a) AND (t3_2.c = t1_2.c))`}, {`->  Seq Scan on plt1_adv_p2 t3_2`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t3_3.a = t1_3.a) AND (t3_3.c = t1_3.c))`}, {`->  Seq Scan on plt1_adv_p3 t3_3`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_3.a = t1_3.a) AND (t2_3.c = t1_3.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_3`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_3`}, {`Filter: (b < 10)`}, {`->  Nested Loop Left Join`}, {`Join Filter: ((t1_4.a = t3_4.a) AND (t1_4.c = t3_4.c))`}, {`->  Nested Loop Left Join`}, {`Join Filter: ((t1_4.a = t2_4.a) AND (t1_4.c = t2_4.c))`}, {`->  Seq Scan on plt1_adv_extra t1_4`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt2_adv_extra t2_4`}, {`->  Seq Scan on plt1_adv_extra t3_4`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c, t3.a, t3.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) LEFT JOIN plt1_adv t3 ON (t1.a = t3.a AND t1.c = t3.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{-1, ``, ``, ``, ``, ``}, {1, "0001", ``, ``, 1, "0001"}, {3, "0003", 3, "0003", 3, "0003"}, {4, "0004", 4, "0004", 4, "0004"}, {6, "0006", 6, "0006", 6, "0006"}, {8, "0008", ``, ``, 8, "0008"}, {9, "0009", 9, "0009", 9, "0009"}},
			},
			{
				Statement: `DROP TABLE plt1_adv_extra;`,
			},
			{
				Statement: `DROP TABLE plt2_adv_extra;`,
			},
			{
				Statement: `ALTER TABLE plt1_adv DETACH PARTITION plt1_adv_p1;`,
			},
			{
				Statement: `ALTER TABLE plt1_adv ATTACH PARTITION plt1_adv_p1 DEFAULT;`,
			},
			{
				Statement: `DROP TABLE plt1_adv_p3;`,
			},
			{
				Statement: `ANALYZE plt1_adv;`,
			},
			{
				Statement: `DROP TABLE plt2_adv_p3;`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_2.a) AND (t2_1.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_1.a) AND (t2_2.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_1`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}},
			},
			{
				Statement: `ALTER TABLE plt2_adv DETACH PARTITION plt2_adv_p2;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p2_ext PARTITION OF plt2_adv FOR VALUES IN ('0004', '0005', '0006');`,
			},
			{
				Statement: `INSERT INTO plt2_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (4, 5, 6);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2_ext t2_2`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p2 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p1 t1_2`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `ALTER TABLE plt2_adv DETACH PARTITION plt2_adv_p2_ext;`,
			},
			{
				Statement: `ALTER TABLE plt2_adv ATTACH PARTITION plt2_adv_p2_ext DEFAULT;`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Append`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Seq Scan on plt2_adv_p2_ext t2_2`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on plt1_adv_p2 t1_1`}, {`Filter: (b < 10)`}, {`->  Seq Scan on plt1_adv_p1 t1_2`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `DROP TABLE plt2_adv_p2_ext;`,
			},
			{
				Statement: `ALTER TABLE plt2_adv ATTACH PARTITION plt2_adv_p2 FOR VALUES IN ('0004', '0006');`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `CREATE TABLE plt3_adv (a int, b int, c text) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE plt3_adv_p1 PARTITION OF plt3_adv FOR VALUES IN ('0004', '0006');`,
			},
			{
				Statement: `CREATE TABLE plt3_adv_p2 PARTITION OF plt3_adv FOR VALUES IN ('0007', '0009');`,
			},
			{
				Statement: `INSERT INTO plt3_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (4, 6, 7, 9);`,
			},
			{
				Statement: `ANALYZE plt3_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c, t3.a, t3.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) LEFT JOIN plt3_adv t3 ON (t1.a = t3.a AND t1.c = t3.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Right Join`}, {`Hash Cond: ((t3_1.a = t1_1.a) AND (t3_1.c = t1_1.c))`}, {`->  Seq Scan on plt3_adv_p1 t3_1`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_2.a = t1_1.a) AND (t2_2.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_1`}, {`Filter: (b < 10)`}, {`->  Hash Right Join`}, {`Hash Cond: ((t3_2.a = t1_2.a) AND (t3_2.c = t1_2.c))`}, {`->  Seq Scan on plt3_adv_p2 t3_2`}, {`->  Hash`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2_1.a = t1_2.a) AND (t2_1.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_2`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c, t3.a, t3.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) LEFT JOIN plt3_adv t3 ON (t1.a = t3.a AND t1.c = t3.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{1, "0001", ``, ``, ``, ``}, {3, "0003", 3, "0003", ``, ``}, {4, "0004", 4, "0004", 4, "0004"}, {6, "0006", 6, "0006", 6, "0006"}},
			},
			{
				Statement: `DROP TABLE plt2_adv_p1;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p1_null PARTITION OF plt2_adv FOR VALUES IN (NULL, '0001', '0003');`,
			},
			{
				Statement: `INSERT INTO plt2_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (1, 3);`,
			},
			{
				Statement: `INSERT INTO plt2_adv VALUES (-1, -1, NULL);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_2.a) AND (t2_1.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p1_null t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p1 t1_2`}, {`Filter: (b < 10)`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_1.a) AND (t2_2.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1_1`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{1, "0001", 1, "0001"}, {3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}},
			},
			{
				Statement: `DROP TABLE plt2_adv_p1_null;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p1_null PARTITION OF plt2_adv FOR VALUES IN (NULL);`,
			},
			{
				Statement: `INSERT INTO plt2_adv VALUES (-1, -1, NULL);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Seq Scan on plt2_adv_p2 t2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p2 t1`}, {`Filter: (b < 10)`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{4, "0004", 4, "0004"}, {6, "0006", 6, "0006"}},
			},
			{
				Statement: `DROP TABLE plt1_adv;`,
			},
			{
				Statement: `DROP TABLE plt2_adv;`,
			},
			{
				Statement: `DROP TABLE plt3_adv;`,
			},
			{
				Statement: `CREATE TABLE plt1_adv (a int, b int, c text) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p1 PARTITION OF plt1_adv FOR VALUES IN ('0001');`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p2 PARTITION OF plt1_adv FOR VALUES IN ('0002');`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p3 PARTITION OF plt1_adv FOR VALUES IN ('0003');`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p4 PARTITION OF plt1_adv FOR VALUES IN (NULL, '0004', '0005');`,
			},
			{
				Statement: `INSERT INTO plt1_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (1, 2, 3, 4, 5);`,
			},
			{
				Statement: `INSERT INTO plt1_adv VALUES (-1, -1, NULL);`,
			},
			{
				Statement: `ANALYZE plt1_adv;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv (a int, b int, c text) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p1 PARTITION OF plt2_adv FOR VALUES IN ('0001', '0002');`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p2 PARTITION OF plt2_adv FOR VALUES IN (NULL);`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p3 PARTITION OF plt2_adv FOR VALUES IN ('0003');`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p4 PARTITION OF plt2_adv FOR VALUES IN ('0004', '0005');`,
			},
			{
				Statement: `INSERT INTO plt2_adv SELECT i, i, to_char(i % 10, 'FM0000') FROM generate_series(1, 299) i WHERE i % 10 IN (1, 2, 3, 4, 5);`,
			},
			{
				Statement: `INSERT INTO plt2_adv VALUES (-1, -1, NULL);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.c IN ('0003', '0004', '0005') AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_1`}, {`Filter: ((b < 10) AND (c = ANY ('{"0003","0004","0005"}'::text[])))`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p4 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p4 t1_2`}, {`Filter: ((b < 10) AND (c = ANY ('{"0003","0004","0005"}'::text[])))`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.c IN ('0003', '0004', '0005') AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {5, "0005", 5, "0005"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.c IS NULL AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Seq Scan on plt2_adv_p4 t2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p4 t1`}, {`Filter: ((c IS NULL) AND (b < 10))`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.c IS NULL AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{-1, ``, ``, ``}},
			},
			{
				Statement: `CREATE TABLE plt1_adv_default PARTITION OF plt1_adv DEFAULT;`,
			},
			{
				Statement: `ANALYZE plt1_adv;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_default PARTITION OF plt2_adv DEFAULT;`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.c IN ('0003', '0004', '0005') AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t2_1.a = t1_1.a) AND (t2_1.c = t1_1.c))`}, {`->  Seq Scan on plt2_adv_p3 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p3 t1_1`}, {`Filter: ((b < 10) AND (c = ANY ('{"0003","0004","0005"}'::text[])))`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.c = t1_2.c))`}, {`->  Seq Scan on plt2_adv_p4 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p4 t1_2`}, {`Filter: ((b < 10) AND (c = ANY ('{"0003","0004","0005"}'::text[])))`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 INNER JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.c IN ('0003', '0004', '0005') AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{3, "0003", 3, "0003"}, {4, "0004", 4, "0004"}, {5, "0005", 5, "0005"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.c IS NULL AND t1.b < 10 ORDER BY t1.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a`}, {`->  Hash Right Join`}, {`Hash Cond: ((t2.a = t1.a) AND (t2.c = t1.c))`}, {`->  Seq Scan on plt2_adv_p4 t2`}, {`->  Hash`}, {`->  Seq Scan on plt1_adv_p4 t1`}, {`Filter: ((c IS NULL) AND (b < 10))`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c FROM plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE t1.c IS NULL AND t1.b < 10 ORDER BY t1.a;`,
				Results:   []sql.Row{{-1, ``, ``, ``}},
			},
			{
				Statement: `DROP TABLE plt1_adv;`,
			},
			{
				Statement: `DROP TABLE plt2_adv;`,
			},
			{
				Statement: `CREATE TABLE plt1_adv (a int, b int, c text) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p1 PARTITION OF plt1_adv FOR VALUES IN ('0000', '0001', '0002');`,
			},
			{
				Statement: `CREATE TABLE plt1_adv_p2 PARTITION OF plt1_adv FOR VALUES IN ('0003', '0004');`,
			},
			{
				Statement: `INSERT INTO plt1_adv SELECT i, i, to_char(i % 5, 'FM0000') FROM generate_series(0, 24) i;`,
			},
			{
				Statement: `ANALYZE plt1_adv;`,
			},
			{
				Statement: `CREATE TABLE plt2_adv (a int, b int, c text) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p1 PARTITION OF plt2_adv FOR VALUES IN ('0002');`,
			},
			{
				Statement: `CREATE TABLE plt2_adv_p2 PARTITION OF plt2_adv FOR VALUES IN ('0003', '0004');`,
			},
			{
				Statement: `INSERT INTO plt2_adv SELECT i, i, to_char(i % 5, 'FM0000') FROM generate_series(0, 24) i WHERE i % 5 IN (2, 3, 4);`,
			},
			{
				Statement: `ANALYZE plt2_adv;`,
			},
			{
				Statement: `CREATE TABLE plt3_adv (a int, b int, c text) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE plt3_adv_p1 PARTITION OF plt3_adv FOR VALUES IN ('0001');`,
			},
			{
				Statement: `CREATE TABLE plt3_adv_p2 PARTITION OF plt3_adv FOR VALUES IN ('0003', '0004');`,
			},
			{
				Statement: `INSERT INTO plt3_adv SELECT i, i, to_char(i % 5, 'FM0000') FROM generate_series(0, 24) i WHERE i % 5 IN (1, 3, 4);`,
			},
			{
				Statement: `ANALYZE plt3_adv;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.a, t1.c, t2.a, t2.c, t3.a, t3.c FROM (plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.c = t2.c)) FULL JOIN plt3_adv t3 ON (t1.c = t3.c) WHERE coalesce(t1.a, 0) % 5 != 3 AND coalesce(t1.a, 0) % 5 != 4 ORDER BY t1.c, t1.a, t2.a, t3.a;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.c, t1.a, t2.a, t3.a`}, {`->  Append`}, {`->  Hash Full Join`}, {`Hash Cond: (t1_1.c = t3_1.c)`}, {`Filter: (((COALESCE(t1_1.a, 0) % 5) <> 3) AND ((COALESCE(t1_1.a, 0) % 5) <> 4))`}, {`->  Hash Left Join`}, {`Hash Cond: (t1_1.c = t2_1.c)`}, {`->  Seq Scan on plt1_adv_p1 t1_1`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p1 t2_1`}, {`->  Hash`}, {`->  Seq Scan on plt3_adv_p1 t3_1`}, {`->  Hash Full Join`}, {`Hash Cond: (t1_2.c = t3_2.c)`}, {`Filter: (((COALESCE(t1_2.a, 0) % 5) <> 3) AND ((COALESCE(t1_2.a, 0) % 5) <> 4))`}, {`->  Hash Left Join`}, {`Hash Cond: (t1_2.c = t2_2.c)`}, {`->  Seq Scan on plt1_adv_p2 t1_2`}, {`->  Hash`}, {`->  Seq Scan on plt2_adv_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on plt3_adv_p2 t3_2`}},
			},
			{
				Statement: `SELECT t1.a, t1.c, t2.a, t2.c, t3.a, t3.c FROM (plt1_adv t1 LEFT JOIN plt2_adv t2 ON (t1.c = t2.c)) FULL JOIN plt3_adv t3 ON (t1.c = t3.c) WHERE coalesce(t1.a, 0) % 5 != 3 AND coalesce(t1.a, 0) % 5 != 4 ORDER BY t1.c, t1.a, t2.a, t3.a;`,
				Results:   []sql.Row{{0, "0000", ``, ``, ``, ``}, {5, "0000", ``, ``, ``, ``}, {10, "0000", ``, ``, ``, ``}, {15, "0000", ``, ``, ``, ``}, {20, "0000", ``, ``, ``, ``}, {1, "0001", ``, ``, 1, "0001"}, {1, "0001", ``, ``, 6, "0001"}, {1, "0001", ``, ``, 11, "0001"}, {1, "0001", ``, ``, 16, "0001"}, {1, "0001", ``, ``, 21, "0001"}, {6, "0001", ``, ``, 1, "0001"}, {6, "0001", ``, ``, 6, "0001"}, {6, "0001", ``, ``, 11, "0001"}, {6, "0001", ``, ``, 16, "0001"}, {6, "0001", ``, ``, 21, "0001"}, {11, "0001", ``, ``, 1, "0001"}, {11, "0001", ``, ``, 6, "0001"}, {11, "0001", ``, ``, 11, "0001"}, {11, "0001", ``, ``, 16, "0001"}, {11, "0001", ``, ``, 21, "0001"}, {16, "0001", ``, ``, 1, "0001"}, {16, "0001", ``, ``, 6, "0001"}, {16, "0001", ``, ``, 11, "0001"}, {16, "0001", ``, ``, 16, "0001"}, {16, "0001", ``, ``, 21, "0001"}, {21, "0001", ``, ``, 1, "0001"}, {21, "0001", ``, ``, 6, "0001"}, {21, "0001", ``, ``, 11, "0001"}, {21, "0001", ``, ``, 16, "0001"}, {21, "0001", ``, ``, 21, "0001"}, {2, "0002", 2, "0002", ``, ``}, {2, "0002", 7, "0002", ``, ``}, {2, "0002", 12, "0002", ``, ``}, {2, "0002", 17, "0002", ``, ``}, {2, "0002", 22, "0002", ``, ``}, {7, "0002", 2, "0002", ``, ``}, {7, "0002", 7, "0002", ``, ``}, {7, "0002", 12, "0002", ``, ``}, {7, "0002", 17, "0002", ``, ``}, {7, "0002", 22, "0002", ``, ``}, {12, "0002", 2, "0002", ``, ``}, {12, "0002", 7, "0002", ``, ``}, {12, "0002", 12, "0002", ``, ``}, {12, "0002", 17, "0002", ``, ``}, {12, "0002", 22, "0002", ``, ``}, {17, "0002", 2, "0002", ``, ``}, {17, "0002", 7, "0002", ``, ``}, {17, "0002", 12, "0002", ``, ``}, {17, "0002", 17, "0002", ``, ``}, {17, "0002", 22, "0002", ``, ``}, {22, "0002", 2, "0002", ``, ``}, {22, "0002", 7, "0002", ``, ``}, {22, "0002", 12, "0002", ``, ``}, {22, "0002", 17, "0002", ``, ``}, {22, "0002", 22, "0002", ``, ``}},
			},
			{
				Statement: `DROP TABLE plt1_adv;`,
			},
			{
				Statement: `DROP TABLE plt2_adv;`,
			},
			{
				Statement: `DROP TABLE plt3_adv;`,
			},
			{
				Statement: `CREATE TABLE alpha (a double precision, b int, c text) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE alpha_neg PARTITION OF alpha FOR VALUES FROM ('-Infinity') TO (0) PARTITION BY RANGE (b);`,
			},
			{
				Statement: `CREATE TABLE alpha_pos PARTITION OF alpha FOR VALUES FROM (0) TO (10.0) PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE alpha_neg_p1 PARTITION OF alpha_neg FOR VALUES FROM (100) TO (200);`,
			},
			{
				Statement: `CREATE TABLE alpha_neg_p2 PARTITION OF alpha_neg FOR VALUES FROM (200) TO (300);`,
			},
			{
				Statement: `CREATE TABLE alpha_neg_p3 PARTITION OF alpha_neg FOR VALUES FROM (300) TO (400);`,
			},
			{
				Statement: `CREATE TABLE alpha_pos_p1 PARTITION OF alpha_pos FOR VALUES IN ('0001', '0003');`,
			},
			{
				Statement: `CREATE TABLE alpha_pos_p2 PARTITION OF alpha_pos FOR VALUES IN ('0004', '0006');`,
			},
			{
				Statement: `CREATE TABLE alpha_pos_p3 PARTITION OF alpha_pos FOR VALUES IN ('0008', '0009');`,
			},
			{
				Statement: `INSERT INTO alpha_neg SELECT -1.0, i, to_char(i % 10, 'FM0000') FROM generate_series(100, 399) i WHERE i % 10 IN (1, 3, 4, 6, 8, 9);`,
			},
			{
				Statement: `INSERT INTO alpha_pos SELECT  1.0, i, to_char(i % 10, 'FM0000') FROM generate_series(100, 399) i WHERE i % 10 IN (1, 3, 4, 6, 8, 9);`,
			},
			{
				Statement: `ANALYZE alpha;`,
			},
			{
				Statement: `CREATE TABLE beta (a double precision, b int, c text) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE beta_neg PARTITION OF beta FOR VALUES FROM (-10.0) TO (0) PARTITION BY RANGE (b);`,
			},
			{
				Statement: `CREATE TABLE beta_pos PARTITION OF beta FOR VALUES FROM (0) TO ('Infinity') PARTITION BY LIST (c);`,
			},
			{
				Statement: `CREATE TABLE beta_neg_p1 PARTITION OF beta_neg FOR VALUES FROM (100) TO (150);`,
			},
			{
				Statement: `CREATE TABLE beta_neg_p2 PARTITION OF beta_neg FOR VALUES FROM (200) TO (300);`,
			},
			{
				Statement: `CREATE TABLE beta_neg_p3 PARTITION OF beta_neg FOR VALUES FROM (350) TO (500);`,
			},
			{
				Statement: `CREATE TABLE beta_pos_p1 PARTITION OF beta_pos FOR VALUES IN ('0002', '0003');`,
			},
			{
				Statement: `CREATE TABLE beta_pos_p2 PARTITION OF beta_pos FOR VALUES IN ('0004', '0006');`,
			},
			{
				Statement: `CREATE TABLE beta_pos_p3 PARTITION OF beta_pos FOR VALUES IN ('0007', '0009');`,
			},
			{
				Statement: `INSERT INTO beta_neg SELECT -1.0, i, to_char(i % 10, 'FM0000') FROM generate_series(100, 149) i WHERE i % 10 IN (2, 3, 4, 6, 7, 9);`,
			},
			{
				Statement: `INSERT INTO beta_neg SELECT -1.0, i, to_char(i % 10, 'FM0000') FROM generate_series(200, 299) i WHERE i % 10 IN (2, 3, 4, 6, 7, 9);`,
			},
			{
				Statement: `INSERT INTO beta_neg SELECT -1.0, i, to_char(i % 10, 'FM0000') FROM generate_series(350, 499) i WHERE i % 10 IN (2, 3, 4, 6, 7, 9);`,
			},
			{
				Statement: `INSERT INTO beta_pos SELECT  1.0, i, to_char(i % 10, 'FM0000') FROM generate_series(100, 149) i WHERE i % 10 IN (2, 3, 4, 6, 7, 9);`,
			},
			{
				Statement: `INSERT INTO beta_pos SELECT  1.0, i, to_char(i % 10, 'FM0000') FROM generate_series(200, 299) i WHERE i % 10 IN (2, 3, 4, 6, 7, 9);`,
			},
			{
				Statement: `INSERT INTO beta_pos SELECT  1.0, i, to_char(i % 10, 'FM0000') FROM generate_series(350, 499) i WHERE i % 10 IN (2, 3, 4, 6, 7, 9);`,
			},
			{
				Statement: `ANALYZE beta;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.*, t2.* FROM alpha t1 INNER JOIN beta t2 ON (t1.a = t2.a AND t1.b = t2.b) WHERE t1.b >= 125 AND t1.b < 225 ORDER BY t1.a, t1.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t1.b`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t1_1.a = t2_1.a) AND (t1_1.b = t2_1.b))`}, {`->  Seq Scan on alpha_neg_p1 t1_1`}, {`Filter: ((b >= 125) AND (b < 225))`}, {`->  Hash`}, {`->  Seq Scan on beta_neg_p1 t2_1`}, {`->  Hash Join`}, {`Hash Cond: ((t2_2.a = t1_2.a) AND (t2_2.b = t1_2.b))`}, {`->  Seq Scan on beta_neg_p2 t2_2`}, {`->  Hash`}, {`->  Seq Scan on alpha_neg_p2 t1_2`}, {`Filter: ((b >= 125) AND (b < 225))`}, {`->  Hash Join`}, {`Hash Cond: ((t2_4.a = t1_4.a) AND (t2_4.b = t1_4.b))`}, {`->  Append`}, {`->  Seq Scan on beta_pos_p1 t2_4`}, {`->  Seq Scan on beta_pos_p2 t2_5`}, {`->  Seq Scan on beta_pos_p3 t2_6`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on alpha_pos_p1 t1_4`}, {`Filter: ((b >= 125) AND (b < 225))`}, {`->  Seq Scan on alpha_pos_p2 t1_5`}, {`Filter: ((b >= 125) AND (b < 225))`}, {`->  Seq Scan on alpha_pos_p3 t1_6`}, {`Filter: ((b >= 125) AND (b < 225))`}},
			},
			{
				Statement: `SELECT t1.*, t2.* FROM alpha t1 INNER JOIN beta t2 ON (t1.a = t2.a AND t1.b = t2.b) WHERE t1.b >= 125 AND t1.b < 225 ORDER BY t1.a, t1.b;`,
				Results:   []sql.Row{{-1, 126, "0006", -1, 126, "0006"}, {-1, 129, "0009", -1, 129, "0009"}, {-1, 133, "0003", -1, 133, "0003"}, {-1, 134, "0004", -1, 134, "0004"}, {-1, 136, "0006", -1, 136, "0006"}, {-1, 139, "0009", -1, 139, "0009"}, {-1, 143, "0003", -1, 143, "0003"}, {-1, 144, "0004", -1, 144, "0004"}, {-1, 146, "0006", -1, 146, "0006"}, {-1, 149, "0009", -1, 149, "0009"}, {-1, 203, "0003", -1, 203, "0003"}, {-1, 204, "0004", -1, 204, "0004"}, {-1, 206, "0006", -1, 206, "0006"}, {-1, 209, "0009", -1, 209, "0009"}, {-1, 213, "0003", -1, 213, "0003"}, {-1, 214, "0004", -1, 214, "0004"}, {-1, 216, "0006", -1, 216, "0006"}, {-1, 219, "0009", -1, 219, "0009"}, {-1, 223, "0003", -1, 223, "0003"}, {-1, 224, "0004", -1, 224, "0004"}, {1, 126, "0006", 1, 126, "0006"}, {1, 129, "0009", 1, 129, "0009"}, {1, 133, "0003", 1, 133, "0003"}, {1, 134, "0004", 1, 134, "0004"}, {1, 136, "0006", 1, 136, "0006"}, {1, 139, "0009", 1, 139, "0009"}, {1, 143, "0003", 1, 143, "0003"}, {1, 144, "0004", 1, 144, "0004"}, {1, 146, "0006", 1, 146, "0006"}, {1, 149, "0009", 1, 149, "0009"}, {1, 203, "0003", 1, 203, "0003"}, {1, 204, "0004", 1, 204, "0004"}, {1, 206, "0006", 1, 206, "0006"}, {1, 209, "0009", 1, 209, "0009"}, {1, 213, "0003", 1, 213, "0003"}, {1, 214, "0004", 1, 214, "0004"}, {1, 216, "0006", 1, 216, "0006"}, {1, 219, "0009", 1, 219, "0009"}, {1, 223, "0003", 1, 223, "0003"}, {1, 224, "0004", 1, 224, "0004"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.*, t2.* FROM alpha t1 INNER JOIN beta t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE ((t1.b >= 100 AND t1.b < 110) OR (t1.b >= 200 AND t1.b < 210)) AND ((t2.b >= 100 AND t2.b < 110) OR (t2.b >= 200 AND t2.b < 210)) AND t1.c IN ('0004', '0009') ORDER BY t1.a, t1.b, t2.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t1.b, t2.b`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t1_2.a = t2_2.a) AND (t1_2.c = t2_2.c))`}, {`->  Append`}, {`->  Seq Scan on alpha_neg_p1 t1_2`}, {`Filter: ((c = ANY ('{"0004","0009"}'::text[])) AND (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210))))`}, {`->  Seq Scan on alpha_neg_p2 t1_3`}, {`Filter: ((c = ANY ('{"0004","0009"}'::text[])) AND (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210))))`}, {`->  Hash`}, {`->  Append`}, {`->  Seq Scan on beta_neg_p1 t2_2`}, {`Filter: (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210)))`}, {`->  Seq Scan on beta_neg_p2 t2_3`}, {`Filter: (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210)))`}, {`->  Nested Loop`}, {`Join Filter: ((t1_4.a = t2_4.a) AND (t1_4.c = t2_4.c))`}, {`->  Seq Scan on alpha_pos_p2 t1_4`}, {`Filter: ((c = ANY ('{"0004","0009"}'::text[])) AND (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210))))`}, {`->  Seq Scan on beta_pos_p2 t2_4`}, {`Filter: (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210)))`}, {`->  Nested Loop`}, {`Join Filter: ((t1_5.a = t2_5.a) AND (t1_5.c = t2_5.c))`}, {`->  Seq Scan on alpha_pos_p3 t1_5`}, {`Filter: ((c = ANY ('{"0004","0009"}'::text[])) AND (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210))))`}, {`->  Seq Scan on beta_pos_p3 t2_5`}, {`Filter: (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210)))`}},
			},
			{
				Statement: `SELECT t1.*, t2.* FROM alpha t1 INNER JOIN beta t2 ON (t1.a = t2.a AND t1.c = t2.c) WHERE ((t1.b >= 100 AND t1.b < 110) OR (t1.b >= 200 AND t1.b < 210)) AND ((t2.b >= 100 AND t2.b < 110) OR (t2.b >= 200 AND t2.b < 210)) AND t1.c IN ('0004', '0009') ORDER BY t1.a, t1.b, t2.b;`,
				Results:   []sql.Row{{-1, 104, "0004", -1, 104, "0004"}, {-1, 104, "0004", -1, 204, "0004"}, {-1, 109, "0009", -1, 109, "0009"}, {-1, 109, "0009", -1, 209, "0009"}, {-1, 204, "0004", -1, 104, "0004"}, {-1, 204, "0004", -1, 204, "0004"}, {-1, 209, "0009", -1, 109, "0009"}, {-1, 209, "0009", -1, 209, "0009"}, {1, 104, "0004", 1, 104, "0004"}, {1, 104, "0004", 1, 204, "0004"}, {1, 109, "0009", 1, 109, "0009"}, {1, 109, "0009", 1, 209, "0009"}, {1, 204, "0004", 1, 104, "0004"}, {1, 204, "0004", 1, 204, "0004"}, {1, 209, "0009", 1, 109, "0009"}, {1, 209, "0009", 1, 209, "0009"}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT t1.*, t2.* FROM alpha t1 INNER JOIN beta t2 ON (t1.a = t2.a AND t1.b = t2.b AND t1.c = t2.c) WHERE ((t1.b >= 100 AND t1.b < 110) OR (t1.b >= 200 AND t1.b < 210)) AND ((t2.b >= 100 AND t2.b < 110) OR (t2.b >= 200 AND t2.b < 210)) AND t1.c IN ('0004', '0009') ORDER BY t1.a, t1.b;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: t1.a, t1.b`}, {`->  Append`}, {`->  Hash Join`}, {`Hash Cond: ((t1_1.a = t2_1.a) AND (t1_1.b = t2_1.b) AND (t1_1.c = t2_1.c))`}, {`->  Seq Scan on alpha_neg_p1 t1_1`}, {`Filter: ((c = ANY ('{"0004","0009"}'::text[])) AND (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210))))`}, {`->  Hash`}, {`->  Seq Scan on beta_neg_p1 t2_1`}, {`Filter: (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210)))`}, {`->  Hash Join`}, {`Hash Cond: ((t1_2.a = t2_2.a) AND (t1_2.b = t2_2.b) AND (t1_2.c = t2_2.c))`}, {`->  Seq Scan on alpha_neg_p2 t1_2`}, {`Filter: ((c = ANY ('{"0004","0009"}'::text[])) AND (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210))))`}, {`->  Hash`}, {`->  Seq Scan on beta_neg_p2 t2_2`}, {`Filter: (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210)))`}, {`->  Nested Loop`}, {`Join Filter: ((t1_3.a = t2_3.a) AND (t1_3.b = t2_3.b) AND (t1_3.c = t2_3.c))`}, {`->  Seq Scan on alpha_pos_p2 t1_3`}, {`Filter: ((c = ANY ('{"0004","0009"}'::text[])) AND (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210))))`}, {`->  Seq Scan on beta_pos_p2 t2_3`}, {`Filter: (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210)))`}, {`->  Nested Loop`}, {`Join Filter: ((t1_4.a = t2_4.a) AND (t1_4.b = t2_4.b) AND (t1_4.c = t2_4.c))`}, {`->  Seq Scan on alpha_pos_p3 t1_4`}, {`Filter: ((c = ANY ('{"0004","0009"}'::text[])) AND (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210))))`}, {`->  Seq Scan on beta_pos_p3 t2_4`}, {`Filter: (((b >= 100) AND (b < 110)) OR ((b >= 200) AND (b < 210)))`}},
			},
			{
				Statement: `SELECT t1.*, t2.* FROM alpha t1 INNER JOIN beta t2 ON (t1.a = t2.a AND t1.b = t2.b AND t1.c = t2.c) WHERE ((t1.b >= 100 AND t1.b < 110) OR (t1.b >= 200 AND t1.b < 210)) AND ((t2.b >= 100 AND t2.b < 110) OR (t2.b >= 200 AND t2.b < 210)) AND t1.c IN ('0004', '0009') ORDER BY t1.a, t1.b;`,
				Results:   []sql.Row{{-1, 104, "0004", -1, 104, "0004"}, {-1, 109, "0009", -1, 109, "0009"}, {-1, 204, "0004", -1, 204, "0004"}, {-1, 209, "0009", -1, 209, "0009"}, {1, 104, "0004", 1, 104, "0004"}, {1, 109, "0009", 1, 109, "0009"}, {1, 204, "0004", 1, 204, "0004"}, {1, 209, "0009", 1, 209, "0009"}},
			},
			{
				Statement: `CREATE TABLE fract_t (id BIGINT, PRIMARY KEY (id)) PARTITION BY RANGE (id);`,
			},
			{
				Statement: `CREATE TABLE fract_t0 PARTITION OF fract_t FOR VALUES FROM ('0') TO ('1000');`,
			},
			{
				Statement: `CREATE TABLE fract_t1 PARTITION OF fract_t FOR VALUES FROM ('1000') TO ('2000');`,
			},
			{
				Statement: `INSERT INTO fract_t (id) (SELECT generate_series(0, 1999));`,
			},
			{
				Statement: `ANALYZE fract_t;`,
			},
			{
				Statement: `SET max_parallel_workers_per_gather = 0;`,
			},
			{
				Statement: `SET enable_partitionwise_join = on;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM fract_t x LEFT JOIN fract_t y USING (id) ORDER BY id ASC LIMIT 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Merge Append`}, {`Sort Key: x.id`}, {`->  Merge Left Join`}, {`Merge Cond: (x_1.id = y_1.id)`}, {`->  Index Only Scan using fract_t0_pkey on fract_t0 x_1`}, {`->  Index Only Scan using fract_t0_pkey on fract_t0 y_1`}, {`->  Merge Left Join`}, {`Merge Cond: (x_2.id = y_2.id)`}, {`->  Index Only Scan using fract_t1_pkey on fract_t1 x_2`}, {`->  Index Only Scan using fract_t1_pkey on fract_t1 y_2`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM fract_t x LEFT JOIN fract_t y USING (id) ORDER BY id DESC LIMIT 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Merge Append`}, {`Sort Key: x.id DESC`}, {`->  Nested Loop Left Join`}, {`->  Index Only Scan Backward using fract_t0_pkey on fract_t0 x_1`}, {`->  Index Only Scan using fract_t0_pkey on fract_t0 y_1`}, {`Index Cond: (id = x_1.id)`}, {`->  Nested Loop Left Join`}, {`->  Index Only Scan Backward using fract_t1_pkey on fract_t1 x_2`}, {`->  Index Only Scan using fract_t1_pkey on fract_t1 y_2`}, {`Index Cond: (id = x_2.id)`}},
			},
			{
				Statement: `DROP TABLE fract_t;`,
			},
			{
				Statement: `RESET max_parallel_workers_per_gather;`,
			},
			{
				Statement: `RESET enable_partitionwise_join;`,
			},
		},
	})
}
