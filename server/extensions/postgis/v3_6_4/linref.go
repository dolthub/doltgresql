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
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// interpolatePoint returns the point at the given fraction of the way from `a` to `b`, interpolating every ordinate.
func interpolatePoint(a []float64, b []float64, fraction float64) []float64 {
	point := make([]float64, len(a))
	for i := range point {
		point[i] = a[i] + float64((b[i]-a[i])*fraction)
	}
	return point
}

// segmentLength returns the length of the segment from `a` to `b`, in three dimensions when `use3D` is true.
func segmentLength(a []float64, b []float64, use3D bool) float64 {
	if use3D {
		dx, dy, dz := b[0]-a[0], b[1]-a[1], b[2]-a[2]
		return math.Sqrt(float64(dx*dx) + float64(dy*dy) + float64(dz*dz))
	}
	return math.Hypot(b[0]-a[0], b[1]-a[1])
}

// pointAlong returns the point at the distance `scale` times `fraction` along a nonempty sequence, measured in three
// dimensions when `use3D` is true and in two otherwise. A distance past the end of the sequence returns its last
// point.
func pointAlong(seq []float64, dims int, fraction float64, scale float64, use3D bool) []float64 {
	distance := fraction * scale
	traveled := 0.0
	for i := dims; i+dims <= len(seq); i += dims {
		a, b := seq[i-dims:i], seq[i:i+dims]
		length := segmentLength(a, b, use3D)
		if distance < traveled+length {
			return interpolatePoint(a, b, (distance-traveled)/length)
		}
		traveled += length
	}
	return append([]float64(nil), seq[len(seq)-dims:]...)
}

// checkFraction returns an error unless the fraction lies within zero and one. A NaN fraction passes, as in PostGIS.
func checkFraction(fraction float64, argument string) error {
	if fraction < 0 || fraction > 1 {
		return pgerror.Newf(pgcode.Internal, "line_interpolate_point: %s arg isn't within [0,1]", argument)
	}
	return nil
}

// lineInterpolatePoints returns the points at each multiple of the fraction along a linestring, or only the first when
// `repeat` is false, measuring in three dimensions when `use3D` is true and the linestring has Z coordinates. A single
// point is returned as a point, and several as a multipoint.
func lineInterpolatePoints(g Geometry, fraction float64, repeat bool, use3D bool) (any, error) {
	if err := checkFraction(fraction, "2nd"); err != nil {
		return nil, err
	}
	line := g.Shape
	if line.typeNum != lineType {
		return nil, pgerror.New(pgcode.Internal, "line_interpolate_point: 1st arg isn't a line")
	}
	if line.isEmpty() {
		return g.withShape(shape{typeNum: pointType, ctype: line.ctype}), nil
	}
	count := 1
	if repeat && fraction > 0 {
		count = int(math.Floor(1 / fraction))
	}
	use3D = use3D && line.hasZ()
	dims := line.ctype.Dimension()
	total := 0.0
	for i := dims; i+dims <= len(line.coords); i += dims {
		total += segmentLength(line.coords[i-dims:i], line.coords[i:i+dims], use3D)
	}
	points := shape{typeNum: multiPointType, ctype: line.ctype}
	for i := 1; i <= count; i++ {
		coords := pointAlong(line.coords, dims, float64(i)*fraction, total, use3D)
		points.members = append(points.members, shape{typeNum: pointType, ctype: line.ctype, coords: coords})
	}
	if count == 1 {
		return g.withShape(points.members[0]), nil
	}
	return g.withShape(points), nil
}

// stLineInterpolatePoints implements ST_LineInterpolatePoint and ST_LineInterpolatePoints.
func stLineInterpolatePoints(ctx *sql.Context, args ...any) (any, error) {
	return lineInterpolatePoints(args[0].(Geometry), args[1].(float64), len(args) > 2 && args[2].(bool), false)
}

// st3DLineInterpolatePoint implements ST_3DLineInterpolatePoint.
func st3DLineInterpolatePoint(ctx *sql.Context, args ...any) (any, error) {
	return lineInterpolatePoints(args[0].(Geometry), args[1].(float64), false, true)
}

// stLineLocatePoint implements ST_LineLocatePoint, which returns the fraction of a linestring's length at which it
// comes closest to a point. The first of several equally close segments is used, and a linestring without length
// returns one.
func stLineLocatePoint(ctx *sql.Context, args ...any) (any, error) {
	g, p := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("LWGEOM_line_locate_point", g, p); err != nil {
		return nil, err
	}
	line, point := g.Shape, p.Shape
	if line.typeNum != lineType {
		return nil, pgerror.New(pgcode.Internal, "line_locate_point: 1st arg isn't a line")
	}
	if point.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, "line_locate_point: 2nd arg isn't a point")
	}
	if line.isEmpty() || point.isEmpty() {
		return 0.0, nil
	}
	dims := line.ctype.Dimension()
	closest, closestDistance := 0, math.Inf(1)
	for i := dims; i+dims <= len(line.coords); i += dims {
		if d := segmentDistanceSquared(point.coords, line.coords[i-dims:], line.coords[i:]); d < closestDistance {
			closest, closestDistance = i, d
		}
	}
	total := curveLength(line)
	if total == 0 {
		return 1.0, nil
	}
	traveled := 0.0
	for i := dims; i < closest; i += dims {
		traveled += math.Hypot(line.coords[i]-line.coords[i-dims], line.coords[i+1]-line.coords[i-dims+1])
	}
	a, b := line.coords[closest-dims:], line.coords[closest:]
	dx, dy := b[0]-a[0], b[1]-a[1]
	r := (float64((point.coords[0]-a[0])*dx) + float64((point.coords[1]-a[1])*dy)) / (float64(dx*dx) + float64(dy*dy))
	r = math.Max(0, math.Min(1, r))
	return (traveled + float64(r*math.Hypot(dx, dy))) / total, nil
}

// substring returns the part of a nonempty linestring between two fractions of its length, which is a point when the
// part has a single location.
func substring(line shape, from float64, to float64) shape {
	dims := line.ctype.Dimension()
	total := curveLength(line)
	start, end := from*total, to*total
	coords := pointAlong(line.coords, dims, start, 1, false)
	traveled := 0.0
	for i := dims; i+dims <= len(line.coords); i += dims {
		traveled += math.Hypot(line.coords[i]-line.coords[i-dims], line.coords[i+1]-line.coords[i-dims+1])
		if traveled > start && traveled < end {
			coords = appendDistinct(coords, line.coords[i:i+dims], line.ctype)
		}
	}
	coords = appendDistinct(coords, pointAlong(line.coords, dims, end, 1, false), line.ctype)
	if len(coords) == dims {
		return shape{typeNum: pointType, ctype: line.ctype, coords: coords}
	}
	return shape{typeNum: lineType, ctype: line.ctype, coords: coords}
}

// stLineSubstring implements ST_LineSubstring for geometry. A multilinestring returns the parts of its members within
// the fractions of its total length, as a multilinestring, or as a geometry collection when a part is a point.
func stLineSubstring(ctx *sql.Context, args ...any) (any, error) {
	g, from, to := args[0].(Geometry), args[1].(float64), args[2].(float64)
	if err := checkFraction(from, "2nd"); err != nil {
		return nil, err
	}
	if err := checkFraction(to, "3rd"); err != nil {
		return nil, err
	}
	if from > to {
		return nil, pgerror.New(pgcode.Internal, "2nd arg must be smaller then 3rd arg")
	}
	s := g.Shape
	switch s.typeNum {
	case lineType:
		if s.isEmpty() {
			return nil, nil
		}
		return g.withShape(substring(s, from, to)), nil
	case multiLineType:
		if s.isEmpty() {
			return nil, nil
		}
		total := 0.0
		for _, member := range s.members {
			if member.isEmpty() {
				return nil, pgerror.New(pgcode.Internal, "lwgeom_api.c [138] called with n=0 and npoints=0")
			}
			total += curveLength(member)
		}
		parts := shape{typeNum: multiLineType, ctype: s.ctype}
		memberStart := 0.0
		for _, member := range s.members {
			memberEnd := memberStart + curveLength(member)/total
			if from <= memberEnd && to >= memberStart {
				subFrom, subTo := 0.0, 1.0
				if from > memberStart {
					subFrom = (from - memberStart) / (memberEnd - memberStart)
				}
				if to < memberEnd {
					subTo = (to - memberStart) / (memberEnd - memberStart)
				}
				part := substring(member, subFrom, subTo)
				if part.typeNum == pointType {
					parts.typeNum = collectionType
				}
				parts.members = append(parts.members, part)
			}
			memberStart = memberEnd
		}
		return g.withShape(parts), nil
	default:
		return nil, pgerror.New(pgcode.Internal, "line_substring: 1st arg isn't a line")
	}
}

// linrefRoutines returns the routines that locate positions along linestrings.
func linrefRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	return []extdef.Routine{
		{Name: "st_3dlineinterpolatepoint", Symbol: "ST_3DLineInterpolatePoint", Parameters: params("geometry", "float8"), Returns: "geometry", Strict: true, Impl: st3DLineInterpolatePoint},
		{Name: "st_lineinterpolatepoint", Symbol: "st_lineinterpolatepoint_geometry_float8", Parameters: params("geometry", "float8"), Returns: "geometry", Strict: true, Impl: stLineInterpolatePoints},
		{Name: "st_lineinterpolatepoint", Symbol: "st_lineinterpolatepoint_text_float8", Parameters: params("text", "float8"), Returns: "geometry", Impl: textArgsAsGeometry(1, stLineInterpolatePoints)},
		{Name: "st_lineinterpolatepoints", Symbol: "st_lineinterpolatepoints_geometry_float8_bool", Parameters: []extdef.Parameter{{Type: "geometry"}, {Type: "float8"}, {Name: "repeat", Type: "bool", Default: `true`}}, Returns: "geometry", Strict: true, Impl: stLineInterpolatePoints},
		{Name: "st_lineinterpolatepoints", Symbol: "st_lineinterpolatepoints_text_float8", Parameters: params("text", "float8"), Returns: "geometry", Impl: textArgsAsGeometry(1, func(ctx *sql.Context, args ...any) (any, error) {
			return stLineInterpolatePoints(ctx, args[0], args[1], true)
		})},
		{Name: "st_linelocatepoint", Symbol: "LWGEOM_line_locate_point", Parameters: geomParams, Returns: "float8", Strict: true, Impl: stLineLocatePoint},
		{Name: "st_linelocatepoint", Symbol: "st_linelocatepoint_text_text", Parameters: params("text", "text"), Returns: "float8", Impl: textArgsAsGeometry(2, stLineLocatePoint)},
		{Name: "st_linesubstring", Symbol: "LWGEOM_line_substring", Parameters: params("geometry", "float8", "float8"), Returns: "geometry", Strict: true, Impl: stLineSubstring},
		{Name: "st_linesubstring", Symbol: "st_linesubstring_text_float8_float8", Parameters: params("text", "float8", "float8"), Returns: "geometry", Impl: textArgsAsGeometry(1, stLineSubstring)},
	}
}
