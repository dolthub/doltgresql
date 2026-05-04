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

package functions

import (
	"math"
	"strings"
	"time"

	cerrors "github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/postgres/parser/timetz"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initExtract registers the functions to the catalog.
func initExtract() {
	framework.RegisterFunction(extract_text_date)
	framework.RegisterFunction(extract_text_time)
	framework.RegisterFunction(extract_text_timetz)
	framework.RegisterFunction(extract_text_timestamp)
	framework.RegisterFunction(extract_text_timestamptz)
	framework.RegisterFunction(extract_text_interval)
}

var ErrUnitNotSupported = errors.NewKind("unit \"%s\" not supported for type %s")

// extract_text_date represents the PostgreSQL date/time function, taking {text, date}
var extract_text_date = framework.Function2{
	Name:               "extract",
	Return:             pgtypes.Numeric,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Date},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		dateVal := val2.(time.Time)
		switch strings.ToLower(field) {
		case "hour", "hours", "microsecond", "microseconds", "millisecond", "milliseconds",
			"minute", "minutes", "second", "seconds", "timezone", "timezone_hour", "timezone_minute":
			return nil, ErrUnitNotSupported.New(field, "date")
		case "epoch":
			return pgtypes.AnyToNumeric(float64(dateVal.UnixMicro()) / 1000000)
		default:
			return getFieldFromTimeVal(field, dateVal)
		}
	},
}

// extract_text_time represents the PostgreSQL date/time function, taking {text, time without time zone}
var extract_text_time = framework.Function2{
	Name:               "extract",
	Return:             pgtypes.Numeric,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Time},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		timeVal := val2.(timeofday.TimeOfDay).ToTime()
		switch strings.ToLower(field) {
		case "century", "centuries", "day", "days", "decade", "decades", "dow", "doy",
			"isodow", "isoyear", "julian", "millennium", "millenniums", "month", "months",
			"quarter", "timezone", "timezone_hour", "timezone_minute", "week", "year", "years":
			return nil, ErrUnitNotSupported.New(field, "time without time zone")
		default:
			return getFieldFromTimeVal(field, timeVal)
		}
	},
}

// extract_text_timetz represents the PostgreSQL date/time function, taking {text, time with time zone}
var extract_text_timetz = framework.Function2{
	Name:               "extract",
	Return:             pgtypes.Numeric,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.TimeTZ},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		timetzVal := val2.(timetz.TimeTZ).ToTime()
		_, currentOffset := timetzVal.Zone()
		switch strings.ToLower(field) {
		case "century", "centuries", "day", "days", "decade", "decades", "dow", "doy",
			"isodow", "isoyear", "julian", "millennium", "millenniums", "month", "months",
			"quarter", "week", "year", "years":
			return nil, ErrUnitNotSupported.New(field, "time with time zone")
		case "timezone":
			return pgtypes.AnyToNumeric(-int64(-currentOffset))
		case "timezone_hour":
			return pgtypes.AnyToNumeric(-int64(-currentOffset / 3600))
		case "timezone_minute":
			return pgtypes.AnyToNumeric(-int64((-currentOffset % 3600) / 60))
		default:
			return getFieldFromTimeVal(field, timetzVal)
		}
	},
}

// extract_text_timestamp represents the PostgreSQL date/time function, taking {text, timestamp without time zone}
var extract_text_timestamp = framework.Function2{
	Name:               "extract",
	Return:             pgtypes.Numeric,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Timestamp},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		tsVal := val2.(time.Time)
		switch strings.ToLower(field) {
		case "timezone", "timezone_hour", "timezone_minute":
			return nil, ErrUnitNotSupported.New(field, "timestamp without time zone")
		default:
			return getFieldFromTimeVal(field, tsVal)
		}
	},
}

// extract_text_timestamptz represents the PostgreSQL date/time function, taking {text, timestamp with time zone}
var extract_text_timestamptz = framework.Function2{
	Name:               "extract",
	Return:             pgtypes.Numeric,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.TimestampTZ},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		loc, err := GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}
		tstzVal := val2.(time.Time).In(loc)
		switch strings.ToLower(field) {
		case "timezone":
			// TODO: postgres seem to use server timezone regardless of input value
			return pgtypes.AnyToNumeric(-28800)
		case "timezone_hour":
			// TODO: postgres seem to use server timezone regardless of input value
			return pgtypes.AnyToNumeric(-8)
		case "timezone_minute":
			// TODO: postgres seem to use server timezone regardless of input value
			return pgtypes.AnyToNumeric(0)
		default:
			return getFieldFromTimeVal(field, tstzVal)
		}
	},
}

const (
	NanosPerMicro = 1000
	NanosPerMilli = NanosPerMicro * duration.MicrosPerMilli
	NanosPerSec   = NanosPerMicro * duration.MicrosPerMilli * duration.MillisPerSec
)

// extract_text_interval represents the PostgreSQL date/time function, taking {text, interval}
var extract_text_interval = framework.Function2{
	Name:               "extract",
	Return:             pgtypes.Numeric,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Interval},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		dur := val2.(duration.Duration)
		var decVal decimal.Decimal
		var err error
		switch strings.ToLower(field) {
		case "century", "centuries":
			decVal = decimal.NewFromFloat(math.Floor(float64(dur.Months) / 12 / 100))
		case "day", "days":
			decVal = decimal.NewFromInt(dur.Days)
		case "decade", "decades":
			decVal = decimal.NewFromFloat(math.Floor(float64(dur.Months) / 12 / 10))
		case "epoch":
			epoch := float64(duration.SecsPerDay*duration.DaysPerMonth*dur.Months) + float64(duration.SecsPerDay*dur.Days) +
				(float64(dur.Nanos()) / (NanosPerSec))
			decVal, err = decimal.NewFromString(decimal.NewFromFloat(epoch).StringFixed(6))
		case "hour", "hours":
			hours := math.Floor(float64(dur.Nanos()) / (NanosPerSec * duration.SecsPerHour))
			decVal = decimal.NewFromFloat(hours)
		case "microsecond", "microseconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			microseconds := float64(secondsInNanos) / NanosPerMicro
			decVal = decimal.NewFromFloat(microseconds)
		case "millennium", "millenniums":
			decVal = decimal.NewFromFloat(math.Floor(float64(dur.Months) / 12 / 1000))
		case "millisecond", "milliseconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			milliseconds := float64(secondsInNanos) / NanosPerMilli
			decVal, err = decimal.NewFromString(decimal.NewFromFloat(milliseconds).StringFixed(3))
		case "minute", "minutes":
			minutesInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerHour)
			minutes := math.Floor(float64(minutesInNanos) / (NanosPerSec * duration.SecsPerMinute))
			decVal = decimal.NewFromFloat(minutes)
		case "month", "months":
			decVal = decimal.NewFromInt(dur.Months % 12)
		case "quarter":
			decVal = decimal.NewFromInt((dur.Months%12-1)/3 + 1)
		case "second", "seconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			seconds := float64(secondsInNanos) / NanosPerSec
			decVal, err = decimal.NewFromString(decimal.NewFromFloat(seconds).StringFixed(6))
		case "year", "years":
			decVal = decimal.NewFromFloat(math.Floor(float64(dur.Months) / 12))
		case "dow", "doy", "isodow", "isoyear", "julian", "timezone", "timezone_hour", "timezone_minute", "week":
			return nil, ErrUnitNotSupported.New(field, "interval")
		default:
			return nil, cerrors.Errorf("unknown field given: %s", field)
		}
		if err != nil {
			return nil, err
		}
		return pgtypes.AnyToNumeric(decVal)
	},
}

// getFieldFromTimeVal returns the value for given field extracted from non-interval values.
// It returns pgtype.Numeric value.
func getFieldFromTimeVal(field string, tVal time.Time) (pgtype.Numeric, error) {
	var dec decimal.Decimal
	var err error
	switch strings.ToLower(field) {
	case "century", "centuries":
		if year := tVal.Year(); year <= 0 {
			dec = decimal.NewFromFloat(math.Floor(float64(year-1) / 100))
		} else {
			dec = decimal.NewFromFloat(math.Ceil(float64(year) / 100))
		}
	case "day", "days":
		dec = decimal.NewFromInt(int64(tVal.Day()))
	case "decade", "decades":
		dec = decimal.NewFromFloat(math.Floor(float64(tVal.Year()) / 10))
	case "dow":
		dec = decimal.NewFromInt(int64(tVal.Weekday()))
	case "doy":
		dec = decimal.NewFromInt(int64(tVal.YearDay()))
	case "epoch":
		dec, err = decimal.NewFromString(decimal.NewFromFloat(float64(tVal.UnixMicro()) / 1000000).StringFixed(6))
	case "hour", "hours":
		dec = decimal.NewFromInt(int64(tVal.Hour()))
	case "isodow":
		wd := int64(tVal.Weekday())
		if wd == 0 {
			wd = 7
		}
		dec = decimal.NewFromInt(wd)
	case "isoyear":
		year, _ := tVal.ISOWeek()
		dec = decimal.NewFromInt(int64(year))
	case "julian":
		dec = decimal.NewFromInt(int64(date2J(tVal.Year(), int(tVal.Month()), tVal.Day())))
	case "microsecond", "microseconds", "usec", "usecs":
		w := float64(tVal.Second() * 1000000)
		f := float64(tVal.Nanosecond()) / float64(1000)
		dec = decimal.NewFromFloat(w + f)
	case "millennium", "millenniums":
		dec = decimal.NewFromFloat(math.Ceil(float64(tVal.Year()) / 1000))
	case "millisecond", "milliseconds", "msec", "msecs":
		w := float64(tVal.Second() * 1000)
		f := float64(tVal.Nanosecond()) / float64(1000000)
		dec, err = decimal.NewFromString(decimal.NewFromFloat(w + f).StringFixed(3))
	case "minute", "minutes":
		dec = decimal.NewFromInt(int64(tVal.Minute()))
	case "month", "months":
		dec = decimal.NewFromInt(int64(tVal.Month()))
	case "quarter":
		q := (int(tVal.Month())-1)/3 + 1
		dec = decimal.NewFromInt(int64(q))
	case "second", "seconds":
		w := float64(tVal.Second())
		f := float64(tVal.Nanosecond()) / float64(1000000000)
		dec, err = decimal.NewFromString(decimal.NewFromFloat(w + f).StringFixed(6))
	case "week":
		_, week := tVal.ISOWeek()
		dec = decimal.NewFromInt(int64(week))
	case "year", "years":
		dec = decimal.NewFromInt(int64(tVal.Year()))
	default:
		return pgtype.Numeric{}, cerrors.Errorf("unknown field given: %s", field)
	}
	if err != nil {
		return pgtype.Numeric{}, err
	}
	return pgtypes.AnyToNumeric(dec)
}
