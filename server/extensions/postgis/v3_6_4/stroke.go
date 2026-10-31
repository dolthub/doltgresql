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
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// strokeTolerance controls how arcs are approximated by straight segments.
type strokeTolerance struct {
	// kind is 0 for segments per quarter circle, 1 for the largest distance between the arc and a segment, and 2 for
	// the largest angle that a segment may span.
	kind int
	// value is the number of segments, the distance, or the angle in radians.
	value float64
	// symmetric makes the result independent of the arc's direction by always sweeping counterclockwise and spacing
	// the segments evenly.
	symmetric bool
	// retainAngle keeps the step angle of a symmetric stroke, splitting the leftover angle between the first and last
	// segments.
	retainAngle bool
}

// defaultStroke is the tolerance PostGIS uses whenever it linearizes curves for its own computations: 32 segments per
// quarter circle.
var defaultStroke = strokeTolerance{value: 32}

// strokeShape returns the shape with every circular string replaced by a linestring, compound curves joined into
// linestrings, curve polygons turned into polygons, and multicurves and multisurfaces turned into their linear
// equivalents.
func strokeShape(s shape, tol strokeTolerance) (shape, error) {
	switch s.typeNum {
	case circStringType:
		coords, err := strokeArcs(s.coords, s.ctype, tol)
		return shape{typeNum: lineType, ctype: s.ctype, coords: coords}, err
	case compoundType:
		var coords []float64
		for _, member := range s.members {
			stroked, err := strokeShape(member, tol)
			if err != nil {
				return shape{}, err
			}
			coords = appendDistinct(coords, stroked.coords, s.ctype)
		}
		return shape{typeNum: lineType, ctype: s.ctype, coords: coords}, nil
	case curvePolyType:
		polygon := shape{typeNum: polygonType, ctype: s.ctype}
		for _, member := range s.members {
			stroked, err := strokeShape(member, tol)
			if err != nil {
				return shape{}, err
			}
			polygon.rings = append(polygon.rings, stroked.coords)
		}
		return polygon, nil
	case multiCurveType, multiSurfaceType, collectionType:
		result := shape{typeNum: s.typeNum, ctype: s.ctype, members: make([]shape, len(s.members))}
		if s.typeNum == multiCurveType {
			result.typeNum = multiLineType
		} else if s.typeNum == multiSurfaceType {
			result.typeNum = multiPolygonType
		}
		for i, member := range s.members {
			stroked, err := strokeShape(member, tol)
			if err != nil {
				return shape{}, err
			}
			result.members[i] = stroked
		}
		return result, nil
	default:
		return s, nil
	}
}

// appendDistinct appends each point of `src` to `dst`, skipping a point that repeats the one before it.
func appendDistinct(dst []float64, src []float64, ctype coordinatesType) []float64 {
	dims := ctype.Dimension()
	for i := 0; i+dims <= len(src); i += dims {
		if len(dst) >= dims && equalOrdinates(dst[len(dst)-dims:], src[i:i+dims]) {
			continue
		}
		dst = append(dst, src[i:i+dims]...)
	}
	return dst
}

// equalOrdinates returns whether two points have identical ordinates.
func equalOrdinates(a []float64, b []float64) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// strokeArcs linearizes the arcs of a circular string, where each arc runs through three consecutive points and
// shares its last point with the next arc.
func strokeArcs(coords []float64, ctype coordinatesType, tol strokeTolerance) ([]float64, error) {
	dims := ctype.Dimension()
	count := len(coords) / dims
	if count == 0 {
		return nil, nil
	}
	var out []float64
	for i := 2; i < count; i += 2 {
		start, mid, end := coords[(i-2)*dims:(i-1)*dims], coords[(i-1)*dims:i*dims], coords[i*dims:(i+1)*dims]
		arc, err := strokeArc(start, mid, end, ctype, tol)
		if err != nil {
			return nil, err
		}
		if arc == nil {
			out = append(out, start...)
			out = append(out, mid...)
			continue
		}
		out = appendDistinct(out, arc, ctype)
	}
	return appendDistinct(out, coords[(count-1)*dims:], ctype), nil
}

// strokeArc returns the start of the arc through the three points followed by the points that approximate it,
// leaving out the end. It returns nil when the points lie on a straight line, unless the arc is a full circle, which
// starts and ends at the same point.
func strokeArc(start []float64, mid []float64, end []float64, ctype coordinatesType, tol strokeTolerance) ([]float64, error) {
	side := pointSide(start, end, mid)
	reversed := false
	if side < 0 && tol.symmetric {
		start, end, side, reversed = end, start, 1, true
	}
	centerX, centerY, radius := circumcircle(start, mid, end)
	isCircle := start[0] == end[0] && start[1] == end[1]
	if (radius < 0 || side == 0) && !isCircle {
		return nil, nil
	}
	clockwise := side < 0
	increment, err := strokeIncrement(tol, radius)
	if err != nil {
		return nil, err
	}
	angleStart := math.Atan2(start[1]-centerY, start[0]-centerX)
	angleMid := math.Atan2(mid[1]-centerY, mid[0]-centerX)
	angleEnd := math.Atan2(end[1]-centerY, end[0]-centerX)
	total := angleEnd - angleStart
	if clockwise {
		total = angleStart - angleEnd
	}
	if total <= 0 {
		total += 2 * math.Pi
	}
	minSegments := 2
	if isCircle {
		minSegments = 3
	}
	segments := int(math.Ceil(total / increment))
	if segments < minSegments {
		segments, increment = minSegments, total/float64(minSegments)
	}
	shift := 0.0
	if tol.symmetric {
		if tol.retainAngle {
			segments = int(math.Trunc(total / increment))
			shift = (total - increment*float64(segments)) / 2
		} else {
			segments = int(math.Ceil(total / increment))
			increment = total / float64(segments)
		}
	}
	if clockwise {
		increment, shift = -increment, -shift
		if angleEnd > angleStart {
			angleEnd -= 2 * math.Pi
		}
		if angleMid > angleStart {
			angleMid -= 2 * math.Pi
		}
	} else {
		if angleEnd < angleStart {
			angleEnd += 2 * math.Pi
		}
		if angleMid < angleStart {
			angleMid += 2 * math.Pi
		}
	}
	if isCircle {
		increment = math.Abs(increment)
		segments = int(math.Ceil(total / increment))
		if segments < 3 {
			segments, increment = 3, total/3
		}
		angleEnd, angleMid, shift = angleStart+2*math.Pi, angleStart+math.Pi, 0
	}
	first, last := 1, segments
	if shift != 0 {
		first, last = 0, segments+1
	}
	dims := ctype.Dimension()
	points := append([]float64(nil), start...)
	for step := first; step < last; step++ {
		angle := angleStart + float64(increment*float64(step)) + shift
		point := []float64{centerX + float64(radius*math.Cos(angle)), centerY + float64(radius*math.Sin(angle))}
		for extra := 2; extra < dims; extra++ {
			point = append(point, interpolateByAngle(angle, angleStart, angleMid, angleEnd, start[extra], mid[extra], end[extra]))
		}
		points = appendDistinct(points, point, ctype)
	}
	if isCircle {
		points = append(points[:len(points)-dims], start...)
	}
	if !reversed {
		return points, nil
	}
	reversedPoints := append([]float64(nil), end...)
	for i := len(points) - dims; i >= dims; i -= dims {
		reversedPoints = appendDistinct(reversedPoints, points[i:i+dims], ctype)
	}
	return reversedPoints, nil
}

// strokeIncrement returns the angle that each segment spans for the given tolerance and arc radius.
func strokeIncrement(tol strokeTolerance, radius float64) (float64, error) {
	switch tol.kind {
	case 0:
		perQuad := math.RoundToEven(tol.value)
		if perQuad != tol.value {
			return 0, pgerror.Newf(pgcode.Internal, "lwarc_linearize: segments per quadrant must be an integer value, got %s", formatG(tol.value, 15))
		}
		if perQuad < 1 {
			return 0, pgerror.Newf(pgcode.Internal, "lwarc_linearize: segments per quadrant must be at least 1, got %d", int(perQuad))
		}
		return math.Pi / 2 / perQuad, nil
	case 1:
		if tol.value <= 0 {
			return 0, pgerror.Newf(pgcode.Internal, "lwarc_linearize: max deviation must be bigger than 0, got %s", formatG(tol.value, 15))
		}
		maxDeviation := math.Min(tol.value, 2*radius)
		halfAngle := math.Acos(1 - maxDeviation/radius)
		for halfAngle == 0 {
			maxDeviation *= 2
			halfAngle = math.Acos(1 - maxDeviation/radius)
		}
		return 2 * halfAngle, nil
	case 2:
		if tol.value <= 0 {
			return 0, pgerror.Newf(pgcode.Internal, "lwarc_linearize: max angle must be bigger than 0, got %s", formatG(tol.value, 15))
		}
		return tol.value, nil
	default:
		return 0, pgerror.Newf(pgcode.Internal, "lwarc_linearize: unsupported tolerance type %d", tol.kind)
	}
}

// interpolateByAngle returns the value at the given angle along an arc, interpolating linearly between the values at
// the start and middle points, or between the middle and end points.
func interpolateByAngle(angle float64, angleStart float64, angleMid float64, angleEnd float64, valStart float64, valMid float64, valEnd float64) float64 {
	increasing := angleStart < angleEnd
	if (increasing && angle <= angleMid) || (!increasing && angle >= angleMid) {
		return valStart + (valMid-valStart)*(angle-angleStart)/(angleMid-angleStart)
	}
	return valMid + (valEnd-valMid)*(angle-angleMid)/(angleEnd-angleMid)
}

// pointSide returns which side of the line from `a` to `b` the point `p` lies on: negative for the left, positive for
// the right, and zero when it lies on the line.
func pointSide(a []float64, b []float64, p []float64) int {
	cross := float64((p[0]-a[0])*(b[1]-a[1])) - float64((b[0]-a[0])*(p[1]-a[1]))
	if cross < 0 {
		return -1
	}
	if cross > 0 {
		return 1
	}
	return 0
}

// circumcircle returns the center and radius of the circle through three points, or a negative radius when the
// points lie on a straight line. When the first and last points coincide, the circle has the other point opposite
// them.
func circumcircle(p1 []float64, p2 []float64, p3 []float64) (float64, float64, float64) {
	if p1[0] == p3[0] && p1[1] == p3[1] {
		centerX, centerY := (p1[0]+p2[0])/2, (p1[1]+p2[1])/2
		return centerX, centerY, math.Hypot(centerX-p1[0], centerY-p1[1])
	}
	dx21, dy21 := p2[0]-p1[0], p2[1]-p1[1]
	dx31, dy31 := p3[0]-p1[0], p3[1]-p1[1]
	h21, h31 := float64(dx21*dx21)+float64(dy21*dy21), float64(dx31*dx31)+float64(dy31*dy31)
	d := 2 * (float64(dx21*dy31) - float64(dx31*dy21))
	if math.Abs(d) < 1e-8 {
		return 0, 0, -1
	}
	centerX := p1[0] + (float64(h21*dy31)-float64(h31*dy21))/d
	centerY := p1[1] - (float64(h21*dx31)-float64(h31*dx21))/d
	dx, dy := centerX-p1[0], centerY-p1[1]
	return centerX, centerY, math.Sqrt(float64(dx*dx) + float64(dy*dy))
}

// formatG formats a number the way C's "%g" format does with the given precision.
func formatG(val float64, precision int) string {
	return strconv.FormatFloat(val, 'g', precision, 64)
}

// stCurveToLine implements ST_CurveToLine, which linearizes every curve in the geometry.
func stCurveToLine(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	tol := strokeTolerance{value: args[1].(float64), kind: int(args[2].(int32))}
	flags := args[3].(int32)
	tol.symmetric, tol.retainAngle = flags&1 != 0, flags&2 != 0
	stroked, err := strokeShape(g.Shape, tol)
	if err != nil {
		return nil, err
	}
	return g.withShape(stroked), nil
}

// arcRunEnd returns the index of the last point of the arc that starts at the point at index `start` of a sequence, in
// which every point lies on the circle through the first three and each step turns by the same angle, of at most 45
// degrees. It returns `start` when fewer than four points form such an arc.
func arcRunEnd(seq []float64, dims int, start int) int {
	const tolerance = 1e-8
	count := len(seq) / dims
	if start+3 >= count {
		return start
	}
	point := func(i int) []float64 {
		return seq[i*dims : i*dims+2]
	}
	centerX, centerY, radius := circumcircle(point(start), point(start+1), point(start+2))
	if radius < 0 {
		return start
	}
	step := func(i int) float64 {
		a, b := point(i), point(i+1)
		turn := math.Atan2(b[1]-centerY, b[0]-centerX) - math.Atan2(a[1]-centerY, a[0]-centerX)
		return math.Remainder(turn, 2*math.Pi)
	}
	first := step(start)
	if math.Abs(first) > math.Pi/4+tolerance {
		return start
	}
	end := start + 1
	for end+1 < count {
		next := point(end + 1)
		if math.Abs(math.Hypot(next[0]-centerX, next[1]-centerY)-radius) > tolerance*radius || math.Abs(step(end)-first) > tolerance {
			break
		}
		end++
	}
	if end-start < 3 {
		return start
	}
	return end
}

// curveSequence returns the curve that a sequence of points describes, replacing each run of at least four points
// that lie evenly spaced on a circle with a circular string through its first, middle, and last points. The
// boolean is false when the sequence holds no such run.
func curveSequence(seq []float64, ctype coordinatesType) (shape, bool) {
	dims := ctype.Dimension()
	count := len(seq) / dims
	point := func(i int) []float64 {
		return seq[i*dims : (i+1)*dims]
	}
	var parts []shape
	var linear []float64
	flush := func() {
		if len(linear) > 0 {
			parts = append(parts, shape{typeNum: lineType, ctype: ctype, coords: linear})
			linear = nil
		}
	}
	hasArc := false
	for i := 0; i+1 < count; {
		if end := arcRunEnd(seq, dims, i); end > i {
			flush()
			var coords []float64
			coords = append(append(append(coords, point(i)...), point(i+(end-i)/2)...), point(end)...)
			parts = append(parts, shape{typeNum: circStringType, ctype: ctype, coords: coords})
			hasArc, i = true, end
			continue
		}
		if len(linear) == 0 {
			linear = append(linear, point(i)...)
		}
		linear = append(linear, point(i+1)...)
		i++
	}
	flush()
	if !hasArc {
		return shape{typeNum: lineType, ctype: ctype, coords: seq}, false
	}
	if len(parts) == 1 {
		return parts[0], true
	}
	return shape{typeNum: compoundType, ctype: ctype, members: parts}, true
}

// curveShape returns the shape with the evenly spaced runs of points on circles in its linestrings and polygons
// replaced by circular arcs, turning each part that gains an arc into the matching curve type. Every other type is
// kept as it is.
func curveShape(s shape) shape {
	switch s.typeNum {
	case lineType:
		curved, _ := curveSequence(s.coords, s.ctype)
		return curved
	case polygonType:
		rings := make([]shape, len(s.rings))
		hasArc := false
		for i, ring := range s.rings {
			var curved bool
			rings[i], curved = curveSequence(ring, s.ctype)
			hasArc = hasArc || curved
		}
		if hasArc {
			return shape{typeNum: curvePolyType, ctype: s.ctype, members: rings}
		}
	case multiLineType, multiPolygonType:
		members := make([]shape, len(s.members))
		changed := false
		for i, member := range s.members {
			members[i] = curveShape(member)
			changed = changed || members[i].typeNum != member.typeNum
		}
		if changed {
			curvedType := map[int]int{multiLineType: multiCurveType, multiPolygonType: multiSurfaceType}[s.typeNum]
			return shape{typeNum: curvedType, ctype: s.ctype, members: members}
		}
	}
	return s
}

// stLineToCurve implements ST_LineToCurve, which turns evenly spaced runs of points on circles back into arcs.
func stLineToCurve(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	return g.withShape(curveShape(g.Shape)), nil
}

// strokeRoutines returns the routines that linearize curves.
func strokeRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_curvetoline", Symbol: "ST_CurveToLine", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"},
			{Name: "tol", Type: "float8", Default: `32`}, {Name: "toltype", Type: "int4", Default: `0`}, {Name: "flags", Type: "int4", Default: `0`}},
			Returns: "geometry", Strict: true, Impl: stCurveToLine},
		{Name: "st_linetocurve", Symbol: "LWGEOM_line_desegmentize", Parameters: []extdef.Parameter{{Name: "geometry", Type: "geometry"}}, Returns: "geometry", Strict: true, Impl: stLineToCurve},
	}
}
