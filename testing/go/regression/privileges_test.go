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

func TestPrivileges(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_privileges)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_privileges,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET client_min_messages TO 'warning';`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_group1;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_group2;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_user1;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_user2;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_user3;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_user4;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_user5;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_user6;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_priv_user7;`,
			},
			{
				Statement: `SELECT lo_unlink(oid) FROM pg_largeobject_metadata WHERE oid >= 1000 AND oid < 3000 ORDER BY oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE USER regress_priv_user1;`,
			},
			{
				Statement: `CREATE USER regress_priv_user2;`,
			},
			{
				Statement: `CREATE USER regress_priv_user3;`,
			},
			{
				Statement: `CREATE USER regress_priv_user4;`,
			},
			{
				Statement: `CREATE USER regress_priv_user5;`,
			},
			{
				Statement:   `CREATE USER regress_priv_user5;	-- duplicate`,
				ErrorString: `role "regress_priv_user5" already exists`,
			},
			{
				Statement: `CREATE USER regress_priv_user6;`,
			},
			{
				Statement: `CREATE USER regress_priv_user7;`,
			},
			{
				Statement: `CREATE USER regress_priv_user8;`,
			},
			{
				Statement: `CREATE USER regress_priv_user9;`,
			},
			{
				Statement: `CREATE USER regress_priv_user10;`,
			},
			{
				Statement: `CREATE ROLE regress_priv_role;`,
			},
			{
				Statement: `GRANT pg_read_all_data TO regress_priv_user6;`,
			},
			{
				Statement: `GRANT pg_write_all_data TO regress_priv_user7;`,
			},
			{
				Statement: `GRANT pg_read_all_settings TO regress_priv_user8 WITH ADMIN OPTION;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user8;`,
			},
			{
				Statement: `GRANT pg_read_all_settings TO regress_priv_user9 WITH ADMIN OPTION;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user9;`,
			},
			{
				Statement: `GRANT pg_read_all_settings TO regress_priv_user10;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user8;`,
			},
			{
				Statement: `REVOKE pg_read_all_settings FROM regress_priv_user10;`,
			},
			{
				Statement: `REVOKE ADMIN OPTION FOR pg_read_all_settings FROM regress_priv_user9;`,
			},
			{
				Statement: `REVOKE pg_read_all_settings FROM regress_priv_user9;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `REVOKE ADMIN OPTION FOR pg_read_all_settings FROM regress_priv_user8;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user8;`,
			},
			{
				Statement: `SET ROLE pg_read_all_settings;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `REVOKE pg_read_all_settings FROM regress_priv_user8;`,
			},
			{
				Statement: `DROP USER regress_priv_user10;`,
			},
			{
				Statement: `DROP USER regress_priv_user9;`,
			},
			{
				Statement: `DROP USER regress_priv_user8;`,
			},
			{
				Statement: `CREATE GROUP regress_priv_group1;`,
			},
			{
				Statement: `CREATE GROUP regress_priv_group2 WITH USER regress_priv_user1, regress_priv_user2;`,
			},
			{
				Statement: `ALTER GROUP regress_priv_group1 ADD USER regress_priv_user4;`,
			},
			{
				Statement: `ALTER GROUP regress_priv_group2 ADD USER regress_priv_user2;	-- duplicate`,
			},
			{
				Statement: `ALTER GROUP regress_priv_group2 DROP USER regress_priv_user2;`,
			},
			{
				Statement: `GRANT regress_priv_group2 TO regress_priv_user4 WITH ADMIN OPTION;`,
			},
			{
				Statement: `CREATE FUNCTION leak(integer,integer) RETURNS boolean
  AS 'int4lt'
  LANGUAGE internal IMMUTABLE STRICT;  -- but deliberately not LEAKPROOF`,
			},
			{
				Statement: `ALTER FUNCTION leak(integer,integer) OWNER TO regress_priv_user1;`,
			},
			{
				Statement: `GRANT regress_priv_role TO regress_priv_user1 WITH ADMIN OPTION GRANTED BY CURRENT_ROLE;`,
			},
			{
				Statement: `REVOKE ADMIN OPTION FOR regress_priv_role FROM regress_priv_user1 GRANTED BY foo; -- error`,
			},
			{
				Statement: `REVOKE ADMIN OPTION FOR regress_priv_role FROM regress_priv_user1 GRANTED BY regress_priv_user2; -- error`,
			},
			{
				Statement: `REVOKE ADMIN OPTION FOR regress_priv_role FROM regress_priv_user1 GRANTED BY CURRENT_USER;`,
			},
			{
				Statement: `REVOKE regress_priv_role FROM regress_priv_user1 GRANTED BY CURRENT_ROLE;`,
			},
			{
				Statement: `DROP ROLE regress_priv_role;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `SELECT session_user, current_user;`,
				Results:   []sql.Row{{`regress_priv_user1`, `regress_priv_user1`}},
			},
			{
				Statement: `CREATE TABLE atest1 ( a int, b text );`,
			},
			{
				Statement: `SELECT * FROM atest1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `INSERT INTO atest1 VALUES (1, 'one');`,
			},
			{
				Statement: `DELETE FROM atest1;`,
			},
			{
				Statement: `UPDATE atest1 SET a = 1 WHERE b = 'blech';`,
			},
			{
				Statement: `TRUNCATE atest1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK atest1 IN ACCESS EXCLUSIVE MODE;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `REVOKE ALL ON atest1 FROM PUBLIC;`,
			},
			{
				Statement: `SELECT * FROM atest1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `GRANT ALL ON atest1 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT SELECT ON atest1 TO regress_priv_user3, regress_priv_user4;`,
			},
			{
				Statement: `SELECT * FROM atest1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `CREATE TABLE atest2 (col1 varchar(10), col2 boolean);`,
			},
			{
				Statement: `GRANT SELECT ON atest2 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT UPDATE ON atest2 TO regress_priv_user3;`,
			},
			{
				Statement: `GRANT INSERT ON atest2 TO regress_priv_user4 GRANTED BY CURRENT_USER;`,
			},
			{
				Statement: `GRANT TRUNCATE ON atest2 TO regress_priv_user5 GRANTED BY CURRENT_ROLE;`,
			},
			{
				Statement:   `GRANT TRUNCATE ON atest2 TO regress_priv_user4 GRANTED BY regress_priv_user5;  -- error`,
				ErrorString: `grantor must be current user`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `SELECT session_user, current_user;`,
				Results:   []sql.Row{{`regress_priv_user2`, `regress_priv_user2`}},
			},
			{
				Statement: `SELECT * FROM atest1; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM atest2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `INSERT INTO atest1 VALUES (2, 'two'); -- ok`,
			},
			{
				Statement:   `INSERT INTO atest2 VALUES ('foo', true); -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `INSERT INTO atest1 SELECT 1, b FROM atest1; -- ok`,
			},
			{
				Statement: `UPDATE atest1 SET a = 1 WHERE a = 2; -- ok`,
			},
			{
				Statement:   `UPDATE atest2 SET col2 = NOT col2; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `SELECT * FROM atest1 FOR UPDATE; -- ok`,
				Results:   []sql.Row{{1, `two`}, {1, `two`}},
			},
			{
				Statement:   `SELECT * FROM atest2 FOR UPDATE; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement:   `DELETE FROM atest2; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement:   `TRUNCATE atest2; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK atest2 IN ACCESS EXCLUSIVE MODE; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement:   `COPY atest2 FROM stdin; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `GRANT ALL ON atest1 TO PUBLIC; -- fail`,
			},
			{
				Statement: `SELECT * FROM atest1 WHERE ( b IN ( SELECT col1 FROM atest2 ) );`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM atest2 WHERE ( col1 IN ( SELECT b FROM atest1 ) );`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user6;`,
			},
			{
				Statement: `SELECT * FROM atest1; -- ok`,
				Results:   []sql.Row{{1, `two`}, {1, `two`}},
			},
			{
				Statement: `SELECT * FROM atest2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `INSERT INTO atest2 VALUES ('foo', true); -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user7;`,
			},
			{
				Statement:   `SELECT * FROM atest1; -- fail`,
				ErrorString: `permission denied for table atest1`,
			},
			{
				Statement:   `SELECT * FROM atest2; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `INSERT INTO atest2 VALUES ('foo', true); -- ok`,
			},
			{
				Statement: `UPDATE atest2 SET col2 = true; -- ok`,
			},
			{
				Statement: `DELETE FROM atest2; -- ok`,
			},
			{
				Statement:   `UPDATE pg_catalog.pg_class SET relname = '123'; -- fail`,
				ErrorString: `permission denied for table pg_class`,
			},
			{
				Statement:   `DELETE FROM pg_catalog.pg_class; -- fail`,
				ErrorString: `permission denied for table pg_class`,
			},
			{
				Statement:   `UPDATE pg_toast.pg_toast_1213 SET chunk_id = 1; -- fail`,
				ErrorString: `permission denied for table pg_toast_1213`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user3;`,
			},
			{
				Statement: `SELECT session_user, current_user;`,
				Results:   []sql.Row{{`regress_priv_user3`, `regress_priv_user3`}},
			},
			{
				Statement: `SELECT * FROM atest1; -- ok`,
				Results:   []sql.Row{{1, `two`}, {1, `two`}},
			},
			{
				Statement:   `SELECT * FROM atest2; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement:   `INSERT INTO atest1 VALUES (2, 'two'); -- fail`,
				ErrorString: `permission denied for table atest1`,
			},
			{
				Statement:   `INSERT INTO atest2 VALUES ('foo', true); -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement:   `INSERT INTO atest1 SELECT 1, b FROM atest1; -- fail`,
				ErrorString: `permission denied for table atest1`,
			},
			{
				Statement:   `UPDATE atest1 SET a = 1 WHERE a = 2; -- fail`,
				ErrorString: `permission denied for table atest1`,
			},
			{
				Statement: `UPDATE atest2 SET col2 = NULL; -- ok`,
			},
			{
				Statement:   `UPDATE atest2 SET col2 = NOT col2; -- fails; requires SELECT on atest2`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `UPDATE atest2 SET col2 = true FROM atest1 WHERE atest1.a = 5; -- ok`,
			},
			{
				Statement:   `SELECT * FROM atest1 FOR UPDATE; -- fail`,
				ErrorString: `permission denied for table atest1`,
			},
			{
				Statement:   `SELECT * FROM atest2 FOR UPDATE; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement:   `DELETE FROM atest2; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement:   `TRUNCATE atest2; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK atest2 IN ACCESS EXCLUSIVE MODE; -- ok`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement:   `COPY atest2 FROM stdin; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement:   `SELECT * FROM atest1 WHERE ( b IN ( SELECT col1 FROM atest2 ) );`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement:   `SELECT * FROM atest2 WHERE ( col1 IN ( SELECT b FROM atest1 ) );`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `COPY atest2 FROM stdin; -- ok`,
			},
			{
				Statement: `SELECT * FROM atest1; -- ok`,
				Results:   []sql.Row{{1, `two`}, {1, `two`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `CREATE TABLE atest12 as
  SELECT x AS a, 10001 - x AS b FROM generate_series(1,10000) x;`,
			},
			{
				Statement: `CREATE INDEX ON atest12 (a);`,
			},
			{
				Statement: `CREATE INDEX ON atest12 (abs(a));`,
			},
			{
				Statement: `ALTER TABLE atest12 SET (autovacuum_enabled = off);`,
			},
			{
				Statement: `SET default_statistics_target = 10000;`,
			},
			{
				Statement: `VACUUM ANALYZE atest12;`,
			},
			{
				Statement: `RESET default_statistics_target;`,
			},
			{
				Statement: `CREATE OPERATOR <<< (procedure = leak, leftarg = integer, rightarg = integer,
                     restrict = scalarltsel);`,
			},
			{
				Statement: `CREATE VIEW atest12v AS
  SELECT * FROM atest12 WHERE b <<< 5;`,
			},
			{
				Statement: `CREATE VIEW atest12sbv WITH (security_barrier=true) AS
  SELECT * FROM atest12 WHERE b <<< 5;`,
			},
			{
				Statement: `GRANT SELECT ON atest12v TO PUBLIC;`,
			},
			{
				Statement: `GRANT SELECT ON atest12sbv TO PUBLIC;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12v x, atest12v y WHERE x.a = y.b;`,
				Results:   []sql.Row{{`Nested Loop`}, {`->  Seq Scan on atest12 atest12_1`}, {`Filter: (b <<< 5)`}, {`->  Index Scan using atest12_a_idx on atest12`}, {`Index Cond: (a = atest12_1.b)`}, {`Filter: (b <<< 5)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12 x, atest12 y
  WHERE x.a = y.b and abs(y.a) <<< 5;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Seq Scan on atest12 y`}, {`Filter: (abs(a) <<< 5)`}, {`->  Index Scan using atest12_a_idx on atest12 x`}, {`Index Cond: (a = y.b)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12sbv x, atest12sbv y WHERE x.a = y.b;`,
				Results:   []sql.Row{{`Nested Loop`}, {`Join Filter: (atest12.a = atest12_1.b)`}, {`->  Seq Scan on atest12`}, {`Filter: (b <<< 5)`}, {`->  Materialize`}, {`->  Seq Scan on atest12 atest12_1`}, {`Filter: (b <<< 5)`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `CREATE FUNCTION leak2(integer,integer) RETURNS boolean
  AS $$begin raise notice 'leak % %', $1, $2; return $1 > $2; end$$
  LANGUAGE plpgsql immutable;`,
			},
			{
				Statement: `CREATE OPERATOR >>> (procedure = leak2, leftarg = integer, rightarg = integer,
                     restrict = scalargtsel);`,
			},
			{
				Statement:   `EXPLAIN (COSTS OFF) SELECT * FROM atest12 WHERE a >>> 0;`,
				ErrorString: `permission denied for table atest12`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12v x, atest12v y WHERE x.a = y.b;`,
				Results:   []sql.Row{{`Nested Loop`}, {`->  Seq Scan on atest12 atest12_1`}, {`Filter: (b <<< 5)`}, {`->  Index Scan using atest12_a_idx on atest12`}, {`Index Cond: (a = atest12_1.b)`}, {`Filter: (b <<< 5)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12sbv x, atest12sbv y WHERE x.a = y.b;`,
				Results:   []sql.Row{{`Nested Loop`}, {`Join Filter: (atest12.a = atest12_1.b)`}, {`->  Seq Scan on atest12`}, {`Filter: (b <<< 5)`}, {`->  Materialize`}, {`->  Seq Scan on atest12 atest12_1`}, {`Filter: (b <<< 5)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12v x, atest12v y
  WHERE x.a = y.b and abs(y.a) <<< 5;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Seq Scan on atest12 atest12_1`}, {`Filter: ((b <<< 5) AND (abs(a) <<< 5))`}, {`->  Index Scan using atest12_a_idx on atest12`}, {`Index Cond: (a = atest12_1.b)`}, {`Filter: (b <<< 5)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12sbv x, atest12sbv y
  WHERE x.a = y.b and abs(y.a) <<< 5;`,
				Results: []sql.Row{{`Nested Loop`}, {`Join Filter: (atest12_1.a = y.b)`}, {`->  Subquery Scan on y`}, {`Filter: (abs(y.a) <<< 5)`}, {`->  Seq Scan on atest12`}, {`Filter: (b <<< 5)`}, {`->  Seq Scan on atest12 atest12_1`}, {`Filter: (b <<< 5)`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT SELECT (a, b) ON atest12 TO PUBLIC;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12v x, atest12v y WHERE x.a = y.b;`,
				Results:   []sql.Row{{`Nested Loop`}, {`->  Seq Scan on atest12 atest12_1`}, {`Filter: (b <<< 5)`}, {`->  Index Scan using atest12_a_idx on atest12`}, {`Index Cond: (a = atest12_1.b)`}, {`Filter: (b <<< 5)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM atest12 x, atest12 y
  WHERE x.a = y.b and abs(y.a) <<< 5;`,
				Results: []sql.Row{{`Hash Join`}, {`Hash Cond: (x.a = y.b)`}, {`->  Seq Scan on atest12 x`}, {`->  Hash`}, {`->  Seq Scan on atest12 y`}, {`Filter: (abs(a) <<< 5)`}},
			},
			{
				Statement: `DROP FUNCTION leak2(integer, integer) CASCADE;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user3;`,
			},
			{
				Statement: `CREATE TABLE atest3 (one int, two int, three int);`,
			},
			{
				Statement: `GRANT DELETE ON atest3 TO GROUP regress_priv_group2;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement:   `SELECT * FROM atest3; -- fail`,
				ErrorString: `permission denied for table atest3`,
			},
			{
				Statement: `DELETE FROM atest3; -- ok`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `ALTER ROLE regress_priv_user1 NOINHERIT;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement:   `DELETE FROM atest3;`,
				ErrorString: `permission denied for table atest3`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user3;`,
			},
			{
				Statement: `CREATE VIEW atestv1 AS SELECT * FROM atest1; -- ok`,
			},
			{
				Statement: `/* The next *should* fail, but it's not implemented that way yet. */
CREATE VIEW atestv2 AS SELECT * FROM atest2;`,
			},
			{
				Statement: `CREATE VIEW atestv3 AS SELECT * FROM atest3; -- ok`,
			},
			{
				Statement: `/* Empty view is a corner case that failed in 9.2. */
CREATE VIEW atestv0 AS SELECT 0 as x WHERE false; -- ok`,
			},
			{
				Statement: `SELECT * FROM atestv1; -- ok`,
				Results:   []sql.Row{{1, `two`}, {1, `two`}},
			},
			{
				Statement:   `SELECT * FROM atestv2; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `GRANT SELECT ON atestv1, atestv3 TO regress_priv_user4;`,
			},
			{
				Statement: `GRANT SELECT ON atestv2 TO regress_priv_user2;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `SELECT * FROM atestv1; -- ok`,
				Results:   []sql.Row{{1, `two`}, {1, `two`}},
			},
			{
				Statement:   `SELECT * FROM atestv2; -- fail`,
				ErrorString: `permission denied for view atestv2`,
			},
			{
				Statement: `SELECT * FROM atestv3; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT * FROM atestv0; -- fail`,
				ErrorString: `permission denied for view atestv0`,
			},
			{
				Statement: `select * from
  ((select a.q1 as x from int8_tbl a offset 0)
   union all
   (select b.q2 as x from int8_tbl b offset 0)) ss
where false;`,
				ErrorString: `permission denied for table int8_tbl`,
			},
			{
				Statement: `set constraint_exclusion = on;`,
			},
			{
				Statement: `select * from
  ((select a.q1 as x, random() from int8_tbl a where q1 > 0)
   union all
   (select b.q2 as x, random() from int8_tbl b where q2 > 0)) ss
where x < 0;`,
				ErrorString: `permission denied for table int8_tbl`,
			},
			{
				Statement: `reset constraint_exclusion;`,
			},
			{
				Statement: `CREATE VIEW atestv4 AS SELECT * FROM atestv3; -- nested view`,
			},
			{
				Statement: `SELECT * FROM atestv4; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `GRANT SELECT ON atestv4 TO regress_priv_user2;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement:   `SELECT * FROM atestv3; -- fail`,
				ErrorString: `permission denied for view atestv3`,
			},
			{
				Statement: `SELECT * FROM atestv4; -- ok (even though regress_priv_user2 cannot access underlying atestv3)`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM atest2; -- ok`,
				Results:   []sql.Row{{`bar`, true}},
			},
			{
				Statement:   `SELECT * FROM atestv2; -- fail (even though regress_priv_user2 can access underlying atest2)`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `CREATE TABLE atest5 (one int, two int unique, three int, four int unique);`,
			},
			{
				Statement: `CREATE TABLE atest6 (one int, two int, blue int);`,
			},
			{
				Statement: `GRANT SELECT (one), INSERT (two), UPDATE (three) ON atest5 TO regress_priv_user4;`,
			},
			{
				Statement: `GRANT ALL (one) ON atest5 TO regress_priv_user3;`,
			},
			{
				Statement: `INSERT INTO atest5 VALUES (1,2,3);`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement:   `SELECT * FROM atest5; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SELECT one FROM atest5; -- ok`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `COPY atest5 (one) TO stdout; -- ok`,
			},
			{
				Statement: `1
SELECT two FROM atest5; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `COPY atest5 (two) TO stdout; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT atest5 FROM atest5; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `COPY atest5 (one,two) TO stdout; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SELECT 1 FROM atest5; -- ok`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT 1 FROM atest5 a JOIN atest5 b USING (one); -- ok`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT 1 FROM atest5 a JOIN atest5 b USING (two); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT 1 FROM atest5 a NATURAL JOIN atest5 b; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT * FROM (atest5 a JOIN atest5 b USING (one)) j; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT j.* FROM (atest5 a JOIN atest5 b USING (one)) j; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT (j.*) IS NULL FROM (atest5 a JOIN atest5 b USING (one)) j; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SELECT one FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)) j; -- ok`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT j.one FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)) j; -- ok`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT two FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)) j; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT j.two FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)) j; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT y FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)) j; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT j.y FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)) j; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT * FROM (atest5 a JOIN atest5 b USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT a.* FROM (atest5 a JOIN atest5 b USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT (a.*) IS NULL FROM (atest5 a JOIN atest5 b USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT two FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT a.two FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT y FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT b.y FROM (atest5 a JOIN atest5 b(one,x,y,z) USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT y FROM (atest5 a LEFT JOIN atest5 b(one,x,y,z) USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT b.y FROM (atest5 a LEFT JOIN atest5 b(one,x,y,z) USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT y FROM (atest5 a FULL JOIN atest5 b(one,x,y,z) USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT b.y FROM (atest5 a FULL JOIN atest5 b(one,x,y,z) USING (one)); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT 1 FROM atest5 WHERE two = 2; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `SELECT * FROM atest1, atest5; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SELECT atest1.* FROM atest1, atest5; -- ok`,
				Results:   []sql.Row{{1, `two`}, {1, `two`}},
			},
			{
				Statement: `SELECT atest1.*,atest5.one FROM atest1, atest5; -- ok`,
				Results:   []sql.Row{{1, `two`, 1}, {1, `two`, 1}},
			},
			{
				Statement:   `SELECT atest1.*,atest5.one FROM atest1 JOIN atest5 ON (atest1.a = atest5.two); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SELECT atest1.*,atest5.one FROM atest1 JOIN atest5 ON (atest1.a = atest5.one); -- ok`,
				Results:   []sql.Row{{1, `two`, 1}, {1, `two`, 1}},
			},
			{
				Statement:   `SELECT one, two FROM atest5; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT SELECT (one,two) ON atest6 TO regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement:   `SELECT one, two FROM atest5 NATURAL JOIN atest6; -- fail still`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT SELECT (two) ON atest5 TO regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `SELECT one, two FROM atest5 NATURAL JOIN atest6; -- ok now`,
				Results:   []sql.Row{},
			},
			{
				Statement: `INSERT INTO atest5 (two) VALUES (3); -- ok`,
			},
			{
				Statement:   `COPY atest5 FROM stdin; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `COPY atest5 (two) FROM stdin; -- ok`,
			},
			{
				Statement:   `INSERT INTO atest5 (three) VALUES (4); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `INSERT INTO atest5 VALUES (5,5,5); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `UPDATE atest5 SET three = 10; -- ok`,
			},
			{
				Statement:   `UPDATE atest5 SET one = 8; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `UPDATE atest5 SET three = 5, one = 2; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `INSERT INTO atest5(two) VALUES (6) ON CONFLICT (two) DO UPDATE set three = 10;`,
			},
			{
				Statement:   `INSERT INTO atest5(two) VALUES (6) ON CONFLICT (two) DO UPDATE set three = 10 RETURNING atest5.three;`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `INSERT INTO atest5(two) VALUES (6) ON CONFLICT (two) DO UPDATE set three = 10 RETURNING atest5.one;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `INSERT INTO atest5(two) VALUES (6) ON CONFLICT (two) DO UPDATE set three = EXCLUDED.one;`,
			},
			{
				Statement:   `INSERT INTO atest5(two) VALUES (6) ON CONFLICT (two) DO UPDATE set three = EXCLUDED.three;`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `INSERT INTO atest5(two) VALUES (6) ON CONFLICT (two) DO UPDATE set one = 8; -- fails (due to UPDATE)`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `INSERT INTO atest5(three) VALUES (4) ON CONFLICT (two) DO UPDATE set three = 10; -- fails (due to INSERT)`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `INSERT INTO atest5(four) VALUES (4); -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT INSERT (four) ON atest5 TO regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement:   `INSERT INTO atest5(four) VALUES (4) ON CONFLICT (four) DO UPDATE set three = 3; -- fails (due to SELECT)`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `INSERT INTO atest5(four) VALUES (4) ON CONFLICT ON CONSTRAINT atest5_four_key DO UPDATE set three = 3; -- fails (due to SELECT)`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `INSERT INTO atest5(four) VALUES (4); -- ok`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT SELECT (four) ON atest5 TO regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `INSERT INTO atest5(four) VALUES (4) ON CONFLICT (four) DO UPDATE set three = 3; -- ok`,
			},
			{
				Statement: `INSERT INTO atest5(four) VALUES (4) ON CONFLICT ON CONSTRAINT atest5_four_key DO UPDATE set three = 3; -- ok`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `REVOKE ALL (one) ON atest5 FROM regress_priv_user4;`,
			},
			{
				Statement: `GRANT SELECT (one,two,blue) ON atest6 TO regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement:   `SELECT one FROM atest5; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement:   `UPDATE atest5 SET one = 1; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SELECT atest6 FROM atest6; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `COPY atest6 TO stdout; -- ok`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `CREATE TABLE mtarget (a int, b text);`,
			},
			{
				Statement: `CREATE TABLE msource (a int, b text);`,
			},
			{
				Statement: `INSERT INTO mtarget VALUES (1, 'init1'), (2, 'init2');`,
			},
			{
				Statement: `INSERT INTO msource VALUES (1, 'source1'), (2, 'source2'), (3, 'source3');`,
			},
			{
				Statement: `GRANT SELECT (a) ON msource TO regress_priv_user4;`,
			},
			{
				Statement: `GRANT SELECT (a) ON mtarget TO regress_priv_user4;`,
			},
			{
				Statement: `GRANT INSERT (a,b) ON mtarget TO regress_priv_user4;`,
			},
			{
				Statement: `GRANT UPDATE (b) ON mtarget TO regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED THEN
	UPDATE SET b = s.b
WHEN NOT MATCHED THEN
	INSERT VALUES (a, NULL);`,
				ErrorString: `permission denied for table msource`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED THEN
	UPDATE SET b = 'x'
WHEN NOT MATCHED THEN
	INSERT VALUES (a, b);`,
				ErrorString: `permission denied for table msource`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED AND s.b = 'x' THEN
	UPDATE SET b = 'x'
WHEN NOT MATCHED THEN
	INSERT VALUES (a, NULL);`,
				ErrorString: `permission denied for table msource`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED THEN
	UPDATE SET b = 'ok'
WHEN NOT MATCHED THEN
	INSERT VALUES (a, NULL);`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT SELECT (b) ON msource TO regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED THEN
	UPDATE SET b = s.b
WHEN NOT MATCHED THEN
	INSERT VALUES (a, b);`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED THEN
	UPDATE SET b = t.b
WHEN NOT MATCHED THEN
	INSERT VALUES (a, NULL);`,
				ErrorString: `permission denied for table mtarget`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED THEN
	UPDATE SET b = s.b, a = t.a + 1
WHEN NOT MATCHED THEN
	INSERT VALUES (a, b);`,
				ErrorString: `permission denied for table mtarget`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED AND t.b IS NOT NULL THEN
	UPDATE SET b = s.b
WHEN NOT MATCHED THEN
	INSERT VALUES (a, b);`,
				ErrorString: `permission denied for table mtarget`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED THEN
	UPDATE SET b = s.b;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED AND t.b IS NOT NULL THEN
	DELETE;`,
				ErrorString: `permission denied for table mtarget`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT DELETE ON mtarget TO regress_priv_user4;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO mtarget t USING msource s ON t.a = s.a
WHEN MATCHED AND t.b IS NOT NULL THEN
	DELETE;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `CREATE TABLE t1 (c1 int, c2 int, c3 int check (c3 < 5), primary key (c1, c2));`,
			},
			{
				Statement: `GRANT SELECT (c1) ON t1 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT INSERT (c1, c2, c3) ON t1 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT UPDATE (c1, c2, c3) ON t1 TO regress_priv_user2;`,
			},
			{
				Statement: `INSERT INTO t1 VALUES (1, 1, 1);`,
			},
			{
				Statement: `INSERT INTO t1 VALUES (1, 2, 1);`,
			},
			{
				Statement: `INSERT INTO t1 VALUES (2, 1, 2);`,
			},
			{
				Statement: `INSERT INTO t1 VALUES (2, 2, 2);`,
			},
			{
				Statement: `INSERT INTO t1 VALUES (3, 1, 3);`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement:   `INSERT INTO t1 (c1, c2) VALUES (1, 1); -- fail, but row not shown`,
				ErrorString: `duplicate key value violates unique constraint "t1_pkey"`,
			},
			{
				Statement:   `UPDATE t1 SET c2 = 1; -- fail, but row not shown`,
				ErrorString: `duplicate key value violates unique constraint "t1_pkey"`,
			},
			{
				Statement:   `INSERT INTO t1 (c1, c2) VALUES (null, null); -- fail, but see columns being inserted`,
				ErrorString: `null value in column "c1" of relation "t1" violates not-null constraint`,
			},
			{
				Statement:   `INSERT INTO t1 (c3) VALUES (null); -- fail, but see columns being inserted or have SELECT`,
				ErrorString: `null value in column "c1" of relation "t1" violates not-null constraint`,
			},
			{
				Statement:   `INSERT INTO t1 (c1) VALUES (5); -- fail, but see columns being inserted or have SELECT`,
				ErrorString: `null value in column "c2" of relation "t1" violates not-null constraint`,
			},
			{
				Statement:   `UPDATE t1 SET c3 = 10; -- fail, but see columns with SELECT rights, or being modified`,
				ErrorString: `new row for relation "t1" violates check constraint "t1_c3_check"`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `DROP TABLE t1;`,
			},
			{
				Statement: `CREATE TABLE errtst(a text, b text NOT NULL, c text, secret1 text, secret2 text) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE errtst_part_1(secret2 text, c text, a text, b text NOT NULL, secret1 text);`,
			},
			{
				Statement: `CREATE TABLE errtst_part_2(secret1 text, secret2 text, a text, c text, b text NOT NULL);`,
			},
			{
				Statement: `ALTER TABLE errtst ATTACH PARTITION errtst_part_1 FOR VALUES IN ('aaa');`,
			},
			{
				Statement: `ALTER TABLE errtst ATTACH PARTITION errtst_part_2 FOR VALUES IN ('aaaa');`,
			},
			{
				Statement: `GRANT SELECT (a, b, c) ON TABLE errtst TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT UPDATE (a, b, c) ON TABLE errtst TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT INSERT (a, b, c) ON TABLE errtst TO regress_priv_user2;`,
			},
			{
				Statement: `INSERT INTO errtst_part_1 (a, b, c, secret1, secret2)
VALUES ('aaa', 'bbb', 'ccc', 'the body', 'is in the attic');`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement:   `INSERT INTO errtst (a, b) VALUES ('aaa', NULL);`,
				ErrorString: `null value in column "b" of relation "errtst_part_1" violates not-null constraint`,
			},
			{
				Statement:   `UPDATE errtst SET b = NULL;`,
				ErrorString: `null value in column "b" of relation "errtst_part_1" violates not-null constraint`,
			},
			{
				Statement:   `UPDATE errtst SET a = 'aaa', b = NULL;`,
				ErrorString: `null value in column "b" of relation "errtst_part_1" violates not-null constraint`,
			},
			{
				Statement:   `UPDATE errtst SET a = 'aaaa', b = NULL;`,
				ErrorString: `null value in column "b" of relation "errtst_part_2" violates not-null constraint`,
			},
			{
				Statement:   `UPDATE errtst SET a = 'aaaa', b = NULL WHERE a = 'aaa';`,
				ErrorString: `null value in column "b" of relation "errtst_part_2" violates not-null constraint`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `DROP TABLE errtst;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `ALTER TABLE atest6 ADD COLUMN three integer;`,
			},
			{
				Statement: `GRANT DELETE ON atest5 TO regress_priv_user3;`,
			},
			{
				Statement: `GRANT SELECT (two) ON atest5 TO regress_priv_user3;`,
			},
			{
				Statement: `REVOKE ALL (one) ON atest5 FROM regress_priv_user3;`,
			},
			{
				Statement: `GRANT SELECT (one) ON atest5 TO regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement:   `SELECT atest6 FROM atest6; -- fail`,
				ErrorString: `permission denied for table atest6`,
			},
			{
				Statement:   `SELECT one FROM atest5 NATURAL JOIN atest6; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `ALTER TABLE atest6 DROP COLUMN three;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `SELECT atest6 FROM atest6; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT one FROM atest5 NATURAL JOIN atest6; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `ALTER TABLE atest6 DROP COLUMN two;`,
			},
			{
				Statement: `REVOKE SELECT (one,blue) ON atest6 FROM regress_priv_user4;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement:   `SELECT * FROM atest6; -- fail`,
				ErrorString: `permission denied for table atest6`,
			},
			{
				Statement:   `SELECT 1 FROM atest6; -- fail`,
				ErrorString: `permission denied for table atest6`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user3;`,
			},
			{
				Statement:   `DELETE FROM atest5 WHERE one = 1; -- fail`,
				ErrorString: `permission denied for table atest5`,
			},
			{
				Statement: `DELETE FROM atest5 WHERE two = 2; -- ok`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `CREATE TABLE atestp1 (f1 int, f2 int);`,
			},
			{
				Statement: `CREATE TABLE atestp2 (fx int, fy int);`,
			},
			{
				Statement: `CREATE TABLE atestc (fz int) INHERITS (atestp1, atestp2);`,
			},
			{
				Statement: `GRANT SELECT(fx,fy,tableoid) ON atestp2 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT SELECT(fx) ON atestc TO regress_priv_user2;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `SELECT fx FROM atestp2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT fy FROM atestp2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT atestp2 FROM atestp2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT tableoid FROM atestp2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT fy FROM atestc; -- fail`,
				ErrorString: `permission denied for table atestc`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT SELECT(fy,tableoid) ON atestc TO regress_priv_user2;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `SELECT fx FROM atestp2; -- still ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT fy FROM atestp2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT atestp2 FROM atestp2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT tableoid FROM atestp2; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `REVOKE ALL ON atestc FROM regress_priv_user2;`,
			},
			{
				Statement: `GRANT ALL ON atestp1 TO regress_priv_user2;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `SELECT f2 FROM atestp1; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT f2 FROM atestc; -- fail`,
				ErrorString: `permission denied for table atestc`,
			},
			{
				Statement: `DELETE FROM atestp1; -- ok`,
			},
			{
				Statement:   `DELETE FROM atestc; -- fail`,
				ErrorString: `permission denied for table atestc`,
			},
			{
				Statement: `UPDATE atestp1 SET f1 = 1; -- ok`,
			},
			{
				Statement:   `UPDATE atestc SET f1 = 1; -- fail`,
				ErrorString: `permission denied for table atestc`,
			},
			{
				Statement: `TRUNCATE atestp1; -- ok`,
			},
			{
				Statement:   `TRUNCATE atestc; -- fail`,
				ErrorString: `permission denied for table atestc`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK atestp1;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK atestc;`,
				ErrorString: `permission denied for table atestc`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `\c -
REVOKE ALL PRIVILEGES ON LANGUAGE sql FROM PUBLIC;`,
			},
			{
				Statement: `GRANT USAGE ON LANGUAGE sql TO regress_priv_user1; -- ok`,
			},
			{
				Statement:   `GRANT USAGE ON LANGUAGE c TO PUBLIC; -- fail`,
				ErrorString: `language "c" is not trusted`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `GRANT USAGE ON LANGUAGE sql TO regress_priv_user2; -- fail`,
			},
			{
				Statement: `CREATE FUNCTION priv_testfunc1(int) RETURNS int AS 'select 2 * $1;' LANGUAGE sql;`,
			},
			{
				Statement: `CREATE FUNCTION priv_testfunc2(int) RETURNS int AS 'select 3 * $1;' LANGUAGE sql;`,
			},
			{
				Statement: `CREATE AGGREGATE priv_testagg1(int) (sfunc = int4pl, stype = int4);`,
			},
			{
				Statement: `CREATE PROCEDURE priv_testproc1(int) AS 'select $1;' LANGUAGE sql;`,
			},
			{
				Statement: `REVOKE ALL ON FUNCTION priv_testfunc1(int), priv_testfunc2(int), priv_testagg1(int) FROM PUBLIC;`,
			},
			{
				Statement: `GRANT EXECUTE ON FUNCTION priv_testfunc1(int), priv_testfunc2(int), priv_testagg1(int) TO regress_priv_user2;`,
			},
			{
				Statement:   `REVOKE ALL ON FUNCTION priv_testproc1(int) FROM PUBLIC; -- fail, not a function`,
				ErrorString: `priv_testproc1(integer) is not a function`,
			},
			{
				Statement: `REVOKE ALL ON PROCEDURE priv_testproc1(int) FROM PUBLIC;`,
			},
			{
				Statement: `GRANT EXECUTE ON PROCEDURE priv_testproc1(int) TO regress_priv_user2;`,
			},
			{
				Statement:   `GRANT USAGE ON FUNCTION priv_testfunc1(int) TO regress_priv_user3; -- semantic error`,
				ErrorString: `invalid privilege type USAGE for function`,
			},
			{
				Statement:   `GRANT USAGE ON FUNCTION priv_testagg1(int) TO regress_priv_user3; -- semantic error`,
				ErrorString: `invalid privilege type USAGE for function`,
			},
			{
				Statement:   `GRANT USAGE ON PROCEDURE priv_testproc1(int) TO regress_priv_user3; -- semantic error`,
				ErrorString: `invalid privilege type USAGE for procedure`,
			},
			{
				Statement: `GRANT ALL PRIVILEGES ON FUNCTION priv_testfunc1(int) TO regress_priv_user4;`,
			},
			{
				Statement:   `GRANT ALL PRIVILEGES ON FUNCTION priv_testfunc_nosuch(int) TO regress_priv_user4;`,
				ErrorString: `function priv_testfunc_nosuch(integer) does not exist`,
			},
			{
				Statement: `GRANT ALL PRIVILEGES ON FUNCTION priv_testagg1(int) TO regress_priv_user4;`,
			},
			{
				Statement: `GRANT ALL PRIVILEGES ON PROCEDURE priv_testproc1(int) TO regress_priv_user4;`,
			},
			{
				Statement: `CREATE FUNCTION priv_testfunc4(boolean) RETURNS text
  AS 'select col1 from atest2 where col2 = $1;'
  LANGUAGE sql SECURITY DEFINER;`,
			},
			{
				Statement: `GRANT EXECUTE ON FUNCTION priv_testfunc4(boolean) TO regress_priv_user3;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `SELECT priv_testfunc1(5), priv_testfunc2(5); -- ok`,
				Results:   []sql.Row{{10, 15}},
			},
			{
				Statement:   `CREATE FUNCTION priv_testfunc3(int) RETURNS int AS 'select 2 * $1;' LANGUAGE sql; -- fail`,
				ErrorString: `permission denied for language sql`,
			},
			{
				Statement: `SELECT priv_testagg1(x) FROM (VALUES (1), (2), (3)) _(x); -- ok`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `CALL priv_testproc1(6); -- ok`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user3;`,
			},
			{
				Statement:   `SELECT priv_testfunc1(5); -- fail`,
				ErrorString: `permission denied for function priv_testfunc1`,
			},
			{
				Statement:   `SELECT priv_testagg1(x) FROM (VALUES (1), (2), (3)) _(x); -- fail`,
				ErrorString: `permission denied for aggregate priv_testagg1`,
			},
			{
				Statement:   `CALL priv_testproc1(6); -- fail`,
				ErrorString: `permission denied for procedure priv_testproc1`,
			},
			{
				Statement:   `SELECT col1 FROM atest2 WHERE col2 = true; -- fail`,
				ErrorString: `permission denied for table atest2`,
			},
			{
				Statement: `SELECT priv_testfunc4(true); -- ok`,
				Results:   []sql.Row{{`bar`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `SELECT priv_testfunc1(5); -- ok`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `SELECT priv_testagg1(x) FROM (VALUES (1), (2), (3)) _(x); -- ok`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `CALL priv_testproc1(6); -- ok`,
			},
			{
				Statement:   `DROP FUNCTION priv_testfunc1(int); -- fail`,
				ErrorString: `must be owner of function priv_testfunc1`,
			},
			{
				Statement:   `DROP AGGREGATE priv_testagg1(int); -- fail`,
				ErrorString: `must be owner of aggregate priv_testagg1`,
			},
			{
				Statement:   `DROP PROCEDURE priv_testproc1(int); -- fail`,
				ErrorString: `must be owner of procedure priv_testproc1`,
			},
			{
				Statement: `\c -
DROP FUNCTION priv_testfunc1(int); -- ok`,
			},
			{
				Statement: `GRANT ALL PRIVILEGES ON LANGUAGE sql TO PUBLIC;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT '{1}'::int4[]::int8[];`,
				Results:   []sql.Row{{`{1}`}},
			},
			{
				Statement: `REVOKE ALL ON FUNCTION int8(integer) FROM PUBLIC;`,
			},
			{
				Statement: `SELECT '{1}'::int4[]::int8[]; --superuser, succeed
 int8 
------
 {1}
(1 row)
SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `SELECT '{1}'::int4[]::int8[]; --other user, fail
ERROR:  permission denied for function int8
ROLLBACK;`,
			},
			{
				Statement: `\c -
CREATE TYPE priv_testtype1 AS (a int, b text);`,
			},
			{
				Statement: `REVOKE USAGE ON TYPE priv_testtype1 FROM PUBLIC;`,
			},
			{
				Statement: `GRANT USAGE ON TYPE priv_testtype1 TO regress_priv_user2;`,
			},
			{
				Statement:   `GRANT USAGE ON TYPE _priv_testtype1 TO regress_priv_user2; -- fail`,
				ErrorString: `cannot set privileges of array types`,
			},
			{
				Statement:   `GRANT USAGE ON DOMAIN priv_testtype1 TO regress_priv_user2; -- fail`,
				ErrorString: `"priv_testtype1" is not a domain`,
			},
			{
				Statement: `CREATE DOMAIN priv_testdomain1 AS int;`,
			},
			{
				Statement: `REVOKE USAGE on DOMAIN priv_testdomain1 FROM PUBLIC;`,
			},
			{
				Statement: `GRANT USAGE ON DOMAIN priv_testdomain1 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT USAGE ON TYPE priv_testdomain1 TO regress_priv_user2; -- ok`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement:   `CREATE AGGREGATE priv_testagg1a(priv_testdomain1) (sfunc = int4_sum, stype = bigint);`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `CREATE DOMAIN priv_testdomain2a AS priv_testdomain1;`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement: `CREATE DOMAIN priv_testdomain3a AS int;`,
			},
			{
				Statement: `CREATE FUNCTION castfunc(int) RETURNS priv_testdomain3a AS $$ SELECT $1::priv_testdomain3a $$ LANGUAGE SQL;`,
			},
			{
				Statement:   `CREATE CAST (priv_testdomain1 AS priv_testdomain3a) WITH FUNCTION castfunc(int);`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement: `DROP FUNCTION castfunc(int) CASCADE;`,
			},
			{
				Statement: `DROP DOMAIN priv_testdomain3a;`,
			},
			{
				Statement:   `CREATE FUNCTION priv_testfunc5a(a priv_testdomain1) RETURNS int LANGUAGE SQL AS $$ SELECT $1 $$;`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `CREATE FUNCTION priv_testfunc6a(b int) RETURNS priv_testdomain1 LANGUAGE SQL AS $$ SELECT $1::priv_testdomain1 $$;`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `CREATE OPERATOR !+! (PROCEDURE = int4pl, LEFTARG = priv_testdomain1, RIGHTARG = priv_testdomain1);`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `CREATE TABLE test5a (a int, b priv_testdomain1);`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `CREATE TABLE test6a OF priv_testtype1;`,
				ErrorString: `permission denied for type priv_testtype1`,
			},
			{
				Statement:   `CREATE TABLE test10a (a int[], b priv_testtype1[]);`,
				ErrorString: `permission denied for type priv_testtype1`,
			},
			{
				Statement: `CREATE TABLE test9a (a int, b int);`,
			},
			{
				Statement:   `ALTER TABLE test9a ADD COLUMN c priv_testdomain1;`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `ALTER TABLE test9a ALTER COLUMN b TYPE priv_testdomain1;`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `CREATE TYPE test7a AS (a int, b priv_testdomain1);`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement: `CREATE TYPE test8a AS (a int, b int);`,
			},
			{
				Statement:   `ALTER TYPE test8a ADD ATTRIBUTE c priv_testdomain1;`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `ALTER TYPE test8a ALTER ATTRIBUTE b TYPE priv_testdomain1;`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `CREATE TABLE test11a AS (SELECT 1::priv_testdomain1 AS a);`,
				ErrorString: `permission denied for type priv_testdomain1`,
			},
			{
				Statement:   `REVOKE ALL ON TYPE priv_testtype1 FROM PUBLIC;`,
				ErrorString: `permission denied for type priv_testtype1`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `CREATE AGGREGATE priv_testagg1b(priv_testdomain1) (sfunc = int4_sum, stype = bigint);`,
			},
			{
				Statement: `CREATE DOMAIN priv_testdomain2b AS priv_testdomain1;`,
			},
			{
				Statement: `CREATE DOMAIN priv_testdomain3b AS int;`,
			},
			{
				Statement: `CREATE FUNCTION castfunc(int) RETURNS priv_testdomain3b AS $$ SELECT $1::priv_testdomain3b $$ LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE CAST (priv_testdomain1 AS priv_testdomain3b) WITH FUNCTION castfunc(int);`,
			},
			{
				Statement: `CREATE FUNCTION priv_testfunc5b(a priv_testdomain1) RETURNS int LANGUAGE SQL AS $$ SELECT $1 $$;`,
			},
			{
				Statement: `CREATE FUNCTION priv_testfunc6b(b int) RETURNS priv_testdomain1 LANGUAGE SQL AS $$ SELECT $1::priv_testdomain1 $$;`,
			},
			{
				Statement: `CREATE OPERATOR !! (PROCEDURE = priv_testfunc5b, RIGHTARG = priv_testdomain1);`,
			},
			{
				Statement: `CREATE TABLE test5b (a int, b priv_testdomain1);`,
			},
			{
				Statement: `CREATE TABLE test6b OF priv_testtype1;`,
			},
			{
				Statement: `CREATE TABLE test10b (a int[], b priv_testtype1[]);`,
			},
			{
				Statement: `CREATE TABLE test9b (a int, b int);`,
			},
			{
				Statement: `ALTER TABLE test9b ADD COLUMN c priv_testdomain1;`,
			},
			{
				Statement: `ALTER TABLE test9b ALTER COLUMN b TYPE priv_testdomain1;`,
			},
			{
				Statement: `CREATE TYPE test7b AS (a int, b priv_testdomain1);`,
			},
			{
				Statement: `CREATE TYPE test8b AS (a int, b int);`,
			},
			{
				Statement: `ALTER TYPE test8b ADD ATTRIBUTE c priv_testdomain1;`,
			},
			{
				Statement: `ALTER TYPE test8b ALTER ATTRIBUTE b TYPE priv_testdomain1;`,
			},
			{
				Statement: `CREATE TABLE test11b AS (SELECT 1::priv_testdomain1 AS a);`,
			},
			{
				Statement: `REVOKE ALL ON TYPE priv_testtype1 FROM PUBLIC;`,
			},
			{
				Statement: `\c -
DROP AGGREGATE priv_testagg1b(priv_testdomain1);`,
			},
			{
				Statement: `DROP DOMAIN priv_testdomain2b;`,
			},
			{
				Statement: `DROP OPERATOR !! (NONE, priv_testdomain1);`,
			},
			{
				Statement: `DROP FUNCTION priv_testfunc5b(a priv_testdomain1);`,
			},
			{
				Statement: `DROP FUNCTION priv_testfunc6b(b int);`,
			},
			{
				Statement: `DROP TABLE test5b;`,
			},
			{
				Statement: `DROP TABLE test6b;`,
			},
			{
				Statement: `DROP TABLE test9b;`,
			},
			{
				Statement: `DROP TABLE test10b;`,
			},
			{
				Statement: `DROP TYPE test7b;`,
			},
			{
				Statement: `DROP TYPE test8b;`,
			},
			{
				Statement: `DROP CAST (priv_testdomain1 AS priv_testdomain3b);`,
			},
			{
				Statement: `DROP FUNCTION castfunc(int) CASCADE;`,
			},
			{
				Statement: `DROP DOMAIN priv_testdomain3b;`,
			},
			{
				Statement: `DROP TABLE test11b;`,
			},
			{
				Statement: `DROP TYPE priv_testtype1; -- ok`,
			},
			{
				Statement: `DROP DOMAIN priv_testdomain1; -- ok`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user5;`,
			},
			{
				Statement: `TRUNCATE atest2; -- ok`,
			},
			{
				Statement:   `TRUNCATE atest3; -- fail`,
				ErrorString: `permission denied for table atest3`,
			},
			{
				Statement: `select has_table_privilege(NULL,'pg_authid','select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `select has_table_privilege('pg_shad','select');`,
				ErrorString: `relation "pg_shad" does not exist`,
			},
			{
				Statement:   `select has_table_privilege('nosuchuser','pg_authid','select');`,
				ErrorString: `role "nosuchuser" does not exist`,
			},
			{
				Statement:   `select has_table_privilege('pg_authid','sel');`,
				ErrorString: `unrecognized privilege type: "sel"`,
			},
			{
				Statement: `select has_table_privilege(-999999,'pg_authid','update');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(1,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `\c -
select has_table_privilege(current_user,'pg_authid','select');`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(current_user,'pg_authid','insert');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,'pg_authid','update')
from (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,'pg_authid','delete')
from (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(current_user,t1.oid,'rule')
from (select oid from pg_class where relname = 'pg_authid') as t1;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(current_user,t1.oid,'references')
from (select oid from pg_class where relname = 'pg_authid') as t1;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,t1.oid,'select')
from (select oid from pg_class where relname = 'pg_authid') as t1,
  (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,t1.oid,'insert')
from (select oid from pg_class where relname = 'pg_authid') as t1,
  (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege('pg_authid','update');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege('pg_authid','delete');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege('pg_authid','truncate');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t1.oid,'select')
from (select oid from pg_class where relname = 'pg_authid') as t1;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t1.oid,'trigger')
from (select oid from pg_class where relname = 'pg_authid') as t1;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user3;`,
			},
			{
				Statement: `select has_table_privilege(current_user,'pg_class','select');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(current_user,'pg_class','insert');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,'pg_class','update')
from (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,'pg_class','delete')
from (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(current_user,t1.oid,'references')
from (select oid from pg_class where relname = 'pg_class') as t1;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,t1.oid,'select')
from (select oid from pg_class where relname = 'pg_class') as t1,
  (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,t1.oid,'insert')
from (select oid from pg_class where relname = 'pg_class') as t1,
  (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege('pg_class','update');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege('pg_class','delete');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege('pg_class','truncate');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(t1.oid,'select')
from (select oid from pg_class where relname = 'pg_class') as t1;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t1.oid,'trigger')
from (select oid from pg_class where relname = 'pg_class') as t1;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(current_user,'atest1','select');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(current_user,'atest1','insert');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,'atest1','update')
from (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,'atest1','delete')
from (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(current_user,t1.oid,'references')
from (select oid from pg_class where relname = 'atest1') as t1;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,t1.oid,'select')
from (select oid from pg_class where relname = 'atest1') as t1,
  (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t2.oid,t1.oid,'insert')
from (select oid from pg_class where relname = 'atest1') as t1,
  (select oid from pg_roles where rolname = current_user) as t2;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege('atest1','update');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege('atest1','delete');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege('atest1','truncate');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select has_table_privilege(t1.oid,'select')
from (select oid from pg_class where relname = 'atest1') as t1;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `select has_table_privilege(t1.oid,'trigger')
from (select oid from pg_class where relname = 'atest1') as t1;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `select has_column_privilege('pg_authid',NULL,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `select has_column_privilege('pg_authid','nosuchcol','select');`,
				ErrorString: `column "nosuchcol" of relation "pg_authid" does not exist`,
			},
			{
				Statement: `select has_column_privilege(9999,'nosuchcol','select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select has_column_privilege(9999,99::int2,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select has_column_privilege('pg_authid',99::int2,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select has_column_privilege(9999,99::int2,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create temp table mytable(f1 int, f2 int, f3 int);`,
			},
			{
				Statement: `alter table mytable drop column f2;`,
			},
			{
				Statement:   `select has_column_privilege('mytable','f2','select');`,
				ErrorString: `column "f2" of relation "mytable" does not exist`,
			},
			{
				Statement: `select has_column_privilege('mytable','........pg.dropped.2........','select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select has_column_privilege('mytable',2::int2,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select has_column_privilege('mytable',99::int2,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `revoke select on table mytable from regress_priv_user3;`,
			},
			{
				Statement: `select has_column_privilege('mytable',2::int2,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select has_column_privilege('mytable',99::int2,'select');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `drop table mytable;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `CREATE TABLE atest4 (a int);`,
			},
			{
				Statement: `GRANT SELECT ON atest4 TO regress_priv_user2 WITH GRANT OPTION;`,
			},
			{
				Statement: `GRANT UPDATE ON atest4 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT SELECT ON atest4 TO GROUP regress_priv_group1 WITH GRANT OPTION;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `GRANT SELECT ON atest4 TO regress_priv_user3;`,
			},
			{
				Statement: `GRANT UPDATE ON atest4 TO regress_priv_user3; -- fail`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `REVOKE SELECT ON atest4 FROM regress_priv_user3; -- does nothing`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user3', 'atest4', 'SELECT'); -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `REVOKE SELECT ON atest4 FROM regress_priv_user2; -- fail`,
				ErrorString: `dependent privileges exist`,
			},
			{
				Statement: `REVOKE GRANT OPTION FOR SELECT ON atest4 FROM regress_priv_user2 CASCADE; -- ok`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user2', 'atest4', 'SELECT'); -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user3', 'atest4', 'SELECT'); -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'atest4', 'SELECT WITH GRANT OPTION'); -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `\c -
CREATE ROLE regress_sro_user;`,
			},
			{
				Statement: `CREATE FUNCTION sro_ifun(int) RETURNS int AS $$
BEGIN
	-- Below we set the table's owner to regress_sro_user
	ASSERT current_user = 'regress_sro_user',
		format('sro_ifun(%s) called by %s', $1, current_user);`,
			},
			{
				Statement: `	RETURN $1;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql IMMUTABLE;`,
			},
			{
				Statement: `CREATE TABLE sro_tab (a int);`,
			},
			{
				Statement: `ALTER TABLE sro_tab OWNER TO regress_sro_user;`,
			},
			{
				Statement: `INSERT INTO sro_tab VALUES (1), (2), (3);`,
			},
			{
				Statement: `CREATE INDEX sro_idx ON sro_tab ((sro_ifun(a) + sro_ifun(0)))
	WHERE sro_ifun(a + 10) > sro_ifun(10);`,
			},
			{
				Statement: `DROP INDEX sro_idx;`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY sro_idx ON sro_tab ((sro_ifun(a) + sro_ifun(0)))
	WHERE sro_ifun(a + 10) > sro_ifun(10);`,
			},
			{
				Statement: `REINDEX TABLE sro_tab;`,
			},
			{
				Statement: `REINDEX INDEX sro_idx;`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY sro_tab;`,
			},
			{
				Statement: `DROP INDEX sro_idx;`,
			},
			{
				Statement: `CREATE INDEX sro_cluster_idx ON sro_tab ((sro_ifun(a) + sro_ifun(0)));`,
			},
			{
				Statement: `CLUSTER sro_tab USING sro_cluster_idx;`,
			},
			{
				Statement: `DROP INDEX sro_cluster_idx;`,
			},
			{
				Statement: `CREATE INDEX sro_brin ON sro_tab USING brin ((sro_ifun(a) + sro_ifun(0)));`,
			},
			{
				Statement: `SELECT brin_desummarize_range('sro_brin', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT brin_summarize_range('sro_brin', 0);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP TABLE sro_tab;`,
			},
			{
				Statement: `CREATE TABLE sro_ptab (a int) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `ALTER TABLE sro_ptab OWNER TO regress_sro_user;`,
			},
			{
				Statement: `CREATE TABLE sro_part PARTITION OF sro_ptab FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `ALTER TABLE sro_part OWNER TO regress_sro_user;`,
			},
			{
				Statement: `INSERT INTO sro_ptab VALUES (1), (2), (3);`,
			},
			{
				Statement: `CREATE INDEX sro_pidx ON sro_ptab ((sro_ifun(a) + sro_ifun(0)))
	WHERE sro_ifun(a + 10) > sro_ifun(10);`,
			},
			{
				Statement: `REINDEX TABLE sro_ptab;`,
			},
			{
				Statement: `REINDEX INDEX CONCURRENTLY sro_pidx;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_sro_user;`,
			},
			{
				Statement: `CREATE FUNCTION unwanted_grant() RETURNS void LANGUAGE sql AS
	'GRANT regress_priv_group2 TO regress_sro_user';`,
			},
			{
				Statement: `CREATE FUNCTION mv_action() RETURNS bool LANGUAGE sql AS
	'DECLARE c CURSOR WITH HOLD FOR SELECT unwanted_grant(); SELECT true';`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW sro_mv AS SELECT mv_action() WITH NO DATA;`,
			},
			{
				Statement:   `REFRESH MATERIALIZED VIEW sro_mv;`,
				ErrorString: `cannot create a cursor WITH HOLD within security-restricted operation`,
			},
			{
				Statement: `CONTEXT:  SQL function "mv_action" statement 1
\c -
REFRESH MATERIALIZED VIEW sro_mv;`,
				ErrorString: `cannot create a cursor WITH HOLD within security-restricted operation`,
			},
			{
				Statement: `CONTEXT:  SQL function "mv_action" statement 1
SET SESSION AUTHORIZATION regress_sro_user;`,
			},
			{
				Statement: `CREATE TABLE sro_trojan_table ();`,
			},
			{
				Statement: `CREATE FUNCTION sro_trojan() RETURNS trigger LANGUAGE plpgsql AS
	'BEGIN PERFORM unwanted_grant(); RETURN NULL; END';`,
			},
			{
				Statement: `CREATE CONSTRAINT TRIGGER t AFTER INSERT ON sro_trojan_table
    INITIALLY DEFERRED FOR EACH ROW EXECUTE PROCEDURE sro_trojan();`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION mv_action() RETURNS bool LANGUAGE sql AS
	'INSERT INTO sro_trojan_table DEFAULT VALUES; SELECT true';`,
			},
			{
				Statement:   `REFRESH MATERIALIZED VIEW sro_mv;`,
				ErrorString: `cannot fire deferred trigger within security-restricted operation`,
			},
			{
				Statement: `CONTEXT:  SQL function "mv_action" statement 1
\c -
REFRESH MATERIALIZED VIEW sro_mv;`,
				ErrorString: `cannot fire deferred trigger within security-restricted operation`,
			},
			{
				Statement: `CONTEXT:  SQL function "mv_action" statement 1
BEGIN; SET CONSTRAINTS ALL IMMEDIATE; REFRESH MATERIALIZED VIEW sro_mv; COMMIT;`,
				ErrorString: `must have admin option on role "regress_priv_group2"`,
			},
			{
				Statement: `CONTEXT:  SQL function "unwanted_grant" statement 1
SQL statement "SELECT unwanted_grant()"
PL/pgSQL function sro_trojan() line 1 at PERFORM
SQL function "mv_action" statement 1
SET SESSION AUTHORIZATION regress_sro_user;`,
			},
			{
				Statement: `CREATE FUNCTION unwanted_grant_nofail(int) RETURNS int
	IMMUTABLE LANGUAGE plpgsql AS $$
BEGIN
	PERFORM unwanted_grant();`,
			},
			{
				Statement: `	RAISE WARNING 'owned';`,
			},
			{
				Statement: `	RETURN 1;`,
			},
			{
				Statement: `EXCEPTION WHEN OTHERS THEN
	RETURN 2;`,
			},
			{
				Statement: `END$$;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW sro_index_mv AS SELECT 1 AS c;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX ON sro_index_mv (c) WHERE unwanted_grant_nofail(1) > 0;`,
			},
			{
				Statement: `\c -
REFRESH MATERIALIZED VIEW CONCURRENTLY sro_index_mv;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW sro_index_mv;`,
			},
			{
				Statement: `DROP OWNED BY regress_sro_user;`,
			},
			{
				Statement: `DROP ROLE regress_sro_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `CREATE FUNCTION dogrant_ok() RETURNS void LANGUAGE sql SECURITY DEFINER AS
	'GRANT regress_priv_group2 TO regress_priv_user5';`,
			},
			{
				Statement: `GRANT regress_priv_group2 TO regress_priv_user5; -- ok: had ADMIN OPTION`,
			},
			{
				Statement: `SET ROLE regress_priv_group2;`,
			},
			{
				Statement:   `GRANT regress_priv_group2 TO regress_priv_user5; -- fails: SET ROLE suspended privilege`,
				ErrorString: `must have admin option on role "regress_priv_group2"`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement:   `GRANT regress_priv_group2 TO regress_priv_user5; -- fails: no ADMIN OPTION`,
				ErrorString: `must have admin option on role "regress_priv_group2"`,
			},
			{
				Statement: `SELECT dogrant_ok();			-- ok: SECURITY DEFINER conveys ADMIN`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SET ROLE regress_priv_group2;`,
			},
			{
				Statement:   `GRANT regress_priv_group2 TO regress_priv_user5; -- fails: SET ROLE did not help`,
				ErrorString: `must have admin option on role "regress_priv_group2"`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_group2;`,
			},
			{
				Statement:   `GRANT regress_priv_group2 TO regress_priv_user5; -- fails: no self-admin`,
				ErrorString: `must have admin option on role "regress_priv_group2"`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `DROP FUNCTION dogrant_ok();`,
			},
			{
				Statement: `REVOKE regress_priv_group2 FROM regress_priv_user5;`,
			},
			{
				Statement: `\c -
CREATE SEQUENCE x_seq;`,
			},
			{
				Statement: `GRANT USAGE on x_seq to regress_priv_user2;`,
			},
			{
				Statement:   `SELECT has_sequence_privilege('regress_priv_user1', 'atest1', 'SELECT');`,
				ErrorString: `"atest1" is not a sequence`,
			},
			{
				Statement:   `SELECT has_sequence_privilege('regress_priv_user1', 'x_seq', 'INSERT');`,
				ErrorString: `unrecognized privilege type: "INSERT"`,
			},
			{
				Statement: `SELECT has_sequence_privilege('regress_priv_user1', 'x_seq', 'SELECT');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `SELECT has_sequence_privilege('x_seq', 'USAGE');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `\c -
SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `SELECT lo_create(1001);`,
				Results:   []sql.Row{{1001}},
			},
			{
				Statement: `SELECT lo_create(1002);`,
				Results:   []sql.Row{{1002}},
			},
			{
				Statement: `SELECT lo_create(1003);`,
				Results:   []sql.Row{{1003}},
			},
			{
				Statement: `SELECT lo_create(1004);`,
				Results:   []sql.Row{{1004}},
			},
			{
				Statement: `SELECT lo_create(1005);`,
				Results:   []sql.Row{{1005}},
			},
			{
				Statement: `GRANT ALL ON LARGE OBJECT 1001 TO PUBLIC;`,
			},
			{
				Statement: `GRANT SELECT ON LARGE OBJECT 1003 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT SELECT,UPDATE ON LARGE OBJECT 1004 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT ALL ON LARGE OBJECT 1005 TO regress_priv_user2;`,
			},
			{
				Statement: `GRANT SELECT ON LARGE OBJECT 1005 TO regress_priv_user2 WITH GRANT OPTION;`,
			},
			{
				Statement:   `GRANT SELECT, INSERT ON LARGE OBJECT 1001 TO PUBLIC;	-- to be failed`,
				ErrorString: `invalid privilege type INSERT for large object`,
			},
			{
				Statement:   `GRANT SELECT, UPDATE ON LARGE OBJECT 1001 TO nosuchuser;	-- to be failed`,
				ErrorString: `role "nosuchuser" does not exist`,
			},
			{
				Statement:   `GRANT SELECT, UPDATE ON LARGE OBJECT  999 TO PUBLIC;	-- to be failed`,
				ErrorString: `large object 999 does not exist`,
			},
			{
				Statement: `\c -
SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `SELECT lo_create(2001);`,
				Results:   []sql.Row{{2001}},
			},
			{
				Statement: `SELECT lo_create(2002);`,
				Results:   []sql.Row{{2002}},
			},
			{
				Statement: `SELECT loread(lo_open(1001, x'20000'::int), 32);	-- allowed, for now`,
				Results:   []sql.Row{{`\x`}},
			},
			{
				Statement:   `SELECT lowrite(lo_open(1001, x'40000'::int), 'abcd');	-- fail, wrong mode`,
				ErrorString: `large object descriptor 0 was not opened for writing`,
			},
			{
				Statement: `SELECT loread(lo_open(1001, x'40000'::int), 32);`,
				Results:   []sql.Row{{`\x`}},
			},
			{
				Statement:   `SELECT loread(lo_open(1002, x'40000'::int), 32);	-- to be denied`,
				ErrorString: `permission denied for large object 1002`,
			},
			{
				Statement: `SELECT loread(lo_open(1003, x'40000'::int), 32);`,
				Results:   []sql.Row{{`\x`}},
			},
			{
				Statement: `SELECT loread(lo_open(1004, x'40000'::int), 32);`,
				Results:   []sql.Row{{`\x`}},
			},
			{
				Statement: `SELECT lowrite(lo_open(1001, x'20000'::int), 'abcd');`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement:   `SELECT lowrite(lo_open(1002, x'20000'::int), 'abcd');	-- to be denied`,
				ErrorString: `permission denied for large object 1002`,
			},
			{
				Statement:   `SELECT lowrite(lo_open(1003, x'20000'::int), 'abcd');	-- to be denied`,
				ErrorString: `permission denied for large object 1003`,
			},
			{
				Statement: `SELECT lowrite(lo_open(1004, x'20000'::int), 'abcd');`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `GRANT SELECT ON LARGE OBJECT 1005 TO regress_priv_user3;`,
			},
			{
				Statement:   `GRANT UPDATE ON LARGE OBJECT 1006 TO regress_priv_user3;	-- to be denied`,
				ErrorString: `large object 1006 does not exist`,
			},
			{
				Statement: `REVOKE ALL ON LARGE OBJECT 2001, 2002 FROM PUBLIC;`,
			},
			{
				Statement: `GRANT ALL ON LARGE OBJECT 2001 TO regress_priv_user3;`,
			},
			{
				Statement:   `SELECT lo_unlink(1001);		-- to be denied`,
				ErrorString: `must be owner of large object 1001`,
			},
			{
				Statement: `SELECT lo_unlink(2002);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `\c -
SELECT oid, pg_get_userbyid(lomowner) ownername, lomacl FROM pg_largeobject_metadata WHERE oid >= 1000 AND oid < 3000 ORDER BY oid;`,
				Results: []sql.Row{{1001, `regress_priv_user1`, `{regress_priv_user1=rw/regress_priv_user1,=rw/regress_priv_user1}`}, {1002, `regress_priv_user1`, ``}, {1003, `regress_priv_user1`, `{regress_priv_user1=rw/regress_priv_user1,regress_priv_user2=r/regress_priv_user1}`}, {1004, `regress_priv_user1`, `{regress_priv_user1=rw/regress_priv_user1,regress_priv_user2=rw/regress_priv_user1}`}, {1005, `regress_priv_user1`, `{regress_priv_user1=rw/regress_priv_user1,regress_priv_user2=r*w/regress_priv_user1,regress_priv_user3=r/regress_priv_user2}`}, {2001, `regress_priv_user2`, `{regress_priv_user2=rw/regress_priv_user2,regress_priv_user3=rw/regress_priv_user2}`}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user3;`,
			},
			{
				Statement: `SELECT loread(lo_open(1001, x'40000'::int), 32);`,
				Results:   []sql.Row{{`\x61626364`}},
			},
			{
				Statement:   `SELECT loread(lo_open(1003, x'40000'::int), 32);	-- to be denied`,
				ErrorString: `permission denied for large object 1003`,
			},
			{
				Statement: `SELECT loread(lo_open(1005, x'40000'::int), 32);`,
				Results:   []sql.Row{{`\x`}},
			},
			{
				Statement:   `SELECT lo_truncate(lo_open(1005, x'20000'::int), 10);	-- to be denied`,
				ErrorString: `permission denied for large object 1005`,
			},
			{
				Statement: `SELECT lo_truncate(lo_open(2001, x'20000'::int), 10);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `\c -
SET lo_compat_privileges = false;	-- default setting`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement:   `SELECT loread(lo_open(1002, x'40000'::int), 32);	-- to be denied`,
				ErrorString: `permission denied for large object 1002`,
			},
			{
				Statement:   `SELECT lowrite(lo_open(1002, x'20000'::int), 'abcd');	-- to be denied`,
				ErrorString: `permission denied for large object 1002`,
			},
			{
				Statement:   `SELECT lo_truncate(lo_open(1002, x'20000'::int), 10);	-- to be denied`,
				ErrorString: `permission denied for large object 1002`,
			},
			{
				Statement:   `SELECT lo_put(1002, 1, 'abcd');				-- to be denied`,
				ErrorString: `permission denied for large object 1002`,
			},
			{
				Statement:   `SELECT lo_unlink(1002);					-- to be denied`,
				ErrorString: `must be owner of large object 1002`,
			},
			{
				Statement:   `SELECT lo_export(1001, '/dev/null');			-- to be denied`,
				ErrorString: `permission denied for function lo_export`,
			},
			{
				Statement:   `SELECT lo_import('/dev/null');				-- to be denied`,
				ErrorString: `permission denied for function lo_import`,
			},
			{
				Statement:   `SELECT lo_import('/dev/null', 2003);			-- to be denied`,
				ErrorString: `permission denied for function lo_import`,
			},
			{
				Statement: `\c -
SET lo_compat_privileges = true;	-- compatibility mode`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user4;`,
			},
			{
				Statement: `SELECT loread(lo_open(1002, x'40000'::int), 32);`,
				Results:   []sql.Row{{`\x`}},
			},
			{
				Statement: `SELECT lowrite(lo_open(1002, x'20000'::int), 'abcd');`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `SELECT lo_truncate(lo_open(1002, x'20000'::int), 10);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT lo_unlink(1002);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT lo_export(1001, '/dev/null');			-- to be denied`,
				ErrorString: `permission denied for function lo_export`,
			},
			{
				Statement: `\c -
SELECT * FROM pg_largeobject LIMIT 0;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement:   `SELECT * FROM pg_largeobject LIMIT 0;			-- to be denied`,
				ErrorString: `permission denied for table pg_largeobject`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement:   `GRANT pg_database_owner TO regress_priv_user1;`,
				ErrorString: `role "pg_database_owner" cannot have explicit members`,
			},
			{
				Statement:   `GRANT regress_priv_user1 TO pg_database_owner;`,
				ErrorString: `role "pg_database_owner" cannot be a member of any role`,
			},
			{
				Statement: `CREATE TABLE datdba_only ();`,
			},
			{
				Statement: `ALTER TABLE datdba_only OWNER TO pg_database_owner;`,
			},
			{
				Statement: `REVOKE DELETE ON datdba_only FROM pg_database_owner;`,
			},
			{
				Statement: `SELECT
	pg_has_role('regress_priv_user1', 'pg_database_owner', 'USAGE') as priv,
	pg_has_role('regress_priv_user1', 'pg_database_owner', 'MEMBER') as mem,
	pg_has_role('regress_priv_user1', 'pg_database_owner',
				'MEMBER WITH ADMIN OPTION') as admin;`,
				Results: []sql.Row{{false, false, false}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DO $$BEGIN EXECUTE format(
	'ALTER DATABASE %I OWNER TO regress_priv_group2', current_catalog); END$$;`,
			},
			{
				Statement: `SELECT
	pg_has_role('regress_priv_user1', 'pg_database_owner', 'USAGE') as priv,
	pg_has_role('regress_priv_user1', 'pg_database_owner', 'MEMBER') as mem,
	pg_has_role('regress_priv_user1', 'pg_database_owner',
				'MEMBER WITH ADMIN OPTION') as admin;`,
				Results: []sql.Row{{true, true, false}},
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user1;`,
			},
			{
				Statement: `TABLE information_schema.enabled_roles ORDER BY role_name COLLATE "C";`,
				Results:   []sql.Row{{`pg_database_owner`}, {`regress_priv_group2`}, {`regress_priv_user1`}},
			},
			{
				Statement: `TABLE information_schema.applicable_roles ORDER BY role_name COLLATE "C";`,
				Results:   []sql.Row{{`regress_priv_group2`, `pg_database_owner`, `NO`}, {`regress_priv_user1`, `regress_priv_group2`, `NO`}},
			},
			{
				Statement: `INSERT INTO datdba_only DEFAULT VALUES;`,
			},
			{
				Statement:   `SAVEPOINT q; DELETE FROM datdba_only; ROLLBACK TO q;`,
				ErrorString: `permission denied for table datdba_only`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_priv_user2;`,
			},
			{
				Statement: `TABLE information_schema.enabled_roles;`,
				Results:   []sql.Row{{`regress_priv_user2`}},
			},
			{
				Statement:   `INSERT INTO datdba_only DEFAULT VALUES;`,
				ErrorString: `permission denied for table datdba_only`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `\c -
CREATE SCHEMA testns;`,
			},
			{
				Statement: `GRANT ALL ON SCHEMA testns TO regress_priv_user1;`,
			},
			{
				Statement: `CREATE TABLE testns.acltest1 (x int);`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'SELECT'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'INSERT'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES IN SCHEMA testns,testns GRANT SELECT ON TABLES TO public,public;`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'SELECT'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'INSERT'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `DROP TABLE testns.acltest1;`,
			},
			{
				Statement: `CREATE TABLE testns.acltest1 (x int);`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'SELECT'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'INSERT'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES IN SCHEMA testns GRANT INSERT ON TABLES TO regress_priv_user1;`,
			},
			{
				Statement: `DROP TABLE testns.acltest1;`,
			},
			{
				Statement: `CREATE TABLE testns.acltest1 (x int);`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'SELECT'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'INSERT'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES IN SCHEMA testns REVOKE INSERT ON TABLES FROM regress_priv_user1;`,
			},
			{
				Statement: `DROP TABLE testns.acltest1;`,
			},
			{
				Statement: `CREATE TABLE testns.acltest1 (x int);`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'SELECT'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.acltest1', 'INSERT'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_priv_user1 REVOKE EXECUTE ON FUNCTIONS FROM public;`,
			},
			{
				Statement:   `ALTER DEFAULT PRIVILEGES IN SCHEMA testns GRANT USAGE ON SCHEMAS TO regress_priv_user2; -- error`,
				ErrorString: `cannot use IN SCHEMA clause when using GRANT/REVOKE ON SCHEMAS`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES GRANT USAGE ON SCHEMAS TO regress_priv_user2;`,
			},
			{
				Statement: `CREATE SCHEMA testns2;`,
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user2', 'testns2', 'USAGE'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user6', 'testns2', 'USAGE'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user2', 'testns2', 'CREATE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES REVOKE USAGE ON SCHEMAS FROM regress_priv_user2;`,
			},
			{
				Statement: `CREATE SCHEMA testns3;`,
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user2', 'testns3', 'USAGE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user2', 'testns3', 'CREATE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES GRANT ALL ON SCHEMAS TO regress_priv_user2;`,
			},
			{
				Statement: `CREATE SCHEMA testns4;`,
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user2', 'testns4', 'USAGE'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user2', 'testns4', 'CREATE'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES REVOKE ALL ON SCHEMAS FROM regress_priv_user2;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES GRANT ALL ON FUNCTIONS TO regress_priv_user2;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES GRANT ALL ON SCHEMAS TO regress_priv_user2;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES GRANT ALL ON SEQUENCES TO regress_priv_user2;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES GRANT ALL ON TABLES TO regress_priv_user2;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES GRANT ALL ON TYPES TO regress_priv_user2;`,
			},
			{
				Statement: `SELECT count(*) FROM pg_shdepend
  WHERE deptype = 'a' AND
        refobjid = 'regress_priv_user2'::regrole AND
	classid = 'pg_default_acl'::regclass;`,
				Results: []sql.Row{{5}},
			},
			{
				Statement: `DROP OWNED BY regress_priv_user2, regress_priv_user2;`,
			},
			{
				Statement: `SELECT count(*) FROM pg_shdepend
  WHERE deptype = 'a' AND
        refobjid = 'regress_priv_user2'::regrole AND
	classid = 'pg_default_acl'::regclass;`,
				Results: []sql.Row{{0}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE SCHEMA testns5;`,
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user2', 'testns5', 'USAGE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_schema_privilege('regress_priv_user2', 'testns5', 'CREATE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SET ROLE regress_priv_user1;`,
			},
			{
				Statement: `CREATE FUNCTION testns.foo() RETURNS int AS 'select 1' LANGUAGE sql;`,
			},
			{
				Statement: `CREATE AGGREGATE testns.agg1(int) (sfunc = int4pl, stype = int4);`,
			},
			{
				Statement: `CREATE PROCEDURE testns.bar() AS 'select 1' LANGUAGE sql;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user2', 'testns.foo()', 'EXECUTE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user2', 'testns.agg1(int)', 'EXECUTE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user2', 'testns.bar()', 'EXECUTE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES IN SCHEMA testns GRANT EXECUTE ON ROUTINES to public;`,
			},
			{
				Statement: `DROP FUNCTION testns.foo();`,
			},
			{
				Statement: `CREATE FUNCTION testns.foo() RETURNS int AS 'select 1' LANGUAGE sql;`,
			},
			{
				Statement: `DROP AGGREGATE testns.agg1(int);`,
			},
			{
				Statement: `CREATE AGGREGATE testns.agg1(int) (sfunc = int4pl, stype = int4);`,
			},
			{
				Statement: `DROP PROCEDURE testns.bar();`,
			},
			{
				Statement: `CREATE PROCEDURE testns.bar() AS 'select 1' LANGUAGE sql;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user2', 'testns.foo()', 'EXECUTE'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user2', 'testns.agg1(int)', 'EXECUTE'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user2', 'testns.bar()', 'EXECUTE'); -- yes (counts as function here)`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP FUNCTION testns.foo();`,
			},
			{
				Statement: `DROP AGGREGATE testns.agg1(int);`,
			},
			{
				Statement: `DROP PROCEDURE testns.bar();`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_priv_user1 REVOKE USAGE ON TYPES FROM public;`,
			},
			{
				Statement: `CREATE DOMAIN testns.priv_testdomain1 AS int;`,
			},
			{
				Statement: `SELECT has_type_privilege('regress_priv_user2', 'testns.priv_testdomain1', 'USAGE'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES IN SCHEMA testns GRANT USAGE ON TYPES to public;`,
			},
			{
				Statement: `DROP DOMAIN testns.priv_testdomain1;`,
			},
			{
				Statement: `CREATE DOMAIN testns.priv_testdomain1 AS int;`,
			},
			{
				Statement: `SELECT has_type_privilege('regress_priv_user2', 'testns.priv_testdomain1', 'USAGE'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP DOMAIN testns.priv_testdomain1;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `SELECT count(*)
  FROM pg_default_acl d LEFT JOIN pg_namespace n ON defaclnamespace = n.oid
  WHERE nspname = 'testns';`,
				Results: []sql.Row{{3}},
			},
			{
				Statement: `DROP SCHEMA testns CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA testns2 CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA testns3 CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA testns4 CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA testns5 CASCADE;`,
			},
			{
				Statement: `SELECT d.*     -- check that entries went away
  FROM pg_default_acl d LEFT JOIN pg_namespace n ON defaclnamespace = n.oid
  WHERE nspname IS NULL AND defaclnamespace != 0;`,
				Results: []sql.Row{},
			},
			{
				Statement: `\c -
CREATE SCHEMA testns;`,
			},
			{
				Statement: `CREATE TABLE testns.t1 (f1 int);`,
			},
			{
				Statement: `CREATE TABLE testns.t2 (f1 int);`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.t1', 'SELECT'); -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `GRANT ALL ON ALL TABLES IN SCHEMA testns TO regress_priv_user1;`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.t1', 'SELECT'); -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.t2', 'SELECT'); -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `REVOKE ALL ON ALL TABLES IN SCHEMA testns FROM regress_priv_user1;`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.t1', 'SELECT'); -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_priv_user1', 'testns.t2', 'SELECT'); -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `CREATE FUNCTION testns.priv_testfunc(int) RETURNS int AS 'select 3 * $1;' LANGUAGE sql;`,
			},
			{
				Statement: `CREATE AGGREGATE testns.priv_testagg(int) (sfunc = int4pl, stype = int4);`,
			},
			{
				Statement: `CREATE PROCEDURE testns.priv_testproc(int) AS 'select 3' LANGUAGE sql;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testfunc(int)', 'EXECUTE'); -- true by default`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testagg(int)', 'EXECUTE'); -- true by default`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testproc(int)', 'EXECUTE'); -- true by default`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `REVOKE ALL ON ALL FUNCTIONS IN SCHEMA testns FROM PUBLIC;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testfunc(int)', 'EXECUTE'); -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testagg(int)', 'EXECUTE'); -- false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testproc(int)', 'EXECUTE'); -- still true, not a function`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `REVOKE ALL ON ALL PROCEDURES IN SCHEMA testns FROM PUBLIC;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testproc(int)', 'EXECUTE'); -- now false`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `GRANT ALL ON ALL ROUTINES IN SCHEMA testns TO PUBLIC;`,
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testfunc(int)', 'EXECUTE'); -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testagg(int)', 'EXECUTE'); -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_function_privilege('regress_priv_user1', 'testns.priv_testproc(int)', 'EXECUTE'); -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP SCHEMA testns CASCADE;`,
			},
			{
				Statement: `\c -
CREATE ROLE regress_schemauser1 superuser login;`,
			},
			{
				Statement: `CREATE ROLE regress_schemauser2 superuser login;`,
			},
			{
				Statement: `SET SESSION ROLE regress_schemauser1;`,
			},
			{
				Statement: `CREATE SCHEMA testns;`,
			},
			{
				Statement: `SELECT nspname, rolname FROM pg_namespace, pg_roles WHERE pg_namespace.nspname = 'testns' AND pg_namespace.nspowner = pg_roles.oid;`,
				Results:   []sql.Row{{`testns`, `regress_schemauser1`}},
			},
			{
				Statement: `ALTER SCHEMA testns OWNER TO regress_schemauser2;`,
			},
			{
				Statement: `ALTER ROLE regress_schemauser2 RENAME TO regress_schemauser_renamed;`,
			},
			{
				Statement: `SELECT nspname, rolname FROM pg_namespace, pg_roles WHERE pg_namespace.nspname = 'testns' AND pg_namespace.nspowner = pg_roles.oid;`,
				Results:   []sql.Row{{`testns`, `regress_schemauser_renamed`}},
			},
			{
				Statement: `set session role regress_schemauser_renamed;`,
			},
			{
				Statement: `DROP SCHEMA testns CASCADE;`,
			},
			{
				Statement: `\c -
DROP ROLE regress_schemauser1;`,
			},
			{
				Statement: `DROP ROLE regress_schemauser_renamed;`,
			},
			{
				Statement: `\c -
set session role regress_priv_user1;`,
			},
			{
				Statement: `create table dep_priv_test (a int);`,
			},
			{
				Statement: `grant select on dep_priv_test to regress_priv_user2 with grant option;`,
			},
			{
				Statement: `grant select on dep_priv_test to regress_priv_user3 with grant option;`,
			},
			{
				Statement: `set session role regress_priv_user2;`,
			},
			{
				Statement: `grant select on dep_priv_test to regress_priv_user4 with grant option;`,
			},
			{
				Statement: `set session role regress_priv_user3;`,
			},
			{
				Statement: `grant select on dep_priv_test to regress_priv_user4 with grant option;`,
			},
			{
				Statement: `set session role regress_priv_user4;`,
			},
			{
				Statement: `grant select on dep_priv_test to regress_priv_user5;`,
			},
			{
				Statement: `\dp dep_priv_test
                                               Access privileges
 Schema |     Name      | Type  |               Access privileges               | Column privileges | Policies 
--------+---------------+-------+-----------------------------------------------+-------------------+----------
 public | dep_priv_test | table | regress_priv_user1=arwdDxt/regress_priv_user1+|                   | 
        |               |       | regress_priv_user2=r*/regress_priv_user1     +|                   | 
        |               |       | regress_priv_user3=r*/regress_priv_user1     +|                   | 
        |               |       | regress_priv_user4=r*/regress_priv_user2     +|                   | 
        |               |       | regress_priv_user4=r*/regress_priv_user3     +|                   | 
        |               |       | regress_priv_user5=r/regress_priv_user4       |                   | 
(1 row)
set session role regress_priv_user2;`,
			},
			{
				Statement: `revoke select on dep_priv_test from regress_priv_user4 cascade;`,
			},
			{
				Statement: `\dp dep_priv_test
                                               Access privileges
 Schema |     Name      | Type  |               Access privileges               | Column privileges | Policies 
--------+---------------+-------+-----------------------------------------------+-------------------+----------
 public | dep_priv_test | table | regress_priv_user1=arwdDxt/regress_priv_user1+|                   | 
        |               |       | regress_priv_user2=r*/regress_priv_user1     +|                   | 
        |               |       | regress_priv_user3=r*/regress_priv_user1     +|                   | 
        |               |       | regress_priv_user4=r*/regress_priv_user3     +|                   | 
        |               |       | regress_priv_user5=r/regress_priv_user4       |                   | 
(1 row)
set session role regress_priv_user3;`,
			},
			{
				Statement: `revoke select on dep_priv_test from regress_priv_user4 cascade;`,
			},
			{
				Statement: `\dp dep_priv_test
                                               Access privileges
 Schema |     Name      | Type  |               Access privileges               | Column privileges | Policies 
--------+---------------+-------+-----------------------------------------------+-------------------+----------
 public | dep_priv_test | table | regress_priv_user1=arwdDxt/regress_priv_user1+|                   | 
        |               |       | regress_priv_user2=r*/regress_priv_user1     +|                   | 
        |               |       | regress_priv_user3=r*/regress_priv_user1      |                   | 
(1 row)
set session role regress_priv_user1;`,
			},
			{
				Statement: `drop table dep_priv_test;`,
			},
			{
				Statement: `\c
drop sequence x_seq;`,
			},
			{
				Statement: `DROP AGGREGATE priv_testagg1(int);`,
			},
			{
				Statement: `DROP FUNCTION priv_testfunc2(int);`,
			},
			{
				Statement: `DROP FUNCTION priv_testfunc4(boolean);`,
			},
			{
				Statement: `DROP PROCEDURE priv_testproc1(int);`,
			},
			{
				Statement: `DROP VIEW atestv0;`,
			},
			{
				Statement: `DROP VIEW atestv1;`,
			},
			{
				Statement: `DROP VIEW atestv2;`,
			},
			{
				Statement: `DROP VIEW atestv3 CASCADE;`,
			},
			{
				Statement:   `DROP VIEW atestv4;`,
				ErrorString: `view "atestv4" does not exist`,
			},
			{
				Statement: `DROP TABLE atest1;`,
			},
			{
				Statement: `DROP TABLE atest2;`,
			},
			{
				Statement: `DROP TABLE atest3;`,
			},
			{
				Statement: `DROP TABLE atest4;`,
			},
			{
				Statement: `DROP TABLE atest5;`,
			},
			{
				Statement: `DROP TABLE atest6;`,
			},
			{
				Statement: `DROP TABLE atestc;`,
			},
			{
				Statement: `DROP TABLE atestp1;`,
			},
			{
				Statement: `DROP TABLE atestp2;`,
			},
			{
				Statement: `SELECT lo_unlink(oid) FROM pg_largeobject_metadata WHERE oid >= 1000 AND oid < 3000 ORDER BY oid;`,
				Results:   []sql.Row{{1}, {1}, {1}, {1}, {1}},
			},
			{
				Statement: `DROP GROUP regress_priv_group1;`,
			},
			{
				Statement: `DROP GROUP regress_priv_group2;`,
			},
			{
				Statement: `REVOKE USAGE ON LANGUAGE sql FROM regress_priv_user1;`,
			},
			{
				Statement: `DROP OWNED BY regress_priv_user1;`,
			},
			{
				Statement: `DROP USER regress_priv_user1;`,
			},
			{
				Statement: `DROP USER regress_priv_user2;`,
			},
			{
				Statement: `DROP USER regress_priv_user3;`,
			},
			{
				Statement: `DROP USER regress_priv_user4;`,
			},
			{
				Statement: `DROP USER regress_priv_user5;`,
			},
			{
				Statement: `DROP USER regress_priv_user6;`,
			},
			{
				Statement: `DROP USER regress_priv_user7;`,
			},
			{
				Statement:   `DROP USER regress_priv_user8; -- does not exist`,
				ErrorString: `role "regress_priv_user8" does not exist`,
			},
			{
				Statement: `CREATE USER regress_locktable_user;`,
			},
			{
				Statement: `CREATE TABLE lock_table (a int);`,
			},
			{
				Statement: `GRANT SELECT ON lock_table TO regress_locktable_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_locktable_user;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_table IN ROW EXCLUSIVE MODE; -- should fail`,
				ErrorString: `permission denied for table lock_table`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_table IN ACCESS SHARE MODE; -- should pass`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_table IN ACCESS EXCLUSIVE MODE; -- should fail`,
				ErrorString: `permission denied for table lock_table`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `\c
REVOKE SELECT ON lock_table FROM regress_locktable_user;`,
			},
			{
				Statement: `GRANT INSERT ON lock_table TO regress_locktable_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_locktable_user;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_table IN ROW EXCLUSIVE MODE; -- should pass`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_table IN ACCESS SHARE MODE; -- should fail`,
				ErrorString: `permission denied for table lock_table`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_table IN ACCESS EXCLUSIVE MODE; -- should fail`,
				ErrorString: `permission denied for table lock_table`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `\c
REVOKE INSERT ON lock_table FROM regress_locktable_user;`,
			},
			{
				Statement: `GRANT UPDATE ON lock_table TO regress_locktable_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_locktable_user;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_table IN ROW EXCLUSIVE MODE; -- should pass`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_table IN ACCESS SHARE MODE; -- should fail`,
				ErrorString: `permission denied for table lock_table`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_table IN ACCESS EXCLUSIVE MODE; -- should pass`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `\c
REVOKE UPDATE ON lock_table FROM regress_locktable_user;`,
			},
			{
				Statement: `GRANT DELETE ON lock_table TO regress_locktable_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_locktable_user;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_table IN ROW EXCLUSIVE MODE; -- should pass`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_table IN ACCESS SHARE MODE; -- should fail`,
				ErrorString: `permission denied for table lock_table`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_table IN ACCESS EXCLUSIVE MODE; -- should pass`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `\c
REVOKE DELETE ON lock_table FROM regress_locktable_user;`,
			},
			{
				Statement: `GRANT TRUNCATE ON lock_table TO regress_locktable_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_locktable_user;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_table IN ROW EXCLUSIVE MODE; -- should pass`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_table IN ACCESS SHARE MODE; -- should fail`,
				ErrorString: `permission denied for table lock_table`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_table IN ACCESS EXCLUSIVE MODE; -- should pass`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `\c
REVOKE TRUNCATE ON lock_table FROM regress_locktable_user;`,
			},
			{
				Statement: `DROP TABLE lock_table;`,
			},
			{
				Statement: `DROP USER regress_locktable_user;`,
			},
			{
				Statement: `\c -
CREATE ROLE regress_readallstats;`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_readallstats','pg_backend_memory_contexts','SELECT'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_readallstats','pg_shmem_allocations','SELECT'); -- no`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `GRANT pg_read_all_stats TO regress_readallstats;`,
			},
			{
				Statement: `SELECT has_table_privilege('regress_readallstats','pg_backend_memory_contexts','SELECT'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT has_table_privilege('regress_readallstats','pg_shmem_allocations','SELECT'); -- yes`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SET ROLE regress_readallstats;`,
			},
			{
				Statement: `SELECT COUNT(*) >= 0 AS ok FROM pg_backend_memory_contexts;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT COUNT(*) >= 0 AS ok FROM pg_shmem_allocations;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `DROP ROLE regress_readallstats;`,
			},
		},
	})
}
