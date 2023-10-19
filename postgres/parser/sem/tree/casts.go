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

// Copyright 2020 The Cockroach Authors.
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
	"github.com/dolthub/doltgresql/postgres/parser/types"
)

type castInfo struct {
	from       types.Family
	to         types.Family
	volatility Volatility

	// If set, the volatility of this cast is not cross-checked against postgres.
	// Use this with caution.
	ignoreVolatilityCheck bool
}

// validCasts lists all valid explicit casts.
//
// This list must be kept in sync with the capabilities of PerformCast.
//
// Each cast defines a volatility:
//
//  - immutable casts yield the same result on the same arguments in whatever
//    context they are evaluated.
//
//  - stable casts can yield a different result depending on the evaluation context:
//    - session settings (e.g. bytes encoding format)
//    - current timezone
//    - current time (e.g. 'now'::string).
//
// TODO(radu): move the PerformCast code for each cast into functions defined
// within each cast.
//
var validCasts = []castInfo{
	// Casts to BitFamily.
	{from: types.UnknownFamily, to: types.BitFamily, volatility: VolatilityImmutable},
	{from: types.BitFamily, to: types.BitFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.BitFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.BitFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.BitFamily, volatility: VolatilityImmutable},

	// Casts to BoolFamily.
	{from: types.UnknownFamily, to: types.BoolFamily, volatility: VolatilityImmutable},
	{from: types.BoolFamily, to: types.BoolFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.BoolFamily, volatility: VolatilityImmutable},
	{from: types.FloatFamily, to: types.BoolFamily, volatility: VolatilityImmutable},
	{from: types.DecimalFamily, to: types.BoolFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.BoolFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.BoolFamily, volatility: VolatilityImmutable},

	// Casts to IntFamily.
	{from: types.UnknownFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.BoolFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.FloatFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.DecimalFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.TimestampFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.DateFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.IntervalFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.OidFamily, to: types.IntFamily, volatility: VolatilityImmutable},
	{from: types.BitFamily, to: types.IntFamily, volatility: VolatilityImmutable},

	// Casts to FloatFamily.
	{from: types.UnknownFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.BoolFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.FloatFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.DecimalFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.TimestampFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.DateFamily, to: types.FloatFamily, volatility: VolatilityImmutable},
	{from: types.IntervalFamily, to: types.FloatFamily, volatility: VolatilityImmutable},

	// Casts to Box2D Family.
	{from: types.UnknownFamily, to: types.Box2DFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.Box2DFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.Box2DFamily, volatility: VolatilityImmutable},
	{from: types.GeometryFamily, to: types.Box2DFamily, volatility: VolatilityImmutable},
	{from: types.Box2DFamily, to: types.Box2DFamily, volatility: VolatilityImmutable},

	// Casts to GeographyFamily.
	{from: types.UnknownFamily, to: types.GeographyFamily, volatility: VolatilityImmutable},
	{from: types.BytesFamily, to: types.GeographyFamily, volatility: VolatilityImmutable},
	{from: types.JsonFamily, to: types.GeographyFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.GeographyFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.GeographyFamily, volatility: VolatilityImmutable},
	{from: types.GeographyFamily, to: types.GeographyFamily, volatility: VolatilityImmutable},
	{from: types.GeometryFamily, to: types.GeographyFamily, volatility: VolatilityImmutable},

	// Casts to GeometryFamily.
	{from: types.UnknownFamily, to: types.GeometryFamily, volatility: VolatilityImmutable},
	{from: types.Box2DFamily, to: types.GeometryFamily, volatility: VolatilityImmutable},
	{from: types.BytesFamily, to: types.GeometryFamily, volatility: VolatilityImmutable},
	{from: types.JsonFamily, to: types.GeometryFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.GeometryFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.GeometryFamily, volatility: VolatilityImmutable},
	{from: types.GeographyFamily, to: types.GeometryFamily, volatility: VolatilityImmutable},
	{from: types.GeometryFamily, to: types.GeometryFamily, volatility: VolatilityImmutable},

	// Casts to DecimalFamily.
	{from: types.UnknownFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.BoolFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.FloatFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.DecimalFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.TimestampFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.DateFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},
	{from: types.IntervalFamily, to: types.DecimalFamily, volatility: VolatilityImmutable},

	// Casts to StringFamily.
	{from: types.UnknownFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.BoolFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.FloatFamily, to: types.StringFamily, volatility: VolatilityStable},
	{from: types.DecimalFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.BitFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.ArrayFamily, to: types.StringFamily, volatility: VolatilityStable},
	{from: types.TupleFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.GeometryFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.Box2DFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.GeographyFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.BytesFamily, to: types.StringFamily, volatility: VolatilityStable},
	{from: types.TimestampFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.StringFamily, volatility: VolatilityStable},
	{from: types.IntervalFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.UuidFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.DateFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.TimeFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.TimeTZFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.OidFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.INetFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.JsonFamily, to: types.StringFamily, volatility: VolatilityImmutable},
	{from: types.EnumFamily, to: types.StringFamily, volatility: VolatilityImmutable},

	// Casts to CollatedStringFamily.
	{from: types.UnknownFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.BoolFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.FloatFamily, to: types.CollatedStringFamily, volatility: VolatilityStable},
	{from: types.DecimalFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.BitFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.ArrayFamily, to: types.CollatedStringFamily, volatility: VolatilityStable},
	{from: types.TupleFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.Box2DFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.GeometryFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.GeographyFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.BytesFamily, to: types.CollatedStringFamily, volatility: VolatilityStable},
	{from: types.TimestampFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.CollatedStringFamily, volatility: VolatilityStable},
	{from: types.IntervalFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.UuidFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.DateFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.TimeFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.TimeTZFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.OidFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.INetFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.JsonFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},
	{from: types.EnumFamily, to: types.CollatedStringFamily, volatility: VolatilityImmutable},

	// Casts to BytesFamily.
	{from: types.UnknownFamily, to: types.BytesFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.BytesFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.BytesFamily, volatility: VolatilityImmutable},
	{from: types.BytesFamily, to: types.BytesFamily, volatility: VolatilityImmutable},
	{from: types.UuidFamily, to: types.BytesFamily, volatility: VolatilityImmutable},
	{from: types.GeometryFamily, to: types.BytesFamily, volatility: VolatilityImmutable},
	{from: types.GeographyFamily, to: types.BytesFamily, volatility: VolatilityImmutable},

	// Casts to DateFamily.
	{from: types.UnknownFamily, to: types.DateFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.DateFamily, volatility: VolatilityStable},
	{from: types.CollatedStringFamily, to: types.DateFamily, volatility: VolatilityStable},
	{from: types.DateFamily, to: types.DateFamily, volatility: VolatilityImmutable},
	{from: types.TimestampFamily, to: types.DateFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.DateFamily, volatility: VolatilityStable},
	{from: types.IntFamily, to: types.DateFamily, volatility: VolatilityImmutable},

	// Casts to TimeFamily.
	{from: types.UnknownFamily, to: types.TimeFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.TimeFamily, volatility: VolatilityStable},
	{from: types.CollatedStringFamily, to: types.TimeFamily, volatility: VolatilityStable},
	{from: types.TimeFamily, to: types.TimeFamily, volatility: VolatilityImmutable},
	{from: types.TimeTZFamily, to: types.TimeFamily, volatility: VolatilityImmutable},
	{from: types.TimestampFamily, to: types.TimeFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.TimeFamily, volatility: VolatilityStable},
	{from: types.IntervalFamily, to: types.TimeFamily, volatility: VolatilityImmutable},

	// Casts to TimeTZFamily.
	{from: types.UnknownFamily, to: types.TimeTZFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.TimeTZFamily, volatility: VolatilityStable},
	{from: types.CollatedStringFamily, to: types.TimeTZFamily, volatility: VolatilityStable},
	{from: types.TimeFamily, to: types.TimeTZFamily, volatility: VolatilityStable},
	{from: types.TimeTZFamily, to: types.TimeTZFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.TimeTZFamily, volatility: VolatilityStable},

	// Casts to TimestampFamily.
	{from: types.UnknownFamily, to: types.TimestampFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.TimestampFamily, volatility: VolatilityStable},
	{from: types.CollatedStringFamily, to: types.TimestampFamily, volatility: VolatilityStable},
	{from: types.DateFamily, to: types.TimestampFamily, volatility: VolatilityImmutable},
	{from: types.TimestampFamily, to: types.TimestampFamily, volatility: VolatilityImmutable},
	{from: types.TimestampTZFamily, to: types.TimestampFamily, volatility: VolatilityStable},
	{from: types.IntFamily, to: types.TimestampFamily, volatility: VolatilityImmutable},

	// Casts to TimestampTZFamily.
	{from: types.UnknownFamily, to: types.TimestampTZFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.TimestampTZFamily, volatility: VolatilityStable},
	{from: types.CollatedStringFamily, to: types.TimestampTZFamily, volatility: VolatilityStable},
	{from: types.DateFamily, to: types.TimestampTZFamily, volatility: VolatilityStable},
	{from: types.TimestampFamily, to: types.TimestampTZFamily, volatility: VolatilityStable},
	{from: types.TimestampTZFamily, to: types.TimestampTZFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.TimestampTZFamily, volatility: VolatilityImmutable},

	// Casts to IntervalFamily.
	{from: types.UnknownFamily, to: types.IntervalFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.IntervalFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.IntervalFamily, volatility: VolatilityImmutable},
	{from: types.IntFamily, to: types.IntervalFamily, volatility: VolatilityImmutable},
	{from: types.TimeFamily, to: types.IntervalFamily, volatility: VolatilityImmutable},
	{from: types.IntervalFamily, to: types.IntervalFamily, volatility: VolatilityImmutable},
	{from: types.FloatFamily, to: types.IntervalFamily, volatility: VolatilityImmutable},
	{from: types.DecimalFamily, to: types.IntervalFamily, volatility: VolatilityImmutable},

	// Casts to OidFamily.
	{from: types.UnknownFamily, to: types.OidFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.OidFamily, volatility: VolatilityStable},
	{from: types.CollatedStringFamily, to: types.OidFamily, volatility: VolatilityStable},
	{from: types.IntFamily, to: types.OidFamily, volatility: VolatilityStable, ignoreVolatilityCheck: true},
	{from: types.OidFamily, to: types.OidFamily, volatility: VolatilityStable},

	// Casts to UuidFamily.
	{from: types.UnknownFamily, to: types.UuidFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.UuidFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.UuidFamily, volatility: VolatilityImmutable},
	{from: types.BytesFamily, to: types.UuidFamily, volatility: VolatilityImmutable},
	{from: types.UuidFamily, to: types.UuidFamily, volatility: VolatilityImmutable},

	// Casts to INetFamily.
	{from: types.UnknownFamily, to: types.INetFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.INetFamily, volatility: VolatilityImmutable},
	{from: types.CollatedStringFamily, to: types.INetFamily, volatility: VolatilityImmutable},
	{from: types.INetFamily, to: types.INetFamily, volatility: VolatilityImmutable},

	// Casts to ArrayFamily.
	{from: types.UnknownFamily, to: types.ArrayFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.ArrayFamily, volatility: VolatilityStable},

	// Casts to JsonFamily.
	{from: types.UnknownFamily, to: types.JsonFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.JsonFamily, volatility: VolatilityImmutable},
	{from: types.JsonFamily, to: types.JsonFamily, volatility: VolatilityImmutable},
	{from: types.GeometryFamily, to: types.JsonFamily, volatility: VolatilityImmutable},
	{from: types.GeographyFamily, to: types.JsonFamily, volatility: VolatilityImmutable},

	// Casts to EnumFamily.
	{from: types.UnknownFamily, to: types.EnumFamily, volatility: VolatilityImmutable},
	{from: types.StringFamily, to: types.EnumFamily, volatility: VolatilityImmutable},
	{from: types.EnumFamily, to: types.EnumFamily, volatility: VolatilityImmutable},
	{from: types.BytesFamily, to: types.EnumFamily, volatility: VolatilityImmutable},

	// Casts to TupleFamily.
	{from: types.UnknownFamily, to: types.TupleFamily, volatility: VolatilityImmutable},
}

type castsMapKey struct {
	from, to types.Family
}

var castsMap map[castsMapKey]*castInfo

func init() {
	castsMap = make(map[castsMapKey]*castInfo, len(validCasts))
	for i := range validCasts {
		c := &validCasts[i]

		key := castsMapKey{from: c.from, to: c.to}
		castsMap[key] = c
	}
}

// lookupCast returns the information for a valid cast.
// Returns nil if this is not a valid cast.
// Does not handle array and tuple casts.
func lookupCast(from, to types.Family) *castInfo {
	return castsMap[castsMapKey{from: from, to: to}]
}
