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
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgSleep registers the functions to the catalog.
func initPgSleep() {
	framework.RegisterFunction(pg_sleep_float64)
	framework.RegisterFunction(pg_sleep_for_interval)
}

// pg_sleep_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var pg_sleep_float64 = framework.Function1{
	Name:               "pg_sleep",
	Return:             pgtypes.Void,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Float64},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		f := val.(float64)
		time.Sleep(time.Duration(f * float64(time.Second)))
		return nil, nil
	},
}

// pg_sleep_for_interval represents the PostgreSQL function of the same name, taking the same parameters.
var pg_sleep_for_interval = framework.Function1{
	Name:               "pg_sleep_for",
	Return:             pgtypes.Void,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Interval},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		d := val.(duration.Duration)
		time.Sleep(time.Duration(d.Nanos()))
		return nil, nil
	},
}
