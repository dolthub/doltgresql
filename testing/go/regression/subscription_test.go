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

func TestSubscription(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_subscription)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_subscription,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE ROLE regress_subscription_user LOGIN SUPERUSER;`,
			},
			{
				Statement: `CREATE ROLE regress_subscription_user2;`,
			},
			{
				Statement: `CREATE ROLE regress_subscription_user_dummy LOGIN NOSUPERUSER;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION 'regress_subscription_user';`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'foo';`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub PUBLICATION foo;`,
				ErrorString: `syntax error at or near "PUBLICATION"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'testconn' PUBLICATION testpub WITH (create_slot);`,
				ErrorString: `CREATE SUBSCRIPTION ... WITH (create_slot = true) cannot run inside a transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'testconn' PUBLICATION testpub;`,
				ErrorString: `invalid connection string syntax: missing "=" after "testconn" in connection info string`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION foo, testpub, foo WITH (connect = false);`,
				ErrorString: `publication name "foo" used more than once`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false);`,
			},
			{
				Statement: `COMMENT ON SUBSCRIPTION regress_testsub IS 'test subscription';`,
			},
			{
				Statement: `SELECT obj_description(s.oid, 'pg_subscription') FROM pg_subscription s;`,
				Results:   []sql.Row{{`test subscription`}},
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false);`,
				ErrorString: `subscription "regress_testsub" already exists`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION 'regress_subscription_user2';`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION foo WITH (connect = false);`,
				ErrorString: `must be superuser to create subscriptions`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION 'regress_subscription_user';`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, copy_data = true);`,
				ErrorString: `connect = false and copy_data = true are mutually exclusive options`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, enabled = true);`,
				ErrorString: `connect = false and enabled = true are mutually exclusive options`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, create_slot = true);`,
				ErrorString: `connect = false and create_slot = true are mutually exclusive options`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (slot_name = NONE, enabled = true);`,
				ErrorString: `slot_name = NONE and enabled = true are mutually exclusive options`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (slot_name = NONE, enabled = false, create_slot = true);`,
				ErrorString: `slot_name = NONE and create_slot = true are mutually exclusive options`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (slot_name = NONE);`,
				ErrorString: `subscription with slot_name = NONE must also set enabled = false`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (slot_name = NONE, enabled = false);`,
				ErrorString: `subscription with slot_name = NONE must also set create_slot = false`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub2 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (slot_name = NONE, create_slot = false);`,
				ErrorString: `subscription with slot_name = NONE must also set enabled = false`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_testsub3 CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (slot_name = NONE, connect = false);`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub3 ENABLE;`,
				ErrorString: `cannot enable subscription that does not have a slot name`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub3 REFRESH PUBLICATION;`,
				ErrorString: `ALTER SUBSCRIPTION ... REFRESH is not allowed for disabled subscriptions`,
			},
			{
				Statement: `DROP SUBSCRIPTION regress_testsub3;`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub5 CONNECTION 'i_dont_exist=param' PUBLICATION testpub;`,
				ErrorString: `invalid connection string syntax: invalid connection option "i_dont_exist"`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub5 CONNECTION 'port=-1' PUBLICATION testpub;`,
				ErrorString: `could not connect to the publisher: invalid port number: "-1"`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub CONNECTION 'foobar';`,
				ErrorString: `invalid connection string syntax: missing "=" after "foobar" in connection info string`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | f         | d                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub SET PUBLICATION testpub2, testpub3 WITH (refresh = false);`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist2';`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub SET (slot_name = 'newname');`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub SET (slot_name = '');`,
				ErrorString: `replication slot name "" is too short`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_doesnotexist CONNECTION 'dbname=regress_doesnotexist2';`,
				ErrorString: `subscription "regress_doesnotexist" does not exist`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub SET (create_slot = false);`,
				ErrorString: `unrecognized subscription parameter: "create_slot"`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub SKIP (lsn = '0/12345');`,
			},
			{
				Statement: `\dRs+
                                                                                         List of subscriptions
      Name       |           Owner           | Enabled |     Publication     | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |           Conninfo           | Skip LSN 
-----------------+---------------------------+---------+---------------------+--------+-----------+------------------+------------------+--------------------+------------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub2,testpub3} | f      | f         | d                | f                | off                | dbname=regress_doesnotexist2 | 0/12345
(1 row)
ALTER SUBSCRIPTION regress_testsub SKIP (lsn = NONE);`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub SKIP (lsn = '0/0');`,
				ErrorString: `invalid WAL location (LSN): 0/0`,
			},
			{
				Statement: `\dRs+
                                                                                         List of subscriptions
      Name       |           Owner           | Enabled |     Publication     | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |           Conninfo           | Skip LSN 
-----------------+---------------------------+---------+---------------------+--------+-----------+------------------+------------------+--------------------+------------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub2,testpub3} | f      | f         | d                | f                | off                | dbname=regress_doesnotexist2 | 0/0
(1 row)
BEGIN;`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub ENABLE;`,
			},
			{
				Statement: `\dRs
                            List of subscriptions
      Name       |           Owner           | Enabled |     Publication     
-----------------+---------------------------+---------+---------------------
 regress_testsub | regress_subscription_user | t       | {testpub2,testpub3}
(1 row)
ALTER SUBSCRIPTION regress_testsub DISABLE;`,
			},
			{
				Statement: `\dRs
                            List of subscriptions
      Name       |           Owner           | Enabled |     Publication     
-----------------+---------------------------+---------+---------------------
 regress_testsub | regress_subscription_user | f       | {testpub2,testpub3}
(1 row)
COMMIT;`,
			},
			{
				Statement: `SET ROLE regress_subscription_user_dummy;`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub RENAME TO regress_testsub_dummy;`,
				ErrorString: `must be owner of subscription regress_testsub`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub RENAME TO regress_testsub_foo;`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub_foo SET (synchronous_commit = local);`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub_foo SET (synchronous_commit = foobar);`,
				ErrorString: `invalid value for parameter "synchronous_commit": "foobar"`,
			},
			{
				Statement: `\dRs+
                                                                                           List of subscriptions
        Name         |           Owner           | Enabled |     Publication     | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |           Conninfo           | Skip LSN 
---------------------+---------------------------+---------+---------------------+--------+-----------+------------------+------------------+--------------------+------------------------------+----------
 regress_testsub_foo | regress_subscription_user | f       | {testpub2,testpub3} | f      | f         | d                | f                | local              | dbname=regress_doesnotexist2 | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub_foo RENAME TO regress_testsub;`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub OWNER TO regress_subscription_user2;`,
				ErrorString: `permission denied to change owner of subscription "regress_testsub"`,
			},
			{
				Statement: `ALTER ROLE regress_subscription_user2 SUPERUSER;`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub OWNER TO regress_subscription_user2;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `DROP SUBSCRIPTION regress_testsub;`,
				ErrorString: `DROP SUBSCRIPTION cannot run inside a transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub SET (slot_name = NONE);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DROP SUBSCRIPTION regress_testsub;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP SUBSCRIPTION IF EXISTS regress_testsub;`,
			},
			{
				Statement:   `DROP SUBSCRIPTION regress_testsub;  -- fail`,
				ErrorString: `subscription "regress_testsub" does not exist`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, binary = foo);`,
				ErrorString: `binary requires a Boolean value`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, binary = true);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | t      | f         | d                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub SET (binary = false);`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub SET (slot_name = NONE);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | f         | d                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
DROP SUBSCRIPTION regress_testsub;`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, streaming = foo);`,
				ErrorString: `streaming requires a Boolean value`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, streaming = true);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | t         | d                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub SET (streaming = false);`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub SET (slot_name = NONE);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | f         | d                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub ADD PUBLICATION testpub WITH (refresh = false);`,
				ErrorString: `publication "testpub" is already in subscription "regress_testsub"`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub ADD PUBLICATION testpub1, testpub1 WITH (refresh = false);`,
				ErrorString: `publication name "testpub1" used more than once`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub ADD PUBLICATION testpub1, testpub2 WITH (refresh = false);`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub ADD PUBLICATION testpub1, testpub2 WITH (refresh = false);`,
				ErrorString: `publication "testpub1" is already in subscription "regress_testsub"`,
			},
			{
				Statement: `\dRs+
                                                                                            List of subscriptions
      Name       |           Owner           | Enabled |         Publication         | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-----------------------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub,testpub1,testpub2} | f      | f         | d                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub DROP PUBLICATION testpub1, testpub1 WITH (refresh = false);`,
				ErrorString: `publication name "testpub1" used more than once`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub DROP PUBLICATION testpub, testpub1, testpub2 WITH (refresh = false);`,
				ErrorString: `cannot drop all the publications from a subscription`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub DROP PUBLICATION testpub3 WITH (refresh = false);`,
				ErrorString: `publication "testpub3" is not in subscription "regress_testsub"`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub DROP PUBLICATION testpub1, testpub2 WITH (refresh = false);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | f         | d                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
DROP SUBSCRIPTION regress_testsub;`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION mypub
       WITH (connect = false, create_slot = false, copy_data = false);`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub ENABLE;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub SET PUBLICATION mypub WITH (refresh = true);`,
				ErrorString: `ALTER SUBSCRIPTION with refresh cannot run inside a transaction block`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `ALTER SUBSCRIPTION regress_testsub REFRESH PUBLICATION;`,
				ErrorString: `ALTER SUBSCRIPTION ... REFRESH cannot run inside a transaction block`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `CREATE FUNCTION func() RETURNS VOID AS
$$ ALTER SUBSCRIPTION regress_testsub SET PUBLICATION mypub WITH (refresh = true) $$ LANGUAGE SQL;`,
			},
			{
				Statement:   `SELECT func();`,
				ErrorString: `ALTER SUBSCRIPTION with refresh cannot be executed from a function`,
			},
			{
				Statement: `CONTEXT:  SQL function "func" statement 1
ALTER SUBSCRIPTION regress_testsub DISABLE;`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub SET (slot_name = NONE);`,
			},
			{
				Statement: `DROP SUBSCRIPTION regress_testsub;`,
			},
			{
				Statement: `DROP FUNCTION func;`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, two_phase = foo);`,
				ErrorString: `two_phase requires a Boolean value`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, two_phase = true);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | f         | p                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub SET (two_phase = false);`,
				ErrorString: `unrecognized subscription parameter: "two_phase"`,
			},
			{
				Statement: `ALTER SUBSCRIPTION regress_testsub SET (streaming = true);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | t         | p                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub SET (slot_name = NONE);`,
			},
			{
				Statement: `DROP SUBSCRIPTION regress_testsub;`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, streaming = true, two_phase = true);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | t         | p                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub SET (slot_name = NONE);`,
			},
			{
				Statement: `DROP SUBSCRIPTION regress_testsub;`,
			},
			{
				Statement:   `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, disable_on_error = foo);`,
				ErrorString: `disable_on_error requires a Boolean value`,
			},
			{
				Statement: `CREATE SUBSCRIPTION regress_testsub CONNECTION 'dbname=regress_doesnotexist' PUBLICATION testpub WITH (connect = false, disable_on_error = false);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | f         | d                | f                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub SET (disable_on_error = true);`,
			},
			{
				Statement: `\dRs+
                                                                                    List of subscriptions
      Name       |           Owner           | Enabled | Publication | Binary | Streaming | Two-phase commit | Disable on error | Synchronous commit |          Conninfo           | Skip LSN 
-----------------+---------------------------+---------+-------------+--------+-----------+------------------+------------------+--------------------+-----------------------------+----------
 regress_testsub | regress_subscription_user | f       | {testpub}   | f      | f         | d                | t                | off                | dbname=regress_doesnotexist | 0/0
(1 row)
ALTER SUBSCRIPTION regress_testsub SET (slot_name = NONE);`,
			},
			{
				Statement: `DROP SUBSCRIPTION regress_testsub;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP ROLE regress_subscription_user;`,
			},
			{
				Statement: `DROP ROLE regress_subscription_user2;`,
			},
			{
				Statement: `DROP ROLE regress_subscription_user_dummy;`,
			},
		},
	})
}
