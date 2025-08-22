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

	cerrors "github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initDateBin registers the functions to the catalog.
func initDateBin() {
	framework.RegisterFunction(date_bin_interval_timestamp_timestamp)
	framework.RegisterFunction(date_bin_interval_timestamptz_timestamptz)
}

// date_bin_interval_timestamp_timestamp represents the PostgreSQL date_bin function for timestamp type.
var date_bin_interval_timestamp_timestamp = framework.Function3{
	Name:       "date_bin",
	Return:     pgtypes.Timestamp,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Timestamp, pgtypes.Timestamp},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		interval := val1.(duration.Duration)
		timestamp := val2.(time.Time)
		origin := val3.(time.Time)
		return binTimestamp(interval, timestamp, origin)
	},
}

// date_bin_interval_timestamptz_timestamptz represents the PostgreSQL date_bin function for timestamptz type.
var date_bin_interval_timestamptz_timestamptz = framework.Function3{
	Name:       "date_bin",
	Return:     pgtypes.TimestampTZ,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.TimestampTZ, pgtypes.TimestampTZ},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		interval := val1.(duration.Duration)
		timestamp := val2.(time.Time)
		origin := val3.(time.Time)
		return binTimestamp(interval, timestamp, origin)
	},
}

// binTimestamp implements the core logic for date_bin function.
func binTimestamp(interval duration.Duration, timestamp time.Time, origin time.Time) (time.Time, error) {
	if interval.Months != 0 {
		return time.Time{}, cerrors.Errorf("timestamps cannot be binned into intervals containing months or years")
	}

	// Calculate total nanoseconds in the interval
	intervalNanos := interval.Nanos() + int64(interval.Days)*24*3600*1000000000

	// Check for zero or negative interval
	if intervalNanos <= 0 {
		return time.Time{}, cerrors.Errorf("stride must be greater than zero")
	}

	// Convert times to nanoseconds since origin
	timestampNanos := timestamp.UnixNano()
	originNanos := origin.UnixNano()

	// Calculate the difference from origin
	diffNanos := timestampNanos - originNanos

	// Calculate how many complete intervals have passed
	binCount := diffNanos / intervalNanos
	if diffNanos < 0 {
		// For negative differences, we need to round down (towards negative infinity)
		binCount = int64(math.Floor(float64(diffNanos) / float64(intervalNanos)))
	}

	// Calculate the bin start time
	binStartNanos := originNanos + binCount*intervalNanos

	return time.Unix(0, binStartNanos).In(timestamp.Location()), nil
}
