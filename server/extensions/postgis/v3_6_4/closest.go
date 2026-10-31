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
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// pointArg returns the X and Y of a geometry argument, which must be a nonempty point. The error text is the one
// that the calling function reports, and an empty point sends the notice that PostGIS sends.
func pointArg(ctx *sql.Context, g Geometry, notPoint string) ([]float64, error) {
	if g.Shape.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, notPoint)
	}
	if g.Shape.isEmpty() {
		sendNotice(ctx, "NOTICE", "lwgeom_api.c [352] called with n=0 and npoints=0")
		return nil, pgerror.New(pgcode.Internal, "Error extracting point")
	}
	return g.Shape.coords, nil
}

// azimuth returns the angle clockwise from north of the direction from `a` to `b`, from zero up to two pi, or false
// when the points are the same.
func azimuth(a []float64, b []float64) (float64, bool) {
	if a[0] == b[0] && a[1] == b[1] {
		return 0, false
	}
	az := math.Atan2(b[0]-a[0], b[1]-a[1])
	if az < 0 {
		az += 2 * math.Pi
	}
	return az, true
}

// stAzimuth implements ST_Azimuth for geometry.
func stAzimuth(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if a.SRID != b.SRID {
		return nil, pgerror.New(pgcode.Internal, "Operation on mixed SRID geometries")
	}
	pa, err := pointArg(ctx, a, "Argument must be POINT geometries")
	if err != nil {
		return nil, err
	}
	pb, err := pointArg(ctx, b, "Argument must be POINT geometries")
	if err != nil {
		return nil, err
	}
	if az, ok := azimuth(pa, pb); ok {
		return az, nil
	}
	return nil, nil
}

// stAngle implements ST_Angle. Three points give the clockwise angle at the second point from the first to the third,
// four points give the angle between the directions from the first to the second and from the third to the fourth,
// and two linestrings give the angle between the directions from their start points to their end points.
func stAngle(ctx *sql.Context, args ...any) (any, error) {
	var points [4][]float64
	if len(args) == 2 {
		for i, arg := range args {
			line := arg.(Geometry).Shape
			if line.typeNum != lineType || line.isEmpty() {
				return nil, pgerror.New(pgcode.Internal, "Argument must be LINESTRING geometries")
			}
			dims := line.ctype.Dimension()
			points[2*i], points[2*i+1] = line.coords[:dims], line.coords[len(line.coords)-dims:]
		}
	} else {
		for i, arg := range args {
			point, err := pointArg(ctx, arg.(Geometry), "Argument must be POINT geometries")
			if err != nil {
				return nil, err
			}
			points[i] = point
		}
	}
	first, second := points[1], points[0]
	third, fourth := points[1], points[2]
	if len(args) != 3 {
		first, second, third, fourth = points[0], points[1], points[2], points[3]
	}
	az1, ok1 := azimuth(first, second)
	az2, ok2 := azimuth(third, fourth)
	if !ok1 || !ok2 {
		return nil, nil
	}
	return math.Mod(2*math.Pi+2*math.Pi-az1+az2, 2*math.Pi), nil
}

// stProject implements ST_Project for geometry, which moves a point by the distance in the direction of the
// azimuth. Given a second point instead of an azimuth, it moves the second point by the distance in the direction
// from the first point to the second.
func stProject(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	origin, err := pointArg(ctx, g, "Argument must be POINT geometry")
	if err != nil {
		return nil, err
	}
	var distance, az float64
	if target, ok := args[1].(Geometry); ok {
		to, err := pointArg(ctx, target, "Argument must be POINT geometry")
		if err != nil {
			return nil, err
		}
		distance = args[2].(float64)
		az, _ = azimuth(origin, to)
		g, origin = target, to
	} else {
		distance, az = args[1].(float64), args[2].(float64)
	}
	theta := 2*math.Pi - az + math.Pi/2
	if theta > 2*math.Pi {
		theta -= 2 * math.Pi
	}
	projected := append([]float64(nil), origin...)
	projected[0] = origin[0] + float64(distance*math.Cos(theta))
	projected[1] = origin[1] + float64(distance*math.Sin(theta))
	return g.withShape(shape{typeNum: pointType, ctype: g.Shape.ctype, coords: projected}), nil
}

// closestPair is the pair of points that are closest or farthest apart on two geometries, and their distance.
type closestPair struct {
	a, b     []float64
	distance float64
	found    bool
}

// consider records a pair of points when it is closer than the current pair, or farther when `farthest` is true.
func (p *closestPair) consider(a []float64, b []float64, farthest bool) {
	dx, dy := b[0]-a[0], b[1]-a[1]
	d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
	if !p.found || (farthest && d > p.distance) || (!farthest && d < p.distance) {
		p.a, p.b, p.distance, p.found = a, b, d, true
	}
}

// segmentsOf returns the segments of a shape's linework as pairs of points, and its vertices as degenerate segments
// when it has no linework.
func segmentsOf(s shape) [][2][]float64 {
	dims := s.ctype.Dimension()
	var segments [][2][]float64
	for _, seq := range s.sequences() {
		if len(seq) == dims {
			segments = append(segments, [2][]float64{seq, seq})
		}
		for i := dims; i+dims <= len(seq); i += dims {
			segments = append(segments, [2][]float64{seq[i-dims : i], seq[i : i+dims]})
		}
	}
	return segments
}

// closestOnSegment returns the point on the segment from `a` to `b` that is closest to `p`.
func closestOnSegment(p []float64, a []float64, b []float64) []float64 {
	dx, dy := b[0]-a[0], b[1]-a[1]
	if dx == 0 && dy == 0 {
		return a[:2]
	}
	r := (float64((p[0]-a[0])*dx) + float64((p[1]-a[1])*dy)) / (float64(dx*dx) + float64(dy*dy))
	if r <= 0 {
		return a[:2]
	}
	if r >= 1 {
		return b[:2]
	}
	return []float64{a[0] + float64(r*dx), a[1] + float64(r*dy)}
}

// segmentCrossing returns the point where two segments cross, or false when they do not.
func segmentCrossing(a1 []float64, a2 []float64, b1 []float64, b2 []float64) ([]float64, bool) {
	dax, day := a2[0]-a1[0], a2[1]-a1[1]
	dbx, dby := b2[0]-b1[0], b2[1]-b1[1]
	denominator := float64(dax*dby) - float64(day*dbx)
	if denominator == 0 {
		return nil, false
	}
	s := (float64((b1[0]-a1[0])*dby) - float64((b1[1]-a1[1])*dbx)) / denominator
	t := (float64((b1[0]-a1[0])*day) - float64((b1[1]-a1[1])*dax)) / denominator
	if s < 0 || s > 1 || t < 0 || t > 1 {
		return nil, false
	}
	return []float64{a1[0] + float64(s*dax), a1[1] + float64(s*day)}, true
}

// closestPoints returns the closest pair of points on two shapes in two dimensions, keeping the first pair when
// several are equally close. A point inside a polygon of the other shape is at distance zero from it.
func closestPoints(a shape, b shape) closestPair {
	var pair closestPair
	inside := func(point []float64, area shape) bool {
		if area.dimension() != 2 {
			return false
		}
		geosArea, err := area.toGEOS()
		if err != nil {
			return false
		}
		g := spatial.NewGEOS()
		defer g.Close()
		covered, err := g.Covers(geosArea, spatial.Geometry{Type: spatial.Point, Coords: point[:2]})
		return err == nil && covered
	}
	for _, sa := range segmentsOf(a) {
		for _, sb := range segmentsOf(b) {
			if crossing, ok := segmentCrossing(sa[0], sa[1], sb[0], sb[1]); ok {
				pair.consider(crossing, crossing, false)
			}
			pair.consider(sa[0][:2], closestOnSegment(sa[0], sb[0], sb[1]), false)
			pair.consider(sa[1][:2], closestOnSegment(sa[1], sb[0], sb[1]), false)
			pair.consider(closestOnSegment(sb[0], sa[0], sa[1]), sb[0][:2], false)
			pair.consider(closestOnSegment(sb[1], sa[0], sa[1]), sb[1][:2], false)
		}
	}
	if pair.found && pair.distance > 0 {
		for _, sb := range segmentsOf(b) {
			if inside(sb[0], a) {
				pair.a, pair.b, pair.distance = sb[0][:2], sb[0][:2], 0
				break
			}
		}
		if pair.distance > 0 {
			for _, sa := range segmentsOf(a) {
				if inside(sa[0], b) {
					pair.a, pair.b, pair.distance = sa[0][:2], sa[0][:2], 0
					break
				}
			}
		}
	}
	return pair
}

// farthestPoints returns the farthest pair of vertices of two shapes in two dimensions, keeping the first pair when
// several are equally far.
func farthestPoints(a shape, b shape) closestPair {
	var pair closestPair
	dims := func(s shape) int { return s.ctype.Dimension() }
	for _, seqA := range a.sequences() {
		for i := 0; i+dims(a) <= len(seqA); i += dims(a) {
			for _, seqB := range b.sequences() {
				for j := 0; j+dims(b) <= len(seqB); j += dims(b) {
					pair.consider(seqA[i:i+2], seqB[j:j+2], true)
				}
			}
		}
	}
	return pair
}

// pairImpl returns an implementation of a function of the closest or farthest pair of points on two geometries,
// which is null when either geometry is empty.
func pairImpl(funcName string, farthest bool, result func(a Geometry, pair closestPair, args []any) any) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		if err := checkSameSRID(funcName, a, b); err != nil {
			return nil, err
		}
		var pair closestPair
		if farthest {
			pair = farthestPoints(a.Shape, b.Shape)
		} else {
			pair = closestPoints(a.Shape, b.Shape)
		}
		if !pair.found {
			return nil, nil
		}
		return result(a, pair, args), nil
	}
}

// Each of these builds the result of a closest or farthest pair function from the pair.
var (
	pairPoint = func(a Geometry, pair closestPair, args []any) any {
		return a.withShape(shape{typeNum: pointType, coords: pair.a})
	}
	pairLine = func(a Geometry, pair closestPair, args []any) any {
		return a.withShape(shape{typeNum: lineType, coords: append(append([]float64(nil), pair.a...), pair.b...)})
	}
	pairDistance = func(a Geometry, pair closestPair, args []any) any {
		return pair.distance
	}
)

// stDFullyWithin implements ST_DFullyWithin, which returns whether every vertex of the second geometry lies within
// the distance of the first.
func stDFullyWithin(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("LWGEOM_dfullywithin", a, b); err != nil {
		return nil, err
	}
	if a.Shape.isEmpty() || b.Shape.isEmpty() {
		return false, nil
	}
	//TODO: check the points between the second geometry's vertices too, as GEOS does with a buffer
	distance := args[2].(float64)
	for _, vertex := range dumpPoints(b.Shape, nil) {
		if closestPoints(a.Shape, vertex.part).distance > distance {
			return false, nil
		}
	}
	return true, nil
}

// stPointInsideCircle implements ST_PointInsideCircle, which returns whether a point lies strictly inside the circle
// with the given center and radius, or null for anything other than a nonempty point.
func stPointInsideCircle(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.typeNum != pointType || s.isEmpty() {
		return nil, nil
	}
	dx, dy := s.coords[0]-args[1].(float64), s.coords[1]-args[2].(float64)
	return math.Sqrt(float64(dx*dx)+float64(dy*dy)) < args[3].(float64), nil
}

// closestRoutines returns the routines that find the closest and farthest points of geometries, and the routines
// that measure directions.
func closestRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	pointParams := func(names ...string) []extdef.Parameter {
		ps := make([]extdef.Parameter, len(names))
		for i, name := range names {
			ps[i] = extdef.Parameter{Name: name, Type: "geometry"}
		}
		return ps
	}
	closestPointImpl := pairImpl("LWGEOM_closestpoint", false, pairPoint)
	shortestLineImpl := pairImpl("LWGEOM_shortestline2d", false, pairLine)
	longestLineImpl := pairImpl("LWGEOM_longestline2d", true, pairLine)
	maxDistanceImpl := pairImpl("LWGEOM_maxdistance2d_linestring", true, pairDistance)
	return []extdef.Routine{
		{Name: "_st_dfullywithin", Symbol: "_st_dfullywithin_geometry_geometry_float8", Parameters: append(geomParams, extdef.Parameter{Type: "float8"}), Returns: "bool", Strict: true, Impl: stDFullyWithin},
		{Name: "_st_longestline", Symbol: "LWGEOM_longestline2d", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: longestLineImpl},
		{Name: "_st_maxdistance", Symbol: "LWGEOM_maxdistance2d_linestring", Parameters: geomParams, Returns: "float8", Strict: true, Impl: maxDistanceImpl},
		{Name: "st_angle", Symbol: "st_angle_geometry_geometry", Parameters: pointParams("line1", "line2"), Returns: "float8", Strict: true, Impl: stAngle},
		{Name: "st_angle", Symbol: "LWGEOM_angle", Parameters: append(pointParams("pt1", "pt2", "pt3"), extdef.Parameter{Name: "pt4", Type: "geometry", Default: `'0101000000000000000000F87F000000000000F87F'::geometry`}), Returns: "float8", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			if fourth := args[3].(Geometry).Shape; fourth.typeNum == pointType && fourth.isEmpty() {
				args = args[:3]
			}
			return stAngle(ctx, args...)
		}},
		{Name: "st_azimuth", Symbol: "LWGEOM_azimuth", Parameters: geomParams, Returns: "float8", Strict: true, Impl: stAzimuth},
		{Name: "st_closestpoint", Symbol: "LWGEOM_closestpoint", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: closestPointImpl},
		{Name: "st_closestpoint", Symbol: "st_closestpoint_text_text", Parameters: params("text", "text"), Returns: "geometry", Impl: textArgsAsGeometry(2, closestPointImpl)},
		{Name: "st_dfullywithin", Symbol: "st_dfullywithin_geometry_geometry_float8", Parameters: append(geomParams, extdef.Parameter{Type: "float8"}), Returns: "bool", Strict: true, Impl: stDFullyWithin},
		{Name: "st_longestline", Symbol: "st_longestline_geometry_geometry", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: longestLineImpl},
		{Name: "st_maxdistance", Symbol: "st_maxdistance_geometry_geometry", Parameters: geomParams, Returns: "float8", Strict: true, Impl: maxDistanceImpl},
		{Name: "st_pointinsidecircle", Symbol: "LWGEOM_inside_circle_point", Parameters: params("geometry", "float8", "float8", "float8"), Returns: "bool", Strict: true, Impl: stPointInsideCircle},
		{Name: "st_project", Symbol: "geometry_project_direction", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "distance", Type: "float8"}, {Name: "azimuth", Type: "float8"}}, Returns: "geometry", Strict: true, Impl: stProject},
		{Name: "st_project", Symbol: "geometry_project_geometry", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}, {Name: "distance", Type: "float8"}}, Returns: "geometry", Strict: true, Impl: stProject},
		{Name: "st_shortestline", Symbol: "LWGEOM_shortestline2d", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: shortestLineImpl},
		{Name: "st_shortestline", Symbol: "st_shortestline_text_text", Parameters: params("text", "text"), Returns: "geometry", Impl: textArgsAsGeometry(2, shortestLineImpl)},
	}
}
