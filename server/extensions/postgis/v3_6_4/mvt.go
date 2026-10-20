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
	"encoding/binary"
	"math"

	"github.com/cockroachdb/apd/v3"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// highestDimensionParts returns the parts of a collection that have its highest dimension, as a single part or as a
// multi geometry. Any other shape is returned as it is.
func highestDimensionParts(s shape) shape {
	if s.typeNum != collectionType {
		return s
	}
	var parts []shape
	var collect func(member shape)
	collect = func(member shape) {
		switch member.typeNum {
		case collectionType, multiPointType, multiLineType, multiPolygonType:
			for _, inner := range member.members {
				collect(inner)
			}
		default:
			if !member.isEmpty() {
				parts = append(parts, member)
			}
		}
	}
	collect(s)
	highest := -1
	for _, part := range parts {
		highest = max(highest, part.dimension())
	}
	var kept []shape
	for _, part := range parts {
		if part.dimension() == highest {
			kept = append(kept, part)
		}
	}
	if len(kept) == 1 {
		return kept[0]
	}
	multiType := map[int]int{0: multiPointType, 1: multiLineType, 2: multiPolygonType}[highest]
	return shape{typeNum: multiType, ctype: s.ctype, members: kept}
}

// stAsMVTGeom implements ST_AsMVTGeom, which moves a geometry into the integer coordinate space of a vector tile that
// covers the bounds, with Y running down. The geometry is clipped to the tile and its buffer unless `clip_geom` is
// false, snapped to whole units, and stripped of repeated points. A geometry that collapses is null, and polygons are
// made valid as mvtPolygons describes.
func stAsMVTGeom(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	g, bounds := args[0].(Geometry), args[1].(box2d)
	extent, buffer, clip := args[2].(int32), args[3].(int32), args[4].(bool)
	width, height := bounds.maxX-bounds.minX, bounds.maxY-bounds.minY
	if width <= 0 || height <= 0 {
		return nil, pgerror.New(pgcode.Internal, "ST_AsMVTGeom: Geometric bounds are too small")
	}
	if extent <= 0 {
		return nil, pgerror.New(pgcode.Internal, "ST_AsMVTGeom: Extent must be greater than 0")
	}
	if g.Shape.isEmpty() {
		return nil, nil
	}
	s := highestDimensionParts(g.Shape)
	scaleX, scaleY := float64(extent)/width, float64(extent)/height
	s = s.mapCoords(func(point []float64) []float64 {
		point[0] = (point[0] - bounds.minX) * scaleX
		point[1] = (bounds.maxY - point[1]) * scaleY
		return point
	})
	low, high := -float64(buffer), float64(extent)+float64(buffer)
	if s.dimension() == 2 {
		return mvtPolygons(g, s, clip, low, high)
	}
	if clip {
		clipped, ok, err := clipToTile(s, low, high)
		if err != nil || !ok {
			return nil, err
		}
		s = clipped
	}
	s = s.mapCoords(func(point []float64) []float64 {
		point[0], point[1] = math.RoundToEven(point[0]), math.RoundToEven(point[1])
		return point
	})
	s, err := removeRepeatedShape(s, 0)
	if err != nil {
		return nil, err
	}
	s = dropCollapsed(s)
	if s.isEmpty() {
		return nil, nil
	}
	return g.withShape(s), nil
}

// mvtPolygons finishes ST_AsMVTGeom for polygons, which are snapped to whole units, clipped ring by ring to the
// square from `low` to `high` on both axes when `clip` is true, and then made valid by Wagyu, which fills them by the
// even-odd rule. Points that clipping adds are truncated to whole units. The result keeps only X and Y, and polygons
// that vanish give null.
func mvtPolygons(g Geometry, s shape, clip bool, low float64, high float64) (any, error) {
	dims := s.ctype.Dimension()
	var rings [][]int64
	addRings := func(polygon shape) {
		for _, ring := range polygon.rings {
			var points [][2]float64
			for i := 0; i+dims <= len(ring); i += dims {
				points = append(points, [2]float64{math.RoundToEven(ring[i]), math.RoundToEven(ring[i+1])})
			}
			if clip {
				points = clipRingToSquare(points, low, high)
			}
			if len(points) == 0 {
				continue
			}
			truncated := make([]int64, 0, 2*len(points))
			for _, p := range points {
				truncated = append(truncated, int64(p[0]), int64(p[1]))
			}
			rings = append(rings, truncated)
		}
	}
	if s.typeNum == polygonType {
		addRings(s)
	}
	for _, member := range s.members {
		addRings(member)
	}
	box := [4]int64{int64(low), int64(low), int64(high), int64(high)}
	if !clip {
		box = [4]int64{math.MaxInt64, math.MaxInt64, math.MinInt64, math.MinInt64}
		for _, ring := range rings {
			for i := 0; i+1 < len(ring); i += 2 {
				box[0], box[1] = min(box[0], ring[i]), min(box[1], ring[i+1])
				box[2], box[3] = max(box[2], ring[i]), max(box[3], ring[i+1])
			}
		}
	}
	polygons, err := spatial.WagyuClip(rings, box[0], box[1], box[2], box[3])
	if err != nil {
		return nil, pgerror.New(pgcode.Internal, err.Error())
	}
	toFloats := func(ring []int64) []float64 {
		floats := make([]float64, len(ring))
		for i, v := range ring {
			floats[i] = float64(v)
		}
		return floats
	}
	members := make([]shape, len(polygons))
	for i, polygon := range polygons {
		members[i] = shape{typeNum: polygonType}
		for _, ring := range polygon {
			members[i].rings = append(members[i].rings, toFloats(ring))
		}
	}
	switch len(members) {
	case 0:
		return nil, nil
	case 1:
		return g.withShape(members[0]), nil
	default:
		return g.withShape(shape{typeNum: multiPolygonType, members: members}), nil
	}
}

// clipRingToSquare clips a ring to the square from `low` to `high` on both axes, one side at a time in the order of
// the bottom, right, top, and left sides, keeping the points inside each side and adding the points where the ring
// crosses it. A ring left with fewer than three points is empty, and any other ring is closed.
func clipRingToSquare(ring [][2]float64, low float64, high float64) [][2]float64 {
	edges := [4]float64{low, high, high, low}
	inside := func(p [2]float64, side int) bool {
		return []bool{p[1] > low, p[0] < high, p[1] < high, p[0] > low}[side]
	}
	crossing := func(a [2]float64, b [2]float64, side int) [2]float64 {
		edge := edges[side]
		if side%2 == 0 {
			return [2]float64{a[0] + (b[0]-a[0])*(edge-a[1])/(b[1]-a[1]), edge}
		}
		return [2]float64{edge, a[1] + (b[1]-a[1])*(edge-a[0])/(b[0]-a[0])}
	}
	out := ring
	for side := 0; side < 4 && len(out) > 0; side++ {
		in := out
		out = nil
		previous := in[len(in)-1]
		for _, current := range in {
			if inside(current, side) {
				if !inside(previous, side) {
					out = append(out, crossing(previous, current, side))
				}
				out = append(out, current)
			} else if inside(previous, side) {
				out = append(out, crossing(previous, current, side))
			}
			previous = current
		}
	}
	if len(out) < 3 {
		return nil
	}
	if out[0] != out[len(out)-1] {
		out = append(out, out[0])
	}
	return out
}

// clipToTile clips a shape to the square from `low` to `high` on both axes, returning false when nothing remains.
// Points outside the square are dropped, and linework is clipped by GEOS.
func clipToTile(s shape, low float64, high float64) (shape, bool, error) {
	inside := func(point shape) bool {
		x, y := point.coords[0], point.coords[1]
		return x >= low && x <= high && y >= low && y <= high
	}
	switch s.typeNum {
	case pointType:
		return s, inside(s), nil
	case multiPointType:
		var members []shape
		for _, member := range s.members {
			if !member.isEmpty() && inside(member) {
				members = append(members, member)
			}
		}
		s.members = members
		return s, len(members) > 0, nil
	}
	if b, ok := s.bounds(); ok && b.minX >= low && b.maxX <= high && b.minY >= low && b.maxY <= high {
		return s, true, nil
	}
	geosS, err := s.toGEOS()
	if err != nil {
		return shape{}, false, err
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	clipped, err := geos.ClipByRect(geosS, low, low, high, high)
	if err != nil {
		return shape{}, false, geosError("GEOSClipByRect", err)
	}
	result := highestDimensionParts(fromGEOS(clipped, 0, s.hasZ()).Shape)
	return result, !result.isEmpty() && result.dimension() == s.dimension(), nil
}

// dropCollapsed removes the linestrings with fewer than two distinct points from a shape.
func dropCollapsed(s shape) shape {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case lineType:
		distinct := false
		for i := dims; i+dims <= len(s.coords); i += dims {
			distinct = distinct || s.coords[i] != s.coords[0] || s.coords[i+1] != s.coords[1]
		}
		if !distinct {
			s.coords = nil
		}
	case multiLineType:
		var members []shape
		for _, member := range s.members {
			if kept := dropCollapsed(member); !kept.isEmpty() {
				members = append(members, kept)
			}
		}
		s.members = members
	}
	return s
}

// appendVarint appends a protocol buffers varint.
func appendVarint(buf []byte, v uint64) []byte {
	return binary.AppendUvarint(buf, v)
}

// appendField appends the key of a protocol buffers field with the given number and wire type.
func appendField(buf []byte, field int, wireType int) []byte {
	return appendVarint(buf, uint64(field<<3|wireType))
}

// appendBytesField appends a length-delimited protocol buffers field.
func appendBytesField(buf []byte, field int, data []byte) []byte {
	return append(appendVarint(appendField(buf, field, 2), uint64(len(data))), data...)
}

// zigzag maps a signed integer onto an unsigned one, as protocol buffers' sint encodings do.
func zigzag(v int64) uint64 {
	return uint64((v << 1) ^ (v >> 63))
}

// mvtLayer is the state of ST_AsMVT: the layer's settings, the keys and values that its features refer to, and the
// features encoded so far.
type mvtLayer struct {
	name         string
	extent       uint64
	geomName     string
	featureID    string
	geomColumn   int
	idColumn     int
	resolved     bool
	keys         []string
	keyIndexes   map[string]int
	values       [][]byte
	valueIndexes map[string]int
	features     [][]byte
}

// keyIndex returns the index of a key, adding it when it is new.
func (l *mvtLayer) keyIndex(key string) int {
	index, ok := l.keyIndexes[key]
	if !ok {
		index = len(l.keys)
		l.keys = append(l.keys, key)
		l.keyIndexes[key] = index
	}
	return index
}

// valueIndex returns the index of an encoded value, adding it when it is new.
func (l *mvtLayer) valueIndex(encoded []byte) int {
	index, ok := l.valueIndexes[string(encoded)]
	if !ok {
		index = len(l.values)
		l.values = append(l.values, encoded)
		l.valueIndexes[string(encoded)] = index
	}
	return index
}

// mvtIntegerValue encodes an integer as a tile value, unsigned when it is not negative.
func mvtIntegerValue(v int64) []byte {
	if v >= 0 {
		return appendVarint(appendField(nil, 5, 0), uint64(v))
	}
	return appendVarint(appendField(nil, 6, 0), zigzag(v))
}

// mvtDoubleValue encodes a double as a tile value.
func mvtDoubleValue(v float64) []byte {
	return binary.LittleEndian.AppendUint64(appendField(nil, 3, 1), math.Float64bits(v))
}

// mvtStringValue encodes a string as a tile value.
func mvtStringValue(v string) []byte {
	return appendBytesField(nil, 1, []byte(v))
}

// addJSONTags adds a tag for each member of a JSON object whose value is a string, number, or boolean. Every member
// adds its key, including those whose values are skipped.
func (l *mvtLayer) addJSONTags(tags []uint64, object pgtypes.JsonValueObject) []uint64 {
	for _, item := range object.Items {
		key := l.keyIndex(item.Key)
		var encoded []byte
		switch value := item.Value.(type) {
		case pgtypes.JsonValueString:
			encoded = mvtStringValue(string(value))
		case pgtypes.JsonValueBoolean:
			encoded = appendVarint(appendField(nil, 7, 0), map[bool]uint64{false: 0, true: 1}[bool(value)])
		case pgtypes.JsonValueNumber:
			number := apd.Decimal(value)
			if integer, err := number.Int64(); err == nil {
				encoded = mvtIntegerValue(integer)
			} else if double, err := number.Float64(); err == nil {
				encoded = mvtDoubleValue(double)
			}
		}
		if encoded != nil {
			tags = append(tags, uint64(key), uint64(l.valueIndex(encoded)))
		}
	}
	return tags
}

// mvtGeometry encodes a shape as the type and commands of a tile feature. Coordinates are truncated to integers, and
// each move is relative to the end of the move before it.
func mvtGeometry(s shape) (uint64, []uint64, error) {
	dims := s.ctype.Dimension()
	var commands []uint64
	var cursorX, cursorY int64
	command := func(id int, count int) {
		commands = append(commands, uint64(id&7|count<<3))
	}
	move := func(seq []float64, from int, to int) {
		for i := from; i < to; i++ {
			x, y := int64(seq[i*dims]), int64(seq[i*dims+1])
			commands = append(commands, zigzag(x-cursorX), zigzag(y-cursorY))
			cursorX, cursorY = x, y
		}
	}
	line := func(seq []float64, closed bool) {
		count := len(seq) / dims
		if closed {
			count--
		}
		command(1, 1)
		move(seq, 0, 1)
		command(2, count-1)
		move(seq, 1, count)
		if closed {
			command(7, 1)
		}
	}
	switch s.typeNum {
	case pointType:
		command(1, 1)
		move(s.coords, 0, 1)
		return 1, commands, nil
	case multiPointType:
		var points []float64
		for _, member := range s.members {
			points = append(points, member.coords...)
		}
		command(1, len(points)/dims)
		move(points, 0, len(points)/dims)
		return 1, commands, nil
	case lineType:
		line(s.coords, false)
		return 2, commands, nil
	case multiLineType:
		for _, member := range s.members {
			line(member.coords, false)
		}
		return 2, commands, nil
	case polygonType:
		for _, ring := range s.rings {
			line(ring, true)
		}
		return 3, commands, nil
	case multiPolygonType:
		for _, member := range s.members {
			for _, ring := range member.rings {
				line(ring, true)
			}
		}
		return 3, commands, nil
	default:
		return 0, nil, pgerror.Newf(pgcode.Internal, "encode_feature_geometry: '%s' geometry type not supported", typeNameForNumber(s.typeNum))
	}
}

// jsonbDocument returns the document of a jsonb value, reading it from the value's text when the value holds it in
// another form.
func jsonbDocument(ctx *sql.Context, t *pgtypes.DoltgresType, value any) (pgtypes.JsonDocument, error) {
	if document, ok := value.(pgtypes.JsonDocument); ok {
		return document, nil
	}
	text, err := t.IoOutput(ctx, value)
	if err != nil {
		return pgtypes.JsonDocument{}, err
	}
	return pgtypes.UnmarshalToJsonDocument([]byte(text))
}

// integerValue returns the value of an integer column, which may hold it as a decimal, or false when it holds no
// integer.
func integerValue(value any) (int64, bool) {
	switch value := value.(type) {
	case int16:
		return int64(value), true
	case int32:
		return int64(value), true
	case int64:
		return value, true
	case decimal.Decimal:
		return value.IntPart(), value.IsInteger()
	default:
		return 0, false
	}
}

// isIntegerType returns whether a type is one of the integer types.
func isIntegerType(t *pgtypes.DoltgresType) bool {
	return t.ID == pgtypes.Int16.ID || t.ID == pgtypes.Int32.ID || t.ID == pgtypes.Int64.ID
}

// mvtTransition implements the transition functions of ST_AsMVT, which add a feature for each row that has a
// geometry. The row's other columns become the feature's tags, with the members of a jsonb column each becoming a
// tag of their own, and the column named by `feature_id_name` becomes the feature's ID when it is not negative.
func mvtTransition(ctx *sql.Context, types []*pgtypes.DoltgresType, args ...any) (any, error) {
	layer, _ := args[0].(*mvtLayer)
	if layer == nil {
		layer = &mvtLayer{name: "default", extent: 4096, keyIndexes: map[string]int{}, valueIndexes: map[string]int{}}
		if len(args) > 2 && args[2] != nil {
			layer.name = args[2].(string)
		}
		if len(args) > 3 && args[3] != nil {
			layer.extent = uint64(args[3].(int32))
		}
		if len(args) > 4 && args[4] != nil {
			layer.geomName = args[4].(string)
		}
		if len(args) > 5 && args[5] != nil {
			layer.featureID = args[5].(string)
		}
	}
	values, ok := args[1].([]pgtypes.RecordValue)
	if !ok {
		return layer, nil
	}
	names := make([]string, len(values))
	if len(types) > 1 && types[1] != nil && len(types[1].CompositeAttrs) == len(values) {
		for i, attr := range types[1].CompositeAttrs {
			names[i] = attr.Name
		}
	}
	fieldType := func(i int) *pgtypes.DoltgresType {
		t, _ := values[i].Type.(*pgtypes.DoltgresType)
		return t
	}
	if !layer.resolved {
		layer.geomColumn, layer.idColumn = -1, -1
		for i := range values {
			if t := fieldType(i); t != nil && t.Name() == "geometry" && layer.geomColumn < 0 && (layer.geomName == "" || names[i] == layer.geomName) {
				layer.geomColumn = i
			}
			if layer.featureID != "" && names[i] == layer.featureID {
				if t := fieldType(i); t == nil || !isIntegerType(t) {
					return nil, pgerror.Newf(pgcode.Internal, "mvt_agg_transfn: Could not find column '%s' of integer type", layer.featureID)
				}
				layer.idColumn = i
			}
		}
		if layer.geomColumn < 0 {
			return nil, pgerror.New(pgcode.Internal, "parse_column_keys: no geometry column found")
		}
		if layer.featureID != "" && layer.idColumn < 0 {
			return nil, pgerror.Newf(pgcode.Internal, "mvt_agg_transfn: Could not find column '%s' of integer type", layer.featureID)
		}
		for i := range values {
			if i != layer.geomColumn && i != layer.idColumn && fieldType(i) != nil && fieldType(i).ID != pgtypes.JsonB.ID {
				layer.keyIndex(names[i])
			}
		}
		layer.resolved = true
	}
	geometry, ok := values[layer.geomColumn].Value.(Geometry)
	if !ok {
		return layer, nil
	}
	geomType, commands, err := mvtGeometry(geometry.Shape)
	if err != nil {
		return nil, err
	}
	var tags []uint64
	for i, field := range values {
		if i == layer.geomColumn || i == layer.idColumn || field.Value == nil {
			continue
		}
		t := fieldType(i)
		var encoded []byte
		if integer, ok := integerValue(field.Value); ok && t != nil && isIntegerType(t) {
			tags = append(tags, uint64(layer.keyIndex(names[i])), uint64(layer.valueIndex(mvtIntegerValue(integer))))
			continue
		}
		switch value := field.Value.(type) {
		case float32:
			encoded = binary.LittleEndian.AppendUint32(appendField(nil, 2, 5), math.Float32bits(value))
		case float64:
			encoded = mvtDoubleValue(value)
		case bool:
			encoded = appendVarint(appendField(nil, 7, 0), map[bool]uint64{false: 0, true: 1}[value])
		default:
			if t == nil {
				continue
			}
			if t.ID == pgtypes.JsonB.ID {
				document, err := jsonbDocument(ctx, t, value)
				if err != nil {
					return nil, err
				}
				if object, ok := document.Value.(pgtypes.JsonValueObject); ok {
					tags = layer.addJSONTags(tags, object)
				}
				continue
			}
			text, err := t.IoOutput(ctx, value)
			if err != nil {
				return nil, err
			}
			encoded = mvtStringValue(text)
		}
		tags = append(tags, uint64(layer.keyIndex(names[i])), uint64(layer.valueIndex(encoded)))
	}
	var feature []byte
	if layer.idColumn >= 0 && values[layer.idColumn].Value != nil {
		if id, ok := integerValue(values[layer.idColumn].Value); ok && id >= 0 {
			feature = appendVarint(appendField(feature, 1, 0), uint64(id))
		}
	}
	if len(tags) > 0 {
		var packed []byte
		for _, tag := range tags {
			packed = appendVarint(packed, tag)
		}
		feature = appendBytesField(feature, 2, packed)
	}
	feature = appendVarint(appendField(feature, 3, 0), geomType)
	var packed []byte
	for _, c := range commands {
		packed = appendVarint(packed, c)
	}
	feature = appendBytesField(feature, 4, packed)
	layer.features = append(layer.features, feature)
	return layer, nil
}

// mvtFinal implements the final function of ST_AsMVT, which encodes the layer as a vector tile. A layer without
// features is an empty tile.
func mvtFinal(ctx *sql.Context, args ...any) (any, error) {
	layer, _ := args[0].(*mvtLayer)
	if layer == nil || len(layer.features) == 0 {
		return []byte{}, nil
	}
	var encoded []byte
	encoded = appendBytesField(encoded, 1, []byte(layer.name))
	for _, feature := range layer.features {
		encoded = appendBytesField(encoded, 2, feature)
	}
	for _, key := range layer.keys {
		encoded = appendBytesField(encoded, 3, []byte(key))
	}
	for _, value := range layer.values {
		encoded = appendBytesField(encoded, 4, value)
	}
	encoded = appendVarint(appendField(encoded, 5, 0), layer.extent)
	encoded = appendVarint(appendField(encoded, 15, 0), 2)
	return appendBytesField(nil, 3, encoded), nil
}

// mvtRoutines returns the routines that build vector tiles.
func mvtRoutines() []extdef.Routine {
	transition := func(symbol string, extra ...string) extdef.Routine {
		return extdef.Routine{Name: "pgis_asmvt_transfn", Symbol: symbol, Parameters: params(append([]string{"internal", "anyelement"}, extra...)...), Returns: "internal", TypedImpl: mvtTransition}
	}
	return []extdef.Routine{
		transition("pgis_asmvt_transfn_internal_anyelement"),
		transition("pgis_asmvt_transfn_internal_anyelement_text", "text"),
		transition("pgis_asmvt_transfn_internal_anyelement_text_int4", "text", "int4"),
		transition("pgis_asmvt_transfn_internal_anyelement_text_int4_text", "text", "int4", "text"),
		transition("pgis_asmvt_transfn", "text", "int4", "text", "text"),
		{Name: "pgis_asmvt_finalfn", Symbol: "pgis_asmvt_finalfn", Parameters: params("internal"), Returns: "bytea", Impl: mvtFinal},
		{Name: "st_asmvtgeom", Symbol: "ST_AsMVTGeom", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "bounds", Type: "box2d"}, {Name: "extent", Type: "int4", Default: "4096"}, {Name: "buffer", Type: "int4", Default: "256"}, {Name: "clip_geom", Type: "bool", Default: "true"}}, Returns: "geometry", Impl: stAsMVTGeom},
	}
}

// mvtAggregates returns the aggregate that builds vector tiles.
func mvtAggregates() []extdef.Aggregate {
	aggregate := func(transition string, extra ...string) extdef.Aggregate {
		return extdef.Aggregate{Name: "st_asmvt", Parameters: params(append([]string{"anyelement"}, extra...)...), Returns: "bytea", StateType: "internal", Transition: transition, Final: "pgis_asmvt_finalfn"}
	}
	return []extdef.Aggregate{
		aggregate("pgis_asmvt_transfn_internal_anyelement"),
		aggregate("pgis_asmvt_transfn_internal_anyelement_text", "text"),
		aggregate("pgis_asmvt_transfn_internal_anyelement_text_int4", "text", "int4"),
		aggregate("pgis_asmvt_transfn_internal_anyelement_text_int4_text", "text", "int4", "text"),
		aggregate("pgis_asmvt_transfn", "text", "int4", "text", "text"),
	}
}
