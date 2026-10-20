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

/*
#include <stdlib.h>
#include <string.h>
#include <geos_c.h>

static void spatial_geos_message(const char *message, void *buffer) {
	strncpy((char *)buffer, message, 1023);
	((char *)buffer)[1023] = 0;
}

static GEOSContextHandle_t spatial_geos_init(char *buffer) {
	GEOSContextHandle_t handle = GEOS_init_r();
	GEOSContext_setErrorMessageHandler_r(handle, spatial_geos_message, buffer);
	return handle;
}

typedef GEOSGeometry *(*spatial_unary)(GEOSContextHandle_t, const GEOSGeometry *);
typedef GEOSGeometry *(*spatial_binary)(GEOSContextHandle_t, const GEOSGeometry *, const GEOSGeometry *);
typedef GEOSGeometry *(*spatial_binary_prec)(GEOSContextHandle_t, const GEOSGeometry *, const GEOSGeometry *, double);
typedef char (*spatial_predicate)(GEOSContextHandle_t, const GEOSGeometry *, const GEOSGeometry *);
typedef char (*spatial_unary_predicate)(GEOSContextHandle_t, const GEOSGeometry *);

static GEOSGeometry *spatial_call_unary(spatial_unary f, GEOSContextHandle_t h, const GEOSGeometry *a) {
	return f(h, a);
}

static GEOSGeometry *spatial_call_binary(spatial_binary f, GEOSContextHandle_t h, const GEOSGeometry *a, const GEOSGeometry *b) {
	return f(h, a, b);
}

static GEOSGeometry *spatial_call_binary_prec(spatial_binary_prec f, GEOSContextHandle_t h, const GEOSGeometry *a, const GEOSGeometry *b, double size) {
	return f(h, a, b, size);
}

static char spatial_call_predicate(spatial_predicate f, GEOSContextHandle_t h, const GEOSGeometry *a, const GEOSGeometry *b) {
	return f(h, a, b);
}

static char spatial_call_unary_predicate(spatial_unary_predicate f, GEOSContextHandle_t h, const GEOSGeometry *a) {
	return f(h, a);
}
*/
import "C"

import (
	"unsafe"

	"github.com/cockroachdb/errors"
)

// GeometryType is the type of a Geometry, numbered as GEOS numbers them.
type GeometryType int

// The geometry types that GEOS supports.
const (
	Point GeometryType = iota
	LineString
	LinearRing
	Polygon
	MultiPoint
	MultiLineString
	MultiPolygon
	GeometryCollection
)

// Geometry is a geometry that GEOS can operate on. Points, linestrings, and linear rings hold their interleaved
// ordinates in `Coords`, with three ordinates per position when `HasZ` is set and two otherwise. Polygons hold their
// rings in `Members`, the shell first, and collections hold their members there. An empty geometry holds neither.
type Geometry struct {
	Type    GeometryType
	HasZ    bool
	Coords  []float64
	Members []Geometry
}

// GEOS is a GEOS context, which operations run in. A context may only be used by one goroutine at a time.
type GEOS struct {
	handle  C.GEOSContextHandle_t
	message *C.char
}

// NewGEOS returns a new GEOS context, which must be closed once it is no longer needed.
func NewGEOS() *GEOS {
	message := (*C.char)(C.calloc(1024, 1))
	return &GEOS{handle: C.spatial_geos_init(message), message: message}
}

// Close releases the context.
func (g *GEOS) Close() {
	C.GEOS_finish_r(g.handle)
	C.free(unsafe.Pointer(g.message))
}

// err returns the last error that GEOS reported.
func (g *GEOS) err() error {
	return errors.New(C.GoString(g.message))
}

// toGEOS returns the GEOS geometry for the given geometry, which the caller must destroy.
func (g *GEOS) toGEOS(geometry Geometry) (*C.GEOSGeometry, error) {
	var result *C.GEOSGeometry
	switch geometry.Type {
	case Point, LineString, LinearRing:
		if len(geometry.Coords) == 0 {
			if geometry.Type == Point {
				result = C.GEOSGeom_createEmptyPoint_r(g.handle)
			} else {
				result = C.GEOSGeom_createEmptyLineString_r(g.handle)
			}
			break
		}
		dims, hasZ := 2, C.int(0)
		if geometry.HasZ {
			dims, hasZ = 3, 1
		}
		sequence := C.GEOSCoordSeq_copyFromBuffer_r(g.handle, (*C.double)(&geometry.Coords[0]), C.uint(len(geometry.Coords)/dims), hasZ, 0)
		if sequence == nil {
			return nil, g.err()
		}
		switch geometry.Type {
		case Point:
			result = C.GEOSGeom_createPoint_r(g.handle, sequence)
		case LineString:
			result = C.GEOSGeom_createLineString_r(g.handle, sequence)
		default:
			result = C.GEOSGeom_createLinearRing_r(g.handle, sequence)
		}
	case Polygon:
		if len(geometry.Members) == 0 {
			result = C.GEOSGeom_createEmptyPolygon_r(g.handle)
			break
		}
		rings, err := g.toGEOSMembers(geometry.Members, LinearRing)
		if err != nil {
			return nil, err
		}
		holes := (**C.GEOSGeometry)(nil)
		if len(rings) > 1 {
			holes = &rings[1]
		}
		result = C.GEOSGeom_createPolygon_r(g.handle, rings[0], holes, C.uint(len(rings)-1))
		C.free(unsafe.Pointer(&rings[0]))
	default:
		if len(geometry.Members) == 0 {
			result = C.GEOSGeom_createEmptyCollection_r(g.handle, C.int(geometry.Type))
			break
		}
		members, err := g.toGEOSMembers(geometry.Members, -1)
		if err != nil {
			return nil, err
		}
		result = C.GEOSGeom_createCollection_r(g.handle, C.int(geometry.Type), &members[0], C.uint(len(members)))
		C.free(unsafe.Pointer(&members[0]))
	}
	if result == nil {
		return nil, g.err()
	}
	return result, nil
}

// toGEOSMembers returns the GEOS geometries for the given members in an array allocated by C, converting each to
// `asType` unless it is negative. The caller owns the geometries and must free the array.
func (g *GEOS) toGEOSMembers(members []Geometry, asType GeometryType) ([]*C.GEOSGeometry, error) {
	pointers := unsafe.Slice((**C.GEOSGeometry)(C.calloc(C.size_t(len(members)), C.size_t(unsafe.Sizeof(uintptr(0))))), len(members))
	for i, member := range members {
		if asType >= 0 {
			member.Type = asType
		}
		converted, err := g.toGEOS(member)
		if err != nil {
			g.destroy(pointers[:i])
			C.free(unsafe.Pointer(&pointers[0]))
			return nil, err
		}
		pointers[i] = converted
	}
	return pointers, nil
}

// destroy destroys the given GEOS geometries.
func (g *GEOS) destroy(geometries []*C.GEOSGeometry) {
	for _, geometry := range geometries {
		C.GEOSGeom_destroy_r(g.handle, geometry)
	}
}

// fromGEOS returns the geometry for the given GEOS geometry, whose members all take the Z flag that is given.
func (g *GEOS) fromGEOS(geometry *C.GEOSGeometry, hasZ bool) (Geometry, error) {
	result := Geometry{Type: GeometryType(C.GEOSGeomTypeId_r(g.handle, geometry)), HasZ: hasZ}
	if C.GEOSisEmpty_r(g.handle, geometry) == 1 {
		return result, nil
	}
	switch result.Type {
	case Point, LineString, LinearRing:
		sequence := C.GEOSGeom_getCoordSeq_r(g.handle, geometry)
		var size C.uint
		if sequence == nil || C.GEOSCoordSeq_getSize_r(g.handle, sequence, &size) == 0 {
			return Geometry{}, g.err()
		}
		dims, hasZ := 2, C.int(0)
		if result.HasZ {
			dims, hasZ = 3, 1
		}
		result.Coords = make([]float64, int(size)*dims)
		if size > 0 && C.GEOSCoordSeq_copyToBuffer_r(g.handle, sequence, (*C.double)(&result.Coords[0]), hasZ, 0) == 0 {
			return Geometry{}, g.err()
		}
	case Polygon:
		holes := int(C.GEOSGetNumInteriorRings_r(g.handle, geometry))
		result.Members = make([]Geometry, holes+1)
		for i := range result.Members {
			var ring *C.GEOSGeometry
			if i == 0 {
				ring = C.GEOSGetExteriorRing_r(g.handle, geometry)
			} else {
				ring = C.GEOSGetInteriorRingN_r(g.handle, geometry, C.int(i-1))
			}
			member, err := g.fromGEOS(ring, hasZ)
			if err != nil {
				return Geometry{}, err
			}
			result.Members[i] = member
		}
	default:
		result.Members = make([]Geometry, int(C.GEOSGetNumGeometries_r(g.handle, geometry)))
		for i := range result.Members {
			member, err := g.fromGEOS(C.GEOSGetGeometryN_r(g.handle, geometry, C.int(i)), hasZ)
			if err != nil {
				return Geometry{}, err
			}
			result.Members[i] = member
		}
	}
	return result, nil
}

// result converts and destroys a GEOS geometry that an operation returned, which is nil when the operation failed.
func (g *GEOS) result(geometry *C.GEOSGeometry) (Geometry, error) {
	if geometry == nil {
		return Geometry{}, g.err()
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometry)
	return g.fromGEOS(geometry, C.GEOSHasZ_r(g.handle, geometry) == 1)
}

// unary runs a GEOS operation that takes one geometry and returns another.
func (g *GEOS) unary(operation C.spatial_unary, a Geometry) (Geometry, error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	return g.result(C.spatial_call_unary(operation, g.handle, geometryA))
}

// binary runs a GEOS operation that takes two geometries and returns another. A negative grid size runs `operation`,
// and any other size runs `precise` with that size.
func (g *GEOS) binary(operation C.spatial_binary, precise C.spatial_binary_prec, a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	geometryA, geometryB, err := g.toGEOSPair(a, b)
	if err != nil {
		return Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	defer C.GEOSGeom_destroy_r(g.handle, geometryB)
	if gridSize < 0 {
		return g.result(C.spatial_call_binary(operation, g.handle, geometryA, geometryB))
	}
	return g.result(C.spatial_call_binary_prec(precise, g.handle, geometryA, geometryB, C.double(gridSize)))
}

// predicate runs a GEOS predicate on two geometries.
func (g *GEOS) predicate(predicate C.spatial_predicate, a Geometry, b Geometry) (bool, error) {
	geometryA, geometryB, err := g.toGEOSPair(a, b)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	defer C.GEOSGeom_destroy_r(g.handle, geometryB)
	return g.boolean(C.spatial_call_predicate(predicate, g.handle, geometryA, geometryB))
}

// unaryPredicate runs a GEOS predicate on one geometry.
func (g *GEOS) unaryPredicate(predicate C.spatial_unary_predicate, a Geometry) (bool, error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	return g.boolean(C.spatial_call_unary_predicate(predicate, g.handle, geometryA))
}

// boolean converts the result of a GEOS function returning a boolean, which is 2 when the function failed.
func (g *GEOS) boolean(result C.char) (bool, error) {
	if result == 2 {
		return false, g.err()
	}
	return result == 1, nil
}

// toGEOSPair returns the GEOS geometries for two geometries, which the caller must destroy.
func (g *GEOS) toGEOSPair(a Geometry, b Geometry) (*C.GEOSGeometry, *C.GEOSGeometry, error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return nil, nil, err
	}
	geometryB, err := g.toGEOS(b)
	if err != nil {
		C.GEOSGeom_destroy_r(g.handle, geometryA)
		return nil, nil, err
	}
	return geometryA, geometryB, nil
}

// Intersection returns the intersection of two geometries, snapped to the given grid size unless it is negative.
func (g *GEOS) Intersection(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return g.binary(C.spatial_binary(C.GEOSIntersection_r), C.spatial_binary_prec(C.GEOSIntersectionPrec_r), a, b, gridSize)
}

// Union returns the union of two geometries, snapped to the given grid size unless it is negative.
func (g *GEOS) Union(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return g.binary(C.spatial_binary(C.GEOSUnion_r), C.spatial_binary_prec(C.GEOSUnionPrec_r), a, b, gridSize)
}

// Difference returns the part of `a` outside `b`, snapped to the given grid size unless it is negative.
func (g *GEOS) Difference(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return g.binary(C.spatial_binary(C.GEOSDifference_r), C.spatial_binary_prec(C.GEOSDifferencePrec_r), a, b, gridSize)
}

// SymDifference returns the parts of two geometries outside each other, snapped to the given grid size unless it is
// negative.
func (g *GEOS) SymDifference(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return g.binary(C.spatial_binary(C.GEOSSymDifference_r), C.spatial_binary_prec(C.GEOSSymDifferencePrec_r), a, b, gridSize)
}

// UnaryUnion returns the union of a geometry's members, snapped to the given grid size unless it is negative.
func (g *GEOS) UnaryUnion(a Geometry, gridSize float64) (Geometry, error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	if gridSize < 0 {
		return g.result(C.GEOSUnaryUnion_r(g.handle, geometryA))
	}
	return g.result(C.GEOSUnaryUnionPrec_r(g.handle, geometryA, C.double(gridSize)))
}

// ConvexHull returns the convex hull of a geometry.
func (g *GEOS) ConvexHull(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSConvexHull_r), a)
}

// Centroid returns the centroid of a geometry.
func (g *GEOS) Centroid(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSGetCentroid_r), a)
}

// PointOnSurface returns a point inside a geometry.
func (g *GEOS) PointOnSurface(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSPointOnSurface_r), a)
}

// Intersects returns whether two geometries share a point.
func (g *GEOS) Intersects(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSIntersects_r), a, b)
}

// Disjoint returns whether two geometries share no point.
func (g *GEOS) Disjoint(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSDisjoint_r), a, b)
}

// Touches returns whether two geometries share boundary points and no interior points.
func (g *GEOS) Touches(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSTouches_r), a, b)
}

// Crosses returns whether two geometries cross.
func (g *GEOS) Crosses(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSCrosses_r), a, b)
}

// Within returns whether `a` lies within `b`.
func (g *GEOS) Within(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSWithin_r), a, b)
}

// Contains returns whether `a` contains `b`.
func (g *GEOS) Contains(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSContains_r), a, b)
}

// Overlaps returns whether two geometries overlap.
func (g *GEOS) Overlaps(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSOverlaps_r), a, b)
}

// Equals returns whether two geometries cover the same points.
func (g *GEOS) Equals(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSEquals_r), a, b)
}

// Covers returns whether every point of `b` lies in `a`.
func (g *GEOS) Covers(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSCovers_r), a, b)
}

// CoveredBy returns whether every point of `a` lies in `b`.
func (g *GEOS) CoveredBy(a Geometry, b Geometry) (bool, error) {
	return g.predicate(C.spatial_predicate(C.GEOSCoveredBy_r), a, b)
}

// Relate returns the DE-9IM intersection matrix of two geometries under the given boundary node rule.
func (g *GEOS) Relate(a Geometry, b Geometry, boundaryNodeRule int) (string, error) {
	geometryA, geometryB, err := g.toGEOSPair(a, b)
	if err != nil {
		return "", err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	defer C.GEOSGeom_destroy_r(g.handle, geometryB)
	matrix := C.GEOSRelateBoundaryNodeRule_r(g.handle, geometryA, geometryB, C.int(boundaryNodeRule))
	if matrix == nil {
		return "", g.err()
	}
	defer C.GEOSFree_r(g.handle, unsafe.Pointer(matrix))
	return C.GoString(matrix), nil
}

// RelatePatternMatch returns whether a DE-9IM intersection matrix matches a pattern.
func (g *GEOS) RelatePatternMatch(matrix string, pattern string) (bool, error) {
	cMatrix, cPattern := C.CString(matrix), C.CString(pattern)
	defer C.free(unsafe.Pointer(cMatrix))
	defer C.free(unsafe.Pointer(cPattern))
	return g.boolean(C.GEOSRelatePatternMatch_r(g.handle, cMatrix, cPattern))
}

// The end cap and join styles of a buffer, numbered as GEOS numbers them.
const (
	BufferRound  = 1
	BufferFlat   = 2
	BufferSquare = 3
	BufferMitre  = 2
	BufferBevel  = 3
)

// BufferParams holds the options of a buffer: the number of segments per quarter circle, the end cap and join
// styles, the mitre limit, and whether only one side is buffered, which is the left side for positive distances.
type BufferParams struct {
	QuadSegs    int
	EndCap      int
	Join        int
	MitreLimit  float64
	SingleSided bool
}

// Buffer returns the area within the given distance of a geometry.
func (g *GEOS) Buffer(a Geometry, distance float64, params BufferParams) (Geometry, error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	bufferParams := C.GEOSBufferParams_create_r(g.handle)
	if bufferParams == nil {
		return Geometry{}, g.err()
	}
	defer C.GEOSBufferParams_destroy_r(g.handle, bufferParams)
	singleSided := C.int(0)
	if params.SingleSided {
		singleSided = 1
	}
	C.GEOSBufferParams_setQuadrantSegments_r(g.handle, bufferParams, C.int(params.QuadSegs))
	C.GEOSBufferParams_setEndCapStyle_r(g.handle, bufferParams, C.int(params.EndCap))
	C.GEOSBufferParams_setJoinStyle_r(g.handle, bufferParams, C.int(params.Join))
	C.GEOSBufferParams_setMitreLimit_r(g.handle, bufferParams, C.double(params.MitreLimit))
	C.GEOSBufferParams_setSingleSided_r(g.handle, bufferParams, singleSided)
	return g.result(C.GEOSBufferWithParams_r(g.handle, geometryA, bufferParams, C.double(distance)))
}

// IsValidDetail returns whether a geometry is valid under the given flags, and otherwise the reason and the location
// of the problem.
func (g *GEOS) IsValidDetail(a Geometry, flags int) (valid bool, reason string, location Geometry, err error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return false, "", Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	var cReason *C.char
	var cLocation *C.GEOSGeometry
	switch C.GEOSisValidDetail_r(g.handle, geometryA, C.int(flags), &cReason, &cLocation) {
	case 1:
		return true, "", Geometry{}, nil
	case 0:
		reason = C.GoString(cReason)
		C.GEOSFree_r(g.handle, unsafe.Pointer(cReason))
		if cLocation != nil {
			location, err = g.result(cLocation)
		}
		return false, reason, location, err
	default:
		return false, "", Geometry{}, g.err()
	}
}

// CoverageUnion returns the union of a collection of polygons that only share edges.
func (g *GEOS) CoverageUnion(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSCoverageUnion_r), a)
}

// Group holds geometries converted to GEOS once, for operations between many pairs of them.
type Group struct {
	g          *GEOS
	geometries []*C.GEOSGeometry
}

// NewGroup converts the given geometries for repeated operations. The group must be closed once it is no longer
// needed.
func (g *GEOS) NewGroup(geometries []Geometry) (*Group, error) {
	group := &Group{g: g, geometries: make([]*C.GEOSGeometry, 0, len(geometries))}
	for _, geometry := range geometries {
		converted, err := g.toGEOS(geometry)
		if err != nil {
			group.Close()
			return nil, err
		}
		group.geometries = append(group.geometries, converted)
	}
	return group, nil
}

// Close releases the group's geometries.
func (group *Group) Close() {
	group.g.destroy(group.geometries)
}

// Intersects returns whether the geometries at the given indexes share a point.
func (group *Group) Intersects(i int, j int) (bool, error) {
	return group.g.boolean(C.GEOSIntersects_r(group.g.handle, group.geometries[i], group.geometries[j]))
}

// Distance returns the distance between the geometries at the given indexes.
func (group *Group) Distance(i int, j int) (float64, error) {
	var distance C.double
	if C.GEOSDistance_r(group.g.handle, group.geometries[i], group.geometries[j], &distance) == 0 {
		return 0, group.g.err()
	}
	return float64(distance), nil
}

// Distance returns the shortest distance between two geometries.
func (g *GEOS) Distance(a Geometry, b Geometry) (float64, error) {
	geometryA, geometryB, err := g.toGEOSPair(a, b)
	if err != nil {
		return 0, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	defer C.GEOSGeom_destroy_r(g.handle, geometryB)
	var distance C.double
	if C.GEOSDistance_r(g.handle, geometryA, geometryB, &distance) == 0 {
		return 0, g.err()
	}
	return float64(distance), nil
}

// MinimumRotatedRectangle returns the rectangle of least width that holds a geometry.
func (g *GEOS) MinimumRotatedRectangle(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSMinimumRotatedRectangle_r), a)
}

// IsSimple returns whether a geometry has no anomalous points, such as self-intersections.
func (g *GEOS) IsSimple(a Geometry) (bool, error) {
	return g.unaryPredicate(C.spatial_unary_predicate(C.GEOSisSimple_r), a)
}

// IsRing returns whether a linestring is closed and simple.
func (g *GEOS) IsRing(a Geometry) (bool, error) {
	return g.unaryPredicate(C.spatial_unary_predicate(C.GEOSisRing_r), a)
}

// ContainsProperly returns whether `a` contains `b` without `b` touching the boundary of `a`.
func (g *GEOS) ContainsProperly(a Geometry, b Geometry) (bool, error) {
	geometryA, geometryB, err := g.toGEOSPair(a, b)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	defer C.GEOSGeom_destroy_r(g.handle, geometryB)
	prepared := C.GEOSPrepare_r(g.handle, geometryA)
	if prepared == nil {
		return false, g.err()
	}
	defer C.GEOSPreparedGeom_destroy_r(g.handle, prepared)
	return g.boolean(C.GEOSPreparedContainsProperly_r(g.handle, prepared, geometryB))
}

// Normalize returns a geometry in GEOS' normal form, with its members, rings, and points in a canonical order.
func (g *GEOS) Normalize(a Geometry) (Geometry, error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	if C.GEOSNormalize_r(g.handle, geometryA) != 0 {
		return Geometry{}, g.err()
	}
	return g.fromGEOS(geometryA, C.GEOSHasZ_r(g.handle, geometryA) == 1)
}

// withGEOS converts a geometry to GEOS, runs an operation on it that returns a new GEOS geometry or nil on failure,
// and converts the result back.
func (g *GEOS) withGEOS(a Geometry, op func(geometryA *C.GEOSGeometry) *C.GEOSGeometry) (Geometry, error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	return g.result(op(geometryA))
}

// withGEOSPair converts two geometries to GEOS, runs an operation on them that returns a new GEOS geometry or nil on
// failure, and converts the result back.
func (g *GEOS) withGEOSPair(a Geometry, b Geometry, op func(geometryA *C.GEOSGeometry, geometryB *C.GEOSGeometry) *C.GEOSGeometry) (Geometry, error) {
	geometryA, geometryB, err := g.toGEOSPair(a, b)
	if err != nil {
		return Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	defer C.GEOSGeom_destroy_r(g.handle, geometryB)
	return g.result(op(geometryA, geometryB))
}

// pairDistance converts two geometries to GEOS and runs a distance operation on them, which returns 0 on failure.
func (g *GEOS) pairDistance(a Geometry, b Geometry, op func(geometryA *C.GEOSGeometry, geometryB *C.GEOSGeometry, distance *C.double) C.int) (float64, error) {
	geometryA, geometryB, err := g.toGEOSPair(a, b)
	if err != nil {
		return 0, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	defer C.GEOSGeom_destroy_r(g.handle, geometryB)
	var distance C.double
	if op(geometryA, geometryB, &distance) == 0 {
		return 0, g.err()
	}
	return float64(distance), nil
}

// MakeValid returns a valid geometry covering the same points, built from the input's linework when `structure` is
// false and from its rings' structure otherwise, dropping collapsed parts unless `keepCollapsed` is set.
func (g *GEOS) MakeValid(a Geometry, structure bool, keepCollapsed bool) (Geometry, error) {
	params := C.GEOSMakeValidParams_create_r(g.handle)
	defer C.GEOSMakeValidParams_destroy_r(g.handle, params)
	method, keep := C.enum_GEOSMakeValidMethods(C.GEOS_MAKE_VALID_LINEWORK), C.int(0)
	if structure {
		method = C.GEOS_MAKE_VALID_STRUCTURE
	}
	if keepCollapsed {
		keep = 1
	}
	C.GEOSMakeValidParams_setMethod_r(g.handle, params, method)
	C.GEOSMakeValidParams_setKeepCollapsed_r(g.handle, params, keep)
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSMakeValidWithParams_r(g.handle, geometryA, params)
	})
}

// SimplifyPreserveTopology returns a simplified geometry that keeps the input's topology.
func (g *GEOS) SimplifyPreserveTopology(a Geometry, tolerance float64) (Geometry, error) {
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSTopologyPreserveSimplify_r(g.handle, geometryA, C.double(tolerance))
	})
}

// LineMerge returns the lines of a geometry joined where they meet, keeping each line's direction when `directed`
// is set.
func (g *GEOS) LineMerge(a Geometry, directed bool) (Geometry, error) {
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		if directed {
			return C.GEOSLineMergeDirected_r(g.handle, geometryA)
		}
		return C.GEOSLineMerge_r(g.handle, geometryA)
	})
}

// Node returns the lines of a geometry split wherever they cross or touch.
func (g *GEOS) Node(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSNode_r), a)
}

// SharedPaths returns the paths that two linear geometries share, as a collection of the paths running in the same
// direction and the paths running in opposite directions.
func (g *GEOS) SharedPaths(a Geometry, b Geometry) (Geometry, error) {
	return g.withGEOSPair(a, b, func(geometryA *C.GEOSGeometry, geometryB *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSharedPaths_r(g.handle, geometryA, geometryB)
	})
}

// Snap returns `a` with its vertices and segments snapped to the vertices of `b` within the tolerance.
func (g *GEOS) Snap(a Geometry, b Geometry, tolerance float64) (Geometry, error) {
	return g.withGEOSPair(a, b, func(geometryA *C.GEOSGeometry, geometryB *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSnap_r(g.handle, geometryA, geometryB, C.double(tolerance))
	})
}

// OffsetCurve returns the line at the given distance from a line, on its left for positive distances.
func (g *GEOS) OffsetCurve(a Geometry, distance float64, params BufferParams) (Geometry, error) {
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSOffsetCurve_r(g.handle, geometryA, C.double(distance), C.int(params.QuadSegs), C.int(params.Join), C.double(params.MitreLimit))
	})
}

// BuildArea returns the polygons that the lines of a geometry enclose.
func (g *GEOS) BuildArea(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSBuildArea_r), a)
}

// ConcaveHull returns a concave polygon holding a geometry, as concave as `ratio` allows, from 0 for the most concave
// to 1 for the convex hull.
func (g *GEOS) ConcaveHull(a Geometry, ratio float64, allowHoles bool) (Geometry, error) {
	holes := C.uint(0)
	if allowHoles {
		holes = 1
	}
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSConcaveHull_r(g.handle, geometryA, C.double(ratio), holes)
	})
}

// PolygonHullSimplify returns a simplified polygon that holds the input when `outer` is set and lies within it
// otherwise, keeping the given fraction of its vertices.
func (g *GEOS) PolygonHullSimplify(a Geometry, outer bool, vertexFraction float64) (Geometry, error) {
	isOuter := C.uint(0)
	if outer {
		isOuter = 1
	}
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSPolygonHullSimplify_r(g.handle, geometryA, isOuter, C.double(vertexFraction))
	})
}

// DelaunayTriangulation returns the Delaunay triangulation of a geometry's vertices, as triangles or as their edges
// when `onlyEdges` is set.
func (g *GEOS) DelaunayTriangulation(a Geometry, tolerance float64, onlyEdges bool) (Geometry, error) {
	edges := C.int(0)
	if onlyEdges {
		edges = 1
	}
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSDelaunayTriangulation_r(g.handle, geometryA, C.double(tolerance), edges)
	})
}

// ConstrainedDelaunayTriangulation returns the triangles of a polygon's constrained Delaunay triangulation.
func (g *GEOS) ConstrainedDelaunayTriangulation(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSConstrainedDelaunayTriangulation_r), a)
}

// VoronoiDiagram returns the Voronoi diagram of a geometry's vertices, extended to hold `envelope` when it is not
// nil, as polygons or as their edges when `onlyEdges` is set.
func (g *GEOS) VoronoiDiagram(a Geometry, envelope *Geometry, tolerance float64, onlyEdges bool) (Geometry, error) {
	edges := C.int(0)
	if onlyEdges {
		edges = 1
	}
	if envelope == nil {
		return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
			return C.GEOSVoronoiDiagram_r(g.handle, geometryA, nil, C.double(tolerance), edges)
		})
	}
	return g.withGEOSPair(a, *envelope, func(geometryA *C.GEOSGeometry, geometryB *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSVoronoiDiagram_r(g.handle, geometryA, geometryB, C.double(tolerance), edges)
	})
}

// MinimumClearance returns the smallest distance that moving a vertex of the geometry by would make it invalid.
func (g *GEOS) MinimumClearance(a Geometry) (float64, error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return 0, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	var distance C.double
	if C.GEOSMinimumClearance_r(g.handle, geometryA, &distance) == 2 {
		return 0, g.err()
	}
	return float64(distance), nil
}

// MinimumClearanceLine returns the line that spans a geometry's minimum clearance.
func (g *GEOS) MinimumClearanceLine(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSMinimumClearanceLine_r), a)
}

// FrechetDistance returns the discrete Fréchet distance between two geometries, after densifying their segments into
// the given fraction of their length unless it is not positive.
func (g *GEOS) FrechetDistance(a Geometry, b Geometry, densifyFraction float64) (float64, error) {
	return g.pairDistance(a, b, func(geometryA *C.GEOSGeometry, geometryB *C.GEOSGeometry, distance *C.double) C.int {
		if densifyFraction > 0 {
			return C.GEOSFrechetDistanceDensify_r(g.handle, geometryA, geometryB, C.double(densifyFraction), distance)
		}
		return C.GEOSFrechetDistance_r(g.handle, geometryA, geometryB, distance)
	})
}

// HausdorffDistance returns the discrete Hausdorff distance between two geometries, after densifying their segments
// into the given fraction of their length unless it is not positive.
func (g *GEOS) HausdorffDistance(a Geometry, b Geometry, densifyFraction float64) (float64, error) {
	return g.pairDistance(a, b, func(geometryA *C.GEOSGeometry, geometryB *C.GEOSGeometry, distance *C.double) C.int {
		if densifyFraction > 0 {
			return C.GEOSHausdorffDistanceDensify_r(g.handle, geometryA, geometryB, C.double(densifyFraction), distance)
		}
		return C.GEOSHausdorffDistance_r(g.handle, geometryA, geometryB, distance)
	})
}

// ClipByRect returns the part of a geometry within a rectangle, quickly and possibly invalid.
func (g *GEOS) ClipByRect(a Geometry, minX float64, minY float64, maxX float64, maxY float64) (Geometry, error) {
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSClipByRect_r(g.handle, geometryA, C.double(minX), C.double(minY), C.double(maxX), C.double(maxY))
	})
}

// MaximumInscribedCircle returns the line from the center of the largest circle within a polygon to the nearest
// point of its boundary, accurate to the tolerance.
func (g *GEOS) MaximumInscribedCircle(a Geometry, tolerance float64) (Geometry, error) {
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSMaximumInscribedCircle_r(g.handle, geometryA, C.double(tolerance))
	})
}

// LargestEmptyCircle returns the line from the center of the largest circle that holds none of the obstacles to the
// nearest obstacle, accurate to the tolerance. The center lies within the boundary when one is given, and within the
// convex hull of the obstacles otherwise.
func (g *GEOS) LargestEmptyCircle(obstacles Geometry, boundary *Geometry, tolerance float64) (Geometry, error) {
	if boundary == nil {
		return g.withGEOS(obstacles, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
			return C.GEOSLargestEmptyCircle_r(g.handle, geometryA, nil, C.double(tolerance))
		})
	}
	return g.withGEOSPair(obstacles, *boundary, func(geometryA *C.GEOSGeometry, geometryB *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSLargestEmptyCircle_r(g.handle, geometryA, geometryB, C.double(tolerance))
	})
}

// ConcaveHullOfPolygons returns a concave polygon holding the polygons of a geometry, filling the gaps between them as
// `ratio` allows, from 0 for the least filling to 1 for the convex hull. A tight hull follows the outer edges of the
// polygons.
func (g *GEOS) ConcaveHullOfPolygons(a Geometry, ratio float64, tight bool, allowHoles bool) (Geometry, error) {
	isTight, holes := C.uint(0), C.uint(0)
	if tight {
		isTight = 1
	}
	if allowHoles {
		holes = 1
	}
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSConcaveHullOfPolygons_r(g.handle, geometryA, C.double(ratio), isTight, holes)
	})
}

// MinimumBoundingCircle returns the smallest circle holding a geometry, as a polygon, along with its center and
// radius.
func (g *GEOS) MinimumBoundingCircle(a Geometry) (circle Geometry, center Geometry, radius float64, err error) {
	geometryA, err := g.toGEOS(a)
	if err != nil {
		return Geometry{}, Geometry{}, 0, err
	}
	defer C.GEOSGeom_destroy_r(g.handle, geometryA)
	var cRadius C.double
	var cCenter *C.GEOSGeometry
	circle, err = g.result(C.GEOSMinimumBoundingCircle_r(g.handle, geometryA, &cRadius, &cCenter))
	if err != nil {
		return Geometry{}, Geometry{}, 0, err
	}
	if center, err = g.result(cCenter); err != nil {
		return Geometry{}, Geometry{}, 0, err
	}
	return circle, center, float64(cRadius), nil
}

// SetPrecision returns a geometry with its coordinates rounded to the grid size, keeping it valid. The flags are
// GEOS' precision flags.
func (g *GEOS) SetPrecision(a Geometry, gridSize float64, flags int) (Geometry, error) {
	return g.withGEOS(a, func(geometryA *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSGeom_setPrecision_r(g.handle, geometryA, C.double(gridSize), C.int(flags))
	})
}

// Polygonize returns the polygons that the lines of the given geometries enclose, as a collection.
func (g *GEOS) Polygonize(geometries []Geometry) (Geometry, error) {
	if len(geometries) == 0 {
		return g.result(C.GEOSGeom_createEmptyCollection_r(g.handle, C.GEOS_GEOMETRYCOLLECTION))
	}
	members, err := g.toGEOSMembers(geometries, -1)
	if err != nil {
		return Geometry{}, err
	}
	defer C.free(unsafe.Pointer(&members[0]))
	defer g.destroy(members)
	return g.result(C.GEOSPolygonize_r(g.handle, &members[0], C.uint(len(members))))
}

// Boundary returns the boundary of a geometry.
func (g *GEOS) Boundary(a Geometry) (Geometry, error) {
	return g.unary(C.spatial_unary(C.GEOSBoundary_r), a)
}
