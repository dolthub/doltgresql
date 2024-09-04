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

func TestAlterGeneric(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_alter_generic)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_alter_generic,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set regresslib :libdir '/regress' :dlsuffix
CREATE FUNCTION test_opclass_options_func(internal)
    RETURNS void
    AS :'regresslib', 'test_opclass_options_func'
    LANGUAGE C;`,
			},
			{
				Statement: `SET client_min_messages TO 'warning';`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_alter_generic_user1;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_alter_generic_user2;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_alter_generic_user3;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE USER regress_alter_generic_user3;`,
			},
			{
				Statement: `CREATE USER regress_alter_generic_user2;`,
			},
			{
				Statement: `CREATE USER regress_alter_generic_user1 IN ROLE regress_alter_generic_user3;`,
			},
			{
				Statement: `CREATE SCHEMA alt_nsp1;`,
			},
			{
				Statement: `CREATE SCHEMA alt_nsp2;`,
			},
			{
				Statement: `GRANT ALL ON SCHEMA alt_nsp1, alt_nsp2 TO public;`,
			},
			{
				Statement: `SET search_path = alt_nsp1, public;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user1;`,
			},
			{
				Statement: `CREATE FUNCTION alt_func1(int) RETURNS int LANGUAGE sql
  AS 'SELECT $1 + 1';`,
			},
			{
				Statement: `CREATE FUNCTION alt_func2(int) RETURNS int LANGUAGE sql
  AS 'SELECT $1 - 1';`,
			},
			{
				Statement: `CREATE AGGREGATE alt_agg1 (
  sfunc1 = int4pl, basetype = int4, stype1 = int4, initcond = 0
);`,
			},
			{
				Statement: `CREATE AGGREGATE alt_agg2 (
  sfunc1 = int4mi, basetype = int4, stype1 = int4, initcond = 0
);`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_func1(int) RENAME TO alt_func3;  -- failed (not aggregate)`,
				ErrorString: `function alt_func1(integer) is not an aggregate`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_func1(int) OWNER TO regress_alter_generic_user3;  -- failed (not aggregate)`,
				ErrorString: `function alt_func1(integer) is not an aggregate`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_func1(int) SET SCHEMA alt_nsp2;  -- failed (not aggregate)`,
				ErrorString: `function alt_func1(integer) is not an aggregate`,
			},
			{
				Statement:   `ALTER FUNCTION alt_func1(int) RENAME TO alt_func2;  -- failed (name conflict)`,
				ErrorString: `function alt_func2(integer) already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER FUNCTION alt_func1(int) RENAME TO alt_func3;  -- OK`,
			},
			{
				Statement:   `ALTER FUNCTION alt_func2(int) OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER FUNCTION alt_func2(int) OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER FUNCTION alt_func2(int) SET SCHEMA alt_nsp1;  -- OK, already there`,
			},
			{
				Statement: `ALTER FUNCTION alt_func2(int) SET SCHEMA alt_nsp2;  -- OK`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_agg1(int) RENAME TO alt_agg2;   -- failed (name conflict)`,
				ErrorString: `function alt_agg2(integer) already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER AGGREGATE alt_agg1(int) RENAME TO alt_agg3;   -- OK`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_agg2(int) OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER AGGREGATE alt_agg2(int) OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER AGGREGATE alt_agg2(int) SET SCHEMA alt_nsp2;  -- OK`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user2;`,
			},
			{
				Statement: `CREATE FUNCTION alt_func1(int) RETURNS int LANGUAGE sql
  AS 'SELECT $1 + 2';`,
			},
			{
				Statement: `CREATE FUNCTION alt_func2(int) RETURNS int LANGUAGE sql
  AS 'SELECT $1 - 2';`,
			},
			{
				Statement: `CREATE AGGREGATE alt_agg1 (
  sfunc1 = int4pl, basetype = int4, stype1 = int4, initcond = 100
);`,
			},
			{
				Statement: `CREATE AGGREGATE alt_agg2 (
  sfunc1 = int4mi, basetype = int4, stype1 = int4, initcond = -100
);`,
			},
			{
				Statement:   `ALTER FUNCTION alt_func3(int) RENAME TO alt_func4;	-- failed (not owner)`,
				ErrorString: `must be owner of function alt_func3`,
			},
			{
				Statement: `ALTER FUNCTION alt_func1(int) RENAME TO alt_func4;	-- OK`,
			},
			{
				Statement:   `ALTER FUNCTION alt_func3(int) OWNER TO regress_alter_generic_user2;	-- failed (not owner)`,
				ErrorString: `must be owner of function alt_func3`,
			},
			{
				Statement:   `ALTER FUNCTION alt_func2(int) OWNER TO regress_alter_generic_user3;	-- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER FUNCTION alt_func3(int) SET SCHEMA alt_nsp2;      -- failed (not owner)`,
				ErrorString: `must be owner of function alt_func3`,
			},
			{
				Statement:   `ALTER FUNCTION alt_func2(int) SET SCHEMA alt_nsp2;	-- failed (name conflicts)`,
				ErrorString: `function alt_func2(integer) already exists in schema "alt_nsp2"`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_agg3(int) RENAME TO alt_agg4;   -- failed (not owner)`,
				ErrorString: `must be owner of function alt_agg3`,
			},
			{
				Statement: `ALTER AGGREGATE alt_agg1(int) RENAME TO alt_agg4;   -- OK`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_agg3(int) OWNER TO regress_alter_generic_user2;  -- failed (not owner)`,
				ErrorString: `must be owner of function alt_agg3`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_agg2(int) OWNER TO regress_alter_generic_user3;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_agg3(int) SET SCHEMA alt_nsp2;  -- failed (not owner)`,
				ErrorString: `must be owner of function alt_agg3`,
			},
			{
				Statement:   `ALTER AGGREGATE alt_agg2(int) SET SCHEMA alt_nsp2;  -- failed (name conflict)`,
				ErrorString: `function alt_agg2(integer) already exists in schema "alt_nsp2"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT n.nspname, proname, prorettype::regtype, prokind, a.rolname
  FROM pg_proc p, pg_namespace n, pg_authid a
  WHERE p.pronamespace = n.oid AND p.proowner = a.oid
    AND n.nspname IN ('alt_nsp1', 'alt_nsp2')
  ORDER BY nspname, proname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_agg2`, `integer`, `a`, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_agg3`, `integer`, `a`, `regress_alter_generic_user1`}, {`alt_nsp1`, `alt_agg4`, `integer`, `a`, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_func2`, `integer`, false, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_func3`, `integer`, false, `regress_alter_generic_user1`}, {`alt_nsp1`, `alt_func4`, `integer`, false, `regress_alter_generic_user2`}, {`alt_nsp2`, `alt_agg2`, `integer`, `a`, `regress_alter_generic_user3`}, {`alt_nsp2`, `alt_func2`, `integer`, false, `regress_alter_generic_user3`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user1;`,
			},
			{
				Statement: `CREATE CONVERSION alt_conv1 FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
			},
			{
				Statement: `CREATE CONVERSION alt_conv2 FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
			},
			{
				Statement:   `ALTER CONVERSION alt_conv1 RENAME TO alt_conv2;  -- failed (name conflict)`,
				ErrorString: `conversion "alt_conv2" already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER CONVERSION alt_conv1 RENAME TO alt_conv3;  -- OK`,
			},
			{
				Statement:   `ALTER CONVERSION alt_conv2 OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER CONVERSION alt_conv2 OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER CONVERSION alt_conv2 SET SCHEMA alt_nsp2;  -- OK`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user2;`,
			},
			{
				Statement: `CREATE CONVERSION alt_conv1 FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
			},
			{
				Statement: `CREATE CONVERSION alt_conv2 FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
			},
			{
				Statement:   `ALTER CONVERSION alt_conv3 RENAME TO alt_conv4;  -- failed (not owner)`,
				ErrorString: `must be owner of conversion alt_conv3`,
			},
			{
				Statement: `ALTER CONVERSION alt_conv1 RENAME TO alt_conv4;  -- OK`,
			},
			{
				Statement:   `ALTER CONVERSION alt_conv3 OWNER TO regress_alter_generic_user2;  -- failed (not owner)`,
				ErrorString: `must be owner of conversion alt_conv3`,
			},
			{
				Statement:   `ALTER CONVERSION alt_conv2 OWNER TO regress_alter_generic_user3;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER CONVERSION alt_conv3 SET SCHEMA alt_nsp2;  -- failed (not owner)`,
				ErrorString: `must be owner of conversion alt_conv3`,
			},
			{
				Statement:   `ALTER CONVERSION alt_conv2 SET SCHEMA alt_nsp2;  -- failed (name conflict)`,
				ErrorString: `conversion "alt_conv2" already exists in schema "alt_nsp2"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT n.nspname, c.conname, a.rolname
  FROM pg_conversion c, pg_namespace n, pg_authid a
  WHERE c.connamespace = n.oid AND c.conowner = a.oid
    AND n.nspname IN ('alt_nsp1', 'alt_nsp2')
  ORDER BY nspname, conname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_conv2`, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_conv3`, `regress_alter_generic_user1`}, {`alt_nsp1`, `alt_conv4`, `regress_alter_generic_user2`}, {`alt_nsp2`, `alt_conv2`, `regress_alter_generic_user3`}},
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER alt_fdw1;`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER alt_fdw2;`,
			},
			{
				Statement: `CREATE SERVER alt_fserv1 FOREIGN DATA WRAPPER alt_fdw1;`,
			},
			{
				Statement: `CREATE SERVER alt_fserv2 FOREIGN DATA WRAPPER alt_fdw2;`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER alt_fdw1 RENAME TO alt_fdw2;  -- failed (name conflict)`,
				ErrorString: `foreign-data wrapper "alt_fdw2" already exists`,
			},
			{
				Statement: `ALTER FOREIGN DATA WRAPPER alt_fdw1 RENAME TO alt_fdw3;  -- OK`,
			},
			{
				Statement:   `ALTER SERVER alt_fserv1 RENAME TO alt_fserv2;   -- failed (name conflict)`,
				ErrorString: `server "alt_fserv2" already exists`,
			},
			{
				Statement: `ALTER SERVER alt_fserv1 RENAME TO alt_fserv3;   -- OK`,
			},
			{
				Statement: `SELECT fdwname FROM pg_foreign_data_wrapper WHERE fdwname like 'alt_fdw%';`,
				Results:   []sql.Row{{`alt_fdw2`}, {`alt_fdw3`}},
			},
			{
				Statement: `SELECT srvname FROM pg_foreign_server WHERE srvname like 'alt_fserv%';`,
				Results:   []sql.Row{{`alt_fserv2`}, {`alt_fserv3`}},
			},
			{
				Statement: `CREATE LANGUAGE alt_lang1 HANDLER plpgsql_call_handler;`,
			},
			{
				Statement: `CREATE LANGUAGE alt_lang2 HANDLER plpgsql_call_handler;`,
			},
			{
				Statement: `ALTER LANGUAGE alt_lang1 OWNER TO regress_alter_generic_user1;  -- OK`,
			},
			{
				Statement: `ALTER LANGUAGE alt_lang2 OWNER TO regress_alter_generic_user2;  -- OK`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user1;`,
			},
			{
				Statement:   `ALTER LANGUAGE alt_lang1 RENAME TO alt_lang2;   -- failed (name conflict)`,
				ErrorString: `language "alt_lang2" already exists`,
			},
			{
				Statement:   `ALTER LANGUAGE alt_lang2 RENAME TO alt_lang3;   -- failed (not owner)`,
				ErrorString: `must be owner of language alt_lang2`,
			},
			{
				Statement: `ALTER LANGUAGE alt_lang1 RENAME TO alt_lang3;   -- OK`,
			},
			{
				Statement:   `ALTER LANGUAGE alt_lang2 OWNER TO regress_alter_generic_user3;  -- failed (not owner)`,
				ErrorString: `must be owner of language alt_lang2`,
			},
			{
				Statement:   `ALTER LANGUAGE alt_lang3 OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER LANGUAGE alt_lang3 OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT lanname, a.rolname
  FROM pg_language l, pg_authid a
  WHERE l.lanowner = a.oid AND l.lanname like 'alt_lang%'
  ORDER BY lanname;`,
				Results: []sql.Row{{`alt_lang2`, `regress_alter_generic_user2`}, {`alt_lang3`, `regress_alter_generic_user3`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user1;`,
			},
			{
				Statement: `CREATE OPERATOR @-@ ( leftarg = int4, rightarg = int4, procedure = int4mi );`,
			},
			{
				Statement: `CREATE OPERATOR @+@ ( leftarg = int4, rightarg = int4, procedure = int4pl );`,
			},
			{
				Statement:   `ALTER OPERATOR @+@(int4, int4) OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER OPERATOR @+@(int4, int4) OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER OPERATOR @-@(int4, int4) SET SCHEMA alt_nsp2;           -- OK`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user2;`,
			},
			{
				Statement: `CREATE OPERATOR @-@ ( leftarg = int4, rightarg = int4, procedure = int4mi );`,
			},
			{
				Statement:   `ALTER OPERATOR @+@(int4, int4) OWNER TO regress_alter_generic_user2;  -- failed (not owner)`,
				ErrorString: `must be owner of operator @+@`,
			},
			{
				Statement:   `ALTER OPERATOR @-@(int4, int4) OWNER TO regress_alter_generic_user3;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER OPERATOR @+@(int4, int4) SET SCHEMA alt_nsp2;   -- failed (not owner)`,
				ErrorString: `must be owner of operator @+@`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT n.nspname, oprname, a.rolname,
    oprleft::regtype, oprright::regtype, oprcode::regproc
  FROM pg_operator o, pg_namespace n, pg_authid a
  WHERE o.oprnamespace = n.oid AND o.oprowner = a.oid
    AND n.nspname IN ('alt_nsp1', 'alt_nsp2')
  ORDER BY nspname, oprname;`,
				Results: []sql.Row{{`alt_nsp1`, `@+@`, `regress_alter_generic_user3`, `integer`, `integer`, `int4pl`}, {`alt_nsp1`, `@-@`, `regress_alter_generic_user2`, `integer`, `integer`, `int4mi`}, {`alt_nsp2`, `@-@`, `regress_alter_generic_user1`, `integer`, `integer`, `int4mi`}},
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf1 USING hash;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf2 USING hash;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf1 USING hash OWNER TO regress_alter_generic_user1;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf2 USING hash OWNER TO regress_alter_generic_user1;`,
			},
			{
				Statement: `CREATE OPERATOR CLASS alt_opc1 FOR TYPE uuid USING hash AS STORAGE uuid;`,
			},
			{
				Statement: `CREATE OPERATOR CLASS alt_opc2 FOR TYPE uuid USING hash AS STORAGE uuid;`,
			},
			{
				Statement: `ALTER OPERATOR CLASS alt_opc1 USING hash OWNER TO regress_alter_generic_user1;`,
			},
			{
				Statement: `ALTER OPERATOR CLASS alt_opc2 USING hash OWNER TO regress_alter_generic_user1;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user1;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf1 USING hash RENAME TO alt_opf2;  -- failed (name conflict)`,
				ErrorString: `operator family "alt_opf2" for access method "hash" already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf1 USING hash RENAME TO alt_opf3;  -- OK`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf2 USING hash OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf2 USING hash OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf2 USING hash SET SCHEMA alt_nsp2;  -- OK`,
			},
			{
				Statement:   `ALTER OPERATOR CLASS alt_opc1 USING hash RENAME TO alt_opc2;  -- failed (name conflict)`,
				ErrorString: `operator class "alt_opc2" for access method "hash" already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER OPERATOR CLASS alt_opc1 USING hash RENAME TO alt_opc3;  -- OK`,
			},
			{
				Statement:   `ALTER OPERATOR CLASS alt_opc2 USING hash OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER OPERATOR CLASS alt_opc2 USING hash OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER OPERATOR CLASS alt_opc2 USING hash SET SCHEMA alt_nsp2;  -- OK`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf1 USING hash;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf2 USING hash;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf1 USING hash OWNER TO regress_alter_generic_user2;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf2 USING hash OWNER TO regress_alter_generic_user2;`,
			},
			{
				Statement: `CREATE OPERATOR CLASS alt_opc1 FOR TYPE macaddr USING hash AS STORAGE macaddr;`,
			},
			{
				Statement: `CREATE OPERATOR CLASS alt_opc2 FOR TYPE macaddr USING hash AS STORAGE macaddr;`,
			},
			{
				Statement: `ALTER OPERATOR CLASS alt_opc1 USING hash OWNER TO regress_alter_generic_user2;`,
			},
			{
				Statement: `ALTER OPERATOR CLASS alt_opc2 USING hash OWNER TO regress_alter_generic_user2;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user2;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf3 USING hash RENAME TO alt_opf4;	-- failed (not owner)`,
				ErrorString: `must be owner of operator family alt_opf3`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf1 USING hash RENAME TO alt_opf4;  -- OK`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf3 USING hash OWNER TO regress_alter_generic_user2;  -- failed (not owner)`,
				ErrorString: `must be owner of operator family alt_opf3`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf2 USING hash OWNER TO regress_alter_generic_user3;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf3 USING hash SET SCHEMA alt_nsp2;  -- failed (not owner)`,
				ErrorString: `must be owner of operator family alt_opf3`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf2 USING hash SET SCHEMA alt_nsp2;  -- failed (name conflict)`,
				ErrorString: `operator family "alt_opf2" for access method "hash" already exists in schema "alt_nsp2"`,
			},
			{
				Statement:   `ALTER OPERATOR CLASS alt_opc3 USING hash RENAME TO alt_opc4;	-- failed (not owner)`,
				ErrorString: `must be owner of operator class alt_opc3`,
			},
			{
				Statement: `ALTER OPERATOR CLASS alt_opc1 USING hash RENAME TO alt_opc4;  -- OK`,
			},
			{
				Statement:   `ALTER OPERATOR CLASS alt_opc3 USING hash OWNER TO regress_alter_generic_user2;  -- failed (not owner)`,
				ErrorString: `must be owner of operator class alt_opc3`,
			},
			{
				Statement:   `ALTER OPERATOR CLASS alt_opc2 USING hash OWNER TO regress_alter_generic_user3;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER OPERATOR CLASS alt_opc3 USING hash SET SCHEMA alt_nsp2;  -- failed (not owner)`,
				ErrorString: `must be owner of operator class alt_opc3`,
			},
			{
				Statement:   `ALTER OPERATOR CLASS alt_opc2 USING hash SET SCHEMA alt_nsp2;  -- failed (name conflict)`,
				ErrorString: `operator class "alt_opc2" for access method "hash" already exists in schema "alt_nsp2"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT nspname, opfname, amname, rolname
  FROM pg_opfamily o, pg_am m, pg_namespace n, pg_authid a
  WHERE o.opfmethod = m.oid AND o.opfnamespace = n.oid AND o.opfowner = a.oid
    AND n.nspname IN ('alt_nsp1', 'alt_nsp2')
	AND NOT opfname LIKE 'alt_opc%'
  ORDER BY nspname, opfname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_opf2`, `hash`, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_opf3`, `hash`, `regress_alter_generic_user1`}, {`alt_nsp1`, `alt_opf4`, `hash`, `regress_alter_generic_user2`}, {`alt_nsp2`, `alt_opf2`, `hash`, `regress_alter_generic_user3`}},
			},
			{
				Statement: `SELECT nspname, opcname, amname, rolname
  FROM pg_opclass o, pg_am m, pg_namespace n, pg_authid a
  WHERE o.opcmethod = m.oid AND o.opcnamespace = n.oid AND o.opcowner = a.oid
    AND n.nspname IN ('alt_nsp1', 'alt_nsp2')
  ORDER BY nspname, opcname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_opc2`, `hash`, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_opc3`, `hash`, `regress_alter_generic_user1`}, {`alt_nsp1`, `alt_opc4`, `hash`, `regress_alter_generic_user2`}, {`alt_nsp2`, `alt_opc2`, `hash`, `regress_alter_generic_user3`}},
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf4 USING btree;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf4 USING btree ADD
  -- int4 vs int2
  OPERATOR 1 < (int4, int2) ,
  OPERATOR 2 <= (int4, int2) ,
  OPERATOR 3 = (int4, int2) ,
  OPERATOR 4 >= (int4, int2) ,
  OPERATOR 5 > (int4, int2) ,
  FUNCTION 1 btint42cmp(int4, int2);`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf4 USING btree DROP
  -- int4 vs int2
  OPERATOR 1 (int4, int2) ,
  OPERATOR 2 (int4, int2) ,
  OPERATOR 3 (int4, int2) ,
  OPERATOR 4 (int4, int2) ,
  OPERATOR 5 (int4, int2) ,
  FUNCTION 1 (int4, int2) ;`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf4 USING btree;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf4 USING btree;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf4 USING invalid_index_method ADD  OPERATOR 1 < (int4, int2); -- invalid indexing_method`,
				ErrorString: `access method "invalid_index_method" does not exist`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf4 USING btree ADD OPERATOR 6 < (int4, int2); -- operator number should be between 1 and 5`,
				ErrorString: `invalid operator number 6, must be between 1 and 5`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf4 USING btree ADD OPERATOR 0 < (int4, int2); -- operator number should be between 1 and 5`,
				ErrorString: `invalid operator number 0, must be between 1 and 5`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf4 USING btree ADD OPERATOR 1 < ; -- operator without argument types`,
				ErrorString: `operator argument types must be specified in ALTER OPERATOR FAMILY`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf4 USING btree ADD FUNCTION 0 btint42cmp(int4, int2); -- invalid options parsing function`,
				ErrorString: `invalid function number 0, must be between 1 and 5`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf4 USING btree ADD FUNCTION 6 btint42cmp(int4, int2); -- function number should be between 1 and 5`,
				ErrorString: `invalid function number 6, must be between 1 and 5`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf4 USING btree ADD STORAGE invalid_storage; -- Ensure STORAGE is not a part of ALTER OPERATOR FAMILY`,
				ErrorString: `STORAGE cannot be specified in ALTER OPERATOR FAMILY`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf4 USING btree;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE ROLE regress_alter_generic_user5 NOSUPERUSER;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf5 USING btree;`,
			},
			{
				Statement: `SET ROLE regress_alter_generic_user5;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf5 USING btree ADD OPERATOR 1 < (int4, int2), FUNCTION 1 btint42cmp(int4, int2);`,
				ErrorString: `must be superuser to alter an operator family`,
			},
			{
				Statement:   `RESET ROLE;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement:   `DROP OPERATOR FAMILY alt_opf5 USING btree;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE ROLE regress_alter_generic_user6;`,
			},
			{
				Statement: `CREATE SCHEMA alt_nsp6;`,
			},
			{
				Statement: `REVOKE ALL ON SCHEMA alt_nsp6 FROM regress_alter_generic_user6;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_nsp6.alt_opf6 USING btree;`,
			},
			{
				Statement: `SET ROLE regress_alter_generic_user6;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_nsp6.alt_opf6 USING btree ADD OPERATOR 1 < (int4, int2);`,
				ErrorString: `permission denied for schema alt_nsp6`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf7 USING btree;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf7 USING btree ADD OPERATOR 1 < (int4, int2);`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf7 USING btree DROP OPERATOR 1 (int4, int2, int8);`,
				ErrorString: `one or two argument types must be specified`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf7 USING btree;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf8 USING btree;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf8 USING btree ADD OPERATOR 1 < (int4, int4);`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf8 USING btree;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf9 USING gist;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf9 USING gist ADD OPERATOR 1 < (int4, int4) FOR ORDER BY float_ops;`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf9 USING gist;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf10 USING btree;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf10 USING btree ADD OPERATOR 1 < (int4, int4) FOR ORDER BY float_ops;`,
				ErrorString: `access method "btree" does not support ordering operators`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf10 USING btree;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf11 USING gist;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf11 USING gist ADD OPERATOR 1 < (int4, int4) FOR ORDER BY float_ops;`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf11 USING gist DROP OPERATOR 1 (int4, int4);`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf11 USING gist;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf12 USING btree;`,
			},
			{
				Statement: `CREATE FUNCTION fn_opf12  (int4, int2) RETURNS BIGINT AS 'SELECT NULL::BIGINT;' LANGUAGE SQL;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf12 USING btree ADD FUNCTION 1 fn_opf12(int4, int2);`,
				ErrorString: `btree comparison functions must return integer`,
			},
			{
				Statement:   `DROP OPERATOR FAMILY alt_opf12 USING btree;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf13 USING hash;`,
			},
			{
				Statement: `CREATE FUNCTION fn_opf13  (int4) RETURNS BIGINT AS 'SELECT NULL::BIGINT;' LANGUAGE SQL;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf13 USING hash ADD FUNCTION 1 fn_opf13(int4);`,
				ErrorString: `hash function 1 must return integer`,
			},
			{
				Statement:   `DROP OPERATOR FAMILY alt_opf13 USING hash;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf14 USING btree;`,
			},
			{
				Statement: `CREATE FUNCTION fn_opf14 (int4) RETURNS BIGINT AS 'SELECT NULL::BIGINT;' LANGUAGE SQL;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf14 USING btree ADD FUNCTION 1 fn_opf14(int4);`,
				ErrorString: `btree comparison functions must have two arguments`,
			},
			{
				Statement:   `DROP OPERATOR FAMILY alt_opf14 USING btree;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf15 USING hash;`,
			},
			{
				Statement: `CREATE FUNCTION fn_opf15 (int4, int2) RETURNS BIGINT AS 'SELECT NULL::BIGINT;' LANGUAGE SQL;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf15 USING hash ADD FUNCTION 1 fn_opf15(int4, int2);`,
				ErrorString: `hash function 1 must have one argument`,
			},
			{
				Statement:   `DROP OPERATOR FAMILY alt_opf15 USING hash;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf16 USING gist;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf16 USING gist ADD FUNCTION 1 btint42cmp(int4, int2);`,
				ErrorString: `associated data types must be specified for index support function`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf16 USING gist;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf17 USING btree;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf17 USING btree ADD OPERATOR 1 < (int4, int4), OPERATOR 1 < (int4, int4); -- operator # appears twice in same statement`,
				ErrorString: `operator number 1 for (integer,integer) appears more than once`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf17 USING btree ADD OPERATOR 1 < (int4, int4); -- operator 1 requested first-time`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf17 USING btree ADD OPERATOR 1 < (int4, int4); -- operator 1 requested again in separate statement`,
				ErrorString: `operator 1(integer,integer) already exists in operator family "alt_opf17"`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf17 USING btree ADD
  OPERATOR 1 < (int4, int2) ,
  OPERATOR 2 <= (int4, int2) ,
  OPERATOR 3 = (int4, int2) ,
  OPERATOR 4 >= (int4, int2) ,
  OPERATOR 5 > (int4, int2) ,
  FUNCTION 1 btint42cmp(int4, int2) ,
  FUNCTION 1 btint42cmp(int4, int2);    -- procedure 1 appears twice in same statement`,
				ErrorString: `function number 1 for (integer,smallint) appears more than once`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf17 USING btree ADD
  OPERATOR 1 < (int4, int2) ,
  OPERATOR 2 <= (int4, int2) ,
  OPERATOR 3 = (int4, int2) ,
  OPERATOR 4 >= (int4, int2) ,
  OPERATOR 5 > (int4, int2) ,
  FUNCTION 1 btint42cmp(int4, int2);    -- procedure 1 appears first time`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf17 USING btree ADD
  OPERATOR 1 < (int4, int2) ,
  OPERATOR 2 <= (int4, int2) ,
  OPERATOR 3 = (int4, int2) ,
  OPERATOR 4 >= (int4, int2) ,
  OPERATOR 5 > (int4, int2) ,
  FUNCTION 1 btint42cmp(int4, int2);    -- procedure 1 requested again in separate statement`,
				ErrorString: `operator 1(integer,smallint) already exists in operator family "alt_opf17"`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf17 USING btree;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf18 USING btree;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf18 USING btree DROP OPERATOR 1 (int4, int4);`,
				ErrorString: `operator 1(integer,integer) does not exist in operator family "alt_opf18"`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf18 USING btree ADD
  OPERATOR 1 < (int4, int2) ,
  OPERATOR 2 <= (int4, int2) ,
  OPERATOR 3 = (int4, int2) ,
  OPERATOR 4 >= (int4, int2) ,
  OPERATOR 5 > (int4, int2) ,
  FUNCTION 1 btint42cmp(int4, int2);`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf18 USING btree
  ADD FUNCTION 4 (int4, int2) btequalimage(oid);`,
				ErrorString: `btree equal image functions must not be cross-type`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf18 USING btree DROP FUNCTION 2 (int4, int4);`,
				ErrorString: `function 2(integer,integer) does not exist in operator family "alt_opf18"`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf18 USING btree;`,
			},
			{
				Statement: `CREATE OPERATOR FAMILY alt_opf19 USING btree;`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf19 USING btree ADD FUNCTION 5 test_opclass_options_func(internal, text[], bool);`,
				ErrorString: `function test_opclass_options_func(internal, text[], boolean) does not exist`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf19 USING btree ADD FUNCTION 5 (int4) btint42cmp(int4, int2);`,
				ErrorString: `invalid operator class options parsing function`,
			},
			{
				Statement:   `ALTER OPERATOR FAMILY alt_opf19 USING btree ADD FUNCTION 5 (int4, int2) btint42cmp(int4, int2);`,
				ErrorString: `left and right associated data types for operator class options parsing functions must match`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf19 USING btree ADD FUNCTION 5 (int4) test_opclass_options_func(internal); -- Ok`,
			},
			{
				Statement: `ALTER OPERATOR FAMILY alt_opf19 USING btree DROP FUNCTION 5 (int4, int4);`,
			},
			{
				Statement: `DROP OPERATOR FAMILY alt_opf19 USING btree;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user1;`,
			},
			{
				Statement: `CREATE TABLE alt_regress_1 (a INTEGER, b INTEGER);`,
			},
			{
				Statement: `CREATE STATISTICS alt_stat1 ON a, b FROM alt_regress_1;`,
			},
			{
				Statement: `CREATE STATISTICS alt_stat2 ON a, b FROM alt_regress_1;`,
			},
			{
				Statement:   `ALTER STATISTICS alt_stat1 RENAME TO alt_stat2;   -- failed (name conflict)`,
				ErrorString: `statistics object "alt_stat2" already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER STATISTICS alt_stat1 RENAME TO alt_stat3;   -- OK`,
			},
			{
				Statement:   `ALTER STATISTICS alt_stat2 OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER STATISTICS alt_stat2 OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER STATISTICS alt_stat2 SET SCHEMA alt_nsp2;    -- OK`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user2;`,
			},
			{
				Statement: `CREATE TABLE alt_regress_2 (a INTEGER, b INTEGER);`,
			},
			{
				Statement: `CREATE STATISTICS alt_stat1 ON a, b FROM alt_regress_2;`,
			},
			{
				Statement: `CREATE STATISTICS alt_stat2 ON a, b FROM alt_regress_2;`,
			},
			{
				Statement:   `ALTER STATISTICS alt_stat3 RENAME TO alt_stat4;    -- failed (not owner)`,
				ErrorString: `must be owner of statistics object alt_stat3`,
			},
			{
				Statement: `ALTER STATISTICS alt_stat1 RENAME TO alt_stat4;    -- OK`,
			},
			{
				Statement:   `ALTER STATISTICS alt_stat3 OWNER TO regress_alter_generic_user2; -- failed (not owner)`,
				ErrorString: `must be owner of statistics object alt_stat3`,
			},
			{
				Statement:   `ALTER STATISTICS alt_stat2 OWNER TO regress_alter_generic_user3; -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER STATISTICS alt_stat3 SET SCHEMA alt_nsp2;		-- failed (not owner)`,
				ErrorString: `must be owner of statistics object alt_stat3`,
			},
			{
				Statement:   `ALTER STATISTICS alt_stat2 SET SCHEMA alt_nsp2;		-- failed (name conflict)`,
				ErrorString: `statistics object "alt_stat2" already exists in schema "alt_nsp2"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT nspname, stxname, rolname
  FROM pg_statistic_ext s, pg_namespace n, pg_authid a
 WHERE s.stxnamespace = n.oid AND s.stxowner = a.oid
   AND n.nspname in ('alt_nsp1', 'alt_nsp2')
 ORDER BY nspname, stxname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_stat2`, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_stat3`, `regress_alter_generic_user1`}, {`alt_nsp1`, `alt_stat4`, `regress_alter_generic_user2`}, {`alt_nsp2`, `alt_stat2`, `regress_alter_generic_user3`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user1;`,
			},
			{
				Statement: `CREATE TEXT SEARCH DICTIONARY alt_ts_dict1 (template=simple);`,
			},
			{
				Statement: `CREATE TEXT SEARCH DICTIONARY alt_ts_dict2 (template=simple);`,
			},
			{
				Statement:   `ALTER TEXT SEARCH DICTIONARY alt_ts_dict1 RENAME TO alt_ts_dict2;  -- failed (name conflict)`,
				ErrorString: `text search dictionary "alt_ts_dict2" already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER TEXT SEARCH DICTIONARY alt_ts_dict1 RENAME TO alt_ts_dict3;  -- OK`,
			},
			{
				Statement:   `ALTER TEXT SEARCH DICTIONARY alt_ts_dict2 OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER TEXT SEARCH DICTIONARY alt_ts_dict2 OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER TEXT SEARCH DICTIONARY alt_ts_dict2 SET SCHEMA alt_nsp2;  -- OK`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user2;`,
			},
			{
				Statement: `CREATE TEXT SEARCH DICTIONARY alt_ts_dict1 (template=simple);`,
			},
			{
				Statement: `CREATE TEXT SEARCH DICTIONARY alt_ts_dict2 (template=simple);`,
			},
			{
				Statement:   `ALTER TEXT SEARCH DICTIONARY alt_ts_dict3 RENAME TO alt_ts_dict4;  -- failed (not owner)`,
				ErrorString: `must be owner of text search dictionary alt_ts_dict3`,
			},
			{
				Statement: `ALTER TEXT SEARCH DICTIONARY alt_ts_dict1 RENAME TO alt_ts_dict4;  -- OK`,
			},
			{
				Statement:   `ALTER TEXT SEARCH DICTIONARY alt_ts_dict3 OWNER TO regress_alter_generic_user2;  -- failed (not owner)`,
				ErrorString: `must be owner of text search dictionary alt_ts_dict3`,
			},
			{
				Statement:   `ALTER TEXT SEARCH DICTIONARY alt_ts_dict2 OWNER TO regress_alter_generic_user3;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER TEXT SEARCH DICTIONARY alt_ts_dict3 SET SCHEMA alt_nsp2;  -- failed (not owner)`,
				ErrorString: `must be owner of text search dictionary alt_ts_dict3`,
			},
			{
				Statement:   `ALTER TEXT SEARCH DICTIONARY alt_ts_dict2 SET SCHEMA alt_nsp2;  -- failed (name conflict)`,
				ErrorString: `text search dictionary "alt_ts_dict2" already exists in schema "alt_nsp2"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT nspname, dictname, rolname
  FROM pg_ts_dict t, pg_namespace n, pg_authid a
  WHERE t.dictnamespace = n.oid AND t.dictowner = a.oid
    AND n.nspname in ('alt_nsp1', 'alt_nsp2')
  ORDER BY nspname, dictname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_ts_dict2`, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_ts_dict3`, `regress_alter_generic_user1`}, {`alt_nsp1`, `alt_ts_dict4`, `regress_alter_generic_user2`}, {`alt_nsp2`, `alt_ts_dict2`, `regress_alter_generic_user3`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user1;`,
			},
			{
				Statement: `CREATE TEXT SEARCH CONFIGURATION alt_ts_conf1 (copy=english);`,
			},
			{
				Statement: `CREATE TEXT SEARCH CONFIGURATION alt_ts_conf2 (copy=english);`,
			},
			{
				Statement:   `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf1 RENAME TO alt_ts_conf2;  -- failed (name conflict)`,
				ErrorString: `text search configuration "alt_ts_conf2" already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf1 RENAME TO alt_ts_conf3;  -- OK`,
			},
			{
				Statement:   `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf2 OWNER TO regress_alter_generic_user2;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user2"`,
			},
			{
				Statement: `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf2 OWNER TO regress_alter_generic_user3;  -- OK`,
			},
			{
				Statement: `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf2 SET SCHEMA alt_nsp2;  -- OK`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_generic_user2;`,
			},
			{
				Statement: `CREATE TEXT SEARCH CONFIGURATION alt_ts_conf1 (copy=english);`,
			},
			{
				Statement: `CREATE TEXT SEARCH CONFIGURATION alt_ts_conf2 (copy=english);`,
			},
			{
				Statement:   `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf3 RENAME TO alt_ts_conf4;  -- failed (not owner)`,
				ErrorString: `must be owner of text search configuration alt_ts_conf3`,
			},
			{
				Statement: `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf1 RENAME TO alt_ts_conf4;  -- OK`,
			},
			{
				Statement:   `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf3 OWNER TO regress_alter_generic_user2;  -- failed (not owner)`,
				ErrorString: `must be owner of text search configuration alt_ts_conf3`,
			},
			{
				Statement:   `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf2 OWNER TO regress_alter_generic_user3;  -- failed (no role membership)`,
				ErrorString: `must be member of role "regress_alter_generic_user3"`,
			},
			{
				Statement:   `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf3 SET SCHEMA alt_nsp2;  -- failed (not owner)`,
				ErrorString: `must be owner of text search configuration alt_ts_conf3`,
			},
			{
				Statement:   `ALTER TEXT SEARCH CONFIGURATION alt_ts_conf2 SET SCHEMA alt_nsp2;  -- failed (name conflict)`,
				ErrorString: `text search configuration "alt_ts_conf2" already exists in schema "alt_nsp2"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SELECT nspname, cfgname, rolname
  FROM pg_ts_config t, pg_namespace n, pg_authid a
  WHERE t.cfgnamespace = n.oid AND t.cfgowner = a.oid
    AND n.nspname in ('alt_nsp1', 'alt_nsp2')
  ORDER BY nspname, cfgname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_ts_conf2`, `regress_alter_generic_user2`}, {`alt_nsp1`, `alt_ts_conf3`, `regress_alter_generic_user1`}, {`alt_nsp1`, `alt_ts_conf4`, `regress_alter_generic_user2`}, {`alt_nsp2`, `alt_ts_conf2`, `regress_alter_generic_user3`}},
			},
			{
				Statement: `CREATE TEXT SEARCH TEMPLATE alt_ts_temp1 (lexize=dsimple_lexize);`,
			},
			{
				Statement: `CREATE TEXT SEARCH TEMPLATE alt_ts_temp2 (lexize=dsimple_lexize);`,
			},
			{
				Statement:   `ALTER TEXT SEARCH TEMPLATE alt_ts_temp1 RENAME TO alt_ts_temp2; -- failed (name conflict)`,
				ErrorString: `text search template "alt_ts_temp2" already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER TEXT SEARCH TEMPLATE alt_ts_temp1 RENAME TO alt_ts_temp3; -- OK`,
			},
			{
				Statement: `ALTER TEXT SEARCH TEMPLATE alt_ts_temp2 SET SCHEMA alt_nsp2;    -- OK`,
			},
			{
				Statement: `CREATE TEXT SEARCH TEMPLATE alt_ts_temp2 (lexize=dsimple_lexize);`,
			},
			{
				Statement:   `ALTER TEXT SEARCH TEMPLATE alt_ts_temp2 SET SCHEMA alt_nsp2;    -- failed (name conflict)`,
				ErrorString: `text search template "alt_ts_temp2" already exists in schema "alt_nsp2"`,
			},
			{
				Statement:   `CREATE TEXT SEARCH TEMPLATE tstemp_case ("Init" = init_function);`,
				ErrorString: `text search template parameter "Init" not recognized`,
			},
			{
				Statement: `SELECT nspname, tmplname
  FROM pg_ts_template t, pg_namespace n
  WHERE t.tmplnamespace = n.oid AND nspname like 'alt_nsp%'
  ORDER BY nspname, tmplname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_ts_temp2`}, {`alt_nsp1`, `alt_ts_temp3`}, {`alt_nsp2`, `alt_ts_temp2`}},
			},
			{
				Statement: `CREATE TEXT SEARCH PARSER alt_ts_prs1
    (start = prsd_start, gettoken = prsd_nexttoken, end = prsd_end, lextypes = prsd_lextype);`,
			},
			{
				Statement: `CREATE TEXT SEARCH PARSER alt_ts_prs2
    (start = prsd_start, gettoken = prsd_nexttoken, end = prsd_end, lextypes = prsd_lextype);`,
			},
			{
				Statement:   `ALTER TEXT SEARCH PARSER alt_ts_prs1 RENAME TO alt_ts_prs2; -- failed (name conflict)`,
				ErrorString: `text search parser "alt_ts_prs2" already exists in schema "alt_nsp1"`,
			},
			{
				Statement: `ALTER TEXT SEARCH PARSER alt_ts_prs1 RENAME TO alt_ts_prs3; -- OK`,
			},
			{
				Statement: `ALTER TEXT SEARCH PARSER alt_ts_prs2 SET SCHEMA alt_nsp2;   -- OK`,
			},
			{
				Statement: `CREATE TEXT SEARCH PARSER alt_ts_prs2
    (start = prsd_start, gettoken = prsd_nexttoken, end = prsd_end, lextypes = prsd_lextype);`,
			},
			{
				Statement:   `ALTER TEXT SEARCH PARSER alt_ts_prs2 SET SCHEMA alt_nsp2;   -- failed (name conflict)`,
				ErrorString: `text search parser "alt_ts_prs2" already exists in schema "alt_nsp2"`,
			},
			{
				Statement:   `CREATE TEXT SEARCH PARSER tspars_case ("Start" = start_function);`,
				ErrorString: `text search parser parameter "Start" not recognized`,
			},
			{
				Statement: `SELECT nspname, prsname
  FROM pg_ts_parser t, pg_namespace n
  WHERE t.prsnamespace = n.oid AND nspname like 'alt_nsp%'
  ORDER BY nspname, prsname;`,
				Results: []sql.Row{{`alt_nsp1`, `alt_ts_prs2`}, {`alt_nsp1`, `alt_ts_prs3`}, {`alt_nsp2`, `alt_ts_prs2`}},
			},
			{
				Statement: `---
---
DROP FOREIGN DATA WRAPPER alt_fdw2 CASCADE;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER alt_fdw3 CASCADE;`,
			},
			{
				Statement: `DROP LANGUAGE alt_lang2 CASCADE;`,
			},
			{
				Statement: `DROP LANGUAGE alt_lang3 CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA alt_nsp1 CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA alt_nsp2 CASCADE;`,
			},
			{
				Statement: `DROP USER regress_alter_generic_user1;`,
			},
			{
				Statement: `DROP USER regress_alter_generic_user2;`,
			},
			{
				Statement: `DROP USER regress_alter_generic_user3;`,
			},
		},
	})
}
