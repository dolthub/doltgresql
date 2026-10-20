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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// isMultiType returns whether shapes of the type number are collections in the sense of ST_IsCollection: the
// multi-geometries, multicurves, multisurfaces, and geometry collections.
func isMultiType(typeNum int) bool {
	switch typeNum {
	case multiPointType, multiLineType, multiPolygonType, collectionType, multiCurveType, multiSurfaceType:
		return true
	default:
		return false
	}
}

// pointShape returns a point holding the ordinates of the point at the given index of a sequence.
func pointShape(seq []float64, ctype coordinatesType, index int) shape {
	dims := ctype.Dimension()
	return shape{typeNum: pointType, ctype: ctype, coords: append([]float64(nil), seq[index*dims:(index+1)*dims]...)}
}

// withShape returns a geometry with the given shape and the original geometry's SRID.
func (g Geometry) withShape(s shape) Geometry {
	return Geometry{SRID: g.SRID, Shape: s}
}

// dimension returns the topological dimension of the shape: zero for points, one for curves, two for surfaces, and
// the largest member dimension for a collection. A closed polyhedral surface is a solid, with dimension three.
func (s shape) dimension() int {
	switch s.typeNum {
	case pointType, multiPointType:
		return 0
	case lineType, circStringType, compoundType, multiLineType, multiCurveType:
		return 1
	case polygonType, curvePolyType, multiPolygonType, multiSurfaceType, triangleType, tinType:
		return 2
	case polyhedralSurfaceType:
		if s.isClosedSurface() {
			return 3
		}
		return 2
	default:
		dim := 0
		for _, member := range s.members {
			dim = max(dim, member.dimension())
		}
		return dim
	}
}

// isClosedSurface returns whether every edge of a polyhedral surface or TIN is shared by exactly two of its patches.
func (s shape) isClosedSurface() bool {
	type edge struct{ a, b string }
	counts := map[edge]int{}
	dims := s.ctype.Dimension()
	for _, ring := range s.sequences() {
		for i := 0; i+2*dims <= len(ring); i += dims {
			a := fmt.Sprint(ring[i : i+min(dims, 3)])
			b := fmt.Sprint(ring[i+dims : i+dims+min(dims, 3)])
			if b < a {
				a, b = b, a
			}
			counts[edge{a, b}]++
		}
	}
	if len(counts) == 0 {
		return false
	}
	for _, count := range counts {
		if count != 2 {
			return false
		}
	}
	return true
}

// isClosedShape returns whether the shape is closed as ST_IsClosed defines it: curves end where they start, polygon
// rings are closed, polyhedral surfaces and TINs enclose a volume, collections hold only closed members, and points
// are always closed.
func (s shape) isClosedShape() bool {
	switch s.typeNum {
	case lineType, circStringType, compoundType:
		return isClosedCurve(s)
	case polygonType:
		for _, ring := range s.rings {
			if !isClosed(ring, s.ctype, true) {
				return false
			}
		}
		return true
	case polyhedralSurfaceType, tinType:
		return s.isClosedSurface()
	}
	for _, member := range s.members {
		if !member.isClosedShape() {
			return false
		}
	}
	return true
}

// numGeometries implements ST_NumGeometries.
func numGeometries(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.isEmpty() {
		return int32(0), nil
	}
	if isMultiType(s.typeNum) {
		return int32(len(s.members)), nil
	}
	return int32(1), nil
}

// geometryN implements ST_GeometryN, which returns the 1-based member of a collection, or the geometry itself when it
// is not a collection and the index is one.
func geometryN(ctx *sql.Context, args ...any) (any, error) {
	g, n := args[0].(Geometry), int(args[1].(int32))
	if g.Shape.isEmpty() {
		return nil, nil
	}
	if !isMultiType(g.Shape.typeNum) {
		if n == 1 {
			return g, nil
		}
		return nil, nil
	}
	if n < 1 || n > len(g.Shape.members) {
		return nil, nil
	}
	return g.withShape(g.Shape.members[n-1]), nil
}

// curveVertices returns the ordinates of every vertex of a linestring, circular string, or compound curve in order.
func curveVertices(s shape) []float64 {
	var vertices []float64
	for _, seq := range s.sequences() {
		vertices = append(vertices, seq...)
	}
	return vertices
}

// pointN implements ST_PointN, which returns the 1-based vertex of a curve, counting backwards from the end when the
// index is negative.
func pointN(ctx *sql.Context, args ...any) (any, error) {
	g, n := args[0].(Geometry), int(args[1].(int32))
	if g.Shape.typeNum != lineType && g.Shape.typeNum != circStringType && g.Shape.typeNum != compoundType {
		return nil, nil
	}
	vertices := curveVertices(g.Shape)
	count := len(vertices) / g.Shape.ctype.Dimension()
	if n < 0 {
		n += count + 1
	}
	if n < 1 || n > count {
		return nil, nil
	}
	return g.withShape(pointShape(vertices, g.Shape.ctype, n-1)), nil
}

// startPoint returns the first vertex of a shape, or false when the shape is empty. Collections of curves are not
// supported, as in PostGIS.
func startPoint(s shape) (shape, bool, error) {
	switch s.typeNum {
	case pointType:
		return s, len(s.coords) > 0, nil
	case lineType, circStringType, triangleType:
		if len(s.coords) == 0 {
			return shape{}, false, nil
		}
		return pointShape(s.coords, s.ctype, 0), true, nil
	case polygonType:
		if len(s.rings) == 0 || len(s.rings[0]) == 0 {
			return shape{}, false, nil
		}
		return pointShape(s.rings[0], s.ctype, 0), true, nil
	case multiCurveType, multiSurfaceType:
		return shape{}, false, pgerror.Newf(pgcode.Internal, "lwgeom_startpoint: unsupported geometry type: %s", typeNameForNumber(s.typeNum))
	default:
		for _, member := range s.members {
			if point, ok, err := startPoint(member); ok || err != nil {
				return point, ok, err
			}
		}
		return shape{}, false, nil
	}
}

// stStartPoint implements ST_StartPoint.
func stStartPoint(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	point, ok, err := startPoint(g.Shape)
	if err != nil || !ok {
		return nil, err
	}
	return g.withShape(point), nil
}

// stEndPoint implements ST_EndPoint, which returns the last vertex of a curve.
func stEndPoint(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if g.Shape.typeNum != lineType && g.Shape.typeNum != circStringType && g.Shape.typeNum != compoundType {
		return nil, nil
	}
	vertices := curveVertices(g.Shape)
	count := len(vertices) / g.Shape.ctype.Dimension()
	if count == 0 {
		return nil, nil
	}
	return g.withShape(pointShape(vertices, g.Shape.ctype, count-1)), nil
}

// ring returns the ring at the given index of a polygon, curve polygon, or triangle as a curve, or false when there
// is no such ring.
func ring(s shape, index int) (shape, bool) {
	switch s.typeNum {
	case polygonType:
		if index < len(s.rings) {
			return shape{typeNum: lineType, ctype: s.ctype, coords: s.rings[index]}, true
		}
		if index == 0 {
			return shape{typeNum: lineType, ctype: s.ctype}, true
		}
	case curvePolyType:
		if index < len(s.members) {
			return s.members[index], true
		}
	case triangleType:
		if index == 0 {
			return shape{typeNum: lineType, ctype: s.ctype, coords: s.coords}, true
		}
	}
	return shape{}, false
}

// stExteriorRing implements ST_ExteriorRing.
func stExteriorRing(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	exterior, ok := ring(g.Shape, 0)
	if !ok {
		return nil, nil
	}
	return g.withShape(exterior), nil
}

// stInteriorRingN implements ST_InteriorRingN, which returns the 1-based interior ring of a polygon or curve polygon.
func stInteriorRingN(ctx *sql.Context, args ...any) (any, error) {
	g, n := args[0].(Geometry), int(args[1].(int32))
	if g.Shape.typeNum == triangleType || n < 1 {
		return nil, nil
	}
	interior, ok := ring(g.Shape, n)
	if !ok || g.Shape.isEmpty() {
		return nil, nil
	}
	return g.withShape(interior), nil
}

// numRings returns the number of rings in a polygon, curve polygon, or triangle.
func numRings(s shape) int {
	switch s.typeNum {
	case polygonType:
		return len(s.rings)
	case curvePolyType:
		return len(s.members)
	case triangleType:
		if len(s.coords) > 0 {
			return 1
		}
	}
	return 0
}

// stNumInteriorRings implements ST_NumInteriorRings, which returns null for anything but a polygon, curve polygon, or
// triangle.
func stNumInteriorRings(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.typeNum != polygonType && s.typeNum != curvePolyType && s.typeNum != triangleType {
		return nil, nil
	}
	return int32(max(numRings(s)-1, 0)), nil
}

// stNRings implements ST_NRings, which counts the rings of every polygon, curve polygon, triangle, and patch.
func stNRings(ctx *sql.Context, args ...any) (any, error) {
	var count func(s shape) int
	count = func(s shape) int {
		if s.typeNum == polygonType || s.typeNum == curvePolyType || s.typeNum == triangleType {
			return numRings(s)
		}
		total := 0
		for _, member := range s.members {
			total += count(member)
		}
		return total
	}
	return int32(count(args[0].(Geometry).Shape)), nil
}

// stNPoints implements ST_NPoints.
func stNPoints(ctx *sql.Context, args ...any) (any, error) {
	return int32(args[0].(Geometry).Shape.numPoints()), nil
}

// stNumPoints implements ST_NumPoints, which returns null for anything but a curve.
func stNumPoints(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.typeNum != lineType && s.typeNum != circStringType && s.typeNum != compoundType {
		return nil, nil
	}
	return int32(s.numPoints()), nil
}

// errPatchesOnly is returned when a patch function is given anything but a polyhedral surface or TIN.
func errPatchesOnly(funcName string) error {
	return pgerror.Newf(pgcode.Internal, "%s only supports TIN and PolyhedralSurface geometries", funcName)
}

// stNumPatches implements ST_NumPatches.
func stNumPatches(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.typeNum != polyhedralSurfaceType && s.typeNum != tinType {
		return nil, errPatchesOnly("ST_NumPatches")
	}
	return int32(len(s.members)), nil
}

// stPatchN implements ST_PatchN, which returns the 1-based patch of a polyhedral surface or TIN.
func stPatchN(ctx *sql.Context, args ...any) (any, error) {
	g, n := args[0].(Geometry), int(args[1].(int32))
	if g.Shape.typeNum != polyhedralSurfaceType && g.Shape.typeNum != tinType {
		return nil, errPatchesOnly("ST_PatchN")
	}
	if n < 1 || n > len(g.Shape.members) {
		return nil, nil
	}
	return g.withShape(g.Shape.members[n-1]), nil
}

// stNumCurves implements ST_NumCurves, which returns the number of curves in a compound curve.
func stNumCurves(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.typeNum != compoundType {
		return nil, nil
	}
	return int32(len(s.members)), nil
}

// stCurveN implements ST_CurveN, which returns the 1-based curve of a compound curve.
func stCurveN(ctx *sql.Context, args ...any) (any, error) {
	g, n := args[0].(Geometry), int(args[1].(int32))
	if g.Shape.typeNum != compoundType || n < 1 || n > len(g.Shape.members) {
		return nil, nil
	}
	return g.withShape(g.Shape.members[n-1]), nil
}

// hasArc returns whether the shape contains a circular string.
func hasArc(s shape) bool {
	if s.typeNum == circStringType {
		return true
	}
	for _, member := range s.members {
		if hasArc(member) {
			return true
		}
	}
	return false
}

// summaryFlags returns the flags that ST_Summary shows in brackets: Z and M for the dimensions, B when PostGIS stores
// a bounding box, G for a geography, and S when the SRID is known.
func summaryFlags(s shape, srid int32, hasBox bool, geodetic bool) string {
	flags := ""
	if s.hasZ() {
		flags += "Z"
	}
	if s.hasM() {
		flags += "M"
	}
	if hasBox {
		flags += "B"
	}
	if geodetic {
		flags += "G"
	}
	if srid != sridUnknown {
		flags += "S"
	}
	return flags
}

// summary writes the ST_Summary description of a shape, indented by `offset` spaces. Only the outermost shape can
// have a bounding box or an SRID.
func summary(sb *strings.Builder, s shape, offset int, srid int32, hasBox bool, geodetic bool) {
	sb.WriteString(strings.Repeat(" ", offset))
	flags := summaryFlags(s, srid, hasBox, geodetic)
	name := typeNameForNumber(s.typeNum)
	plural := func(count int, noun string) string {
		if count == 0 {
			return fmt.Sprintf(" with 0 %ss", noun)
		}
		if count == 1 {
			return fmt.Sprintf(" with 1 %s:\n", noun)
		}
		return fmt.Sprintf(" with %d %ss:\n", count, noun)
	}
	switch s.typeNum {
	case pointType:
		fmt.Fprintf(sb, "%s[%s]", name, flags)
	case lineType, circStringType, triangleType:
		fmt.Fprintf(sb, "%s[%s] with %d points", name, flags, s.numPoints())
	case polygonType:
		fmt.Fprintf(sb, "%s[%s]%s", name, flags, plural(len(s.rings), "ring"))
		for i, ring := range s.rings {
			if i > 0 {
				sb.WriteByte('\n')
			}
			fmt.Fprintf(sb, "   ring %d has %d points", i, len(ring)/s.ctype.Dimension())
		}
	case polyhedralSurfaceType:
		fmt.Fprintf(sb, "Object is of unknown type: %d", s.typeNum)
	default:
		fmt.Fprintf(sb, "%s[%s]%s", name, flags, plural(len(s.members), "element"))
		for i, member := range s.members {
			if i > 0 {
				sb.WriteByte('\n')
			}
			summary(sb, member, offset+2, sridUnknown, false, geodetic)
		}
	}
}

// summaryImpl returns the implementation of ST_Summary for geometry, or for geography when `geodetic` is true.
func summaryImpl(geodetic bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		g := args[0].(Geometry)
		var sb strings.Builder
		summary(&sb, g.Shape, 0, g.SRID, !g.Shape.isEmpty() && !hasExactBox(g.Shape), geodetic)
		return sb.String(), nil
	}
}

// stIsRing implements ST_IsRing, which returns whether a curve is closed and simple.
func stIsRing(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.isEmpty() {
		return false, nil
	}
	geosLine, err := s.toGEOS()
	if err != nil {
		return nil, err
	}
	if geosLine.Type != spatial.LineString {
		return nil, pgerror.New(pgcode.Internal, "ST_IsRing() should only be called on a linear feature")
	}
	g := spatial.NewGEOS()
	defer g.Close()
	ring, err := g.IsRing(geosLine)
	if err != nil {
		return nil, geosError("GEOSisRing", err)
	}
	return ring, nil
}

// stIsSimple implements ST_IsSimple.
func stIsSimple(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.isEmpty() {
		return true, nil
	}
	geosShape, err := s.toGEOS()
	if err != nil {
		return nil, err
	}
	g := spatial.NewGEOS()
	defer g.Close()
	simple, err := g.IsSimple(geosShape)
	if err != nil {
		return nil, geosError("GEOSisSimple", err)
	}
	return simple, nil
}

// isClockwise returns whether every polygon or triangle in the shape has a clockwise exterior ring and
// counterclockwise interior rings, or, when `clockwise` is false, the reverse. Shapes without polygons pass.
func isClockwise(s shape, clockwise bool) bool {
	switch s.typeNum {
	case polygonType:
		for i, ring := range s.rings {
			if (signedRingArea(ring, s.ctype) < 0) != (clockwise == (i == 0)) {
				return false
			}
		}
		return true
	case triangleType:
		return len(s.coords) == 0 || (signedRingArea(s.coords, s.ctype) < 0) == clockwise
	case multiPolygonType, collectionType:
		for _, member := range s.members {
			if !isClockwise(member, clockwise) {
				return false
			}
		}
	}
	return true
}

// zmFlag returns the number that PostGIS uses for a coordinates type: zero for XY, one for XYM, two for XYZ, and three
// for XYZM.
func zmFlag(ctype coordinatesType) int {
	return int(ctype&dimXYZ)*2 + int(ctype&dimXYM)/2
}

// signedRingArea returns the two-dimensional area enclosed by a ring, which is positive when the ring runs
// counterclockwise and negative when it runs clockwise.
func signedRingArea(ring []float64, ctype coordinatesType) float64 {
	dims := ctype.Dimension()
	area := 0.0
	for i := 0; i+2*dims <= len(ring); i += dims {
		area += ring[i]*ring[i+dims+1] - ring[i+dims]*ring[i+1]
	}
	return area / 2
}

// stMemSize implements ST_MemSize, which returns the number of bytes that PostGIS uses to store the geometry: an
// eight-byte header, the float32 bounding box when PostGIS keeps one, and the body.
func stMemSize(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	size := 8 + len(serializedBody(nil, s))
	if !s.isEmpty() && !hasExactBox(s) {
		size += 8 * s.ctype.Dimension()
	}
	return int32(size), nil
}

// stPoints implements ST_Points, which returns every vertex of the geometry as a multipoint.
func stPoints(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	multi := shape{typeNum: multiPointType, ctype: g.Shape.ctype}
	dims := g.Shape.ctype.Dimension()
	for _, seq := range g.Shape.sequences() {
		for i := 0; i+dims <= len(seq); i += dims {
			multi.members = append(multi.members, pointShape(seq, g.Shape.ctype, i/dims))
		}
	}
	return g.withShape(multi), nil
}

// structureRoutines returns the routines that inspect the structure of geometries.
func structureRoutines() []extdef.Routine {
	geometryParam := []extdef.Parameter{{Name: "geometry", Type: "geometry"}}
	boolProperty := func(test func(s shape) bool) extdef.Function {
		return func(ctx *sql.Context, args ...any) (any, error) {
			return test(args[0].(Geometry).Shape), nil
		}
	}
	int16Property := func(property func(s shape) int) extdef.Function {
		return func(ctx *sql.Context, args ...any) (any, error) {
			return int16(property(args[0].(Geometry).Shape)), nil
		}
	}
	dims := func(s shape) int { return s.ctype.Dimension() }
	return []extdef.Routine{
		{Name: "st_coorddim", Symbol: "st_coorddim_geometry", Parameters: geometryParam, Returns: "int2", Strict: true, Impl: int16Property(dims)},
		{Name: "st_curven", Symbol: "ST_CurveN", Parameters: []extdef.Parameter{{Name: "geometry", Type: "geometry"}, {Name: "i", Type: "int4"}}, Returns: "geometry", Strict: true, Impl: stCurveN},
		{Name: "st_dimension", Symbol: "LWGEOM_dimension", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return int32(args[0].(Geometry).Shape.dimension()), nil
		}},
		{Name: "st_endpoint", Symbol: "LWGEOM_endpoint_linestring", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stEndPoint},
		{Name: "st_exteriorring", Symbol: "LWGEOM_exteriorring_polygon", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stExteriorRing},
		{Name: "st_geometryn", Symbol: "LWGEOM_geometryn_collection", Parameters: params("geometry", "int4"), Returns: "geometry", Strict: true, Impl: geometryN},
		{Name: "st_hasarc", Symbol: "LWGEOM_has_arc", Parameters: geometryParam, Returns: "bool", Strict: true, Impl: boolProperty(hasArc)},
		{Name: "st_hasm", Symbol: "LWGEOM_hasm", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: boolProperty(shape.hasM)},
		{Name: "st_hasz", Symbol: "LWGEOM_hasz", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: boolProperty(shape.hasZ)},
		{Name: "st_interiorringn", Symbol: "LWGEOM_interiorringn_polygon", Parameters: params("geometry", "int4"), Returns: "geometry", Strict: true, Impl: stInteriorRingN},
		{Name: "st_isclosed", Symbol: "LWGEOM_isclosed", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: boolProperty(func(s shape) bool {
			return !s.isEmpty() && s.isClosedShape()
		})},
		{Name: "st_iscollection", Symbol: "ST_IsCollection", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: boolProperty(func(s shape) bool {
			return isMultiType(s.typeNum)
		})},
		{Name: "st_isempty", Symbol: "LWGEOM_isempty", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: boolProperty(shape.isEmpty)},
		{Name: "st_ispolygonccw", Symbol: "ST_IsPolygonCCW", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: boolProperty(func(s shape) bool {
			return isClockwise(s, false)
		})},
		{Name: "st_ispolygoncw", Symbol: "ST_IsPolygonCW", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: boolProperty(func(s shape) bool {
			return isClockwise(s, true)
		})},
		{Name: "st_isring", Symbol: "isring", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: stIsRing},
		{Name: "st_issimple", Symbol: "issimple", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: stIsSimple},
		{Name: "st_memsize", Symbol: "LWGEOM_mem_size", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: stMemSize},
		{Name: "st_ndims", Symbol: "st_ndims_geometry", Parameters: params("geometry"), Returns: "int2", Strict: true, Impl: int16Property(dims)},
		{Name: "st_npoints", Symbol: "LWGEOM_npoints", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: stNPoints},
		{Name: "st_nrings", Symbol: "LWGEOM_nrings", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: stNRings},
		{Name: "st_numcurves", Symbol: "ST_NumCurves", Parameters: geometryParam, Returns: "int4", Strict: true, Impl: stNumCurves},
		{Name: "st_numgeometries", Symbol: "LWGEOM_numgeometries_collection", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: numGeometries},
		{Name: "st_numinteriorring", Symbol: "st_numinteriorring_geometry", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: stNumInteriorRings},
		{Name: "st_numinteriorrings", Symbol: "st_numinteriorrings_geometry", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: stNumInteriorRings},
		{Name: "st_numpatches", Symbol: "LWGEOM_numpatches", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: stNumPatches},
		{Name: "st_numpoints", Symbol: "LWGEOM_numpoints_linestring", Parameters: params("geometry"), Returns: "int4", Strict: true, Impl: stNumPoints},
		{Name: "st_patchn", Symbol: "LWGEOM_patchn", Parameters: params("geometry", "int4"), Returns: "geometry", Strict: true, Impl: stPatchN},
		{Name: "st_pointn", Symbol: "LWGEOM_pointn_linestring", Parameters: params("geometry", "int4"), Returns: "geometry", Strict: true, Impl: pointN},
		{Name: "st_points", Symbol: "ST_Points", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stPoints},
		{Name: "st_startpoint", Symbol: "LWGEOM_startpoint_linestring", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stStartPoint},
		{Name: "st_summary", Symbol: "LWGEOM_summary", Parameters: params("geometry"), Returns: "text", Strict: true, Impl: summaryImpl(false)},
		{Name: "st_summary", Symbol: "LWGEOM_summary_geography", Parameters: params("geography"), Returns: "text", Strict: true, Impl: summaryImpl(true)},
		{Name: "st_zmflag", Symbol: "LWGEOM_zmflag", Parameters: params("geometry"), Returns: "int2", Strict: true, Impl: int16Property(func(s shape) int {
			return zmFlag(s.ctype)
		})},
	}
}
