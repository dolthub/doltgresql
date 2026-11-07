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
	"bytes"
	"encoding/binary"
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/geo"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// compareGeometries orders two geometries the way PostGIS' btree operator class does. Geometries are equal only when
// their stored forms match exactly. Otherwise, empty geometries come first, and the rest are ordered by the position
// of their bounding box centers along a Hilbert curve. Ties fall back to the stored length, the SRID, the dimensions,
// and finally the stored bytes. Geographies place their box centers on the sphere instead.
func compareGeometries(a Geometry, b Geometry, geodetic bool) int {
	bodyA, bodyB := serializedBody(nil, a.Shape), serializedBody(nil, b.Shape)
	cmp := bytes.Compare(bodyA[:min(len(bodyA), len(bodyB))], bodyB[:min(len(bodyA), len(bodyB))])
	zA, mA, zB, mB := a.Shape.hasZ(), a.Shape.hasM(), b.Shape.hasZ(), b.Shape.hasM()
	if len(bodyA) == len(bodyB) && a.SRID == b.SRID && cmp == 0 && zA == zB && mA == mB {
		return 0
	}
	emptyA, emptyB := a.Shape.isEmpty(), b.Shape.isEmpty()
	if emptyA != emptyB {
		if emptyA {
			return -1
		}
		return 1
	}
	if !emptyA {
		hashA, hashB := sortableHash(a, geodetic), sortableHash(b, geodetic)
		if hashA != hashB {
			if hashA > hashB {
				return 1
			}
			return -1
		}
	}
	if len(bodyA) != len(bodyB) && cmp == 0 {
		if len(bodyA) < len(bodyB) {
			return -1
		}
		return 1
	}
	if a.SRID != b.SRID {
		if a.SRID > b.SRID {
			return 1
		}
		return -1
	}
	if zA != zB {
		if zA {
			return 1
		}
		return -1
	}
	if mA != mB {
		if mA {
			return 1
		}
		return -1
	}
	return cmp
}

// serializedBody appends the bytes that PostGIS stores for the shape after its header, which are compared bytewise
// to break ties. Each geometry is written as a type number and a count, followed by the ordinates of a sequence, by
// the point counts and then the ordinates of a polygon's rings, or by the members of a collection.
func serializedBody(dst []byte, s shape) []byte {
	dst = binary.LittleEndian.AppendUint32(dst, uint32(s.typeNum))
	switch s.typeNum {
	case pointType, lineType, circStringType, triangleType:
		dst = binary.LittleEndian.AppendUint32(dst, uint32(len(s.coords)/s.ctype.Dimension()))
		return appendFloats(dst, s.coords, binary.LittleEndian)
	case polygonType:
		dst = binary.LittleEndian.AppendUint32(dst, uint32(len(s.rings)))
		for _, ring := range s.rings {
			dst = binary.LittleEndian.AppendUint32(dst, uint32(len(ring)/s.ctype.Dimension()))
		}
		if len(s.rings)%2 == 1 {
			dst = binary.LittleEndian.AppendUint32(dst, 0)
		}
		for _, ring := range s.rings {
			dst = appendFloats(dst, ring, binary.LittleEndian)
		}
		return dst
	default:
		dst = binary.LittleEndian.AppendUint32(dst, uint32(len(s.members)))
		for _, member := range s.members {
			dst = serializedBody(dst, member)
		}
		return dst
	}
}

// sortableHash returns the position of the center of a non-empty geometry's bounding box along a Hilbert curve,
// computed from the float32 bits of the center's coordinates. Coordinates in common SRIDs are first scaled into a
// range where the float exponent is constant.
func sortableHash(g Geometry, geodetic bool) uint64 {
	var x, y float32
	if geodetic {
		minPt, maxPt := geocentricBox(g.Shape)
		lon, lat := geocentricToLonLat((minPt[0]+maxPt[0])/2, (minPt[1]+maxPt[1])/2, (minPt[2]+maxPt[2])/2)
		x = float32(1.5 + lon/512)
		y = float32(1.5 + lat/256)
	} else {
		b, _ := g.Shape.bounds()
		if !hasExactBox(g.Shape) {
			b = b.floatRounded()
		}
		x = float32((b.maxX + b.minX) / 2)
		y = float32((b.maxY + b.minY) / 2)
		if g.SRID == 3857 || g.SRID == 3395 {
			x = float32(1.5 + float64(x)/67108864)
			y = float32(1.5 + float64(y)/67108864)
		} else if g.SRID == sridDefault {
			x = float32(1.5 + float64(x)/512)
			y = float32(1.5 + float64(y)/256)
		}
	}
	return geo.HilbertInverse(1<<32, uint64(math.Float32bits(y)), uint64(math.Float32bits(x)))
}

// hasExactBox returns whether PostGIS stores the shape without a bounding box, deriving an exact one when needed: a
// point, a linestring of at most two points, or a multipoint or multilinestring holding only one of those. Every other
// shape keeps its box widened to the nearest float32 values.
func hasExactBox(s shape) bool {
	switch s.typeNum {
	case pointType:
		return true
	case lineType:
		return s.numPoints() <= 2
	case multiPointType:
		return len(s.members) == 1
	case multiLineType:
		return len(s.members) == 1 && s.members[0].numPoints() <= 2
	default:
		return false
	}
}

// geocentricBox returns the corners of the box that contains the unit-sphere positions of the shape's vertices.
func geocentricBox(s shape) ([3]float64, [3]float64) {
	minPt := [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxPt := [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	dims := s.ctype.Dimension()
	for _, seq := range s.sequences() {
		for i := 0; i+dims <= len(seq); i += dims {
			pt := lonLatToGeocentric(seq[i], seq[i+1])
			for axis := range pt {
				minPt[axis] = math.Min(minPt[axis], pt[axis])
				maxPt[axis] = math.Max(maxPt[axis], pt[axis])
			}
		}
	}
	//TODO: widen the box to contain the great circle arcs between vertices, as PostGIS does
	return minPt, maxPt
}

// lonLatToGeocentric returns the position on the unit sphere of a longitude and latitude given in degrees.
func lonLatToGeocentric(lonDegrees float64, latDegrees float64) [3]float64 {
	lon, lat := lonDegrees*math.Pi/180, latDegrees*math.Pi/180
	return [3]float64{math.Cos(lat) * math.Cos(lon), math.Cos(lat) * math.Sin(lon), math.Sin(lat)}
}

// geocentricToLonLat normalizes a geocentric position onto the unit sphere and returns its longitude and latitude in
// radians.
func geocentricToLonLat(x float64, y float64, z float64) (float64, float64) {
	length := math.Sqrt(float64(x*x) + float64(y*y) + float64(z*z))
	if length != 0 {
		x, y, z = x/length, y/length, z/length
	}
	return math.Atan2(y, x), math.Asin(z)
}

// comparisonImpl returns an implementation that compares its two arguments and reports whether the result passes
// the given test.
func comparisonImpl(geodetic bool, test func(int) bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		return test(compareGeometries(args[0].(Geometry), args[1].(Geometry), geodetic)), nil
	}
}

// cmpImpl returns an implementation that returns the comparison of its two arguments.
func cmpImpl(geodetic bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		return int32(compareGeometries(args[0].(Geometry), args[1].(Geometry), geodetic)), nil
	}
}

// comparisonRoutines returns the btree comparison routines of both types.
func comparisonRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	lt := func(c int) bool { return c < 0 }
	le := func(c int) bool { return c <= 0 }
	eq := func(c int) bool { return c == 0 }
	ne := func(c int) bool { return c != 0 }
	ge := func(c int) bool { return c >= 0 }
	gt := func(c int) bool { return c > 0 }
	return []extdef.Routine{
		{Name: "geometry_lt", Symbol: "lwgeom_lt", Parameters: geomParams, Returns: "bool", Strict: true, Impl: comparisonImpl(false, lt)},
		{Name: "geometry_le", Symbol: "lwgeom_le", Parameters: geomParams, Returns: "bool", Strict: true, Impl: comparisonImpl(false, le)},
		{Name: "geometry_eq", Symbol: "lwgeom_eq", Parameters: geomParams, Returns: "bool", Strict: true, Impl: comparisonImpl(false, eq)},
		{Name: "geometry_neq", Symbol: "lwgeom_neq", Parameters: geomParams, Returns: "bool", Strict: true, Impl: comparisonImpl(false, ne)},
		{Name: "geometry_ge", Symbol: "lwgeom_ge", Parameters: geomParams, Returns: "bool", Strict: true, Impl: comparisonImpl(false, ge)},
		{Name: "geometry_gt", Symbol: "lwgeom_gt", Parameters: geomParams, Returns: "bool", Strict: true, Impl: comparisonImpl(false, gt)},
		{Name: "geometry_cmp", Symbol: "lwgeom_cmp", Parameters: geomParams, Returns: "int4", Strict: true, Impl: cmpImpl(false)},
		{Name: "geography_lt", Symbol: "geography_lt", Parameters: params("geography", "geography"), Returns: "bool", Strict: true, Impl: comparisonImpl(true, lt)},
		{Name: "geography_le", Symbol: "geography_le", Parameters: params("geography", "geography"), Returns: "bool", Strict: true, Impl: comparisonImpl(true, le)},
		{Name: "geography_eq", Symbol: "geography_eq", Parameters: params("geography", "geography"), Returns: "bool", Strict: true, Impl: comparisonImpl(true, eq)},
		{Name: "geography_ge", Symbol: "geography_ge", Parameters: params("geography", "geography"), Returns: "bool", Strict: true, Impl: comparisonImpl(true, ge)},
		{Name: "geography_gt", Symbol: "geography_gt", Parameters: params("geography", "geography"), Returns: "bool", Strict: true, Impl: comparisonImpl(true, gt)},
		{Name: "geography_cmp", Symbol: "geography_cmp", Parameters: params("geography", "geography"), Returns: "int4", Strict: true, Impl: cmpImpl(true)},
	}
}

// comparisonOperators returns the comparison operators of both types.
func comparisonOperators() []extdef.Operator {
	return []extdef.Operator{
		{Symbol: "<", Left: "geometry", Right: "geometry", Routine: "lwgeom_lt", Commutator: ">", Negator: ">="},
		{Symbol: "<=", Left: "geometry", Right: "geometry", Routine: "lwgeom_le", Commutator: ">=", Negator: ">"},
		{Symbol: "=", Left: "geometry", Right: "geometry", Routine: "lwgeom_eq", Commutator: "=", Negator: "<>", Hashes: true, Merges: true},
		{Symbol: "<>", Left: "geometry", Right: "geometry", Routine: "lwgeom_neq", Commutator: "<>", Negator: "="},
		{Symbol: ">=", Left: "geometry", Right: "geometry", Routine: "lwgeom_ge", Commutator: "<=", Negator: "<"},
		{Symbol: ">", Left: "geometry", Right: "geometry", Routine: "lwgeom_gt", Commutator: "<", Negator: "<="},
		{Symbol: "<", Left: "geography", Right: "geography", Routine: "geography_lt", Commutator: ">", Negator: ">="},
		{Symbol: "<=", Left: "geography", Right: "geography", Routine: "geography_le", Commutator: ">=", Negator: ">"},
		{Symbol: "=", Left: "geography", Right: "geography", Routine: "geography_eq", Commutator: "="},
		{Symbol: ">=", Left: "geography", Right: "geography", Routine: "geography_ge", Commutator: "<=", Negator: "<"},
		{Symbol: ">", Left: "geography", Right: "geography", Routine: "geography_gt", Commutator: "<", Negator: "<="},
	}
}
