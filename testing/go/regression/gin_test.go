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

func TestGin(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_gin)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_gin,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create table gin_test_tbl(i int4[]) with (autovacuum_enabled = off);`,
			},
			{
				Statement: `create index gin_test_idx on gin_test_tbl using gin (i)
  with (fastupdate = on, gin_pending_list_limit = 4096);`,
			},
			{
				Statement: `insert into gin_test_tbl select array[1, 2, g] from generate_series(1, 20000) g;`,
			},
			{
				Statement: `insert into gin_test_tbl select array[1, 3, g] from generate_series(1, 1000) g;`,
			},
			{
				Statement: `select gin_clean_pending_list('gin_test_idx')>10 as many; -- flush the fastupdate buffers`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `insert into gin_test_tbl select array[3, 1, g] from generate_series(1, 1000) g;`,
			},
			{
				Statement: `vacuum gin_test_tbl; -- flush the fastupdate buffers`,
			},
			{
				Statement: `select gin_clean_pending_list('gin_test_idx'); -- nothing to flush`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `delete from gin_test_tbl where i @> array[2];`,
			},
			{
				Statement: `vacuum gin_test_tbl;`,
			},
			{
				Statement: `alter index gin_test_idx set (fastupdate = off);`,
			},
			{
				Statement: `insert into gin_test_tbl select array[1, 2, g] from generate_series(1, 1000) g;`,
			},
			{
				Statement: `insert into gin_test_tbl select array[1, 3, g] from generate_series(1, 1000) g;`,
			},
			{
				Statement: `delete from gin_test_tbl where i @> array[2];`,
			},
			{
				Statement: `vacuum gin_test_tbl;`,
			},
			{
				Statement: `explain (costs off)
select count(*) from gin_test_tbl where i @> array[1, 999];`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on gin_test_tbl`}, {`Recheck Cond: (i @> '{1,999}'::integer[])`}, {`->  Bitmap Index Scan on gin_test_idx`}, {`Index Cond: (i @> '{1,999}'::integer[])`}},
			},
			{
				Statement: `select count(*) from gin_test_tbl where i @> array[1, 999];`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `set gin_fuzzy_search_limit = 1000;`,
			},
			{
				Statement: `explain (costs off)
select count(*) > 0 as ok from gin_test_tbl where i @> array[1];`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on gin_test_tbl`}, {`Recheck Cond: (i @> '{1}'::integer[])`}, {`->  Bitmap Index Scan on gin_test_idx`}, {`Index Cond: (i @> '{1}'::integer[])`}},
			},
			{
				Statement: `select count(*) > 0 as ok from gin_test_tbl where i @> array[1];`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `reset gin_fuzzy_search_limit;`,
			},
			{
				Statement: `create temp table t_gin_test_tbl(i int4[], j int4[]);`,
			},
			{
				Statement: `create index on t_gin_test_tbl using gin (i, j);`,
			},
			{
				Statement: `insert into t_gin_test_tbl
values
  (null,    null),
  ('{}',    null),
  ('{1}',   null),
  ('{1,2}', null),
  (null,    '{}'),
  (null,    '{10}'),
  ('{1,2}', '{10}'),
  ('{2}',   '{10}'),
  ('{1,3}', '{}'),
  ('{1,1}', '{10}');`,
			},
			{
				Statement: `set enable_seqscan = off;`,
			},
			{
				Statement: `explain (costs off)
select * from t_gin_test_tbl where array[0] <@ i;`,
				Results: []sql.Row{{`Bitmap Heap Scan on t_gin_test_tbl`}, {`Recheck Cond: ('{0}'::integer[] <@ i)`}, {`->  Bitmap Index Scan on t_gin_test_tbl_i_j_idx`}, {`Index Cond: (i @> '{0}'::integer[])`}},
			},
			{
				Statement: `select * from t_gin_test_tbl where array[0] <@ i;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from t_gin_test_tbl where array[0] <@ i and '{}'::int4[] <@ j;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `explain (costs off)
select * from t_gin_test_tbl where i @> '{}';`,
				Results: []sql.Row{{`Bitmap Heap Scan on t_gin_test_tbl`}, {`Recheck Cond: (i @> '{}'::integer[])`}, {`->  Bitmap Index Scan on t_gin_test_tbl_i_j_idx`}, {`Index Cond: (i @> '{}'::integer[])`}},
			},
			{
				Statement: `select * from t_gin_test_tbl where i @> '{}';`,
				Results:   []sql.Row{{`{}`, ``}, {`{1}`, ``}, {`{1,2}`, ``}, {`{1,2}`, `{10}`}, {`{2}`, `{10}`}, {`{1,3}`, `{}`}, {`{1,1}`, `{10}`}},
			},
			{
				Statement: `create function explain_query_json(query_sql text)
returns table (explain_line json)
language plpgsql as
$$
begin
  set enable_seqscan = off;`,
			},
			{
				Statement: `  set enable_bitmapscan = on;`,
			},
			{
				Statement: `  return query execute 'EXPLAIN (ANALYZE, FORMAT json) ' || query_sql;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create function execute_text_query_index(query_sql text)
returns setof text
language plpgsql
as
$$
begin
  set enable_seqscan = off;`,
			},
			{
				Statement: `  set enable_bitmapscan = on;`,
			},
			{
				Statement: `  return query execute query_sql;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create function execute_text_query_heap(query_sql text)
returns setof text
language plpgsql
as
$$
begin
  set enable_seqscan = on;`,
			},
			{
				Statement: `  set enable_bitmapscan = off;`,
			},
			{
				Statement: `  return query execute query_sql;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `select
  query,
  js->0->'Plan'->'Plans'->0->'Actual Rows' as "return by index",
  js->0->'Plan'->'Rows Removed by Index Recheck' as "removed by recheck",
  (res_index = res_heap) as "match"
from
  (values
    ($$ i @> '{}' $$),
    ($$ j @> '{}' $$),
    ($$ i @> '{}' and j @> '{}' $$),
    ($$ i @> '{1}' $$),
    ($$ i @> '{1}' and j @> '{}' $$),
    ($$ i @> '{1}' and i @> '{}' and j @> '{}' $$),
    ($$ j @> '{10}' $$),
    ($$ j @> '{10}' and i @> '{}' $$),
    ($$ j @> '{10}' and j @> '{}' and i @> '{}' $$),
    ($$ i @> '{1}' and j @> '{10}' $$)
  ) q(query),
  lateral explain_query_json($$select * from t_gin_test_tbl where $$ || query) js,
  lateral execute_text_query_index($$select string_agg((i, j)::text, ' ') from t_gin_test_tbl where $$ || query) res_index,
  lateral execute_text_query_heap($$select string_agg((i, j)::text, ' ') from t_gin_test_tbl where $$ || query) res_heap;`,
				Results: []sql.Row{{`i @> '{}'`, 7, 0, true}, {`j @> '{}'`, 6, 0, true}, {`i @> '{}' and j @> '{}'`, 4, 0, true}, {`i @> '{1}'`, 5, 0, true}, {`i @> '{1}' and j @> '{}'`, 3, 0, true}, {`i @> '{1}' and i @> '{}' and j @> '{}'`, 3, 0, true}, {`j @> '{10}'`, 4, 0, true}, {`j @> '{10}' and i @> '{}'`, 3, 0, true}, {`j @> '{10}' and j @> '{}' and i @> '{}'`, 3, 0, true}, {`i @> '{1}' and j @> '{10}'`, 2, 0, true}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `insert into t_gin_test_tbl select array[1, g, g/10], array[2, g, g/10]
  from generate_series(1, 20000) g;`,
			},
			{
				Statement: `select gin_clean_pending_list('t_gin_test_tbl_i_j_idx') is not null;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `analyze t_gin_test_tbl;`,
			},
			{
				Statement: `set enable_seqscan = off;`,
			},
			{
				Statement: `set enable_bitmapscan = on;`,
			},
			{
				Statement: `explain (costs off)
select count(*) from t_gin_test_tbl where j @> array[50];`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on t_gin_test_tbl`}, {`Recheck Cond: (j @> '{50}'::integer[])`}, {`->  Bitmap Index Scan on t_gin_test_tbl_i_j_idx`}, {`Index Cond: (j @> '{50}'::integer[])`}},
			},
			{
				Statement: `select count(*) from t_gin_test_tbl where j @> array[50];`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `explain (costs off)
select count(*) from t_gin_test_tbl where j @> array[2];`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on t_gin_test_tbl`}, {`Recheck Cond: (j @> '{2}'::integer[])`}, {`->  Bitmap Index Scan on t_gin_test_tbl_i_j_idx`}, {`Index Cond: (j @> '{2}'::integer[])`}},
			},
			{
				Statement: `select count(*) from t_gin_test_tbl where j @> array[2];`,
				Results:   []sql.Row{{20000}},
			},
			{
				Statement: `explain (costs off)
select count(*) from t_gin_test_tbl where j @> '{}'::int[];`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on t_gin_test_tbl`}, {`Recheck Cond: (j @> '{}'::integer[])`}, {`->  Bitmap Index Scan on t_gin_test_tbl_i_j_idx`}, {`Index Cond: (j @> '{}'::integer[])`}},
			},
			{
				Statement: `select count(*) from t_gin_test_tbl where j @> '{}'::int[];`,
				Results:   []sql.Row{{20006}},
			},
			{
				Statement: `delete from t_gin_test_tbl where j @> array[2];`,
			},
			{
				Statement: `vacuum t_gin_test_tbl;`,
			},
			{
				Statement: `select count(*) from t_gin_test_tbl where j @> array[50];`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from t_gin_test_tbl where j @> array[2];`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select count(*) from t_gin_test_tbl where j @> '{}'::int[];`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `drop table t_gin_test_tbl;`,
			},
			{
				Statement: `create unlogged table t_gin_test_tbl(i int4[], j int4[]);`,
			},
			{
				Statement: `create index on t_gin_test_tbl using gin (i, j);`,
			},
			{
				Statement: `insert into t_gin_test_tbl
values
  (null,    null),
  ('{}',    null),
  ('{1}',   '{2,3}');`,
			},
			{
				Statement: `drop table t_gin_test_tbl;`,
			},
		},
	})
}
