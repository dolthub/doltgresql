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

func TestDependency(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_dependency)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_dependency,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE USER regress_dep_user;`,
			},
			{
				Statement: `CREATE USER regress_dep_user2;`,
			},
			{
				Statement: `CREATE USER regress_dep_user3;`,
			},
			{
				Statement: `CREATE GROUP regress_dep_group;`,
			},
			{
				Statement: `CREATE TABLE deptest (f1 serial primary key, f2 text);`,
			},
			{
				Statement: `GRANT SELECT ON TABLE deptest TO GROUP regress_dep_group;`,
			},
			{
				Statement: `GRANT ALL ON TABLE deptest TO regress_dep_user, regress_dep_user2;`,
			},
			{
				Statement:   `DROP USER regress_dep_user;`,
				ErrorString: `role "regress_dep_user" cannot be dropped because some objects depend on it`,
			},
			{
				Statement:   `DROP GROUP regress_dep_group;`,
				ErrorString: `role "regress_dep_group" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `REVOKE SELECT ON deptest FROM GROUP regress_dep_group;`,
			},
			{
				Statement: `DROP GROUP regress_dep_group;`,
			},
			{
				Statement: `REVOKE SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES ON deptest FROM regress_dep_user;`,
			},
			{
				Statement:   `DROP USER regress_dep_user;`,
				ErrorString: `role "regress_dep_user" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `REVOKE TRIGGER ON deptest FROM regress_dep_user;`,
			},
			{
				Statement: `DROP USER regress_dep_user;`,
			},
			{
				Statement: `REVOKE ALL ON deptest FROM regress_dep_user2;`,
			},
			{
				Statement: `DROP USER regress_dep_user2;`,
			},
			{
				Statement: `\set VERBOSITY terse
ALTER TABLE deptest OWNER TO regress_dep_user3;`,
			},
			{
				Statement:   `DROP USER regress_dep_user3;`,
				ErrorString: `role "regress_dep_user3" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `\set VERBOSITY default
DROP TABLE deptest;`,
			},
			{
				Statement: `DROP USER regress_dep_user3;`,
			},
			{
				Statement: `CREATE USER regress_dep_user0;`,
			},
			{
				Statement: `CREATE USER regress_dep_user1;`,
			},
			{
				Statement: `CREATE USER regress_dep_user2;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_dep_user0;`,
			},
			{
				Statement:   `DROP OWNED BY regress_dep_user1;`,
				ErrorString: `permission denied to drop objects`,
			},
			{
				Statement:   `DROP OWNED BY regress_dep_user0, regress_dep_user2;`,
				ErrorString: `permission denied to drop objects`,
			},
			{
				Statement:   `REASSIGN OWNED BY regress_dep_user0 TO regress_dep_user1;`,
				ErrorString: `permission denied to reassign objects`,
			},
			{
				Statement:   `REASSIGN OWNED BY regress_dep_user1 TO regress_dep_user0;`,
				ErrorString: `permission denied to reassign objects`,
			},
			{
				Statement: `DROP OWNED BY regress_dep_user0;`,
			},
			{
				Statement: `CREATE TABLE deptest1 (f1 int unique);`,
			},
			{
				Statement: `GRANT ALL ON deptest1 TO regress_dep_user1 WITH GRANT OPTION;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_dep_user1;`,
			},
			{
				Statement: `CREATE TABLE deptest (a serial primary key, b text);`,
			},
			{
				Statement: `GRANT ALL ON deptest1 TO regress_dep_user2;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `\z deptest1
                                               Access privileges
 Schema |   Name   | Type  |                 Access privileges                  | Column privileges | Policies 
--------+----------+-------+----------------------------------------------------+-------------------+----------
 public | deptest1 | table | regress_dep_user0=arwdDxt/regress_dep_user0       +|                   | 
        |          |       | regress_dep_user1=a*r*w*d*D*x*t*/regress_dep_user0+|                   | 
        |          |       | regress_dep_user2=arwdDxt/regress_dep_user1        |                   | 
(1 row)
DROP OWNED BY regress_dep_user1;`,
			},
			{
				Statement: `\z deptest1
                                           Access privileges
 Schema |   Name   | Type  |              Access privileges              | Column privileges | Policies 
--------+----------+-------+---------------------------------------------+-------------------+----------
 public | deptest1 | table | regress_dep_user0=arwdDxt/regress_dep_user0 |                   | 
(1 row)
\d deptest
GRANT ALL ON deptest1 TO regress_dep_user1;`,
			},
			{
				Statement: `GRANT CREATE ON DATABASE regression TO regress_dep_user1;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_dep_user1;`,
			},
			{
				Statement: `CREATE SCHEMA deptest;`,
			},
			{
				Statement: `CREATE TABLE deptest (a serial primary key, b text);`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_dep_user1 IN SCHEMA deptest
  GRANT ALL ON TABLES TO regress_dep_user2;`,
			},
			{
				Statement: `CREATE FUNCTION deptest_func() RETURNS void LANGUAGE plpgsql
  AS $$ BEGIN END; $$;`,
			},
			{
				Statement: `CREATE TYPE deptest_enum AS ENUM ('red');`,
			},
			{
				Statement: `CREATE TYPE deptest_range AS RANGE (SUBTYPE = int4);`,
			},
			{
				Statement: `CREATE TABLE deptest2 (f1 int);`,
			},
			{
				Statement: `CREATE SEQUENCE ss1;`,
			},
			{
				Statement: `ALTER TABLE deptest2 ALTER f1 SET DEFAULT nextval('ss1');`,
			},
			{
				Statement: `ALTER SEQUENCE ss1 OWNED BY deptest2.f1;`,
			},
			{
				Statement: `CREATE TYPE deptest_t AS (a int);`,
			},
			{
				Statement: `SELECT typowner = relowner
FROM pg_type JOIN pg_class c ON typrelid = c.oid WHERE typname = 'deptest_t';`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `REASSIGN OWNED BY regress_dep_user1 TO regress_dep_user2;`,
			},
			{
				Statement: `\dt deptest
              List of relations
 Schema |  Name   | Type  |       Owner       
--------+---------+-------+-------------------
 public | deptest | table | regress_dep_user2
(1 row)
SELECT typowner = relowner
FROM pg_type JOIN pg_class c ON typrelid = c.oid WHERE typname = 'deptest_t';`,
				Results: []sql.Row{{true}},
			},
			{
				Statement:   `DROP USER regress_dep_user1;`,
				ErrorString: `role "regress_dep_user1" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `privileges for table deptest1
owner of default privileges on new relations belonging to role regress_dep_user1 in schema deptest
DROP OWNED BY regress_dep_user1;`,
			},
			{
				Statement: `DROP USER regress_dep_user1;`,
			},
			{
				Statement:   `DROP USER regress_dep_user2;`,
				ErrorString: `role "regress_dep_user2" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `owner of sequence deptest_a_seq
owner of table deptest
owner of function deptest_func()
owner of type deptest_enum
owner of type deptest_multirange
owner of type deptest_range
owner of table deptest2
owner of sequence ss1
owner of type deptest_t
DROP OWNED BY regress_dep_user2, regress_dep_user0;`,
			},
			{
				Statement: `DROP USER regress_dep_user2;`,
			},
			{
				Statement: `DROP USER regress_dep_user0;`,
			},
		},
	})
}
