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

func TestAdvisoryLock(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_advisory_lock)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_advisory_lock,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT
	pg_advisory_xact_lock(1), pg_advisory_xact_lock_shared(2),
	pg_advisory_xact_lock(1, 1), pg_advisory_xact_lock_shared(2, 2);`,
				Results: []sql.Row{{``, ``, ``, ``}},
			},
			{
				Statement: `SELECT locktype, classid, objid, objsubid, mode, granted
	FROM pg_locks WHERE locktype = 'advisory'
	ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`advisory`, 0, 1, 1, `ExclusiveLock`, true}, {`advisory`, 0, 2, 1, `ShareLock`, true}, {`advisory`, 1, 1, 2, `ExclusiveLock`, true}, {`advisory`, 2, 2, 2, `ShareLock`, true}},
			},
			{
				Statement: `SELECT pg_advisory_unlock_all();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory';`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `SELECT
	pg_advisory_unlock(1), pg_advisory_unlock_shared(2),
	pg_advisory_unlock(1, 1), pg_advisory_unlock_shared(2, 2);`,
				Results: []sql.Row{{false, false, false, false}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT
	pg_advisory_xact_lock(1), pg_advisory_xact_lock_shared(2),
	pg_advisory_xact_lock(1, 1), pg_advisory_xact_lock_shared(2, 2);`,
				Results: []sql.Row{{``, ``, ``, ``}},
			},
			{
				Statement: `SELECT locktype, classid, objid, objsubid, mode, granted
	FROM pg_locks WHERE locktype = 'advisory'
	ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`advisory`, 0, 1, 1, `ExclusiveLock`, true}, {`advisory`, 0, 2, 1, `ShareLock`, true}, {`advisory`, 1, 1, 2, `ExclusiveLock`, true}, {`advisory`, 2, 2, 2, `ShareLock`, true}},
			},
			{
				Statement: `SELECT
	pg_advisory_lock(1), pg_advisory_lock_shared(2),
	pg_advisory_lock(1, 1), pg_advisory_lock_shared(2, 2);`,
				Results: []sql.Row{{``, ``, ``, ``}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT locktype, classid, objid, objsubid, mode, granted
	FROM pg_locks WHERE locktype = 'advisory'
	ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`advisory`, 0, 1, 1, `ExclusiveLock`, true}, {`advisory`, 0, 2, 1, `ShareLock`, true}, {`advisory`, 1, 1, 2, `ExclusiveLock`, true}, {`advisory`, 2, 2, 2, `ShareLock`, true}},
			},
			{
				Statement: `SELECT
	pg_advisory_unlock(1), pg_advisory_unlock(1),
	pg_advisory_unlock_shared(2), pg_advisory_unlock_shared(2),
	pg_advisory_unlock(1, 1), pg_advisory_unlock(1, 1),
	pg_advisory_unlock_shared(2, 2), pg_advisory_unlock_shared(2, 2);`,
				Results: []sql.Row{{true, false, true, false, true, false, true, false}},
			},
			{
				Statement: `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT
	pg_advisory_lock(1), pg_advisory_lock_shared(2),
	pg_advisory_lock(1, 1), pg_advisory_lock_shared(2, 2);`,
				Results: []sql.Row{{``, ``, ``, ``}},
			},
			{
				Statement: `SELECT locktype, classid, objid, objsubid, mode, granted
	FROM pg_locks WHERE locktype = 'advisory'
	ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`advisory`, 0, 1, 1, `ExclusiveLock`, true}, {`advisory`, 0, 2, 1, `ShareLock`, true}, {`advisory`, 1, 1, 2, `ExclusiveLock`, true}, {`advisory`, 2, 2, 2, `ShareLock`, true}},
			},
			{
				Statement: `SELECT
	pg_advisory_xact_lock(1), pg_advisory_xact_lock_shared(2),
	pg_advisory_xact_lock(1, 1), pg_advisory_xact_lock_shared(2, 2);`,
				Results: []sql.Row{{``, ``, ``, ``}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT locktype, classid, objid, objsubid, mode, granted
	FROM pg_locks WHERE locktype = 'advisory'
	ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`advisory`, 0, 1, 1, `ExclusiveLock`, true}, {`advisory`, 0, 2, 1, `ShareLock`, true}, {`advisory`, 1, 1, 2, `ExclusiveLock`, true}, {`advisory`, 2, 2, 2, `ShareLock`, true}},
			},
			{
				Statement: `SELECT pg_advisory_unlock_all();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT
	pg_advisory_xact_lock(1), pg_advisory_xact_lock(1),
	pg_advisory_xact_lock_shared(2), pg_advisory_xact_lock_shared(2),
	pg_advisory_xact_lock(1, 1), pg_advisory_xact_lock(1, 1),
	pg_advisory_xact_lock_shared(2, 2), pg_advisory_xact_lock_shared(2, 2);`,
				Results: []sql.Row{{``, ``, ``, ``, ``, ``, ``, ``}},
			},
			{
				Statement: `SELECT locktype, classid, objid, objsubid, mode, granted
	FROM pg_locks WHERE locktype = 'advisory'
	ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`advisory`, 0, 1, 1, `ExclusiveLock`, true}, {`advisory`, 0, 2, 1, `ShareLock`, true}, {`advisory`, 1, 1, 2, `ExclusiveLock`, true}, {`advisory`, 2, 2, 2, `ShareLock`, true}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT
	pg_advisory_lock(1), pg_advisory_lock(1),
	pg_advisory_lock_shared(2), pg_advisory_lock_shared(2),
	pg_advisory_lock(1, 1), pg_advisory_lock(1, 1),
	pg_advisory_lock_shared(2, 2), pg_advisory_lock_shared(2, 2);`,
				Results: []sql.Row{{``, ``, ``, ``, ``, ``, ``, ``}},
			},
			{
				Statement: `SELECT locktype, classid, objid, objsubid, mode, granted
	FROM pg_locks WHERE locktype = 'advisory'
	ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`advisory`, 0, 1, 1, `ExclusiveLock`, true}, {`advisory`, 0, 2, 1, `ShareLock`, true}, {`advisory`, 1, 1, 2, `ExclusiveLock`, true}, {`advisory`, 2, 2, 2, `ShareLock`, true}},
			},
			{
				Statement: `SELECT
	pg_advisory_unlock(1), pg_advisory_unlock(1),
	pg_advisory_unlock_shared(2), pg_advisory_unlock_shared(2),
	pg_advisory_unlock(1, 1), pg_advisory_unlock(1, 1),
	pg_advisory_unlock_shared(2, 2), pg_advisory_unlock_shared(2, 2);`,
				Results: []sql.Row{{true, true, true, true, true, true, true, true}},
			},
			{
				Statement: `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory';`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT
	pg_advisory_lock(1), pg_advisory_lock(1),
	pg_advisory_lock_shared(2), pg_advisory_lock_shared(2),
	pg_advisory_lock(1, 1), pg_advisory_lock(1, 1),
	pg_advisory_lock_shared(2, 2), pg_advisory_lock_shared(2, 2);`,
				Results: []sql.Row{{``, ``, ``, ``, ``, ``, ``, ``}},
			},
			{
				Statement: `SELECT locktype, classid, objid, objsubid, mode, granted
	FROM pg_locks WHERE locktype = 'advisory'
	ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`advisory`, 0, 1, 1, `ExclusiveLock`, true}, {`advisory`, 0, 2, 1, `ShareLock`, true}, {`advisory`, 1, 1, 2, `ExclusiveLock`, true}, {`advisory`, 2, 2, 2, `ShareLock`, true}},
			},
			{
				Statement: `SELECT pg_advisory_unlock_all();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory';`,
				Results:   []sql.Row{{0}},
			},
		},
	})
}
