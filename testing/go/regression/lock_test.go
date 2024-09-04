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

func TestLock(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_lock)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_lock,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set regresslib :libdir '/regress' :dlsuffix
CREATE SCHEMA lock_schema1;`,
			},
			{
				Statement: `SET search_path = lock_schema1;`,
			},
			{
				Statement: `CREATE TABLE lock_tbl1 (a BIGINT);`,
			},
			{
				Statement: `CREATE TABLE lock_tbl1a (a BIGINT);`,
			},
			{
				Statement: `CREATE VIEW lock_view1 AS SELECT * FROM lock_tbl1;`,
			},
			{
				Statement: `CREATE VIEW lock_view2(a,b) AS SELECT * FROM lock_tbl1, lock_tbl1a;`,
			},
			{
				Statement: `CREATE VIEW lock_view3 AS SELECT * from lock_view2;`,
			},
			{
				Statement: `CREATE VIEW lock_view4 AS SELECT (select a from lock_tbl1a limit 1) from lock_tbl1;`,
			},
			{
				Statement: `CREATE VIEW lock_view5 AS SELECT * from lock_tbl1 where a in (select * from lock_tbl1a);`,
			},
			{
				Statement: `CREATE VIEW lock_view6 AS SELECT * from (select * from lock_tbl1) sub;`,
			},
			{
				Statement: `CREATE ROLE regress_rol_lock1;`,
			},
			{
				Statement: `ALTER ROLE regress_rol_lock1 SET search_path = lock_schema1;`,
			},
			{
				Statement: `GRANT USAGE ON SCHEMA lock_schema1 TO regress_rol_lock1;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN ACCESS SHARE MODE;`,
			},
			{
				Statement: `LOCK lock_tbl1 IN ROW SHARE MODE;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN ROW EXCLUSIVE MODE;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN SHARE UPDATE EXCLUSIVE MODE;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN SHARE MODE;`,
			},
			{
				Statement: `LOCK lock_tbl1 IN SHARE ROW EXCLUSIVE MODE;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN ACCESS EXCLUSIVE MODE;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN ACCESS SHARE MODE NOWAIT;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN ROW SHARE MODE NOWAIT;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN ROW EXCLUSIVE MODE NOWAIT;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN SHARE UPDATE EXCLUSIVE MODE NOWAIT;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN SHARE MODE NOWAIT;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN SHARE ROW EXCLUSIVE MODE NOWAIT;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN EXCLUSIVE MODE NOWAIT;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 IN ACCESS EXCLUSIVE MODE NOWAIT;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_view1 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `select relname from pg_locks l, pg_class c
 where l.relation = c.oid and relname like '%lock_%' and mode = 'ExclusiveLock'
 order by relname;`,
				Results: []sql.Row{{`lock_tbl1`}, {`lock_view1`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_view2 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `select relname from pg_locks l, pg_class c
 where l.relation = c.oid and relname like '%lock_%' and mode = 'ExclusiveLock'
 order by relname;`,
				Results: []sql.Row{{`lock_tbl1`}, {`lock_tbl1a`}, {`lock_view2`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_view3 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `select relname from pg_locks l, pg_class c
 where l.relation = c.oid and relname like '%lock_%' and mode = 'ExclusiveLock'
 order by relname;`,
				Results: []sql.Row{{`lock_tbl1`}, {`lock_tbl1a`}, {`lock_view2`}, {`lock_view3`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_view4 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `select relname from pg_locks l, pg_class c
 where l.relation = c.oid and relname like '%lock_%' and mode = 'ExclusiveLock'
 order by relname;`,
				Results: []sql.Row{{`lock_tbl1`}, {`lock_tbl1a`}, {`lock_view4`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_view5 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `select relname from pg_locks l, pg_class c
 where l.relation = c.oid and relname like '%lock_%' and mode = 'ExclusiveLock'
 order by relname;`,
				Results: []sql.Row{{`lock_tbl1`}, {`lock_tbl1a`}, {`lock_view5`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_view6 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `select relname from pg_locks l, pg_class c
 where l.relation = c.oid and relname like '%lock_%' and mode = 'ExclusiveLock'
 order by relname;`,
				Results: []sql.Row{{`lock_tbl1`}, {`lock_view6`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE OR REPLACE VIEW lock_view2 AS SELECT * from lock_view3;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_view2 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE VIEW lock_view7 AS SELECT * from lock_view2;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_view7 IN EXCLUSIVE MODE;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE TABLE lock_tbl2 (b BIGINT) INHERITS (lock_tbl1);`,
			},
			{
				Statement: `CREATE TABLE lock_tbl3 () INHERITS (lock_tbl2);`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 * IN ACCESS EXCLUSIVE MODE;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `GRANT UPDATE ON TABLE lock_tbl1 TO regress_rol_lock1;`,
			},
			{
				Statement: `SET ROLE regress_rol_lock1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_tbl2;`,
				ErrorString: `permission denied for table lock_tbl2`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_tbl1 * IN ACCESS EXCLUSIVE MODE;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE ONLY lock_tbl1;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `REVOKE UPDATE ON TABLE lock_tbl1 FROM regress_rol_lock1;`,
			},
			{
				Statement: `SET ROLE regress_rol_lock1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_view1;`,
				ErrorString: `permission denied for view lock_view1`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT UPDATE ON TABLE lock_view1 TO regress_rol_lock1;`,
			},
			{
				Statement: `SET ROLE regress_rol_lock1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_view1 IN ACCESS EXCLUSIVE MODE;`,
			},
			{
				Statement: `select relname from pg_locks l, pg_class c
 where l.relation = c.oid and relname like '%lock_%' and mode = 'AccessExclusiveLock'
 order by relname;`,
				Results: []sql.Row{{`lock_tbl1`}, {`lock_tbl2`}, {`lock_tbl3`}, {`lock_view1`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `REVOKE UPDATE ON TABLE lock_view1 FROM regress_rol_lock1;`,
			},
			{
				Statement: `CREATE VIEW lock_view8 WITH (security_invoker) AS SELECT * FROM lock_tbl1;`,
			},
			{
				Statement: `SET ROLE regress_rol_lock1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_view8;`,
				ErrorString: `permission denied for view lock_view8`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT UPDATE ON TABLE lock_view8 TO regress_rol_lock1;`,
			},
			{
				Statement: `SET ROLE regress_rol_lock1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `LOCK TABLE lock_view8;`,
				ErrorString: `permission denied for table lock_tbl1`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT UPDATE ON TABLE lock_tbl1 TO regress_rol_lock1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE lock_view8 IN ACCESS EXCLUSIVE MODE;`,
			},
			{
				Statement: `select relname from pg_locks l, pg_class c
 where l.relation = c.oid and relname like '%lock_%' and mode = 'AccessExclusiveLock'
 order by relname;`,
				Results: []sql.Row{{`lock_tbl1`}, {`lock_tbl2`}, {`lock_tbl3`}, {`lock_view8`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `REVOKE UPDATE ON TABLE lock_view8 FROM regress_rol_lock1;`,
			},
			{
				Statement: `DROP VIEW lock_view8;`,
			},
			{
				Statement: `DROP VIEW lock_view7;`,
			},
			{
				Statement: `DROP VIEW lock_view6;`,
			},
			{
				Statement: `DROP VIEW lock_view5;`,
			},
			{
				Statement: `DROP VIEW lock_view4;`,
			},
			{
				Statement: `DROP VIEW lock_view3 CASCADE;`,
			},
			{
				Statement: `DROP VIEW lock_view1;`,
			},
			{
				Statement: `DROP TABLE lock_tbl3;`,
			},
			{
				Statement: `DROP TABLE lock_tbl2;`,
			},
			{
				Statement: `DROP TABLE lock_tbl1;`,
			},
			{
				Statement: `DROP TABLE lock_tbl1a;`,
			},
			{
				Statement: `DROP SCHEMA lock_schema1 CASCADE;`,
			},
			{
				Statement: `DROP ROLE regress_rol_lock1;`,
			},
			{
				Statement: `RESET search_path;`,
			},
			{
				Statement: `CREATE FUNCTION test_atomic_ops()
    RETURNS bool
    AS :'regresslib'
    LANGUAGE C;`,
			},
			{
				Statement: `SELECT test_atomic_ops();`,
				Results:   []sql.Row{{true}},
			},
		},
	})
}
