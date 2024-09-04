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

func TestMiscFunctions(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_misc_functions)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_misc_functions,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set regresslib :libdir '/regress' :dlsuffix
SELECT num_nonnulls(NULL);`,
				Results: []sql.Row{{0}},
			},
			{
				Statement: `SELECT num_nonnulls('1');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT num_nonnulls(NULL::text);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT num_nonnulls(NULL::text, NULL::int);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT num_nonnulls(1, 2, NULL::text, NULL::point, '', int8 '9', 1.0 / NULL);`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `SELECT num_nonnulls(VARIADIC '{1,2,NULL,3}'::int[]);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT num_nonnulls(VARIADIC '{"1","2","3","4"}'::text[]);`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `SELECT num_nonnulls(VARIADIC ARRAY(SELECT CASE WHEN i <> 40 THEN i END FROM generate_series(1, 100) i));`,
				Results:   []sql.Row{{99}},
			},
			{
				Statement: `SELECT num_nulls(NULL);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT num_nulls('1');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT num_nulls(NULL::text);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT num_nulls(NULL::text, NULL::int);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT num_nulls(1, 2, NULL::text, NULL::point, '', int8 '9', 1.0 / NULL);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT num_nulls(VARIADIC '{1,2,NULL,3}'::int[]);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT num_nulls(VARIADIC '{"1","2","3","4"}'::text[]);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT num_nulls(VARIADIC ARRAY(SELECT CASE WHEN i <> 40 THEN i END FROM generate_series(1, 100) i));`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT num_nonnulls(VARIADIC NULL::text[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT num_nonnulls(VARIADIC '{}'::int[]);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT num_nulls(VARIADIC NULL::text[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT num_nulls(VARIADIC '{}'::int[]);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT num_nonnulls();`,
				ErrorString: `function num_nonnulls() does not exist`,
			},
			{
				Statement:   `SELECT num_nulls();`,
				ErrorString: `function num_nulls() does not exist`,
			},
			{
				Statement: `CREATE FUNCTION test_canonicalize_path(text)
   RETURNS text
   AS :'regresslib'
   LANGUAGE C STRICT IMMUTABLE;`,
			},
			{
				Statement: `SELECT test_canonicalize_path('/');`,
				Results:   []sql.Row{{`/`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('/./abc/def/');`,
				Results:   []sql.Row{{`/abc/def`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('/./../abc/def');`,
				Results:   []sql.Row{{`/abc/def`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('/./../../abc/def/');`,
				Results:   []sql.Row{{`/abc/def`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('/abc/.././def/ghi');`,
				Results:   []sql.Row{{`/def/ghi`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('/abc/./../def/ghi//');`,
				Results:   []sql.Row{{`/def/ghi`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('/abc/def/../..');`,
				Results:   []sql.Row{{`/`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('/abc/def/../../..');`,
				Results:   []sql.Row{{`/`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('/abc/def/../../../../ghi/jkl');`,
				Results:   []sql.Row{{`/ghi/jkl`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('.');`,
				Results:   []sql.Row{{`.`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('./');`,
				Results:   []sql.Row{{`.`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('./abc/..');`,
				Results:   []sql.Row{{`.`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('abc/../');`,
				Results:   []sql.Row{{`.`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('abc/../def');`,
				Results:   []sql.Row{{`def`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('..');`,
				Results:   []sql.Row{{`..`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('../abc/def');`,
				Results:   []sql.Row{{`../abc/def`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('../abc/..');`,
				Results:   []sql.Row{{`..`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('../abc/../def');`,
				Results:   []sql.Row{{`../def`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('../abc/../../def/ghi');`,
				Results:   []sql.Row{{`../../def/ghi`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('./abc/./def/.');`,
				Results:   []sql.Row{{`abc/def`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('./abc/././def/.');`,
				Results:   []sql.Row{{`abc/def`}},
			},
			{
				Statement: `SELECT test_canonicalize_path('./abc/./def/.././ghi/../../../jkl/mno');`,
				Results:   []sql.Row{{`../jkl/mno`}},
			},
			{
				Statement: `SELECT pg_log_backend_memory_contexts(pg_backend_pid());`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT pg_log_backend_memory_contexts(pid) FROM pg_stat_activity
  WHERE backend_type = 'checkpointer';`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `CREATE ROLE regress_log_memory;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_log_memory',
  'pg_log_backend_memory_contexts(integer)', 'EXECUTE'); -- no`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `GRANT EXECUTE ON FUNCTION pg_log_backend_memory_contexts(integer)
  TO regress_log_memory;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_log_memory',
  'pg_log_backend_memory_contexts(integer)', 'EXECUTE'); -- yes`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SET ROLE regress_log_memory;`,
			},
			{
				Statement: `SELECT pg_log_backend_memory_contexts(pg_backend_pid());`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `REVOKE EXECUTE ON FUNCTION pg_log_backend_memory_contexts(integer)
  FROM regress_log_memory;`,
			},
			{
				Statement: `DROP ROLE regress_log_memory;`,
			},
			{
				Statement: `select setting as segsize
from pg_settings where name = 'wal_segment_size'
\gset
select count(*) > 0 as ok from pg_ls_waldir();`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select count(*) > 0 as ok from (select pg_ls_waldir()) ss;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select * from pg_ls_waldir() limit 0;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select count(*) > 0 as ok from (select * from pg_ls_waldir() limit 1) ss;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select (w).size = :segsize as ok
from (select pg_ls_waldir() w) ss where length((w).name) = 24 limit 1;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select count(*) >= 0 as ok from pg_ls_archive_statusdir();`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select * from (select pg_ls_dir('.') a) a where a = 'base' limit 1;`,
				Results:   []sql.Row{{`base`}},
			},
			{
				Statement:   `select pg_ls_dir('does not exist', false, false); -- error`,
				ErrorString: `could not open directory "does not exist": No such file or directory`,
			},
			{
				Statement: `select pg_ls_dir('does not exist', true, false); -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select count(*) = 1 as dot_found
  from pg_ls_dir('.', false, true) as ls where ls = '.';`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select count(*) = 1 as dot_found
  from pg_ls_dir('.', false, false) as ls where ls = '.';`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select * from (select (pg_timezone_names()).name) ptn where name='UTC' limit 1;`,
				Results:   []sql.Row{{`UTC`}},
			},
			{
				Statement: `select count(*) > 0 from
  (select pg_tablespace_databases(oid) as pts from pg_tablespace
   where spcname = 'pg_default') pts
  join pg_database db on pts.pts = db.oid;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `CREATE ROLE regress_slot_dir_funcs;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_slot_dir_funcs',
  'pg_ls_logicalsnapdir()', 'EXECUTE');`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_slot_dir_funcs',
  'pg_ls_logicalmapdir()', 'EXECUTE');`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_slot_dir_funcs',
  'pg_ls_replslotdir(text)', 'EXECUTE');`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `GRANT pg_monitor TO regress_slot_dir_funcs;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_slot_dir_funcs',
  'pg_ls_logicalsnapdir()', 'EXECUTE');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_slot_dir_funcs',
  'pg_ls_logicalmapdir()', 'EXECUTE');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_slot_dir_funcs',
  'pg_ls_replslotdir(text)', 'EXECUTE');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `DROP ROLE regress_slot_dir_funcs;`,
			},
			{
				Statement: `CREATE FUNCTION my_int_eq(int, int) RETURNS bool
  LANGUAGE internal STRICT IMMUTABLE PARALLEL SAFE
  AS $$int4eq$$;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM tenk1 a JOIN tenk1 b ON a.unique1 = b.unique1
WHERE my_int_eq(a.unique2, 42);`,
				Results: []sql.Row{{`Hash Join`}, {`Hash Cond: (b.unique1 = a.unique1)`}, {`->  Seq Scan on tenk1 b`}, {`->  Hash`}, {`->  Seq Scan on tenk1 a`}, {`Filter: my_int_eq(unique2, 42)`}},
			},
			{
				Statement: `CREATE FUNCTION test_support_func(internal)
    RETURNS internal
    AS :'regresslib', 'test_support_func'
    LANGUAGE C STRICT;`,
			},
			{
				Statement: `ALTER FUNCTION my_int_eq(int, int) SUPPORT test_support_func;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM tenk1 a JOIN tenk1 b ON a.unique1 = b.unique1
WHERE my_int_eq(a.unique2, 42);`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Seq Scan on tenk1 a`}, {`Filter: my_int_eq(unique2, 42)`}, {`->  Index Scan using tenk1_unique1 on tenk1 b`}, {`Index Cond: (unique1 = a.unique1)`}},
			},
			{
				Statement: `CREATE FUNCTION my_gen_series(int, int) RETURNS SETOF integer
  LANGUAGE internal STRICT IMMUTABLE PARALLEL SAFE
  AS $$generate_series_int4$$
  SUPPORT test_support_func;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM tenk1 a JOIN my_gen_series(1,1000) g ON a.unique1 = g;`,
				Results: []sql.Row{{`Hash Join`}, {`Hash Cond: (g.g = a.unique1)`}, {`->  Function Scan on my_gen_series g`}, {`->  Hash`}, {`->  Seq Scan on tenk1 a`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM tenk1 a JOIN my_gen_series(1,10) g ON a.unique1 = g;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Function Scan on my_gen_series g`}, {`->  Index Scan using tenk1_unique1 on tenk1 a`}, {`Index Cond: (unique1 = g.g)`}},
			},
		},
	})
}
