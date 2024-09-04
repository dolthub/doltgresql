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

func TestIndexing(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_indexing)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_indexing,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create table idxpart (a int, b int, c text) partition by range (a);`,
			},
			{
				Statement: `create index idxpart_idx on idxpart (a);`,
			},
			{
				Statement: `select relhassubclass from pg_class where relname = 'idxpart_idx';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select indexdef from pg_indexes where indexname like 'idxpart_idx%';`,
				Results:   []sql.Row{{`CREATE INDEX idxpart_idx ON ONLY public.idxpart USING btree (a)`}},
			},
			{
				Statement: `drop index idxpart_idx;`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0) to (10);`,
			},
			{
				Statement: `create table idxpart2 partition of idxpart for values from (10) to (100)
	partition by range (b);`,
			},
			{
				Statement: `create table idxpart21 partition of idxpart2 for values from (0) to (100);`,
			},
			{
				Statement: `create index idxpart_idx on only idxpart(a);`,
			},
			{
				Statement: `select relhassubclass from pg_class where relname = 'idxpart_idx';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `drop index idxpart_idx;`,
			},
			{
				Statement: `create index on idxpart (a);`,
			},
			{
				Statement: `select relname, relkind, relhassubclass, inhparent::regclass
    from pg_class left join pg_index ix on (indexrelid = oid)
	left join pg_inherits on (ix.indexrelid = inhrelid)
	where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart`, `p`, true, ``}, {`idxpart1`, `r`, false, ``}, {`idxpart1_a_idx`, `i`, false, `idxpart_a_idx`}, {`idxpart2`, `p`, true, ``}, {`idxpart21`, `r`, false, ``}, {`idxpart21_a_idx`, `i`, false, `idxpart2_a_idx`}, {`idxpart2_a_idx`, `I`, true, `idxpart_a_idx`}, {`idxpart_a_idx`, `I`, true, ``}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int, c text) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0) to (10);`,
			},
			{
				Statement:   `create index concurrently on idxpart (a);`,
				ErrorString: `cannot create index on partitioned table "idxpart" concurrently`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `CREATE TABLE idxpart (col1 INT) PARTITION BY RANGE (col1);`,
			},
			{
				Statement: `CREATE INDEX ON idxpart (col1);`,
			},
			{
				Statement: `CREATE TABLE idxpart_two (col2 INT);`,
			},
			{
				Statement: `SELECT col2 FROM idxpart_two fk LEFT OUTER JOIN idxpart pk ON (col1 = col2);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP table idxpart, idxpart_two;`,
			},
			{
				Statement: `CREATE TABLE idxpart (a INT, b TEXT, c INT) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE idxpart1 PARTITION OF idxpart FOR VALUES FROM (MINVALUE) TO (MAXVALUE);`,
			},
			{
				Statement: `CREATE INDEX partidx_abc_idx ON idxpart (a, b, c);`,
			},
			{
				Statement: `INSERT INTO idxpart (a, b, c) SELECT i, i, i FROM generate_series(1, 50) i;`,
			},
			{
				Statement: `ALTER TABLE idxpart ALTER COLUMN c TYPE numeric;`,
			},
			{
				Statement: `DROP TABLE idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int, c text) partition by range (a);`,
			},
			{
				Statement: `create index idxparti on idxpart (a);`,
			},
			{
				Statement: `create index idxparti2 on idxpart (b, c);`,
			},
			{
				Statement: `create table idxpart1 (like idxpart);`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | text    |           |          | 
alter table idxpart attach partition idxpart1 for values from (0) to (10);`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | text    |           |          | 
Partition of: idxpart FOR VALUES FROM (0) TO (10)
Indexes:
    "idxpart1_a_idx" btree (a)
    "idxpart1_b_c_idx" btree (b, c)
\d+ idxpart1_a_idx
                 Index "public.idxpart1_a_idx"
 Column |  Type   | Key? | Definition | Storage | Stats target 
--------+---------+------+------------+---------+--------------
 a      | integer | yes  | a          | plain   | 
Partition of: idxparti 
No partition constraint
btree, for table "public.idxpart1"
\d+ idxpart1_b_c_idx
                Index "public.idxpart1_b_c_idx"
 Column |  Type   | Key? | Definition | Storage  | Stats target 
--------+---------+------+------------+----------+--------------
 b      | integer | yes  | b          | plain    | 
 c      | text    | yes  | c          | extended | 
Partition of: idxparti2 
No partition constraint
btree, for table "public.idxpart1"
create index idxpart_c on only idxpart (c);`,
			},
			{
				Statement: `create index idxpart1_c on idxpart1 (c);`,
			},
			{
				Statement:   `alter table idxpart_c attach partition idxpart1_c for values from (10) to (20);`,
				ErrorString: `"idxpart_c" is not a partitioned table`,
			},
			{
				Statement: `alter index idxpart_c attach partition idxpart1_c;`,
			},
			{
				Statement: `select relname, relpartbound from pg_class
  where relname in ('idxpart_c', 'idxpart1_c')
  order by relname;`,
				Results: []sql.Row{{`idxpart1_c`, ``}, {`idxpart_c`, ``}},
			},
			{
				Statement:   `alter table idxpart_c detach partition idxpart1_c;`,
				ErrorString: `ALTER action DETACH PARTITION cannot be performed on relation "idxpart_c"`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a, b);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0, 0) to (10, 10);`,
			},
			{
				Statement: `create index on idxpart1 (a, b);`,
			},
			{
				Statement: `create index on idxpart (a, b);`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: idxpart FOR VALUES FROM (0, 0) TO (10, 10)
Indexes:
    "idxpart1_a_b_idx" btree (a, b)
select relname, relkind, relhassubclass, inhparent::regclass
    from pg_class left join pg_index ix on (indexrelid = oid)
	left join pg_inherits on (ix.indexrelid = inhrelid)
	where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart`, `p`, true, ``}, {`idxpart1`, `r`, false, ``}, {`idxpart1_a_b_idx`, `i`, false, `idxpart_a_b_idx`}, {`idxpart_a_b_idx`, `I`, true, ``}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int) partition by range (a);`,
			},
			{
				Statement: `create index on idxpart (a);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0) to (10);`,
			},
			{
				Statement:   `drop index idxpart1_a_idx;	-- no way`,
				ErrorString: `cannot drop index idxpart1_a_idx because index idxpart_a_idx requires it`,
			},
			{
				Statement:   `drop index concurrently idxpart_a_idx;	-- unsupported`,
				ErrorString: `cannot drop partitioned index "idxpart_a_idx" concurrently`,
			},
			{
				Statement: `drop index idxpart_a_idx;	-- both indexes go away`,
			},
			{
				Statement: `select relname, relkind from pg_class
  where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart`, `p`}, {`idxpart1`, `r`}},
			},
			{
				Statement: `create index on idxpart (a);`,
			},
			{
				Statement: `drop table idxpart1;		-- the index on partition goes away too`,
			},
			{
				Statement: `select relname, relkind from pg_class
  where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart`, `p`}, {`idxpart_a_idx`, `I`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create temp table idxpart_temp (a int) partition by range (a);`,
			},
			{
				Statement: `create index on idxpart_temp(a);`,
			},
			{
				Statement: `create temp table idxpart1_temp partition of idxpart_temp
  for values from (0) to (10);`,
			},
			{
				Statement:   `drop index idxpart1_temp_a_idx; -- error`,
				ErrorString: `cannot drop index idxpart1_temp_a_idx because index idxpart_temp_a_idx requires it`,
			},
			{
				Statement: `drop index concurrently idxpart_temp_a_idx;`,
			},
			{
				Statement: `select relname, relkind from pg_class
  where relname like 'idxpart_temp%' order by relname;`,
				Results: []sql.Row{{`idxpart_temp`, `p`}},
			},
			{
				Statement: `drop table idxpart_temp;`,
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a, b);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0, 0) to (10, 10);`,
			},
			{
				Statement: `create index idxpart_a_b_idx on only idxpart (a, b);`,
			},
			{
				Statement: `create index idxpart1_a_b_idx on idxpart1 (a, b);`,
			},
			{
				Statement: `create index idxpart1_tst1 on idxpart1 (b, a);`,
			},
			{
				Statement: `create index idxpart1_tst2 on idxpart1 using hash (a);`,
			},
			{
				Statement: `create index idxpart1_tst3 on idxpart1 (a, b) where a > 10;`,
			},
			{
				Statement:   `alter index idxpart attach partition idxpart1;`,
				ErrorString: `"idxpart" is not an index`,
			},
			{
				Statement:   `alter index idxpart_a_b_idx attach partition idxpart1;`,
				ErrorString: `"idxpart1" is not an index`,
			},
			{
				Statement:   `alter index idxpart_a_b_idx attach partition idxpart_a_b_idx;`,
				ErrorString: `cannot attach index "idxpart_a_b_idx" as a partition of index "idxpart_a_b_idx"`,
			},
			{
				Statement:   `alter index idxpart_a_b_idx attach partition idxpart1_b_idx;`,
				ErrorString: `relation "idxpart1_b_idx" does not exist`,
			},
			{
				Statement:   `alter index idxpart_a_b_idx attach partition idxpart1_tst1;`,
				ErrorString: `cannot attach index "idxpart1_tst1" as a partition of index "idxpart_a_b_idx"`,
			},
			{
				Statement:   `alter index idxpart_a_b_idx attach partition idxpart1_tst2;`,
				ErrorString: `cannot attach index "idxpart1_tst2" as a partition of index "idxpart_a_b_idx"`,
			},
			{
				Statement:   `alter index idxpart_a_b_idx attach partition idxpart1_tst3;`,
				ErrorString: `cannot attach index "idxpart1_tst3" as a partition of index "idxpart_a_b_idx"`,
			},
			{
				Statement: `alter index idxpart_a_b_idx attach partition idxpart1_a_b_idx;`,
			},
			{
				Statement: `alter index idxpart_a_b_idx attach partition idxpart1_a_b_idx; -- quiet`,
			},
			{
				Statement: `create index idxpart1_2_a_b on idxpart1 (a, b);`,
			},
			{
				Statement:   `alter index idxpart_a_b_idx attach partition idxpart1_2_a_b;`,
				ErrorString: `cannot attach index "idxpart1_2_a_b" as a partition of index "idxpart_a_b_idx"`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `select indexrelid::regclass, indrelid::regclass
  from pg_index where indexrelid::regclass::text like 'idxpart%';`,
				Results: []sql.Row{},
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (a int, b int);`,
			},
			{
				Statement: `create index on idxpart1 using hash (a);`,
			},
			{
				Statement: `create index on idxpart1 (a) where b > 1;`,
			},
			{
				Statement: `create index on idxpart1 ((a + 0));`,
			},
			{
				Statement: `create index on idxpart1 (a, a);`,
			},
			{
				Statement: `create index on idxpart (a);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0) to (1000);`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: idxpart FOR VALUES FROM (0) TO (1000)
Indexes:
    "idxpart1_a_a1_idx" btree (a, a)
    "idxpart1_a_idx" hash (a)
    "idxpart1_a_idx1" btree (a) WHERE b > 1
    "idxpart1_a_idx2" btree (a)
    "idxpart1_expr_idx" btree ((a + 0))
drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0) to (100);`,
			},
			{
				Statement: `create table idxpart2 partition of idxpart for values from (100) to (1000)
  partition by range (a);`,
			},
			{
				Statement: `create table idxpart21 partition of idxpart2 for values from (100) to (200);`,
			},
			{
				Statement: `create table idxpart22 partition of idxpart2 for values from (200) to (300);`,
			},
			{
				Statement: `create index on idxpart22 (a);`,
			},
			{
				Statement: `create index on only idxpart2 (a);`,
			},
			{
				Statement: `create index on idxpart (a);`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: idxpart FOR VALUES FROM (0) TO (100)
Indexes:
    "idxpart1_a_idx" btree (a)
\d idxpart2
        Partitioned table "public.idxpart2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: idxpart FOR VALUES FROM (100) TO (1000)
Partition key: RANGE (a)
Indexes:
    "idxpart2_a_idx" btree (a) INVALID
Number of partitions: 2 (Use \d+ to list them.)
\d idxpart21
             Table "public.idxpart21"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: idxpart2 FOR VALUES FROM (100) TO (200)
select indexrelid::regclass, indrelid::regclass, inhparent::regclass
  from pg_index idx left join pg_inherits inh on (idx.indexrelid = inh.inhrelid)
where indexrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart1_a_idx`, `idxpart1`, `idxpart_a_idx`}, {`idxpart22_a_idx`, `idxpart22`, ``}, {`idxpart2_a_idx`, `idxpart2`, `idxpart_a_idx`}, {`idxpart_a_idx`, `idxpart`, ``}},
			},
			{
				Statement: `alter index idxpart2_a_idx attach partition idxpart22_a_idx;`,
			},
			{
				Statement: `select indexrelid::regclass, indrelid::regclass, inhparent::regclass
  from pg_index idx left join pg_inherits inh on (idx.indexrelid = inh.inhrelid)
where indexrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart1_a_idx`, `idxpart1`, `idxpart_a_idx`}, {`idxpart22_a_idx`, `idxpart22`, `idxpart2_a_idx`}, {`idxpart2_a_idx`, `idxpart2`, `idxpart_a_idx`}, {`idxpart_a_idx`, `idxpart`, ``}},
			},
			{
				Statement: `alter index idxpart2_a_idx attach partition idxpart22_a_idx;`,
			},
			{
				Statement: `\d idxpart2
        Partitioned table "public.idxpart2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: idxpart FOR VALUES FROM (100) TO (1000)
Partition key: RANGE (a)
Indexes:
    "idxpart2_a_idx" btree (a) INVALID
Number of partitions: 2 (Use \d+ to list them.)
create index on idxpart21 (a);`,
			},
			{
				Statement: `alter index idxpart2_a_idx attach partition idxpart21_a_idx;`,
			},
			{
				Statement: `\d idxpart2
        Partitioned table "public.idxpart2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: idxpart FOR VALUES FROM (100) TO (1000)
Partition key: RANGE (a)
Indexes:
    "idxpart2_a_idx" btree (a)
Number of partitions: 2 (Use \d+ to list them.)
drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int, c text, d bool) partition by range (a);`,
			},
			{
				Statement: `create index idxparti on idxpart (a);`,
			},
			{
				Statement: `create index idxparti2 on idxpart (b, c);`,
			},
			{
				Statement: `create table idxpart1 (like idxpart including indexes);`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | text    |           |          | 
 d      | boolean |           |          | 
Indexes:
    "idxpart1_a_idx" btree (a)
    "idxpart1_b_c_idx" btree (b, c)
select relname, relkind, inhparent::regclass
    from pg_class left join pg_index ix on (indexrelid = oid)
	left join pg_inherits on (ix.indexrelid = inhrelid)
	where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart`, `p`, ``}, {`idxpart1`, `r`, ``}, {`idxpart1_a_idx`, `i`, ``}, {`idxpart1_b_c_idx`, `i`, ``}, {`idxparti`, `I`, ``}, {`idxparti2`, `I`, ``}},
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0) to (10);`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | text    |           |          | 
 d      | boolean |           |          | 
Partition of: idxpart FOR VALUES FROM (0) TO (10)
Indexes:
    "idxpart1_a_idx" btree (a)
    "idxpart1_b_c_idx" btree (b, c)
select relname, relkind, inhparent::regclass
    from pg_class left join pg_index ix on (indexrelid = oid)
	left join pg_inherits on (ix.indexrelid = inhrelid)
	where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart`, `p`, ``}, {`idxpart1`, `r`, ``}, {`idxpart1_a_idx`, `i`, `idxparti`}, {`idxpart1_b_c_idx`, `i`, `idxparti2`}, {`idxparti`, `I`, ``}, {`idxparti2`, `I`, ``}},
			},
			{
				Statement: `create index on idxpart1 ((a+b)) where d = true;`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | text    |           |          | 
 d      | boolean |           |          | 
Partition of: idxpart FOR VALUES FROM (0) TO (10)
Indexes:
    "idxpart1_a_idx" btree (a)
    "idxpart1_b_c_idx" btree (b, c)
    "idxpart1_expr_idx" btree ((a + b)) WHERE d = true
select relname, relkind, inhparent::regclass
    from pg_class left join pg_index ix on (indexrelid = oid)
	left join pg_inherits on (ix.indexrelid = inhrelid)
	where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart`, `p`, ``}, {`idxpart1`, `r`, ``}, {`idxpart1_a_idx`, `i`, `idxparti`}, {`idxpart1_b_c_idx`, `i`, `idxparti2`}, {`idxpart1_expr_idx`, `i`, ``}, {`idxparti`, `I`, ``}, {`idxparti2`, `I`, ``}},
			},
			{
				Statement: `create index idxparti3 on idxpart ((a+b)) where d = true;`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | text    |           |          | 
 d      | boolean |           |          | 
Partition of: idxpart FOR VALUES FROM (0) TO (10)
Indexes:
    "idxpart1_a_idx" btree (a)
    "idxpart1_b_c_idx" btree (b, c)
    "idxpart1_expr_idx" btree ((a + b)) WHERE d = true
select relname, relkind, inhparent::regclass
    from pg_class left join pg_index ix on (indexrelid = oid)
	left join pg_inherits on (ix.indexrelid = inhrelid)
	where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart`, `p`, ``}, {`idxpart1`, `r`, ``}, {`idxpart1_a_idx`, `i`, `idxparti`}, {`idxpart1_b_c_idx`, `i`, `idxparti2`}, {`idxpart1_expr_idx`, `i`, `idxparti3`}, {`idxparti`, `I`, ``}, {`idxparti2`, `I`, ``}, {`idxparti3`, `I`, ``}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (1) to (1000) partition by range (a);`,
			},
			{
				Statement: `create table idxpart11 partition of idxpart1 for values from (1) to (100);`,
			},
			{
				Statement: `create index on only idxpart1 (a);`,
			},
			{
				Statement: `create index on only idxpart (a);`,
			},
			{
				Statement: `select relname, indisvalid from pg_class join pg_index on indexrelid = oid
   where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart1_a_idx`, false}, {`idxpart_a_idx`, false}},
			},
			{
				Statement: `alter index idxpart_a_idx attach partition idxpart1_a_idx;`,
			},
			{
				Statement: `select relname, indisvalid from pg_class join pg_index on indexrelid = oid
   where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart1_a_idx`, false}, {`idxpart_a_idx`, false}},
			},
			{
				Statement: `create index on idxpart11 (a);`,
			},
			{
				Statement: `alter index idxpart1_a_idx attach partition idxpart11_a_idx;`,
			},
			{
				Statement: `select relname, indisvalid from pg_class join pg_index on indexrelid = oid
   where relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart11_a_idx`, true}, {`idxpart1_a_idx`, true}, {`idxpart_a_idx`, true}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (like idxpart);`,
			},
			{
				Statement: `create index on idxpart1 (a);`,
			},
			{
				Statement: `create index on idxpart (a);`,
			},
			{
				Statement: `create table idxpart2 (like idxpart);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0000) to (1000);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from (1000) to (2000);`,
			},
			{
				Statement: `create table idxpart3 partition of idxpart for values from (2000) to (3000);`,
			},
			{
				Statement: `select relname, relkind from pg_class where relname like 'idxpart%' order by relname;`,
				Results:   []sql.Row{{`idxpart`, `p`}, {`idxpart1`, `r`}, {`idxpart1_a_idx`, `i`}, {`idxpart2`, `r`}, {`idxpart2_a_idx`, `i`}, {`idxpart3`, `r`}, {`idxpart3_a_idx`, `i`}, {`idxpart_a_idx`, `I`}},
			},
			{
				Statement: `alter table idxpart detach partition idxpart1;`,
			},
			{
				Statement: `alter table idxpart detach partition idxpart2;`,
			},
			{
				Statement: `alter table idxpart detach partition idxpart3;`,
			},
			{
				Statement: `drop index idxpart1_a_idx;`,
			},
			{
				Statement: `drop index idxpart2_a_idx;`,
			},
			{
				Statement: `drop index idxpart3_a_idx;`,
			},
			{
				Statement: `select relname, relkind from pg_class where relname like 'idxpart%' order by relname;`,
				Results:   []sql.Row{{`idxpart`, `p`}, {`idxpart1`, `r`}, {`idxpart2`, `r`}, {`idxpart3`, `r`}, {`idxpart_a_idx`, `I`}},
			},
			{
				Statement: `drop table idxpart, idxpart1, idxpart2, idxpart3;`,
			},
			{
				Statement: `select relname, relkind from pg_class where relname like 'idxpart%' order by relname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `create table idxpart (a int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (like idxpart);`,
			},
			{
				Statement: `create index on idxpart1 (a);`,
			},
			{
				Statement: `create index on idxpart (a);`,
			},
			{
				Statement: `create table idxpart2 (like idxpart);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0000) to (1000);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from (1000) to (2000);`,
			},
			{
				Statement: `create table idxpart3 partition of idxpart for values from (2000) to (3000);`,
			},
			{
				Statement: `select relname, relkind from pg_class where relname like 'idxpart%' order by relname;`,
				Results:   []sql.Row{{`idxpart`, `p`}, {`idxpart1`, `r`}, {`idxpart1_a_idx`, `i`}, {`idxpart2`, `r`}, {`idxpart2_a_idx`, `i`}, {`idxpart3`, `r`}, {`idxpart3_a_idx`, `i`}, {`idxpart_a_idx`, `I`}},
			},
			{
				Statement: `alter table idxpart detach partition idxpart1;`,
			},
			{
				Statement: `alter table idxpart detach partition idxpart2;`,
			},
			{
				Statement: `alter table idxpart detach partition idxpart3;`,
			},
			{
				Statement: `drop index idxpart_a_idx;`,
			},
			{
				Statement: `select relname, relkind from pg_class where relname like 'idxpart%' order by relname;`,
				Results:   []sql.Row{{`idxpart`, `p`}, {`idxpart1`, `r`}, {`idxpart1_a_idx`, `i`}, {`idxpart2`, `r`}, {`idxpart2_a_idx`, `i`}, {`idxpart3`, `r`}, {`idxpart3_a_idx`, `i`}},
			},
			{
				Statement: `drop table idxpart, idxpart1, idxpart2, idxpart3;`,
			},
			{
				Statement: `select relname, relkind from pg_class where relname like 'idxpart%' order by relname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `create table idxpart (a int, b int, c int) partition by range(a);`,
			},
			{
				Statement: `create index on idxpart(c);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0) to (250);`,
			},
			{
				Statement: `create table idxpart2 partition of idxpart for values from (250) to (500);`,
			},
			{
				Statement: `alter table idxpart detach partition idxpart2;`,
			},
			{
				Statement: `\d idxpart2
              Table "public.idxpart2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | integer |           |          | 
Indexes:
    "idxpart2_c_idx" btree (c)
alter table idxpart2 drop column c;`,
			},
			{
				Statement: `\d idxpart2
              Table "public.idxpart2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
drop table idxpart, idxpart2;`,
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (like idxpart);`,
			},
			{
				Statement: `create index on idxpart1 ((a + b));`,
			},
			{
				Statement: `create index on idxpart ((a + b));`,
			},
			{
				Statement: `create table idxpart2 (like idxpart);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0000) to (1000);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from (1000) to (2000);`,
			},
			{
				Statement: `create table idxpart3 partition of idxpart for values from (2000) to (3000);`,
			},
			{
				Statement: `select relname as child, inhparent::regclass as parent, pg_get_indexdef as childdef
  from pg_class join pg_inherits on inhrelid = oid,
  lateral pg_get_indexdef(pg_class.oid)
  where relkind in ('i', 'I') and relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart1_expr_idx`, `idxpart_expr_idx`, `CREATE INDEX idxpart1_expr_idx ON public.idxpart1 USING btree (((a + b)))`}, {`idxpart2_expr_idx`, `idxpart_expr_idx`, `CREATE INDEX idxpart2_expr_idx ON public.idxpart2 USING btree (((a + b)))`}, {`idxpart3_expr_idx`, `idxpart_expr_idx`, `CREATE INDEX idxpart3_expr_idx ON public.idxpart3 USING btree (((a + b)))`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a text) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (like idxpart);`,
			},
			{
				Statement: `create table idxpart2 (like idxpart);`,
			},
			{
				Statement: `create index on idxpart2 (a collate "POSIX");`,
			},
			{
				Statement: `create index on idxpart2 (a);`,
			},
			{
				Statement: `create index on idxpart2 (a collate "C");`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from ('aaa') to ('bbb');`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from ('bbb') to ('ccc');`,
			},
			{
				Statement: `create table idxpart3 partition of idxpart for values from ('ccc') to ('ddd');`,
			},
			{
				Statement: `create index on idxpart (a collate "C");`,
			},
			{
				Statement: `create table idxpart4 partition of idxpart for values from ('ddd') to ('eee');`,
			},
			{
				Statement: `select relname as child, inhparent::regclass as parent, pg_get_indexdef as childdef
  from pg_class left join pg_inherits on inhrelid = oid,
  lateral pg_get_indexdef(pg_class.oid)
  where relkind in ('i', 'I') and relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart1_a_idx`, `idxpart_a_idx`, `CREATE INDEX idxpart1_a_idx ON public.idxpart1 USING btree (a COLLATE "C")`}, {`idxpart2_a_idx`, ``, `CREATE INDEX idxpart2_a_idx ON public.idxpart2 USING btree (a COLLATE "POSIX")`}, {`idxpart2_a_idx1`, ``, `CREATE INDEX idxpart2_a_idx1 ON public.idxpart2 USING btree (a)`}, {`idxpart2_a_idx2`, `idxpart_a_idx`, `CREATE INDEX idxpart2_a_idx2 ON public.idxpart2 USING btree (a COLLATE "C")`}, {`idxpart3_a_idx`, `idxpart_a_idx`, `CREATE INDEX idxpart3_a_idx ON public.idxpart3 USING btree (a COLLATE "C")`}, {`idxpart4_a_idx`, `idxpart_a_idx`, `CREATE INDEX idxpart4_a_idx ON public.idxpart4 USING btree (a COLLATE "C")`}, {`idxpart_a_idx`, ``, `CREATE INDEX idxpart_a_idx ON ONLY public.idxpart USING btree (a COLLATE "C")`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a text) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (like idxpart);`,
			},
			{
				Statement: `create table idxpart2 (like idxpart);`,
			},
			{
				Statement: `create index on idxpart2 (a);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from ('aaa') to ('bbb');`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from ('bbb') to ('ccc');`,
			},
			{
				Statement: `create table idxpart3 partition of idxpart for values from ('ccc') to ('ddd');`,
			},
			{
				Statement: `create index on idxpart (a text_pattern_ops);`,
			},
			{
				Statement: `create table idxpart4 partition of idxpart for values from ('ddd') to ('eee');`,
			},
			{
				Statement: `select relname as child, inhparent::regclass as parent, pg_get_indexdef as childdef
  from pg_class left join pg_inherits on inhrelid = oid,
  lateral pg_get_indexdef(pg_class.oid)
  where relkind in ('i', 'I') and relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart1_a_idx`, `idxpart_a_idx`, `CREATE INDEX idxpart1_a_idx ON public.idxpart1 USING btree (a text_pattern_ops)`}, {`idxpart2_a_idx`, ``, `CREATE INDEX idxpart2_a_idx ON public.idxpart2 USING btree (a)`}, {`idxpart2_a_idx1`, `idxpart_a_idx`, `CREATE INDEX idxpart2_a_idx1 ON public.idxpart2 USING btree (a text_pattern_ops)`}, {`idxpart3_a_idx`, `idxpart_a_idx`, `CREATE INDEX idxpart3_a_idx ON public.idxpart3 USING btree (a text_pattern_ops)`}, {`idxpart4_a_idx`, `idxpart_a_idx`, `CREATE INDEX idxpart4_a_idx ON public.idxpart4 USING btree (a text_pattern_ops)`}, {`idxpart_a_idx`, ``, `CREATE INDEX idxpart_a_idx ON ONLY public.idxpart USING btree (a text_pattern_ops)`}},
			},
			{
				Statement: `drop index idxpart_a_idx;`,
			},
			{
				Statement: `create index on only idxpart (a text_pattern_ops);`,
			},
			{
				Statement:   `alter index idxpart_a_idx attach partition idxpart2_a_idx;`,
				ErrorString: `cannot attach index "idxpart2_a_idx" as a partition of index "idxpart_a_idx"`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (col1 int, a int, col2 int, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (b int, col1 int, col2 int, col3 int, a int);`,
			},
			{
				Statement: `alter table idxpart drop column col1, drop column col2;`,
			},
			{
				Statement: `alter table idxpart1 drop column col1, drop column col2, drop column col3;`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0) to (1000);`,
			},
			{
				Statement: `create index idxpart_1_idx on only idxpart (b, a);`,
			},
			{
				Statement: `create index idxpart1_1_idx on idxpart1 (b, a);`,
			},
			{
				Statement: `create index idxpart1_1b_idx on idxpart1 (b);`,
			},
			{
				Statement: `create index idxpart_2_idx on only idxpart ((b + a)) where a > 1;`,
			},
			{
				Statement: `create index idxpart1_2_idx on idxpart1 ((b + a)) where a > 1;`,
			},
			{
				Statement: `create index idxpart1_2b_idx on idxpart1 ((a + b)) where a > 1;`,
			},
			{
				Statement: `create index idxpart1_2c_idx on idxpart1 ((b + a)) where b > 1;`,
			},
			{
				Statement:   `alter index idxpart_1_idx attach partition idxpart1_1b_idx;	-- fail`,
				ErrorString: `cannot attach index "idxpart1_1b_idx" as a partition of index "idxpart_1_idx"`,
			},
			{
				Statement: `alter index idxpart_1_idx attach partition idxpart1_1_idx;`,
			},
			{
				Statement:   `alter index idxpart_2_idx attach partition idxpart1_2b_idx;	-- fail`,
				ErrorString: `cannot attach index "idxpart1_2b_idx" as a partition of index "idxpart_2_idx"`,
			},
			{
				Statement:   `alter index idxpart_2_idx attach partition idxpart1_2c_idx;	-- fail`,
				ErrorString: `cannot attach index "idxpart1_2c_idx" as a partition of index "idxpart_2_idx"`,
			},
			{
				Statement: `alter index idxpart_2_idx attach partition idxpart1_2_idx;	-- ok`,
			},
			{
				Statement: `select relname as child, inhparent::regclass as parent, pg_get_indexdef as childdef
  from pg_class left join pg_inherits on inhrelid = oid,
  lateral pg_get_indexdef(pg_class.oid)
  where relkind in ('i', 'I') and relname like 'idxpart%' order by relname;`,
				Results: []sql.Row{{`idxpart1_1_idx`, `idxpart_1_idx`, `CREATE INDEX idxpart1_1_idx ON public.idxpart1 USING btree (b, a)`}, {`idxpart1_1b_idx`, ``, `CREATE INDEX idxpart1_1b_idx ON public.idxpart1 USING btree (b)`}, {`idxpart1_2_idx`, `idxpart_2_idx`, `CREATE INDEX idxpart1_2_idx ON public.idxpart1 USING btree (((b + a))) WHERE (a > 1)`}, {`idxpart1_2b_idx`, ``, `CREATE INDEX idxpart1_2b_idx ON public.idxpart1 USING btree (((a + b))) WHERE (a > 1)`}, {`idxpart1_2c_idx`, ``, `CREATE INDEX idxpart1_2c_idx ON public.idxpart1 USING btree (((b + a))) WHERE (b > 1)`}, {`idxpart_1_idx`, ``, `CREATE INDEX idxpart_1_idx ON ONLY public.idxpart USING btree (b, a)`}, {`idxpart_2_idx`, ``, `CREATE INDEX idxpart_2_idx ON ONLY public.idxpart USING btree (((b + a))) WHERE (a > 1)`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int, c text) partition by range (a);`,
			},
			{
				Statement: `create index idxparti on idxpart (a);`,
			},
			{
				Statement: `create index idxparti2 on idxpart (c, b);`,
			},
			{
				Statement: `create table idxpart1 (c text, a int, b int);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0) to (10);`,
			},
			{
				Statement: `create table idxpart2 (c text, a int, b int);`,
			},
			{
				Statement: `create index on idxpart2 (a);`,
			},
			{
				Statement: `create index on idxpart2 (c, b);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from (10) to (20);`,
			},
			{
				Statement: `select c.relname, pg_get_indexdef(indexrelid)
  from pg_class c join pg_index i on c.oid = i.indexrelid
  where indrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart1_a_idx`, `CREATE INDEX idxpart1_a_idx ON public.idxpart1 USING btree (a)`}, {`idxpart1_c_b_idx`, `CREATE INDEX idxpart1_c_b_idx ON public.idxpart1 USING btree (c, b)`}, {`idxpart2_a_idx`, `CREATE INDEX idxpart2_a_idx ON public.idxpart2 USING btree (a)`}, {`idxpart2_c_b_idx`, `CREATE INDEX idxpart2_c_b_idx ON public.idxpart2 USING btree (c, b)`}, {`idxparti`, `CREATE INDEX idxparti ON ONLY public.idxpart USING btree (a)`}, {`idxparti2`, `CREATE INDEX idxparti2 ON ONLY public.idxpart USING btree (c, b)`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (col1 int, col2 int, a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (col2 int, b int, col1 int, a int);`,
			},
			{
				Statement: `create table idxpart2 (col1 int, col2 int, b int, a int);`,
			},
			{
				Statement: `alter table idxpart drop column col1, drop column col2;`,
			},
			{
				Statement: `alter table idxpart1 drop column col1, drop column col2;`,
			},
			{
				Statement: `alter table idxpart2 drop column col1, drop column col2;`,
			},
			{
				Statement: `create index on idxpart2 (abs(b));`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from (0) to (1);`,
			},
			{
				Statement: `create index on idxpart (abs(b));`,
			},
			{
				Statement: `create index on idxpart ((b + 1));`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (1) to (2);`,
			},
			{
				Statement: `select c.relname, pg_get_indexdef(indexrelid)
  from pg_class c join pg_index i on c.oid = i.indexrelid
  where indrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart1_abs_idx`, `CREATE INDEX idxpart1_abs_idx ON public.idxpart1 USING btree (abs(b))`}, {`idxpart1_expr_idx`, `CREATE INDEX idxpart1_expr_idx ON public.idxpart1 USING btree (((b + 1)))`}, {`idxpart2_abs_idx`, `CREATE INDEX idxpart2_abs_idx ON public.idxpart2 USING btree (abs(b))`}, {`idxpart2_expr_idx`, `CREATE INDEX idxpart2_expr_idx ON public.idxpart2 USING btree (((b + 1)))`}, {`idxpart_abs_idx`, `CREATE INDEX idxpart_abs_idx ON ONLY public.idxpart USING btree (abs(b))`}, {`idxpart_expr_idx`, `CREATE INDEX idxpart_expr_idx ON ONLY public.idxpart USING btree (((b + 1)))`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (col1 int, a int, col3 int, b int) partition by range (a);`,
			},
			{
				Statement: `alter table idxpart drop column col1, drop column col3;`,
			},
			{
				Statement: `create table idxpart1 (col1 int, col2 int, col3 int, col4 int, b int, a int);`,
			},
			{
				Statement: `alter table idxpart1 drop column col1, drop column col2, drop column col3, drop column col4;`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0) to (1000);`,
			},
			{
				Statement: `create table idxpart2 (col1 int, col2 int, b int, a int);`,
			},
			{
				Statement: `create index on idxpart2 (a) where b > 1000;`,
			},
			{
				Statement: `alter table idxpart2 drop column col1, drop column col2;`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from (1000) to (2000);`,
			},
			{
				Statement: `create index on idxpart (a) where b > 1000;`,
			},
			{
				Statement: `select c.relname, pg_get_indexdef(indexrelid)
  from pg_class c join pg_index i on c.oid = i.indexrelid
  where indrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart1_a_idx`, `CREATE INDEX idxpart1_a_idx ON public.idxpart1 USING btree (a) WHERE (b > 1000)`}, {`idxpart2_a_idx`, `CREATE INDEX idxpart2_a_idx ON public.idxpart2 USING btree (a) WHERE (b > 1000)`}, {`idxpart_a_idx`, `CREATE INDEX idxpart_a_idx ON ONLY public.idxpart USING btree (a) WHERE (b > 1000)`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart1 (drop_1 int, drop_2 int, col_keep int, drop_3 int);`,
			},
			{
				Statement: `alter table idxpart1 drop column drop_1;`,
			},
			{
				Statement: `alter table idxpart1 drop column drop_2;`,
			},
			{
				Statement: `alter table idxpart1 drop column drop_3;`,
			},
			{
				Statement: `create index on idxpart1 (col_keep);`,
			},
			{
				Statement: `create table idxpart (col_keep int) partition by range (col_keep);`,
			},
			{
				Statement: `create index on idxpart (col_keep);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0) to (1000);`,
			},
			{
				Statement: `\d idxpart
         Partitioned table "public.idxpart"
  Column  |  Type   | Collation | Nullable | Default 
----------+---------+-----------+----------+---------
 col_keep | integer |           |          | 
Partition key: RANGE (col_keep)
Indexes:
    "idxpart_col_keep_idx" btree (col_keep)
Number of partitions: 1 (Use \d+ to list them.)
\d idxpart1
               Table "public.idxpart1"
  Column  |  Type   | Collation | Nullable | Default 
----------+---------+-----------+----------+---------
 col_keep | integer |           |          | 
Partition of: idxpart FOR VALUES FROM (0) TO (1000)
Indexes:
    "idxpart1_col_keep_idx" btree (col_keep)
select attrelid::regclass, attname, attnum from pg_attribute
  where attrelid::regclass::text like 'idxpart%' and attnum > 0
  order by attrelid::regclass, attnum;`,
				Results: []sql.Row{{`idxpart1`, `........pg.dropped.1........`, 1}, {`idxpart1`, `........pg.dropped.2........`, 2}, {`idxpart1`, `col_keep`, 3}, {`idxpart1`, `........pg.dropped.4........`, 4}, {`idxpart1_col_keep_idx`, `col_keep`, 1}, {`idxpart`, `col_keep`, 1}, {`idxpart_col_keep_idx`, `col_keep`, 1}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart(drop_1 int, drop_2 int, col_keep int, drop_3 int) partition by range (col_keep);`,
			},
			{
				Statement: `alter table idxpart drop column drop_1;`,
			},
			{
				Statement: `alter table idxpart drop column drop_2;`,
			},
			{
				Statement: `alter table idxpart drop column drop_3;`,
			},
			{
				Statement: `create table idxpart1 (col_keep int);`,
			},
			{
				Statement: `create index on idxpart1 (col_keep);`,
			},
			{
				Statement: `create index on idxpart (col_keep);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (0) to (1000);`,
			},
			{
				Statement: `\d idxpart
         Partitioned table "public.idxpart"
  Column  |  Type   | Collation | Nullable | Default 
----------+---------+-----------+----------+---------
 col_keep | integer |           |          | 
Partition key: RANGE (col_keep)
Indexes:
    "idxpart_col_keep_idx" btree (col_keep)
Number of partitions: 1 (Use \d+ to list them.)
\d idxpart1
               Table "public.idxpart1"
  Column  |  Type   | Collation | Nullable | Default 
----------+---------+-----------+----------+---------
 col_keep | integer |           |          | 
Partition of: idxpart FOR VALUES FROM (0) TO (1000)
Indexes:
    "idxpart1_col_keep_idx" btree (col_keep)
select attrelid::regclass, attname, attnum from pg_attribute
  where attrelid::regclass::text like 'idxpart%' and attnum > 0
  order by attrelid::regclass, attnum;`,
				Results: []sql.Row{{`idxpart`, `........pg.dropped.1........`, 1}, {`idxpart`, `........pg.dropped.2........`, 2}, {`idxpart`, `col_keep`, 3}, {`idxpart`, `........pg.dropped.4........`, 4}, {`idxpart1`, `col_keep`, 1}, {`idxpart1_col_keep_idx`, `col_keep`, 1}, {`idxpart_col_keep_idx`, `col_keep`, 1}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int primary key, b int) partition by range (a);`,
			},
			{
				Statement: `\d idxpart
        Partitioned table "public.idxpart"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
 b      | integer |           |          | 
Partition key: RANGE (a)
Indexes:
    "idxpart_pkey" PRIMARY KEY, btree (a)
Number of partitions: 0
create table failpart partition of idxpart (b primary key) for values from (0) to (100);`,
				ErrorString: `multiple primary keys for table "failpart" are not allowed`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1pk partition of idxpart (a primary key) for values from (0) to (100);`,
			},
			{
				Statement: `\d idxpart1pk
             Table "public.idxpart1pk"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
Partition of: idxpart FOR VALUES FROM (0) TO (100)
Indexes:
    "idxpart1pk_pkey" PRIMARY KEY, btree (a)
drop table idxpart;`,
			},
			{
				Statement:   `create table idxpart (a int unique, b int) partition by range (a, b);`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement:   `create table idxpart (a int, b int unique) partition by range (a, b);`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement:   `create table idxpart (a int primary key, b int) partition by range (b, a);`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement:   `create table idxpart (a int, b int primary key) partition by range (b, a);`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement: `create table idxpart (a int, b int, c text, primary key  (a, b, c)) partition by range (b, c, a);`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement:   `create table idxpart (a int, exclude (a with = )) partition by range (a);`,
				ErrorString: `exclusion constraints are not supported on partitioned tables`,
			},
			{
				Statement:   `create table idxpart (a int primary key, b int) partition by range ((b + a));`,
				ErrorString: `unsupported PRIMARY KEY constraint with partition key definition`,
			},
			{
				Statement:   `create table idxpart (a int unique, b int) partition by range ((b + a));`,
				ErrorString: `unsupported UNIQUE constraint with partition key definition`,
			},
			{
				Statement: `create table idxpart (a int, b int, c text) partition by range (a, b);`,
			},
			{
				Statement:   `alter table idxpart add primary key (a);	-- not an incomplete one though`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement: `alter table idxpart add primary key (a, b);	-- this works`,
			},
			{
				Statement: `\d idxpart
        Partitioned table "public.idxpart"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
 b      | integer |           | not null | 
 c      | text    |           |          | 
Partition key: RANGE (a, b)
Indexes:
    "idxpart_pkey" PRIMARY KEY, btree (a, b)
Number of partitions: 0
create table idxpart1 partition of idxpart for values from (0, 0) to (1000, 1000);`,
			},
			{
				Statement: `\d idxpart1
              Table "public.idxpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
 b      | integer |           | not null | 
 c      | text    |           |          | 
Partition of: idxpart FOR VALUES FROM (0, 0) TO (1000, 1000)
Indexes:
    "idxpart1_pkey" PRIMARY KEY, btree (a, b)
drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a, b);`,
			},
			{
				Statement:   `alter table idxpart add unique (a);			-- not an incomplete one though`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement: `alter table idxpart add unique (b, a);		-- this works`,
			},
			{
				Statement: `\d idxpart
        Partitioned table "public.idxpart"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition key: RANGE (a, b)
Indexes:
    "idxpart_b_a_key" UNIQUE CONSTRAINT, btree (b, a)
Number of partitions: 0
drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a);`,
			},
			{
				Statement:   `alter table idxpart add exclude (a with =);`,
				ErrorString: `exclusion constraints are not supported on partitioned tables`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int, primary key (a, b)) partition by range (a, b);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (1, 1) to (10, 10);`,
			},
			{
				Statement: `create table idxpart2 partition of idxpart for values from (10, 10) to (20, 20)
  partition by range (b);`,
			},
			{
				Statement: `create table idxpart21 partition of idxpart2 for values from (10) to (15);`,
			},
			{
				Statement: `create table idxpart22 partition of idxpart2 for values from (15) to (20);`,
			},
			{
				Statement: `create table idxpart3 (b int not null, a int not null);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart3 for values from (20, 20) to (30, 30);`,
			},
			{
				Statement: `select conname, contype, conrelid::regclass, conindid::regclass, conkey
  from pg_constraint where conrelid::regclass::text like 'idxpart%'
  order by conname;`,
				Results: []sql.Row{{`idxpart1_pkey`, `p`, `idxpart1`, `idxpart1_pkey`, `{1,2}`}, {`idxpart21_pkey`, `p`, `idxpart21`, `idxpart21_pkey`, `{1,2}`}, {`idxpart22_pkey`, `p`, `idxpart22`, `idxpart22_pkey`, `{1,2}`}, {`idxpart2_pkey`, `p`, `idxpart2`, `idxpart2_pkey`, `{1,2}`}, {`idxpart3_pkey`, `p`, `idxpart3`, `idxpart3_pkey`, `{2,1}`}, {`idxpart_pkey`, `p`, `idxpart`, `idxpart_pkey`, `{1,2}`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int, primary key (a)) partition by range (a);`,
			},
			{
				Statement: `create table idxpart2 partition of idxpart
for values from (0) to (1000) partition by range (b); -- fail`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int unique, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (a int not null, b int, unique (a, b))
  partition by range (a, b);`,
			},
			{
				Statement:   `alter table idxpart attach partition idxpart1 for values from (1) to (1000);`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement: `DROP TABLE idxpart, idxpart1;`,
			},
			{
				Statement: `create table idxpart (a int, b int, primary key (a, b)) partition by range (a);`,
			},
			{
				Statement: `create table idxpart2 partition of idxpart for values from (0) to (1000) partition by range (b);`,
			},
			{
				Statement: `create table idxpart21 partition of idxpart2 for values from (0) to (1000);`,
			},
			{
				Statement: `select conname, contype, conrelid::regclass, conindid::regclass, conkey
  from pg_constraint where conrelid::regclass::text like 'idxpart%'
  order by conname;`,
				Results: []sql.Row{{`idxpart21_pkey`, `p`, `idxpart21`, `idxpart21_pkey`, `{1,2}`}, {`idxpart2_pkey`, `p`, `idxpart2`, `idxpart2_pkey`, `{1,2}`}, {`idxpart_pkey`, `p`, `idxpart`, `idxpart_pkey`, `{1,2}`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (i int) partition by hash (i);`,
			},
			{
				Statement: `create table idxpart0 partition of idxpart (i) for values with (modulus 2, remainder 0);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart (i) for values with (modulus 2, remainder 1);`,
			},
			{
				Statement: `alter table idxpart0 add primary key(i);`,
			},
			{
				Statement: `alter table idxpart add primary key(i);`,
			},
			{
				Statement: `select indrelid::regclass, indexrelid::regclass, inhparent::regclass, indisvalid,
  conname, conislocal, coninhcount, connoinherit, convalidated
  from pg_index idx left join pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  left join pg_constraint con on (idx.indexrelid = con.conindid)
  where indrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart0`, `idxpart0_pkey`, `idxpart_pkey`, true, `idxpart0_pkey`, false, 1, true, true}, {`idxpart1`, `idxpart1_pkey`, `idxpart_pkey`, true, `idxpart1_pkey`, false, 1, false, true}, {`idxpart`, `idxpart_pkey`, ``, true, `idxpart_pkey`, true, 0, true, true}},
			},
			{
				Statement:   `drop index idxpart0_pkey;								-- fail`,
				ErrorString: `cannot drop index idxpart0_pkey because index idxpart_pkey requires it`,
			},
			{
				Statement:   `drop index idxpart1_pkey;								-- fail`,
				ErrorString: `cannot drop index idxpart1_pkey because index idxpart_pkey requires it`,
			},
			{
				Statement:   `alter table idxpart0 drop constraint idxpart0_pkey;		-- fail`,
				ErrorString: `cannot drop inherited constraint "idxpart0_pkey" of relation "idxpart0"`,
			},
			{
				Statement:   `alter table idxpart1 drop constraint idxpart1_pkey;		-- fail`,
				ErrorString: `cannot drop inherited constraint "idxpart1_pkey" of relation "idxpart1"`,
			},
			{
				Statement: `alter table idxpart drop constraint idxpart_pkey;		-- ok`,
			},
			{
				Statement: `select indrelid::regclass, indexrelid::regclass, inhparent::regclass, indisvalid,
  conname, conislocal, coninhcount, connoinherit, convalidated
  from pg_index idx left join pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  left join pg_constraint con on (idx.indexrelid = con.conindid)
  where indrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `CREATE TABLE idxpart (c1 INT PRIMARY KEY, c2 INT, c3 VARCHAR(10)) PARTITION BY RANGE(c1);`,
			},
			{
				Statement: `CREATE TABLE idxpart1 (LIKE idxpart);`,
			},
			{
				Statement: `ALTER TABLE idxpart1 ADD PRIMARY KEY (c1, c2);`,
			},
			{
				Statement:   `ALTER TABLE idxpart ATTACH PARTITION idxpart1 FOR VALUES FROM (100) TO (200);`,
				ErrorString: `multiple primary keys for table "idxpart1" are not allowed`,
			},
			{
				Statement: `DROP TABLE idxpart, idxpart1;`,
			},
			{
				Statement: `create table idxpart (a int, b int, primary key (a)) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (a int not null, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart11 (a int not null, b int primary key);`,
			},
			{
				Statement: `alter table idxpart1 attach partition idxpart11 for values from (0) to (1000);`,
			},
			{
				Statement:   `alter table idxpart attach partition idxpart1 for values from (0) to (10000);`,
				ErrorString: `multiple primary keys for table "idxpart11" are not allowed`,
			},
			{
				Statement: `drop table idxpart, idxpart1, idxpart11;`,
			},
			{
				Statement: `create table idxpart (a int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart0 (like idxpart);`,
			},
			{
				Statement: `alter table idxpart0 add primary key (a);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart0 for values from (0) to (1000);`,
			},
			{
				Statement: `alter table only idxpart add primary key (a);`,
			},
			{
				Statement: `select indrelid::regclass, indexrelid::regclass, inhparent::regclass, indisvalid,
  conname, conislocal, coninhcount, connoinherit, convalidated
  from pg_index idx left join pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  left join pg_constraint con on (idx.indexrelid = con.conindid)
  where indrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart0`, `idxpart0_pkey`, ``, true, `idxpart0_pkey`, true, 0, true, true}, {`idxpart`, `idxpart_pkey`, ``, false, `idxpart_pkey`, true, 0, true, true}},
			},
			{
				Statement: `alter index idxpart_pkey attach partition idxpart0_pkey;`,
			},
			{
				Statement: `select indrelid::regclass, indexrelid::regclass, inhparent::regclass, indisvalid,
  conname, conislocal, coninhcount, connoinherit, convalidated
  from pg_index idx left join pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  left join pg_constraint con on (idx.indexrelid = con.conindid)
  where indrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart0`, `idxpart0_pkey`, `idxpart_pkey`, true, `idxpart0_pkey`, false, 1, true, true}, {`idxpart`, `idxpart_pkey`, ``, true, `idxpart_pkey`, true, 0, true, true}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart0 (like idxpart);`,
			},
			{
				Statement: `alter table idxpart0 add unique (a);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart0 default;`,
			},
			{
				Statement:   `alter table only idxpart add primary key (a);  -- fail, no NOT NULL constraint`,
				ErrorString: `constraint must be added to child tables too`,
			},
			{
				Statement: `alter table idxpart0 alter column a set not null;`,
			},
			{
				Statement: `alter table only idxpart add primary key (a);  -- now it works`,
			},
			{
				Statement:   `alter table idxpart0 alter column a drop not null;  -- fail, pkey needs it`,
				ErrorString: `column "a" is marked NOT NULL in parent table`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (a int not null, b int);`,
			},
			{
				Statement: `create unique index on idxpart1 (a);`,
			},
			{
				Statement: `alter table idxpart add primary key (a);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (1) to (1000);`,
			},
			{
				Statement: `select indrelid::regclass, indexrelid::regclass, inhparent::regclass, indisvalid,
  conname, conislocal, coninhcount, connoinherit, convalidated
  from pg_index idx left join pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  left join pg_constraint con on (idx.indexrelid = con.conindid)
  where indrelid::regclass::text like 'idxpart%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`idxpart1`, `idxpart1_a_idx`, ``, true, ``, ``, ``, ``, ``}, {`idxpart1`, `idxpart1_pkey`, `idxpart_pkey`, true, `idxpart1_pkey`, false, 1, false, true}, {`idxpart`, `idxpart_pkey`, ``, true, `idxpart_pkey`, true, 0, true, true}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 (a int not null, b int);`,
			},
			{
				Statement: `create unique index on idxpart1 (a);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart1 for values from (1) to (1000);`,
			},
			{
				Statement: `alter table only idxpart add primary key (a);`,
			},
			{
				Statement:   `alter index idxpart_pkey attach partition idxpart1_a_idx;	-- fail`,
				ErrorString: `cannot attach index "idxpart1_a_idx" as a partition of index "idxpart_pkey"`,
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int, b text, primary key (a, b)) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0) to (100000);`,
			},
			{
				Statement: `create table idxpart2 (c int, like idxpart);`,
			},
			{
				Statement: `insert into idxpart2 (c, a, b) values (42, 572814, 'inserted first');`,
			},
			{
				Statement: `alter table idxpart2 drop column c;`,
			},
			{
				Statement: `create unique index on idxpart (a);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart2 for values from (100000) to (1000000);`,
			},
			{
				Statement: `insert into idxpart values (0, 'zero'), (42, 'life'), (2^16, 'sixteen');`,
			},
			{
				Statement:   `insert into idxpart select 2^g, format('two to power of %s', g) from generate_series(15, 17) g;`,
				ErrorString: `duplicate key value violates unique constraint "idxpart1_a_idx"`,
			},
			{
				Statement: `insert into idxpart values (16, 'sixteen');`,
			},
			{
				Statement: `insert into idxpart (b, a) values ('one', 142857), ('two', 285714);`,
			},
			{
				Statement:   `insert into idxpart select a * 2, b || b from idxpart where a between 2^16 and 2^19;`,
				ErrorString: `duplicate key value violates unique constraint "idxpart2_a_idx"`,
			},
			{
				Statement:   `insert into idxpart values (572814, 'five');`,
				ErrorString: `duplicate key value violates unique constraint "idxpart2_a_idx"`,
			},
			{
				Statement: `insert into idxpart values (857142, 'six');`,
			},
			{
				Statement: `select tableoid::regclass, * from idxpart order by a;`,
				Results:   []sql.Row{{`idxpart1`, 0, `zero`}, {`idxpart1`, 16, `sixteen`}, {`idxpart1`, 42, `life`}, {`idxpart1`, 65536, `sixteen`}, {`idxpart2`, 142857, `one`}, {`idxpart2`, 285714, `two`}, {`idxpart2`, 572814, `inserted first`}, {`idxpart2`, 857142, `six`}},
			},
			{
				Statement: `drop table idxpart;`,
			},
			{
				Statement: `create table idxpart (a int) partition by range (a);`,
			},
			{
				Statement: `create table idxpart1 partition of idxpart for values from (0) to (100);`,
			},
			{
				Statement: `create table idxpart2 partition of idxpart for values from (100) to (1000)
  partition by range (a);`,
			},
			{
				Statement: `create table idxpart21 partition of idxpart2 for values from (100) to (200);`,
			},
			{
				Statement: `create table idxpart22 partition of idxpart2 for values from (200) to (300);`,
			},
			{
				Statement: `create index on idxpart22 (a);`,
			},
			{
				Statement: `create index on only idxpart2 (a);`,
			},
			{
				Statement: `alter index idxpart2_a_idx attach partition idxpart22_a_idx;`,
			},
			{
				Statement: `create index on idxpart (a);`,
			},
			{
				Statement: `create table idxpart_another (a int, b int, primary key (a, b)) partition by range (a);`,
			},
			{
				Statement: `create table idxpart_another_1 partition of idxpart_another for values from (0) to (100);`,
			},
			{
				Statement: `create table idxpart3 (c int, b int, a int) partition by range (a);`,
			},
			{
				Statement: `alter table idxpart3 drop column b, drop column c;`,
			},
			{
				Statement: `create table idxpart31 partition of idxpart3 for values from (1000) to (1200);`,
			},
			{
				Statement: `create table idxpart32 partition of idxpart3 for values from (1200) to (1400);`,
			},
			{
				Statement: `alter table idxpart attach partition idxpart3 for values from (1000) to (2000);`,
			},
			{
				Statement: `create schema regress_indexing;`,
			},
			{
				Statement: `set search_path to regress_indexing;`,
			},
			{
				Statement: `create table pk (a int primary key) partition by range (a);`,
			},
			{
				Statement: `create table pk1 partition of pk for values from (0) to (1000);`,
			},
			{
				Statement: `create table pk2 (b int, a int);`,
			},
			{
				Statement: `alter table pk2 drop column b;`,
			},
			{
				Statement: `alter table pk2 alter a set not null;`,
			},
			{
				Statement: `alter table pk attach partition pk2 for values from (1000) to (2000);`,
			},
			{
				Statement: `create table pk3 partition of pk for values from (2000) to (3000);`,
			},
			{
				Statement: `create table pk4 (like pk);`,
			},
			{
				Statement: `alter table pk attach partition pk4 for values from (3000) to (4000);`,
			},
			{
				Statement: `create table pk5 (like pk) partition by range (a);`,
			},
			{
				Statement: `create table pk51 partition of pk5 for values from (4000) to (4500);`,
			},
			{
				Statement: `create table pk52 partition of pk5 for values from (4500) to (5000);`,
			},
			{
				Statement: `alter table pk attach partition pk5 for values from (4000) to (5000);`,
			},
			{
				Statement: `reset search_path;`,
			},
			{
				Statement: `create table covidxpart (a int, b int) partition by list (a);`,
			},
			{
				Statement: `create unique index on covidxpart (a) include (b);`,
			},
			{
				Statement: `create table covidxpart1 partition of covidxpart for values in (1);`,
			},
			{
				Statement: `create table covidxpart2 partition of covidxpart for values in (2);`,
			},
			{
				Statement: `insert into covidxpart values (1, 1);`,
			},
			{
				Statement:   `insert into covidxpart values (1, 1);`,
				ErrorString: `duplicate key value violates unique constraint "covidxpart1_a_b_idx"`,
			},
			{
				Statement: `create table covidxpart3 (b int, c int, a int);`,
			},
			{
				Statement: `alter table covidxpart3 drop c;`,
			},
			{
				Statement: `alter table covidxpart attach partition covidxpart3 for values in (3);`,
			},
			{
				Statement: `insert into covidxpart values (3, 1);`,
			},
			{
				Statement:   `insert into covidxpart values (3, 1);`,
				ErrorString: `duplicate key value violates unique constraint "covidxpart3_a_b_idx"`,
			},
			{
				Statement: `create table covidxpart4 (b int, a int);`,
			},
			{
				Statement: `create unique index on covidxpart4 (a) include (b);`,
			},
			{
				Statement: `create unique index on covidxpart4 (a);`,
			},
			{
				Statement: `alter table covidxpart attach partition covidxpart4 for values in (4);`,
			},
			{
				Statement: `insert into covidxpart values (4, 1);`,
			},
			{
				Statement:   `insert into covidxpart values (4, 1);`,
				ErrorString: `duplicate key value violates unique constraint "covidxpart4_a_b_idx"`,
			},
			{
				Statement:   `create unique index on covidxpart (b) include (a); -- should fail`,
				ErrorString: `unique constraint on partitioned table must include all partitioning columns`,
			},
			{
				Statement: `create table parted_pk_detach_test (a int primary key) partition by list (a);`,
			},
			{
				Statement: `create table parted_pk_detach_test1 partition of parted_pk_detach_test for values in (1);`,
			},
			{
				Statement:   `alter table parted_pk_detach_test1 drop constraint parted_pk_detach_test1_pkey;	-- should fail`,
				ErrorString: `cannot drop inherited constraint "parted_pk_detach_test1_pkey" of relation "parted_pk_detach_test1"`,
			},
			{
				Statement: `alter table parted_pk_detach_test detach partition parted_pk_detach_test1;`,
			},
			{
				Statement: `alter table parted_pk_detach_test1 drop constraint parted_pk_detach_test1_pkey;`,
			},
			{
				Statement: `drop table parted_pk_detach_test, parted_pk_detach_test1;`,
			},
			{
				Statement: `create table parted_uniq_detach_test (a int unique) partition by list (a);`,
			},
			{
				Statement: `create table parted_uniq_detach_test1 partition of parted_uniq_detach_test for values in (1);`,
			},
			{
				Statement:   `alter table parted_uniq_detach_test1 drop constraint parted_uniq_detach_test1_a_key;	-- should fail`,
				ErrorString: `cannot drop inherited constraint "parted_uniq_detach_test1_a_key" of relation "parted_uniq_detach_test1"`,
			},
			{
				Statement: `alter table parted_uniq_detach_test detach partition parted_uniq_detach_test1;`,
			},
			{
				Statement: `alter table parted_uniq_detach_test1 drop constraint parted_uniq_detach_test1_a_key;`,
			},
			{
				Statement: `drop table parted_uniq_detach_test, parted_uniq_detach_test1;`,
			},
			{
				Statement: `create table parted_index_col_drop(a int, b int, c int)
  partition by list (a);`,
			},
			{
				Statement: `create table parted_index_col_drop1 partition of parted_index_col_drop
  for values in (1) partition by list (a);`,
			},
			{
				Statement: `create table parted_index_col_drop2 partition of parted_index_col_drop
  for values in (2) partition by list (a);`,
			},
			{
				Statement: `create table parted_index_col_drop11 partition of parted_index_col_drop1
  for values in (1);`,
			},
			{
				Statement: `create index on parted_index_col_drop (b);`,
			},
			{
				Statement: `create index on parted_index_col_drop (c);`,
			},
			{
				Statement: `create index on parted_index_col_drop (b, c);`,
			},
			{
				Statement: `alter table parted_index_col_drop drop column c;`,
			},
			{
				Statement: `\d parted_index_col_drop
 Partitioned table "public.parted_index_col_drop"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition key: LIST (a)
Indexes:
    "parted_index_col_drop_b_idx" btree (b)
Number of partitions: 2 (Use \d+ to list them.)
\d parted_index_col_drop1
 Partitioned table "public.parted_index_col_drop1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: parted_index_col_drop FOR VALUES IN (1)
Partition key: LIST (a)
Indexes:
    "parted_index_col_drop1_b_idx" btree (b)
Number of partitions: 1 (Use \d+ to list them.)
\d parted_index_col_drop2
 Partitioned table "public.parted_index_col_drop2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: parted_index_col_drop FOR VALUES IN (2)
Partition key: LIST (a)
Indexes:
    "parted_index_col_drop2_b_idx" btree (b)
Number of partitions: 0
\d parted_index_col_drop11
      Table "public.parted_index_col_drop11"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: parted_index_col_drop1 FOR VALUES IN (1)
Indexes:
    "parted_index_col_drop11_b_idx" btree (b)
drop table parted_index_col_drop;`,
			},
			{
				Statement: `create table parted_inval_tab (a int) partition by range (a);`,
			},
			{
				Statement: `create index parted_inval_idx on parted_inval_tab (a);`,
			},
			{
				Statement: `create table parted_inval_tab_1 (a int) partition by range (a);`,
			},
			{
				Statement: `create table parted_inval_tab_1_1 partition of parted_inval_tab_1
  for values from (0) to (10);`,
			},
			{
				Statement: `create table parted_inval_tab_1_2 partition of parted_inval_tab_1
  for values from (10) to (20);`,
			},
			{
				Statement: `create index parted_inval_ixd_1 on only parted_inval_tab_1 (a);`,
			},
			{
				Statement: `alter table parted_inval_tab attach partition parted_inval_tab_1
  for values from (1) to (100);`,
			},
			{
				Statement: `select indexrelid::regclass, indisvalid,
       indrelid::regclass, inhparent::regclass
  from pg_index idx left join
       pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  where indexrelid::regclass::text like 'parted_inval%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`parted_inval_idx`, true, `parted_inval_tab`, ``}, {`parted_inval_ixd_1`, false, `parted_inval_tab_1`, ``}, {`parted_inval_tab_1_1_a_idx`, true, `parted_inval_tab_1_1`, `parted_inval_tab_1_a_idx`}, {`parted_inval_tab_1_2_a_idx`, true, `parted_inval_tab_1_2`, `parted_inval_tab_1_a_idx`}, {`parted_inval_tab_1_a_idx`, true, `parted_inval_tab_1`, `parted_inval_idx`}},
			},
			{
				Statement: `drop table parted_inval_tab;`,
			},
			{
				Statement: `create table parted_isvalid_tab (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table parted_isvalid_tab_1 partition of parted_isvalid_tab
  for values from (1) to (10) partition by range (a);`,
			},
			{
				Statement: `create table parted_isvalid_tab_2 partition of parted_isvalid_tab
  for values from (10) to (20) partition by range (a);`,
			},
			{
				Statement: `create table parted_isvalid_tab_11 partition of parted_isvalid_tab_1
  for values from (1) to (5);`,
			},
			{
				Statement: `create table parted_isvalid_tab_12 partition of parted_isvalid_tab_1
  for values from (5) to (10);`,
			},
			{
				Statement: `insert into parted_isvalid_tab_11 values (1, 0);`,
			},
			{
				Statement:   `create index concurrently parted_isvalid_idx_11 on parted_isvalid_tab_11 ((a/b));`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `create index parted_isvalid_idx on parted_isvalid_tab ((a/b));`,
			},
			{
				Statement: `select indexrelid::regclass, indisvalid,
       indrelid::regclass, inhparent::regclass
  from pg_index idx left join
       pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  where indexrelid::regclass::text like 'parted_isvalid%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`parted_isvalid_idx`, false, `parted_isvalid_tab`, ``}, {`parted_isvalid_idx_11`, false, `parted_isvalid_tab_11`, `parted_isvalid_tab_1_expr_idx`}, {`parted_isvalid_tab_12_expr_idx`, true, `parted_isvalid_tab_12`, `parted_isvalid_tab_1_expr_idx`}, {`parted_isvalid_tab_1_expr_idx`, false, `parted_isvalid_tab_1`, `parted_isvalid_idx`}, {`parted_isvalid_tab_2_expr_idx`, true, `parted_isvalid_tab_2`, `parted_isvalid_idx`}},
			},
			{
				Statement: `drop table parted_isvalid_tab;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create table parted_replica_tab (id int not null) partition by range (id);`,
			},
			{
				Statement: `create table parted_replica_tab_1 partition of parted_replica_tab
  for values from (1) to (10) partition by range (id);`,
			},
			{
				Statement: `create table parted_replica_tab_11 partition of parted_replica_tab_1
  for values from (1) to (5);`,
			},
			{
				Statement: `create unique index parted_replica_idx
  on only parted_replica_tab using btree (id);`,
			},
			{
				Statement: `create unique index parted_replica_idx_1
  on only parted_replica_tab_1 using btree (id);`,
			},
			{
				Statement: `alter table only parted_replica_tab_1 replica identity
  using index parted_replica_idx_1;`,
			},
			{
				Statement: `create unique index parted_replica_idx_11 on parted_replica_tab_11 USING btree (id);`,
			},
			{
				Statement: `select indexrelid::regclass, indisvalid, indisreplident,
       indrelid::regclass, inhparent::regclass
  from pg_index idx left join
       pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  where indexrelid::regclass::text like 'parted_replica%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`parted_replica_idx`, false, false, `parted_replica_tab`, ``}, {`parted_replica_idx_1`, false, true, `parted_replica_tab_1`, ``}, {`parted_replica_idx_11`, true, false, `parted_replica_tab_11`, ``}},
			},
			{
				Statement: `alter index parted_replica_idx ATTACH PARTITION parted_replica_idx_1;`,
			},
			{
				Statement: `select indexrelid::regclass, indisvalid, indisreplident,
       indrelid::regclass, inhparent::regclass
  from pg_index idx left join
       pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  where indexrelid::regclass::text like 'parted_replica%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`parted_replica_idx`, false, false, `parted_replica_tab`, ``}, {`parted_replica_idx_1`, false, true, `parted_replica_tab_1`, `parted_replica_idx`}, {`parted_replica_idx_11`, true, false, `parted_replica_tab_11`, ``}},
			},
			{
				Statement: `alter index parted_replica_idx_1 ATTACH PARTITION parted_replica_idx_11;`,
			},
			{
				Statement: `alter table only parted_replica_tab_1 replica identity
  using index parted_replica_idx_1;`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `select indexrelid::regclass, indisvalid, indisreplident,
       indrelid::regclass, inhparent::regclass
  from pg_index idx left join
       pg_inherits inh on (idx.indexrelid = inh.inhrelid)
  where indexrelid::regclass::text like 'parted_replica%'
  order by indexrelid::regclass::text collate "C";`,
				Results: []sql.Row{{`parted_replica_idx`, true, false, `parted_replica_tab`, ``}, {`parted_replica_idx_1`, true, true, `parted_replica_tab_1`, `parted_replica_idx`}, {`parted_replica_idx_11`, true, false, `parted_replica_tab_11`, `parted_replica_idx_1`}},
			},
			{
				Statement: `drop table parted_replica_tab;`,
			},
		},
	})
}
