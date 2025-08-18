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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initCurrentTime registers the functions to the catalog.
func initCurrentTime() {
	framework.RegisterFunction(clock_timestamp)
	framework.RegisterFunction(current_date)
	framework.RegisterFunction(current_time)
	framework.RegisterFunction(current_time_int32)
	framework.RegisterFunction(current_timestamp)
	framework.RegisterFunction(current_timestamp_int32)
	framework.RegisterFunction(localtime)
	framework.RegisterFunction(localtime_int32)
	framework.RegisterFunction(localtimestamp)
	framework.RegisterFunction(localtimestamp_int32)
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

// current_date represents the PostgreSQL function of the same name, taking the same parameters.
var current_date = framework.Function0{
	Name:   "current_date",
	Return: pgtypes.Date,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		qt := ctx.QueryTime()
		year, month, day := qt.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, qt.Location()), nil
	},
}

// current_time represents the PostgreSQL function of the same name, taking the same parameters.
var current_time = framework.Function0{
	Name:   "current_time",
	Return: pgtypes.TimeTZ,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		return ctx.QueryTime(), nil
	},
}

// current_time_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var current_time_int32 = framework.Function1{
	Name:       "current_time",
	Return:     pgtypes.TimeTZ,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: support precision
		return ctx.QueryTime(), nil
	},
}

// current_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var current_timestamp = framework.Function0{
	Name:   "current_timestamp",
	Return: pgtypes.TimestampTZ,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		return ctx.QueryTime(), nil
	},
}

// current_timestamp_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var current_timestamp_int32 = framework.Function1{
	Name:       "current_timestamp",
	Return:     pgtypes.TimestampTZ,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: support precision
		return ctx.QueryTime(), nil
	},
}

// localtime represents the PostgreSQL function of the same name, taking the same parameters.
var localtime = framework.Function0{
	Name:   "localtime",
	Return: pgtypes.Timestamp,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		// Current date and time (start of current transaction)
		return ctx.QueryTime(), nil
	},
}

// localtime_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var localtime_int32 = framework.Function1{
	Name:       "localtime",
	Return:     pgtypes.Timestamp,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// Current date and time (start of current transaction)
		return ctx.QueryTime(), nil
	},
}

// localtimestamp represents the PostgreSQL function of the same name, taking the same parameters.
var localtimestamp = framework.Function0{
	Name:   "localtimestamp",
	Return: pgtypes.Timestamp,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		// Current date and time (start of current transaction)
		return ctx.QueryTime(), nil
	},
}

// localtimestamp_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var localtimestamp_int32 = framework.Function1{
	Name:       "localtimestamp",
	Return:     pgtypes.Timestamp,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// Current date and time (start of current transaction)
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
