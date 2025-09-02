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
		year := val1.(int32)
		if year == 0 {
			return nil, errDateFieldOutOfRange
		} else if year < 0 {
			// PostgreSQL: year 0 = 1 BC, year -1 = 2 BC, etc.
			// which formatting it is handled in FormatDateTimeWithBC using calculation: (1 - year)
			year++
		}
		month := val2.(int32)
		if month < 1 || month > 12 {
			return nil, errDateFieldOutOfRange
		}
		day := val3.(int32)
		if day < 1 || day > 31 {
			return nil, errDateFieldOutOfRange
		}
		hour := val4.(int32)
		if hour < 0 || hour > 23 {
			return nil, errTimeFieldOutOfRange
		}
		minute := val5.(int32)
		if minute < 0 || minute > 59 {
			return nil, errTimeFieldOutOfRange
		}
		partialSecond := val6.(float64)
		second := int(partialSecond)
		if second < 0 || second >= 60 {
			return nil, errTimeFieldOutOfRange
		}
		nsec := int64(partialSecond*float64(time.Second)) % int64(time.Second)

		loc, err := GetServerLocation(ctx)
		if err != nil {
			return nil, err
		}
		return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), second, int(nsec), loc), nil
	},
}
