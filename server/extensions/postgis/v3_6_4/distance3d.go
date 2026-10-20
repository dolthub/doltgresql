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

// point3 is a point in three dimensions.
type point3 [3]float64

// sub returns the vector from `q` to `p`.
func (p point3) sub(q point3) point3 {
	return point3{p[0] - q[0], p[1] - q[1], p[2] - q[2]}
}

// add returns the sum of two vectors.
func (p point3) add(q point3) point3 {
	return point3{p[0] + q[0], p[1] + q[1], p[2] + q[2]}
}

// scale returns the vector multiplied by `f`.
func (p point3) scale(f float64) point3 {
	return point3{p[0] * f, p[1] * f, p[2] * f}
}

// dot returns the dot product of two vectors.
func (p point3) dot(q point3) float64 {
	return p[0]*q[0] + p[1]*q[1] + p[2]*q[2]
}

// distance returns the distance between two points, in two dimensions when `flat` is true.
func (p point3) distance(q point3, flat bool) float64 {
	d := p.sub(q)
	if flat {
		d[2] = 0
	}
	return math.Sqrt(d.dot(d))
}

// polygon3 is a polygon in three dimensions, with the plane through its exterior ring.
type polygon3 struct {
	rings [][]point3
	// origin is a point on the plane, and normal is its unit normal, which is zero when the ring has no area.
	origin point3
	normal point3
}

// primitives3D holds the points, segments, and polygons that make up a shape.
type primitives3D struct {
	points   []point3
	segments [][2]point3
	polygons []polygon3
}

// thirdOrdinate returns the index of the ordinate that serves as the third dimension of a shape's points: its Z when
// `useZ` is true, otherwise its M when `useM` is true, or -1 when the points sit at zero.
func thirdOrdinate(s shape, useZ bool, useM bool) int {
	if useZ && s.hasZ() {
		return 2
	}
	if useM && s.hasM() {
		if s.hasZ() {
			return 3
		}
		return 2
	}
	return -1
}

// newPolygon3 returns the polygon of the rings, finding its plane with Newell's method.
func newPolygon3(rings [][]point3) polygon3 {
	p := polygon3{rings: rings}
	if len(rings) == 0 || len(rings[0]) == 0 {
		return p
	}
	exterior := rings[0]
	var normal point3
	for i := 0; i+1 < len(exterior); i++ {
		a, b := exterior[i], exterior[i+1]
		normal[0] += (a[1] - b[1]) * (a[2] + b[2])
		normal[1] += (a[2] - b[2]) * (a[0] + b[0])
		normal[2] += (a[0] - b[0]) * (a[1] + b[1])
	}
	if length := math.Sqrt(normal.dot(normal)); length > 0 {
		p.normal = normal.scale(1 / length)
	}
	p.origin = exterior[0]
	return p
}

// collectPrimitives3D returns the primitives of a shape, taking the third dimension from the ordinate at the given
// index. Curves are not supported.
func collectPrimitives3D(s shape, third int) (primitives3D, error) {
	var prims primitives3D
	err := addPrimitives3D(&prims, s, third)
	return prims, err
}

// addPrimitives3D adds the primitives of a shape to `prims`.
func addPrimitives3D(prims *primitives3D, s shape, third int) error {
	dims := s.ctype.Dimension()
	points := func(seq []float64) []point3 {
		var pts []point3
		for i := 0; i+dims <= len(seq); i += dims {
			p := point3{seq[i], seq[i+1], 0}
			if third >= 0 {
				p[2] = seq[i+third]
			}
			pts = append(pts, p)
		}
		return pts
	}
	switch s.typeNum {
	case pointType:
		prims.points = append(prims.points, points(s.coords)...)
	case lineType:
		pts := points(s.coords)
		if len(pts) == 1 {
			prims.points = append(prims.points, pts[0])
		}
		for i := 1; i < len(pts); i++ {
			prims.segments = append(prims.segments, [2]point3{pts[i-1], pts[i]})
		}
	case polygonType, triangleType:
		var rings [][]point3
		if s.typeNum == triangleType {
			rings = append(rings, points(s.coords))
		}
		for _, ring := range s.rings {
			rings = append(rings, points(ring))
		}
		if len(rings) > 0 && len(rings[0]) > 0 {
			prims.polygons = append(prims.polygons, newPolygon3(rings))
		}
	case multiPointType, multiLineType, multiPolygonType, collectionType, polyhedralSurfaceType, tinType:
		for _, member := range s.members {
			if err := addPrimitives3D(prims, member, third); err != nil {
				return err
			}
		}
	default:
		return pgerror.Newf(pgcode.Internal, "lw_dist3d_distribute_bruteforce: Unsupported geometry type: %s", typeNameForNumber(s.typeNum))
	}
	return nil
}

// vertices returns every vertex of the primitives.
func (prims primitives3D) vertices() []point3 {
	vertices := append([]point3(nil), prims.points...)
	for _, segment := range prims.segments {
		vertices = append(vertices, segment[0], segment[1])
	}
	for _, polygon := range prims.polygons {
		for _, ring := range polygon.rings {
			vertices = append(vertices, ring...)
		}
	}
	return vertices
}

// edges returns the segments of the polygon's rings.
func (p polygon3) edges() [][2]point3 {
	var edges [][2]point3
	for _, ring := range p.rings {
		for i := 1; i < len(ring); i++ {
			edges = append(edges, [2]point3{ring[i-1], ring[i]})
		}
	}
	return edges
}

// signedDistance returns the distance from the polygon's plane to the point, which is negative on the side opposite
// to the normal.
func (p polygon3) signedDistance(pt point3) float64 {
	return pt.sub(p.origin).dot(p.normal)
}

// contains returns whether a point on the polygon's plane lies on the polygon, including its boundary. The test is
// made in the coordinate plane that the polygon's plane faces most directly.
func (p polygon3) contains(pt point3) bool {
	dropped := 0
	for axis := 1; axis < 3; axis++ {
		if math.Abs(p.normal[axis]) > math.Abs(p.normal[dropped]) {
			dropped = axis
		}
	}
	u, v := (dropped+1)%3, (dropped+2)%3
	ringContains := func(ring []point3) (inside bool, onBoundary bool) {
		for i := 1; i < len(ring); i++ {
			a, b := ring[i-1], ring[i]
			cross := (b[u]-a[u])*(pt[v]-a[v]) - (b[v]-a[v])*(pt[u]-a[u])
			if cross == 0 && math.Min(a[u], b[u]) <= pt[u] && pt[u] <= math.Max(a[u], b[u]) &&
				math.Min(a[v], b[v]) <= pt[v] && pt[v] <= math.Max(a[v], b[v]) {
				return true, true
			}
			if (a[v] > pt[v]) != (b[v] > pt[v]) && pt[u] < a[u]+(pt[v]-a[v])*(b[u]-a[u])/(b[v]-a[v]) {
				inside = !inside
			}
		}
		return inside, false
	}
	if inside, _ := ringContains(p.rings[0]); !inside {
		return false
	}
	for _, hole := range p.rings[1:] {
		if inside, onBoundary := ringContains(hole); inside && !onBoundary {
			return false
		}
	}
	return true
}

// closestOnSegment3 returns the point on the segment from `a` to `b` that is closest to `p`.
func closestOnSegment3(p point3, a point3, b point3) point3 {
	d := b.sub(a)
	length := d.dot(d)
	if length == 0 {
		return a
	}
	t := p.sub(a).dot(d) / length
	return a.add(d.scale(math.Max(0, math.Min(1, t))))
}

// closestOnSegments3 returns the closest points on two segments, the first on the segment from `a0` to `a1`.
func closestOnSegments3(a0 point3, a1 point3, b0 point3, b1 point3) (point3, point3) {
	clamp := func(f float64) float64 {
		return math.Max(0, math.Min(1, f))
	}
	da, db, r := a1.sub(a0), b1.sub(b0), a0.sub(b0)
	lengthA, lengthB := da.dot(da), db.dot(db)
	if lengthA == 0 {
		return a0, closestOnSegment3(a0, b0, b1)
	}
	if lengthB == 0 {
		return closestOnSegment3(b0, a0, a1), b0
	}
	c, f, between := da.dot(r), db.dot(r), da.dot(db)
	s := 0.0
	if denominator := lengthA*lengthB - between*between; denominator != 0 {
		s = clamp((between*f - c*lengthB) / denominator)
	}
	t := (between*s + f) / lengthB
	if t < 0 {
		t, s = 0, clamp(-c/lengthA)
	} else if t > 1 {
		t, s = 1, clamp((between-c)/lengthA)
	}
	return a0.add(da.scale(s)), b0.add(db.scale(t))
}

// pair3 is the pair of points that are closest or farthest apart on two geometries, and their distance.
type pair3 struct {
	a, b     point3
	distance float64
	found    bool
}

// consider records a pair of points when it is closer than the current pair, or farther when `farthest` is true,
// measuring in two dimensions when `flat` is true.
func (p *pair3) consider(a point3, b point3, farthest bool, flat bool) {
	d := a.distance(b, flat)
	if !p.found || (farthest && d > p.distance) || (!farthest && d < p.distance) {
		p.a, p.b, p.distance, p.found = a, b, d, true
	}
}

// pointToPolygon records the closest pair from a point to a polygon. A point whose projection onto the polygon's
// plane lies on the polygon is that far from it, and any other point is as far as the polygon's nearest edge.
func (p *pair3) pointToPolygon(pt point3, polygon polygon3, swap bool) {
	consider := func(a point3, b point3) {
		if swap {
			a, b = b, a
		}
		p.consider(a, b, false, false)
	}
	if polygon.normal != (point3{}) {
		projection := pt.sub(polygon.normal.scale(polygon.signedDistance(pt)))
		if polygon.contains(projection) {
			consider(pt, projection)
			return
		}
	}
	for _, edge := range polygon.edges() {
		consider(pt, closestOnSegment3(pt, edge[0], edge[1]))
	}
}

// segmentToPolygon records the closest pair from a segment to a polygon. A segment that crosses the polygon's plane
// at a point on the polygon touches it. When both ends lie strictly on one side of the plane, each end whose
// projection lies on the polygon is that far from it. The segment is also measured against every edge. An end that
// lies on the plane counts only through the edges, as it does in PostGIS.
func (p *pair3) segmentToPolygon(segment [2]point3, polygon polygon3, swap bool) {
	consider := func(a point3, b point3) {
		if swap {
			a, b = b, a
		}
		p.consider(a, b, false, false)
	}
	if polygon.normal != (point3{}) {
		s0, s1 := polygon.signedDistance(segment[0]), polygon.signedDistance(segment[1])
		if s0*s1 < 0 {
			crossing := segment[0].add(segment[1].sub(segment[0]).scale(s0 / (s0 - s1)))
			if polygon.contains(crossing) {
				consider(crossing, crossing)
				return
			}
		} else if s0*s1 > 0 {
			for i, end := range segment {
				projection := end.sub(polygon.normal.scale([]float64{s0, s1}[i]))
				if polygon.contains(projection) {
					consider(end, projection)
				}
			}
		}
	}
	for _, edge := range polygon.edges() {
		a, b := closestOnSegments3(segment[0], segment[1], edge[0], edge[1])
		consider(a, b)
	}
}

// closestPoints3D returns the closest pair of points on two sets of primitives, keeping the first pair when several
// are equally close.
func closestPoints3D(a primitives3D, b primitives3D) pair3 {
	var pair pair3
	for _, pa := range a.points {
		for _, pb := range b.points {
			pair.consider(pa, pb, false, false)
		}
		for _, sb := range b.segments {
			pair.consider(pa, closestOnSegment3(pa, sb[0], sb[1]), false, false)
		}
		for _, polygon := range b.polygons {
			pair.pointToPolygon(pa, polygon, false)
		}
	}
	for _, sa := range a.segments {
		for _, pb := range b.points {
			pair.consider(closestOnSegment3(pb, sa[0], sa[1]), pb, false, false)
		}
		for _, sb := range b.segments {
			onA, onB := closestOnSegments3(sa[0], sa[1], sb[0], sb[1])
			pair.consider(onA, onB, false, false)
		}
		for _, polygon := range b.polygons {
			pair.segmentToPolygon(sa, polygon, false)
		}
	}
	for _, polygonA := range a.polygons {
		for _, pb := range b.points {
			pair.pointToPolygon(pb, polygonA, true)
		}
		for _, sb := range b.segments {
			pair.segmentToPolygon(sb, polygonA, true)
		}
		for _, polygonB := range b.polygons {
			for _, edge := range polygonA.edges() {
				pair.segmentToPolygon(edge, polygonB, false)
			}
			for _, edge := range polygonB.edges() {
				pair.segmentToPolygon(edge, polygonA, true)
			}
		}
	}
	return pair
}

// farthestPoints3D returns the farthest pair of vertices of two sets of primitives, measuring in two dimensions when
// `flat` is true, and keeping the first pair when several are equally far.
func farthestPoints3D(a primitives3D, b primitives3D, flat bool) pair3 {
	var pair pair3
	verticesB := b.vertices()
	for _, va := range a.vertices() {
		for _, vb := range verticesB {
			pair.consider(va, vb, true, flat)
		}
	}
	return pair
}

// zAt returns the Z of the primitives at the given X and Y, taken from the primitive nearest to it in two
// dimensions: a point's own Z, the Z along a segment, or the Z of a polygon's plane when the polygon covers it.
func (prims primitives3D) zAt(xy point3) float64 {
	nearest, z := math.Inf(1), 0.0
	consider := func(p point3) {
		if d := p.distance(xy, true); d < nearest {
			nearest, z = d, p[2]
		}
	}
	alongSegment := func(a point3, b point3) point3 {
		d := b.sub(a)
		d[2] = 0
		t := 0.0
		if length := d.dot(d); length > 0 {
			t = math.Max(0, math.Min(1, ((xy[0]-a[0])*d[0]+(xy[1]-a[1])*d[1])/length))
		}
		return a.add(b.sub(a).scale(t))
	}
	for _, p := range prims.points {
		consider(p)
	}
	for _, s := range prims.segments {
		consider(alongSegment(s[0], s[1]))
	}
	for _, polygon := range prims.polygons {
		if polygon.normal[2] != 0 {
			offset := (xy[0]-polygon.origin[0])*polygon.normal[0] + (xy[1]-polygon.origin[1])*polygon.normal[1]
			onPlane := point3{xy[0], xy[1], polygon.origin[2] - offset/polygon.normal[2]}
			if polygon.contains(onPlane) {
				consider(onPlane)
				continue
			}
		}
		for _, edge := range polygon.edges() {
			consider(alongSegment(edge[0], edge[1]))
		}
	}
	return z
}

// pair3D returns the closest or farthest pair of points on two geometries in three dimensions. When either geometry
// lacks Z, it sends a notice and measures in two dimensions, and each point of the pair takes its Z from the geometry
// that has one. The boolean reports whether both geometries have Z.
func pair3D(ctx *sql.Context, a Geometry, b Geometry, farthest bool) (pair3, bool, error) {
	withZ := a.Shape.hasZ() && b.Shape.hasZ()
	primsA, err := collectPrimitives3D(a.Shape, thirdOrdinate(a.Shape, true, false))
	if err != nil {
		return pair3{}, false, err
	}
	primsB, err := collectPrimitives3D(b.Shape, thirdOrdinate(b.Shape, true, false))
	if err != nil {
		return pair3{}, false, err
	}
	if withZ {
		if farthest {
			return farthestPoints3D(primsA, primsB, false), true, nil
		}
		return closestPoints3D(primsA, primsB), true, nil
	}
	sendNotice(ctx, "NOTICE", `One or both of the geometries is missing z-value. The unknown z-value will be regarded as "any value"`)
	if farthest {
		pair := farthestPoints3D(primsA, primsB, true)
		if !a.Shape.hasZ() {
			pair.a[2] = pair.b[2]
		} else if !b.Shape.hasZ() {
			pair.b[2] = pair.a[2]
		}
		return pair, false, nil
	}
	flat := closestPoints(a.Shape, b.Shape)
	if !flat.found {
		return pair3{}, false, nil
	}
	pair := pair3{a: point3{flat.a[0], flat.a[1]}, b: point3{flat.b[0], flat.b[1]}, distance: flat.distance, found: true}
	if a.Shape.hasZ() {
		pair.a[2] = primsA.zAt(pair.a)
		pair.b[2] = pair.a[2]
	} else if b.Shape.hasZ() {
		pair.b[2] = primsB.zAt(pair.b)
		pair.a[2] = pair.b[2]
	}
	return pair, false, nil
}

// pair3DImpl returns an implementation of a function of the closest or farthest pair of points on two geometries in
// three dimensions, which is null when either geometry is empty.
func pair3DImpl(funcName string, farthest bool, result func(a Geometry, pair pair3, hasZ bool) any) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		if err := checkSameSRID(funcName, a, b); err != nil {
			return nil, err
		}
		pair, withZ, err := pair3D(ctx, a, b, farthest)
		if err != nil || !pair.found {
			return nil, err
		}
		return result(a, pair, withZ || a.Shape.hasZ() || b.Shape.hasZ()), nil
	}
}

// ordinates returns the ordinates of the point and their coordinates type, keeping Z when `hasZ` is true.
func (p point3) ordinates(hasZ bool) ([]float64, coordinatesType) {
	if hasZ {
		return []float64{p[0], p[1], p[2]}, dimXYZ
	}
	return []float64{p[0], p[1]}, dimXY
}

// Each of these builds the result of a closest or farthest pair function in three dimensions from the pair.
var (
	pair3DPoint = func(a Geometry, pair pair3, hasZ bool) any {
		coords, ctype := pair.a.ordinates(hasZ)
		return a.withShape(pointShape(coords, ctype, 0))
	}
	pair3DLine = func(a Geometry, pair pair3, hasZ bool) any {
		start, ctype := pair.a.ordinates(hasZ)
		end, _ := pair.b.ordinates(hasZ)
		return a.withShape(shape{typeNum: lineType, ctype: ctype, coords: append(start, end...)})
	}
	pair3DDistance = func(a Geometry, pair pair3, hasZ bool) any {
		return pair.distance
	}
)

// within3DImpl returns an implementation of ST_3DDWithin, ST_3DDFullyWithin, or ST_3DIntersects, which compare the
// distance between the closest points, or the farthest when `farthest` is true, with a tolerance. ST_3DIntersects
// has no tolerance argument and uses zero. Empty geometries are never within any distance.
func within3DImpl(funcName string, farthest bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		if err := checkSameSRID(funcName, a, b); err != nil {
			return nil, err
		}
		tolerance := 0.0
		if len(args) > 2 {
			tolerance = args[2].(float64)
		}
		pair, _, err := pair3D(ctx, a, b, farthest)
		if err != nil {
			return nil, err
		}
		return pair.found && pair.distance <= tolerance, nil
	}
}

// geometryDistanceND implements the <<->> operator, which measures the distance between two geometries in every
// dimension that both of them have. Empty geometries are zero apart.
func geometryDistanceND(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	useZ, useM := a.Shape.hasZ() && b.Shape.hasZ(), a.Shape.hasM() && b.Shape.hasM()
	if !useZ && !useM {
		pair := closestPoints(a.Shape, b.Shape)
		if !pair.found {
			return 0.0, nil
		}
		return pair.distance, nil
	}
	if useZ && useM && a.Shape.typeNum == pointType && b.Shape.typeNum == pointType && !a.Shape.isEmpty() && !b.Shape.isEmpty() {
		var sum float64
		for i := range 4 {
			d := a.Shape.coords[i] - b.Shape.coords[i]
			sum += d * d
		}
		return math.Sqrt(sum), nil
	}
	//TODO: measure in four dimensions between geometries other than points that both have Z and M
	primsA, err := collectPrimitives3D(a.Shape, thirdOrdinate(a.Shape, useZ, useM))
	if err != nil {
		return nil, err
	}
	primsB, err := collectPrimitives3D(b.Shape, thirdOrdinate(b.Shape, useZ, useM))
	if err != nil {
		return nil, err
	}
	pair := closestPoints3D(primsA, primsB)
	if !pair.found {
		return 0.0, nil
	}
	return pair.distance, nil
}

// distance3DRoutines returns the routines that measure geometries in three dimensions.
func distance3DRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	withTolerance := append(append([]extdef.Parameter(nil), geomParams...), extdef.Parameter{Type: "float8"})
	dwithin := within3DImpl("LWGEOM_dwithin3d", false)
	dfullywithin := within3DImpl("LWGEOM_dfullywithin3d", true)
	intersects := within3DImpl("ST_3DIntersects", false)
	return []extdef.Routine{
		{Name: "_st_3ddfullywithin", Symbol: "_st_3ddfullywithin_geometry_geometry_float8", Parameters: withTolerance, Returns: "bool", Strict: true, Impl: dfullywithin},
		{Name: "_st_3ddwithin", Symbol: "_st_3ddwithin_geometry_geometry_float8", Parameters: withTolerance, Returns: "bool", Strict: true, Impl: dwithin},
		{Name: "_st_3dintersects", Symbol: "_st_3dintersects_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: intersects},
		{Name: "geometry_distance_centroid_nd", Symbol: "gserialized_distance_nd", Parameters: params("geometry", "geometry"), Returns: "float8", Strict: true, Impl: geometryDistanceND},
		{Name: "st_3dclosestpoint", Symbol: "LWGEOM_closestpoint3d", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: pair3DImpl("LWGEOM_closestpoint3d", false, pair3DPoint)},
		{Name: "st_3ddfullywithin", Symbol: "st_3ddfullywithin_geometry_geometry_float8", Parameters: withTolerance, Returns: "bool", Strict: true, Impl: dfullywithin},
		{Name: "st_3ddistance", Symbol: "ST_3DDistance", Parameters: geomParams, Returns: "float8", Strict: true, Impl: pair3DImpl("ST_3DDistance", false, pair3DDistance)},
		{Name: "st_3ddwithin", Symbol: "st_3ddwithin_geometry_geometry_float8", Parameters: withTolerance, Returns: "bool", Strict: true, Impl: dwithin},
		{Name: "st_3dintersects", Symbol: "st_3dintersects_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: intersects},
		{Name: "st_3dlongestline", Symbol: "LWGEOM_longestline3d", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: pair3DImpl("LWGEOM_longestline3d", true, pair3DLine)},
		{Name: "st_3dmaxdistance", Symbol: "LWGEOM_maxdistance3d", Parameters: geomParams, Returns: "float8", Strict: true, Impl: pair3DImpl("LWGEOM_maxdistance3d", true, pair3DDistance)},
		{Name: "st_3dshortestline", Symbol: "LWGEOM_shortestline3d", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: pair3DImpl("LWGEOM_shortestline3d", false, pair3DLine)},
	}
}
