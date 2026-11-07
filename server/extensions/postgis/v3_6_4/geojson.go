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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// The option bits of ST_AsGeoJSON.
const (
	geojsonBBox            = 1
	geojsonShortCRS        = 2
	geojsonLongCRS         = 4
	geojsonShortCRSNot4326 = 8
)

// sridAuthority returns the authority name and code that the spatial reference system table holds for the SRID, or an
// error when the table does not hold the SRID.
func sridAuthority(srid int32) (string, int, error) {
	//TODO: read the spatial_ref_sys table, so that systems that users add are found, once casts can run queries
	found, ok := databaseSRSFor(srid)
	if !ok {
		return "", 0, pgerror.Newf(pgcode.Internal, "SRID %d unknown in spatial_ref_sys table", srid)
	}
	return found.authority, int(srid), nil
}

// sridForAuthority returns the SRID that the spatial reference system table holds for an authority name and code, or
// zero when it holds none.
func sridForAuthority(name string, code int) int32 {
	//TODO: read the spatial_ref_sys table, so that systems that users add are found, once casts can run queries
	if found, ok := databaseSRSFor(int32(code)); ok && strings.EqualFold(found.authority, name) {
		return int32(code)
	}
	return sridUnknown
}

// geojsonWriter writes geometries as GeoJSON.
type geojsonWriter struct {
	sb        strings.Builder
	precision int
}

// writeGeoJSON returns the GeoJSON of a geometry with the given number of decimal digits and option bits.
func writeGeoJSON(g Geometry, precision int, options int) (string, error) {
	w := &geojsonWriter{precision: precision}
	var crs string
	if g.SRID != sridUnknown {
		if options&geojsonLongCRS != 0 || options&geojsonShortCRS != 0 || (options&geojsonShortCRSNot4326 != 0 && g.SRID != sridDefault) {
			name, code, err := sridAuthority(g.SRID)
			if err != nil {
				return "", err
			}
			if options&geojsonLongCRS != 0 {
				crs = fmt.Sprintf("urn:ogc:def:crs:%s::%d", name, code)
			} else {
				crs = fmt.Sprintf("%s:%d", name, code)
			}
		}
	}
	if err := w.write(g.Shape, crs, options&geojsonBBox != 0, true); err != nil {
		return "", err
	}
	return w.sb.String(), nil
}

// write writes a shape as a GeoJSON object, adding the CRS and bounding box members only to the top-level object.
func (w *geojsonWriter) write(s shape, crs string, withBBox bool, isTop bool) error {
	typeName := ""
	switch s.typeNum {
	case pointType:
		typeName = "Point"
	case lineType:
		typeName = "LineString"
	case polygonType, triangleType:
		typeName = "Polygon"
	case multiPointType:
		typeName = "MultiPoint"
	case multiLineType:
		typeName = "MultiLineString"
	case multiPolygonType:
		typeName = "MultiPolygon"
	case collectionType, tinType:
		if !isTop {
			return pgerror.New(pgcode.Internal, "GeoJson: geometry not supported.")
		}
		typeName = "GeometryCollection"
	default:
		return pgerror.Newf(pgcode.Internal, "lwgeom_to_geojson: '%s' geometry type not supported", typeNameForNumber(s.typeNum))
	}
	w.sb.WriteString(`{"type":"` + typeName + `"`)
	if crs != "" {
		w.sb.WriteString(`,"crs":{"type":"name","properties":{"name":"` + crs + `"}}`)
	}
	if withBBox {
		w.writeBBox(s)
	}
	if typeName == "GeometryCollection" {
		w.sb.WriteString(`,"geometries":[`)
		for i, member := range s.members {
			if i > 0 {
				w.sb.WriteByte(',')
			}
			if member.typeNum == triangleType {
				member = shape{typeNum: polygonType, ctype: member.ctype, rings: [][]float64{member.coords}}
			}
			if err := w.write(member, "", false, false); err != nil {
				return err
			}
		}
		w.sb.WriteString("]}")
		return nil
	}
	w.sb.WriteString(`,"coordinates":`)
	if s.isEmpty() {
		w.sb.WriteString("[]}")
		return nil
	}
	w.writeCoordinates(s)
	w.sb.WriteByte('}')
	return nil
}

// writeBBox writes the bounding box member, in three dimensions when the shape has Z coordinates. An empty shape has
// a box of zeros unless it is a collection without members, which has no box.
func (w *geojsonWriter) writeBBox(s shape) {
	b, ok := s.bounds()
	if !ok {
		if !hasSequence(s.typeNum) && s.typeNum != polygonType && len(s.members) == 0 {
			return
		}
		b = bbox{}
	}
	vals := []float64{b.minX, b.minY, b.maxX, b.maxY}
	if s.hasZ() {
		vals = []float64{b.minX, b.minY, b.minZ, b.maxX, b.maxY, b.maxZ}
	}
	w.sb.WriteString(`,"bbox":[`)
	for i, val := range vals {
		if i > 0 {
			w.sb.WriteByte(',')
		}
		w.sb.WriteString(strconv.FormatFloat(val, 'f', max(w.precision, 0), 64))
	}
	w.sb.WriteByte(']')
}

// writeCoordinates writes the coordinates array of a nonempty shape, where empty members are empty arrays.
func (w *geojsonWriter) writeCoordinates(s shape) {
	switch s.typeNum {
	case pointType:
		w.writePositions(s.coords, s, false)
	case lineType:
		w.writePositions(s.coords, s, true)
	case triangleType:
		w.sb.WriteByte('[')
		w.writePositions(s.coords, s, true)
		w.sb.WriteByte(']')
	case polygonType:
		w.sb.WriteByte('[')
		for i, ring := range s.rings {
			if i > 0 {
				w.sb.WriteByte(',')
			}
			w.writePositions(ring, s, true)
		}
		w.sb.WriteByte(']')
	default:
		w.sb.WriteByte('[')
		for i, member := range s.members {
			if i > 0 {
				w.sb.WriteByte(',')
			}
			if member.isEmpty() {
				w.sb.WriteString("[]")
			} else {
				w.writeCoordinates(member)
			}
		}
		w.sb.WriteByte(']')
	}
}

// writePositions writes the positions of a sequence, each holding X, Y, and Z when the shape has Z coordinates. A
// single position is written alone unless `asArray` is true.
func (w *geojsonWriter) writePositions(seq []float64, s shape, asArray bool) {
	dims := s.ctype.Dimension()
	if asArray {
		w.sb.WriteByte('[')
	}
	for i := 0; i+dims <= len(seq); i += dims {
		if i > 0 {
			w.sb.WriteByte(',')
		}
		w.sb.WriteString("[" + formatOrdinate(seq[i], w.precision) + "," + formatOrdinate(seq[i+1], w.precision))
		if s.hasZ() {
			w.sb.WriteString("," + formatOrdinate(seq[i+2], w.precision))
		}
		w.sb.WriteByte(']')
	}
	if asArray {
		w.sb.WriteByte(']')
	}
}

// stAsGeoJSON implements ST_AsGeoJSON for geometry and geography.
func stAsGeoJSON(ctx *sql.Context, args ...any) (any, error) {
	return writeGeoJSON(args[0].(Geometry), int(args[1].(int32)), int(args[2].(int32)))
}

// geometryToJSON implements the json function and cast for geometry, which writes GeoJSON with a short CRS and 15
// decimal digits.
func geometryToJSON(ctx *sql.Context, args ...any) (any, error) {
	return writeGeoJSON(args[0].(Geometry), 15, geojsonShortCRS)
}

// stAsGeoJSONRow implements ST_AsGeoJSON for a row, which writes it as a GeoJSON Feature. Its geometry is the geometry
// or geography column that `geom_column` names, or the first one when that is empty, written with the given number of
// decimal digits. The column that `id_column` names becomes the feature's id, and every other column becomes a
// property, written as to_json writes it, except that a geometry is written as its json cast writes it. A pretty row
// puts each property after the first on its own line.
func stAsGeoJSONRow(ctx *sql.Context, types []*pgtypes.DoltgresType, args ...any) (any, error) {
	values := args[0].([]pgtypes.RecordValue)
	geomColumn, idColumn := args[1].(string), args[4].(string)
	names := make([]string, len(values))
	columnTypes := make([]*pgtypes.DoltgresType, len(values))
	for i, value := range values {
		names[i] = fmt.Sprintf("f%d", i+1)
		if len(types[0].CompositeAttrs) == len(values) {
			names[i] = types[0].CompositeAttrs[i].Name
		}
		columnTypes[i], _ = value.Type.(*pgtypes.DoltgresType)
	}
	geomIndex, idIndex := -1, -1
	for i, t := range columnTypes {
		if geomIndex < 0 && t != nil && (t.Name() == "geometry" || t.Name() == "geography") && (geomColumn == "" || names[i] == geomColumn) {
			geomIndex = i
		}
	}
	if geomIndex < 0 {
		return nil, pgerror.New(pgcode.InvalidParameterValue, "geometry column is missing")
	}
	if idColumn != "" {
		for i, name := range names {
			if i != geomIndex && name == idColumn {
				idIndex = i
			}
		}
		if idIndex < 0 {
			return nil, pgerror.Newf(pgcode.InvalidParameterValue, `Specified id column "%s" is missing`, idColumn)
		}
	}
	toJSON := func(i int) (string, error) {
		if g, ok := values[i].Value.(Geometry); ok && columnTypes[i].Name() == "geometry" {
			text, err := geometryToJSON(ctx, g)
			if err != nil {
				return "", err
			}
			return text.(string), nil
		}
		raw, err := pgtypes.ValueToJSON(ctx, columnTypes[i], values[i].Value)
		return string(raw), err
	}
	var sb strings.Builder
	sb.WriteString(`{"type": "Feature", "geometry": `)
	if g, ok := values[geomIndex].Value.(Geometry); ok {
		text, err := writeGeoJSON(g, int(args[2].(int32)), geojsonShortCRSNot4326)
		if err != nil {
			return nil, err
		}
		sb.WriteString(text)
	} else {
		sb.WriteString("null")
	}
	if idIndex >= 0 {
		id, err := toJSON(idIndex)
		if err != nil {
			return nil, err
		}
		sb.WriteString(`, "id": ` + id)
	}
	sb.WriteString(`, "properties": {`)
	separator := ""
	for i := range values {
		if i == geomIndex || i == idIndex {
			continue
		}
		key, err := pgtypes.ValueToJSON(ctx, pgtypes.Text, names[i])
		if err != nil {
			return nil, err
		}
		value, err := toJSON(i)
		if err != nil {
			return nil, err
		}
		sb.WriteString(separator + string(key) + ": " + value)
		separator = ", "
		if args[3].(bool) {
			separator = ",\n "
		}
	}
	sb.WriteString("}}")
	return sb.String(), nil
}

// defaultGeoJSONPrecision is the number of decimal digits that GeoJSON is written with by default.
const defaultGeoJSONPrecision = 9

// errInvalidGeoJSON is returned for GeoJSON that is not a geometry object of a known type.
var errInvalidGeoJSON = pgerror.New(pgcode.Internal, "invalid GeoJson representation")

// parseGeoJSON reads a GeoJSON geometry object. Its SRID belongs to the authority and code that its "crs" member
// names, written as "<authority>:<code>" or "urn:ogc:def:crs:<authority>::<code>", and is zero for any other name. It
// returns false when there is no "crs" member.
func parseGeoJSON(ctx *sql.Context, input string) (Geometry, bool, error) {
	var obj map[string]any
	if err := json.Unmarshal([]byte(input), &obj); err != nil {
		return Geometry{}, false, pgerror.New(pgcode.Internal, err.Error())
	}
	s, err := geojsonShape(obj)
	if err != nil {
		return Geometry{}, false, err
	}
	if s.hasZ() {
		s = promoteGeoJSONZ(s)
	} else {
		s = dropGeoJSONZ(s)
	}
	crs, hasCRS := obj["crs"].(map[string]any)
	if !hasCRS {
		return Geometry{Shape: s}, false, nil
	}
	properties, _ := crs["properties"].(map[string]any)
	name, _ := properties["name"].(string)
	authority, code, _ := strings.Cut(strings.Replace(strings.TrimPrefix(name, "urn:ogc:def:crs:"), "::", ":", 1), ":")
	parsed, err := strconv.Atoi(code)
	if err != nil {
		return Geometry{Shape: s}, true, nil
	}
	return Geometry{SRID: sridForAuthority(authority, parsed), Shape: s}, true, nil
}

// geojsonShape reads the shape of a GeoJSON geometry object. Positions with a third ordinate mark the shape as having
// Z coordinates, which promoteGeoJSONZ then gives to every position.
func geojsonShape(obj map[string]any) (shape, error) {
	typeName, _ := obj["type"].(string)
	typeNums := map[string]int{"Point": pointType, "LineString": lineType, "Polygon": polygonType, "MultiPoint": multiPointType,
		"MultiLineString": multiLineType, "MultiPolygon": multiPolygonType, "GeometryCollection": collectionType}
	typeNum, ok := typeNums[typeName]
	if !ok {
		return shape{}, errInvalidGeoJSON
	}
	if typeNum == collectionType {
		geometries, ok := obj["geometries"].([]any)
		if !ok {
			return shape{}, pgerror.New(pgcode.Internal, "Unable to find 'geometries' in GeoJSON string")
		}
		collection := shape{typeNum: collectionType}
		for _, member := range geometries {
			memberObj, ok := member.(map[string]any)
			if !ok {
				return shape{}, errInvalidGeoJSON
			}
			memberShape, err := geojsonShape(memberObj)
			if err != nil {
				return shape{}, err
			}
			collection.ctype |= memberShape.ctype
			collection.members = append(collection.members, memberShape)
		}
		return collection, nil
	}
	coordinates, ok := obj["coordinates"].([]any)
	if !ok {
		return shape{}, pgerror.New(pgcode.Internal, "Unable to find 'coordinates' in GeoJSON string")
	}
	return geojsonCoordinates(typeNum, coordinates)
}

// geojsonCoordinates reads the coordinates array of a GeoJSON geometry of the given type.
func geojsonCoordinates(typeNum int, coordinates []any) (shape, error) {
	s := shape{typeNum: typeNum}
	switch typeNum {
	case pointType:
		if len(coordinates) == 0 {
			return s, nil
		}
		return s, appendGeoJSONPosition(&s, &s.coords, coordinates)
	case lineType:
		for _, position := range coordinates {
			if err := appendGeoJSONPosition(&s, &s.coords, position); err != nil {
				return shape{}, err
			}
		}
		return s, nil
	case polygonType:
		for _, ring := range coordinates {
			ringShape, err := geojsonCoordinates(lineType, asArray(ring))
			if err != nil {
				return shape{}, err
			}
			s.ctype |= ringShape.ctype
			s.rings = append(s.rings, ringShape.coords)
		}
		return s, nil
	default:
		memberType := typeNum - 3
		for _, member := range coordinates {
			memberShape, err := geojsonCoordinates(memberType, asArray(member))
			if err != nil {
				return shape{}, err
			}
			s.ctype |= memberShape.ctype
			s.members = append(s.members, memberShape)
		}
		return s, nil
	}
}

// asArray returns a JSON value as an array, or an empty array when it is not one.
func asArray(value any) []any {
	array, _ := value.([]any)
	return array
}

// appendGeoJSONPosition appends the X, Y, and Z of a GeoJSON position to a sequence, where a missing Z is zero, and
// marks the shape as having Z coordinates when the position has a third ordinate. Ordinates that are not numbers read
// as zero.
func appendGeoJSONPosition(s *shape, seq *[]float64, position any) error {
	ordinates := asArray(position)
	if len(ordinates) < 2 {
		return pgerror.New(pgcode.Internal, "Too few ordinates in GeoJSON")
	}
	for i := 0; i < len(ordinates) && i < 3; i++ {
		value, _ := ordinates[i].(float64)
		*seq = append(*seq, value)
	}
	if len(ordinates) == 2 {
		*seq = append(*seq, 0)
	} else {
		s.ctype = dimXYZ
	}
	return nil
}

// promoteGeoJSONZ returns the shape with Z coordinates throughout, after reading set a Z coordinate for every position.
func promoteGeoJSONZ(s shape) shape {
	s.ctype = dimXYZ
	for i, member := range s.members {
		s.members[i] = promoteGeoJSONZ(member)
	}
	return s
}

// dropGeoJSONZ returns the shape without the Z coordinates that reading set to zero for every position.
func dropGeoJSONZ(s shape) shape {
	return forceDims(s.withDims(dimXYZ), dimXY, 0, 0)
}

// stGeomFromGeoJSON implements ST_GeomFromGeoJSON for text, json, and jsonb, whose geometry is in SRID 4326 unless
// the GeoJSON names another.
func stGeomFromGeoJSON(ctx *sql.Context, args ...any) (any, error) {
	input, ok := args[0].(string)
	if doc, isJSONB := args[0].(pgtypes.JsonDocument); isJSONB {
		var sb strings.Builder
		pgtypes.JsonValueFormatter(&sb, doc.Value)
		input, ok = sb.String(), true
	}
	if !ok {
		input = fmt.Sprint(args[0])
	}
	g, hasCRS, err := parseGeoJSON(ctx, input)
	if err != nil {
		return nil, err
	}
	if !hasCRS {
		g.SRID = sridDefault
	}
	return g, nil
}

// geojsonRoutines returns the routines that write GeoJSON.
func geojsonRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "json", Symbol: "geometry_to_json", Parameters: params("geometry"), Returns: "json", Strict: true, Impl: geometryToJSON},
		{Name: "jsonb", Symbol: "geometry_to_jsonb", Parameters: params("geometry"), Returns: "jsonb", Strict: true, Impl: geometryToJSON},
		{Name: "st_asgeojson", Symbol: "geography_as_geojson", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, {Name: "maxdecimaldigits", Type: "int4", Default: `9`}, {Name: "options", Type: "int4", Default: `0`}}, Returns: "text", Strict: true, Impl: stAsGeoJSON},
		{Name: "st_asgeojson", Symbol: "LWGEOM_asGeoJson", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "maxdecimaldigits", Type: "int4", Default: `9`}, {Name: "options", Type: "int4", Default: `8`}}, Returns: "text", Strict: true, Impl: stAsGeoJSON},
		{Name: "st_geomfromgeojson", Symbol: "geom_from_geojson", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: stGeomFromGeoJSON},
		{Name: "st_geomfromgeojson", Symbol: "st_geomfromgeojson_json", Parameters: params("json"), Returns: "geometry", Strict: true, Impl: stGeomFromGeoJSON},
		{Name: "st_geomfromgeojson", Symbol: "st_geomfromgeojson_jsonb", Parameters: params("jsonb"), Returns: "geometry", Strict: true, Impl: stGeomFromGeoJSON},
		{Name: "st_asgeojson", Symbol: "ST_AsGeoJsonRow", Parameters: []extdef.Parameter{{Name: "r", Type: "record"}, {Name: "geom_column", Type: "text", Default: `''::text`}, {Name: "maxdecimaldigits", Type: "int4", Default: `9`}, {Name: "pretty_bool", Type: "bool", Default: `false`}, {Name: "id_column", Type: "text", Default: `''::text`}}, Returns: "text", Strict: true, TypedImpl: stAsGeoJSONRow},
		{Name: "st_asgeojson", Symbol: "st_asgeojson_text", Parameters: params("text"), Returns: "text", Strict: true, Impl: textArgsAsGeometry(1, func(ctx *sql.Context, args ...any) (any, error) {
			return writeGeoJSON(args[0].(Geometry), defaultGeoJSONPrecision, 0)
		})},
	}
}
