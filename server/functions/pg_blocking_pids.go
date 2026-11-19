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

package functions

import (
	"sort"

	"github.com/dolthub/dolt/go/libraries/doltcore/rowjournal"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgBlockingPids registers the functions to the catalog.
func initPgBlockingPids() {
	framework.RegisterFunction(pg_blocking_pids)
	framework.RegisterFunction(pg_isolation_test_session_is_blocked)
}

// pg_blocking_pids represents the PostgreSQL function of the same name, which returns the process IDs of the sessions
// that are keeping the session with the process ID given from acquiring a lock. It reports row and table locks.
var pg_blocking_pids = framework.Function1{
	Name:               "pg_blocking_pids",
	Return:             pgtypes.Int32Array,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Int32},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		blocking := BlockingSessions(ctx, uint32(val.(int32)))
		pids := make([]any, len(blocking))
		for i, pid := range blocking {
			pids[i] = int32(pid)
		}

		return pids, nil
	},
}

// pg_isolation_test_session_is_blocked represents the PostgreSQL function of the same name, which the isolation
// tester uses to tell that a session is waiting for a lock held by one of the other sessions of the test.
var pg_isolation_test_session_is_blocked = framework.Function2{
	Name:               "pg_isolation_test_session_is_blocked",
	Return:             pgtypes.Bool,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32Array},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		interesting := make(map[uint32]struct{})
		for _, pid := range val2.([]any) {
			if pid, ok := pid.(int32); ok {
				interesting[uint32(pid)] = struct{}{}
			}
		}

		for _, pid := range BlockingSessions(ctx, uint32(val1.(int32))) {
			if _, ok := interesting[pid]; ok {
				return true, nil
			}
		}

		return false, nil
	},
}

// RowJournal returns the journal of row locks for the server that |ctx| belongs to, or nil if it has none.
func RowJournal(ctx *sql.Context) *rowjournal.Journal {
	sess, ok := ctx.Session.(*dsess.DoltSession)
	if !ok || sess.Provider() == nil {
		return nil
	}

	return sess.Provider().RowJournal()
}

// BlockingSessions returns the IDs of the sessions that hold a row lock, a table lock or an advisory lock that the
// session given is waiting for, or are ahead of it in the queue for one, in ascending order.
func BlockingSessions(ctx *sql.Context, sessionID uint32) []uint32 {
	blocking := make(map[uint32]struct{})
	if owner, ok := advisoryLockBlocker(sessionID); ok {
		blocking[owner] = struct{}{}
	}

	if journal := RowJournal(ctx); journal != nil {
		sessions := make(map[rowjournal.TxnID]uint32)
		for _, txn := range journal.ActiveTxns() {
			sessions[txn.ID] = txn.SessionID
		}

		for _, lock := range journal.Locks() {
			if lock.Granted || lock.SessionID != sessionID {
				continue
			}

			for _, txn := range lock.BlockedBy {
				if session, ok := sessions[txn]; ok {
					blocking[session] = struct{}{}
				}
			}
		}
	}

	ids := make([]uint32, 0, len(blocking))
	for id := range blocking {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
