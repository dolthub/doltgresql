// Copyright 2023 Dolthub, Inc.
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

// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package timetz

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgdate"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
)

var (
	// MaxTimeTZOffsetSecs is the maximum offset TimeTZ allows in seconds.
	// NOTE: postgres documentation mentions 14:59, but up to 15:59 is accepted.
	MaxTimeTZOffsetSecs = int32((15*time.Hour + 59*time.Minute) / time.Second)
	// MinTimeTZOffsetSecs is the minimum offset TimeTZ allows in seconds.
	// NOTE: postgres documentation mentions -14:59, but up to -15:59 is accepted.
	MinTimeTZOffsetSecs = -1 * MaxTimeTZOffsetSecs

	// timeTZMaxTimeRegex is a compiled regex for parsing the 24:00 timetz value.
	timeTZMaxTimeRegex = regexp.MustCompile(`^([0-9-]*T?)?\s*24:`)

	// timeTZIncludesDateRegex is a regex to check whether there is a date
	// associated with the given string when attempting to parse it.
	timeTZIncludesDateRegex = regexp.MustCompile(`^\d{4}-`)
	// timeTZHasTimeComponent determines whether there is a time component at all
	// in a given string.
	timeTZHasTimeComponent = regexp.MustCompile(`\d:`)
	// LibPQTimePrefix is the prefix lib/pq prints time-type datatypes with.
	LibPQTimePrefix = "0000-01-01"

	fixedOffsetPrefix = "fixed offset:"
)

// TimeTZ is an implementation of postgres' TimeTZ.
// Note that in this implementation, if time is equal in terms of UTC time
// the zone offset is further used to differentiate.
type TimeTZ struct {
	// TimeOfDay is the time since midnight in a given zone
	// dictated by OffsetSecs.
	timeofday.TimeOfDay
	// OffsetSecs is the offset of the zone, with the sign reversed.
	// e.g. -0800 (PDT) would have OffsetSecs of +8*60*60.
	// This is in line with the postgres implementation.
	// This means timeofday.Secs() + OffsetSecs = UTC secs.
	OffsetSecs int32
}

// MakeTimeTZ creates a TimeTZ from a TimeOfDay and offset.
func MakeTimeTZ(t timeofday.TimeOfDay, offsetSecs int32) TimeTZ {
	return TimeTZ{TimeOfDay: t, OffsetSecs: offsetSecs}
}

// MakeTimeTZFromLocation creates a TimeTZ from a TimeOfDay and time.Location.
func MakeTimeTZFromLocation(t timeofday.TimeOfDay, loc *time.Location) TimeTZ {
	_, zoneOffsetSecs := time.Now().UTC().In(loc).Zone()
	return TimeTZ{TimeOfDay: t, OffsetSecs: -int32(zoneOffsetSecs)}
}

// MakeTimeTZFromTime creates a TimeTZ from a time.Time.
// It will be trimmed to microsecond precision.
// 2400 time will overflow to 0000. If 2400 is needed, use
// MakeTimeTZFromTimeAllow2400.
func MakeTimeTZFromTime(t time.Time) TimeTZ {
	return MakeTimeTZFromLocation(
		timeofday.New(t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000),
		t.Location(),
	)
}

// MakeTimeTZFromTimeAllow2400 creates a TimeTZ from a time.Time,
// but factors in that Time2400 may be possible.
// This assumes either a lib/pq time or unix time is set.
// This should be used for storage and network deserialization, where
// 2400 time is allowed.
func MakeTimeTZFromTimeAllow2400(t time.Time) TimeTZ {
	if t.Day() != 1 {
		return MakeTimeTZFromLocation(timeofday.Time2400, t.Location())
	}
	return MakeTimeTZFromTime(t)
}

// Now returns the TimeTZ of the current location.
func Now() TimeTZ {
	return MakeTimeTZFromTime(time.Now().UTC())
}

// ParseTimeTZ parses and returns the TimeTZ represented by the
// provided string, or an error if parsing is unsuccessful.
//
// The dependsOnContext return value indicates if we had to consult the given
// `now` value (either for the time or the local timezone).
func ParseTimeTZ(
	now time.Time, s string, precision time.Duration,
) (_ TimeTZ, dependsOnContext bool, _ error) {
	// Special case as we have to use `ParseTimestamp` to get the date.
	// We cannot use `ParseTime` as it does not have timezone awareness.
	if !timeTZHasTimeComponent.MatchString(s) {
		return TimeTZ{}, false, pgerror.Newf(
			pgcode.InvalidTextRepresentation,
			"could not parse %q as TimeTZ",
			s,
		)
	}

	// ParseTimestamp requires a date field -- append date at the beginning
	// if a date has not been included.
	if !timeTZIncludesDateRegex.MatchString(s) {
		s = currentDateAsPrefix() + " " + s
	} else {
		s = ReplaceLibPQTimePrefix(s)
	}

	t, dependsOnContext, err := pgdate.ParseTimestamp(now, pgdate.ParseModeYMD, s)
	if err != nil {
		// Build our own error message to avoid exposing the dummy date.
		return TimeTZ{}, false, pgerror.Newf(
			pgcode.InvalidTextRepresentation,
			"could not parse %q as TimeTZ",
			s,
		)
	}
	retTime := timeofday.FromTime(t.Round(precision))
	// Special case on 24:00 and 24:00:00 as the parser
	// does not handle these correctly.
	if timeTZMaxTimeRegex.MatchString(s) {
		retTime = timeofday.Time2400
	}

	_, offsetSecsUnconverted := t.Zone()
	offsetSecs := int32(-offsetSecsUnconverted)
	if offsetSecs > MaxTimeTZOffsetSecs || offsetSecs < MinTimeTZOffsetSecs {
		return TimeTZ{}, false, pgerror.Newf(
			pgcode.NumericValueOutOfRange,
			"time zone displacement out of range: %q",
			s,
		)
	}
	return MakeTimeTZ(retTime, offsetSecs), dependsOnContext, nil
}

// String implements the Stringer interface.
func (t TimeTZ) String() string {
	tTime := t.ToTime()
	timeComponent := tTime.Format("15:04:05.999999")
	// 24:00:00 gets returned as 00:00:00, which is incorrect.
	if t.TimeOfDay == timeofday.Time2400 {
		timeComponent = "24:00:00"
	}
	timeZoneComponent := tTime.Format("Z07")
	// If it is UTC, .Format converts it to "Z".
	// Fully expand this component.
	if t.OffsetSecs == 0 {
		timeZoneComponent = "+00"
	} else if 0 < t.OffsetSecs && t.OffsetSecs < 60 {
		// Go's time.Format functionality does not work for offsets which
		// in the range -0s < offsetSecs < -60s, e.g. -22s offset prints as 00:00:-22.
		// Manually correct for this.
		timeZoneComponent = fmt.Sprintf("-00:00:%02d", t.OffsetSecs)
	} else if t.OffsetSecs%60 > 0 {
		timeZoneComponent = tTime.Format("Z07:00:00")
	} else if t.OffsetSecs%3600 > 0 {
		timeZoneComponent = tTime.Format("Z07:00")
	}

	return timeComponent + timeZoneComponent
}

// ToTime converts a DTimeTZ to a time.Time, corrected to the given location.
func (t TimeTZ) ToTime() time.Time {
	loc := FixedOffsetTimeZoneToLocation(-int(t.OffsetSecs), "TimeTZ")
	return t.TimeOfDay.ToTime().Add(time.Duration(t.OffsetSecs) * time.Second).In(loc)
}

// Round rounds a DTimeTZ to the given duration.
func (t TimeTZ) Round(precision time.Duration) TimeTZ {
	return MakeTimeTZ(t.TimeOfDay.Round(precision), t.OffsetSecs)
}

// ToDuration returns the TimeTZ as an offset duration from UTC midnight.
func (t TimeTZ) ToDuration() time.Duration {
	return t.ToTime().Sub(time.Unix(0, 0).UTC())
}

// Before returns whether the current is before the other TimeTZ.
func (t TimeTZ) Before(other TimeTZ) bool {
	return t.ToTime().Before(other.ToTime()) || (t.ToTime().Equal(other.ToTime()) && t.OffsetSecs < other.OffsetSecs)
}

// After returns whether the TimeTZ is after the other TimeTZ.
func (t *TimeTZ) After(other TimeTZ) bool {
	return t.ToTime().After(other.ToTime()) || (t.ToTime().Equal(other.ToTime()) && t.OffsetSecs > other.OffsetSecs)
}

// Equal returns whether the TimeTZ is equal to the other TimeTZ.
func (t TimeTZ) Equal(other TimeTZ) bool {
	return t.TimeOfDay == other.TimeOfDay && t.OffsetSecs == other.OffsetSecs
}

// ReplaceLibPQTimePrefix replaces unparsable lib/pq dates used for timestamps
// (0000-01-01) with timestamps that can be parsed by date libraries.
func ReplaceLibPQTimePrefix(s string) string {
	if strings.HasPrefix(s, LibPQTimePrefix) {
		return currentDateAsPrefix() + s[len(LibPQTimePrefix):]
	}
	return s
}

// FixedOffsetTimeZoneToLocation creates a time.Location with a set offset and
// with a name that can be marshaled by crdb between nodes.
func FixedOffsetTimeZoneToLocation(offset int, origRepr string) *time.Location {
	return time.FixedZone(
		fmt.Sprintf("%s%d (%s)", fixedOffsetPrefix, offset, origRepr),
		offset)
}

// currentDateAsPrefix will allow the timetz value to have the current time zone
// rather than timezone of 1970-01-01 in cases of timezone undefined.
// Go uses location to define the timezone, which can differ in cases of
// standard vs daylight saving time.
func currentDateAsPrefix() string {
	return time.Now().Format("2006-01-02")
}
