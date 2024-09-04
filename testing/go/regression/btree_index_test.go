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

func TestBtreeIndex(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_btree_index)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_btree_index,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv abs_srcdir PG_ABS_SRCDIR
CREATE TABLE bt_i4_heap (
	seqno 		int4,
	random 		int4
);`,
			},
			{
				Statement: `CREATE TABLE bt_name_heap (
	seqno 		name,
	random 		int4
);`,
			},
			{
				Statement: `CREATE TABLE bt_txt_heap (
	seqno 		text,
	random 		int4
);`,
			},
			{
				Statement: `CREATE TABLE bt_f8_heap (
	seqno 		float8,
	random 		int4
);`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/desc.data'
COPY bt_i4_heap FROM :'filename';`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/hash.data'
COPY bt_name_heap FROM :'filename';`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/desc.data'
COPY bt_txt_heap FROM :'filename';`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/hash.data'
COPY bt_f8_heap FROM :'filename';`,
			},
			{
				Statement: `ANALYZE bt_i4_heap;`,
			},
			{
				Statement: `ANALYZE bt_name_heap;`,
			},
			{
				Statement: `ANALYZE bt_txt_heap;`,
			},
			{
				Statement: `ANALYZE bt_f8_heap;`,
			},
			{
				Statement: `CREATE INDEX bt_i4_index ON bt_i4_heap USING btree (seqno int4_ops);`,
			},
			{
				Statement: `CREATE INDEX bt_name_index ON bt_name_heap USING btree (seqno name_ops);`,
			},
			{
				Statement: `CREATE INDEX bt_txt_index ON bt_txt_heap USING btree (seqno text_ops);`,
			},
			{
				Statement: `CREATE INDEX bt_f8_index ON bt_f8_heap USING btree (seqno float8_ops);`,
			},
			{
				Statement: `SELECT b.*
   FROM bt_i4_heap b
   WHERE b.seqno < 1;`,
				Results: []sql.Row{{0, 1935401906}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_i4_heap b
   WHERE b.seqno >= 9999;`,
				Results: []sql.Row{{9999, 1227676208}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_i4_heap b
   WHERE b.seqno = 4500;`,
				Results: []sql.Row{{4500, 2080851358}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_name_heap b
   WHERE b.seqno < '1'::name;`,
				Results: []sql.Row{{0, 1935401906}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_name_heap b
   WHERE b.seqno >= '9999'::name;`,
				Results: []sql.Row{{9999, 1227676208}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_name_heap b
   WHERE b.seqno = '4500'::name;`,
				Results: []sql.Row{{4500, 2080851358}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_txt_heap b
   WHERE b.seqno < '1'::text;`,
				Results: []sql.Row{{0, 1935401906}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_txt_heap b
   WHERE b.seqno >= '9999'::text;`,
				Results: []sql.Row{{9999, 1227676208}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_txt_heap b
   WHERE b.seqno = '4500'::text;`,
				Results: []sql.Row{{4500, 2080851358}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_f8_heap b
   WHERE b.seqno < '1'::float8;`,
				Results: []sql.Row{{0, 1935401906}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_f8_heap b
   WHERE b.seqno >= '9999'::float8;`,
				Results: []sql.Row{{9999, 1227676208}},
			},
			{
				Statement: `SELECT b.*
   FROM bt_f8_heap b
   WHERE b.seqno = '4500'::float8;`,
				Results: []sql.Row{{4500, 2080851358}},
			},
			{
				Statement: `set enable_seqscan to false;`,
			},
			{
				Statement: `set enable_indexscan to true;`,
			},
			{
				Statement: `set enable_bitmapscan to false;`,
			},
			{
				Statement: `explain (costs off)
select proname from pg_proc where proname like E'RI\\_FKey%del' order by 1;`,
				Results: []sql.Row{{`Index Only Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: ((proname >= 'RI_FKey'::text) AND (proname < 'RI_FKez'::text))`}, {`Filter: (proname ~~ 'RI\_FKey%del'::text)`}},
			},
			{
				Statement: `select proname from pg_proc where proname like E'RI\\_FKey%del' order by 1;`,
				Results:   []sql.Row{{`RI_FKey_cascade_del`}, {`RI_FKey_noaction_del`}, {`RI_FKey_restrict_del`}, {`RI_FKey_setdefault_del`}, {`RI_FKey_setnull_del`}},
			},
			{
				Statement: `explain (costs off)
select proname from pg_proc where proname ilike '00%foo' order by 1;`,
				Results: []sql.Row{{`Index Only Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: ((proname >= '00'::text) AND (proname < '01'::text))`}, {`Filter: (proname ~~* '00%foo'::text)`}},
			},
			{
				Statement: `select proname from pg_proc where proname ilike '00%foo' order by 1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `explain (costs off)
select proname from pg_proc where proname ilike 'ri%foo' order by 1;`,
				Results: []sql.Row{{`Index Only Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Filter: (proname ~~* 'ri%foo'::text)`}},
			},
			{
				Statement: `set enable_indexscan to false;`,
			},
			{
				Statement: `set enable_bitmapscan to true;`,
			},
			{
				Statement: `explain (costs off)
select proname from pg_proc where proname like E'RI\\_FKey%del' order by 1;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: proname`}, {`->  Bitmap Heap Scan on pg_proc`}, {`Filter: (proname ~~ 'RI\_FKey%del'::text)`}, {`->  Bitmap Index Scan on pg_proc_proname_args_nsp_index`}, {`Index Cond: ((proname >= 'RI_FKey'::text) AND (proname < 'RI_FKez'::text))`}},
			},
			{
				Statement: `select proname from pg_proc where proname like E'RI\\_FKey%del' order by 1;`,
				Results:   []sql.Row{{`RI_FKey_cascade_del`}, {`RI_FKey_noaction_del`}, {`RI_FKey_restrict_del`}, {`RI_FKey_setdefault_del`}, {`RI_FKey_setnull_del`}},
			},
			{
				Statement: `explain (costs off)
select proname from pg_proc where proname ilike '00%foo' order by 1;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: proname`}, {`->  Bitmap Heap Scan on pg_proc`}, {`Filter: (proname ~~* '00%foo'::text)`}, {`->  Bitmap Index Scan on pg_proc_proname_args_nsp_index`}, {`Index Cond: ((proname >= '00'::text) AND (proname < '01'::text))`}},
			},
			{
				Statement: `select proname from pg_proc where proname ilike '00%foo' order by 1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `explain (costs off)
select proname from pg_proc where proname ilike 'ri%foo' order by 1;`,
				Results: []sql.Row{{`Index Only Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Filter: (proname ~~* 'ri%foo'::text)`}},
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
				Statement: `create temp table btree_bpchar (f1 text collate "C");`,
			},
			{
				Statement: `create index on btree_bpchar(f1 bpchar_ops) WITH (deduplicate_items=on);`,
			},
			{
				Statement: `insert into btree_bpchar values ('foo'), ('fool'), ('bar'), ('quux');`,
			},
			{
				Statement: `explain (costs off)
select * from btree_bpchar where f1 like 'foo';`,
				Results: []sql.Row{{`Seq Scan on btree_bpchar`}, {`Filter: (f1 ~~ 'foo'::text)`}},
			},
			{
				Statement: `select * from btree_bpchar where f1 like 'foo';`,
				Results:   []sql.Row{{`foo`}},
			},
			{
				Statement: `explain (costs off)
select * from btree_bpchar where f1 like 'foo%';`,
				Results: []sql.Row{{`Seq Scan on btree_bpchar`}, {`Filter: (f1 ~~ 'foo%'::text)`}},
			},
			{
				Statement: `select * from btree_bpchar where f1 like 'foo%';`,
				Results:   []sql.Row{{`foo`}, {`fool`}},
			},
			{
				Statement: `explain (costs off)
select * from btree_bpchar where f1::bpchar like 'foo';`,
				Results: []sql.Row{{`Bitmap Heap Scan on btree_bpchar`}, {`Filter: ((f1)::bpchar ~~ 'foo'::text)`}, {`->  Bitmap Index Scan on btree_bpchar_f1_idx`}, {`Index Cond: ((f1)::bpchar = 'foo'::bpchar)`}},
			},
			{
				Statement: `select * from btree_bpchar where f1::bpchar like 'foo';`,
				Results:   []sql.Row{{`foo`}},
			},
			{
				Statement: `explain (costs off)
select * from btree_bpchar where f1::bpchar like 'foo%';`,
				Results: []sql.Row{{`Bitmap Heap Scan on btree_bpchar`}, {`Filter: ((f1)::bpchar ~~ 'foo%'::text)`}, {`->  Bitmap Index Scan on btree_bpchar_f1_idx`}, {`Index Cond: (((f1)::bpchar >= 'foo'::bpchar) AND ((f1)::bpchar < 'fop'::bpchar))`}},
			},
			{
				Statement: `select * from btree_bpchar where f1::bpchar like 'foo%';`,
				Results:   []sql.Row{{`foo`}, {`fool`}},
			},
			{
				Statement: `insert into btree_bpchar select 'foo' from generate_series(1,1500);`,
			},
			{
				Statement: `CREATE TABLE dedup_unique_test_table (a int) WITH (autovacuum_enabled=false);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX dedup_unique ON dedup_unique_test_table (a) WITH (deduplicate_items=on);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX plain_unique ON dedup_unique_test_table (a) WITH (deduplicate_items=off);`,
			},
			{
				Statement: `DO $$
BEGIN
    FOR r IN 1..1350 LOOP
        DELETE FROM dedup_unique_test_table;`,
			},
			{
				Statement: `        INSERT INTO dedup_unique_test_table SELECT 1;`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement: `END$$;`,
			},
			{
				Statement: `DROP INDEX plain_unique;`,
			},
			{
				Statement: `DELETE FROM dedup_unique_test_table WHERE a = 1;`,
			},
			{
				Statement: `INSERT INTO dedup_unique_test_table SELECT i FROM generate_series(0,450) i;`,
			},
			{
				Statement: `create table btree_tall_tbl(id int4, t text);`,
			},
			{
				Statement: `alter table btree_tall_tbl alter COLUMN t set storage plain;`,
			},
			{
				Statement: `create index btree_tall_idx on btree_tall_tbl (t, id) with (fillfactor = 10);`,
			},
			{
				Statement: `insert into btree_tall_tbl select g, repeat('x', 250)
from generate_series(1, 130) g;`,
			},
			{
				Statement: `CREATE TABLE delete_test_table (a bigint, b bigint, c bigint, d bigint);`,
			},
			{
				Statement: `INSERT INTO delete_test_table SELECT i, 1, 2, 3 FROM generate_series(1,80000) i;`,
			},
			{
				Statement: `ALTER TABLE delete_test_table ADD PRIMARY KEY (a,b,c,d);`,
			},
			{
				Statement: `DELETE FROM delete_test_table WHERE a < 79990;`,
			},
			{
				Statement: `VACUUM delete_test_table;`,
			},
			{
				Statement: `INSERT INTO delete_test_table SELECT i, 1, 2, 3 FROM generate_series(1,1000) i;`,
			},
			{
				Statement:   `create index on btree_tall_tbl (id int4_ops(foo=1));`,
				ErrorString: `operator class int4_ops has no options`,
			},
			{
				Statement: `CREATE INDEX btree_tall_idx2 ON btree_tall_tbl (id);`,
			},
			{
				Statement:   `ALTER INDEX btree_tall_idx2 ALTER COLUMN id SET (n_distinct=100);`,
				ErrorString: `ALTER action ALTER COLUMN ... SET cannot be performed on relation "btree_tall_idx2"`,
			},
			{
				Statement: `DROP INDEX btree_tall_idx2;`,
			},
			{
				Statement: `CREATE TABLE btree_part (id int4) PARTITION BY RANGE (id);`,
			},
			{
				Statement: `CREATE INDEX btree_part_idx ON btree_part(id);`,
			},
			{
				Statement:   `ALTER INDEX btree_part_idx ALTER COLUMN id SET (n_distinct=100);`,
				ErrorString: `ALTER action ALTER COLUMN ... SET cannot be performed on relation "btree_part_idx"`,
			},
			{
				Statement: `DROP TABLE btree_part;`,
			},
		},
	})
}
