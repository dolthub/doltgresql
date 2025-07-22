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
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '+' ORDER BY o.oprcode::varchar;

// initBinaryPlus registers the functions to the catalog.
func initBinaryPlus() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, date_pl_interval)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, date_pli)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, datetime_pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, datetimetz_pl)
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
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, integer_pl_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_time)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_date)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_timetz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_timestamp)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, interval_pl_timestamptz)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, numeric_add)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, time_pl_interval)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, timedate_pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, timetz_pl_interval)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, timetzdate_pl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, timestamp_pl_interval)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, timestamptz_pl_interval)
}

// float4pl_callable is the callable logic for the float4pl function.
func float4pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return val1.(float32) + val2.(float32), nil
}

// float4pl represents the PostgreSQL function of the same name, taking the same parameters.
var float4pl = framework.Function2{
	Name:       "float4pl",
	Return:     pgtypes.Float32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable:   float4pl_callable,
}

// float48pl_callable is the callable logic for the float48pl function.
func float48pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return float64(val1.(float32)) + val2.(float64), nil
}

// float48pl represents the PostgreSQL function of the same name, taking the same parameters.
var float48pl = framework.Function2{
	Name:       "float48pl",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable:   float48pl_callable,
}

// float8pl_callable is the callable logic for the float8pl function.
func float8pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return val1.(float64) + val2.(float64), nil
}

// float8pl represents the PostgreSQL function of the same name, taking the same parameters.
var float8pl = framework.Function2{
	Name:       "float8pl",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable:   float8pl_callable,
}

// float84pl_callable is the callable logic for the float84pl function.
func float84pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return val1.(float64) + float64(val2.(float32)), nil
}

// float84pl represents the PostgreSQL function of the same name, taking the same parameters.
var float84pl = framework.Function2{
	Name:       "float84pl",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable:   float84pl_callable,
}

// int2pl_callable is the callable logic for the int2pl function.
func int2pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	result := int64(val1.(int16)) + int64(val2.(int16))
	if result > math.MaxInt16 || result < math.MinInt16 {
		return nil, errors.Errorf("smallint out of range")
	}
	return int16(result), nil
}

// int2pl represents the PostgreSQL function of the same name, taking the same parameters.
var int2pl = framework.Function2{
	Name:       "int2pl",
	Return:     pgtypes.Int16,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable:   int2pl_callable,
}

// int24pl_callable is the callable logic for the int24pl function.
func int24pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	result := int64(val1.(int16)) + int64(val2.(int32))
	if result > math.MaxInt16 || result < math.MinInt16 {
		return nil, errors.Errorf("integer out of range")
	}
	return int32(result), nil
}

// int24pl represents the PostgreSQL function of the same name, taking the same parameters.
var int24pl = framework.Function2{
	Name:       "int24pl",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable:   int24pl_callable,
}

// int28pl_callable is the callable logic for the int28pl function.
func int28pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return plusOverflow(int64(val1.(int16)), val2.(int64))
}

// int28pl represents the PostgreSQL function of the same name, taking the same parameters.
var int28pl = framework.Function2{
	Name:       "int28pl",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable:   int28pl_callable,
}

// int4pl_callable is the callable logic for the int4pl function.
func int4pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	result := int64(val1.(int32)) + int64(val2.(int32))
	if result > math.MaxInt32 || result < math.MinInt32 {
		return nil, errors.Errorf("integer out of range")
	}
	return int32(result), nil
}

// int4pl represents the PostgreSQL function of the same name, taking the same parameters.
var int4pl = framework.Function2{
	Name:       "int4pl",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable:   int4pl_callable,
}

// int42pl_callable is the callable logic for the int42pl function.
func int42pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	result := int64(val1.(int32)) + int64(val2.(int16))
	if result > math.MaxInt32 || result < math.MinInt32 {
		return nil, errors.Errorf("integer out of range")
	}
	return int32(result), nil
}

// int42pl represents the PostgreSQL function of the same name, taking the same parameters.
var int42pl = framework.Function2{
	Name:       "int42pl",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable:   int42pl_callable,
}

// int48pl_callable is the callable logic for the int48pl function.
func int48pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return plusOverflow(int64(val1.(int32)), val2.(int64))
}

// int48pl represents the PostgreSQL function of the same name, taking the same parameters.
var int48pl = framework.Function2{
	Name:       "int48pl",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable:   int48pl_callable,
}

// int8pl_callable is the callable logic for the int8pl function.
func int8pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return plusOverflow(val1.(int64), val2.(int64))
}

// int8pl represents the PostgreSQL function of the same name, taking the same parameters.
var int8pl = framework.Function2{
	Name:       "int8pl",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable:   int8pl_callable,
}

// int82pl_callable is the callable logic for the int82pl function.
func int82pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return plusOverflow(val1.(int64), int64(val2.(int16)))
}

// int82pl represents the PostgreSQL function of the same name, taking the same parameters.
var int82pl = framework.Function2{
	Name:       "int82pl",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable:   int82pl_callable,
}

// int84pl_callable is the callable logic for the int84pl function.
func int84pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return plusOverflow(val1.(int64), int64(val2.(int32)))
}

// int84pl represents the PostgreSQL function of the same name, taking the same parameters.
var int84pl = framework.Function2{
	Name:       "int84pl",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable:   int84pl_callable,
}

// integer_pl_date represents the PostgreSQL function of the same name, taking the same parameters.
var integer_pl_date = framework.Function2{
	Name:       "integer_pl_date",
	Return:     pgtypes.Date,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		days := val1.(int32)
		date := val2.(time.Time)

		// Add the specified number of days to the date (reverse of date_pli)
		result := date.AddDate(0, 0, int(days))

		return result, nil
	},
}

// interval_pl_callable is the callable logic for the interval_pl function.
func interval_pl_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	dur1 := val1.(duration.Duration)
	dur2 := val2.(duration.Duration)
	return dur1.Add(dur2), nil
}

// interval_pl represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl = framework.Function2{
	Name:       "interval_pl",
	Return:     pgtypes.Interval,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Interval},
	Strict:     true,
	Callable:   interval_pl_callable,
}

// interval_pl_time_callable is the callable logic for the interval_pl_time function.
func interval_pl_time_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return val2.(time.Time).Add(time.Duration(val1.(duration.Duration).Nanos())), nil
}

// interval_pl_time represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_time = framework.Function2{
	Name:       "interval_pl_time",
	Return:     pgtypes.Time,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Time},
	Strict:     true,
	Callable:   interval_pl_time_callable,
}

// interval_pl_date_callable is the callable logic for the interval_pl_date function.
func interval_pl_date_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return intervalPlusNonInterval(val1.(duration.Duration), val2.(time.Time))
}

// interval_pl_date represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_date = framework.Function2{
	Name:       "interval_pl_date",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Date},
	Strict:     true,
	Callable:   interval_pl_date_callable,
}

// interval_pl_timetz_callable is the callable logic for the interval_pl_timetz function.
func interval_pl_timetz_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	ttz := val2.(time.Time)
	return ttz.Add(time.Duration(val1.(duration.Duration).Nanos())), nil
}

// interval_pl_timetz represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_timetz = framework.Function2{
	Name:       "interval_pl_timetz",
	Return:     pgtypes.TimeTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.TimeTZ},
	Strict:     true,
	Callable:   interval_pl_timetz_callable,
}

// interval_pl_timestamp_callable is the callable logic for the interval_pl_timestamp function.
func interval_pl_timestamp_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return intervalPlusNonInterval(val1.(duration.Duration), val2.(time.Time))
}

// interval_pl_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_timestamp = framework.Function2{
	Name:       "interval_pl_timestamp",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Timestamp},
	Strict:     true,
	Callable:   interval_pl_timestamp_callable,
}

// interval_pl_timestamptz_callable is the callable logic for the interval_pl_timestamptz function.
func interval_pl_timestamptz_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return intervalPlusNonInterval(val1.(duration.Duration), val2.(time.Time))
}

// interval_pl_timestamptz represents the PostgreSQL function of the same name, taking the same parameters.
var interval_pl_timestamptz = framework.Function2{
	Name:       "interval_pl_timestamptz",
	Return:     pgtypes.TimestampTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.TimestampTZ},
	Strict:     true,
	Callable:   interval_pl_timestamptz_callable,
}

// numeric_add_callable is the callable logic for the numeric_add function.
func numeric_add_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return val1.(decimal.Decimal).Add(val2.(decimal.Decimal)), nil
}

// numeric_add represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_add = framework.Function2{
	Name:       "numeric_add",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable:   numeric_add_callable,
}

// plusOverflow is a convenience function that checks for overflow for int64 addition.
func plusOverflow(val1 int64, val2 int64) (any, error) {
	if val2 > 0 {
		if val1 > math.MaxInt64-val2 {
			return nil, errors.Errorf("bigint out of range")
		}
	} else {
		if val1 < math.MinInt64-val2 {
			return nil, errors.Errorf("bigint out of range")
		}
	}
	return val1 + val2, nil
}

// date_pl_interval represents the PostgreSQL function of the same name, taking the same parameters.
var date_pl_interval = framework.Function2{
	Name:       "date_pl_interval",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		date := val1.(time.Time)
		interval := val2.(duration.Duration)

		// Add the interval to the date using the existing helper function
		return intervalPlusNonInterval(interval, date)
	},
}

// date_pli represents the PostgreSQL function of the same name, taking the same parameters.
var date_pli = framework.Function2{
	Name:       "date_pli",
	Return:     pgtypes.Date,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		date := val1.(time.Time)
		days := val2.(int32)

		// Add the specified number of days to the date
		result := date.AddDate(0, 0, int(days))

		return result, nil
	},
}

// datetime_pl represents the PostgreSQL function of the same name, taking the same parameters.
var datetime_pl = framework.Function2{
	Name:       "datetime_pl",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		date := val1.(time.Time)
		timeVal := val2.(time.Time)

		// Combine date from first parameter with time from second parameter
		// Extract hour, minute, second, nanosecond from time
		hour, min, sec := timeVal.Clock()
		nsec := timeVal.Nanosecond()

		// Create new timestamp with date components from date and time components from time
		result := time.Date(date.Year(), date.Month(), date.Day(), hour, min, sec, nsec, date.Location())

		return result, nil
	},
}

// datetimetz_pl represents the PostgreSQL function of the same name, taking the same parameters.
var datetimetz_pl = framework.Function2{
	Name:       "datetimetz_pl",
	Return:     pgtypes.TimestampTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Date, pgtypes.TimeTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		date := val1.(time.Time)
		timetzVal := val2.(time.Time)

		// Combine date from first parameter with time+timezone from second parameter
		// Extract hour, minute, second, nanosecond, and timezone from timetz
		hour, min, sec := timetzVal.Clock()
		nsec := timetzVal.Nanosecond()
		location := timetzVal.Location()

		// Create new timestamptz with date components from date and time+timezone components from timetz
		result := time.Date(date.Year(), date.Month(), date.Day(), hour, min, sec, nsec, location)

		return result, nil
	},
}

// timedate_pl represents the PostgreSQL function of the same name, taking the same parameters.
var timedate_pl = framework.Function2{
	Name:       "timedate_pl",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Time, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timeVal := val1.(time.Time)
		date := val2.(time.Time)

		// Combine time from first parameter with date from second parameter
		// Extract hour, minute, second, nanosecond from time
		hour, min, sec := timeVal.Clock()
		nsec := timeVal.Nanosecond()

		// Create new timestamp with time components from time and date components from date
		result := time.Date(date.Year(), date.Month(), date.Day(), hour, min, sec, nsec, date.Location())

		return result, nil
	},
}

// timetzdate_pl represents the PostgreSQL function of the same name, taking the same parameters.
var timetzdate_pl = framework.Function2{
	Name:       "timetzdate_pl",
	Return:     pgtypes.TimestampTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timetzVal := val1.(time.Time)
		date := val2.(time.Time)

		// Combine timetz from first parameter with date from second parameter
		// Extract hour, minute, second, nanosecond, and timezone from timetz
		hour, min, sec := timetzVal.Clock()
		nsec := timetzVal.Nanosecond()
		location := timetzVal.Location()

		// Create new timestamptz with time+timezone components from timetz and date components from date
		result := time.Date(date.Year(), date.Month(), date.Day(), hour, min, sec, nsec, location)

		return result, nil
	},
}

// timestamp_pl_interval represents the PostgreSQL function of the same name, taking the same parameters.
var timestamp_pl_interval = framework.Function2{
	Name:       "timestamp_pl_interval",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timestamp := val1.(time.Time)
		interval := val2.(duration.Duration)

		// Add the interval to the timestamp using the existing helper function
		return intervalPlusNonInterval(interval, timestamp)
	},
}

// timestamptz_pl_interval represents the PostgreSQL function of the same name, taking the same parameters.
var timestamptz_pl_interval = framework.Function2{
	Name:       "timestamptz_pl_interval",
	Return:     pgtypes.TimestampTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timestamptz := val1.(time.Time)
		interval := val2.(duration.Duration)

		// Add the interval to the timestamptz using the existing helper function
		return intervalPlusNonInterval(interval, timestamptz)
	},
}

// time_pl_interval represents the PostgreSQL function of the same name, taking the same parameters.
var time_pl_interval = framework.Function2{
	Name:       "time_pl_interval",
	Return:     pgtypes.Time,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Time, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timeVal := val1.(time.Time)
		interval := val2.(duration.Duration)

		// Add the interval to the time
		// Convert interval to duration and add to time
		return timeVal.Add(time.Duration(interval.Nanos())), nil
	},
}

// timetz_pl_interval represents the PostgreSQL function of the same name, taking the same parameters.
var timetz_pl_interval = framework.Function2{
	Name:       "timetz_pl_interval",
	Return:     pgtypes.TimeTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimeTZ, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		timetzVal := val1.(time.Time)
		interval := val2.(duration.Duration)

		// Add the interval to the timetz
		// Convert interval to duration and add to timetz
		return timetzVal.Add(time.Duration(interval.Nanos())), nil
	},
}

// intervalPlusNonInterval adds given interval duration to the given time.Time value.
// During converting interval duration to time.Duration type, it can overflow.
func intervalPlusNonInterval(d duration.Duration, t time.Time) (time.Time, error) {
	seconds, ok := d.AsInt64()
	if !ok {
		if !ok {
			return time.Time{}, errors.Errorf("interval overflow")
		}
	}
	nanos := float64(seconds) * functions.NanosPerSec
	if nanos > float64(math.MaxInt64) || nanos < float64(math.MinInt64) {
		return time.Time{}, errors.Errorf("interval overflow")
	}
	return t.Add(time.Duration(nanos)), nil
}
