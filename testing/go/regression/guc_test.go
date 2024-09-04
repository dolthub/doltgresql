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

func TestGuc(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_guc)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_guc,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`Postgres, MDY`}},
			},
			{
				Statement: `SET vacuum_cost_delay TO 40;`,
			},
			{
				Statement: `SET datestyle = 'ISO, YMD';`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `SET LOCAL vacuum_cost_delay TO 50;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SET LOCAL datestyle = 'SQL';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL vacuum_cost_delay TO 50;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`50ms`}},
			},
			{
				Statement: `SET LOCAL datestyle = 'SQL';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`SQL, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`08/13/2006 12:34:56 PDT`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET vacuum_cost_delay TO 60;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`60ms`}},
			},
			{
				Statement: `SET datestyle = 'German';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`German, DMY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`13.08.2006 12:34:56 PDT`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET vacuum_cost_delay TO 70;`,
			},
			{
				Statement: `SET datestyle = 'MDY';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `SAVEPOINT first_sp;`,
			},
			{
				Statement: `SET vacuum_cost_delay TO 80.1;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`80100us`}},
			},
			{
				Statement: `SET datestyle = 'German, DMY';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`German, DMY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`13.08.2006 12:34:56 PDT`}},
			},
			{
				Statement: `ROLLBACK TO first_sp;`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `SAVEPOINT second_sp;`,
			},
			{
				Statement: `SET vacuum_cost_delay TO '900us';`,
			},
			{
				Statement: `SET datestyle = 'SQL, YMD';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`SQL, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`08/13/2006 12:34:56 PDT`}},
			},
			{
				Statement: `SAVEPOINT third_sp;`,
			},
			{
				Statement: `SET vacuum_cost_delay TO 100;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`100ms`}},
			},
			{
				Statement: `SET datestyle = 'Postgres, MDY';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`Postgres, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`Sun Aug 13 12:34:56 2006 PDT`}},
			},
			{
				Statement: `ROLLBACK TO third_sp;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`900us`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`SQL, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`08/13/2006 12:34:56 PDT`}},
			},
			{
				Statement: `ROLLBACK TO second_sp;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`70ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `SAVEPOINT sp;`,
			},
			{
				Statement: `SET LOCAL vacuum_cost_delay TO 30;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`30ms`}},
			},
			{
				Statement: `SET LOCAL datestyle = 'Postgres, MDY';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`Postgres, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`Sun Aug 13 12:34:56 2006 PDT`}},
			},
			{
				Statement: `ROLLBACK TO sp;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `SAVEPOINT sp;`,
			},
			{
				Statement: `SET LOCAL vacuum_cost_delay TO 30;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`30ms`}},
			},
			{
				Statement: `SET LOCAL datestyle = 'Postgres, MDY';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`Postgres, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`Sun Aug 13 12:34:56 2006 PDT`}},
			},
			{
				Statement: `RELEASE SAVEPOINT sp;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`30ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`Postgres, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`Sun Aug 13 12:34:56 2006 PDT`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET vacuum_cost_delay TO 40;`,
			},
			{
				Statement: `SET LOCAL vacuum_cost_delay TO 50;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`50ms`}},
			},
			{
				Statement: `SET datestyle = 'ISO, DMY';`,
			},
			{
				Statement: `SET LOCAL datestyle = 'Postgres, MDY';`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`Postgres, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`Sun Aug 13 12:34:56 2006 PDT`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`40ms`}},
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, DMY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `SET datestyle = iso, ymd;`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`ISO, YMD`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`2006-08-13 12:34:56-07`}},
			},
			{
				Statement: `RESET datestyle;`,
			},
			{
				Statement: `SHOW datestyle;`,
				Results:   []sql.Row{{`Postgres, MDY`}},
			},
			{
				Statement: `SELECT '2006-08-13 12:34:56'::timestamptz;`,
				Results:   []sql.Row{{`Sun Aug 13 12:34:56 2006 PDT`}},
			},
			{
				Statement:   `SET seq_page_cost TO 'NaN';`,
				ErrorString: `invalid value for parameter "seq_page_cost": "NaN"`,
			},
			{
				Statement:   `SET vacuum_cost_delay TO '10s';`,
				ErrorString: `10000 ms is outside the valid range for parameter "vacuum_cost_delay" (0 .. 100)`,
			},
			{
				Statement:   `SET no_such_variable TO 42;`,
				ErrorString: `unrecognized configuration parameter "no_such_variable"`,
			},
			{
				Statement:   `SHOW custom.my_guc;  -- error, not known yet`,
				ErrorString: `unrecognized configuration parameter "custom.my_guc"`,
			},
			{
				Statement: `SET custom.my_guc = 42;`,
			},
			{
				Statement: `SHOW custom.my_guc;`,
				Results:   []sql.Row{{42}},
			},
			{
				Statement: `RESET custom.my_guc;  -- this makes it go to empty, not become unknown again`,
			},
			{
				Statement: `SHOW custom.my_guc;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SET custom.my.qualified.guc = 'foo';`,
			},
			{
				Statement: `SHOW custom.my.qualified.guc;`,
				Results:   []sql.Row{{`foo`}},
			},
			{
				Statement:   `SET custom."bad-guc" = 42;  -- disallowed because -c cannot set this name`,
				ErrorString: `invalid configuration parameter name "custom.bad-guc"`,
			},
			{
				Statement:   `SHOW custom."bad-guc";`,
				ErrorString: `unrecognized configuration parameter "custom.bad-guc"`,
			},
			{
				Statement:   `SET special."weird name" = 'foo';  -- could be allowed, but we choose not to`,
				ErrorString: `invalid configuration parameter name "special.weird name"`,
			},
			{
				Statement:   `SHOW special."weird name";`,
				ErrorString: `unrecognized configuration parameter "special.weird name"`,
			},
			{
				Statement: `SET plpgsql.extra_foo_warnings = true;  -- allowed if plpgsql is not loaded yet`,
			},
			{
				Statement: `LOAD 'plpgsql';  -- this will throw a warning and delete the variable`,
			},
			{
				Statement:   `SET plpgsql.extra_foo_warnings = true;  -- now, it's an error`,
				ErrorString: `invalid configuration parameter name "plpgsql.extra_foo_warnings"`,
			},
			{
				Statement:   `SHOW plpgsql.extra_foo_warnings;`,
				ErrorString: `unrecognized configuration parameter "plpgsql.extra_foo_warnings"`,
			},
			{
				Statement: `CREATE TEMP TABLE reset_test ( data text ) ON COMMIT DELETE ROWS;`,
			},
			{
				Statement: `SELECT relname FROM pg_class WHERE relname = 'reset_test';`,
				Results:   []sql.Row{{`reset_test`}},
			},
			{
				Statement: `DISCARD TEMP;`,
			},
			{
				Statement: `SELECT relname FROM pg_class WHERE relname = 'reset_test';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DECLARE foo CURSOR WITH HOLD FOR SELECT 1;`,
			},
			{
				Statement: `PREPARE foo AS SELECT 1;`,
			},
			{
				Statement: `LISTEN foo_event;`,
			},
			{
				Statement: `SET vacuum_cost_delay = 13;`,
			},
			{
				Statement: `CREATE TEMP TABLE tmp_foo (data text) ON COMMIT DELETE ROWS;`,
			},
			{
				Statement: `CREATE ROLE regress_guc_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_guc_user;`,
			},
			{
				Statement: `SELECT pg_listening_channels();`,
				Results:   []sql.Row{{`foo_event`}},
			},
			{
				Statement: `SELECT name FROM pg_prepared_statements;`,
				Results:   []sql.Row{{`foo`}},
			},
			{
				Statement: `SELECT name FROM pg_cursors;`,
				Results:   []sql.Row{{`foo`}},
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{`13ms`}},
			},
			{
				Statement: `SELECT relname from pg_class where relname = 'tmp_foo';`,
				Results:   []sql.Row{{`tmp_foo`}},
			},
			{
				Statement: `SELECT current_user = 'regress_guc_user';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DISCARD ALL;`,
			},
			{
				Statement: `SELECT pg_listening_channels();`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT name FROM pg_prepared_statements;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT name FROM pg_cursors;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SHOW vacuum_cost_delay;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT relname from pg_class where relname = 'tmp_foo';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT current_user = 'regress_guc_user';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `DROP ROLE regress_guc_user;`,
			},
			{
				Statement: `set search_path = foo, public, not_there_initially;`,
			},
			{
				Statement: `select current_schemas(false);`,
				Results:   []sql.Row{{`{public}`}},
			},
			{
				Statement: `create schema not_there_initially;`,
			},
			{
				Statement: `select current_schemas(false);`,
				Results:   []sql.Row{{`{public,not_there_initially}`}},
			},
			{
				Statement: `drop schema not_there_initially;`,
			},
			{
				Statement: `select current_schemas(false);`,
				Results:   []sql.Row{{`{public}`}},
			},
			{
				Statement: `reset search_path;`,
			},
			{
				Statement: `set work_mem = '3MB';`,
			},
			{
				Statement: `create function report_guc(text) returns text as
$$ select current_setting($1) $$ language sql
set work_mem = '1MB';`,
			},
			{
				Statement: `select report_guc('work_mem'), current_setting('work_mem');`,
				Results:   []sql.Row{{`1MB`, `3MB`}},
			},
			{
				Statement: `alter function report_guc(text) set work_mem = '2MB';`,
			},
			{
				Statement: `select report_guc('work_mem'), current_setting('work_mem');`,
				Results:   []sql.Row{{`2MB`, `3MB`}},
			},
			{
				Statement: `alter function report_guc(text) reset all;`,
			},
			{
				Statement: `select report_guc('work_mem'), current_setting('work_mem');`,
				Results:   []sql.Row{{`3MB`, `3MB`}},
			},
			{
				Statement: `create or replace function myfunc(int) returns text as $$
begin
  set local work_mem = '2MB';`,
			},
			{
				Statement: `  return current_setting('work_mem');`,
			},
			{
				Statement: `end $$
language plpgsql
set work_mem = '1MB';`,
			},
			{
				Statement: `select myfunc(0), current_setting('work_mem');`,
				Results:   []sql.Row{{`2MB`, `3MB`}},
			},
			{
				Statement: `alter function myfunc(int) reset all;`,
			},
			{
				Statement: `select myfunc(0), current_setting('work_mem');`,
				Results:   []sql.Row{{`2MB`, `2MB`}},
			},
			{
				Statement: `set work_mem = '3MB';`,
			},
			{
				Statement: `create or replace function myfunc(int) returns text as $$
begin
  set work_mem = '2MB';`,
			},
			{
				Statement: `  return current_setting('work_mem');`,
			},
			{
				Statement: `end $$
language plpgsql
set work_mem = '1MB';`,
			},
			{
				Statement: `select myfunc(0), current_setting('work_mem');`,
				Results:   []sql.Row{{`2MB`, `2MB`}},
			},
			{
				Statement: `set work_mem = '3MB';`,
			},
			{
				Statement: `create or replace function myfunc(int) returns text as $$
begin
  set work_mem = '2MB';`,
			},
			{
				Statement: `  perform 1/$1;`,
			},
			{
				Statement: `  return current_setting('work_mem');`,
			},
			{
				Statement: `end $$
language plpgsql
set work_mem = '1MB';`,
			},
			{
				Statement:   `select myfunc(0);`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `CONTEXT:  SQL statement "SELECT 1/$1"
PL/pgSQL function myfunc(integer) line 4 at PERFORM
select current_setting('work_mem');`,
				Results: []sql.Row{{`3MB`}},
			},
			{
				Statement: `select myfunc(1), current_setting('work_mem');`,
				Results:   []sql.Row{{`2MB`, `2MB`}},
			},
			{
				Statement:   `select current_setting('nosuch.setting');  -- FAIL`,
				ErrorString: `unrecognized configuration parameter "nosuch.setting"`,
			},
			{
				Statement:   `select current_setting('nosuch.setting', false);  -- FAIL`,
				ErrorString: `unrecognized configuration parameter "nosuch.setting"`,
			},
			{
				Statement: `select current_setting('nosuch.setting', true) is null;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `set nosuch.setting = 'nada';`,
			},
			{
				Statement: `select current_setting('nosuch.setting');`,
				Results:   []sql.Row{{`nada`}},
			},
			{
				Statement: `select current_setting('nosuch.setting', false);`,
				Results:   []sql.Row{{`nada`}},
			},
			{
				Statement: `select current_setting('nosuch.setting', true);`,
				Results:   []sql.Row{{`nada`}},
			},
			{
				Statement: `create function func_with_bad_set() returns int as $$ select 1 $$
language sql
set default_text_search_config = no_such_config;`,
				ErrorString: `invalid value for parameter "default_text_search_config": "no_such_config"`,
			},
			{
				Statement: `set check_function_bodies = off;`,
			},
			{
				Statement: `create function func_with_bad_set() returns int as $$ select 1 $$
language sql
set default_text_search_config = no_such_config;`,
			},
			{
				Statement:   `select func_with_bad_set();`,
				ErrorString: `invalid value for parameter "default_text_search_config": "no_such_config"`,
			},
			{
				Statement: `reset check_function_bodies;`,
			},
			{
				Statement: `set default_with_oids to f;`,
			},
			{
				Statement:   `set default_with_oids to t;`,
				ErrorString: `tables declared WITH OIDS are not supported`,
			},
			{
				Statement: `SELECT pg_settings_get_flags(NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_settings_get_flags('does_not_exist');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `CREATE TABLE tab_settings_flags AS SELECT name, category,
    'EXPLAIN'          = ANY(flags) AS explain,
    'NO_RESET_ALL'     = ANY(flags) AS no_reset_all,
    'NOT_IN_SAMPLE'    = ANY(flags) AS not_in_sample,
    'RUNTIME_COMPUTED' = ANY(flags) AS runtime_computed
  FROM pg_show_all_settings() AS psas,
    pg_settings_get_flags(psas.name) AS flags;`,
			},
			{
				Statement: `SELECT name FROM tab_settings_flags
  WHERE category = 'Developer Options' AND NOT not_in_sample
  ORDER BY 1;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT name FROM tab_settings_flags
  WHERE category ~ '^Query Tuning' AND NOT explain
  ORDER BY 1;`,
				Results: []sql.Row{{`default_statistics_target`}},
			},
			{
				Statement: `SELECT name FROM tab_settings_flags
  WHERE NOT category = 'Preset Options' AND runtime_computed
  ORDER BY 1;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT name FROM tab_settings_flags
  WHERE category = 'Preset Options' AND NOT not_in_sample
  ORDER BY 1;`,
				Results: []sql.Row{},
			},
			{
				Statement: `DROP TABLE tab_settings_flags;`,
			},
		},
	})
}
