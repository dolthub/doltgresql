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

package binary

import (
	"fmt"
	"math"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '+' ORDER BY o.oprcode::varchar;

// initBinaryPlus registers the functions to the catalog.
func initBinaryPlus() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, float4pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, float48pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, float8pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, float84pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int2pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int24pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int28pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int4pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int42pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int48pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int8pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int82pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int84pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_time)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_timetz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, numeric_add)
}

// float4pl represents the PostgreSQL function of the same name, taking the same parameters.
var float4pl = framework.Function2{
	Name:       "float4pl",
	Return:     pgtypes.Float32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float32) + val2.(float32), nil
	},
}

// float48pl represents the PostgreSQL function of the same name, taking the same parameters.
var float48pl = framework.Function2{
	Name:       "float48pl",
	Return:     pgtypes.Float64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return float64(val1.(float32)) + val2.(float64), nil
	},
}

// float8pl represents the PostgreSQL function of the same name, taking the same parameters.
var float8pl = framework.Function2{
	Name:       "float8pl",
	Return:     pgtypes.Float64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float64) + val2.(float64), nil
	},
}

// float84pl represents the PostgreSQL function of the same name, taking the same parameters.
var float84pl = framework.Function2{
	Name:       "float84pl",
	Return:     pgtypes.Float64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float64) + float64(val2.(float32)), nil
	},
}

// int2pl represents the PostgreSQL function of the same name, taking the same parameters.
var int2pl = framework.Function2{
	Name:       "int2pl",
	Return:     pgtypes.Int16,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int16)) + int64(val2.(int16))
		if result > math.MaxInt16 || result < math.MinInt16 {
			return nil, fmt.Errorf("smallint out of range")
		}
		return int16(result), nil
	},
}

// int24pl represents the PostgreSQL function of the same name, taking the same parameters.
var int24pl = framework.Function2{
	Name:       "int24pl",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int16)) + int64(val2.(int32))
		if result > math.MaxInt16 || result < math.MinInt16 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int28pl represents the PostgreSQL function of the same name, taking the same parameters.
var int28pl = framework.Function2{
	Name:       "int28pl",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return plusOverflow(int64(val1.(int16)), val2.(int64))
	},
}

// int4pl represents the PostgreSQL function of the same name, taking the same parameters.
var int4pl = framework.Function2{
	Name:       "int4pl",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int32)) + int64(val2.(int32))
		if result > math.MaxInt32 || result < math.MinInt32 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int42pl represents the PostgreSQL function of the same name, taking the same parameters.
var int42pl = framework.Function2{
	Name:       "int42pl",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int32)) + int64(val2.(int16))
		if result > math.MaxInt32 || result < math.MinInt32 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int48pl represents the PostgreSQL function of the same name, taking the same parameters.
var int48pl = framework.Function2{
	Name:       "int48pl",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return plusOverflow(int64(val1.(int32)), val2.(int64))
	},
}

// int8pl represents the PostgreSQL function of the same name, taking the same parameters.
var int8pl = framework.Function2{
	Name:       "int8pl",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return plusOverflow(val1.(int64), val2.(int64))
	},
}

// int82pl represents the PostgreSQL function of the same name, taking the same parameters.
var int82pl = framework.Function2{
	Name:       "int82pl",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return plusOverflow(val1.(int64), int64(val2.(int16)))
	},
}

// int84pl represents the PostgreSQL function of the same name, taking the same parameters.
var int84pl = framework.Function2{
	Name:       "int84pl",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return plusOverflow(val1.(int64), int64(val2.(int32)))
	},
}

// interval_pl represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl = framework.Function2{
	Name:       "interval_pl",
	Return:     pgtypes.Interval,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		dur1 := val1.(duration.Duration)
		dur2 := val2.(duration.Duration)
		return dur1.Add(dur2), nil
	},
}

// interval_pl_time represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_time = framework.Function2{
	Name:       "interval_pl_time",
	Return:     pgtypes.Time,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return intervalPlusNonInterval(val1.(duration.Duration), val2.(time.Time))
	},
}

// interval_pl_date represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_date = framework.Function2{
	Name:       "interval_pl_date",
	Return:     pgtypes.Timestamp,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return intervalPlusNonInterval(val1.(duration.Duration), val2.(time.Time))
	},
}

// interval_pl_timetz represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_timetz = framework.Function2{
	Name:       "interval_pl_timetz",
	Return:     pgtypes.TimeTZ,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return intervalPlusNonInterval(val1.(duration.Duration), val2.(time.Time))
	},
}

// interval_pl_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_timestamp = framework.Function2{
	Name:       "interval_pl_timestamp",
	Return:     pgtypes.Timestamp,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return intervalPlusNonInterval(val1.(duration.Duration), val2.(time.Time))
	},
}

// interval_pl_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_timestamptz = framework.Function2{
	Name:       "interval_pl_timestamptz",
	Return:     pgtypes.TimestampTZ,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return intervalPlusNonInterval(val1.(duration.Duration), val2.(time.Time))
	},
}

// numeric_add represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_add = framework.Function2{
	Name:       "numeric_add",
	Return:     pgtypes.Numeric,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(decimal.Decimal).Add(val2.(decimal.Decimal)), nil
	},
}

// plusOverflow is a convenience function that checks for overflow for int64 addition.
func plusOverflow(val1 int64, val2 int64) (any, error) {
	if val2 > 0 {
		if val1 > math.MaxInt64-val2 {
			return nil, fmt.Errorf("bigint out of range")
		}
	} else {
		if val1 < math.MinInt64-val2 {
			return nil, fmt.Errorf("bigint out of range")
		}
	}
	return val1 + val2, nil
}

// intervalPlusNonInterval adds given interval duration to the given time.Time value.
// During converting interval duration to time.Duration type, it can overflow.
func intervalPlusNonInterval(d duration.Duration, t time.Time) (time.Time, error) {
	seconds, ok := d.AsInt64()
	if !ok {
		if !ok {
			return time.Time{}, fmt.Errorf("interval overflow")
		}
	}
	nanos := float64(seconds) * functions.NanosPerSec
	if nanos > float64(math.MaxInt64) || nanos < float64(math.MinInt64) {
		return time.Time{}, fmt.Errorf("interval overflow")
	}
	return t.Add(time.Duration(nanos)), nil
}
