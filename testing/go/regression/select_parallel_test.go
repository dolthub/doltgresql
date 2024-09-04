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

func TestSelectParallel(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_select_parallel)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_select_parallel,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup, RegressionFileName_create_misc},
		Statements: []RegressionFileStatement{
			{
				Statement: `create function sp_parallel_restricted(int) returns int as
  $$begin return $1; end$$ language plpgsql parallel restricted;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `set parallel_setup_cost=0;`,
			},
			{
				Statement: `set parallel_tuple_cost=0;`,
			},
			{
				Statement: `set min_parallel_table_scan_size=0;`,
			},
			{
				Statement: `set max_parallel_workers_per_gather=4;`,
			},
			{
				Statement: `explain (costs off)
  select round(avg(aa)), sum(aa) from a_star;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 3`}, {`->  Partial Aggregate`}, {`->  Parallel Append`}, {`->  Parallel Seq Scan on d_star a_star_4`}, {`->  Parallel Seq Scan on f_star a_star_6`}, {`->  Parallel Seq Scan on e_star a_star_5`}, {`->  Parallel Seq Scan on b_star a_star_2`}, {`->  Parallel Seq Scan on c_star a_star_3`}, {`->  Parallel Seq Scan on a_star a_star_1`}},
			},
			{
				Statement: `select round(avg(aa)), sum(aa) from a_star a1;`,
				Results:   []sql.Row{{14, 355}},
			},
			{
				Statement: `alter table c_star set (parallel_workers = 0);`,
			},
			{
				Statement: `alter table d_star set (parallel_workers = 0);`,
			},
			{
				Statement: `explain (costs off)
  select round(avg(aa)), sum(aa) from a_star;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 3`}, {`->  Partial Aggregate`}, {`->  Parallel Append`}, {`->  Seq Scan on d_star a_star_4`}, {`->  Seq Scan on c_star a_star_3`}, {`->  Parallel Seq Scan on f_star a_star_6`}, {`->  Parallel Seq Scan on e_star a_star_5`}, {`->  Parallel Seq Scan on b_star a_star_2`}, {`->  Parallel Seq Scan on a_star a_star_1`}},
			},
			{
				Statement: `select round(avg(aa)), sum(aa) from a_star a2;`,
				Results:   []sql.Row{{14, 355}},
			},
			{
				Statement: `alter table a_star set (parallel_workers = 0);`,
			},
			{
				Statement: `alter table b_star set (parallel_workers = 0);`,
			},
			{
				Statement: `alter table e_star set (parallel_workers = 0);`,
			},
			{
				Statement: `alter table f_star set (parallel_workers = 0);`,
			},
			{
				Statement: `explain (costs off)
  select round(avg(aa)), sum(aa) from a_star;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 3`}, {`->  Partial Aggregate`}, {`->  Parallel Append`}, {`->  Seq Scan on d_star a_star_4`}, {`->  Seq Scan on f_star a_star_6`}, {`->  Seq Scan on e_star a_star_5`}, {`->  Seq Scan on b_star a_star_2`}, {`->  Seq Scan on c_star a_star_3`}, {`->  Seq Scan on a_star a_star_1`}},
			},
			{
				Statement: `select round(avg(aa)), sum(aa) from a_star a3;`,
				Results:   []sql.Row{{14, 355}},
			},
			{
				Statement: `alter table a_star reset (parallel_workers);`,
			},
			{
				Statement: `alter table b_star reset (parallel_workers);`,
			},
			{
				Statement: `alter table c_star reset (parallel_workers);`,
			},
			{
				Statement: `alter table d_star reset (parallel_workers);`,
			},
			{
				Statement: `alter table e_star reset (parallel_workers);`,
			},
			{
				Statement: `alter table f_star reset (parallel_workers);`,
			},
			{
				Statement: `set enable_parallel_append to off;`,
			},
			{
				Statement: `explain (costs off)
  select round(avg(aa)), sum(aa) from a_star;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 1`}, {`->  Partial Aggregate`}, {`->  Append`}, {`->  Parallel Seq Scan on a_star a_star_1`}, {`->  Parallel Seq Scan on b_star a_star_2`}, {`->  Parallel Seq Scan on c_star a_star_3`}, {`->  Parallel Seq Scan on d_star a_star_4`}, {`->  Parallel Seq Scan on e_star a_star_5`}, {`->  Parallel Seq Scan on f_star a_star_6`}},
			},
			{
				Statement: `select round(avg(aa)), sum(aa) from a_star a4;`,
				Results:   []sql.Row{{14, 355}},
			},
			{
				Statement: `reset enable_parallel_append;`,
			},
			{
				Statement: `create function sp_test_func() returns setof text as
$$ select 'foo'::varchar union all select 'bar'::varchar $$
language sql stable;`,
			},
			{
				Statement: `select sp_test_func() order by 1;`,
				Results:   []sql.Row{{`bar`}, {`foo`}},
			},
			{
				Statement: `create table part_pa_test(a int, b int) partition by range(a);`,
			},
			{
				Statement: `create table part_pa_test_p1 partition of part_pa_test for values from (minvalue) to (0);`,
			},
			{
				Statement: `create table part_pa_test_p2 partition of part_pa_test for values from (0) to (maxvalue);`,
			},
			{
				Statement: `explain (costs off)
	select (select max((select pa1.b from part_pa_test pa1 where pa1.a = pa2.a)))
	from part_pa_test pa2;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Gather`}, {`Workers Planned: 3`}, {`->  Parallel Append`}, {`->  Parallel Seq Scan on part_pa_test_p1 pa2_1`}, {`->  Parallel Seq Scan on part_pa_test_p2 pa2_2`}, {`SubPlan 2`}, {`->  Result`}, {`SubPlan 1`}, {`->  Append`}, {`->  Seq Scan on part_pa_test_p1 pa1_1`}, {`Filter: (a = pa2.a)`}, {`->  Seq Scan on part_pa_test_p2 pa1_2`}, {`Filter: (a = pa2.a)`}},
			},
			{
				Statement: `drop table part_pa_test;`,
			},
			{
				Statement: `set parallel_leader_participation = off;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from tenk1 where stringu1 = 'GRAAAA';`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Seq Scan on tenk1`}, {`Filter: (stringu1 = 'GRAAAA'::name)`}},
			},
			{
				Statement: `select count(*) from tenk1 where stringu1 = 'GRAAAA';`,
				Results:   []sql.Row{{15}},
			},
			{
				Statement: `set max_parallel_workers = 0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from tenk1 where stringu1 = 'GRAAAA';`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Seq Scan on tenk1`}, {`Filter: (stringu1 = 'GRAAAA'::name)`}},
			},
			{
				Statement: `select count(*) from tenk1 where stringu1 = 'GRAAAA';`,
				Results:   []sql.Row{{15}},
			},
			{
				Statement: `reset max_parallel_workers;`,
			},
			{
				Statement: `reset parallel_leader_participation;`,
			},
			{
				Statement: `alter table tenk1 set (parallel_workers = 4);`,
			},
			{
				Statement: `explain (verbose, costs off)
select sp_parallel_restricted(unique1) from tenk1
  where stringu1 = 'GRAAAA' order by 1;`,
				Results: []sql.Row{{`Sort`}, {`Output: (sp_parallel_restricted(unique1))`}, {`Sort Key: (sp_parallel_restricted(tenk1.unique1))`}, {`->  Gather`}, {`Output: sp_parallel_restricted(unique1)`}, {`Workers Planned: 4`}, {`->  Parallel Seq Scan on public.tenk1`}, {`Output: unique1`}, {`Filter: (tenk1.stringu1 = 'GRAAAA'::name)`}},
			},
			{
				Statement: `explain (costs off)
	select length(stringu1) from tenk1 group by length(stringu1);`,
				Results: []sql.Row{{`Finalize HashAggregate`}, {`Group Key: (length((stringu1)::text))`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial HashAggregate`}, {`Group Key: length((stringu1)::text)`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select length(stringu1) from tenk1 group by length(stringu1);`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `explain (costs off)
	select stringu1, count(*) from tenk1 group by stringu1 order by stringu1;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: stringu1`}, {`->  Finalize HashAggregate`}, {`Group Key: stringu1`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial HashAggregate`}, {`Group Key: stringu1`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `explain (costs off)
	select  sum(sp_parallel_restricted(unique1)) from tenk1
	group by(sp_parallel_restricted(unique1));`,
				Results: []sql.Row{{`HashAggregate`}, {`Group Key: sp_parallel_restricted(unique1)`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Index Only Scan using tenk1_unique1 on tenk1`}},
			},
			{
				Statement: `prepare tenk1_count(integer) As select  count((unique1)) from tenk1 where hundred > $1;`,
			},
			{
				Statement: `explain (costs off) execute tenk1_count(1);`,
				Results:   []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Seq Scan on tenk1`}, {`Filter: (hundred > 1)`}},
			},
			{
				Statement: `execute tenk1_count(1);`,
				Results:   []sql.Row{{9800}},
			},
			{
				Statement: `deallocate tenk1_count;`,
			},
			{
				Statement: `alter table tenk2 set (parallel_workers = 0);`,
			},
			{
				Statement: `explain (costs off)
	select count(*) from tenk1 where (two, four) not in
	(select hundred, thousand from tenk2 where thousand > 100);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Seq Scan on tenk1`}, {`Filter: (NOT (hashed SubPlan 1))`}, {`SubPlan 1`}, {`->  Seq Scan on tenk2`}, {`Filter: (thousand > 100)`}},
			},
			{
				Statement: `select count(*) from tenk1 where (two, four) not in
	(select hundred, thousand from tenk2 where thousand > 100);`,
				Results: []sql.Row{{10000}},
			},
			{
				Statement: `explain (costs off)
	select * from tenk1 where (unique1 + random())::integer not in
	(select ten from tenk2);`,
				Results: []sql.Row{{`Seq Scan on tenk1`}, {`Filter: (NOT (hashed SubPlan 1))`}, {`SubPlan 1`}, {`->  Seq Scan on tenk2`}},
			},
			{
				Statement: `alter table tenk2 reset (parallel_workers);`,
			},
			{
				Statement: `set enable_indexscan = off;`,
			},
			{
				Statement: `set enable_indexonlyscan = off;`,
			},
			{
				Statement: `set enable_bitmapscan = off;`,
			},
			{
				Statement: `alter table tenk2 set (parallel_workers = 2);`,
			},
			{
				Statement: `explain (costs off)
	select count(*) from tenk1
        where tenk1.unique1 = (Select max(tenk2.unique1) from tenk2);`,
				Results: []sql.Row{{`Aggregate`}, {`InitPlan 1 (returns $2)`}, {`->  Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Partial Aggregate`}, {`->  Parallel Seq Scan on tenk2`}, {`->  Gather`}, {`Workers Planned: 4`}, {`Params Evaluated: $2`}, {`->  Parallel Seq Scan on tenk1`}, {`Filter: (unique1 = $2)`}},
			},
			{
				Statement: `select count(*) from tenk1
    where tenk1.unique1 = (Select max(tenk2.unique1) from tenk2);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `reset enable_indexscan;`,
			},
			{
				Statement: `reset enable_indexonlyscan;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `alter table tenk2 reset (parallel_workers);`,
			},
			{
				Statement: `set enable_seqscan to off;`,
			},
			{
				Statement: `set enable_bitmapscan to off;`,
			},
			{
				Statement: `explain (costs off)
	select  count((unique1)) from tenk1 where hundred > 1;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Index Scan using tenk1_hundred on tenk1`}, {`Index Cond: (hundred > 1)`}},
			},
			{
				Statement: `select  count((unique1)) from tenk1 where hundred > 1;`,
				Results:   []sql.Row{{9800}},
			},
			{
				Statement: `explain (costs off)
	select  count(*) from tenk1 where thousand > 95;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`Index Cond: (thousand > 95)`}},
			},
			{
				Statement: `select  count(*) from tenk1 where thousand > 95;`,
				Results:   []sql.Row{{9040}},
			},
			{
				Statement: `set enable_material = false;`,
			},
			{
				Statement: `explain (costs off)
select * from
  (select count(unique1) from tenk1 where hundred > 10) ss
  right join (values (1),(2),(3)) v(x) on true;`,
				Results: []sql.Row{{`Nested Loop Left Join`}, {`->  Values Scan on "*VALUES*"`}, {`->  Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Index Scan using tenk1_hundred on tenk1`}, {`Index Cond: (hundred > 10)`}},
			},
			{
				Statement: `select * from
  (select count(unique1) from tenk1 where hundred > 10) ss
  right join (values (1),(2),(3)) v(x) on true;`,
				Results: []sql.Row{{8900, 1}, {8900, 2}, {8900, 3}},
			},
			{
				Statement: `explain (costs off)
select * from
  (select count(*) from tenk1 where thousand > 99) ss
  right join (values (1),(2),(3)) v(x) on true;`,
				Results: []sql.Row{{`Nested Loop Left Join`}, {`->  Values Scan on "*VALUES*"`}, {`->  Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`Index Cond: (thousand > 99)`}},
			},
			{
				Statement: `select * from
  (select count(*) from tenk1 where thousand > 99) ss
  right join (values (1),(2),(3)) v(x) on true;`,
				Results: []sql.Row{{9000, 1}, {9000, 2}, {9000, 3}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `set enable_indexonlyscan to off;`,
			},
			{
				Statement: `set enable_indexscan to off;`,
			},
			{
				Statement: `alter table tenk1 set (parallel_workers = 0);`,
			},
			{
				Statement: `alter table tenk2 set (parallel_workers = 1);`,
			},
			{
				Statement: `explain (costs off)
select count(*) from tenk1
  left join (select tenk2.unique1 from tenk2 order by 1 limit 1000) ss
  on tenk1.unique1 < ss.unique1 + 1
  where tenk1.unique1 < 2;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Nested Loop Left Join`}, {`Join Filter: (tenk1.unique1 < (tenk2.unique1 + 1))`}, {`->  Seq Scan on tenk1`}, {`Filter: (unique1 < 2)`}, {`->  Limit`}, {`->  Gather Merge`}, {`Workers Planned: 1`}, {`->  Sort`}, {`Sort Key: tenk2.unique1`}, {`->  Parallel Seq Scan on tenk2`}},
			},
			{
				Statement: `select count(*) from tenk1
  left join (select tenk2.unique1 from tenk2 order by 1 limit 1000) ss
  on tenk1.unique1 < ss.unique1 + 1
  where tenk1.unique1 < 2;`,
				Results: []sql.Row{{1999}},
			},
			{
				Statement: `alter table tenk1 set (parallel_workers = 4);`,
			},
			{
				Statement: `alter table tenk2 reset (parallel_workers);`,
			},
			{
				Statement: `reset enable_material;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `reset enable_indexonlyscan;`,
			},
			{
				Statement: `reset enable_indexscan;`,
			},
			{
				Statement: `set enable_seqscan to off;`,
			},
			{
				Statement: `set enable_indexscan to off;`,
			},
			{
				Statement: `set enable_hashjoin to off;`,
			},
			{
				Statement: `set enable_mergejoin to off;`,
			},
			{
				Statement: `set enable_material to off;`,
			},
			{
				Statement: `DO $$
BEGIN
 SET effective_io_concurrency = 50;`,
			},
			{
				Statement: `EXCEPTION WHEN invalid_parameter_value THEN
END $$;`,
			},
			{
				Statement: `set work_mem='64kB';  --set small work mem to force lossy pages
explain (costs off)
	select count(*) from tenk1, tenk2 where tenk1.hundred > 1 and tenk2.thousand=0;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Nested Loop`}, {`->  Seq Scan on tenk2`}, {`Filter: (thousand = 0)`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Bitmap Heap Scan on tenk1`}, {`Recheck Cond: (hundred > 1)`}, {`->  Bitmap Index Scan on tenk1_hundred`}, {`Index Cond: (hundred > 1)`}},
			},
			{
				Statement: `select count(*) from tenk1, tenk2 where tenk1.hundred > 1 and tenk2.thousand=0;`,
				Results:   []sql.Row{{98000}},
			},
			{
				Statement: `create table bmscantest (a int, t text);`,
			},
			{
				Statement: `insert into bmscantest select r, 'fooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo' FROM generate_series(1,100000) r;`,
			},
			{
				Statement: `create index i_bmtest ON bmscantest(a);`,
			},
			{
				Statement: `select count(*) from bmscantest where a>1;`,
				Results:   []sql.Row{{99999}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `alter table tenk2 set (parallel_workers = 0);`,
			},
			{
				Statement: `explain (analyze, timing off, summary off, costs off)
   select count(*) from tenk1, tenk2 where tenk1.hundred > 1
        and tenk2.thousand=0;`,
				Results: []sql.Row{{`Aggregate (actual rows=1 loops=1)`}, {`->  Nested Loop (actual rows=98000 loops=1)`}, {`->  Seq Scan on tenk2 (actual rows=10 loops=1)`}, {`Filter: (thousand = 0)`}, {`Rows Removed by Filter: 9990`}, {`->  Gather (actual rows=9800 loops=10)`}, {`Workers Planned: 4`}, {`Workers Launched: 4`}, {`->  Parallel Seq Scan on tenk1 (actual rows=1960 loops=50)`}, {`Filter: (hundred > 1)`}, {`Rows Removed by Filter: 40`}},
			},
			{
				Statement: `alter table tenk2 reset (parallel_workers);`,
			},
			{
				Statement: `reset work_mem;`,
			},
			{
				Statement: `create function explain_parallel_sort_stats() returns setof text
language plpgsql as
$$
declare ln text;`,
			},
			{
				Statement: `begin
    for ln in
        explain (analyze, timing off, summary off, costs off)
          select * from
          (select ten from tenk1 where ten < 100 order by ten) ss
          right join (values (1),(2),(3)) v(x) on true
    loop
        ln := regexp_replace(ln, 'Memory: \S*',  'Memory: xxx');`,
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
				Statement: `select * from explain_parallel_sort_stats();`,
				Results:   []sql.Row{{`Nested Loop Left Join (actual rows=30000 loops=1)`}, {`->  Values Scan on "*VALUES*" (actual rows=3 loops=1)`}, {`->  Gather Merge (actual rows=10000 loops=3)`}, {`Workers Planned: 4`}, {`Workers Launched: 4`}, {`->  Sort (actual rows=2000 loops=15)`}, {`Sort Key: tenk1.ten`}, {`Sort Method: quicksort  Memory: xxx`}, {`Worker 0:  Sort Method: quicksort  Memory: xxx`}, {`Worker 1:  Sort Method: quicksort  Memory: xxx`}, {`Worker 2:  Sort Method: quicksort  Memory: xxx`}, {`Worker 3:  Sort Method: quicksort  Memory: xxx`}, {`->  Parallel Seq Scan on tenk1 (actual rows=2000 loops=15)`}, {`Filter: (ten < 100)`}},
			},
			{
				Statement: `reset enable_indexscan;`,
			},
			{
				Statement: `reset enable_hashjoin;`,
			},
			{
				Statement: `reset enable_mergejoin;`,
			},
			{
				Statement: `reset enable_material;`,
			},
			{
				Statement: `reset effective_io_concurrency;`,
			},
			{
				Statement: `drop table bmscantest;`,
			},
			{
				Statement: `drop function explain_parallel_sort_stats();`,
			},
			{
				Statement: `set enable_hashjoin to off;`,
			},
			{
				Statement: `set enable_nestloop to off;`,
			},
			{
				Statement: `explain (costs off)
	select  count(*) from tenk1, tenk2 where tenk1.unique1 = tenk2.unique1;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Merge Join`}, {`Merge Cond: (tenk1.unique1 = tenk2.unique1)`}, {`->  Parallel Index Only Scan using tenk1_unique1 on tenk1`}, {`->  Index Only Scan using tenk2_unique1 on tenk2`}},
			},
			{
				Statement: `select  count(*) from tenk1, tenk2 where tenk1.unique1 = tenk2.unique1;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement: `reset enable_hashjoin;`,
			},
			{
				Statement: `reset enable_nestloop;`,
			},
			{
				Statement: `set enable_hashagg = false;`,
			},
			{
				Statement: `explain (costs off)
   select count(*) from tenk1 group by twenty;`,
				Results: []sql.Row{{`Finalize GroupAggregate`}, {`Group Key: twenty`}, {`->  Gather Merge`}, {`Workers Planned: 4`}, {`->  Partial GroupAggregate`}, {`Group Key: twenty`}, {`->  Sort`}, {`Sort Key: twenty`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select count(*) from tenk1 group by twenty;`,
				Results:   []sql.Row{{500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}},
			},
			{
				Statement: `create function sp_simple_func(var1 integer) returns integer
as $$
begin
        return var1 + 10;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql PARALLEL SAFE;`,
			},
			{
				Statement: `explain (costs off, verbose)
    select ten, sp_simple_func(ten) from tenk1 where ten < 100 order by ten;`,
				Results: []sql.Row{{`Gather Merge`}, {`Output: ten, (sp_simple_func(ten))`}, {`Workers Planned: 4`}, {`->  Result`}, {`Output: ten, sp_simple_func(ten)`}, {`->  Sort`}, {`Output: ten`}, {`Sort Key: tenk1.ten`}, {`->  Parallel Seq Scan on public.tenk1`}, {`Output: ten`}, {`Filter: (tenk1.ten < 100)`}},
			},
			{
				Statement: `drop function sp_simple_func(integer);`,
			},
			{
				Statement: `explain (costs off)
   select count(*), generate_series(1,2) from tenk1 group by twenty;`,
				Results: []sql.Row{{`ProjectSet`}, {`->  Finalize GroupAggregate`}, {`Group Key: twenty`}, {`->  Gather Merge`}, {`Workers Planned: 4`}, {`->  Partial GroupAggregate`}, {`Group Key: twenty`}, {`->  Sort`}, {`Sort Key: twenty`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select count(*), generate_series(1,2) from tenk1 group by twenty;`,
				Results:   []sql.Row{{500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}, {500, 1}, {500, 2}},
			},
			{
				Statement: `set parallel_leader_participation = off;`,
			},
			{
				Statement: `explain (costs off)
   select count(*) from tenk1 group by twenty;`,
				Results: []sql.Row{{`Finalize GroupAggregate`}, {`Group Key: twenty`}, {`->  Gather Merge`}, {`Workers Planned: 4`}, {`->  Partial GroupAggregate`}, {`Group Key: twenty`}, {`->  Sort`}, {`Sort Key: twenty`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select count(*) from tenk1 group by twenty;`,
				Results:   []sql.Row{{500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}, {500}},
			},
			{
				Statement: `reset parallel_leader_participation;`,
			},
			{
				Statement: `set enable_material = false;`,
			},
			{
				Statement: `explain (costs off)
select * from
  (select string4, count(unique2)
   from tenk1 group by string4 order by string4) ss
  right join (values (1),(2),(3)) v(x) on true;`,
				Results: []sql.Row{{`Nested Loop Left Join`}, {`->  Values Scan on "*VALUES*"`}, {`->  Finalize GroupAggregate`}, {`Group Key: tenk1.string4`}, {`->  Gather Merge`}, {`Workers Planned: 4`}, {`->  Partial GroupAggregate`}, {`Group Key: tenk1.string4`}, {`->  Sort`}, {`Sort Key: tenk1.string4`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select * from
  (select string4, count(unique2)
   from tenk1 group by string4 order by string4) ss
  right join (values (1),(2),(3)) v(x) on true;`,
				Results: []sql.Row{{`AAAAxx`, 2500, 1}, {`HHHHxx`, 2500, 1}, {`OOOOxx`, 2500, 1}, {`VVVVxx`, 2500, 1}, {`AAAAxx`, 2500, 2}, {`HHHHxx`, 2500, 2}, {`OOOOxx`, 2500, 2}, {`VVVVxx`, 2500, 2}, {`AAAAxx`, 2500, 3}, {`HHHHxx`, 2500, 3}, {`OOOOxx`, 2500, 3}, {`VVVVxx`, 2500, 3}},
			},
			{
				Statement: `reset enable_material;`,
			},
			{
				Statement: `reset enable_hashagg;`,
			},
			{
				Statement: `explain (costs off)
select avg(unique1::int8) from tenk1;`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial Aggregate`}, {`->  Parallel Index Only Scan using tenk1_unique1 on tenk1`}},
			},
			{
				Statement: `select avg(unique1::int8) from tenk1;`,
				Results:   []sql.Row{{4999.5000000000000000}},
			},
			{
				Statement: `explain (costs off)
  select fivethous from tenk1 order by fivethous limit 4;`,
				Results: []sql.Row{{`Limit`}, {`->  Gather Merge`}, {`Workers Planned: 4`}, {`->  Sort`}, {`Sort Key: fivethous`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select fivethous from tenk1 order by fivethous limit 4;`,
				Results:   []sql.Row{{0}, {0}, {1}, {1}},
			},
			{
				Statement: `set max_parallel_workers = 0;`,
			},
			{
				Statement: `explain (costs off)
   select string4 from tenk1 order by string4 limit 5;`,
				Results: []sql.Row{{`Limit`}, {`->  Gather Merge`}, {`Workers Planned: 4`}, {`->  Sort`}, {`Sort Key: string4`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select string4 from tenk1 order by string4 limit 5;`,
				Results:   []sql.Row{{`AAAAxx`}, {`AAAAxx`}, {`AAAAxx`}, {`AAAAxx`}, {`AAAAxx`}},
			},
			{
				Statement: `set parallel_leader_participation = off;`,
			},
			{
				Statement: `explain (costs off)
   select string4 from tenk1 order by string4 limit 5;`,
				Results: []sql.Row{{`Limit`}, {`->  Gather Merge`}, {`Workers Planned: 4`}, {`->  Sort`}, {`Sort Key: string4`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select string4 from tenk1 order by string4 limit 5;`,
				Results:   []sql.Row{{`AAAAxx`}, {`AAAAxx`}, {`AAAAxx`}, {`AAAAxx`}, {`AAAAxx`}},
			},
			{
				Statement: `reset parallel_leader_participation;`,
			},
			{
				Statement: `reset max_parallel_workers;`,
			},
			{
				Statement: `SAVEPOINT settings;`,
			},
			{
				Statement: `SET LOCAL force_parallel_mode = 1;`,
			},
			{
				Statement: `explain (costs off)
  select stringu1::int2 from tenk1 where unique1 = 1;`,
				Results: []sql.Row{{`Gather`}, {`Workers Planned: 1`}, {`Single Copy: true`}, {`->  Index Scan using tenk1_unique1 on tenk1`}, {`Index Cond: (unique1 = 1)`}},
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT settings;`,
			},
			{
				Statement: `CREATE FUNCTION make_record(n int)
  RETURNS RECORD LANGUAGE plpgsql PARALLEL SAFE AS
$$
BEGIN
  RETURN CASE n
           WHEN 1 THEN ROW(1)
           WHEN 2 THEN ROW(1, 2)
           WHEN 3 THEN ROW(1, 2, 3)
           WHEN 4 THEN ROW(1, 2, 3, 4)
           ELSE ROW(1, 2, 3, 4, 5)
         END;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SAVEPOINT settings;`,
			},
			{
				Statement: `SET LOCAL force_parallel_mode = 1;`,
			},
			{
				Statement: `SELECT make_record(x) FROM (SELECT generate_series(1, 5) x) ss ORDER BY x;`,
				Results:   []sql.Row{{`(1)`}, {`(1,2)`}, {`(1,2,3)`}, {`(1,2,3,4)`}, {`(1,2,3,4,5)`}},
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT settings;`,
			},
			{
				Statement: `DROP function make_record(n int);`,
			},
			{
				Statement: `drop role if exists regress_parallel_worker;`,
			},
			{
				Statement: `create role regress_parallel_worker;`,
			},
			{
				Statement: `set role regress_parallel_worker;`,
			},
			{
				Statement: `reset session authorization;`,
			},
			{
				Statement: `drop role regress_parallel_worker;`,
			},
			{
				Statement: `set force_parallel_mode = 1;`,
			},
			{
				Statement: `select count(*) from tenk1;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement: `reset force_parallel_mode;`,
			},
			{
				Statement: `reset role;`,
			},
			{
				Statement: `explain (costs off, verbose)
  select count(*) from tenk1 a where (unique1, two) in
    (select unique1, row_number() over() from tenk1 b);`,
				Results: []sql.Row{{`Aggregate`}, {`Output: count(*)`}, {`->  Hash Semi Join`}, {`Hash Cond: ((a.unique1 = b.unique1) AND (a.two = (row_number() OVER (?))))`}, {`->  Gather`}, {`Output: a.unique1, a.two`}, {`Workers Planned: 4`}, {`->  Parallel Seq Scan on public.tenk1 a`}, {`Output: a.unique1, a.two`}, {`->  Hash`}, {`Output: b.unique1, (row_number() OVER (?))`}, {`->  WindowAgg`}, {`Output: b.unique1, row_number() OVER (?)`}, {`->  Gather`}, {`Output: b.unique1`}, {`Workers Planned: 4`}, {`->  Parallel Index Only Scan using tenk1_unique1 on public.tenk1 b`}, {`Output: b.unique1`}},
			},
			{
				Statement: `explain (costs off)
  select * from tenk1 a where two in
    (select two from tenk1 b where stringu1 like '%AAAA' limit 3);`,
				Results: []sql.Row{{`Hash Semi Join`}, {`Hash Cond: (a.two = b.two)`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Seq Scan on tenk1 a`}, {`->  Hash`}, {`->  Limit`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Seq Scan on tenk1 b`}, {`Filter: (stringu1 ~~ '%AAAA'::text)`}},
			},
			{
				Statement: `SAVEPOINT settings;`,
			},
			{
				Statement: `SET LOCAL force_parallel_mode = 1;`,
			},
			{
				Statement: `EXPLAIN (analyze, timing off, summary off, costs off) SELECT * FROM tenk1;`,
				Results:   []sql.Row{{`Gather (actual rows=10000 loops=1)`}, {`Workers Planned: 4`}, {`Workers Launched: 4`}, {`->  Parallel Seq Scan on tenk1 (actual rows=2000 loops=5)`}},
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT settings;`,
			},
			{
				Statement: `SAVEPOINT settings;`,
			},
			{
				Statement: `SET LOCAL force_parallel_mode = 1;`,
			},
			{
				Statement:   `select (stringu1 || repeat('abcd', 5000))::int2 from tenk1 where unique1 = 1;`,
				ErrorString: `invalid input syntax for type smallint: "BAAAAAabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd"`,
			},
			{
				Statement: `CONTEXT:  parallel worker
ROLLBACK TO SAVEPOINT settings;`,
			},
			{
				Statement: `SAVEPOINT settings;`,
			},
			{
				Statement: `SET LOCAL parallel_setup_cost = 10;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT unique1 FROM tenk1 WHERE fivethous = tenthous + 1
UNION ALL
SELECT unique1 FROM tenk1 WHERE fivethous = tenthous + 1;`,
				Results: []sql.Row{{`Gather`}, {`Workers Planned: 4`}, {`->  Parallel Append`}, {`->  Parallel Seq Scan on tenk1`}, {`Filter: (fivethous = (tenthous + 1))`}, {`->  Parallel Seq Scan on tenk1 tenk1_1`}, {`Filter: (fivethous = (tenthous + 1))`}},
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT settings;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT unique1 FROM tenk1 WHERE fivethous =
	(SELECT unique1 FROM tenk1 WHERE fivethous = 1 LIMIT 1)
UNION ALL
SELECT unique1 FROM tenk1 WHERE fivethous =
	(SELECT unique2 FROM tenk1 WHERE fivethous = 1 LIMIT 1)
ORDER BY 1;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: tenk1.unique1`}, {`->  Append`}, {`->  Gather`}, {`Workers Planned: 4`}, {`Params Evaluated: $1`}, {`InitPlan 1 (returns $1)`}, {`->  Limit`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Seq Scan on tenk1 tenk1_2`}, {`Filter: (fivethous = 1)`}, {`->  Parallel Seq Scan on tenk1`}, {`Filter: (fivethous = $1)`}, {`->  Gather`}, {`Workers Planned: 4`}, {`Params Evaluated: $3`}, {`InitPlan 2 (returns $3)`}, {`->  Limit`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Seq Scan on tenk1 tenk1_3`}, {`Filter: (fivethous = 1)`}, {`->  Parallel Seq Scan on tenk1 tenk1_1`}, {`Filter: (fivethous = $3)`}},
			},
			{
				Statement: `SELECT * FROM information_schema.foreign_data_wrapper_options
ORDER BY 1, 2, 3;`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
SELECT generate_series(1, two), array(select generate_series(1, two))
  FROM tenk1 ORDER BY tenthous;`,
				Results: []sql.Row{{`ProjectSet`}, {`Output: generate_series(1, tenk1.two), (SubPlan 1), tenk1.tenthous`}, {`->  Gather Merge`}, {`Output: tenk1.two, tenk1.tenthous`}, {`Workers Planned: 4`}, {`->  Result`}, {`Output: tenk1.two, tenk1.tenthous`}, {`->  Sort`}, {`Output: tenk1.tenthous, tenk1.two`}, {`Sort Key: tenk1.tenthous`}, {`->  Parallel Seq Scan on public.tenk1`}, {`Output: tenk1.tenthous, tenk1.two`}, {`SubPlan 1`}, {`->  ProjectSet`}, {`Output: generate_series(1, tenk1.two)`}, {`->  Result`}},
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
SELECT unnest(ARRAY[]::integer[]) + 1 AS pathkey
  FROM tenk1 t1 JOIN tenk1 t2 ON TRUE
  ORDER BY pathkey;`,
				Results: []sql.Row{{`Sort`}, {`Output: (((unnest('{}'::integer[])) + 1))`}, {`Sort Key: (((unnest('{}'::integer[])) + 1))`}, {`->  Result`}, {`Output: ((unnest('{}'::integer[])) + 1)`}, {`->  ProjectSet`}, {`Output: unnest('{}'::integer[])`}, {`->  Nested Loop`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Index Only Scan using tenk1_hundred on public.tenk1 t1`}, {`->  Materialize`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Index Only Scan using tenk1_hundred on public.tenk1 t2`}},
			},
			{
				Statement: `CREATE FUNCTION make_some_array(int,int) returns int[] as
$$declare x int[];`,
			},
			{
				Statement: `  begin
    x[1] := $1;`,
			},
			{
				Statement: `    x[2] := $2;`,
			},
			{
				Statement: `    return x;`,
			},
			{
				Statement: `  end$$ language plpgsql parallel safe;`,
			},
			{
				Statement: `CREATE TABLE fooarr(f1 text, f2 int[], f3 text);`,
			},
			{
				Statement: `INSERT INTO fooarr VALUES('1', ARRAY[1,2], 'one');`,
			},
			{
				Statement: `PREPARE pstmt(text, int[]) AS SELECT * FROM fooarr WHERE f1 = $1 AND f2 = $2;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) EXECUTE pstmt('1', make_some_array(1,2));`,
				Results:   []sql.Row{{`Gather`}, {`Workers Planned: 3`}, {`->  Parallel Seq Scan on fooarr`}, {`Filter: ((f1 = '1'::text) AND (f2 = '{1,2}'::integer[]))`}},
			},
			{
				Statement: `EXECUTE pstmt('1', make_some_array(1,2));`,
				Results:   []sql.Row{{1, `{1,2}`, `one`}},
			},
			{
				Statement: `DEALLOCATE pstmt;`,
			},
			{
				Statement: `CREATE VIEW tenk1_vw_sec WITH (security_barrier) AS SELECT * FROM tenk1;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT 1 FROM tenk1_vw_sec
  WHERE (SELECT sum(f1) FROM int4_tbl WHERE f1 < unique1) < 100;`,
				Results: []sql.Row{{`Subquery Scan on tenk1_vw_sec`}, {`Filter: ((SubPlan 1) < 100)`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Parallel Index Only Scan using tenk1_unique1 on tenk1`}, {`SubPlan 1`}, {`->  Aggregate`}, {`->  Seq Scan on int4_tbl`}, {`Filter: (f1 < tenk1_vw_sec.unique1)`}},
			},
			{
				Statement: `rollback;`,
			},
		},
	})
}
