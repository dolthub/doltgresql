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
	"math"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initToTimestamp registers the functions to the catalog.
func initToTimestamp() {
	framework.RegisterFunction(to_timestamp_float8)
	framework.RegisterFunction(to_timestamp_text_text)
}

// to_timestamp_float8 represents the PostgreSQL function of the same name, taking a Unix timestamp as float8.
var to_timestamp_float8 = framework.Function1{
	Name:       "to_timestamp",
	Return:     pgtypes.TimestampTZ,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		timestamp := val.(float64)

		// Handle special values - PostgreSQL returns special strings for infinity
		if math.IsInf(timestamp, 1) {
			return nil, nil
			// TODO: handle special case
			//return "infinity", nil
		} else if math.IsInf(timestamp, -1) {
			return nil, nil
			// TODO: handle special case
			//return "-infinity", nil
		} else if math.IsNaN(timestamp) {
			return nil, errors.Errorf("timestamp cannot be NaN")
		}

		// Convert Unix timestamp to time.Time
		sec := int64(timestamp)
		nsec := int64((timestamp - float64(sec)) * 1e9)

		// Ensure nsec is positive for proper time handling
		if nsec < 0 {
			sec--
			nsec += 1e9
		}

		// Check for valid timestamp range (PostgreSQL limits)
		if timestamp < -210866803200 || timestamp > 9223372036.854775 {
			return nil, errors.Errorf("timestamp out of range")
		}

		return time.Unix(sec, nsec).UTC(), nil
	},
}

// to_timestamp_text_text represents the PostgreSQL function of the same name, taking text input and format pattern.
var to_timestamp_text_text = framework.Function2{
	Name:       "to_timestamp",
	Return:     pgtypes.TimestampTZ,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		input := val1.(string)
		format := val2.(string)

		// Parse the timestamp using PostgreSQL format patterns
		return getDateTimeFromFormat(ctx, input, format)
	},
}
