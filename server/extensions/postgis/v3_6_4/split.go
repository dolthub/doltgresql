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
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// splitLineAtPoint returns the two pieces of a line split at a point that lies on it, or false when the point lies
// off the line or at one of its ends. The point splits the first segment that holds it, which is the first segment
// whose two-dimensional cross product with the point is zero, and takes the Z and M coordinates interpolated along
// that segment.
func splitLineAtPoint(line shape, point []float64) (shape, shape, bool) {
	dims := line.ctype.Dimension()
	count := len(line.coords) / dims
	if count < 2 {
		return shape{}, shape{}, false
	}
	segment, fraction := -1, 0.0
	for i := 0; i+1 < count; i++ {
		a, b := line.coords[i*dims:], line.coords[(i+1)*dims:]
		dx, dy := b[0]-a[0], b[1]-a[1]
		px, py := point[0]-a[0], point[1]-a[1]
		lengthSquared, along := float64(dx*dx)+float64(dy*dy), float64(px*dx)+float64(py*dy)
		if lengthSquared == 0 || along <= 0 || along >= lengthSquared {
			if point[0] == a[0] && point[1] == a[1] {
				segment, fraction = i, 0
				break
			}
			if point[0] == b[0] && point[1] == b[1] {
				segment, fraction = i, 1
				break
			}
			continue
		}
		if float64(px*dy)-float64(py*dx) == 0 {
			segment, fraction = i, along/lengthSquared
			break
		}
	}
	first, last := line.coords[:dims], line.coords[(count-1)*dims:]
	if segment < 0 || point[0] == first[0] && point[1] == first[1] || point[0] == last[0] && point[1] == last[1] {
		return shape{}, shape{}, false
	}
	a, b := line.coords[segment*dims:(segment+1)*dims], line.coords[(segment+1)*dims:(segment+2)*dims]
	at := []float64{point[0], point[1]}
	for j := 2; j < dims; j++ {
		at = append(at, a[j]+float64(fraction*(b[j]-a[j])))
	}
	sameXY := func(p []float64) bool {
		return p[0] == at[0] && p[1] == at[1]
	}
	head := append([]float64(nil), line.coords[:(segment+1)*dims]...)
	if !sameXY(a) {
		head = append(head, at...)
	}
	tail := append([]float64(nil), at...)
	rest := line.coords[(segment+1)*dims:]
	if sameXY(b) {
		rest = rest[dims:]
	}
	tail = append(tail, rest...)
	return shape{typeNum: lineType, ctype: line.ctype, coords: head}, shape{typeNum: lineType, ctype: line.ctype, coords: tail}, true
}

// splitLineByPoints splits a line at each point of a multipoint in turn. Each split replaces the piece that held the
// point with the piece after the point, and appends the piece before it.
func splitLineByPoints(line shape, points []shape) []shape {
	pieces := []shape{line}
	for _, point := range points {
		if len(point.coords) < 2 {
			continue
		}
		for j, piece := range pieces {
			if before, after, ok := splitLineAtPoint(piece, point.coords); ok {
				pieces[j] = after
				pieces = append(pieces, before)
				break
			}
		}
	}
	return pieces
}

// splitLineByLines splits a line wherever it crosses the lines of the blade, which may not overlap the line.
func splitLineByLines(geos *spatial.GEOS, line shape, blade spatial.Geometry) ([]shape, error) {
	geosLine := line.linearToGEOS()
	crossing, err := geos.Intersection(geosLine, blade, -1)
	if err != nil {
		return nil, geosError("GEOSIntersection", err)
	}
	if shapeFromGEOS(crossing, false).dimension() > 0 {
		return nil, pgerror.New(pgcode.Internal, "Splitter line has linear intersection with input")
	}
	difference, err := geos.Difference(geosLine, blade, -1)
	if err != nil {
		return nil, geosError("GEOSDifference", err)
	}
	var pieces []shape
	atomicMembers(shapeFromGEOS(difference, line.hasZ()), false, func(member shape) {
		pieces = append(pieces, member)
	})
	return pieces, nil
}

// splitPolygonByLines splits a polygon along the lines of the blade, returning the polygons that the blade and the
// polygon's boundary enclose within the polygon.
func splitPolygonByLines(geos *spatial.GEOS, polygon shape, blade spatial.Geometry) ([]shape, error) {
	geosPolygon := polygon.linearToGEOS()
	boundary, err := geos.Boundary(geosPolygon)
	if err != nil {
		return nil, geosError("GEOSBoundary", err)
	}
	noded, err := geos.Union(boundary, blade, -1)
	if err != nil {
		return nil, geosError("GEOSUnion", err)
	}
	areas, err := geos.Polygonize([]spatial.Geometry{noded})
	if err != nil {
		return nil, geosError("GEOSPolygonize", err)
	}
	var pieces []shape
	for _, area := range areas.Members {
		inside, err := geos.PointOnSurface(area)
		if err != nil {
			return nil, geosError("GEOSPointOnSurface", err)
		}
		covered, err := geos.Covers(geosPolygon, inside)
		if err != nil {
			return nil, geosError("GEOSCovers", err)
		}
		if covered {
			pieces = append(pieces, shapeFromGEOS(area, polygon.hasZ()))
		}
	}
	return pieces, nil
}

// splitShape returns the pieces of a shape split by a blade, splitting the members of multi-part shapes one at a
// time.
func splitShape(geos *spatial.GEOS, s shape, blade shape) ([]shape, error) {
	unsupported := func(kind string) error {
		return pgerror.Newf(pgcode.Internal, "Splitting a %s by a %s is unsupported", kind, typeNameForNumber(blade.typeNum))
	}
	switch s.typeNum {
	case lineType:
		switch blade.typeNum {
		case pointType:
			if before, after, ok := splitLineAtPoint(s, blade.coords); ok {
				return []shape{before, after}, nil
			}
			return []shape{s}, nil
		case multiPointType:
			return splitLineByPoints(s, blade.members), nil
		case lineType, multiLineType, polygonType, multiPolygonType:
			geosBlade := blade.linearToGEOS()
			if blade.dimension() == 2 {
				var err error
				if geosBlade, err = geos.Boundary(geosBlade); err != nil {
					return nil, geosError("GEOSBoundary", err)
				}
			}
			return splitLineByLines(geos, s, geosBlade)
		default:
			return nil, unsupported("Line")
		}
	case polygonType:
		switch blade.typeNum {
		case lineType, multiLineType:
			return splitPolygonByLines(geos, s, blade.linearToGEOS())
		default:
			return nil, unsupported("Polygon")
		}
	case multiLineType, multiPolygonType, collectionType:
		var pieces []shape
		for _, member := range s.members {
			memberPieces, err := splitShape(geos, member, blade)
			if err != nil {
				return nil, err
			}
			pieces = append(pieces, memberPieces...)
		}
		return pieces, nil
	default:
		return nil, pgerror.Newf(pgcode.Internal, "Splitting of %s geometries is unsupported", typeNameForNumber(s.typeNum))
	}
}

// stSplit implements ST_Split, which returns the pieces of a line or polygon split by a blade as a collection.
func stSplit(ctx *sql.Context, args ...any) (any, error) {
	g, blade := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("ST_Split", g, blade); err != nil {
		return nil, err
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	pieces, err := splitShape(geos, g.Shape, blade.Shape)
	if err != nil {
		return nil, err
	}
	return g.withShape(shape{typeNum: collectionType, ctype: g.Shape.ctype, members: pieces}), nil
}

// splitRoutines returns the routines that split geometries.
func splitRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_split", Symbol: "ST_Split", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "geometry", Strict: true, Impl: stSplit},
	}
}
