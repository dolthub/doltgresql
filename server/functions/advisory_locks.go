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

	"github.com/cockroachdb/errors"
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
		if lockSubsystem == nil {
			return false, errors.Errorf("lock subsystem not available")
		}

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
		if lockSubsystem == nil {
			return false, errors.Errorf("lock subsystem not available")
		}

		return lockSubsystem.TryLock(ctx, lockName)
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
		if lockSubsystem == nil {
			return false, errors.Errorf("lock subsystem not available")
		}

		err := lockSubsystem.Unlock(ctx, lockName)
		if sql.ErrLockDoesNotExist.Is(err) {
			return false, nil
		}

		return err == nil, err
	},
}

// getLockSubsystem returns the active lock system for the SQL engine.
func getLockSubsystem() *sql.LockSubsystem {
	engine := sqlserver.GetRunningServer().Engine
	// This should be impossible if the server was initialized correctly, but for some test harnesses we
	// take shortcuts that might invalidate that assumption
	if engine == nil {
		return nil
	}
	return engine.LS
}
