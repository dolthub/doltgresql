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

func TestEventTrigger(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_event_trigger)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_event_trigger,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create event trigger regress_event_trigger
   on ddl_command_start
   execute procedure pg_backend_pid();`,
				ErrorString: `function pg_backend_pid must return type event_trigger`,
			},
			{
				Statement: `create function test_event_trigger() returns event_trigger as $$
BEGIN
    RAISE NOTICE 'test_event_trigger: % %', tg_event, tg_tag;`,
			},
			{
				Statement: `END
$$ language plpgsql;`,
			},
			{
				Statement:   `SELECT test_event_trigger();`,
				ErrorString: `trigger functions can only be called as triggers`,
			},
			{
				Statement: `CONTEXT:  compilation of PL/pgSQL function "test_event_trigger" near line 1
create function test_event_trigger_arg(name text)
returns event_trigger as $$ BEGIN RETURN 1; END $$ language plpgsql;`,
				ErrorString: `event trigger functions cannot have declared arguments`,
			},
			{
				Statement: `CONTEXT:  compilation of PL/pgSQL function "test_event_trigger_arg" near line 1
create function test_event_trigger_sql() returns event_trigger as $$
SELECT 1 $$ language sql;`,
				ErrorString: `SQL functions cannot return type event_trigger`,
			},
			{
				Statement: `create event trigger regress_event_trigger on elephant_bootstrap
   execute procedure test_event_trigger();`,
				ErrorString: `unrecognized event name "elephant_bootstrap"`,
			},
			{
				Statement: `create event trigger regress_event_trigger on ddl_command_start
   execute procedure test_event_trigger();`,
			},
			{
				Statement: `create event trigger regress_event_trigger_end on ddl_command_end
   execute function test_event_trigger();`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when food in ('sandwich')
   execute procedure test_event_trigger();`,
				ErrorString: `unrecognized filter variable "food"`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when tag in ('sandwich')
   execute procedure test_event_trigger();`,
				ErrorString: `filter value "sandwich" not recognized for filter variable "tag"`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when tag in ('create table', 'create skunkcabbage')
   execute procedure test_event_trigger();`,
				ErrorString: `filter value "create skunkcabbage" not recognized for filter variable "tag"`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when tag in ('DROP EVENT TRIGGER')
   execute procedure test_event_trigger();`,
				ErrorString: `event triggers are not supported for DROP EVENT TRIGGER`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when tag in ('CREATE ROLE')
   execute procedure test_event_trigger();`,
				ErrorString: `event triggers are not supported for CREATE ROLE`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when tag in ('CREATE DATABASE')
   execute procedure test_event_trigger();`,
				ErrorString: `event triggers are not supported for CREATE DATABASE`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when tag in ('CREATE TABLESPACE')
   execute procedure test_event_trigger();`,
				ErrorString: `event triggers are not supported for CREATE TABLESPACE`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when tag in ('create table') and tag in ('CREATE FUNCTION')
   execute procedure test_event_trigger();`,
				ErrorString: `filter variable "tag" specified more than once`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   execute procedure test_event_trigger('argument not allowed');`,
				ErrorString: `syntax error at or near "'argument not allowed'"`,
			},
			{
				Statement: `create event trigger regress_event_trigger2 on ddl_command_start
   when tag in ('create table', 'CREATE FUNCTION')
   execute procedure test_event_trigger();`,
			},
			{
				Statement: `comment on event trigger regress_event_trigger is 'test comment';`,
			},
			{
				Statement: `create role regress_evt_user;`,
			},
			{
				Statement: `set role regress_evt_user;`,
			},
			{
				Statement: `create event trigger regress_event_trigger_noperms on ddl_command_start
   execute procedure test_event_trigger();`,
				ErrorString: `permission denied to create event trigger "regress_event_trigger_noperms"`,
			},
			{
				Statement: `reset role;`,
			},
			{
				Statement: `alter event trigger regress_event_trigger disable;`,
			},
			{
				Statement: `create table event_trigger_fire1 (a int);`,
			},
			{
				Statement: `alter event trigger regress_event_trigger enable;`,
			},
			{
				Statement: `set session_replication_role = replica;`,
			},
			{
				Statement: `create table event_trigger_fire2 (a int);`,
			},
			{
				Statement: `alter event trigger regress_event_trigger enable replica;`,
			},
			{
				Statement: `create table event_trigger_fire3 (a int);`,
			},
			{
				Statement: `alter event trigger regress_event_trigger enable always;`,
			},
			{
				Statement: `create table event_trigger_fire4 (a int);`,
			},
			{
				Statement: `reset session_replication_role;`,
			},
			{
				Statement: `create table event_trigger_fire5 (a int);`,
			},
			{
				Statement: `create function f1() returns int
language plpgsql
as $$
begin
  create table event_trigger_fire6 (a int);`,
			},
			{
				Statement: `  return 0;`,
			},
			{
				Statement: `end $$;`,
			},
			{
				Statement: `select f1();`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `create procedure p1()
language plpgsql
as $$
begin
  create table event_trigger_fire7 (a int);`,
			},
			{
				Statement: `end $$;`,
			},
			{
				Statement: `call p1();`,
			},
			{
				Statement: `alter event trigger regress_event_trigger disable;`,
			},
			{
				Statement: `drop table event_trigger_fire2, event_trigger_fire3, event_trigger_fire4, event_trigger_fire5, event_trigger_fire6, event_trigger_fire7;`,
			},
			{
				Statement: `drop routine f1(), p1();`,
			},
			{
				Statement: `grant all on table event_trigger_fire1 to public;`,
			},
			{
				Statement: `comment on table event_trigger_fire1 is 'here is a comment';`,
			},
			{
				Statement: `revoke all on table event_trigger_fire1 from public;`,
			},
			{
				Statement: `drop table event_trigger_fire1;`,
			},
			{
				Statement: `create foreign data wrapper useless;`,
			},
			{
				Statement: `create server useless_server foreign data wrapper useless;`,
			},
			{
				Statement: `create user mapping for regress_evt_user server useless_server;`,
			},
			{
				Statement: `alter default privileges for role regress_evt_user
 revoke delete on tables from regress_evt_user;`,
			},
			{
				Statement:   `alter event trigger regress_event_trigger owner to regress_evt_user;`,
				ErrorString: `permission denied to change owner of event trigger "regress_event_trigger"`,
			},
			{
				Statement: `alter role regress_evt_user superuser;`,
			},
			{
				Statement: `alter event trigger regress_event_trigger owner to regress_evt_user;`,
			},
			{
				Statement:   `alter event trigger regress_event_trigger rename to regress_event_trigger2;`,
				ErrorString: `event trigger "regress_event_trigger2" already exists`,
			},
			{
				Statement: `alter event trigger regress_event_trigger rename to regress_event_trigger3;`,
			},
			{
				Statement:   `drop event trigger regress_event_trigger;`,
				ErrorString: `event trigger "regress_event_trigger" does not exist`,
			},
			{
				Statement:   `drop role regress_evt_user;`,
				ErrorString: `role "regress_evt_user" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `owner of user mapping for regress_evt_user on server useless_server
owner of default privileges on new relations belonging to role regress_evt_user
drop event trigger if exists regress_event_trigger2;`,
			},
			{
				Statement: `drop event trigger if exists regress_event_trigger2;`,
			},
			{
				Statement: `drop event trigger regress_event_trigger3;`,
			},
			{
				Statement: `drop event trigger regress_event_trigger_end;`,
			},
			{
				Statement: `CREATE SCHEMA schema_one authorization regress_evt_user;`,
			},
			{
				Statement: `CREATE SCHEMA schema_two authorization regress_evt_user;`,
			},
			{
				Statement: `CREATE SCHEMA audit_tbls authorization regress_evt_user;`,
			},
			{
				Statement: `CREATE TEMP TABLE a_temp_tbl ();`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_evt_user;`,
			},
			{
				Statement: `CREATE TABLE schema_one.table_one(a int);`,
			},
			{
				Statement: `CREATE TABLE schema_one."table two"(a int);`,
			},
			{
				Statement: `CREATE TABLE schema_one.table_three(a int);`,
			},
			{
				Statement: `CREATE TABLE audit_tbls.schema_one_table_two(the_value text);`,
			},
			{
				Statement: `CREATE TABLE schema_two.table_two(a int);`,
			},
			{
				Statement: `CREATE TABLE schema_two.table_three(a int, b text);`,
			},
			{
				Statement: `CREATE TABLE audit_tbls.schema_two_table_three(the_value text);`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION schema_two.add(int, int) RETURNS int LANGUAGE plpgsql
  CALLED ON NULL INPUT
  AS $$ BEGIN RETURN coalesce($1,0) + coalesce($2,0); END; $$;`,
			},
			{
				Statement: `CREATE AGGREGATE schema_two.newton
  (BASETYPE = int, SFUNC = schema_two.add, STYPE = int);`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `CREATE TABLE undroppable_objs (
	object_type text,
	object_identity text
);`,
			},
			{
				Statement: `INSERT INTO undroppable_objs VALUES
('table', 'schema_one.table_three'),
('table', 'audit_tbls.schema_two_table_three');`,
			},
			{
				Statement: `CREATE TABLE dropped_objects (
	type text,
	schema text,
	object text
);`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION undroppable() RETURNS event_trigger
LANGUAGE plpgsql AS $$
DECLARE
	obj record;`,
			},
			{
				Statement: `BEGIN
	PERFORM 1 FROM pg_tables WHERE tablename = 'undroppable_objs';`,
			},
			{
				Statement: `	IF NOT FOUND THEN
		RAISE NOTICE 'table undroppable_objs not found, skipping';`,
			},
			{
				Statement: `		RETURN;`,
			},
			{
				Statement: `	END IF;`,
			},
			{
				Statement: `	FOR obj IN
		SELECT * FROM pg_event_trigger_dropped_objects() JOIN
			undroppable_objs USING (object_type, object_identity)
	LOOP
		RAISE EXCEPTION 'object % of type % cannot be dropped',
			obj.object_identity, obj.object_type;`,
			},
			{
				Statement: `	END LOOP;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE EVENT TRIGGER undroppable ON sql_drop
	EXECUTE PROCEDURE undroppable();`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION test_evtrig_dropped_objects() RETURNS event_trigger
LANGUAGE plpgsql AS $$
DECLARE
    obj record;`,
			},
			{
				Statement: `BEGIN
    FOR obj IN SELECT * FROM pg_event_trigger_dropped_objects()
    LOOP
        IF obj.object_type = 'table' THEN
                EXECUTE format('DROP TABLE IF EXISTS audit_tbls.%I',
					format('%s_%s', obj.schema_name, obj.object_name));`,
			},
			{
				Statement: `        END IF;`,
			},
			{
				Statement: `	INSERT INTO dropped_objects
		(type, schema, object) VALUES
		(obj.object_type, obj.schema_name, obj.object_identity);`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement: `END
$$;`,
			},
			{
				Statement: `CREATE EVENT TRIGGER regress_event_trigger_drop_objects ON sql_drop
	WHEN TAG IN ('drop table', 'drop function', 'drop view',
		'drop owned', 'drop schema', 'alter table')
	EXECUTE PROCEDURE test_evtrig_dropped_objects();`,
			},
			{
				Statement: `ALTER TABLE schema_one.table_one DROP COLUMN a;`,
			},
			{
				Statement:   `DROP SCHEMA schema_one, schema_two CASCADE;`,
				ErrorString: `object audit_tbls.schema_two_table_three of type table cannot be dropped`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function undroppable() line 14 at RAISE
SQL statement "DROP TABLE IF EXISTS audit_tbls.schema_two_table_three"
PL/pgSQL function test_evtrig_dropped_objects() line 8 at EXECUTE
DELETE FROM undroppable_objs WHERE object_identity = 'audit_tbls.schema_two_table_three';`,
			},
			{
				Statement:   `DROP SCHEMA schema_one, schema_two CASCADE;`,
				ErrorString: `object schema_one.table_three of type table cannot be dropped`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function undroppable() line 14 at RAISE
DELETE FROM undroppable_objs WHERE object_identity = 'schema_one.table_three';`,
			},
			{
				Statement: `DROP SCHEMA schema_one, schema_two CASCADE;`,
			},
			{
				Statement: `SELECT * FROM dropped_objects WHERE schema IS NULL OR schema <> 'pg_toast';`,
				Results:   []sql.Row{{`table column`, `schema_one`, `schema_one.table_one.a`}, {`schema`, ``, `schema_two`}, {`table`, `schema_two`, `schema_two.table_two`}, {`type`, `schema_two`, `schema_two.table_two`}, {`type`, `schema_two`, `schema_two.table_two[]`}, {`table`, `audit_tbls`, `audit_tbls.schema_two_table_three`}, {`type`, `audit_tbls`, `audit_tbls.schema_two_table_three`}, {`type`, `audit_tbls`, `audit_tbls.schema_two_table_three[]`}, {`table`, `schema_two`, `schema_two.table_three`}, {`type`, `schema_two`, `schema_two.table_three`}, {`type`, `schema_two`, `schema_two.table_three[]`}, {`function`, `schema_two`, `schema_two.add(integer,integer)`}, {`aggregate`, `schema_two`, `schema_two.newton(integer)`}, {`schema`, ``, `schema_one`}, {`table`, `schema_one`, `schema_one.table_one`}, {`type`, `schema_one`, `schema_one.table_one`}, {`type`, `schema_one`, `schema_one.table_one[]`}, {`table`, `schema_one`, `schema_one."table two"`}, {`type`, `schema_one`, `schema_one."table two"`}, {`type`, `schema_one`, `schema_one."table two"[]`}, {`table`, `schema_one`, `schema_one.table_three`}, {`type`, `schema_one`, `schema_one.table_three`}, {`type`, `schema_one`, `schema_one.table_three[]`}},
			},
			{
				Statement: `DROP OWNED BY regress_evt_user;`,
			},
			{
				Statement: `SELECT * FROM dropped_objects WHERE type = 'schema';`,
				Results:   []sql.Row{{`schema`, ``, `schema_two`}, {`schema`, ``, `schema_one`}, {`schema`, ``, `audit_tbls`}},
			},
			{
				Statement: `DROP ROLE regress_evt_user;`,
			},
			{
				Statement: `DROP EVENT TRIGGER regress_event_trigger_drop_objects;`,
			},
			{
				Statement: `DROP EVENT TRIGGER undroppable;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION event_trigger_report_dropped()
 RETURNS event_trigger
 LANGUAGE plpgsql
AS $$
DECLARE r record;`,
			},
			{
				Statement: `BEGIN
    FOR r IN SELECT * from pg_event_trigger_dropped_objects()
    LOOP
    IF NOT r.normal AND NOT r.original THEN
        CONTINUE;`,
			},
			{
				Statement: `    END IF;`,
			},
			{
				Statement: `    RAISE NOTICE 'NORMAL: orig=% normal=% istemp=% type=% identity=% name=% args=%',
        r.original, r.normal, r.is_temporary, r.object_type,
        r.object_identity, r.address_names, r.address_args;`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement: `END; $$;`,
			},
			{
				Statement: `CREATE EVENT TRIGGER regress_event_trigger_report_dropped ON sql_drop
    EXECUTE PROCEDURE event_trigger_report_dropped();`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION event_trigger_report_end()
 RETURNS event_trigger
 LANGUAGE plpgsql
AS $$
DECLARE r RECORD;`,
			},
			{
				Statement: `BEGIN
    FOR r IN SELECT * FROM pg_event_trigger_ddl_commands()
    LOOP
        RAISE NOTICE 'END: command_tag=% type=% identity=%',
            r.command_tag, r.object_type, r.object_identity;`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement: `END; $$;`,
			},
			{
				Statement: `CREATE EVENT TRIGGER regress_event_trigger_report_end ON ddl_command_end
  EXECUTE PROCEDURE event_trigger_report_end();`,
			},
			{
				Statement: `CREATE SCHEMA evttrig
	CREATE TABLE one (col_a SERIAL PRIMARY KEY, col_b text DEFAULT 'forty two', col_c SERIAL)
	CREATE INDEX one_idx ON one (col_b)
	CREATE TABLE two (col_c INTEGER CHECK (col_c > 0) REFERENCES one DEFAULT 42)
	CREATE TABLE id (col_d int NOT NULL GENERATED ALWAYS AS IDENTITY);`,
			},
			{
				Statement: `CREATE TABLE evttrig.parted (
    id int PRIMARY KEY)
    PARTITION BY RANGE (id);`,
			},
			{
				Statement: `CREATE TABLE evttrig.part_1_10 PARTITION OF evttrig.parted (id)
  FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `CREATE TABLE evttrig.part_10_20 PARTITION OF evttrig.parted (id)
  FOR VALUES FROM (10) TO (20) PARTITION BY RANGE (id);`,
			},
			{
				Statement: `CREATE TABLE evttrig.part_10_15 PARTITION OF evttrig.part_10_20 (id)
  FOR VALUES FROM (10) TO (15);`,
			},
			{
				Statement: `CREATE TABLE evttrig.part_15_20 PARTITION OF evttrig.part_10_20 (id)
  FOR VALUES FROM (15) TO (20);`,
			},
			{
				Statement: `ALTER TABLE evttrig.two DROP COLUMN col_c;`,
			},
			{
				Statement: `ALTER TABLE evttrig.one ALTER COLUMN col_b DROP DEFAULT;`,
			},
			{
				Statement: `ALTER TABLE evttrig.one DROP CONSTRAINT one_pkey;`,
			},
			{
				Statement: `ALTER TABLE evttrig.one DROP COLUMN col_c;`,
			},
			{
				Statement: `ALTER TABLE evttrig.id ALTER COLUMN col_d SET DATA TYPE bigint;`,
			},
			{
				Statement: `ALTER TABLE evttrig.id ALTER COLUMN col_d DROP IDENTITY,
  ALTER COLUMN col_d SET DATA TYPE int;`,
			},
			{
				Statement: `DROP INDEX evttrig.one_idx;`,
			},
			{
				Statement: `DROP SCHEMA evttrig CASCADE;`,
			},
			{
				Statement: `DROP TABLE a_temp_tbl;`,
			},
			{
				Statement: `CREATE OPERATOR CLASS evttrigopclass FOR TYPE int USING btree AS STORAGE int;`,
			},
			{
				Statement: `DROP EVENT TRIGGER regress_event_trigger_report_dropped;`,
			},
			{
				Statement: `DROP EVENT TRIGGER regress_event_trigger_report_end;`,
			},
			{
				Statement:   `select pg_event_trigger_table_rewrite_oid();`,
				ErrorString: `pg_event_trigger_table_rewrite_oid() can only be called in a table_rewrite event trigger function`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION test_evtrig_no_rewrite() RETURNS event_trigger
LANGUAGE plpgsql AS $$
BEGIN
  RAISE EXCEPTION 'rewrites not allowed';`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create event trigger no_rewrite_allowed on table_rewrite
  execute procedure test_evtrig_no_rewrite();`,
			},
			{
				Statement: `create table rewriteme (id serial primary key, foo float, bar timestamptz);`,
			},
			{
				Statement: `insert into rewriteme
     select x * 1.001 from generate_series(1, 500) as t(x);`,
			},
			{
				Statement:   `alter table rewriteme alter column foo type numeric;`,
				ErrorString: `rewrites not allowed`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function test_evtrig_no_rewrite() line 3 at RAISE
alter table rewriteme add column baz int default 0;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION test_evtrig_no_rewrite() RETURNS event_trigger
LANGUAGE plpgsql AS $$
BEGIN
  RAISE NOTICE 'Table ''%'' is being rewritten (reason = %)',
               pg_event_trigger_table_rewrite_oid()::regclass,
               pg_event_trigger_table_rewrite_reason();`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `alter table rewriteme
 add column onemore int default 0,
 add column another int default -1,
 alter column foo type numeric(10,4);`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW heapmv USING heap AS SELECT 1 AS a;`,
			},
			{
				Statement: `ALTER MATERIALIZED VIEW heapmv SET ACCESS METHOD heap2;`,
			},
			{
				Statement: `DROP MATERIALIZED VIEW heapmv;`,
			},
			{
				Statement: `alter table rewriteme alter column foo type numeric(12,4);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `set timezone to 'UTC';`,
			},
			{
				Statement: `alter table rewriteme alter column bar type timestamp;`,
			},
			{
				Statement: `set timezone to '0';`,
			},
			{
				Statement: `alter table rewriteme alter column bar type timestamptz;`,
			},
			{
				Statement: `set timezone to 'Europe/London';`,
			},
			{
				Statement: `alter table rewriteme alter column bar type timestamp; -- does rewrite`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION test_evtrig_no_rewrite() RETURNS event_trigger
LANGUAGE plpgsql AS $$
BEGIN
  RAISE NOTICE 'Table is being rewritten (reason = %)',
               pg_event_trigger_table_rewrite_reason();`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create type rewritetype as (a int);`,
			},
			{
				Statement: `create table rewritemetoo1 of rewritetype;`,
			},
			{
				Statement: `create table rewritemetoo2 of rewritetype;`,
			},
			{
				Statement: `alter type rewritetype alter attribute a type text cascade;`,
			},
			{
				Statement: `create table rewritemetoo3 (a rewritetype);`,
			},
			{
				Statement:   `alter type rewritetype alter attribute a type varchar cascade;`,
				ErrorString: `cannot alter type "rewritetype" because column "rewritemetoo3.a" uses it`,
			},
			{
				Statement: `drop table rewriteme;`,
			},
			{
				Statement: `drop event trigger no_rewrite_allowed;`,
			},
			{
				Statement: `drop function test_evtrig_no_rewrite();`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `CREATE TABLE event_trigger_test (a integer, b text);`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION start_command()
RETURNS event_trigger AS $$
BEGIN
RAISE NOTICE '% - ddl_command_start', tg_tag;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION end_command()
RETURNS event_trigger AS $$
BEGIN
RAISE NOTICE '% - ddl_command_end', tg_tag;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION drop_sql_command()
RETURNS event_trigger AS $$
BEGIN
RAISE NOTICE '% - sql_drop', tg_tag;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE EVENT TRIGGER start_rls_command ON ddl_command_start
    WHEN TAG IN ('CREATE POLICY', 'ALTER POLICY', 'DROP POLICY') EXECUTE PROCEDURE start_command();`,
			},
			{
				Statement: `CREATE EVENT TRIGGER end_rls_command ON ddl_command_end
    WHEN TAG IN ('CREATE POLICY', 'ALTER POLICY', 'DROP POLICY') EXECUTE PROCEDURE end_command();`,
			},
			{
				Statement: `CREATE EVENT TRIGGER sql_drop_command ON sql_drop
    WHEN TAG IN ('DROP POLICY') EXECUTE PROCEDURE drop_sql_command();`,
			},
			{
				Statement: `CREATE POLICY p1 ON event_trigger_test USING (FALSE);`,
			},
			{
				Statement: `ALTER POLICY p1 ON event_trigger_test USING (TRUE);`,
			},
			{
				Statement: `ALTER POLICY p1 ON event_trigger_test RENAME TO p2;`,
			},
			{
				Statement: `DROP POLICY p2 ON event_trigger_test;`,
			},
			{
				Statement: `SELECT
    e.evtname,
    pg_describe_object('pg_event_trigger'::regclass, e.oid, 0) as descr,
    b.type, b.object_names, b.object_args,
    pg_identify_object(a.classid, a.objid, a.objsubid) as ident
  FROM pg_event_trigger as e,
    LATERAL pg_identify_object_as_address('pg_event_trigger'::regclass, e.oid, 0) as b,
    LATERAL pg_get_object_address(b.type, b.object_names, b.object_args) as a
  ORDER BY e.evtname;`,
				Results: []sql.Row{{`end_rls_command`, `event trigger end_rls_command`, `event trigger`, `{end_rls_command}`, `{}`, `("event trigger",,end_rls_command,end_rls_command)`}, {`sql_drop_command`, `event trigger sql_drop_command`, `event trigger`, `{sql_drop_command}`, `{}`, `("event trigger",,sql_drop_command,sql_drop_command)`}, {`start_rls_command`, `event trigger start_rls_command`, `event trigger`, `{start_rls_command}`, `{}`, `("event trigger",,start_rls_command,start_rls_command)`}},
			},
			{
				Statement: `DROP EVENT TRIGGER start_rls_command;`,
			},
			{
				Statement: `DROP EVENT TRIGGER end_rls_command;`,
			},
			{
				Statement: `DROP EVENT TRIGGER sql_drop_command;`,
			},
		},
	})
}
