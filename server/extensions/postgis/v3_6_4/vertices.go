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
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// maxAllocSize is the largest allocation that Postgres permits, in bytes.
const maxAllocSize = 0x3fffffff

// pointsNear returns whether two points lie within the tolerance of each other in two dimensions, or whether they are
// identical when the tolerance is zero.
func pointsNear(a []float64, b []float64, tolerance float64) bool {
	if tolerance == 0 {
		return equalOrdinates(a, b)
	}
	dx, dy := a[0]-b[0], a[1]-b[1]
	return float64(dx*dx)+float64(dy*dy) <= tolerance*tolerance
}

// removeRepeated returns a copy of a sequence without the points that lie within the tolerance of the point kept
// before them. A point is only removed while enough points remain to keep `minPoints` points. The last point is always
// kept, and it replaces the point kept before it when the two lie within the tolerance.
func removeRepeated(seq []float64, dims int, tolerance float64, minPoints int) []float64 {
	count := len(seq) / dims
	if count == 0 {
		return seq
	}
	kept := append([]float64(nil), seq[:dims]...)
	for i := 1; i < count; i++ {
		point := seq[i*dims : (i+1)*dims]
		keptCount := len(kept) / dims
		if count-i > minPoints-keptCount && pointsNear(point, kept[len(kept)-dims:], tolerance) {
			if i < count-1 {
				continue
			}
			if keptCount > 1 {
				kept = kept[:len(kept)-dims]
			}
		}
		kept = append(kept, point...)
	}
	return kept
}

// removeRepeatedShape removes repeated points from the linestrings and polygon rings of a shape, keeping at least two
// points in a linestring and four in a polygon ring. A multipoint drops its empty points and every point that lies
// within the tolerance of a point that it already kept, in two dimensions.
func removeRepeatedShape(s shape, tolerance float64) (shape, error) {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case lineType:
		s.coords = removeRepeated(s.coords, dims, tolerance, 2)
	case polygonType:
		rings := make([][]float64, len(s.rings))
		for i, ring := range s.rings {
			rings[i] = removeRepeated(ring, dims, tolerance, 4)
		}
		s.rings = rings
	case multiPointType:
		var members []shape
		for _, member := range s.members {
			if member.isEmpty() {
				continue
			}
			isRepeated := false
			for _, existing := range members {
				dx, dy := member.coords[0]-existing.coords[0], member.coords[1]-existing.coords[1]
				if float64(dx*dx)+float64(dy*dy) <= tolerance*tolerance {
					isRepeated = true
					break
				}
			}
			if !isRepeated {
				members = append(members, member)
			}
		}
		s.members = members
	case multiLineType, multiPolygonType, collectionType, compoundType, curvePolyType, multiCurveType, multiSurfaceType, tinType:
		members := make([]shape, len(s.members))
		for i, member := range s.members {
			var err error
			if members[i], err = removeRepeatedShape(member, tolerance); err != nil {
				return shape{}, err
			}
		}
		s.members = members
	case polyhedralSurfaceType:
		return shape{}, pgerror.New(pgcode.Internal, "lwgeom_remove_repeated_points_in_place: unsupported geometry type: PolyhedralSurface")
	}
	return s, nil
}

// stRemoveRepeatedPoints implements ST_RemoveRepeatedPoints.
func stRemoveRepeatedPoints(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	removed, err := removeRepeatedShape(g.Shape, args[1].(float64))
	if err != nil {
		return nil, err
	}
	return g.withShape(removed), nil
}

// segmentizeSequence returns a copy of a sequence where every segment longer than the maximum length is divided into
// equal pieces that are no longer than it, skipping any point that repeats the one before it except the last point of
// a two-point sequence. It returns false, after
// sending the notice that PostGIS sends, when a segment would need more pieces than PostGIS allows. Like PostGIS, it
// fails once the sequence outgrows the largest allocation that Postgres permits, doubling its capacity as it grows.
func segmentizeSequence(ctx *sql.Context, seq []float64, dims int, maxLength float64) ([]float64, bool, error) {
	count := len(seq) / dims
	if count < 2 {
		return seq, true, nil
	}
	segmentized := append([]float64(nil), seq[:dims]...)
	appendPoint := func(point []float64) {
		if !equalOrdinates(segmentized[len(segmentized)-dims:], point) {
			segmentized = append(segmentized, point...)
		}
	}
	capacity := count
	for i := 1; i < count; i++ {
		start, end := seq[(i-1)*dims:i*dims], seq[i*dims:(i+1)*dims]
		length := math.Hypot(end[0]-start[0], end[1]-start[1])
		pieces := 1.0
		if length > maxLength {
			if length/maxLength > math.MaxInt32 {
				sendNotice(ctx, "NOTICE", fmt.Sprintf("ptarray.c:457 - ptarray_segmentize2d: Too many segments required (%e)", length/maxLength))
				return nil, false, nil
			}
			pieces = math.Ceil(length / maxLength)
		}
		for float64(len(segmentized)/dims)+pieces > float64(capacity) {
			capacity *= 2
			if size := capacity * dims * 8; size > maxAllocSize {
				return nil, false, pgerror.Newf(pgcode.Internal, "invalid memory alloc request size %d", size)
			}
		}
		for piece := 1.0; piece < pieces; piece++ {
			point := make([]float64, dims)
			for j := range point {
				point[j] = start[j] + (end[j]-start[j])*piece/pieces
			}
			appendPoint(point)
		}
		appendPoint(end)
	}
	if count == 2 && len(segmentized) == dims {
		segmentized = append(segmentized, seq[dims:]...)
	}
	return segmentized, true, nil
}

// segmentizeShape segmentizes the linestrings and polygon rings of a shape, returning false when PostGIS would return
// null.
func segmentizeShape(ctx *sql.Context, s shape, maxLength float64) (shape, bool, error) {
	dims := s.ctype.Dimension()
	var ok bool
	var err error
	switch s.typeNum {
	case lineType:
		s.coords, ok, err = segmentizeSequence(ctx, s.coords, dims, maxLength)
		return s, ok, err
	case polygonType:
		rings := make([][]float64, len(s.rings))
		for i, ring := range s.rings {
			if rings[i], ok, err = segmentizeSequence(ctx, ring, dims, maxLength); !ok || err != nil {
				return shape{}, ok, err
			}
		}
		s.rings = rings
	case multiLineType, multiPolygonType, collectionType:
		members := make([]shape, len(s.members))
		for i, member := range s.members {
			if members[i], ok, err = segmentizeShape(ctx, member, maxLength); !ok || err != nil {
				return shape{}, ok, err
			}
		}
		s.members = members
	}
	return s, true, nil
}

// stSegmentize implements ST_Segmentize for geometry.
func stSegmentize(ctx *sql.Context, args ...any) (any, error) {
	g, maxLength := args[0].(Geometry), args[1].(float64)
	if maxLength <= 0 {
		return nil, pgerror.Newf(pgcode.Internal, "ST_Segmentize: invalid max_distance %s (must be >= 0)", formatG(maxLength, 6))
	}
	segmentized, ok, err := segmentizeShape(ctx, g.Shape, maxLength)
	if !ok || err != nil {
		return nil, err
	}
	return g.withShape(segmentized), nil
}

// grid is a snapping grid, holding the origin and the cell size for the X, Y, Z, and M ordinates in that order. An
// ordinate whose cell size is not positive is left unsnapped.
type grid struct {
	origin [4]float64
	size   [4]float64
}

// snapPoint returns a point whose ordinates have been moved to the nearest grid line, rounding halves to even. Each
// ordinate is scaled by the reciprocal of the cell size, rounded, and scaled back.
func (g grid) snapPoint(point []float64, ctype coordinatesType) []float64 {
	snap := func(index int, gridIndex int) {
		if g.size[gridIndex] > 0 {
			scale := 1 / g.size[gridIndex]
			point[index] = math.RoundToEven(float64((point[index]-g.origin[gridIndex])*scale))/scale + g.origin[gridIndex]
		}
	}
	snap(0, 0)
	snap(1, 1)
	if ctype.Is3D() {
		snap(2, 2)
	}
	if ctype.IsMeasured() {
		snap(ctype.Dimension()-1, 3)
	}
	return point
}

// snapSequence returns a copy of a sequence snapped to the grid, without the points that repeat the one before them.
func (g grid) snapSequence(seq []float64, ctype coordinatesType) []float64 {
	dims := ctype.Dimension()
	var snapped []float64
	for i := 0; i+dims <= len(seq); i += dims {
		point := g.snapPoint(append([]float64(nil), seq[i:i+dims]...), ctype)
		if len(snapped) < dims || !equalOrdinates(snapped[len(snapped)-dims:], point) {
			snapped = append(snapped, point...)
		}
	}
	return snapped
}

// snapShape snaps a shape to the grid. Linestrings and circular strings left with fewer than two points become empty,
// rings left with fewer than four points are dropped along with the rest of their polygon when the exterior ring is
// dropped, and empty members are dropped from collections.
func (g grid) snapShape(s shape) (shape, error) {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case pointType:
		if len(s.coords) > 0 {
			s.coords = g.snapPoint(append([]float64(nil), s.coords...), s.ctype)
		}
	case lineType, circStringType:
		if s.coords = g.snapSequence(s.coords, s.ctype); len(s.coords) < 2*dims {
			s.coords = nil
		}
	case triangleType:
		if s.coords = g.snapSequence(s.coords, s.ctype); len(s.coords) < 4*dims {
			s.coords = nil
		}
	case polygonType:
		var rings [][]float64
		for i, ring := range s.rings {
			snapped := g.snapSequence(ring, s.ctype)
			if len(snapped) >= 4*dims {
				rings = append(rings, snapped)
			} else if i == 0 {
				break
			}
		}
		s.rings = rings
	case multiPointType, multiLineType, multiPolygonType, collectionType, compoundType, tinType:
		var members []shape
		for _, member := range s.members {
			snapped, err := g.snapShape(member)
			if err != nil {
				return shape{}, err
			}
			if !snapped.isEmpty() {
				members = append(members, snapped)
			}
		}
		s.members = members
	default:
		return shape{}, pgerror.Newf(pgcode.Internal, "lwgeom_grid_in_place: Unsupported geometry type: %s", typeNameForNumber(s.typeNum))
	}
	return s, nil
}

// stSnapToGrid implements every form of ST_SnapToGrid: a single cell size for X and Y, separate X and Y cell sizes,
// an origin with X and Y cell sizes, or an origin point with cell sizes for every ordinate. A grid without any
// positive cell size returns the geometry unchanged.
func stSnapToGrid(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	var snapGrid grid
	switch len(args) {
	case 2:
		snapGrid.size = [4]float64{args[1].(float64), args[1].(float64)}
	case 3:
		snapGrid.size = [4]float64{args[1].(float64), args[2].(float64)}
	case 5:
		snapGrid.origin = [4]float64{args[1].(float64), args[2].(float64)}
		snapGrid.size = [4]float64{args[3].(float64), args[4].(float64)}
	default:
		offset := args[1].(Geometry).Shape
		if offset.typeNum != pointType {
			return nil, pgerror.New(pgcode.Internal, "Offset geometry must be a point")
		}
		if offset.isEmpty() {
			return nil, pgerror.New(pgcode.Internal, "lwgeom_api.c [138] called with n=0 and npoints=0")
		}
		origin := convertPoint(offset.coords, offset.ctype, dimXYZM, 0, 0)
		copy(snapGrid.origin[:], origin)
		snapGrid.size = [4]float64{args[2].(float64), args[3].(float64), args[4].(float64), args[5].(float64)}
	}
	if !(snapGrid.size[0] > 0 || snapGrid.size[1] > 0 || snapGrid.size[2] > 0 || snapGrid.size[3] > 0) {
		return g, nil
	}
	snapped, err := snapGrid.snapShape(g.Shape)
	if err != nil {
		return nil, err
	}
	return g.withShape(snapped), nil
}

// lineArg returns the shape of a geometry argument after checking that it is a linestring.
func lineArg(g Geometry) (shape, error) {
	if g.Shape.typeNum != lineType {
		return shape{}, pgerror.New(pgcode.Internal, "First argument must be a LINESTRING")
	}
	return g.Shape, nil
}

// stAddPoint implements ST_AddPoint, which inserts a point before the 0-based position, or appends it when the
// position is -1 or omitted. The point takes the dimensions of the linestring, and an empty point changes nothing.
func stAddPoint(ctx *sql.Context, args ...any) (any, error) {
	line, err := lineArg(args[0].(Geometry))
	if err != nil {
		return nil, err
	}
	point := args[1].(Geometry).Shape
	if point.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, "Second argument must be a POINT")
	}
	if point.isEmpty() {
		return args[0], nil
	}
	dims := line.ctype.Dimension()
	count := len(line.coords) / dims
	position := count
	if len(args) > 2 && args[2].(int32) != -1 {
		position = int(args[2].(int32))
		if position < 0 || position > count {
			return nil, pgerror.New(pgcode.Internal, "LWGEOM_addpoint: Invalid offset")
		}
	}
	coords := append([]float64(nil), line.coords[:position*dims]...)
	coords = append(coords, convertPoint(point.coords, point.ctype, line.ctype, 0, 0)...)
	line.coords = append(coords, line.coords[position*dims:]...)
	return args[0].(Geometry).withShape(line), nil
}

// stRemovePoint implements ST_RemovePoint, which removes the point at the 0-based position.
func stRemovePoint(ctx *sql.Context, args ...any) (any, error) {
	line, err := lineArg(args[0].(Geometry))
	if err != nil {
		return nil, err
	}
	dims := line.ctype.Dimension()
	count := len(line.coords) / dims
	if count < 3 {
		return nil, pgerror.New(pgcode.Internal, "Can't remove points from a single segment line")
	}
	position := int(args[1].(int32))
	if position < 0 || position >= count {
		return nil, pgerror.Newf(pgcode.Internal, "Point index out of range (0..%d)", count-1)
	}
	coords := append([]float64(nil), line.coords[:position*dims]...)
	line.coords = append(coords, line.coords[(position+1)*dims:]...)
	return args[0].(Geometry).withShape(line), nil
}

// stSetPoint implements ST_SetPoint, which replaces the point at the 0-based position, counting backwards from the end
// when the position is negative. The new point takes the dimensions of the linestring.
func stSetPoint(ctx *sql.Context, args ...any) (any, error) {
	line, err := lineArg(args[0].(Geometry))
	if err != nil {
		return nil, err
	}
	point := args[2].(Geometry).Shape
	if point.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, "Third argument must be a POINT")
	}
	if point.isEmpty() {
		return nil, pgerror.New(pgcode.Internal, "lwgeom_api.c [138] called with n=0 and npoints=0")
	}
	dims := line.ctype.Dimension()
	count := len(line.coords) / dims
	if count == 0 {
		return nil, pgerror.New(pgcode.Internal, "Line has no points")
	}
	position := int(args[1].(int32))
	if position < 0 {
		position += count
	}
	if position < 0 || position >= count {
		return nil, pgerror.Newf(pgcode.Internal, "abs(Point index) out of range (-)(0..%d)", count-1)
	}
	coords := append([]float64(nil), line.coords...)
	copy(coords[position*dims:], convertPoint(point.coords, point.ctype, line.ctype, 0, 0))
	line.coords = coords
	return args[0].(Geometry).withShape(line), nil
}

// vertexRoutines returns the routines that add, remove, move, and snap the vertices of geometries.
func vertexRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	return []extdef.Routine{
		{Name: "st_addpoint", Symbol: "st_addpoint_geometry_geometry", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: stAddPoint},
		{Name: "st_addpoint", Symbol: "st_addpoint_geometry_geometry_int4", Parameters: append(geomParams, extdef.Parameter{Type: "int4"}), Returns: "geometry", Strict: true, Impl: stAddPoint},
		{Name: "st_removepoint", Symbol: "LWGEOM_removepoint", Parameters: params("geometry", "int4"), Returns: "geometry", Strict: true, Impl: stRemovePoint},
		{Name: "st_removerepeatedpoints", Symbol: "ST_RemoveRepeatedPoints", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "tolerance", Type: "float8", Default: `0.0`}}, Returns: "geometry", Strict: true, Impl: stRemoveRepeatedPoints},
		{Name: "st_segmentize", Symbol: "LWGEOM_segmentize2d", Parameters: params("geometry", "float8"), Returns: "geometry", Strict: true, Impl: stSegmentize},
		{Name: "st_setpoint", Symbol: "LWGEOM_setpoint_linestring", Parameters: params("geometry", "int4", "geometry"), Returns: "geometry", Strict: true, Impl: stSetPoint},
		{Name: "st_snaptogrid", Symbol: "LWGEOM_snaptogrid_pointoff", Parameters: append(geomParams, params("float8", "float8", "float8", "float8")...), Returns: "geometry", Strict: true, Impl: stSnapToGrid},
		{Name: "st_snaptogrid", Symbol: "st_snaptogrid_geometry_float8", Parameters: params("geometry", "float8"), Returns: "geometry", Strict: true, Impl: stSnapToGrid},
		{Name: "st_snaptogrid", Symbol: "st_snaptogrid_geometry_float8_float8", Parameters: params("geometry", "float8", "float8"), Returns: "geometry", Strict: true, Impl: stSnapToGrid},
		{Name: "st_snaptogrid", Symbol: "LWGEOM_snaptogrid", Parameters: params("geometry", "float8", "float8", "float8", "float8"), Returns: "geometry", Strict: true, Impl: stSnapToGrid},
	}
}
