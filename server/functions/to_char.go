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
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initToChar registers the functions to the catalog.
func initToChar() {
	framework.RegisterFunction(to_char_timestamp_text)
	framework.RegisterFunction(to_char_timestamptz_text)
	framework.RegisterFunction(to_char_interval_text)
	framework.RegisterFunction(to_char_numeric_text)
}

// to_char_timestamp_text represents the PostgreSQL function of the same name, taking the same parameters.
// Postgres date formatting: https://www.postgresql.org/docs/15/functions-formatting.html
var to_char_timestamp_text = framework.Function2{
	Name:       "to_char",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		timestamp := val1.(time.Time)
		format := val2.(string)

		loc, err := GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}

		ttc := timestampTtc(time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), timestamp.Nanosecond(), loc))
		return tsToChar(ttc, format, false)
	},
}

// to_char_timestamptz_text represents the PostgreSQL function of the same name, taking the same parameters.
var to_char_timestamptz_text = framework.Function2{
	Name:       "to_char",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.TimestampTZ, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		timestamp := val1.(time.Time)
		format := val2.(string)

		loc, err := GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}

		ttc := timestampTtc(timestamp.In(loc))
		return tsToChar(ttc, format, false)
	},
}

func timestampTtc(ts time.Time) *tmToChar {
	ttc := &tmToChar{}
	ttc.year = ts.Year()
	ttc.mon = int(ts.Month())
	ttc.mday = ts.Day()
	ttc.hour = ts.Hour()
	ttc.min = ts.Minute()
	ttc.sec = ts.Second()
	ttc.fsec = int64(ts.Nanosecond() / 1000)
	_, ttc.gmtoff = ts.Zone()
	tzn := ts.Location().String() // TODO: should be timezone abbreviation
	if strings.HasPrefix(tzn, "fixed") {
		tzn = " "
	}
	ttc.tzn = tzn

	// calculate wday and yday
	thisDate := date2J(ttc.year, ttc.mon, ttc.mday)
	ttc.wday = j2Day(thisDate)
	ttc.yday = thisDate - date2J(ttc.year, 1, 1) + 1
	return ttc
}

// to_char_interval_text represents the PostgreSQL function of the same name, taking the same parameters.
var to_char_interval_text = framework.Function2{
	Name:       "to_char",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		interval := val1.(duration.Duration)
		format := val2.(string)

		ttc := &tmToChar{}
		ttc.year = int(interval.Months) / monthsPerYear
		ttc.mon = int(interval.Months) % monthsPerYear
		ttc.mday = int(interval.Days)
		t := interval.Nanos()

		tFrac := t / (usecsPerSecs * duration.SecsPerHour)
		t -= tFrac * (usecsPerSecs * duration.SecsPerHour)
		ttc.hour = int(tFrac)
		tFrac = t / (usecsPerSecs * duration.SecsPerMinute)
		t -= tFrac * (usecsPerSecs * duration.SecsPerMinute)
		ttc.min = int(tFrac)
		tFrac = t / usecsPerSecs
		t -= tFrac * usecsPerSecs
		ttc.sec = int(tFrac)
		//ttc.usec = int(t)

		return tsToChar(ttc, format, true)
	},
}

// to_char_numeric_text represents the PostgreSQL function of the same name, taking the same parameters.
var to_char_numeric_text = framework.Function2{
	Name:       "to_char",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		//timestamp := val1.(decimal.Decimal)
		//format := val2.(string)

		return nil, errors.Errorf(`to_char(numeric,text) is not supported yet`)
	},
}
