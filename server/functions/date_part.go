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
	"strings"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initDatePart registers the functions to the catalog.
func initDatePart() {
	framework.RegisterFunction(date_part_text_date)
	framework.RegisterFunction(date_part_text_time)
	framework.RegisterFunction(date_part_text_timetz)
	framework.RegisterFunction(date_part_text_timestamp)
	framework.RegisterFunction(date_part_text_timestamptz)
	framework.RegisterFunction(date_part_text_interval)
}

// date_part_text_date represents the PostgreSQL date_part function for date type.
var date_part_text_date = framework.Function2{
	Name:       "date_part",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Date},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		dateVal := val2.(time.Time)
		switch strings.ToLower(field) {
		case "hour", "hours", "microsecond", "microseconds", "millisecond", "milliseconds",
			"minute", "minutes", "second", "seconds", "timezone", "timezone_hour", "timezone_minute":
			return nil, ErrUnitNotSupported.New(field, "date")
		case "epoch":
			return float64(dateVal.UnixMicro()) / 1000000, nil
		default:
			result, err := getFieldFromTimeVal(field, dateVal)
			if err != nil {
				return nil, err
			}
			f, _ := result.Float64()
			return f, nil
		}
	},
}

// date_part_text_time represents the PostgreSQL date_part function for time type.
var date_part_text_time = framework.Function2{
	Name:       "date_part",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Time},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		timeVal := val2.(time.Time)
		switch strings.ToLower(field) {
		case "century", "centuries", "day", "days", "decade", "decades", "dow", "doy",
			"isodow", "isoyear", "julian", "millennium", "millenniums", "month", "months",
			"quarter", "timezone", "timezone_hour", "timezone_minute", "week", "year", "years":
			return nil, ErrUnitNotSupported.New(field, "time without time zone")
		default:
			result, err := getFieldFromTimeVal(field, timeVal)
			if err != nil {
				return nil, err
			}
			f, _ := result.Float64()
			return f, nil
		}
	},
}

// date_part_text_timetz represents the PostgreSQL date_part function for time with time zone type.
var date_part_text_timetz = framework.Function2{
	Name:       "date_part",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.TimeTZ},
	Strict:     true,
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
			return float64(-currentOffset), nil
		case "timezone_hour":
			return float64(-currentOffset / 3600), nil
		case "timezone_minute":
			return float64((-currentOffset % 3600) / 60), nil
		default:
			result, err := getFieldFromTimeVal(field, timetzVal)
			if err != nil {
				return nil, err
			}
			f, _ := result.Float64()
			return f, nil
		}
	},
}

// date_part_text_timestamp represents the PostgreSQL date_part function for timestamp type.
var date_part_text_timestamp = framework.Function2{
	Name:       "date_part",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		tsVal := val2.(time.Time)
		switch strings.ToLower(field) {
		case "timezone", "timezone_hour", "timezone_minute":
			return nil, ErrUnitNotSupported.New(field, "timestamp without time zone")
		default:
			result, err := getFieldFromTimeVal(field, tsVal)
			if err != nil {
				return nil, err
			}
			f, _ := result.Float64()
			return f, nil
		}
	},
}

// date_part_text_timestamptz represents the PostgreSQL date_part function for timestamp with time zone type.
var date_part_text_timestamptz = framework.Function2{
	Name:       "date_part",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.TimestampTZ},
	Strict:     true,
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
			return float64(-28800), nil
		case "timezone_hour":
			// TODO: postgres seem to use server timezone regardless of input value
			return float64(-8), nil
		case "timezone_minute":
			// TODO: postgres seem to use server timezone regardless of input value
			return float64(0), nil
		default:
			result, err := getFieldFromTimeVal(field, tstzVal)
			if err != nil {
				return nil, err
			}
			f, _ := result.Float64()
			return f, nil
		}
	},
}

// date_part_text_interval represents the PostgreSQL date_part function for interval type.
var date_part_text_interval = framework.Function2{
	Name:       "date_part",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		field := val1.(string)
		dur := val2.(duration.Duration)

		// This mirrors the exact logic from extract_text_interval
		switch strings.ToLower(field) {
		case "century", "centuries":
			result := decimal.NewFromFloat(float64(dur.Months) / 12 / 100).Floor()
			f, _ := result.Float64()
			return f, nil
		case "day", "days":
			return float64(dur.Days), nil
		case "decade", "decades":
			result := decimal.NewFromFloat(float64(dur.Months) / 12 / 10).Floor()
			f, _ := result.Float64()
			return f, nil
		case "epoch":
			epoch := float64(duration.SecsPerDay*duration.DaysPerMonth*dur.Months) + float64(duration.SecsPerDay*dur.Days) +
				(float64(dur.Nanos()) / float64(NanosPerSec))
			return epoch, nil
		case "hour", "hours":
			hours := float64(dur.Nanos()) / float64(NanosPerSec*duration.SecsPerHour)
			result := decimal.NewFromFloat(hours).Floor()
			f, _ := result.Float64()
			return f, nil
		case "microsecond", "microseconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			microseconds := float64(secondsInNanos) / float64(NanosPerMicro)
			return microseconds, nil
		case "millennium", "millenniums":
			result := decimal.NewFromFloat(float64(dur.Months) / 12 / 1000).Floor()
			f, _ := result.Float64()
			return f, nil
		case "millisecond", "milliseconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			milliseconds := float64(secondsInNanos) / float64(NanosPerMilli)
			return milliseconds, nil
		case "minute", "minutes":
			minutesInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerHour)
			minutes := float64(minutesInNanos) / float64(NanosPerSec*duration.SecsPerMinute)
			result := decimal.NewFromFloat(minutes).Floor()
			f, _ := result.Float64()
			return f, nil
		case "month", "months":
			return float64(dur.Months % 12), nil
		case "quarter":
			return float64((dur.Months%12-1)/3 + 1), nil
		case "second", "seconds":
			secondsInNanos := dur.Nanos() % (NanosPerSec * duration.SecsPerMinute)
			seconds := float64(secondsInNanos) / float64(NanosPerSec)
			return seconds, nil
		case "year", "years":
			result := decimal.NewFromFloat(float64(dur.Months) / 12).Floor()
			f, _ := result.Float64()
			return f, nil
		case "dow", "doy", "isodow", "isoyear", "julian", "timezone", "timezone_hour", "timezone_minute", "week":
			return nil, ErrUnitNotSupported.New(field, "interval")
		default:
			return nil, ErrUnitNotSupported.New(field, "interval")
		}
	},
}
