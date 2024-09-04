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

func TestStats(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_stats)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_stats,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SHOW track_counts;  -- must be on`,
				Results:   []sql.Row{{`on`}},
			},
			{
				Statement: `SET enable_seqscan TO on;`,
			},
			{
				Statement: `SET enable_indexscan TO on;`,
			},
			{
				Statement: `SET enable_indexonlyscan TO off;`,
			},
			{
				Statement: `SET track_functions TO 'all';`,
			},
			{
				Statement: `SELECT oid AS dboid from pg_database where datname = current_database() \gset
BEGIN;`,
			},
			{
				Statement: `SET LOCAL stats_fetch_consistency = snapshot;`,
			},
			{
				Statement: `CREATE TABLE prevstats AS
SELECT t.seq_scan, t.seq_tup_read, t.idx_scan, t.idx_tup_fetch,
       (b.heap_blks_read + b.heap_blks_hit) AS heap_blks,
       (b.idx_blks_read + b.idx_blks_hit) AS idx_blks,
       pg_stat_get_snapshot_timestamp() as snap_ts
  FROM pg_catalog.pg_stat_user_tables AS t,
       pg_catalog.pg_statio_user_tables AS b
 WHERE t.relname='tenk2' AND b.relname='tenk2';`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE TABLE trunc_stats_test(id serial);`,
			},
			{
				Statement: `CREATE TABLE trunc_stats_test1(id serial, stuff text);`,
			},
			{
				Statement: `CREATE TABLE trunc_stats_test2(id serial);`,
			},
			{
				Statement: `CREATE TABLE trunc_stats_test3(id serial, stuff text);`,
			},
			{
				Statement: `CREATE TABLE trunc_stats_test4(id serial);`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test DEFAULT VALUES;`,
			},
			{
				Statement: `TRUNCATE trunc_stats_test;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test1 DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test1 DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test1 DEFAULT VALUES;`,
			},
			{
				Statement: `UPDATE trunc_stats_test1 SET id = id + 10 WHERE id IN (1, 2);`,
			},
			{
				Statement: `DELETE FROM trunc_stats_test1 WHERE id = 3;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE trunc_stats_test1 SET id = id + 100;`,
			},
			{
				Statement: `TRUNCATE trunc_stats_test1;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test1 DEFAULT VALUES;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test2 DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test2 DEFAULT VALUES;`,
			},
			{
				Statement: `SAVEPOINT p1;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test2 DEFAULT VALUES;`,
			},
			{
				Statement: `TRUNCATE trunc_stats_test2;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test2 DEFAULT VALUES;`,
			},
			{
				Statement: `RELEASE SAVEPOINT p1;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test3 DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test3 DEFAULT VALUES;`,
			},
			{
				Statement: `SAVEPOINT p1;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test3 DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test3 DEFAULT VALUES;`,
			},
			{
				Statement: `TRUNCATE trunc_stats_test3;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test3 DEFAULT VALUES;`,
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT p1;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test4 DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test4 DEFAULT VALUES;`,
			},
			{
				Statement: `TRUNCATE trunc_stats_test4;`,
			},
			{
				Statement: `INSERT INTO trunc_stats_test4 DEFAULT VALUES;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT count(*) FROM tenk2;`,
				Results:   []sql.Row{{10000}},
			},
			{
				Statement: `SET enable_bitmapscan TO off;`,
			},
			{
				Statement: `SELECT count(*) FROM tenk2 WHERE unique1 = 1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
			{
				Statement: `SELECT pg_stat_force_next_flush();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL stats_fetch_consistency = snapshot;`,
			},
			{
				Statement: `SELECT relname, n_tup_ins, n_tup_upd, n_tup_del, n_live_tup, n_dead_tup
  FROM pg_stat_user_tables
 WHERE relname like 'trunc_stats_test%' order by relname;`,
				Results: []sql.Row{{`trunc_stats_test`, 3, 0, 0, 0, 0}, {`trunc_stats_test1`, 4, 2, 1, 1, 0}, {`trunc_stats_test2`, 1, 0, 0, 1, 0}, {`trunc_stats_test3`, 4, 0, 0, 2, 2}, {`trunc_stats_test4`, 2, 0, 0, 0, 2}},
			},
			{
				Statement: `SELECT st.seq_scan >= pr.seq_scan + 1,
       st.seq_tup_read >= pr.seq_tup_read + cl.reltuples,
       st.idx_scan >= pr.idx_scan + 1,
       st.idx_tup_fetch >= pr.idx_tup_fetch + 1
  FROM pg_stat_user_tables AS st, pg_class AS cl, prevstats AS pr
 WHERE st.relname='tenk2' AND cl.relname='tenk2';`,
				Results: []sql.Row{{true, true, true, true}},
			},
			{
				Statement: `SELECT st.heap_blks_read + st.heap_blks_hit >= pr.heap_blks + cl.relpages,
       st.idx_blks_read + st.idx_blks_hit >= pr.idx_blks + 1
  FROM pg_statio_user_tables AS st, pg_class AS cl, prevstats AS pr
 WHERE st.relname='tenk2' AND cl.relname='tenk2';`,
				Results: []sql.Row{{true, true}},
			},
			{
				Statement: `SELECT pr.snap_ts < pg_stat_get_snapshot_timestamp() as snapshot_newer
FROM prevstats AS pr;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `----
---
CREATE FUNCTION stats_test_func1() RETURNS VOID LANGUAGE plpgsql AS $$BEGIN END;$$;`,
			},
			{
				Statement: `SELECT 'stats_test_func1()'::regprocedure::oid AS stats_test_func1_oid \gset
CREATE FUNCTION stats_test_func2() RETURNS VOID LANGUAGE plpgsql AS $$BEGIN END;$$;`,
			},
			{
				Statement: `SELECT 'stats_test_func2()'::regprocedure::oid AS stats_test_func2_oid \gset
BEGIN;`,
			},
			{
				Statement: `SET LOCAL stats_fetch_consistency = none;`,
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(:stats_test_func1_oid);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_xact_function_calls(:stats_test_func1_oid);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_test_func1();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_xact_function_calls(:stats_test_func1_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT stats_test_func1();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_xact_function_calls(:stats_test_func1_oid);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(:stats_test_func1_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT stats_test_func2();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SAVEPOINT foo;`,
			},
			{
				Statement: `SELECT stats_test_func2();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT foo;`,
			},
			{
				Statement: `SELECT pg_stat_get_xact_function_calls(:stats_test_func2_oid);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT stats_test_func2();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT stats_test_func2();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT pg_stat_force_next_flush();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT funcname, calls FROM pg_stat_user_functions WHERE funcid = :stats_test_func1_oid;`,
				Results:   []sql.Row{{`stats_test_func1`, 2}},
			},
			{
				Statement: `SELECT funcname, calls FROM pg_stat_user_functions WHERE funcid = :stats_test_func2_oid;`,
				Results:   []sql.Row{{`stats_test_func2`, 4}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT funcname, calls FROM pg_stat_user_functions WHERE funcid = :stats_test_func1_oid;`,
				Results:   []sql.Row{{`stats_test_func1`, 2}},
			},
			{
				Statement: `DROP FUNCTION stats_test_func1();`,
			},
			{
				Statement: `SELECT funcname, calls FROM pg_stat_user_functions WHERE funcid = :stats_test_func1_oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(:stats_test_func1_oid);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT funcname, calls FROM pg_stat_user_functions WHERE funcid = :stats_test_func1_oid;`,
				Results:   []sql.Row{{`stats_test_func1`, 2}},
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(:stats_test_func1_oid);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DROP FUNCTION stats_test_func1();`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT funcname, calls FROM pg_stat_user_functions WHERE funcid = :stats_test_func1_oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(:stats_test_func1_oid);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT stats_test_func2();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SAVEPOINT a;`,
			},
			{
				Statement: `SELECT stats_test_func2();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SAVEPOINT b;`,
			},
			{
				Statement: `DROP FUNCTION stats_test_func2();`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT funcname, calls FROM pg_stat_user_functions WHERE funcid = :stats_test_func2_oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(:stats_test_func2_oid);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `CREATE TABLE drop_stats_test();`,
			},
			{
				Statement: `INSERT INTO drop_stats_test DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT 'drop_stats_test'::regclass::oid AS drop_stats_test_oid \gset
CREATE TABLE drop_stats_test_xact();`,
			},
			{
				Statement: `INSERT INTO drop_stats_test_xact DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT 'drop_stats_test_xact'::regclass::oid AS drop_stats_test_xact_oid \gset
CREATE TABLE drop_stats_test_subxact();`,
			},
			{
				Statement: `INSERT INTO drop_stats_test_subxact DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT 'drop_stats_test_subxact'::regclass::oid AS drop_stats_test_subxact_oid \gset
SELECT pg_stat_force_next_flush();`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP TABLE drop_stats_test;`,
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT pg_stat_get_xact_tuples_inserted(:drop_stats_test_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT pg_stat_get_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT pg_stat_get_xact_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO drop_stats_test_xact DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT pg_stat_get_xact_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP TABLE drop_stats_test_xact;`,
			},
			{
				Statement: `SELECT pg_stat_get_xact_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT pg_stat_force_next_flush();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT pg_stat_get_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT pg_stat_get_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO drop_stats_test_xact DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT pg_stat_get_xact_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP TABLE drop_stats_test_xact;`,
			},
			{
				Statement: `SELECT pg_stat_get_xact_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT pg_stat_force_next_flush();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT pg_stat_get_tuples_inserted(:drop_stats_test_xact_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_subxact_oid);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO drop_stats_test_subxact DEFAULT VALUES;`,
			},
			{
				Statement: `SAVEPOINT sp1;`,
			},
			{
				Statement: `INSERT INTO drop_stats_test_subxact DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT pg_stat_get_xact_tuples_inserted(:drop_stats_test_subxact_oid);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SAVEPOINT sp2;`,
			},
			{
				Statement: `DROP TABLE drop_stats_test_subxact;`,
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT sp2;`,
			},
			{
				Statement: `SELECT pg_stat_get_xact_tuples_inserted(:drop_stats_test_subxact_oid);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT pg_stat_force_next_flush();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_subxact_oid);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_subxact_oid);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SAVEPOINT sp1;`,
			},
			{
				Statement: `DROP TABLE drop_stats_test_subxact;`,
			},
			{
				Statement: `SAVEPOINT sp2;`,
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT sp1;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_subxact_oid);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_subxact_oid);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SAVEPOINT sp1;`,
			},
			{
				Statement: `DROP TABLE drop_stats_test_subxact;`,
			},
			{
				Statement: `SAVEPOINT sp2;`,
			},
			{
				Statement: `RELEASE SAVEPOINT sp1;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT pg_stat_get_live_tuples(:drop_stats_test_subxact_oid);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `DROP TABLE trunc_stats_test, trunc_stats_test1, trunc_stats_test2, trunc_stats_test3, trunc_stats_test4;`,
			},
			{
				Statement: `DROP TABLE prevstats;`,
			},
			{
				Statement: `-----
-----
SELECT sessions AS db_stat_sessions FROM pg_stat_database WHERE datname = (SELECT current_database()) \gset
\c
SELECT pg_stat_force_next_flush();`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT sessions > :db_stat_sessions FROM pg_stat_database WHERE datname = (SELECT current_database());`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT checkpoints_req AS rqst_ckpts_before FROM pg_stat_bgwriter \gset
SELECT wal_bytes AS wal_bytes_before FROM pg_stat_wal \gset
CREATE TABLE test_stats_temp AS SELECT 17;`,
			},
			{
				Statement: `DROP TABLE test_stats_temp;`,
			},
			{
				Statement: `CHECKPOINT;`,
			},
			{
				Statement: `CHECKPOINT;`,
			},
			{
				Statement: `SELECT checkpoints_req > :rqst_ckpts_before FROM pg_stat_bgwriter;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT wal_bytes > :wal_bytes_before FROM pg_stat_wal;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `-----
-----
SELECT stats_reset AS slru_commit_ts_reset_ts FROM pg_stat_slru WHERE name = 'CommitTs' \gset
SELECT stats_reset AS slru_notify_reset_ts FROM pg_stat_slru WHERE name = 'Notify' \gset
SELECT pg_stat_reset_slru('CommitTs');`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_reset > :'slru_commit_ts_reset_ts'::timestamptz FROM pg_stat_slru WHERE name = 'CommitTs';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT stats_reset AS slru_commit_ts_reset_ts FROM pg_stat_slru WHERE name = 'CommitTs' \gset
SELECT pg_stat_reset_slru(NULL);`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_reset > :'slru_commit_ts_reset_ts'::timestamptz FROM pg_stat_slru WHERE name = 'CommitTs';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT stats_reset > :'slru_notify_reset_ts'::timestamptz FROM pg_stat_slru WHERE name = 'Notify';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT stats_reset AS archiver_reset_ts FROM pg_stat_archiver \gset
SELECT pg_stat_reset_shared('archiver');`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_reset > :'archiver_reset_ts'::timestamptz FROM pg_stat_archiver;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT stats_reset AS archiver_reset_ts FROM pg_stat_archiver \gset
SELECT stats_reset AS bgwriter_reset_ts FROM pg_stat_bgwriter \gset
SELECT pg_stat_reset_shared('bgwriter');`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_reset > :'bgwriter_reset_ts'::timestamptz FROM pg_stat_bgwriter;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT stats_reset AS bgwriter_reset_ts FROM pg_stat_bgwriter \gset
SELECT stats_reset AS wal_reset_ts FROM pg_stat_wal \gset
SELECT pg_stat_reset_shared('wal');`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_reset > :'wal_reset_ts'::timestamptz FROM pg_stat_wal;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT stats_reset AS wal_reset_ts FROM pg_stat_wal \gset
SELECT pg_stat_reset_shared(NULL);`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_reset = :'archiver_reset_ts'::timestamptz FROM pg_stat_archiver;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT stats_reset = :'bgwriter_reset_ts'::timestamptz FROM pg_stat_bgwriter;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT stats_reset = :'wal_reset_ts'::timestamptz FROM pg_stat_wal;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT pg_stat_reset();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_reset AS db_reset_ts FROM pg_stat_database WHERE datname = (SELECT current_database()) \gset
SELECT pg_stat_reset();`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT stats_reset > :'db_reset_ts'::timestamptz FROM pg_stat_database WHERE datname = (SELECT current_database());`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `----
----
BEGIN;`,
			},
			{
				Statement: `SET LOCAL stats_fetch_consistency = snapshot;`,
			},
			{
				Statement: `SELECT pg_stat_get_snapshot_timestamp();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_snapshot_timestamp() >= NOW();`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT pg_stat_clear_snapshot();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_snapshot_timestamp();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `----
----
BEGIN;`,
			},
			{
				Statement: `SET LOCAL stats_fetch_consistency = cache;`,
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_snapshot_timestamp() IS NOT NULL AS snapshot_ok;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SET LOCAL stats_fetch_consistency = snapshot;`,
			},
			{
				Statement: `SELECT pg_stat_get_snapshot_timestamp() IS NOT NULL AS snapshot_ok;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_snapshot_timestamp() IS NOT NULL AS snapshot_ok;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SET LOCAL stats_fetch_consistency = none;`,
			},
			{
				Statement: `SELECT pg_stat_get_snapshot_timestamp() IS NOT NULL AS snapshot_ok;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT pg_stat_get_function_calls(0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_snapshot_timestamp() IS NOT NULL AS snapshot_ok;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `----
----
SELECT pg_stat_have_stats('bgwriter', 0, 0);`,
				Results: []sql.Row{{true}},
			},
			{
				Statement:   `SELECT pg_stat_have_stats('zaphod', 0, 0);`,
				ErrorString: `invalid statistics kind: "zaphod"`,
			},
			{
				Statement: `SELECT pg_stat_have_stats('database', :dboid, 1);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT pg_stat_have_stats('database', :dboid, 0);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `CREATE table stats_test_tab1 as select generate_series(1,10) a;`,
			},
			{
				Statement: `CREATE index stats_test_idx1 on stats_test_tab1(a);`,
			},
			{
				Statement: `SELECT 'stats_test_idx1'::regclass::oid AS stats_test_idx1_oid \gset
SET enable_seqscan TO off;`,
			},
			{
				Statement: `select a from stats_test_tab1 where a = 3;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP index stats_test_idx1;`,
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE index stats_test_idx1 on stats_test_tab1(a);`,
			},
			{
				Statement: `SELECT 'stats_test_idx1'::regclass::oid AS stats_test_idx1_oid \gset
select a from stats_test_tab1 where a = 3;`,
				Results: []sql.Row{{3}},
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `CREATE index stats_test_idx1 on stats_test_tab1(a);`,
			},
			{
				Statement: `SELECT 'stats_test_idx1'::regclass::oid AS stats_test_idx1_oid \gset
select a from stats_test_tab1 where a = 3;`,
				Results: []sql.Row{{3}},
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `REINDEX index CONCURRENTLY stats_test_idx1;`,
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT 'stats_test_idx1'::regclass::oid AS stats_test_idx1_oid \gset
SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP index stats_test_idx1;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT pg_stat_have_stats('relation', :dboid, :stats_test_idx1_oid);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SET enable_seqscan TO on;`,
			},
			{
				Statement: `SELECT pg_stat_get_replication_slot(NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_stat_get_subscription_stats(NULL);`,
				Results:   []sql.Row{{``}},
			},
		},
	})
}
