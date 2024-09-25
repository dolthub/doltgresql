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

// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import (
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/postgres/parser/types"
)

// TruncateString truncates a string to a given number of runes.
func truncateString(s string, maxRunes int) string {
	// This is a fast path (len(s) is an upper bound for RuneCountInString).
	if len(s) <= maxRunes {
		return s
	}
	n := utf8.RuneCountInString(s)
	if n <= maxRunes {
		return s
	}
	// Fast path for ASCII strings.
	if len(s) == n {
		return s[:maxRunes]
	}
	i := 0
	for pos := range s {
		if i == maxRunes {
			return s[:pos]
		}
		i++
	}
	// This code should be unreachable.
	return s
}

// ParseAndRequireString parses s as type t for simple types. Arrays and collated
// strings are not handled.
//
// The dependsOnContext return value indicates if we had to consult the
// ParseTimeContext (either for the time or the local timezone).
func ParseAndRequireString(
	t *types.T, s string, ctx ParseTimeContext,
) (d Datum, dependsOnContext bool, err error) {
	switch t.Family() {
	case types.ArrayFamily:
		d, dependsOnContext, err = ParseDArrayFromString(ctx, s, t.ArrayContents())
	case types.BitFamily:
		d, err = ParseDBitArray(s)
	case types.BoolFamily:
		d, err = ParseDBool(s)
	case types.BytesFamily:
		d, err = ParseDByte(s)
	case types.DateFamily:
		d, dependsOnContext, err = ParseDDate(ctx, s)
	case types.DecimalFamily:
		d, err = ParseDDecimal(s)
	case types.FloatFamily:
		d, err = ParseDFloat(s)
	case types.INetFamily:
		d, err = ParseDIPAddrFromINetString(s)
	case types.IntFamily:
		d, err = ParseDInt(s)
	case types.IntervalFamily:
		itm, typErr := t.IntervalTypeMetadata()
		if typErr != nil {
			return nil, false, typErr
		}
		d, err = ParseDIntervalWithTypeMetadata(s, itm)
	case types.Box2DFamily:
		d, err = ParseDBox2D(s)
	case types.GeographyFamily:
		d, err = ParseDGeography(s)
	case types.GeometryFamily:
		d, err = ParseDGeometry(s)
	case types.JsonFamily:
		d, err = ParseDJSON(s)
	case types.OidFamily:
		i, err := ParseDInt(s)
		if err != nil {
			return nil, false, err
		}
		d = NewDOid(*i)
	case types.StringFamily:
		// If the string type specifies a limit we truncate to that limit:
		//   'hello'::CHAR(2) -> 'he'
		// This is true of all the string type variants.
		if t.Width() > 0 {
			s = truncateString(s, int(t.Width()))
		}
		return NewDString(s), false, nil
	case types.TimeFamily:
		d, dependsOnContext, err = ParseDTime(ctx, s, TimeFamilyPrecisionToRoundDuration(t.Precision()))
	case types.TimeTZFamily:
		d, dependsOnContext, err = ParseDTimeTZ(ctx, s, TimeFamilyPrecisionToRoundDuration(t.Precision()))
	case types.TimestampFamily:
		d, dependsOnContext, err = ParseDTimestamp(ctx, s, TimeFamilyPrecisionToRoundDuration(t.Precision()))
	case types.TimestampTZFamily:
		d, dependsOnContext, err = ParseDTimestampTZ(ctx, s, TimeFamilyPrecisionToRoundDuration(t.Precision()), time.Local)
	case types.UuidFamily:
		d, err = ParseDUuidFromString(s)
	case types.EnumFamily:
		d, err = MakeDEnumFromLogicalRepresentation(t, s)
	default:
		return nil, false, errors.AssertionFailedf("unknown type %s (%T)", t, t)
	}
	return d, dependsOnContext, err
}
