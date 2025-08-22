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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initNow registers the functions to the catalog.
func initNow() {
	framework.RegisterFunction(clock_timestamp)
	framework.RegisterFunction(now)
	framework.RegisterFunction(timeofday_)
}

// clock_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var clock_timestamp = framework.Function0{
	Name:   "clock_timestamp",
	Return: pgtypes.TimestampTZ,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		// Current date and time (changes during statement execution)
		return ctx.QueryTime(), nil
	},
}

// now represents the PostgreSQL function of the same name, taking the same parameters.
var now = framework.Function0{
	Name:   "now",
	Return: pgtypes.TimestampTZ,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		// Current date and time (start of current transaction)
		return ctx.QueryTime(), nil
	},
}

// timeofday_ represents the PostgreSQL function of the same name, taking the same parameters.
var timeofday_ = framework.Function0{
	Name:   "timeofday",
	Return: pgtypes.TimestampTZ,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		// Current date and time (like clock_timestamp, but as a text string)
		return ctx.QueryTime().Format(`Wed Feb 25 11:06:39.999999 2015 PST`), nil
	},
}
