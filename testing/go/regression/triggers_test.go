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

func TestTriggers(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_triggers)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_triggers,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set autoinclib :libdir '/autoinc' :dlsuffix
\set refintlib :libdir '/refint' :dlsuffix
\set regresslib :libdir '/regress' :dlsuffix
CREATE FUNCTION autoinc ()
	RETURNS trigger
	AS :'autoinclib'
	LANGUAGE C;`,
			},
			{
				Statement: `CREATE FUNCTION check_primary_key ()
	RETURNS trigger
	AS :'refintlib'
	LANGUAGE C;`,
			},
			{
				Statement: `CREATE FUNCTION check_foreign_key ()
	RETURNS trigger
	AS :'refintlib'
	LANGUAGE C;`,
			},
			{
				Statement: `CREATE FUNCTION trigger_return_old ()
        RETURNS trigger
        AS :'regresslib'
        LANGUAGE C;`,
			},
			{
				Statement: `CREATE FUNCTION set_ttdummy (int4)
        RETURNS int4
        AS :'regresslib'
        LANGUAGE C STRICT;`,
			},
			{
				Statement: `create table pkeys (pkey1 int4 not null, pkey2 text not null);`,
			},
			{
				Statement: `create table fkeys (fkey1 int4, fkey2 text, fkey3 int);`,
			},
			{
				Statement: `create table fkeys2 (fkey21 int4, fkey22 text, pkey23 int not null);`,
			},
			{
				Statement: `create index fkeys_i on fkeys (fkey1, fkey2);`,
			},
			{
				Statement: `create index fkeys2_i on fkeys2 (fkey21, fkey22);`,
			},
			{
				Statement: `create index fkeys2p_i on fkeys2 (pkey23);`,
			},
			{
				Statement: `insert into pkeys values (10, '1');`,
			},
			{
				Statement: `insert into pkeys values (20, '2');`,
			},
			{
				Statement: `insert into pkeys values (30, '3');`,
			},
			{
				Statement: `insert into pkeys values (40, '4');`,
			},
			{
				Statement: `insert into pkeys values (50, '5');`,
			},
			{
				Statement: `insert into pkeys values (60, '6');`,
			},
			{
				Statement: `create unique index pkeys_i on pkeys (pkey1, pkey2);`,
			},
			{
				Statement: `create trigger check_fkeys_pkey_exist
	before insert or update on fkeys
	for each row
	execute function
	check_primary_key ('fkey1', 'fkey2', 'pkeys', 'pkey1', 'pkey2');`,
			},
			{
				Statement: `create trigger check_fkeys_pkey2_exist
	before insert or update on fkeys
	for each row
	execute function check_primary_key ('fkey3', 'fkeys2', 'pkey23');`,
			},
			{
				Statement: `create trigger check_fkeys2_pkey_exist
	before insert or update on fkeys2
	for each row
	execute procedure
	check_primary_key ('fkey21', 'fkey22', 'pkeys', 'pkey1', 'pkey2');`,
			},
			{
				Statement:   `COMMENT ON TRIGGER check_fkeys2_pkey_bad ON fkeys2 IS 'wrong';`,
				ErrorString: `trigger "check_fkeys2_pkey_bad" for table "fkeys2" does not exist`,
			},
			{
				Statement: `COMMENT ON TRIGGER check_fkeys2_pkey_exist ON fkeys2 IS 'right';`,
			},
			{
				Statement: `COMMENT ON TRIGGER check_fkeys2_pkey_exist ON fkeys2 IS NULL;`,
			},
			{
				Statement: `create trigger check_pkeys_fkey_cascade
	before delete or update on pkeys
	for each row
	execute procedure
	check_foreign_key (2, 'cascade', 'pkey1', 'pkey2',
	'fkeys', 'fkey1', 'fkey2', 'fkeys2', 'fkey21', 'fkey22');`,
			},
			{
				Statement: `create trigger check_fkeys2_fkey_restrict
	before delete or update on fkeys2
	for each row
	execute procedure check_foreign_key (1, 'restrict', 'pkey23', 'fkeys', 'fkey3');`,
			},
			{
				Statement: `insert into fkeys2 values (10, '1', 1);`,
			},
			{
				Statement: `insert into fkeys2 values (30, '3', 2);`,
			},
			{
				Statement: `insert into fkeys2 values (40, '4', 5);`,
			},
			{
				Statement: `insert into fkeys2 values (50, '5', 3);`,
			},
			{
				Statement:   `insert into fkeys2 values (70, '5', 3);`,
				ErrorString: `tuple references non-existent key`,
			},
			{
				Statement: `insert into fkeys values (10, '1', 2);`,
			},
			{
				Statement: `insert into fkeys values (30, '3', 3);`,
			},
			{
				Statement: `insert into fkeys values (40, '4', 2);`,
			},
			{
				Statement: `insert into fkeys values (50, '5', 2);`,
			},
			{
				Statement:   `insert into fkeys values (70, '5', 1);`,
				ErrorString: `tuple references non-existent key`,
			},
			{
				Statement:   `insert into fkeys values (60, '6', 4);`,
				ErrorString: `tuple references non-existent key`,
			},
			{
				Statement:   `delete from pkeys where pkey1 = 30 and pkey2 = '3';`,
				ErrorString: `"check_fkeys2_fkey_restrict": tuple is referenced in "fkeys"`,
			},
			{
				Statement: `CONTEXT:  SQL statement "delete from fkeys2 where fkey21 = $1 and fkey22 = $2 "
delete from pkeys where pkey1 = 40 and pkey2 = '4';`,
			},
			{
				Statement:   `update pkeys set pkey1 = 7, pkey2 = '70' where pkey1 = 50 and pkey2 = '5';`,
				ErrorString: `"check_fkeys2_fkey_restrict": tuple is referenced in "fkeys"`,
			},
			{
				Statement: `CONTEXT:  SQL statement "delete from fkeys2 where fkey21 = $1 and fkey22 = $2 "
update pkeys set pkey1 = 7, pkey2 = '70' where pkey1 = 10 and pkey2 = '1';`,
			},
			{
				Statement: `SELECT trigger_name, event_manipulation, event_object_schema, event_object_table,
       action_order, action_condition, action_orientation, action_timing,
       action_reference_old_table, action_reference_new_table
  FROM information_schema.triggers
  WHERE event_object_table in ('pkeys', 'fkeys', 'fkeys2')
  ORDER BY trigger_name COLLATE "C", 2;`,
				Results: []sql.Row{{`check_fkeys2_fkey_restrict`, `DELETE`, `public`, `fkeys2`, 1, ``, `ROW`, `BEFORE`, ``, ``}, {`check_fkeys2_fkey_restrict`, `UPDATE`, `public`, `fkeys2`, 1, ``, `ROW`, `BEFORE`, ``, ``}, {`check_fkeys2_pkey_exist`, `INSERT`, `public`, `fkeys2`, 1, ``, `ROW`, `BEFORE`, ``, ``}, {`check_fkeys2_pkey_exist`, `UPDATE`, `public`, `fkeys2`, 2, ``, `ROW`, `BEFORE`, ``, ``}, {`check_fkeys_pkey2_exist`, `INSERT`, `public`, `fkeys`, 1, ``, `ROW`, `BEFORE`, ``, ``}, {`check_fkeys_pkey2_exist`, `UPDATE`, `public`, `fkeys`, 1, ``, `ROW`, `BEFORE`, ``, ``}, {`check_fkeys_pkey_exist`, `INSERT`, `public`, `fkeys`, 2, ``, `ROW`, `BEFORE`, ``, ``}, {`check_fkeys_pkey_exist`, `UPDATE`, `public`, `fkeys`, 2, ``, `ROW`, `BEFORE`, ``, ``}, {`check_pkeys_fkey_cascade`, `DELETE`, `public`, `pkeys`, 1, ``, `ROW`, `BEFORE`, ``, ``}, {`check_pkeys_fkey_cascade`, `UPDATE`, `public`, `pkeys`, 1, ``, `ROW`, `BEFORE`, ``, ``}},
			},
			{
				Statement: `DROP TABLE pkeys;`,
			},
			{
				Statement: `DROP TABLE fkeys;`,
			},
			{
				Statement: `DROP TABLE fkeys2;`,
			},
			{
				Statement: `create table trigtest (f1 int, f2 text);`,
			},
			{
				Statement: `create trigger trigger_return_old
	before insert or delete or update on trigtest
	for each row execute procedure trigger_return_old();`,
			},
			{
				Statement: `insert into trigtest values(1, 'foo');`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{1, `foo`}},
			},
			{
				Statement: `update trigtest set f2 = f2 || 'bar';`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{1, `foo`}},
			},
			{
				Statement: `delete from trigtest;`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `create function f1_times_10() returns trigger as
$$ begin new.f1 := new.f1 * 10; return new; end $$ language plpgsql;`,
			},
			{
				Statement: `create trigger trigger_alpha
	before insert or update on trigtest
	for each row execute procedure f1_times_10();`,
			},
			{
				Statement: `insert into trigtest values(1, 'foo');`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{10, `foo`}},
			},
			{
				Statement: `update trigtest set f2 = f2 || 'bar';`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{10, `foo`}},
			},
			{
				Statement: `delete from trigtest;`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `create trigger trigger_zed
	before insert or update on trigtest
	for each row execute procedure f1_times_10();`,
			},
			{
				Statement: `insert into trigtest values(1, 'foo');`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{100, `foo`}},
			},
			{
				Statement: `update trigtest set f2 = f2 || 'bar';`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{1000, `foo`}},
			},
			{
				Statement: `delete from trigtest;`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `drop trigger trigger_alpha on trigtest;`,
			},
			{
				Statement: `insert into trigtest values(1, 'foo');`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{10, `foo`}},
			},
			{
				Statement: `update trigtest set f2 = f2 || 'bar';`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{100, `foo`}},
			},
			{
				Statement: `delete from trigtest;`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `drop table trigtest;`,
			},
			{
				Statement: `create table trigtest (
  a integer,
  b bool default true not null,
  c text default 'xyzzy' not null);`,
			},
			{
				Statement: `create trigger trigger_return_old
	before insert or delete or update on trigtest
	for each row execute procedure trigger_return_old();`,
			},
			{
				Statement: `insert into trigtest values(1);`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{1, true, `xyzzy`}},
			},
			{
				Statement: `alter table trigtest add column d integer default 42 not null;`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{1, true, `xyzzy`, 42}},
			},
			{
				Statement: `update trigtest set a = 2 where a = 1 returning *;`,
				Results:   []sql.Row{{1, true, `xyzzy`, 42}},
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{1, true, `xyzzy`, 42}},
			},
			{
				Statement: `alter table trigtest drop column b;`,
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{1, `xyzzy`, 42}},
			},
			{
				Statement: `update trigtest set a = 2 where a = 1 returning *;`,
				Results:   []sql.Row{{1, `xyzzy`, 42}},
			},
			{
				Statement: `select * from trigtest;`,
				Results:   []sql.Row{{1, `xyzzy`, 42}},
			},
			{
				Statement: `drop table trigtest;`,
			},
			{
				Statement: `create sequence ttdummy_seq increment 10 start 0 minvalue 0;`,
			},
			{
				Statement: `create table tttest (
	price_id	int4,
	price_val	int4,
	price_on	int4,
	price_off	int4 default 999999
);`,
			},
			{
				Statement: `create trigger ttdummy
	before delete or update on tttest
	for each row
	execute procedure
	ttdummy (price_on, price_off);`,
			},
			{
				Statement: `create trigger ttserial
	before insert or update on tttest
	for each row
	execute procedure
	autoinc (price_on, ttdummy_seq);`,
			},
			{
				Statement: `insert into tttest values (1, 1, null);`,
			},
			{
				Statement: `insert into tttest values (2, 2, null);`,
			},
			{
				Statement: `insert into tttest values (3, 3, 0);`,
			},
			{
				Statement: `select * from tttest;`,
				Results:   []sql.Row{{1, 1, 10, 999999}, {2, 2, 20, 999999}, {3, 3, 30, 999999}},
			},
			{
				Statement: `delete from tttest where price_id = 2;`,
			},
			{
				Statement: `select * from tttest;`,
				Results:   []sql.Row{{1, 1, 10, 999999}, {3, 3, 30, 999999}, {2, 2, 20, 40}},
			},
			{
				Statement: `select * from tttest where price_off = 999999;`,
				Results:   []sql.Row{{1, 1, 10, 999999}, {3, 3, 30, 999999}},
			},
			{
				Statement: `update tttest set price_val = 30 where price_id = 3;`,
			},
			{
				Statement: `select * from tttest;`,
				Results:   []sql.Row{{1, 1, 10, 999999}, {2, 2, 20, 40}, {3, 30, 50, 999999}, {3, 3, 30, 50}},
			},
			{
				Statement: `update tttest set price_id = 5 where price_id = 3;`,
			},
			{
				Statement: `select * from tttest;`,
				Results:   []sql.Row{{1, 1, 10, 999999}, {2, 2, 20, 40}, {3, 3, 30, 50}, {5, 30, 60, 999999}, {3, 30, 50, 60}},
			},
			{
				Statement: `select set_ttdummy(0);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `delete from tttest where price_id = 5;`,
			},
			{
				Statement: `update tttest set price_off = 999999 where price_val = 30;`,
			},
			{
				Statement: `select * from tttest;`,
				Results:   []sql.Row{{1, 1, 10, 999999}, {2, 2, 20, 40}, {3, 3, 30, 50}, {3, 30, 50, 999999}},
			},
			{
				Statement: `update tttest set price_id = 5 where price_id = 3;`,
			},
			{
				Statement: `select * from tttest;`,
				Results:   []sql.Row{{1, 1, 10, 999999}, {2, 2, 20, 40}, {5, 3, 30, 50}, {5, 30, 50, 999999}},
			},
			{
				Statement: `select set_ttdummy(1);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `update tttest set price_on = -1 where price_id = 1;`,
				ErrorString: `ttdummy (tttest): you cannot change price_on and/or price_off columns (use set_ttdummy)`,
			},
			{
				Statement: `select set_ttdummy(0);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `update tttest set price_on = -1 where price_id = 1;`,
			},
			{
				Statement: `select * from tttest;`,
				Results:   []sql.Row{{2, 2, 20, 40}, {5, 3, 30, 50}, {5, 30, 50, 999999}, {1, 1, -1, 999999}},
			},
			{
				Statement: `select * from tttest where price_on <= 35 and price_off > 35 and price_id = 5;`,
				Results:   []sql.Row{{5, 3, 30, 50}},
			},
			{
				Statement: `drop table tttest;`,
			},
			{
				Statement: `drop sequence ttdummy_seq;`,
			},
			{
				Statement: `CREATE TABLE log_table (tstamp timestamp default timeofday()::timestamp);`,
			},
			{
				Statement: `CREATE TABLE main_table (a int unique, b int);`,
			},
			{
				Statement: `COPY main_table (a,b) FROM stdin;`,
			},
			{
				Statement: `CREATE FUNCTION trigger_func() RETURNS trigger LANGUAGE plpgsql AS '
BEGIN
	RAISE NOTICE ''trigger_func(%) called: action = %, when = %, level = %'', TG_ARGV[0], TG_OP, TG_WHEN, TG_LEVEL;`,
			},
			{
				Statement: `	RETURN NULL;`,
			},
			{
				Statement: `END;';`,
			},
			{
				Statement: `CREATE TRIGGER before_ins_stmt_trig BEFORE INSERT ON main_table
FOR EACH STATEMENT EXECUTE PROCEDURE trigger_func('before_ins_stmt');`,
			},
			{
				Statement: `CREATE TRIGGER after_ins_stmt_trig AFTER INSERT ON main_table
FOR EACH STATEMENT EXECUTE PROCEDURE trigger_func('after_ins_stmt');`,
			},
			{
				Statement: `CREATE TRIGGER after_upd_stmt_trig AFTER UPDATE ON main_table
EXECUTE PROCEDURE trigger_func('after_upd_stmt');`,
			},
			{
				Statement: `INSERT INTO main_table (a, b) VALUES (5, 10) ON CONFLICT (a)
  DO UPDATE SET b = EXCLUDED.b;`,
			},
			{
				Statement: `CREATE TRIGGER after_upd_row_trig AFTER UPDATE ON main_table
FOR EACH ROW EXECUTE PROCEDURE trigger_func('after_upd_row');`,
			},
			{
				Statement: `INSERT INTO main_table DEFAULT VALUES;`,
			},
			{
				Statement: `UPDATE main_table SET a = a + 1 WHERE b < 30;`,
			},
			{
				Statement: `UPDATE main_table SET a = a + 2 WHERE b > 100;`,
			},
			{
				Statement: `ALTER TABLE main_table DROP CONSTRAINT main_table_a_key;`,
			},
			{
				Statement: `COPY main_table (a, b) FROM stdin;`,
			},
			{
				Statement: `SELECT * FROM main_table ORDER BY a, b;`,
				Results:   []sql.Row{{6, 10}, {21, 20}, {30, 40}, {31, 10}, {50, 35}, {50, 60}, {81, 15}, {``, ``}},
			},
			{
				Statement: `CREATE TRIGGER modified_a BEFORE UPDATE OF a ON main_table
FOR EACH ROW WHEN (OLD.a <> NEW.a) EXECUTE PROCEDURE trigger_func('modified_a');`,
			},
			{
				Statement: `CREATE TRIGGER modified_any BEFORE UPDATE OF a ON main_table
FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM NEW.*) EXECUTE PROCEDURE trigger_func('modified_any');`,
			},
			{
				Statement: `CREATE TRIGGER insert_a AFTER INSERT ON main_table
FOR EACH ROW WHEN (NEW.a = 123) EXECUTE PROCEDURE trigger_func('insert_a');`,
			},
			{
				Statement: `CREATE TRIGGER delete_a AFTER DELETE ON main_table
FOR EACH ROW WHEN (OLD.a = 123) EXECUTE PROCEDURE trigger_func('delete_a');`,
			},
			{
				Statement: `CREATE TRIGGER insert_when BEFORE INSERT ON main_table
FOR EACH STATEMENT WHEN (true) EXECUTE PROCEDURE trigger_func('insert_when');`,
			},
			{
				Statement: `CREATE TRIGGER delete_when AFTER DELETE ON main_table
FOR EACH STATEMENT WHEN (true) EXECUTE PROCEDURE trigger_func('delete_when');`,
			},
			{
				Statement: `SELECT trigger_name, event_manipulation, event_object_schema, event_object_table,
       action_order, action_condition, action_orientation, action_timing,
       action_reference_old_table, action_reference_new_table
  FROM information_schema.triggers
  WHERE event_object_table IN ('main_table')
  ORDER BY trigger_name COLLATE "C", 2;`,
				Results: []sql.Row{{`after_ins_stmt_trig`, `INSERT`, `public`, `main_table`, 1, ``, `STATEMENT`, `AFTER`, ``, ``}, {`after_upd_row_trig`, `UPDATE`, `public`, `main_table`, 1, ``, `ROW`, `AFTER`, ``, ``}, {`after_upd_stmt_trig`, `UPDATE`, `public`, `main_table`, 1, ``, `STATEMENT`, `AFTER`, ``, ``}, {`before_ins_stmt_trig`, `INSERT`, `public`, `main_table`, 1, ``, `STATEMENT`, `BEFORE`, ``, ``}, {`delete_a`, `DELETE`, `public`, `main_table`, 1, `(old.a = 123)`, `ROW`, `AFTER`, ``, ``}, {`delete_when`, `DELETE`, `public`, `main_table`, 1, `true`, `STATEMENT`, `AFTER`, ``, ``}, {`insert_a`, `INSERT`, `public`, `main_table`, 1, `(new.a = 123)`, `ROW`, `AFTER`, ``, ``}, {`insert_when`, `INSERT`, `public`, `main_table`, 2, `true`, `STATEMENT`, `BEFORE`, ``, ``}, {`modified_a`, `UPDATE`, `public`, `main_table`, 1, `(old.a <> new.a)`, `ROW`, `BEFORE`, ``, ``}, {`modified_any`, `UPDATE`, `public`, `main_table`, 2, `(old.* IS DISTINCT FROM new.*)`, `ROW`, `BEFORE`, ``, ``}},
			},
			{
				Statement: `INSERT INTO main_table (a) VALUES (123), (456);`,
			},
			{
				Statement: `COPY main_table FROM stdin;`,
			},
			{
				Statement: `DELETE FROM main_table WHERE a IN (123, 456);`,
			},
			{
				Statement: `UPDATE main_table SET a = 50, b = 60;`,
			},
			{
				Statement: `SELECT * FROM main_table ORDER BY a, b;`,
				Results:   []sql.Row{{6, 10}, {21, 20}, {30, 40}, {31, 10}, {50, 35}, {50, 60}, {81, 15}, {``, ``}},
			},
			{
				Statement: `SELECT pg_get_triggerdef(oid, true) FROM pg_trigger WHERE tgrelid = 'main_table'::regclass AND tgname = 'modified_a';`,
				Results:   []sql.Row{{`CREATE TRIGGER modified_a BEFORE UPDATE OF a ON main_table FOR EACH ROW WHEN (old.a <> new.a) EXECUTE FUNCTION trigger_func('modified_a')`}},
			},
			{
				Statement: `SELECT pg_get_triggerdef(oid, false) FROM pg_trigger WHERE tgrelid = 'main_table'::regclass AND tgname = 'modified_a';`,
				Results:   []sql.Row{{`CREATE TRIGGER modified_a BEFORE UPDATE OF a ON public.main_table FOR EACH ROW WHEN ((old.a <> new.a)) EXECUTE FUNCTION trigger_func('modified_a')`}},
			},
			{
				Statement: `SELECT pg_get_triggerdef(oid, true) FROM pg_trigger WHERE tgrelid = 'main_table'::regclass AND tgname = 'modified_any';`,
				Results:   []sql.Row{{`CREATE TRIGGER modified_any BEFORE UPDATE OF a ON main_table FOR EACH ROW WHEN (old.* IS DISTINCT FROM new.*) EXECUTE FUNCTION trigger_func('modified_any')`}},
			},
			{
				Statement: `ALTER TRIGGER modified_a ON main_table RENAME TO modified_modified_a;`,
			},
			{
				Statement: `SELECT count(*) FROM pg_trigger WHERE tgrelid = 'main_table'::regclass AND tgname = 'modified_a';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT count(*) FROM pg_trigger WHERE tgrelid = 'main_table'::regclass AND tgname = 'modified_modified_a';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP TRIGGER modified_modified_a ON main_table;`,
			},
			{
				Statement: `DROP TRIGGER modified_any ON main_table;`,
			},
			{
				Statement: `DROP TRIGGER insert_a ON main_table;`,
			},
			{
				Statement: `DROP TRIGGER delete_a ON main_table;`,
			},
			{
				Statement: `DROP TRIGGER insert_when ON main_table;`,
			},
			{
				Statement: `DROP TRIGGER delete_when ON main_table;`,
			},
			{
				Statement: `create table table_with_oids(a int);`,
			},
			{
				Statement: `insert into table_with_oids values (1);`,
			},
			{
				Statement: `create trigger oid_unchanged_trig after update on table_with_oids
	for each row
	when (new.tableoid = old.tableoid AND new.tableoid <> 0)
	execute procedure trigger_func('after_upd_oid_unchanged');`,
			},
			{
				Statement: `update table_with_oids set a = a + 1;`,
			},
			{
				Statement: `drop table table_with_oids;`,
			},
			{
				Statement: `DROP TRIGGER after_upd_row_trig ON main_table;`,
			},
			{
				Statement: `CREATE TRIGGER before_upd_a_row_trig BEFORE UPDATE OF a ON main_table
FOR EACH ROW EXECUTE PROCEDURE trigger_func('before_upd_a_row');`,
			},
			{
				Statement: `CREATE TRIGGER after_upd_b_row_trig AFTER UPDATE OF b ON main_table
FOR EACH ROW EXECUTE PROCEDURE trigger_func('after_upd_b_row');`,
			},
			{
				Statement: `CREATE TRIGGER after_upd_a_b_row_trig AFTER UPDATE OF a, b ON main_table
FOR EACH ROW EXECUTE PROCEDURE trigger_func('after_upd_a_b_row');`,
			},
			{
				Statement: `CREATE TRIGGER before_upd_a_stmt_trig BEFORE UPDATE OF a ON main_table
FOR EACH STATEMENT EXECUTE PROCEDURE trigger_func('before_upd_a_stmt');`,
			},
			{
				Statement: `CREATE TRIGGER after_upd_b_stmt_trig AFTER UPDATE OF b ON main_table
FOR EACH STATEMENT EXECUTE PROCEDURE trigger_func('after_upd_b_stmt');`,
			},
			{
				Statement: `SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE tgrelid = 'main_table'::regclass AND tgname = 'after_upd_a_b_row_trig';`,
				Results:   []sql.Row{{`CREATE TRIGGER after_upd_a_b_row_trig AFTER UPDATE OF a, b ON public.main_table FOR EACH ROW EXECUTE FUNCTION trigger_func('after_upd_a_b_row')`}},
			},
			{
				Statement: `UPDATE main_table SET a = 50;`,
			},
			{
				Statement: `UPDATE main_table SET b = 10;`,
			},
			{
				Statement: `CREATE TABLE some_t (some_col boolean NOT NULL);`,
			},
			{
				Statement: `CREATE FUNCTION dummy_update_func() RETURNS trigger AS $$
BEGIN
  RAISE NOTICE 'dummy_update_func(%) called: action = %, old = %, new = %',
    TG_ARGV[0], TG_OP, OLD, NEW;`,
			},
			{
				Statement: `  RETURN NEW;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE TRIGGER some_trig_before BEFORE UPDATE ON some_t FOR EACH ROW
  EXECUTE PROCEDURE dummy_update_func('before');`,
			},
			{
				Statement: `CREATE TRIGGER some_trig_aftera AFTER UPDATE ON some_t FOR EACH ROW
  WHEN (NOT OLD.some_col AND NEW.some_col)
  EXECUTE PROCEDURE dummy_update_func('aftera');`,
			},
			{
				Statement: `CREATE TRIGGER some_trig_afterb AFTER UPDATE ON some_t FOR EACH ROW
  WHEN (NOT NEW.some_col)
  EXECUTE PROCEDURE dummy_update_func('afterb');`,
			},
			{
				Statement: `INSERT INTO some_t VALUES (TRUE);`,
			},
			{
				Statement: `UPDATE some_t SET some_col = TRUE;`,
			},
			{
				Statement: `UPDATE some_t SET some_col = FALSE;`,
			},
			{
				Statement: `UPDATE some_t SET some_col = TRUE;`,
			},
			{
				Statement: `DROP TABLE some_t;`,
			},
			{
				Statement: `CREATE TRIGGER error_upd_and_col BEFORE UPDATE OR UPDATE OF a ON main_table
FOR EACH ROW EXECUTE PROCEDURE trigger_func('error_upd_and_col');`,
				ErrorString: `duplicate trigger events specified at or near "ON"`,
			},
			{
				Statement: `CREATE TRIGGER error_upd_a_a BEFORE UPDATE OF a, a ON main_table
FOR EACH ROW EXECUTE PROCEDURE trigger_func('error_upd_a_a');`,
				ErrorString: `column "a" specified more than once`,
			},
			{
				Statement: `CREATE TRIGGER error_ins_a BEFORE INSERT OF a ON main_table
FOR EACH ROW EXECUTE PROCEDURE trigger_func('error_ins_a');`,
				ErrorString: `syntax error at or near "OF"`,
			},
			{
				Statement: `CREATE TRIGGER error_ins_when BEFORE INSERT OR UPDATE ON main_table
FOR EACH ROW WHEN (OLD.a <> NEW.a)
EXECUTE PROCEDURE trigger_func('error_ins_old');`,
				ErrorString: `INSERT trigger's WHEN condition cannot reference OLD values`,
			},
			{
				Statement: `CREATE TRIGGER error_del_when BEFORE DELETE OR UPDATE ON main_table
FOR EACH ROW WHEN (OLD.a <> NEW.a)
EXECUTE PROCEDURE trigger_func('error_del_new');`,
				ErrorString: `DELETE trigger's WHEN condition cannot reference NEW values`,
			},
			{
				Statement: `CREATE TRIGGER error_del_when BEFORE INSERT OR UPDATE ON main_table
FOR EACH ROW WHEN (NEW.tableoid <> 0)
EXECUTE PROCEDURE trigger_func('error_when_sys_column');`,
				ErrorString: `BEFORE trigger's WHEN condition cannot reference NEW system columns`,
			},
			{
				Statement: `CREATE TRIGGER error_stmt_when BEFORE UPDATE OF a ON main_table
FOR EACH STATEMENT WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE trigger_func('error_stmt_when');`,
				ErrorString: `statement trigger's WHEN condition cannot reference column values`,
			},
			{
				Statement:   `ALTER TABLE main_table DROP COLUMN b;`,
				ErrorString: `cannot drop column b of table main_table because other objects depend on it`,
			},
			{
				Statement: `trigger after_upd_a_b_row_trig on table main_table depends on column b of table main_table
trigger after_upd_b_stmt_trig on table main_table depends on column b of table main_table
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
begin;`,
			},
			{
				Statement: `DROP TRIGGER after_upd_a_b_row_trig ON main_table;`,
			},
			{
				Statement: `DROP TRIGGER after_upd_b_row_trig ON main_table;`,
			},
			{
				Statement: `DROP TRIGGER after_upd_b_stmt_trig ON main_table;`,
			},
			{
				Statement: `ALTER TABLE main_table DROP COLUMN b;`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `create table trigtest (i serial primary key);`,
			},
			{
				Statement: `create table trigtest2 (i int references trigtest(i) on delete cascade);`,
			},
			{
				Statement: `create function trigtest() returns trigger as $$
begin
	raise notice '% % % %', TG_TABLE_NAME, TG_OP, TG_WHEN, TG_LEVEL;`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `end;$$ language plpgsql;`,
			},
			{
				Statement: `create trigger trigtest_b_row_tg before insert or update or delete on trigtest
for each row execute procedure trigtest();`,
			},
			{
				Statement: `create trigger trigtest_a_row_tg after insert or update or delete on trigtest
for each row execute procedure trigtest();`,
			},
			{
				Statement: `create trigger trigtest_b_stmt_tg before insert or update or delete on trigtest
for each statement execute procedure trigtest();`,
			},
			{
				Statement: `create trigger trigtest_a_stmt_tg after insert or update or delete on trigtest
for each statement execute procedure trigtest();`,
			},
			{
				Statement: `insert into trigtest default values;`,
			},
			{
				Statement: `alter table trigtest disable trigger trigtest_b_row_tg;`,
			},
			{
				Statement: `insert into trigtest default values;`,
			},
			{
				Statement: `alter table trigtest disable trigger user;`,
			},
			{
				Statement: `insert into trigtest default values;`,
			},
			{
				Statement: `alter table trigtest enable trigger trigtest_a_stmt_tg;`,
			},
			{
				Statement: `insert into trigtest default values;`,
			},
			{
				Statement: `set session_replication_role = replica;`,
			},
			{
				Statement: `insert into trigtest default values;  -- does not trigger`,
			},
			{
				Statement: `alter table trigtest enable always trigger trigtest_a_stmt_tg;`,
			},
			{
				Statement: `insert into trigtest default values;  -- now it does`,
			},
			{
				Statement: `reset session_replication_role;`,
			},
			{
				Statement: `insert into trigtest2 values(1);`,
			},
			{
				Statement: `insert into trigtest2 values(2);`,
			},
			{
				Statement: `delete from trigtest where i=2;`,
			},
			{
				Statement: `select * from trigtest2;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `alter table trigtest disable trigger all;`,
			},
			{
				Statement: `delete from trigtest where i=1;`,
			},
			{
				Statement: `select * from trigtest2;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `insert into trigtest default values;`,
			},
			{
				Statement: `select *  from trigtest;`,
				Results:   []sql.Row{{3}, {4}, {5}, {6}, {7}},
			},
			{
				Statement: `drop table trigtest2;`,
			},
			{
				Statement: `drop table trigtest;`,
			},
			{
				Statement: `CREATE TABLE trigger_test (
        i int,
        v varchar
);`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION trigger_data()  RETURNS trigger
LANGUAGE plpgsql AS $$
declare
	argstr text;`,
			},
			{
				Statement: `	relid text;`,
			},
			{
				Statement: `begin
	relid := TG_relid::regclass;`,
			},
			{
				Statement: `	-- plpgsql can't discover its trigger data in a hash like perl and python
	-- can, or by a sort of reflection like tcl can,
	-- so we have to hard code the names.
	raise NOTICE 'TG_NAME: %', TG_name;`,
			},
			{
				Statement: `	raise NOTICE 'TG_WHEN: %', TG_when;`,
			},
			{
				Statement: `	raise NOTICE 'TG_LEVEL: %', TG_level;`,
			},
			{
				Statement: `	raise NOTICE 'TG_OP: %', TG_op;`,
			},
			{
				Statement: `	raise NOTICE 'TG_RELID::regclass: %', relid;`,
			},
			{
				Statement: `	raise NOTICE 'TG_RELNAME: %', TG_relname;`,
			},
			{
				Statement: `	raise NOTICE 'TG_TABLE_NAME: %', TG_table_name;`,
			},
			{
				Statement: `	raise NOTICE 'TG_TABLE_SCHEMA: %', TG_table_schema;`,
			},
			{
				Statement: `	raise NOTICE 'TG_NARGS: %', TG_nargs;`,
			},
			{
				Statement: `	argstr := '[';`,
			},
			{
				Statement: `	for i in 0 .. TG_nargs - 1 loop
		if i > 0 then
			argstr := argstr || ', ';`,
			},
			{
				Statement: `		end if;`,
			},
			{
				Statement: `		argstr := argstr || TG_argv[i];`,
			},
			{
				Statement: `	end loop;`,
			},
			{
				Statement: `	argstr := argstr || ']';`,
			},
			{
				Statement: `	raise NOTICE 'TG_ARGV: %', argstr;`,
			},
			{
				Statement: `	if TG_OP != 'INSERT' then
		raise NOTICE 'OLD: %', OLD;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if TG_OP != 'DELETE' then
		raise NOTICE 'NEW: %', NEW;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	if TG_OP = 'DELETE' then
		return OLD;`,
			},
			{
				Statement: `	else
		return NEW;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER show_trigger_data_trig
BEFORE INSERT OR UPDATE OR DELETE ON trigger_test
FOR EACH ROW EXECUTE PROCEDURE trigger_data(23,'skidoo');`,
			},
			{
				Statement: `insert into trigger_test values(1,'insert');`,
			},
			{
				Statement: `update trigger_test set v = 'update' where i = 1;`,
			},
			{
				Statement: `delete from trigger_test;`,
			},
			{
				Statement: `DROP TRIGGER show_trigger_data_trig on trigger_test;`,
			},
			{
				Statement: `DROP FUNCTION trigger_data();`,
			},
			{
				Statement: `DROP TABLE trigger_test;`,
			},
			{
				Statement: `CREATE TABLE trigger_test (f1 int, f2 text, f3 text);`,
			},
			{
				Statement: `CREATE FUNCTION mytrigger() RETURNS trigger LANGUAGE plpgsql as $$
begin
	if row(old.*) = row(new.*) then
		raise notice 'row % not changed', new.f1;`,
			},
			{
				Statement: `	else
		raise notice 'row % changed', new.f1;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `end$$;`,
			},
			{
				Statement: `CREATE TRIGGER t
BEFORE UPDATE ON trigger_test
FOR EACH ROW EXECUTE PROCEDURE mytrigger();`,
			},
			{
				Statement: `INSERT INTO trigger_test VALUES(1, 'foo', 'bar');`,
			},
			{
				Statement: `INSERT INTO trigger_test VALUES(2, 'baz', 'quux');`,
			},
			{
				Statement: `UPDATE trigger_test SET f3 = 'bar';`,
			},
			{
				Statement: `UPDATE trigger_test SET f3 = NULL;`,
			},
			{
				Statement: `UPDATE trigger_test SET f3 = NULL;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION mytrigger() RETURNS trigger LANGUAGE plpgsql as $$
begin
	if row(old.*) is distinct from row(new.*) then
		raise notice 'row % changed', new.f1;`,
			},
			{
				Statement: `	else
		raise notice 'row % not changed', new.f1;`,
			},
			{
				Statement: `	end if;`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `end$$;`,
			},
			{
				Statement: `UPDATE trigger_test SET f3 = 'bar';`,
			},
			{
				Statement: `UPDATE trigger_test SET f3 = NULL;`,
			},
			{
				Statement: `UPDATE trigger_test SET f3 = NULL;`,
			},
			{
				Statement: `DROP TABLE trigger_test;`,
			},
			{
				Statement: `DROP FUNCTION mytrigger();`,
			},
			{
				Statement: `CREATE FUNCTION serializable_update_trig() RETURNS trigger LANGUAGE plpgsql AS
$$
declare
	rec record;`,
			},
			{
				Statement: `begin
	new.description = 'updated in trigger';`,
			},
			{
				Statement: `	return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TABLE serializable_update_tab (
	id int,
	filler  text,
	description text
);`,
			},
			{
				Statement: `CREATE TRIGGER serializable_update_trig BEFORE UPDATE ON serializable_update_tab
	FOR EACH ROW EXECUTE PROCEDURE serializable_update_trig();`,
			},
			{
				Statement: `INSERT INTO serializable_update_tab SELECT a, repeat('xyzxz', 100), 'new'
	FROM generate_series(1, 50) a;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `UPDATE serializable_update_tab SET description = 'no no', id = 1 WHERE id = 1;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT description FROM serializable_update_tab WHERE id = 1;`,
				Results:   []sql.Row{{`updated in trigger`}},
			},
			{
				Statement: `DROP TABLE serializable_update_tab;`,
			},
			{
				Statement: `CREATE TABLE min_updates_test (
	f1	text,
	f2 int,
	f3 int);`,
			},
			{
				Statement: `INSERT INTO min_updates_test VALUES ('a',1,2),('b','2',null);`,
			},
			{
				Statement: `CREATE TRIGGER z_min_update
BEFORE UPDATE ON min_updates_test
FOR EACH ROW EXECUTE PROCEDURE suppress_redundant_updates_trigger();`,
			},
			{
				Statement: `\set QUIET false
UPDATE min_updates_test SET f1 = f1;`,
			},
			{
				Statement: `UPDATE 0
UPDATE min_updates_test SET f2 = f2 + 1;`,
			},
			{
				Statement: `UPDATE 2
UPDATE min_updates_test SET f3 = 2 WHERE f3 is null;`,
			},
			{
				Statement: `UPDATE 1
\set QUIET true
SELECT * FROM min_updates_test;`,
				Results: []sql.Row{{`a`, 2, 2}, {`b`, 3, 2}},
			},
			{
				Statement: `DROP TABLE min_updates_test;`,
			},
			{
				Statement: `CREATE VIEW main_view AS SELECT a, b FROM main_table;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION view_trigger() RETURNS trigger
LANGUAGE plpgsql AS $$
declare
    argstr text := '';`,
			},
			{
				Statement: `begin
    for i in 0 .. TG_nargs - 1 loop
        if i > 0 then
            argstr := argstr || ', ';`,
			},
			{
				Statement: `        end if;`,
			},
			{
				Statement: `        argstr := argstr || TG_argv[i];`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `    raise notice '% % % % (%)', TG_TABLE_NAME, TG_WHEN, TG_OP, TG_LEVEL, argstr;`,
			},
			{
				Statement: `    if TG_LEVEL = 'ROW' then
        if TG_OP = 'INSERT' then
            raise NOTICE 'NEW: %', NEW;`,
			},
			{
				Statement: `            INSERT INTO main_table VALUES (NEW.a, NEW.b);`,
			},
			{
				Statement: `            RETURN NEW;`,
			},
			{
				Statement: `        end if;`,
			},
			{
				Statement: `        if TG_OP = 'UPDATE' then
            raise NOTICE 'OLD: %, NEW: %', OLD, NEW;`,
			},
			{
				Statement: `            UPDATE main_table SET a = NEW.a, b = NEW.b WHERE a = OLD.a AND b = OLD.b;`,
			},
			{
				Statement: `            if NOT FOUND then RETURN NULL; end if;`,
			},
			{
				Statement: `            RETURN NEW;`,
			},
			{
				Statement: `        end if;`,
			},
			{
				Statement: `        if TG_OP = 'DELETE' then
            raise NOTICE 'OLD: %', OLD;`,
			},
			{
				Statement: `            DELETE FROM main_table WHERE a = OLD.a AND b = OLD.b;`,
			},
			{
				Statement: `            if NOT FOUND then RETURN NULL; end if;`,
			},
			{
				Statement: `            RETURN OLD;`,
			},
			{
				Statement: `        end if;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    RETURN NULL;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig BEFORE INSERT ON main_view
FOR EACH ROW EXECUTE PROCEDURE trigger_func('before_ins_row');`,
				ErrorString: `"main_view" is a view`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig BEFORE UPDATE ON main_view
FOR EACH ROW EXECUTE PROCEDURE trigger_func('before_upd_row');`,
				ErrorString: `"main_view" is a view`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig BEFORE DELETE ON main_view
FOR EACH ROW EXECUTE PROCEDURE trigger_func('before_del_row');`,
				ErrorString: `"main_view" is a view`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig AFTER INSERT ON main_view
FOR EACH ROW EXECUTE PROCEDURE trigger_func('before_ins_row');`,
				ErrorString: `"main_view" is a view`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig AFTER UPDATE ON main_view
FOR EACH ROW EXECUTE PROCEDURE trigger_func('before_upd_row');`,
				ErrorString: `"main_view" is a view`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig AFTER DELETE ON main_view
FOR EACH ROW EXECUTE PROCEDURE trigger_func('before_del_row');`,
				ErrorString: `"main_view" is a view`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig BEFORE TRUNCATE ON main_view
EXECUTE PROCEDURE trigger_func('before_tru_row');`,
				ErrorString: `"main_view" is a view`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig AFTER TRUNCATE ON main_view
EXECUTE PROCEDURE trigger_func('before_tru_row');`,
				ErrorString: `"main_view" is a view`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig INSTEAD OF INSERT ON main_table
FOR EACH ROW EXECUTE PROCEDURE view_trigger('instead_of_ins');`,
				ErrorString: `"main_table" is a table`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig INSTEAD OF UPDATE ON main_table
FOR EACH ROW EXECUTE PROCEDURE view_trigger('instead_of_upd');`,
				ErrorString: `"main_table" is a table`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig INSTEAD OF DELETE ON main_table
FOR EACH ROW EXECUTE PROCEDURE view_trigger('instead_of_del');`,
				ErrorString: `"main_table" is a table`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig INSTEAD OF UPDATE ON main_view
FOR EACH ROW WHEN (OLD.a <> NEW.a) EXECUTE PROCEDURE view_trigger('instead_of_upd');`,
				ErrorString: `INSTEAD OF triggers cannot have WHEN conditions`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig INSTEAD OF UPDATE OF a ON main_view
FOR EACH ROW EXECUTE PROCEDURE view_trigger('instead_of_upd');`,
				ErrorString: `INSTEAD OF triggers cannot have column lists`,
			},
			{
				Statement: `CREATE TRIGGER invalid_trig INSTEAD OF UPDATE ON main_view
EXECUTE PROCEDURE view_trigger('instead_of_upd');`,
				ErrorString: `INSTEAD OF triggers must be FOR EACH ROW`,
			},
			{
				Statement: `CREATE TRIGGER instead_of_insert_trig INSTEAD OF INSERT ON main_view
FOR EACH ROW EXECUTE PROCEDURE view_trigger('instead_of_ins');`,
			},
			{
				Statement: `CREATE TRIGGER instead_of_update_trig INSTEAD OF UPDATE ON main_view
FOR EACH ROW EXECUTE PROCEDURE view_trigger('instead_of_upd');`,
			},
			{
				Statement: `CREATE TRIGGER instead_of_delete_trig INSTEAD OF DELETE ON main_view
FOR EACH ROW EXECUTE PROCEDURE view_trigger('instead_of_del');`,
			},
			{
				Statement: `CREATE TRIGGER before_ins_stmt_trig BEFORE INSERT ON main_view
FOR EACH STATEMENT EXECUTE PROCEDURE view_trigger('before_view_ins_stmt');`,
			},
			{
				Statement: `CREATE TRIGGER before_upd_stmt_trig BEFORE UPDATE ON main_view
FOR EACH STATEMENT EXECUTE PROCEDURE view_trigger('before_view_upd_stmt');`,
			},
			{
				Statement: `CREATE TRIGGER before_del_stmt_trig BEFORE DELETE ON main_view
FOR EACH STATEMENT EXECUTE PROCEDURE view_trigger('before_view_del_stmt');`,
			},
			{
				Statement: `CREATE TRIGGER after_ins_stmt_trig AFTER INSERT ON main_view
FOR EACH STATEMENT EXECUTE PROCEDURE view_trigger('after_view_ins_stmt');`,
			},
			{
				Statement: `CREATE TRIGGER after_upd_stmt_trig AFTER UPDATE ON main_view
FOR EACH STATEMENT EXECUTE PROCEDURE view_trigger('after_view_upd_stmt');`,
			},
			{
				Statement: `CREATE TRIGGER after_del_stmt_trig AFTER DELETE ON main_view
FOR EACH STATEMENT EXECUTE PROCEDURE view_trigger('after_view_del_stmt');`,
			},
			{
				Statement: `\set QUIET false
INSERT INTO main_view VALUES (20, 30);`,
			},
			{
				Statement: `INSERT 0 1
INSERT INTO main_view VALUES (21, 31) RETURNING a, b;`,
				Results: []sql.Row{{21, 31}},
			},
			{
				Statement: `INSERT 0 1
UPDATE main_view SET b = 31 WHERE a = 20;`,
			},
			{
				Statement: `UPDATE 0
UPDATE main_view SET b = 32 WHERE a = 21 AND b = 31 RETURNING a, b;`,
				Results: []sql.Row{},
			},
			{
				Statement: `UPDATE 0
DROP TRIGGER before_upd_a_row_trig ON main_table;`,
			},
			{
				Statement: `DROP TRIGGER
UPDATE main_view SET b = 31 WHERE a = 20;`,
			},
			{
				Statement: `UPDATE 1
UPDATE main_view SET b = 32 WHERE a = 21 AND b = 31 RETURNING a, b;`,
				Results: []sql.Row{{21, 32}},
			},
			{
				Statement: `UPDATE 1
UPDATE main_view SET b = 0 WHERE false;`,
			},
			{
				Statement: `UPDATE 0
DELETE FROM main_view WHERE a IN (20,21);`,
			},
			{
				Statement: `DELETE 3
DELETE FROM main_view WHERE a = 31 RETURNING a, b;`,
				Results: []sql.Row{{31, 10}},
			},
			{
				Statement: `DELETE 1
\set QUIET true
\d main_view
              View "public.main_view"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Triggers:
    after_del_stmt_trig AFTER DELETE ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('after_view_del_stmt')
    after_ins_stmt_trig AFTER INSERT ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('after_view_ins_stmt')
    after_upd_stmt_trig AFTER UPDATE ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('after_view_upd_stmt')
    before_del_stmt_trig BEFORE DELETE ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('before_view_del_stmt')
    before_ins_stmt_trig BEFORE INSERT ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('before_view_ins_stmt')
    before_upd_stmt_trig BEFORE UPDATE ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('before_view_upd_stmt')
    instead_of_delete_trig INSTEAD OF DELETE ON main_view FOR EACH ROW EXECUTE FUNCTION view_trigger('instead_of_del')
    instead_of_insert_trig INSTEAD OF INSERT ON main_view FOR EACH ROW EXECUTE FUNCTION view_trigger('instead_of_ins')
    instead_of_update_trig INSTEAD OF UPDATE ON main_view FOR EACH ROW EXECUTE FUNCTION view_trigger('instead_of_upd')
DROP TRIGGER instead_of_insert_trig ON main_view;`,
			},
			{
				Statement: `DROP TRIGGER instead_of_delete_trig ON main_view;`,
			},
			{
				Statement: `\d+ main_view
                          View "public.main_view"
 Column |  Type   | Collation | Nullable | Default | Storage | Description 
--------+---------+-----------+----------+---------+---------+-------------
 a      | integer |           |          |         | plain   | 
 b      | integer |           |          |         | plain   | 
View definition:
 SELECT main_table.a,
    main_table.b
   FROM main_table;`,
			},
			{
				Statement: `Triggers:
    after_del_stmt_trig AFTER DELETE ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('after_view_del_stmt')
    after_ins_stmt_trig AFTER INSERT ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('after_view_ins_stmt')
    after_upd_stmt_trig AFTER UPDATE ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('after_view_upd_stmt')
    before_del_stmt_trig BEFORE DELETE ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('before_view_del_stmt')
    before_ins_stmt_trig BEFORE INSERT ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('before_view_ins_stmt')
    before_upd_stmt_trig BEFORE UPDATE ON main_view FOR EACH STATEMENT EXECUTE FUNCTION view_trigger('before_view_upd_stmt')
    instead_of_update_trig INSTEAD OF UPDATE ON main_view FOR EACH ROW EXECUTE FUNCTION view_trigger('instead_of_upd')
DROP VIEW main_view;`,
			},
			{
				Statement: `CREATE TABLE country_table (
    country_id        serial primary key,
    country_name    text unique not null,
    continent        text not null
);`,
			},
			{
				Statement: `INSERT INTO country_table (country_name, continent)
    VALUES ('Japan', 'Asia'),
           ('UK', 'Europe'),
           ('USA', 'North America')
    RETURNING *;`,
				Results: []sql.Row{{1, `Japan`, `Asia`}, {2, `UK`, `Europe`}, {3, `USA`, `North America`}},
			},
			{
				Statement: `CREATE TABLE city_table (
    city_id        serial primary key,
    city_name    text not null,
    population    bigint,
    country_id    int references country_table
);`,
			},
			{
				Statement: `CREATE VIEW city_view AS
    SELECT city_id, city_name, population, country_name, continent
    FROM city_table ci
    LEFT JOIN country_table co ON co.country_id = ci.country_id;`,
			},
			{
				Statement: `CREATE FUNCTION city_insert() RETURNS trigger LANGUAGE plpgsql AS $$
declare
    ctry_id int;`,
			},
			{
				Statement: `begin
    if NEW.country_name IS NOT NULL then
        SELECT country_id, continent INTO ctry_id, NEW.continent
            FROM country_table WHERE country_name = NEW.country_name;`,
			},
			{
				Statement: `        if NOT FOUND then
            raise exception 'No such country: "%"', NEW.country_name;`,
			},
			{
				Statement: `        end if;`,
			},
			{
				Statement: `    else
        NEW.continent := NULL;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if NEW.city_id IS NOT NULL then
        INSERT INTO city_table
            VALUES(NEW.city_id, NEW.city_name, NEW.population, ctry_id);`,
			},
			{
				Statement: `    else
        INSERT INTO city_table(city_name, population, country_id)
            VALUES(NEW.city_name, NEW.population, ctry_id)
            RETURNING city_id INTO NEW.city_id;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    RETURN NEW;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER city_insert_trig INSTEAD OF INSERT ON city_view
FOR EACH ROW EXECUTE PROCEDURE city_insert();`,
			},
			{
				Statement: `CREATE FUNCTION city_delete() RETURNS trigger LANGUAGE plpgsql AS $$
begin
    DELETE FROM city_table WHERE city_id = OLD.city_id;`,
			},
			{
				Statement: `    if NOT FOUND then RETURN NULL; end if;`,
			},
			{
				Statement: `    RETURN OLD;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER city_delete_trig INSTEAD OF DELETE ON city_view
FOR EACH ROW EXECUTE PROCEDURE city_delete();`,
			},
			{
				Statement: `CREATE FUNCTION city_update() RETURNS trigger LANGUAGE plpgsql AS $$
declare
    ctry_id int;`,
			},
			{
				Statement: `begin
    if NEW.country_name IS DISTINCT FROM OLD.country_name then
        SELECT country_id, continent INTO ctry_id, NEW.continent
            FROM country_table WHERE country_name = NEW.country_name;`,
			},
			{
				Statement: `        if NOT FOUND then
            raise exception 'No such country: "%"', NEW.country_name;`,
			},
			{
				Statement: `        end if;`,
			},
			{
				Statement: `        UPDATE city_table SET city_name = NEW.city_name,
                              population = NEW.population,
                              country_id = ctry_id
            WHERE city_id = OLD.city_id;`,
			},
			{
				Statement: `    else
        UPDATE city_table SET city_name = NEW.city_name,
                              population = NEW.population
            WHERE city_id = OLD.city_id;`,
			},
			{
				Statement: `        NEW.continent := OLD.continent;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if NOT FOUND then RETURN NULL; end if;`,
			},
			{
				Statement: `    RETURN NEW;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER city_update_trig INSTEAD OF UPDATE ON city_view
FOR EACH ROW EXECUTE PROCEDURE city_update();`,
			},
			{
				Statement: `\set QUIET false
INSERT INTO city_view(city_name) VALUES('Tokyo') RETURNING *;`,
				Results: []sql.Row{{1, `Tokyo`, ``, ``, ``}},
			},
			{
				Statement: `INSERT 0 1
INSERT INTO city_view(city_name, population) VALUES('London', 7556900) RETURNING *;`,
				Results: []sql.Row{{2, `London`, 7556900, ``, ``}},
			},
			{
				Statement: `INSERT 0 1
INSERT INTO city_view(city_name, country_name) VALUES('Washington DC', 'USA') RETURNING *;`,
				Results: []sql.Row{{3, `Washington DC`, ``, `USA`, `North America`}},
			},
			{
				Statement: `INSERT 0 1
INSERT INTO city_view(city_id, city_name) VALUES(123456, 'New York') RETURNING *;`,
				Results: []sql.Row{{123456, `New York`, ``, ``, ``}},
			},
			{
				Statement: `INSERT 0 1
INSERT INTO city_view VALUES(234567, 'Birmingham', 1016800, 'UK', 'EU') RETURNING *;`,
				Results: []sql.Row{{234567, `Birmingham`, 1016800, `UK`, `Europe`}},
			},
			{
				Statement: `INSERT 0 1
UPDATE city_view SET country_name = 'Japon' WHERE city_name = 'Tokyo'; -- error`,
				ErrorString: `No such country: "Japon"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function city_update() line 9 at RAISE
UPDATE city_view SET country_name = 'Japan' WHERE city_name = 'Takyo'; -- no match`,
			},
			{
				Statement: `UPDATE 0
UPDATE city_view SET country_name = 'Japan' WHERE city_name = 'Tokyo' RETURNING *; -- OK`,
				Results: []sql.Row{{1, `Tokyo`, ``, `Japan`, `Asia`}},
			},
			{
				Statement: `UPDATE 1
UPDATE city_view SET population = 13010279 WHERE city_name = 'Tokyo' RETURNING *;`,
				Results: []sql.Row{{1, `Tokyo`, 13010279, `Japan`, `Asia`}},
			},
			{
				Statement: `UPDATE 1
UPDATE city_view SET country_name = 'UK' WHERE city_name = 'New York' RETURNING *;`,
				Results: []sql.Row{{123456, `New York`, ``, `UK`, `Europe`}},
			},
			{
				Statement: `UPDATE 1
UPDATE city_view SET country_name = 'USA', population = 8391881 WHERE city_name = 'New York' RETURNING *;`,
				Results: []sql.Row{{123456, `New York`, 8391881, `USA`, `North America`}},
			},
			{
				Statement: `UPDATE 1
UPDATE city_view SET continent = 'EU' WHERE continent = 'Europe' RETURNING *;`,
				Results: []sql.Row{{234567, `Birmingham`, 1016800, `UK`, `Europe`}},
			},
			{
				Statement: `UPDATE 1
UPDATE city_view v1 SET country_name = v2.country_name FROM city_view v2
    WHERE v2.city_name = 'Birmingham' AND v1.city_name = 'London' RETURNING *;`,
				Results: []sql.Row{{2, `London`, 7556900, `UK`, `Europe`, 234567, `Birmingham`, 1016800, `UK`, `Europe`}},
			},
			{
				Statement: `UPDATE 1
DELETE FROM city_view WHERE city_name = 'Birmingham' RETURNING *;`,
				Results: []sql.Row{{234567, `Birmingham`, 1016800, `UK`, `Europe`}},
			},
			{
				Statement: `DELETE 1
\set QUIET true
CREATE VIEW european_city_view AS
    SELECT * FROM city_view WHERE continent = 'Europe';`,
			},
			{
				Statement: `SELECT count(*) FROM european_city_view;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `CREATE FUNCTION no_op_trig_fn() RETURNS trigger LANGUAGE plpgsql
AS 'begin RETURN NULL; end';`,
			},
			{
				Statement: `CREATE TRIGGER no_op_trig INSTEAD OF INSERT OR UPDATE OR DELETE
ON european_city_view FOR EACH ROW EXECUTE PROCEDURE no_op_trig_fn();`,
			},
			{
				Statement: `\set QUIET false
INSERT INTO european_city_view VALUES (0, 'x', 10000, 'y', 'z');`,
			},
			{
				Statement: `INSERT 0 0
UPDATE european_city_view SET population = 10000;`,
			},
			{
				Statement: `UPDATE 0
DELETE FROM european_city_view;`,
			},
			{
				Statement: `DELETE 0
\set QUIET true
CREATE RULE european_city_insert_rule AS ON INSERT TO european_city_view
DO INSTEAD INSERT INTO city_view
VALUES (NEW.city_id, NEW.city_name, NEW.population, NEW.country_name, NEW.continent)
RETURNING *;`,
			},
			{
				Statement: `CREATE RULE european_city_update_rule AS ON UPDATE TO european_city_view
DO INSTEAD UPDATE city_view SET
    city_name = NEW.city_name,
    population = NEW.population,
    country_name = NEW.country_name
WHERE city_id = OLD.city_id
RETURNING NEW.*;`,
			},
			{
				Statement: `CREATE RULE european_city_delete_rule AS ON DELETE TO european_city_view
DO INSTEAD DELETE FROM city_view WHERE city_id = OLD.city_id RETURNING *;`,
			},
			{
				Statement: `\set QUIET false
INSERT INTO european_city_view(city_name, country_name)
    VALUES ('Cambridge', 'USA') RETURNING *;`,
				Results: []sql.Row{{4, `Cambridge`, ``, `USA`, `North America`}},
			},
			{
				Statement: `INSERT 0 1
UPDATE european_city_view SET country_name = 'UK'
    WHERE city_name = 'Cambridge';`,
			},
			{
				Statement: `UPDATE 0
DELETE FROM european_city_view WHERE city_name = 'Cambridge';`,
			},
			{
				Statement: `DELETE 0
UPDATE city_view SET country_name = 'UK'
    WHERE city_name = 'Cambridge' RETURNING *;`,
				Results: []sql.Row{{4, `Cambridge`, ``, `UK`, `Europe`}},
			},
			{
				Statement: `UPDATE 1
UPDATE european_city_view SET population = 122800
    WHERE city_name = 'Cambridge' RETURNING *;`,
				Results: []sql.Row{{4, `Cambridge`, 122800, `UK`, `Europe`}},
			},
			{
				Statement: `UPDATE 1
DELETE FROM european_city_view WHERE city_name = 'Cambridge' RETURNING *;`,
				Results: []sql.Row{{4, `Cambridge`, 122800, `UK`, `Europe`}},
			},
			{
				Statement: `DELETE 1
UPDATE city_view v SET population = 599657
    FROM city_table ci, country_table co
    WHERE ci.city_name = 'Washington DC' and co.country_name = 'USA'
    AND v.city_id = ci.city_id AND v.country_name = co.country_name
    RETURNING co.country_id, v.country_name,
              v.city_id, v.city_name, v.population;`,
				Results: []sql.Row{{3, `USA`, 3, `Washington DC`, 599657}},
			},
			{
				Statement: `UPDATE 1
\set QUIET true
SELECT * FROM city_view;`,
				Results: []sql.Row{{1, `Tokyo`, 13010279, `Japan`, `Asia`}, {123456, `New York`, 8391881, `USA`, `North America`}, {2, `London`, 7556900, `UK`, `Europe`}, {3, `Washington DC`, 599657, `USA`, `North America`}},
			},
			{
				Statement: `DROP TABLE city_table CASCADE;`,
			},
			{
				Statement: `DROP TABLE country_table;`,
			},
			{
				Statement: `create table depth_a (id int not null primary key);`,
			},
			{
				Statement: `create table depth_b (id int not null primary key);`,
			},
			{
				Statement: `create table depth_c (id int not null primary key);`,
			},
			{
				Statement: `create function depth_a_tf() returns trigger
  language plpgsql as $$
begin
  raise notice '%: depth = %', tg_name, pg_trigger_depth();`,
			},
			{
				Statement: `  insert into depth_b values (new.id);`,
			},
			{
				Statement: `  raise notice '%: depth = %', tg_name, pg_trigger_depth();`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger depth_a_tr before insert on depth_a
  for each row execute procedure depth_a_tf();`,
			},
			{
				Statement: `create function depth_b_tf() returns trigger
  language plpgsql as $$
begin
  raise notice '%: depth = %', tg_name, pg_trigger_depth();`,
			},
			{
				Statement: `  begin
    execute 'insert into depth_c values (' || new.id::text || ')';`,
			},
			{
				Statement: `  exception
    when sqlstate 'U9999' then
      raise notice 'SQLSTATE = U9999: depth = %', pg_trigger_depth();`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `  raise notice '%: depth = %', tg_name, pg_trigger_depth();`,
			},
			{
				Statement: `  if new.id = 1 then
    execute 'insert into depth_c values (' || new.id::text || ')';`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger depth_b_tr before insert on depth_b
  for each row execute procedure depth_b_tf();`,
			},
			{
				Statement: `create function depth_c_tf() returns trigger
  language plpgsql as $$
begin
  raise notice '%: depth = %', tg_name, pg_trigger_depth();`,
			},
			{
				Statement: `  if new.id = 1 then
    raise exception sqlstate 'U9999';`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  raise notice '%: depth = %', tg_name, pg_trigger_depth();`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger depth_c_tr before insert on depth_c
  for each row execute procedure depth_c_tf();`,
			},
			{
				Statement: `select pg_trigger_depth();`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `insert into depth_a values (1);`,
				ErrorString: `U9999`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function depth_c_tf() line 5 at RAISE
SQL statement "insert into depth_c values (1)"
PL/pgSQL function depth_b_tf() line 12 at EXECUTE
SQL statement "insert into depth_b values (new.id)"
PL/pgSQL function depth_a_tf() line 4 at SQL statement
select pg_trigger_depth();`,
				Results: []sql.Row{{0}},
			},
			{
				Statement: `insert into depth_a values (2);`,
			},
			{
				Statement: `select pg_trigger_depth();`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `drop table depth_a, depth_b, depth_c;`,
			},
			{
				Statement: `drop function depth_a_tf();`,
			},
			{
				Statement: `drop function depth_b_tf();`,
			},
			{
				Statement: `drop function depth_c_tf();`,
			},
			{
				Statement: `create temp table parent (
    aid int not null primary key,
    val1 text,
    val2 text,
    val3 text,
    val4 text,
    bcnt int not null default 0);`,
			},
			{
				Statement: `create temp table child (
    bid int not null primary key,
    aid int not null,
    val1 text);`,
			},
			{
				Statement: `create function parent_upd_func()
  returns trigger language plpgsql as
$$
begin
  if old.val1 <> new.val1 then
    new.val2 = new.val1;`,
			},
			{
				Statement: `    delete from child where child.aid = new.aid and child.val1 = new.val1;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger parent_upd_trig before update on parent
  for each row execute procedure parent_upd_func();`,
			},
			{
				Statement: `create function parent_del_func()
  returns trigger language plpgsql as
$$
begin
  delete from child where aid = old.aid;`,
			},
			{
				Statement: `  return old;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger parent_del_trig before delete on parent
  for each row execute procedure parent_del_func();`,
			},
			{
				Statement: `create function child_ins_func()
  returns trigger language plpgsql as
$$
begin
  update parent set bcnt = bcnt + 1 where aid = new.aid;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger child_ins_trig after insert on child
  for each row execute procedure child_ins_func();`,
			},
			{
				Statement: `create function child_del_func()
  returns trigger language plpgsql as
$$
begin
  update parent set bcnt = bcnt - 1 where aid = old.aid;`,
			},
			{
				Statement: `  return old;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger child_del_trig after delete on child
  for each row execute procedure child_del_func();`,
			},
			{
				Statement: `insert into parent values (1, 'a', 'a', 'a', 'a', 0);`,
			},
			{
				Statement: `insert into child values (10, 1, 'b');`,
			},
			{
				Statement: `select * from parent; select * from child;`,
				Results:   []sql.Row{{1, `a`, `a`, `a`, `a`, 1}},
			},
			{
				Statement: ` bid | aid | val1 
-----+-----+------
  10 |   1 | b
(1 row)
update parent set val1 = 'b' where aid = 1; -- should fail`,
				ErrorString: `tuple to be updated was already modified by an operation triggered by the current command`,
			},
			{
				Statement: `select * from parent; select * from child;`,
				Results:   []sql.Row{{1, `a`, `a`, `a`, `a`, 1}},
			},
			{
				Statement: ` bid | aid | val1 
-----+-----+------
  10 |   1 | b
(1 row)
delete from parent where aid = 1; -- should fail`,
				ErrorString: `tuple to be deleted was already modified by an operation triggered by the current command`,
			},
			{
				Statement: `select * from parent; select * from child;`,
				Results:   []sql.Row{{1, `a`, `a`, `a`, `a`, 1}},
			},
			{
				Statement: ` bid | aid | val1 
-----+-----+------
  10 |   1 | b
(1 row)
create or replace function parent_del_func()
  returns trigger language plpgsql as
$$
begin
  delete from child where aid = old.aid;`,
			},
			{
				Statement: `  if found then
    delete from parent where aid = old.aid;`,
			},
			{
				Statement: `    return null; -- cancel outer deletion`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return old;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `delete from parent where aid = 1;`,
			},
			{
				Statement: `select * from parent; select * from child;`,
				Results:   []sql.Row{},
			},
			{
				Statement: ` bid | aid | val1 
-----+-----+------
(0 rows)
drop table parent, child;`,
			},
			{
				Statement: `drop function parent_upd_func();`,
			},
			{
				Statement: `drop function parent_del_func();`,
			},
			{
				Statement: `drop function child_ins_func();`,
			},
			{
				Statement: `drop function child_del_func();`,
			},
			{
				Statement: `create temp table self_ref_trigger (
    id int primary key,
    parent int references self_ref_trigger,
    data text,
    nchildren int not null default 0
);`,
			},
			{
				Statement: `create function self_ref_trigger_ins_func()
  returns trigger language plpgsql as
$$
begin
  if new.parent is not null then
    update self_ref_trigger set nchildren = nchildren + 1
      where id = new.parent;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger self_ref_trigger_ins_trig before insert on self_ref_trigger
  for each row execute procedure self_ref_trigger_ins_func();`,
			},
			{
				Statement: `create function self_ref_trigger_del_func()
  returns trigger language plpgsql as
$$
begin
  if old.parent is not null then
    update self_ref_trigger set nchildren = nchildren - 1
      where id = old.parent;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return old;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger self_ref_trigger_del_trig before delete on self_ref_trigger
  for each row execute procedure self_ref_trigger_del_func();`,
			},
			{
				Statement: `insert into self_ref_trigger values (1, null, 'root');`,
			},
			{
				Statement: `insert into self_ref_trigger values (2, 1, 'root child A');`,
			},
			{
				Statement: `insert into self_ref_trigger values (3, 1, 'root child B');`,
			},
			{
				Statement: `insert into self_ref_trigger values (4, 2, 'grandchild 1');`,
			},
			{
				Statement: `insert into self_ref_trigger values (5, 3, 'grandchild 2');`,
			},
			{
				Statement: `update self_ref_trigger set data = 'root!' where id = 1;`,
			},
			{
				Statement: `select * from self_ref_trigger;`,
				Results:   []sql.Row{{2, 1, `root child A`, 1}, {4, 2, `grandchild 1`, 0}, {3, 1, `root child B`, 1}, {5, 3, `grandchild 2`, 0}, {1, ``, `root!`, 2}},
			},
			{
				Statement:   `delete from self_ref_trigger;`,
				ErrorString: `tuple to be updated was already modified by an operation triggered by the current command`,
			},
			{
				Statement: `select * from self_ref_trigger;`,
				Results:   []sql.Row{{2, 1, `root child A`, 1}, {4, 2, `grandchild 1`, 0}, {3, 1, `root child B`, 1}, {5, 3, `grandchild 2`, 0}, {1, ``, `root!`, 2}},
			},
			{
				Statement: `drop table self_ref_trigger;`,
			},
			{
				Statement: `drop function self_ref_trigger_ins_func();`,
			},
			{
				Statement: `drop function self_ref_trigger_del_func();`,
			},
			{
				Statement: `create table stmt_trig_on_empty_upd (a int);`,
			},
			{
				Statement: `create table stmt_trig_on_empty_upd1 () inherits (stmt_trig_on_empty_upd);`,
			},
			{
				Statement: `create function update_stmt_notice() returns trigger as $$
begin
	raise notice 'updating %', TG_TABLE_NAME;`,
			},
			{
				Statement: `	return null;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create trigger before_stmt_trigger
	before update on stmt_trig_on_empty_upd
	execute procedure update_stmt_notice();`,
			},
			{
				Statement: `create trigger before_stmt_trigger
	before update on stmt_trig_on_empty_upd1
	execute procedure update_stmt_notice();`,
			},
			{
				Statement: `update stmt_trig_on_empty_upd set a = a where false returning a+1 as aa;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `update stmt_trig_on_empty_upd1 set a = a where false returning a+1 as aa;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `drop table stmt_trig_on_empty_upd cascade;`,
			},
			{
				Statement: `drop function update_stmt_notice();`,
			},
			{
				Statement: `create table trigger_ddl_table (
   col1 integer,
   col2 integer
);`,
			},
			{
				Statement: `create function trigger_ddl_func() returns trigger as $$
begin
  alter table trigger_ddl_table add primary key (col1);`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `create trigger trigger_ddl_func before insert on trigger_ddl_table for each row
  execute procedure trigger_ddl_func();`,
			},
			{
				Statement:   `insert into trigger_ddl_table values (1, 42);  -- fail`,
				ErrorString: `cannot ALTER TABLE "trigger_ddl_table" because it is being used by active queries in this session`,
			},
			{
				Statement: `CONTEXT:  SQL statement "alter table trigger_ddl_table add primary key (col1)"
PL/pgSQL function trigger_ddl_func() line 3 at SQL statement
create or replace function trigger_ddl_func() returns trigger as $$
begin
  create index on trigger_ddl_table (col2);`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `insert into trigger_ddl_table values (1, 42);  -- fail`,
				ErrorString: `cannot CREATE INDEX "trigger_ddl_table" because it is being used by active queries in this session`,
			},
			{
				Statement: `CONTEXT:  SQL statement "create index on trigger_ddl_table (col2)"
PL/pgSQL function trigger_ddl_func() line 3 at SQL statement
drop table trigger_ddl_table;`,
			},
			{
				Statement: `drop function trigger_ddl_func();`,
			},
			{
				Statement: `create table upsert (key int4 primary key, color text);`,
			},
			{
				Statement: `create function upsert_before_func()
  returns trigger language plpgsql as
$$
begin
  if (TG_OP = 'UPDATE') then
    raise warning 'before update (old): %', old.*::text;`,
			},
			{
				Statement: `    raise warning 'before update (new): %', new.*::text;`,
			},
			{
				Statement: `  elsif (TG_OP = 'INSERT') then
    raise warning 'before insert (new): %', new.*::text;`,
			},
			{
				Statement: `    if new.key % 2 = 0 then
      new.key := new.key + 1;`,
			},
			{
				Statement: `      new.color := new.color || ' trig modified';`,
			},
			{
				Statement: `      raise warning 'before insert (new, modified): %', new.*::text;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger upsert_before_trig before insert or update on upsert
  for each row execute procedure upsert_before_func();`,
			},
			{
				Statement: `create function upsert_after_func()
  returns trigger language plpgsql as
$$
begin
  if (TG_OP = 'UPDATE') then
    raise warning 'after update (old): %', old.*::text;`,
			},
			{
				Statement: `    raise warning 'after update (new): %', new.*::text;`,
			},
			{
				Statement: `  elsif (TG_OP = 'INSERT') then
    raise warning 'after insert (new): %', new.*::text;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return null;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger upsert_after_trig after insert or update on upsert
  for each row execute procedure upsert_after_func();`,
			},
			{
				Statement: `insert into upsert values(1, 'black') on conflict (key) do update set color = 'updated ' || upsert.color;`,
			},
			{
				Statement: `insert into upsert values(2, 'red') on conflict (key) do update set color = 'updated ' || upsert.color;`,
			},
			{
				Statement: `insert into upsert values(3, 'orange') on conflict (key) do update set color = 'updated ' || upsert.color;`,
			},
			{
				Statement: `insert into upsert values(4, 'green') on conflict (key) do update set color = 'updated ' || upsert.color;`,
			},
			{
				Statement: `insert into upsert values(5, 'purple') on conflict (key) do update set color = 'updated ' || upsert.color;`,
			},
			{
				Statement: `insert into upsert values(6, 'white') on conflict (key) do update set color = 'updated ' || upsert.color;`,
			},
			{
				Statement: `insert into upsert values(7, 'pink') on conflict (key) do update set color = 'updated ' || upsert.color;`,
			},
			{
				Statement: `insert into upsert values(8, 'yellow') on conflict (key) do update set color = 'updated ' || upsert.color;`,
			},
			{
				Statement: `select * from upsert;`,
				Results:   []sql.Row{{1, `black`}, {3, `updated red trig modified`}, {5, `updated green trig modified`}, {7, `updated white trig modified`}, {9, `yellow trig modified`}},
			},
			{
				Statement: `drop table upsert;`,
			},
			{
				Statement: `drop function upsert_before_func();`,
			},
			{
				Statement: `drop function upsert_after_func();`,
			},
			{
				Statement: `create table my_table (i int);`,
			},
			{
				Statement: `create view my_view as select * from my_table;`,
			},
			{
				Statement: `create function my_trigger_function() returns trigger as $$ begin end; $$ language plpgsql;`,
			},
			{
				Statement: `create trigger my_trigger after update on my_view referencing old table as old_table
   for each statement execute procedure my_trigger_function();`,
				ErrorString: `"my_view" is a view`,
			},
			{
				Statement: `drop function my_trigger_function();`,
			},
			{
				Statement: `drop view my_view;`,
			},
			{
				Statement: `drop table my_table;`,
			},
			{
				Statement: `create table parted_trig (a int) partition by list (a);`,
			},
			{
				Statement: `create function trigger_nothing() returns trigger
  language plpgsql as $$ begin end; $$;`,
			},
			{
				Statement: `create trigger failed instead of update on parted_trig
  for each row execute procedure trigger_nothing();`,
				ErrorString: `"parted_trig" is a table`,
			},
			{
				Statement: `create trigger failed after update on parted_trig
  referencing old table as old_table
  for each row execute procedure trigger_nothing();`,
				ErrorString: `"parted_trig" is a partitioned table`,
			},
			{
				Statement: `drop table parted_trig;`,
			},
			{
				Statement: `create table trigpart (a int, b int) partition by range (a);`,
			},
			{
				Statement: `create table trigpart1 partition of trigpart for values from (0) to (1000);`,
			},
			{
				Statement: `create trigger trg1 after insert on trigpart for each row execute procedure trigger_nothing();`,
			},
			{
				Statement: `create table trigpart2 partition of trigpart for values from (1000) to (2000);`,
			},
			{
				Statement: `create table trigpart3 (like trigpart);`,
			},
			{
				Statement: `alter table trigpart attach partition trigpart3 for values from (2000) to (3000);`,
			},
			{
				Statement: `create table trigpart4 partition of trigpart for values from (3000) to (4000) partition by range (a);`,
			},
			{
				Statement: `create table trigpart41 partition of trigpart4 for values from (3000) to (3500);`,
			},
			{
				Statement: `create table trigpart42 (like trigpart);`,
			},
			{
				Statement: `alter table trigpart4 attach partition trigpart42 for values from (3500) to (4000);`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgfoid::regproc from pg_trigger
  where tgrelid::regclass::text like 'trigpart%' order by tgrelid::regclass::text;`,
				Results: []sql.Row{{`trigpart`, `trg1`, `trigger_nothing`}, {`trigpart1`, `trg1`, `trigger_nothing`}, {`trigpart2`, `trg1`, `trigger_nothing`}, {`trigpart3`, `trg1`, `trigger_nothing`}, {`trigpart4`, `trg1`, `trigger_nothing`}, {`trigpart41`, `trg1`, `trigger_nothing`}, {`trigpart42`, `trg1`, `trigger_nothing`}},
			},
			{
				Statement:   `drop trigger trg1 on trigpart1;	-- fail`,
				ErrorString: `cannot drop trigger trg1 on table trigpart1 because trigger trg1 on table trigpart requires it`,
			},
			{
				Statement:   `drop trigger trg1 on trigpart2;	-- fail`,
				ErrorString: `cannot drop trigger trg1 on table trigpart2 because trigger trg1 on table trigpart requires it`,
			},
			{
				Statement:   `drop trigger trg1 on trigpart3;	-- fail`,
				ErrorString: `cannot drop trigger trg1 on table trigpart3 because trigger trg1 on table trigpart requires it`,
			},
			{
				Statement: `drop table trigpart2;			-- ok, trigger should be gone in that partition`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgfoid::regproc from pg_trigger
  where tgrelid::regclass::text like 'trigpart%' order by tgrelid::regclass::text;`,
				Results: []sql.Row{{`trigpart`, `trg1`, `trigger_nothing`}, {`trigpart1`, `trg1`, `trigger_nothing`}, {`trigpart3`, `trg1`, `trigger_nothing`}, {`trigpart4`, `trg1`, `trigger_nothing`}, {`trigpart41`, `trg1`, `trigger_nothing`}, {`trigpart42`, `trg1`, `trigger_nothing`}},
			},
			{
				Statement: `drop trigger trg1 on trigpart;		-- ok, all gone`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgfoid::regproc from pg_trigger
  where tgrelid::regclass::text like 'trigpart%' order by tgrelid::regclass::text;`,
				Results: []sql.Row{},
			},
			{
				Statement: `create trigger trg1 after insert on trigpart for each row execute procedure trigger_nothing();`,
			},
			{
				Statement: `\d trigpart3
             Table "public.trigpart3"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: trigpart FOR VALUES FROM (2000) TO (3000)
Triggers:
    trg1 AFTER INSERT ON trigpart3 FOR EACH ROW EXECUTE FUNCTION trigger_nothing(), ON TABLE trigpart
alter table trigpart detach partition trigpart3;`,
			},
			{
				Statement:   `drop trigger trg1 on trigpart3; -- fail due to "does not exist"`,
				ErrorString: `trigger "trg1" for table "trigpart3" does not exist`,
			},
			{
				Statement: `alter table trigpart detach partition trigpart4;`,
			},
			{
				Statement:   `drop trigger trg1 on trigpart41; -- fail due to "does not exist"`,
				ErrorString: `trigger "trg1" for table "trigpart41" does not exist`,
			},
			{
				Statement: `drop table trigpart4;`,
			},
			{
				Statement: `alter table trigpart attach partition trigpart3 for values from (2000) to (3000);`,
			},
			{
				Statement: `alter table trigpart detach partition trigpart3;`,
			},
			{
				Statement: `alter table trigpart attach partition trigpart3 for values from (2000) to (3000);`,
			},
			{
				Statement: `drop table trigpart3;`,
			},
			{
				Statement: `select tgrelid::regclass::text, tgname, tgfoid::regproc, tgenabled, tgisinternal from pg_trigger
  where tgname ~ '^trg1' order by 1;`,
				Results: []sql.Row{{`trigpart`, `trg1`, `trigger_nothing`, `O`, false}, {`trigpart1`, `trg1`, `trigger_nothing`, `O`, false}},
			},
			{
				Statement: `create table trigpart3 (like trigpart);`,
			},
			{
				Statement: `create trigger trg1 after insert on trigpart3 for each row execute procedure trigger_nothing();`,
			},
			{
				Statement: `\d trigpart3
             Table "public.trigpart3"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Triggers:
    trg1 AFTER INSERT ON trigpart3 FOR EACH ROW EXECUTE FUNCTION trigger_nothing()
alter table trigpart attach partition trigpart3 FOR VALUES FROM (2000) to (3000); -- fail`,
				ErrorString: `trigger "trg1" for relation "trigpart3" already exists`,
			},
			{
				Statement: `drop table trigpart3;`,
			},
			{
				Statement: `create trigger samename after delete on trigpart execute function trigger_nothing();`,
			},
			{
				Statement: `create trigger samename after delete on trigpart1 execute function trigger_nothing();`,
			},
			{
				Statement: `\d trigpart1
             Table "public.trigpart1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Partition of: trigpart FOR VALUES FROM (0) TO (1000)
Triggers:
    samename AFTER DELETE ON trigpart1 FOR EACH STATEMENT EXECUTE FUNCTION trigger_nothing()
    trg1 AFTER INSERT ON trigpart1 FOR EACH ROW EXECUTE FUNCTION trigger_nothing(), ON TABLE trigpart
drop table trigpart;`,
			},
			{
				Statement: `drop function trigger_nothing();`,
			},
			{
				Statement: `create table parted_stmt_trig (a int) partition by list (a);`,
			},
			{
				Statement: `create table parted_stmt_trig1 partition of parted_stmt_trig for values in (1);`,
			},
			{
				Statement: `create table parted_stmt_trig2 partition of parted_stmt_trig for values in (2);`,
			},
			{
				Statement: `create table parted2_stmt_trig (a int) partition by list (a);`,
			},
			{
				Statement: `create table parted2_stmt_trig1 partition of parted2_stmt_trig for values in (1);`,
			},
			{
				Statement: `create table parted2_stmt_trig2 partition of parted2_stmt_trig for values in (2);`,
			},
			{
				Statement: `create or replace function trigger_notice() returns trigger as $$
  begin
    raise notice 'trigger % on % % % for %', TG_NAME, TG_TABLE_NAME, TG_WHEN, TG_OP, TG_LEVEL;`,
			},
			{
				Statement: `    if TG_LEVEL = 'ROW' then
       return NEW;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `  $$ language plpgsql;`,
			},
			{
				Statement: `create trigger trig_ins_before before insert on parted_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_ins_after after insert on parted_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_upd_before before update on parted_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_upd_after after update on parted_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_del_before before delete on parted_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_del_after after delete on parted_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_ins_after_parent after insert on parted_stmt_trig
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_upd_after_parent after update on parted_stmt_trig
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_del_after_parent after delete on parted_stmt_trig
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_ins_before_child before insert on parted_stmt_trig1
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_ins_after_child after insert on parted_stmt_trig1
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_upd_before_child before update on parted_stmt_trig1
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_upd_after_child after update on parted_stmt_trig1
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_del_before_child before delete on parted_stmt_trig1
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_del_after_child after delete on parted_stmt_trig1
  for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_ins_before_3 before insert on parted2_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_ins_after_3 after insert on parted2_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_upd_before_3 before update on parted2_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_upd_after_3 after update on parted2_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_del_before_3 before delete on parted2_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger trig_del_after_3 after delete on parted2_stmt_trig
  for each statement execute procedure trigger_notice();`,
			},
			{
				Statement: `with ins (a) as (
  insert into parted2_stmt_trig values (1), (2) returning a
) insert into parted_stmt_trig select a from ins returning tableoid::regclass, a;`,
				Results: []sql.Row{{`parted_stmt_trig1`, 1}, {`parted_stmt_trig2`, 2}},
			},
			{
				Statement: `with upd as (
  update parted2_stmt_trig set a = a
) update parted_stmt_trig  set a = a;`,
			},
			{
				Statement: `delete from parted_stmt_trig;`,
			},
			{
				Statement: `copy parted_stmt_trig(a) from stdin;`,
			},
			{
				Statement: `copy parted_stmt_trig1(a) from stdin;`,
			},
			{
				Statement: `alter table parted_stmt_trig disable trigger trig_ins_after_parent;`,
			},
			{
				Statement: `insert into parted_stmt_trig values (1);`,
			},
			{
				Statement: `alter table parted_stmt_trig enable trigger trig_ins_after_parent;`,
			},
			{
				Statement: `insert into parted_stmt_trig values (1);`,
			},
			{
				Statement: `drop table parted_stmt_trig, parted2_stmt_trig;`,
			},
			{
				Statement: `create table parted_trig (a int) partition by range (a);`,
			},
			{
				Statement: `create table parted_trig_1 partition of parted_trig for values from (0) to (1000)
   partition by range (a);`,
			},
			{
				Statement: `create table parted_trig_1_1 partition of parted_trig_1 for values from (0) to (100);`,
			},
			{
				Statement: `create table parted_trig_2 partition of parted_trig for values from (1000) to (2000);`,
			},
			{
				Statement: `create trigger zzz after insert on parted_trig for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger mmm after insert on parted_trig_1_1 for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger aaa after insert on parted_trig_1 for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger bbb after insert on parted_trig for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `create trigger qqq after insert on parted_trig_1_1 for each row execute procedure trigger_notice();`,
			},
			{
				Statement: `insert into parted_trig values (50), (1500);`,
			},
			{
				Statement: `drop table parted_trig;`,
			},
			{
				Statement: `create table parted_trig (a int) partition by list (a);`,
			},
			{
				Statement: `create table parted_trig1 partition of parted_trig for values in (1);`,
			},
			{
				Statement: `create or replace function trigger_notice() returns trigger as $$
  declare
    arg1 text = TG_ARGV[0];`,
			},
			{
				Statement: `    arg2 integer = TG_ARGV[1];`,
			},
			{
				Statement: `  begin
    raise notice 'trigger % on % % % for % args % %',
		TG_NAME, TG_TABLE_NAME, TG_WHEN, TG_OP, TG_LEVEL, arg1, arg2;`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `  $$ language plpgsql;`,
			},
			{
				Statement: `create trigger aaa after insert on parted_trig
   for each row execute procedure trigger_notice('quirky', 1);`,
			},
			{
				Statement: `create table parted_trig2 partition of parted_trig for values in (2);`,
			},
			{
				Statement: `create table parted_trig3 (like parted_trig);`,
			},
			{
				Statement: `alter table parted_trig attach partition parted_trig3 for values in (3);`,
			},
			{
				Statement: `insert into parted_trig values (1), (2), (3);`,
			},
			{
				Statement: `drop table parted_trig;`,
			},
			{
				Statement: `create function bark(text) returns bool language plpgsql immutable
  as $$ begin raise notice '% <- woof!', $1; return true; end; $$;`,
			},
			{
				Statement: `create or replace function trigger_notice_ab() returns trigger as $$
  begin
    raise notice 'trigger % on % % % for %: (a,b)=(%,%)',
		TG_NAME, TG_TABLE_NAME, TG_WHEN, TG_OP, TG_LEVEL,
		NEW.a, NEW.b;`,
			},
			{
				Statement: `    if TG_LEVEL = 'ROW' then
       return NEW;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `  $$ language plpgsql;`,
			},
			{
				Statement: `create table parted_irreg_ancestor (fd text, b text, fd2 int, fd3 int, a int)
  partition by range (b);`,
			},
			{
				Statement: `alter table parted_irreg_ancestor drop column fd,
  drop column fd2, drop column fd3;`,
			},
			{
				Statement: `create table parted_irreg (fd int, a int, fd2 int, b text)
  partition by range (b);`,
			},
			{
				Statement: `alter table parted_irreg drop column fd, drop column fd2;`,
			},
			{
				Statement: `alter table parted_irreg_ancestor attach partition parted_irreg
  for values from ('aaaa') to ('zzzz');`,
			},
			{
				Statement: `create table parted1_irreg (b text, fd int, a int);`,
			},
			{
				Statement: `alter table parted1_irreg drop column fd;`,
			},
			{
				Statement: `alter table parted_irreg attach partition parted1_irreg
  for values from ('aaaa') to ('bbbb');`,
			},
			{
				Statement: `create trigger parted_trig after insert on parted_irreg
  for each row execute procedure trigger_notice_ab();`,
			},
			{
				Statement: `create trigger parted_trig_odd after insert on parted_irreg for each row
  when (bark(new.b) AND new.a % 2 = 1) execute procedure trigger_notice_ab();`,
			},
			{
				Statement: `insert into parted_irreg values (1, 'aardvark'), (2, 'aanimals');`,
			},
			{
				Statement: `insert into parted1_irreg values ('aardwolf', 2);`,
			},
			{
				Statement: `insert into parted_irreg_ancestor values ('aasvogel', 3);`,
			},
			{
				Statement: `drop table parted_irreg_ancestor;`,
			},
			{
				Statement: `create table parted (a int, b int, c text) partition by list (a);`,
			},
			{
				Statement: `create table parted_1 partition of parted for values in (1)
  partition by list (b);`,
			},
			{
				Statement: `create table parted_1_1 partition of parted_1 for values in (1);`,
			},
			{
				Statement: `create function parted_trigfunc() returns trigger language plpgsql as $$
begin
  new.a = new.a + 1;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `insert into parted values (1, 1, 'uno uno v1');    -- works`,
			},
			{
				Statement: `create trigger t before insert or update or delete on parted
  for each row execute function parted_trigfunc();`,
			},
			{
				Statement:   `insert into parted values (1, 1, 'uno uno v2');    -- fail`,
				ErrorString: `moving row to another partition during a BEFORE FOR EACH ROW trigger is not supported`,
			},
			{
				Statement:   `update parted set c = c || 'v3';                   -- fail`,
				ErrorString: `no partition of relation "parted" found for row`,
			},
			{
				Statement: `create or replace function parted_trigfunc() returns trigger language plpgsql as $$
begin
  new.b = new.b + 1;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement:   `insert into parted values (1, 1, 'uno uno v4');    -- fail`,
				ErrorString: `moving row to another partition during a BEFORE FOR EACH ROW trigger is not supported`,
			},
			{
				Statement:   `update parted set c = c || 'v5';                   -- fail`,
				ErrorString: `no partition of relation "parted_1" found for row`,
			},
			{
				Statement: `create or replace function parted_trigfunc() returns trigger language plpgsql as $$
begin
  new.c = new.c || ' did '|| TG_OP;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `insert into parted values (1, 1, 'uno uno');       -- works`,
			},
			{
				Statement: `update parted set c = c || ' v6';                   -- works`,
			},
			{
				Statement: `select tableoid::regclass, * from parted;`,
				Results:   []sql.Row{{`parted_1_1`, 1, 1, `uno uno v1 v6 did UPDATE`}, {`parted_1_1`, 1, 1, `uno uno did INSERT v6 did UPDATE`}},
			},
			{
				Statement: `truncate table parted;`,
			},
			{
				Statement: `create table parted_2 partition of parted for values in (2);`,
			},
			{
				Statement: `insert into parted values (1, 1, 'uno uno v5');`,
			},
			{
				Statement: `update parted set a = 2;`,
			},
			{
				Statement: `select tableoid::regclass, * from parted;`,
				Results:   []sql.Row{{`parted_2`, 2, 1, `uno uno v5 did INSERT did UPDATE did INSERT`}},
			},
			{
				Statement: `create or replace function parted_trigfunc2() returns trigger language plpgsql as $$
begin
  new.a = new.a + 1;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger t2 before update on parted
  for each row execute function parted_trigfunc2();`,
			},
			{
				Statement: `truncate table parted;`,
			},
			{
				Statement: `insert into parted values (1, 1, 'uno uno v6');`,
			},
			{
				Statement: `create table parted_3 partition of parted for values in (3);`,
			},
			{
				Statement: `update parted set a = a + 1;`,
			},
			{
				Statement: `select tableoid::regclass, * from parted;`,
				Results:   []sql.Row{{`parted_3`, 3, 1, `uno uno v6 did INSERT did UPDATE did INSERT`}},
			},
			{
				Statement: `update parted set a = 0;`,
			},
			{
				Statement: `select tableoid::regclass, * from parted;`,
				Results:   []sql.Row{{`parted_1_1`, 1, 1, `uno uno v6 did INSERT did UPDATE did INSERT did UPDATE did INSERT`}},
			},
			{
				Statement: `drop table parted;`,
			},
			{
				Statement: `create table parted (a int, b int, c text) partition by list ((a + b));`,
			},
			{
				Statement: `create or replace function parted_trigfunc() returns trigger language plpgsql as $$
begin
  new.a = new.a + new.b;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create table parted_1 partition of parted for values in (1, 2);`,
			},
			{
				Statement: `create table parted_2 partition of parted for values in (3, 4);`,
			},
			{
				Statement: `create trigger t before insert or update on parted
  for each row execute function parted_trigfunc();`,
			},
			{
				Statement: `insert into parted values (0, 1, 'zero win');`,
			},
			{
				Statement:   `insert into parted values (1, 1, 'one fail');`,
				ErrorString: `moving row to another partition during a BEFORE FOR EACH ROW trigger is not supported`,
			},
			{
				Statement:   `insert into parted values (1, 2, 'two fail');`,
				ErrorString: `moving row to another partition during a BEFORE FOR EACH ROW trigger is not supported`,
			},
			{
				Statement: `select * from parted;`,
				Results:   []sql.Row{{1, 1, `zero win`}},
			},
			{
				Statement: `drop table parted;`,
			},
			{
				Statement: `drop function parted_trigfunc();`,
			},
			{
				Statement: `create table parted_constr_ancestor (a int, b text)
  partition by range (b);`,
			},
			{
				Statement: `create table parted_constr (a int, b text)
  partition by range (b);`,
			},
			{
				Statement: `alter table parted_constr_ancestor attach partition parted_constr
  for values from ('aaaa') to ('zzzz');`,
			},
			{
				Statement: `create table parted1_constr (a int, b text);`,
			},
			{
				Statement: `alter table parted_constr attach partition parted1_constr
  for values from ('aaaa') to ('bbbb');`,
			},
			{
				Statement: `create constraint trigger parted_trig after insert on parted_constr_ancestor
  deferrable
  for each row execute procedure trigger_notice_ab();`,
			},
			{
				Statement: `create constraint trigger parted_trig_two after insert on parted_constr
  deferrable initially deferred
  for each row when (bark(new.b) AND new.a % 2 = 1)
  execute procedure trigger_notice_ab();`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `insert into parted_constr values (1, 'aardvark');`,
			},
			{
				Statement: `insert into parted1_constr values (2, 'aardwolf');`,
			},
			{
				Statement: `insert into parted_constr_ancestor values (3, 'aasvogel');`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `set constraints parted_trig deferred;`,
			},
			{
				Statement: `insert into parted_constr values (1, 'aardvark');`,
			},
			{
				Statement: `insert into parted1_constr values (2, 'aardwolf'), (3, 'aasvogel');`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `drop table parted_constr_ancestor;`,
			},
			{
				Statement: `drop function bark(text);`,
			},
			{
				Statement: `create table parted_trigger (a int, b text) partition by range (a);`,
			},
			{
				Statement: `create table parted_trigger_1 partition of parted_trigger for values from (0) to (1000);`,
			},
			{
				Statement: `create table parted_trigger_2 (drp int, a int, b text);`,
			},
			{
				Statement: `alter table parted_trigger_2 drop column drp;`,
			},
			{
				Statement: `alter table parted_trigger attach partition parted_trigger_2 for values from (1000) to (2000);`,
			},
			{
				Statement: `create trigger parted_trigger after update on parted_trigger
  for each row when (new.a % 2 = 1 and length(old.b) >= 2) execute procedure trigger_notice_ab();`,
			},
			{
				Statement: `create table parted_trigger_3 (b text, a int) partition by range (length(b));`,
			},
			{
				Statement: `create table parted_trigger_3_1 partition of parted_trigger_3 for values from (1) to (3);`,
			},
			{
				Statement: `create table parted_trigger_3_2 partition of parted_trigger_3 for values from (3) to (5);`,
			},
			{
				Statement: `alter table parted_trigger attach partition parted_trigger_3 for values from (2000) to (3000);`,
			},
			{
				Statement: `insert into parted_trigger values
    (0, 'a'), (1, 'bbb'), (2, 'bcd'), (3, 'c'),
	(1000, 'c'), (1001, 'ddd'), (1002, 'efg'), (1003, 'f'),
	(2000, 'e'), (2001, 'fff'), (2002, 'ghi'), (2003, 'h');`,
			},
			{
				Statement: `update parted_trigger set a = a + 2; -- notice for odd 'a' values, long 'b' values`,
			},
			{
				Statement: `drop table parted_trigger;`,
			},
			{
				Statement: `create table parted_referenced (a int);`,
			},
			{
				Statement: `create table unparted_trigger (a int, b text);	-- for comparison purposes`,
			},
			{
				Statement: `create table parted_trigger (a int, b text) partition by range (a);`,
			},
			{
				Statement: `create table parted_trigger_1 partition of parted_trigger for values from (0) to (1000);`,
			},
			{
				Statement: `create table parted_trigger_2 (drp int, a int, b text);`,
			},
			{
				Statement: `alter table parted_trigger_2 drop column drp;`,
			},
			{
				Statement: `alter table parted_trigger attach partition parted_trigger_2 for values from (1000) to (2000);`,
			},
			{
				Statement: `create constraint trigger parted_trigger after update on parted_trigger
  from parted_referenced
  for each row execute procedure trigger_notice_ab();`,
			},
			{
				Statement: `create constraint trigger parted_trigger after update on unparted_trigger
  from parted_referenced
  for each row execute procedure trigger_notice_ab();`,
			},
			{
				Statement: `create table parted_trigger_3 (b text, a int) partition by range (length(b));`,
			},
			{
				Statement: `create table parted_trigger_3_1 partition of parted_trigger_3 for values from (1) to (3);`,
			},
			{
				Statement: `create table parted_trigger_3_2 partition of parted_trigger_3 for values from (3) to (5);`,
			},
			{
				Statement: `alter table parted_trigger attach partition parted_trigger_3 for values from (2000) to (3000);`,
			},
			{
				Statement: `select tgname, conname, t.tgrelid::regclass, t.tgconstrrelid::regclass,
  c.conrelid::regclass, c.confrelid::regclass
  from pg_trigger t join pg_constraint c on (t.tgconstraint = c.oid)
  where tgname = 'parted_trigger'
  order by t.tgrelid::regclass::text;`,
				Results: []sql.Row{{`parted_trigger`, `parted_trigger`, `parted_trigger`, `parted_referenced`, `parted_trigger`, `-`}, {`parted_trigger`, `parted_trigger`, `parted_trigger_1`, `parted_referenced`, `parted_trigger_1`, `-`}, {`parted_trigger`, `parted_trigger`, `parted_trigger_2`, `parted_referenced`, `parted_trigger_2`, `-`}, {`parted_trigger`, `parted_trigger`, `parted_trigger_3`, `parted_referenced`, `parted_trigger_3`, `-`}, {`parted_trigger`, `parted_trigger`, `parted_trigger_3_1`, `parted_referenced`, `parted_trigger_3_1`, `-`}, {`parted_trigger`, `parted_trigger`, `parted_trigger_3_2`, `parted_referenced`, `parted_trigger_3_2`, `-`}, {`parted_trigger`, `parted_trigger`, `unparted_trigger`, `parted_referenced`, `unparted_trigger`, `-`}},
			},
			{
				Statement: `drop table parted_referenced, parted_trigger, unparted_trigger;`,
			},
			{
				Statement: `create table parted_trigger (a int, b text) partition by range (a);`,
			},
			{
				Statement: `create table parted_trigger_1 partition of parted_trigger for values from (0) to (1000);`,
			},
			{
				Statement: `create table parted_trigger_2 (drp int, a int, b text);`,
			},
			{
				Statement: `alter table parted_trigger_2 drop column drp;`,
			},
			{
				Statement: `alter table parted_trigger attach partition parted_trigger_2 for values from (1000) to (2000);`,
			},
			{
				Statement: `create trigger parted_trigger after update of b on parted_trigger
  for each row execute procedure trigger_notice_ab();`,
			},
			{
				Statement: `create table parted_trigger_3 (b text, a int) partition by range (length(b));`,
			},
			{
				Statement: `create table parted_trigger_3_1 partition of parted_trigger_3 for values from (1) to (4);`,
			},
			{
				Statement: `create table parted_trigger_3_2 partition of parted_trigger_3 for values from (4) to (8);`,
			},
			{
				Statement: `alter table parted_trigger attach partition parted_trigger_3 for values from (2000) to (3000);`,
			},
			{
				Statement: `insert into parted_trigger values (0, 'a'), (1000, 'c'), (2000, 'e'), (2001, 'eeee');`,
			},
			{
				Statement: `update parted_trigger set a = a + 2;	-- no notices here`,
			},
			{
				Statement: `update parted_trigger set b = b || 'b';	-- all triggers should fire`,
			},
			{
				Statement: `drop table parted_trigger;`,
			},
			{
				Statement: `drop function trigger_notice_ab();`,
			},
			{
				Statement: `create table trg_clone (a int) partition by range (a);`,
			},
			{
				Statement: `create table trg_clone1 partition of trg_clone for values from (0) to (1000);`,
			},
			{
				Statement: `alter table trg_clone add constraint uniq unique (a) deferrable;`,
			},
			{
				Statement: `create table trg_clone2 partition of trg_clone for values from (1000) to (2000);`,
			},
			{
				Statement: `create table trg_clone3 partition of trg_clone for values from (2000) to (3000)
  partition by range (a);`,
			},
			{
				Statement: `create table trg_clone_3_3 partition of trg_clone3 for values from (2000) to (2100);`,
			},
			{
				Statement: `select tgrelid::regclass, count(*) from pg_trigger
  where tgrelid::regclass in ('trg_clone', 'trg_clone1', 'trg_clone2',
	'trg_clone3', 'trg_clone_3_3')
  group by tgrelid::regclass order by tgrelid::regclass;`,
				Results: []sql.Row{{`trg_clone`, 1}, {`trg_clone1`, 1}, {`trg_clone2`, 1}, {`trg_clone3`, 1}, {`trg_clone_3_3`, 1}},
			},
			{
				Statement: `drop table trg_clone;`,
			},
			{
				Statement: `create table parent (a int);`,
			},
			{
				Statement: `create table child1 () inherits (parent);`,
			},
			{
				Statement: `create function trig_nothing() returns trigger language plpgsql
  as $$ begin return null; end $$;`,
			},
			{
				Statement: `create trigger tg after insert on parent
  for each row execute function trig_nothing();`,
			},
			{
				Statement: `create trigger tg after insert on child1
  for each row execute function trig_nothing();`,
			},
			{
				Statement: `alter table parent disable trigger tg;`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgenabled from pg_trigger
  where tgrelid in ('parent'::regclass, 'child1'::regclass)
  order by tgrelid::regclass::text;`,
				Results: []sql.Row{{`child1`, `tg`, `O`}, {`parent`, `tg`, `D`}},
			},
			{
				Statement: `alter table only parent enable always trigger tg;`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgenabled from pg_trigger
  where tgrelid in ('parent'::regclass, 'child1'::regclass)
  order by tgrelid::regclass::text;`,
				Results: []sql.Row{{`child1`, `tg`, `O`}, {`parent`, `tg`, `A`}},
			},
			{
				Statement: `drop table parent, child1;`,
			},
			{
				Statement: `create table parent (a int) partition by list (a);`,
			},
			{
				Statement: `create table child1 partition of parent for values in (1);`,
			},
			{
				Statement: `create trigger tg after insert on parent
  for each row execute procedure trig_nothing();`,
			},
			{
				Statement: `create trigger tg_stmt after insert on parent
  for statement execute procedure trig_nothing();`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgenabled from pg_trigger
  where tgrelid in ('parent'::regclass, 'child1'::regclass)
  order by tgrelid::regclass::text, tgname;`,
				Results: []sql.Row{{`child1`, `tg`, `O`}, {`parent`, `tg`, `O`}, {`parent`, `tg_stmt`, `O`}},
			},
			{
				Statement: `alter table only parent enable always trigger tg;	-- no recursion because ONLY`,
			},
			{
				Statement: `alter table parent enable always trigger tg_stmt;	-- no recursion because statement trigger`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgenabled from pg_trigger
  where tgrelid in ('parent'::regclass, 'child1'::regclass)
  order by tgrelid::regclass::text, tgname;`,
				Results: []sql.Row{{`child1`, `tg`, `O`}, {`parent`, `tg`, `A`}, {`parent`, `tg_stmt`, `A`}},
			},
			{
				Statement: `alter table parent enable always trigger tg;`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgenabled from pg_trigger
  where tgrelid in ('parent'::regclass, 'child1'::regclass)
  order by tgrelid::regclass::text, tgname;`,
				Results: []sql.Row{{`child1`, `tg`, `A`}, {`parent`, `tg`, `A`}, {`parent`, `tg_stmt`, `A`}},
			},
			{
				Statement: `alter table parent disable trigger user;`,
			},
			{
				Statement: `select tgrelid::regclass, tgname, tgenabled from pg_trigger
  where tgrelid in ('parent'::regclass, 'child1'::regclass)
  order by tgrelid::regclass::text, tgname;`,
				Results: []sql.Row{{`child1`, `tg`, `D`}, {`parent`, `tg`, `D`}, {`parent`, `tg_stmt`, `D`}},
			},
			{
				Statement: `drop table parent, child1;`,
			},
			{
				Statement: `create table parent (a int primary key, f int references parent)
  partition by list (a);`,
			},
			{
				Statement: `create table child1 partition of parent for values in (1);`,
			},
			{
				Statement: `select tgrelid::regclass, rtrim(tgname, '0123456789') as tgname,
  tgfoid::regproc, tgenabled
  from pg_trigger where tgrelid in ('parent'::regclass, 'child1'::regclass)
  order by tgrelid::regclass::text, tgfoid;`,
				Results: []sql.Row{{`child1`, `RI_ConstraintTrigger_c_`, "RI_FKey_check_ins", `O`}, {`child1`, `RI_ConstraintTrigger_c_`, "RI_FKey_check_upd", `O`}, {`parent`, `RI_ConstraintTrigger_c_`, "RI_FKey_check_ins", `O`}, {`parent`, `RI_ConstraintTrigger_c_`, "RI_FKey_check_upd", `O`}, {`parent`, `RI_ConstraintTrigger_a_`, "RI_FKey_noaction_del", `O`}, {`parent`, `RI_ConstraintTrigger_a_`, "RI_FKey_noaction_upd", `O`}},
			},
			{
				Statement: `alter table parent disable trigger all;`,
			},
			{
				Statement: `select tgrelid::regclass, rtrim(tgname, '0123456789') as tgname,
  tgfoid::regproc, tgenabled
  from pg_trigger where tgrelid in ('parent'::regclass, 'child1'::regclass)
  order by tgrelid::regclass::text, tgfoid;`,
				Results: []sql.Row{{`child1`, `RI_ConstraintTrigger_c_`, "RI_FKey_check_ins", `D`}, {`child1`, `RI_ConstraintTrigger_c_`, "RI_FKey_check_upd", `D`}, {`parent`, `RI_ConstraintTrigger_c_`, "RI_FKey_check_ins", `D`}, {`parent`, `RI_ConstraintTrigger_c_`, "RI_FKey_check_upd", `D`}, {`parent`, `RI_ConstraintTrigger_a_`, "RI_FKey_noaction_del", `D`}, {`parent`, `RI_ConstraintTrigger_a_`, "RI_FKey_noaction_upd", `D`}},
			},
			{
				Statement: `drop table parent, child1;`,
			},
			{
				Statement: `CREATE TABLE trgfire (i int) PARTITION BY RANGE (i);`,
			},
			{
				Statement: `CREATE TABLE trgfire1 PARTITION OF trgfire FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION tgf() RETURNS trigger LANGUAGE plpgsql
  AS $$ begin raise exception 'except'; end $$;`,
			},
			{
				Statement: `CREATE TRIGGER tg AFTER INSERT ON trgfire FOR EACH ROW EXECUTE FUNCTION tgf();`,
			},
			{
				Statement:   `INSERT INTO trgfire VALUES (1);`,
				ErrorString: `except`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tgf() line 1 at RAISE
ALTER TABLE trgfire DISABLE TRIGGER tg;`,
			},
			{
				Statement: `INSERT INTO trgfire VALUES (1);`,
			},
			{
				Statement: `CREATE TABLE trgfire2 PARTITION OF trgfire FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `INSERT INTO trgfire VALUES (11);`,
			},
			{
				Statement: `CREATE TABLE trgfire3 (LIKE trgfire);`,
			},
			{
				Statement: `ALTER TABLE trgfire ATTACH PARTITION trgfire3 FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `INSERT INTO trgfire VALUES (21);`,
			},
			{
				Statement: `CREATE TABLE trgfire4 PARTITION OF trgfire FOR VALUES FROM (30) TO (40) PARTITION BY LIST (i);`,
			},
			{
				Statement: `CREATE TABLE trgfire4_30 PARTITION OF trgfire4 FOR VALUES IN (30);`,
			},
			{
				Statement: `INSERT INTO trgfire VALUES (30);`,
			},
			{
				Statement: `CREATE TABLE trgfire5 (LIKE trgfire) PARTITION BY LIST (i);`,
			},
			{
				Statement: `CREATE TABLE trgfire5_40 PARTITION OF trgfire5 FOR VALUES IN (40);`,
			},
			{
				Statement: `ALTER TABLE trgfire ATTACH PARTITION trgfire5 FOR VALUES FROM (40) TO (50);`,
			},
			{
				Statement: `INSERT INTO trgfire VALUES (40);`,
			},
			{
				Statement: `SELECT tgrelid::regclass, tgenabled FROM pg_trigger
  WHERE tgrelid::regclass IN (SELECT oid from pg_class where relname LIKE 'trgfire%')
  ORDER BY tgrelid::regclass::text;`,
				Results: []sql.Row{{`trgfire`, `D`}, {`trgfire1`, `D`}, {`trgfire2`, `D`}, {`trgfire3`, `D`}, {`trgfire4`, `D`}, {`trgfire4_30`, `D`}, {`trgfire5`, `D`}, {`trgfire5_40`, `D`}},
			},
			{
				Statement: `ALTER TABLE trgfire ENABLE TRIGGER tg;`,
			},
			{
				Statement:   `INSERT INTO trgfire VALUES (1);`,
				ErrorString: `except`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tgf() line 1 at RAISE
INSERT INTO trgfire VALUES (11);`,
				ErrorString: `except`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tgf() line 1 at RAISE
INSERT INTO trgfire VALUES (21);`,
				ErrorString: `except`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tgf() line 1 at RAISE
INSERT INTO trgfire VALUES (30);`,
				ErrorString: `except`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tgf() line 1 at RAISE
INSERT INTO trgfire VALUES (40);`,
				ErrorString: `except`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function tgf() line 1 at RAISE
DROP TABLE trgfire;`,
			},
			{
				Statement: `DROP FUNCTION tgf();`,
			},
			{
				Statement: `create or replace function dump_insert() returns trigger language plpgsql as
$$
  begin
    raise notice 'trigger = %, new table = %',
                 TG_NAME,
                 (select string_agg(new_table::text, ', ' order by a) from new_table);`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create or replace function dump_update() returns trigger language plpgsql as
$$
  begin
    raise notice 'trigger = %, old table = %, new table = %',
                 TG_NAME,
                 (select string_agg(old_table::text, ', ' order by a) from old_table),
                 (select string_agg(new_table::text, ', ' order by a) from new_table);`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create or replace function dump_delete() returns trigger language plpgsql as
$$
  begin
    raise notice 'trigger = %, old table = %',
                 TG_NAME,
                 (select string_agg(old_table::text, ', ' order by a) from old_table);`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create table parent (a text, b int) partition by list (a);`,
			},
			{
				Statement: `create table child1 partition of parent for values in ('AAA');`,
			},
			{
				Statement: `create table child2 (x int, a text, b int);`,
			},
			{
				Statement: `alter table child2 drop column x;`,
			},
			{
				Statement: `alter table parent attach partition child2 for values in ('BBB');`,
			},
			{
				Statement: `create table child3 (b int, a text);`,
			},
			{
				Statement: `alter table parent attach partition child3 for values in ('CCC');`,
			},
			{
				Statement: `create trigger parent_insert_trig
  after insert on parent referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger parent_update_trig
  after update on parent referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger parent_delete_trig
  after delete on parent referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `create trigger child1_insert_trig
  after insert on child1 referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger child1_update_trig
  after update on child1 referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger child1_delete_trig
  after delete on child1 referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `create trigger child2_insert_trig
  after insert on child2 referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger child2_update_trig
  after update on child2 referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger child2_delete_trig
  after delete on child2 referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `create trigger child3_insert_trig
  after insert on child3 referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger child3_update_trig
  after update on child3 referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger child3_delete_trig
  after delete on child3 referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `SELECT trigger_name, event_manipulation, event_object_schema, event_object_table,
       action_order, action_condition, action_orientation, action_timing,
       action_reference_old_table, action_reference_new_table
  FROM information_schema.triggers
  WHERE event_object_table IN ('parent', 'child1', 'child2', 'child3')
  ORDER BY trigger_name COLLATE "C", 2;`,
				Results: []sql.Row{{`child1_delete_trig`, `DELETE`, `public`, `child1`, 1, ``, `STATEMENT`, `AFTER`, `old_table`, ``}, {`child1_insert_trig`, `INSERT`, `public`, `child1`, 1, ``, `STATEMENT`, `AFTER`, ``, `new_table`}, {`child1_update_trig`, `UPDATE`, `public`, `child1`, 1, ``, `STATEMENT`, `AFTER`, `old_table`, `new_table`}, {`child2_delete_trig`, `DELETE`, `public`, `child2`, 1, ``, `STATEMENT`, `AFTER`, `old_table`, ``}, {`child2_insert_trig`, `INSERT`, `public`, `child2`, 1, ``, `STATEMENT`, `AFTER`, ``, `new_table`}, {`child2_update_trig`, `UPDATE`, `public`, `child2`, 1, ``, `STATEMENT`, `AFTER`, `old_table`, `new_table`}, {`child3_delete_trig`, `DELETE`, `public`, `child3`, 1, ``, `STATEMENT`, `AFTER`, `old_table`, ``}, {`child3_insert_trig`, `INSERT`, `public`, `child3`, 1, ``, `STATEMENT`, `AFTER`, ``, `new_table`}, {`child3_update_trig`, `UPDATE`, `public`, `child3`, 1, ``, `STATEMENT`, `AFTER`, `old_table`, `new_table`}, {`parent_delete_trig`, `DELETE`, `public`, `parent`, 1, ``, `STATEMENT`, `AFTER`, `old_table`, ``}, {`parent_insert_trig`, `INSERT`, `public`, `parent`, 1, ``, `STATEMENT`, `AFTER`, ``, `new_table`}, {`parent_update_trig`, `UPDATE`, `public`, `parent`, 1, ``, `STATEMENT`, `AFTER`, `old_table`, `new_table`}},
			},
			{
				Statement: `insert into child1 values ('AAA', 42);`,
			},
			{
				Statement: `insert into child2 values ('BBB', 42);`,
			},
			{
				Statement: `insert into child3 values (42, 'CCC');`,
			},
			{
				Statement: `update parent set b = b + 1;`,
			},
			{
				Statement: `delete from parent;`,
			},
			{
				Statement: `insert into parent values ('AAA', 42);`,
			},
			{
				Statement: `insert into parent values ('BBB', 42);`,
			},
			{
				Statement: `insert into parent values ('CCC', 42);`,
			},
			{
				Statement: `delete from child1;`,
			},
			{
				Statement: `delete from child2;`,
			},
			{
				Statement: `delete from child3;`,
			},
			{
				Statement: `copy parent (a, b) from stdin;`,
			},
			{
				Statement: `drop trigger child1_insert_trig on child1;`,
			},
			{
				Statement: `drop trigger child1_update_trig on child1;`,
			},
			{
				Statement: `drop trigger child1_delete_trig on child1;`,
			},
			{
				Statement: `drop trigger child2_insert_trig on child2;`,
			},
			{
				Statement: `drop trigger child2_update_trig on child2;`,
			},
			{
				Statement: `drop trigger child2_delete_trig on child2;`,
			},
			{
				Statement: `drop trigger child3_insert_trig on child3;`,
			},
			{
				Statement: `drop trigger child3_update_trig on child3;`,
			},
			{
				Statement: `drop trigger child3_delete_trig on child3;`,
			},
			{
				Statement: `delete from parent;`,
			},
			{
				Statement: `copy parent (a, b) from stdin;`,
			},
			{
				Statement: `create or replace function intercept_insert() returns trigger language plpgsql as
$$
  begin
    new.b = new.b + 1000;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger intercept_insert_child3
  before insert on child3
  for each row execute procedure intercept_insert();`,
			},
			{
				Statement: `insert into parent values ('AAA', 42), ('BBB', 42), ('CCC', 66);`,
			},
			{
				Statement: `copy parent (a, b) from stdin;`,
			},
			{
				Statement: `drop table child1, child2, child3, parent;`,
			},
			{
				Statement: `drop function intercept_insert();`,
			},
			{
				Statement: `create table parent (a text, b int) partition by list (a);`,
			},
			{
				Statement: `create table child partition of parent for values in ('AAA');`,
			},
			{
				Statement: `create trigger child_row_trig
  after insert on child referencing new table as new_table
  for each row execute procedure dump_insert();`,
				ErrorString: `ROW triggers with transition tables are not supported on partitions`,
			},
			{
				Statement: `alter table parent detach partition child;`,
			},
			{
				Statement: `create trigger child_row_trig
  after insert on child referencing new table as new_table
  for each row execute procedure dump_insert();`,
			},
			{
				Statement:   `alter table parent attach partition child for values in ('AAA');`,
				ErrorString: `trigger "child_row_trig" prevents table "child" from becoming a partition`,
			},
			{
				Statement: `drop trigger child_row_trig on child;`,
			},
			{
				Statement: `alter table parent attach partition child for values in ('AAA');`,
			},
			{
				Statement: `drop table child, parent;`,
			},
			{
				Statement: `create table parent (a text, b int);`,
			},
			{
				Statement: `create table child1 () inherits (parent);`,
			},
			{
				Statement: `create table child2 (b int, a text);`,
			},
			{
				Statement: `alter table child2 inherit parent;`,
			},
			{
				Statement: `create table child3 (c text) inherits (parent);`,
			},
			{
				Statement: `create trigger parent_insert_trig
  after insert on parent referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger parent_update_trig
  after update on parent referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger parent_delete_trig
  after delete on parent referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `create trigger child1_insert_trig
  after insert on child1 referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger child1_update_trig
  after update on child1 referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger child1_delete_trig
  after delete on child1 referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `create trigger child2_insert_trig
  after insert on child2 referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger child2_update_trig
  after update on child2 referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger child2_delete_trig
  after delete on child2 referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `create trigger child3_insert_trig
  after insert on child3 referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger child3_update_trig
  after update on child3 referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger child3_delete_trig
  after delete on child3 referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `insert into child1 values ('AAA', 42);`,
			},
			{
				Statement: `insert into child2 values (42, 'BBB');`,
			},
			{
				Statement: `insert into child3 values ('CCC', 42, 'foo');`,
			},
			{
				Statement: `update parent set b = b + 1;`,
			},
			{
				Statement: `delete from parent;`,
			},
			{
				Statement: `insert into child1 values ('AAA', 42);`,
			},
			{
				Statement: `insert into child2 values (42, 'BBB');`,
			},
			{
				Statement: `insert into child3 values ('CCC', 42, 'foo');`,
			},
			{
				Statement: `delete from child1;`,
			},
			{
				Statement: `delete from child2;`,
			},
			{
				Statement: `delete from child3;`,
			},
			{
				Statement: `copy parent (a, b) from stdin;`,
			},
			{
				Statement: `create index on parent(b);`,
			},
			{
				Statement: `copy parent (a, b) from stdin;`,
			},
			{
				Statement: `drop trigger child1_insert_trig on child1;`,
			},
			{
				Statement: `drop trigger child1_update_trig on child1;`,
			},
			{
				Statement: `drop trigger child1_delete_trig on child1;`,
			},
			{
				Statement: `drop trigger child2_insert_trig on child2;`,
			},
			{
				Statement: `drop trigger child2_update_trig on child2;`,
			},
			{
				Statement: `drop trigger child2_delete_trig on child2;`,
			},
			{
				Statement: `drop trigger child3_insert_trig on child3;`,
			},
			{
				Statement: `drop trigger child3_update_trig on child3;`,
			},
			{
				Statement: `drop trigger child3_delete_trig on child3;`,
			},
			{
				Statement: `delete from parent;`,
			},
			{
				Statement: `drop table child1, child2, child3, parent;`,
			},
			{
				Statement: `create table parent (a text, b int);`,
			},
			{
				Statement: `create table child () inherits (parent);`,
			},
			{
				Statement: `create trigger child_row_trig
  after insert on child referencing new table as new_table
  for each row execute procedure dump_insert();`,
				ErrorString: `ROW triggers with transition tables are not supported on inheritance children`,
			},
			{
				Statement: `alter table child no inherit parent;`,
			},
			{
				Statement: `create trigger child_row_trig
  after insert on child referencing new table as new_table
  for each row execute procedure dump_insert();`,
			},
			{
				Statement:   `alter table child inherit parent;`,
				ErrorString: `trigger "child_row_trig" prevents table "child" from becoming an inheritance child`,
			},
			{
				Statement: `drop trigger child_row_trig on child;`,
			},
			{
				Statement: `alter table child inherit parent;`,
			},
			{
				Statement: `drop table child, parent;`,
			},
			{
				Statement: `create table table1 (a int);`,
			},
			{
				Statement: `create table table2 (a text);`,
			},
			{
				Statement: `create trigger table1_trig
  after insert on table1 referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger table2_trig
  after insert on table2 referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `with wcte as (insert into table1 values (42))
  insert into table2 values ('hello world');`,
			},
			{
				Statement: `with wcte as (insert into table1 values (43))
  insert into table1 values (44);`,
			},
			{
				Statement: `select * from table1;`,
				Results:   []sql.Row{{42}, {44}, {43}},
			},
			{
				Statement: `select * from table2;`,
				Results:   []sql.Row{{`hello world`}},
			},
			{
				Statement: `drop table table1;`,
			},
			{
				Statement: `drop table table2;`,
			},
			{
				Statement: `create table my_table (a int primary key, b text);`,
			},
			{
				Statement: `create trigger my_table_insert_trig
  after insert on my_table referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger my_table_update_trig
  after update on my_table referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `insert into my_table values (1, 'AAA'), (2, 'BBB')
  on conflict (a) do
  update set b = my_table.b || ':' || excluded.b;`,
			},
			{
				Statement: `insert into my_table values (1, 'AAA'), (2, 'BBB'), (3, 'CCC'), (4, 'DDD')
  on conflict (a) do
  update set b = my_table.b || ':' || excluded.b;`,
			},
			{
				Statement: `insert into my_table values (3, 'CCC'), (4, 'DDD')
  on conflict (a) do
  update set b = my_table.b || ':' || excluded.b;`,
			},
			{
				Statement: `create table iocdu_tt_parted (a int primary key, b text) partition by list (a);`,
			},
			{
				Statement: `create table iocdu_tt_parted1 partition of iocdu_tt_parted for values in (1);`,
			},
			{
				Statement: `create table iocdu_tt_parted2 partition of iocdu_tt_parted for values in (2);`,
			},
			{
				Statement: `create table iocdu_tt_parted3 partition of iocdu_tt_parted for values in (3);`,
			},
			{
				Statement: `create table iocdu_tt_parted4 partition of iocdu_tt_parted for values in (4);`,
			},
			{
				Statement: `create trigger iocdu_tt_parted_insert_trig
  after insert on iocdu_tt_parted referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger iocdu_tt_parted_update_trig
  after update on iocdu_tt_parted referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `insert into iocdu_tt_parted values (1, 'AAA'), (2, 'BBB')
  on conflict (a) do
  update set b = iocdu_tt_parted.b || ':' || excluded.b;`,
			},
			{
				Statement: `insert into iocdu_tt_parted values (1, 'AAA'), (2, 'BBB'), (3, 'CCC'), (4, 'DDD')
  on conflict (a) do
  update set b = iocdu_tt_parted.b || ':' || excluded.b;`,
			},
			{
				Statement: `insert into iocdu_tt_parted values (3, 'CCC'), (4, 'DDD')
  on conflict (a) do
  update set b = iocdu_tt_parted.b || ':' || excluded.b;`,
			},
			{
				Statement: `drop table iocdu_tt_parted;`,
			},
			{
				Statement: `create trigger my_table_multievent_trig
  after insert or update on my_table referencing new table as new_table
  for each statement execute procedure dump_insert();`,
				ErrorString: `transition tables cannot be specified for triggers with more than one event`,
			},
			{
				Statement: `create trigger my_table_col_update_trig
  after update of b on my_table referencing new table as new_table
  for each statement execute procedure dump_insert();`,
				ErrorString: `transition tables cannot be specified for triggers with column lists`,
			},
			{
				Statement: `drop table my_table;`,
			},
			{
				Statement: `create table refd_table (a int primary key, b text);`,
			},
			{
				Statement: `create table trig_table (a int, b text,
  foreign key (a) references refd_table on update cascade on delete cascade
);`,
			},
			{
				Statement: `create trigger trig_table_before_trig
  before insert or update or delete on trig_table
  for each statement execute procedure trigger_func('trig_table');`,
			},
			{
				Statement: `create trigger trig_table_insert_trig
  after insert on trig_table referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger trig_table_update_trig
  after update on trig_table referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger trig_table_delete_trig
  after delete on trig_table referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `insert into refd_table values
  (1, 'one'),
  (2, 'two'),
  (3, 'three');`,
			},
			{
				Statement: `insert into trig_table values
  (1, 'one a'),
  (1, 'one b'),
  (2, 'two a'),
  (2, 'two b'),
  (3, 'three a'),
  (3, 'three b');`,
			},
			{
				Statement: `update refd_table set a = 11 where b = 'one';`,
			},
			{
				Statement: `select * from trig_table;`,
				Results:   []sql.Row{{2, `two a`}, {2, `two b`}, {3, `three a`}, {3, `three b`}, {11, `one a`}, {11, `one b`}},
			},
			{
				Statement: `delete from refd_table where length(b) = 3;`,
			},
			{
				Statement: `select * from trig_table;`,
				Results:   []sql.Row{{3, `three a`}, {3, `three b`}},
			},
			{
				Statement: `drop table refd_table, trig_table;`,
			},
			{
				Statement: `create table self_ref (a int primary key,
                       b int references self_ref(a) on delete cascade);`,
			},
			{
				Statement: `create trigger self_ref_before_trig
  before delete on self_ref
  for each statement execute procedure trigger_func('self_ref');`,
			},
			{
				Statement: `create trigger self_ref_r_trig
  after delete on self_ref referencing old table as old_table
  for each row execute procedure dump_delete();`,
			},
			{
				Statement: `create trigger self_ref_s_trig
  after delete on self_ref referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `insert into self_ref values (1, null), (2, 1), (3, 2);`,
			},
			{
				Statement: `delete from self_ref where a = 1;`,
			},
			{
				Statement: `drop trigger self_ref_r_trig on self_ref;`,
			},
			{
				Statement: `insert into self_ref values (1, null), (2, 1), (3, 2), (4, 3);`,
			},
			{
				Statement: `delete from self_ref where a = 1;`,
			},
			{
				Statement: `drop table self_ref;`,
			},
			{
				Statement: `create table merge_target_table (a int primary key, b text);`,
			},
			{
				Statement: `create trigger merge_target_table_insert_trig
  after insert on merge_target_table referencing new table as new_table
  for each statement execute procedure dump_insert();`,
			},
			{
				Statement: `create trigger merge_target_table_update_trig
  after update on merge_target_table referencing old table as old_table new table as new_table
  for each statement execute procedure dump_update();`,
			},
			{
				Statement: `create trigger merge_target_table_delete_trig
  after delete on merge_target_table referencing old table as old_table
  for each statement execute procedure dump_delete();`,
			},
			{
				Statement: `create table merge_source_table (a int, b text);`,
			},
			{
				Statement: `insert into merge_source_table
  values (1, 'initial1'), (2, 'initial2'),
		 (3, 'initial3'), (4, 'initial4');`,
			},
			{
				Statement: `merge into merge_target_table t
using merge_source_table s
on t.a = s.a
when not matched then
  insert values (a, b);`,
			},
			{
				Statement: `merge into merge_target_table t
using merge_source_table s
on t.a = s.a
when matched and s.a <= 2 then
	update set b = t.b || ' updated by merge'
when matched and s.a > 2 then
	delete
when not matched then
  insert values (a, b);`,
			},
			{
				Statement: `merge into merge_target_table t
using merge_source_table s
on t.a = s.a
when matched and s.a <= 2 then
	update set b = t.b || ' updated again by merge'
when matched and s.a > 2 then
	delete
when not matched then
  insert values (a, b);`,
			},
			{
				Statement: `drop table merge_source_table, merge_target_table;`,
			},
			{
				Statement: `drop function dump_insert();`,
			},
			{
				Statement: `drop function dump_update();`,
			},
			{
				Statement: `drop function dump_delete();`,
			},
			{
				Statement: `create table my_table (id integer);`,
			},
			{
				Statement: `create function funcA() returns trigger as $$
begin
  raise notice 'hello from funcA';`,
			},
			{
				Statement: `  return null;`,
			},
			{
				Statement: `end; $$ language plpgsql;`,
			},
			{
				Statement: `create function funcB() returns trigger as $$
begin
  raise notice 'hello from funcB';`,
			},
			{
				Statement: `  return null;`,
			},
			{
				Statement: `end; $$ language plpgsql;`,
			},
			{
				Statement: `create trigger my_trig
  after insert on my_table
  for each row execute procedure funcA();`,
			},
			{
				Statement: `create trigger my_trig
  before insert on my_table
  for each row execute procedure funcB();  -- should fail`,
				ErrorString: `trigger "my_trig" for relation "my_table" already exists`,
			},
			{
				Statement: `insert into my_table values (1);`,
			},
			{
				Statement: `create or replace trigger my_trig
  before insert on my_table
  for each row execute procedure funcB();  -- OK`,
			},
			{
				Statement: `insert into my_table values (2);  -- this insert should become a no-op`,
			},
			{
				Statement: `table my_table;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `drop table my_table;`,
			},
			{
				Statement: `create table parted_trig (a int) partition by range (a);`,
			},
			{
				Statement: `create table parted_trig_1 partition of parted_trig
       for values from (0) to (1000) partition by range (a);`,
			},
			{
				Statement: `create table parted_trig_1_1 partition of parted_trig_1 for values from (0) to (100);`,
			},
			{
				Statement: `create table parted_trig_2 partition of parted_trig for values from (1000) to (2000);`,
			},
			{
				Statement: `create table default_parted_trig partition of parted_trig default;`,
			},
			{
				Statement: `create or replace trigger my_trig
  after insert on parted_trig
  for each row execute procedure funcA();`,
			},
			{
				Statement: `insert into parted_trig (a) values (50);`,
			},
			{
				Statement: `create or replace trigger my_trig
  after insert on parted_trig
  for each row execute procedure funcB();`,
			},
			{
				Statement: `insert into parted_trig (a) values (50);`,
			},
			{
				Statement: `create or replace trigger my_trig
  after insert on parted_trig
  for each row execute procedure funcA();`,
			},
			{
				Statement: `insert into parted_trig (a) values (50);`,
			},
			{
				Statement: `create or replace trigger my_trig
  after insert on parted_trig_1
  for each row execute procedure funcB();  -- should fail`,
				ErrorString: `trigger "my_trig" for relation "parted_trig_1" is an internal or a child trigger`,
			},
			{
				Statement: `insert into parted_trig (a) values (50);`,
			},
			{
				Statement: `drop trigger my_trig on parted_trig;`,
			},
			{
				Statement: `insert into parted_trig (a) values (50);`,
			},
			{
				Statement: `create trigger my_trig
  after insert on parted_trig_1
  for each row execute procedure funcA();`,
			},
			{
				Statement: `insert into parted_trig (a) values (50);`,
			},
			{
				Statement: `create trigger my_trig
  after insert on parted_trig
  for each row execute procedure funcB();  -- should fail`,
				ErrorString: `trigger "my_trig" for relation "parted_trig_1" already exists`,
			},
			{
				Statement: `insert into parted_trig (a) values (50);`,
			},
			{
				Statement: `create or replace trigger my_trig
  after insert on parted_trig
  for each row execute procedure funcB();`,
			},
			{
				Statement: `insert into parted_trig (a) values (50);`,
			},
			{
				Statement: `drop table parted_trig;`,
			},
			{
				Statement: `drop function funcA();`,
			},
			{
				Statement: `drop function funcB();`,
			},
			{
				Statement: `create table trigger_parted (a int primary key) partition by list (a);`,
			},
			{
				Statement: `create function trigger_parted_trigfunc() returns trigger language plpgsql as
  $$ begin end; $$;`,
			},
			{
				Statement: `create trigger aft_row after insert or update on trigger_parted
  for each row execute function trigger_parted_trigfunc();`,
			},
			{
				Statement: `create table trigger_parted_p1 partition of trigger_parted for values in (1)
  partition by list (a);`,
			},
			{
				Statement: `create table trigger_parted_p1_1 partition of trigger_parted_p1 for values in (1);`,
			},
			{
				Statement: `create table trigger_parted_p2 partition of trigger_parted for values in (2)
  partition by list (a);`,
			},
			{
				Statement: `create table trigger_parted_p2_2 partition of trigger_parted_p2 for values in (2);`,
			},
			{
				Statement: `alter table only trigger_parted_p2 disable trigger aft_row;`,
			},
			{
				Statement: `alter table trigger_parted_p2_2 enable always trigger aft_row;`,
			},
			{
				Statement: `create table convslot_test_parent (col1 text primary key);`,
			},
			{
				Statement: `create table convslot_test_child (col1 text primary key,
	foreign key (col1) references convslot_test_parent(col1) on delete cascade on update cascade
);`,
			},
			{
				Statement: `alter table convslot_test_child add column col2 text not null default 'tutu';`,
			},
			{
				Statement: `insert into convslot_test_parent(col1) values ('1');`,
			},
			{
				Statement: `insert into convslot_test_child(col1) values ('1');`,
			},
			{
				Statement: `insert into convslot_test_parent(col1) values ('3');`,
			},
			{
				Statement: `insert into convslot_test_child(col1) values ('3');`,
			},
			{
				Statement: `create function convslot_trig1()
returns trigger
language plpgsql
AS $$
begin
raise notice 'trigger = %, old_table = %',
          TG_NAME,
          (select string_agg(old_table::text, ', ' order by col1) from old_table);`,
			},
			{
				Statement: `return null;`,
			},
			{
				Statement: `end; $$;`,
			},
			{
				Statement: `create function convslot_trig2()
returns trigger
language plpgsql
AS $$
begin
raise notice 'trigger = %, new table = %',
          TG_NAME,
          (select string_agg(new_table::text, ', ' order by col1) from new_table);`,
			},
			{
				Statement: `return null;`,
			},
			{
				Statement: `end; $$;`,
			},
			{
				Statement: `create trigger but_trigger after update on convslot_test_child
referencing new table as new_table
for each statement execute function convslot_trig2();`,
			},
			{
				Statement: `update convslot_test_parent set col1 = col1 || '1';`,
			},
			{
				Statement: `create function convslot_trig3()
returns trigger
language plpgsql
AS $$
begin
raise notice 'trigger = %, old_table = %, new table = %',
          TG_NAME,
          (select string_agg(old_table::text, ', ' order by col1) from old_table),
          (select string_agg(new_table::text, ', ' order by col1) from new_table);`,
			},
			{
				Statement: `return null;`,
			},
			{
				Statement: `end; $$;`,
			},
			{
				Statement: `create trigger but_trigger2 after update on convslot_test_child
referencing old table as old_table new table as new_table
for each statement execute function convslot_trig3();`,
			},
			{
				Statement: `update convslot_test_parent set col1 = col1 || '1';`,
			},
			{
				Statement: `create trigger bdt_trigger after delete on convslot_test_child
referencing old table as old_table
for each statement execute function convslot_trig1();`,
			},
			{
				Statement: `delete from convslot_test_parent;`,
			},
			{
				Statement: `drop table convslot_test_child, convslot_test_parent;`,
			},
			{
				Statement: `drop function convslot_trig1();`,
			},
			{
				Statement: `drop function convslot_trig2();`,
			},
			{
				Statement: `drop function convslot_trig3();`,
			},
			{
				Statement: `create table convslot_test_parent (id int primary key, val int)
partition by range (id);`,
			},
			{
				Statement: `create table convslot_test_part (val int, id int not null);`,
			},
			{
				Statement: `alter table convslot_test_parent
  attach partition convslot_test_part for values from (1) to (1000);`,
			},
			{
				Statement: `create function convslot_trig4() returns trigger as
$$begin raise exception 'BOOM!'; end$$ language plpgsql;`,
			},
			{
				Statement: `create trigger convslot_test_parent_update
    after update on convslot_test_parent
    referencing old table as old_rows new table as new_rows
    for each statement execute procedure convslot_trig4();`,
			},
			{
				Statement: `insert into convslot_test_parent (id, val) values (1, 2);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `savepoint svp;`,
			},
			{
				Statement:   `update convslot_test_parent set val = 3;  -- error expected`,
				ErrorString: `BOOM!`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function convslot_trig4() line 1 at RAISE
rollback to savepoint svp;`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `drop table convslot_test_parent;`,
			},
			{
				Statement: `drop function convslot_trig4();`,
			},
			{
				Statement: `create table grandparent (id int, primary key (id)) partition by range (id);`,
			},
			{
				Statement: `create table middle partition of grandparent for values from (1) to (10)
partition by range (id);`,
			},
			{
				Statement: `create table chi partition of middle for values from (1) to (5);`,
			},
			{
				Statement: `create table cho partition of middle for values from (6) to (10);`,
			},
			{
				Statement: `create function f () returns trigger as
$$ begin return new; end; $$
language plpgsql;`,
			},
			{
				Statement: `create trigger a after insert on grandparent
for each row execute procedure f();`,
			},
			{
				Statement: `alter trigger a on grandparent rename to b;`,
			},
			{
				Statement: `select tgrelid::regclass, tgname,
(select tgname from pg_trigger tr where tr.oid = pg_trigger.tgparentid) parent_tgname
from pg_trigger where tgrelid in (select relid from pg_partition_tree('grandparent'))
order by tgname, tgrelid::regclass::text COLLATE "C";`,
				Results: []sql.Row{{`chi`, `b`, `b`}, {`cho`, `b`, `b`}, {`grandparent`, `b`, ``}, {`middle`, `b`, `b`}},
			},
			{
				Statement:   `alter trigger a on only grandparent rename to b;	-- ONLY not supported`,
				ErrorString: `syntax error at or near "only"`,
			},
			{
				Statement:   `alter trigger b on middle rename to c;	-- can't rename trigger on partition`,
				ErrorString: `cannot rename trigger "b" on table "middle"`,
			},
			{
				Statement: `create trigger c after insert on middle
for each row execute procedure f();`,
			},
			{
				Statement:   `alter trigger b on grandparent rename to c;`,
				ErrorString: `trigger "c" for relation "middle" already exists`,
			},
			{
				Statement: `create trigger p after insert on grandparent for each statement execute function f();`,
			},
			{
				Statement: `create trigger p after insert on middle for each statement execute function f();`,
			},
			{
				Statement: `alter trigger p on grandparent rename to q;`,
			},
			{
				Statement: `select tgrelid::regclass, tgname,
(select tgname from pg_trigger tr where tr.oid = pg_trigger.tgparentid) parent_tgname
from pg_trigger where tgrelid in (select relid from pg_partition_tree('grandparent'))
order by tgname, tgrelid::regclass::text COLLATE "C";`,
				Results: []sql.Row{{`chi`, `b`, `b`}, {`cho`, `b`, `b`}, {`grandparent`, `b`, ``}, {`middle`, `b`, `b`}, {`chi`, `c`, `c`}, {`cho`, `c`, `c`}, {`middle`, `c`, ``}, {`middle`, `p`, ``}, {`grandparent`, `q`, ``}},
			},
			{
				Statement: `drop table grandparent;`,
			},
			{
				Statement: `create table parent (a int);`,
			},
			{
				Statement: `create table child () inherits (parent);`,
			},
			{
				Statement: `create trigger parenttrig after insert on parent
for each row execute procedure f();`,
			},
			{
				Statement: `create trigger parenttrig after insert on child
for each row execute procedure f();`,
			},
			{
				Statement: `alter trigger parenttrig on parent rename to anothertrig;`,
			},
			{
				Statement: `\d+ child
                                   Table "public.child"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
Triggers:
    parenttrig AFTER INSERT ON child FOR EACH ROW EXECUTE FUNCTION f()
Inherits: parent
drop table parent, child;`,
			},
			{
				Statement: `drop function f();`,
			},
		},
	})
}
