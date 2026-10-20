// Copyright 2026 Dolthub, Inc.
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

package v3_6_4

import (
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// geodeticTolerance is how far a coordinate may stray outside the longitude and latitude ranges before it is
// wrapped back into them, rather than simply clamped to the nearest bound.
const geodeticTolerance = 1e-10

// asGeography converts a geometry into a geography: the SRID must be a longitude/latitude system and defaults to
// 4326, and coordinates outside [-180 -90, 180 90] are brought back into range with a notice.
func asGeography(ctx *sql.Context, g Geometry) (Geometry, error) {
	if err := checkLatLongSRID(g.SRID); err != nil {
		return Geometry{}, err
	}
	if g.SRID <= sridUnknown {
		g.SRID = sridDefault
	}
	if !geographyAllowed(g.Shape.typeNum) {
		return Geometry{}, pgerror.Newf(pgcode.InvalidParameterValue, "Geography type does not support %s", typeNameForNumber(g.Shape.typeNum))
	}
	coerced := false
	g.Shape = g.Shape.mapCoords(func(point []float64) []float64 {
		x, y := nudgeOrdinate(point[0], 180), nudgeOrdinate(point[1], 90)
		if x < -180 || x > 180 || y < -90 || y > 90 {
			coerced = true
			x, y = normalizeLongitude(x), normalizeLatitude(y)
		}
		point[0], point[1] = x, y
		return point
	})
	if coerced {
		sendNotice(ctx, "NOTICE", "Coordinate values were coerced into range [-180 -90, 180 90] for GEOGRAPHY")
	}
	return g, nil
}

// geographyAllowed returns whether the geography type can hold a shape of the given type number, which excludes the
// curve, surface, and triangle types.
func geographyAllowed(typeNum int) bool {
	return typeNum >= pointType && typeNum <= collectionType
}

// nudgeOrdinate clamps an ordinate that lies within the geodetic tolerance outside [-bound, bound] onto the bound.
func nudgeOrdinate(val float64, bound float64) float64 {
	if val > bound && val-bound <= geodeticTolerance {
		return bound
	}
	if val < -bound && -bound-val <= geodeticTolerance {
		return -bound
	}
	return val
}

// normalizeLongitude wraps a longitude into the range (-180, 180].
func normalizeLongitude(lon float64) float64 {
	if lon > 360 {
		lon = math.Remainder(lon, 360)
	}
	if lon < -360 {
		lon = math.Remainder(lon, -360)
	}
	if lon > 180 {
		lon -= 360
	}
	if lon < -180 {
		lon += 360
	}
	if lon == -180 {
		return 180
	}
	if lon == -360 {
		return 0
	}
	return lon
}

// normalizeLatitude reflects a latitude back into the range [-90, 90], as though it continued over a pole.
func normalizeLatitude(lat float64) float64 {
	if lat > 360 {
		lat = math.Remainder(lat, 360)
	}
	if lat < -360 {
		lat = math.Remainder(lat, -360)
	}
	if lat > 180 {
		lat = 180 - lat
	}
	if lat < -180 {
		lat = -180 - lat
	}
	if lat > 90 {
		lat = 180 - lat
	}
	if lat < -90 {
		lat = -180 - lat
	}
	return lat
}

// geographyIn implements geography_in, which accepts hex-encoded WKB or WKT.
func geographyIn(ctx *sql.Context, args ...any) (any, error) {
	input := args[0].(string)
	if input == "" {
		return nil, errWKTInvalid
	}
	var g Geometry
	var err error
	if input[0] == '0' {
		g, err = decodeHexWKB(ctx, input)
	} else {
		g, err = parseEWKT(input)
		if err == nil {
			g.SRID = clampSRID(ctx, g.SRID)
		}
	}
	if err != nil {
		return nil, err
	}
	if g, err = asGeography(ctx, g); err != nil {
		return nil, err
	}
	return enforceTypmod(g, typmodArg(args))
}

// geographyRecv implements geography_recv.
func geographyRecv(ctx *sql.Context, args ...any) (any, error) {
	g, err := decodeWKB(ctx, args[0].([]byte))
	if err != nil {
		return nil, err
	}
	if g, err = asGeography(ctx, g); err != nil {
		return nil, err
	}
	return enforceTypmod(g, typmodArg(args))
}

// geographyFromGeometry implements geography(geometry), the cast from geometry to geography.
func geographyFromGeometry(ctx *sql.Context, args ...any) (any, error) {
	return asGeography(ctx, args[0].(Geometry))
}

// geometryFromGeography implements geometry(geography), the cast from geography to geometry, which keeps the
// geography as it is.
func geometryFromGeography(ctx *sql.Context, args ...any) (any, error) {
	return args[0].(Geometry), nil
}
