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
)

// The bits of the options argument of ST_AsX3D.
const (
	x3dFlipXY         = 1
	x3dGeoCoordinates = 2
)

// x3dGeoSystemPrefix is the start of a geographic coordinate node, which its axis order completes.
const x3dGeoSystemPrefix = `<GeoCoordinate geoSystem='"GD" "WE" `

// x3dWriter writes geometries as X3D 3 nodes, matching the spacing that PostGIS writes. The identifier, when there is
// one, is written into every node's opening tag followed by a colon.
type x3dWriter struct {
	sb        strings.Builder
	precision int
	options   int32
	id        string
}

// writeX3D returns the X3D of a geometry, which is empty for an empty geometry.
func writeX3D(g Geometry, version int32, precision int, options int32, id string) (any, error) {
	if version != 3 {
		return nil, pgerror.New(pgcode.Internal, "Only X3D version 3 are supported")
	}
	if options&x3dGeoCoordinates != 0 && g.SRID != sridDefault {
		return nil, pgerror.New(pgcode.Internal, "Only SRID 4326 is supported for geocoordinates.")
	}
	if id != "" {
		id += ":"
	}
	w := &x3dWriter{precision: precision, options: options, id: id}
	if g.Shape.isEmpty() {
		return "", nil
	}
	if err := w.write(g.Shape); err != nil {
		return nil, err
	}
	return w.sb.String(), nil
}

// write writes a shape as an X3D node. Points and triangles are written as bare coordinates.
func (w *x3dWriter) write(s shape) error {
	switch s.typeNum {
	case pointType, triangleType:
		w.sb.WriteString(w.points(s.coords, s.ctype, s.typeNum == triangleType))
	case lineType:
		w.sb.WriteString(fmt.Sprintf("<LineSet %s vertexCount='%d'>", w.id, len(s.coords)/s.ctype.Dimension()))
		w.sb.WriteString(w.coordinateStart() + w.points(s.coords, s.ctype, false) + "' /></LineSet>")
	case polygonType, multiPolygonType:
		polygons := []shape{s}
		if s.typeNum == multiPolygonType {
			polygons = s.members
		}
		var indexes, points []string
		next := 0
		for _, polygon := range polygons {
			var ringIndexes []string
			for _, ring := range polygon.rings {
				count := len(ring)/s.ctype.Dimension() - 1
				ringIndexes = append(ringIndexes, x3dIndexes(next, count))
				next += count
			}
			indexes = append(indexes, strings.Join(ringIndexes, " -1 "))
			points = append(points, w.rings(polygon.rings, s.ctype)+" ")
		}
		w.indexedSet(fmt.Sprintf("<IndexedFaceSet %s convex='false' coordIndex='%s'>", w.id, strings.Join(indexes, " -1 ")), "IndexedFaceSet", strings.Join(points, ""), s.ctype)
	case multiPointType:
		var points strings.Builder
		for _, member := range s.members {
			points.WriteString(w.points(member.coords, member.ctype, false) + " ")
		}
		if s.hasZ() {
			w.sb.WriteString("<PointSet " + w.id + ">" + w.coordinateStart() + points.String() + "' /></PointSet>")
		} else {
			w.sb.WriteString("<Polypoint2D " + w.id + " point='" + points.String() + "' />")
		}
	case multiLineType:
		var indexes []string
		var points strings.Builder
		next := 0
		for _, line := range s.members {
			count := len(line.coords) / line.ctype.Dimension()
			indexes = append(indexes, x3dIndexes(next, count))
			next += count
			points.WriteString(w.points(line.coords, line.ctype, false) + " ")
		}
		w.indexedSet(fmt.Sprintf("<IndexedLineSet %s coordIndex='%s'>", w.id, strings.Join(indexes, " -1 ")), "IndexedLineSet", points.String(), s.ctype)
	case tinType:
		var points []string
		for _, triangle := range s.members {
			points = append(points, w.points(triangle.coords, triangle.ctype, true))
		}
		w.sb.WriteString(fmt.Sprintf("<IndexedTriangleSet %s index='%s'>", w.id, x3dIndexes(0, 3*len(s.members))))
		w.sb.WriteString(w.coordinateStart() + strings.Join(points, " ") + "'/></IndexedTriangleSet>")
	case polyhedralSurfaceType:
		var indexes, points []string
		next := 0
		for _, patch := range s.members {
			count := len(patch.rings[0])/patch.ctype.Dimension() - 1
			indexes = append(indexes, x3dIndexes(next, count))
			next += count
			points = append(points, w.rings(patch.rings, patch.ctype))
		}
		w.sb.WriteString(fmt.Sprintf("<IndexedFaceSet convex='false' %s coordIndex='%s'>", w.id, strings.Join(indexes, " -1 ")))
		w.sb.WriteString(w.coordinateStart() + strings.Join(points, " ") + "' /></IndexedFaceSet>")
	case collectionType:
		for _, member := range s.members {
			w.sb.WriteString("<Shape" + w.id + ">")
			switch member.typeNum {
			case polygonType:
				w.sb.WriteString(w.rings(member.rings, member.ctype))
			case pointType, lineType, multiPointType, multiLineType, multiPolygonType, polyhedralSurfaceType, tinType, collectionType:
				if err := w.write(member); err != nil {
					return err
				}
			default:
				return pgerror.New(pgcode.Internal, "asx3d3_collection_buf: unknown geometry type")
			}
			w.sb.WriteString("</Shape>")
		}
	default:
		return pgerror.Newf(pgcode.Internal, "lwgeom_to_x3d3: '%s' geometry type not supported", typeNameForNumber(s.typeNum))
	}
	return nil
}

// indexedSet writes an indexed face or line set from its opening tag and points. Only sets with Z coordinates wrap
// their points in a coordinate node and close the set, as PostGIS does.
func (w *x3dWriter) indexedSet(open string, name string, points string, ctype coordinatesType) {
	w.sb.WriteString(open)
	if ctype.Is3D() {
		w.sb.WriteString(w.coordinateStart() + points + "' /></" + name + ">")
	} else {
		w.sb.WriteString(points + "' />")
	}
}

// coordinateStart returns the start of the node that holds a set's points, which is a geographic coordinate node when
// the options ask for one.
func (w *x3dWriter) coordinateStart() string {
	if w.options&x3dGeoCoordinates == 0 {
		return "<Coordinate point='"
	}
	if w.options&x3dFlipXY != 0 {
		return x3dGeoSystemPrefix + `"latitude_first"' point='`
	}
	return x3dGeoSystemPrefix + `"longitude_first"' point='`
}

// rings returns the points of every ring without their closing points, separated by spaces.
func (w *x3dWriter) rings(rings [][]float64, ctype coordinatesType) string {
	points := make([]string, len(rings))
	for i, ring := range rings {
		points[i] = w.points(ring, ctype, true)
	}
	return strings.Join(points, " ")
}

// points returns the ordinates of a sequence separated by spaces, leaving out M, and the last point when `dropLast`
// is true. X and Y are swapped when the options ask for it.
func (w *x3dWriter) points(seq []float64, ctype coordinatesType, dropLast bool) string {
	dims := ctype.Dimension()
	end := len(seq)
	if dropLast {
		end -= dims
	}
	var ordinates []string
	for i := 0; i+dims <= end; i += dims {
		x, y := seq[i], seq[i+1]
		if w.options&x3dFlipXY != 0 {
			x, y = y, x
		}
		ordinates = append(ordinates, formatOrdinate(x, w.precision), formatOrdinate(y, w.precision))
		if ctype.Is3D() {
			ordinates = append(ordinates, formatOrdinate(seq[i+2], w.precision))
		}
	}
	return strings.Join(ordinates, " ")
}

// x3dIndexes returns `count` consecutive indexes starting at `start`, separated by spaces.
func x3dIndexes(start int, count int) string {
	indexes := make([]string, count)
	for i := range indexes {
		indexes[i] = fmt.Sprint(start + i)
	}
	return strings.Join(indexes, " ")
}

// x3dImpl implements ST_AsX3D and _ST_AsX3D. The arguments start with the X3D version when `hasVersion` is true, and
// the version is 3 otherwise. Null arguments after the geometry take their defaults.
func x3dImpl(hasVersion bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		version := int32(3)
		if hasVersion {
			version = 0
			if args[0] != nil {
				version = args[0].(int32)
			}
			args = args[1:]
		}
		if args[0] == nil {
			return nil, nil
		}
		precision, options, id := 15, int32(0), ""
		if args[1] != nil {
			precision = int(args[1].(int32))
		}
		if args[2] != nil {
			options = args[2].(int32)
		}
		if len(args) > 3 && args[3] != nil {
			id = args[3].(string)
		}
		return writeX3D(args[0].(Geometry), version, precision, options, id)
	}
}

// x3dRoutines returns the routines that write X3D.
func x3dRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "_st_asx3d", Symbol: "LWGEOM_asX3D", Parameters: params("int4", "geometry", "int4", "int4", "text"), Returns: "text", Impl: x3dImpl(true)},
		{Name: "st_asx3d", Symbol: "st_asx3d_geometry_int4_int4", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "maxdecimaldigits", Type: "int4", Default: `15`}, {Name: "options", Type: "int4", Default: `0`}}, Returns: "text", Impl: x3dImpl(false)},
	}
}
