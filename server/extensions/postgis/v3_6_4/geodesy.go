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

// sphereRadius is the radius of the sphere that PostGIS measures geographies on when told not to use the spheroid:
// the mean of the WGS84 spheroid's three semi-axes.
var sphereRadius = (2*spatial.WGS84.Radius() + spatial.WGS84.Radius()*(1-spatial.WGS84.Flattening())) / 3

// vec3 is a point on the unit sphere in Cartesian coordinates.
type vec3 [3]float64

// toVec3 returns the point on the unit sphere at a longitude and latitude in degrees.
func toVec3(point []float64) vec3 {
	lon, lat := point[0]*math.Pi/180, point[1]*math.Pi/180
	return vec3{math.Cos(lat) * math.Cos(lon), math.Cos(lat) * math.Sin(lon), math.Sin(lat)}
}

// toLonLat returns the longitude and latitude in degrees of a point on the unit sphere.
func (v vec3) toLonLat() []float64 {
	return []float64{math.Atan2(v[1], v[0]) * 180 / math.Pi, math.Asin(v[2]) * 180 / math.Pi}
}

// dot returns the dot product of two vectors.
func (v vec3) dot(o vec3) float64 {
	return float64(v[0]*o[0]) + float64(v[1]*o[1]) + float64(v[2]*o[2])
}

// cross returns the cross product of two vectors.
func (v vec3) cross(o vec3) vec3 {
	return vec3{float64(v[1]*o[2]) - float64(v[2]*o[1]), float64(v[2]*o[0]) - float64(v[0]*o[2]), float64(v[0]*o[1]) - float64(v[1]*o[0])}
}

// normalized returns the vector scaled to unit length, or false when it has no length.
func (v vec3) normalized() (vec3, bool) {
	length := math.Sqrt(v.dot(v))
	if length == 0 {
		return v, false
	}
	return vec3{v[0] / length, v[1] / length, v[2] / length}, true
}

// angleTo returns the angle between two unit vectors, which is their great-circle distance on the unit sphere.
func (v vec3) angleTo(o vec3) float64 {
	c := v.cross(o)
	return math.Atan2(math.Sqrt(c.dot(c)), v.dot(o))
}

// closestOnArc returns the point of the shorter great-circle arc from `a` to `b` that is closest to `p`.
func closestOnArc(p vec3, a vec3, b vec3) vec3 {
	normal, ok := a.cross(b).normalized()
	if !ok {
		return a
	}
	along := p.dot(normal)
	projected, ok := vec3{p[0] - float64(along*normal[0]), p[1] - float64(along*normal[1]), p[2] - float64(along*normal[2])}.normalized()
	if ok && a.cross(projected).dot(normal) >= 0 && projected.cross(b).dot(normal) >= 0 {
		return projected
	}
	if p.angleTo(a) <= p.angleTo(b) {
		return a
	}
	return b
}

// closestOnArcSegment returns the point on the arc from `a` to `b` that is closest to `p`, as in closestOnArc. When that
// point is an end of the arc, it also returns the end's coordinates from the segment, and otherwise nil.
func closestOnArcSegment(p vec3, a vec3, b vec3, segment [2][]float64) (vec3, []float64) {
	closest := closestOnArc(p, a, b)
	if closest == a {
		return closest, segment[0]
	}
	if closest == b {
		return closest, segment[1]
	}
	return closest, nil
}

// arcCrossing returns the point where two great-circle arcs cross, or false when they do not.
func arcCrossing(a1 vec3, a2 vec3, b1 vec3, b2 vec3) (vec3, bool) {
	na, nb := a1.cross(a2), b1.cross(b2)
	line, ok := na.cross(nb).normalized()
	if !ok {
		return vec3{}, false
	}
	for _, candidate := range []vec3{line, {-line[0], -line[1], -line[2]}} {
		if a1.cross(candidate).dot(na) >= 0 && candidate.cross(a2).dot(na) >= 0 &&
			b1.cross(candidate).dot(nb) >= 0 && candidate.cross(b2).dot(nb) >= 0 {
			return candidate, true
		}
	}
	return vec3{}, false
}

// insideSphericalRing returns whether a point lies inside a ring on the sphere, by summing the signed angles that
// the ring's edges turn through as seen from the point.
func insideSphericalRing(p vec3, ring []float64, dims int) bool {
	total := 0.0
	for i := dims; i+dims <= len(ring); i += dims {
		a, b := p.cross(toVec3(ring[i-dims:])), p.cross(toVec3(ring[i:]))
		total += math.Atan2(a.cross(b).dot(p), a.dot(b))
	}
	return math.Abs(total) > math.Pi
}

// insideSphericalShape returns whether a point lies inside a polygon of the shape: inside an exterior ring and
// outside its interior rings.
func insideSphericalShape(p vec3, s shape) bool {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case polygonType:
		if len(s.rings) == 0 || !insideSphericalRing(p, s.rings[0], dims) {
			return false
		}
		for _, hole := range s.rings[1:] {
			if insideSphericalRing(p, hole, dims) {
				return false
			}
		}
		return true
	case triangleType:
		return len(s.coords) > 0 && insideSphericalRing(p, s.coords, dims)
	default:
		for _, member := range s.members {
			if insideSphericalShape(p, member) {
				return true
			}
		}
		return false
	}
}

// sphericalClosest returns the closest pair of points on two geographies, found on the sphere, or false when either
// is empty. A vertex inside a polygon of the other geography is a pair with itself.
func sphericalClosest(a shape, b shape) ([]float64, []float64, bool) {
	segmentsA, segmentsB := segmentsOf(a), segmentsOf(b)
	if len(segmentsA) == 0 || len(segmentsB) == 0 {
		return nil, nil, false
	}
	for _, segment := range segmentsB {
		if insideSphericalShape(toVec3(segment[0]), a) {
			return segment[0], segment[0], true
		}
	}
	for _, segment := range segmentsA {
		if insideSphericalShape(toVec3(segment[0]), b) {
			return segment[0], segment[0], true
		}
	}
	var bestA, bestB []float64
	best := math.Inf(1)
	consider := func(pa vec3, lonLatA []float64, pb vec3, lonLatB []float64) {
		if d := pa.angleTo(pb); d < best {
			if lonLatA == nil {
				lonLatA = pa.toLonLat()
			}
			if lonLatB == nil {
				lonLatB = pb.toLonLat()
			}
			bestA, bestB, best = lonLatA, lonLatB, d
		}
	}
	for _, sa := range segmentsA {
		a1, a2 := toVec3(sa[0]), toVec3(sa[1])
		for _, sb := range segmentsB {
			b1, b2 := toVec3(sb[0]), toVec3(sb[1])
			if crossing, ok := arcCrossing(a1, a2, b1, b2); ok {
				consider(crossing, nil, crossing, nil)
			}
			closestB1, lonLatB1 := closestOnArcSegment(a1, b1, b2, sb)
			consider(a1, sa[0], closestB1, lonLatB1)
			closestB2, lonLatB2 := closestOnArcSegment(a2, b1, b2, sb)
			consider(a2, sa[1], closestB2, lonLatB2)
			closestA1, lonLatA1 := closestOnArcSegment(b1, a1, a2, sa)
			consider(closestA1, lonLatA1, b1, sb[0])
			closestA2, lonLatA2 := closestOnArcSegment(b2, a1, a2, sa)
			consider(closestA2, lonLatA2, b2, sb[1])
		}
	}
	return bestA, bestB, true
}

// geodesicDistance returns the distance in meters between two points given in degrees, on the WGS84 spheroid or on
// the sphere.
func geodesicDistance(a []float64, b []float64, useSpheroid bool) float64 {
	if !useSpheroid {
		return toVec3(a).angleTo(toVec3(b)) * sphereRadius
	}
	return spheroidDistance(spatial.WGS84, a, b)
}

// spheroidDistance returns the distance in meters between two points given in degrees, on the ellipsoid. The degrees
// pass through radians first, as PostGIS stores geographic points in radians.
func spheroidDistance(ellipsoid *spatial.Ellipsoid, a []float64, b []float64) float64 {
	if ellipsoid.Flattening() == 0 {
		radius := ellipsoid.Radius()
		return (2*radius + radius) / 3 * greatCircleAngle(a, b)
	}
	viaRadians := func(degrees float64) float64 {
		return degrees * math.Pi / 180 * 180 / math.Pi
	}
	distance, _, _ := ellipsoid.Inverse(viaRadians(a[1]), viaRadians(a[0]), viaRadians(b[1]), viaRadians(b[0]))
	return distance
}

// greatCircleAngle returns the angle in radians between two points given in degrees, along the great circle through
// them. A latitude past a pole is folded back over it, keeping its longitude, as PostGIS does. Each product is rounded
// on its own, so that a fused multiply-add cannot change the result.
func greatCircleAngle(a []float64, b []float64) float64 {
	foldLatitude := func(lat float64) float64 {
		if lat > math.Pi/2 {
			return math.Pi - lat
		}
		if lat < -math.Pi/2 {
			return -math.Pi - lat
		}
		return lat
	}
	lonA, latA := a[0]*math.Pi/180, foldLatitude(a[1]*math.Pi/180)
	lonB, latB := b[0]*math.Pi/180, foldLatitude(b[1]*math.Pi/180)
	sinLatA, cosLatA := math.Sin(latA), math.Cos(latA)
	sinLatB, cosLatB := math.Sin(latB), math.Cos(latB)
	deltaLon := lonB - lonA
	cosDeltaLon := math.Cos(deltaLon)
	across := float64(cosLatB * math.Sin(deltaLon))
	along := float64(cosLatA*sinLatB) - float64(float64(sinLatA*cosLatB)*cosDeltaLon)
	toward := float64(sinLatA*sinLatB) + float64(float64(cosLatA*cosLatB)*cosDeltaLon)
	return math.Atan2(math.Sqrt(float64(across*across)+float64(along*along)), toward)
}

// geographyDistance returns the distance in meters between two geographies, or false when either is empty.
func geographyDistance(a shape, b shape, useSpheroid bool) (float64, bool) {
	pa, pb, ok := sphericalClosest(a, b)
	if !ok {
		return 0, false
	}
	return geodesicDistance(pa, pb, useSpheroid), true
}

// useSpheroidArg returns the optional use_spheroid argument at the index, which defaults to true.
func useSpheroidArg(args []any, index int) bool {
	return len(args) <= index || args[index].(bool)
}

// geographyDistanceImpl implements ST_Distance for geography, rounded to ten nanometers as PostGIS does.
func geographyDistanceImpl(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("geography_distance", a, b); err != nil {
		return nil, err
	}
	distance, ok := geographyDistance(a.Shape, b.Shape, useSpheroidArg(args, 2))
	if !ok {
		return nil, nil
	}
	return math.Round(distance*1e8) / 1e8, nil
}

// geographyDWithin implements ST_DWithin for geography.
func geographyDWithin(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("geography_dwithin", a, b); err != nil {
		return nil, err
	}
	distance, ok := geographyDistance(a.Shape, b.Shape, useSpheroidArg(args, 3))
	return ok && distance <= args[2].(float64), nil
}

// geodesicLength returns the length in meters of a sequence of points in degrees.
func geodesicLength(seq []float64, dims int, useSpheroid bool) float64 {
	length := 0.0
	for i := dims; i+dims <= len(seq); i += dims {
		length += geodesicDistance(seq[i-dims:], seq[i:], useSpheroid)
	}
	return length
}

// geographyLengths returns the total length of a geography's linestrings and the total perimeter of its polygons.
func geographyLengths(s shape, useSpheroid bool) (float64, float64) {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case lineType:
		return geodesicLength(s.coords, dims, useSpheroid), 0
	case polygonType:
		perimeter := 0.0
		for _, ring := range s.rings {
			perimeter += geodesicLength(ring, dims, useSpheroid)
		}
		return 0, perimeter
	case triangleType:
		return 0, geodesicLength(s.coords, dims, useSpheroid)
	default:
		var length, perimeter float64
		for _, member := range s.members {
			l, p := geographyLengths(member, useSpheroid)
			length, perimeter = length+l, perimeter+p
		}
		return length, perimeter
	}
}

// ringArea returns the area in square meters enclosed by a ring of points in degrees, on the WGS84 spheroid or on
// the sphere.
func ringArea(ring []float64, dims int, useSpheroid bool) float64 {
	if !useSpheroid {
		excess := 0.0
		first := toVec3(ring)
		for i := 2 * dims; i+dims <= len(ring); i += dims {
			b, c := toVec3(ring[i-dims:]), toVec3(ring[i:])
			excess += 2 * math.Atan2(first.dot(b.cross(c)), 1+first.dot(b)+b.dot(c)+c.dot(first))
		}
		return math.Abs(excess) * sphereRadius * sphereRadius
	}
	var lats, lons []float64
	for i := 0; i+dims <= len(ring); i += dims {
		lats, lons = append(lats, ring[i+1]), append(lons, ring[i])
	}
	return math.Abs(spatial.WGS84.PolygonArea(lats, lons))
}

// geographyArea returns the area of a geography's polygons in square meters.
func geographyArea(s shape, useSpheroid bool) float64 {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case polygonType:
		area := 0.0
		for i, ring := range s.rings {
			if i == 0 {
				area += ringArea(ring, dims, useSpheroid)
			} else {
				area -= ringArea(ring, dims, useSpheroid)
			}
		}
		return area
	case triangleType:
		if len(s.coords) == 0 {
			return 0
		}
		return ringArea(s.coords, dims, useSpheroid)
	default:
		area := 0.0
		for _, member := range s.members {
			area += geographyArea(member, useSpheroid)
		}
		return area
	}
}

// geographyAzimuth implements ST_Azimuth for geography, the initial bearing of the geodesic between two points in
// radians, from zero up to two pi.
func geographyAzimuth(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	pa, err := pointArg(ctx, a, "ST_Azimuth: Arguments must be POINT geometries")
	if err != nil {
		return nil, err
	}
	pb, err := pointArg(ctx, b, "ST_Azimuth: Arguments must be POINT geometries")
	if err != nil {
		return nil, err
	}
	if pa[0] == pb[0] && pa[1] == pb[1] {
		return nil, nil
	}
	_, azimuth, _ := spatial.WGS84.Inverse(pa[1], pa[0], pb[1], pb[0])
	azimuth = azimuth * math.Pi / 180
	if azimuth < 0 {
		azimuth += 2 * math.Pi
	}
	return azimuth, nil
}

// geographyProject implements ST_Project for geography, which follows the geodesic from a point for the distance in
// the direction of the azimuth in radians. Given a second point instead, it moves the second point for the distance
// in the direction that the geodesic from the first point to the second starts out in.
func geographyProject(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil || args[1] == nil || args[2] == nil {
		return nil, nil
	}
	g := args[0].(Geometry)
	origin, err := pointArg(ctx, g, "ST_Project(geography) is only valid for point inputs")
	if err != nil {
		return nil, err
	}
	var distance, azimuth float64
	if target, ok := args[1].(Geometry); ok {
		to, err := pointArg(ctx, target, "ST_Project(geography) is only valid for point inputs")
		if err != nil {
			return nil, err
		}
		_, azimuth, _ = spatial.WGS84.Inverse(origin[1], origin[0], to[1], to[0])
		distance = args[2].(float64)
		g, origin = target, to
	} else {
		distance, azimuth = args[1].(float64), args[2].(float64)*180/math.Pi
	}
	lat, lon := spatial.WGS84.Direct(origin[1], origin[0], azimuth, distance)
	projected := append([]float64{lon, lat}, origin[2:]...)
	return g.withShape(shape{typeNum: pointType, ctype: g.Shape.ctype, coords: projected}), nil
}

// geographyCovers returns whether the first geography covers the second: every point of the second lies inside a
// polygon of the first without any edge crossing the polygon's boundary, or on the first's points or linestrings.
func geographyCovers(a shape, b shape) bool {
	if a.isEmpty() || b.isEmpty() {
		return false
	}
	if a.dimension() == 2 {
		for _, sb := range segmentsOf(b) {
			if !insideSphericalShape(toVec3(sb[0]), a) || !insideSphericalShape(toVec3(sb[1]), a) {
				return false
			}
			for _, sa := range segmentsOf(a) {
				crossing, ok := arcCrossing(toVec3(sa[0]), toVec3(sa[1]), toVec3(sb[0]), toVec3(sb[1]))
				if ok && crossing.angleTo(toVec3(sb[0])) > 0 && crossing.angleTo(toVec3(sb[1])) > 0 {
					return false
				}
			}
		}
		return true
	}
	if b.dimension() > a.dimension() {
		return false
	}
	for _, vertex := range dumpPoints(b, nil) {
		if distance, ok := geographyDistance(a, vertex.part, false); !ok || distance > 0 {
			return false
		}
	}
	return true
}

// geographyPredicate returns an implementation of a spatial relationship test between geographies.
func geographyPredicate(funcName string, test func(a shape, b shape) bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		if err := checkSameSRID(funcName, a, b); err != nil {
			return nil, err
		}
		return test(a.Shape, b.Shape), nil
	}
}

// geographyClosestImpl returns an implementation of ST_ClosestPoint or ST_ShortestLine for geography, where `line`
// chooses the shortest line.
func geographyClosestImpl(line bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		if err := checkSameSRID("geography_closestpoint", a, b); err != nil {
			return nil, err
		}
		pa, pb, ok := sphericalClosest(a.Shape, b.Shape)
		if !ok {
			return nil, nil
		}
		if line {
			return a.withShape(shape{typeNum: lineType, coords: append(append([]float64(nil), pa...), pb...)}), nil
		}
		return a.withShape(shape{typeNum: pointType, coords: pa}), nil
	}
}

// slerp returns the point at the fraction of the great-circle arc from `a` to `b`, whose angle is `angle`.
func slerp(a vec3, b vec3, angle float64, fraction float64) vec3 {
	sinAngle := math.Sin(angle)
	wa, wb := math.Sin((1-fraction)*angle)/sinAngle, math.Sin(fraction*angle)/sinAngle
	return vec3{float64(wa*a[0]) + float64(wb*b[0]), float64(wa*a[1]) + float64(wb*b[1]), float64(wa*a[2]) + float64(wb*b[2])}
}

// bisectGeography appends the points that divide the great-circle segment from `start` to `end` by halving it until
// every piece is at most the maximum length on the sphere. Z and M coordinates take the mean of their ends.
func bisectGeography(dst []float64, start []float64, end []float64, maxLength float64) []float64 {
	a, b := toVec3(start), toVec3(end)
	if a.angleTo(b)*sphereRadius <= maxLength {
		return dst
	}
	middle, ok := vec3{a[0] + b[0], a[1] + b[1], a[2] + b[2]}.normalized()
	if !ok {
		return dst
	}
	mid := middle.toLonLat()
	for j := 2; j < len(start); j++ {
		mid = append(mid, (start[j]+end[j])/2)
	}
	dst = bisectGeography(dst, start, mid, maxLength)
	dst = append(dst, mid...)
	return bisectGeography(dst, mid, end, maxLength)
}

// segmentizeGeographySequence returns a copy of a sequence of points in degrees where every great-circle segment
// longer than the maximum length on the sphere is halved until no piece is longer.
func segmentizeGeographySequence(seq []float64, dims int, maxLength float64) []float64 {
	if len(seq) < 2*dims {
		return seq
	}
	segmentized := append([]float64(nil), seq[:dims]...)
	for i := dims; i+dims <= len(seq); i += dims {
		segmentized = bisectGeography(segmentized, seq[i-dims:i], seq[i:i+dims], maxLength)
		segmentized = append(segmentized, seq[i:i+dims]...)
	}
	return segmentized
}

// segmentizeGeography returns a copy of the shape with every linestring and ring segmentized on the sphere.
func segmentizeGeography(s shape, maxLength float64) shape {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case lineType:
		s.coords = segmentizeGeographySequence(s.coords, dims, maxLength)
	case polygonType:
		rings := make([][]float64, len(s.rings))
		for i, ring := range s.rings {
			rings[i] = segmentizeGeographySequence(ring, dims, maxLength)
		}
		s.rings = rings
	case pointType, multiPointType:
	default:
		members := make([]shape, len(s.members))
		for i, member := range s.members {
			members[i] = segmentizeGeography(member, maxLength)
		}
		s.members = members
	}
	return s
}

// geographyCentroid returns the centroid of a geography's parts of highest dimension on the sphere: the mean
// direction of its points, of the midpoints of its segments weighted by their lengths, or of the centers of the
// triangles that fan out from each ring's first point weighted by their areas, which count against the ring when they
// turn the other way and against the polygon in a hole. It returns false for an empty geography.
func geographyCentroid(s shape, useSpheroid bool) ([]float64, bool) {
	var sum vec3
	add := func(v vec3, weight float64) {
		sum = vec3{sum[0] + float64(weight*v[0]), sum[1] + float64(weight*v[1]), sum[2] + float64(weight*v[2])}
	}
	dims := s.ctype.Dimension()
	switch s.dimension() {
	case 0:
		for _, segment := range segmentsOf(s) {
			add(toVec3(segment[0]), 1)
		}
	case 1:
		for _, segment := range segmentsOf(s) {
			if middle, ok := toVec3(segment[0]).plus(toVec3(segment[1])).normalized(); ok {
				add(middle, geodesicDistance(segment[0], segment[1], useSpheroid))
			}
		}
	default:
		var addRing func(ring []float64, sign float64)
		addRing = func(ring []float64, sign float64) {
			sign *= math.Copysign(1, signedRingArea(ring, s.ctype))
			first := toVec3(ring)
			for i := 2 * dims; i+dims <= len(ring); i += dims {
				triangle := append(append(append([]float64(nil), ring[:dims]...), ring[i-dims:i]...), ring[i:i+dims]...)
				triangle = append(triangle, ring[:dims]...)
				area := ringArea(triangle, dims, useSpheroid)
				if center, ok := first.plus(toVec3(ring[i-dims:])).plus(toVec3(ring[i:])).normalized(); ok {
					add(center, sign*area*math.Copysign(1, signedRingArea(triangle, s.ctype)))
				}
			}
		}
		var visit func(member shape)
		visit = func(member shape) {
			switch member.typeNum {
			case polygonType:
				for i, ring := range member.rings {
					sign := 1.0
					if i > 0 {
						sign = -1
					}
					addRing(ring, sign)
				}
			case triangleType:
				if len(member.coords) > 0 {
					addRing(member.coords, 1)
				}
			default:
				for _, child := range member.members {
					visit(child)
				}
			}
		}
		visit(s)
	}
	centroid, ok := sum.normalized()
	if !ok {
		return nil, false
	}
	return centroid.toLonLat(), true
}

// plus returns the sum of two vectors.
func (v vec3) plus(o vec3) vec3 {
	return vec3{v[0] + o[0], v[1] + o[1], v[2] + o[2]}
}

// geographySegment is a segment of a geography's linestring with its length in meters.
type geographySegment struct {
	start, end []float64
	length     float64
}

// geographySegments returns the segments of a geography linestring with their lengths, and their total length.
func geographySegments(line shape, useSpheroid bool) ([]geographySegment, float64) {
	dims := line.ctype.Dimension()
	var segments []geographySegment
	total := 0.0
	for i := dims; i+dims <= len(line.coords); i += dims {
		segment := geographySegment{start: line.coords[i-dims : i], end: line.coords[i : i+dims]}
		segment.length = geodesicDistance(segment.start, segment.end, useSpheroid)
		segments = append(segments, segment)
		total += segment.length
	}
	return segments, total
}

// geographyPointAlong returns the point at the distance along a geography linestring's segments, following the
// geodesic or the great circle of the segment that holds it. A distance past the end returns the last point.
func geographyPointAlong(segments []geographySegment, distance float64, useSpheroid bool) []float64 {
	traveled := 0.0
	for _, segment := range segments {
		if distance < traveled+segment.length {
			fraction := (distance - traveled) / segment.length
			var point []float64
			if useSpheroid {
				lat, lon := spatial.WGS84.Position(segment.start[1], segment.start[0], segment.end[1], segment.end[0], fraction*segment.length)
				point = []float64{lon, lat}
			} else {
				a, b := toVec3(segment.start), toVec3(segment.end)
				point = slerp(a, b, a.angleTo(b), fraction).toLonLat()
			}
			for j := 2; j < len(segment.start); j++ {
				point = append(point, segment.start[j]+float64((segment.end[j]-segment.start[j])*fraction))
			}
			return point
		}
		traveled += segment.length
	}
	last := segments[len(segments)-1].end
	return append([]float64(nil), last...)
}

// geographyLineInterpolate implements ST_LineInterpolatePoint and ST_LineInterpolatePoints for geography.
func geographyLineInterpolate(ctx *sql.Context, args ...any) (any, error) {
	g, fraction := args[0].(Geometry), args[1].(float64)
	useSpheroid := useSpheroidArg(args, 2)
	repeat := len(args) > 3 && args[3].(bool)
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
	segments, total := geographySegments(line, useSpheroid)
	if len(segments) == 0 {
		return g.withShape(pointShape(line.coords, line.ctype, 0)), nil
	}
	count := 1
	if repeat && fraction > 0 {
		count = int(math.Floor(1 / fraction))
	}
	points := shape{typeNum: multiPointType, ctype: line.ctype}
	for i := 1; i <= count; i++ {
		coords := geographyPointAlong(segments, float64(i)*fraction*total, useSpheroid)
		points.members = append(points.members, shape{typeNum: pointType, ctype: line.ctype, coords: coords})
	}
	if count == 1 {
		return g.withShape(points.members[0]), nil
	}
	return g.withShape(points), nil
}

// geographyLineLocate implements ST_LineLocatePoint for geography, which returns the fraction of a linestring's
// length at which it comes closest to a point on the sphere.
func geographyLineLocate(ctx *sql.Context, args ...any) (any, error) {
	g, p := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("geography_line_locate_point", g, p); err != nil {
		return nil, err
	}
	if g.Shape.typeNum != lineType {
		return nil, pgerror.New(pgcode.Internal, "line_locate_point: 1st arg isn't a line")
	}
	if p.Shape.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, "line_locate_point: 2nd arg isn't a point")
	}
	if g.Shape.isEmpty() || p.Shape.isEmpty() {
		return 0.0, nil
	}
	useSpheroid := useSpheroidArg(args, 2)
	segments, total := geographySegments(g.Shape, useSpheroid)
	if total == 0 {
		return 1.0, nil
	}
	target := toVec3(p.Shape.coords)
	best, located, traveled := math.Inf(1), 0.0, 0.0
	for _, segment := range segments {
		closest := closestOnArc(target, toVec3(segment.start), toVec3(segment.end))
		if d := target.angleTo(closest); d < best {
			best = d
			located = traveled + geodesicDistance(segment.start, closest.toLonLat(), useSpheroid)
		}
		traveled += segment.length
	}
	return located / total, nil
}

// geographyLineSubstring implements ST_LineSubstring for geography.
func geographyLineSubstring(ctx *sql.Context, args ...any) (any, error) {
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
	line := g.Shape
	if line.typeNum != lineType {
		return nil, pgerror.New(pgcode.Internal, "line_substring: 1st arg isn't a line")
	}
	if line.isEmpty() {
		return nil, nil
	}
	segments, total := geographySegments(line, true)
	if len(segments) == 0 {
		return g.withShape(pointShape(line.coords, line.ctype, 0)), nil
	}
	start, end := from*total, to*total
	coords := geographyPointAlong(segments, start, true)
	traveled := 0.0
	for _, segment := range segments {
		traveled += segment.length
		if traveled > start && traveled < end {
			coords = appendDistinct(coords, segment.end, line.ctype)
		}
	}
	coords = appendDistinct(coords, geographyPointAlong(segments, end, true), line.ctype)
	if len(coords) == line.ctype.Dimension() {
		return g.withShape(shape{typeNum: pointType, ctype: line.ctype, coords: coords}), nil
	}
	return g.withShape(shape{typeNum: lineType, ctype: line.ctype, coords: coords}), nil
}

// geodesyRoutines returns the routines that measure geographies on the WGS84 spheroid or a sphere.
func geodesyRoutines() []extdef.Routine {
	useSpheroid := extdef.Parameter{Name: "use_spheroid", Type: "bool", Default: `true`}
	geogParams := []extdef.Parameter{{Name: "geog1", Type: "geography"}, {Name: "geog2", Type: "geography"}}
	covers := geographyPredicate("geography_covers", geographyCovers)
	coveredBy := geographyPredicate("geography_coveredby", func(a shape, b shape) bool {
		return geographyCovers(b, a)
	})
	uncachedDistance := func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		if err := checkSameSRID("geography_distance_uncached", a, b); err != nil {
			return nil, err
		}
		useSpheroid := true
		if len(args) > 2 {
			useSpheroid = args[len(args)-1].(bool)
		}
		distance, ok := geographyDistance(a.Shape, b.Shape, useSpheroid)
		if !ok {
			return nil, nil
		}
		return distance, nil
	}
	treeDistance := func(ctx *sql.Context, args ...any) (any, error) {
		if len(args) == 4 {
			args = []any{args[0], args[1], args[3]}
		}
		return geographyDistanceImpl(ctx, args...)
	}
	return []extdef.Routine{
		{Name: "_st_coveredby", Symbol: "_st_coveredby_geography_geography", Parameters: geogParams, Returns: "bool", Strict: true, Impl: coveredBy},
		{Name: "_st_covers", Symbol: "_st_covers_geography_geography", Parameters: geogParams, Returns: "bool", Strict: true, Impl: covers},
		{Name: "_st_distancetree", Symbol: "_st_distancetree_geography_geography", Parameters: params("geography", "geography"), Returns: "float8", Strict: true, Impl: treeDistance},
		{Name: "_st_distancetree", Symbol: "geography_distance_tree", Parameters: params("geography", "geography", "float8", "bool"), Returns: "float8", Strict: true, Impl: treeDistance},
		{Name: "_st_distanceuncached", Symbol: "_st_distanceuncached_geography_geography", Parameters: params("geography", "geography"), Returns: "float8", Strict: true, Impl: uncachedDistance},
		{Name: "_st_distanceuncached", Symbol: "_st_distanceuncached_geography_geography_bool", Parameters: params("geography", "geography", "bool"), Returns: "float8", Strict: true, Impl: uncachedDistance},
		{Name: "_st_distanceuncached", Symbol: "geography_distance_uncached", Parameters: params("geography", "geography", "float8", "bool"), Returns: "float8", Strict: true, Impl: uncachedDistance},
		{Name: "_st_dwithin", Symbol: "_st_dwithin_geography_geography_float8_bool", Parameters: append(geogParams, extdef.Parameter{Name: "tolerance", Type: "float8"}, useSpheroid), Returns: "bool", Strict: true, Impl: geographyDWithin},
		{Name: "_st_dwithinuncached", Symbol: "_st_dwithinuncached_geography_geography_float8", Parameters: params("geography", "geography", "float8"), Returns: "bool", Impl: func(ctx *sql.Context, args ...any) (any, error) {
			if args[0] == nil || args[1] == nil || args[2] == nil {
				return nil, nil
			}
			return geographyDWithin(ctx, args...)
		}},
		{Name: "_st_dwithinuncached", Symbol: "geography_dwithin_uncached", Parameters: params("geography", "geography", "float8", "bool"), Returns: "bool", Strict: true, Impl: geographyDWithin},
		{Name: "st_area", Symbol: "geography_area", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, useSpheroid}, Returns: "float8", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return geographyArea(args[0].(Geometry).Shape, useSpheroidArg(args, 1)), nil
		}},
		{Name: "st_azimuth", Symbol: "geography_azimuth", Parameters: []extdef.Parameter{{Name: "geog1", Type: "geography"}, {Name: "geog2", Type: "geography"}}, Returns: "float8", Strict: true, Impl: geographyAzimuth},
		{Name: "st_distance", Symbol: "geography_distance", Parameters: []extdef.Parameter{{Name: "geog1", Type: "geography"}, {Name: "geog2", Type: "geography"}, useSpheroid}, Returns: "float8", Strict: true, Impl: geographyDistanceImpl},
		{Name: "st_dwithin", Symbol: "st_dwithin_geography_geography_float8_bool", Parameters: []extdef.Parameter{{Name: "geog1", Type: "geography"}, {Name: "geog2", Type: "geography"}, {Name: "tolerance", Type: "float8"}, useSpheroid}, Returns: "bool", Strict: true, Impl: geographyDWithin},
		{Name: "st_length", Symbol: "geography_length", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, useSpheroid}, Returns: "float8", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			length, _ := geographyLengths(args[0].(Geometry).Shape, useSpheroidArg(args, 1))
			return length, nil
		}},
		{Name: "st_perimeter", Symbol: "geography_perimeter", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, useSpheroid}, Returns: "float8", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			_, perimeter := geographyLengths(args[0].(Geometry).Shape, useSpheroidArg(args, 1))
			return perimeter, nil
		}},
		{Name: "st_project", Symbol: "geography_project", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, {Name: "distance", Type: "float8"}, {Name: "azimuth", Type: "float8"}}, Returns: "geography", Impl: geographyProject},
		{Name: "st_project", Symbol: "geography_project_geography", Parameters: []extdef.Parameter{{Name: "geog_from", Type: "geography"}, {Name: "geog_to", Type: "geography"}, {Name: "distance", Type: "float8"}}, Returns: "geography", Strict: true, Impl: geographyProject},
		{Name: "st_centroid", Symbol: "geography_centroid", Parameters: []extdef.Parameter{{Type: "geography"}, useSpheroid}, Returns: "geography", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			g := args[0].(Geometry)
			centroid, ok := geographyCentroid(g.Shape, useSpheroidArg(args, 1))
			if !ok {
				return g.withShape(shape{typeNum: pointType}), nil
			}
			return g.withShape(shape{typeNum: pointType, coords: centroid}), nil
		}},
		{Name: "st_lineinterpolatepoint", Symbol: "st_lineinterpolatepoint_geography_float8_bool", Parameters: []extdef.Parameter{{Type: "geography"}, {Type: "float8"}, useSpheroid}, Returns: "geography", Strict: true, Impl: geographyLineInterpolate},
		{Name: "st_lineinterpolatepoints", Symbol: "st_lineinterpolatepoints_geography_float8_bool_bool", Parameters: []extdef.Parameter{{Type: "geography"}, {Type: "float8"}, useSpheroid, {Name: "repeat", Type: "bool", Default: `true`}}, Returns: "geography", Strict: true, Impl: geographyLineInterpolate},
		{Name: "st_linelocatepoint", Symbol: "geography_line_locate_point", Parameters: []extdef.Parameter{{Type: "geography"}, {Type: "geography"}, useSpheroid}, Returns: "float8", Strict: true, Impl: geographyLineLocate},
		{Name: "st_linesubstring", Symbol: "geography_line_substring", Parameters: params("geography", "float8", "float8"), Returns: "geography", Strict: true, Impl: geographyLineSubstring},
		{Name: "geometrytype", Symbol: "geometrytype_geography", Parameters: params("geography"), Returns: "text", Strict: true, Impl: geometryType},
		{Name: "st_closestpoint", Symbol: "geography_closestpoint", Parameters: []extdef.Parameter{{Type: "geography"}, {Type: "geography"}, useSpheroid}, Returns: "geography", Strict: true, Impl: geographyClosestImpl(false)},
		{Name: "st_coveredby", Symbol: "geography_coveredby", Parameters: []extdef.Parameter{{Name: "geog1", Type: "geography"}, {Name: "geog2", Type: "geography"}}, Returns: "bool", Strict: true, Impl: coveredBy},
		{Name: "st_covers", Symbol: "geography_covers", Parameters: []extdef.Parameter{{Name: "geog1", Type: "geography"}, {Name: "geog2", Type: "geography"}}, Returns: "bool", Strict: true, Impl: covers},
		{Name: "st_intersects", Symbol: "geography_intersects", Parameters: []extdef.Parameter{{Name: "geog1", Type: "geography"}, {Name: "geog2", Type: "geography"}}, Returns: "bool", Strict: true, Impl: geographyPredicate("geography_intersects", func(a shape, b shape) bool {
			distance, ok := geographyDistance(a, b, false)
			return ok && distance < 0.00001
		})},
		{Name: "st_segmentize", Symbol: "geography_segmentize", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, {Name: "max_segment_length", Type: "float8"}}, Returns: "geography", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			g := args[0].(Geometry)
			return g.withShape(segmentizeGeography(g.Shape, args[1].(float64))), nil
		}},
		{Name: "st_shortestline", Symbol: "geography_shortestline", Parameters: []extdef.Parameter{{Type: "geography"}, {Type: "geography"}, useSpheroid}, Returns: "geography", Strict: true, Impl: geographyClosestImpl(true)},
		{Name: "geography_distance_knn", Symbol: "geography_distance_knn", Parameters: params("geography", "geography"), Returns: "float8", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			a, b := args[0].(Geometry), args[1].(Geometry)
			if err := checkSameSRID("geography_distance_knn", a, b); err != nil {
				return nil, err
			}
			distance, ok := geographyDistance(a.Shape, b.Shape, false)
			if !ok {
				return nil, nil
			}
			return distance, nil
		}},
		{Name: "st_distancesphere", Symbol: "st_distancesphere_geometry_geometry", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "float8", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			distance, ok := geographyDistance(args[0].(Geometry).Shape, args[1].(Geometry).Shape, false)
			if !ok {
				return nil, nil
			}
			return math.Round(distance*1e8) / 1e8, nil
		}},
		{Name: "st_distancesphere", Symbol: "LWGEOM_distance_sphere", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}, {Name: "radius", Type: "float8"}}, Returns: "float8", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return stDistanceSpheroid(ctx, args[0], args[1], spheroid{a: args[2].(float64)})
		}},
	}
}
