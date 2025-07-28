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

	cerrors "github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initDateTrunc registers the functions to the catalog.
func initDateTrunc() {
	framework.RegisterFunction(date_trunc_text_timestamp)
	framework.RegisterFunction(date_trunc_text_timestamptz)
	framework.RegisterFunction(date_trunc_text_timestamptz_text)
	framework.RegisterFunction(date_trunc_text_interval)
}

// date_trunc_text_timestamp represents the PostgreSQL date_trunc function for timestamp type.
var date_trunc_text_timestamp = framework.Function2{
	Name:       "date_trunc",
	Return:     pgtypes.Timestamp,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		unit := val1.(string)
		ts := val2.(time.Time)
		return truncateTime(unit, ts)
	},
}

// date_trunc_text_timestamptz represents the PostgreSQL date_trunc function for timestamptz type.
var date_trunc_text_timestamptz = framework.Function2{
	Name:       "date_trunc",
	Return:     pgtypes.TimestampTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		unit := val1.(string)
		ts := val2.(time.Time)
		// Convert to server location for consistent behavior
		loc, err := GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}
		localTs := ts.In(loc)
		truncated, err := truncateTime(unit, localTs)
		if err != nil {
			return nil, err
		}
		return truncated, nil
	},
}

// date_trunc_text_timestamptz_text represents the PostgreSQL date_trunc function with timezone parameter.
var date_trunc_text_timestamptz_text = framework.Function3{
	Name:       "date_trunc",
	Return:     pgtypes.TimestampTZ,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.TimestampTZ, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		unit := val1.(string)
		ts := val2.(time.Time)
		timezone := val3.(string)

		// Convert timezone string to offset
		newOffset, err := convertTzToOffsetSecs(timezone)
		if err != nil {
			return nil, err
		}

		// Create a location with the specified offset
		loc := time.FixedZone("", int(newOffset))

		// Convert timestamp to the specified timezone
		tsInTz := ts.In(loc)

		// Truncate in the specified timezone
		truncated, err := truncateTime(unit, tsInTz)
		if err != nil {
			return nil, err
		}

		// Convert back to server timezone
		serverLoc, err := GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}
		return truncated.In(serverLoc), nil
	},
}

// date_trunc_text_interval represents the PostgreSQL date_trunc function for interval type.
var date_trunc_text_interval = framework.Function2{
	Name:       "date_trunc",
	Return:     pgtypes.Interval,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		unit := val1.(string)
		interval := val2.(duration.Duration)
		return truncateInterval(unit, interval)
	},
}

// truncateTime truncates a time value to the specified unit.
func truncateTime(unit string, t time.Time) (time.Time, error) {
	switch strings.ToLower(unit) {
	case "microsecond", "microseconds":
		// Truncate to microseconds - remove nanoseconds beyond microseconds
		return t.Truncate(time.Microsecond), nil
	case "millisecond", "milliseconds":
		// Truncate to milliseconds
		return t.Truncate(time.Millisecond), nil
	case "second", "seconds":
		// Truncate to seconds
		return t.Truncate(time.Second), nil
	case "minute", "minutes":
		// Truncate to minutes
		return t.Truncate(time.Minute), nil
	case "hour", "hours":
		// Truncate to hours
		return t.Truncate(time.Hour), nil
	case "day", "days":
		// Truncate to beginning of day
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), nil
	case "week":
		// Truncate to beginning of week (Monday)
		// ISO week starts on Monday
		days := int(t.Weekday())
		if days == 0 {
			days = 7 // Sunday becomes 7
		}
		days-- // Make Monday = 0
		weekStart := t.AddDate(0, 0, -days)
		return time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, t.Location()), nil
	case "month", "months":
		// Truncate to beginning of month
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()), nil
	case "quarter":
		// Truncate to beginning of quarter
		quarterMonth := ((int(t.Month())-1)/3)*3 + 1
		return time.Date(t.Year(), time.Month(quarterMonth), 1, 0, 0, 0, 0, t.Location()), nil
	case "year", "years":
		// Truncate to beginning of year
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()), nil
	case "decade", "decades":
		// Truncate to beginning of decade
		decade := (t.Year() / 10) * 10
		return time.Date(decade, 1, 1, 0, 0, 0, 0, t.Location()), nil
	case "century", "centuries":
		// Truncate to beginning of century
		// Century 1 is years 1-100, century 2 is years 101-200, etc.
		century := ((t.Year()-1)/100)*100 + 1
		return time.Date(century, 1, 1, 0, 0, 0, 0, t.Location()), nil
	case "millennium", "millenniums":
		// Truncate to beginning of millennium
		// Millennium 1 is years 1-1000, millennium 2 is years 1001-2000, etc.
		millennium := ((t.Year()-1)/1000)*1000 + 1
		return time.Date(millennium, 1, 1, 0, 0, 0, 0, t.Location()), nil
	default:
		return time.Time{}, cerrors.Errorf("date_trunc units \"%s\" not supported", unit)
	}
}

// truncateInterval truncates an interval value to the specified unit.
func truncateInterval(unit string, interval duration.Duration) (duration.Duration, error) {
	switch strings.ToLower(unit) {
	case "microsecond", "microseconds":
		// Keep only microseconds and larger units, zero out nanoseconds
		nanos := interval.Nanos()
		truncatedNanos := (nanos / 1000) * 1000 // Truncate to microseconds
		return duration.MakeDuration(truncatedNanos, interval.Days, interval.Months), nil
	case "millisecond", "milliseconds":
		// Keep only milliseconds and larger units
		nanos := interval.Nanos()
		truncatedNanos := (nanos / 1000000) * 1000000 // Truncate to milliseconds
		return duration.MakeDuration(truncatedNanos, interval.Days, interval.Months), nil
	case "second", "seconds":
		// Keep only seconds and larger units
		nanos := interval.Nanos()
		truncatedNanos := (nanos / 1000000000) * 1000000000 // Truncate to seconds
		return duration.MakeDuration(truncatedNanos, interval.Days, interval.Months), nil
	case "minute", "minutes":
		// Keep only minutes and larger units
		nanos := interval.Nanos()
		truncatedNanos := (nanos / (60 * 1000000000)) * (60 * 1000000000) // Truncate to minutes
		return duration.MakeDuration(truncatedNanos, interval.Days, interval.Months), nil
	case "hour", "hours":
		// Keep only hours and larger units
		nanos := interval.Nanos()
		truncatedNanos := (nanos / (3600 * 1000000000)) * (3600 * 1000000000) // Truncate to hours
		return duration.MakeDuration(truncatedNanos, interval.Days, interval.Months), nil
	case "day", "days":
		// Keep only days and larger units, zero out time portion
		return duration.MakeDuration(0, interval.Days, interval.Months), nil
	case "month", "months":
		// Keep only months and larger units, zero out days and time
		return duration.MakeDuration(0, 0, interval.Months), nil
	case "year", "years":
		// Keep only whole years, truncate months
		years := interval.Months / 12
		return duration.MakeDuration(0, 0, years*12), nil
	case "decade", "decades":
		// Keep only whole decades
		years := interval.Months / 12
		decades := years / 10
		return duration.MakeDuration(0, 0, decades*10*12), nil
	case "century", "centuries":
		// Keep only whole centuries
		years := interval.Months / 12
		centuries := years / 100
		return duration.MakeDuration(0, 0, centuries*100*12), nil
	case "millennium", "millenniums":
		// Keep only whole millenniums
		years := interval.Months / 12
		millenniums := years / 1000
		return duration.MakeDuration(0, 0, millenniums*1000*12), nil
	default:
		return duration.Duration{}, cerrors.Errorf("date_trunc units \"%s\" not supported for interval", unit)
	}
}
