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
	"sort"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// stShiftLongitude implements ST_ShiftLongitude, which moves negative longitudes up by 360 degrees and longitudes
// past 180 degrees down by 360 degrees.
func stShiftLongitude(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	return g.withShape(g.Shape.mapCoords(func(point []float64) []float64 {
		if point[0] < 0 {
			point[0] += 360
		} else if point[0] > 180 {
			point[0] -= 360
		}
		return point
	})), nil
}

// wrapShape splits a shape at the wrapping line, moving the parts on the side that the move points away from: the
// parts left of the line for a positive move and the parts right of it for a negative move.
func wrapShape(s shape, wrap float64, move float64) (shape, error) {
	b, ok := s.bounds()
	if !ok {
		return s, nil
	}
	moved := func(part shape) shape {
		return part.mapCoords(func(point []float64) []float64 {
			point[0] += move
			return point
		})
	}
	if (move > 0 && b.maxX <= wrap) || (move < 0 && b.minX >= wrap) {
		return moved(s), nil
	}
	if b.maxX <= wrap || b.minX >= wrap || hasMembers(s.typeNum) && s.typeNum != compoundType && s.typeNum != curvePolyType {
		if !hasMembers(s.typeNum) {
			return s, nil
		}
		members := make([]shape, len(s.members))
		for i, member := range s.members {
			wrapped, err := wrapShape(member, wrap, move)
			if err != nil {
				return shape{}, err
			}
			members[i] = wrapped
		}
		s.members = members
		return s, nil
	}
	geosShape, err := s.toGEOS()
	if err != nil {
		return shape{}, err
	}
	g := spatial.NewGEOS()
	defer g.Close()
	var parts []shape
	for _, side := range []shape{rectanglePolygon(b.minX-1, b.minY-1, wrap, b.maxY+1), rectanglePolygon(wrap, b.minY-1, b.maxX+1, b.maxY+1)} {
		part, err := g.Intersection(geosShape, side.linearToGEOS(), -1)
		if err != nil {
			return shape{}, geosError("GEOSIntersection", err)
		}
		parts = append(parts, shapeFromGEOS(part, part.HasZ))
	}
	if move > 0 {
		parts[0] = moved(parts[0])
	} else {
		parts[1] = moved(parts[1])
	}
	wrapped := shape{typeNum: collectionType, ctype: s.ctype}
	for _, part := range parts {
		atomicMembers(part, false, func(member shape) {
			if !member.isEmpty() {
				wrapped.members = append(wrapped.members, member)
			}
		})
	}
	if len(wrapped.members) == 1 {
		return wrapped.members[0], nil
	}
	if typeNum := wrapped.members[0].typeNum; multiTypeNumber(typeNum) != typeNum {
		wrapped.typeNum = multiTypeNumber(typeNum)
		for _, member := range wrapped.members {
			if member.typeNum != typeNum {
				wrapped.typeNum = collectionType
			}
		}
	}
	return wrapped, nil
}

// stWrapX implements ST_WrapX.
func stWrapX(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	wrapped, err := wrapShape(g.Shape, args[1].(float64), args[2].(float64))
	if err != nil {
		return nil, err
	}
	return g.withShape(wrapped), nil
}

// stScroll implements ST_Scroll, which rotates a closed linestring to start at one of its vertices.
func stScroll(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	point, err := pointArg(ctx, args[1].(Geometry), "ST_Scroll: second argument must be a point")
	if err != nil {
		return nil, err
	}
	line := g.Shape
	if line.typeNum != lineType {
		return nil, pgerror.Newf(pgcode.Internal, "ST_Scroll: unsupported geometry type: %s", typeNameForNumber(line.typeNum))
	}
	dims := line.ctype.Dimension()
	if !isClosed(line.coords, line.ctype, true) {
		return nil, pgerror.New(pgcode.Internal, "ptarray_scroll_in_place: input POINTARRAY is not closed")
	}
	open := line.coords[:len(line.coords)-dims]
	for i := 0; i < len(open); i += dims {
		if open[i] == point[0] && open[i+1] == point[1] {
			scrolled := append(append([]float64(nil), open[i:]...), open[:i]...)
			line.coords = append(scrolled, scrolled[:dims]...)
			return g.withShape(line), nil
		}
	}
	return nil, pgerror.New(pgcode.Internal, "ptarray_scroll_in_place: input POINTARRAY does not contain the given point")
}

// filterByM returns a copy of the shape keeping only the vertices whose M coordinate lies within the range, dropping
// the M coordinates unless `keepM` is true. Rings left with fewer than four points empty their polygons.
func filterByM(s shape, minM float64, maxM float64, keepM bool) shape {
	dims := s.ctype.Dimension()
	ctype := s.ctype
	if !keepM {
		ctype = s.ctype &^ dimXYM
	}
	filter := func(seq []float64) []float64 {
		var kept []float64
		for i := 0; i+dims <= len(seq); i += dims {
			if m := seq[i+dims-1]; m >= minM && m <= maxM {
				kept = append(kept, convertPoint(seq[i:i+dims], s.ctype, ctype, 0, 0)...)
			}
		}
		return kept
	}
	result := shape{typeNum: s.typeNum, ctype: ctype}
	switch {
	case hasSequence(s.typeNum):
		result.coords = filter(s.coords)
	case s.typeNum == polygonType:
		for _, ring := range s.rings {
			filtered := filter(ring)
			if len(filtered) < 4*ctype.Dimension() {
				result.rings = nil
				break
			}
			result.rings = append(result.rings, filtered)
		}
	default:
		for _, member := range s.members {
			result.members = append(result.members, filterByM(member, minM, maxM, keepM))
		}
	}
	return result
}

// stFilterByM implements ST_FilterByM, which returns geometries without M coordinates unchanged with a notice.
func stFilterByM(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	g := args[0].(Geometry)
	if !g.Shape.hasM() {
		sendNotice(ctx, "NOTICE", "No M-value, No vertex removed")
		return g, nil
	}
	maxM := math.MaxFloat64
	if args[2] != nil {
		maxM = args[2].(float64)
	}
	keepM := args[3] != nil && args[3].(bool)
	return g.withShape(filterByM(g.Shape, args[1].(float64), maxM, keepM)), nil
}

// stLineExtend implements ST_LineExtend, which extends a linestring past its end and before its start, in the
// directions of its first and last segments that have length.
func stLineExtend(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	forward, backward := args[1].(float64), args[2].(float64)
	line := g.Shape
	if line.typeNum != lineType {
		return nil, pgerror.New(pgcode.Internal, "Argument must be LINESTRING geometry")
	}
	if forward < 0 || backward < 0 {
		return nil, pgerror.New(pgcode.Internal, "lwline_extend: distances must be non-negative")
	}
	dims := line.ctype.Dimension()
	count := len(line.coords) / dims
	extend := func(from int, step int, distance float64) []float64 {
		if distance == 0 {
			return nil
		}
		end := line.coords[from*dims : (from+1)*dims]
		for i := from + step; i >= 0 && i < count; i += step {
			prev := line.coords[i*dims : (i+1)*dims]
			length := math.Hypot(end[0]-prev[0], end[1]-prev[1])
			if length == 0 {
				continue
			}
			extended := make([]float64, dims)
			for j := range extended {
				extended[j] = end[j] + float64((end[j]-prev[j])/length*distance)
			}
			return extended
		}
		return nil
	}
	coords := extend(0, 1, backward)
	coords = append(coords, line.coords...)
	coords = append(coords, extend(count-1, -1, forward)...)
	line.coords = coords
	return g.withShape(line), nil
}

// stBoundingDiagonal implements ST_BoundingDiagonal, the linestring from the minimum to the maximum corner of the
// geometry's bounding box, which is widened to float32 values unless `fits` asks for the exact one.
func stBoundingDiagonal(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	b, ok := g.Shape.bounds()
	if !args[1].(bool) {
		b = b.floatRounded()
	}
	ctype := g.Shape.ctype &^ dimXYM
	if !ok {
		return g.withShape(shape{typeNum: lineType, ctype: ctype}), nil
	}
	coords := []float64{b.minX, b.minY, b.maxX, b.maxY}
	if ctype.Is3D() {
		coords = []float64{b.minX, b.minY, b.minZ, b.maxX, b.maxY, b.maxZ}
	}
	return g.withShape(shape{typeNum: lineType, ctype: ctype, coords: coords}), nil
}

// chaikin returns a sequence smoothed once by cutting each corner at a quarter and three quarters of each segment.
// A ring wraps around, keeping its first point when `keepEnds` is true, and an open sequence always keeps its ends.
func chaikin(seq []float64, dims int, isRing bool, keepEnds bool) []float64 {
	count := len(seq) / dims
	if count < 3 {
		return seq
	}
	point := func(a []float64, b []float64, weight float64) []float64 {
		p := make([]float64, dims)
		for j := range p {
			p[j] = float64((1-weight)*a[j]) + float64(weight*b[j])
		}
		return p
	}
	var smoothed []float64
	keep := !isRing || keepEnds
	if keep {
		smoothed = append(smoothed, seq[:dims]...)
	}
	for i := 0; i+1 < count; i++ {
		a, b := seq[i*dims:(i+1)*dims], seq[(i+1)*dims:(i+2)*dims]
		if !(keep && i == 0) {
			smoothed = append(smoothed, point(a, b, 0.25)...)
		}
		if !(keep && i == count-2) {
			smoothed = append(smoothed, point(a, b, 0.75)...)
		}
	}
	if keep {
		return append(smoothed, seq[len(seq)-dims:]...)
	}
	return append(smoothed, smoothed[:dims]...)
}

// chaikinShape smooths every linestring and ring of the shape.
func chaikinShape(s shape, iterations int, keepEnds bool) shape {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case lineType:
		for i := 0; i < iterations; i++ {
			s.coords = chaikin(s.coords, dims, false, keepEnds)
		}
	case polygonType:
		rings := make([][]float64, len(s.rings))
		for r, ring := range s.rings {
			for i := 0; i < iterations; i++ {
				ring = chaikin(ring, dims, true, keepEnds)
			}
			rings[r] = ring
		}
		s.rings = rings
	case pointType, multiPointType:
	default:
		members := make([]shape, len(s.members))
		for i, member := range s.members {
			members[i] = chaikinShape(member, iterations, keepEnds)
		}
		s.members = members
	}
	return s
}

// stChaikinSmoothing implements ST_ChaikinSmoothing.
func stChaikinSmoothing(ctx *sql.Context, args ...any) (any, error) {
	g, iterations := args[0].(Geometry), int(args[1].(int32))
	if iterations < 1 || iterations > 5 {
		return nil, pgerror.New(pgcode.Internal, "Number of iterations must be between 1 and 5 : LWGEOM_ChaikinSmoothing")
	}
	return g.withShape(chaikinShape(g.Shape, iterations, args[2].(bool))), nil
}

// effectiveAreas returns the Visvalingam-Whyatt effective area of each point of a sequence: the area of the triangle
// that it forms with its neighbors when it becomes the point of least area, removing points one at a time, but never
// less than the area of a point removed before it. The end points never go and have the largest float32 area.
func effectiveAreas(seq []float64, dims int) []float64 {
	count := len(seq) / dims
	areas := make([]float64, count)
	prev, next := make([]int, count), make([]int, count)
	for i := range areas {
		prev[i], next[i] = i-1, i+1
	}
	triangle := func(i int) float64 {
		a, b, c := seq[prev[i]*dims:], seq[i*dims:], seq[next[i]*dims:]
		return math.Abs(float64((b[0]-a[0])*(c[1]-a[1]))-float64((c[0]-a[0])*(b[1]-a[1]))) / 2
	}
	remaining := make([]int, 0, count)
	for i := 1; i+1 < count; i++ {
		areas[i] = triangle(i)
		remaining = append(remaining, i)
	}
	for len(remaining) > 0 {
		sort.SliceStable(remaining, func(x, y int) bool { return areas[remaining[x]] < areas[remaining[y]] })
		i := remaining[0]
		remaining = remaining[1:]
		p, n := prev[i], next[i]
		next[p], prev[n] = n, p
		for _, neighbor := range []int{p, n} {
			if neighbor > 0 && neighbor < count-1 {
				areas[neighbor] = math.Max(triangle(neighbor), areas[i])
			}
		}
	}
	areas[0], areas[count-1] = math.MaxFloat32, math.MaxFloat32
	return areas
}

// simplifyVW returns a copy of the shape without the points whose effective area is less than the threshold,
// keeping at least two points in a linestring and four in a ring. With `setArea` true, every point keeps its effective
// area as its M coordinate instead.
func simplifyVW(s shape, threshold float64, setArea bool) shape {
	dims := s.ctype.Dimension()
	ctype := s.ctype
	if setArea {
		ctype |= dimXYM
	}
	simplifySeq := func(seq []float64, minPoints int) []float64 {
		count := len(seq) / dims
		if count < 3 {
			return seq
		}
		areas := effectiveAreas(seq, dims)
		order := make([]int, 0, count)
		for i := 1; i+1 < count; i++ {
			order = append(order, i)
		}
		sort.SliceStable(order, func(x, y int) bool { return areas[order[x]] < areas[order[y]] })
		dropped := make([]bool, count)
		left := count
		for _, i := range order {
			if areas[i] >= threshold || left <= minPoints {
				break
			}
			dropped[i] = true
			left--
		}
		var kept []float64
		for i := 0; i < count; i++ {
			if !dropped[i] {
				point := convertPoint(seq[i*dims:(i+1)*dims], s.ctype, ctype, 0, 0)
				if setArea {
					point[len(point)-1] = areas[i]
				}
				kept = append(kept, point...)
			}
		}
		return kept
	}
	result := shape{typeNum: s.typeNum, ctype: ctype}
	switch s.typeNum {
	case lineType:
		result.coords = simplifySeq(s.coords, 2)
	case polygonType:
		for _, ring := range s.rings {
			result.rings = append(result.rings, simplifySeq(ring, 4))
		}
	case pointType:
		result.coords = convertPoint(s.coords, s.ctype, ctype, 0, 0)
		if len(s.coords) == 0 {
			result.coords = nil
		}
	default:
		for _, member := range s.members {
			result.members = append(result.members, simplifyVW(member, threshold, setArea))
		}
	}
	return result
}

// quantize returns a value with the bits of its mantissa that lie below the precision of the given number of decimal
// digits cleared.
func quantize(val float64, digits int) float64 {
	if val == 0 || math.IsNaN(val) || math.IsInf(val, 0) {
		return val
	}
	_, exponent := math.Frexp(val)
	keep := exponent - 1 + int(math.Ceil(float64(digits)*math.Log2(10))) + 1
	if keep >= 52 {
		return val
	}
	keep = max(keep, 0)
	bits := math.Float64bits(val)
	return math.Float64frombits(bits &^ ((uint64(1) << uint(52-keep)) - 1))
}

// stQuantizeCoordinates implements ST_QuantizeCoordinates, whose Y, Z, and M precisions default to the X precision.
func stQuantizeCoordinates(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	g := args[0].(Geometry)
	digits := make([]int, 4)
	for i := range digits {
		digits[i] = int(args[1].(int32))
		if args[1+i] != nil {
			digits[i] = int(args[1+i].(int32))
		}
	}
	ctype := g.Shape.ctype
	return g.withShape(g.Shape.mapCoords(func(point []float64) []float64 {
		point[0], point[1] = quantize(point[0], digits[0]), quantize(point[1], digits[1])
		if ctype.Is3D() {
			point[2] = quantize(point[2], digits[2])
		}
		if ctype.IsMeasured() {
			point[len(point)-1] = quantize(point[len(point)-1], digits[3])
		}
		return point
	})), nil
}

// removeSmallParts returns a copy of the shape without the linestrings, polygons, and interior rings whose bounding
// box is narrower or shorter than the minimums. A polygon whose exterior ring is too small is removed entirely.
func removeSmallParts(s shape, minWidth float64, minHeight float64) shape {
	dims := s.ctype.Dimension()
	isSmall := func(seq []float64) bool {
		b, ok := shape{typeNum: lineType, ctype: s.ctype, coords: seq}.bounds()
		return !ok || b.maxX-b.minX < minWidth || b.maxY-b.minY < minHeight
	}
	switch s.typeNum {
	case lineType:
		if isSmall(s.coords) {
			s.coords = nil
		}
	case polygonType:
		if len(s.rings) == 0 || isSmall(s.rings[0]) {
			s.rings = nil
			break
		}
		rings := [][]float64{s.rings[0]}
		for _, hole := range s.rings[1:] {
			if !isSmall(hole) {
				rings = append(rings, hole)
			}
		}
		s.rings = rings
	case pointType, multiPointType:
	default:
		var members []shape
		for _, member := range s.members {
			if kept := removeSmallParts(member, minWidth, minHeight); !kept.isEmpty() {
				members = append(members, kept)
			}
		}
		s.members = members
	}
	_ = dims
	return s
}

// reshapeRoutines returns the routines that wrap, smooth, filter, and quantize geometries.
func reshapeRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_boundingdiagonal", Symbol: "ST_BoundingDiagonal", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "fits", Type: "bool", Default: `false`}}, Returns: "geometry", Strict: true, Impl: stBoundingDiagonal},
		{Name: "st_chaikinsmoothing", Symbol: "LWGEOM_ChaikinSmoothing", Parameters: []extdef.Parameter{{Type: "geometry"}, {Type: "int4", Default: `1`}, {Type: "bool", Default: `false`}}, Returns: "geometry", Strict: true, Impl: stChaikinSmoothing},
		{Name: "st_filterbym", Symbol: "LWGEOM_FilterByM", Parameters: []extdef.Parameter{{Type: "geometry"}, {Type: "float8"}, {Type: "float8", Default: `NULL::double precision`}, {Type: "bool", Default: `false`}}, Returns: "geometry", Impl: stFilterByM},
		{Name: "st_lineextend", Symbol: "geometry_line_extend", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "distance_forward", Type: "float8"}, {Name: "distance_backward", Type: "float8", Default: `0.0`}}, Returns: "geometry", Strict: true, Impl: stLineExtend},
		{Name: "st_quantizecoordinates", Symbol: "ST_QuantizeCoordinates", Parameters: []extdef.Parameter{{Name: "g", Type: "geometry"}, {Name: "prec_x", Type: "int4"}, {Name: "prec_y", Type: "int4", Default: `NULL::integer`}, {Name: "prec_z", Type: "int4", Default: `NULL::integer`}, {Name: "prec_m", Type: "int4", Default: `NULL::integer`}}, Returns: "geometry", Impl: stQuantizeCoordinates},
		{Name: "st_removesmallparts", Symbol: "ST_RemoveSmallParts", Parameters: params("geometry", "float8", "float8"), Returns: "geometry", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			g := args[0].(Geometry)
			return g.withShape(removeSmallParts(g.Shape, args[1].(float64), args[2].(float64))), nil
		}},
		{Name: "st_scroll", Symbol: "ST_Scroll", Parameters: params("geometry", "geometry"), Returns: "geometry", Strict: true, Impl: stScroll},
		{Name: "st_seteffectivearea", Symbol: "st_seteffectivearea_geometry_float8_int4", Parameters: []extdef.Parameter{{Type: "geometry"}, {Type: "float8", Default: `'-1'::integer`}, {Type: "int4", Default: `1`}}, Returns: "geometry", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			g := args[0].(Geometry)
			return g.withShape(simplifyVW(g.Shape, args[1].(float64), true)), nil
		}},
		{Name: "st_shiftlongitude", Symbol: "LWGEOM_longitude_shift", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stShiftLongitude},
		{Name: "st_simplifyvw", Symbol: "st_simplifyvw_geometry_float8", Parameters: params("geometry", "float8"), Returns: "geometry", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			g := args[0].(Geometry)
			return g.withShape(simplifyVW(g.Shape, args[1].(float64), false)), nil
		}},
		{Name: "st_wrapx", Symbol: "ST_WrapX", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "wrap", Type: "float8"}, {Name: "move", Type: "float8"}}, Returns: "geometry", Strict: true, Impl: stWrapX},
	}
}
