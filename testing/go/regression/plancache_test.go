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

func TestPlancache(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_plancache)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_plancache,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TEMP TABLE pcachetest AS SELECT * FROM int8_tbl;`,
			},
			{
				Statement: `PREPARE prepstmt AS SELECT * FROM pcachetest;`,
			},
			{
				Statement: `EXECUTE prepstmt;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `PREPARE prepstmt2(bigint) AS SELECT * FROM pcachetest WHERE q1 = $1;`,
			},
			{
				Statement: `EXECUTE prepstmt2(123);`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `DROP TABLE pcachetest;`,
			},
			{
				Statement:   `EXECUTE prepstmt;`,
				ErrorString: `relation "pcachetest" does not exist`,
			},
			{
				Statement:   `EXECUTE prepstmt2(123);`,
				ErrorString: `relation "pcachetest" does not exist`,
			},
			{
				Statement: `CREATE TEMP TABLE pcachetest AS SELECT * FROM int8_tbl ORDER BY 2;`,
			},
			{
				Statement: `EXECUTE prepstmt;`,
				Results:   []sql.Row{{4567890123456789, -4567890123456789}, {4567890123456789, 123}, {123, 456}, {123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `EXECUTE prepstmt2(123);`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `ALTER TABLE pcachetest ADD COLUMN q3 bigint;`,
			},
			{
				Statement:   `EXECUTE prepstmt;`,
				ErrorString: `cached plan must not change result type`,
			},
			{
				Statement:   `EXECUTE prepstmt2(123);`,
				ErrorString: `cached plan must not change result type`,
			},
			{
				Statement: `ALTER TABLE pcachetest DROP COLUMN q3;`,
			},
			{
				Statement: `EXECUTE prepstmt;`,
				Results:   []sql.Row{{4567890123456789, -4567890123456789}, {4567890123456789, 123}, {123, 456}, {123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `EXECUTE prepstmt2(123);`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `CREATE TEMP VIEW pcacheview AS
  SELECT * FROM pcachetest;`,
			},
			{
				Statement: `PREPARE vprep AS SELECT * FROM pcacheview;`,
			},
			{
				Statement: `EXECUTE vprep;`,
				Results:   []sql.Row{{4567890123456789, -4567890123456789}, {4567890123456789, 123}, {123, 456}, {123, 4567890123456789}, {4567890123456789, 4567890123456789}},
			},
			{
				Statement: `CREATE OR REPLACE TEMP VIEW pcacheview AS
  SELECT q1, q2/2 AS q2 FROM pcachetest;`,
			},
			{
				Statement: `EXECUTE vprep;`,
				Results:   []sql.Row{{4567890123456789, -2283945061728394}, {4567890123456789, 61}, {123, 228}, {123, 2283945061728394}, {4567890123456789, 2283945061728394}},
			},
			{
				Statement: `create function cache_test(int) returns int as $$
declare total int;`,
			},
			{
				Statement: `begin
	create temp table t1(f1 int);`,
			},
			{
				Statement: `	insert into t1 values($1);`,
			},
			{
				Statement: `	insert into t1 values(11);`,
			},
			{
				Statement: `	insert into t1 values(12);`,
			},
			{
				Statement: `	insert into t1 values(13);`,
			},
			{
				Statement: `	select sum(f1) into total from t1;`,
			},
			{
				Statement: `	drop table t1;`,
			},
			{
				Statement: `	return total;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `select cache_test(1);`,
				Results:   []sql.Row{{37}},
			},
			{
				Statement: `select cache_test(2);`,
				Results:   []sql.Row{{38}},
			},
			{
				Statement: `select cache_test(3);`,
				Results:   []sql.Row{{39}},
			},
			{
				Statement: `create temp view v1 as
  select 2+2 as f1;`,
			},
			{
				Statement: `create function cache_test_2() returns int as $$
begin
	return f1 from v1;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select cache_test_2();`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `create or replace temp view v1 as
  select 2+2+4 as f1;`,
			},
			{
				Statement: `select cache_test_2();`,
				Results:   []sql.Row{{8}},
			},
			{
				Statement: `create or replace temp view v1 as
  select 2+2+4+(select max(unique1) from tenk1) as f1;`,
			},
			{
				Statement: `select cache_test_2();`,
				Results:   []sql.Row{{10007}},
			},
			{
				Statement: `create schema s1
  create table abc (f1 int);`,
			},
			{
				Statement: `create schema s2
  create table abc (f1 int);`,
			},
			{
				Statement: `insert into s1.abc values(123);`,
			},
			{
				Statement: `insert into s2.abc values(456);`,
			},
			{
				Statement: `set search_path = s1;`,
			},
			{
				Statement: `prepare p1 as select f1 from abc;`,
			},
			{
				Statement: `execute p1;`,
				Results:   []sql.Row{{123}},
			},
			{
				Statement: `set search_path = s2;`,
			},
			{
				Statement: `select f1 from abc;`,
				Results:   []sql.Row{{456}},
			},
			{
				Statement: `execute p1;`,
				Results:   []sql.Row{{456}},
			},
			{
				Statement: `alter table s1.abc add column f2 float8;   -- force replan`,
			},
			{
				Statement: `execute p1;`,
				Results:   []sql.Row{{456}},
			},
			{
				Statement: `drop schema s1 cascade;`,
			},
			{
				Statement: `drop schema s2 cascade;`,
			},
			{
				Statement: `reset search_path;`,
			},
			{
				Statement: `create temp sequence seq;`,
			},
			{
				Statement: `prepare p2 as select nextval('seq');`,
			},
			{
				Statement: `execute p2;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `drop sequence seq;`,
			},
			{
				Statement: `create temp sequence seq;`,
			},
			{
				Statement: `execute p2;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `create function cachebug() returns void as $$
declare r int;`,
			},
			{
				Statement: `begin
  drop table if exists temptable cascade;`,
			},
			{
				Statement: `  create temp table temptable as select * from generate_series(1,3) as f1;`,
			},
			{
				Statement: `  create temp view vv as select * from temptable;`,
			},
			{
				Statement: `  for r in select * from vv loop
    raise notice '%', r;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select cachebug();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select cachebug();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create table pc_list_parted (a int) partition by list(a);`,
			},
			{
				Statement: `create table pc_list_part_null partition of pc_list_parted for values in (null);`,
			},
			{
				Statement: `create table pc_list_part_1 partition of pc_list_parted for values in (1);`,
			},
			{
				Statement: `create table pc_list_part_def partition of pc_list_parted default;`,
			},
			{
				Statement: `prepare pstmt_def_insert (int) as insert into pc_list_part_def values($1);`,
			},
			{
				Statement:   `execute pstmt_def_insert(null);`,
				ErrorString: `new row for relation "pc_list_part_def" violates partition constraint`,
			},
			{
				Statement:   `execute pstmt_def_insert(1);`,
				ErrorString: `new row for relation "pc_list_part_def" violates partition constraint`,
			},
			{
				Statement: `create table pc_list_part_2 partition of pc_list_parted for values in (2);`,
			},
			{
				Statement:   `execute pstmt_def_insert(2);`,
				ErrorString: `new row for relation "pc_list_part_def" violates partition constraint`,
			},
			{
				Statement: `alter table pc_list_parted detach partition pc_list_part_null;`,
			},
			{
				Statement: `execute pstmt_def_insert(null);`,
			},
			{
				Statement: `drop table pc_list_part_1;`,
			},
			{
				Statement: `execute pstmt_def_insert(1);`,
			},
			{
				Statement: `drop table pc_list_parted, pc_list_part_null;`,
			},
			{
				Statement: `deallocate pstmt_def_insert;`,
			},
			{
				Statement: `create table test_mode (a int);`,
			},
			{
				Statement: `insert into test_mode select 1 from generate_series(1,1000) union all select 2;`,
			},
			{
				Statement: `create index on test_mode (a);`,
			},
			{
				Statement: `analyze test_mode;`,
			},
			{
				Statement: `prepare test_mode_pp (int) as select count(*) from test_mode where a = $1;`,
			},
			{
				Statement: `select name, generic_plans, custom_plans from pg_prepared_statements
  where  name = 'test_mode_pp';`,
				Results: []sql.Row{{`test_mode_pp`, 0, 0}},
			},
			{
				Statement: `set plan_cache_mode to auto;`,
			},
			{
				Statement: `explain (costs off) execute test_mode_pp(2);`,
				Results:   []sql.Row{{`Aggregate`}, {`->  Index Only Scan using test_mode_a_idx on test_mode`}, {`Index Cond: (a = 2)`}},
			},
			{
				Statement: `select name, generic_plans, custom_plans from pg_prepared_statements
  where  name = 'test_mode_pp';`,
				Results: []sql.Row{{`test_mode_pp`, 0, 1}},
			},
			{
				Statement: `set plan_cache_mode to force_generic_plan;`,
			},
			{
				Statement: `explain (costs off) execute test_mode_pp(2);`,
				Results:   []sql.Row{{`Aggregate`}, {`->  Seq Scan on test_mode`}, {`Filter: (a = $1)`}},
			},
			{
				Statement: `select name, generic_plans, custom_plans from pg_prepared_statements
  where  name = 'test_mode_pp';`,
				Results: []sql.Row{{`test_mode_pp`, 1, 1}},
			},
			{
				Statement: `set plan_cache_mode to auto;`,
			},
			{
				Statement: `execute test_mode_pp(1); -- 1x`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `execute test_mode_pp(1); -- 2x`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `execute test_mode_pp(1); -- 3x`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `execute test_mode_pp(1); -- 4x`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `select name, generic_plans, custom_plans from pg_prepared_statements
  where  name = 'test_mode_pp';`,
				Results: []sql.Row{{`test_mode_pp`, 1, 5}},
			},
			{
				Statement: `execute test_mode_pp(1); -- 5x`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `select name, generic_plans, custom_plans from pg_prepared_statements
  where  name = 'test_mode_pp';`,
				Results: []sql.Row{{`test_mode_pp`, 2, 5}},
			},
			{
				Statement: `explain (costs off) execute test_mode_pp(2);`,
				Results:   []sql.Row{{`Aggregate`}, {`->  Seq Scan on test_mode`}, {`Filter: (a = $1)`}},
			},
			{
				Statement: `set plan_cache_mode to force_custom_plan;`,
			},
			{
				Statement: `explain (costs off) execute test_mode_pp(2);`,
				Results:   []sql.Row{{`Aggregate`}, {`->  Index Only Scan using test_mode_a_idx on test_mode`}, {`Index Cond: (a = 2)`}},
			},
			{
				Statement: `select name, generic_plans, custom_plans from pg_prepared_statements
  where  name = 'test_mode_pp';`,
				Results: []sql.Row{{`test_mode_pp`, 3, 6}},
			},
			{
				Statement: `drop table test_mode;`,
			},
		},
	})
}
