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
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/postgres/parser/timetz"
	"strings"
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTimezone registers the functions to the catalog.
func initTimezone() {
	framework.RegisterFunction(timezone_interval_timestamptz)
	framework.RegisterFunction(timezone_text_timestamptz)
	framework.RegisterFunction(timezone_text_timetz)
	framework.RegisterFunction(timezone_interval_timetz)
	framework.RegisterFunction(timezone_text_timestamp)
	framework.RegisterFunction(timezone_interval_timestamp)
}

// timezone_interval_timestamptz represents the PostgreSQL date/time function, taking {interval, timestamp with time zone}
var timezone_interval_timestamptz = framework.Function2{
	Name:               "timezone",
	Return:             pgtypes.Timestamp,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.TimestampTZ},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		dur := val1.(duration.Duration)
		t := val2.(time.Time)
		return t.UTC().Add(time.Duration(dur.Nanos())), nil
	},
}

// timezone_text_timestamptz represents the PostgreSQL date/time function, taking {text, timestamp with time zone}
var timezone_text_timestamptz = framework.Function2{
	Name:               "timezone",
	Return:             pgtypes.Timestamp,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.TimestampTZ},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		tz := val1.(string)
		timeVal := val2.(time.Time)
		newOffset, err := convertTzToOffsetSecs(tz)
		if err != nil {
			return nil, err
		}
		return timeVal.UTC().Add(time.Duration(-int64(newOffset) * NanosPerSec)), nil
	},
}

// timezone_text_timetz represents the PostgreSQL date/time function, taking {text, time with time zone}
var timezone_text_timetz = framework.Function2{
	Name:               "timezone",
	Return:             pgtypes.TimeTZ,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.TimeTZ},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		tz := val1.(string)
		timeVal := val2.(time.Time)
		newOffset, err := convertTzToOffsetSecs(tz)
		if err != nil {
			return nil, err
		}
		_, currentOffset := timeVal.Zone()
		t := timeVal.Add(time.Duration((-int64(currentOffset) + int64(newOffset)) * NanosPerSec))
		return timetz.MakeTimeTZ(timeofday.FromTime(t), -newOffset).ToTime(), nil
	},
}

// timezone_interval_timetz represents the PostgreSQL date/time function, taking {interval, time with time zone}
var timezone_interval_timetz = framework.Function2{
	Name:               "timezone",
	Return:             pgtypes.TimeTZ,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.TimeTZ},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		dur := val1.(duration.Duration)
		timeVal := val2.(time.Time)
		newOffset := int32(dur.Nanos() / NanosPerSec)
		_, currentOffset := timeVal.Zone()
		t := timeVal.Add(time.Duration((-int64(currentOffset) + int64(newOffset)) * NanosPerSec))
		return timetz.MakeTimeTZ(timeofday.FromTime(t), -newOffset).ToTime(), nil
	},
}

// timezone_text_timestamp represents the PostgreSQL date/time function, taking {text, timestamp without time zone}
var timezone_text_timestamp = framework.Function2{
	Name:               "timezone",
	Return:             pgtypes.TimestampTZ,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Timestamp},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		tz := val1.(string)
		timeVal := val2.(time.Time)
		newOffset, err := convertTzToOffsetSecs(tz)
		if err != nil {
			return nil, err
		}
		serverLoc, err := pgtypes.GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}
		return timeVal.Add(time.Duration(-int64(newOffset) * NanosPerSec)).In(serverLoc), nil
	},
}

// timezone_interval_timestamp represents the PostgreSQL date/time function, taking {interval, timestamp without time zone}
var timezone_interval_timestamp = framework.Function2{
	Name:               "timezone",
	Return:             pgtypes.TimestampTZ,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Timestamp},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		dur := val1.(duration.Duration)
		timeVal := val2.(time.Time)
		serverLoc, err := pgtypes.GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}
		return timeVal.Add(time.Duration(-dur.Nanos())).In(serverLoc), nil
	},
}

// TZ input can be in format of 'UTC' or '-04:45:33'
func convertTzToOffsetSecs(tz string) (int32, error) {
	if strings.ToLower(tz) == "utc" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err == nil {
		_, offsetSecsUnconverted := time.Now().In(loc).Zone()
		return int32(-offsetSecsUnconverted), nil
	}

	var t time.Time
	if t, err = time.Parse("Z07", tz); err == nil {
	} else if t, err = time.Parse("Z07:00", tz); err == nil {
	} else if t, err = time.Parse("Z07:00:00", tz); err != nil {
		return 0, err
	}

	_, offsetSecsUnconverted := t.Zone()
	return int32(-offsetSecsUnconverted), nil
}
