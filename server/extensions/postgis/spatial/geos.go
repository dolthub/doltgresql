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

// This file implements the GEOS operations that simplefeatures provides. Its overlay, relate, and buffer operations
// are a Go port of JTS, which GEOS is a C++ port of, so they follow the same algorithms.

import (
	"sort"

	"github.com/cockroachdb/errors"
	"github.com/peterstace/simplefeatures/geom"
)

// NewGEOS returns a context for running GEOS operations.
func NewGEOS() *GEOS {
	return &GEOS{}
}

// Close releases the context.
func (g *GEOS) Close() {}

// toGeom converts a geometry to its simplefeatures form.
func toGeom(g Geometry) (geom.Geometry, error) {
	ctype := geom.DimXY
	if g.HasZ {
		ctype = geom.DimXYZ
	}
	sequence := func(coords []float64) geom.Sequence {
		return geom.NewSequence(coords, ctype)
	}
	switch g.Type {
	case Point:
		if len(g.Coords) == 0 {
			return geom.Point{}.ForceCoordinatesType(ctype).AsGeometry(), nil
		}
		return geom.NewPoint(sequence(g.Coords).Get(0)).AsGeometry(), nil
	case LineString, LinearRing:
		return geom.NewLineString(sequence(g.Coords)).AsGeometry(), nil
	case Polygon:
		polygon, err := toGeomPolygon(g, ctype)
		if err != nil {
			return geom.Geometry{}, err
		}
		return polygon.AsGeometry(), nil
	case MultiPoint:
		points := make([]geom.Point, len(g.Members))
		for i, member := range g.Members {
			converted, err := toGeom(member)
			if err != nil {
				return geom.Geometry{}, err
			}
			point, ok := converted.AsPoint()
			if !ok {
				return geom.Geometry{}, errors.New("a multipoint member is not a point")
			}
			points[i] = point
		}
		return geom.NewMultiPoint(points).ForceCoordinatesType(ctype).AsGeometry(), nil
	case MultiLineString:
		lines := make([]geom.LineString, len(g.Members))
		for i, member := range g.Members {
			lines[i] = geom.NewLineString(sequence(member.Coords))
		}
		return geom.NewMultiLineString(lines).ForceCoordinatesType(ctype).AsGeometry(), nil
	case MultiPolygon:
		polygons := make([]geom.Polygon, len(g.Members))
		for i, member := range g.Members {
			polygon, err := toGeomPolygon(member, ctype)
			if err != nil {
				return geom.Geometry{}, err
			}
			polygons[i] = polygon
		}
		return geom.NewMultiPolygon(polygons).ForceCoordinatesType(ctype).AsGeometry(), nil
	case GeometryCollection:
		members := make([]geom.Geometry, len(g.Members))
		for i, member := range g.Members {
			converted, err := toGeom(member)
			if err != nil {
				return geom.Geometry{}, err
			}
			members[i] = converted
		}
		return geom.NewGeometryCollection(members).ForceCoordinatesType(ctype).AsGeometry(), nil
	}
	return geom.Geometry{}, errors.Errorf("unknown geometry type: %d", g.Type)
}

// toGeomPolygon converts a polygon to its simplefeatures form.
func toGeomPolygon(g Geometry, ctype geom.CoordinatesType) (geom.Polygon, error) {
	if g.Type != Polygon {
		return geom.Polygon{}, errors.New("a multipolygon member is not a polygon")
	}
	rings := make([]geom.LineString, len(g.Members))
	for i, ring := range g.Members {
		rings[i] = geom.NewLineString(geom.NewSequence(ring.Coords, ctype))
	}
	return geom.NewPolygon(rings).ForceCoordinatesType(ctype), nil
}

// fromGeom converts a simplefeatures geometry to a geometry, keeping Z when it has it.
func fromGeom(g geom.Geometry) Geometry {
	hasZ := g.CoordinatesType().Is3D()
	coords := func(s geom.Sequence) []float64 {
		if s.Length() == 0 {
			return nil
		}
		if hasZ {
			s = s.ForceCoordinatesType(geom.DimXYZ)
		} else {
			s = s.ForceCoordinatesType(geom.DimXY)
		}
		var out []float64
		for i := 0; i < s.Length(); i++ {
			c := s.Get(i)
			out = append(out, c.X, c.Y)
			if hasZ {
				out = append(out, c.Z)
			}
		}
		return out
	}
	polygon := func(p geom.Polygon) Geometry {
		result := Geometry{Type: Polygon, HasZ: hasZ}
		if p.IsEmpty() {
			return result
		}
		result.Members = append(result.Members, Geometry{Type: LinearRing, HasZ: hasZ, Coords: coords(p.ExteriorRing().Coordinates())})
		for i := 0; i < p.NumInteriorRings(); i++ {
			result.Members = append(result.Members, Geometry{Type: LinearRing, HasZ: hasZ, Coords: coords(p.InteriorRingN(i).Coordinates())})
		}
		return result
	}
	switch g.Type() {
	case geom.TypePoint:
		point := g.MustAsPoint()
		c, ok := point.Coordinates()
		if !ok {
			return Geometry{Type: Point, HasZ: hasZ}
		}
		result := Geometry{Type: Point, HasZ: hasZ, Coords: []float64{c.X, c.Y}}
		if hasZ {
			result.Coords = append(result.Coords, c.Z)
		}
		return result
	case geom.TypeLineString:
		return Geometry{Type: LineString, HasZ: hasZ, Coords: coords(g.MustAsLineString().Coordinates())}
	case geom.TypePolygon:
		return polygon(g.MustAsPolygon())
	case geom.TypeMultiPoint:
		mp := g.MustAsMultiPoint()
		result := Geometry{Type: MultiPoint, HasZ: hasZ}
		for i := 0; i < mp.NumPoints(); i++ {
			result.Members = append(result.Members, fromGeom(mp.PointN(i).AsGeometry()))
		}
		return result
	case geom.TypeMultiLineString:
		ml := g.MustAsMultiLineString()
		result := Geometry{Type: MultiLineString, HasZ: hasZ}
		for i := 0; i < ml.NumLineStrings(); i++ {
			result.Members = append(result.Members, Geometry{Type: LineString, HasZ: hasZ, Coords: coords(ml.LineStringN(i).Coordinates())})
		}
		return result
	case geom.TypeMultiPolygon:
		mp := g.MustAsMultiPolygon()
		result := Geometry{Type: MultiPolygon, HasZ: hasZ}
		for i := 0; i < mp.NumPolygons(); i++ {
			result.Members = append(result.Members, polygon(mp.PolygonN(i)))
		}
		return result
	default:
		gc := g.MustAsGeometryCollection()
		result := Geometry{Type: GeometryCollection, HasZ: hasZ}
		for i := 0; i < gc.NumGeometries(); i++ {
			result.Members = append(result.Members, fromGeom(gc.GeometryN(i)))
		}
		return result
	}
}

// toGeomPair converts two geometries to their simplefeatures forms.
func toGeomPair(a Geometry, b Geometry) (geom.Geometry, geom.Geometry, error) {
	geomA, err := toGeom(a)
	if err != nil {
		return geom.Geometry{}, geom.Geometry{}, err
	}
	geomB, err := toGeom(b)
	if err != nil {
		return geom.Geometry{}, geom.Geometry{}, err
	}
	return geomA, geomB, nil
}

// hasZ returns whether the geometry or any of its members has Z.
func hasZ(g Geometry) bool {
	if g.HasZ {
		return true
	}
	for _, member := range g.Members {
		if hasZ(member) {
			return true
		}
	}
	return false
}

// overlay runs a binary overlay operation. Only floating precision, given by a negative grid size, is supported, as
// are only geometries without Z, since GEOS interpolates the Z of overlay results, which simplefeatures does not.
func overlay(op func(a geom.Geometry, b geom.Geometry) (geom.Geometry, error), a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	if gridSize >= 0 || hasZ(a) || hasZ(b) {
		return Geometry{}, ErrNotSupported
	}
	geomA, geomB, err := toGeomPair(a, b)
	if err != nil {
		return Geometry{}, err
	}
	result, err := op(geomA, geomB)
	if err != nil {
		return Geometry{}, err
	}
	return sortPoints(fromGeom(result)), nil
}

// sortPoints sorts the points of an overlay result by their X and then Y, which is the order that GEOS gives them in,
// since it collects them in maps ordered by coordinate.
func sortPoints(g Geometry) Geometry {
	switch g.Type {
	case MultiPoint:
		sort.SliceStable(g.Members, func(i, j int) bool { return pointLess(g.Members[i], g.Members[j]) })
	case GeometryCollection:
		var indexes []int
		var points []Geometry
		for i, member := range g.Members {
			if member.Type == Point {
				indexes = append(indexes, i)
				points = append(points, member)
			}
		}
		sort.SliceStable(points, func(i, j int) bool { return pointLess(points[i], points[j]) })
		for k, i := range indexes {
			g.Members[i] = points[k]
		}
	}
	return g
}

// pointLess returns whether the first point sorts before the second by X and then Y.
func pointLess(a Geometry, b Geometry) bool {
	if len(a.Coords) < 2 || len(b.Coords) < 2 {
		return len(a.Coords) < len(b.Coords)
	}
	if a.Coords[0] != b.Coords[0] {
		return a.Coords[0] < b.Coords[0]
	}
	return a.Coords[1] < b.Coords[1]
}

// Intersection returns the points that two geometries share, snapped to the grid size when it is not negative.
func (g *GEOS) Intersection(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return overlay(geom.Intersection, a, b, gridSize)
}

// Union returns the points of either geometry.
func (g *GEOS) Union(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return overlay(geom.Union, a, b, gridSize)
}

// Difference returns the points of the first geometry that the second does not have.
func (g *GEOS) Difference(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return overlay(geom.Difference, a, b, gridSize)
}

// SymDifference returns the points that exactly one of the geometries has.
func (g *GEOS) SymDifference(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return overlay(geom.SymmetricDifference, a, b, gridSize)
}

// UnaryUnion returns the union of a geometry's members.
func (g *GEOS) UnaryUnion(a Geometry, gridSize float64) (Geometry, error) {
	if gridSize >= 0 || hasZ(a) {
		return Geometry{}, ErrNotSupported
	}
	geomA, err := toGeom(a)
	if err != nil {
		return Geometry{}, err
	}
	result, err := geom.UnaryUnion(geomA)
	if err != nil {
		return Geometry{}, err
	}
	return sortPoints(fromGeom(result)), nil
}

// unary runs an operation on a single geometry.
func unary(a Geometry, op func(geom.Geometry) geom.Geometry) (Geometry, error) {
	geomA, err := toGeom(a)
	if err != nil {
		return Geometry{}, err
	}
	return fromGeom(op(geomA)), nil
}

// ConvexHull returns the smallest convex geometry that holds the geometry.
func (g *GEOS) ConvexHull(a Geometry) (Geometry, error) {
	return convexHull(a), nil
}

// Centroid returns the center of mass of the geometry.
func (g *GEOS) Centroid(a Geometry) (Geometry, error) {
	return unary(a, func(g geom.Geometry) geom.Geometry { return g.Centroid().AsGeometry() })
}

// PointOnSurface returns a point that the geometry holds.
func (g *GEOS) PointOnSurface(a Geometry) (Geometry, error) {
	return unary(a, func(g geom.Geometry) geom.Geometry { return g.PointOnSurface().AsGeometry() })
}

// Boundary returns the boundary of the geometry.
func (g *GEOS) Boundary(a Geometry) (Geometry, error) {
	return unary(a, geom.Geometry.Boundary)
}

// predicate runs a binary predicate.
func predicate(op func(a geom.Geometry, b geom.Geometry) (bool, error), a Geometry, b Geometry) (bool, error) {
	geomA, geomB, err := toGeomPair(a, b)
	if err != nil {
		return false, err
	}
	return op(geomA, geomB)
}

// Intersects returns whether two geometries share a point.
func (g *GEOS) Intersects(a Geometry, b Geometry) (bool, error) {
	return predicate(func(a geom.Geometry, b geom.Geometry) (bool, error) { return geom.Intersects(a, b), nil }, a, b)
}

// Disjoint returns whether two geometries share no point.
func (g *GEOS) Disjoint(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.Disjoint, a, b)
}

// Touches returns whether two geometries share boundary points but no interior points.
func (g *GEOS) Touches(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.Touches, a, b)
}

// Crosses returns whether two geometries share some interior points but not all.
func (g *GEOS) Crosses(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.Crosses, a, b)
}

// Within returns whether the first geometry lies in the second.
func (g *GEOS) Within(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.Within, a, b)
}

// Contains returns whether the second geometry lies in the first.
func (g *GEOS) Contains(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.Contains, a, b)
}

// Overlaps returns whether two geometries of the same dimension share some interior points but not all.
func (g *GEOS) Overlaps(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.Overlaps, a, b)
}

// Equals returns whether two geometries are topologically equal.
func (g *GEOS) Equals(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.Equals, a, b)
}

// Covers returns whether no point of the second geometry lies outside the first.
func (g *GEOS) Covers(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.Covers, a, b)
}

// CoveredBy returns whether no point of the first geometry lies outside the second.
func (g *GEOS) CoveredBy(a Geometry, b Geometry) (bool, error) {
	return predicate(geom.CoveredBy, a, b)
}

// ContainsProperly returns whether the second geometry lies in the interior of the first.
func (g *GEOS) ContainsProperly(a Geometry, b Geometry) (bool, error) {
	matrix, err := g.Relate(a, b, 1)
	if err != nil {
		return false, err
	}
	return geom.RelateMatches(matrix, "T**FF*FF*")
}

// Relate returns the DE-9IM intersection matrix of two geometries. Only the Mod-2 boundary node rule, numbered 1 as
// GEOS numbers it, is supported.
func (g *GEOS) Relate(a Geometry, b Geometry, boundaryNodeRule int) (string, error) {
	if boundaryNodeRule != 1 {
		return "", ErrNotSupported
	}
	geomA, geomB, err := toGeomPair(a, b)
	if err != nil {
		return "", err
	}
	return geom.Relate(geomA, geomB)
}

// RelatePatternMatch returns whether a DE-9IM intersection matrix matches a pattern.
func (g *GEOS) RelatePatternMatch(matrix string, pattern string) (bool, error) {
	return geom.RelateMatches(matrix, pattern)
}

// Buffer returns the area within the given distance of a geometry.
func (g *GEOS) Buffer(a Geometry, distance float64, params BufferParams) (Geometry, error) {
	geomA, err := toGeom(a)
	if err != nil {
		return Geometry{}, err
	}
	opts := []geom.BufferOption{geom.BufferQuadSegments(params.QuadSegs)}
	switch params.EndCap {
	case BufferFlat:
		opts = append(opts, geom.BufferEndCapFlat())
	case BufferSquare:
		opts = append(opts, geom.BufferEndCapSquare())
	default:
		opts = append(opts, geom.BufferEndCapRound())
	}
	switch params.Join {
	case BufferMitre:
		opts = append(opts, geom.BufferJoinStyleMitre(params.MitreLimit))
	case BufferBevel:
		opts = append(opts, geom.BufferJoinStyleBevel())
	default:
		opts = append(opts, geom.BufferJoinStyleRound())
	}
	if params.SingleSided {
		opts = append(opts, geom.BufferSingleSided())
	}
	result, err := geom.Buffer(geomA, distance, opts...)
	if err != nil {
		return Geometry{}, err
	}
	return fromGeom(result), nil
}

// Distance returns the shortest distance between two geometries.
func (g *GEOS) Distance(a Geometry, b Geometry) (float64, error) {
	geomA, geomB, err := toGeomPair(a, b)
	if err != nil {
		return 0, err
	}
	distance, ok := geom.Distance(geomA, geomB)
	if !ok {
		return 0, nil
	}
	return distance, nil
}

// IsSimple returns whether the geometry has no anomalous points, such as self-intersections.
func (g *GEOS) IsSimple(a Geometry) (bool, error) {
	geomA, err := toGeom(a)
	if err != nil {
		return false, err
	}
	simple, wellDefined := geomA.IsSimple()
	if !wellDefined {
		return true, nil
	}
	return simple, nil
}

// IsRing returns whether the geometry is a closed and simple linestring.
func (g *GEOS) IsRing(a Geometry) (bool, error) {
	geomA, err := toGeom(a)
	if err != nil {
		return false, err
	}
	line, ok := geomA.AsLineString()
	if !ok || line.IsEmpty() || !line.IsClosed() {
		return false, nil
	}
	simple, _ := geomA.IsSimple()
	return simple, nil
}

// MinimumRotatedRectangle returns the rectangle of least width that holds the geometry.
func (g *GEOS) MinimumRotatedRectangle(a Geometry) (Geometry, error) {
	return minimumRectangle(a), nil
}

// Group is a group of converted geometries for repeated operations.
type Group struct {
	geometries []geom.Geometry
}

// NewGroup converts the given geometries for repeated operations. The group must be closed once it is no longer
// needed.
func (g *GEOS) NewGroup(geometries []Geometry) (*Group, error) {
	group := &Group{geometries: make([]geom.Geometry, len(geometries))}
	for i, geometry := range geometries {
		converted, err := toGeom(geometry)
		if err != nil {
			return nil, err
		}
		group.geometries[i] = converted
	}
	return group, nil
}

// Close releases the group's geometries.
func (group *Group) Close() {}

// Intersects returns whether the geometries at the given indexes share a point.
func (group *Group) Intersects(i int, j int) (bool, error) {
	return geom.Intersects(group.geometries[i], group.geometries[j]), nil
}

// Distance returns the distance between the geometries at the given indexes.
func (group *Group) Distance(i int, j int) (float64, error) {
	distance, _ := geom.Distance(group.geometries[i], group.geometries[j])
	return distance, nil
}
