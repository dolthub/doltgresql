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
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initMakeTimestamp registers the functions to the catalog.
func initMakeTimestamp() {
	framework.RegisterFunction(make_timestamp)
	framework.RegisterFunction(make_timestamptz_int32_int32_int32_int32_int32_float64)
	framework.RegisterFunction(make_timestamptz_int32_int32_int32_int32_int32_float64_text)
}

var errDateFieldOutOfRange = errors.Errorf(`date field value out of range`)
var errTimeFieldOutOfRange = errors.Errorf(`time field value out of range`)

// make_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
var make_timestamp = framework.Function6{
	Name:               "make_timestamp",
	Return:             pgtypes.Timestamp,
	Parameters:         [6]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Float64},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [7]*pgtypes.DoltgresType, val1, val2, val3, val4, val5, val6 any) (any, error) {
		loc, err := GetServerLocation(ctx)
		if err != nil {
			return time.Time{}, err
		}
		return getTimestampInServerLocation(val1.(int32), val2.(int32), val3.(int32), val4.(int32), val5.(int32), val6.(float64), loc)
	},
}

// make_timestamptz_int32_int32_int32_int32_int32_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var make_timestamptz_int32_int32_int32_int32_int32_float64 = framework.Function6{
	Name:               "make_timestamptz",
	Return:             pgtypes.TimestampTZ,
	Parameters:         [6]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Float64},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [7]*pgtypes.DoltgresType, val1, val2, val3, val4, val5, val6 any) (any, error) {
		loc, err := GetServerLocation(ctx)
		if err != nil {
			return time.Time{}, err
		}
		return getTimestampInServerLocation(val1.(int32), val2.(int32), val3.(int32), val4.(int32), val5.(int32), val6.(float64), loc)
	},
}

// make_timestamptz_int32_int32_int32_int32_int32_float64_text represents the PostgreSQL function of the same name, taking the same parameters.
var make_timestamptz_int32_int32_int32_int32_int32_float64_text = framework.Function7{
	Name:               "make_timestamptz",
	Return:             pgtypes.TimestampTZ,
	Parameters:         [7]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Int32, pgtypes.Float64, pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [8]*pgtypes.DoltgresType, val1, val2, val3, val4, val5, val6, val7 any) (any, error) {
		tz := val7.(string)
		loc, _, _, err := convertTzToOffsetSecs(time.Now().UTC(), tz)
		if err != nil {
			return nil, errors.Errorf(`time zone "%s" not recognized`, tz)
		}
		return getTimestampInServerLocation(val1.(int32), val2.(int32), val3.(int32), val4.(int32), val5.(int32), val6.(float64), loc)
	},
}

func getTimestampInServerLocation(year, month, day, hour, minute int32, partialSecond float64, loc *time.Location) (time.Time, error) {
	if year == 0 {
		return time.Time{}, errDateFieldOutOfRange
	} else if year < 0 {
		// PostgreSQL: year 0 = 1 BC, year -1 = 2 BC, etc.
		// which formatting it is handled in FormatDateTimeWithBC using calculation: (1 - year)
		year++
	}

	if month < 1 || month > 12 {
		return time.Time{}, errDateFieldOutOfRange
	}

	if day < 1 || day > 31 {
		return time.Time{}, errDateFieldOutOfRange
	}

	if hour < 0 || hour > 23 {
		return time.Time{}, errTimeFieldOutOfRange
	}

	if minute < 0 || minute > 59 {
		return time.Time{}, errTimeFieldOutOfRange
	}

	second := int(partialSecond)
	if second < 0 || second >= 60 {
		return time.Time{}, errTimeFieldOutOfRange
	}
	nsec := int64(partialSecond*float64(time.Second)) % int64(time.Second)

	return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), second, int(nsec), loc), nil
}
