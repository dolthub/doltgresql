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

package spatial

// This file ports more GEOS 3.11.1 operations: normalization, the discrete Hausdorff and Fréchet distances, and the
// minimum bounding circle.

import (
	"math"
	"sort"

	"github.com/cockroachdb/errors"
)

// geosVersion is the version of GEOS whose behavior the ported operations follow.
const geosVersion = "3.11.1-CAPI-1.17.1"

// GEOSVersion returns the version of GEOS whose behavior the ported operations follow.
func GEOSVersion() string {
	return geosVersion
}

// dims returns the number of ordinates per position of a geometry's coordinates.
func (g Geometry) dims() int {
	if g.HasZ {
		return 3
	}
	return 2
}

// isEmpty returns whether the geometry has no coordinates.
func (g Geometry) isEmpty() bool {
	if len(g.Coords) > 0 {
		return false
	}
	for _, member := range g.Members {
		if !member.isEmpty() {
			return false
		}
	}
	return true
}

// sortIndex is GEOS's SORTINDEX of a geometry type.
func sortIndex(t GeometryType) int {
	return map[GeometryType]int{Point: 0, MultiPoint: 1, LineString: 2, LinearRing: 3, MultiLineString: 4, Polygon: 5,
		MultiPolygon: 6, GeometryCollection: 7}[t]
}

// compareCoords compares two positions by X and then Y, as Coordinate::compareTo.
func compareCoords(a []float64, b []float64) int {
	switch {
	case a[0] < b[0]:
		return -1
	case a[0] > b[0]:
		return 1
	case a[1] < b[1]:
		return -1
	case a[1] > b[1]:
		return 1
	}
	return 0
}

// compareGeometries is GEOS's Geometry::compareTo.
func compareGeometries(a Geometry, b Geometry) int {
	if ia, ib := sortIndex(a.Type), sortIndex(b.Type); ia != ib {
		return signum(float64(ia - ib))
	}
	emptyA, emptyB := a.isEmpty(), b.isEmpty()
	switch {
	case emptyA && emptyB:
		return 0
	case emptyA:
		return -1
	case emptyB:
		return 1
	}
	switch a.Type {
	case Point:
		return compareCoords(a.Coords, b.Coords)
	case LineString, LinearRing:
		da, db := a.dims(), b.dims()
		na, nb := len(a.Coords)/da, len(b.Coords)/db
		if na != nb {
			return signum(float64(na - nb))
		}
		for i := 0; i < na; i++ {
			if c := compareCoords(a.Coords[i*da:], b.Coords[i*db:]); c != 0 {
				return c
			}
		}
		return 0
	case Polygon:
		if c := compareGeometries(a.Members[0], b.Members[0]); c != 0 {
			return c
		}
		if len(a.Members) != len(b.Members) {
			return signum(float64(len(a.Members) - len(b.Members)))
		}
		for i := 1; i < len(a.Members); i++ {
			if c := compareGeometries(a.Members[i], b.Members[i]); c != 0 {
				return c
			}
		}
		return 0
	}
	for i := 0; i < len(a.Members) && i < len(b.Members); i++ {
		if c := compareGeometries(a.Members[i], b.Members[i]); c != 0 {
			return c
		}
	}
	return signum(float64(len(a.Members) - len(b.Members)))
}

// isCCW is GEOS's Orientation::isCCW of a closed ring, which returns false for flat rings.
func isCCW(ring []xy) bool {
	nPts := len(ring) - 1
	upHiPt := ring[0]
	var upLowPt xy
	prevY := upHiPt.y
	iUpHi := 0
	for i := 1; i <= nPts; i++ {
		py := ring[i].y
		if py > prevY && py >= upHiPt.y {
			iUpHi = i
			upHiPt = ring[i]
			upLowPt = ring[i-1]
		}
		prevY = py
	}
	if iUpHi == 0 {
		return false
	}
	iDownLow := iUpHi
	for {
		iDownLow = (iDownLow + 1) % nPts
		if iDownLow == iUpHi || ring[iDownLow].y != upHiPt.y {
			break
		}
	}
	downLowPt := ring[iDownLow]
	iDownHi := nPts - 1
	if iDownLow > 0 {
		iDownHi = iDownLow - 1
	}
	downHiPt := ring[iDownHi]
	equal := func(a xy, b xy) bool { return a.x == b.x && a.y == b.y }
	if equal(upHiPt, downHiPt) {
		if equal(upLowPt, upHiPt) || equal(downLowPt, upHiPt) || equal(upLowPt, downLowPt) {
			return false
		}
		return orientationIndex(upLowPt, upHiPt, downLowPt) == 1
	}
	return downHiPt.x-upHiPt.x < 0
}

// normalizeRing is GEOS's normalization of a closed sequence: it starts at its least position and runs clockwise
// when `clockwise` is true and counterclockwise otherwise.
func normalizeRing(coords []float64, dims int, clockwise bool, requireRing bool) ([]float64, error) {
	n := len(coords)/dims - 1
	if n < 0 {
		return coords, nil
	}
	if requireRing && n < 3 {
		return nil, errors.New("IllegalArgumentException: Ring has fewer than 4 points, so orientation cannot be determined")
	}
	minIndex := 0
	for i := 1; i < n; i++ {
		if compareCoords(coords[i*dims:], coords[minIndex*dims:]) < 0 {
			minIndex = i
		}
	}
	result := make([]float64, 0, len(coords))
	result = append(result, coords[minIndex*dims:n*dims]...)
	result = append(result, coords[:minIndex*dims]...)
	result = append(result, result[:dims]...)
	ring := make([]xy, n+1)
	for i := range ring {
		ring[i] = xy{x: result[i*dims], y: result[i*dims+1]}
	}
	if (requireRing || n+1 >= 4) && isCCW(ring) == clockwise {
		for i, j := 0, n; i < j; i, j = i+1, j-1 {
			for d := 0; d < dims; d++ {
				result[i*dims+d], result[j*dims+d] = result[j*dims+d], result[i*dims+d]
			}
		}
	}
	return result, nil
}

// normalize is GEOS's Geometry::normalize.
func normalize(g Geometry) (Geometry, error) {
	dims := g.dims()
	switch g.Type {
	case Point:
		return g, nil
	case LineString, LinearRing:
		n := len(g.Coords) / dims
		if n == 0 {
			return g, nil
		}
		if compareCoords(g.Coords, g.Coords[(n-1)*dims:]) == 0 {
			coords, err := normalizeRing(g.Coords, dims, true, false)
			if err != nil {
				return Geometry{}, err
			}
			g.Coords = coords
			return g, nil
		}
		for i := 0; i < n/2; i++ {
			j := n - 1 - i
			if c := compareCoords(g.Coords[i*dims:], g.Coords[j*dims:]); c != 0 {
				if c > 0 {
					reversed := make([]float64, len(g.Coords))
					for k := 0; k < n; k++ {
						copy(reversed[k*dims:(k+1)*dims], g.Coords[(n-1-k)*dims:(n-k)*dims])
					}
					g.Coords = reversed
				}
				break
			}
		}
		return g, nil
	case Polygon:
		members := make([]Geometry, len(g.Members))
		for i, ring := range g.Members {
			coords, err := normalizeRing(ring.Coords, dims, i == 0, true)
			if err != nil {
				return Geometry{}, err
			}
			members[i] = Geometry{Type: LinearRing, HasZ: ring.HasZ, Coords: coords}
		}
		if len(members) > 1 {
			holes := members[1:]
			sort.SliceStable(holes, func(i, j int) bool { return compareGeometries(holes[i], holes[j]) > 0 })
		}
		g.Members = members
		return g, nil
	}
	members := make([]Geometry, len(g.Members))
	for i, member := range g.Members {
		normalized, err := normalize(member)
		if err != nil {
			return Geometry{}, err
		}
		members[i] = normalized
	}
	sort.SliceStable(members, func(i, j int) bool { return compareGeometries(members[i], members[j]) > 0 })
	g.Members = members
	return g, nil
}

// Normalize returns the geometry in GEOS's normal form.
func (g *GEOS) Normalize(a Geometry) (Geometry, error) {
	return normalize(a)
}

// pointDistance is GEOS's Coordinate::distance.
func pointDistance(a xy, b xy) float64 {
	dx := a.x - b.x
	dy := a.y - b.y
	return math.Sqrt(float64(dx*dx) + float64(dy*dy))
}

// closestPoint is GEOS's LineSegment::closestPoint.
func closestPoint(p0 xy, p1 xy, p xy) xy {
	var factor float64
	switch {
	case p.x == p0.x && p.y == p0.y:
		factor = 0
	case p.x == p1.x && p.y == p1.y:
		factor = 1
	default:
		dx := p1.x - p0.x
		dy := p1.y - p0.y
		len2 := float64(dx*dx) + float64(dy*dy)
		factor = (float64((p.x-p0.x)*dx) + float64((p.y-p0.y)*dy)) / len2
	}
	if factor > 0 && factor < 1 {
		return xy{x: p0.x + float64(factor*(p1.x-p0.x)), y: p0.y + float64(factor*(p1.y-p0.y))}
	}
	if pointDistance(p0, p) < pointDistance(p1, p) {
		return p0
	}
	return p1
}

// minimumDistanceTo is GEOS's DistanceToPoint::computeDistance, which returns the least distance from a point to a
// geometry, measuring polygons by their rings, and false when the geometry is empty.
func minimumDistanceTo(g Geometry, p xy) (float64, bool) {
	best, found := math.Inf(1), false
	consider := func(d float64) {
		if !found || d < best {
			best, found = d, true
		}
	}
	var visit func(g Geometry)
	visit = func(g Geometry) {
		switch g.Type {
		case LineString, LinearRing:
			points := verticesOf(g)
			for i := 1; i < len(points); i++ {
				consider(pointDistance(closestPoint(points[i-1], points[i], p), p))
			}
		case Polygon:
			for _, ring := range g.Members {
				visit(ring)
			}
		case Point:
			if len(g.Coords) >= 2 {
				consider(pointDistance(xy{x: g.Coords[0], y: g.Coords[1]}, p))
			}
		default:
			for _, member := range g.Members {
				visit(member)
			}
		}
	}
	visit(g)
	return best, found
}

// sequences returns the coordinate sequences of a geometry, in order.
func sequences(g Geometry) [][]xy {
	if g.Type == Point || g.Type == LineString || g.Type == LinearRing {
		if len(g.Coords) == 0 {
			return nil
		}
		return [][]xy{verticesOf(g)}
	}
	var result [][]xy
	for _, member := range g.Members {
		result = append(result, sequences(member)...)
	}
	return result
}

// javaRound is GEOS's java_math_round.
func javaRound(v float64) float64 {
	n, frac := math.Modf(v)
	f := math.Abs(frac)
	if v >= 0 {
		switch {
		case f < 0.5:
			return math.Floor(v)
		case f > 0.5:
			return math.Ceil(v)
		}
		return n + 1.0
	}
	switch {
	case f < 0.5:
		return math.Ceil(v)
	case f > 0.5:
		return math.Floor(v)
	}
	return n
}

// subSegments returns the number of parts that a densify fraction splits segments into, or an error when the fraction
// is out of range.
func subSegments(densifyFraction float64) (int, error) {
	if densifyFraction > 1.0 || !(densifyFraction > 0.0) {
		return 0, errors.New("IllegalArgumentException: Fraction is not in range (0.0 - 1.0]")
	}
	return int(javaRound(1.0 / densifyFraction)), nil
}

// orientedHausdorff returns the greatest distance from the vertices of a geometry, and from the points that densify
// its segments when numSubSegs is positive, to another geometry.
func orientedHausdorff(discrete Geometry, other Geometry, numSubSegs int, max *float64, found *bool) {
	if discrete.isEmpty() || other.isEmpty() {
		return
	}
	consider := func(p xy) {
		if d, ok := minimumDistanceTo(other, p); ok && (!*found || d > *max) {
			*max, *found = d, true
		}
	}
	for _, seq := range sequences(discrete) {
		for _, p := range seq {
			consider(p)
		}
	}
	if numSubSegs > 0 {
		for _, seq := range sequences(discrete) {
			for index := 1; index < len(seq); index++ {
				p0, p1 := seq[index-1], seq[index]
				delx := (p1.x - p0.x) / float64(numSubSegs)
				dely := (p1.y - p0.y) / float64(numSubSegs)
				for i := 0; i < numSubSegs; i++ {
					consider(xy{x: p0.x + float64(float64(i)*delx), y: p0.y + float64(float64(i)*dely)})
				}
			}
		}
	}
}

// HausdorffDistance returns GEOS's discrete Hausdorff distance between two geometries, densifying their segments by
// the fraction when it is positive.
func (g *GEOS) HausdorffDistance(a Geometry, b Geometry, densifyFraction float64) (float64, error) {
	numSubSegs := 0
	if densifyFraction > 0 {
		var err error
		if numSubSegs, err = subSegments(densifyFraction); err != nil {
			return 0, err
		}
	}
	var max float64
	found := false
	orientedHausdorff(a, b, numSubSegs, &max, &found)
	orientedHausdorff(b, a, numSubSegs, &max, &found)
	if !found {
		return math.NaN(), nil
	}
	return max, nil
}

// FrechetDistance returns GEOS's discrete Fréchet distance between two geometries, densifying their segments by the
// fraction when it is positive.
func (g *GEOS) FrechetDistance(a Geometry, b Geometry, densifyFraction float64) (float64, error) {
	numSubSegs := 0
	if densifyFraction > 0 {
		var err error
		if numSubSegs, err = subSegments(densifyFraction); err != nil {
			return 0, err
		}
	}
	var lp, lq []xy
	for _, seq := range sequences(a) {
		lp = append(lp, seq...)
	}
	for _, seq := range sequences(b) {
		lq = append(lq, seq...)
	}
	if len(lp) == 0 || len(lq) == 0 {
		return math.NaN(), nil
	}
	pointAt := func(seq []xy, index int) xy {
		if numSubSegs == 0 {
			return seq[index]
		}
		i, j := index/numSubSegs, index%numSubSegs
		if i >= len(seq)-1 {
			return seq[len(seq)-1]
		}
		p0, p1 := seq[i], seq[i+1]
		delx := (p1.x - p0.x) / float64(numSubSegs)
		dely := (p1.y - p0.y) / float64(numSubSegs)
		return xy{x: p0.x + float64(float64(j)*delx), y: p0.y + float64(float64(j)*dely)}
	}
	pSize, qSize := len(lp), len(lq)
	if numSubSegs > 0 {
		pSize = numSubSegs*(len(lp)-1) + 1
		qSize = numSubSegs*(len(lq)-1) + 1
	}
	// The coupling distances are computed in the order that GEOS's memoized recursion fills them in, keeping the
	// larger of the best predecessor and the distance between the current points
	larger := func(previous float64, d float64) float64 {
		if previous > d {
			return previous
		}
		return d
	}
	ca := make([][]float64, pSize)
	for i := range ca {
		ca[i] = make([]float64, qSize)
	}
	for i := 0; i < pSize; i++ {
		for j := 0; j < qSize; j++ {
			d := pointDistance(pointAt(lp, i), pointAt(lq, j))
			switch {
			case i == 0 && j == 0:
				ca[i][j] = d
			case j == 0:
				ca[i][j] = larger(ca[i-1][0], d)
			case i == 0:
				ca[i][j] = larger(ca[0][j-1], d)
			default:
				minDist := ca[i-1][j]
				if !(ca[i-1][j] < ca[i-1][j-1]) {
					minDist = ca[i-1][j-1]
				}
				if ca[i][j-1] < minDist {
					minDist = ca[i][j-1]
				}
				ca[i][j] = larger(minDist, d)
			}
		}
	}
	return ca[pSize-1][qSize-1], nil
}

// MinimumBoundingCircle returns GEOS's minimum bounding circle of a geometry as a polygon, along with its center and
// radius.
func (g *GEOS) MinimumBoundingCircle(a Geometry) (circle Geometry, center Geometry, radius float64, err error) {
	var extremal []xy
	if !a.isEmpty() {
		points := uniqueVertices(a)
		if countPoints(a) == 1 {
			extremal = points[:1]
		} else {
			hull := convexHull(a)
			var pts []xy
			if hull.Type == Polygon {
				pts = verticesOf(hull.Members[0])
			} else {
				pts = verticesOf(hull)
			}
			if len(pts) > 1 && pts[0].x == pts[len(pts)-1].x && pts[0].y == pts[len(pts)-1].y {
				pts = pts[:len(pts)-1]
			}
			if extremal, err = circlePoints(pts); err != nil {
				return Geometry{}, Geometry{}, 0, err
			}
		}
	}
	var centre xy
	switch len(extremal) {
	case 0:
		return Geometry{Type: Polygon}, Geometry{Type: Point}, 0, nil
	case 1:
		centre = extremal[0]
	case 2:
		centre = xy{x: (extremal[0].x + extremal[1].x) / 2.0, y: (extremal[0].y + extremal[1].y) / 2.0}
	default:
		centre = circumcentre(extremal[0], extremal[1], extremal[2])
	}
	radius = pointDistance(centre, extremal[0])
	center = Geometry{Type: Point, Coords: []float64{centre.x, centre.y}}
	if radius == 0.0 {
		return center, center, radius, nil
	}
	circle, err = g.Buffer(center, radius, BufferParams{QuadSegs: 8, EndCap: BufferRound, Join: BufferRound, MitreLimit: 5})
	return circle, center, radius, err
}

// countPoints returns the number of positions that a geometry has.
func countPoints(g Geometry) int {
	count := len(g.Coords) / g.dims()
	for _, member := range g.Members {
		count += countPoints(member)
	}
	return count
}

// circlePoints is MinimumBoundingCircle::computeCirclePoints, which returns the two or three points of a convex hull
// that determine its minimum bounding circle.
func circlePoints(pts []xy) ([]xy, error) {
	if len(pts) <= 2 {
		return pts, nil
	}
	p := pts[0]
	for _, pt := range pts {
		if pt.y < p.y {
			p = pt
		}
	}
	equal := func(a xy, b xy) bool { return a.x == b.x && a.y == b.y }
	minSin := math.Inf(1)
	var q xy
	for _, pt := range pts {
		if equal(pt, p) {
			continue
		}
		dx := pt.x - p.x
		dy := math.Abs(pt.y - p.y)
		sin := dy / math.Sqrt(float64(dx*dx)+float64(dy*dy))
		if sin < minSin {
			minSin, q = sin, pt
		}
	}
	isObtuse := func(p0 xy, p1 xy, p2 xy) bool {
		return float64((p0.x-p1.x)*(p2.x-p1.x))+float64((p0.y-p1.y)*(p2.y-p1.y)) < 0
	}
	angleBetween := func(tip1 xy, tail xy, tip2 xy) float64 {
		a1 := math.Atan2(tip1.y-tail.y, tip1.x-tail.x)
		a2 := math.Atan2(tip2.y-tail.y, tip2.x-tail.x)
		d := a1 - a2
		if a1 < a2 {
			d = a2 - a1
		}
		if d > math.Pi {
			d = 2*math.Pi - d
		}
		return d
	}
	for i := 0; i < len(pts); i++ {
		minAng := math.Inf(1)
		r := pts[0]
		for _, pt := range pts {
			if equal(pt, p) || equal(pt, q) {
				continue
			}
			if ang := angleBetween(p, pt, q); ang < minAng {
				minAng, r = ang, pt
			}
		}
		switch {
		case isObtuse(p, r, q):
			return []xy{p, q}, nil
		case isObtuse(r, p, q):
			p = r
		case isObtuse(r, q, p):
			q = r
		default:
			return []xy{p, q, r}, nil
		}
	}
	return nil, errors.New("Logic failure in MinimumBoundingCircle algorithm!")
}

// circumcentre is GEOS's Triangle::circumcentre.
func circumcentre(a xy, b xy, c xy) xy {
	det := func(m00 float64, m01 float64, m10 float64, m11 float64) float64 {
		return float64(m00*m11) - float64(m01*m10)
	}
	ax, ay := a.x-c.x, a.y-c.y
	bx, by := b.x-c.x, b.y-c.y
	denom := 2 * det(ax, ay, bx, by)
	numx := det(ay, float64(ax*ax)+float64(ay*ay), by, float64(bx*bx)+float64(by*by))
	numy := det(ax, float64(ax*ax)+float64(ay*ay), bx, float64(bx*bx)+float64(by*by))
	return xy{x: c.x - numx/denom, y: c.y + numy/denom}
}
