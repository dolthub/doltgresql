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

func TestMemoize(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_memoize)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_memoize,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create function explain_memoize(query text, hide_hitmiss bool) returns setof text
language plpgsql as
$$
declare
    ln text;`,
			},
			{
				Statement: `begin
    for ln in
        execute format('explain (analyze, costs off, summary off, timing off) %s',
            query)
    loop
        if hide_hitmiss = true then
                ln := regexp_replace(ln, 'Hits: 0', 'Hits: Zero');`,
			},
			{
				Statement: `                ln := regexp_replace(ln, 'Hits: \d+', 'Hits: N');`,
			},
			{
				Statement: `                ln := regexp_replace(ln, 'Misses: 0', 'Misses: Zero');`,
			},
			{
				Statement: `                ln := regexp_replace(ln, 'Misses: \d+', 'Misses: N');`,
			},
			{
				Statement: `        end if;`,
			},
			{
				Statement: `        ln := regexp_replace(ln, 'Evictions: 0', 'Evictions: Zero');`,
			},
			{
				Statement: `        ln := regexp_replace(ln, 'Evictions: \d+', 'Evictions: N');`,
			},
			{
				Statement: `        ln := regexp_replace(ln, 'Memory Usage: \d+', 'Memory Usage: N');`,
			},
			{
				Statement: `	ln := regexp_replace(ln, 'Heap Fetches: \d+', 'Heap Fetches: N');`,
			},
			{
				Statement: `	ln := regexp_replace(ln, 'loops=\d+', 'loops=N');`,
			},
			{
				Statement: `        return next ln;`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SET enable_hashjoin TO off;`,
			},
			{
				Statement: `SET enable_bitmapscan TO off;`,
			},
			{
				Statement: `SELECT explain_memoize('
SELECT COUNT(*),AVG(t1.unique1) FROM tenk1 t1
INNER JOIN tenk1 t2 ON t1.unique1 = t2.twenty
WHERE t2.unique1 < 1000;', false);`,
				Results: []sql.Row{{`Aggregate (actual rows=1 loops=N)`}, {`->  Nested Loop (actual rows=1000 loops=N)`}, {`->  Seq Scan on tenk1 t2 (actual rows=1000 loops=N)`}, {`Filter: (unique1 < 1000)`}, {`Rows Removed by Filter: 9000`}, {`->  Memoize (actual rows=1 loops=N)`}, {`Cache Key: t2.twenty`}, {`Cache Mode: logical`}, {`Hits: 980  Misses: 20  Evictions: Zero  Overflows: 0  Memory Usage: NkB`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 t1 (actual rows=1 loops=N)`}, {`Index Cond: (unique1 = t2.twenty)`}, {`Heap Fetches: N`}},
			},
			{
				Statement: `SELECT COUNT(*),AVG(t1.unique1) FROM tenk1 t1
INNER JOIN tenk1 t2 ON t1.unique1 = t2.twenty
WHERE t2.unique1 < 1000;`,
				Results: []sql.Row{{1000, 9.5000000000000000}},
			},
			{
				Statement: `SELECT explain_memoize('
SELECT COUNT(*),AVG(t2.unique1) FROM tenk1 t1,
LATERAL (SELECT t2.unique1 FROM tenk1 t2
         WHERE t1.twenty = t2.unique1 OFFSET 0) t2
WHERE t1.unique1 < 1000;', false);`,
				Results: []sql.Row{{`Aggregate (actual rows=1 loops=N)`}, {`->  Nested Loop (actual rows=1000 loops=N)`}, {`->  Seq Scan on tenk1 t1 (actual rows=1000 loops=N)`}, {`Filter: (unique1 < 1000)`}, {`Rows Removed by Filter: 9000`}, {`->  Memoize (actual rows=1 loops=N)`}, {`Cache Key: t1.twenty`}, {`Cache Mode: binary`}, {`Hits: 980  Misses: 20  Evictions: Zero  Overflows: 0  Memory Usage: NkB`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 t2 (actual rows=1 loops=N)`}, {`Index Cond: (unique1 = t1.twenty)`}, {`Heap Fetches: N`}},
			},
			{
				Statement: `SELECT COUNT(*),AVG(t2.unique1) FROM tenk1 t1,
LATERAL (SELECT t2.unique1 FROM tenk1 t2
         WHERE t1.twenty = t2.unique1 OFFSET 0) t2
WHERE t1.unique1 < 1000;`,
				Results: []sql.Row{{1000, 9.5000000000000000}},
			},
			{
				Statement: `SET work_mem TO '64kB';`,
			},
			{
				Statement: `SET hash_mem_multiplier TO 1.0;`,
			},
			{
				Statement: `SET enable_mergejoin TO off;`,
			},
			{
				Statement: `SELECT explain_memoize('
SELECT COUNT(*),AVG(t1.unique1) FROM tenk1 t1
INNER JOIN tenk1 t2 ON t1.unique1 = t2.thousand
WHERE t2.unique1 < 1200;', true);`,
				Results: []sql.Row{{`Aggregate (actual rows=1 loops=N)`}, {`->  Nested Loop (actual rows=1200 loops=N)`}, {`->  Seq Scan on tenk1 t2 (actual rows=1200 loops=N)`}, {`Filter: (unique1 < 1200)`}, {`Rows Removed by Filter: 8800`}, {`->  Memoize (actual rows=1 loops=N)`}, {`Cache Key: t2.thousand`}, {`Cache Mode: logical`}, {`Hits: N  Misses: N  Evictions: N  Overflows: 0  Memory Usage: NkB`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 t1 (actual rows=1 loops=N)`}, {`Index Cond: (unique1 = t2.thousand)`}, {`Heap Fetches: N`}},
			},
			{
				Statement: `CREATE TABLE flt (f float);`,
			},
			{
				Statement: `CREATE INDEX flt_f_idx ON flt (f);`,
			},
			{
				Statement: `INSERT INTO flt VALUES('-0.0'::float),('+0.0'::float);`,
			},
			{
				Statement: `ANALYZE flt;`,
			},
			{
				Statement: `SET enable_seqscan TO off;`,
			},
			{
				Statement: `SELECT explain_memoize('
SELECT * FROM flt f1 INNER JOIN flt f2 ON f1.f = f2.f;', false);`,
				Results: []sql.Row{{`Nested Loop (actual rows=4 loops=N)`}, {`->  Index Only Scan using flt_f_idx on flt f1 (actual rows=2 loops=N)`}, {`Heap Fetches: N`}, {`->  Memoize (actual rows=2 loops=N)`}, {`Cache Key: f1.f`}, {`Cache Mode: logical`}, {`Hits: 1  Misses: 1  Evictions: Zero  Overflows: 0  Memory Usage: NkB`}, {`->  Index Only Scan using flt_f_idx on flt f2 (actual rows=2 loops=N)`}, {`Index Cond: (f = f1.f)`}, {`Heap Fetches: N`}},
			},
			{
				Statement: `SELECT explain_memoize('
SELECT * FROM flt f1 INNER JOIN flt f2 ON f1.f >= f2.f;', false);`,
				Results: []sql.Row{{`Nested Loop (actual rows=4 loops=N)`}, {`->  Index Only Scan using flt_f_idx on flt f1 (actual rows=2 loops=N)`}, {`Heap Fetches: N`}, {`->  Memoize (actual rows=2 loops=N)`}, {`Cache Key: f1.f`}, {`Cache Mode: binary`}, {`Hits: 0  Misses: 2  Evictions: Zero  Overflows: 0  Memory Usage: NkB`}, {`->  Index Only Scan using flt_f_idx on flt f2 (actual rows=2 loops=N)`}, {`Index Cond: (f <= f1.f)`}, {`Heap Fetches: N`}},
			},
			{
				Statement: `DROP TABLE flt;`,
			},
			{
				Statement: `CREATE TABLE strtest (n name, t text);`,
			},
			{
				Statement: `CREATE INDEX strtest_n_idx ON strtest (n);`,
			},
			{
				Statement: `CREATE INDEX strtest_t_idx ON strtest (t);`,
			},
			{
				Statement: `INSERT INTO strtest VALUES('one','one'),('two','two'),('three',repeat(md5('three'),100));`,
			},
			{
				Statement: `INSERT INTO strtest SELECT * FROM strtest;`,
			},
			{
				Statement: `ANALYZE strtest;`,
			},
			{
				Statement: `SELECT explain_memoize('
SELECT * FROM strtest s1 INNER JOIN strtest s2 ON s1.n >= s2.n;', false);`,
				Results: []sql.Row{{`Nested Loop (actual rows=24 loops=N)`}, {`->  Seq Scan on strtest s1 (actual rows=6 loops=N)`}, {`->  Memoize (actual rows=4 loops=N)`}, {`Cache Key: s1.n`}, {`Cache Mode: binary`}, {`Hits: 3  Misses: 3  Evictions: Zero  Overflows: 0  Memory Usage: NkB`}, {`->  Index Scan using strtest_n_idx on strtest s2 (actual rows=4 loops=N)`}, {`Index Cond: (n <= s1.n)`}},
			},
			{
				Statement: `SELECT explain_memoize('
SELECT * FROM strtest s1 INNER JOIN strtest s2 ON s1.t >= s2.t;', false);`,
				Results: []sql.Row{{`Nested Loop (actual rows=24 loops=N)`}, {`->  Seq Scan on strtest s1 (actual rows=6 loops=N)`}, {`->  Memoize (actual rows=4 loops=N)`}, {`Cache Key: s1.t`}, {`Cache Mode: binary`}, {`Hits: 3  Misses: 3  Evictions: Zero  Overflows: 0  Memory Usage: NkB`}, {`->  Index Scan using strtest_t_idx on strtest s2 (actual rows=4 loops=N)`}, {`Index Cond: (t <= s1.t)`}},
			},
			{
				Statement: `DROP TABLE strtest;`,
			},
			{
				Statement: `SET enable_partitionwise_join TO on;`,
			},
			{
				Statement: `CREATE TABLE prt (a int) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE prt_p1 PARTITION OF prt FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `CREATE TABLE prt_p2 PARTITION OF prt FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `INSERT INTO prt VALUES (0), (0), (0), (0);`,
			},
			{
				Statement: `INSERT INTO prt VALUES (10), (10), (10), (10);`,
			},
			{
				Statement: `CREATE INDEX iprt_p1_a ON prt_p1 (a);`,
			},
			{
				Statement: `CREATE INDEX iprt_p2_a ON prt_p2 (a);`,
			},
			{
				Statement: `ANALYZE prt;`,
			},
			{
				Statement: `SELECT explain_memoize('
SELECT * FROM prt t1 INNER JOIN prt t2 ON t1.a = t2.a;', false);`,
				Results: []sql.Row{{`Append (actual rows=32 loops=N)`}, {`->  Nested Loop (actual rows=16 loops=N)`}, {`->  Index Only Scan using iprt_p1_a on prt_p1 t1_1 (actual rows=4 loops=N)`}, {`Heap Fetches: N`}, {`->  Memoize (actual rows=4 loops=N)`}, {`Cache Key: t1_1.a`}, {`Cache Mode: logical`}, {`Hits: 3  Misses: 1  Evictions: Zero  Overflows: 0  Memory Usage: NkB`}, {`->  Index Only Scan using iprt_p1_a on prt_p1 t2_1 (actual rows=4 loops=N)`}, {`Index Cond: (a = t1_1.a)`}, {`Heap Fetches: N`}, {`->  Nested Loop (actual rows=16 loops=N)`}, {`->  Index Only Scan using iprt_p2_a on prt_p2 t1_2 (actual rows=4 loops=N)`}, {`Heap Fetches: N`}, {`->  Memoize (actual rows=4 loops=N)`}, {`Cache Key: t1_2.a`}, {`Cache Mode: logical`}, {`Hits: 3  Misses: 1  Evictions: Zero  Overflows: 0  Memory Usage: NkB`}, {`->  Index Only Scan using iprt_p2_a on prt_p2 t2_2 (actual rows=4 loops=N)`}, {`Index Cond: (a = t1_2.a)`}, {`Heap Fetches: N`}},
			},
			{
				Statement: `DROP TABLE prt;`,
			},
			{
				Statement: `RESET enable_partitionwise_join;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT unique1 FROM tenk1 t0
WHERE unique1 < 3
  AND EXISTS (
	SELECT 1 FROM tenk1 t1
	INNER JOIN tenk1 t2 ON t1.unique1 = t2.hundred
	WHERE t0.ten = t1.twenty AND t0.two <> t2.four OFFSET 0);`,
				Results: []sql.Row{{`Index Scan using tenk1_unique1 on tenk1 t0`}, {`Index Cond: (unique1 < 3)`}, {`Filter: (SubPlan 1)`}, {`SubPlan 1`}, {`->  Nested Loop`}, {`->  Index Scan using tenk1_hundred on tenk1 t2`}, {`Filter: (t0.two <> four)`}, {`->  Memoize`}, {`Cache Key: t2.hundred`}, {`Cache Mode: logical`}, {`->  Index Scan using tenk1_unique1 on tenk1 t1`}, {`Index Cond: (unique1 = t2.hundred)`}, {`Filter: (t0.ten = twenty)`}},
			},
			{
				Statement: `SELECT unique1 FROM tenk1 t0
WHERE unique1 < 3
  AND EXISTS (
	SELECT 1 FROM tenk1 t1
	INNER JOIN tenk1 t2 ON t1.unique1 = t2.hundred
	WHERE t0.ten = t1.twenty AND t0.two <> t2.four OFFSET 0);`,
				Results: []sql.Row{{2}},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `RESET enable_mergejoin;`,
			},
			{
				Statement: `RESET work_mem;`,
			},
			{
				Statement: `RESET hash_mem_multiplier;`,
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
			{
				Statement: `RESET enable_hashjoin;`,
			},
			{
				Statement: `SET min_parallel_table_scan_size TO 0;`,
			},
			{
				Statement: `SET parallel_setup_cost TO 0;`,
			},
			{
				Statement: `SET parallel_tuple_cost TO 0;`,
			},
			{
				Statement: `SET max_parallel_workers_per_gather TO 2;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT COUNT(*),AVG(t2.unique1) FROM tenk1 t1,
LATERAL (SELECT t2.unique1 FROM tenk1 t2 WHERE t1.twenty = t2.unique1) t2
WHERE t1.unique1 < 1000;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Partial Aggregate`}, {`->  Nested Loop`}, {`->  Parallel Bitmap Heap Scan on tenk1 t1`}, {`Recheck Cond: (unique1 < 1000)`}, {`->  Bitmap Index Scan on tenk1_unique1`}, {`Index Cond: (unique1 < 1000)`}, {`->  Memoize`}, {`Cache Key: t1.twenty`}, {`Cache Mode: logical`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 t2`}, {`Index Cond: (unique1 = t1.twenty)`}},
			},
			{
				Statement: `SELECT COUNT(*),AVG(t2.unique1) FROM tenk1 t1,
LATERAL (SELECT t2.unique1 FROM tenk1 t2 WHERE t1.twenty = t2.unique1) t2
WHERE t1.unique1 < 1000;`,
				Results: []sql.Row{{1000, 9.5000000000000000}},
			},
			{
				Statement: `RESET max_parallel_workers_per_gather;`,
			},
			{
				Statement: `RESET parallel_tuple_cost;`,
			},
			{
				Statement: `RESET parallel_setup_cost;`,
			},
			{
				Statement: `RESET min_parallel_table_scan_size;`,
			},
		},
	})
}
