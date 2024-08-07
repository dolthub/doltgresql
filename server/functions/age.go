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
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	duration "github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initAge registers the functions to the catalog.
func initAge() {
	framework.RegisterFunction(age_timestamp_timestamp)
	framework.RegisterFunction(age_timestamp)
}

// age_timestamp_timestamp represents the PostgreSQL date/time function.
var age_timestamp_timestamp = framework.Function2{
	Name:               "age",
	Return:             pgtypes.Interval,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Timestamp},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		t1 := val1.(time.Time)
		t2 := val2.(time.Time)
		return diffTimes(t1, t2), nil
	},
}

// age_timestamp_timestamp represents the PostgreSQL date/time function.
var age_timestamp = framework.Function1{
	Name:               "age",
	Return:             pgtypes.Interval,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Timestamp},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		t := val.(time.Time)
		// current_date (at midnight)
		cur, err := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		return diffTimes(cur, t), nil
	},
}

// diffTimes returns the duration t1-t2. It subtracts each time component separately,
// unlike time.Sub() function.
func diffTimes(t1, t2 time.Time) duration.Duration {
	// if t1 is before t2, then negate the result.
	negate := t1.Before(t2)
	if negate {
		t1, t2 = t2, t1
	}

	// Calculate difference in each unit
	years := int64(t1.Year() - t2.Year())
	months := int64(t1.Month() - t2.Month())
	days := int64(t1.Day() - t2.Day())
	hours := int64(t1.Hour() - t2.Hour())
	minutes := int64(t1.Minute() - t2.Minute())
	seconds := int64(t1.Second() - t2.Second())
	nanoseconds := int64(t1.Nanosecond() - t2.Nanosecond())

	// Adjust for any negative values
	if nanoseconds < 0 {
		nanoseconds += 1e9
		seconds--
	}
	if seconds < 0 {
		seconds += 60
		minutes--
	}
	if minutes < 0 {
		minutes += 60
		hours--
	}
	if hours < 0 {
		hours += 24
		days--
	}
	if days < 0 {
		days += 30
		months--
	}
	if months < 0 {
		months += 12
		years--
	}

	durNanos := nanoseconds + seconds*NanosPerSec + minutes*NanosPerSec*duration.SecsPerMinute + hours*NanosPerSec*duration.SecsPerHour
	durDays := days
	durMonths := months + years*duration.MonthsPerYear

	if negate {
		return duration.MakeDuration(-durNanos, -durDays, -durMonths)
	} else {
		return duration.MakeDuration(durNanos, durDays, durMonths)
	}
}
