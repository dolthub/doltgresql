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

func TestObjectAddress(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_object_address)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_object_address,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET client_min_messages TO 'warning';`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_addr_user;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE USER regress_addr_user;`,
			},
			{
				Statement: `CREATE SCHEMA addr_nsp;`,
			},
			{
				Statement: `SET search_path TO 'addr_nsp';`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER addr_fdw;`,
			},
			{
				Statement: `CREATE SERVER addr_fserv FOREIGN DATA WRAPPER addr_fdw;`,
			},
			{
				Statement: `CREATE TEXT SEARCH DICTIONARY addr_ts_dict (template=simple);`,
			},
			{
				Statement: `CREATE TEXT SEARCH CONFIGURATION addr_ts_conf (copy=english);`,
			},
			{
				Statement: `CREATE TEXT SEARCH TEMPLATE addr_ts_temp (lexize=dsimple_lexize);`,
			},
			{
				Statement: `CREATE TEXT SEARCH PARSER addr_ts_prs
    (start = prsd_start, gettoken = prsd_nexttoken, end = prsd_end, lextypes = prsd_lextype);`,
			},
			{
				Statement: `CREATE TABLE addr_nsp.gentable (
	a serial primary key CONSTRAINT a_chk CHECK (a > 0),
	b text DEFAULT 'hello');`,
			},
			{
				Statement: `CREATE TABLE addr_nsp.parttable (
	a int PRIMARY KEY
) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE VIEW addr_nsp.genview AS SELECT * from addr_nsp.gentable;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW addr_nsp.genmatview AS SELECT * FROM addr_nsp.gentable;`,
			},
			{
				Statement: `CREATE TYPE addr_nsp.gencomptype AS (a int);`,
			},
			{
				Statement: `CREATE TYPE addr_nsp.genenum AS ENUM ('one', 'two');`,
			},
			{
				Statement: `CREATE FOREIGN TABLE addr_nsp.genftable (a int) SERVER addr_fserv;`,
			},
			{
				Statement: `CREATE AGGREGATE addr_nsp.genaggr(int4) (sfunc = int4pl, stype = int4);`,
			},
			{
				Statement: `CREATE DOMAIN addr_nsp.gendomain AS int4 CONSTRAINT domconstr CHECK (value > 0);`,
			},
			{
				Statement: `CREATE FUNCTION addr_nsp.trig() RETURNS TRIGGER LANGUAGE plpgsql AS $$ BEGIN END; $$;`,
			},
			{
				Statement: `CREATE TRIGGER t BEFORE INSERT ON addr_nsp.gentable FOR EACH ROW EXECUTE PROCEDURE addr_nsp.trig();`,
			},
			{
				Statement: `CREATE POLICY genpol ON addr_nsp.gentable;`,
			},
			{
				Statement: `CREATE PROCEDURE addr_nsp.proc(int4) LANGUAGE SQL AS $$ $$;`,
			},
			{
				Statement: `CREATE SERVER "integer" FOREIGN DATA WRAPPER addr_fdw;`,
			},
			{
				Statement: `CREATE USER MAPPING FOR regress_addr_user SERVER "integer";`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_addr_user IN SCHEMA public GRANT ALL ON TABLES TO regress_addr_user;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_addr_user REVOKE DELETE ON TABLES FROM regress_addr_user;`,
			},
			{
				Statement: `CREATE TRANSFORM FOR int LANGUAGE SQL (
	FROM SQL WITH FUNCTION prsd_lextype(internal),
	TO SQL WITH FUNCTION int4recv(internal));`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION addr_pub FOR TABLE addr_nsp.gentable;`,
			},
			{
				Statement: `CREATE PUBLICATION addr_pub_schema FOR TABLES IN SCHEMA addr_nsp;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_addr_sub CONNECTION '' PUBLICATION bar WITH (connect = false, slot_name = NONE);`,
			},
			{
				Statement: `CREATE STATISTICS addr_nsp.gentable_stat ON a, b FROM addr_nsp.gentable;`,
			},
			{
				Statement:   `SELECT pg_get_object_address('stone', '{}', '{}');`,
				ErrorString: `unrecognized object type "stone"`,
			},
			{
				Statement:   `SELECT pg_get_object_address('table', '{}', '{}');`,
				ErrorString: `name list length must be at least 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('table', '{NULL}', '{}');`,
				ErrorString: `name or argument lists may not contain nulls`,
			},
			{
				Statement: `DO $$
DECLARE
	objtype text;`,
			},
			{
				Statement: `BEGIN
	FOR objtype IN VALUES ('toast table'), ('index column'), ('sequence column'),
		('toast table column'), ('view column'), ('materialized view column')
	LOOP
		BEGIN
			PERFORM pg_get_object_address(objtype, '{one}', '{}');`,
			},
			{
				Statement: `		EXCEPTION WHEN invalid_parameter_value THEN
			RAISE WARNING 'error for %: %', objtype, sqlerrm;`,
			},
			{
				Statement: `		END;`,
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
				Statement:   `select * from pg_get_object_address('operator of access method', '{btree,integer_ops,1}', '{int4,bool}');`,
				ErrorString: `operator 1 (int4, bool) of operator family integer_ops for access method btree does not exist`,
			},
			{
				Statement:   `select * from pg_get_object_address('operator of access method', '{btree,integer_ops,99}', '{int4,int4}');`,
				ErrorString: `operator 99 (int4, int4) of operator family integer_ops for access method btree does not exist`,
			},
			{
				Statement:   `select * from pg_get_object_address('function of access method', '{btree,integer_ops,1}', '{int4,bool}');`,
				ErrorString: `function 1 (int4, bool) of operator family integer_ops for access method btree does not exist`,
			},
			{
				Statement:   `select * from pg_get_object_address('function of access method', '{btree,integer_ops,99}', '{int4,int4}');`,
				ErrorString: `function 99 (int4, int4) of operator family integer_ops for access method btree does not exist`,
			},
			{
				Statement: `DO $$
DECLARE
	objtype text;`,
			},
			{
				Statement: `	names	text[];`,
			},
			{
				Statement: `	args	text[];`,
			},
			{
				Statement: `BEGIN
	FOR objtype IN VALUES
		('table'), ('index'), ('sequence'), ('view'),
		('materialized view'), ('foreign table'),
		('table column'), ('foreign table column'),
		('aggregate'), ('function'), ('procedure'), ('type'), ('cast'),
		('table constraint'), ('domain constraint'), ('conversion'), ('default value'),
		('operator'), ('operator class'), ('operator family'), ('rule'), ('trigger'),
		('text search parser'), ('text search dictionary'),
		('text search template'), ('text search configuration'),
		('policy'), ('user mapping'), ('default acl'), ('transform'),
		('operator of access method'), ('function of access method'),
		('publication namespace'), ('publication relation')
	LOOP
		FOR names IN VALUES ('{eins}'), ('{addr_nsp, zwei}'), ('{eins, zwei, drei}')
		LOOP
			FOR args IN VALUES ('{}'), ('{integer}')
			LOOP
				BEGIN
					PERFORM pg_get_object_address(objtype, names, args);`,
			},
			{
				Statement: `				EXCEPTION WHEN OTHERS THEN
						RAISE WARNING 'error for %,%,%: %', objtype, names, args, sqlerrm;`,
			},
			{
				Statement: `				END;`,
			},
			{
				Statement: `			END LOOP;`,
			},
			{
				Statement: `		END LOOP;`,
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
				Statement:   `SELECT pg_get_object_address('language', '{one}', '{}');`,
				ErrorString: `language "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('language', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('large object', '{123}', '{}');`,
				ErrorString: `large object 123 does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('large object', '{123,456}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('large object', '{blargh}', '{}');`,
				ErrorString: `invalid input syntax for type oid: "blargh"`,
			},
			{
				Statement:   `SELECT pg_get_object_address('schema', '{one}', '{}');`,
				ErrorString: `schema "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('schema', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('role', '{one}', '{}');`,
				ErrorString: `role "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('role', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('database', '{one}', '{}');`,
				ErrorString: `database "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('database', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('tablespace', '{one}', '{}');`,
				ErrorString: `tablespace "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('tablespace', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('foreign-data wrapper', '{one}', '{}');`,
				ErrorString: `foreign-data wrapper "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('foreign-data wrapper', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('server', '{one}', '{}');`,
				ErrorString: `server "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('server', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('extension', '{one}', '{}');`,
				ErrorString: `extension "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('extension', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('event trigger', '{one}', '{}');`,
				ErrorString: `event trigger "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('event trigger', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('access method', '{one}', '{}');`,
				ErrorString: `access method "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('access method', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('publication', '{one}', '{}');`,
				ErrorString: `publication "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('publication', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement:   `SELECT pg_get_object_address('subscription', '{one}', '{}');`,
				ErrorString: `subscription "one" does not exist`,
			},
			{
				Statement:   `SELECT pg_get_object_address('subscription', '{one,two}', '{}');`,
				ErrorString: `name list length must be exactly 1`,
			},
			{
				Statement: `WITH objects (type, name, args) AS (VALUES
				('table', '{addr_nsp, gentable}'::text[], '{}'::text[]),
				('table', '{addr_nsp, parttable}'::text[], '{}'::text[]),
				('index', '{addr_nsp, gentable_pkey}', '{}'),
				('index', '{addr_nsp, parttable_pkey}', '{}'),
				('sequence', '{addr_nsp, gentable_a_seq}', '{}'),
				-- toast table
				('view', '{addr_nsp, genview}', '{}'),
				('materialized view', '{addr_nsp, genmatview}', '{}'),
				('foreign table', '{addr_nsp, genftable}', '{}'),
				('table column', '{addr_nsp, gentable, b}', '{}'),
				('foreign table column', '{addr_nsp, genftable, a}', '{}'),
				('aggregate', '{addr_nsp, genaggr}', '{int4}'),
				('function', '{pg_catalog, pg_identify_object}', '{pg_catalog.oid, pg_catalog.oid, int4}'),
				('procedure', '{addr_nsp, proc}', '{int4}'),
				('type', '{pg_catalog._int4}', '{}'),
				('type', '{addr_nsp.gendomain}', '{}'),
				('type', '{addr_nsp.gencomptype}', '{}'),
				('type', '{addr_nsp.genenum}', '{}'),
				('cast', '{int8}', '{int4}'),
				('collation', '{default}', '{}'),
				('table constraint', '{addr_nsp, gentable, a_chk}', '{}'),
				('domain constraint', '{addr_nsp.gendomain}', '{domconstr}'),
				('conversion', '{pg_catalog, koi8_r_to_mic}', '{}'),
				('default value', '{addr_nsp, gentable, b}', '{}'),
				('language', '{plpgsql}', '{}'),
				-- large object
				('operator', '{+}', '{int4, int4}'),
				('operator class', '{btree, int4_ops}', '{}'),
				('operator family', '{btree, integer_ops}', '{}'),
				('operator of access method', '{btree,integer_ops,1}', '{integer,integer}'),
				('function of access method', '{btree,integer_ops,2}', '{integer,integer}'),
				('rule', '{addr_nsp, genview, _RETURN}', '{}'),
				('trigger', '{addr_nsp, gentable, t}', '{}'),
				('schema', '{addr_nsp}', '{}'),
				('text search parser', '{addr_ts_prs}', '{}'),
				('text search dictionary', '{addr_ts_dict}', '{}'),
				('text search template', '{addr_ts_temp}', '{}'),
				('text search configuration', '{addr_ts_conf}', '{}'),
				('role', '{regress_addr_user}', '{}'),
				-- database
				-- tablespace
				('foreign-data wrapper', '{addr_fdw}', '{}'),
				('server', '{addr_fserv}', '{}'),
				('user mapping', '{regress_addr_user}', '{integer}'),
				('default acl', '{regress_addr_user,public}', '{r}'),
				('default acl', '{regress_addr_user}', '{r}'),
				-- extension
				-- event trigger
				('policy', '{addr_nsp, gentable, genpol}', '{}'),
				('transform', '{int}', '{sql}'),
				('access method', '{btree}', '{}'),
				('publication', '{addr_pub}', '{}'),
				('publication namespace', '{addr_nsp}', '{addr_pub_schema}'),
				('publication relation', '{addr_nsp, gentable}', '{addr_pub}'),
				('subscription', '{regress_addr_sub}', '{}'),
				('statistics object', '{addr_nsp, gentable_stat}', '{}')
        )
SELECT (pg_identify_object(addr1.classid, addr1.objid, addr1.objsubid)).*,
	-- test roundtrip through pg_identify_object_as_address
	ROW(pg_identify_object(addr1.classid, addr1.objid, addr1.objsubid)) =
	ROW(pg_identify_object(addr2.classid, addr2.objid, addr2.objsubid))
	  FROM objects, pg_get_object_address(type, name, args) addr1,
			pg_identify_object_as_address(classid, objid, objsubid) ioa(typ,nms,args),
			pg_get_object_address(typ, nms, ioa.args) as addr2
	ORDER BY addr1.classid, addr1.objid, addr1.objsubid;`,
				Results: []sql.Row{{`default acl`, ``, ``, `for role regress_addr_user in schema public on tables`, true}, {`default acl`, ``, ``, `for role regress_addr_user on tables`, true}, {`type`, `pg_catalog`, `_int4`, `integer[]`, true}, {`type`, `addr_nsp`, `gencomptype`, `addr_nsp.gencomptype`, true}, {`type`, `addr_nsp`, `genenum`, `addr_nsp.genenum`, true}, {`type`, `addr_nsp`, `gendomain`, `addr_nsp.gendomain`, true}, {`function`, `pg_catalog`, ``, `pg_catalog.pg_identify_object(pg_catalog.oid,pg_catalog.oid,integer)`, true}, {`aggregate`, `addr_nsp`, ``, `addr_nsp.genaggr(integer)`, true}, {`procedure`, `addr_nsp`, ``, `addr_nsp.proc(integer)`, true}, {`sequence`, `addr_nsp`, `gentable_a_seq`, `addr_nsp.gentable_a_seq`, true}, {`table`, `addr_nsp`, `gentable`, `addr_nsp.gentable`, true}, {`table column`, `addr_nsp`, `gentable`, `addr_nsp.gentable.b`, true}, {`index`, `addr_nsp`, `gentable_pkey`, `addr_nsp.gentable_pkey`, true}, {`table`, `addr_nsp`, `parttable`, `addr_nsp.parttable`, true}, {`index`, `addr_nsp`, `parttable_pkey`, `addr_nsp.parttable_pkey`, true}, {`view`, `addr_nsp`, `genview`, `addr_nsp.genview`, true}, {`materialized view`, `addr_nsp`, `genmatview`, `addr_nsp.genmatview`, true}, {`foreign table`, `addr_nsp`, `genftable`, `addr_nsp.genftable`, true}, {`foreign table column`, `addr_nsp`, `genftable`, `addr_nsp.genftable.a`, true}, {`role`, ``, `regress_addr_user`, `regress_addr_user`, true}, {`server`, ``, `addr_fserv`, `addr_fserv`, true}, {`user mapping`, ``, ``, `regress_addr_user on server integer`, true}, {`foreign-data wrapper`, ``, `addr_fdw`, `addr_fdw`, true}, {`access method`, ``, `btree`, `btree`, true}, {`operator of access method`, ``, ``, `operator 1 (integer, integer) of pg_catalog.integer_ops USING btree`, true}, {`function of access method`, ``, ``, `function 2 (integer, integer) of pg_catalog.integer_ops USING btree`, true}, {`default value`, ``, ``, `for addr_nsp.gentable.b`, true}, {`cast`, ``, ``, `(bigint AS integer)`, true}, {`table constraint`, `addr_nsp`, ``, `a_chk on addr_nsp.gentable`, true}, {`domain constraint`, `addr_nsp`, ``, `domconstr on addr_nsp.gendomain`, true}, {`conversion`, `pg_catalog`, `koi8_r_to_mic`, `pg_catalog.koi8_r_to_mic`, true}, {`language`, ``, `plpgsql`, `plpgsql`, true}, {`schema`, ``, `addr_nsp`, `addr_nsp`, true}, {`operator class`, `pg_catalog`, `int4_ops`, `pg_catalog.int4_ops USING btree`, true}, {`operator`, `pg_catalog`, ``, `pg_catalog.+(integer,integer)`, true}, {`rule`, ``, ``, `"_RETURN" on addr_nsp.genview`, true}, {`trigger`, ``, ``, `t on addr_nsp.gentable`, true}, {`operator family`, `pg_catalog`, `integer_ops`, `pg_catalog.integer_ops USING btree`, true}, {`policy`, ``, ``, `genpol on addr_nsp.gentable`, true}, {`statistics object`, `addr_nsp`, `gentable_stat`, `addr_nsp.gentable_stat`, true}, {`collation`, `pg_catalog`, "default", `pg_catalog."default"`, true}, {`transform`, ``, ``, `for integer on language sql`, true}, {`text search dictionary`, `addr_nsp`, `addr_ts_dict`, `addr_nsp.addr_ts_dict`, true}, {`text search parser`, `addr_nsp`, `addr_ts_prs`, `addr_nsp.addr_ts_prs`, true}, {`text search configuration`, `addr_nsp`, `addr_ts_conf`, `addr_nsp.addr_ts_conf`, true}, {`text search template`, `addr_nsp`, `addr_ts_temp`, `addr_nsp.addr_ts_temp`, true}, {`subscription`, ``, `regress_addr_sub`, `regress_addr_sub`, true}, {`publication`, ``, `addr_pub`, `addr_pub`, true}, {`publication relation`, ``, ``, `addr_nsp.gentable in publication addr_pub`, true}, {`publication namespace`, ``, ``, `addr_nsp in publication addr_pub_schema`, true}},
			},
			{
				Statement: `---
---
DROP FOREIGN DATA WRAPPER addr_fdw CASCADE;`,
			},
			{
				Statement: `DROP PUBLICATION addr_pub;`,
			},
			{
				Statement: `DROP PUBLICATION addr_pub_schema;`,
			},
			{
				Statement: `DROP SUBSCRIPTION regress_addr_sub;`,
			},
			{
				Statement: `DROP SCHEMA addr_nsp CASCADE;`,
			},
			{
				Statement: `DROP OWNED BY regress_addr_user;`,
			},
			{
				Statement: `DROP USER regress_addr_user;`,
			},
			{
				Statement: `\pset null 'NULL'
\a\t
WITH objects (classid, objid, objsubid) AS (VALUES
    ('pg_class'::regclass, 0, 0), -- no relation
    ('pg_class'::regclass, 'pg_class'::regclass, 100), -- no column for relation
    ('pg_proc'::regclass, 0, 0), -- no function
    ('pg_type'::regclass, 0, 0), -- no type
    ('pg_cast'::regclass, 0, 0), -- no cast
    ('pg_collation'::regclass, 0, 0), -- no collation
    ('pg_constraint'::regclass, 0, 0), -- no constraint
    ('pg_conversion'::regclass, 0, 0), -- no conversion
    ('pg_attrdef'::regclass, 0, 0), -- no default attribute
    ('pg_language'::regclass, 0, 0), -- no language
    ('pg_largeobject'::regclass, 0, 0), -- no large object, no error
    ('pg_operator'::regclass, 0, 0), -- no operator
    ('pg_opclass'::regclass, 0, 0), -- no opclass, no need to check for no access method
    ('pg_opfamily'::regclass, 0, 0), -- no opfamily
    ('pg_am'::regclass, 0, 0), -- no access method
    ('pg_amop'::regclass, 0, 0), -- no AM operator
    ('pg_amproc'::regclass, 0, 0), -- no AM proc
    ('pg_rewrite'::regclass, 0, 0), -- no rewrite
    ('pg_trigger'::regclass, 0, 0), -- no trigger
    ('pg_namespace'::regclass, 0, 0), -- no schema
    ('pg_statistic_ext'::regclass, 0, 0), -- no statistics
    ('pg_ts_parser'::regclass, 0, 0), -- no TS parser
    ('pg_ts_dict'::regclass, 0, 0), -- no TS dictionary
    ('pg_ts_template'::regclass, 0, 0), -- no TS template
    ('pg_ts_config'::regclass, 0, 0), -- no TS configuration
    ('pg_authid'::regclass, 0, 0), -- no role
    ('pg_database'::regclass, 0, 0), -- no database
    ('pg_tablespace'::regclass, 0, 0), -- no tablespace
    ('pg_foreign_data_wrapper'::regclass, 0, 0), -- no FDW
    ('pg_foreign_server'::regclass, 0, 0), -- no server
    ('pg_user_mapping'::regclass, 0, 0), -- no user mapping
    ('pg_default_acl'::regclass, 0, 0), -- no default ACL
    ('pg_extension'::regclass, 0, 0), -- no extension
    ('pg_event_trigger'::regclass, 0, 0), -- no event trigger
    ('pg_policy'::regclass, 0, 0), -- no policy
    ('pg_publication'::regclass, 0, 0), -- no publication
    ('pg_publication_rel'::regclass, 0, 0), -- no publication relation
    ('pg_subscription'::regclass, 0, 0), -- no subscription
    ('pg_transform'::regclass, 0, 0) -- no transformation
  )
SELECT ROW(pg_identify_object(objects.classid, objects.objid, objects.objsubid))
         AS ident,
       ROW(pg_identify_object_as_address(objects.classid, objects.objid, objects.objsubid))
         AS addr,
       pg_describe_object(objects.classid, objects.objid, objects.objsubid)
         AS descr
FROM objects
ORDER BY objects.classid, objects.objid, objects.objsubid;`,
			},
			{
				Statement: `("(""default acl"",,,)")|("(""default acl"",,)")|NULL
("(tablespace,,,)")|("(tablespace,,)")|NULL
("(type,,,)")|("(type,,)")|NULL
("(routine,,,)")|("(routine,,)")|NULL
("(relation,,,)")|("(relation,,)")|NULL
("(""table column"",,,)")|("(""table column"",,)")|NULL
("(role,,,)")|("(role,,)")|NULL
("(database,,,)")|("(database,,)")|NULL
("(server,,,)")|("(server,,)")|NULL
("(""user mapping"",,,)")|("(""user mapping"",,)")|NULL
("(""foreign-data wrapper"",,,)")|("(""foreign-data wrapper"",,)")|NULL
("(""access method"",,,)")|("(""access method"",,)")|NULL
("(""operator of access method"",,,)")|("(""operator of access method"",,)")|NULL
("(""function of access method"",,,)")|("(""function of access method"",,)")|NULL
("(""default value"",,,)")|("(""default value"",,)")|NULL
("(cast,,,)")|("(cast,,)")|NULL
("(constraint,,,)")|("(constraint,,)")|NULL
("(conversion,,,)")|("(conversion,,)")|NULL
("(language,,,)")|("(language,,)")|NULL
("(""large object"",,,)")|("(""large object"",,)")|NULL
("(schema,,,)")|("(schema,,)")|NULL
("(""operator class"",,,)")|("(""operator class"",,)")|NULL
("(operator,,,)")|("(operator,,)")|NULL
("(rule,,,)")|("(rule,,)")|NULL
("(trigger,,,)")|("(trigger,,)")|NULL
("(""operator family"",,,)")|("(""operator family"",,)")|NULL
("(extension,,,)")|("(extension,,)")|NULL
("(policy,,,)")|("(policy,,)")|NULL
("(""statistics object"",,,)")|("(""statistics object"",,)")|NULL
("(collation,,,)")|("(collation,,)")|NULL
("(""event trigger"",,,)")|("(""event trigger"",,)")|NULL
("(transform,,,)")|("(transform,,)")|NULL
("(""text search dictionary"",,,)")|("(""text search dictionary"",,)")|NULL
("(""text search parser"",,,)")|("(""text search parser"",,)")|NULL
("(""text search configuration"",,,)")|("(""text search configuration"",,)")|NULL
("(""text search template"",,,)")|("(""text search template"",,)")|NULL
("(subscription,,,)")|("(subscription,,)")|NULL
("(publication,,,)")|("(publication,,)")|NULL
("(""publication relation"",,,)")|("(""publication relation"",,)")|NULL
\a\t`,
			},
		},
	})
}
