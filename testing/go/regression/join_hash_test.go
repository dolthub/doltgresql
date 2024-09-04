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

func TestJoinHash(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_join_hash)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_join_hash,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `begin;`,
			},
			{
				Statement: `set local min_parallel_table_scan_size = 0;`,
			},
			{
				Statement: `set local parallel_setup_cost = 0;`,
			},
			{
				Statement: `set local enable_hashjoin = on;`,
			},
			{
				Statement: `create or replace function find_hash(node json)
returns json language plpgsql
as
$$
declare
  x json;`,
			},
			{
				Statement: `  child json;`,
			},
			{
				Statement: `begin
  if node->>'Node Type' = 'Hash' then
    return node;`,
			},
			{
				Statement: `  else
    for child in select json_array_elements(node->'Plans')
    loop
      x := find_hash(child);`,
			},
			{
				Statement: `      if x is not null then
        return x;`,
			},
			{
				Statement: `      end if;`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create or replace function hash_join_batches(query text)
returns table (original int, final int) language plpgsql
as
$$
declare
  whole_plan json;`,
			},
			{
				Statement: `  hash_node json;`,
			},
			{
				Statement: `begin
  for whole_plan in
    execute 'explain (analyze, format ''json'') ' || query
  loop
    hash_node := find_hash(json_extract_path(whole_plan, '0', 'Plan'));`,
			},
			{
				Statement: `    original := hash_node->>'Original Hash Batches';`,
			},
			{
				Statement: `    final := hash_node->>'Hash Batches';`,
			},
			{
				Statement: `    return next;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create table simple as
  select generate_series(1, 20000) AS id, 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa';`,
			},
			{
				Statement: `alter table simple set (parallel_workers = 2);`,
			},
			{
				Statement: `analyze simple;`,
			},
			{
				Statement: `create table bigger_than_it_looks as
  select generate_series(1, 20000) as id, 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa';`,
			},
			{
				Statement: `alter table bigger_than_it_looks set (autovacuum_enabled = 'false');`,
			},
			{
				Statement: `alter table bigger_than_it_looks set (parallel_workers = 2);`,
			},
			{
				Statement: `analyze bigger_than_it_looks;`,
			},
			{
				Statement: `update pg_class set reltuples = 1000 where relname = 'bigger_than_it_looks';`,
			},
			{
				Statement: `create table extremely_skewed (id int, t text);`,
			},
			{
				Statement: `alter table extremely_skewed set (autovacuum_enabled = 'false');`,
			},
			{
				Statement: `alter table extremely_skewed set (parallel_workers = 2);`,
			},
			{
				Statement: `analyze extremely_skewed;`,
			},
			{
				Statement: `insert into extremely_skewed
  select 42 as id, 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
  from generate_series(1, 20000);`,
			},
			{
				Statement: `update pg_class
  set reltuples = 2, relpages = pg_relation_size('extremely_skewed') / 8192
  where relname = 'extremely_skewed';`,
			},
			{
				Statement: `create table wide as select generate_series(1, 2) as id, rpad('', 320000, 'x') as t;`,
			},
			{
				Statement: `alter table wide set (parallel_workers = 2);`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 0;`,
			},
			{
				Statement: `set local work_mem = '4MB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join simple s using (id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on simple s`}},
			},
			{
				Statement: `select count(*) from simple r join simple s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) from simple r join simple s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false, false}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set local work_mem = '4MB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local enable_parallel_hash = off;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join simple s using (id);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Partial Aggregate`}, {`->  Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Parallel Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on simple s`}},
			},
			{
				Statement: `select count(*) from simple r join simple s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) from simple r join simple s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false, false}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set local work_mem = '4MB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local enable_parallel_hash = on;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join simple s using (id);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Partial Aggregate`}, {`->  Parallel Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Parallel Seq Scan on simple r`}, {`->  Parallel Hash`}, {`->  Parallel Seq Scan on simple s`}},
			},
			{
				Statement: `select count(*) from simple r join simple s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) from simple r join simple s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false, false}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 0;`,
			},
			{
				Statement: `set local work_mem = '128kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join simple s using (id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on simple s`}},
			},
			{
				Statement: `select count(*) from simple r join simple s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) from simple r join simple s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true, false}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set local work_mem = '128kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local enable_parallel_hash = off;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join simple s using (id);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Partial Aggregate`}, {`->  Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Parallel Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on simple s`}},
			},
			{
				Statement: `select count(*) from simple r join simple s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) from simple r join simple s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true, false}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set local work_mem = '192kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local enable_parallel_hash = on;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join simple s using (id);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Partial Aggregate`}, {`->  Parallel Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Parallel Seq Scan on simple r`}, {`->  Parallel Hash`}, {`->  Parallel Seq Scan on simple s`}},
			},
			{
				Statement: `select count(*) from simple r join simple s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) from simple r join simple s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true, false}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 0;`,
			},
			{
				Statement: `set local work_mem = '128kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) FROM simple r JOIN bigger_than_it_looks s USING (id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on bigger_than_it_looks s`}},
			},
			{
				Statement: `select count(*) FROM simple r JOIN bigger_than_it_looks s USING (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) FROM simple r JOIN bigger_than_it_looks s USING (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false, true}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set local work_mem = '128kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local enable_parallel_hash = off;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join bigger_than_it_looks s using (id);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Partial Aggregate`}, {`->  Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Parallel Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on bigger_than_it_looks s`}},
			},
			{
				Statement: `select count(*) from simple r join bigger_than_it_looks s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) from simple r join bigger_than_it_looks s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false, true}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 1;`,
			},
			{
				Statement: `set local work_mem = '192kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local enable_parallel_hash = on;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join bigger_than_it_looks s using (id);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 1`}, {`->  Partial Aggregate`}, {`->  Parallel Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Parallel Seq Scan on simple r`}, {`->  Parallel Hash`}, {`->  Parallel Seq Scan on bigger_than_it_looks s`}},
			},
			{
				Statement: `select count(*) from simple r join bigger_than_it_looks s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select original > 1 as initially_multibatch, final > original as increased_batches
  from hash_join_batches(
$$
  select count(*) from simple r join bigger_than_it_looks s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false, true}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 0;`,
			},
			{
				Statement: `set local work_mem = '128kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join extremely_skewed s using (id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on extremely_skewed s`}},
			},
			{
				Statement: `select count(*) from simple r join extremely_skewed s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select * from hash_join_batches(
$$
  select count(*) from simple r join extremely_skewed s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set local work_mem = '128kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local enable_parallel_hash = off;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join extremely_skewed s using (id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Parallel Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on extremely_skewed s`}},
			},
			{
				Statement: `select count(*) from simple r join extremely_skewed s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select * from hash_join_batches(
$$
  select count(*) from simple r join extremely_skewed s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 1;`,
			},
			{
				Statement: `set local work_mem = '128kB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local enable_parallel_hash = on;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from simple r join extremely_skewed s using (id);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 1`}, {`->  Partial Aggregate`}, {`->  Parallel Hash Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Parallel Seq Scan on simple r`}, {`->  Parallel Hash`}, {`->  Parallel Seq Scan on extremely_skewed s`}},
			},
			{
				Statement: `select count(*) from simple r join extremely_skewed s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `select * from hash_join_batches(
$$
  select count(*) from simple r join extremely_skewed s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{1, 4}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set local work_mem = '4MB';`,
			},
			{
				Statement: `set local hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `set local parallel_leader_participation = off;`,
			},
			{
				Statement: `select * from hash_join_batches(
$$
  select count(*) from simple r join simple s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{1, 1}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `create table join_foo as select generate_series(1, 3) as id, 'xxxxx'::text as t;`,
			},
			{
				Statement: `alter table join_foo set (parallel_workers = 0);`,
			},
			{
				Statement: `create table join_bar as select generate_series(1, 10000) as id, 'xxxxx'::text as t;`,
			},
			{
				Statement: `alter table join_bar set (parallel_workers = 2);`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set enable_parallel_hash = off;`,
			},
			{
				Statement: `set parallel_leader_participation = off;`,
			},
			{
				Statement: `set min_parallel_table_scan_size = 0;`,
			},
			{
				Statement: `set parallel_setup_cost = 0;`,
			},
			{
				Statement: `set parallel_tuple_cost = 0;`,
			},
			{
				Statement: `set max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set enable_material = off;`,
			},
			{
				Statement: `set enable_mergejoin = off;`,
			},
			{
				Statement: `set work_mem = '64kB';`,
			},
			{
				Statement: `set hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from join_foo
    left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
    on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Nested Loop Left Join`}, {`Join Filter: ((join_foo.id < (b1.id + 1)) AND (join_foo.id > (b1.id - 1)))`}, {`->  Seq Scan on join_foo`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Hash Join`}, {`Hash Cond: (b1.id = b2.id)`}, {`->  Parallel Seq Scan on join_bar b1`}, {`->  Hash`}, {`->  Seq Scan on join_bar b2`}},
			},
			{
				Statement: `select count(*) from join_foo
  left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
  on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
				Results: []sql.Row{{3}},
			},
			{
				Statement: `select final > 1 as multibatch
  from hash_join_batches(
$$
  select count(*) from join_foo
    left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
    on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set enable_parallel_hash = off;`,
			},
			{
				Statement: `set parallel_leader_participation = off;`,
			},
			{
				Statement: `set min_parallel_table_scan_size = 0;`,
			},
			{
				Statement: `set parallel_setup_cost = 0;`,
			},
			{
				Statement: `set parallel_tuple_cost = 0;`,
			},
			{
				Statement: `set max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set enable_material = off;`,
			},
			{
				Statement: `set enable_mergejoin = off;`,
			},
			{
				Statement: `set work_mem = '4MB';`,
			},
			{
				Statement: `set hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from join_foo
    left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
    on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Nested Loop Left Join`}, {`Join Filter: ((join_foo.id < (b1.id + 1)) AND (join_foo.id > (b1.id - 1)))`}, {`->  Seq Scan on join_foo`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Hash Join`}, {`Hash Cond: (b1.id = b2.id)`}, {`->  Parallel Seq Scan on join_bar b1`}, {`->  Hash`}, {`->  Seq Scan on join_bar b2`}},
			},
			{
				Statement: `select count(*) from join_foo
  left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
  on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
				Results: []sql.Row{{3}},
			},
			{
				Statement: `select final > 1 as multibatch
  from hash_join_batches(
$$
  select count(*) from join_foo
    left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
    on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set enable_parallel_hash = on;`,
			},
			{
				Statement: `set parallel_leader_participation = off;`,
			},
			{
				Statement: `set min_parallel_table_scan_size = 0;`,
			},
			{
				Statement: `set parallel_setup_cost = 0;`,
			},
			{
				Statement: `set parallel_tuple_cost = 0;`,
			},
			{
				Statement: `set max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set enable_material = off;`,
			},
			{
				Statement: `set enable_mergejoin = off;`,
			},
			{
				Statement: `set work_mem = '64kB';`,
			},
			{
				Statement: `set hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from join_foo
    left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
    on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Nested Loop Left Join`}, {`Join Filter: ((join_foo.id < (b1.id + 1)) AND (join_foo.id > (b1.id - 1)))`}, {`->  Seq Scan on join_foo`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Parallel Hash Join`}, {`Hash Cond: (b1.id = b2.id)`}, {`->  Parallel Seq Scan on join_bar b1`}, {`->  Parallel Hash`}, {`->  Parallel Seq Scan on join_bar b2`}},
			},
			{
				Statement: `select count(*) from join_foo
  left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
  on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
				Results: []sql.Row{{3}},
			},
			{
				Statement: `select final > 1 as multibatch
  from hash_join_batches(
$$
  select count(*) from join_foo
    left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
    on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set enable_parallel_hash = on;`,
			},
			{
				Statement: `set parallel_leader_participation = off;`,
			},
			{
				Statement: `set min_parallel_table_scan_size = 0;`,
			},
			{
				Statement: `set parallel_setup_cost = 0;`,
			},
			{
				Statement: `set parallel_tuple_cost = 0;`,
			},
			{
				Statement: `set max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set enable_material = off;`,
			},
			{
				Statement: `set enable_mergejoin = off;`,
			},
			{
				Statement: `set work_mem = '4MB';`,
			},
			{
				Statement: `set hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select count(*) from join_foo
    left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
    on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Nested Loop Left Join`}, {`Join Filter: ((join_foo.id < (b1.id + 1)) AND (join_foo.id > (b1.id - 1)))`}, {`->  Seq Scan on join_foo`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Parallel Hash Join`}, {`Hash Cond: (b1.id = b2.id)`}, {`->  Parallel Seq Scan on join_bar b1`}, {`->  Parallel Hash`}, {`->  Parallel Seq Scan on join_bar b2`}},
			},
			{
				Statement: `select count(*) from join_foo
  left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
  on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
				Results: []sql.Row{{3}},
			},
			{
				Statement: `select final > 1 as multibatch
  from hash_join_batches(
$$
  select count(*) from join_foo
    left join (select b1.id, b1.t from join_bar b1 join join_bar b2 using (id)) ss
    on join_foo.id < ss.id + 1 and join_foo.id > ss.id - 1;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 0;`,
			},
			{
				Statement: `explain (costs off)
     select  count(*) from simple r full outer join simple s using (id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Full Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on simple s`}},
			},
			{
				Statement: `select  count(*) from simple r full outer join simple s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `explain (costs off)
     select  count(*) from simple r full outer join simple s using (id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Full Join`}, {`Hash Cond: (r.id = s.id)`}, {`->  Seq Scan on simple r`}, {`->  Hash`}, {`->  Seq Scan on simple s`}},
			},
			{
				Statement: `select  count(*) from simple r full outer join simple s using (id);`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 0;`,
			},
			{
				Statement: `explain (costs off)
     select  count(*) from simple r full outer join simple s on (r.id = 0 - s.id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Full Join`}, {`Hash Cond: ((0 - s.id) = r.id)`}, {`->  Seq Scan on simple s`}, {`->  Hash`}, {`->  Seq Scan on simple r`}},
			},
			{
				Statement: `select  count(*) from simple r full outer join simple s on (r.id = 0 - s.id);`,
				Results:   []sql.Row{{40000}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set local max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `explain (costs off)
     select  count(*) from simple r full outer join simple s on (r.id = 0 - s.id);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Hash Full Join`}, {`Hash Cond: ((0 - s.id) = r.id)`}, {`->  Seq Scan on simple s`}, {`->  Hash`}, {`->  Seq Scan on simple r`}},
			},
			{
				Statement: `select  count(*) from simple r full outer join simple s on (r.id = 0 - s.id);`,
				Results:   []sql.Row{{40000}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `savepoint settings;`,
			},
			{
				Statement: `set max_parallel_workers_per_gather = 2;`,
			},
			{
				Statement: `set enable_parallel_hash = on;`,
			},
			{
				Statement: `set work_mem = '128kB';`,
			},
			{
				Statement: `set hash_mem_multiplier = 1.0;`,
			},
			{
				Statement: `explain (costs off)
  select length(max(s.t))
  from wide left join (select id, coalesce(t, '') || '' as t from wide) s using (id);`,
				Results: []sql.Row{{`Finalize Aggregate`}, {`->  Gather`}, {`Workers Planned: 2`}, {`->  Partial Aggregate`}, {`->  Parallel Hash Left Join`}, {`Hash Cond: (wide.id = wide_1.id)`}, {`->  Parallel Seq Scan on wide`}, {`->  Parallel Hash`}, {`->  Parallel Seq Scan on wide wide_1`}},
			},
			{
				Statement: `select length(max(s.t))
from wide left join (select id, coalesce(t, '') || '' as t from wide) s using (id);`,
				Results: []sql.Row{{320000}},
			},
			{
				Statement: `select final > 1 as multibatch
  from hash_join_batches(
$$
  select length(max(s.t))
  from wide left join (select id, coalesce(t, '') || '' as t from wide) s using (id);`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `rollback to settings;`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_sort = OFF; -- avoid mergejoins`,
			},
			{
				Statement: `SET LOCAL from_collapse_limit = 1; -- allows easy changing of join order`,
			},
			{
				Statement: `CREATE TABLE hjtest_1 (a text, b int, id int, c bool);`,
			},
			{
				Statement: `CREATE TABLE hjtest_2 (a bool, id int, b text, c int);`,
			},
			{
				Statement: `INSERT INTO hjtest_1(a, b, id, c) VALUES ('text', 2, 1, false); -- matches`,
			},
			{
				Statement: `INSERT INTO hjtest_1(a, b, id, c) VALUES ('text', 1, 2, false); -- fails id join condition`,
			},
			{
				Statement: `INSERT INTO hjtest_1(a, b, id, c) VALUES ('text', 20, 1, false); -- fails < 50`,
			},
			{
				Statement: `INSERT INTO hjtest_1(a, b, id, c) VALUES ('text', 1, 1, false); -- fails (SELECT hjtest_1.b * 5) = (SELECT hjtest_2.c*5)`,
			},
			{
				Statement: `INSERT INTO hjtest_2(a, id, b, c) VALUES (true, 1, 'another', 2); -- matches`,
			},
			{
				Statement: `INSERT INTO hjtest_2(a, id, b, c) VALUES (true, 3, 'another', 7); -- fails id join condition`,
			},
			{
				Statement: `INSERT INTO hjtest_2(a, id, b, c) VALUES (true, 1, 'another', 90);  -- fails < 55`,
			},
			{
				Statement: `INSERT INTO hjtest_2(a, id, b, c) VALUES (true, 1, 'another', 3); -- fails (SELECT hjtest_1.b * 5) = (SELECT hjtest_2.c*5)`,
			},
			{
				Statement: `INSERT INTO hjtest_2(a, id, b, c) VALUES (true, 1, 'text', 1); --  fails hjtest_1.a <> hjtest_2.b;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF, VERBOSE)
SELECT hjtest_1.a a1, hjtest_2.a a2,hjtest_1.tableoid::regclass t1, hjtest_2.tableoid::regclass t2
FROM hjtest_1, hjtest_2
WHERE
    hjtest_1.id = (SELECT 1 WHERE hjtest_2.id = 1)
    AND (SELECT hjtest_1.b * 5) = (SELECT hjtest_2.c*5)
    AND (SELECT hjtest_1.b * 5) < 50
    AND (SELECT hjtest_2.c * 5) < 55
    AND hjtest_1.a <> hjtest_2.b;`,
				Results: []sql.Row{{`Hash Join`}, {`Output: hjtest_1.a, hjtest_2.a, (hjtest_1.tableoid)::regclass, (hjtest_2.tableoid)::regclass`}, {`Hash Cond: ((hjtest_1.id = (SubPlan 1)) AND ((SubPlan 2) = (SubPlan 3)))`}, {`Join Filter: (hjtest_1.a <> hjtest_2.b)`}, {`->  Seq Scan on public.hjtest_1`}, {`Output: hjtest_1.a, hjtest_1.tableoid, hjtest_1.id, hjtest_1.b`}, {`Filter: ((SubPlan 4) < 50)`}, {`SubPlan 4`}, {`->  Result`}, {`Output: (hjtest_1.b * 5)`}, {`->  Hash`}, {`Output: hjtest_2.a, hjtest_2.tableoid, hjtest_2.id, hjtest_2.c, hjtest_2.b`}, {`->  Seq Scan on public.hjtest_2`}, {`Output: hjtest_2.a, hjtest_2.tableoid, hjtest_2.id, hjtest_2.c, hjtest_2.b`}, {`Filter: ((SubPlan 5) < 55)`}, {`SubPlan 5`}, {`->  Result`}, {`Output: (hjtest_2.c * 5)`}, {`SubPlan 1`}, {`->  Result`}, {`Output: 1`}, {`One-Time Filter: (hjtest_2.id = 1)`}, {`SubPlan 3`}, {`->  Result`}, {`Output: (hjtest_2.c * 5)`}, {`SubPlan 2`}, {`->  Result`}, {`Output: (hjtest_1.b * 5)`}},
			},
			{
				Statement: `SELECT hjtest_1.a a1, hjtest_2.a a2,hjtest_1.tableoid::regclass t1, hjtest_2.tableoid::regclass t2
FROM hjtest_1, hjtest_2
WHERE
    hjtest_1.id = (SELECT 1 WHERE hjtest_2.id = 1)
    AND (SELECT hjtest_1.b * 5) = (SELECT hjtest_2.c*5)
    AND (SELECT hjtest_1.b * 5) < 50
    AND (SELECT hjtest_2.c * 5) < 55
    AND hjtest_1.a <> hjtest_2.b;`,
				Results: []sql.Row{{`text`, true, `hjtest_1`, `hjtest_2`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF, VERBOSE)
SELECT hjtest_1.a a1, hjtest_2.a a2,hjtest_1.tableoid::regclass t1, hjtest_2.tableoid::regclass t2
FROM hjtest_2, hjtest_1
WHERE
    hjtest_1.id = (SELECT 1 WHERE hjtest_2.id = 1)
    AND (SELECT hjtest_1.b * 5) = (SELECT hjtest_2.c*5)
    AND (SELECT hjtest_1.b * 5) < 50
    AND (SELECT hjtest_2.c * 5) < 55
    AND hjtest_1.a <> hjtest_2.b;`,
				Results: []sql.Row{{`Hash Join`}, {`Output: hjtest_1.a, hjtest_2.a, (hjtest_1.tableoid)::regclass, (hjtest_2.tableoid)::regclass`}, {`Hash Cond: (((SubPlan 1) = hjtest_1.id) AND ((SubPlan 3) = (SubPlan 2)))`}, {`Join Filter: (hjtest_1.a <> hjtest_2.b)`}, {`->  Seq Scan on public.hjtest_2`}, {`Output: hjtest_2.a, hjtest_2.tableoid, hjtest_2.id, hjtest_2.c, hjtest_2.b`}, {`Filter: ((SubPlan 5) < 55)`}, {`SubPlan 5`}, {`->  Result`}, {`Output: (hjtest_2.c * 5)`}, {`->  Hash`}, {`Output: hjtest_1.a, hjtest_1.tableoid, hjtest_1.id, hjtest_1.b`}, {`->  Seq Scan on public.hjtest_1`}, {`Output: hjtest_1.a, hjtest_1.tableoid, hjtest_1.id, hjtest_1.b`}, {`Filter: ((SubPlan 4) < 50)`}, {`SubPlan 4`}, {`->  Result`}, {`Output: (hjtest_1.b * 5)`}, {`SubPlan 2`}, {`->  Result`}, {`Output: (hjtest_1.b * 5)`}, {`SubPlan 1`}, {`->  Result`}, {`Output: 1`}, {`One-Time Filter: (hjtest_2.id = 1)`}, {`SubPlan 3`}, {`->  Result`}, {`Output: (hjtest_2.c * 5)`}},
			},
			{
				Statement: `SELECT hjtest_1.a a1, hjtest_2.a a2,hjtest_1.tableoid::regclass t1, hjtest_2.tableoid::regclass t2
FROM hjtest_2, hjtest_1
WHERE
    hjtest_1.id = (SELECT 1 WHERE hjtest_2.id = 1)
    AND (SELECT hjtest_1.b * 5) = (SELECT hjtest_2.c*5)
    AND (SELECT hjtest_1.b * 5) < 50
    AND (SELECT hjtest_2.c * 5) < 55
    AND hjtest_1.a <> hjtest_2.b;`,
				Results: []sql.Row{{`text`, true, `hjtest_1`, `hjtest_2`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `set local enable_hashjoin = on;`,
			},
			{
				Statement: `explain (costs off)
select i8.q2, ss.* from
int8_tbl i8,
lateral (select t1.fivethous, i4.f1 from tenk1 t1 join int4_tbl i4
         on t1.fivethous = i4.f1+i8.q2 order by 1,2) ss;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Seq Scan on int8_tbl i8`}, {`->  Sort`}, {`Sort Key: t1.fivethous, i4.f1`}, {`->  Hash Join`}, {`Hash Cond: (t1.fivethous = (i4.f1 + i8.q2))`}, {`->  Seq Scan on tenk1 t1`}, {`->  Hash`}, {`->  Seq Scan on int4_tbl i4`}},
			},
			{
				Statement: `select i8.q2, ss.* from
int8_tbl i8,
lateral (select t1.fivethous, i4.f1 from tenk1 t1 join int4_tbl i4
         on t1.fivethous = i4.f1+i8.q2 order by 1,2) ss;`,
				Results: []sql.Row{{456, 456, 0}, {456, 456, 0}, {123, 123, 0}, {123, 123, 0}},
			},
			{
				Statement: `rollback;`,
			},
		},
	})
}
