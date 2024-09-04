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

func TestTemp(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_temp)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_temp,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE temptest(col int);`,
			},
			{
				Statement: `CREATE INDEX i_temptest ON temptest(col);`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest(tcol int);`,
			},
			{
				Statement: `CREATE INDEX i_temptest ON temptest(tcol);`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP INDEX i_temptest;`,
			},
			{
				Statement: `DROP TABLE temptest;`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP INDEX i_temptest;`,
			},
			{
				Statement: `DROP TABLE temptest;`,
			},
			{
				Statement: `CREATE TABLE temptest(col int);`,
			},
			{
				Statement: `INSERT INTO temptest VALUES (1);`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest(tcol float);`,
			},
			{
				Statement: `INSERT INTO temptest VALUES (2.1);`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{{2.1}},
			},
			{
				Statement: `DROP TABLE temptest;`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP TABLE temptest;`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest(col int);`,
			},
			{
				Statement: `\c
SELECT * FROM temptest;`,
				ErrorString: `relation "temptest" does not exist`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest(col int) ON COMMIT DELETE ROWS;`,
			},
			{
				Statement: `CREATE INDEX ON temptest(bit_length(''));`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO temptest VALUES (1);`,
			},
			{
				Statement: `INSERT INTO temptest VALUES (2);`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE temptest;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest(col) ON COMMIT DELETE ROWS AS SELECT 1;`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE temptest;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest(col int) ON COMMIT DROP;`,
			},
			{
				Statement: `INSERT INTO temptest VALUES (1);`,
			},
			{
				Statement: `INSERT INTO temptest VALUES (2);`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement:   `SELECT * FROM temptest;`,
				ErrorString: `relation "temptest" does not exist`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest(col) ON COMMIT DROP AS SELECT 1;`,
			},
			{
				Statement: `SELECT * FROM temptest;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement:   `SELECT * FROM temptest;`,
				ErrorString: `relation "temptest" does not exist`,
			},
			{
				Statement:   `CREATE TABLE temptest(col int) ON COMMIT DELETE ROWS;`,
				ErrorString: `ON COMMIT can only be used on temporary tables`,
			},
			{
				Statement:   `CREATE TABLE temptest(col) ON COMMIT DELETE ROWS AS SELECT 1;`,
				ErrorString: `ON COMMIT can only be used on temporary tables`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest1(col int PRIMARY KEY);`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest2(col int REFERENCES temptest1)
  ON COMMIT DELETE ROWS;`,
			},
			{
				Statement: `INSERT INTO temptest1 VALUES (1);`,
			},
			{
				Statement: `INSERT INTO temptest2 VALUES (1);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM temptest1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM temptest2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest3(col int PRIMARY KEY) ON COMMIT DELETE ROWS;`,
			},
			{
				Statement: `CREATE TEMP TABLE temptest4(col int REFERENCES temptest3);`,
			},
			{
				Statement:   `COMMIT;`,
				ErrorString: `unsupported ON COMMIT and foreign key combination`,
			},
			{
				Statement: `create table public.whereami (f1 text);`,
			},
			{
				Statement: `insert into public.whereami values ('public');`,
			},
			{
				Statement: `create temp table whereami (f1 text);`,
			},
			{
				Statement: `insert into whereami values ('temp');`,
			},
			{
				Statement: `create function public.whoami() returns text
  as $$select 'public'::text$$ language sql;`,
			},
			{
				Statement: `create function pg_temp.whoami() returns text
  as $$select 'temp'::text$$ language sql;`,
			},
			{
				Statement: `select * from whereami;`,
				Results:   []sql.Row{{`temp`}},
			},
			{
				Statement: `select whoami();`,
				Results:   []sql.Row{{`public`}},
			},
			{
				Statement: `set search_path = pg_temp, public;`,
			},
			{
				Statement: `select * from whereami;`,
				Results:   []sql.Row{{`temp`}},
			},
			{
				Statement: `select whoami();`,
				Results:   []sql.Row{{`public`}},
			},
			{
				Statement: `set search_path = public, pg_temp;`,
			},
			{
				Statement: `select * from whereami;`,
				Results:   []sql.Row{{`public`}},
			},
			{
				Statement: `select whoami();`,
				Results:   []sql.Row{{`public`}},
			},
			{
				Statement: `select pg_temp.whoami();`,
				Results:   []sql.Row{{`temp`}},
			},
			{
				Statement: `drop table public.whereami;`,
			},
			{
				Statement: `set search_path = pg_temp, public;`,
			},
			{
				Statement: `create domain pg_temp.nonempty as text check (value <> '');`,
			},
			{
				Statement:   `select nonempty('');`,
				ErrorString: `function nonempty(unknown) does not exist`,
			},
			{
				Statement:   `select pg_temp.nonempty('');`,
				ErrorString: `value for domain nonempty violates check constraint "nonempty_check"`,
			},
			{
				Statement:   `select ''::nonempty;`,
				ErrorString: `value for domain nonempty violates check constraint "nonempty_check"`,
			},
			{
				Statement: `reset search_path;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create temp table temp_parted_oncommit (a int)
  partition by list (a) on commit delete rows;`,
			},
			{
				Statement: `create temp table temp_parted_oncommit_1
  partition of temp_parted_oncommit
  for values in (1) on commit delete rows;`,
			},
			{
				Statement: `insert into temp_parted_oncommit values (1);`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `select * from temp_parted_oncommit;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `drop table temp_parted_oncommit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create temp table temp_parted_oncommit_test (a int)
  partition by list (a) on commit drop;`,
			},
			{
				Statement: `create temp table temp_parted_oncommit_test1
  partition of temp_parted_oncommit_test
  for values in (1) on commit delete rows;`,
			},
			{
				Statement: `create temp table temp_parted_oncommit_test2
  partition of temp_parted_oncommit_test
  for values in (2) on commit drop;`,
			},
			{
				Statement: `insert into temp_parted_oncommit_test values (1), (2);`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `select relname from pg_class where relname ~ '^temp_parted_oncommit_test';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create temp table temp_parted_oncommit_test (a int)
  partition by list (a) on commit delete rows;`,
			},
			{
				Statement: `create temp table temp_parted_oncommit_test1
  partition of temp_parted_oncommit_test
  for values in (1) on commit preserve rows;`,
			},
			{
				Statement: `create temp table temp_parted_oncommit_test2
  partition of temp_parted_oncommit_test
  for values in (2) on commit drop;`,
			},
			{
				Statement: `insert into temp_parted_oncommit_test values (1), (2);`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `select * from temp_parted_oncommit_test;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select relname from pg_class where relname ~ '^temp_parted_oncommit_test'
  order by relname;`,
				Results: []sql.Row{{`temp_parted_oncommit_test`}, {`temp_parted_oncommit_test1`}},
			},
			{
				Statement: `drop table temp_parted_oncommit_test;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create temp table temp_inh_oncommit_test (a int) on commit drop;`,
			},
			{
				Statement: `create temp table temp_inh_oncommit_test1 ()
  inherits(temp_inh_oncommit_test) on commit delete rows;`,
			},
			{
				Statement: `insert into temp_inh_oncommit_test1 values (1);`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `select relname from pg_class where relname ~ '^temp_inh_oncommit_test';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create temp table temp_inh_oncommit_test (a int) on commit delete rows;`,
			},
			{
				Statement: `create temp table temp_inh_oncommit_test1 ()
  inherits(temp_inh_oncommit_test) on commit drop;`,
			},
			{
				Statement: `insert into temp_inh_oncommit_test1 values (1);`,
			},
			{
				Statement: `insert into temp_inh_oncommit_test values (1);`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `select * from temp_inh_oncommit_test;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select relname from pg_class where relname ~ '^temp_inh_oncommit_test';`,
				Results:   []sql.Row{{`temp_inh_oncommit_test`}},
			},
			{
				Statement: `drop table temp_inh_oncommit_test;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create function pg_temp.twophase_func() returns void as
  $$ select '2pc_func'::text $$ language sql;`,
			},
			{
				Statement:   `prepare transaction 'twophase_func';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `create function pg_temp.twophase_func() returns void as
  $$ select '2pc_func'::text $$ language sql;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `drop function pg_temp.twophase_func();`,
			},
			{
				Statement:   `prepare transaction 'twophase_func';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create operator pg_temp.@@ (leftarg = int4, rightarg = int4, procedure = int4mi);`,
			},
			{
				Statement:   `prepare transaction 'twophase_operator';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create type pg_temp.twophase_type as (a int);`,
			},
			{
				Statement:   `prepare transaction 'twophase_type';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create view pg_temp.twophase_view as select 1;`,
			},
			{
				Statement:   `prepare transaction 'twophase_view';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create sequence pg_temp.twophase_seq;`,
			},
			{
				Statement:   `prepare transaction 'twophase_sequence';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `create temp table twophase_tab (a int);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `select a from twophase_tab;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `prepare transaction 'twophase_tab';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `insert into twophase_tab values (1);`,
			},
			{
				Statement:   `prepare transaction 'twophase_tab';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `lock twophase_tab in access exclusive mode;`,
			},
			{
				Statement:   `prepare transaction 'twophase_tab';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `drop table twophase_tab;`,
			},
			{
				Statement:   `prepare transaction 'twophase_tab';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
			{
				Statement: `\c -
SET search_path TO 'pg_temp';`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT current_schema() ~ 'pg_temp' AS is_temp_schema;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `PREPARE TRANSACTION 'twophase_search';`,
				ErrorString: `cannot PREPARE a transaction that has operated on temporary objects`,
			},
		},
	})
}
