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

func TestSelectDistinct(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_select_distinct)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_select_distinct,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT DISTINCT two FROM onek ORDER BY 1;`,
				Results:   []sql.Row{{0}, {1}},
			},
			{
				Statement: `SELECT DISTINCT ten FROM onek ORDER BY 1;`,
				Results:   []sql.Row{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}},
			},
			{
				Statement: `SELECT DISTINCT string4 FROM onek ORDER BY 1;`,
				Results:   []sql.Row{{`AAAAxx`}, {`HHHHxx`}, {`OOOOxx`}, {`VVVVxx`}},
			},
			{
				Statement: `SELECT DISTINCT two, string4, ten
   FROM onek
   ORDER BY two using <, string4 using <, ten using <;`,
				Results: []sql.Row{{0, `AAAAxx`, 0}, {0, `AAAAxx`, 2}, {0, `AAAAxx`, 4}, {0, `AAAAxx`, 6}, {0, `AAAAxx`, 8}, {0, `HHHHxx`, 0}, {0, `HHHHxx`, 2}, {0, `HHHHxx`, 4}, {0, `HHHHxx`, 6}, {0, `HHHHxx`, 8}, {0, `OOOOxx`, 0}, {0, `OOOOxx`, 2}, {0, `OOOOxx`, 4}, {0, `OOOOxx`, 6}, {0, `OOOOxx`, 8}, {0, `VVVVxx`, 0}, {0, `VVVVxx`, 2}, {0, `VVVVxx`, 4}, {0, `VVVVxx`, 6}, {0, `VVVVxx`, 8}, {1, `AAAAxx`, 1}, {1, `AAAAxx`, 3}, {1, `AAAAxx`, 5}, {1, `AAAAxx`, 7}, {1, `AAAAxx`, 9}, {1, `HHHHxx`, 1}, {1, `HHHHxx`, 3}, {1, `HHHHxx`, 5}, {1, `HHHHxx`, 7}, {1, `HHHHxx`, 9}, {1, `OOOOxx`, 1}, {1, `OOOOxx`, 3}, {1, `OOOOxx`, 5}, {1, `OOOOxx`, 7}, {1, `OOOOxx`, 9}, {1, `VVVVxx`, 1}, {1, `VVVVxx`, 3}, {1, `VVVVxx`, 5}, {1, `VVVVxx`, 7}, {1, `VVVVxx`, 9}},
			},
			{
				Statement: `SELECT DISTINCT p.age FROM person* p ORDER BY age using >;`,
				Results:   []sql.Row{{98}, {88}, {78}, {68}, {60}, {58}, {50}, {48}, {40}, {38}, {34}, {30}, {28}, {25}, {24}, {23}, {20}, {19}, {18}, {8}},
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
SELECT count(*) FROM
  (SELECT DISTINCT two, four, two FROM tenk1) ss;`,
				Results: []sql.Row{{`Aggregate`}, {`Output: count(*)`}, {`->  HashAggregate`}, {`Output: tenk1.two, tenk1.four, tenk1.two`}, {`Group Key: tenk1.two, tenk1.four, tenk1.two`}, {`->  Seq Scan on public.tenk1`}, {`Output: tenk1.two, tenk1.four, tenk1.two`}},
			},
			{
				Statement: `SELECT count(*) FROM
  (SELECT DISTINCT two, four, two FROM tenk1) ss;`,
				Results: []sql.Row{{4}},
			},
			{
				Statement: `SET work_mem='64kB';`,
			},
			{
				Statement: `SET enable_hashagg=FALSE;`,
			},
			{
				Statement: `SET jit_above_cost=0;`,
			},
			{
				Statement: `EXPLAIN (costs off)
SELECT DISTINCT g%1000 FROM generate_series(0,9999) g;`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: ((g % 1000))`}, {`->  Function Scan on generate_series g`}},
			},
			{
				Statement: `CREATE TABLE distinct_group_1 AS
SELECT DISTINCT g%1000 FROM generate_series(0,9999) g;`,
			},
			{
				Statement: `SET jit_above_cost TO DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE distinct_group_2 AS
SELECT DISTINCT (g%1000)::text FROM generate_series(0,9999) g;`,
			},
			{
				Statement: `SET enable_hashagg=TRUE;`,
			},
			{
				Statement: `SET enable_sort=FALSE;`,
			},
			{
				Statement: `SET jit_above_cost=0;`,
			},
			{
				Statement: `EXPLAIN (costs off)
SELECT DISTINCT g%1000 FROM generate_series(0,9999) g;`,
				Results: []sql.Row{{`HashAggregate`}, {`Group Key: (g % 1000)`}, {`->  Function Scan on generate_series g`}},
			},
			{
				Statement: `CREATE TABLE distinct_hash_1 AS
SELECT DISTINCT g%1000 FROM generate_series(0,9999) g;`,
			},
			{
				Statement: `SET jit_above_cost TO DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE distinct_hash_2 AS
SELECT DISTINCT (g%1000)::text FROM generate_series(0,9999) g;`,
			},
			{
				Statement: `SET enable_sort=TRUE;`,
			},
			{
				Statement: `SET work_mem TO DEFAULT;`,
			},
			{
				Statement: `(SELECT * FROM distinct_hash_1 EXCEPT SELECT * FROM distinct_group_1)
  UNION ALL
(SELECT * FROM distinct_group_1 EXCEPT SELECT * FROM distinct_hash_1);`,
				Results: []sql.Row{},
			},
			{
				Statement: `(SELECT * FROM distinct_hash_1 EXCEPT SELECT * FROM distinct_group_1)
  UNION ALL
(SELECT * FROM distinct_group_1 EXCEPT SELECT * FROM distinct_hash_1);`,
				Results: []sql.Row{},
			},
			{
				Statement: `DROP TABLE distinct_hash_1;`,
			},
			{
				Statement: `DROP TABLE distinct_hash_2;`,
			},
			{
				Statement: `DROP TABLE distinct_group_1;`,
			},
			{
				Statement: `DROP TABLE distinct_group_2;`,
			},
			{
				Statement: `SET parallel_tuple_cost=0;`,
			},
			{
				Statement: `SET parallel_setup_cost=0;`,
			},
			{
				Statement: `SET min_parallel_table_scan_size=0;`,
			},
			{
				Statement: `SET max_parallel_workers_per_gather=2;`,
			},
			{
				Statement: `EXPLAIN (costs off)
SELECT DISTINCT four FROM tenk1;`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: four`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  HashAggregate`}, {`Group Key: four`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `SELECT DISTINCT four FROM tenk1;`,
				Results:   []sql.Row{{0}, {1}, {2}, {3}},
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION distinct_func(a INT) RETURNS INT AS $$
  BEGIN
    RETURN a;`,
			},
			{
				Statement: `  END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql PARALLEL UNSAFE;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT DISTINCT distinct_func(1) FROM tenk1;`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: (distinct_func(1))`}, {`->  Index Only Scan using tenk1_hundred on tenk1`}},
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION distinct_func(a INT) RETURNS INT AS $$
  BEGIN
    RETURN a;`,
			},
			{
				Statement: `  END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql PARALLEL SAFE;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT DISTINCT distinct_func(1) FROM tenk1;`,
				Results: []sql.Row{{`Unique`}, {`->  Sort`}, {`Sort Key: (distinct_func(1))`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `RESET max_parallel_workers_per_gather;`,
			},
			{
				Statement: `RESET min_parallel_table_scan_size;`,
			},
			{
				Statement: `RESET parallel_setup_cost;`,
			},
			{
				Statement: `RESET parallel_tuple_cost;`,
			},
			{
				Statement: `CREATE TEMP TABLE disttable (f1 integer);`,
			},
			{
				Statement: `INSERT INTO DISTTABLE VALUES(1);`,
			},
			{
				Statement: `INSERT INTO DISTTABLE VALUES(2);`,
			},
			{
				Statement: `INSERT INTO DISTTABLE VALUES(3);`,
			},
			{
				Statement: `INSERT INTO DISTTABLE VALUES(NULL);`,
			},
			{
				Statement: `SELECT f1, f1 IS DISTINCT FROM 2 as "not 2" FROM disttable;`,
				Results:   []sql.Row{{1, true}, {2, false}, {3, true}, {``, true}},
			},
			{
				Statement: `SELECT f1, f1 IS DISTINCT FROM NULL as "not null" FROM disttable;`,
				Results:   []sql.Row{{1, true}, {2, true}, {3, true}, {``, false}},
			},
			{
				Statement: `SELECT f1, f1 IS DISTINCT FROM f1 as "false" FROM disttable;`,
				Results:   []sql.Row{{1, false}, {2, false}, {3, false}, {``, false}},
			},
			{
				Statement: `SELECT f1, f1 IS DISTINCT FROM f1+1 as "not null" FROM disttable;`,
				Results:   []sql.Row{{1, true}, {2, true}, {3, true}, {``, false}},
			},
			{
				Statement: `SELECT 1 IS DISTINCT FROM 2 as "yes";`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT 2 IS DISTINCT FROM 2 as "no";`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 2 IS DISTINCT FROM null as "yes";`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT null IS DISTINCT FROM null as "no";`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 1 IS NOT DISTINCT FROM 2 as "no";`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 2 IS NOT DISTINCT FROM 2 as "yes";`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT 2 IS NOT DISTINCT FROM null as "no";`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT null IS NOT DISTINCT FROM null as "yes";`,
				Results:   []sql.Row{{true}},
			},
		},
	})
}
