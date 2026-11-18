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

// This file ports the GEOS 3.11.1 algorithms whose results simplefeatures computes differently, so that results agree
// with GEOS's, including the order and starting point of their coordinates. Products are converted to float64
// explicitly wherever a fused multiply-add would change the rounding.

import (
	"math"
	"sort"
)

// xy is a two-dimensional coordinate that remembers the Z of the vertex it came from.
type xy struct {
	x, y, z float64
}

// dd is a double-double number, as GEOS's geos::math::DD.
type dd struct {
	hi, lo float64
}

// ddSplit is the splitting constant of GEOS's DD multiplication, 2^27+1.
const ddSplit = 134217729.0

// add returns the sum of two double-double numbers, as DD::selfAdd.
func (a dd) add(b dd) dd {
	S := a.hi + b.hi
	T := a.lo + b.lo
	e := S - a.hi
	f := T - a.lo
	s := S - e
	t := T - f
	s = (b.hi - e) + (a.hi - s)
	t = (b.lo - f) + (a.lo - t)
	e = s + T
	H := S + e
	h := e + (S - H)
	e = t + h
	zhi := H + e
	zlo := e + (H - zhi)
	return dd{zhi, zlo}
}

// mul returns the product of two double-double numbers, as DD::selfMultiply.
func (a dd) mul(b dd) dd {
	C := float64(ddSplit * a.hi)
	hx := C - a.hi
	c := float64(ddSplit * b.hi)
	hx = C - hx
	tx := a.hi - hx
	hy := c - b.hi
	C = float64(a.hi * b.hi)
	hy = c - hy
	ty := b.hi - hy
	c = ((((float64(hx*hy) - C) + float64(hx*ty)) + float64(tx*hy)) + float64(tx*ty)) + (float64(a.hi*b.lo) + float64(a.lo*b.hi))
	zhi := C + c
	hx = C - zhi
	zlo := c + hx
	return dd{zhi, zlo}
}

// orientationIndex returns whether q lies to the left (1) or right (-1) of the line from p1 to p2, or on it (0), with
// GEOS's robust CGAlgorithmsDD::orientationIndex.
func orientationIndex(p1 xy, p2 xy, q xy) int {
	if index := orientationIndexFilter(p1, p2, q); index <= 1 {
		return index
	}
	dx1 := dd{p2.x, 0}.add(dd{-p1.x, 0})
	dy1 := dd{p2.y, 0}.add(dd{-p1.y, 0})
	dx2 := dd{q.x, 0}.add(dd{-p2.x, 0})
	dy2 := dd{q.y, 0}.add(dd{-p2.y, 0})
	mx1y2 := dx1.mul(dy2)
	my1x2 := dy1.mul(dx2)
	d := mx1y2.add(dd{-my1x2.hi, -my1x2.lo})
	switch {
	case d.hi < 0 || (d.hi == 0 && d.lo < 0):
		return -1
	case d.hi > 0 || (d.hi == 0 && d.lo > 0):
		return 1
	}
	return 0
}

// orientationIndexFilter is GEOS's orientationIndexFilter, which returns 2 when double precision cannot decide.
func orientationIndexFilter(pa xy, pb xy, pc xy) int {
	const dpSafeEpsilon = 1e-15
	detleft := float64((pa.x - pc.x) * (pb.y - pc.y))
	detright := float64((pa.y - pc.y) * (pb.x - pc.x))
	det := detleft - detright
	var detsum float64
	switch {
	case detleft > 0.0:
		if detright <= 0.0 {
			return signum(det)
		}
		detsum = detleft + detright
	case detleft < 0.0:
		if detright >= 0.0 {
			return signum(det)
		}
		detsum = -detleft - detright
	default:
		return signum(det)
	}
	errbound := float64(dpSafeEpsilon * detsum)
	if det >= errbound || -det >= errbound {
		return signum(det)
	}
	return 2
}

// signum returns the sign of a number.
func signum(x float64) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// inRing returns whether a point lies in or on a closed ring, as GEOS's PointLocation::isInRing.
func inRing(p xy, ring []xy) bool {
	crossings := 0
	for i := 1; i < len(ring); i++ {
		p1, p2 := ring[i-1], ring[i]
		if p1.x < p.x && p2.x < p.x {
			continue
		}
		if p.x == p2.x && p.y == p2.y {
			return true
		}
		if p1.y == p.y && p2.y == p.y {
			minx, maxx := p1.x, p2.x
			if minx > maxx {
				minx, maxx = p2.x, p1.x
			}
			if p.x >= minx && p.x <= maxx {
				return true
			}
			continue
		}
		if (p1.y > p.y && p2.y <= p.y) || (p2.y > p.y && p1.y <= p.y) {
			sign := orientationIndex(p1, p2, p)
			if sign == 0 {
				return true
			}
			if p2.y < p1.y {
				sign = -sign
			}
			if sign > 0 {
				crossings++
			}
		}
	}
	return crossings%2 == 1
}

// uniqueVertices returns the distinct vertices of a geometry, by their X and Y, in the order that they first appear.
func uniqueVertices(g Geometry) []xy {
	var points []xy
	seen := make(map[[2]float64]bool)
	var visit func(g Geometry)
	visit = func(g Geometry) {
		dims := 2
		if g.HasZ {
			dims = 3
		}
		for i := 0; i+dims <= len(g.Coords); i += dims {
			p := xy{x: g.Coords[i], y: g.Coords[i+1], z: math.NaN()}
			if g.HasZ {
				p.z = g.Coords[i+2]
			}
			key := [2]float64{p.x, p.y}
			if !seen[key] {
				seen[key] = true
				points = append(points, p)
			}
		}
		for _, member := range g.Members {
			visit(member)
		}
	}
	visit(g)
	return points
}

// geometryFromPoints returns a point, linestring, or polygon of the given vertices, keeping Z when the geometry has
// it.
func geometryFromPoints(typ GeometryType, points []xy, hasZ bool) Geometry {
	result := Geometry{Type: typ, HasZ: hasZ}
	coords := make([]float64, 0, len(points)*3)
	for _, p := range points {
		coords = append(coords, p.x, p.y)
		if hasZ {
			coords = append(coords, p.z)
		}
	}
	if typ == Polygon {
		result.Members = []Geometry{{Type: LinearRing, HasZ: hasZ, Coords: coords}}
	} else {
		result.Coords = coords
	}
	return result
}

// convexHull is GEOS's ConvexHull::getConvexHull.
func convexHull(g Geometry) Geometry {
	points := uniqueVertices(g)
	switch len(points) {
	case 0:
		return Geometry{Type: GeometryCollection}
	case 1:
		return geometryFromPoints(Point, points, g.HasZ)
	case 2:
		return geometryFromPoints(LineString, points, g.HasZ)
	}
	if len(points) > 50 {
		points = reduceHullPoints(points)
	}
	points = sortRadially(points)
	hull := grahamScan(points)
	cleaned := cleanRing(hull)
	if len(cleaned) == 3 {
		return geometryFromPoints(LineString, cleaned[:2], g.HasZ)
	}
	return geometryFromPoints(Polygon, cleaned, g.HasZ)
}

// reduceHullPoints keeps the points that lie outside the octagon of extreme points, along with the octagon's own, as
// ConvexHull::reduce does.
func reduceHullPoints(points []xy) []xy {
	oct := make([]int, 8)
	for i := 1; i < len(points); i++ {
		p := points[i]
		if p.x < points[oct[0]].x {
			oct[0] = i
		}
		if p.x-p.y < points[oct[1]].x-points[oct[1]].y {
			oct[1] = i
		}
		if p.y > points[oct[2]].y {
			oct[2] = i
		}
		if p.x+p.y > points[oct[3]].x+points[oct[3]].y {
			oct[3] = i
		}
		if p.x > points[oct[4]].x {
			oct[4] = i
		}
		if p.x-p.y > points[oct[5]].x-points[oct[5]].y {
			oct[5] = i
		}
		if p.y < points[oct[6]].y {
			oct[6] = i
		}
		if p.x+p.y < points[oct[7]].x+points[oct[7]].y {
			oct[7] = i
		}
	}
	var ring []int
	for i, index := range oct {
		if i == 0 || index != oct[i-1] {
			ring = append(ring, index)
		}
	}
	if len(ring) < 3 {
		return points
	}
	ring = append(ring, ring[0])
	ringPoints := make([]xy, len(ring))
	for i, index := range ring {
		ringPoints[i] = points[index]
	}
	reduced := make(map[[2]float64]xy)
	for _, p := range ringPoints {
		reduced[[2]float64{p.x, p.y}] = p
	}
	for _, p := range points {
		if !inRing(p, ringPoints) {
			key := [2]float64{p.x, p.y}
			if _, ok := reduced[key]; !ok {
				reduced[key] = p
			}
		}
	}
	result := make([]xy, 0, len(reduced))
	for _, p := range reduced {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].x != result[j].x {
			return result[i].x < result[j].x
		}
		return result[i].y < result[j].y
	})
	for len(result) < 3 {
		result = append(result, result[0])
	}
	return result
}

// sortRadially moves the lowest point first and sorts the rest by their angle around it, as ConvexHull::preSort.
func sortRadially(points []xy) []xy {
	points = append([]xy(nil), points...)
	for i := 1; i < len(points); i++ {
		p0, pi := points[0], points[i]
		if pi.y < p0.y || (pi.y == p0.y && pi.x < p0.x) {
			points[0], points[i] = pi, p0
		}
	}
	origin := points[0]
	polarCompare := func(p xy, q xy) int {
		switch orientationIndex(origin, p, q) {
		case 1:
			return 1
		case -1:
			return -1
		}
		if p.y > q.y {
			return 1
		}
		if p.y < q.y {
			return -1
		}
		if p.x > q.x {
			return 1
		}
		if p.x < q.x {
			return -1
		}
		return 0
	}
	sort.SliceStable(points, func(i, j int) bool {
		return polarCompare(points[i], points[j]) == -1
	})
	return points
}

// grahamScan is ConvexHull::grahamScan, which returns the closed hull of radially sorted points.
func grahamScan(c []xy) []xy {
	ps := []xy{c[0], c[1], c[2]}
	for i := 3; i < len(c); i++ {
		p := ps[len(ps)-1]
		ps = ps[:len(ps)-1]
		for len(ps) > 0 && orientationIndex(ps[len(ps)-1], p, c[i]) > 0 {
			p = ps[len(ps)-1]
			ps = ps[:len(ps)-1]
		}
		ps = append(ps, p, c[i])
	}
	return append(ps, c[0])
}

// isBetween is ConvexHull::isBetween.
func isBetween(c1 xy, c2 xy, c3 xy) bool {
	if orientationIndex(c1, c2, c3) != 0 {
		return false
	}
	if c1.x != c3.x {
		if (c1.x <= c2.x && c2.x <= c3.x) || (c3.x <= c2.x && c2.x <= c1.x) {
			return true
		}
	}
	if c1.y != c3.y {
		if (c1.y <= c2.y && c2.y <= c3.y) || (c3.y <= c2.y && c2.y <= c1.y) {
			return true
		}
	}
	return false
}

// cleanRing is ConvexHull::cleanRing, which drops repeated and collinear points from a closed ring.
func cleanRing(original []xy) []xy {
	last := original[len(original)-1]
	var cleaned []xy
	var prev *xy
	for i := 0; i < len(original)-1; i++ {
		curr, next := original[i], original[i+1]
		if curr.x == next.x && curr.y == next.y {
			continue
		}
		if prev != nil && isBetween(*prev, curr, next) {
			continue
		}
		cleaned = append(cleaned, curr)
		prev = &cleaned[len(cleaned)-1]
	}
	return append(cleaned, last)
}

// minimumRectangle is GEOS's MinimumDiameter::getMinimumRectangle, which returns the rectangle of least width that
// holds a geometry.
func minimumRectangle(g Geometry) Geometry {
	hull := convexHull(g)
	var pts []xy
	switch hull.Type {
	case Polygon:
		pts = verticesOf(hull.Members[0])
	default:
		pts = verticesOf(hull)
	}
	var minWidth float64
	var minWidthPt *xy
	var base [2]xy
	switch len(pts) {
	case 0:
		return Geometry{Type: Polygon}
	case 1:
		minWidthPt = &pts[0]
		base = [2]xy{pts[0], pts[0]}
	case 2, 3:
		minWidthPt = &pts[0]
		base = [2]xy{pts[0], pts[1]}
	default:
		minWidth = math.Inf(1)
		currMaxIndex := 1
		for i := 1; i < len(pts); i++ {
			seg := [2]xy{pts[i-1], pts[i]}
			maxPerpDistance := perpendicularDistance(pts[currMaxIndex], seg)
			nextPerpDistance := maxPerpDistance
			maxIndex, nextIndex := currMaxIndex, currMaxIndex
			for nextPerpDistance >= maxPerpDistance {
				maxPerpDistance = nextPerpDistance
				maxIndex = nextIndex
				nextIndex = maxIndex + 1
				if nextIndex >= len(pts) {
					nextIndex = 0
				}
				if nextIndex == currMaxIndex {
					break
				}
				nextPerpDistance = perpendicularDistance(pts[nextIndex], seg)
			}
			if maxPerpDistance < minWidth {
				minWidth = maxPerpDistance
				minWidthPt = &pts[maxIndex]
				base = seg
			}
			currMaxIndex = maxIndex
		}
	}
	if minWidthPt == nil {
		return Geometry{Type: Polygon}
	}
	if minWidth == 0.0 {
		if base[0].x == base[1].x && base[0].y == base[1].y {
			return geometryFromPoints(Point, base[:1], false)
		}
		return maximumLine(pts)
	}
	dx := base[1].x - base[0].x
	dy := base[1].y - base[0].y
	minPara, maxPara := math.Inf(1), math.Inf(-1)
	minPerp, maxPerp := math.Inf(1), math.Inf(-1)
	computeC := func(a float64, b float64, p xy) float64 {
		return float64(a*p.y) - float64(b*p.x)
	}
	for _, p := range pts {
		paraC := computeC(dx, dy, p)
		maxPara = math.Max(maxPara, paraC)
		minPara = math.Min(minPara, paraC)
		perpC := computeC(-dy, dx, p)
		maxPerp = math.Max(maxPerp, perpC)
		minPerp = math.Min(minPerp, perpC)
	}
	maxPerpLine := segmentForLine(-dx, -dy, maxPerp)
	minPerpLine := segmentForLine(-dx, -dy, minPerp)
	maxParaLine := segmentForLine(-dy, dx, maxPara)
	minParaLine := segmentForLine(-dy, dx, minPara)
	p0 := lineIntersection(maxParaLine, maxPerpLine)
	p1 := lineIntersection(minParaLine, maxPerpLine)
	p2 := lineIntersection(minParaLine, minPerpLine)
	p3 := lineIntersection(maxParaLine, minPerpLine)
	return geometryFromPoints(Polygon, []xy{p0, p1, p2, p3, p0}, false)
}

// verticesOf returns the vertices of a point, linestring, or ring.
func verticesOf(g Geometry) []xy {
	dims := 2
	if g.HasZ {
		dims = 3
	}
	var points []xy
	for i := 0; i+dims <= len(g.Coords); i += dims {
		points = append(points, xy{x: g.Coords[i], y: g.Coords[i+1]})
	}
	return points
}

// perpendicularDistance is GEOS's Distance::pointToLinePerpendicular.
func perpendicularDistance(p xy, seg [2]xy) float64 {
	a, b := seg[0], seg[1]
	lengthSquared := float64((b.x-a.x)*(b.x-a.x)) + float64((b.y-a.y)*(b.y-a.y))
	s := (float64((a.y-p.y)*(b.x-a.x)) - float64((a.x-p.x)*(b.y-a.y))) / lengthSquared
	return math.Abs(s) * math.Sqrt(lengthSquared)
}

// segmentForLine is MinimumDiameter::computeSegmentForLine, which returns a segment on the line ax + by = c.
func segmentForLine(a float64, b float64, c float64) [2]xy {
	if math.Abs(b) > math.Abs(a) {
		return [2]xy{{x: 0.0, y: c / b}, {x: 1.0, y: c/b - a/b}}
	}
	return [2]xy{{x: c / a, y: 0.0}, {x: c/a - b/a, y: 1.0}}
}

// lineIntersection is GEOS's Intersection::intersection of the lines through two segments.
func lineIntersection(p [2]xy, q [2]xy) xy {
	p1, p2, q1, q2 := p[0], p[1], q[0], q[1]
	minX0, maxX0 := math.Min(p1.x, p2.x), math.Max(p1.x, p2.x)
	minY0, maxY0 := math.Min(p1.y, p2.y), math.Max(p1.y, p2.y)
	minX1, maxX1 := math.Min(q1.x, q2.x), math.Max(q1.x, q2.x)
	minY1, maxY1 := math.Min(q1.y, q2.y), math.Max(q1.y, q2.y)
	intMinX, intMaxX := math.Max(minX0, minX1), math.Min(maxX0, maxX1)
	intMinY, intMaxY := math.Max(minY0, minY1), math.Min(maxY0, maxY1)
	midx := (intMinX + intMaxX) / 2.0
	midy := (intMinY + intMaxY) / 2.0
	p1x, p1y := p1.x-midx, p1.y-midy
	p2x, p2y := p2.x-midx, p2.y-midy
	q1x, q1y := q1.x-midx, q1.y-midy
	q2x, q2y := q2.x-midx, q2.y-midy
	px := p1y - p2y
	py := p2x - p1x
	pw := float64(p1x*p2y) - float64(p2x*p1y)
	qx := q1y - q2y
	qy := q2x - q1x
	qw := float64(q1x*q2y) - float64(q2x*q1y)
	x := float64(py*qw) - float64(qy*pw)
	y := float64(qx*pw) - float64(px*qw)
	w := float64(px*qy) - float64(qx*py)
	return xy{x: x/w + midx, y: y/w + midy}
}

// maximumLine is MinimumDiameter::computeMaximumLine, which returns the line between the extreme points of a
// degenerate hull.
func maximumLine(pts []xy) Geometry {
	ptMinX, ptMaxX, ptMinY, ptMaxY := pts[0], pts[0], pts[0], pts[0]
	for _, p := range pts[1:] {
		if p.x < ptMinX.x {
			ptMinX = p
		}
		if p.x > ptMaxX.x {
			ptMaxX = p
		}
		if p.y < ptMinY.y {
			ptMinY = p
		}
		if p.y > ptMaxY.y {
			ptMaxY = p
		}
	}
	p0, p1 := ptMinX, ptMaxX
	if p0.x == p1.x {
		p0, p1 = ptMinY, ptMaxY
	}
	return geometryFromPoints(LineString, []xy{p0, p1}, false)
}
