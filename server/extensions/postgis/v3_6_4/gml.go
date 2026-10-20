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
	"regexp"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// The bits of the options argument of ST_AsGML.
const (
	gmlLongSRS     = 1
	gmlNoDimension = 2
	gmlShortLine   = 4
	gmlLatLon      = 16
	gmlEnvelope    = 32
)

// gmlWriter writes geometries as GML 2 or GML 3, prefixing every tag with the namespace prefix. GML 3 geometry
// elements also carry the identifier when there is one.
type gmlWriter struct {
	sb        strings.Builder
	version   int32
	precision int
	options   int32
	prefix    string
	id        string
}

// writeGML returns the GML of a geometry, which is null for an empty geometry unless the options ask for its
// envelope.
func writeGML(g Geometry, version int32, precision int, options int32, prefix string, id string) (any, error) {
	if version != 2 && version != 3 {
		return nil, pgerror.New(pgcode.Internal, "Only GML 2 and GML 3 are supported")
	}
	if prefix != "" {
		prefix += ":"
	}
	w := &gmlWriter{version: version, precision: precision, options: options, prefix: prefix, id: id}
	srsName := ""
	if g.SRID != sridUnknown {
		srsName = fmt.Sprintf("EPSG:%d", g.SRID)
		if options&gmlLongSRS != 0 {
			srsName = fmt.Sprintf("urn:ogc:def:crs:EPSG::%d", g.SRID)
		}
	}
	if options&gmlEnvelope != 0 {
		w.writeEnvelope(g.Shape, srsName)
		return w.sb.String(), nil
	}
	if g.Shape.isEmpty() {
		return nil, nil
	}
	var err error
	if version == 2 {
		err = w.write2(g.Shape, srsName, true)
	} else {
		err = w.write3(g.Shape, srsName)
	}
	if err != nil {
		return nil, err
	}
	return w.sb.String(), nil
}

// tag writes an opening or closing tag with the namespace prefix.
func (w *gmlWriter) tag(name string, closing bool) {
	if closing {
		w.sb.WriteString("</" + w.prefix + name + ">")
	} else {
		w.sb.WriteString("<" + w.prefix + name + ">")
	}
}

// startGeometry writes the opening tag of a geometry element, with the SRS name when one is given and, in GML 3, the
// identifier. The element of an empty geometry is closed at once, in which case startGeometry returns false.
func (w *gmlWriter) startGeometry(name string, srsName string, isEmpty bool) bool {
	w.sb.WriteString("<" + w.prefix + name)
	if srsName != "" {
		w.sb.WriteString(` srsName="` + srsName + `"`)
	}
	if w.version == 3 && w.id != "" {
		w.sb.WriteString(" " + w.prefix + `id="` + w.id + `"`)
	}
	if isEmpty {
		w.sb.WriteString("/>")
		return false
	}
	w.sb.WriteString(">")
	return true
}

// write2 writes a shape as a GML 2 element. Collections skip the members that GML 2 cannot represent, while a
// top-level shape of such a type is an error.
func (w *gmlWriter) write2(s shape, srsName string, top bool) error {
	switch s.typeNum {
	case pointType, lineType:
		name := "Point"
		if s.typeNum == lineType {
			name = "LineString"
		}
		if w.startGeometry(name, srsName, s.isEmpty()) {
			w.coordinates2(s.coords, s.ctype)
			w.tag(name, true)
		}
	case polygonType:
		if w.startGeometry("Polygon", srsName, s.isEmpty()) {
			for i, ring := range s.rings {
				boundary := "innerBoundaryIs"
				if i == 0 {
					boundary = "outerBoundaryIs"
				}
				w.tag(boundary, false)
				w.tag("LinearRing", false)
				w.coordinates2(ring, s.ctype)
				w.tag("LinearRing", true)
				w.tag(boundary, true)
			}
			w.tag("Polygon", true)
		}
	case multiPointType, multiLineType, multiPolygonType, collectionType:
		names := map[int][2]string{
			multiPointType:   {"MultiPoint", "pointMember"},
			multiLineType:    {"MultiLineString", "lineStringMember"},
			multiPolygonType: {"MultiPolygon", "polygonMember"},
			collectionType:   {"MultiGeometry", "geometryMember"},
		}[s.typeNum]
		name, memberName := names[0], names[1]
		if w.startGeometry(name, srsName, s.isEmpty()) {
			for _, member := range s.members {
				w.tag(memberName, false)
				if err := w.write2(member, "", false); err != nil {
					return err
				}
				w.tag(memberName, true)
			}
			w.tag(name, true)
		}
	case triangleType, tinType, polyhedralSurfaceType:
		if top {
			return pgerror.Newf(pgcode.Internal, "Cannot convert %s to GML2. Try ST_AsGML(3, <geometry>) to generate GML3.", typeNameForNumber(s.typeNum))
		}
	default:
		if top {
			return pgerror.Newf(pgcode.Internal, "lwgeom_to_gml2: '%s' geometry type not supported", typeNameForNumber(s.typeNum))
		}
	}
	return nil
}

// coordinates2 writes a GML 2 coordinates element, with X, Y, and Z when present separated by commas, and points
// separated by spaces.
func (w *gmlWriter) coordinates2(seq []float64, ctype coordinatesType) {
	w.tag("coordinates", false)
	dims := ctype.Dimension()
	for i := 0; i+dims <= len(seq); i += dims {
		if i > 0 {
			w.sb.WriteByte(' ')
		}
		w.sb.WriteString(formatOrdinate(seq[i], w.precision) + "," + formatOrdinate(seq[i+1], w.precision))
		if ctype.Is3D() {
			w.sb.WriteString("," + formatOrdinate(seq[i+2], w.precision))
		}
	}
	w.tag("coordinates", true)
}

// write3 writes a shape as a GML 3 element. Lines are curves made of segments unless the options ask for plain
// linestrings.
func (w *gmlWriter) write3(s shape, srsName string) error {
	switch s.typeNum {
	case pointType:
		if w.startGeometry("Point", srsName, s.isEmpty()) {
			w.positions3("pos", s.coords, s.ctype)
			w.tag("Point", true)
		}
	case lineType, circStringType, compoundType:
		if s.typeNum == lineType && w.options&gmlShortLine != 0 {
			if w.startGeometry("LineString", srsName, s.isEmpty()) {
				w.positions3("posList", s.coords, s.ctype)
				w.tag("LineString", true)
			}
			return nil
		}
		if w.startGeometry("Curve", srsName, s.isEmpty()) {
			w.tag("segments", false)
			segments := []shape{s}
			if s.typeNum == compoundType {
				segments = s.members
			}
			for _, segment := range segments {
				name := "LineStringSegment"
				if segment.typeNum == circStringType {
					name = "ArcString"
				}
				w.tag(name, false)
				w.positions3("posList", segment.coords, segment.ctype)
				w.tag(name, true)
			}
			w.tag("segments", true)
			w.tag("Curve", true)
		}
	case polygonType, triangleType:
		name, rings := "Polygon", s.rings
		if s.typeNum == triangleType {
			name, rings = "Triangle", [][]float64{s.coords}
		}
		if w.startGeometry(name, srsName, s.isEmpty()) {
			w.rings3(rings, s.ctype)
			w.tag(name, true)
		}
	case curvePolyType:
		if w.startGeometry("Polygon", srsName, s.isEmpty()) {
			for i, ring := range s.members {
				boundary := "interior"
				if i == 0 {
					boundary = "exterior"
				}
				w.tag(boundary, false)
				if ring.typeNum == lineType {
					w.tag("LinearRing", false)
					w.positions3("posList", ring.coords, ring.ctype)
					w.tag("LinearRing", true)
				} else {
					w.tag("Ring", false)
					w.tag("curveMember", false)
					if err := w.write3(ring, ""); err != nil {
						return err
					}
					w.tag("curveMember", true)
					w.tag("Ring", true)
				}
				w.tag(boundary, true)
			}
			w.tag("Polygon", true)
		}
	case polyhedralSurfaceType, tinType:
		name, patches, patchName := "PolyhedralSurface", "polygonPatches", "PolygonPatch"
		if s.typeNum == tinType {
			name, patches, patchName = "Tin", "trianglePatches", "Triangle"
		}
		if w.startGeometry(name, srsName, s.isEmpty()) {
			w.tag(patches, false)
			for _, patch := range s.members {
				rings := patch.rings
				if patch.typeNum == triangleType {
					rings = [][]float64{patch.coords}
				}
				if w.startGeometry(patchName, "", patch.isEmpty()) {
					w.rings3(rings, patch.ctype)
					w.tag(patchName, true)
				}
			}
			w.tag(patches, true)
			w.tag(name, true)
		}
	default:
		names := map[int][2]string{
			multiPointType:   {"MultiPoint", "pointMember"},
			multiLineType:    {"MultiCurve", "curveMember"},
			multiPolygonType: {"MultiSurface", "surfaceMember"},
			multiCurveType:   {"MultiCurve", "curveMember"},
			multiSurfaceType: {"MultiSurface", ""},
			collectionType:   {"MultiGeometry", "geometryMember"},
		}[s.typeNum]
		name, memberName := names[0], names[1]
		if w.startGeometry(name, srsName, s.isEmpty()) {
			for _, member := range s.members {
				if s.typeNum == collectionType {
					switch member.typeNum {
					case pointType, lineType, polygonType, multiPointType, multiLineType, multiPolygonType, collectionType:
					default:
						return pgerror.New(pgcode.Internal, "asgml3_collection: unknown geometry type")
					}
				}
				if memberName != "" {
					w.tag(memberName, false)
				}
				if err := w.write3(member, ""); err != nil {
					return err
				}
				if memberName != "" {
					w.tag(memberName, true)
				}
			}
			w.tag(name, true)
		}
	}
	return nil
}

// rings3 writes the exterior and interior linear rings of a GML 3 polygon or triangle.
func (w *gmlWriter) rings3(rings [][]float64, ctype coordinatesType) {
	for i, ring := range rings {
		boundary := "interior"
		if i == 0 {
			boundary = "exterior"
		}
		w.tag(boundary, false)
		w.tag("LinearRing", false)
		w.positions3("posList", ring, ctype)
		w.tag("LinearRing", true)
		w.tag(boundary, true)
	}
}

// positions3 writes a GML 3 pos or posList element, with every ordinate separated by a space. X and Y are swapped
// when the options declare latitude and longitude.
func (w *gmlWriter) positions3(name string, seq []float64, ctype coordinatesType) {
	w.sb.WriteString("<" + w.prefix + name)
	if w.options&gmlNoDimension == 0 {
		w.sb.WriteString(fmt.Sprintf(` srsDimension="%d"`, w.dimensions(ctype)))
	}
	w.sb.WriteString(">")
	dims := ctype.Dimension()
	for i := 0; i+dims <= len(seq); i += dims {
		if i > 0 {
			w.sb.WriteByte(' ')
		}
		w.point3(seq[i : i+w.dimensions(ctype)])
	}
	w.tag(name, true)
}

// point3 writes the ordinates of one GML 3 position separated by spaces, with X and Y swapped when the options
// declare latitude and longitude.
func (w *gmlWriter) point3(point []float64) {
	x, y := point[0], point[1]
	if w.options&gmlLatLon != 0 {
		x, y = y, x
	}
	w.sb.WriteString(formatOrdinate(x, w.precision) + " " + formatOrdinate(y, w.precision))
	for _, val := range point[2:] {
		w.sb.WriteString(" " + formatOrdinate(val, w.precision))
	}
}

// dimensions returns the number of ordinates that GML writes for each point, which leaves out M.
func (w *gmlWriter) dimensions(ctype coordinatesType) int {
	if ctype.Is3D() {
		return 3
	}
	return 2
}

// writeEnvelope writes the bounding box of a shape, as a GML 2 Box or a GML 3 Envelope.
func (w *gmlWriter) writeEnvelope(s shape, srsName string) {
	b, ok := s.bounds()
	name := "Envelope"
	if w.version == 2 {
		name = "Box"
	}
	w.sb.WriteString("<" + w.prefix + name)
	if srsName != "" {
		w.sb.WriteString(` srsName="` + srsName + `"`)
	}
	if !ok {
		w.sb.WriteString("/>")
		return
	}
	lower, upper := []float64{b.minX, b.minY}, []float64{b.maxX, b.maxY}
	if s.hasZ() {
		lower, upper = append(lower, b.minZ), append(upper, b.maxZ)
	}
	if w.version == 2 {
		w.sb.WriteString(">")
		ctype := dimXY
		if s.hasZ() {
			ctype = dimXYZ
		}
		w.coordinates2(append(lower, upper...), ctype)
		w.tag(name, true)
		return
	}
	if w.options&gmlNoDimension == 0 {
		w.sb.WriteString(fmt.Sprintf(` srsDimension="%d"`, len(lower)))
	}
	w.sb.WriteString(">")
	for i, corner := range [][]float64{lower, upper} {
		cornerName := "lowerCorner"
		if i == 1 {
			cornerName = "upperCorner"
		}
		w.tag(cornerName, false)
		w.point3(corner)
		w.tag(cornerName, true)
	}
	w.tag(name, true)
}

// gmlImpl returns an implementation of ST_AsGML. The arguments start with the GML version when `hasVersion` is true,
// and the version is 2 otherwise. Null arguments after the geometry take their defaults. Geography rejects the
// options that only apply to geometry, and declares latitude and longitude when it uses the long SRS name.
func gmlImpl(hasVersion bool, isGeography bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		version := int32(2)
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
		precision, options, prefix, id := 15, int32(0), "gml", ""
		if len(args) > 1 && args[1] != nil {
			precision = int(args[1].(int32))
		}
		if len(args) > 2 && args[2] != nil {
			options = args[2].(int32)
		}
		if len(args) > 3 && args[3] != nil {
			prefix = args[3].(string)
		}
		if len(args) > 4 && args[4] != nil {
			id = args[4].(string)
		}
		if isGeography {
			if options&(gmlLatLon|gmlEnvelope) != 0 {
				return nil, pgerror.Newf(pgcode.Internal, "Options %d passed to ST_AsGML(geography) but are only applicable to ST_AsGML(geometry)", options)
			}
			if options&gmlLongSRS != 0 {
				options |= gmlLatLon
			}
		}
		return writeGML(args[0].(Geometry), version, precision, options, prefix, id)
	}
}

// gmlRoutines returns the routines that read and write GML.
func gmlRoutines() []extdef.Routine {
	precision := extdef.Parameter{Name: "maxdecimaldigits", Type: "int4", Default: `15`}
	options := extdef.Parameter{Name: "options", Type: "int4", Default: `0`}
	geographyTail := []extdef.Parameter{precision, options, {Name: "nprefix", Type: "text", Default: `'gml'::text`}, {Name: "id", Type: "text", Default: `''::text`}}
	return []extdef.Routine{
		{Name: "_st_geomfromgml", Symbol: "_st_geomfromgml_text_int4", Parameters: params("text", "int4"), Returns: "geometry", QueryImpl: stGeomFromGML},
		{Name: "st_geomfromgml", Symbol: "st_geomfromgml_text", Parameters: params("text"), Returns: "geometry", Strict: true, QueryImpl: stGeomFromGML},
		{Name: "st_geomfromgml", Symbol: "st_geomfromgml_text_int4", Parameters: params("text", "int4"), Returns: "geometry", Strict: true, QueryImpl: stGeomFromGML},
		{Name: "st_gmltosql", Symbol: "st_gmltosql_text", Parameters: params("text"), Returns: "geometry", Strict: true, QueryImpl: stGeomFromGML},
		{Name: "st_gmltosql", Symbol: "st_gmltosql_text_int4", Parameters: params("text", "int4"), Returns: "geometry", Strict: true, QueryImpl: stGeomFromGML},
		{Name: "_st_asgml", Symbol: "_st_asgml_int4_geometry_int4_int4_text_text", Parameters: params("int4", "geometry", "int4", "int4", "text", "text"), Returns: "text", Impl: gmlImpl(true, false)},
		{Name: "st_asgml", Symbol: "st_asgml_geography_int4_int4_text_text", Parameters: append([]extdef.Parameter{{Name: "geog", Type: "geography"}}, geographyTail...), Returns: "text", Strict: true, Impl: gmlImpl(false, true)},
		{Name: "st_asgml", Symbol: "st_asgml_geometry_int4_int4", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, precision, options}, Returns: "text", Impl: gmlImpl(false, false)},
		{Name: "st_asgml", Symbol: "st_asgml_text", Parameters: params("text"), Returns: "text", Strict: true, Impl: textArgsAsGeometry(1, gmlImpl(false, false))},
		{Name: "st_asgml", Symbol: "st_asgml_int4_geography_int4_int4_text_text", Parameters: append([]extdef.Parameter{{Name: "version", Type: "int4"}, {Name: "geog", Type: "geography"}}, geographyTail...), Returns: "text", Strict: true, Impl: gmlImpl(true, true)},
		{Name: "st_asgml", Symbol: "st_asgml_int4_geometry_int4_int4_text_text", Parameters: []extdef.Parameter{{Name: "version", Type: "int4"}, {Name: "geom", Type: "geometry"}, precision, options, {Name: "nprefix", Type: "text", Default: `NULL::text`}, {Name: "id", Type: "text", Default: `NULL::text`}}, Returns: "text", Impl: gmlImpl(true, false)},
	}
}

// gmlNamespaces are the namespaces of GML elements.
var gmlNamespaces = []string{"http://www.opengis.net/gml", "http://www.opengis.net/gml/3.2"}

// errInvalidGML is the error for any GML that cannot be read as a geometry.
var errInvalidGML = pgerror.New(pgcode.Internal, "invalid GML representation")

// gmlNumber matches the numbers that GML coordinates may hold.
var gmlNumber = regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]*)?([eE][+-]?[0-9]+)?$`)

// gmlReader reads GML geometries, keeping the SRID of the first element that declares a spatial reference system.
// The runner reads spatial_ref_sys.
type gmlReader struct {
	ctx    *sql.Context
	runner sql.StatementRunner
	srid   int32
	// flat is whether any coordinates lack Z, which leaves the whole geometry without Z.
	flat bool
}

// stGeomFromGML implements ST_GeomFromGML and ST_GMLToSQL, which read a GML geometry. A nonzero SRID argument replaces
// the SRID that the GML declares, without changing the coordinates.
func stGeomFromGML(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	if args[0] == nil {
		return nil, nil
	}
	root, ok := parseXMLTree(args[0].(string), gmlNamespaces...)
	if !ok {
		return nil, errInvalidGML
	}
	r := &gmlReader{ctx: ctx, runner: runner}
	s, err := r.read(root, false)
	if err != nil {
		return nil, err
	}
	if r.flat {
		s = forceDims(s, dimXY, 0, 0)
	} else if gmlHasEmptyMember(s) {
		return nil, pgerror.New(pgcode.Internal, "Dimensions mismatch in lwcollection")
	}
	srid := r.srid
	if len(args) > 1 && args[1] != nil && args[1].(int32) != 0 {
		srid = args[1].(int32)
	}
	return Geometry{SRID: srid, Shape: s}, nil
}

// read reads a GML geometry element. `latLon` is whether the coordinates give latitude before longitude, which an
// element's spatial reference system decides for it and the elements inside it. An element whose spatial reference
// system differs from the geometry's is transformed into the geometry's.
func (r *gmlReader) read(n *xmlNode, latLon bool) (shape, error) {
	if !n.inSpace {
		return shape{}, errInvalidGML
	}
	outer := r.srid
	latLon, srid, err := r.srs(n, latLon)
	if err != nil {
		return shape{}, err
	}
	s, err := r.element(n, latLon)
	if err != nil || outer == sridUnknown || srid == sridUnknown || srid == outer {
		return s, err
	}
	transformed, err := stTransform(r.ctx, r.runner, Geometry{SRID: srid, Shape: s}, outer)
	if err != nil {
		return shape{}, err
	}
	return transformed.(Geometry).Shape, nil
}

// element reads a GML geometry element, whose coordinates give latitude before longitude when `latLon` is true.
func (r *gmlReader) element(n *xmlNode, latLon bool) (shape, error) {
	switch n.name {
	case "Point", "LineString":
		seq, ctype, found, err := r.positions(n, latLon)
		if err != nil {
			return shape{}, err
		}
		typeNum, minPoints, maxPoints := pointType, 1, 1
		if n.name == "LineString" {
			typeNum, minPoints, maxPoints = lineType, 2, math.MaxInt
		}
		if !found {
			return shape{typeNum: typeNum}, nil
		}
		if count := len(seq) / ctype.Dimension(); count < minPoints || count > maxPoints {
			return shape{}, errInvalidGML
		}
		return shape{typeNum: typeNum, ctype: ctype, coords: seq}, nil
	case "Curve":
		var curve shape
		if segments := n.childrenNamed("segments"); len(segments) > 0 {
			for _, segment := range segments[0].childrenNamed("LineStringSegment") {
				seq, ctype, _, err := r.positions(segment, latLon)
				if err != nil {
					return shape{}, err
				}
				dims := ctype.Dimension()
				if len(seq)/dims < 2 {
					return shape{}, errInvalidGML
				}
				segmentShape := shape{typeNum: lineType, ctype: ctype, coords: seq}
				if len(curve.coords) == 0 {
					curve = segmentShape
					continue
				}
				if ctype != curve.ctype {
					curve, segmentShape = forceDims(curve, dimXY, 0, 0), forceDims(segmentShape, dimXY, 0, 0)
					dims = 2
				}
				if !equalOrdinates(curve.coords[len(curve.coords)-dims:], segmentShape.coords[:dims]) {
					return shape{}, errInvalidGML
				}
				curve.coords = append(curve.coords, segmentShape.coords[dims:]...)
			}
		}
		if len(curve.coords) == 0 {
			return shape{}, errInvalidGML
		}
		return curve, nil
	case "LinearRing":
		ring, err := r.ring(n, latLon)
		if err != nil {
			return shape{}, err
		}
		return shape{typeNum: polygonType, ctype: ring.ctype, rings: [][]float64{ring.coords}}, nil
	case "Polygon", "PolygonPatch":
		if len(n.children) == 0 {
			return shape{typeNum: polygonType}, nil
		}
		return r.polygon(n, latLon)
	case "Surface":
		var patches []*xmlNode
		for _, container := range n.childrenNamed("patches") {
			patches = append(patches, container.childrenNamed("PolygonPatch")...)
		}
		if len(patches) != 1 {
			return shape{}, errInvalidGML
		}
		return r.read(patches[0], latLon)
	case "Triangle":
		polygon, err := r.polygon(n, latLon)
		if err != nil || len(polygon.rings) != 1 || len(polygon.rings[0])/polygon.ctype.Dimension() != 4 {
			return shape{}, errInvalidGML
		}
		return shape{typeNum: triangleType, ctype: polygon.ctype, coords: polygon.rings[0]}, nil
	case "Tin", "TriangulatedSurface", "PolyhedralSurface":
		typeNum, patchesName, patchName := tinType, "trianglePatches", "Triangle"
		if n.name == "PolyhedralSurface" {
			typeNum, patchesName, patchName = polyhedralSurfaceType, "polygonPatches", "PolygonPatch"
		}
		var patches []*xmlNode
		for _, container := range n.childrenNamed(patchesName) {
			patches = append(patches, container.childrenNamed(patchName)...)
		}
		return r.collection(typeNum, patches, latLon)
	case "MultiPoint", "MultiLineString", "MultiCurve", "MultiPolygon", "MultiSurface", "MultiGeometry":
		typeNum := map[string]int{"MultiPoint": multiPointType, "MultiLineString": multiLineType, "MultiCurve": multiLineType,
			"MultiPolygon": multiPolygonType, "MultiSurface": multiPolygonType, "MultiGeometry": collectionType}[n.name]
		memberNames := map[string][2]string{"MultiPoint": {"pointMember", "pointMembers"}, "MultiLineString": {"lineStringMember", ""},
			"MultiCurve": {"curveMember", "curveMembers"}, "MultiPolygon": {"polygonMember", ""},
			"MultiSurface": {"surfaceMember", "surfaceMembers"}, "MultiGeometry": {"geometryMember", ""}}[n.name]
		var members []*xmlNode
		for _, child := range n.children {
			if !child.inSpace {
				continue
			}
			if child.name == memberNames[0] && len(child.children) > 0 {
				members = append(members, child.children[0])
			} else if child.name == memberNames[1] {
				members = append(members, child.children...)
			}
		}
		return r.collection(typeNum, members, latLon)
	default:
		return shape{}, errInvalidGML
	}
}

// polygon reads the exterior and interior linear rings of a polygon, patch, or triangle, which must all have the same
// dimensions.
func (r *gmlReader) polygon(n *xmlNode, latLon bool) (shape, error) {
	var exteriors, interiors []*xmlNode
	for _, child := range n.children {
		if child.inSpace && (child.name == "exterior" || child.name == "outerBoundaryIs") {
			exteriors = append(exteriors, child)
		} else if child.inSpace && (child.name == "interior" || child.name == "innerBoundaryIs") {
			interiors = append(interiors, child)
		}
	}
	if len(exteriors) != 1 {
		return shape{}, errInvalidGML
	}
	var rings []shape
	for _, boundary := range append(exteriors, interiors...) {
		linearRings := boundary.childrenNamed("LinearRing")
		if len(linearRings) == 0 {
			return shape{}, errInvalidGML
		}
		ring, err := r.ring(linearRings[0], latLon)
		if err != nil {
			return shape{}, err
		}
		rings = append(rings, ring)
	}
	polygon := shape{typeNum: polygonType, ctype: commonCoordinatesType(rings)}
	for _, ring := range rings {
		polygon.rings = append(polygon.rings, forceDims(ring, polygon.ctype, 0, 0).coords)
	}
	return polygon, nil
}

// collection reads the members of a multi-geometry, collection, or surface of the given type. Points, lines, and
// polygons must be the members of their multi-geometries, and members keep Z only when all of them have it.
func (r *gmlReader) collection(typeNum int, members []*xmlNode, latLon bool) (shape, error) {
	var shapes []shape
	for _, node := range members {
		member, err := r.read(node, latLon)
		if err != nil {
			return shape{}, err
		}
		expected, isMulti := map[int]int{multiPointType: pointType, multiLineType: lineType, multiPolygonType: polygonType}[typeNum]
		if isMulti && member.typeNum != expected {
			return shape{}, pgerror.Newf(pgcode.Internal, "%s cannot contain %s element", typeNameForNumber(typeNum), typeNameForNumber(member.typeNum))
		}
		shapes = append(shapes, member)
	}
	collection := shape{typeNum: typeNum, ctype: commonCoordinatesType(shapes)}
	for _, member := range shapes {
		collection.members = append(collection.members, forceDims(member, collection.ctype, 0, 0))
	}
	return collection, nil
}

// gmlHasEmptyMember returns whether any collection inside the shape has an empty member, which PostGIS cannot add to a
// collection with Z.
func gmlHasEmptyMember(s shape) bool {
	for _, member := range s.members {
		if member.isEmpty() || gmlHasEmptyMember(member) {
			return true
		}
	}
	return false
}

// ring reads a closed linear ring of at least four points.
func (r *gmlReader) ring(n *xmlNode, latLon bool) (shape, error) {
	latLon, _, err := r.srs(n, latLon)
	if err != nil {
		return shape{}, err
	}
	seq, ctype, _, err := r.positions(n, latLon)
	if err != nil {
		return shape{}, err
	}
	dims := ctype.Dimension()
	if len(seq)/dims < 4 || !equalOrdinates(seq[:dims], seq[len(seq)-dims:]) {
		return shape{}, errInvalidGML
	}
	return shape{typeNum: lineType, ctype: ctype, coords: seq}, nil
}

// srs reads the spatial reference system that an element declares, returning whether its coordinates give latitude
// before longitude, or the inherited answer when it declares none, along with the SRID it declares. EPSG codes in URN
// form follow the axis order of their system, while other forms always give longitude first. The first declared SRID
// becomes the geometry's.
func (r *gmlReader) srs(n *xmlNode, latLon bool) (bool, int32, error) {
	name, ok := n.attrs["srsName"]
	if !ok {
		return latLon, sridUnknown, nil
	}
	var code string
	urn := false
	if rest, found := strings.CutPrefix(name, "EPSG:"); found {
		code = rest
	} else if rest, found = strings.CutPrefix(name, "http://www.opengis.net/gml/srs/epsg.xml#"); found {
		code = rest
	} else {
		for _, prefix := range []string{"urn:ogc:def:crs:EPSG:", "urn:EPSG:geographicCRS:", "urn:x-ogc:def:crs:EPSG:"} {
			if strings.HasPrefix(name, prefix) {
				code, urn = name[strings.LastIndexByte(name, ':')+1:], true
			}
		}
	}
	srid, err := strconv.ParseInt(code, 10, 32)
	if err != nil {
		return false, sridUnknown, pgerror.New(pgcode.Internal, "unknown spatial reference system")
	}
	definitions, err := lookUpSRS(r.ctx, r.runner, int32(srid))
	if err != nil {
		return false, sridUnknown, pgerror.New(pgcode.Internal, "unknown spatial reference system")
	}
	if r.srid == sridUnknown {
		r.srid = int32(srid)
	}
	geographic := false
	for _, definition := range definitions {
		if definition != "" {
			geographic = spatial.IsGeographic(definition)
			break
		}
	}
	return urn && geographic, int32(srid), nil
}

// positions reads the points of an element's coordinates, pos, posList, coord, and pointProperty children, in order,
// and returns whether it has any. The points keep Z only when all of them have it. X and Y are swapped when `latLon`
// is true, and a pointProperty's point is read with its own spatial reference system.
func (r *gmlReader) positions(n *xmlNode, latLon bool) ([]float64, coordinatesType, bool, error) {
	var points [][]float64
	found := false
	for _, child := range n.children {
		if !child.inSpace {
			continue
		}
		var lists [][]float64
		switch child.name {
		case "coordinates":
			cs, ts, decimal := gmlAttr(child, "cs", ","), gmlAttr(child, "ts", " "), gmlAttr(child, "decimal", ".")
			if len(cs) != 1 || len(ts) != 1 || len(decimal) != 1 || decimal == cs || decimal == ts {
				return nil, dimXY, false, errInvalidGML
			}
			for _, tuple := range strings.Split(child.text.String(), ts) {
				var point []float64
				for _, val := range strings.Split(tuple, cs) {
					number, ok := gmlParseNumber(strings.ReplaceAll(val, decimal, "."))
					if !ok {
						return nil, dimXY, false, errInvalidGML
					}
					point = append(point, number)
				}
				lists = append(lists, point)
			}
		case "pos", "posList":
			dims := 2
			if text, ok := child.attrs["srsDimension"]; ok {
				dims, _ = strconv.Atoi(text)
			} else if text, ok = child.attrs["dimension"]; ok {
				dims, _ = strconv.Atoi(text)
			}
			fields := strings.Fields(child.text.String())
			if dims < 2 || dims > 3 || len(fields) == 0 || len(fields)%dims != 0 || (child.name == "pos" && len(fields) != dims) {
				return nil, dimXY, false, errInvalidGML
			}
			for i := 0; i < len(fields); i += dims {
				var point []float64
				for _, field := range fields[i : i+dims] {
					number, ok := gmlParseNumber(field)
					if !ok {
						return nil, dimXY, false, errInvalidGML
					}
					point = append(point, number)
				}
				lists = append(lists, point)
			}
		case "coord":
			if len(child.childrenNamed("X")) == 0 || len(child.childrenNamed("Y")) == 0 {
				return nil, dimXY, false, errInvalidGML
			}
			var point []float64
			for _, axis := range []string{"X", "Y", "Z"} {
				if elements := child.childrenNamed(axis); len(elements) > 0 {
					number, ok := gmlParseNumber(strings.TrimSpace(elements[0].text.String()))
					if !ok {
						return nil, dimXY, false, errInvalidGML
					}
					point = append(point, number)
				}
			}
			lists = append(lists, point)
		case "pointProperty", "pointRep":
			pointNodes := child.childrenNamed("Point")
			if len(pointNodes) == 0 {
				return nil, dimXY, false, errInvalidGML
			}
			point, err := r.read(pointNodes[0], latLon)
			if err != nil || point.isEmpty() {
				return nil, dimXY, false, errInvalidGML
			}
			points, found = append(points, point.coords), true
			continue
		default:
			continue
		}
		for _, point := range lists {
			if len(point) < 2 || len(point) > 3 {
				return nil, dimXY, false, errInvalidGML
			}
			if latLon {
				point[0], point[1] = point[1], point[0]
			}
		}
		points, found = append(points, lists...), true
	}
	ctype := dimXYZ
	for _, point := range points {
		if len(point) == 2 {
			ctype, r.flat = dimXY, true
		}
	}
	var seq []float64
	for _, point := range points {
		seq = append(seq, point[:ctype.Dimension()]...)
	}
	if len(points) == 0 {
		ctype = dimXY
	}
	return seq, ctype, found, nil
}

// gmlAttr returns the value of an element's attribute, or the default when the element lacks it.
func gmlAttr(n *xmlNode, name string, defaultValue string) string {
	if val, ok := n.attrs[name]; ok {
		return val
	}
	return defaultValue
}

// gmlParseNumber parses a number of GML coordinates, which must start with a digit after an optional sign.
func gmlParseNumber(text string) (float64, bool) {
	if !gmlNumber.MatchString(text) {
		return 0, false
	}
	val, err := strconv.ParseFloat(text, 64)
	return val, err == nil
}
