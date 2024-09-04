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

func TestPartitionPrune(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_partition_prune)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_partition_prune,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `set plan_cache_mode = force_generic_plan;`,
			},
			{
				Statement: `create table lp (a char) partition by list (a);`,
			},
			{
				Statement: `create table lp_default partition of lp default;`,
			},
			{
				Statement: `create table lp_ef partition of lp for values in ('e', 'f');`,
			},
			{
				Statement: `create table lp_ad partition of lp for values in ('a', 'd');`,
			},
			{
				Statement: `create table lp_bc partition of lp for values in ('b', 'c');`,
			},
			{
				Statement: `create table lp_g partition of lp for values in ('g');`,
			},
			{
				Statement: `create table lp_null partition of lp for values in (null);`,
			},
			{
				Statement: `explain (costs off) select * from lp;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_ad lp_1`}, {`->  Seq Scan on lp_bc lp_2`}, {`->  Seq Scan on lp_ef lp_3`}, {`->  Seq Scan on lp_g lp_4`}, {`->  Seq Scan on lp_null lp_5`}, {`->  Seq Scan on lp_default lp_6`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a > 'a' and a < 'd';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_bc lp_1`}, {`Filter: ((a > 'a'::bpchar) AND (a < 'd'::bpchar))`}, {`->  Seq Scan on lp_default lp_2`}, {`Filter: ((a > 'a'::bpchar) AND (a < 'd'::bpchar))`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a > 'a' and a <= 'd';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_ad lp_1`}, {`Filter: ((a > 'a'::bpchar) AND (a <= 'd'::bpchar))`}, {`->  Seq Scan on lp_bc lp_2`}, {`Filter: ((a > 'a'::bpchar) AND (a <= 'd'::bpchar))`}, {`->  Seq Scan on lp_default lp_3`}, {`Filter: ((a > 'a'::bpchar) AND (a <= 'd'::bpchar))`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a = 'a';`,
				Results:   []sql.Row{{`Seq Scan on lp_ad lp`}, {`Filter: (a = 'a'::bpchar)`}},
			},
			{
				Statement: `explain (costs off) select * from lp where 'a' = a;	/* commuted */
         QUERY PLAN          
-----------------------------
 Seq Scan on lp_ad lp
   Filter: ('a'::bpchar = a)
(2 rows)
explain (costs off) select * from lp where a is not null;`,
				Results: []sql.Row{{`Append`}, {`->  Seq Scan on lp_ad lp_1`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on lp_bc lp_2`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on lp_ef lp_3`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on lp_g lp_4`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on lp_default lp_5`}, {`Filter: (a IS NOT NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a is null;`,
				Results:   []sql.Row{{`Seq Scan on lp_null lp`}, {`Filter: (a IS NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a = 'a' or a = 'c';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_ad lp_1`}, {`Filter: ((a = 'a'::bpchar) OR (a = 'c'::bpchar))`}, {`->  Seq Scan on lp_bc lp_2`}, {`Filter: ((a = 'a'::bpchar) OR (a = 'c'::bpchar))`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a is not null and (a = 'a' or a = 'c');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_ad lp_1`}, {`Filter: ((a IS NOT NULL) AND ((a = 'a'::bpchar) OR (a = 'c'::bpchar)))`}, {`->  Seq Scan on lp_bc lp_2`}, {`Filter: ((a IS NOT NULL) AND ((a = 'a'::bpchar) OR (a = 'c'::bpchar)))`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a <> 'g';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_ad lp_1`}, {`Filter: (a <> 'g'::bpchar)`}, {`->  Seq Scan on lp_bc lp_2`}, {`Filter: (a <> 'g'::bpchar)`}, {`->  Seq Scan on lp_ef lp_3`}, {`Filter: (a <> 'g'::bpchar)`}, {`->  Seq Scan on lp_default lp_4`}, {`Filter: (a <> 'g'::bpchar)`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a <> 'a' and a <> 'd';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_bc lp_1`}, {`Filter: ((a <> 'a'::bpchar) AND (a <> 'd'::bpchar))`}, {`->  Seq Scan on lp_ef lp_2`}, {`Filter: ((a <> 'a'::bpchar) AND (a <> 'd'::bpchar))`}, {`->  Seq Scan on lp_g lp_3`}, {`Filter: ((a <> 'a'::bpchar) AND (a <> 'd'::bpchar))`}, {`->  Seq Scan on lp_default lp_4`}, {`Filter: ((a <> 'a'::bpchar) AND (a <> 'd'::bpchar))`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a not in ('a', 'd');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_bc lp_1`}, {`Filter: (a <> ALL ('{a,d}'::bpchar[]))`}, {`->  Seq Scan on lp_ef lp_2`}, {`Filter: (a <> ALL ('{a,d}'::bpchar[]))`}, {`->  Seq Scan on lp_g lp_3`}, {`Filter: (a <> ALL ('{a,d}'::bpchar[]))`}, {`->  Seq Scan on lp_default lp_4`}, {`Filter: (a <> ALL ('{a,d}'::bpchar[]))`}},
			},
			{
				Statement: `create table coll_pruning (a text collate "C") partition by list (a);`,
			},
			{
				Statement: `create table coll_pruning_a partition of coll_pruning for values in ('a');`,
			},
			{
				Statement: `create table coll_pruning_b partition of coll_pruning for values in ('b');`,
			},
			{
				Statement: `create table coll_pruning_def partition of coll_pruning default;`,
			},
			{
				Statement: `explain (costs off) select * from coll_pruning where a collate "C" = 'a' collate "C";`,
				Results:   []sql.Row{{`Seq Scan on coll_pruning_a coll_pruning`}, {`Filter: (a = 'a'::text COLLATE "C")`}},
			},
			{
				Statement: `explain (costs off) select * from coll_pruning where a collate "POSIX" = 'a' collate "POSIX";`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coll_pruning_a coll_pruning_1`}, {`Filter: ((a)::text = 'a'::text COLLATE "POSIX")`}, {`->  Seq Scan on coll_pruning_b coll_pruning_2`}, {`Filter: ((a)::text = 'a'::text COLLATE "POSIX")`}, {`->  Seq Scan on coll_pruning_def coll_pruning_3`}, {`Filter: ((a)::text = 'a'::text COLLATE "POSIX")`}},
			},
			{
				Statement: `create table rlp (a int, b varchar) partition by range (a);`,
			},
			{
				Statement: `create table rlp_default partition of rlp default partition by list (a);`,
			},
			{
				Statement: `create table rlp_default_default partition of rlp_default default;`,
			},
			{
				Statement: `create table rlp_default_10 partition of rlp_default for values in (10);`,
			},
			{
				Statement: `create table rlp_default_30 partition of rlp_default for values in (30);`,
			},
			{
				Statement: `create table rlp_default_null partition of rlp_default for values in (null);`,
			},
			{
				Statement: `create table rlp1 partition of rlp for values from (minvalue) to (1);`,
			},
			{
				Statement: `create table rlp2 partition of rlp for values from (1) to (10);`,
			},
			{
				Statement: `create table rlp3 (b varchar, a int) partition by list (b varchar_ops);`,
			},
			{
				Statement: `create table rlp3_default partition of rlp3 default;`,
			},
			{
				Statement: `create table rlp3abcd partition of rlp3 for values in ('ab', 'cd');`,
			},
			{
				Statement: `create table rlp3efgh partition of rlp3 for values in ('ef', 'gh');`,
			},
			{
				Statement: `create table rlp3nullxy partition of rlp3 for values in (null, 'xy');`,
			},
			{
				Statement: `alter table rlp attach partition rlp3 for values from (15) to (20);`,
			},
			{
				Statement: `create table rlp4 partition of rlp for values from (20) to (30) partition by range (a);`,
			},
			{
				Statement: `create table rlp4_default partition of rlp4 default;`,
			},
			{
				Statement: `create table rlp4_1 partition of rlp4 for values from (20) to (25);`,
			},
			{
				Statement: `create table rlp4_2 partition of rlp4 for values from (25) to (29);`,
			},
			{
				Statement: `create table rlp5 partition of rlp for values from (31) to (maxvalue) partition by range (a);`,
			},
			{
				Statement: `create table rlp5_default partition of rlp5 default;`,
			},
			{
				Statement: `create table rlp5_1 partition of rlp5 for values from (31) to (40);`,
			},
			{
				Statement: `explain (costs off) select * from rlp where a < 1;`,
				Results:   []sql.Row{{`Seq Scan on rlp1 rlp`}, {`Filter: (a < 1)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where 1 > a;	/* commuted */
      QUERY PLAN      
----------------------
 Seq Scan on rlp1 rlp
   Filter: (1 > a)
(2 rows)
explain (costs off) select * from rlp where a <= 1;`,
				Results: []sql.Row{{`Append`}, {`->  Seq Scan on rlp1 rlp_1`}, {`Filter: (a <= 1)`}, {`->  Seq Scan on rlp2 rlp_2`}, {`Filter: (a <= 1)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 1;`,
				Results:   []sql.Row{{`Seq Scan on rlp2 rlp`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 1::bigint;		/* same as above */
         QUERY PLAN          
-----------------------------
 Seq Scan on rlp2 rlp
   Filter: (a = '1'::bigint)
(2 rows)
explain (costs off) select * from rlp where a = 1::numeric;		/* no pruning */
                  QUERY PLAN                   
-----------------------------------------------
 Append
   ->  Seq Scan on rlp1 rlp_1
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp2 rlp_2
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp3abcd rlp_3
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp3efgh rlp_4
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp3nullxy rlp_5
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp3_default rlp_6
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp4_1 rlp_7
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp4_2 rlp_8
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp4_default rlp_9
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp5_1 rlp_10
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp5_default rlp_11
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp_default_10 rlp_12
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp_default_30 rlp_13
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp_default_null rlp_14
         Filter: ((a)::numeric = '1'::numeric)
   ->  Seq Scan on rlp_default_default rlp_15
         Filter: ((a)::numeric = '1'::numeric)
(31 rows)
explain (costs off) select * from rlp where a <= 10;`,
				Results: []sql.Row{{`Append`}, {`->  Seq Scan on rlp1 rlp_1`}, {`Filter: (a <= 10)`}, {`->  Seq Scan on rlp2 rlp_2`}, {`Filter: (a <= 10)`}, {`->  Seq Scan on rlp_default_10 rlp_3`}, {`Filter: (a <= 10)`}, {`->  Seq Scan on rlp_default_default rlp_4`}, {`Filter: (a <= 10)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a > 10;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp3abcd rlp_1`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp3efgh rlp_2`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp3nullxy rlp_3`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp3_default rlp_4`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp4_1 rlp_5`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp4_2 rlp_6`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp4_default rlp_7`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp5_1 rlp_8`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp5_default rlp_9`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp_default_30 rlp_10`}, {`Filter: (a > 10)`}, {`->  Seq Scan on rlp_default_default rlp_11`}, {`Filter: (a > 10)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a < 15;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp1 rlp_1`}, {`Filter: (a < 15)`}, {`->  Seq Scan on rlp2 rlp_2`}, {`Filter: (a < 15)`}, {`->  Seq Scan on rlp_default_10 rlp_3`}, {`Filter: (a < 15)`}, {`->  Seq Scan on rlp_default_default rlp_4`}, {`Filter: (a < 15)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a <= 15;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp1 rlp_1`}, {`Filter: (a <= 15)`}, {`->  Seq Scan on rlp2 rlp_2`}, {`Filter: (a <= 15)`}, {`->  Seq Scan on rlp3abcd rlp_3`}, {`Filter: (a <= 15)`}, {`->  Seq Scan on rlp3efgh rlp_4`}, {`Filter: (a <= 15)`}, {`->  Seq Scan on rlp3nullxy rlp_5`}, {`Filter: (a <= 15)`}, {`->  Seq Scan on rlp3_default rlp_6`}, {`Filter: (a <= 15)`}, {`->  Seq Scan on rlp_default_10 rlp_7`}, {`Filter: (a <= 15)`}, {`->  Seq Scan on rlp_default_default rlp_8`}, {`Filter: (a <= 15)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a > 15 and b = 'ab';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp3abcd rlp_1`}, {`Filter: ((a > 15) AND ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp4_1 rlp_2`}, {`Filter: ((a > 15) AND ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp4_2 rlp_3`}, {`Filter: ((a > 15) AND ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp4_default rlp_4`}, {`Filter: ((a > 15) AND ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp5_1 rlp_5`}, {`Filter: ((a > 15) AND ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp5_default rlp_6`}, {`Filter: ((a > 15) AND ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp_default_30 rlp_7`}, {`Filter: ((a > 15) AND ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp_default_default rlp_8`}, {`Filter: ((a > 15) AND ((b)::text = 'ab'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 16;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp3abcd rlp_1`}, {`Filter: (a = 16)`}, {`->  Seq Scan on rlp3efgh rlp_2`}, {`Filter: (a = 16)`}, {`->  Seq Scan on rlp3nullxy rlp_3`}, {`Filter: (a = 16)`}, {`->  Seq Scan on rlp3_default rlp_4`}, {`Filter: (a = 16)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 16 and b in ('not', 'in', 'here');`,
				Results:   []sql.Row{{`Seq Scan on rlp3_default rlp`}, {`Filter: ((a = 16) AND ((b)::text = ANY ('{not,in,here}'::text[])))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 16 and b < 'ab';`,
				Results:   []sql.Row{{`Seq Scan on rlp3_default rlp`}, {`Filter: (((b)::text < 'ab'::text) AND (a = 16))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 16 and b <= 'ab';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp3abcd rlp_1`}, {`Filter: (((b)::text <= 'ab'::text) AND (a = 16))`}, {`->  Seq Scan on rlp3_default rlp_2`}, {`Filter: (((b)::text <= 'ab'::text) AND (a = 16))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 16 and b is null;`,
				Results:   []sql.Row{{`Seq Scan on rlp3nullxy rlp`}, {`Filter: ((b IS NULL) AND (a = 16))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 16 and b is not null;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp3abcd rlp_1`}, {`Filter: ((b IS NOT NULL) AND (a = 16))`}, {`->  Seq Scan on rlp3efgh rlp_2`}, {`Filter: ((b IS NOT NULL) AND (a = 16))`}, {`->  Seq Scan on rlp3nullxy rlp_3`}, {`Filter: ((b IS NOT NULL) AND (a = 16))`}, {`->  Seq Scan on rlp3_default rlp_4`}, {`Filter: ((b IS NOT NULL) AND (a = 16))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a is null;`,
				Results:   []sql.Row{{`Seq Scan on rlp_default_null rlp`}, {`Filter: (a IS NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a is not null;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp1 rlp_1`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp2 rlp_2`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp3abcd rlp_3`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp3efgh rlp_4`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp3nullxy rlp_5`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp3_default rlp_6`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp4_1 rlp_7`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp4_2 rlp_8`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp4_default rlp_9`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp5_1 rlp_10`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp5_default rlp_11`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp_default_10 rlp_12`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp_default_30 rlp_13`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on rlp_default_default rlp_14`}, {`Filter: (a IS NOT NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a > 30;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp5_1 rlp_1`}, {`Filter: (a > 30)`}, {`->  Seq Scan on rlp5_default rlp_2`}, {`Filter: (a > 30)`}, {`->  Seq Scan on rlp_default_default rlp_3`}, {`Filter: (a > 30)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 30;	/* only default is scanned */
           QUERY PLAN           
--------------------------------
 Seq Scan on rlp_default_30 rlp
   Filter: (a = 30)
(2 rows)
explain (costs off) select * from rlp where a <= 31;`,
				Results: []sql.Row{{`Append`}, {`->  Seq Scan on rlp1 rlp_1`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp2 rlp_2`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp3abcd rlp_3`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp3efgh rlp_4`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp3nullxy rlp_5`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp3_default rlp_6`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp4_1 rlp_7`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp4_2 rlp_8`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp4_default rlp_9`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp5_1 rlp_10`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp_default_10 rlp_11`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp_default_30 rlp_12`}, {`Filter: (a <= 31)`}, {`->  Seq Scan on rlp_default_default rlp_13`}, {`Filter: (a <= 31)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 1 or a = 7;`,
				Results:   []sql.Row{{`Seq Scan on rlp2 rlp`}, {`Filter: ((a = 1) OR (a = 7))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 1 or b = 'ab';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp1 rlp_1`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp2 rlp_2`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp3abcd rlp_3`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp4_1 rlp_4`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp4_2 rlp_5`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp4_default rlp_6`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp5_1 rlp_7`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp5_default rlp_8`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp_default_10 rlp_9`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp_default_30 rlp_10`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp_default_null rlp_11`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}, {`->  Seq Scan on rlp_default_default rlp_12`}, {`Filter: ((a = 1) OR ((b)::text = 'ab'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a > 20 and a < 27;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp4_1 rlp_1`}, {`Filter: ((a > 20) AND (a < 27))`}, {`->  Seq Scan on rlp4_2 rlp_2`}, {`Filter: ((a > 20) AND (a < 27))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 29;`,
				Results:   []sql.Row{{`Seq Scan on rlp4_default rlp`}, {`Filter: (a = 29)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a >= 29;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp4_default rlp_1`}, {`Filter: (a >= 29)`}, {`->  Seq Scan on rlp5_1 rlp_2`}, {`Filter: (a >= 29)`}, {`->  Seq Scan on rlp5_default rlp_3`}, {`Filter: (a >= 29)`}, {`->  Seq Scan on rlp_default_30 rlp_4`}, {`Filter: (a >= 29)`}, {`->  Seq Scan on rlp_default_default rlp_5`}, {`Filter: (a >= 29)`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a < 1 or (a > 20 and a < 25);`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp1 rlp_1`}, {`Filter: ((a < 1) OR ((a > 20) AND (a < 25)))`}, {`->  Seq Scan on rlp4_1 rlp_2`}, {`Filter: ((a < 1) OR ((a > 20) AND (a < 25)))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 20 or a = 40;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp4_1 rlp_1`}, {`Filter: ((a = 20) OR (a = 40))`}, {`->  Seq Scan on rlp5_default rlp_2`}, {`Filter: ((a = 20) OR (a = 40))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp3 where a = 20;   /* empty */
        QUERY PLAN        
--------------------------
 Result
   One-Time Filter: false
(2 rows)
explain (costs off) select * from rlp where a > 1 and a = 10;	/* only default */
            QUERY PLAN            
----------------------------------
 Seq Scan on rlp_default_10 rlp
   Filter: ((a > 1) AND (a = 10))
(2 rows)
explain (costs off) select * from rlp where a > 1 and a >=15;	/* rlp3 onwards, including default */
                  QUERY PLAN                  
----------------------------------------------
 Append
   ->  Seq Scan on rlp3abcd rlp_1
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp3efgh rlp_2
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp3nullxy rlp_3
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp3_default rlp_4
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp4_1 rlp_5
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp4_2 rlp_6
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp4_default rlp_7
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp5_1 rlp_8
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp5_default rlp_9
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp_default_30 rlp_10
         Filter: ((a > 1) AND (a >= 15))
   ->  Seq Scan on rlp_default_default rlp_11
         Filter: ((a > 1) AND (a >= 15))
(23 rows)
explain (costs off) select * from rlp where a = 1 and a = 3;	/* empty */
        QUERY PLAN        
--------------------------
 Result
   One-Time Filter: false
(2 rows)
explain (costs off) select * from rlp where (a = 1 and a = 3) or (a > 1 and a = 15);`,
				Results: []sql.Row{{`Append`}, {`->  Seq Scan on rlp2 rlp_1`}, {`Filter: (((a = 1) AND (a = 3)) OR ((a > 1) AND (a = 15)))`}, {`->  Seq Scan on rlp3abcd rlp_2`}, {`Filter: (((a = 1) AND (a = 3)) OR ((a > 1) AND (a = 15)))`}, {`->  Seq Scan on rlp3efgh rlp_3`}, {`Filter: (((a = 1) AND (a = 3)) OR ((a > 1) AND (a = 15)))`}, {`->  Seq Scan on rlp3nullxy rlp_4`}, {`Filter: (((a = 1) AND (a = 3)) OR ((a > 1) AND (a = 15)))`}, {`->  Seq Scan on rlp3_default rlp_5`}, {`Filter: (((a = 1) AND (a = 3)) OR ((a > 1) AND (a = 15)))`}},
			},
			{
				Statement: `create table mc3p (a int, b int, c int) partition by range (a, abs(b), c);`,
			},
			{
				Statement: `create table mc3p_default partition of mc3p default;`,
			},
			{
				Statement: `create table mc3p0 partition of mc3p for values from (minvalue, minvalue, minvalue) to (1, 1, 1);`,
			},
			{
				Statement: `create table mc3p1 partition of mc3p for values from (1, 1, 1) to (10, 5, 10);`,
			},
			{
				Statement: `create table mc3p2 partition of mc3p for values from (10, 5, 10) to (10, 10, 10);`,
			},
			{
				Statement: `create table mc3p3 partition of mc3p for values from (10, 10, 10) to (10, 10, 20);`,
			},
			{
				Statement: `create table mc3p4 partition of mc3p for values from (10, 10, 20) to (10, maxvalue, maxvalue);`,
			},
			{
				Statement: `create table mc3p5 partition of mc3p for values from (11, 1, 1) to (20, 10, 10);`,
			},
			{
				Statement: `create table mc3p6 partition of mc3p for values from (20, 10, 10) to (20, 20, 20);`,
			},
			{
				Statement: `create table mc3p7 partition of mc3p for values from (20, 20, 20) to (maxvalue, maxvalue, maxvalue);`,
			},
			{
				Statement: `explain (costs off) select * from mc3p where a = 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: (a = 1)`}, {`->  Seq Scan on mc3p_default mc3p_3`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a = 1 and abs(b) < 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: ((a = 1) AND (abs(b) < 1))`}, {`->  Seq Scan on mc3p_default mc3p_2`}, {`Filter: ((a = 1) AND (abs(b) < 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a = 1 and abs(b) = 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: ((a = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: ((a = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p_default mc3p_3`}, {`Filter: ((a = 1) AND (abs(b) = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a = 1 and abs(b) = 1 and c < 8;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: ((c < 8) AND (a = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: ((c < 8) AND (a = 1) AND (abs(b) = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a = 10 and abs(b) between 5 and 35;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p1 mc3p_1`}, {`Filter: ((a = 10) AND (abs(b) >= 5) AND (abs(b) <= 35))`}, {`->  Seq Scan on mc3p2 mc3p_2`}, {`Filter: ((a = 10) AND (abs(b) >= 5) AND (abs(b) <= 35))`}, {`->  Seq Scan on mc3p3 mc3p_3`}, {`Filter: ((a = 10) AND (abs(b) >= 5) AND (abs(b) <= 35))`}, {`->  Seq Scan on mc3p4 mc3p_4`}, {`Filter: ((a = 10) AND (abs(b) >= 5) AND (abs(b) <= 35))`}, {`->  Seq Scan on mc3p_default mc3p_5`}, {`Filter: ((a = 10) AND (abs(b) >= 5) AND (abs(b) <= 35))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a > 10;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p5 mc3p_1`}, {`Filter: (a > 10)`}, {`->  Seq Scan on mc3p6 mc3p_2`}, {`Filter: (a > 10)`}, {`->  Seq Scan on mc3p7 mc3p_3`}, {`Filter: (a > 10)`}, {`->  Seq Scan on mc3p_default mc3p_4`}, {`Filter: (a > 10)`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a >= 10;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p1 mc3p_1`}, {`Filter: (a >= 10)`}, {`->  Seq Scan on mc3p2 mc3p_2`}, {`Filter: (a >= 10)`}, {`->  Seq Scan on mc3p3 mc3p_3`}, {`Filter: (a >= 10)`}, {`->  Seq Scan on mc3p4 mc3p_4`}, {`Filter: (a >= 10)`}, {`->  Seq Scan on mc3p5 mc3p_5`}, {`Filter: (a >= 10)`}, {`->  Seq Scan on mc3p6 mc3p_6`}, {`Filter: (a >= 10)`}, {`->  Seq Scan on mc3p7 mc3p_7`}, {`Filter: (a >= 10)`}, {`->  Seq Scan on mc3p_default mc3p_8`}, {`Filter: (a >= 10)`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a < 10;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: (a < 10)`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: (a < 10)`}, {`->  Seq Scan on mc3p_default mc3p_3`}, {`Filter: (a < 10)`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a <= 10 and abs(b) < 10;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: ((a <= 10) AND (abs(b) < 10))`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: ((a <= 10) AND (abs(b) < 10))`}, {`->  Seq Scan on mc3p2 mc3p_3`}, {`Filter: ((a <= 10) AND (abs(b) < 10))`}, {`->  Seq Scan on mc3p_default mc3p_4`}, {`Filter: ((a <= 10) AND (abs(b) < 10))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a = 11 and abs(b) = 0;`,
				Results:   []sql.Row{{`Seq Scan on mc3p_default mc3p`}, {`Filter: ((a = 11) AND (abs(b) = 0))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a = 20 and abs(b) = 10 and c = 100;`,
				Results:   []sql.Row{{`Seq Scan on mc3p6 mc3p`}, {`Filter: ((a = 20) AND (c = 100) AND (abs(b) = 10))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a > 20;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p7 mc3p_1`}, {`Filter: (a > 20)`}, {`->  Seq Scan on mc3p_default mc3p_2`}, {`Filter: (a > 20)`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a >= 20;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p5 mc3p_1`}, {`Filter: (a >= 20)`}, {`->  Seq Scan on mc3p6 mc3p_2`}, {`Filter: (a >= 20)`}, {`->  Seq Scan on mc3p7 mc3p_3`}, {`Filter: (a >= 20)`}, {`->  Seq Scan on mc3p_default mc3p_4`}, {`Filter: (a >= 20)`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where (a = 1 and abs(b) = 1 and c = 1) or (a = 10 and abs(b) = 5 and c = 10) or (a > 11 and a < 20);`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p1 mc3p_1`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)))`}, {`->  Seq Scan on mc3p2 mc3p_2`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)))`}, {`->  Seq Scan on mc3p5 mc3p_3`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)))`}, {`->  Seq Scan on mc3p_default mc3p_4`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where (a = 1 and abs(b) = 1 and c = 1) or (a = 10 and abs(b) = 5 and c = 10) or (a > 11 and a < 20) or a < 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1))`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1))`}, {`->  Seq Scan on mc3p2 mc3p_3`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1))`}, {`->  Seq Scan on mc3p5 mc3p_4`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1))`}, {`->  Seq Scan on mc3p_default mc3p_5`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where (a = 1 and abs(b) = 1 and c = 1) or (a = 10 and abs(b) = 5 and c = 10) or (a > 11 and a < 20) or a < 1 or a = 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1) OR (a = 1))`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1) OR (a = 1))`}, {`->  Seq Scan on mc3p2 mc3p_3`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1) OR (a = 1))`}, {`->  Seq Scan on mc3p5 mc3p_4`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1) OR (a = 1))`}, {`->  Seq Scan on mc3p_default mc3p_5`}, {`Filter: (((a = 1) AND (abs(b) = 1) AND (c = 1)) OR ((a = 10) AND (abs(b) = 5) AND (c = 10)) OR ((a > 11) AND (a < 20)) OR (a < 1) OR (a = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where a = 1 or abs(b) = 1 or c = 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}, {`->  Seq Scan on mc3p2 mc3p_3`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}, {`->  Seq Scan on mc3p3 mc3p_4`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}, {`->  Seq Scan on mc3p4 mc3p_5`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}, {`->  Seq Scan on mc3p5 mc3p_6`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}, {`->  Seq Scan on mc3p6 mc3p_7`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}, {`->  Seq Scan on mc3p7 mc3p_8`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}, {`->  Seq Scan on mc3p_default mc3p_9`}, {`Filter: ((a = 1) OR (abs(b) = 1) OR (c = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where (a = 1 and abs(b) = 1) or (a = 10 and abs(b) = 10);`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 10)))`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 10)))`}, {`->  Seq Scan on mc3p2 mc3p_3`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 10)))`}, {`->  Seq Scan on mc3p3 mc3p_4`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 10)))`}, {`->  Seq Scan on mc3p4 mc3p_5`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 10)))`}, {`->  Seq Scan on mc3p_default mc3p_6`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 10)))`}},
			},
			{
				Statement: `explain (costs off) select * from mc3p where (a = 1 and abs(b) = 1) or (a = 10 and abs(b) = 9);`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc3p0 mc3p_1`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 9)))`}, {`->  Seq Scan on mc3p1 mc3p_2`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 9)))`}, {`->  Seq Scan on mc3p2 mc3p_3`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 9)))`}, {`->  Seq Scan on mc3p_default mc3p_4`}, {`Filter: (((a = 1) AND (abs(b) = 1)) OR ((a = 10) AND (abs(b) = 9)))`}},
			},
			{
				Statement: `create table mc2p (a int, b int) partition by range (a, b);`,
			},
			{
				Statement: `create table mc2p_default partition of mc2p default;`,
			},
			{
				Statement: `create table mc2p0 partition of mc2p for values from (minvalue, minvalue) to (1, minvalue);`,
			},
			{
				Statement: `create table mc2p1 partition of mc2p for values from (1, minvalue) to (1, 1);`,
			},
			{
				Statement: `create table mc2p2 partition of mc2p for values from (1, 1) to (2, minvalue);`,
			},
			{
				Statement: `create table mc2p3 partition of mc2p for values from (2, minvalue) to (2, 1);`,
			},
			{
				Statement: `create table mc2p4 partition of mc2p for values from (2, 1) to (2, maxvalue);`,
			},
			{
				Statement: `create table mc2p5 partition of mc2p for values from (2, maxvalue) to (maxvalue, maxvalue);`,
			},
			{
				Statement: `explain (costs off) select * from mc2p where a < 2;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc2p0 mc2p_1`}, {`Filter: (a < 2)`}, {`->  Seq Scan on mc2p1 mc2p_2`}, {`Filter: (a < 2)`}, {`->  Seq Scan on mc2p2 mc2p_3`}, {`Filter: (a < 2)`}, {`->  Seq Scan on mc2p_default mc2p_4`}, {`Filter: (a < 2)`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p where a = 2 and b < 1;`,
				Results:   []sql.Row{{`Seq Scan on mc2p3 mc2p`}, {`Filter: ((b < 1) AND (a = 2))`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p where a > 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mc2p2 mc2p_1`}, {`Filter: (a > 1)`}, {`->  Seq Scan on mc2p3 mc2p_2`}, {`Filter: (a > 1)`}, {`->  Seq Scan on mc2p4 mc2p_3`}, {`Filter: (a > 1)`}, {`->  Seq Scan on mc2p5 mc2p_4`}, {`Filter: (a > 1)`}, {`->  Seq Scan on mc2p_default mc2p_5`}, {`Filter: (a > 1)`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p where a = 1 and b > 1;`,
				Results:   []sql.Row{{`Seq Scan on mc2p2 mc2p`}, {`Filter: ((b > 1) AND (a = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p where a = 1 and b is null;`,
				Results:   []sql.Row{{`Seq Scan on mc2p_default mc2p`}, {`Filter: ((b IS NULL) AND (a = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p where a is null and b is null;`,
				Results:   []sql.Row{{`Seq Scan on mc2p_default mc2p`}, {`Filter: ((a IS NULL) AND (b IS NULL))`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p where a is null and b = 1;`,
				Results:   []sql.Row{{`Seq Scan on mc2p_default mc2p`}, {`Filter: ((a IS NULL) AND (b = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p where a is null;`,
				Results:   []sql.Row{{`Seq Scan on mc2p_default mc2p`}, {`Filter: (a IS NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p where b is null;`,
				Results:   []sql.Row{{`Seq Scan on mc2p_default mc2p`}, {`Filter: (b IS NULL)`}},
			},
			{
				Statement: `create table boolpart (a bool) partition by list (a);`,
			},
			{
				Statement: `create table boolpart_default partition of boolpart default;`,
			},
			{
				Statement: `create table boolpart_t partition of boolpart for values in ('true');`,
			},
			{
				Statement: `create table boolpart_f partition of boolpart for values in ('false');`,
			},
			{
				Statement: `insert into boolpart values (true), (false), (null);`,
			},
			{
				Statement: `explain (costs off) select * from boolpart where a in (true, false);`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on boolpart_f boolpart_1`}, {`Filter: (a = ANY ('{t,f}'::boolean[]))`}, {`->  Seq Scan on boolpart_t boolpart_2`}, {`Filter: (a = ANY ('{t,f}'::boolean[]))`}},
			},
			{
				Statement: `explain (costs off) select * from boolpart where a = false;`,
				Results:   []sql.Row{{`Seq Scan on boolpart_f boolpart`}, {`Filter: (NOT a)`}},
			},
			{
				Statement: `explain (costs off) select * from boolpart where not a = false;`,
				Results:   []sql.Row{{`Seq Scan on boolpart_t boolpart`}, {`Filter: a`}},
			},
			{
				Statement: `explain (costs off) select * from boolpart where a is true or a is not true;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on boolpart_f boolpart_1`}, {`Filter: ((a IS TRUE) OR (a IS NOT TRUE))`}, {`->  Seq Scan on boolpart_t boolpart_2`}, {`Filter: ((a IS TRUE) OR (a IS NOT TRUE))`}, {`->  Seq Scan on boolpart_default boolpart_3`}, {`Filter: ((a IS TRUE) OR (a IS NOT TRUE))`}},
			},
			{
				Statement: `explain (costs off) select * from boolpart where a is not true;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on boolpart_f boolpart_1`}, {`Filter: (a IS NOT TRUE)`}, {`->  Seq Scan on boolpart_default boolpart_2`}, {`Filter: (a IS NOT TRUE)`}},
			},
			{
				Statement: `explain (costs off) select * from boolpart where a is not true and a is not false;`,
				Results:   []sql.Row{{`Seq Scan on boolpart_default boolpart`}, {`Filter: ((a IS NOT TRUE) AND (a IS NOT FALSE))`}},
			},
			{
				Statement: `explain (costs off) select * from boolpart where a is unknown;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on boolpart_f boolpart_1`}, {`Filter: (a IS UNKNOWN)`}, {`->  Seq Scan on boolpart_t boolpart_2`}, {`Filter: (a IS UNKNOWN)`}, {`->  Seq Scan on boolpart_default boolpart_3`}, {`Filter: (a IS UNKNOWN)`}},
			},
			{
				Statement: `explain (costs off) select * from boolpart where a is not unknown;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on boolpart_f boolpart_1`}, {`Filter: (a IS NOT UNKNOWN)`}, {`->  Seq Scan on boolpart_t boolpart_2`}, {`Filter: (a IS NOT UNKNOWN)`}, {`->  Seq Scan on boolpart_default boolpart_3`}, {`Filter: (a IS NOT UNKNOWN)`}},
			},
			{
				Statement: `select * from boolpart where a in (true, false);`,
				Results:   []sql.Row{{false}, {true}},
			},
			{
				Statement: `select * from boolpart where a = false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select * from boolpart where not a = false;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select * from boolpart where a is true or a is not true;`,
				Results:   []sql.Row{{false}, {true}, {``}},
			},
			{
				Statement: `select * from boolpart where a is not true;`,
				Results:   []sql.Row{{false}, {``}},
			},
			{
				Statement: `select * from boolpart where a is not true and a is not false;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select * from boolpart where a is unknown;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select * from boolpart where a is not unknown;`,
				Results:   []sql.Row{{false}, {true}},
			},
			{
				Statement: `create table iboolpart (a bool) partition by list ((not a));`,
			},
			{
				Statement: `create table iboolpart_default partition of iboolpart default;`,
			},
			{
				Statement: `create table iboolpart_f partition of iboolpart for values in ('true');`,
			},
			{
				Statement: `create table iboolpart_t partition of iboolpart for values in ('false');`,
			},
			{
				Statement: `insert into iboolpart values (true), (false), (null);`,
			},
			{
				Statement: `explain (costs off) select * from iboolpart where a in (true, false);`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on iboolpart_t iboolpart_1`}, {`Filter: (a = ANY ('{t,f}'::boolean[]))`}, {`->  Seq Scan on iboolpart_f iboolpart_2`}, {`Filter: (a = ANY ('{t,f}'::boolean[]))`}, {`->  Seq Scan on iboolpart_default iboolpart_3`}, {`Filter: (a = ANY ('{t,f}'::boolean[]))`}},
			},
			{
				Statement: `explain (costs off) select * from iboolpart where a = false;`,
				Results:   []sql.Row{{`Seq Scan on iboolpart_f iboolpart`}, {`Filter: (NOT a)`}},
			},
			{
				Statement: `explain (costs off) select * from iboolpart where not a = false;`,
				Results:   []sql.Row{{`Seq Scan on iboolpart_t iboolpart`}, {`Filter: a`}},
			},
			{
				Statement: `explain (costs off) select * from iboolpart where a is true or a is not true;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on iboolpart_t iboolpart_1`}, {`Filter: ((a IS TRUE) OR (a IS NOT TRUE))`}, {`->  Seq Scan on iboolpart_f iboolpart_2`}, {`Filter: ((a IS TRUE) OR (a IS NOT TRUE))`}, {`->  Seq Scan on iboolpart_default iboolpart_3`}, {`Filter: ((a IS TRUE) OR (a IS NOT TRUE))`}},
			},
			{
				Statement: `explain (costs off) select * from iboolpart where a is not true;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on iboolpart_t iboolpart_1`}, {`Filter: (a IS NOT TRUE)`}, {`->  Seq Scan on iboolpart_f iboolpart_2`}, {`Filter: (a IS NOT TRUE)`}, {`->  Seq Scan on iboolpart_default iboolpart_3`}, {`Filter: (a IS NOT TRUE)`}},
			},
			{
				Statement: `explain (costs off) select * from iboolpart where a is not true and a is not false;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on iboolpart_t iboolpart_1`}, {`Filter: ((a IS NOT TRUE) AND (a IS NOT FALSE))`}, {`->  Seq Scan on iboolpart_f iboolpart_2`}, {`Filter: ((a IS NOT TRUE) AND (a IS NOT FALSE))`}, {`->  Seq Scan on iboolpart_default iboolpart_3`}, {`Filter: ((a IS NOT TRUE) AND (a IS NOT FALSE))`}},
			},
			{
				Statement: `explain (costs off) select * from iboolpart where a is unknown;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on iboolpart_t iboolpart_1`}, {`Filter: (a IS UNKNOWN)`}, {`->  Seq Scan on iboolpart_f iboolpart_2`}, {`Filter: (a IS UNKNOWN)`}, {`->  Seq Scan on iboolpart_default iboolpart_3`}, {`Filter: (a IS UNKNOWN)`}},
			},
			{
				Statement: `explain (costs off) select * from iboolpart where a is not unknown;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on iboolpart_t iboolpart_1`}, {`Filter: (a IS NOT UNKNOWN)`}, {`->  Seq Scan on iboolpart_f iboolpart_2`}, {`Filter: (a IS NOT UNKNOWN)`}, {`->  Seq Scan on iboolpart_default iboolpart_3`}, {`Filter: (a IS NOT UNKNOWN)`}},
			},
			{
				Statement: `select * from iboolpart where a in (true, false);`,
				Results:   []sql.Row{{true}, {false}},
			},
			{
				Statement: `select * from iboolpart where a = false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select * from iboolpart where not a = false;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select * from iboolpart where a is true or a is not true;`,
				Results:   []sql.Row{{true}, {false}, {``}},
			},
			{
				Statement: `select * from iboolpart where a is not true;`,
				Results:   []sql.Row{{false}, {``}},
			},
			{
				Statement: `select * from iboolpart where a is not true and a is not false;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select * from iboolpart where a is unknown;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select * from iboolpart where a is not unknown;`,
				Results:   []sql.Row{{true}, {false}},
			},
			{
				Statement: `create table boolrangep (a bool, b bool, c int) partition by range (a,b,c);`,
			},
			{
				Statement: `create table boolrangep_tf partition of boolrangep for values from ('true', 'false', 0) to ('true', 'false', 100);`,
			},
			{
				Statement: `create table boolrangep_ft partition of boolrangep for values from ('false', 'true', 0) to ('false', 'true', 100);`,
			},
			{
				Statement: `create table boolrangep_ff1 partition of boolrangep for values from ('false', 'false', 0) to ('false', 'false', 50);`,
			},
			{
				Statement: `create table boolrangep_ff2 partition of boolrangep for values from ('false', 'false', 50) to ('false', 'false', 100);`,
			},
			{
				Statement: `explain (costs off)  select * from boolrangep where not a and not b and c = 25;`,
				Results:   []sql.Row{{`Seq Scan on boolrangep_ff1 boolrangep`}, {`Filter: ((NOT a) AND (NOT b) AND (c = 25))`}},
			},
			{
				Statement: `create table coercepart (a varchar) partition by list (a);`,
			},
			{
				Statement: `create table coercepart_ab partition of coercepart for values in ('ab');`,
			},
			{
				Statement: `create table coercepart_bc partition of coercepart for values in ('bc');`,
			},
			{
				Statement: `create table coercepart_cd partition of coercepart for values in ('cd');`,
			},
			{
				Statement: `explain (costs off) select * from coercepart where a in ('ab', to_char(125, '999'));`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coercepart_ab coercepart_1`}, {`Filter: ((a)::text = ANY ((ARRAY['ab'::character varying, (to_char(125, '999'::text))::character varying])::text[]))`}, {`->  Seq Scan on coercepart_bc coercepart_2`}, {`Filter: ((a)::text = ANY ((ARRAY['ab'::character varying, (to_char(125, '999'::text))::character varying])::text[]))`}, {`->  Seq Scan on coercepart_cd coercepart_3`}, {`Filter: ((a)::text = ANY ((ARRAY['ab'::character varying, (to_char(125, '999'::text))::character varying])::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a ~ any ('{ab}');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coercepart_ab coercepart_1`}, {`Filter: ((a)::text ~ ANY ('{ab}'::text[]))`}, {`->  Seq Scan on coercepart_bc coercepart_2`}, {`Filter: ((a)::text ~ ANY ('{ab}'::text[]))`}, {`->  Seq Scan on coercepart_cd coercepart_3`}, {`Filter: ((a)::text ~ ANY ('{ab}'::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a !~ all ('{ab}');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coercepart_ab coercepart_1`}, {`Filter: ((a)::text !~ ALL ('{ab}'::text[]))`}, {`->  Seq Scan on coercepart_bc coercepart_2`}, {`Filter: ((a)::text !~ ALL ('{ab}'::text[]))`}, {`->  Seq Scan on coercepart_cd coercepart_3`}, {`Filter: ((a)::text !~ ALL ('{ab}'::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a ~ any ('{ab,bc}');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coercepart_ab coercepart_1`}, {`Filter: ((a)::text ~ ANY ('{ab,bc}'::text[]))`}, {`->  Seq Scan on coercepart_bc coercepart_2`}, {`Filter: ((a)::text ~ ANY ('{ab,bc}'::text[]))`}, {`->  Seq Scan on coercepart_cd coercepart_3`}, {`Filter: ((a)::text ~ ANY ('{ab,bc}'::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a !~ all ('{ab,bc}');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coercepart_ab coercepart_1`}, {`Filter: ((a)::text !~ ALL ('{ab,bc}'::text[]))`}, {`->  Seq Scan on coercepart_bc coercepart_2`}, {`Filter: ((a)::text !~ ALL ('{ab,bc}'::text[]))`}, {`->  Seq Scan on coercepart_cd coercepart_3`}, {`Filter: ((a)::text !~ ALL ('{ab,bc}'::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a = any ('{ab,bc}');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coercepart_ab coercepart_1`}, {`Filter: ((a)::text = ANY ('{ab,bc}'::text[]))`}, {`->  Seq Scan on coercepart_bc coercepart_2`}, {`Filter: ((a)::text = ANY ('{ab,bc}'::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a = any ('{ab,null}');`,
				Results:   []sql.Row{{`Seq Scan on coercepart_ab coercepart`}, {`Filter: ((a)::text = ANY ('{ab,NULL}'::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a = any (null::text[]);`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a = all ('{ab}');`,
				Results:   []sql.Row{{`Seq Scan on coercepart_ab coercepart`}, {`Filter: ((a)::text = ALL ('{ab}'::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a = all ('{ab,bc}');`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a = all ('{ab,null}');`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off) select * from coercepart where a = all (null::text[]);`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `drop table coercepart;`,
			},
			{
				Statement: `CREATE TABLE part (a INT, b INT) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE part_p1 PARTITION OF part FOR VALUES IN (-2,-1,0,1,2);`,
			},
			{
				Statement: `CREATE TABLE part_p2 PARTITION OF part DEFAULT PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE part_p2_p1 PARTITION OF part_p2 DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE part_rev (b INT, c INT, a INT);`,
			},
			{
				Statement:   `ALTER TABLE part ATTACH PARTITION part_rev FOR VALUES IN (3);  -- fail`,
				ErrorString: `table "part_rev" contains column "c" not found in parent "part"`,
			},
			{
				Statement: `ALTER TABLE part_rev DROP COLUMN c;`,
			},
			{
				Statement: `ALTER TABLE part ATTACH PARTITION part_rev FOR VALUES IN (3);  -- now it's ok`,
			},
			{
				Statement: `INSERT INTO part VALUES (-1,-1), (1,1), (2,NULL), (NULL,-2),(NULL,NULL);`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT tableoid::regclass as part, a, b FROM part WHERE a IS NULL ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{`Sort`}, {`Sort Key: ((part.tableoid)::regclass), part.a, part.b`}, {`->  Seq Scan on part_p2_p1 part`}, {`Filter: (a IS NULL)`}},
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF) SELECT * FROM part p(x) ORDER BY x;`,
				Results:   []sql.Row{{`Sort`}, {`Output: p.x, p.b`}, {`Sort Key: p.x`}, {`->  Append`}, {`->  Seq Scan on public.part_p1 p_1`}, {`Output: p_1.x, p_1.b`}, {`->  Seq Scan on public.part_rev p_2`}, {`Output: p_2.x, p_2.b`}, {`->  Seq Scan on public.part_p2_p1 p_3`}, {`Output: p_3.x, p_3.b`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p t1, lateral (select count(*) from mc3p t2 where t2.a = t1.b and abs(t2.b) = 1 and t2.c = 1) s where t1.a = 1;`,
				Results:   []sql.Row{{`Nested Loop`}, {`->  Append`}, {`->  Seq Scan on mc2p1 t1_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on mc2p2 t1_2`}, {`Filter: (a = 1)`}, {`->  Seq Scan on mc2p_default t1_3`}, {`Filter: (a = 1)`}, {`->  Aggregate`}, {`->  Append`}, {`->  Seq Scan on mc3p0 t2_1`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p1 t2_2`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p2 t2_3`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p3 t2_4`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p4 t2_5`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p5 t2_6`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p6 t2_7`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p7 t2_8`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p_default t2_9`}, {`Filter: ((a = t1.b) AND (c = 1) AND (abs(b) = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p t1, lateral (select count(*) from mc3p t2 where t2.c = t1.b and abs(t2.b) = 1 and t2.a = 1) s where t1.a = 1;`,
				Results:   []sql.Row{{`Nested Loop`}, {`->  Append`}, {`->  Seq Scan on mc2p1 t1_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on mc2p2 t1_2`}, {`Filter: (a = 1)`}, {`->  Seq Scan on mc2p_default t1_3`}, {`Filter: (a = 1)`}, {`->  Aggregate`}, {`->  Append`}, {`->  Seq Scan on mc3p0 t2_1`}, {`Filter: ((c = t1.b) AND (a = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p1 t2_2`}, {`Filter: ((c = t1.b) AND (a = 1) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p_default t2_3`}, {`Filter: ((c = t1.b) AND (a = 1) AND (abs(b) = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from mc2p t1, lateral (select count(*) from mc3p t2 where t2.a = 1 and abs(t2.b) = 1 and t2.c = 1) s where t1.a = 1;`,
				Results:   []sql.Row{{`Nested Loop`}, {`->  Aggregate`}, {`->  Seq Scan on mc3p1 t2`}, {`Filter: ((a = 1) AND (c = 1) AND (abs(b) = 1))`}, {`->  Append`}, {`->  Seq Scan on mc2p1 t1_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on mc2p2 t1_2`}, {`Filter: (a = 1)`}, {`->  Seq Scan on mc2p_default t1_3`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `create table rp (a int) partition by range (a);`,
			},
			{
				Statement: `create table rp0 partition of rp for values from (minvalue) to (1);`,
			},
			{
				Statement: `create table rp1 partition of rp for values from (1) to (2);`,
			},
			{
				Statement: `create table rp2 partition of rp for values from (2) to (maxvalue);`,
			},
			{
				Statement: `explain (costs off) select * from rp where a <> 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rp0 rp_1`}, {`Filter: (a <> 1)`}, {`->  Seq Scan on rp1 rp_2`}, {`Filter: (a <> 1)`}, {`->  Seq Scan on rp2 rp_3`}, {`Filter: (a <> 1)`}},
			},
			{
				Statement: `explain (costs off) select * from rp where a <> 1 and a <> 2;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rp0 rp_1`}, {`Filter: ((a <> 1) AND (a <> 2))`}, {`->  Seq Scan on rp1 rp_2`}, {`Filter: ((a <> 1) AND (a <> 2))`}, {`->  Seq Scan on rp2 rp_3`}, {`Filter: ((a <> 1) AND (a <> 2))`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a <> 'a';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_ad lp_1`}, {`Filter: (a <> 'a'::bpchar)`}, {`->  Seq Scan on lp_bc lp_2`}, {`Filter: (a <> 'a'::bpchar)`}, {`->  Seq Scan on lp_ef lp_3`}, {`Filter: (a <> 'a'::bpchar)`}, {`->  Seq Scan on lp_g lp_4`}, {`Filter: (a <> 'a'::bpchar)`}, {`->  Seq Scan on lp_default lp_5`}, {`Filter: (a <> 'a'::bpchar)`}},
			},
			{
				Statement: `explain (costs off) select * from lp where a <> 'a' and a is null;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off) select * from lp where (a <> 'a' and a <> 'd') or a is null;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on lp_bc lp_1`}, {`Filter: (((a <> 'a'::bpchar) AND (a <> 'd'::bpchar)) OR (a IS NULL))`}, {`->  Seq Scan on lp_ef lp_2`}, {`Filter: (((a <> 'a'::bpchar) AND (a <> 'd'::bpchar)) OR (a IS NULL))`}, {`->  Seq Scan on lp_g lp_3`}, {`Filter: (((a <> 'a'::bpchar) AND (a <> 'd'::bpchar)) OR (a IS NULL))`}, {`->  Seq Scan on lp_null lp_4`}, {`Filter: (((a <> 'a'::bpchar) AND (a <> 'd'::bpchar)) OR (a IS NULL))`}, {`->  Seq Scan on lp_default lp_5`}, {`Filter: (((a <> 'a'::bpchar) AND (a <> 'd'::bpchar)) OR (a IS NULL))`}},
			},
			{
				Statement: `explain (costs off) select * from rlp where a = 15 and b <> 'ab' and b <> 'cd' and b <> 'xy' and b is not null;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on rlp3efgh rlp_1`}, {`Filter: ((b IS NOT NULL) AND ((b)::text <> 'ab'::text) AND ((b)::text <> 'cd'::text) AND ((b)::text <> 'xy'::text) AND (a = 15))`}, {`->  Seq Scan on rlp3_default rlp_2`}, {`Filter: ((b IS NOT NULL) AND ((b)::text <> 'ab'::text) AND ((b)::text <> 'cd'::text) AND ((b)::text <> 'xy'::text) AND (a = 15))`}},
			},
			{
				Statement: `create table coll_pruning_multi (a text) partition by range (substr(a, 1) collate "POSIX", substr(a, 1) collate "C");`,
			},
			{
				Statement: `create table coll_pruning_multi1 partition of coll_pruning_multi for values from ('a', 'a') to ('a', 'e');`,
			},
			{
				Statement: `create table coll_pruning_multi2 partition of coll_pruning_multi for values from ('a', 'e') to ('a', 'z');`,
			},
			{
				Statement: `create table coll_pruning_multi3 partition of coll_pruning_multi for values from ('b', 'a') to ('b', 'e');`,
			},
			{
				Statement: `explain (costs off) select * from coll_pruning_multi where substr(a, 1) = 'e' collate "C";`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coll_pruning_multi1 coll_pruning_multi_1`}, {`Filter: (substr(a, 1) = 'e'::text COLLATE "C")`}, {`->  Seq Scan on coll_pruning_multi2 coll_pruning_multi_2`}, {`Filter: (substr(a, 1) = 'e'::text COLLATE "C")`}, {`->  Seq Scan on coll_pruning_multi3 coll_pruning_multi_3`}, {`Filter: (substr(a, 1) = 'e'::text COLLATE "C")`}},
			},
			{
				Statement: `explain (costs off) select * from coll_pruning_multi where substr(a, 1) = 'a' collate "POSIX";`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on coll_pruning_multi1 coll_pruning_multi_1`}, {`Filter: (substr(a, 1) = 'a'::text COLLATE "POSIX")`}, {`->  Seq Scan on coll_pruning_multi2 coll_pruning_multi_2`}, {`Filter: (substr(a, 1) = 'a'::text COLLATE "POSIX")`}},
			},
			{
				Statement: `explain (costs off) select * from coll_pruning_multi where substr(a, 1) = 'e' collate "C" and substr(a, 1) = 'a' collate "POSIX";`,
				Results:   []sql.Row{{`Seq Scan on coll_pruning_multi2 coll_pruning_multi`}, {`Filter: ((substr(a, 1) = 'e'::text COLLATE "C") AND (substr(a, 1) = 'a'::text COLLATE "POSIX"))`}},
			},
			{
				Statement: `create table like_op_noprune (a text) partition by list (a);`,
			},
			{
				Statement: `create table like_op_noprune1 partition of like_op_noprune for values in ('ABC');`,
			},
			{
				Statement: `create table like_op_noprune2 partition of like_op_noprune for values in ('BCD');`,
			},
			{
				Statement: `explain (costs off) select * from like_op_noprune where a like '%BC';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on like_op_noprune1 like_op_noprune_1`}, {`Filter: (a ~~ '%BC'::text)`}, {`->  Seq Scan on like_op_noprune2 like_op_noprune_2`}, {`Filter: (a ~~ '%BC'::text)`}},
			},
			{
				Statement: `create table lparted_by_int2 (a smallint) partition by list (a);`,
			},
			{
				Statement: `create table lparted_by_int2_1 partition of lparted_by_int2 for values in (1);`,
			},
			{
				Statement: `create table lparted_by_int2_16384 partition of lparted_by_int2 for values in (16384);`,
			},
			{
				Statement: `explain (costs off) select * from lparted_by_int2 where a = 100000000000000;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `create table rparted_by_int2 (a smallint) partition by range (a);`,
			},
			{
				Statement: `create table rparted_by_int2_1 partition of rparted_by_int2 for values from (1) to (10);`,
			},
			{
				Statement: `create table rparted_by_int2_16384 partition of rparted_by_int2 for values from (10) to (16384);`,
			},
			{
				Statement: `explain (costs off) select * from rparted_by_int2 where a > 100000000000000;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `create table rparted_by_int2_maxvalue partition of rparted_by_int2 for values from (16384) to (maxvalue);`,
			},
			{
				Statement: `explain (costs off) select * from rparted_by_int2 where a > 100000000000000;`,
				Results:   []sql.Row{{`Seq Scan on rparted_by_int2_maxvalue rparted_by_int2`}, {`Filter: (a > '100000000000000'::bigint)`}},
			},
			{
				Statement: `drop table lp, coll_pruning, rlp, mc3p, mc2p, boolpart, iboolpart, boolrangep, rp, coll_pruning_multi, like_op_noprune, lparted_by_int2, rparted_by_int2;`,
			},
			{
				Statement: `create table hp (a int, b text, c int)
  partition by hash (a part_test_int4_ops, b part_test_text_ops);`,
			},
			{
				Statement: `create table hp0 partition of hp for values with (modulus 4, remainder 0);`,
			},
			{
				Statement: `create table hp3 partition of hp for values with (modulus 4, remainder 3);`,
			},
			{
				Statement: `create table hp1 partition of hp for values with (modulus 4, remainder 1);`,
			},
			{
				Statement: `create table hp2 partition of hp for values with (modulus 4, remainder 2);`,
			},
			{
				Statement: `insert into hp values (null, null, 0);`,
			},
			{
				Statement: `insert into hp values (1, null, 1);`,
			},
			{
				Statement: `insert into hp values (1, 'xxx', 2);`,
			},
			{
				Statement: `insert into hp values (null, 'xxx', 3);`,
			},
			{
				Statement: `insert into hp values (2, 'xxx', 4);`,
			},
			{
				Statement: `insert into hp values (1, 'abcde', 5);`,
			},
			{
				Statement: `select tableoid::regclass, * from hp order by c;`,
				Results:   []sql.Row{{`hp0`, ``, ``, 0}, {`hp1`, 1, ``, 1}, {`hp0`, 1, `xxx`, 2}, {`hp2`, ``, `xxx`, 3}, {`hp3`, 2, `xxx`, 4}, {`hp2`, 1, `abcde`, 5}},
			},
			{
				Statement: `explain (costs off) select * from hp where a = 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on hp0 hp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on hp1 hp_2`}, {`Filter: (a = 1)`}, {`->  Seq Scan on hp2 hp_3`}, {`Filter: (a = 1)`}, {`->  Seq Scan on hp3 hp_4`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) select * from hp where b = 'xxx';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on hp0 hp_1`}, {`Filter: (b = 'xxx'::text)`}, {`->  Seq Scan on hp1 hp_2`}, {`Filter: (b = 'xxx'::text)`}, {`->  Seq Scan on hp2 hp_3`}, {`Filter: (b = 'xxx'::text)`}, {`->  Seq Scan on hp3 hp_4`}, {`Filter: (b = 'xxx'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a is null;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on hp0 hp_1`}, {`Filter: (a IS NULL)`}, {`->  Seq Scan on hp1 hp_2`}, {`Filter: (a IS NULL)`}, {`->  Seq Scan on hp2 hp_3`}, {`Filter: (a IS NULL)`}, {`->  Seq Scan on hp3 hp_4`}, {`Filter: (a IS NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from hp where b is null;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on hp0 hp_1`}, {`Filter: (b IS NULL)`}, {`->  Seq Scan on hp1 hp_2`}, {`Filter: (b IS NULL)`}, {`->  Seq Scan on hp2 hp_3`}, {`Filter: (b IS NULL)`}, {`->  Seq Scan on hp3 hp_4`}, {`Filter: (b IS NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a < 1 and b = 'xxx';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on hp0 hp_1`}, {`Filter: ((a < 1) AND (b = 'xxx'::text))`}, {`->  Seq Scan on hp1 hp_2`}, {`Filter: ((a < 1) AND (b = 'xxx'::text))`}, {`->  Seq Scan on hp2 hp_3`}, {`Filter: ((a < 1) AND (b = 'xxx'::text))`}, {`->  Seq Scan on hp3 hp_4`}, {`Filter: ((a < 1) AND (b = 'xxx'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a <> 1 and b = 'yyy';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on hp0 hp_1`}, {`Filter: ((a <> 1) AND (b = 'yyy'::text))`}, {`->  Seq Scan on hp1 hp_2`}, {`Filter: ((a <> 1) AND (b = 'yyy'::text))`}, {`->  Seq Scan on hp2 hp_3`}, {`Filter: ((a <> 1) AND (b = 'yyy'::text))`}, {`->  Seq Scan on hp3 hp_4`}, {`Filter: ((a <> 1) AND (b = 'yyy'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a <> 1 and b <> 'xxx';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on hp0 hp_1`}, {`Filter: ((a <> 1) AND (b <> 'xxx'::text))`}, {`->  Seq Scan on hp1 hp_2`}, {`Filter: ((a <> 1) AND (b <> 'xxx'::text))`}, {`->  Seq Scan on hp2 hp_3`}, {`Filter: ((a <> 1) AND (b <> 'xxx'::text))`}, {`->  Seq Scan on hp3 hp_4`}, {`Filter: ((a <> 1) AND (b <> 'xxx'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a is null and b is null;`,
				Results:   []sql.Row{{`Seq Scan on hp0 hp`}, {`Filter: ((a IS NULL) AND (b IS NULL))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a = 1 and b is null;`,
				Results:   []sql.Row{{`Seq Scan on hp1 hp`}, {`Filter: ((b IS NULL) AND (a = 1))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a = 1 and b = 'xxx';`,
				Results:   []sql.Row{{`Seq Scan on hp0 hp`}, {`Filter: ((a = 1) AND (b = 'xxx'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a is null and b = 'xxx';`,
				Results:   []sql.Row{{`Seq Scan on hp2 hp`}, {`Filter: ((a IS NULL) AND (b = 'xxx'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a = 2 and b = 'xxx';`,
				Results:   []sql.Row{{`Seq Scan on hp3 hp`}, {`Filter: ((a = 2) AND (b = 'xxx'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a = 1 and b = 'abcde';`,
				Results:   []sql.Row{{`Seq Scan on hp2 hp`}, {`Filter: ((a = 1) AND (b = 'abcde'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where (a = 1 and b = 'abcde') or (a = 2 and b = 'xxx') or (a is null and b is null);`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on hp0 hp_1`}, {`Filter: (((a = 1) AND (b = 'abcde'::text)) OR ((a = 2) AND (b = 'xxx'::text)) OR ((a IS NULL) AND (b IS NULL)))`}, {`->  Seq Scan on hp2 hp_2`}, {`Filter: (((a = 1) AND (b = 'abcde'::text)) OR ((a = 2) AND (b = 'xxx'::text)) OR ((a IS NULL) AND (b IS NULL)))`}, {`->  Seq Scan on hp3 hp_3`}, {`Filter: (((a = 1) AND (b = 'abcde'::text)) OR ((a = 2) AND (b = 'xxx'::text)) OR ((a IS NULL) AND (b IS NULL)))`}},
			},
			{
				Statement: `drop table hp1;`,
			},
			{
				Statement: `drop table hp3;`,
			},
			{
				Statement: `explain (costs off) select * from hp where a = 1 and b = 'abcde';`,
				Results:   []sql.Row{{`Seq Scan on hp2 hp`}, {`Filter: ((a = 1) AND (b = 'abcde'::text))`}},
			},
			{
				Statement: `explain (costs off) select * from hp where a = 1 and b = 'abcde' and
  (c = 2 or c = 3);`,
				Results: []sql.Row{{`Seq Scan on hp2 hp`}, {`Filter: ((a = 1) AND (b = 'abcde'::text) AND ((c = 2) OR (c = 3)))`}},
			},
			{
				Statement: `drop table hp2;`,
			},
			{
				Statement: `explain (costs off) select * from hp where a = 1 and b = 'abcde' and
  (c = 2 or c = 3);`,
				Results: []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `drop table hp;`,
			},
			{
				Statement: `create table ab (a int not null, b int not null) partition by list (a);`,
			},
			{
				Statement: `create table ab_a2 partition of ab for values in(2) partition by list (b);`,
			},
			{
				Statement: `create table ab_a2_b1 partition of ab_a2 for values in (1);`,
			},
			{
				Statement: `create table ab_a2_b2 partition of ab_a2 for values in (2);`,
			},
			{
				Statement: `create table ab_a2_b3 partition of ab_a2 for values in (3);`,
			},
			{
				Statement: `create table ab_a1 partition of ab for values in(1) partition by list (b);`,
			},
			{
				Statement: `create table ab_a1_b1 partition of ab_a1 for values in (1);`,
			},
			{
				Statement: `create table ab_a1_b2 partition of ab_a1 for values in (2);`,
			},
			{
				Statement: `create table ab_a1_b3 partition of ab_a1 for values in (3);`,
			},
			{
				Statement: `create table ab_a3 partition of ab for values in(3) partition by list (b);`,
			},
			{
				Statement: `create table ab_a3_b1 partition of ab_a3 for values in (1);`,
			},
			{
				Statement: `create table ab_a3_b2 partition of ab_a3 for values in (2);`,
			},
			{
				Statement: `create table ab_a3_b3 partition of ab_a3 for values in (3);`,
			},
			{
				Statement: `set enable_indexonlyscan = off;`,
			},
			{
				Statement: `prepare ab_q1 (int, int, int) as
select * from ab where a between $1 and $2 and b <= $3;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute ab_q1 (2, 2, 3);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 6`}, {`->  Seq Scan on ab_a2_b1 ab_1 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}, {`->  Seq Scan on ab_a2_b2 ab_2 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}, {`->  Seq Scan on ab_a2_b3 ab_3 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute ab_q1 (1, 2, 3);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 3`}, {`->  Seq Scan on ab_a1_b1 ab_1 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}, {`->  Seq Scan on ab_a1_b2 ab_2 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}, {`->  Seq Scan on ab_a1_b3 ab_3 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}, {`->  Seq Scan on ab_a2_b1 ab_4 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}, {`->  Seq Scan on ab_a2_b2 ab_5 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}, {`->  Seq Scan on ab_a2_b3 ab_6 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b <= $3))`}},
			},
			{
				Statement: `deallocate ab_q1;`,
			},
			{
				Statement: `prepare ab_q1 (int, int) as
select a from ab where a between $1 and $2 and b < 3;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute ab_q1 (2, 2);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 4`}, {`->  Seq Scan on ab_a2_b1 ab_1 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 3))`}, {`->  Seq Scan on ab_a2_b2 ab_2 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 3))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute ab_q1 (2, 4);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 2`}, {`->  Seq Scan on ab_a2_b1 ab_1 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 3))`}, {`->  Seq Scan on ab_a2_b2 ab_2 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 3))`}, {`->  Seq Scan on ab_a3_b1 ab_3 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 3))`}, {`->  Seq Scan on ab_a3_b2 ab_4 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 3))`}},
			},
			{
				Statement: `prepare ab_q2 (int, int) as
select a from ab where a between $1 and $2 and b < (select 3);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute ab_q2 (2, 2);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 6`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Seq Scan on ab_a2_b1 ab_1 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < $0))`}, {`->  Seq Scan on ab_a2_b2 ab_2 (actual rows=0 loops=1)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < $0))`}, {`->  Seq Scan on ab_a2_b3 ab_3 (never executed)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < $0))`}},
			},
			{
				Statement: `prepare ab_q3 (int, int) as
select a from ab where b between $1 and $2 and a < (select 3);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute ab_q3 (2, 2);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 6`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Seq Scan on ab_a1_b2 ab_1 (actual rows=0 loops=1)`}, {`Filter: ((b >= $1) AND (b <= $2) AND (a < $0))`}, {`->  Seq Scan on ab_a2_b2 ab_2 (actual rows=0 loops=1)`}, {`Filter: ((b >= $1) AND (b <= $2) AND (a < $0))`}, {`->  Seq Scan on ab_a3_b2 ab_3 (never executed)`}, {`Filter: ((b >= $1) AND (b <= $2) AND (a < $0))`}},
			},
			{
				Statement: `create table list_part (a int) partition by list (a);`,
			},
			{
				Statement: `create table list_part1 partition of list_part for values in (1);`,
			},
			{
				Statement: `create table list_part2 partition of list_part for values in (2);`,
			},
			{
				Statement: `create table list_part3 partition of list_part for values in (3);`,
			},
			{
				Statement: `create table list_part4 partition of list_part for values in (4);`,
			},
			{
				Statement: `insert into list_part select generate_series(1,4);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `declare cur SCROLL CURSOR for select 1 from list_part where a > (select 1) and a < (select 4);`,
			},
			{
				Statement: `move 3 from cur;`,
			},
			{
				Statement: `fetch backward all from cur;`,
				Results:   []sql.Row{{1}, {1}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create function list_part_fn(int) returns int as $$ begin return $1; end;$$ language plpgsql stable;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) select * from list_part where a = list_part_fn(1);`,
				Results:   []sql.Row{{`Append (actual rows=1 loops=1)`}, {`Subplans Removed: 3`}, {`->  Seq Scan on list_part1 list_part_1 (actual rows=1 loops=1)`}, {`Filter: (a = list_part_fn(1))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) select * from list_part where a = list_part_fn(a);`,
				Results:   []sql.Row{{`Append (actual rows=4 loops=1)`}, {`->  Seq Scan on list_part1 list_part_1 (actual rows=1 loops=1)`}, {`Filter: (a = list_part_fn(a))`}, {`->  Seq Scan on list_part2 list_part_2 (actual rows=1 loops=1)`}, {`Filter: (a = list_part_fn(a))`}, {`->  Seq Scan on list_part3 list_part_3 (actual rows=1 loops=1)`}, {`Filter: (a = list_part_fn(a))`}, {`->  Seq Scan on list_part4 list_part_4 (actual rows=1 loops=1)`}, {`Filter: (a = list_part_fn(a))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) select * from list_part where a = list_part_fn(1) + a;`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`->  Seq Scan on list_part1 list_part_1 (actual rows=0 loops=1)`}, {`Filter: (a = (list_part_fn(1) + a))`}, {`Rows Removed by Filter: 1`}, {`->  Seq Scan on list_part2 list_part_2 (actual rows=0 loops=1)`}, {`Filter: (a = (list_part_fn(1) + a))`}, {`Rows Removed by Filter: 1`}, {`->  Seq Scan on list_part3 list_part_3 (actual rows=0 loops=1)`}, {`Filter: (a = (list_part_fn(1) + a))`}, {`Rows Removed by Filter: 1`}, {`->  Seq Scan on list_part4 list_part_4 (actual rows=0 loops=1)`}, {`Filter: (a = (list_part_fn(1) + a))`}, {`Rows Removed by Filter: 1`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `drop table list_part;`,
			},
			{
				Statement: `create function explain_parallel_append(text) returns setof text
language plpgsql as
$$
declare
    ln text;`,
			},
			{
				Statement: `begin
    for ln in
        execute format('explain (analyze, costs off, summary off, timing off) %s',
            $1)
    loop
        ln := regexp_replace(ln, 'Workers Launched: \d+', 'Workers Launched: N');`,
			},
			{
				Statement: `        ln := regexp_replace(ln, 'actual rows=\d+ loops=\d+', 'actual rows=N loops=N');`,
			},
			{
				Statement: `        ln := regexp_replace(ln, 'Rows Removed by Filter: \d+', 'Rows Removed by Filter: N');`,
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
				Statement: `prepare ab_q4 (int, int) as
select avg(a) from ab where a between $1 and $2 and b < 4;`,
			},
			{
				Statement: `set parallel_setup_cost = 0;`,
			},
			{
				Statement: `set parallel_tuple_cost = 0;`,
			},
			{
				Statement: `set min_parallel_table_scan_size = 0;`,
			},
			{
				Statement: `set max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `select explain_parallel_append('execute ab_q4 (2, 2)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 2`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Parallel Append (actual rows=N loops=N)`}, {`Subplans Removed: 6`}, {`->  Parallel Seq Scan on ab_a2_b1 ab_1 (actual rows=N loops=N)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 4))`}, {`->  Parallel Seq Scan on ab_a2_b2 ab_2 (actual rows=N loops=N)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 4))`}, {`->  Parallel Seq Scan on ab_a2_b3 ab_3 (actual rows=N loops=N)`}, {`Filter: ((a >= $1) AND (a <= $2) AND (b < 4))`}},
			},
			{
				Statement: `prepare ab_q5 (int, int, int) as
select avg(a) from ab where a in($1,$2,$3) and b < 4;`,
			},
			{
				Statement: `select explain_parallel_append('execute ab_q5 (1, 1, 1)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 2`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Parallel Append (actual rows=N loops=N)`}, {`Subplans Removed: 6`}, {`->  Parallel Seq Scan on ab_a1_b1 ab_1 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}, {`->  Parallel Seq Scan on ab_a1_b2 ab_2 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}, {`->  Parallel Seq Scan on ab_a1_b3 ab_3 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}},
			},
			{
				Statement: `select explain_parallel_append('execute ab_q5 (2, 3, 3)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 2`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Parallel Append (actual rows=N loops=N)`}, {`Subplans Removed: 3`}, {`->  Parallel Seq Scan on ab_a2_b1 ab_1 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}, {`->  Parallel Seq Scan on ab_a2_b2 ab_2 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}, {`->  Parallel Seq Scan on ab_a2_b3 ab_3 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}, {`->  Parallel Seq Scan on ab_a3_b1 ab_4 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}, {`->  Parallel Seq Scan on ab_a3_b2 ab_5 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}, {`->  Parallel Seq Scan on ab_a3_b3 ab_6 (actual rows=N loops=N)`}, {`Filter: ((b < 4) AND (a = ANY (ARRAY[$1, $2, $3])))`}},
			},
			{
				Statement: `select explain_parallel_append('execute ab_q5 (33, 44, 55)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 2`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Parallel Append (actual rows=N loops=N)`}, {`Subplans Removed: 9`}},
			},
			{
				Statement: `select explain_parallel_append('select count(*) from ab where (a = (select 1) or a = (select 3)) and b = 2');`,
				Results:   []sql.Row{{`Aggregate (actual rows=N loops=N)`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=N loops=N)`}, {`InitPlan 2 (returns $1)`}, {`->  Result (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 2`}, {`Params Evaluated: $0, $1`}, {`Workers Launched: N`}, {`->  Parallel Append (actual rows=N loops=N)`}, {`->  Parallel Seq Scan on ab_a1_b2 ab_1 (actual rows=N loops=N)`}, {`Filter: ((b = 2) AND ((a = $0) OR (a = $1)))`}, {`->  Parallel Seq Scan on ab_a2_b2 ab_2 (never executed)`}, {`Filter: ((b = 2) AND ((a = $0) OR (a = $1)))`}, {`->  Parallel Seq Scan on ab_a3_b2 ab_3 (actual rows=N loops=N)`}, {`Filter: ((b = 2) AND ((a = $0) OR (a = $1)))`}},
			},
			{
				Statement: `create table lprt_a (a int not null);`,
			},
			{
				Statement: `insert into lprt_a select 0 from generate_series(1,100);`,
			},
			{
				Statement: `insert into lprt_a values(1),(1);`,
			},
			{
				Statement: `analyze lprt_a;`,
			},
			{
				Statement: `create index ab_a2_b1_a_idx on ab_a2_b1 (a);`,
			},
			{
				Statement: `create index ab_a2_b2_a_idx on ab_a2_b2 (a);`,
			},
			{
				Statement: `create index ab_a2_b3_a_idx on ab_a2_b3 (a);`,
			},
			{
				Statement: `create index ab_a1_b1_a_idx on ab_a1_b1 (a);`,
			},
			{
				Statement: `create index ab_a1_b2_a_idx on ab_a1_b2 (a);`,
			},
			{
				Statement: `create index ab_a1_b3_a_idx on ab_a1_b3 (a);`,
			},
			{
				Statement: `create index ab_a3_b1_a_idx on ab_a3_b1 (a);`,
			},
			{
				Statement: `create index ab_a3_b2_a_idx on ab_a3_b2 (a);`,
			},
			{
				Statement: `create index ab_a3_b3_a_idx on ab_a3_b3 (a);`,
			},
			{
				Statement: `set enable_hashjoin = 0;`,
			},
			{
				Statement: `set enable_mergejoin = 0;`,
			},
			{
				Statement: `set enable_memoize = 0;`,
			},
			{
				Statement: `select explain_parallel_append('select avg(ab.a) from ab inner join lprt_a a on ab.a = a.a where a.a in(0, 0, 1)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 1`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Nested Loop (actual rows=N loops=N)`}, {`->  Parallel Seq Scan on lprt_a a (actual rows=N loops=N)`}, {`Filter: (a = ANY ('{0,0,1}'::integer[]))`}, {`->  Append (actual rows=N loops=N)`}, {`->  Index Scan using ab_a1_b1_a_idx on ab_a1_b1 ab_1 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a1_b2_a_idx on ab_a1_b2 ab_2 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a1_b3_a_idx on ab_a1_b3 ab_3 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b1_a_idx on ab_a2_b1 ab_4 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b2_a_idx on ab_a2_b2 ab_5 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b3_a_idx on ab_a2_b3 ab_6 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b1_a_idx on ab_a3_b1 ab_7 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b2_a_idx on ab_a3_b2 ab_8 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b3_a_idx on ab_a3_b3 ab_9 (never executed)`}, {`Index Cond: (a = a.a)`}},
			},
			{
				Statement: `select explain_parallel_append('select avg(ab.a) from ab inner join lprt_a a on ab.a = a.a + 0 where a.a in(0, 0, 1)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 1`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Nested Loop (actual rows=N loops=N)`}, {`->  Parallel Seq Scan on lprt_a a (actual rows=N loops=N)`}, {`Filter: (a = ANY ('{0,0,1}'::integer[]))`}, {`->  Append (actual rows=N loops=N)`}, {`->  Index Scan using ab_a1_b1_a_idx on ab_a1_b1 ab_1 (actual rows=N loops=N)`}, {`Index Cond: (a = (a.a + 0))`}, {`->  Index Scan using ab_a1_b2_a_idx on ab_a1_b2 ab_2 (actual rows=N loops=N)`}, {`Index Cond: (a = (a.a + 0))`}, {`->  Index Scan using ab_a1_b3_a_idx on ab_a1_b3 ab_3 (actual rows=N loops=N)`}, {`Index Cond: (a = (a.a + 0))`}, {`->  Index Scan using ab_a2_b1_a_idx on ab_a2_b1 ab_4 (never executed)`}, {`Index Cond: (a = (a.a + 0))`}, {`->  Index Scan using ab_a2_b2_a_idx on ab_a2_b2 ab_5 (never executed)`}, {`Index Cond: (a = (a.a + 0))`}, {`->  Index Scan using ab_a2_b3_a_idx on ab_a2_b3 ab_6 (never executed)`}, {`Index Cond: (a = (a.a + 0))`}, {`->  Index Scan using ab_a3_b1_a_idx on ab_a3_b1 ab_7 (never executed)`}, {`Index Cond: (a = (a.a + 0))`}, {`->  Index Scan using ab_a3_b2_a_idx on ab_a3_b2 ab_8 (never executed)`}, {`Index Cond: (a = (a.a + 0))`}, {`->  Index Scan using ab_a3_b3_a_idx on ab_a3_b3 ab_9 (never executed)`}, {`Index Cond: (a = (a.a + 0))`}},
			},
			{
				Statement: `insert into lprt_a values(3),(3);`,
			},
			{
				Statement: `select explain_parallel_append('select avg(ab.a) from ab inner join lprt_a a on ab.a = a.a where a.a in(1, 0, 3)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 1`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Nested Loop (actual rows=N loops=N)`}, {`->  Parallel Seq Scan on lprt_a a (actual rows=N loops=N)`}, {`Filter: (a = ANY ('{1,0,3}'::integer[]))`}, {`->  Append (actual rows=N loops=N)`}, {`->  Index Scan using ab_a1_b1_a_idx on ab_a1_b1 ab_1 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a1_b2_a_idx on ab_a1_b2 ab_2 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a1_b3_a_idx on ab_a1_b3 ab_3 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b1_a_idx on ab_a2_b1 ab_4 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b2_a_idx on ab_a2_b2 ab_5 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b3_a_idx on ab_a2_b3 ab_6 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b1_a_idx on ab_a3_b1 ab_7 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b2_a_idx on ab_a3_b2 ab_8 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b3_a_idx on ab_a3_b3 ab_9 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}},
			},
			{
				Statement: `select explain_parallel_append('select avg(ab.a) from ab inner join lprt_a a on ab.a = a.a where a.a in(1, 0, 0)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 1`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Nested Loop (actual rows=N loops=N)`}, {`->  Parallel Seq Scan on lprt_a a (actual rows=N loops=N)`}, {`Filter: (a = ANY ('{1,0,0}'::integer[]))`}, {`Rows Removed by Filter: N`}, {`->  Append (actual rows=N loops=N)`}, {`->  Index Scan using ab_a1_b1_a_idx on ab_a1_b1 ab_1 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a1_b2_a_idx on ab_a1_b2 ab_2 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a1_b3_a_idx on ab_a1_b3 ab_3 (actual rows=N loops=N)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b1_a_idx on ab_a2_b1 ab_4 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b2_a_idx on ab_a2_b2 ab_5 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b3_a_idx on ab_a2_b3 ab_6 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b1_a_idx on ab_a3_b1 ab_7 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b2_a_idx on ab_a3_b2 ab_8 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b3_a_idx on ab_a3_b3 ab_9 (never executed)`}, {`Index Cond: (a = a.a)`}},
			},
			{
				Statement: `delete from lprt_a where a = 1;`,
			},
			{
				Statement: `select explain_parallel_append('select avg(ab.a) from ab inner join lprt_a a on ab.a = a.a where a.a in(1, 0, 0)');`,
				Results:   []sql.Row{{`Finalize Aggregate (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 1`}, {`Workers Launched: N`}, {`->  Partial Aggregate (actual rows=N loops=N)`}, {`->  Nested Loop (actual rows=N loops=N)`}, {`->  Parallel Seq Scan on lprt_a a (actual rows=N loops=N)`}, {`Filter: (a = ANY ('{1,0,0}'::integer[]))`}, {`Rows Removed by Filter: N`}, {`->  Append (actual rows=N loops=N)`}, {`->  Index Scan using ab_a1_b1_a_idx on ab_a1_b1 ab_1 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a1_b2_a_idx on ab_a1_b2 ab_2 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a1_b3_a_idx on ab_a1_b3 ab_3 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b1_a_idx on ab_a2_b1 ab_4 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b2_a_idx on ab_a2_b2 ab_5 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a2_b3_a_idx on ab_a2_b3 ab_6 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b1_a_idx on ab_a3_b1 ab_7 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b2_a_idx on ab_a3_b2 ab_8 (never executed)`}, {`Index Cond: (a = a.a)`}, {`->  Index Scan using ab_a3_b3_a_idx on ab_a3_b3 ab_9 (never executed)`}, {`Index Cond: (a = a.a)`}},
			},
			{
				Statement: `reset enable_hashjoin;`,
			},
			{
				Statement: `reset enable_mergejoin;`,
			},
			{
				Statement: `reset enable_memoize;`,
			},
			{
				Statement: `reset parallel_setup_cost;`,
			},
			{
				Statement: `reset parallel_tuple_cost;`,
			},
			{
				Statement: `reset min_parallel_table_scan_size;`,
			},
			{
				Statement: `reset max_parallel_workers_per_gather;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from ab where a = (select max(a) from lprt_a) and b = (select max(a)-1 from lprt_a);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`InitPlan 1 (returns $0)`}, {`->  Aggregate (actual rows=1 loops=1)`}, {`->  Seq Scan on lprt_a (actual rows=102 loops=1)`}, {`InitPlan 2 (returns $1)`}, {`->  Aggregate (actual rows=1 loops=1)`}, {`->  Seq Scan on lprt_a lprt_a_1 (actual rows=102 loops=1)`}, {`->  Bitmap Heap Scan on ab_a1_b1 ab_1 (never executed)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a1_b1_a_idx (never executed)`}, {`Index Cond: (a = $0)`}, {`->  Bitmap Heap Scan on ab_a1_b2 ab_2 (never executed)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a1_b2_a_idx (never executed)`}, {`Index Cond: (a = $0)`}, {`->  Bitmap Heap Scan on ab_a1_b3 ab_3 (never executed)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a1_b3_a_idx (never executed)`}, {`Index Cond: (a = $0)`}, {`->  Bitmap Heap Scan on ab_a2_b1 ab_4 (never executed)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a2_b1_a_idx (never executed)`}, {`Index Cond: (a = $0)`}, {`->  Bitmap Heap Scan on ab_a2_b2 ab_5 (never executed)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a2_b2_a_idx (never executed)`}, {`Index Cond: (a = $0)`}, {`->  Bitmap Heap Scan on ab_a2_b3 ab_6 (never executed)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a2_b3_a_idx (never executed)`}, {`Index Cond: (a = $0)`}, {`->  Bitmap Heap Scan on ab_a3_b1 ab_7 (never executed)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a3_b1_a_idx (never executed)`}, {`Index Cond: (a = $0)`}, {`->  Bitmap Heap Scan on ab_a3_b2 ab_8 (actual rows=0 loops=1)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a3_b2_a_idx (actual rows=0 loops=1)`}, {`Index Cond: (a = $0)`}, {`->  Bitmap Heap Scan on ab_a3_b3 ab_9 (never executed)`}, {`Recheck Cond: (a = $0)`}, {`Filter: (b = $1)`}, {`->  Bitmap Index Scan on ab_a3_b3_a_idx (never executed)`}, {`Index Cond: (a = $0)`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from (select * from ab where a = 1 union all select * from ab) ab where b = (select 1);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Append (actual rows=0 loops=1)`}, {`->  Bitmap Heap Scan on ab_a1_b1 ab_11 (actual rows=0 loops=1)`}, {`Recheck Cond: (a = 1)`}, {`Filter: (b = $0)`}, {`->  Bitmap Index Scan on ab_a1_b1_a_idx (actual rows=0 loops=1)`}, {`Index Cond: (a = 1)`}, {`->  Bitmap Heap Scan on ab_a1_b2 ab_12 (never executed)`}, {`Recheck Cond: (a = 1)`}, {`Filter: (b = $0)`}, {`->  Bitmap Index Scan on ab_a1_b2_a_idx (never executed)`}, {`Index Cond: (a = 1)`}, {`->  Bitmap Heap Scan on ab_a1_b3 ab_13 (never executed)`}, {`Recheck Cond: (a = 1)`}, {`Filter: (b = $0)`}, {`->  Bitmap Index Scan on ab_a1_b3_a_idx (never executed)`}, {`Index Cond: (a = 1)`}, {`->  Seq Scan on ab_a1_b1 ab_1 (actual rows=0 loops=1)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a1_b2 ab_2 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a1_b3 ab_3 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a2_b1 ab_4 (actual rows=0 loops=1)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a2_b2 ab_5 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a2_b3 ab_6 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a3_b1 ab_7 (actual rows=0 loops=1)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a3_b2 ab_8 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a3_b3 ab_9 (never executed)`}, {`Filter: (b = $0)`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from (select * from ab where a = 1 union all (values(10,5)) union all select * from ab) ab where b = (select 1);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Append (actual rows=0 loops=1)`}, {`->  Bitmap Heap Scan on ab_a1_b1 ab_11 (actual rows=0 loops=1)`}, {`Recheck Cond: (a = 1)`}, {`Filter: (b = $0)`}, {`->  Bitmap Index Scan on ab_a1_b1_a_idx (actual rows=0 loops=1)`}, {`Index Cond: (a = 1)`}, {`->  Bitmap Heap Scan on ab_a1_b2 ab_12 (never executed)`}, {`Recheck Cond: (a = 1)`}, {`Filter: (b = $0)`}, {`->  Bitmap Index Scan on ab_a1_b2_a_idx (never executed)`}, {`Index Cond: (a = 1)`}, {`->  Bitmap Heap Scan on ab_a1_b3 ab_13 (never executed)`}, {`Recheck Cond: (a = 1)`}, {`Filter: (b = $0)`}, {`->  Bitmap Index Scan on ab_a1_b3_a_idx (never executed)`}, {`Index Cond: (a = 1)`}, {`->  Result (actual rows=0 loops=1)`}, {`One-Time Filter: (5 = $0)`}, {`->  Seq Scan on ab_a1_b1 ab_1 (actual rows=0 loops=1)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a1_b2 ab_2 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a1_b3 ab_3 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a2_b1 ab_4 (actual rows=0 loops=1)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a2_b2 ab_5 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a2_b3 ab_6 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a3_b1 ab_7 (actual rows=0 loops=1)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a3_b2 ab_8 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a3_b3 ab_9 (never executed)`}, {`Filter: (b = $0)`}},
			},
			{
				Statement: `create table xy_1 (x int, y int);`,
			},
			{
				Statement: `insert into xy_1 values(100,-10);`,
			},
			{
				Statement: `set enable_bitmapscan = 0;`,
			},
			{
				Statement: `set enable_indexscan = 0;`,
			},
			{
				Statement: `prepare ab_q6 as
select * from (
	select tableoid::regclass,a,b from ab
union all
	select tableoid::regclass,x,y from xy_1
union all
	select tableoid::regclass,a,b from ab
) ab where a = $1 and b = (select -10);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute ab_q6(1);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 12`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Seq Scan on ab_a1_b1 ab_1 (never executed)`}, {`Filter: ((a = $1) AND (b = $0))`}, {`->  Seq Scan on ab_a1_b2 ab_2 (never executed)`}, {`Filter: ((a = $1) AND (b = $0))`}, {`->  Seq Scan on ab_a1_b3 ab_3 (never executed)`}, {`Filter: ((a = $1) AND (b = $0))`}, {`->  Seq Scan on xy_1 (actual rows=0 loops=1)`}, {`Filter: ((x = $1) AND (y = $0))`}, {`Rows Removed by Filter: 1`}, {`->  Seq Scan on ab_a1_b1 ab_4 (never executed)`}, {`Filter: ((a = $1) AND (b = $0))`}, {`->  Seq Scan on ab_a1_b2 ab_5 (never executed)`}, {`Filter: ((a = $1) AND (b = $0))`}, {`->  Seq Scan on ab_a1_b3 ab_6 (never executed)`}, {`Filter: ((a = $1) AND (b = $0))`}},
			},
			{
				Statement: `execute ab_q6(100);`,
				Results:   []sql.Row{{`xy_1`, 100, -10}},
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `reset enable_indexscan;`,
			},
			{
				Statement: `deallocate ab_q1;`,
			},
			{
				Statement: `deallocate ab_q2;`,
			},
			{
				Statement: `deallocate ab_q3;`,
			},
			{
				Statement: `deallocate ab_q4;`,
			},
			{
				Statement: `deallocate ab_q5;`,
			},
			{
				Statement: `deallocate ab_q6;`,
			},
			{
				Statement: `insert into ab values (1,2);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
update ab_a1 set b = 3 from ab where ab.a = 1 and ab.a = ab_a1.a;`,
				Results: []sql.Row{{`Update on ab_a1 (actual rows=0 loops=1)`}, {`Update on ab_a1_b1 ab_a1_1`}, {`Update on ab_a1_b2 ab_a1_2`}, {`Update on ab_a1_b3 ab_a1_3`}, {`->  Nested Loop (actual rows=1 loops=1)`}, {`->  Append (actual rows=1 loops=1)`}, {`->  Bitmap Heap Scan on ab_a1_b1 ab_a1_1 (actual rows=0 loops=1)`}, {`Recheck Cond: (a = 1)`}, {`->  Bitmap Index Scan on ab_a1_b1_a_idx (actual rows=0 loops=1)`}, {`Index Cond: (a = 1)`}, {`->  Bitmap Heap Scan on ab_a1_b2 ab_a1_2 (actual rows=1 loops=1)`}, {`Recheck Cond: (a = 1)`}, {`Heap Blocks: exact=1`}, {`->  Bitmap Index Scan on ab_a1_b2_a_idx (actual rows=1 loops=1)`}, {`Index Cond: (a = 1)`}, {`->  Bitmap Heap Scan on ab_a1_b3 ab_a1_3 (actual rows=0 loops=1)`}, {`Recheck Cond: (a = 1)`}, {`->  Bitmap Index Scan on ab_a1_b3_a_idx (actual rows=1 loops=1)`}, {`Index Cond: (a = 1)`}, {`->  Materialize (actual rows=1 loops=1)`}, {`->  Append (actual rows=1 loops=1)`}, {`->  Bitmap Heap Scan on ab_a1_b1 ab_1 (actual rows=0 loops=1)`}, {`Recheck Cond: (a = 1)`}, {`->  Bitmap Index Scan on ab_a1_b1_a_idx (actual rows=0 loops=1)`}, {`Index Cond: (a = 1)`}, {`->  Bitmap Heap Scan on ab_a1_b2 ab_2 (actual rows=1 loops=1)`}, {`Recheck Cond: (a = 1)`}, {`Heap Blocks: exact=1`}, {`->  Bitmap Index Scan on ab_a1_b2_a_idx (actual rows=1 loops=1)`}, {`Index Cond: (a = 1)`}, {`->  Bitmap Heap Scan on ab_a1_b3 ab_3 (actual rows=0 loops=1)`}, {`Recheck Cond: (a = 1)`}, {`->  Bitmap Index Scan on ab_a1_b3_a_idx (actual rows=1 loops=1)`}, {`Index Cond: (a = 1)`}},
			},
			{
				Statement: `table ab;`,
				Results:   []sql.Row{{1, 3}},
			},
			{
				Statement: `truncate ab;`,
			},
			{
				Statement: `insert into ab values (1, 1), (1, 2), (1, 3), (2, 1);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
update ab_a1 set b = 3 from ab_a2 where ab_a2.b = (select 1);`,
				Results: []sql.Row{{`Update on ab_a1 (actual rows=0 loops=1)`}, {`Update on ab_a1_b1 ab_a1_1`}, {`Update on ab_a1_b2 ab_a1_2`}, {`Update on ab_a1_b3 ab_a1_3`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Nested Loop (actual rows=3 loops=1)`}, {`->  Append (actual rows=3 loops=1)`}, {`->  Seq Scan on ab_a1_b1 ab_a1_1 (actual rows=1 loops=1)`}, {`->  Seq Scan on ab_a1_b2 ab_a1_2 (actual rows=1 loops=1)`}, {`->  Seq Scan on ab_a1_b3 ab_a1_3 (actual rows=1 loops=1)`}, {`->  Materialize (actual rows=1 loops=3)`}, {`->  Append (actual rows=1 loops=1)`}, {`->  Seq Scan on ab_a2_b1 ab_a2_1 (actual rows=1 loops=1)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a2_b2 ab_a2_2 (never executed)`}, {`Filter: (b = $0)`}, {`->  Seq Scan on ab_a2_b3 ab_a2_3 (never executed)`}, {`Filter: (b = $0)`}},
			},
			{
				Statement: `select tableoid::regclass, * from ab;`,
				Results:   []sql.Row{{`ab_a1_b3`, 1, 3}, {`ab_a1_b3`, 1, 3}, {`ab_a1_b3`, 1, 3}, {`ab_a2_b1`, 2, 1}},
			},
			{
				Statement: `drop table ab, lprt_a;`,
			},
			{
				Statement: `create table tbl1(col1 int);`,
			},
			{
				Statement: `insert into tbl1 values (501), (505);`,
			},
			{
				Statement: `create table tprt (col1 int) partition by range (col1);`,
			},
			{
				Statement: `create table tprt_1 partition of tprt for values from (1) to (501);`,
			},
			{
				Statement: `create table tprt_2 partition of tprt for values from (501) to (1001);`,
			},
			{
				Statement: `create table tprt_3 partition of tprt for values from (1001) to (2001);`,
			},
			{
				Statement: `create table tprt_4 partition of tprt for values from (2001) to (3001);`,
			},
			{
				Statement: `create table tprt_5 partition of tprt for values from (3001) to (4001);`,
			},
			{
				Statement: `create table tprt_6 partition of tprt for values from (4001) to (5001);`,
			},
			{
				Statement: `create index tprt1_idx on tprt_1 (col1);`,
			},
			{
				Statement: `create index tprt2_idx on tprt_2 (col1);`,
			},
			{
				Statement: `create index tprt3_idx on tprt_3 (col1);`,
			},
			{
				Statement: `create index tprt4_idx on tprt_4 (col1);`,
			},
			{
				Statement: `create index tprt5_idx on tprt_5 (col1);`,
			},
			{
				Statement: `create index tprt6_idx on tprt_6 (col1);`,
			},
			{
				Statement: `insert into tprt values (10), (20), (501), (502), (505), (1001), (4500);`,
			},
			{
				Statement: `set enable_hashjoin = off;`,
			},
			{
				Statement: `set enable_mergejoin = off;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from tbl1 join tprt on tbl1.col1 > tprt.col1;`,
				Results: []sql.Row{{`Nested Loop (actual rows=6 loops=1)`}, {`->  Seq Scan on tbl1 (actual rows=2 loops=1)`}, {`->  Append (actual rows=3 loops=2)`}, {`->  Index Scan using tprt1_idx on tprt_1 (actual rows=2 loops=2)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt2_idx on tprt_2 (actual rows=2 loops=1)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt3_idx on tprt_3 (never executed)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt4_idx on tprt_4 (never executed)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt5_idx on tprt_5 (never executed)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt6_idx on tprt_6 (never executed)`}, {`Index Cond: (col1 < tbl1.col1)`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from tbl1 join tprt on tbl1.col1 = tprt.col1;`,
				Results: []sql.Row{{`Nested Loop (actual rows=2 loops=1)`}, {`->  Seq Scan on tbl1 (actual rows=2 loops=1)`}, {`->  Append (actual rows=1 loops=2)`}, {`->  Index Scan using tprt1_idx on tprt_1 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt2_idx on tprt_2 (actual rows=1 loops=2)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt3_idx on tprt_3 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt4_idx on tprt_4 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt5_idx on tprt_5 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt6_idx on tprt_6 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}},
			},
			{
				Statement: `select tbl1.col1, tprt.col1 from tbl1
inner join tprt on tbl1.col1 > tprt.col1
order by tbl1.col1, tprt.col1;`,
				Results: []sql.Row{{501, 10}, {501, 20}, {505, 10}, {505, 20}, {505, 501}, {505, 502}},
			},
			{
				Statement: `select tbl1.col1, tprt.col1 from tbl1
inner join tprt on tbl1.col1 = tprt.col1
order by tbl1.col1, tprt.col1;`,
				Results: []sql.Row{{501, 501}, {505, 505}},
			},
			{
				Statement: `insert into tbl1 values (1001), (1010), (1011);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from tbl1 inner join tprt on tbl1.col1 > tprt.col1;`,
				Results: []sql.Row{{`Nested Loop (actual rows=23 loops=1)`}, {`->  Seq Scan on tbl1 (actual rows=5 loops=1)`}, {`->  Append (actual rows=5 loops=5)`}, {`->  Index Scan using tprt1_idx on tprt_1 (actual rows=2 loops=5)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt2_idx on tprt_2 (actual rows=3 loops=4)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt3_idx on tprt_3 (actual rows=1 loops=2)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt4_idx on tprt_4 (never executed)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt5_idx on tprt_5 (never executed)`}, {`Index Cond: (col1 < tbl1.col1)`}, {`->  Index Scan using tprt6_idx on tprt_6 (never executed)`}, {`Index Cond: (col1 < tbl1.col1)`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from tbl1 inner join tprt on tbl1.col1 = tprt.col1;`,
				Results: []sql.Row{{`Nested Loop (actual rows=3 loops=1)`}, {`->  Seq Scan on tbl1 (actual rows=5 loops=1)`}, {`->  Append (actual rows=1 loops=5)`}, {`->  Index Scan using tprt1_idx on tprt_1 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt2_idx on tprt_2 (actual rows=1 loops=2)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt3_idx on tprt_3 (actual rows=0 loops=3)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt4_idx on tprt_4 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt5_idx on tprt_5 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt6_idx on tprt_6 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}},
			},
			{
				Statement: `select tbl1.col1, tprt.col1 from tbl1
inner join tprt on tbl1.col1 > tprt.col1
order by tbl1.col1, tprt.col1;`,
				Results: []sql.Row{{501, 10}, {501, 20}, {505, 10}, {505, 20}, {505, 501}, {505, 502}, {1001, 10}, {1001, 20}, {1001, 501}, {1001, 502}, {1001, 505}, {1010, 10}, {1010, 20}, {1010, 501}, {1010, 502}, {1010, 505}, {1010, 1001}, {1011, 10}, {1011, 20}, {1011, 501}, {1011, 502}, {1011, 505}, {1011, 1001}},
			},
			{
				Statement: `select tbl1.col1, tprt.col1 from tbl1
inner join tprt on tbl1.col1 = tprt.col1
order by tbl1.col1, tprt.col1;`,
				Results: []sql.Row{{501, 501}, {505, 505}, {1001, 1001}},
			},
			{
				Statement: `delete from tbl1;`,
			},
			{
				Statement: `insert into tbl1 values (4400);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from tbl1 join tprt on tbl1.col1 < tprt.col1;`,
				Results: []sql.Row{{`Nested Loop (actual rows=1 loops=1)`}, {`->  Seq Scan on tbl1 (actual rows=1 loops=1)`}, {`->  Append (actual rows=1 loops=1)`}, {`->  Index Scan using tprt1_idx on tprt_1 (never executed)`}, {`Index Cond: (col1 > tbl1.col1)`}, {`->  Index Scan using tprt2_idx on tprt_2 (never executed)`}, {`Index Cond: (col1 > tbl1.col1)`}, {`->  Index Scan using tprt3_idx on tprt_3 (never executed)`}, {`Index Cond: (col1 > tbl1.col1)`}, {`->  Index Scan using tprt4_idx on tprt_4 (never executed)`}, {`Index Cond: (col1 > tbl1.col1)`}, {`->  Index Scan using tprt5_idx on tprt_5 (never executed)`}, {`Index Cond: (col1 > tbl1.col1)`}, {`->  Index Scan using tprt6_idx on tprt_6 (actual rows=1 loops=1)`}, {`Index Cond: (col1 > tbl1.col1)`}},
			},
			{
				Statement: `select tbl1.col1, tprt.col1 from tbl1
inner join tprt on tbl1.col1 < tprt.col1
order by tbl1.col1, tprt.col1;`,
				Results: []sql.Row{{4400, 4500}},
			},
			{
				Statement: `delete from tbl1;`,
			},
			{
				Statement: `insert into tbl1 values (10000);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from tbl1 join tprt on tbl1.col1 = tprt.col1;`,
				Results: []sql.Row{{`Nested Loop (actual rows=0 loops=1)`}, {`->  Seq Scan on tbl1 (actual rows=1 loops=1)`}, {`->  Append (actual rows=0 loops=1)`}, {`->  Index Scan using tprt1_idx on tprt_1 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt2_idx on tprt_2 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt3_idx on tprt_3 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt4_idx on tprt_4 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt5_idx on tprt_5 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}, {`->  Index Scan using tprt6_idx on tprt_6 (never executed)`}, {`Index Cond: (col1 = tbl1.col1)`}},
			},
			{
				Statement: `select tbl1.col1, tprt.col1 from tbl1
inner join tprt on tbl1.col1 = tprt.col1
order by tbl1.col1, tprt.col1;`,
				Results: []sql.Row{},
			},
			{
				Statement: `drop table tbl1, tprt;`,
			},
			{
				Statement: `create table part_abc (a int not null, b int not null, c int not null) partition by list (a);`,
			},
			{
				Statement: `create table part_bac (b int not null, a int not null, c int not null) partition by list (b);`,
			},
			{
				Statement: `create table part_cab (c int not null, a int not null, b int not null) partition by list (c);`,
			},
			{
				Statement: `create table part_abc_p1 (a int not null, b int not null, c int not null);`,
			},
			{
				Statement: `alter table part_abc attach partition part_bac for values in(1);`,
			},
			{
				Statement: `alter table part_bac attach partition part_cab for values in(2);`,
			},
			{
				Statement: `alter table part_cab attach partition part_abc_p1 for values in(3);`,
			},
			{
				Statement: `prepare part_abc_q1 (int, int, int) as
select * from part_abc where a = $1 and b = $2 and c = $3;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute part_abc_q1 (1, 2, 3);`,
				Results:   []sql.Row{{`Seq Scan on part_abc_p1 part_abc (actual rows=0 loops=1)`}, {`Filter: ((a = $1) AND (b = $2) AND (c = $3))`}},
			},
			{
				Statement: `deallocate part_abc_q1;`,
			},
			{
				Statement: `drop table part_abc;`,
			},
			{
				Statement: `create table listp (a int, b int) partition by list (a);`,
			},
			{
				Statement: `create table listp_1 partition of listp for values in(1) partition by list (b);`,
			},
			{
				Statement: `create table listp_1_1 partition of listp_1 for values in(1);`,
			},
			{
				Statement: `create table listp_2 partition of listp for values in(2) partition by list (b);`,
			},
			{
				Statement: `create table listp_2_1 partition of listp_2 for values in(2);`,
			},
			{
				Statement: `select * from listp where b = 1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `prepare q1 (int,int) as select * from listp where b in ($1,$2);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)  execute q1 (1,1);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 1`}, {`->  Seq Scan on listp_1_1 listp_1 (actual rows=0 loops=1)`}, {`Filter: (b = ANY (ARRAY[$1, $2]))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)  execute q1 (2,2);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 1`}, {`->  Seq Scan on listp_2_1 listp_1 (actual rows=0 loops=1)`}, {`Filter: (b = ANY (ARRAY[$1, $2]))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)  execute q1 (0,0);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 2`}},
			},
			{
				Statement: `deallocate q1;`,
			},
			{
				Statement: `prepare q1 (int,int,int,int) as select * from listp where b in($1,$2) and $3 <> b and $4 <> b;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)  execute q1 (1,2,2,0);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 1`}, {`->  Seq Scan on listp_1_1 listp_1 (actual rows=0 loops=1)`}, {`Filter: ((b = ANY (ARRAY[$1, $2])) AND ($3 <> b) AND ($4 <> b))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)  execute q1 (1,2,2,1);`,
				Results:   []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 2`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from listp where a = (select null::int);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Seq Scan on listp_1_1 listp_1 (never executed)`}, {`Filter: (a = $0)`}, {`->  Seq Scan on listp_2_1 listp_2 (never executed)`}, {`Filter: (a = $0)`}},
			},
			{
				Statement: `drop table listp;`,
			},
			{
				Statement: `create table stable_qual_pruning (a timestamp) partition by range (a);`,
			},
			{
				Statement: `create table stable_qual_pruning1 partition of stable_qual_pruning
  for values from ('2000-01-01') to ('2000-02-01');`,
			},
			{
				Statement: `create table stable_qual_pruning2 partition of stable_qual_pruning
  for values from ('2000-02-01') to ('2000-03-01');`,
			},
			{
				Statement: `create table stable_qual_pruning3 partition of stable_qual_pruning
  for values from ('3000-02-01') to ('3000-03-01');`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from stable_qual_pruning where a < localtimestamp;`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 1`}, {`->  Seq Scan on stable_qual_pruning1 stable_qual_pruning_1 (actual rows=0 loops=1)`}, {`Filter: (a < LOCALTIMESTAMP)`}, {`->  Seq Scan on stable_qual_pruning2 stable_qual_pruning_2 (actual rows=0 loops=1)`}, {`Filter: (a < LOCALTIMESTAMP)`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from stable_qual_pruning where a < '2000-02-01'::timestamptz;`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 2`}, {`->  Seq Scan on stable_qual_pruning1 stable_qual_pruning_1 (actual rows=0 loops=1)`}, {`Filter: (a < 'Tue Feb 01 00:00:00 2000 PST'::timestamp with time zone)`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from stable_qual_pruning
  where a = any(array['2010-02-01', '2020-01-01']::timestamp[]);`,
				Results: []sql.Row{{`Result (actual rows=0 loops=1)`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from stable_qual_pruning
  where a = any(array['2000-02-01', '2010-01-01']::timestamp[]);`,
				Results: []sql.Row{{`Seq Scan on stable_qual_pruning2 stable_qual_pruning (actual rows=0 loops=1)`}, {`Filter: (a = ANY ('{"Tue Feb 01 00:00:00 2000","Fri Jan 01 00:00:00 2010"}'::timestamp without time zone[]))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from stable_qual_pruning
  where a = any(array['2000-02-01', localtimestamp]::timestamp[]);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 2`}, {`->  Seq Scan on stable_qual_pruning2 stable_qual_pruning_1 (actual rows=0 loops=1)`}, {`Filter: (a = ANY (ARRAY['Tue Feb 01 00:00:00 2000'::timestamp without time zone, LOCALTIMESTAMP]))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from stable_qual_pruning
  where a = any(array['2010-02-01', '2020-01-01']::timestamptz[]);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 3`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from stable_qual_pruning
  where a = any(array['2000-02-01', '2010-01-01']::timestamptz[]);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`Subplans Removed: 2`}, {`->  Seq Scan on stable_qual_pruning2 stable_qual_pruning_1 (actual rows=0 loops=1)`}, {`Filter: (a = ANY ('{"Tue Feb 01 00:00:00 2000 PST","Fri Jan 01 00:00:00 2010 PST"}'::timestamp with time zone[]))`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from stable_qual_pruning
  where a = any(null::timestamptz[]);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`->  Seq Scan on stable_qual_pruning1 stable_qual_pruning_1 (actual rows=0 loops=1)`}, {`Filter: (a = ANY (NULL::timestamp with time zone[]))`}, {`->  Seq Scan on stable_qual_pruning2 stable_qual_pruning_2 (actual rows=0 loops=1)`}, {`Filter: (a = ANY (NULL::timestamp with time zone[]))`}, {`->  Seq Scan on stable_qual_pruning3 stable_qual_pruning_3 (actual rows=0 loops=1)`}, {`Filter: (a = ANY (NULL::timestamp with time zone[]))`}},
			},
			{
				Statement: `drop table stable_qual_pruning;`,
			},
			{
				Statement: `create table mc3p (a int, b int, c int) partition by range (a, abs(b), c);`,
			},
			{
				Statement: `create table mc3p0 partition of mc3p
  for values from (0, 0, 0) to (0, maxvalue, maxvalue);`,
			},
			{
				Statement: `create table mc3p1 partition of mc3p
  for values from (1, 1, 1) to (2, minvalue, minvalue);`,
			},
			{
				Statement: `create table mc3p2 partition of mc3p
  for values from (2, minvalue, minvalue) to (3, maxvalue, maxvalue);`,
			},
			{
				Statement: `insert into mc3p values (0, 1, 1), (1, 1, 1), (2, 1, 1);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from mc3p where a < 3 and abs(b) = 1;`,
				Results: []sql.Row{{`Append (actual rows=3 loops=1)`}, {`->  Seq Scan on mc3p0 mc3p_1 (actual rows=1 loops=1)`}, {`Filter: ((a < 3) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p1 mc3p_2 (actual rows=1 loops=1)`}, {`Filter: ((a < 3) AND (abs(b) = 1))`}, {`->  Seq Scan on mc3p2 mc3p_3 (actual rows=1 loops=1)`}, {`Filter: ((a < 3) AND (abs(b) = 1))`}},
			},
			{
				Statement: `prepare ps1 as
  select * from mc3p where a = $1 and abs(b) < (select 3);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
execute ps1(1);`,
				Results: []sql.Row{{`Append (actual rows=1 loops=1)`}, {`Subplans Removed: 2`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Seq Scan on mc3p1 mc3p_1 (actual rows=1 loops=1)`}, {`Filter: ((a = $1) AND (abs(b) < $0))`}},
			},
			{
				Statement: `deallocate ps1;`,
			},
			{
				Statement: `prepare ps2 as
  select * from mc3p where a <= $1 and abs(b) < (select 3);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
execute ps2(1);`,
				Results: []sql.Row{{`Append (actual rows=2 loops=1)`}, {`Subplans Removed: 1`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Seq Scan on mc3p0 mc3p_1 (actual rows=1 loops=1)`}, {`Filter: ((a <= $1) AND (abs(b) < $0))`}, {`->  Seq Scan on mc3p1 mc3p_2 (actual rows=1 loops=1)`}, {`Filter: ((a <= $1) AND (abs(b) < $0))`}},
			},
			{
				Statement: `deallocate ps2;`,
			},
			{
				Statement: `drop table mc3p;`,
			},
			{
				Statement: `create table boolvalues (value bool not null);`,
			},
			{
				Statement: `insert into boolvalues values('t'),('f');`,
			},
			{
				Statement: `create table boolp (a bool) partition by list (a);`,
			},
			{
				Statement: `create table boolp_t partition of boolp for values in('t');`,
			},
			{
				Statement: `create table boolp_f partition of boolp for values in('f');`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from boolp where a = (select value from boolvalues where value);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`InitPlan 1 (returns $0)`}, {`->  Seq Scan on boolvalues (actual rows=1 loops=1)`}, {`Filter: value`}, {`Rows Removed by Filter: 1`}, {`->  Seq Scan on boolp_f boolp_1 (never executed)`}, {`Filter: (a = $0)`}, {`->  Seq Scan on boolp_t boolp_2 (actual rows=0 loops=1)`}, {`Filter: (a = $0)`}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from boolp where a = (select value from boolvalues where not value);`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`InitPlan 1 (returns $0)`}, {`->  Seq Scan on boolvalues (actual rows=1 loops=1)`}, {`Filter: (NOT value)`}, {`Rows Removed by Filter: 1`}, {`->  Seq Scan on boolp_f boolp_1 (actual rows=0 loops=1)`}, {`Filter: (a = $0)`}, {`->  Seq Scan on boolp_t boolp_2 (never executed)`}, {`Filter: (a = $0)`}},
			},
			{
				Statement: `drop table boolp;`,
			},
			{
				Statement: `set enable_seqscan = off;`,
			},
			{
				Statement: `set enable_sort = off;`,
			},
			{
				Statement: `create table ma_test (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table ma_test_p1 partition of ma_test for values from (0) to (10);`,
			},
			{
				Statement: `create table ma_test_p2 partition of ma_test for values from (10) to (20);`,
			},
			{
				Statement: `create table ma_test_p3 partition of ma_test for values from (20) to (30);`,
			},
			{
				Statement: `insert into ma_test select x,x from generate_series(0,29) t(x);`,
			},
			{
				Statement: `create index on ma_test (b);`,
			},
			{
				Statement: `analyze ma_test;`,
			},
			{
				Statement: `prepare mt_q1 (int) as select a from ma_test where a >= $1 and a % 10 = 5 order by b;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute mt_q1(15);`,
				Results:   []sql.Row{{`Merge Append (actual rows=2 loops=1)`}, {`Sort Key: ma_test.b`}, {`Subplans Removed: 1`}, {`->  Index Scan using ma_test_p2_b_idx on ma_test_p2 ma_test_1 (actual rows=1 loops=1)`}, {`Filter: ((a >= $1) AND ((a % 10) = 5))`}, {`Rows Removed by Filter: 9`}, {`->  Index Scan using ma_test_p3_b_idx on ma_test_p3 ma_test_2 (actual rows=1 loops=1)`}, {`Filter: ((a >= $1) AND ((a % 10) = 5))`}, {`Rows Removed by Filter: 9`}},
			},
			{
				Statement: `execute mt_q1(15);`,
				Results:   []sql.Row{{15}, {25}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute mt_q1(25);`,
				Results:   []sql.Row{{`Merge Append (actual rows=1 loops=1)`}, {`Sort Key: ma_test.b`}, {`Subplans Removed: 2`}, {`->  Index Scan using ma_test_p3_b_idx on ma_test_p3 ma_test_1 (actual rows=1 loops=1)`}, {`Filter: ((a >= $1) AND ((a % 10) = 5))`}, {`Rows Removed by Filter: 9`}},
			},
			{
				Statement: `execute mt_q1(25);`,
				Results:   []sql.Row{{25}},
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) execute mt_q1(35);`,
				Results:   []sql.Row{{`Merge Append (actual rows=0 loops=1)`}, {`Sort Key: ma_test.b`}, {`Subplans Removed: 3`}},
			},
			{
				Statement: `execute mt_q1(35);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `deallocate mt_q1;`,
			},
			{
				Statement: `prepare mt_q2 (int) as select * from ma_test where a >= $1 order by b limit 1;`,
			},
			{
				Statement: `explain (analyze, verbose, costs off, summary off, timing off) execute mt_q2 (35);`,
				Results:   []sql.Row{{`Limit (actual rows=0 loops=1)`}, {`Output: ma_test.a, ma_test.b`}, {`->  Merge Append (actual rows=0 loops=1)`}, {`Sort Key: ma_test.b`}, {`Subplans Removed: 3`}},
			},
			{
				Statement: `deallocate mt_q2;`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off) select * from ma_test where a >= (select min(b) from ma_test_p2) order by b;`,
				Results:   []sql.Row{{`Merge Append (actual rows=20 loops=1)`}, {`Sort Key: ma_test.b`}, {`InitPlan 2 (returns $1)`}, {`->  Result (actual rows=1 loops=1)`}, {`InitPlan 1 (returns $0)`}, {`->  Limit (actual rows=1 loops=1)`}, {`->  Index Scan using ma_test_p2_b_idx on ma_test_p2 (actual rows=1 loops=1)`}, {`Index Cond: (b IS NOT NULL)`}, {`->  Index Scan using ma_test_p1_b_idx on ma_test_p1 ma_test_1 (never executed)`}, {`Filter: (a >= $1)`}, {`->  Index Scan using ma_test_p2_b_idx on ma_test_p2 ma_test_2 (actual rows=10 loops=1)`}, {`Filter: (a >= $1)`}, {`->  Index Scan using ma_test_p3_b_idx on ma_test_p3 ma_test_3 (actual rows=10 loops=1)`}, {`Filter: (a >= $1)`}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_sort;`,
			},
			{
				Statement: `drop table ma_test;`,
			},
			{
				Statement: `reset enable_indexonlyscan;`,
			},
			{
				Statement: `create table pp_arrpart (a int[]) partition by list (a);`,
			},
			{
				Statement: `create table pp_arrpart1 partition of pp_arrpart for values in ('{1}');`,
			},
			{
				Statement: `create table pp_arrpart2 partition of pp_arrpart for values in ('{2, 3}', '{4, 5}');`,
			},
			{
				Statement: `explain (costs off) select * from pp_arrpart where a = '{1}';`,
				Results:   []sql.Row{{`Seq Scan on pp_arrpart1 pp_arrpart`}, {`Filter: (a = '{1}'::integer[])`}},
			},
			{
				Statement: `explain (costs off) select * from pp_arrpart where a = '{1, 2}';`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off) select * from pp_arrpart where a in ('{4, 5}', '{1}');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on pp_arrpart1 pp_arrpart_1`}, {`Filter: ((a = '{4,5}'::integer[]) OR (a = '{1}'::integer[]))`}, {`->  Seq Scan on pp_arrpart2 pp_arrpart_2`}, {`Filter: ((a = '{4,5}'::integer[]) OR (a = '{1}'::integer[]))`}},
			},
			{
				Statement: `explain (costs off) update pp_arrpart set a = a where a = '{1}';`,
				Results:   []sql.Row{{`Update on pp_arrpart`}, {`Update on pp_arrpart1 pp_arrpart_1`}, {`->  Seq Scan on pp_arrpart1 pp_arrpart_1`}, {`Filter: (a = '{1}'::integer[])`}},
			},
			{
				Statement: `explain (costs off) delete from pp_arrpart where a = '{1}';`,
				Results:   []sql.Row{{`Delete on pp_arrpart`}, {`Delete on pp_arrpart1 pp_arrpart_1`}, {`->  Seq Scan on pp_arrpart1 pp_arrpart_1`}, {`Filter: (a = '{1}'::integer[])`}},
			},
			{
				Statement: `drop table pp_arrpart;`,
			},
			{
				Statement: `create table pph_arrpart (a int[]) partition by hash (a);`,
			},
			{
				Statement: `create table pph_arrpart1 partition of pph_arrpart for values with (modulus 2, remainder 0);`,
			},
			{
				Statement: `create table pph_arrpart2 partition of pph_arrpart for values with (modulus 2, remainder 1);`,
			},
			{
				Statement: `insert into pph_arrpart values ('{1}'), ('{1, 2}'), ('{4, 5}');`,
			},
			{
				Statement: `select tableoid::regclass, * from pph_arrpart order by 1;`,
				Results:   []sql.Row{{`pph_arrpart1`, `{1,2}`}, {`pph_arrpart1`, `{4,5}`}, {`pph_arrpart2`, `{1}`}},
			},
			{
				Statement: `explain (costs off) select * from pph_arrpart where a = '{1}';`,
				Results:   []sql.Row{{`Seq Scan on pph_arrpart2 pph_arrpart`}, {`Filter: (a = '{1}'::integer[])`}},
			},
			{
				Statement: `explain (costs off) select * from pph_arrpart where a = '{1, 2}';`,
				Results:   []sql.Row{{`Seq Scan on pph_arrpart1 pph_arrpart`}, {`Filter: (a = '{1,2}'::integer[])`}},
			},
			{
				Statement: `explain (costs off) select * from pph_arrpart where a in ('{4, 5}', '{1}');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on pph_arrpart1 pph_arrpart_1`}, {`Filter: ((a = '{4,5}'::integer[]) OR (a = '{1}'::integer[]))`}, {`->  Seq Scan on pph_arrpart2 pph_arrpart_2`}, {`Filter: ((a = '{4,5}'::integer[]) OR (a = '{1}'::integer[]))`}},
			},
			{
				Statement: `drop table pph_arrpart;`,
			},
			{
				Statement: `create type pp_colors as enum ('green', 'blue', 'black');`,
			},
			{
				Statement: `create table pp_enumpart (a pp_colors) partition by list (a);`,
			},
			{
				Statement: `create table pp_enumpart_green partition of pp_enumpart for values in ('green');`,
			},
			{
				Statement: `create table pp_enumpart_blue partition of pp_enumpart for values in ('blue');`,
			},
			{
				Statement: `explain (costs off) select * from pp_enumpart where a = 'blue';`,
				Results:   []sql.Row{{`Seq Scan on pp_enumpart_blue pp_enumpart`}, {`Filter: (a = 'blue'::pp_colors)`}},
			},
			{
				Statement: `explain (costs off) select * from pp_enumpart where a = 'black';`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `drop table pp_enumpart;`,
			},
			{
				Statement: `drop type pp_colors;`,
			},
			{
				Statement: `create type pp_rectype as (a int, b int);`,
			},
			{
				Statement: `create table pp_recpart (a pp_rectype) partition by list (a);`,
			},
			{
				Statement: `create table pp_recpart_11 partition of pp_recpart for values in ('(1,1)');`,
			},
			{
				Statement: `create table pp_recpart_23 partition of pp_recpart for values in ('(2,3)');`,
			},
			{
				Statement: `explain (costs off) select * from pp_recpart where a = '(1,1)'::pp_rectype;`,
				Results:   []sql.Row{{`Seq Scan on pp_recpart_11 pp_recpart`}, {`Filter: (a = '(1,1)'::pp_rectype)`}},
			},
			{
				Statement: `explain (costs off) select * from pp_recpart where a = '(1,2)'::pp_rectype;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `drop table pp_recpart;`,
			},
			{
				Statement: `drop type pp_rectype;`,
			},
			{
				Statement: `create table pp_intrangepart (a int4range) partition by list (a);`,
			},
			{
				Statement: `create table pp_intrangepart12 partition of pp_intrangepart for values in ('[1,2]');`,
			},
			{
				Statement: `create table pp_intrangepart2inf partition of pp_intrangepart for values in ('[2,)');`,
			},
			{
				Statement: `explain (costs off) select * from pp_intrangepart where a = '[1,2]'::int4range;`,
				Results:   []sql.Row{{`Seq Scan on pp_intrangepart12 pp_intrangepart`}, {`Filter: (a = '[1,3)'::int4range)`}},
			},
			{
				Statement: `explain (costs off) select * from pp_intrangepart where a = '(1,2)'::int4range;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `drop table pp_intrangepart;`,
			},
			{
				Statement: `create table pp_lp (a int, value int) partition by list (a);`,
			},
			{
				Statement: `create table pp_lp1 partition of pp_lp for values in(1);`,
			},
			{
				Statement: `create table pp_lp2 partition of pp_lp for values in(2);`,
			},
			{
				Statement: `explain (costs off) select * from pp_lp where a = 1;`,
				Results:   []sql.Row{{`Seq Scan on pp_lp1 pp_lp`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) update pp_lp set value = 10 where a = 1;`,
				Results:   []sql.Row{{`Update on pp_lp`}, {`Update on pp_lp1 pp_lp_1`}, {`->  Seq Scan on pp_lp1 pp_lp_1`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) delete from pp_lp where a = 1;`,
				Results:   []sql.Row{{`Delete on pp_lp`}, {`Delete on pp_lp1 pp_lp_1`}, {`->  Seq Scan on pp_lp1 pp_lp_1`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `set enable_partition_pruning = off;`,
			},
			{
				Statement: `set constraint_exclusion = 'partition'; -- this should not affect the result.`,
			},
			{
				Statement: `explain (costs off) select * from pp_lp where a = 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on pp_lp1 pp_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on pp_lp2 pp_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) update pp_lp set value = 10 where a = 1;`,
				Results:   []sql.Row{{`Update on pp_lp`}, {`Update on pp_lp1 pp_lp_1`}, {`Update on pp_lp2 pp_lp_2`}, {`->  Append`}, {`->  Seq Scan on pp_lp1 pp_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on pp_lp2 pp_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) delete from pp_lp where a = 1;`,
				Results:   []sql.Row{{`Delete on pp_lp`}, {`Delete on pp_lp1 pp_lp_1`}, {`Delete on pp_lp2 pp_lp_2`}, {`->  Append`}, {`->  Seq Scan on pp_lp1 pp_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on pp_lp2 pp_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `set constraint_exclusion = 'off'; -- this should not affect the result.`,
			},
			{
				Statement: `explain (costs off) select * from pp_lp where a = 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on pp_lp1 pp_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on pp_lp2 pp_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) update pp_lp set value = 10 where a = 1;`,
				Results:   []sql.Row{{`Update on pp_lp`}, {`Update on pp_lp1 pp_lp_1`}, {`Update on pp_lp2 pp_lp_2`}, {`->  Append`}, {`->  Seq Scan on pp_lp1 pp_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on pp_lp2 pp_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) delete from pp_lp where a = 1;`,
				Results:   []sql.Row{{`Delete on pp_lp`}, {`Delete on pp_lp1 pp_lp_1`}, {`Delete on pp_lp2 pp_lp_2`}, {`->  Append`}, {`->  Seq Scan on pp_lp1 pp_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on pp_lp2 pp_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `drop table pp_lp;`,
			},
			{
				Statement: `create table inh_lp (a int, value int);`,
			},
			{
				Statement: `create table inh_lp1 (a int, value int, check(a = 1)) inherits (inh_lp);`,
			},
			{
				Statement: `create table inh_lp2 (a int, value int, check(a = 2)) inherits (inh_lp);`,
			},
			{
				Statement: `set constraint_exclusion = 'partition';`,
			},
			{
				Statement: `explain (costs off) select * from inh_lp where a = 1;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on inh_lp inh_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on inh_lp1 inh_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) update inh_lp set value = 10 where a = 1;`,
				Results:   []sql.Row{{`Update on inh_lp`}, {`Update on inh_lp inh_lp_1`}, {`Update on inh_lp1 inh_lp_2`}, {`->  Result`}, {`->  Append`}, {`->  Seq Scan on inh_lp inh_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on inh_lp1 inh_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) delete from inh_lp where a = 1;`,
				Results:   []sql.Row{{`Delete on inh_lp`}, {`Delete on inh_lp inh_lp_1`}, {`Delete on inh_lp1 inh_lp_2`}, {`->  Append`}, {`->  Seq Scan on inh_lp inh_lp_1`}, {`Filter: (a = 1)`}, {`->  Seq Scan on inh_lp1 inh_lp_2`}, {`Filter: (a = 1)`}},
			},
			{
				Statement: `explain (costs off) update inh_lp1 set value = 10 where a = 2;`,
				Results:   []sql.Row{{`Update on inh_lp1`}, {`->  Seq Scan on inh_lp1`}, {`Filter: (a = 2)`}},
			},
			{
				Statement: `drop table inh_lp cascade;`,
			},
			{
				Statement: `reset enable_partition_pruning;`,
			},
			{
				Statement: `reset constraint_exclusion;`,
			},
			{
				Statement: `create temp table pp_temp_parent (a int) partition by list (a);`,
			},
			{
				Statement: `create temp table pp_temp_part_1 partition of pp_temp_parent for values in (1);`,
			},
			{
				Statement: `create temp table pp_temp_part_def partition of pp_temp_parent default;`,
			},
			{
				Statement: `explain (costs off) select * from pp_temp_parent where true;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on pp_temp_part_1 pp_temp_parent_1`}, {`->  Seq Scan on pp_temp_part_def pp_temp_parent_2`}},
			},
			{
				Statement: `explain (costs off) select * from pp_temp_parent where a = 2;`,
				Results:   []sql.Row{{`Seq Scan on pp_temp_part_def pp_temp_parent`}, {`Filter: (a = 2)`}},
			},
			{
				Statement: `drop table pp_temp_parent;`,
			},
			{
				Statement: `create temp table p (a int, b int, c int) partition by list (a);`,
			},
			{
				Statement: `create temp table p1 partition of p for values in (1);`,
			},
			{
				Statement: `create temp table p2 partition of p for values in (2);`,
			},
			{
				Statement: `create temp table q (a int, b int, c int) partition by list (a);`,
			},
			{
				Statement: `create temp table q1 partition of q for values in (1) partition by list (b);`,
			},
			{
				Statement: `create temp table q11 partition of q1 for values in (1) partition by list (c);`,
			},
			{
				Statement: `create temp table q111 partition of q11 for values in (1);`,
			},
			{
				Statement: `create temp table q2 partition of q for values in (2) partition by list (b);`,
			},
			{
				Statement: `create temp table q21 partition of q2 for values in (1);`,
			},
			{
				Statement: `create temp table q22 partition of q2 for values in (2);`,
			},
			{
				Statement: `insert into q22 values (2, 2, 3);`,
			},
			{
				Statement: `explain (costs off)
select *
from (
      select * from p
      union all
      select * from q1
      union all
      select 1, 1, 1
     ) s(a, b, c)
where s.a = 1 and s.b = 1 and s.c = (select 1);`,
				Results: []sql.Row{{`Append`}, {`InitPlan 1 (returns $0)`}, {`->  Result`}, {`->  Seq Scan on p1 p`}, {`Filter: ((a = 1) AND (b = 1) AND (c = $0))`}, {`->  Seq Scan on q111 q1`}, {`Filter: ((a = 1) AND (b = 1) AND (c = $0))`}, {`->  Result`}, {`One-Time Filter: (1 = $0)`}},
			},
			{
				Statement: `select *
from (
      select * from p
      union all
      select * from q1
      union all
      select 1, 1, 1
     ) s(a, b, c)
where s.a = 1 and s.b = 1 and s.c = (select 1);`,
				Results: []sql.Row{{1, 1, 1}},
			},
			{
				Statement: `prepare q (int, int) as
select *
from (
      select * from p
      union all
      select * from q1
      union all
      select 1, 1, 1
     ) s(a, b, c)
where s.a = $1 and s.b = $2 and s.c = (select 1);`,
			},
			{
				Statement: `explain (costs off) execute q (1, 1);`,
				Results:   []sql.Row{{`Append`}, {`Subplans Removed: 1`}, {`InitPlan 1 (returns $0)`}, {`->  Result`}, {`->  Seq Scan on p1 p`}, {`Filter: ((a = $1) AND (b = $2) AND (c = $0))`}, {`->  Seq Scan on q111 q1`}, {`Filter: ((a = $1) AND (b = $2) AND (c = $0))`}, {`->  Result`}, {`One-Time Filter: ((1 = $1) AND (1 = $2) AND (1 = $0))`}},
			},
			{
				Statement: `execute q (1, 1);`,
				Results:   []sql.Row{{1, 1, 1}},
			},
			{
				Statement: `drop table p, q;`,
			},
			{
				Statement: `create table listp (a int, b int) partition by list (a);`,
			},
			{
				Statement: `create table listp1 partition of listp for values in(1);`,
			},
			{
				Statement: `create table listp2 partition of listp for values in(2) partition by list(b);`,
			},
			{
				Statement: `create table listp2_10 partition of listp2 for values in (10);`,
			},
			{
				Statement: `explain (analyze, costs off, summary off, timing off)
select * from listp where a = (select 2) and b <> 10;`,
				Results: []sql.Row{{`Seq Scan on listp1 listp (actual rows=0 loops=1)`}, {`Filter: ((b <> 10) AND (a = $0))`}, {`InitPlan 1 (returns $0)`}, {`->  Result (never executed)`}},
			},
			{
				Statement: `set enable_partition_pruning to off;`,
			},
			{
				Statement: `set constraint_exclusion to 'partition';`,
			},
			{
				Statement: `explain (costs off) select * from listp1 where a = 2;`,
				Results:   []sql.Row{{`Seq Scan on listp1`}, {`Filter: (a = 2)`}},
			},
			{
				Statement: `explain (costs off) update listp1 set a = 1 where a = 2;`,
				Results:   []sql.Row{{`Update on listp1`}, {`->  Seq Scan on listp1`}, {`Filter: (a = 2)`}},
			},
			{
				Statement: `set constraint_exclusion to 'on';`,
			},
			{
				Statement: `explain (costs off) select * from listp1 where a = 2;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off) update listp1 set a = 1 where a = 2;`,
				Results:   []sql.Row{{`Update on listp1`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `reset constraint_exclusion;`,
			},
			{
				Statement: `reset enable_partition_pruning;`,
			},
			{
				Statement: `drop table listp;`,
			},
			{
				Statement: `set parallel_setup_cost to 0;`,
			},
			{
				Statement: `set parallel_tuple_cost to 0;`,
			},
			{
				Statement: `create table listp (a int) partition by list(a);`,
			},
			{
				Statement: `create table listp_12 partition of listp for values in(1,2) partition by list(a);`,
			},
			{
				Statement: `create table listp_12_1 partition of listp_12 for values in(1);`,
			},
			{
				Statement: `create table listp_12_2 partition of listp_12 for values in(2);`,
			},
			{
				Statement: `alter table listp_12_1 set (parallel_workers = 0);`,
			},
			{
				Statement: `select explain_parallel_append('select * from listp where a = (select 1);');`,
				Results:   []sql.Row{{`Gather (actual rows=N loops=N)`}, {`Workers Planned: 2`}, {`Params Evaluated: $0`}, {`Workers Launched: N`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=N loops=N)`}, {`->  Parallel Append (actual rows=N loops=N)`}, {`->  Seq Scan on listp_12_1 listp_1 (actual rows=N loops=N)`}, {`Filter: (a = $0)`}, {`->  Parallel Seq Scan on listp_12_2 listp_2 (never executed)`}, {`Filter: (a = $0)`}},
			},
			{
				Statement: `select explain_parallel_append(
'select * from listp where a = (select 1)
  union all
select * from listp where a = (select 2);');`,
				Results: []sql.Row{{`Append (actual rows=N loops=N)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 2`}, {`Params Evaluated: $0`}, {`Workers Launched: N`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=N loops=N)`}, {`->  Parallel Append (actual rows=N loops=N)`}, {`->  Seq Scan on listp_12_1 listp_1 (actual rows=N loops=N)`}, {`Filter: (a = $0)`}, {`->  Parallel Seq Scan on listp_12_2 listp_2 (never executed)`}, {`Filter: (a = $0)`}, {`->  Gather (actual rows=N loops=N)`}, {`Workers Planned: 2`}, {`Params Evaluated: $1`}, {`Workers Launched: N`}, {`InitPlan 2 (returns $1)`}, {`->  Result (actual rows=N loops=N)`}, {`->  Parallel Append (actual rows=N loops=N)`}, {`->  Seq Scan on listp_12_1 listp_4 (never executed)`}, {`Filter: (a = $1)`}, {`->  Parallel Seq Scan on listp_12_2 listp_5 (actual rows=N loops=N)`}, {`Filter: (a = $1)`}},
			},
			{
				Statement: `drop table listp;`,
			},
			{
				Statement: `reset parallel_tuple_cost;`,
			},
			{
				Statement: `reset parallel_setup_cost;`,
			},
			{
				Statement: `set enable_sort to 0;`,
			},
			{
				Statement: `create table rangep (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table rangep_0_to_100 partition of rangep for values from (0) to (100) partition by list (b);`,
			},
			{
				Statement: `create table rangep_0_to_100_1 partition of rangep_0_to_100 for values in(1);`,
			},
			{
				Statement: `create table rangep_0_to_100_2 partition of rangep_0_to_100 for values in(2);`,
			},
			{
				Statement: `create table rangep_0_to_100_3 partition of rangep_0_to_100 for values in(3);`,
			},
			{
				Statement: `create table rangep_100_to_200 partition of rangep for values from (100) to (200);`,
			},
			{
				Statement: `create index on rangep (a);`,
			},
			{
				Statement: `explain (analyze on, costs off, timing off, summary off)
select * from rangep where b IN((select 1),(select 2)) order by a;`,
				Results: []sql.Row{{`Append (actual rows=0 loops=1)`}, {`InitPlan 1 (returns $0)`}, {`->  Result (actual rows=1 loops=1)`}, {`InitPlan 2 (returns $1)`}, {`->  Result (actual rows=1 loops=1)`}, {`->  Merge Append (actual rows=0 loops=1)`}, {`Sort Key: rangep_2.a`}, {`->  Index Scan using rangep_0_to_100_1_a_idx on rangep_0_to_100_1 rangep_2 (actual rows=0 loops=1)`}, {`Filter: (b = ANY (ARRAY[$0, $1]))`}, {`->  Index Scan using rangep_0_to_100_2_a_idx on rangep_0_to_100_2 rangep_3 (actual rows=0 loops=1)`}, {`Filter: (b = ANY (ARRAY[$0, $1]))`}, {`->  Index Scan using rangep_0_to_100_3_a_idx on rangep_0_to_100_3 rangep_4 (never executed)`}, {`Filter: (b = ANY (ARRAY[$0, $1]))`}, {`->  Index Scan using rangep_100_to_200_a_idx on rangep_100_to_200 rangep_5 (actual rows=0 loops=1)`}, {`Filter: (b = ANY (ARRAY[$0, $1]))`}},
			},
			{
				Statement: `reset enable_sort;`,
			},
			{
				Statement: `drop table rangep;`,
			},
			{
				Statement: `create table rp_prefix_test1 (a int, b varchar) partition by range(a, b);`,
			},
			{
				Statement: `create table rp_prefix_test1_p1 partition of rp_prefix_test1 for values from (1, 'a') to (1, 'b');`,
			},
			{
				Statement: `create table rp_prefix_test1_p2 partition of rp_prefix_test1 for values from (2, 'a') to (2, 'b');`,
			},
			{
				Statement: `explain (costs off) select * from rp_prefix_test1 where a <= 1 and b = 'a';`,
				Results:   []sql.Row{{`Seq Scan on rp_prefix_test1_p1 rp_prefix_test1`}, {`Filter: ((a <= 1) AND ((b)::text = 'a'::text))`}},
			},
			{
				Statement: `create table rp_prefix_test2 (a int, b int, c int) partition by range(a, b, c);`,
			},
			{
				Statement: `create table rp_prefix_test2_p1 partition of rp_prefix_test2 for values from (1, 1, 0) to (1, 1, 10);`,
			},
			{
				Statement: `create table rp_prefix_test2_p2 partition of rp_prefix_test2 for values from (2, 2, 0) to (2, 2, 10);`,
			},
			{
				Statement: `explain (costs off) select * from rp_prefix_test2 where a <= 1 and b = 1 and c >= 0;`,
				Results:   []sql.Row{{`Seq Scan on rp_prefix_test2_p1 rp_prefix_test2`}, {`Filter: ((a <= 1) AND (c >= 0) AND (b = 1))`}},
			},
			{
				Statement: `create table rp_prefix_test3 (a int, b int, c int, d int) partition by range(a, b, c, d);`,
			},
			{
				Statement: `create table rp_prefix_test3_p1 partition of rp_prefix_test3 for values from (1, 1, 1, 0) to (1, 1, 1, 10);`,
			},
			{
				Statement: `create table rp_prefix_test3_p2 partition of rp_prefix_test3 for values from (2, 2, 2, 0) to (2, 2, 2, 10);`,
			},
			{
				Statement: `explain (costs off) select * from rp_prefix_test3 where a >= 1 and b >= 1 and b >= 2 and c >= 2 and d >= 0;`,
				Results:   []sql.Row{{`Seq Scan on rp_prefix_test3_p2 rp_prefix_test3`}, {`Filter: ((a >= 1) AND (b >= 1) AND (b >= 2) AND (c >= 2) AND (d >= 0))`}},
			},
			{
				Statement: `explain (costs off) select * from rp_prefix_test3 where a >= 1 and b >= 1 and b = 2 and c = 2 and d >= 0;`,
				Results:   []sql.Row{{`Seq Scan on rp_prefix_test3_p2 rp_prefix_test3`}, {`Filter: ((a >= 1) AND (b >= 1) AND (d >= 0) AND (b = 2) AND (c = 2))`}},
			},
			{
				Statement: `create table hp_prefix_test (a int, b int, c int, d int) partition by hash (a part_test_int4_ops, b part_test_int4_ops, c part_test_int4_ops, d part_test_int4_ops);`,
			},
			{
				Statement: `create table hp_prefix_test_p1 partition of hp_prefix_test for values with (modulus 2, remainder 0);`,
			},
			{
				Statement: `create table hp_prefix_test_p2 partition of hp_prefix_test for values with (modulus 2, remainder 1);`,
			},
			{
				Statement: `explain (costs off) select * from hp_prefix_test where a = 1 and b is null and c = 1 and d = 1;`,
				Results:   []sql.Row{{`Seq Scan on hp_prefix_test_p1 hp_prefix_test`}, {`Filter: ((b IS NULL) AND (a = 1) AND (c = 1) AND (d = 1))`}},
			},
			{
				Statement: `drop table rp_prefix_test1;`,
			},
			{
				Statement: `drop table rp_prefix_test2;`,
			},
			{
				Statement: `drop table rp_prefix_test3;`,
			},
			{
				Statement: `drop table hp_prefix_test;`,
			},
			{
				Statement: `create operator === (
   leftarg = int4,
   rightarg = int4,
   procedure = int4eq,
   commutator = ===,
   hashes
);`,
			},
			{
				Statement: `create operator class part_test_int4_ops2
for type int4
using hash as
operator 1 ===,
function 2 part_hashint4_noop(int4, int8);`,
			},
			{
				Statement: `create table hp_contradict_test (a int, b int) partition by hash (a part_test_int4_ops2, b part_test_int4_ops2);`,
			},
			{
				Statement: `create table hp_contradict_test_p1 partition of hp_contradict_test for values with (modulus 2, remainder 0);`,
			},
			{
				Statement: `create table hp_contradict_test_p2 partition of hp_contradict_test for values with (modulus 2, remainder 1);`,
			},
			{
				Statement: `explain (costs off) select * from hp_contradict_test where a is null and a === 1 and b === 1;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `explain (costs off) select * from hp_contradict_test where a === 1 and b === 1 and a is null;`,
				Results:   []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `drop table hp_contradict_test;`,
			},
			{
				Statement: `drop operator class part_test_int4_ops2 using hash;`,
			},
			{
				Statement: `drop operator ===(int4, int4);`,
			},
		},
	})
}
