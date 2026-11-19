// Copyright 2026 Dolthub, Inc.
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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

var rowLockSetup = []string{
	"CREATE TABLE t (pk INT PRIMARY KEY, v INT)",
	"INSERT INTO t VALUES (1, 10), (2, 20), (3, 30)",
	"CREATE TABLE u (pk INT PRIMARY KEY, t_pk INT)",
	"INSERT INTO u VALUES (11, 1), (12, 2), (13, 3)",
}

// TestRowLocks tests row-level locking: the locks that statements take on the rows they change, and the locks that
// SELECT takes with a locking clause.
func TestRowLocks(t *testing.T) {
	RunTransactionTests(t, []ScriptTest{
		{
			Name:        "FOR UPDATE makes an update of the row wait until the reader commits",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT * FROM t WHERE pk = 1 FOR UPDATE", Expected: []sql.Row{{1, 10}}},
				{Query: "/* client b */ UPDATE t SET v = 21 WHERE pk = 2", Expected: []sql.Row{}},
				{Query: "/* client b */ UPDATE t SET v = 11 WHERE pk = 1", ExpectedBlocking: true},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT * FROM t ORDER BY pk", Expected: []sql.Row{{1, 11}, {2, 21}, {3, 30}}},
			},
		},
		{
			Name:        "an update waits for an update of the same row, and fails when it commits",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client b */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ UPDATE t SET v = v + 1 WHERE pk = 1", Expected: []sql.Row{}},
				{
					Query:            "/* client b */ UPDATE t SET v = v + 1 WHERE pk = 1",
					ExpectedBlocking: true,
					ExpectedErr:      "could not serialize access due to concurrent update",
					ExpectedErrCode:  "40001",
				},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ ROLLBACK", Expected: []sql.Row{}},
				{Query: "/* client b */ UPDATE t SET v = v + 1 WHERE pk = 1", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT v FROM t WHERE pk = 1", Expected: []sql.Row{{12}}},
			},
		},
		{
			Name:        "an update waits for an update of the same row, and proceeds when it rolls back",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ UPDATE t SET v = v + 1 WHERE pk = 1", Expected: []sql.Row{}},
				{Query: "/* client b */ UPDATE t SET v = v + 5 WHERE pk = 1", ExpectedBlocking: true},
				{Query: "/* client a */ ROLLBACK", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 1", Expected: []sql.Row{{15}}},
			},
		},
		{
			Name:        "a row changed since the transaction began can't be locked or changed",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT v FROM t WHERE pk = 1", Expected: []sql.Row{{10}}},
				{Query: "/* client b */ UPDATE t SET v = 11 WHERE pk = 1", Expected: []sql.Row{}},
				{Query: "/* client b */ DELETE FROM t WHERE pk = 2", Expected: []sql.Row{}},
				{
					Query:           "/* client a */ SELECT v FROM t WHERE pk = 1 FOR UPDATE",
					ExpectedErr:     "could not serialize access due to concurrent update",
					ExpectedErrCode: "40001",
				},
				{Query: "/* client a */ ROLLBACK", Expected: []sql.Row{}},
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client b */ DELETE FROM t WHERE pk = 3", Expected: []sql.Row{}},
				{
					Query:           "/* client a */ UPDATE t SET v = 0 WHERE pk = 3",
					ExpectedErr:     "could not serialize access due to concurrent delete",
					ExpectedErrCode: "40001",
				},
				{Query: "/* client a */ ROLLBACK", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT * FROM t ORDER BY pk FOR UPDATE", Expected: []sql.Row{{1, 11}}},
			},
		},
		{
			Name:        "lock strengths conflict as they do in Postgres",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client b */ BEGIN", Expected: []sql.Row{}},
				// KEY SHARE allows everything but a lock for deleting the row or changing its key.
				{Query: "/* client a */ SELECT v FROM t WHERE pk = 1 FOR KEY SHARE", Expected: []sql.Row{{10}}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 1 FOR KEY SHARE", Expected: []sql.Row{{10}}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 1 FOR SHARE", Expected: []sql.Row{{10}}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 1 FOR NO KEY UPDATE", Expected: []sql.Row{{10}}},
				{Query: "/* client b */ UPDATE t SET v = 11 WHERE pk = 1", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 1 FOR UPDATE NOWAIT", ExpectedErr: "could not obtain lock on row", ExpectedErrCode: "55P03"},
				{Query: "/* client b */ ROLLBACK", Expected: []sql.Row{}},
				{Query: "/* client b */ DELETE FROM t WHERE pk = 1", ExpectedBlocking: true},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT count(*) FROM t WHERE pk = 1", Expected: []sql.Row{{0}}},
				// SHARE allows other SHARE and KEY SHARE locks, and no writes.
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client b */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT v FROM t WHERE pk = 2 FOR SHARE", Expected: []sql.Row{{20}}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 2 FOR SHARE", Expected: []sql.Row{{20}}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 2 FOR KEY SHARE", Expected: []sql.Row{{20}}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 2 FOR NO KEY UPDATE NOWAIT", ExpectedErr: "could not obtain lock on row", ExpectedErrCode: "55P03"},
				{Query: "/* client b */ ROLLBACK", Expected: []sql.Row{}},
				{Query: "/* client b */ UPDATE t SET v = 21 WHERE pk = 2", ExpectedBlocking: true},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 2", Expected: []sql.Row{{21}}},
				// NO KEY UPDATE allows only KEY SHARE.
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT v FROM t WHERE pk = 3 FOR NO KEY UPDATE", Expected: []sql.Row{{30}}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 3 FOR KEY SHARE", Expected: []sql.Row{{30}}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 3 FOR SHARE NOWAIT", ExpectedErr: "could not obtain lock on row", ExpectedErrCode: "55P03"},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 3 FOR NO KEY UPDATE NOWAIT", ExpectedErr: "could not obtain lock on row", ExpectedErrCode: "55P03"},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
			},
		},
		{
			Name:        "NOWAIT fails and SKIP LOCKED leaves out a row that is locked",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT * FROM t WHERE pk = 1 FOR UPDATE", Expected: []sql.Row{{1, 10}}},
				{
					Query:           "/* client b */ SELECT * FROM t WHERE pk = 1 FOR UPDATE NOWAIT",
					ExpectedErr:     "could not obtain lock on row",
					ExpectedErrCode: "55P03",
				},
				{Query: "/* client b */ SELECT * FROM t WHERE pk = 1 FOR UPDATE SKIP LOCKED", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT pk FROM t ORDER BY pk LIMIT 1 FOR UPDATE SKIP LOCKED", Expected: []sql.Row{{2}}},
				{Query: "/* client b */ SELECT pk FROM t ORDER BY pk FOR SHARE SKIP LOCKED", Expected: []sql.Row{{2}, {3}}},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT pk FROM t ORDER BY pk FOR UPDATE NOWAIT", Expected: []sql.Row{{1}, {2}, {3}}},
			},
		},
		{
			Name:        "a locking clause locks the tables it names, each as strongly as it says",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{
					Query:    "/* client a */ SELECT t.pk, u.pk FROM t JOIN u ON u.t_pk = t.pk WHERE t.pk = 1 AND u.pk = 11 FOR UPDATE OF u",
					Expected: []sql.Row{{1, 11}},
				},
				{Query: "/* client b */ UPDATE t SET v = 11 WHERE pk = 1", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT pk FROM u WHERE pk = 11 FOR KEY SHARE NOWAIT", ExpectedErr: "could not obtain lock on row", ExpectedErrCode: "55P03"},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{
					Query:    "/* client a */ SELECT x.pk, y.pk FROM t x JOIN u y ON y.t_pk = x.pk WHERE x.pk = 2 AND y.pk = 12 FOR UPDATE OF x FOR SHARE OF y",
					Expected: []sql.Row{{2, 12}},
				},
				{Query: "/* client b */ SELECT pk FROM u WHERE pk = 12 FOR SHARE NOWAIT", Expected: []sql.Row{{12}}},
				{Query: "/* client b */ SELECT pk FROM u WHERE pk = 12 FOR UPDATE NOWAIT", ExpectedErr: "could not obtain lock on row", ExpectedErrCode: "55P03"},
				{Query: "/* client b */ SELECT pk FROM t WHERE pk = 2 FOR KEY SHARE NOWAIT", ExpectedErr: "could not obtain lock on row", ExpectedErrCode: "55P03"},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{
					Query:       "/* client a */ SELECT * FROM t FOR UPDATE OF u",
					ExpectedErr: "u",
				},
				{
					Query:           "/* client a */ SELECT * FROM t FOR UPDATE OF public.t",
					ExpectedErr:     "FOR UPDATE must specify unqualified relation names",
					ExpectedErrCode: "42601",
				},
			},
		},
		{
			Name:        "a locking clause is refused where rows of the result aren't rows of a table",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{
					Query:           "/* client a */ SELECT DISTINCT v FROM t FOR UPDATE",
					ExpectedErr:     "FOR UPDATE is not allowed with DISTINCT clause",
					ExpectedErrCode: "0A000",
				},
				{
					Query:           "/* client a */ SELECT v FROM t GROUP BY v FOR SHARE",
					ExpectedErr:     "FOR SHARE is not allowed with GROUP BY clause",
					ExpectedErrCode: "0A000",
				},
				{
					Query:           "/* client a */ SELECT count(*) FROM t FOR NO KEY UPDATE",
					ExpectedErr:     "FOR NO KEY UPDATE is not allowed with aggregate functions",
					ExpectedErrCode: "0A000",
				},
				{
					Query:           "/* client a */ SELECT pk FROM t UNION SELECT pk FROM u FOR KEY SHARE",
					ExpectedErr:     "FOR KEY SHARE is not allowed with UNION/INTERSECT/EXCEPT",
					ExpectedErrCode: "0A000",
				},
				{
					Query:           "/* client a */ SELECT pk, row_number() OVER (ORDER BY pk) FROM t FOR UPDATE",
					ExpectedErr:     "FOR UPDATE is not allowed with window functions",
					ExpectedErrCode: "0A000",
				},
				// A subquery may aggregate without affecting the rows of the outer query.
				{
					Query:    "/* client a */ SELECT pk FROM t WHERE v = (SELECT max(v) FROM t) FOR UPDATE",
					Expected: []sql.Row{{3}},
				},
				{Query: "/* client a */ SELECT 1 FOR UPDATE", Expected: []sql.Row{{1}}},
			},
		},
		{
			Name:        "the transaction that completes a cycle of waits gets a deadlock error",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client b */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ UPDATE t SET v = 11 WHERE pk = 1", Expected: []sql.Row{}},
				{Query: "/* client b */ UPDATE t SET v = 22 WHERE pk = 2", Expected: []sql.Row{}},
				{Query: "/* client a */ UPDATE t SET v = 21 WHERE pk = 2", ExpectedBlocking: true},
				{
					Query:           "/* client b */ UPDATE t SET v = 12 WHERE pk = 1",
					ExpectedErr:     "deadlock detected",
					ExpectedErrCode: "40P01",
				},
				{Query: "/* client b */ ROLLBACK", Expected: []sql.Row{}},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT * FROM t ORDER BY pk", Expected: []sql.Row{{1, 11}, {2, 21}, {3, 30}}},
			},
		},
		{
			Name:        "lock_timeout bounds the wait for a row lock",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ UPDATE t SET v = 11 WHERE pk = 1", Expected: []sql.Row{}},
				{Query: "/* client b */ SET lock_timeout = 100", Expected: []sql.Row{}},
				{
					Query:           "/* client b */ UPDATE t SET v = 12 WHERE pk = 1",
					ExpectedErr:     "canceling statement due to lock timeout",
					ExpectedErrCode: "55P03",
				},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 1 FOR SHARE", ExpectedErr: "canceling statement due to lock timeout", ExpectedErrCode: "55P03"},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 1 FOR SHARE", Expected: []sql.Row{{11}}},
			},
		},
		{
			Name: "rows of a table without a primary key are locked by their contents",
			SetUpScript: []string{
				"CREATE TABLE k (c1 INT, c2 INT)",
				"CREATE INDEX k_c1 ON k (c1)",
				"INSERT INTO k VALUES (1, 1), (1, 1), (2, 2)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT * FROM k WHERE c1 = 1 FOR UPDATE", Expected: []sql.Row{{1, 1}, {1, 1}}},
				{Query: "/* client b */ SELECT * FROM k FOR UPDATE SKIP LOCKED", Expected: []sql.Row{{2, 2}}},
				{Query: "/* client b */ UPDATE k SET c2 = 5 WHERE c1 = 1", ExpectedBlocking: true},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT * FROM k ORDER BY c1, c2", Expected: []sql.Row{{1, 5}, {1, 5}, {2, 2}}},
			},
		},
		{
			Name:        "a session that disconnects gives up its row locks",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT * FROM t WHERE pk = 1 FOR UPDATE", Expected: []sql.Row{{1, 10}}},
				{Query: "/* client b */ UPDATE t SET v = 11 WHERE pk = 1", ExpectedBlocking: true},
				{Query: "/* client a */", CloseClient: true},
				{Query: "/* client b */ SELECT v FROM t WHERE pk = 1", Expected: []sql.Row{{11}}},
			},
		},
	})
}

// TestTransactionIDs tests that every transaction has an ID, and that IDs ascend in the order transactions start.
func TestTransactionIDs(t *testing.T) {
	RunTransactionTests(t, []ScriptTest{
		{
			Name: "txid_current returns the ID of the current transaction",
			SetUpScript: []string{
				"CREATE TABLE ids (name TEXT PRIMARY KEY, id BIGINT)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ INSERT INTO ids VALUES ('a1', txid_current())", Expected: []sql.Row{}},
				{Query: "/* client b */ INSERT INTO ids VALUES ('b', txid_current())", Expected: []sql.Row{}},
				{Query: "/* client a */ INSERT INTO ids VALUES ('a2', txid_current())", Expected: []sql.Row{}},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client a */ INSERT INTO ids VALUES ('a3', txid_current())", Expected: []sql.Row{}},
				{
					// A transaction keeps its ID for as long as it runs, and one that starts later has a higher ID.
					Query: `/* client a */ SELECT a1.id > 0, a1.id = a2.id, b.id > a1.id, a3.id > b.id
						FROM ids a1, ids a2, ids a3, ids b
						WHERE a1.name = 'a1' AND a2.name = 'a2' AND a3.name = 'a3' AND b.name = 'b'`,
					Expected: []sql.Row{{"t", "t", "t", "t"}},
				},
			},
		},
	})
}

// TestRowLockIntrospection tests the views and functions that report who holds and who waits for row locks.
func TestRowLockIntrospection(t *testing.T) {
	RunTransactionTests(t, []ScriptTest{
		{
			Name:        "pg_locks and pg_blocking_pids report row locks",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ SELECT count(*) FROM pg_locks", Expected: []sql.Row{{0}}},
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT * FROM t WHERE pk = 1 FOR SHARE", Expected: []sql.Row{{1, 10}}},
				{Query: "/* client a */ UPDATE t SET v = 21 WHERE pk = 2", Expected: []sql.Row{}},
				{
					Query: `/* client a */ SELECT locktype, relation::regclass::text, mode, granted, pid = pg_backend_pid(), transactionid::text = txid_current()::text
						FROM pg_locks ORDER BY mode`,
					Expected: []sql.Row{
						{"tuple", "t", "ExclusiveLock", "t", "t", "t"},
						{"tuple", "t", "RowShareLock", "t", "t", "t"},
					},
				},
				{Query: "/* client a */ SELECT pg_blocking_pids(pg_backend_pid())", Expected: []sql.Row{{"{}"}}},
				{Query: "/* client b */ UPDATE t SET v = 11 WHERE pk = 1", ExpectedBlocking: true},
				{
					// The waiting request is listed with the lock it wants, and is blocked by the session that holds
					// the row.
					Query: `/* client a */ SELECT mode, granted, pid <> pg_backend_pid(), pg_blocking_pids(pid) = ARRAY[pg_backend_pid()],
							pg_isolation_test_session_is_blocked(pid, ARRAY[pg_backend_pid()]),
							pg_isolation_test_session_is_blocked(pid, ARRAY[pid]),
							pg_isolation_test_session_is_blocked(pg_backend_pid(), ARRAY[pid])
						FROM pg_locks WHERE NOT granted`,
					Expected: []sql.Row{{"ExclusiveLock", "f", "t", "t", "t", "f", "f"}},
				},
				{
					// pg_stat_activity shows the waiting session as waiting for a lock, and both sessions'
					// transactions.
					Query: `/* client a */ SELECT pid = pg_backend_pid(), wait_event_type, wait_event, backend_xid IS NOT NULL
						FROM pg_stat_activity WHERE pid IN (SELECT pid FROM pg_locks) ORDER BY wait_event_type NULLS FIRST`,
					Expected: []sql.Row{{"t", nil, nil, "t"}, {"f", "Lock", "tuple", "t"}},
				},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT count(*) FROM pg_locks", Expected: []sql.Row{{0}}},
			},
		},
		{
			Name: "pg_blocking_pids reports the holder of an advisory lock that a session waits for",
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ CREATE TABLE pids (name TEXT PRIMARY KEY, pid INT)", Expected: []sql.Row{}},
				{Query: "/* client a */ INSERT INTO pids VALUES ('a', pg_backend_pid())", Expected: []sql.Row{}},
				{Query: "/* client b */ INSERT INTO pids VALUES ('b', pg_backend_pid())", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT pg_advisory_lock(42)", Expected: []sql.Row{{"t"}}},
				{Query: "/* client b */ SELECT pg_advisory_lock(42)", ExpectedBlocking: true},
				{
					Query: `/* client a */ SELECT pg_blocking_pids(b.pid) = ARRAY[a.pid],
							pg_isolation_test_session_is_blocked(b.pid, ARRAY[a.pid]),
							pg_isolation_test_session_is_blocked(a.pid, ARRAY[b.pid])
						FROM pids a, pids b WHERE a.name = 'a' AND b.name = 'b'`,
					Expected: []sql.Row{{"t", "t", "f"}},
				},
				{Query: "/* client a */ SELECT pg_advisory_unlock(42)", Expected: []sql.Row{{"t"}}},
				{
					Query:    "/* client b */ SELECT cardinality(pg_blocking_pids(pg_backend_pid()))",
					Expected: []sql.Row{{0}},
				},
				{Query: "/* client b */ SELECT pg_advisory_unlock(42)", Expected: []sql.Row{{"t"}}},
			},
		},
		{
			Name:        "pg_locks reports the lock that a DDL statement takes on a table",
			SetUpScript: rowLockSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN", Expected: []sql.Row{}},
				{Query: "/* client a */ SELECT * FROM t WHERE pk = 1 FOR KEY SHARE", Expected: []sql.Row{{1, 10}}},
				{Query: "/* client b */ ALTER TABLE t ADD COLUMN w INT", ExpectedBlocking: true},
				{
					Query: "/* client a */ SELECT locktype, relation::regclass::text, mode, granted FROM pg_locks ORDER BY granted DESC",
					Expected: []sql.Row{
						{"tuple", "t", "AccessShareLock", "t"},
						{"relation", "t", "AccessExclusiveLock", "f"},
					},
				},
				{Query: "/* client a */ COMMIT", Expected: []sql.Row{}},
				{Query: "/* client b */ SELECT pk, v, w FROM t WHERE pk = 1", Expected: []sql.Row{{1, 10, nil}}},
			},
		},
	})
}
