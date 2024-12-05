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
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
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
			return decimal.NewFromFloat(float64(dateVal.UnixMicro()) / 1000000), nil
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
		timeVal := val2.(time.Time)
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
		timetzVal := val2.(time.Time)
		_, currentOffset := timetzVal.Zone()
		switch strings.ToLower(field) {
		case "century", "centuries", "day", "days", "decade", "decades", "dow", "doy",
			"isodow", "isoyear", "julian", "millennium", "millenniums", "month", "months",
			"quarter", "week", "year", "years":
			return nil, ErrUnitNotSupported.New(field, "time with time zone")
		case "timezone":
			return decimal.NewFromInt(-int64(-currentOffset)), nil
		case "timezone_hour":
			return decimal.NewFromInt(-int64(-currentOffset / 3600)), nil
		case "timezone_minute":
			return decimal.NewFromInt(-int64((-currentOffset % 3600) / 60)), nil
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
			return decimal.NewFromInt(-28800), nil
		case "timezone_hour":
			// TODO: postgres seem to use server timezone regardless of input value
			return decimal.NewFromInt(-8), nil
		case "timezone_minute":
			// TODO: postgres seem to use server timezone regardless of input value
			return decimal.NewFromInt(0), nil
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
		switch strings.ToLower(field) {
		case "century", "centuries":
			return decimal.NewFromFloat(math.Floor(float64(dur.Months) / 12 / 100)), nil
		case "day", "days":
			return decimal.NewFromInt(dur.Days), nil
		case "decade", "decades":
			return decimal.NewFromFloat(math.Floor(float64(dur.Months) / 12 / 10)), nil
		case "epoch":
			epoch := float64(duration.SecsPerDay*duration.DaysPerMonth*dur.Months) + float64(duration.SecsPerDay*dur.Days) +
				(float64(dur.Nanos()) / (NanosPerSec))
			return decimal.NewFromString(decimal.NewFromFloat(epoch).StringFixed(6))
		case "hour", "hours":
			hours := math.Floor(float64(dur.Nanos()) / (NanosPerSec * duration.SecsPerHour))
			return decimal.NewFromFloat(hours), nil
		case "microsecond", "microseconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			microseconds := float64(secondsInNanos) / NanosPerMicro
			return decimal.NewFromFloat(microseconds), nil
		case "millennium", "millenniums":
			return decimal.NewFromFloat(math.Floor(float64(dur.Months) / 12 / 1000)), nil
		case "millisecond", "milliseconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			milliseconds := float64(secondsInNanos) / NanosPerMilli
			return decimal.NewFromString(decimal.NewFromFloat(milliseconds).StringFixed(3))
		case "minute", "minutes":
			minutesInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerHour)
			minutes := math.Floor(float64(minutesInNanos) / (NanosPerSec * duration.SecsPerMinute))
			return decimal.NewFromFloat(minutes), nil
		case "month", "months":
			return decimal.NewFromInt(dur.Months % 12), nil
		case "quarter":
			return decimal.NewFromInt((dur.Months%12-1)/3 + 1), nil
		case "second", "seconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			seconds := float64(secondsInNanos) / NanosPerSec
			return decimal.NewFromString(decimal.NewFromFloat(seconds).StringFixed(6))
		case "year", "years":
			return decimal.NewFromFloat(math.Floor(float64(dur.Months) / 12)), nil
		case "dow", "doy", "isodow", "isoyear", "julian", "timezone", "timezone_hour", "timezone_minute", "week":
			return nil, ErrUnitNotSupported.New(field, "interval")
		default:
			return nil, fmt.Errorf("unknown field given: %s", field)
		}
	},
}

// getFieldFromTimeVal returns the value for given field extracted from non-interval values.
func getFieldFromTimeVal(field string, tVal time.Time) (decimal.Decimal, error) {
	switch strings.ToLower(field) {
	case "century", "centuries":
		if year := tVal.Year(); year <= 0 {
			return decimal.NewFromFloat(math.Floor(float64(year-1) / 100)), nil
		} else {
			return decimal.NewFromFloat(math.Ceil(float64(year) / 100)), nil
		}
	case "day", "days":
		return decimal.NewFromInt(int64(tVal.Day())), nil
	case "decade", "decades":
		return decimal.NewFromFloat(math.Floor(float64(tVal.Year()) / 10)), nil
	case "dow":
		return decimal.NewFromInt(int64(tVal.Weekday())), nil
	case "doy":
		return decimal.NewFromInt(int64(tVal.YearDay())), nil
	case "epoch":
		return decimal.NewFromString(decimal.NewFromFloat(float64(tVal.UnixMicro()) / 1000000).StringFixed(6))
	case "hour", "hours":
		return decimal.NewFromInt(int64(tVal.Hour())), nil
	case "isodow":
		wd := int64(tVal.Weekday())
		if wd == 0 {
			wd = 7
		}
		return decimal.NewFromInt(wd), nil
	case "isoyear":
		year, _ := tVal.ISOWeek()
		return decimal.NewFromInt(int64(year)), nil
	case "julian":
		return decimal.Decimal{}, fmt.Errorf("'julian' field extraction not supported yet")
	case "microsecond", "microseconds":
		w := float64(tVal.Second() * 1000000)
		f := float64(tVal.Nanosecond()) / float64(1000)
		return decimal.NewFromFloat(w + f), nil
	case "millennium", "millenniums":
		return decimal.NewFromFloat(math.Ceil(float64(tVal.Year()) / 1000)), nil
	case "millisecond", "milliseconds":
		w := float64(tVal.Second() * 1000)
		f := float64(tVal.Nanosecond()) / float64(1000000)
		return decimal.NewFromString(decimal.NewFromFloat(w + f).StringFixed(3))
	case "minute", "minutes":
		return decimal.NewFromInt(int64(tVal.Minute())), nil
	case "month", "months":
		return decimal.NewFromInt(int64(tVal.Month())), nil
	case "quarter":
		q := (int(tVal.Month())-1)/3 + 1
		return decimal.NewFromInt(int64(q)), nil
	case "second", "seconds":
		w := float64(tVal.Second())
		f := float64(tVal.Nanosecond()) / float64(1000000000)
		return decimal.NewFromString(decimal.NewFromFloat(w + f).StringFixed(6))
	case "week":
		_, week := tVal.ISOWeek()
		return decimal.NewFromInt(int64(week)), nil
	case "year", "years":
		return decimal.NewFromInt(int64(tVal.Year())), nil
	default:
		return decimal.Decimal{}, fmt.Errorf("unknown field given: %s", field)
	}
}
