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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// stArea implements ST_Area and ST_Area2D for geometry, which sum the two-dimensional area of every polygon.
func stArea(ctx *sql.Context, args ...any) (any, error) {
	return shapeArea(args[0].(Geometry).Shape)
}

// shapeArea returns the two-dimensional area of every surface in the shape, linearizing curve polygons first.
func shapeArea(s shape) (float64, error) {
	switch s.typeNum {
	case polygonType:
		area := 0.0
		for i, ring := range s.rings {
			if i == 0 {
				area += math.Abs(signedRingArea(ring, s.ctype))
			} else {
				area -= math.Abs(signedRingArea(ring, s.ctype))
			}
		}
		return area, nil
	case triangleType:
		return math.Abs(signedRingArea(s.coords, s.ctype)), nil
	case curvePolyType:
		stroked, err := strokeShape(s, defaultStroke)
		if err != nil {
			return 0, err
		}
		return shapeArea(stroked)
	case multiPolygonType, multiSurfaceType, collectionType, polyhedralSurfaceType, tinType:
		area := 0.0
		for _, member := range s.members {
			memberArea, err := shapeArea(member)
			if err != nil {
				return 0, err
			}
			area += memberArea
		}
		return area, nil
	default:
		return 0, nil
	}
}

// stLength implements ST_Length and ST_Length2D for geometry, which sum the two-dimensional length of every curve
// and ignore surfaces.
func stLength(ctx *sql.Context, args ...any) (any, error) {
	return curveLength(args[0].(Geometry).Shape), nil
}

// curveLength returns the two-dimensional length of every curve in the shape, measuring circular arcs exactly.
func curveLength(s shape) float64 {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case lineType:
		length := 0.0
		for i := dims; i+dims <= len(s.coords); i += dims {
			dx, dy := s.coords[i]-s.coords[i-dims], s.coords[i+1]-s.coords[i-dims+1]
			length += math.Sqrt(float64(dx*dx) + float64(dy*dy))
		}
		return length
	case circStringType:
		length := 0.0
		for i := 2 * dims; i+dims <= len(s.coords); i += 2 * dims {
			length += arcLength(s.coords[i-2*dims:], s.coords[i-dims:], s.coords[i:])
		}
		return length
	case pointType, polygonType, triangleType, curvePolyType, multiPolygonType, multiSurfaceType, polyhedralSurfaceType, tinType:
		return 0
	default:
		length := 0.0
		for _, member := range s.members {
			length += curveLength(member)
		}
		return length
	}
}

// arcLength returns the length of the circular arc through three points, or the length of the straight path through
// them when they lie on a line.
func arcLength(start []float64, mid []float64, end []float64) float64 {
	centerX, centerY, radius := circumcircle(start, mid, end)
	isCircle := start[0] == end[0] && start[1] == end[1]
	distance := func(a []float64, b []float64) float64 {
		dx, dy := b[0]-a[0], b[1]-a[1]
		return math.Sqrt(float64(dx*dx) + float64(dy*dy))
	}
	if (radius < 0 || pointSide(start, end, mid) == 0) && !isCircle {
		return distance(start, mid) + distance(mid, end)
	}
	if isCircle {
		return math.Pi * distance(start, mid)
	}
	angleStart := math.Atan2(start[1]-centerY, start[0]-centerX)
	angleEnd := math.Atan2(end[1]-centerY, end[0]-centerX)
	total := angleEnd - angleStart
	if pointSide(start, end, mid) < 0 {
		total = angleStart - angleEnd
	}
	if total <= 0 {
		total += 2 * math.Pi
	}
	return radius * total
}

// stPerimeter implements ST_Perimeter and ST_Perimeter2D for geometry, which sum the two-dimensional length of every
// polygon's rings and ignore everything else.
func stPerimeter(ctx *sql.Context, args ...any) (any, error) {
	return perimeter2D(args[0].(Geometry).Shape), nil
}

// perimeter2D returns the two-dimensional length of the rings of every surface in the shape, measuring circular arcs
// exactly.
func perimeter2D(s shape) float64 {
	switch s.typeNum {
	case polygonType:
		perimeter := 0.0
		for _, ring := range s.rings {
			perimeter += curveLength(shape{typeNum: lineType, ctype: s.ctype, coords: ring})
		}
		return perimeter
	case triangleType:
		return curveLength(shape{typeNum: lineType, ctype: s.ctype, coords: s.coords})
	case curvePolyType:
		return curveLength(shape{typeNum: multiCurveType, ctype: s.ctype, members: s.members})
	case multiPolygonType, multiSurfaceType, collectionType, polyhedralSurfaceType, tinType:
		perimeter := 0.0
		for _, member := range s.members {
			perimeter += perimeter2D(member)
		}
		return perimeter
	default:
		return 0
	}
}

// sequenceLength3D returns the length of a sequence of points, measured in three dimensions when it has Z
// coordinates.
func sequenceLength3D(seq []float64, ctype coordinatesType) float64 {
	dims := ctype.Dimension()
	length := 0.0
	for i := dims; i+dims <= len(seq); i += dims {
		dx, dy, dz := seq[i]-seq[i-dims], seq[i+1]-seq[i-dims+1], 0.0
		if ctype.Is3D() {
			dz = seq[i+2] - seq[i-dims+2]
		}
		length += math.Sqrt(float64(dx*dx) + float64(dy*dy) + float64(dz*dz))
	}
	return length
}

// stLength3D implements ST_3DLength, which sums the length of every curve in three dimensions. Circular strings and
// compound curves are measured in two dimensions, as PostGIS does.
func stLength3D(ctx *sql.Context, args ...any) (any, error) {
	var length3D func(s shape) float64
	length3D = func(s shape) float64 {
		switch s.typeNum {
		case lineType:
			return sequenceLength3D(s.coords, s.ctype)
		case circStringType, compoundType:
			return curveLength(s)
		case multiLineType, multiCurveType, collectionType:
			length := 0.0
			for _, member := range s.members {
				length += length3D(member)
			}
			return length
		default:
			return 0
		}
	}
	return length3D(args[0].(Geometry).Shape), nil
}

// stPerimeter3D implements ST_3DPerimeter, which sums the length of every surface's rings in three dimensions. Rings
// made of circular strings or compound curves are measured in two dimensions, as PostGIS does.
func stPerimeter3D(ctx *sql.Context, args ...any) (any, error) {
	var perimeter3D func(s shape) float64
	perimeter3D = func(s shape) float64 {
		perimeter := 0.0
		switch s.typeNum {
		case polygonType:
			for _, ring := range s.rings {
				perimeter += sequenceLength3D(ring, s.ctype)
			}
		case triangleType:
			perimeter = sequenceLength3D(s.coords, s.ctype)
		case curvePolyType:
			for _, ring := range s.members {
				if ring.typeNum == lineType {
					perimeter += sequenceLength3D(ring.coords, ring.ctype)
				} else {
					perimeter += curveLength(ring)
				}
			}
		case multiPolygonType, multiSurfaceType, polyhedralSurfaceType, tinType, collectionType:
			for _, member := range s.members {
				perimeter += perimeter3D(member)
			}
		}
		return perimeter
	}
	return perimeter3D(args[0].(Geometry).Shape), nil
}

// geosPair converts two geometries the way PostGIS hands them to GEOS, after checking that their SRIDs match. The
// function name is the one PostGIS reports in its error.
func geosPair(funcName string, a Geometry, b Geometry) (spatial.Geometry, spatial.Geometry, error) {
	if err := checkSameSRID(funcName, a, b); err != nil {
		return spatial.Geometry{}, spatial.Geometry{}, err
	}
	geosA, err := a.Shape.toGEOS()
	if err != nil {
		return spatial.Geometry{}, spatial.Geometry{}, err
	}
	geosB, err := b.Shape.toGEOS()
	if err != nil {
		return spatial.Geometry{}, spatial.Geometry{}, err
	}
	return geosA, geosB, nil
}

// distance2D returns the minimum two-dimensional distance between two geometries, or false when either is empty.
func distance2D(funcName string, a Geometry, b Geometry) (float64, bool, error) {
	geosA, geosB, err := geosPair(funcName, a, b)
	if err != nil || a.Shape.isEmpty() || b.Shape.isEmpty() {
		return 0, false, err
	}
	g := spatial.NewGEOS()
	defer g.Close()
	distance, err := g.Distance(geosA, geosB)
	if err != nil {
		return 0, false, geosError("GEOSDistance", err)
	}
	return distance, true, nil
}

// stDistance implements ST_Distance for geometry, which returns the minimum two-dimensional distance, or null when
// either geometry is empty.
func stDistance(ctx *sql.Context, args ...any) (any, error) {
	distance, ok, err := distance2D("ST_Distance", args[0].(Geometry), args[1].(Geometry))
	if err != nil || !ok {
		return nil, err
	}
	return distance, nil
}

// stDWithin implements ST_DWithin for geometry, which returns whether the geometries lie within the given distance
// of each other. Empty geometries are never within any distance.
func stDWithin(ctx *sql.Context, args ...any) (any, error) {
	tolerance := args[2].(float64)
	distance, ok, err := distance2D("LWGEOM_dwithin", args[0].(Geometry), args[1].(Geometry))
	if err != nil {
		return nil, err
	}
	if tolerance < 0 {
		return nil, pgerror.New(pgcode.Internal, "Tolerance cannot be less than zero\n")
	}
	return ok && distance <= tolerance, nil
}

// textArgsAsGeometry wraps an implementation so that its first `count` arguments may be given as text, which is parsed
// the way the geometry type parses its input. These overloads are not strict, so a null argument returns null.
func textArgsAsGeometry(count int, impl extdef.Function) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		for i := 0; i < count; i++ {
			if args[i] == nil {
				return nil, nil
			}
			g, err := parseGeometryInput(ctx, args[i].(string))
			if err != nil {
				return nil, err
			}
			args[i] = g
		}
		return impl(ctx, args...)
	}
}

// measureRoutines returns the routines that measure geometries.
func measureRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	return []extdef.Routine{
		{Name: "st_3dlength", Symbol: "LWGEOM_length_linestring", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stLength3D},
		{Name: "st_3dperimeter", Symbol: "LWGEOM_perimeter_poly", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stPerimeter3D},
		{Name: "_st_dwithin", Symbol: "_st_dwithin_geometry_geometry_float8", Parameters: append(geomParams, extdef.Parameter{Type: "float8"}), Returns: "bool", Strict: true, Impl: stDWithin},
		{Name: "geometry_distance_centroid", Symbol: "geometry_distance_centroid_geometry_geometry", Parameters: geomParams, Returns: "float8", Strict: true, Impl: stDistance},
		{Name: "st_area", Symbol: "ST_Area", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stArea},
		{Name: "st_area", Symbol: "st_area_text", Parameters: params("text"), Returns: "float8", Strict: true, Impl: textArgsAsGeometry(1, stArea)},
		{Name: "st_area2d", Symbol: "st_area2d_geometry", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stArea},
		{Name: "st_distance", Symbol: "ST_Distance", Parameters: geomParams, Returns: "float8", Strict: true, Impl: stDistance},
		{Name: "st_distance", Symbol: "st_distance_text_text", Parameters: params("text", "text"), Returns: "float8", Strict: true, Impl: textArgsAsGeometry(2, stDistance)},
		{Name: "st_dwithin", Symbol: "LWGEOM_dwithin", Parameters: append(geomParams, extdef.Parameter{Type: "float8"}), Returns: "bool", Strict: true, Impl: stDWithin},
		{Name: "st_dwithin", Symbol: "st_dwithin_text_text_float8", Parameters: params("text", "text", "float8"), Returns: "bool", Impl: textArgsAsGeometry(2, stDWithin)},
		{Name: "st_length", Symbol: "LWGEOM_length2d_linestring", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stLength},
		{Name: "st_length", Symbol: "st_length_text", Parameters: params("text"), Returns: "float8", Strict: true, Impl: textArgsAsGeometry(1, stLength)},
		{Name: "st_length2d", Symbol: "st_length2d_geometry", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stLength},
		{Name: "st_perimeter", Symbol: "LWGEOM_perimeter2d_poly", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stPerimeter},
		{Name: "st_perimeter2d", Symbol: "st_perimeter2d_geometry", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stPerimeter},
	}
}
