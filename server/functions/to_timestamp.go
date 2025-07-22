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
	"strconv"
	"strings"
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

		// Handle special values
		if math.IsInf(timestamp, 1) {
			return time.Date(294276, time.December, 31, 23, 59, 59, 999999999, time.UTC), nil
		} else if math.IsInf(timestamp, -1) {
			return time.Date(-4713, time.January, 1, 0, 0, 0, 0, time.UTC), nil // 4713 BC
		} else if math.IsNaN(timestamp) {
			return nil, errors.Errorf("timestamp out of range")
		}

		// Convert Unix timestamp to time.Time
		sec := int64(timestamp)
		nsec := int64((timestamp - float64(sec)) * 1e9)

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
		parsedTime, err := parseTimestampWithFormat(input, format, ctx)
		if err != nil {
			return nil, err
		}

		return parsedTime, nil
	},
}

// parseTimestampWithFormat parses a timestamp string using PostgreSQL format patterns
func parseTimestampWithFormat(input, format string, ctx *sql.Context) (time.Time, error) {
	// This is a simplified implementation of PostgreSQL's to_timestamp format parsing
	// In a full implementation, this would need to handle all PostgreSQL format patterns

	// Initialize default values
	year := 1970
	month := 1
	day := 1
	hour := 0
	minute := 0
	second := 0
	nanosecond := 0

	inputPos := 0
	formatPos := 0

	// Skip leading whitespace in input
	for inputPos < len(input) && (input[inputPos] == ' ' || input[inputPos] == '\t') {
		inputPos++
	}

	for formatPos < len(format) && inputPos < len(input) {
		if formatPos >= len(format) {
			break
		}

		switch {
		case strings.HasPrefix(format[formatPos:], "YYYY"):
			// Parse 4-digit year
			if inputPos+4 > len(input) {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			yearStr := input[inputPos : inputPos+4]
			var err error
			year, err = strconv.Atoi(yearStr)
			if err != nil {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			inputPos += 4
			formatPos += 4

		case strings.HasPrefix(format[formatPos:], "MM"):
			// Parse 2-digit month
			if inputPos+2 > len(input) {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			monthStr := input[inputPos : inputPos+2]
			var err error
			month, err = strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			inputPos += 2
			formatPos += 2

		case strings.HasPrefix(format[formatPos:], "DD"):
			// Parse 2-digit day
			if inputPos+2 > len(input) {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			dayStr := input[inputPos : inputPos+2]
			var err error
			day, err = strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 31 {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			inputPos += 2
			formatPos += 2

		case strings.HasPrefix(format[formatPos:], "HH24"):
			// Parse 2-digit 24-hour format hour
			if inputPos+2 > len(input) {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			hourStr := input[inputPos : inputPos+2]
			var err error
			hour, err = strconv.Atoi(hourStr)
			if err != nil || hour < 0 || hour > 23 {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			inputPos += 2
			formatPos += 4

		case strings.HasPrefix(format[formatPos:], "MI"):
			// Parse 2-digit minute
			if inputPos+2 > len(input) {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			minuteStr := input[inputPos : inputPos+2]
			var err error
			minute, err = strconv.Atoi(minuteStr)
			if err != nil || minute < 0 || minute > 59 {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			inputPos += 2
			formatPos += 2

		case strings.HasPrefix(format[formatPos:], "SS"):
			// Parse 2-digit second
			if inputPos+2 > len(input) {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			secondStr := input[inputPos : inputPos+2]
			var err error
			second, err = strconv.Atoi(secondStr)
			if err != nil || second < 0 || second > 59 {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			inputPos += 2
			formatPos += 2

		case strings.HasPrefix(format[formatPos:], "MS"):
			// Parse 3-digit millisecond
			if inputPos+3 > len(input) {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			msStr := input[inputPos : inputPos+3]
			ms, err := strconv.Atoi(msStr)
			if err != nil || ms < 0 || ms > 999 {
				return time.Time{}, errors.Errorf("invalid input string for type timestamp")
			}
			nanosecond = ms * 1000000
			inputPos += 3
			formatPos += 2

		default:
			// Handle literal characters
			if formatPos < len(format) && inputPos < len(input) {
				if format[formatPos] == input[inputPos] {
					inputPos++
					formatPos++
				} else if format[formatPos] == ' ' || format[formatPos] == '\t' {
					// Skip whitespace in format
					formatPos++
					// Skip corresponding whitespace in input
					for inputPos < len(input) && (input[inputPos] == ' ' || input[inputPos] == '\t') {
						inputPos++
					}
				} else {
					return time.Time{}, errors.Errorf("invalid input string for type timestamp")
				}
			} else {
				formatPos++
			}
		}
	}

	// Validate the parsed date
	if month < 1 || month > 12 {
		return time.Time{}, errors.Errorf("invalid input string for type timestamp")
	}
	if day < 1 || day > 31 {
		return time.Time{}, errors.Errorf("invalid input string for type timestamp")
	}

	// Get server timezone to interpret the parsed timestamp
	serverLoc, err := GetServerLocation(ctx)
	if err != nil {
		return time.Time{}, err
	}
	
	// Create the time in the server timezone
	result := time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, serverLoc)

	// Validate the created time (handles invalid dates like Feb 30)
	if result.Year() != year || int(result.Month()) != month || result.Day() != day {
		return time.Time{}, errors.Errorf("invalid input string for type timestamp")
	}

	return result, nil
}
