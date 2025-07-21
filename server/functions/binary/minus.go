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
	"math"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '-' ORDER BY o.oprcode::varchar;

// initBinaryMinus registers the functions to the catalog.
func initBinaryMinus() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, date_mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, date_mii)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, date_mi_interval)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, float4mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, float48mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, float8mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, float84mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int2mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int24mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int28mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int4mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int42mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int48mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int8mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int82mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int84mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, interval_mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, numeric_sub)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, time_mi_interval)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, time_mi_time)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, timetz_mi_interval)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, timestamp_mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, timestamp_mi_interval)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, timestamptz_mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, timestamptz_mi_interval)
}

// float4mi represents the PostgreSQL function of the same name, taking the same parameters.
var float4mi = framework.Function2{
	Name:       "float4mi",
	Return:     pgtypes.Float32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float32) - val2.(float32), nil
	},
}

// float48mi represents the PostgreSQL function of the same name, taking the same parameters.
var float48mi = framework.Function2{
	Name:       "float48mi",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return float64(val1.(float32)) - val2.(float64), nil
	},
}

// float8mi represents the PostgreSQL function of the same name, taking the same parameters.
var float8mi = framework.Function2{
	Name:       "float8mi",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float64) - val2.(float64), nil
	},
}

// float84mi represents the PostgreSQL function of the same name, taking the same parameters.
var float84mi = framework.Function2{
	Name:       "float84mi",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float64) - float64(val2.(float32)), nil
	},
}

// int2mi represents the PostgreSQL function of the same name, taking the same parameters.
var int2mi = framework.Function2{
	Name:       "int2mi",
	Return:     pgtypes.Int16,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int16)) - int64(val2.(int16))
		if result > math.MaxInt16 || result < math.MinInt16 {
			return nil, errors.Errorf("smallint out of range")
		}
		return int16(result), nil
	},
}

// int24mi represents the PostgreSQL function of the same name, taking the same parameters.
var int24mi = framework.Function2{
	Name:       "int24mi",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int16)) - int64(val2.(int32))
		if result > math.MaxInt16 || result < math.MinInt16 {
			return nil, errors.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int28mi represents the PostgreSQL function of the same name, taking the same parameters.
var int28mi = framework.Function2{
	Name:       "int28mi",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return minusOverflow(int64(val1.(int16)), val2.(int64))
	},
}

// int4mi represents the PostgreSQL function of the same name, taking the same parameters.
var int4mi = framework.Function2{
	Name:       "int4mi",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int32)) - int64(val2.(int32))
		if result > math.MaxInt32 || result < math.MinInt32 {
			return nil, errors.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int42mi represents the PostgreSQL function of the same name, taking the same parameters.
var int42mi = framework.Function2{
	Name:       "int42mi",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int32)) - int64(val2.(int16))
		if result > math.MaxInt32 || result < math.MinInt32 {
			return nil, errors.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int48mi represents the PostgreSQL function of the same name, taking the same parameters.
var int48mi = framework.Function2{
	Name:       "int48mi",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return minusOverflow(int64(val1.(int32)), val2.(int64))
	},
}

// int8mi represents the PostgreSQL function of the same name, taking the same parameters.
var int8mi = framework.Function2{
	Name:       "int8mi",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return minusOverflow(val1.(int64), val2.(int64))
	},
}

// int82mi represents the PostgreSQL function of the same name, taking the same parameters.
var int82mi = framework.Function2{
	Name:       "int82mi",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return minusOverflow(val1.(int64), int64(val2.(int16)))
	},
}

// int84mi represents the PostgreSQL function of the same name, taking the same parameters.
var int84mi = framework.Function2{
	Name:       "int84mi",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return minusOverflow(val1.(int64), int64(val2.(int32)))
	},
}

// interval_mi represents the PostgreSQL function of the same name, taking the same parameters.
var interval_mi = framework.Function2{
	Name:       "interval_mi",
	Return:     pgtypes.Interval,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		dur1 := val1.(duration.Duration)
		dur2 := val2.(duration.Duration)
		return dur1.Sub(dur2), nil
	},
}

// numeric_sub represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_sub = framework.Function2{
	Name:       "numeric_sub",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(decimal.Decimal).Sub(val2.(decimal.Decimal)), nil
	},
}

// date_mi represents the PostgreSQL function of the same name, taking the same parameters.
var date_mi = framework.Function2{
	Name:       "date_mi",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		date1 := val1.(time.Time)
		date2 := val2.(time.Time)

		// Calculate the difference in days
		duration := date1.Sub(date2)
		days := int32(duration.Hours() / 24)

		return days, nil
	},
}

// date_mii represents the PostgreSQL function of the same name, taking the same parameters.
var date_mii = framework.Function2{
	Name:       "date_mii",
	Return:     pgtypes.Date,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		date := val1.(time.Time)
		days := val2.(int32)

		// Subtract the specified number of days from the date
		result := date.AddDate(0, 0, -int(days))

		return result, nil
	},
}

// date_mi_interval represents the PostgreSQL function of the same name, taking the same parameters.
var date_mi_interval = framework.Function2{
	Name:       "date_mi_interval",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		date := val1.(time.Time)
		interval := val2.(duration.Duration)
		seconds, ok := interval.AsInt64()
		if !ok {
			return nil, errors.New("overflown interval")
		}
		// above truncates partial seconds.
		nanos := seconds*duration.NanosPerMicro*duration.MicrosPerMilli*duration.MillisPerSec + interval.Nanos()%int64(time.Second)
		// Subtract the interval from the date using negative duration
		result := date.Add(-time.Duration(nanos))

		return result, nil
	},
}

// timestamptz_mi represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_mi = framework.Function2{
	Name:       "timestamptz_mi",
	Return:     pgtypes.Interval,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		ts1 := val1.(time.Time)
		ts2 := val2.(time.Time)

		// Calculate the difference and return as interval
		diff := ts1.Sub(ts2)
		return duration.MakeDuration(diff.Nanoseconds(), 0, 0), nil
	},
}

// timestamptz_mi_interval represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_mi_interval = framework.Function2{
	Name:       "timestamptz_mi_interval",
	Return:     pgtypes.TimestampTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timestamptz := val1.(time.Time)
		interval := val2.(duration.Duration)

		// Subtract the interval from the timestamptz
		return timestamptz.Add(-time.Duration(interval.Nanos())), nil
	},
}

// timestamp_mi represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_mi = framework.Function2{
	Name:       "timestamp_mi",
	Return:     pgtypes.Interval,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		ts1 := val1.(time.Time)
		ts2 := val2.(time.Time)

		// Calculate the difference and return as interval
		diff := ts1.Sub(ts2)
		return duration.MakeDuration(diff.Nanoseconds(), 0, 0), nil
	},
}

// timestamp_mi_interval represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_mi_interval = framework.Function2{
	Name:       "timestamp_mi_interval",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timestamp := val1.(time.Time)
		interval := val2.(duration.Duration)

		// Subtract the interval from the timestamp
		return timestamp.Add(-time.Duration(interval.Nanos())), nil
	},
}

// time_mi_time represents the PostgreSQL function of the same name, taking the same parameters.
var time_mi_time = framework.Function2{
	Name:       "time_mi_time",
	Return:     pgtypes.Interval,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Time, pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		time1 := val1.(time.Time)
		time2 := val2.(time.Time)

		// Calculate the difference and return as interval
		diff := time1.Sub(time2)
		return duration.MakeDuration(diff.Nanoseconds(), 0, 0), nil
	},
}

// time_mi_interval represents the PostgreSQL function of the same name, taking the same parameters.
var time_mi_interval = framework.Function2{
	Name:       "time_mi_interval",
	Return:     pgtypes.Time,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Time, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timeVal := val1.(time.Time)
		interval := val2.(duration.Duration)

		// Subtract the interval from the time
		return timeVal.Add(-time.Duration(interval.Nanos())), nil
	},
}

// timetz_mi_interval represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_mi_interval = framework.Function2{
	Name:       "timetz_mi_interval",
	Return:     pgtypes.TimeTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		interval := val1.(duration.Duration)
		timetzVal := val2.(time.Time)

		// Subtract the interval from the timetz (note: parameters are reversed)
		return timetzVal.Add(-time.Duration(interval.Nanos())), nil
	},
}

// minusOverflow is a convenience function that checks for overflow for int64 subtraction.
func minusOverflow(val1 int64, val2 int64) (any, error) {
	if val2 > 0 {
		if val1 < math.MinInt64+val2 {
			return nil, errors.Errorf("bigint out of range")
		}
	} else {
		if val1 > math.MaxInt64+val2 {
			return nil, errors.Errorf("bigint out of range")
		}
	}
	return val1 - val2, nil
}
