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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// simplifySequence applies the Douglas-Peucker algorithm to a sequence: starting from its two ends, the point farthest
// from the segment between two kept points is kept whenever its distance exceeds the tolerance, and the halves on
// either side are examined the same way. Until `minPoints` points are kept, the farthest point is kept regardless of
// the tolerance. Short sequences are returned unchanged.
func simplifySequence(seq []float64, ctype coordinatesType, tolerance float64, minPoints int) []float64 {
	dims := ctype.Dimension()
	count := len(seq) / dims
	if count < 3 || count <= minPoints {
		return seq
	}
	kept := make([]bool, count)
	kept[0], kept[count-1] = true, true
	keptCount := 2
	threshold := func() float64 {
		if keptCount >= minPoints {
			return tolerance * tolerance
		}
		return -1
	}
	var split func(first int, last int)
	split = func(first int, last int) {
		farthest, farthestDistance := first, threshold()
		for i := first + 1; i < last; i++ {
			if d := segmentDistanceSquared(seq[i*dims:], seq[first*dims:], seq[last*dims:]); d > farthestDistance {
				farthest, farthestDistance = i, d
			}
		}
		if farthest == first {
			return
		}
		kept[farthest] = true
		keptCount++
		split(first, farthest)
		split(farthest, last)
	}
	split(0, count-1)
	simplified := make([]float64, 0, keptCount*dims)
	for i, keep := range kept {
		if keep {
			simplified = append(simplified, seq[i*dims:(i+1)*dims]...)
		}
	}
	return simplified
}

// segmentDistanceSquared returns the squared two-dimensional distance from point `p` to the segment from `a` to `b`.
func segmentDistanceSquared(p []float64, a []float64, b []float64) float64 {
	dx, dy := b[0]-a[0], b[1]-a[1]
	if dx == 0 && dy == 0 {
		return float64((p[0]-a[0])*(p[0]-a[0])) + float64((p[1]-a[1])*(p[1]-a[1]))
	}
	t := (float64((p[0]-a[0])*dx) + float64((p[1]-a[1])*dy)) / (float64(dx*dx) + float64(dy*dy))
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	x, y := a[0]+float64(t*dx), a[1]+float64(t*dy)
	return float64((p[0]-x)*(p[0]-x)) + float64((p[1]-y)*(p[1]-y))
}

// simplifyShape simplifies every linestring, ring, and triangle in the shape. Unless `preserveCollapsed` is true,
// linestrings that collapse to a single location, triangles that lose a vertex, and polygons whose exterior ring
// falls below four points become empty, and empty members are dropped from collections. Interior rings below four
// points are always dropped.
func simplifyShape(s shape, tolerance float64, preserveCollapsed bool) (shape, error) {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case pointType:
		return s, nil
	case triangleType:
		if preserveCollapsed {
			return s, nil
		}
		s.coords = simplifySequence(s.coords, s.ctype, tolerance, 0)
		if len(s.coords)/dims < 3 {
			s.coords = nil
		}
		return s, nil
	case lineType:
		s.coords = simplifySequence(s.coords, s.ctype, tolerance, 2)
		if !preserveCollapsed && len(s.coords) == 2*dims && s.coords[0] == s.coords[dims] && s.coords[1] == s.coords[dims+1] {
			s.coords = nil
		}
		return s, nil
	case polygonType:
		var rings [][]float64
		for i, ring := range s.rings {
			minPoints := 0
			if preserveCollapsed && i == 0 {
				minPoints = 4
			}
			simplified := simplifySequence(ring, s.ctype, tolerance, minPoints)
			if len(simplified)/dims < 4 {
				if i == 0 {
					rings = nil
					break
				}
				continue
			}
			rings = append(rings, simplified)
		}
		s.rings = rings
		return s, nil
	case multiPointType, multiLineType, multiPolygonType, tinType, collectionType:
		members := make([]shape, 0, len(s.members))
		for _, member := range s.members {
			simplified, err := simplifyShape(member, tolerance, preserveCollapsed)
			if err != nil {
				return shape{}, err
			}
			if !simplified.isEmpty() {
				members = append(members, simplified)
			}
		}
		s.members = members
		return s, nil
	default:
		return shape{}, pgerror.Newf(pgcode.Internal, "lwgeom_simplify_in_place: unsupported geometry type: %s", typeNameForNumber(s.typeNum))
	}
}

// stSimplify implements ST_Simplify, which returns null when the whole geometry collapses.
func stSimplify(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	preserveCollapsed := len(args) > 2 && args[2].(bool)
	simplified, err := simplifyShape(g.Shape, args[1].(float64), preserveCollapsed)
	if err != nil || simplified.isEmpty() {
		return nil, err
	}
	return g.withShape(simplified), nil
}

// boundary returns the boundary of a shape. A point's boundary is an empty point and a multipoint's is an empty
// multipoint. The boundary of curves is the multipoint of the endpoints that occur an odd number of times, so a
// closed curve has an empty boundary. A polygon or triangle has its rings as its boundary, as a linestring when there
// is only one, and a curve polygon has a multicurve of its rings. A collection's boundary is the multipoint of the
// points in the boundaries of its members.
func boundary(s shape) (shape, error) {
	switch s.typeNum {
	case pointType:
		return shape{typeNum: pointType, ctype: s.ctype}, nil
	case multiPointType:
		return shape{typeNum: multiPointType, ctype: s.ctype}, nil
	case lineType, circStringType, multiLineType, compoundType:
		return curveEndpoints(s), nil
	case polygonType, triangleType, multiPolygonType, tinType:
		var rings [][]float64
		var collect func(member shape)
		collect = func(member shape) {
			if member.typeNum == triangleType {
				if len(member.coords) > 0 {
					rings = append(rings, member.coords)
				}
				return
			}
			rings = append(rings, member.rings...)
			for _, child := range member.members {
				collect(child)
			}
		}
		collect(s)
		if len(rings) == 1 {
			return shape{typeNum: lineType, ctype: s.ctype, coords: rings[0]}, nil
		}
		if len(rings) == 0 && s.typeNum == multiPolygonType {
			return shape{typeNum: collectionType, ctype: s.ctype}, nil
		}
		multi := shape{typeNum: multiLineType, ctype: s.ctype}
		for _, ring := range rings {
			multi.members = append(multi.members, shape{typeNum: lineType, ctype: s.ctype, coords: ring})
		}
		return multi, nil
	case curvePolyType:
		return shape{typeNum: multiCurveType, ctype: s.ctype, members: s.members}, nil
	case collectionType:
		if len(s.members) == 0 {
			return s, nil
		}
		multi := shape{typeNum: multiPointType, ctype: s.ctype}
		for _, member := range s.members {
			memberBoundary, err := boundary(member)
			if err != nil {
				return shape{}, err
			}
			if memberBoundary.typeNum == pointType {
				multi.members = append(multi.members, memberBoundary)
				continue
			}
			for _, point := range memberBoundary.members {
				if point.typeNum == pointType {
					multi.members = append(multi.members, point)
				}
			}
		}
		return multi, nil
	default:
		return shape{}, pgerror.Newf(pgcode.Internal, "lwgeom_boundary: unsupported geometry type: %s", typeNameForNumber(s.typeNum))
	}
}

// curveEndpoints returns the multipoint of the endpoints of a shape's curves that occur an odd number of times, in the
// order they last became odd.
func curveEndpoints(s shape) shape {
	var curves []shape
	if s.typeNum == multiLineType {
		curves = s.members
	} else {
		curves = []shape{s}
	}
	dims := s.ctype.Dimension()
	var points [][]float64
	toggle := func(point []float64) {
		for i, existing := range points {
			if existing[0] == point[0] && existing[1] == point[1] {
				points = append(points[:i], points[i+1:]...)
				return
			}
		}
		points = append(points, point)
	}
	for _, curve := range curves {
		vertices := curveVertices(curve)
		if len(vertices) == 0 {
			continue
		}
		toggle(vertices[:dims])
		toggle(vertices[len(vertices)-dims:])
	}
	multi := shape{typeNum: multiPointType, ctype: s.ctype}
	for _, point := range points {
		multi.members = append(multi.members, shape{typeNum: pointType, ctype: s.ctype, coords: point})
	}
	return multi
}

// stBoundary implements ST_Boundary.
func stBoundary(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	result, err := boundary(g.Shape)
	if err != nil {
		return nil, err
	}
	return g.withShape(result), nil
}

// editorRoutines returns the routines that edit geometries with PostGIS' own rules.
func editorRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_boundary", Symbol: "boundary", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stBoundary},
		{Name: "st_simplify", Symbol: "LWGEOM_simplify2d", Parameters: params("geometry", "float8"), Returns: "geometry", Strict: true, Impl: stSimplify},
		{Name: "st_simplify", Symbol: "LWGEOM_simplify2d_bool", Parameters: params("geometry", "float8", "bool"), Returns: "geometry", Strict: true, Impl: stSimplify},
	}
}
