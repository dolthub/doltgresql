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

package pgdate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
		hasError bool
	}{
		// Basic ISO 8601 formats
		{"2025-03-24T19:21:59Z", time.Date(2025, 3, 24, 19, 21, 59, 0, time.UTC), false},
		{"2025-03-24T19:21:59+00:00", time.Date(2025, 3, 24, 19, 21, 59, 0, time.FixedZone("+000000", 0)), false},
		{"2025-03-24T19:21:59.690479+00:00", time.Date(2025, 3, 24, 19, 21, 59, 690479000, time.FixedZone("+000000", 0)), false},
		{"2025-03-24", time.Date(2025, 3, 24, 0, 0, 0, 0, time.Local), false},

		// Different time zone offsets
		{"2025-03-24T19:21:59-05:00", time.Date(2025, 3, 24, 19, 21, 59, 0, time.FixedZone("-050000", -5*60*60)), false},
		{"2025-03-24T19:21:59.123456+02:30", time.Date(2025, 3, 24, 19, 21, 59, 123456000, time.FixedZone("+023000", 2*60*60+30*60)), false},

		// Edge cases
		{"2024-02-29T12:34:56Z", time.Date(2024, 2, 29, 12, 34, 56, 0, time.UTC), false}, // Leap year
		{"0001-01-01T00:00:00Z", time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), false},       // Minimum date
		{"9999-12-31T23:59:59Z", time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC), false}, // Maximum date

		// Invalid formats
		{"24-03-2025T19:21:59Z", time.Time{}, true}, // Wrong date format
		{"2025-03-24T25:00:00Z", time.Time{}, true}, // Invalid hour
		{"2025-13-01T12:00:00Z", time.Time{}, true}, // Invalid month
		{"2025-03-32T12:00:00Z", time.Time{}, true}, // Invalid day
		{"2025-02-29T12:34:56Z", time.Time{}, true},
		{"NotATimestamp", time.Time{}, true},        // Completely invalid
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			parsed, _, err := ParseTimestamp(time.Now(), ParseModeYMD, test.input)
			if test.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, parsed)
			}
		})
	}
}