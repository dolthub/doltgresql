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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// convertPoint returns the ordinates of a point converted from one coordinates type to another. Ordinates that both
// types share are kept, and a Z or M ordinate that the point lacks is set to `z` or `m`.
func convertPoint(point []float64, from coordinatesType, to coordinatesType, z float64, m float64) []float64 {
	converted := []float64{point[0], point[1]}
	if to.Is3D() {
		if from.Is3D() {
			z = point[2]
		}
		converted = append(converted, z)
	}
	if to.IsMeasured() {
		if from.IsMeasured() {
			m = point[from.Dimension()-1]
		}
		converted = append(converted, m)
	}
	return converted
}

// forceDims returns a copy of the shape with the given coordinates type, setting any Z or M ordinate that the shape
// lacks to `z` or `m`.
func forceDims(s shape, ctype coordinatesType, z float64, m float64) shape {
	from := s.ctype
	return s.mapCoords(func(point []float64) []float64 {
		return convertPoint(point, from, ctype, z, m)
	}).withDims(ctype)
}

// forceDimsImpl returns an implementation of a function that forces a geometry to the given coordinates type. Its
// optional arguments hold the Z value and then the M value, or only the M value when the type has no Z.
func forceDimsImpl(ctype coordinatesType) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		g := args[0].(Geometry)
		var z, m float64
		if ctype.Is3D() && len(args) > 1 {
			z = args[1].(float64)
			args = args[1:]
		}
		if ctype.IsMeasured() && len(args) > 1 {
			m = args[1].(float64)
		}
		return g.withShape(forceDims(g.Shape, ctype, z, m)), nil
	}
}

// stMulti implements ST_Multi. As in PostGIS, an empty multicurve becomes an empty polyhedral surface.
func stMulti(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if g.Shape.typeNum == multiCurveType && g.Shape.isEmpty() {
		return g.withShape(shape{typeNum: polyhedralSurfaceType, ctype: g.Shape.ctype}), nil
	}
	return g.withShape(asMulti(g.Shape)), nil
}

// stForceCollection implements ST_ForceCollection. Collections, compound curves, curve polygons, and surfaces hand
// their members to the new collection, and any other shape becomes its only member.
func stForceCollection(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	switch g.Shape.typeNum {
	case collectionType:
		return g, nil
	case pointType, lineType, polygonType, circStringType, triangleType:
		return g.withShape(shape{typeNum: collectionType, ctype: g.Shape.ctype, members: []shape{g.Shape}}), nil
	default:
		return g.withShape(shape{typeNum: collectionType, ctype: g.Shape.ctype, members: g.Shape.members}), nil
	}
}

// stForceCurve implements ST_ForceCurve, which turns linestrings into compound curves, polygons into curve polygons,
// and their multi-geometries into multicurves and multisurfaces.
func stForceCurve(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	s := g.Shape
	switch s.typeNum {
	case lineType:
		curve := shape{typeNum: compoundType, ctype: s.ctype}
		if len(s.coords) > 0 {
			curve.members = []shape{s}
		}
		return g.withShape(curve), nil
	case polygonType:
		curve := shape{typeNum: curvePolyType, ctype: s.ctype}
		for _, ring := range s.rings {
			curve.members = append(curve.members, shape{typeNum: lineType, ctype: s.ctype, coords: ring})
		}
		return g.withShape(curve), nil
	case multiLineType:
		return g.withShape(shape{typeNum: multiCurveType, ctype: s.ctype, members: s.members}), nil
	case multiPolygonType:
		return g.withShape(shape{typeNum: multiSurfaceType, ctype: s.ctype, members: s.members}), nil
	default:
		return g, nil
	}
}

// forceSFS converts a shape to the types of the OGC Simple Features specification: curves are linearized, and unless
// `keepTriangles` is true, triangles become polygons and surfaces become collections of polygons.
func forceSFS(s shape, keepTriangles bool) (shape, error) {
	switch s.typeNum {
	case curvePolyType:
		if s.isEmpty() {
			return shape{}, pgerror.New(pgcode.Internal, "lwpoly_construct: need at least 1 ring")
		}
		return strokeShape(s, defaultStroke)
	case circStringType, compoundType, multiCurveType, multiSurfaceType:
		return strokeShape(s, defaultStroke)
	case triangleType:
		if keepTriangles {
			return s, nil
		}
		if s.isEmpty() {
			return shape{}, pgerror.New(pgcode.Internal, "lwpoly_from_lwlines: shell must have at least 4 points")
		}
		return shape{typeNum: polygonType, ctype: s.ctype, rings: [][]float64{s.coords}}, nil
	case tinType, polyhedralSurfaceType, collectionType:
		if keepTriangles && s.typeNum != collectionType {
			return s, nil
		}
		collection := shape{typeNum: collectionType, ctype: s.ctype, members: make([]shape, len(s.members))}
		for i, member := range s.members {
			converted, err := forceSFS(member, keepTriangles)
			if err != nil {
				return shape{}, err
			}
			collection.members[i] = converted
		}
		return collection, nil
	default:
		return s, nil
	}
}

// stForceSFS implements ST_ForceSFS, which keeps triangles and surfaces only when asked for version 1.2.
func stForceSFS(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	converted, err := forceSFS(g.Shape, len(args) > 1 && args[1].(string) == "1.2")
	if err != nil {
		return nil, err
	}
	return g.withShape(converted), nil
}

// reverseSequence returns a copy of an ordinate sequence with its points in the opposite order.
func reverseSequence(seq []float64, dims int) []float64 {
	reversed := make([]float64, 0, len(seq))
	for i := len(seq) - dims; i >= 0; i -= dims {
		reversed = append(reversed, seq[i:i+dims]...)
	}
	return reversed
}

// reverseShape returns a copy of the shape where every sequence runs in the opposite direction. A compound curve also
// reverses the order of its members.
func reverseShape(s shape) shape {
	dims := s.ctype.Dimension()
	if hasSequence(s.typeNum) {
		s.coords = reverseSequence(s.coords, dims)
		return s
	}
	if s.typeNum == polygonType {
		rings := make([][]float64, len(s.rings))
		for i, ring := range s.rings {
			rings[i] = reverseSequence(ring, dims)
		}
		s.rings = rings
		return s
	}
	members := make([]shape, len(s.members))
	for i, member := range s.members {
		if s.typeNum == compoundType {
			members[len(members)-1-i] = reverseShape(member)
		} else {
			members[i] = reverseShape(member)
		}
	}
	s.members = members
	return s
}

// stReverse implements ST_Reverse.
func stReverse(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	return g.withShape(reverseShape(g.Shape)), nil
}

// stFlipCoordinates implements ST_FlipCoordinates, which swaps the X and Y ordinates of every point.
func stFlipCoordinates(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	return g.withShape(g.Shape.mapCoords(func(point []float64) []float64 {
		point[0], point[1] = point[1], point[0]
		return point
	})), nil
}

// stSwapOrdinates implements ST_SwapOrdinates, which swaps the two ordinates named by a pair of letters from x, y, z,
// and m.
func stSwapOrdinates(ctx *sql.Context, args ...any) (any, error) {
	g, spec := args[0].(Geometry), args[1].(string)
	if len(spec) != 2 {
		return nil, pgerror.Newf(pgcode.Internal, "Invalid ordinate specification. Need two letters from the set (x,y,z,m). Got '%s'", spec)
	}
	var indexes [2]int
	for i, name := range strings.ToLower(spec) {
		switch name {
		case 'x':
			indexes[i] = 0
		case 'y':
			indexes[i] = 1
		case 'z':
			if !g.Shape.hasZ() {
				return nil, pgerror.New(pgcode.Internal, "Geometry does not have a Z ordinate")
			}
			indexes[i] = 2
		case 'm':
			if !g.Shape.hasM() {
				return nil, pgerror.New(pgcode.Internal, "Geometry does not have an M ordinate")
			}
			indexes[i] = g.Shape.ctype.Dimension() - 1
		default:
			return nil, pgerror.Newf(pgcode.Internal, "Invalid ordinate name '%c'. Expected x,y,z or m", spec[i])
		}
	}
	return g.withShape(g.Shape.mapCoords(func(point []float64) []float64 {
		point[indexes[0]], point[indexes[1]] = point[indexes[1]], point[indexes[0]]
		return point
	})), nil
}

// orientRings returns a copy of the shape where the exterior ring of every polygon and triangle runs clockwise and
// every interior ring runs counterclockwise, or the reverse when `clockwise` is false. Rings without area count as
// counterclockwise. Only polygons, triangles, multipolygons, and geometry collections are changed.
func orientRings(s shape, clockwise bool) shape {
	dims := s.ctype.Dimension()
	orient := func(ring []float64, exterior bool) []float64 {
		isCounterclockwise := signedRingArea(ring, s.ctype) >= 0
		if isCounterclockwise == (clockwise == exterior) {
			return reverseSequence(ring, dims)
		}
		return ring
	}
	switch s.typeNum {
	case polygonType:
		rings := make([][]float64, len(s.rings))
		for i, ring := range s.rings {
			rings[i] = orient(ring, i == 0)
		}
		s.rings = rings
	case triangleType:
		s.coords = orient(s.coords, true)
	case multiPolygonType, collectionType:
		members := make([]shape, len(s.members))
		for i, member := range s.members {
			members[i] = orientRings(member, clockwise)
		}
		s.members = members
	}
	return s
}

// orientImpl returns an implementation of ST_ForcePolygonCW when `clockwise` is true, and of ST_ForcePolygonCCW
// otherwise.
func orientImpl(clockwise bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		g := args[0].(Geometry)
		return g.withShape(orientRings(g.Shape, clockwise)), nil
	}
}

// atomicMembers calls `visit` with every member of the shape that is not a collection, in order, descending into
// multi-geometries, geometry collections, and surfaces. Compound curves and curve polygons are visited whole unless
// `intoCurves` is true, in which case their members are visited instead.
func atomicMembers(s shape, intoCurves bool, visit func(member shape)) {
	switch s.typeNum {
	case multiPointType, multiLineType, multiPolygonType, collectionType, multiCurveType, multiSurfaceType,
		polyhedralSurfaceType, tinType:
		for _, member := range s.members {
			atomicMembers(member, intoCurves, visit)
		}
	case compoundType, curvePolyType:
		if !intoCurves {
			visit(s)
			return
		}
		for _, member := range s.members {
			atomicMembers(member, intoCurves, visit)
		}
	default:
		visit(s)
	}
}

// stCollectionExtract implements ST_CollectionExtract. A collection yields the multi-geometry of its nonempty points,
// linestrings, or polygons, choosing the type of highest dimension that it holds when no type is given. Any other
// shape is returned when it has the requested type, and is replaced by an empty shape of that type otherwise.
func stCollectionExtract(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	typeNum := 0
	if len(args) > 1 {
		typeNum = int(args[1].(int32))
	}
	if typeNum < 0 || typeNum > polygonType {
		return nil, pgerror.New(pgcode.Internal, "ST_CollectionExtract: only point, linestring and polygon may be extracted")
	}
	s := g.Shape
	switch s.typeNum {
	case multiPointType, multiLineType, multiPolygonType, collectionType, multiCurveType, multiSurfaceType,
		polyhedralSurfaceType, tinType, compoundType, curvePolyType:
	default:
		if typeNum == 0 || typeNum == s.typeNum {
			return g, nil
		}
		return g.withShape(shape{typeNum: typeNum, ctype: s.ctype}), nil
	}
	if typeNum == 0 {
		atomicMembers(s, true, func(member shape) {
			if member.typeNum <= polygonType && member.typeNum > typeNum {
				typeNum = member.typeNum
			}
		})
		if typeNum == 0 {
			return g.withShape(shape{typeNum: collectionType, ctype: s.ctype}), nil
		}
	}
	extracted := shape{typeNum: multiTypeNumber(typeNum), ctype: s.ctype}
	atomicMembers(s, true, func(member shape) {
		if member.typeNum == typeNum && !member.isEmpty() {
			extracted.members = append(extracted.members, member)
		}
	})
	return g.withShape(extracted), nil
}

// stCollectionHomogenize implements ST_CollectionHomogenize. The members of a collection are gathered by type, in the
// order of the type numbers, and each type becomes its only member or a multi-geometry of its members. A single type
// is returned alone, and several types are returned as a geometry collection. An empty collection stays empty.
func stCollectionHomogenize(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	s := g.Shape
	switch s.typeNum {
	case multiPointType, multiLineType, multiPolygonType, collectionType, multiCurveType, multiSurfaceType,
		polyhedralSurfaceType, tinType:
	default:
		return g, nil
	}
	if s.isEmpty() {
		return g.withShape(shape{typeNum: s.typeNum, ctype: s.ctype}), nil
	}
	return g.withShape(homogenize(s)), nil
}

// homogenize gathers the members of a nonempty collection by type, in the order of the type numbers, and makes each
// type its only member or a multi-geometry of its members. A single type is returned alone, and several types are
// returned as a geometry collection.
func homogenize(s shape) shape {
	groups := make(map[int][]shape)
	atomicMembers(s, false, func(member shape) {
		groups[member.typeNum] = append(groups[member.typeNum], member)
	})
	var homogenized []shape
	for typeNum := pointType; typeNum <= tinType; typeNum++ {
		members := groups[typeNum]
		if len(members) == 1 {
			homogenized = append(homogenized, members[0])
		} else if len(members) > 1 {
			homogenized = append(homogenized, shape{typeNum: multiTypeNumber(typeNum), ctype: s.ctype, members: members})
		}
	}
	if len(homogenized) == 1 {
		return homogenized[0]
	}
	return shape{typeNum: collectionType, ctype: s.ctype, members: homogenized}
}

// coercionRoutines returns the routines that change the dimensions, types, and orientation of geometries.
func coercionRoutines() []extdef.Routine {
	geomParam := extdef.Parameter{Name: "geom", Type: "geometry"}
	zParam := extdef.Parameter{Name: "zvalue", Type: "float8", Default: `0.0`}
	mParam := extdef.Parameter{Name: "mvalue", Type: "float8", Default: `0.0`}
	return []extdef.Routine{
		{Name: "st_collectionextract", Symbol: "st_collectionextract_geometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stCollectionExtract},
		{Name: "st_collectionextract", Symbol: "st_collectionextract_geometry_int4", Parameters: params("geometry", "int4"), Returns: "geometry", Strict: true, Impl: stCollectionExtract},
		{Name: "st_collectionhomogenize", Symbol: "ST_CollectionHomogenize", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stCollectionHomogenize},
		{Name: "st_flipcoordinates", Symbol: "ST_FlipCoordinates", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stFlipCoordinates},
		{Name: "st_force2d", Symbol: "LWGEOM_force_2d", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: forceDimsImpl(dimXY)},
		{Name: "st_force3d", Symbol: "st_force3d_geometry_float8", Parameters: []extdef.Parameter{geomParam, zParam}, Returns: "geometry", Strict: true, Impl: forceDimsImpl(dimXYZ)},
		{Name: "st_force3dm", Symbol: "LWGEOM_force_3dm", Parameters: []extdef.Parameter{geomParam, mParam}, Returns: "geometry", Strict: true, Impl: forceDimsImpl(dimXYM)},
		{Name: "st_force3dz", Symbol: "LWGEOM_force_3dz", Parameters: []extdef.Parameter{geomParam, zParam}, Returns: "geometry", Strict: true, Impl: forceDimsImpl(dimXYZ)},
		{Name: "st_force4d", Symbol: "LWGEOM_force_4d", Parameters: []extdef.Parameter{geomParam, zParam, mParam}, Returns: "geometry", Strict: true, Impl: forceDimsImpl(dimXYZM)},
		{Name: "st_forcecollection", Symbol: "LWGEOM_force_collection", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stForceCollection},
		{Name: "st_forcecurve", Symbol: "LWGEOM_force_curve", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stForceCurve},
		{Name: "st_forcepolygonccw", Symbol: "st_forcepolygonccw_geometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: orientImpl(false)},
		{Name: "st_forcepolygoncw", Symbol: "st_forcepolygoncw_geometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: orientImpl(true)},
		{Name: "st_forcerhr", Symbol: "st_forcerhr_geometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: orientImpl(true)},
		{Name: "st_forcesfs", Symbol: "st_forcesfs_geometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stForceSFS},
		{Name: "st_forcesfs", Symbol: "st_forcesfs_geometry_text", Parameters: []extdef.Parameter{{Type: "geometry"}, {Name: "version", Type: "text"}}, Returns: "geometry", Strict: true, Impl: stForceSFS},
		{Name: "st_multi", Symbol: "LWGEOM_force_multi", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stMulti},
		{Name: "st_reverse", Symbol: "LWGEOM_reverse", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stReverse},
		{Name: "st_swapordinates", Symbol: "ST_SwapOrdinates", Parameters: []extdef.Parameter{geomParam, {Name: "ords", Type: "cstring"}}, Returns: "geometry", Strict: true, Impl: stSwapOrdinates},
	}
}
