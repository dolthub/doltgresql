// Copyright 2025 Dolthub, Inc.
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
	"fmt"
	"time"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqlserver"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initAdvisoryLockFunctions registers the advisory lock functions to the catalog.
func initAdvisoryLockFunctions() {
	framework.RegisterFunction(pg_advisory_lock_bigint)
	framework.RegisterFunction(pg_advisory_unlock_bigint)
	framework.RegisterFunction(pg_try_advisory_lock_bigint)
}

// pg_advisory_lock_bigint represents the pg_advisory_lock(bigint) function.
// https://www.postgresql.org/docs/9.1/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
var pg_advisory_lock_bigint = framework.Function1{
	Name:       "pg_advisory_lock",
	Return:     pgtypes.Bool,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		lockNumericId := val1.(int64)
		lockName := fmt.Sprintf("%v", lockNumericId)

		lockSubsystem := getLockSubsystem()
		// TODO: Postgres supports reentrant locks, meaning if pg_advisory_lock(123) is called multiple times,
		//       pg_advisory_unlock(123) must be called the same number of times to fully release a lock. This
		//       is different from MySQL's locking behavior, so LockSubsystem should be updated to support
		//       a reentrant mode in addition to the current mode.
		err := lockSubsystem.Lock(ctx, lockName, time.Millisecond*-1)
		return err == nil, err
	},
}

// pg_try_advisory_lock_bigint represents the pg_try_advisory_lock(bigint) function.
// https://www.postgresql.org/docs/9.1/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
var pg_try_advisory_lock_bigint = framework.Function1{
	Name:       "pg_try_advisory_lock",
	Return:     pgtypes.Bool,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		lockNumericId := val1.(int64)
		lockName := fmt.Sprintf("%v", lockNumericId)

		lockSubsystem := getLockSubsystem()
		// TODO: We currently need to specify a timeout, but it may be a better mapping to
		//       this function if we had a lockSubsystem.TryLock function that would try
		//       to grab the lock once and then return immediately. Until then, we set a
		//       short timeout and translate any timeout errors into a false return value.
		err := lockSubsystem.Lock(ctx, lockName, time.Millisecond*1)
		if sql.ErrLockTimeout.Is(err) {
			return false, nil
		}

		return err == nil, err
	},
}

// pg_advisory_unlock_bigint represents the pg_advisory_unlock(bigint) function.
// https://www.postgresql.org/docs/9.1/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
var pg_advisory_unlock_bigint = framework.Function1{
	Name:               "pg_advisory_unlock",
	Return:             pgtypes.Bool,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Int64},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		lockNumericId := val1.(int64)
		lockName := fmt.Sprintf("%v", lockNumericId)

		lockSubsystem := getLockSubsystem()
		err := lockSubsystem.Unlock(ctx, lockName)
		if sql.ErrLockDoesNotExist.Is(err) {
			return false, nil
		}

		return err == nil, err
	},
}

// getLockSubsystem returns the active lock system for the SQL engine.
func getLockSubsystem() *sql.LockSubsystem {
	// TODO: Is there a better way to get the active engine? Doesn't seem to be possible from
	//       the context or session.
	engine := sqlserver.GetRunningServer().Engine
	return engine.LS
}
