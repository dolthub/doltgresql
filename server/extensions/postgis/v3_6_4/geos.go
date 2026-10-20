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
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// toGEOS converts a shape to a GEOS geometry the way PostGIS hands geometries to GEOS: curves are first linearized
// with the default tolerance, M coordinates are dropped, triangles become polygons, TINs become collections of
// polygons, and polyhedral surfaces are rejected.
func (s shape) toGEOS() (spatial.Geometry, error) {
	if s.typeNum == polyhedralSurfaceType {
		return spatial.Geometry{}, pgerror.Newf(pgcode.Internal, "Unknown geometry type: %d - %s", s.typeNum, typeNameForNumber(s.typeNum))
	}
	stroked, err := strokeShape(s, defaultStroke)
	if err != nil {
		return spatial.Geometry{}, err
	}
	return stroked.linearToGEOS(), nil
}

// linearToGEOS converts a shape without curves to a GEOS geometry, as described in toGEOS.
func (s shape) linearToGEOS() spatial.Geometry {
	result := spatial.Geometry{HasZ: s.hasZ()}
	switch s.typeNum {
	case pointType, lineType:
		result.Type = spatial.Point
		if s.typeNum == lineType {
			result.Type = spatial.LineString
		}
		result.Coords = s.geosCoords(s.coords)
	case polygonType:
		result.Type = spatial.Polygon
		for _, ring := range s.rings {
			result.Members = append(result.Members, spatial.Geometry{Type: spatial.LinearRing, HasZ: result.HasZ, Coords: s.geosCoords(ring)})
		}
	case triangleType:
		result.Type = spatial.Polygon
		if len(s.coords) > 0 {
			result.Members = []spatial.Geometry{{Type: spatial.LinearRing, HasZ: result.HasZ, Coords: s.geosCoords(s.coords)}}
		}
	default:
		switch s.typeNum {
		case multiPointType:
			result.Type = spatial.MultiPoint
		case multiLineType:
			result.Type = spatial.MultiLineString
		case multiPolygonType:
			result.Type = spatial.MultiPolygon
		default:
			result.Type = spatial.GeometryCollection
		}
		result.Members = make([]spatial.Geometry, len(s.members))
		for i, member := range s.members {
			result.Members[i] = member.linearToGEOS()
		}
	}
	return result
}

// geosCoords returns the ordinates of a sequence in the shape without any M coordinates.
func (s shape) geosCoords(seq []float64) []float64 {
	if !s.hasM() {
		return seq
	}
	dims := s.ctype.Dimension()
	kept := dims - 1
	coords := make([]float64, 0, len(seq)/dims*kept)
	for i := 0; i+dims <= len(seq); i += dims {
		coords = append(coords, seq[i:i+kept]...)
	}
	return coords
}

// fromGEOS converts a GEOS result to a geometry with the given SRID, the way PostGIS reads results back from GEOS:
// Z coordinates are kept only when `keepZ` is true, and linear rings become linestrings.
func fromGEOS(result spatial.Geometry, srid int32, keepZ bool) Geometry {
	return Geometry{SRID: srid, Shape: shapeFromGEOS(result, keepZ && result.HasZ)}
}

// shapeFromGEOS converts a GEOS geometry to a shape, keeping Z coordinates when `hasZ` is true.
func shapeFromGEOS(g spatial.Geometry, hasZ bool) shape {
	ctype := dimXY
	if hasZ {
		ctype = dimXYZ
	}
	coords := func(source spatial.Geometry) []float64 {
		if hasZ || !source.HasZ {
			return source.Coords
		}
		flat := make([]float64, 0, len(source.Coords)/3*2)
		for i := 0; i+3 <= len(source.Coords); i += 3 {
			flat = append(flat, source.Coords[i], source.Coords[i+1])
		}
		return flat
	}
	result := shape{ctype: ctype}
	switch g.Type {
	case spatial.Point:
		result.typeNum, result.coords = pointType, coords(g)
	case spatial.LineString, spatial.LinearRing:
		result.typeNum, result.coords = lineType, coords(g)
	case spatial.Polygon:
		result.typeNum = polygonType
		for _, ring := range g.Members {
			result.rings = append(result.rings, coords(ring))
		}
	default:
		switch g.Type {
		case spatial.MultiPoint:
			result.typeNum = multiPointType
		case spatial.MultiLineString:
			result.typeNum = multiLineType
		case spatial.MultiPolygon:
			result.typeNum = multiPolygonType
		default:
			result.typeNum = collectionType
		}
		result.members = make([]shape, len(g.Members))
		for i, member := range g.Members {
			result.members[i] = shapeFromGEOS(member, hasZ)
		}
	}
	return result
}

// geosError returns the error that PostGIS reports when the named GEOS function fails.
func geosError(function string, err error) error {
	return pgerror.Newf(pgcode.Internal, "%s: %s", function, err.Error())
}
