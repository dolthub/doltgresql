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

func TestMerge(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_merge)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_merge,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE USER regress_merge_privs;`,
			},
			{
				Statement: `CREATE USER regress_merge_no_privs;`,
			},
			{
				Statement: `DROP TABLE IF EXISTS target;`,
			},
			{
				Statement: `DROP TABLE IF EXISTS source;`,
			},
			{
				Statement: `CREATE TABLE target (tid integer, balance integer)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE source (sid integer, delta integer) -- no index
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `INSERT INTO target VALUES (1, 10);`,
			},
			{
				Statement: `INSERT INTO target VALUES (2, 20);`,
			},
			{
				Statement: `INSERT INTO target VALUES (3, 30);`,
			},
			{
				Statement: `SELECT t.ctid is not null as matched, t.*, s.* FROM source s FULL OUTER JOIN target t ON s.sid = t.tid ORDER BY t.tid, s.sid;`,
				Results:   []sql.Row{{true, 1, 10, ``, ``}, {true, 2, 20, ``, ``}, {true, 3, 30, ``, ``}},
			},
			{
				Statement: `ALTER TABLE target OWNER TO regress_merge_privs;`,
			},
			{
				Statement: `ALTER TABLE source OWNER TO regress_merge_privs;`,
			},
			{
				Statement: `CREATE TABLE target2 (tid integer, balance integer)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE source2 (sid integer, delta integer)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `ALTER TABLE target2 OWNER TO regress_merge_no_privs;`,
			},
			{
				Statement: `ALTER TABLE source2 OWNER TO regress_merge_no_privs;`,
			},
			{
				Statement: `GRANT INSERT ON target TO regress_merge_no_privs;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_merge_privs;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	DELETE;`,
				Results: []sql.Row{{`Merge on target t`}, {`->  Merge Join`}, {`Merge Cond: (t.tid = s.sid)`}, {`->  Sort`}, {`Sort Key: t.tid`}, {`->  Seq Scan on target t`}, {`->  Sort`}, {`Sort Key: s.sid`}, {`->  Seq Scan on source s`}},
			},
			{
				Statement: `MERGE INTO target t RANDOMWORD
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
				ErrorString: `syntax error at or near "RANDOMWORD"`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	INSERT DEFAULT VALUES;`,
				ErrorString: `syntax error at or near "INSERT"`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT INTO target DEFAULT VALUES;`,
				ErrorString: `syntax error at or near "INTO"`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT VALUES (1,1), (2,2);`,
				ErrorString: `syntax error at or near ","`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT SELECT (1, 1);`,
				ErrorString: `syntax error at or near "SELECT"`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	UPDATE SET balance = 0;`,
				ErrorString: `syntax error at or near "UPDATE"`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE target SET balance = 0;`,
				ErrorString: `syntax error at or near "target"`,
			},
			{
				Statement: `MERGE INTO target
USING target
ON tid = tid
WHEN MATCHED THEN DO NOTHING;`,
				ErrorString: `name "target" specified more than once`,
			},
			{
				Statement: `WITH foo AS (
  MERGE INTO target USING source ON (true)
  WHEN MATCHED THEN DELETE
) SELECT * FROM foo;`,
				ErrorString: `MERGE not supported in WITH query`,
			},
			{
				Statement: `COPY (
  MERGE INTO target USING source ON (true)
  WHEN MATCHED THEN DELETE
) TO stdout;`,
				ErrorString: `MERGE not supported in COPY`,
			},
			{
				Statement: `CREATE VIEW tv AS SELECT * FROM target;`,
			},
			{
				Statement: `MERGE INTO tv t
USING source s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT DEFAULT VALUES;`,
				ErrorString: `cannot execute MERGE on relation "tv"`,
			},
			{
				Statement: `DROP VIEW tv;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mv AS SELECT * FROM target;`,
			},
			{
				Statement: `MERGE INTO mv t
USING source s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT DEFAULT VALUES;`,
				ErrorString: `cannot execute MERGE on relation "mv"`,
			},
			{
				Statement: `DROP MATERIALIZED VIEW mv;`,
			},
			{
				Statement: `MERGE INTO target
USING source2
ON target.tid = source2.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
				ErrorString: `permission denied for table source2`,
			},
			{
				Statement: `GRANT INSERT ON target TO regress_merge_no_privs;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_merge_no_privs;`,
			},
			{
				Statement: `MERGE INTO target
USING source2
ON target.tid = source2.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
				ErrorString: `permission denied for table target`,
			},
			{
				Statement: `GRANT UPDATE ON target2 TO regress_merge_privs;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_merge_privs;`,
			},
			{
				Statement: `MERGE INTO target2
USING source
ON target2.tid = source.sid
WHEN MATCHED THEN
	DELETE;`,
				ErrorString: `permission denied for table target2`,
			},
			{
				Statement: `MERGE INTO target2
USING source
ON target2.tid = source.sid
WHEN NOT MATCHED THEN
	INSERT DEFAULT VALUES;`,
				ErrorString: `permission denied for table target2`,
			},
			{
				Statement: `MERGE INTO target t
USING (SELECT * FROM source WHERE t.tid > sid) s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT DEFAULT VALUES;`,
				ErrorString: `invalid reference to FROM-clause entry for table "t"`,
			},
			{
				Statement: `MERGE INTO target
USING source
ON target.tid = source.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	DELETE;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT DEFAULT VALUES;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `INSERT INTO source VALUES (4, 40);`,
			},
			{
				Statement: `SELECT * FROM source ORDER BY sid;`,
				Results:   []sql.Row{{4, 40}},
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}},
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	DO NOTHING;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	DELETE;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {``, ``}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `INSERT INTO target SELECT generate_series(1000,2500), 0;`,
			},
			{
				Statement: `ALTER TABLE target ADD PRIMARY KEY (tid);`,
			},
			{
				Statement: `ANALYZE target;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
				Results: []sql.Row{{`Merge on target t`}, {`->  Hash Join`}, {`Hash Cond: (s.sid = t.tid)`}, {`->  Seq Scan on source s`}, {`->  Hash`}, {`->  Seq Scan on target t`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	DELETE;`,
				Results: []sql.Row{{`Merge on target t`}, {`->  Hash Join`}, {`Hash Cond: (s.sid = t.tid)`}, {`->  Seq Scan on source s`}, {`->  Hash`}, {`->  Seq Scan on target t`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT VALUES (4, NULL);`,
				Results: []sql.Row{{`Merge on target t`}, {`->  Hash Left Join`}, {`Hash Cond: (s.sid = t.tid)`}, {`->  Seq Scan on source s`}, {`->  Hash`}, {`->  Seq Scan on target t`}},
			},
			{
				Statement: `DELETE FROM target WHERE tid > 100;`,
			},
			{
				Statement: `ANALYZE target;`,
			},
			{
				Statement: `INSERT INTO source VALUES (2, 5);`,
			},
			{
				Statement: `INSERT INTO source VALUES (3, 20);`,
			},
			{
				Statement: `SELECT * FROM source ORDER BY sid;`,
				Results:   []sql.Row{{2, 5}, {3, 20}, {4, 40}},
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 0}, {3, 0}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	DELETE;`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	DO NOTHING;`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT VALUES (4, NULL);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {4, ``}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `INSERT INTO source VALUES (2, 5);`,
			},
			{
				Statement: `SELECT * FROM source ORDER BY sid;`,
				Results:   []sql.Row{{2, 5}, {2, 5}, {3, 20}, {4, 40}},
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
				ErrorString: `MERGE command cannot affect row a second time`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	DELETE;`,
				ErrorString: `MERGE command cannot affect row a second time`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DELETE FROM source WHERE sid = 2;`,
			},
			{
				Statement: `INSERT INTO source VALUES (2, 5);`,
			},
			{
				Statement: `SELECT * FROM source ORDER BY sid;`,
				Results:   []sql.Row{{2, 5}, {3, 20}, {4, 40}},
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}},
			},
			{
				Statement: `INSERT INTO source VALUES (4, 40);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
  INSERT VALUES (4, NULL);`,
				ErrorString: `duplicate key value violates unique constraint "target_pkey"`,
			},
			{
				Statement:   `SELECT * FROM target ORDER BY tid;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DELETE FROM source WHERE sid = 4;`,
			},
			{
				Statement: `INSERT INTO source VALUES (4, 40);`,
			},
			{
				Statement: `SELECT * FROM source ORDER BY sid;`,
				Results:   []sql.Row{{2, 5}, {3, 20}, {4, 40}},
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}},
			},
			{
				Statement: `alter table target drop CONSTRAINT target_pkey;`,
			},
			{
				Statement: `alter table target alter column tid drop not null;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT VALUES (4, 4)
WHEN MATCHED THEN
	UPDATE SET balance = 0;`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 0}, {3, 0}, {4, 4}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = 0
WHEN NOT MATCHED THEN
	INSERT VALUES (4, 4);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 0}, {3, 0}, {4, 4}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = t.balance + s.delta;`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 25}, {3, 50}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {4, 40}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `INSERT INTO source VALUES (5, 50);`,
			},
			{
				Statement: `INSERT INTO source VALUES (5, 50);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
  INSERT VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {4, 40}, {5, 50}, {5, 50}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DELETE FROM source WHERE sid = 5;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT (tid, balance) VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {4, 40}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT (tid, balance) VALUES (t.tid, s.delta);`,
				ErrorString: `invalid reference to FROM-clause entry for table "t"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON (SELECT true)
WHEN NOT MATCHED THEN
	INSERT (tid, balance) VALUES (t.tid, s.delta);`,
				ErrorString: `invalid reference to FROM-clause entry for table "t"`,
			},
			{
				Statement:   `SELECT * FROM target ORDER BY tid;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = t.balance + s.delta
WHEN NOT MATCHED THEN
	INSERT VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 25}, {3, 50}, {4, 40}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED THEN /* Terminal WHEN clause for MATCHED */
	DELETE
WHEN MATCHED THEN
	UPDATE SET balance = t.balance - s.delta;`,
				ErrorString: `unreachable WHEN clause specified after unconditional WHEN clause`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE TABLE wq_target (tid integer not null, balance integer DEFAULT -1)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE wq_source (balance integer, sid integer)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `INSERT INTO wq_source (sid, balance) VALUES (1, 100);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT (tid) VALUES (s.sid);`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, -1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN NOT MATCHED AND FALSE THEN
	INSERT (tid) VALUES (s.sid);`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN NOT MATCHED AND s.balance <> 100 THEN
	INSERT (tid) VALUES (s.sid);`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN NOT MATCHED AND s.balance = 100 THEN
	INSERT (tid) VALUES (s.sid);`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, -1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN NOT MATCHED AND t.balance = 100 THEN
	INSERT (tid) VALUES (s.sid);`,
				ErrorString: `invalid reference to FROM-clause entry for table "t"`,
			},
			{
				Statement:   `SELECT * FROM wq_target;`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN NOT MATCHED AND s.balance = 100 THEN
	INSERT (tid) VALUES (s.sid);`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, -1}},
			},
			{
				Statement: `SELECT * FROM wq_source;`,
				Results:   []sql.Row{{100, 1}},
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND s.balance = 100 THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, 99}},
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND t.balance = 100 THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, 99}},
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND t.balance = 99 AND s.balance > 100 THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, 99}},
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND t.balance = 99 AND s.balance = 100 THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, 199}},
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND t.balance = 99 OR s.balance > 100 THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, 199}},
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND t.balance = 199 OR s.balance > 100 THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, 299}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON (t.tid = s.sid)
WHEN matched and t = s or t.tid = s.sid THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, 399}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND t.balance > (SELECT max(balance) FROM target) THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND t.xmin = t.xmax THEN
	UPDATE SET balance = t.balance + s.balance;`,
				ErrorString: `cannot use system column "xmin" in MERGE WHEN condition`,
			},
			{
				Statement: `MERGE INTO wq_target t
USING wq_source s ON t.tid = s.sid
WHEN MATCHED AND t.tableoid >= 0 THEN
	UPDATE SET balance = t.balance + s.balance;`,
			},
			{
				Statement: `SELECT * FROM wq_target;`,
				Results:   []sql.Row{{1, 499}},
			},
			{
				Statement: `DROP TABLE wq_target, wq_source;`,
			},
			{
				Statement: `create or replace function merge_trigfunc () returns trigger
language plpgsql as
$$
DECLARE
	line text;`,
			},
			{
				Statement: `BEGIN
	SELECT INTO line format('%s %s %s trigger%s',
		TG_WHEN, TG_OP, TG_LEVEL, CASE
		WHEN TG_OP = 'INSERT' AND TG_LEVEL = 'ROW'
			THEN format(' row: %s', NEW)
		WHEN TG_OP = 'UPDATE' AND TG_LEVEL = 'ROW'
			THEN format(' row: %s -> %s', OLD, NEW)
		WHEN TG_OP = 'DELETE' AND TG_LEVEL = 'ROW'
			THEN format(' row: %s', OLD)
		END);`,
			},
			{
				Statement: `	RAISE NOTICE '%', line;`,
			},
			{
				Statement: `	IF (TG_WHEN = 'BEFORE' AND TG_LEVEL = 'ROW') THEN
		IF (TG_OP = 'DELETE') THEN
			RETURN OLD;`,
			},
			{
				Statement: `		ELSE
			RETURN NEW;`,
			},
			{
				Statement: `		END IF;`,
			},
			{
				Statement: `	ELSE
		RETURN NULL;`,
			},
			{
				Statement: `	END IF;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER merge_bsi BEFORE INSERT ON target FOR EACH STATEMENT EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_bsu BEFORE UPDATE ON target FOR EACH STATEMENT EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_bsd BEFORE DELETE ON target FOR EACH STATEMENT EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_asi AFTER INSERT ON target FOR EACH STATEMENT EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_asu AFTER UPDATE ON target FOR EACH STATEMENT EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_asd AFTER DELETE ON target FOR EACH STATEMENT EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_bri BEFORE INSERT ON target FOR EACH ROW EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_bru BEFORE UPDATE ON target FOR EACH ROW EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_brd BEFORE DELETE ON target FOR EACH ROW EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_ari AFTER INSERT ON target FOR EACH ROW EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_aru AFTER UPDATE ON target FOR EACH ROW EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `CREATE TRIGGER merge_ard AFTER DELETE ON target FOR EACH ROW EXECUTE PROCEDURE merge_trigfunc ();`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE target SET balance = 0 WHERE tid = 3;`,
			},
			{
				Statement: `MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED AND t.balance > s.delta THEN
	UPDATE SET balance = t.balance - s.delta
WHEN MATCHED THEN
	DELETE
WHEN NOT MATCHED THEN
	INSERT VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 15}, {4, 40}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `create or replace function skip_merge_op() returns trigger
language plpgsql as
$$
BEGIN
	RETURN NULL;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT * FROM target full outer join source on (sid = tid);`,
				Results:   []sql.Row{{3, 30, 3, 20}, {2, 20, 2, 5}, {``, ``, 4, 40}, {1, 10, ``, ``}},
			},
			{
				Statement: `create trigger merge_skip BEFORE INSERT OR UPDATE or DELETE
  ON target FOR EACH ROW EXECUTE FUNCTION skip_merge_op();`,
			},
			{
				Statement: `DO $$
DECLARE
  result integer;`,
			},
			{
				Statement: `BEGIN
MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED AND s.sid = 3 THEN UPDATE SET balance = t.balance + s.delta
WHEN MATCHED THEN DELETE
WHEN NOT MATCHED THEN INSERT VALUES (sid, delta);`,
			},
			{
				Statement: `IF FOUND THEN
  RAISE NOTICE 'Found';`,
			},
			{
				Statement: `ELSE
  RAISE NOTICE 'Not found';`,
			},
			{
				Statement: `END IF;`,
			},
			{
				Statement: `GET DIAGNOSTICS result := ROW_COUNT;`,
			},
			{
				Statement: `RAISE NOTICE 'ROW_COUNT = %', result;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT * FROM target FULL OUTER JOIN source ON (sid = tid);`,
				Results:   []sql.Row{{3, 30, 3, 20}, {2, 20, 2, 5}, {``, ``, 4, 40}, {1, 10, ``, ``}},
			},
			{
				Statement: `DROP TRIGGER merge_skip ON target;`,
			},
			{
				Statement: `DROP FUNCTION skip_merge_op();`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DO LANGUAGE plpgsql $$
BEGIN
MERGE INTO target t
USING source AS s
ON t.tid = s.sid
WHEN MATCHED AND t.balance > s.delta THEN
	UPDATE SET balance = t.balance - s.delta;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING (SELECT 9 AS sid, 57 AS delta) AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT (tid, balance) VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {9, 57}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING (SELECT sid, delta FROM source WHERE delta > 0) AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT (tid, balance) VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {4, 40}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING (SELECT sid, delta as newname FROM source WHERE delta > 0) AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT (tid, balance) VALUES (s.sid, s.newname);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {4, 40}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t1
USING target t2
ON t1.tid = t2.tid
WHEN MATCHED THEN
	UPDATE SET balance = t1.balance + t2.balance
WHEN NOT MATCHED THEN
	INSERT VALUES (t2.tid, t2.balance);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 20}, {2, 40}, {3, 60}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING (SELECT tid as sid, balance as delta FROM target WHERE balance > 0) AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT (tid, balance) VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO target t
USING
(SELECT sid, max(delta) AS delta
 FROM source
 GROUP BY sid
 HAVING count(*) = 1
 ORDER BY sid ASC) AS s
ON t.tid = s.sid
WHEN NOT MATCHED THEN
	INSERT (tid, balance) VALUES (s.sid, s.delta);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 30}, {4, 40}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE FUNCTION merge_func (p_id integer, p_bal integer)
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
 result integer;`,
			},
			{
				Statement: `BEGIN
MERGE INTO target t
USING (SELECT p_id AS sid) AS s
ON t.tid = s.sid
WHEN MATCHED THEN
	UPDATE SET balance = t.balance - p_bal;`,
			},
			{
				Statement: `IF FOUND THEN
	GET DIAGNOSTICS result := ROW_COUNT;`,
			},
			{
				Statement: `END IF;`,
			},
			{
				Statement: `RETURN result;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT merge_func(3, 4);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 10}, {2, 20}, {3, 26}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `prepare foom as merge into target t using (select 1 as sid) s on (t.tid = s.sid) when matched then update set balance = 1;`,
			},
			{
				Statement: `execute foom;`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 1}, {2, 20}, {3, 30}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `PREPARE foom2 (integer, integer) AS
MERGE INTO target t
USING (SELECT 1) s
ON t.tid = $1
WHEN MATCHED THEN
UPDATE SET balance = $2;`,
			},
			{
				Statement: `execute foom2 (1, 1);`,
			},
			{
				Statement: `SELECT * FROM target ORDER BY tid;`,
				Results:   []sql.Row{{1, 1}, {2, 20}, {3, 30}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE TABLE sq_target (tid integer NOT NULL, balance integer)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE sq_source (delta integer, sid integer, balance integer DEFAULT 0)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `INSERT INTO sq_target(tid, balance) VALUES (1,100), (2,200), (3,300);`,
			},
			{
				Statement: `INSERT INTO sq_source(sid, delta) VALUES (1,10), (2,20), (4,40);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO sq_target t
USING (SELECT * FROM sq_source) s
ON tid = sid
WHEN MATCHED AND t.balance > delta THEN
	UPDATE SET balance = t.balance + delta;`,
			},
			{
				Statement: `SELECT * FROM sq_target;`,
				Results:   []sql.Row{{3, 300}, {1, 110}, {2, 220}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE VIEW v AS SELECT * FROM sq_source WHERE sid < 2;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO sq_target
USING v
ON tid = sid
WHEN MATCHED THEN
    UPDATE SET balance = v.balance + delta;`,
			},
			{
				Statement: `SELECT * FROM sq_target;`,
				Results:   []sql.Row{{2, 200}, {3, 300}, {1, 10}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO sq_target
USING v
ON tid = sid
WHEN MATCHED AND tid > 2 THEN
    UPDATE SET balance = balance + delta
WHEN NOT MATCHED THEN
	INSERT (balance, tid) VALUES (balance + delta, sid)
WHEN MATCHED AND tid < 2 THEN
	DELETE;`,
				ErrorString: `column reference "balance" is ambiguous`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO sq_source (sid, balance, delta) VALUES (-1, -1, -10);`,
			},
			{
				Statement: `MERGE INTO sq_target t
USING v
ON tid = sid
WHEN MATCHED AND tid > 2 THEN
    UPDATE SET balance = t.balance + delta
WHEN NOT MATCHED THEN
	INSERT (balance, tid) VALUES (balance + delta, sid)
WHEN MATCHED AND tid < 2 THEN
	DELETE;`,
			},
			{
				Statement: `SELECT * FROM sq_target;`,
				Results:   []sql.Row{{2, 200}, {3, 300}, {-1, -11}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO sq_source (sid, balance, delta) VALUES (-1, -1, -10);`,
			},
			{
				Statement: `WITH targq AS (
	SELECT * FROM v
)
MERGE INTO sq_target t
USING v
ON tid = sid
WHEN MATCHED AND tid > 2 THEN
    UPDATE SET balance = t.balance + delta
WHEN NOT MATCHED THEN
	INSERT (balance, tid) VALUES (balance + delta, sid)
WHEN MATCHED AND tid < 2 THEN
	DELETE;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO sq_source (sid, balance, delta) VALUES (-1, -1, -10);`,
			},
			{
				Statement: `MERGE INTO sq_target t
USING v
ON tid = sid
WHEN MATCHED AND tid > 2 THEN
    UPDATE SET balance = t.balance + delta
WHEN NOT MATCHED THEN
	INSERT (balance, tid) VALUES (balance + delta, sid)
WHEN MATCHED AND tid < 2 THEN
	DELETE
RETURNING *;`,
				ErrorString: `syntax error at or near "RETURNING"`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE TABLE ex_mtarget (a int, b int)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE ex_msource (a int, b int)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `INSERT INTO ex_mtarget SELECT i, i*10 FROM generate_series(1,100,2) i;`,
			},
			{
				Statement: `INSERT INTO ex_msource SELECT i, i*10 FROM generate_series(1,100,1) i;`,
			},
			{
				Statement: `CREATE FUNCTION explain_merge(query text) RETURNS SETOF text
LANGUAGE plpgsql AS
$$
DECLARE ln text;`,
			},
			{
				Statement: `BEGIN
    FOR ln IN
        EXECUTE 'explain (analyze, timing off, summary off, costs off) ' ||
		  query
    LOOP
        ln := regexp_replace(ln, '(Memory( Usage)?|Buckets|Batches): \S*',  '\1: xxx', 'g');`,
			},
			{
				Statement: `        RETURN NEXT ln;`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT explain_merge('
MERGE INTO ex_mtarget t USING ex_msource s ON t.a = s.a
WHEN MATCHED THEN
	UPDATE SET b = t.b + 1');`,
				Results: []sql.Row{{`Merge on ex_mtarget t (actual rows=0 loops=1)`}, {`Tuples: updated=50`}, {`->  Merge Join (actual rows=50 loops=1)`}, {`Merge Cond: (t.a = s.a)`}, {`->  Sort (actual rows=50 loops=1)`}, {`Sort Key: t.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_mtarget t (actual rows=50 loops=1)`}, {`->  Sort (actual rows=100 loops=1)`}, {`Sort Key: s.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_msource s (actual rows=100 loops=1)`}},
			},
			{
				Statement: `SELECT explain_merge('
MERGE INTO ex_mtarget t USING ex_msource s ON t.a = s.a
WHEN MATCHED AND t.a < 10 THEN
	UPDATE SET b = t.b + 1');`,
				Results: []sql.Row{{`Merge on ex_mtarget t (actual rows=0 loops=1)`}, {`Tuples: updated=5 skipped=45`}, {`->  Merge Join (actual rows=50 loops=1)`}, {`Merge Cond: (t.a = s.a)`}, {`->  Sort (actual rows=50 loops=1)`}, {`Sort Key: t.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_mtarget t (actual rows=50 loops=1)`}, {`->  Sort (actual rows=100 loops=1)`}, {`Sort Key: s.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_msource s (actual rows=100 loops=1)`}},
			},
			{
				Statement: `SELECT explain_merge('
MERGE INTO ex_mtarget t USING ex_msource s ON t.a = s.a
WHEN MATCHED AND t.a < 10 THEN
	UPDATE SET b = t.b + 1
WHEN MATCHED AND t.a >= 10 AND t.a <= 20 THEN
	DELETE');`,
				Results: []sql.Row{{`Merge on ex_mtarget t (actual rows=0 loops=1)`}, {`Tuples: updated=5 deleted=5 skipped=40`}, {`->  Merge Join (actual rows=50 loops=1)`}, {`Merge Cond: (t.a = s.a)`}, {`->  Sort (actual rows=50 loops=1)`}, {`Sort Key: t.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_mtarget t (actual rows=50 loops=1)`}, {`->  Sort (actual rows=100 loops=1)`}, {`Sort Key: s.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_msource s (actual rows=100 loops=1)`}},
			},
			{
				Statement: `SELECT explain_merge('
MERGE INTO ex_mtarget t USING ex_msource s ON t.a = s.a
WHEN NOT MATCHED AND s.a < 10 THEN
	INSERT VALUES (a, b)');`,
				Results: []sql.Row{{`Merge on ex_mtarget t (actual rows=0 loops=1)`}, {`Tuples: inserted=4 skipped=96`}, {`->  Merge Left Join (actual rows=100 loops=1)`}, {`Merge Cond: (s.a = t.a)`}, {`->  Sort (actual rows=100 loops=1)`}, {`Sort Key: s.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_msource s (actual rows=100 loops=1)`}, {`->  Sort (actual rows=45 loops=1)`}, {`Sort Key: t.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_mtarget t (actual rows=45 loops=1)`}},
			},
			{
				Statement: `SELECT explain_merge('
MERGE INTO ex_mtarget t USING ex_msource s ON t.a = s.a
WHEN MATCHED AND t.a < 10 THEN
	UPDATE SET b = t.b + 1
WHEN MATCHED AND t.a >= 30 AND t.a <= 40 THEN
	DELETE
WHEN NOT MATCHED AND s.a < 20 THEN
	INSERT VALUES (a, b)');`,
				Results: []sql.Row{{`Merge on ex_mtarget t (actual rows=0 loops=1)`}, {`Tuples: inserted=10 updated=9 deleted=5 skipped=76`}, {`->  Merge Left Join (actual rows=100 loops=1)`}, {`Merge Cond: (s.a = t.a)`}, {`->  Sort (actual rows=100 loops=1)`}, {`Sort Key: s.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_msource s (actual rows=100 loops=1)`}, {`->  Sort (actual rows=49 loops=1)`}, {`Sort Key: t.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_mtarget t (actual rows=49 loops=1)`}},
			},
			{
				Statement: `SELECT explain_merge('
MERGE INTO ex_mtarget t USING ex_msource s ON t.a = s.a AND t.a < -1000
WHEN MATCHED AND t.a < 10 THEN
	DO NOTHING');`,
				Results: []sql.Row{{`Merge on ex_mtarget t (actual rows=0 loops=1)`}, {`->  Merge Join (actual rows=0 loops=1)`}, {`Merge Cond: (t.a = s.a)`}, {`->  Sort (actual rows=0 loops=1)`}, {`Sort Key: t.a`}, {`Sort Method: quicksort  Memory: xxx`}, {`->  Seq Scan on ex_mtarget t (actual rows=0 loops=1)`}, {`Filter: (a < '-1000'::integer)`}, {`Rows Removed by Filter: 54`}, {`->  Sort (never executed)`}, {`Sort Key: s.a`}, {`->  Seq Scan on ex_msource s (never executed)`}},
			},
			{
				Statement: `DROP TABLE ex_msource, ex_mtarget;`,
			},
			{
				Statement: `DROP FUNCTION explain_merge(text);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO sq_target t
USING v
ON tid = sid
WHEN MATCHED THEN
    UPDATE SET balance = (SELECT count(*) FROM sq_target);`,
			},
			{
				Statement: `SELECT * FROM sq_target WHERE tid = 1;`,
				Results:   []sql.Row{{1, 3}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO sq_target t
USING v
ON tid = sid
WHEN MATCHED AND (SELECT count(*) > 0 FROM sq_target) THEN
    UPDATE SET balance = 42;`,
			},
			{
				Statement: `SELECT * FROM sq_target WHERE tid = 1;`,
				Results:   []sql.Row{{1, 42}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO sq_target t
USING v
ON tid = sid AND (SELECT count(*) > 0 FROM sq_target)
WHEN MATCHED THEN
    UPDATE SET balance = 42;`,
			},
			{
				Statement: `SELECT * FROM sq_target WHERE tid = 1;`,
				Results:   []sql.Row{{1, 42}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE sq_target, sq_source CASCADE;`,
			},
			{
				Statement: `CREATE TABLE pa_target (tid integer, balance float, val text)
	PARTITION BY LIST (tid);`,
			},
			{
				Statement: `CREATE TABLE part1 PARTITION OF pa_target FOR VALUES IN (1,4)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part2 PARTITION OF pa_target FOR VALUES IN (2,5,6)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part3 PARTITION OF pa_target FOR VALUES IN (3,8,9)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part4 PARTITION OF pa_target DEFAULT
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE pa_source (sid integer, delta float);`,
			},
			{
				Statement: `INSERT INTO pa_source SELECT id, id * 10  FROM generate_series(1,14) AS id;`,
			},
			{
				Statement: `INSERT INTO pa_target SELECT id, id * 100, 'initial' FROM generate_series(1,14,2) AS id;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO pa_target t
  USING pa_source s
  ON t.tid = s.sid
  WHEN MATCHED THEN
    UPDATE SET balance = balance + delta, val = val || ' updated by merge'
  WHEN NOT MATCHED THEN
    INSERT VALUES (sid, delta, 'inserted by merge');`,
			},
			{
				Statement: `SELECT * FROM pa_target ORDER BY tid;`,
				Results:   []sql.Row{{1, 110, `initial updated by merge`}, {2, 20, `inserted by merge`}, {3, 330, `initial updated by merge`}, {4, 40, `inserted by merge`}, {5, 550, `initial updated by merge`}, {6, 60, `inserted by merge`}, {7, 770, `initial updated by merge`}, {8, 80, `inserted by merge`}, {9, 990, `initial updated by merge`}, {10, 100, `inserted by merge`}, {11, 1210, `initial updated by merge`}, {12, 120, `inserted by merge`}, {13, 1430, `initial updated by merge`}, {14, 140, `inserted by merge`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO pa_target t
  USING pa_source s
  ON t.tid = s.sid AND tid = 1
  WHEN MATCHED THEN
    UPDATE SET balance = balance + delta, val = val || ' updated by merge'
  WHEN NOT MATCHED THEN
    INSERT VALUES (sid, delta, 'inserted by merge');`,
			},
			{
				Statement: `SELECT * FROM pa_target ORDER BY tid;`,
				Results:   []sql.Row{{1, 110, `initial updated by merge`}, {2, 20, `inserted by merge`}, {3, 30, `inserted by merge`}, {3, 300, `initial`}, {4, 40, `inserted by merge`}, {5, 500, `initial`}, {5, 50, `inserted by merge`}, {6, 60, `inserted by merge`}, {7, 700, `initial`}, {7, 70, `inserted by merge`}, {8, 80, `inserted by merge`}, {9, 90, `inserted by merge`}, {9, 900, `initial`}, {10, 100, `inserted by merge`}, {11, 1100, `initial`}, {11, 110, `inserted by merge`}, {12, 120, `inserted by merge`}, {13, 1300, `initial`}, {13, 130, `inserted by merge`}, {14, 140, `inserted by merge`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE FUNCTION merge_func() RETURNS integer LANGUAGE plpgsql AS $$
DECLARE
  result integer;`,
			},
			{
				Statement: `BEGIN
MERGE INTO pa_target t
  USING pa_source s
  ON t.tid = s.sid
  WHEN MATCHED THEN
    UPDATE SET tid = tid + 1, balance = balance + delta, val = val || ' updated by merge'
  WHEN NOT MATCHED THEN
    INSERT VALUES (sid, delta, 'inserted by merge');`,
			},
			{
				Statement: `IF FOUND THEN
  GET DIAGNOSTICS result := ROW_COUNT;`,
			},
			{
				Statement: `END IF;`,
			},
			{
				Statement: `RETURN result;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT merge_func();`,
				Results:   []sql.Row{{14}},
			},
			{
				Statement: `SELECT * FROM pa_target ORDER BY tid;`,
				Results:   []sql.Row{{2, 110, `initial updated by merge`}, {2, 20, `inserted by merge`}, {4, 40, `inserted by merge`}, {4, 330, `initial updated by merge`}, {6, 550, `initial updated by merge`}, {6, 60, `inserted by merge`}, {8, 80, `inserted by merge`}, {8, 770, `initial updated by merge`}, {10, 990, `initial updated by merge`}, {10, 100, `inserted by merge`}, {12, 1210, `initial updated by merge`}, {12, 120, `inserted by merge`}, {14, 1430, `initial updated by merge`}, {14, 140, `inserted by merge`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE pa_target CASCADE;`,
			},
			{
				Statement: `CREATE TABLE pa_target (tid integer, balance float, val text)
	PARTITION BY LIST (tid);`,
			},
			{
				Statement: `CREATE TABLE part1 (tid integer, balance float, val text)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part2 (balance float, tid integer, val text)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part3 (tid integer, balance float, val text)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part4 (extraid text, tid integer, balance float, val text)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `ALTER TABLE part4 DROP COLUMN extraid;`,
			},
			{
				Statement: `ALTER TABLE pa_target ATTACH PARTITION part1 FOR VALUES IN (1,4);`,
			},
			{
				Statement: `ALTER TABLE pa_target ATTACH PARTITION part2 FOR VALUES IN (2,5,6);`,
			},
			{
				Statement: `ALTER TABLE pa_target ATTACH PARTITION part3 FOR VALUES IN (3,8,9);`,
			},
			{
				Statement: `ALTER TABLE pa_target ATTACH PARTITION part4 DEFAULT;`,
			},
			{
				Statement: `INSERT INTO pa_target SELECT id, id * 100, 'initial' FROM generate_series(1,14,2) AS id;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO pa_target t
  USING pa_source s
  ON t.tid = s.sid
  WHEN MATCHED THEN
    UPDATE SET balance = balance + delta, val = val || ' updated by merge'
  WHEN NOT MATCHED THEN
    INSERT VALUES (sid, delta, 'inserted by merge');`,
			},
			{
				Statement: `SELECT * FROM pa_target ORDER BY tid;`,
				Results:   []sql.Row{{1, 110, `initial updated by merge`}, {2, 20, `inserted by merge`}, {3, 330, `initial updated by merge`}, {4, 40, `inserted by merge`}, {5, 550, `initial updated by merge`}, {6, 60, `inserted by merge`}, {7, 770, `initial updated by merge`}, {8, 80, `inserted by merge`}, {9, 990, `initial updated by merge`}, {10, 100, `inserted by merge`}, {11, 1210, `initial updated by merge`}, {12, 120, `inserted by merge`}, {13, 1430, `initial updated by merge`}, {14, 140, `inserted by merge`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO pa_target t
  USING pa_source s
  ON t.tid = s.sid AND tid IN (1, 5)
  WHEN MATCHED AND tid % 5 = 0 THEN DELETE
  WHEN MATCHED THEN
    UPDATE SET balance = balance + delta, val = val || ' updated by merge'
  WHEN NOT MATCHED THEN
    INSERT VALUES (sid, delta, 'inserted by merge');`,
			},
			{
				Statement: `SELECT * FROM pa_target ORDER BY tid;`,
				Results:   []sql.Row{{1, 110, `initial updated by merge`}, {2, 20, `inserted by merge`}, {3, 30, `inserted by merge`}, {3, 300, `initial`}, {4, 40, `inserted by merge`}, {6, 60, `inserted by merge`}, {7, 700, `initial`}, {7, 70, `inserted by merge`}, {8, 80, `inserted by merge`}, {9, 900, `initial`}, {9, 90, `inserted by merge`}, {10, 100, `inserted by merge`}, {11, 110, `inserted by merge`}, {11, 1100, `initial`}, {12, 120, `inserted by merge`}, {13, 1300, `initial`}, {13, 130, `inserted by merge`}, {14, 140, `inserted by merge`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO pa_target t
  USING pa_source s
  ON t.tid = s.sid
  WHEN MATCHED THEN
    UPDATE SET tid = tid + 1, balance = balance + delta, val = val || ' updated by merge'
  WHEN NOT MATCHED THEN
    INSERT VALUES (sid, delta, 'inserted by merge');`,
			},
			{
				Statement: `SELECT * FROM pa_target ORDER BY tid;`,
				Results:   []sql.Row{{2, 110, `initial updated by merge`}, {2, 20, `inserted by merge`}, {4, 40, `inserted by merge`}, {4, 330, `initial updated by merge`}, {6, 550, `initial updated by merge`}, {6, 60, `inserted by merge`}, {8, 80, `inserted by merge`}, {8, 770, `initial updated by merge`}, {10, 990, `initial updated by merge`}, {10, 100, `inserted by merge`}, {12, 1210, `initial updated by merge`}, {12, 120, `inserted by merge`}, {14, 1430, `initial updated by merge`}, {14, 140, `inserted by merge`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER TABLE pa_target ENABLE ROW LEVEL SECURITY;`,
			},
			{
				Statement: `ALTER TABLE pa_target FORCE ROW LEVEL SECURITY;`,
			},
			{
				Statement: `CREATE POLICY pa_target_pol ON pa_target USING (tid != 0);`,
			},
			{
				Statement: `MERGE INTO pa_target t
  USING pa_source s
  ON t.tid = s.sid AND t.tid IN (1,2,3,4)
  WHEN MATCHED THEN
    UPDATE SET tid = tid - 1;`,
				ErrorString: `new row violates row-level security policy for table "pa_target"`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE pa_source;`,
			},
			{
				Statement: `DROP TABLE pa_target CASCADE;`,
			},
			{
				Statement: `CREATE TABLE pa_target (logts timestamp, tid integer, balance float, val text)
	PARTITION BY RANGE (logts);`,
			},
			{
				Statement: `CREATE TABLE part_m01 PARTITION OF pa_target
	FOR VALUES FROM ('2017-01-01') TO ('2017-02-01')
	PARTITION BY LIST (tid);`,
			},
			{
				Statement: `CREATE TABLE part_m01_odd PARTITION OF part_m01
	FOR VALUES IN (1,3,5,7,9) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part_m01_even PARTITION OF part_m01
	FOR VALUES IN (2,4,6,8) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part_m02 PARTITION OF pa_target
	FOR VALUES FROM ('2017-02-01') TO ('2017-03-01')
	PARTITION BY LIST (tid);`,
			},
			{
				Statement: `CREATE TABLE part_m02_odd PARTITION OF part_m02
	FOR VALUES IN (1,3,5,7,9) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE part_m02_even PARTITION OF part_m02
	FOR VALUES IN (2,4,6,8) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE pa_source (sid integer, delta float)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `INSERT INTO pa_source SELECT id, id * 10  FROM generate_series(1,14) AS id;`,
			},
			{
				Statement: `INSERT INTO pa_target SELECT '2017-01-31', id, id * 100, 'initial' FROM generate_series(1,9,3) AS id;`,
			},
			{
				Statement: `INSERT INTO pa_target SELECT '2017-02-28', id, id * 100, 'initial' FROM generate_series(2,9,3) AS id;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO pa_target t
  USING (SELECT '2017-01-15' AS slogts, * FROM pa_source WHERE sid < 10) s
  ON t.tid = s.sid
  WHEN MATCHED THEN
    UPDATE SET balance = balance + delta, val = val || ' updated by merge'
  WHEN NOT MATCHED THEN
    INSERT VALUES (slogts::timestamp, sid, delta, 'inserted by merge');`,
			},
			{
				Statement: `SELECT * FROM pa_target ORDER BY tid;`,
				Results:   []sql.Row{{`Tue Jan 31 00:00:00 2017`, 1, 110, `initial updated by merge`}, {`Tue Feb 28 00:00:00 2017`, 2, 220, `initial updated by merge`}, {`Sun Jan 15 00:00:00 2017`, 3, 30, `inserted by merge`}, {`Tue Jan 31 00:00:00 2017`, 4, 440, `initial updated by merge`}, {`Tue Feb 28 00:00:00 2017`, 5, 550, `initial updated by merge`}, {`Sun Jan 15 00:00:00 2017`, 6, 60, `inserted by merge`}, {`Tue Jan 31 00:00:00 2017`, 7, 770, `initial updated by merge`}, {`Tue Feb 28 00:00:00 2017`, 8, 880, `initial updated by merge`}, {`Sun Jan 15 00:00:00 2017`, 9, 90, `inserted by merge`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE pa_source;`,
			},
			{
				Statement: `DROP TABLE pa_target CASCADE;`,
			},
			{
				Statement: `CREATE TABLE pa_target (tid integer PRIMARY KEY) PARTITION BY LIST (tid);`,
			},
			{
				Statement: `CREATE TABLE pa_targetp PARTITION OF pa_target DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE pa_source (sid integer);`,
			},
			{
				Statement: `INSERT INTO pa_source VALUES (1), (2);`,
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
MERGE INTO pa_target t USING pa_source s ON t.tid = s.sid
  WHEN NOT MATCHED THEN INSERT VALUES (s.sid);`,
				Results: []sql.Row{{`Merge on public.pa_target t`}, {`Merge on public.pa_targetp t_1`}, {`->  Nested Loop Left Join`}, {`Output: s.sid, t_1.tableoid, t_1.ctid`}, {`->  Seq Scan on public.pa_source s`}, {`Output: s.sid`}, {`->  Index Scan using pa_targetp_pkey on public.pa_targetp t_1`}, {`Output: t_1.tid, t_1.tableoid, t_1.ctid`}, {`Index Cond: (t_1.tid = s.sid)`}},
			},
			{
				Statement: `MERGE INTO pa_target t USING pa_source s ON t.tid = s.sid
  WHEN NOT MATCHED THEN INSERT VALUES (s.sid);`,
			},
			{
				Statement: `TABLE pa_target;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `DROP TABLE pa_targetp;`,
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
MERGE INTO pa_target t USING pa_source s ON t.tid = s.sid
  WHEN NOT MATCHED THEN INSERT VALUES (s.sid);`,
				Results: []sql.Row{{`Merge on public.pa_target t`}, {`->  Hash Left Join`}, {`Output: s.sid, t.ctid`}, {`Hash Cond: (s.sid = t.tid)`}, {`->  Seq Scan on public.pa_source s`}, {`Output: s.sid`}, {`->  Hash`}, {`Output: t.tid, t.ctid`}, {`->  Result`}, {`Output: t.tid, t.ctid`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `MERGE INTO pa_target t USING pa_source s ON t.tid = s.sid
  WHEN NOT MATCHED THEN INSERT VALUES (s.sid);`,
				ErrorString: `no partition of relation "pa_target" found for row`,
			},
			{
				Statement: `DROP TABLE pa_source;`,
			},
			{
				Statement: `DROP TABLE pa_target CASCADE;`,
			},
			{
				Statement: `CREATE TABLE cj_target (tid integer, balance float, val text)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE cj_source1 (sid1 integer, scat integer, delta integer)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE cj_source2 (sid2 integer, sval text)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `INSERT INTO cj_source1 VALUES (1, 10, 100);`,
			},
			{
				Statement: `INSERT INTO cj_source1 VALUES (1, 20, 200);`,
			},
			{
				Statement: `INSERT INTO cj_source1 VALUES (2, 20, 300);`,
			},
			{
				Statement: `INSERT INTO cj_source1 VALUES (3, 10, 400);`,
			},
			{
				Statement: `INSERT INTO cj_source2 VALUES (1, 'initial source2');`,
			},
			{
				Statement: `INSERT INTO cj_source2 VALUES (2, 'initial source2');`,
			},
			{
				Statement: `INSERT INTO cj_source2 VALUES (3, 'initial source2');`,
			},
			{
				Statement: `MERGE INTO cj_target t
USING cj_source1 s1
	INNER JOIN cj_source2 s2 ON sid1 = sid2
ON t.tid = sid1
WHEN NOT MATCHED THEN
	INSERT VALUES (sid1, delta, sval);`,
			},
			{
				Statement: `MERGE INTO cj_target t
USING cj_source2 s2
	INNER JOIN cj_source1 s1 ON sid1 = sid2 AND scat = 20
ON t.tid = sid1
WHEN NOT MATCHED THEN
	INSERT VALUES (sid2, delta, sval)
WHEN MATCHED THEN
	DELETE;`,
			},
			{
				Statement: `MERGE INTO cj_target t
USING cj_source2 s2
	INNER JOIN cj_source1 s1 ON sid1 = sid2
ON t.tid = sid1
WHEN NOT MATCHED THEN
	INSERT VALUES (sid2, delta + scat, sval)
WHEN MATCHED THEN
	UPDATE SET val = val || ' updated by merge';`,
			},
			{
				Statement: `MERGE INTO cj_target t
USING cj_source2 s2
	INNER JOIN cj_source1 s1 ON sid1 = sid2 AND scat = 20
ON t.tid = sid1
WHEN MATCHED THEN
	UPDATE SET val = val || ' ' || delta::text;`,
			},
			{
				Statement: `SELECT * FROM cj_target;`,
				Results:   []sql.Row{{3, 400, `initial source2 updated by merge`}, {1, 220, `initial source2 200`}, {1, 110, `initial source2 200`}, {2, 320, `initial source2 300`}},
			},
			{
				Statement: `MERGE INTO cj_target t
USING (SELECT *, 'join input'::text AS phv FROM cj_source1) fj
	FULL JOIN cj_source2 fj2 ON fj.scat = fj2.sid2 * 10
ON t.tid = fj.scat
WHEN NOT MATCHED THEN
	INSERT (tid, balance, val) VALUES (fj.scat, fj.delta, fj.phv);`,
			},
			{
				Statement: `SELECT * FROM cj_target;`,
				Results:   []sql.Row{{3, 400, `initial source2 updated by merge`}, {1, 220, `initial source2 200`}, {1, 110, `initial source2 200`}, {2, 320, `initial source2 300`}, {10, 100, `join input`}, {10, 400, `join input`}, {20, 200, `join input`}, {20, 300, `join input`}, {``, ``, ``}},
			},
			{
				Statement: `ALTER TABLE cj_source1 RENAME COLUMN sid1 TO sid;`,
			},
			{
				Statement: `ALTER TABLE cj_source2 RENAME COLUMN sid2 TO sid;`,
			},
			{
				Statement: `TRUNCATE cj_target;`,
			},
			{
				Statement: `MERGE INTO cj_target t
USING cj_source1 s1
	INNER JOIN cj_source2 s2 ON s1.sid = s2.sid
ON t.tid = s1.sid
WHEN NOT MATCHED THEN
	INSERT VALUES (s2.sid, delta, sval);`,
			},
			{
				Statement: `DROP TABLE cj_source2, cj_source1, cj_target;`,
			},
			{
				Statement: `CREATE TABLE fs_target (a int, b int, c text)
  WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `MERGE INTO fs_target t
USING generate_series(1,100,1) AS id
ON t.a = id
WHEN MATCHED THEN
	UPDATE SET b = b + id
WHEN NOT MATCHED THEN
	INSERT VALUES (id, -1);`,
			},
			{
				Statement: `MERGE INTO fs_target t
USING generate_series(1,100,2) AS id
ON t.a = id
WHEN MATCHED THEN
	UPDATE SET b = b + id, c = 'updated '|| id.*::text
WHEN NOT MATCHED THEN
	INSERT VALUES (id, -1, 'inserted ' || id.*::text);`,
			},
			{
				Statement: `SELECT count(*) FROM fs_target;`,
				Results:   []sql.Row{{100}},
			},
			{
				Statement: `DROP TABLE fs_target;`,
			},
			{
				Statement: `CREATE TABLE measurement (
    city_id         int not null,
    logdate         date not null,
    peaktemp        int,
    unitsales       int
) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE measurement_y2006m02 (
    CHECK ( logdate >= DATE '2006-02-01' AND logdate < DATE '2006-03-01' )
) INHERITS (measurement) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE measurement_y2006m03 (
    CHECK ( logdate >= DATE '2006-03-01' AND logdate < DATE '2006-04-01' )
) INHERITS (measurement) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `CREATE TABLE measurement_y2007m01 (
    filler          text,
    peaktemp        int,
    logdate         date not null,
    city_id         int not null,
    unitsales       int
    CHECK ( logdate >= DATE '2007-01-01' AND logdate < DATE '2007-02-01')
) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `ALTER TABLE measurement_y2007m01 DROP COLUMN filler;`,
			},
			{
				Statement: `ALTER TABLE measurement_y2007m01 INHERIT measurement;`,
			},
			{
				Statement: `INSERT INTO measurement VALUES (0, '2005-07-21', 5, 15);`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION measurement_insert_trigger()
RETURNS TRIGGER AS $$
BEGIN
    IF ( NEW.logdate >= DATE '2006-02-01' AND
         NEW.logdate < DATE '2006-03-01' ) THEN
        INSERT INTO measurement_y2006m02 VALUES (NEW.*);`,
			},
			{
				Statement: `    ELSIF ( NEW.logdate >= DATE '2006-03-01' AND
            NEW.logdate < DATE '2006-04-01' ) THEN
        INSERT INTO measurement_y2006m03 VALUES (NEW.*);`,
			},
			{
				Statement: `    ELSIF ( NEW.logdate >= DATE '2007-01-01' AND
            NEW.logdate < DATE '2007-02-01' ) THEN
        INSERT INTO measurement_y2007m01 (city_id, logdate, peaktemp, unitsales)
            VALUES (NEW.*);`,
			},
			{
				Statement: `    ELSE
        RAISE EXCEPTION 'Date out of range.  Fix the measurement_insert_trigger() function!';`,
			},
			{
				Statement: `    END IF;`,
			},
			{
				Statement: `    RETURN NULL;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql ;`,
			},
			{
				Statement: `CREATE TRIGGER insert_measurement_trigger
    BEFORE INSERT ON measurement
    FOR EACH ROW EXECUTE PROCEDURE measurement_insert_trigger();`,
			},
			{
				Statement: `INSERT INTO measurement VALUES (1, '2006-02-10', 35, 10);`,
			},
			{
				Statement: `INSERT INTO measurement VALUES (1, '2006-02-16', 45, 20);`,
			},
			{
				Statement: `INSERT INTO measurement VALUES (1, '2006-03-17', 25, 10);`,
			},
			{
				Statement: `INSERT INTO measurement VALUES (1, '2006-03-27', 15, 40);`,
			},
			{
				Statement: `INSERT INTO measurement VALUES (1, '2007-01-15', 10, 10);`,
			},
			{
				Statement: `INSERT INTO measurement VALUES (1, '2007-01-17', 10, 10);`,
			},
			{
				Statement: `SELECT tableoid::regclass, * FROM measurement ORDER BY city_id, logdate;`,
				Results:   []sql.Row{{`measurement`, 0, `07-21-2005`, 5, 15}, {`measurement_y2006m02`, 1, `02-10-2006`, 35, 10}, {`measurement_y2006m02`, 1, `02-16-2006`, 45, 20}, {`measurement_y2006m03`, 1, `03-17-2006`, 25, 10}, {`measurement_y2006m03`, 1, `03-27-2006`, 15, 40}, {`measurement_y2007m01`, 1, `01-15-2007`, 10, 10}, {`measurement_y2007m01`, 1, `01-17-2007`, 10, 10}},
			},
			{
				Statement: `CREATE TABLE new_measurement (LIKE measurement) WITH (autovacuum_enabled=off);`,
			},
			{
				Statement: `INSERT INTO new_measurement VALUES (0, '2005-07-21', 25, 20);`,
			},
			{
				Statement: `INSERT INTO new_measurement VALUES (1, '2006-03-01', 20, 10);`,
			},
			{
				Statement: `INSERT INTO new_measurement VALUES (1, '2006-02-16', 50, 10);`,
			},
			{
				Statement: `INSERT INTO new_measurement VALUES (2, '2006-02-10', 20, 20);`,
			},
			{
				Statement: `INSERT INTO new_measurement VALUES (1, '2006-03-27', NULL, NULL);`,
			},
			{
				Statement: `INSERT INTO new_measurement VALUES (1, '2007-01-17', NULL, NULL);`,
			},
			{
				Statement: `INSERT INTO new_measurement VALUES (1, '2007-01-15', 5, NULL);`,
			},
			{
				Statement: `INSERT INTO new_measurement VALUES (1, '2007-01-16', 10, 10);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO ONLY measurement m
 USING new_measurement nm ON
      (m.city_id = nm.city_id and m.logdate=nm.logdate)
WHEN MATCHED AND nm.peaktemp IS NULL THEN DELETE
WHEN MATCHED THEN UPDATE
     SET peaktemp = greatest(m.peaktemp, nm.peaktemp),
        unitsales = m.unitsales + coalesce(nm.unitsales, 0)
WHEN NOT MATCHED THEN INSERT
     (city_id, logdate, peaktemp, unitsales)
   VALUES (city_id, logdate, peaktemp, unitsales);`,
			},
			{
				Statement: `SELECT tableoid::regclass, * FROM measurement ORDER BY city_id, logdate, peaktemp;`,
				Results:   []sql.Row{{`measurement`, 0, `07-21-2005`, 25, 35}, {`measurement_y2006m02`, 1, `02-10-2006`, 35, 10}, {`measurement_y2006m02`, 1, `02-16-2006`, 45, 20}, {`measurement_y2006m02`, 1, `02-16-2006`, 50, 10}, {`measurement_y2006m03`, 1, `03-01-2006`, 20, 10}, {`measurement_y2006m03`, 1, `03-17-2006`, 25, 10}, {`measurement_y2006m03`, 1, `03-27-2006`, 15, 40}, {`measurement_y2006m03`, 1, `03-27-2006`, ``, ``}, {`measurement_y2007m01`, 1, `01-15-2007`, 5, ``}, {`measurement_y2007m01`, 1, `01-15-2007`, 10, 10}, {`measurement_y2007m01`, 1, `01-16-2007`, 10, 10}, {`measurement_y2007m01`, 1, `01-17-2007`, 10, 10}, {`measurement_y2007m01`, 1, `01-17-2007`, ``, ``}, {`measurement_y2006m02`, 2, `02-10-2006`, 20, 20}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `MERGE into measurement m
 USING new_measurement nm ON
      (m.city_id = nm.city_id and m.logdate=nm.logdate)
WHEN MATCHED AND nm.peaktemp IS NULL THEN DELETE
WHEN MATCHED THEN UPDATE
     SET peaktemp = greatest(m.peaktemp, nm.peaktemp),
        unitsales = m.unitsales + coalesce(nm.unitsales, 0)
WHEN NOT MATCHED THEN INSERT
     (city_id, logdate, peaktemp, unitsales)
   VALUES (city_id, logdate, peaktemp, unitsales);`,
			},
			{
				Statement: `SELECT tableoid::regclass, * FROM measurement ORDER BY city_id, logdate;`,
				Results:   []sql.Row{{`measurement`, 0, `07-21-2005`, 25, 35}, {`measurement_y2006m02`, 1, `02-10-2006`, 35, 10}, {`measurement_y2006m02`, 1, `02-16-2006`, 50, 30}, {`measurement_y2006m03`, 1, `03-01-2006`, 20, 10}, {`measurement_y2006m03`, 1, `03-17-2006`, 25, 10}, {`measurement_y2007m01`, 1, `01-15-2007`, 10, 10}, {`measurement_y2007m01`, 1, `01-16-2007`, 10, 10}, {`measurement_y2006m02`, 2, `02-10-2006`, 20, 20}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `MERGE INTO new_measurement nm
 USING ONLY measurement m ON
      (nm.city_id = m.city_id and nm.logdate=m.logdate)
WHEN MATCHED THEN DELETE;`,
			},
			{
				Statement: `SELECT * FROM new_measurement ORDER BY city_id, logdate;`,
				Results:   []sql.Row{{1, `02-16-2006`, 50, 10}, {1, `03-01-2006`, 20, 10}, {1, `03-27-2006`, ``, ``}, {1, `01-15-2007`, 5, ``}, {1, `01-16-2007`, 10, 10}, {1, `01-17-2007`, ``, ``}, {2, `02-10-2006`, 20, 20}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `MERGE INTO new_measurement nm
 USING measurement m ON
      (nm.city_id = m.city_id and nm.logdate=m.logdate)
WHEN MATCHED THEN DELETE;`,
			},
			{
				Statement: `SELECT * FROM new_measurement ORDER BY city_id, logdate;`,
				Results:   []sql.Row{{1, `03-27-2006`, ``, ``}, {1, `01-17-2007`, ``, ``}},
			},
			{
				Statement: `DROP TABLE measurement, new_measurement CASCADE;`,
			},
			{
				Statement: `DROP FUNCTION measurement_insert_trigger();`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP TABLE target, target2;`,
			},
			{
				Statement: `DROP TABLE source, source2;`,
			},
			{
				Statement: `DROP FUNCTION merge_trigfunc();`,
			},
			{
				Statement: `DROP USER regress_merge_privs;`,
			},
			{
				Statement: `DROP USER regress_merge_no_privs;`,
			},
		},
	})
}
