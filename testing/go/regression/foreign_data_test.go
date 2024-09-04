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

func TestForeignData(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_foreign_data)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_foreign_data,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set regresslib :libdir '/regress' :dlsuffix
CREATE FUNCTION test_fdw_handler()
    RETURNS fdw_handler
    AS :'regresslib', 'test_fdw_handler'
    LANGUAGE C;`,
			},
			{
				Statement: `SET client_min_messages TO 'warning';`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_foreign_data_user, regress_test_role, regress_test_role2, regress_test_role_super, regress_test_indirect, regress_unprivileged_role;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE ROLE regress_foreign_data_user LOGIN SUPERUSER;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION 'regress_foreign_data_user';`,
			},
			{
				Statement: `CREATE ROLE regress_test_role;`,
			},
			{
				Statement: `CREATE ROLE regress_test_role2;`,
			},
			{
				Statement: `CREATE ROLE regress_test_role_super SUPERUSER;`,
			},
			{
				Statement: `CREATE ROLE regress_test_indirect;`,
			},
			{
				Statement: `CREATE ROLE regress_unprivileged_role;`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER dummy;`,
			},
			{
				Statement: `COMMENT ON FOREIGN DATA WRAPPER dummy IS 'useless';`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER postgresql VALIDATOR postgresql_fdw_validator;`,
			},
			{
				Statement: `SELECT fdwname, fdwhandler::regproc, fdwvalidator::regproc, fdwoptions FROM pg_foreign_data_wrapper ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{`dummy`, `-`, `-`, ``}, {`postgresql`, `-`, `postgresql_fdw_validator`, ``}},
			},
			{
				Statement: `SELECT srvname, srvoptions FROM pg_foreign_server;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_user_mapping;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `CREATE FOREIGN DATA WRAPPER foo VALIDATOR bar;            -- ERROR`,
				ErrorString: `function bar(text[], oid) does not exist`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `\dew
                        List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         
------------+---------------------------+---------+--------------------------
 dummy      | regress_foreign_data_user | -       | -
 foo        | regress_foreign_data_user | -       | -
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator
(3 rows)
CREATE FOREIGN DATA WRAPPER foo; -- duplicate`,
				ErrorString: `foreign-data wrapper "foo" already exists`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER foo OPTIONS (testing '1');`,
			},
			{
				Statement: `\dew+
                                                 List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges |  FDW options  | Description 
------------+---------------------------+---------+--------------------------+-------------------+---------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |               | useless
 foo        | regress_foreign_data_user | -       | -                        |                   | (testing '1') | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |               | 
(3 rows)
DROP FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement:   `CREATE FOREIGN DATA WRAPPER foo OPTIONS (testing '1', testing '2');   -- ERROR`,
				ErrorString: `option "testing" provided more than once`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER foo OPTIONS (testing '1', another '2');`,
			},
			{
				Statement: `\dew+
                                                       List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges |        FDW options         | Description 
------------+---------------------------+---------+--------------------------+-------------------+----------------------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |                            | useless
 foo        | regress_foreign_data_user | -       | -                        |                   | (testing '1', another '2') | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |                            | 
(3 rows)
DROP FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement:   `CREATE FOREIGN DATA WRAPPER foo; -- ERROR`,
				ErrorString: `permission denied to create foreign-data wrapper "foo"`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER foo VALIDATOR postgresql_fdw_validator;`,
			},
			{
				Statement: `\dew+
                                                List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges | FDW options | Description 
------------+---------------------------+---------+--------------------------+-------------------+-------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |             | useless
 foo        | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |             | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |             | 
(3 rows)
CREATE FUNCTION invalid_fdw_handler() RETURNS int LANGUAGE SQL AS 'SELECT 1;';`,
			},
			{
				Statement:   `CREATE FOREIGN DATA WRAPPER test_fdw HANDLER invalid_fdw_handler;  -- ERROR`,
				ErrorString: `function invalid_fdw_handler must return type fdw_handler`,
			},
			{
				Statement:   `CREATE FOREIGN DATA WRAPPER test_fdw HANDLER test_fdw_handler HANDLER invalid_fdw_handler;  -- ERROR`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER test_fdw HANDLER test_fdw_handler;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER test_fdw;`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OPTIONS (nonexistent 'fdw');         -- ERROR`,
				ErrorString: `invalid option "nonexistent"`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo;                             -- ERROR`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo VALIDATOR bar;               -- ERROR`,
				ErrorString: `function bar(text[], oid) does not exist`,
			},
			{
				Statement: `ALTER FOREIGN DATA WRAPPER foo NO VALIDATOR;`,
			},
			{
				Statement: `\dew+
                                                List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges | FDW options | Description 
------------+---------------------------+---------+--------------------------+-------------------+-------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |             | useless
 foo        | regress_foreign_data_user | -       | -                        |                   |             | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |             | 
(3 rows)
ALTER FOREIGN DATA WRAPPER foo OPTIONS (a '1', b '2');`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OPTIONS (SET c '4');         -- ERROR`,
				ErrorString: `option "c" not found`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OPTIONS (DROP c);            -- ERROR`,
				ErrorString: `option "c" not found`,
			},
			{
				Statement: `ALTER FOREIGN DATA WRAPPER foo OPTIONS (ADD x '1', DROP x);`,
			},
			{
				Statement: `\dew+
                                                 List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges |  FDW options   | Description 
------------+---------------------------+---------+--------------------------+-------------------+----------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |                | useless
 foo        | regress_foreign_data_user | -       | -                        |                   | (a '1', b '2') | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |                | 
(3 rows)
ALTER FOREIGN DATA WRAPPER foo OPTIONS (DROP a, SET b '3', ADD c '4');`,
			},
			{
				Statement: `\dew+
                                                 List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges |  FDW options   | Description 
------------+---------------------------+---------+--------------------------+-------------------+----------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |                | useless
 foo        | regress_foreign_data_user | -       | -                        |                   | (b '3', c '4') | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |                | 
(3 rows)
ALTER FOREIGN DATA WRAPPER foo OPTIONS (a '2');`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OPTIONS (b '4');             -- ERROR`,
				ErrorString: `option "b" provided more than once`,
			},
			{
				Statement: `\dew+
                                                     List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges |      FDW options      | Description 
------------+---------------------------+---------+--------------------------+-------------------+-----------------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |                       | useless
 foo        | regress_foreign_data_user | -       | -                        |                   | (b '3', c '4', a '2') | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |                       | 
(3 rows)
SET ROLE regress_test_role;`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OPTIONS (ADD d '5');         -- ERROR`,
				ErrorString: `permission denied to alter foreign-data wrapper "foo"`,
			},
			{
				Statement: `SET ROLE regress_test_role_super;`,
			},
			{
				Statement: `ALTER FOREIGN DATA WRAPPER foo OPTIONS (ADD d '5');`,
			},
			{
				Statement: `\dew+
                                                        List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges |         FDW options          | Description 
------------+---------------------------+---------+--------------------------+-------------------+------------------------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |                              | useless
 foo        | regress_foreign_data_user | -       | -                        |                   | (b '3', c '4', a '2', d '5') | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |                              | 
(3 rows)
ALTER FOREIGN DATA WRAPPER foo OWNER TO regress_test_role;  -- ERROR`,
				ErrorString: `permission denied to change owner of foreign-data wrapper "foo"`,
			},
			{
				Statement: `ALTER FOREIGN DATA WRAPPER foo OWNER TO regress_test_role_super;`,
			},
			{
				Statement: `ALTER ROLE regress_test_role_super NOSUPERUSER;`,
			},
			{
				Statement: `SET ROLE regress_test_role_super;`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OPTIONS (ADD e '6');         -- ERROR`,
				ErrorString: `permission denied to alter foreign-data wrapper "foo"`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `\dew+
                                                        List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges |         FDW options          | Description 
------------+---------------------------+---------+--------------------------+-------------------+------------------------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |                              | useless
 foo        | regress_test_role_super   | -       | -                        |                   | (b '3', c '4', a '2', d '5') | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |                              | 
(3 rows)
ALTER FOREIGN DATA WRAPPER foo RENAME TO foo1;`,
			},
			{
				Statement: `\dew+
                                                        List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges |         FDW options          | Description 
------------+---------------------------+---------+--------------------------+-------------------+------------------------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |                              | useless
 foo1       | regress_test_role_super   | -       | -                        |                   | (b '3', c '4', a '2', d '5') | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |                              | 
(3 rows)
ALTER FOREIGN DATA WRAPPER foo1 RENAME TO foo;`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo HANDLER invalid_fdw_handler;  -- ERROR`,
				ErrorString: `function invalid_fdw_handler must return type fdw_handler`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo HANDLER test_fdw_handler HANDLER anything;  -- ERROR`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement: `ALTER FOREIGN DATA WRAPPER foo HANDLER test_fdw_handler;`,
			},
			{
				Statement: `DROP FUNCTION invalid_fdw_handler();`,
			},
			{
				Statement:   `DROP FOREIGN DATA WRAPPER nonexistent;                      -- ERROR`,
				ErrorString: `foreign-data wrapper "nonexistent" does not exist`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER IF EXISTS nonexistent;`,
			},
			{
				Statement: `\dew+
                                                             List of foreign-data wrappers
    Name    |           Owner           |     Handler      |        Validator         | Access privileges |         FDW options          | Description 
------------+---------------------------+------------------+--------------------------+-------------------+------------------------------+-------------
 dummy      | regress_foreign_data_user | -                | -                        |                   |                              | useless
 foo        | regress_test_role_super   | test_fdw_handler | -                        |                   | (b '3', c '4', a '2', d '5') | 
 postgresql | regress_foreign_data_user | -                | postgresql_fdw_validator |                   |                              | 
(3 rows)
DROP ROLE regress_test_role_super;                          -- ERROR`,
				ErrorString: `role "regress_test_role_super" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `SET ROLE regress_test_role_super;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `DROP ROLE regress_test_role_super;`,
			},
			{
				Statement: `\dew+
                                                List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges | FDW options | Description 
------------+---------------------------+---------+--------------------------+-------------------+-------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |             | useless
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |             | 
(2 rows)
CREATE FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `CREATE SERVER s1 FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `COMMENT ON SERVER s1 IS 'foreign server';`,
			},
			{
				Statement: `CREATE USER MAPPING FOR current_user SERVER s1;`,
			},
			{
				Statement:   `CREATE USER MAPPING FOR current_user SERVER s1;				-- ERROR`,
				ErrorString: `user mapping for "regress_foreign_data_user" already exists for server "s1"`,
			},
			{
				Statement: `CREATE USER MAPPING IF NOT EXISTS FOR current_user SERVER s1; -- NOTICE`,
			},
			{
				Statement: `\dew+
                                                List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges | FDW options | Description 
------------+---------------------------+---------+--------------------------+-------------------+-------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |             | useless
 foo        | regress_foreign_data_user | -       | -                        |                   |             | 
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |             | 
(3 rows)
\des+
                                                   List of foreign servers
 Name |           Owner           | Foreign-data wrapper | Access privileges | Type | Version | FDW options |  Description   
------+---------------------------+----------------------+-------------------+------+---------+-------------+----------------
 s1   | regress_foreign_data_user | foo                  |                   |      |         |             | foreign server
(1 row)
\deu+
              List of user mappings
 Server |         User name         | FDW options 
--------+---------------------------+-------------
 s1     | regress_foreign_data_user | 
(1 row)
DROP FOREIGN DATA WRAPPER foo;                              -- ERROR`,
				ErrorString: `cannot drop foreign-data wrapper foo because other objects depend on it`,
			},
			{
				Statement: `user mapping for regress_foreign_data_user on server s1 depends on server s1
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
SET ROLE regress_test_role;`,
			},
			{
				Statement:   `DROP FOREIGN DATA WRAPPER foo CASCADE;                      -- ERROR`,
				ErrorString: `must be owner of foreign-data wrapper foo`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER foo CASCADE;`,
			},
			{
				Statement: `\dew+
                                                List of foreign-data wrappers
    Name    |           Owner           | Handler |        Validator         | Access privileges | FDW options | Description 
------------+---------------------------+---------+--------------------------+-------------------+-------------+-------------
 dummy      | regress_foreign_data_user | -       | -                        |                   |             | useless
 postgresql | regress_foreign_data_user | -       | postgresql_fdw_validator |                   |             | 
(2 rows)
\des+
                                       List of foreign servers
 Name | Owner | Foreign-data wrapper | Access privileges | Type | Version | FDW options | Description 
------+-------+----------------------+-------------------+------+---------+-------------+-------------
(0 rows)
\deu+
      List of user mappings
 Server | User name | FDW options 
--------+-----------+-------------
(0 rows)
CREATE SERVER s1 FOREIGN DATA WRAPPER foo;                  -- ERROR`,
				ErrorString: `foreign-data wrapper "foo" does not exist`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER foo OPTIONS ("test wrapper" 'true');`,
			},
			{
				Statement: `CREATE SERVER s1 FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement:   `CREATE SERVER s1 FOREIGN DATA WRAPPER foo;                  -- ERROR`,
				ErrorString: `server "s1" already exists`,
			},
			{
				Statement: `CREATE SERVER IF NOT EXISTS s1 FOREIGN DATA WRAPPER foo;	-- No ERROR, just NOTICE`,
			},
			{
				Statement: `CREATE SERVER s2 FOREIGN DATA WRAPPER foo OPTIONS (host 'a', dbname 'b');`,
			},
			{
				Statement: `CREATE SERVER s3 TYPE 'oracle' FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `CREATE SERVER s4 TYPE 'oracle' FOREIGN DATA WRAPPER foo OPTIONS (host 'a', dbname 'b');`,
			},
			{
				Statement: `CREATE SERVER s5 VERSION '15.0' FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `CREATE SERVER s6 VERSION '16.0' FOREIGN DATA WRAPPER foo OPTIONS (host 'a', dbname 'b');`,
			},
			{
				Statement: `CREATE SERVER s7 TYPE 'oracle' VERSION '17.0' FOREIGN DATA WRAPPER foo OPTIONS (host 'a', dbname 'b');`,
			},
			{
				Statement:   `CREATE SERVER s8 FOREIGN DATA WRAPPER postgresql OPTIONS (foo '1'); -- ERROR`,
				ErrorString: `invalid option "foo"`,
			},
			{
				Statement: `CREATE SERVER s8 FOREIGN DATA WRAPPER postgresql OPTIONS (host 'localhost', dbname 's8db');`,
			},
			{
				Statement: `\des+
                                                             List of foreign servers
 Name |           Owner           | Foreign-data wrapper | Access privileges |  Type  | Version |            FDW options            | Description 
------+---------------------------+----------------------+-------------------+--------+---------+-----------------------------------+-------------
 s1   | regress_foreign_data_user | foo                  |                   |        |         |                                   | 
 s2   | regress_foreign_data_user | foo                  |                   |        |         | (host 'a', dbname 'b')            | 
 s3   | regress_foreign_data_user | foo                  |                   | oracle |         |                                   | 
 s4   | regress_foreign_data_user | foo                  |                   | oracle |         | (host 'a', dbname 'b')            | 
 s5   | regress_foreign_data_user | foo                  |                   |        | 15.0    |                                   | 
 s6   | regress_foreign_data_user | foo                  |                   |        | 16.0    | (host 'a', dbname 'b')            | 
 s7   | regress_foreign_data_user | foo                  |                   | oracle | 17.0    | (host 'a', dbname 'b')            | 
 s8   | regress_foreign_data_user | postgresql           |                   |        |         | (host 'localhost', dbname 's8db') | 
(8 rows)
SET ROLE regress_test_role;`,
			},
			{
				Statement:   `CREATE SERVER t1 FOREIGN DATA WRAPPER foo;                 -- ERROR: no usage on FDW`,
				ErrorString: `permission denied for foreign-data wrapper foo`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN DATA WRAPPER foo TO regress_test_role;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `CREATE SERVER t1 FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `\des+
                                                             List of foreign servers
 Name |           Owner           | Foreign-data wrapper | Access privileges |  Type  | Version |            FDW options            | Description 
------+---------------------------+----------------------+-------------------+--------+---------+-----------------------------------+-------------
 s1   | regress_foreign_data_user | foo                  |                   |        |         |                                   | 
 s2   | regress_foreign_data_user | foo                  |                   |        |         | (host 'a', dbname 'b')            | 
 s3   | regress_foreign_data_user | foo                  |                   | oracle |         |                                   | 
 s4   | regress_foreign_data_user | foo                  |                   | oracle |         | (host 'a', dbname 'b')            | 
 s5   | regress_foreign_data_user | foo                  |                   |        | 15.0    |                                   | 
 s6   | regress_foreign_data_user | foo                  |                   |        | 16.0    | (host 'a', dbname 'b')            | 
 s7   | regress_foreign_data_user | foo                  |                   | oracle | 17.0    | (host 'a', dbname 'b')            | 
 s8   | regress_foreign_data_user | postgresql           |                   |        |         | (host 'localhost', dbname 's8db') | 
 t1   | regress_test_role         | foo                  |                   |        |         |                                   | 
(9 rows)
REVOKE USAGE ON FOREIGN DATA WRAPPER foo FROM regress_test_role;`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN DATA WRAPPER foo TO regress_test_indirect;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement:   `CREATE SERVER t2 FOREIGN DATA WRAPPER foo;                 -- ERROR`,
				ErrorString: `permission denied for foreign-data wrapper foo`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT regress_test_indirect TO regress_test_role;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `CREATE SERVER t2 FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `\des+
                                                             List of foreign servers
 Name |           Owner           | Foreign-data wrapper | Access privileges |  Type  | Version |            FDW options            | Description 
------+---------------------------+----------------------+-------------------+--------+---------+-----------------------------------+-------------
 s1   | regress_foreign_data_user | foo                  |                   |        |         |                                   | 
 s2   | regress_foreign_data_user | foo                  |                   |        |         | (host 'a', dbname 'b')            | 
 s3   | regress_foreign_data_user | foo                  |                   | oracle |         |                                   | 
 s4   | regress_foreign_data_user | foo                  |                   | oracle |         | (host 'a', dbname 'b')            | 
 s5   | regress_foreign_data_user | foo                  |                   |        | 15.0    |                                   | 
 s6   | regress_foreign_data_user | foo                  |                   |        | 16.0    | (host 'a', dbname 'b')            | 
 s7   | regress_foreign_data_user | foo                  |                   | oracle | 17.0    | (host 'a', dbname 'b')            | 
 s8   | regress_foreign_data_user | postgresql           |                   |        |         | (host 'localhost', dbname 's8db') | 
 t1   | regress_test_role         | foo                  |                   |        |         |                                   | 
 t2   | regress_test_role         | foo                  |                   |        |         |                                   | 
(10 rows)
RESET ROLE;`,
			},
			{
				Statement: `REVOKE regress_test_indirect FROM regress_test_role;`,
			},
			{
				Statement:   `ALTER SERVER s0;                                            -- ERROR`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `ALTER SERVER s0 OPTIONS (a '1');                            -- ERROR`,
				ErrorString: `server "s0" does not exist`,
			},
			{
				Statement: `ALTER SERVER s1 VERSION '1.0' OPTIONS (servername 's1');`,
			},
			{
				Statement: `ALTER SERVER s2 VERSION '1.1';`,
			},
			{
				Statement: `ALTER SERVER s3 OPTIONS ("tns name" 'orcl', port '1521');`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN SERVER s1 TO regress_test_role;`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN SERVER s6 TO regress_test_role2 WITH GRANT OPTION;`,
			},
			{
				Statement: `\des+
                                                                               List of foreign servers
 Name |           Owner           | Foreign-data wrapper |                   Access privileges                   |  Type  | Version |            FDW options            | Description 
------+---------------------------+----------------------+-------------------------------------------------------+--------+---------+-----------------------------------+-------------
 s1   | regress_foreign_data_user | foo                  | regress_foreign_data_user=U/regress_foreign_data_user+|        | 1.0     | (servername 's1')                 | 
      |                           |                      | regress_test_role=U/regress_foreign_data_user         |        |         |                                   | 
 s2   | regress_foreign_data_user | foo                  |                                                       |        | 1.1     | (host 'a', dbname 'b')            | 
 s3   | regress_foreign_data_user | foo                  |                                                       | oracle |         | ("tns name" 'orcl', port '1521')  | 
 s4   | regress_foreign_data_user | foo                  |                                                       | oracle |         | (host 'a', dbname 'b')            | 
 s5   | regress_foreign_data_user | foo                  |                                                       |        | 15.0    |                                   | 
 s6   | regress_foreign_data_user | foo                  | regress_foreign_data_user=U/regress_foreign_data_user+|        | 16.0    | (host 'a', dbname 'b')            | 
      |                           |                      | regress_test_role2=U*/regress_foreign_data_user       |        |         |                                   | 
 s7   | regress_foreign_data_user | foo                  |                                                       | oracle | 17.0    | (host 'a', dbname 'b')            | 
 s8   | regress_foreign_data_user | postgresql           |                                                       |        |         | (host 'localhost', dbname 's8db') | 
 t1   | regress_test_role         | foo                  |                                                       |        |         |                                   | 
 t2   | regress_test_role         | foo                  |                                                       |        |         |                                   | 
(10 rows)
SET ROLE regress_test_role;`,
			},
			{
				Statement:   `ALTER SERVER s1 VERSION '1.1';                              -- ERROR`,
				ErrorString: `must be owner of foreign server s1`,
			},
			{
				Statement:   `ALTER SERVER s1 OWNER TO regress_test_role;                 -- ERROR`,
				ErrorString: `must be owner of foreign server s1`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `ALTER SERVER s1 OWNER TO regress_test_role;`,
			},
			{
				Statement: `GRANT regress_test_role2 TO regress_test_role;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `ALTER SERVER s1 VERSION '1.1';`,
			},
			{
				Statement:   `ALTER SERVER s1 OWNER TO regress_test_role2;                -- ERROR`,
				ErrorString: `permission denied for foreign-data wrapper foo`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement:   `ALTER SERVER s8 OPTIONS (foo '1');                          -- ERROR option validation`,
				ErrorString: `invalid option "foo"`,
			},
			{
				Statement: `ALTER SERVER s8 OPTIONS (connect_timeout '30', SET dbname 'db1', DROP host);`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement:   `ALTER SERVER s1 OWNER TO regress_test_indirect;             -- ERROR`,
				ErrorString: `must be member of role "regress_test_indirect"`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT regress_test_indirect TO regress_test_role;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `ALTER SERVER s1 OWNER TO regress_test_indirect;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN DATA WRAPPER foo TO regress_test_indirect;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `ALTER SERVER s1 OWNER TO regress_test_indirect;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement:   `DROP ROLE regress_test_indirect;                            -- ERROR`,
				ErrorString: `role "regress_test_indirect" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `owner of server s1
\des+
                                                                                 List of foreign servers
 Name |           Owner           | Foreign-data wrapper |                   Access privileges                   |  Type  | Version |             FDW options              | Description 
------+---------------------------+----------------------+-------------------------------------------------------+--------+---------+--------------------------------------+-------------
 s1   | regress_test_indirect     | foo                  | regress_test_indirect=U/regress_test_indirect         |        | 1.1     | (servername 's1')                    | 
 s2   | regress_foreign_data_user | foo                  |                                                       |        | 1.1     | (host 'a', dbname 'b')               | 
 s3   | regress_foreign_data_user | foo                  |                                                       | oracle |         | ("tns name" 'orcl', port '1521')     | 
 s4   | regress_foreign_data_user | foo                  |                                                       | oracle |         | (host 'a', dbname 'b')               | 
 s5   | regress_foreign_data_user | foo                  |                                                       |        | 15.0    |                                      | 
 s6   | regress_foreign_data_user | foo                  | regress_foreign_data_user=U/regress_foreign_data_user+|        | 16.0    | (host 'a', dbname 'b')               | 
      |                           |                      | regress_test_role2=U*/regress_foreign_data_user       |        |         |                                      | 
 s7   | regress_foreign_data_user | foo                  |                                                       | oracle | 17.0    | (host 'a', dbname 'b')               | 
 s8   | regress_foreign_data_user | postgresql           |                                                       |        |         | (dbname 'db1', connect_timeout '30') | 
 t1   | regress_test_role         | foo                  |                                                       |        |         |                                      | 
 t2   | regress_test_role         | foo                  |                                                       |        |         |                                      | 
(10 rows)
ALTER SERVER s8 RENAME to s8new;`,
			},
			{
				Statement: `\des+
                                                                                 List of foreign servers
 Name  |           Owner           | Foreign-data wrapper |                   Access privileges                   |  Type  | Version |             FDW options              | Description 
-------+---------------------------+----------------------+-------------------------------------------------------+--------+---------+--------------------------------------+-------------
 s1    | regress_test_indirect     | foo                  | regress_test_indirect=U/regress_test_indirect         |        | 1.1     | (servername 's1')                    | 
 s2    | regress_foreign_data_user | foo                  |                                                       |        | 1.1     | (host 'a', dbname 'b')               | 
 s3    | regress_foreign_data_user | foo                  |                                                       | oracle |         | ("tns name" 'orcl', port '1521')     | 
 s4    | regress_foreign_data_user | foo                  |                                                       | oracle |         | (host 'a', dbname 'b')               | 
 s5    | regress_foreign_data_user | foo                  |                                                       |        | 15.0    |                                      | 
 s6    | regress_foreign_data_user | foo                  | regress_foreign_data_user=U/regress_foreign_data_user+|        | 16.0    | (host 'a', dbname 'b')               | 
       |                           |                      | regress_test_role2=U*/regress_foreign_data_user       |        |         |                                      | 
 s7    | regress_foreign_data_user | foo                  |                                                       | oracle | 17.0    | (host 'a', dbname 'b')               | 
 s8new | regress_foreign_data_user | postgresql           |                                                       |        |         | (dbname 'db1', connect_timeout '30') | 
 t1    | regress_test_role         | foo                  |                                                       |        |         |                                      | 
 t2    | regress_test_role         | foo                  |                                                       |        |         |                                      | 
(10 rows)
ALTER SERVER s8new RENAME to s8;`,
			},
			{
				Statement:   `DROP SERVER nonexistent;                                    -- ERROR`,
				ErrorString: `server "nonexistent" does not exist`,
			},
			{
				Statement: `DROP SERVER IF EXISTS nonexistent;`,
			},
			{
				Statement: `\des
                 List of foreign servers
 Name |           Owner           | Foreign-data wrapper 
------+---------------------------+----------------------
 s1   | regress_test_indirect     | foo
 s2   | regress_foreign_data_user | foo
 s3   | regress_foreign_data_user | foo
 s4   | regress_foreign_data_user | foo
 s5   | regress_foreign_data_user | foo
 s6   | regress_foreign_data_user | foo
 s7   | regress_foreign_data_user | foo
 s8   | regress_foreign_data_user | postgresql
 t1   | regress_test_role         | foo
 t2   | regress_test_role         | foo
(10 rows)
SET ROLE regress_test_role;`,
			},
			{
				Statement:   `DROP SERVER s2;                                             -- ERROR`,
				ErrorString: `must be owner of foreign server s2`,
			},
			{
				Statement: `DROP SERVER s1;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `\des
                 List of foreign servers
 Name |           Owner           | Foreign-data wrapper 
------+---------------------------+----------------------
 s2   | regress_foreign_data_user | foo
 s3   | regress_foreign_data_user | foo
 s4   | regress_foreign_data_user | foo
 s5   | regress_foreign_data_user | foo
 s6   | regress_foreign_data_user | foo
 s7   | regress_foreign_data_user | foo
 s8   | regress_foreign_data_user | postgresql
 t1   | regress_test_role         | foo
 t2   | regress_test_role         | foo
(9 rows)
ALTER SERVER s2 OWNER TO regress_test_role;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `DROP SERVER s2;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `\des
                 List of foreign servers
 Name |           Owner           | Foreign-data wrapper 
------+---------------------------+----------------------
 s3   | regress_foreign_data_user | foo
 s4   | regress_foreign_data_user | foo
 s5   | regress_foreign_data_user | foo
 s6   | regress_foreign_data_user | foo
 s7   | regress_foreign_data_user | foo
 s8   | regress_foreign_data_user | postgresql
 t1   | regress_test_role         | foo
 t2   | regress_test_role         | foo
(8 rows)
CREATE USER MAPPING FOR current_user SERVER s3;`,
			},
			{
				Statement: `\deu
       List of user mappings
 Server |         User name         
--------+---------------------------
 s3     | regress_foreign_data_user
(1 row)
DROP SERVER s3;                                             -- ERROR`,
				ErrorString: `cannot drop server s3 because other objects depend on it`,
			},
			{
				Statement: `DROP SERVER s3 CASCADE;`,
			},
			{
				Statement: `\des
                 List of foreign servers
 Name |           Owner           | Foreign-data wrapper 
------+---------------------------+----------------------
 s4   | regress_foreign_data_user | foo
 s5   | regress_foreign_data_user | foo
 s6   | regress_foreign_data_user | foo
 s7   | regress_foreign_data_user | foo
 s8   | regress_foreign_data_user | postgresql
 t1   | regress_test_role         | foo
 t2   | regress_test_role         | foo
(7 rows)
\deu
List of user mappings
 Server | User name 
--------+-----------
(0 rows)
CREATE USER MAPPING FOR regress_test_missing_role SERVER s1;  -- ERROR`,
				ErrorString: `role "regress_test_missing_role" does not exist`,
			},
			{
				Statement:   `CREATE USER MAPPING FOR current_user SERVER s1;             -- ERROR`,
				ErrorString: `server "s1" does not exist`,
			},
			{
				Statement: `CREATE USER MAPPING FOR current_user SERVER s4;`,
			},
			{
				Statement:   `CREATE USER MAPPING FOR user SERVER s4;                     -- ERROR duplicate`,
				ErrorString: `user mapping for "regress_foreign_data_user" already exists for server "s4"`,
			},
			{
				Statement: `CREATE USER MAPPING FOR public SERVER s4 OPTIONS ("this mapping" 'is public');`,
			},
			{
				Statement:   `CREATE USER MAPPING FOR user SERVER s8 OPTIONS (username 'test', password 'secret');    -- ERROR`,
				ErrorString: `invalid option "username"`,
			},
			{
				Statement: `CREATE USER MAPPING FOR user SERVER s8 OPTIONS (user 'test', password 'secret');`,
			},
			{
				Statement: `ALTER SERVER s5 OWNER TO regress_test_role;`,
			},
			{
				Statement: `ALTER SERVER s6 OWNER TO regress_test_indirect;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `CREATE USER MAPPING FOR current_user SERVER s5;`,
			},
			{
				Statement: `CREATE USER MAPPING FOR current_user SERVER s6 OPTIONS (username 'test');`,
			},
			{
				Statement:   `CREATE USER MAPPING FOR current_user SERVER s7;             -- ERROR`,
				ErrorString: `permission denied for foreign server s7`,
			},
			{
				Statement:   `CREATE USER MAPPING FOR public SERVER s8;                   -- ERROR`,
				ErrorString: `must be owner of foreign server s8`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `ALTER SERVER t1 OWNER TO regress_test_indirect;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `CREATE USER MAPPING FOR current_user SERVER t1 OPTIONS (username 'bob', password 'boo');`,
			},
			{
				Statement: `CREATE USER MAPPING FOR public SERVER t1;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `\deu
       List of user mappings
 Server |         User name         
--------+---------------------------
 s4     | public
 s4     | regress_foreign_data_user
 s5     | regress_test_role
 s6     | regress_test_role
 s8     | regress_foreign_data_user
 t1     | public
 t1     | regress_test_role
(7 rows)
ALTER USER MAPPING FOR regress_test_missing_role SERVER s4 OPTIONS (gotcha 'true'); -- ERROR`,
				ErrorString: `role "regress_test_missing_role" does not exist`,
			},
			{
				Statement:   `ALTER USER MAPPING FOR user SERVER ss4 OPTIONS (gotcha 'true'); -- ERROR`,
				ErrorString: `server "ss4" does not exist`,
			},
			{
				Statement:   `ALTER USER MAPPING FOR public SERVER s5 OPTIONS (gotcha 'true');            -- ERROR`,
				ErrorString: `user mapping for "public" does not exist for server "s5"`,
			},
			{
				Statement:   `ALTER USER MAPPING FOR current_user SERVER s8 OPTIONS (username 'test');    -- ERROR`,
				ErrorString: `invalid option "username"`,
			},
			{
				Statement: `ALTER USER MAPPING FOR current_user SERVER s8 OPTIONS (DROP user, SET password 'public');`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `ALTER USER MAPPING FOR current_user SERVER s5 OPTIONS (ADD modified '1');`,
			},
			{
				Statement:   `ALTER USER MAPPING FOR public SERVER s4 OPTIONS (ADD modified '1'); -- ERROR`,
				ErrorString: `must be owner of foreign server s4`,
			},
			{
				Statement: `ALTER USER MAPPING FOR public SERVER t1 OPTIONS (ADD modified '1');`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `\deu+
                         List of user mappings
 Server |         User name         |           FDW options            
--------+---------------------------+----------------------------------
 s4     | public                    | ("this mapping" 'is public')
 s4     | regress_foreign_data_user | 
 s5     | regress_test_role         | (modified '1')
 s6     | regress_test_role         | (username 'test')
 s8     | regress_foreign_data_user | (password 'public')
 t1     | public                    | (modified '1')
 t1     | regress_test_role         | (username 'bob', password 'boo')
(7 rows)
DROP USER MAPPING FOR regress_test_missing_role SERVER s4;  -- ERROR`,
				ErrorString: `role "regress_test_missing_role" does not exist`,
			},
			{
				Statement:   `DROP USER MAPPING FOR user SERVER ss4;`,
				ErrorString: `server "ss4" does not exist`,
			},
			{
				Statement:   `DROP USER MAPPING FOR public SERVER s7;                     -- ERROR`,
				ErrorString: `user mapping for "public" does not exist for server "s7"`,
			},
			{
				Statement: `DROP USER MAPPING IF EXISTS FOR regress_test_missing_role SERVER s4;`,
			},
			{
				Statement: `DROP USER MAPPING IF EXISTS FOR user SERVER ss4;`,
			},
			{
				Statement: `DROP USER MAPPING IF EXISTS FOR public SERVER s7;`,
			},
			{
				Statement: `CREATE USER MAPPING FOR public SERVER s8;`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement:   `DROP USER MAPPING FOR public SERVER s8;                     -- ERROR`,
				ErrorString: `must be owner of foreign server s8`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `DROP SERVER s7;`,
			},
			{
				Statement: `\deu
       List of user mappings
 Server |         User name         
--------+---------------------------
 s4     | public
 s4     | regress_foreign_data_user
 s5     | regress_test_role
 s6     | regress_test_role
 s8     | public
 s8     | regress_foreign_data_user
 t1     | public
 t1     | regress_test_role
(8 rows)
CREATE SCHEMA foreign_schema;`,
			},
			{
				Statement: `CREATE SERVER s0 FOREIGN DATA WRAPPER dummy;`,
			},
			{
				Statement:   `CREATE FOREIGN TABLE ft1 ();                                    -- ERROR`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `CREATE FOREIGN TABLE ft1 () SERVER no_server;                   -- ERROR`,
				ErrorString: `server "no_server" does not exist`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft1 (
	c1 integer OPTIONS ("param 1" 'val1') PRIMARY KEY,
	c2 text OPTIONS (param2 'val2', param3 'val3'),
	c3 date
) SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value'); -- ERROR`,
				ErrorString: `primary key constraints are not supported on foreign tables`,
			},
			{
				Statement: `CREATE TABLE ref_table (id integer PRIMARY KEY);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft1 (
	c1 integer OPTIONS ("param 1" 'val1') REFERENCES ref_table (id),
	c2 text OPTIONS (param2 'val2', param3 'val3'),
	c3 date
) SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value'); -- ERROR`,
				ErrorString: `foreign key constraints are not supported on foreign tables`,
			},
			{
				Statement: `DROP TABLE ref_table;`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft1 (
	c1 integer OPTIONS ("param 1" 'val1') NOT NULL,
	c2 text OPTIONS (param2 'val2', param3 'val3'),
	c3 date,
	UNIQUE (c3)
) SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value'); -- ERROR`,
				ErrorString: `unique constraints are not supported on foreign tables`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft1 (
	c1 integer OPTIONS ("param 1" 'val1') NOT NULL,
	c2 text OPTIONS (param2 'val2', param3 'val3') CHECK (c2 <> ''),
	c3 date,
	CHECK (c3 BETWEEN '1994-01-01'::date AND '1994-01-31'::date)
) SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value');`,
			},
			{
				Statement: `COMMENT ON FOREIGN TABLE ft1 IS 'ft1';`,
			},
			{
				Statement: `COMMENT ON COLUMN ft1.c1 IS 'ft1.c1';`,
			},
			{
				Statement: `\d+ ft1
                                                 Foreign table "public.ft1"
 Column |  Type   | Collation | Nullable | Default |          FDW options           | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+--------------------------------+----------+--------------+-------------
 c1     | integer |           | not null |         | ("param 1" 'val1')             | plain    |              | ft1.c1
 c2     | text    |           |          |         | (param2 'val2', param3 'val3') | extended |              | 
 c3     | date    |           |          |         |                                | plain    |              | 
Check constraints:
    "ft1_c2_check" CHECK (c2 <> ''::text)
    "ft1_c3_check" CHECK (c3 >= '01-01-1994'::date AND c3 <= '01-31-1994'::date)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
\det+
                                 List of foreign tables
 Schema | Table | Server |                   FDW options                   | Description 
--------+-------+--------+-------------------------------------------------+-------------
 public | ft1   | s0     | (delimiter ',', quote '"', "be quoted" 'value') | ft1
(1 row)
CREATE INDEX id_ft1_c2 ON ft1 (c2);                             -- ERROR`,
				ErrorString: `cannot create index on relation "ft1"`,
			},
			{
				Statement:   `SELECT * FROM ft1;                                              -- ERROR`,
				ErrorString: `foreign-data wrapper "dummy" has no handler`,
			},
			{
				Statement:   `EXPLAIN SELECT * FROM ft1;                                      -- ERROR`,
				ErrorString: `foreign-data wrapper "dummy" has no handler`,
			},
			{
				Statement: `CREATE TABLE lt1 (a INT) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part1
  PARTITION OF lt1 FOR VALUES FROM (0) TO (1000) SERVER s0;`,
			},
			{
				Statement: `CREATE INDEX ON lt1 (a);                              -- skips partition`,
			},
			{
				Statement:   `CREATE UNIQUE INDEX ON lt1 (a);                                 -- ERROR`,
				ErrorString: `cannot create unique index on partitioned table "lt1"`,
			},
			{
				Statement:   `ALTER TABLE lt1 ADD PRIMARY KEY (a);                            -- ERROR`,
				ErrorString: `cannot create unique index on partitioned table "lt1"`,
			},
			{
				Statement: `DROP TABLE lt1;`,
			},
			{
				Statement: `CREATE TABLE lt1 (a INT) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE INDEX ON lt1 (a);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part1
  PARTITION OF lt1 FOR VALUES FROM (0) TO (1000) SERVER s0;`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part2 (a INT) SERVER s0;`,
			},
			{
				Statement: `ALTER TABLE lt1 ATTACH PARTITION ft_part2 FOR VALUES FROM (1000) TO (2000);`,
			},
			{
				Statement: `DROP FOREIGN TABLE ft_part1, ft_part2;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX ON lt1 (a);`,
			},
			{
				Statement: `ALTER TABLE lt1 ADD PRIMARY KEY (a);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part1
  PARTITION OF lt1 FOR VALUES FROM (0) TO (1000) SERVER s0;     -- ERROR`,
				ErrorString: `cannot create foreign partition of partitioned table "lt1"`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part2 (a INT NOT NULL) SERVER s0;`,
			},
			{
				Statement: `ALTER TABLE lt1 ATTACH PARTITION ft_part2
  FOR VALUES FROM (1000) TO (2000);                             -- ERROR`,
				ErrorString: `cannot attach foreign table "ft_part2" as partition of partitioned table "lt1"`,
			},
			{
				Statement: `DROP TABLE lt1;`,
			},
			{
				Statement: `DROP FOREIGN TABLE ft_part2;`,
			},
			{
				Statement: `CREATE TABLE lt1 (a INT) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE INDEX ON lt1 (a);`,
			},
			{
				Statement: `CREATE TABLE lt1_part1
  PARTITION OF lt1 FOR VALUES FROM (0) TO (1000)
  PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part_1_1
  PARTITION OF lt1_part1 FOR VALUES FROM (0) TO (100) SERVER s0;`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part_1_2 (a INT) SERVER s0;`,
			},
			{
				Statement: `ALTER TABLE lt1_part1 ATTACH PARTITION ft_part_1_2 FOR VALUES FROM (100) TO (200);`,
			},
			{
				Statement:   `CREATE UNIQUE INDEX ON lt1 (a);`,
				ErrorString: `cannot create unique index on partitioned table "lt1"`,
			},
			{
				Statement:   `ALTER TABLE lt1 ADD PRIMARY KEY (a);`,
				ErrorString: `cannot create unique index on partitioned table "lt1_part1"`,
			},
			{
				Statement: `DROP FOREIGN TABLE ft_part_1_1, ft_part_1_2;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX ON lt1 (a);`,
			},
			{
				Statement: `ALTER TABLE lt1 ADD PRIMARY KEY (a);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part_1_1
  PARTITION OF lt1_part1 FOR VALUES FROM (0) TO (100) SERVER s0;`,
				ErrorString: `cannot create foreign partition of partitioned table "lt1_part1"`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft_part_1_2 (a INT NOT NULL) SERVER s0;`,
			},
			{
				Statement:   `ALTER TABLE lt1_part1 ATTACH PARTITION ft_part_1_2 FOR VALUES FROM (100) TO (200);`,
				ErrorString: `cannot attach foreign table "ft_part_1_2" as partition of partitioned table "lt1_part1"`,
			},
			{
				Statement: `DROP TABLE lt1;`,
			},
			{
				Statement: `DROP FOREIGN TABLE ft_part_1_2;`,
			},
			{
				Statement: `COMMENT ON FOREIGN TABLE ft1 IS 'foreign table';`,
			},
			{
				Statement: `COMMENT ON FOREIGN TABLE ft1 IS NULL;`,
			},
			{
				Statement: `COMMENT ON COLUMN ft1.c1 IS 'foreign column';`,
			},
			{
				Statement: `COMMENT ON COLUMN ft1.c1 IS NULL;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD COLUMN c4 integer;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD COLUMN c5 integer DEFAULT 0;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD COLUMN c6 integer;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD COLUMN c7 integer NOT NULL;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD COLUMN c8 integer;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD COLUMN c9 integer;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD COLUMN c10 integer OPTIONS (p1 'v1');`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c4 SET DEFAULT 0;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c5 DROP DEFAULT;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c6 SET NOT NULL;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c7 DROP NOT NULL;`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft1 ALTER COLUMN c8 TYPE char(10) USING '0'; -- ERROR`,
				ErrorString: `"ft1" is not a table`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c8 TYPE char(10);`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c8 SET DATA TYPE text;`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft1 ALTER COLUMN xmin OPTIONS (ADD p1 'v1'); -- ERROR`,
				ErrorString: `cannot alter system column "xmin"`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c7 OPTIONS (ADD p1 'v1', ADD p2 'v2'),
                        ALTER COLUMN c8 OPTIONS (ADD p1 'v1', ADD p2 'v2');`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c8 OPTIONS (SET p2 'V2', DROP p1);`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c1 SET STATISTICS 10000;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c1 SET (n_distinct = 100);`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c8 SET STATISTICS -1;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c8 SET STORAGE PLAIN;`,
			},
			{
				Statement: `\d+ ft1
                                                 Foreign table "public.ft1"
 Column |  Type   | Collation | Nullable | Default |          FDW options           | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+--------------------------------+----------+--------------+-------------
 c1     | integer |           | not null |         | ("param 1" 'val1')             | plain    | 10000        | 
 c2     | text    |           |          |         | (param2 'val2', param3 'val3') | extended |              | 
 c3     | date    |           |          |         |                                | plain    |              | 
 c4     | integer |           |          | 0       |                                | plain    |              | 
 c5     | integer |           |          |         |                                | plain    |              | 
 c6     | integer |           | not null |         |                                | plain    |              | 
 c7     | integer |           |          |         | (p1 'v1', p2 'v2')             | plain    |              | 
 c8     | text    |           |          |         | (p2 'V2')                      | plain    |              | 
 c9     | integer |           |          |         |                                | plain    |              | 
 c10    | integer |           |          |         | (p1 'v1')                      | plain    |              | 
Check constraints:
    "ft1_c2_check" CHECK (c2 <> ''::text)
    "ft1_c3_check" CHECK (c3 >= '01-01-1994'::date AND c3 <= '01-31-1994'::date)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
CREATE TABLE use_ft1_column_type (x ft1);`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft1 ALTER COLUMN c8 SET DATA TYPE integer;	-- ERROR`,
				ErrorString: `cannot alter foreign table "ft1" because column "use_ft1_column_type.x" uses its row type`,
			},
			{
				Statement: `DROP TABLE use_ft1_column_type;`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft1 ADD PRIMARY KEY (c7);                   -- ERROR`,
				ErrorString: `primary key constraints are not supported on foreign tables`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD CONSTRAINT ft1_c9_check CHECK (c9 < 0) NOT VALID;`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft1 ALTER CONSTRAINT ft1_c9_check DEFERRABLE; -- ERROR`,
				ErrorString: `ALTER action ALTER CONSTRAINT cannot be performed on relation "ft1"`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 DROP CONSTRAINT ft1_c9_check;`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft1 DROP CONSTRAINT no_const;               -- ERROR`,
				ErrorString: `constraint "no_const" of relation "ft1" does not exist`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 DROP CONSTRAINT IF EXISTS no_const;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 OWNER TO regress_test_role;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 OPTIONS (DROP delimiter, SET quote '~', ADD escape '@');`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft1 DROP COLUMN no_column;                  -- ERROR`,
				ErrorString: `column "no_column" of relation "ft1" does not exist`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 DROP COLUMN IF EXISTS no_column;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 DROP COLUMN c9;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 SET SCHEMA foreign_schema;`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft1 SET TABLESPACE ts;                      -- ERROR`,
				ErrorString: `relation "ft1" does not exist`,
			},
			{
				Statement: `ALTER FOREIGN TABLE foreign_schema.ft1 RENAME c1 TO foreign_column_1;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE foreign_schema.ft1 RENAME TO foreign_table_1;`,
			},
			{
				Statement: `\d foreign_schema.foreign_table_1
                        Foreign table "foreign_schema.foreign_table_1"
      Column      |  Type   | Collation | Nullable | Default |          FDW options           
------------------+---------+-----------+----------+---------+--------------------------------
 foreign_column_1 | integer |           | not null |         | ("param 1" 'val1')
 c2               | text    |           |          |         | (param2 'val2', param3 'val3')
 c3               | date    |           |          |         | 
 c4               | integer |           |          | 0       | 
 c5               | integer |           |          |         | 
 c6               | integer |           | not null |         | 
 c7               | integer |           |          |         | (p1 'v1', p2 'v2')
 c8               | text    |           |          |         | (p2 'V2')
 c10              | integer |           |          |         | (p1 'v1')
Check constraints:
    "ft1_c2_check" CHECK (c2 <> ''::text)
    "ft1_c3_check" CHECK (c3 >= '01-01-1994'::date AND c3 <= '01-31-1994'::date)
Server: s0
FDW options: (quote '~', "be quoted" 'value', escape '@')
ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ADD COLUMN c4 integer;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ADD COLUMN c6 integer;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ADD COLUMN c7 integer NOT NULL;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ADD COLUMN c8 integer;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ADD COLUMN c9 integer;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ADD COLUMN c10 integer OPTIONS (p1 'v1');`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ALTER COLUMN c6 SET NOT NULL;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ALTER COLUMN c7 DROP NOT NULL;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ALTER COLUMN c8 TYPE char(10);`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ALTER COLUMN c8 SET DATA TYPE text;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ALTER COLUMN c7 OPTIONS (ADD p1 'v1', ADD p2 'v2'),
                        ALTER COLUMN c8 OPTIONS (ADD p1 'v1', ADD p2 'v2');`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 ALTER COLUMN c8 OPTIONS (SET p2 'V2', DROP p1);`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 DROP CONSTRAINT IF EXISTS no_const;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 DROP CONSTRAINT ft1_c1_check;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 OWNER TO regress_test_role;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 OPTIONS (DROP delimiter, SET quote '~', ADD escape '@');`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 DROP COLUMN IF EXISTS no_column;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 DROP COLUMN c9;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 SET SCHEMA foreign_schema;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 RENAME c1 TO foreign_column_1;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE IF EXISTS doesnt_exist_ft1 RENAME TO foreign_table_1;`,
			},
			{
				Statement: `SELECT * FROM information_schema.foreign_data_wrappers ORDER BY 1, 2;`,
				Results:   []sql.Row{{`regression`, `dummy`, `regress_foreign_data_user`, ``, `c`}, {`regression`, `foo`, `regress_foreign_data_user`, ``, `c`}, {`regression`, `postgresql`, `regress_foreign_data_user`, ``, `c`}},
			},
			{
				Statement: `SELECT * FROM information_schema.foreign_data_wrapper_options ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{`regression`, `foo`, `test wrapper`, `true`}},
			},
			{
				Statement: `SELECT * FROM information_schema.foreign_servers ORDER BY 1, 2;`,
				Results:   []sql.Row{{`regression`, `s0`, `regression`, `dummy`, ``, ``, `regress_foreign_data_user`}, {`regression`, `s4`, `regression`, `foo`, `oracle`, ``, `regress_foreign_data_user`}, {`regression`, `s5`, `regression`, `foo`, ``, 15.0, `regress_test_role`}, {`regression`, `s6`, `regression`, `foo`, ``, 16.0, `regress_test_indirect`}, {`regression`, `s8`, `regression`, `postgresql`, ``, ``, `regress_foreign_data_user`}, {`regression`, `t1`, `regression`, `foo`, ``, ``, `regress_test_indirect`}, {`regression`, `t2`, `regression`, `foo`, ``, ``, `regress_test_role`}},
			},
			{
				Statement: `SELECT * FROM information_schema.foreign_server_options ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{`regression`, `s4`, `dbname`, `b`}, {`regression`, `s4`, `host`, `a`}, {`regression`, `s6`, `dbname`, `b`}, {`regression`, `s6`, `host`, `a`}, {`regression`, `s8`, `connect_timeout`, 30}, {`regression`, `s8`, `dbname`, `db1`}},
			},
			{
				Statement: `SELECT * FROM information_schema.user_mappings ORDER BY lower(authorization_identifier), 2, 3;`,
				Results:   []sql.Row{{`PUBLIC`, `regression`, `s4`}, {`PUBLIC`, `regression`, `s8`}, {`PUBLIC`, `regression`, `t1`}, {`regress_foreign_data_user`, `regression`, `s4`}, {`regress_foreign_data_user`, `regression`, `s8`}, {`regress_test_role`, `regression`, `s5`}, {`regress_test_role`, `regression`, `s6`}, {`regress_test_role`, `regression`, `t1`}},
			},
			{
				Statement: `SELECT * FROM information_schema.user_mapping_options ORDER BY lower(authorization_identifier), 2, 3, 4;`,
				Results:   []sql.Row{{`PUBLIC`, `regression`, `s4`, `this mapping`, `is public`}, {`PUBLIC`, `regression`, `t1`, `modified`, 1}, {`regress_foreign_data_user`, `regression`, `s8`, `password`, `public`}, {`regress_test_role`, `regression`, `s5`, `modified`, 1}, {`regress_test_role`, `regression`, `s6`, `username`, `test`}, {`regress_test_role`, `regression`, `t1`, `password`, `boo`}, {`regress_test_role`, `regression`, `t1`, `username`, `bob`}},
			},
			{
				Statement: `SELECT * FROM information_schema.usage_privileges WHERE object_type LIKE 'FOREIGN%' AND object_name IN ('s6', 'foo') ORDER BY 1, 2, 3, 4, 5;`,
				Results:   []sql.Row{{`regress_foreign_data_user`, `regress_foreign_data_user`, `regression`, ``, `foo`, `FOREIGN DATA WRAPPER`, `USAGE`, `YES`}, {`regress_foreign_data_user`, `regress_test_indirect`, `regression`, ``, `foo`, `FOREIGN DATA WRAPPER`, `USAGE`, `NO`}, {`regress_test_indirect`, `regress_test_indirect`, `regression`, ``, `s6`, `FOREIGN SERVER`, `USAGE`, `YES`}, {`regress_test_indirect`, `regress_test_role2`, `regression`, ``, `s6`, `FOREIGN SERVER`, `USAGE`, `YES`}},
			},
			{
				Statement: `SELECT * FROM information_schema.role_usage_grants WHERE object_type LIKE 'FOREIGN%' AND object_name IN ('s6', 'foo') ORDER BY 1, 2, 3, 4, 5;`,
				Results:   []sql.Row{{`regress_foreign_data_user`, `regress_foreign_data_user`, `regression`, ``, `foo`, `FOREIGN DATA WRAPPER`, `USAGE`, `YES`}, {`regress_foreign_data_user`, `regress_test_indirect`, `regression`, ``, `foo`, `FOREIGN DATA WRAPPER`, `USAGE`, `NO`}, {`regress_test_indirect`, `regress_test_indirect`, `regression`, ``, `s6`, `FOREIGN SERVER`, `USAGE`, `YES`}, {`regress_test_indirect`, `regress_test_role2`, `regression`, ``, `s6`, `FOREIGN SERVER`, `USAGE`, `YES`}},
			},
			{
				Statement: `SELECT * FROM information_schema.foreign_tables ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{`regression`, `foreign_schema`, `foreign_table_1`, `regression`, `s0`}},
			},
			{
				Statement: `SELECT * FROM information_schema.foreign_table_options ORDER BY 1, 2, 3, 4;`,
				Results:   []sql.Row{{`regression`, `foreign_schema`, `foreign_table_1`, `be quoted`, `value`}, {`regression`, `foreign_schema`, `foreign_table_1`, `escape`, `@`}, {`regression`, `foreign_schema`, `foreign_table_1`, `quote`, `~`}},
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `SELECT * FROM information_schema.user_mapping_options ORDER BY 1, 2, 3, 4;`,
				Results:   []sql.Row{{`PUBLIC`, `regression`, `t1`, `modified`, 1}, {`regress_test_role`, `regression`, `s5`, `modified`, 1}, {`regress_test_role`, `regression`, `s6`, `username`, `test`}, {`regress_test_role`, `regression`, `t1`, `password`, `boo`}, {`regress_test_role`, `regression`, `t1`, `username`, `bob`}},
			},
			{
				Statement: `SELECT * FROM information_schema.usage_privileges WHERE object_type LIKE 'FOREIGN%' AND object_name IN ('s6', 'foo') ORDER BY 1, 2, 3, 4, 5;`,
				Results:   []sql.Row{{`regress_foreign_data_user`, `regress_test_indirect`, `regression`, ``, `foo`, `FOREIGN DATA WRAPPER`, `USAGE`, `NO`}, {`regress_test_indirect`, `regress_test_indirect`, `regression`, ``, `s6`, `FOREIGN SERVER`, `USAGE`, `YES`}, {`regress_test_indirect`, `regress_test_role2`, `regression`, ``, `s6`, `FOREIGN SERVER`, `USAGE`, `YES`}},
			},
			{
				Statement: `SELECT * FROM information_schema.role_usage_grants WHERE object_type LIKE 'FOREIGN%' AND object_name IN ('s6', 'foo') ORDER BY 1, 2, 3, 4, 5;`,
				Results:   []sql.Row{{`regress_foreign_data_user`, `regress_test_indirect`, `regression`, ``, `foo`, `FOREIGN DATA WRAPPER`, `USAGE`, `NO`}, {`regress_test_indirect`, `regress_test_indirect`, `regression`, ``, `s6`, `FOREIGN SERVER`, `USAGE`, `YES`}, {`regress_test_indirect`, `regress_test_role2`, `regression`, ``, `s6`, `FOREIGN SERVER`, `USAGE`, `YES`}},
			},
			{
				Statement: `DROP USER MAPPING FOR current_user SERVER t1;`,
			},
			{
				Statement: `SET ROLE regress_test_role2;`,
			},
			{
				Statement: `SELECT * FROM information_schema.user_mapping_options ORDER BY 1, 2, 3, 4;`,
				Results:   []sql.Row{{`regress_test_role`, `regression`, `s6`, `username`, ``}},
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `SELECT has_foreign_data_wrapper_privilege('regress_test_role',
    (SELECT oid FROM pg_foreign_data_wrapper WHERE fdwname='foo'), 'USAGE');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_foreign_data_wrapper_privilege('regress_test_role', 'foo', 'USAGE');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_foreign_data_wrapper_privilege(
    (SELECT oid FROM pg_roles WHERE rolname='regress_test_role'),
    (SELECT oid FROM pg_foreign_data_wrapper WHERE fdwname='foo'), 'USAGE');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_foreign_data_wrapper_privilege(
    (SELECT oid FROM pg_foreign_data_wrapper WHERE fdwname='foo'), 'USAGE');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_foreign_data_wrapper_privilege(
    (SELECT oid FROM pg_roles WHERE rolname='regress_test_role'), 'foo', 'USAGE');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_foreign_data_wrapper_privilege('foo', 'USAGE');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `GRANT USAGE ON FOREIGN DATA WRAPPER foo TO regress_test_role;`,
			},
			{
				Statement: `SELECT has_foreign_data_wrapper_privilege('regress_test_role', 'foo', 'USAGE');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_server_privilege('regress_test_role',
    (SELECT oid FROM pg_foreign_server WHERE srvname='s8'), 'USAGE');`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_server_privilege('regress_test_role', 's8', 'USAGE');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_server_privilege(
    (SELECT oid FROM pg_roles WHERE rolname='regress_test_role'),
    (SELECT oid FROM pg_foreign_server WHERE srvname='s8'), 'USAGE');`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_server_privilege(
    (SELECT oid FROM pg_foreign_server WHERE srvname='s8'), 'USAGE');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_server_privilege(
    (SELECT oid FROM pg_roles WHERE rolname='regress_test_role'), 's8', 'USAGE');`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_server_privilege('s8', 'USAGE');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `GRANT USAGE ON FOREIGN SERVER s8 TO regress_test_role;`,
			},
			{
				Statement: `SELECT has_server_privilege('regress_test_role', 's8', 'USAGE');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `REVOKE USAGE ON FOREIGN SERVER s8 FROM regress_test_role;`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN SERVER s4 TO regress_test_role;`,
			},
			{
				Statement: `DROP USER MAPPING FOR public SERVER s4;`,
			},
			{
				Statement: `ALTER SERVER s6 OPTIONS (DROP host, DROP dbname);`,
			},
			{
				Statement: `ALTER USER MAPPING FOR regress_test_role SERVER s6 OPTIONS (DROP username);`,
			},
			{
				Statement: `ALTER FOREIGN DATA WRAPPER foo VALIDATOR postgresql_fdw_validator;`,
			},
			{
				Statement: `SET ROLE regress_unprivileged_role;`,
			},
			{
				Statement:   `CREATE FOREIGN DATA WRAPPER foobar;                             -- ERROR`,
				ErrorString: `permission denied to create foreign-data wrapper "foobar"`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OPTIONS (gotcha 'true');         -- ERROR`,
				ErrorString: `permission denied to alter foreign-data wrapper "foo"`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OWNER TO regress_unprivileged_role; -- ERROR`,
				ErrorString: `permission denied to change owner of foreign-data wrapper "foo"`,
			},
			{
				Statement:   `DROP FOREIGN DATA WRAPPER foo;                                  -- ERROR`,
				ErrorString: `must be owner of foreign-data wrapper foo`,
			},
			{
				Statement:   `GRANT USAGE ON FOREIGN DATA WRAPPER foo TO regress_test_role;   -- ERROR`,
				ErrorString: `permission denied for foreign-data wrapper foo`,
			},
			{
				Statement:   `CREATE SERVER s9 FOREIGN DATA WRAPPER foo;                      -- ERROR`,
				ErrorString: `permission denied for foreign-data wrapper foo`,
			},
			{
				Statement:   `ALTER SERVER s4 VERSION '0.5';                                  -- ERROR`,
				ErrorString: `must be owner of foreign server s4`,
			},
			{
				Statement:   `ALTER SERVER s4 OWNER TO regress_unprivileged_role;             -- ERROR`,
				ErrorString: `must be owner of foreign server s4`,
			},
			{
				Statement:   `DROP SERVER s4;                                                 -- ERROR`,
				ErrorString: `must be owner of foreign server s4`,
			},
			{
				Statement:   `GRANT USAGE ON FOREIGN SERVER s4 TO regress_test_role;          -- ERROR`,
				ErrorString: `permission denied for foreign server s4`,
			},
			{
				Statement:   `CREATE USER MAPPING FOR public SERVER s4;                       -- ERROR`,
				ErrorString: `must be owner of foreign server s4`,
			},
			{
				Statement:   `ALTER USER MAPPING FOR regress_test_role SERVER s6 OPTIONS (gotcha 'true'); -- ERROR`,
				ErrorString: `must be owner of foreign server s6`,
			},
			{
				Statement:   `DROP USER MAPPING FOR regress_test_role SERVER s6;              -- ERROR`,
				ErrorString: `must be owner of foreign server s6`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN DATA WRAPPER postgresql TO regress_unprivileged_role;`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN DATA WRAPPER foo TO regress_unprivileged_role WITH GRANT OPTION;`,
			},
			{
				Statement: `SET ROLE regress_unprivileged_role;`,
			},
			{
				Statement:   `CREATE FOREIGN DATA WRAPPER foobar;                             -- ERROR`,
				ErrorString: `permission denied to create foreign-data wrapper "foobar"`,
			},
			{
				Statement:   `ALTER FOREIGN DATA WRAPPER foo OPTIONS (gotcha 'true');         -- ERROR`,
				ErrorString: `permission denied to alter foreign-data wrapper "foo"`,
			},
			{
				Statement:   `DROP FOREIGN DATA WRAPPER foo;                                  -- ERROR`,
				ErrorString: `must be owner of foreign-data wrapper foo`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN DATA WRAPPER postgresql TO regress_test_role; -- WARNING`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN DATA WRAPPER foo TO regress_test_role;`,
			},
			{
				Statement: `CREATE SERVER s9 FOREIGN DATA WRAPPER postgresql;`,
			},
			{
				Statement:   `ALTER SERVER s6 VERSION '0.5';                                  -- ERROR`,
				ErrorString: `must be owner of foreign server s6`,
			},
			{
				Statement:   `DROP SERVER s6;                                                 -- ERROR`,
				ErrorString: `must be owner of foreign server s6`,
			},
			{
				Statement:   `GRANT USAGE ON FOREIGN SERVER s6 TO regress_test_role;          -- ERROR`,
				ErrorString: `permission denied for foreign server s6`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN SERVER s9 TO regress_test_role;`,
			},
			{
				Statement:   `CREATE USER MAPPING FOR public SERVER s6;                       -- ERROR`,
				ErrorString: `must be owner of foreign server s6`,
			},
			{
				Statement: `CREATE USER MAPPING FOR public SERVER s9;`,
			},
			{
				Statement:   `ALTER USER MAPPING FOR regress_test_role SERVER s6 OPTIONS (gotcha 'true'); -- ERROR`,
				ErrorString: `must be owner of foreign server s6`,
			},
			{
				Statement:   `DROP USER MAPPING FOR regress_test_role SERVER s6;              -- ERROR`,
				ErrorString: `must be owner of foreign server s6`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement:   `REVOKE USAGE ON FOREIGN DATA WRAPPER foo FROM regress_unprivileged_role; -- ERROR`,
				ErrorString: `dependent privileges exist`,
			},
			{
				Statement: `REVOKE USAGE ON FOREIGN DATA WRAPPER foo FROM regress_unprivileged_role CASCADE;`,
			},
			{
				Statement: `SET ROLE regress_unprivileged_role;`,
			},
			{
				Statement:   `GRANT USAGE ON FOREIGN DATA WRAPPER foo TO regress_test_role;   -- ERROR`,
				ErrorString: `permission denied for foreign-data wrapper foo`,
			},
			{
				Statement:   `CREATE SERVER s10 FOREIGN DATA WRAPPER foo;                     -- ERROR`,
				ErrorString: `permission denied for foreign-data wrapper foo`,
			},
			{
				Statement: `ALTER SERVER s9 VERSION '1.1';`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN SERVER s9 TO regress_test_role;`,
			},
			{
				Statement: `CREATE USER MAPPING FOR current_user SERVER s9;`,
			},
			{
				Statement: `DROP SERVER s9 CASCADE;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `CREATE SERVER s9 FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN SERVER s9 TO regress_unprivileged_role;`,
			},
			{
				Statement: `SET ROLE regress_unprivileged_role;`,
			},
			{
				Statement:   `ALTER SERVER s9 VERSION '1.2';                                  -- ERROR`,
				ErrorString: `must be owner of foreign server s9`,
			},
			{
				Statement: `GRANT USAGE ON FOREIGN SERVER s9 TO regress_test_role;          -- WARNING`,
			},
			{
				Statement: `CREATE USER MAPPING FOR current_user SERVER s9;`,
			},
			{
				Statement:   `DROP SERVER s9 CASCADE;                                         -- ERROR`,
				ErrorString: `must be owner of foreign server s9`,
			},
			{
				Statement: `SET ROLE regress_test_role;`,
			},
			{
				Statement: `CREATE SERVER s10 FOREIGN DATA WRAPPER foo;`,
			},
			{
				Statement: `CREATE USER MAPPING FOR public SERVER s10 OPTIONS (user 'secret');`,
			},
			{
				Statement: `CREATE USER MAPPING FOR regress_unprivileged_role SERVER s10 OPTIONS (user 'secret');`,
			},
			{
				Statement: `\deu+
                 List of user mappings
 Server |         User name         |    FDW options    
--------+---------------------------+-------------------
 s10    | public                    | ("user" 'secret')
 s10    | regress_unprivileged_role | 
 s4     | regress_foreign_data_user | 
 s5     | regress_test_role         | (modified '1')
 s6     | regress_test_role         | 
 s8     | public                    | 
 s8     | regress_foreign_data_user | 
 s9     | regress_unprivileged_role | 
 t1     | public                    | (modified '1')
(9 rows)
RESET ROLE;`,
			},
			{
				Statement: `\deu+
                  List of user mappings
 Server |         User name         |     FDW options     
--------+---------------------------+---------------------
 s10    | public                    | ("user" 'secret')
 s10    | regress_unprivileged_role | ("user" 'secret')
 s4     | regress_foreign_data_user | 
 s5     | regress_test_role         | (modified '1')
 s6     | regress_test_role         | 
 s8     | public                    | 
 s8     | regress_foreign_data_user | (password 'public')
 s9     | regress_unprivileged_role | 
 t1     | public                    | (modified '1')
(9 rows)
SET ROLE regress_unprivileged_role;`,
			},
			{
				Statement: `\deu+
              List of user mappings
 Server |         User name         | FDW options 
--------+---------------------------+-------------
 s10    | public                    | 
 s10    | regress_unprivileged_role | 
 s4     | regress_foreign_data_user | 
 s5     | regress_test_role         | 
 s6     | regress_test_role         | 
 s8     | public                    | 
 s8     | regress_foreign_data_user | 
 s9     | regress_unprivileged_role | 
 t1     | public                    | 
(9 rows)
RESET ROLE;`,
			},
			{
				Statement: `DROP SERVER s10 CASCADE;`,
			},
			{
				Statement: `CREATE FUNCTION dummy_trigger() RETURNS TRIGGER AS $$
  BEGIN
    RETURN NULL;`,
			},
			{
				Statement: `  END
$$ language plpgsql;`,
			},
			{
				Statement: `CREATE TRIGGER trigtest_before_stmt BEFORE INSERT OR UPDATE OR DELETE
ON foreign_schema.foreign_table_1
FOR EACH STATEMENT
EXECUTE PROCEDURE dummy_trigger();`,
			},
			{
				Statement: `CREATE TRIGGER trigtest_after_stmt AFTER INSERT OR UPDATE OR DELETE
ON foreign_schema.foreign_table_1
FOR EACH STATEMENT
EXECUTE PROCEDURE dummy_trigger();`,
			},
			{
				Statement: `CREATE TRIGGER trigtest_after_stmt_tt AFTER INSERT OR UPDATE OR DELETE -- ERROR
ON foreign_schema.foreign_table_1
REFERENCING NEW TABLE AS new_table
FOR EACH STATEMENT
EXECUTE PROCEDURE dummy_trigger();`,
				ErrorString: `"foreign_table_1" is a foreign table`,
			},
			{
				Statement: `CREATE TRIGGER trigtest_before_row BEFORE INSERT OR UPDATE OR DELETE
ON foreign_schema.foreign_table_1
FOR EACH ROW
EXECUTE PROCEDURE dummy_trigger();`,
			},
			{
				Statement: `CREATE TRIGGER trigtest_after_row AFTER INSERT OR UPDATE OR DELETE
ON foreign_schema.foreign_table_1
FOR EACH ROW
EXECUTE PROCEDURE dummy_trigger();`,
			},
			{
				Statement: `CREATE CONSTRAINT TRIGGER trigtest_constraint AFTER INSERT OR UPDATE OR DELETE
ON foreign_schema.foreign_table_1
FOR EACH ROW
EXECUTE PROCEDURE dummy_trigger();`,
				ErrorString: `"foreign_table_1" is a foreign table`,
			},
			{
				Statement: `ALTER FOREIGN TABLE foreign_schema.foreign_table_1
	DISABLE TRIGGER trigtest_before_stmt;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE foreign_schema.foreign_table_1
	ENABLE TRIGGER trigtest_before_stmt;`,
			},
			{
				Statement: `DROP TRIGGER trigtest_before_stmt ON foreign_schema.foreign_table_1;`,
			},
			{
				Statement: `DROP TRIGGER trigtest_before_row ON foreign_schema.foreign_table_1;`,
			},
			{
				Statement: `DROP TRIGGER trigtest_after_stmt ON foreign_schema.foreign_table_1;`,
			},
			{
				Statement: `DROP TRIGGER trigtest_after_row ON foreign_schema.foreign_table_1;`,
			},
			{
				Statement: `DROP FUNCTION dummy_trigger();`,
			},
			{
				Statement: `CREATE TABLE fd_pt1 (
	c1 integer NOT NULL,
	c2 text,
	c3 date
);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft2 () INHERITS (fd_pt1)
  SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value');`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
DROP FOREIGN TABLE ft2;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
CREATE FOREIGN TABLE ft2 (
	c1 integer NOT NULL,
	c2 text,
	c3 date
) SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value');`,
			},
			{
				Statement: `\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
ALTER FOREIGN TABLE ft2 INHERIT fd_pt1;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
CREATE TABLE ct3() INHERITS(ft2);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft3 (
	c1 integer NOT NULL,
	c2 text,
	c3 date
) INHERITS(ft2)
  SERVER s0;`,
			},
			{
				Statement: `\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
Child tables: ct3,
              ft3
\d+ ct3
                                    Table "public.ct3"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Inherits: ft2
\d+ ft3
                                       Foreign table "public.ft3"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Server: s0
Inherits: ft2
ALTER TABLE fd_pt1 ADD COLUMN c4 integer;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ADD COLUMN c5 integer DEFAULT 0;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ADD COLUMN c6 integer;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ADD COLUMN c7 integer NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ADD COLUMN c8 integer;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
 c4     | integer |           |          |         | plain    |              | 
 c5     | integer |           |          | 0       | plain    |              | 
 c6     | integer |           |          |         | plain    |              | 
 c7     | integer |           | not null |         | plain    |              | 
 c8     | integer |           |          |         | plain    |              | 
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
 c4     | integer |           |          |         |             | plain    |              | 
 c5     | integer |           |          | 0       |             | plain    |              | 
 c6     | integer |           |          |         |             | plain    |              | 
 c7     | integer |           | not null |         |             | plain    |              | 
 c8     | integer |           |          |         |             | plain    |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
Child tables: ct3,
              ft3
\d+ ct3
                                    Table "public.ct3"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
 c4     | integer |           |          |         | plain    |              | 
 c5     | integer |           |          | 0       | plain    |              | 
 c6     | integer |           |          |         | plain    |              | 
 c7     | integer |           | not null |         | plain    |              | 
 c8     | integer |           |          |         | plain    |              | 
Inherits: ft2
\d+ ft3
                                       Foreign table "public.ft3"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
 c4     | integer |           |          |         |             | plain    |              | 
 c5     | integer |           |          | 0       |             | plain    |              | 
 c6     | integer |           |          |         |             | plain    |              | 
 c7     | integer |           | not null |         |             | plain    |              | 
 c8     | integer |           |          |         |             | plain    |              | 
Server: s0
Inherits: ft2
ALTER TABLE fd_pt1 ALTER COLUMN c4 SET DEFAULT 0;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c5 DROP DEFAULT;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c6 SET NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c7 DROP NOT NULL;`,
			},
			{
				Statement:   `ALTER TABLE fd_pt1 ALTER COLUMN c8 TYPE char(10) USING '0';        -- ERROR`,
				ErrorString: `"ft2" is not a table`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c8 TYPE char(10);`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c8 SET DATA TYPE text;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c1 SET STATISTICS 10000;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c1 SET (n_distinct = 100);`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c8 SET STATISTICS -1;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ALTER COLUMN c8 SET STORAGE EXTERNAL;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    | 10000        | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
 c4     | integer |           |          | 0       | plain    |              | 
 c5     | integer |           |          |         | plain    |              | 
 c6     | integer |           | not null |         | plain    |              | 
 c7     | integer |           |          |         | plain    |              | 
 c8     | text    |           |          |         | external |              | 
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    | 10000        | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
 c4     | integer |           |          | 0       |             | plain    |              | 
 c5     | integer |           |          |         |             | plain    |              | 
 c6     | integer |           | not null |         |             | plain    |              | 
 c7     | integer |           |          |         |             | plain    |              | 
 c8     | text    |           |          |         |             | external |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
Child tables: ct3,
              ft3
ALTER TABLE fd_pt1 DROP COLUMN c4;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 DROP COLUMN c5;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 DROP COLUMN c6;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 DROP COLUMN c7;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 DROP COLUMN c8;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    | 10000        | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    | 10000        | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
Child tables: ct3,
              ft3
ALTER TABLE fd_pt1 ADD CONSTRAINT fd_pt1chk1 CHECK (c1 > 0) NO INHERIT;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ADD CONSTRAINT fd_pt1chk2 CHECK (c2 <> '');`,
			},
			{
				Statement: `SELECT relname, conname, contype, conislocal, coninhcount, connoinherit
  FROM pg_class AS pc JOIN pg_constraint AS pgc ON (conrelid = pc.oid)
  WHERE pc.relname = 'fd_pt1'
  ORDER BY 1,2;`,
				Results: []sql.Row{{`fd_pt1`, `fd_pt1chk1`, `c`, true, 0, true}, {`fd_pt1`, `fd_pt1chk2`, `c`, true, 0, false}},
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    | 10000        | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Check constraints:
    "fd_pt1chk1" CHECK (c1 > 0) NO INHERIT
    "fd_pt1chk2" CHECK (c2 <> ''::text)
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    | 10000        | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Check constraints:
    "fd_pt1chk2" CHECK (c2 <> ''::text)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
Child tables: ct3,
              ft3
DROP FOREIGN TABLE ft2; -- ERROR`,
				ErrorString: `cannot drop foreign table ft2 because other objects depend on it`,
			},
			{
				Statement: `foreign table ft3 depends on foreign table ft2
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP FOREIGN TABLE ft2 CASCADE;`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft2 (
	c1 integer NOT NULL,
	c2 text,
	c3 date
) SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value');`,
			},
			{
				Statement:   `ALTER FOREIGN TABLE ft2 INHERIT fd_pt1;                            -- ERROR`,
				ErrorString: `child table is missing constraint "fd_pt1chk2"`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft2 ADD CONSTRAINT fd_pt1chk2 CHECK (c2 <> '');`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft2 INHERIT fd_pt1;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    | 10000        | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Check constraints:
    "fd_pt1chk1" CHECK (c1 > 0) NO INHERIT
    "fd_pt1chk2" CHECK (c2 <> ''::text)
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Check constraints:
    "fd_pt1chk2" CHECK (c2 <> ''::text)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
ALTER TABLE fd_pt1 DROP CONSTRAINT fd_pt1chk1 CASCADE;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 DROP CONSTRAINT fd_pt1chk2 CASCADE;`,
			},
			{
				Statement: `INSERT INTO fd_pt1 VALUES (1, 'fd_pt1'::text, '1994-01-01'::date);`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 ADD CONSTRAINT fd_pt1chk3 CHECK (c2 <> '') NOT VALID;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    | 10000        | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Check constraints:
    "fd_pt1chk3" CHECK (c2 <> ''::text) NOT VALID
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Check constraints:
    "fd_pt1chk2" CHECK (c2 <> ''::text)
    "fd_pt1chk3" CHECK (c2 <> ''::text) NOT VALID
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
ALTER TABLE fd_pt1 VALIDATE CONSTRAINT fd_pt1chk3;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    | 10000        | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Check constraints:
    "fd_pt1chk3" CHECK (c2 <> ''::text)
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Check constraints:
    "fd_pt1chk2" CHECK (c2 <> ''::text)
    "fd_pt1chk3" CHECK (c2 <> ''::text)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
ALTER TABLE fd_pt1 RENAME COLUMN c1 TO f1;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 RENAME COLUMN c2 TO f2;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 RENAME COLUMN c3 TO f3;`,
			},
			{
				Statement: `ALTER TABLE fd_pt1 RENAME CONSTRAINT fd_pt1chk3 TO f2_check;`,
			},
			{
				Statement: `\d+ fd_pt1
                                   Table "public.fd_pt1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 f1     | integer |           | not null |         | plain    | 10000        | 
 f2     | text    |           |          |         | extended |              | 
 f3     | date    |           |          |         | plain    |              | 
Check constraints:
    "f2_check" CHECK (f2 <> ''::text)
Child tables: ft2
\d+ ft2
                                       Foreign table "public.ft2"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 f1     | integer |           | not null |         |             | plain    |              | 
 f2     | text    |           |          |         |             | extended |              | 
 f3     | date    |           |          |         |             | plain    |              | 
Check constraints:
    "f2_check" CHECK (f2 <> ''::text)
    "fd_pt1chk2" CHECK (f2 <> ''::text)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
Inherits: fd_pt1
DROP TABLE fd_pt1 CASCADE;`,
			},
			{
				Statement:   `IMPORT FOREIGN SCHEMA s1 FROM SERVER s9 INTO public; -- ERROR`,
				ErrorString: `foreign-data wrapper "foo" has no handler`,
			},
			{
				Statement: `IMPORT FOREIGN SCHEMA s1 LIMIT TO (t1) FROM SERVER s9 INTO public; --ERROR
ERROR:  foreign-data wrapper "foo" has no handler
IMPORT FOREIGN SCHEMA s1 EXCEPT (t1) FROM SERVER s9 INTO public; -- ERROR`,
				ErrorString: `foreign-data wrapper "foo" has no handler`,
			},
			{
				Statement: `IMPORT FOREIGN SCHEMA s1 EXCEPT (t1, t2) FROM SERVER s9 INTO public
OPTIONS (option1 'value1', option2 'value2'); -- ERROR`,
				ErrorString: `foreign-data wrapper "foo" has no handler`,
			},
			{
				Statement:   `DROP FOREIGN TABLE no_table;                                    -- ERROR`,
				ErrorString: `foreign table "no_table" does not exist`,
			},
			{
				Statement: `DROP FOREIGN TABLE IF EXISTS no_table;`,
			},
			{
				Statement: `DROP FOREIGN TABLE foreign_schema.foreign_table_1;`,
			},
			{
				Statement: `REASSIGN OWNED BY regress_test_role TO regress_test_role2;`,
			},
			{
				Statement:   `DROP OWNED BY regress_test_role2;`,
				ErrorString: `cannot drop desired object(s) because other objects depend on them`,
			},
			{
				Statement: `DROP OWNED BY regress_test_role2 CASCADE;`,
			},
			{
				Statement: `CREATE TABLE fd_pt2 (
	c1 integer NOT NULL,
	c2 text,
	c3 date
) PARTITION BY LIST (c1);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE fd_pt2_1 PARTITION OF fd_pt2 FOR VALUES IN (1)
  SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value');`,
			},
			{
				Statement: `\d+ fd_pt2
                             Partitioned table "public.fd_pt2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Partition key: LIST (c1)
Partitions: fd_pt2_1 FOR VALUES IN (1)
\d+ fd_pt2_1
                                     Foreign table "public.fd_pt2_1"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Partition of: fd_pt2 FOR VALUES IN (1)
Partition constraint: ((c1 IS NOT NULL) AND (c1 = 1))
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
DROP FOREIGN TABLE fd_pt2_1;`,
			},
			{
				Statement: `CREATE FOREIGN TABLE fd_pt2_1 (
	c1 integer NOT NULL,
	c2 text,
	c3 date,
	c4 char
) SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value');`,
			},
			{
				Statement: `\d+ fd_pt2_1
                                       Foreign table "public.fd_pt2_1"
 Column |     Type     | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+--------------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer      |           | not null |         |             | plain    |              | 
 c2     | text         |           |          |         |             | extended |              | 
 c3     | date         |           |          |         |             | plain    |              | 
 c4     | character(1) |           |          |         |             | extended |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
ALTER TABLE fd_pt2 ATTACH PARTITION fd_pt2_1 FOR VALUES IN (1);       -- ERROR`,
				ErrorString: `table "fd_pt2_1" contains column "c4" not found in parent "fd_pt2"`,
			},
			{
				Statement: `DROP FOREIGN TABLE fd_pt2_1;`,
			},
			{
				Statement: `\d+ fd_pt2
                             Partitioned table "public.fd_pt2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Partition key: LIST (c1)
Number of partitions: 0
CREATE FOREIGN TABLE fd_pt2_1 (
	c1 integer NOT NULL,
	c2 text,
	c3 date
) SERVER s0 OPTIONS (delimiter ',', quote '"', "be quoted" 'value');`,
			},
			{
				Statement: `\d+ fd_pt2_1
                                     Foreign table "public.fd_pt2_1"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
ALTER TABLE fd_pt2 ATTACH PARTITION fd_pt2_1 FOR VALUES IN (1);`,
			},
			{
				Statement: `\d+ fd_pt2
                             Partitioned table "public.fd_pt2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Partition key: LIST (c1)
Partitions: fd_pt2_1 FOR VALUES IN (1)
\d+ fd_pt2_1
                                     Foreign table "public.fd_pt2_1"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           |          |         |             | plain    |              | 
Partition of: fd_pt2 FOR VALUES IN (1)
Partition constraint: ((c1 IS NOT NULL) AND (c1 = 1))
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
ALTER TABLE fd_pt2_1 ADD c4 char;`,
				ErrorString: `cannot add column to a partition`,
			},
			{
				Statement: `ALTER TABLE fd_pt2_1 ALTER c3 SET NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE fd_pt2_1 ADD CONSTRAINT p21chk CHECK (c2 <> '');`,
			},
			{
				Statement: `\d+ fd_pt2
                             Partitioned table "public.fd_pt2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           |          |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Partition key: LIST (c1)
Partitions: fd_pt2_1 FOR VALUES IN (1)
\d+ fd_pt2_1
                                     Foreign table "public.fd_pt2_1"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           | not null |         |             | plain    |              | 
Partition of: fd_pt2 FOR VALUES IN (1)
Partition constraint: ((c1 IS NOT NULL) AND (c1 = 1))
Check constraints:
    "p21chk" CHECK (c2 <> ''::text)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
ALTER TABLE fd_pt2_1 ALTER c1 DROP NOT NULL;`,
				ErrorString: `column "c1" is marked NOT NULL in parent table`,
			},
			{
				Statement: `ALTER TABLE fd_pt2 DETACH PARTITION fd_pt2_1;`,
			},
			{
				Statement: `ALTER TABLE fd_pt2 ALTER c2 SET NOT NULL;`,
			},
			{
				Statement: `\d+ fd_pt2
                             Partitioned table "public.fd_pt2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           | not null |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Partition key: LIST (c1)
Number of partitions: 0
\d+ fd_pt2_1
                                     Foreign table "public.fd_pt2_1"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           |          |         |             | extended |              | 
 c3     | date    |           | not null |         |             | plain    |              | 
Check constraints:
    "p21chk" CHECK (c2 <> ''::text)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
ALTER TABLE fd_pt2 ATTACH PARTITION fd_pt2_1 FOR VALUES IN (1);       -- ERROR`,
				ErrorString: `column "c2" in child table must be marked NOT NULL`,
			},
			{
				Statement: `ALTER FOREIGN TABLE fd_pt2_1 ALTER c2 SET NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE fd_pt2 ATTACH PARTITION fd_pt2_1 FOR VALUES IN (1);`,
			},
			{
				Statement: `ALTER TABLE fd_pt2 DETACH PARTITION fd_pt2_1;`,
			},
			{
				Statement: `ALTER TABLE fd_pt2 ADD CONSTRAINT fd_pt2chk1 CHECK (c1 > 0);`,
			},
			{
				Statement: `\d+ fd_pt2
                             Partitioned table "public.fd_pt2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 c1     | integer |           | not null |         | plain    |              | 
 c2     | text    |           | not null |         | extended |              | 
 c3     | date    |           |          |         | plain    |              | 
Partition key: LIST (c1)
Check constraints:
    "fd_pt2chk1" CHECK (c1 > 0)
Number of partitions: 0
\d+ fd_pt2_1
                                     Foreign table "public.fd_pt2_1"
 Column |  Type   | Collation | Nullable | Default | FDW options | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+-------------+----------+--------------+-------------
 c1     | integer |           | not null |         |             | plain    |              | 
 c2     | text    |           | not null |         |             | extended |              | 
 c3     | date    |           | not null |         |             | plain    |              | 
Check constraints:
    "p21chk" CHECK (c2 <> ''::text)
Server: s0
FDW options: (delimiter ',', quote '"', "be quoted" 'value')
ALTER TABLE fd_pt2 ATTACH PARTITION fd_pt2_1 FOR VALUES IN (1);       -- ERROR`,
				ErrorString: `child table is missing constraint "fd_pt2chk1"`,
			},
			{
				Statement: `ALTER FOREIGN TABLE fd_pt2_1 ADD CONSTRAINT fd_pt2chk1 CHECK (c1 > 0);`,
			},
			{
				Statement: `ALTER TABLE fd_pt2 ATTACH PARTITION fd_pt2_1 FOR VALUES IN (1);`,
			},
			{
				Statement: `DROP FOREIGN TABLE fd_pt2_1;`,
			},
			{
				Statement: `DROP TABLE fd_pt2;`,
			},
			{
				Statement: `CREATE TEMP TABLE temp_parted (a int) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE FOREIGN TABLE foreign_part PARTITION OF temp_parted DEFAULT
  SERVER s0;  -- ERROR`,
				ErrorString: `cannot create a permanent relation as partition of temporary relation "temp_parted"`,
			},
			{
				Statement: `CREATE FOREIGN TABLE foreign_part (a int) SERVER s0;`,
			},
			{
				Statement:   `ALTER TABLE temp_parted ATTACH PARTITION foreign_part DEFAULT;  -- ERROR`,
				ErrorString: `cannot attach a permanent relation as partition of temporary relation "temp_parted"`,
			},
			{
				Statement: `DROP FOREIGN TABLE foreign_part;`,
			},
			{
				Statement: `DROP TABLE temp_parted;`,
			},
			{
				Statement: `DROP SCHEMA foreign_schema CASCADE;`,
			},
			{
				Statement:   `DROP ROLE regress_test_role;                                -- ERROR`,
				ErrorString: `role "regress_test_role" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `privileges for server s4
owner of user mapping for regress_test_role on server s6
DROP SERVER t1 CASCADE;`,
			},
			{
				Statement: `DROP USER MAPPING FOR regress_test_role SERVER s6;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER foo CASCADE;`,
			},
			{
				Statement: `DROP SERVER s8 CASCADE;`,
			},
			{
				Statement: `DROP ROLE regress_test_indirect;`,
			},
			{
				Statement: `DROP ROLE regress_test_role;`,
			},
			{
				Statement:   `DROP ROLE regress_unprivileged_role;                        -- ERROR`,
				ErrorString: `role "regress_unprivileged_role" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `REVOKE ALL ON FOREIGN DATA WRAPPER postgresql FROM regress_unprivileged_role;`,
			},
			{
				Statement: `DROP ROLE regress_unprivileged_role;`,
			},
			{
				Statement: `DROP ROLE regress_test_role2;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER postgresql CASCADE;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER dummy CASCADE;`,
			},
			{
				Statement: `\c
DROP ROLE regress_foreign_data_user;`,
			},
			{
				Statement: `SELECT fdwname, fdwhandler, fdwvalidator, fdwoptions FROM pg_foreign_data_wrapper;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT srvname, srvoptions FROM pg_foreign_server;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM pg_user_mapping;`,
				Results:   []sql.Row{},
			},
		},
	})
}
