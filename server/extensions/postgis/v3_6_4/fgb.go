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
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"

	flatbuffers "github.com/dolthub/flatbuffers/v23/go"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/server/hook"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// fgbMagic starts every FlatGeobuf file.
var fgbMagic = []byte{'f', 'g', 'b', 3, 'f', 'g', 'b', 1}

// The FlatGeobuf column types that PostgreSQL columns map to.
const (
	fgbBool     = 2
	fgbShort    = 3
	fgbInt      = 5
	fgbLong     = 7
	fgbFloat    = 9
	fgbDouble   = 10
	fgbString   = 11
	fgbJSON     = 12
	fgbDateTime = 13
	fgbBinary   = 14
)

// fgbColumnTypeNames holds the name of each FlatGeobuf column type.
var fgbColumnTypeNames = map[byte]string{0: "Byte", 1: "UByte", fgbBool: "Bool", fgbShort: "Short", 4: "UShort", fgbInt: "Int", 6: "UInt",
	fgbLong: "Long", 8: "ULong", fgbFloat: "Float", fgbDouble: "Double", fgbString: "String", fgbJSON: "Json", fgbDateTime: "DateTime", fgbBinary: "Binary"}

// fgbColumnTypes maps each PostgreSQL type that a FlatGeobuf column can hold to its column type.
var fgbColumnTypes = map[id.Type]byte{
	pgtypes.Bool.ID: fgbBool, pgtypes.Int16.ID: fgbShort, pgtypes.Int32.ID: fgbInt, pgtypes.Int64.ID: fgbLong,
	pgtypes.Float32.ID: fgbFloat, pgtypes.Float64.ID: fgbDouble, pgtypes.Text.ID: fgbString, pgtypes.VarChar.ID: fgbString, pgtypes.Unknown.ID: fgbString,
	pgtypes.JsonB.ID: fgbJSON, pgtypes.Date.ID: fgbDateTime, pgtypes.Timestamp.ID: fgbDateTime, pgtypes.TimestampTZ.ID: fgbDateTime,
	pgtypes.Bytea.ID: fgbBinary,
}

// fgbTableTypes holds the PostgreSQL type that a table created from a file gives each column type.
var fgbTableTypes = map[byte]string{fgbBool: "boolean", fgbShort: "smallint", fgbInt: "integer", fgbLong: "bigint", fgbFloat: "real",
	fgbDouble: "double precision", fgbString: "text", fgbJSON: "jsonb", fgbDateTime: "timestamptz", fgbBinary: "bytea"}

// fgbGeometryTypes maps the shape types that FlatGeobuf can hold to its geometry types.
var fgbGeometryTypes = map[int]byte{pointType: 1, lineType: 2, polygonType: 3, multiPointType: 4, multiLineType: 5, multiPolygonType: 6,
	collectionType: 7, polyhedralSurfaceType: 15, tinType: 16, triangleType: 17}

// fgbColumn is a column of a FlatGeobuf file.
type fgbColumn struct {
	name       string
	columnType byte
}

// fgbFeature is a row that ST_AsFlatGeobuf gathers: its geometry, which is empty when the row has none, and its
// encoded properties.
type fgbFeature struct {
	geometry   shape
	properties []byte
}

// fgbCollection is the state of ST_AsFlatGeobuf.
type fgbCollection struct {
	index        bool
	geomName     string
	resolved     bool
	geomColumn   int
	columns      []fgbColumn
	columnFields []int
	geometryType byte
	hasZ, hasM   bool
	srid         int32
	features     []fgbFeature
}

// fgbTransition implements the transition functions of ST_AsFlatGeobuf, which gather each row. The first row sets the
// file's columns, geometry type, dimensions, and SRID, and every later geometry must have the same type.
func fgbTransition(ctx *sql.Context, types []*pgtypes.DoltgresType, args ...any) (any, error) {
	collection, _ := args[0].(*fgbCollection)
	if collection == nil {
		collection = &fgbCollection{}
		if len(args) > 2 && args[2] != nil {
			collection.index = args[2].(bool)
		}
		if len(args) > 3 && args[3] != nil {
			collection.geomName = args[3].(string)
		}
	}
	values, ok := args[1].([]pgtypes.RecordValue)
	if !ok {
		return collection, nil
	}
	names := make([]string, len(values))
	if len(types) > 1 && types[1] != nil && len(types[1].CompositeAttrs) == len(values) {
		for i, attr := range types[1].CompositeAttrs {
			names[i] = attr.Name
		}
	}
	geometry, hasGeometry := shape{}, false
	if !collection.resolved {
		collection.geomColumn = -1
		for i, value := range values {
			t, _ := value.Type.(*pgtypes.DoltgresType)
			if t != nil && t.Name() == "geometry" && collection.geomColumn < 0 && (collection.geomName == "" || names[i] == collection.geomName) {
				collection.geomColumn = i
				continue
			}
			if t == nil {
				continue
			}
			columnType, ok := fgbColumnTypes[t.ID]
			if !ok {
				return nil, pgerror.Newf(pgcode.Internal, "flatgeobuf: get_column_type: '%d' column type not supported", id.Cache().ToOID(t.ID.AsId()))
			}
			collection.columns = append(collection.columns, fgbColumn{name: names[i], columnType: columnType})
			collection.columnFields = append(collection.columnFields, i)
		}
		if collection.geomColumn < 0 {
			return nil, pgerror.New(pgcode.Internal, "no geom column found")
		}
		if g, ok := values[collection.geomColumn].Value.(Geometry); ok {
			geometry, hasGeometry = g.Shape, true
			geometryType, ok := fgbGeometryTypes[g.Shape.typeNum]
			if !ok {
				return nil, pgerror.Newf(pgcode.Internal, "flatgeobuf: get_geometrytype: '%s' geometry type not supported", typeNameForNumber(g.Shape.typeNum))
			}
			collection.geometryType, collection.hasZ, collection.hasM, collection.srid = geometryType, g.Shape.hasZ(), g.Shape.hasM(), g.SRID
		}
		collection.resolved = true
	} else if g, ok := values[collection.geomColumn].Value.(Geometry); ok {
		geometry, hasGeometry = g.Shape, true
		if geometryType, ok := fgbGeometryTypes[g.Shape.typeNum]; !ok || geometryType != collection.geometryType {
			return nil, pgerror.New(pgcode.Internal, "mixed geometry type is not supported")
		}
	}
	if hasGeometry && (geometry.typeNum == triangleType || geometry.typeNum == polyhedralSurfaceType || geometry.typeNum == tinType) {
		return nil, pgerror.Newf(pgcode.Internal, "flatgeobuf: GeometryWriter::write: '%s' geometry type not supported", typeNameForNumber(geometry.typeNum))
	}
	feature := fgbFeature{geometry: geometry}
	for columnIndex, field := range collection.columnFields {
		value := values[field].Value
		if value == nil {
			continue
		}
		feature.properties = binary.LittleEndian.AppendUint16(feature.properties, uint16(columnIndex))
		encoded, err := fgbPropertyValue(ctx, collection.columns[columnIndex].columnType, values[field])
		if err != nil {
			return nil, err
		}
		feature.properties = append(feature.properties, encoded...)
	}
	collection.features = append(collection.features, feature)
	return collection, nil
}

// fgbPropertyValue encodes a property value of the given column type: numbers and booleans in little-endian order,
// and everything else as its length followed by its text, or its bytes for binary values.
func fgbPropertyValue(ctx *sql.Context, columnType byte, field pgtypes.RecordValue) ([]byte, error) {
	switch columnType {
	case fgbBool:
		return []byte{map[bool]byte{false: 0, true: 1}[field.Value.(bool)]}, nil
	case fgbShort, fgbInt, fgbLong:
		integer, _ := integerValue(field.Value)
		switch columnType {
		case fgbShort:
			return binary.LittleEndian.AppendUint16(nil, uint16(integer)), nil
		case fgbInt:
			return binary.LittleEndian.AppendUint32(nil, uint32(integer)), nil
		default:
			return binary.LittleEndian.AppendUint64(nil, uint64(integer)), nil
		}
	case fgbFloat:
		return binary.LittleEndian.AppendUint32(nil, math.Float32bits(field.Value.(float32))), nil
	case fgbDouble:
		return binary.LittleEndian.AppendUint64(nil, math.Float64bits(field.Value.(float64))), nil
	case fgbBinary:
		data := field.Value.([]byte)
		return append(binary.LittleEndian.AppendUint32(nil, uint32(len(data))), data...), nil
	default:
		t, _ := field.Type.(*pgtypes.DoltgresType)
		text, err := t.IoOutput(ctx, field.Value)
		if err != nil {
			return nil, err
		}
		return append(binary.LittleEndian.AppendUint32(nil, uint32(len(text))), text...), nil
	}
}

// fgbGeometry adds the FlatGeobuf geometry table of a shape to the builder, writing the type of the shape when
// `withType` is true or the shape has parts. Multipolygons and collections hold their members as parts, which always
// carry their types, and the other types hold their points, with the ends of their rings or linestrings when they
// have more than one.
func fgbGeometry(b *flatbuffers.Builder, s shape, withType bool) flatbuffers.UOffsetT {
	var parts []flatbuffers.UOffsetT
	var ends []uint32
	var xy, z, m []float64
	dims := s.ctype.Dimension()
	addSequence := func(seq []float64) {
		for i := 0; i+dims <= len(seq); i += dims {
			xy = append(xy, seq[i], seq[i+1])
			next := i + 2
			if s.hasZ() {
				z = append(z, seq[next])
				next++
			}
			if s.hasM() {
				m = append(m, seq[next])
			}
		}
	}
	switch s.typeNum {
	case pointType, lineType:
		addSequence(s.coords)
	case multiPointType:
		for _, member := range s.members {
			addSequence(member.coords)
		}
	case polygonType:
		for _, ring := range s.rings {
			addSequence(ring)
			ends = append(ends, uint32(len(xy)/2))
		}
	case multiLineType:
		for _, member := range s.members {
			addSequence(member.coords)
			ends = append(ends, uint32(len(xy)/2))
		}
	case multiPolygonType, collectionType:
		for _, member := range s.members {
			parts = append(parts, fgbGeometry(b, member, true))
		}
	}
	if len(ends) < 2 {
		ends = nil
	}
	doubles := func(vals []float64) flatbuffers.UOffsetT {
		if len(vals) == 0 {
			return 0
		}
		b.StartVector(8, len(vals), 8)
		for i := len(vals) - 1; i >= 0; i-- {
			b.PrependFloat64(vals[i])
		}
		return b.EndVector(len(vals))
	}
	var endsVector flatbuffers.UOffsetT
	if len(ends) > 0 {
		b.StartVector(4, len(ends), 4)
		for i := len(ends) - 1; i >= 0; i-- {
			b.PrependUint32(ends[i])
		}
		endsVector = b.EndVector(len(ends))
	}
	xyVector, zVector, mVector := doubles(xy), doubles(z), doubles(m)
	var partsVector flatbuffers.UOffsetT
	if len(parts) > 0 {
		b.StartVector(4, len(parts), 4)
		for i := len(parts) - 1; i >= 0; i-- {
			b.PrependUOffsetT(parts[i])
		}
		partsVector = b.EndVector(len(parts))
	}
	b.StartObject(8)
	if partsVector != 0 {
		b.PrependUOffsetTSlot(7, partsVector, 0)
	}
	if mVector != 0 {
		b.PrependUOffsetTSlot(3, mVector, 0)
	}
	if zVector != 0 {
		b.PrependUOffsetTSlot(2, zVector, 0)
	}
	if xyVector != 0 {
		b.PrependUOffsetTSlot(1, xyVector, 0)
	}
	if endsVector != 0 {
		b.PrependUOffsetTSlot(0, endsVector, 0)
	}
	if withType || len(parts) > 0 {
		b.PrependByteSlot(6, fgbGeometryTypes[s.typeNum], 0)
	}
	return b.EndObject()
}

// fgbFeatureBytes encodes a feature as a size-prefixed FlatGeobuf feature.
func fgbFeatureBytes(feature fgbFeature) []byte {
	b := flatbuffers.NewBuilder(0)
	b.Prep(8, 0)
	var geometry flatbuffers.UOffsetT
	if !feature.geometry.isEmpty() {
		geometry = fgbGeometry(b, feature.geometry, false)
	}
	var properties flatbuffers.UOffsetT
	if len(feature.properties) > 0 {
		properties = b.CreateByteVector(feature.properties)
	}
	b.StartObject(3)
	if geometry != 0 {
		b.PrependUOffsetTSlot(0, geometry, 0)
	}
	if properties != 0 {
		b.PrependUOffsetTSlot(1, properties, 0)
	}
	b.FinishSizePrefixed(b.EndObject())
	return b.FinishedBytes()
}

// fgbHilbert returns the position of a point on the Hilbert curve that fills a grid of 65536 by 65536 cells.
func fgbHilbert(x uint32, y uint32) uint32 {
	a := x ^ y
	b := 0xFFFF ^ a
	c := 0xFFFF ^ (x | y)
	d := x & (y ^ 0xFFFF)
	A := a | (b >> 1)
	B := (a >> 1) ^ a
	C := ((c >> 1) ^ (b & (d >> 1))) ^ c
	D := ((a & (c >> 1)) ^ (d >> 1)) ^ d
	for _, shift := range []uint32{2, 4, 8} {
		a, b, c, d = A, B, C, D
		A = (a & (a >> shift)) ^ (b & (b >> shift))
		B = (a & (b >> shift)) ^ (b & ((a ^ b) >> shift))
		C ^= (a & (c >> shift)) ^ (b & (d >> shift))
		D ^= (b & (c >> shift)) ^ ((a ^ b) & (d >> shift))
	}
	a = C ^ (C >> 1)
	b = D ^ (D >> 1)
	i0 := x ^ y
	i1 := b | (0xFFFF ^ (i0 | a))
	interleave := func(v uint32) uint32 {
		v = (v | (v << 8)) & 0x00FF00FF
		v = (v | (v << 4)) & 0x0F0F0F0F
		v = (v | (v << 2)) & 0x33333333
		v = (v | (v << 1)) & 0x55555555
		return v
	}
	return (interleave(i1) << 1) | interleave(i0)
}

// fgbFinal implements the final function of ST_AsFlatGeobuf, which writes the gathered rows as a FlatGeobuf file, or
// returns null when there are none. With an index, the features are written in descending order along a Hilbert curve
// over their boxes, after a packed R-tree of 16-entry nodes that leads to them.
func fgbFinal(ctx *sql.Context, args ...any) (any, error) {
	collection, _ := args[0].(*fgbCollection)
	if collection == nil {
		return nil, nil
	}
	features := collection.features
	var boxes []bbox
	extent := bbox{minX: math.Inf(1), minY: math.Inf(1), maxX: math.Inf(-1), maxY: math.Inf(-1)}
	indexed := collection.index && len(features) > 0
	if indexed {
		boxes = make([]bbox, len(features))
		for i, feature := range features {
			boxes[i], _ = feature.geometry.bounds()
			extent.minX, extent.minY = math.Min(extent.minX, boxes[i].minX), math.Min(extent.minY, boxes[i].minY)
			extent.maxX, extent.maxY = math.Max(extent.maxX, boxes[i].maxX), math.Max(extent.maxY, boxes[i].maxY)
		}
		width, height := extent.maxX-extent.minX, extent.maxY-extent.minY
		hilberts := make([]uint32, len(features))
		for i, box := range boxes {
			var x, y uint32
			if width != 0 {
				x = uint32(math.Floor(0xFFFF * ((box.minX+box.maxX)/2 - extent.minX) / width))
			}
			if height != 0 {
				y = uint32(math.Floor(0xFFFF * ((box.minY+box.maxY)/2 - extent.minY) / height))
			}
			hilberts[i] = fgbHilbert(x, y)
		}
		order := make([]int, len(features))
		for i := range order {
			order[i] = i
		}
		sort.SliceStable(order, func(i, j int) bool {
			return hilberts[order[i]] > hilberts[order[j]]
		})
		sortedFeatures, sortedBoxes := make([]fgbFeature, len(order)), make([]bbox, len(order))
		for i, original := range order {
			sortedFeatures[i], sortedBoxes[i] = features[original], boxes[original]
		}
		features, boxes = sortedFeatures, sortedBoxes
	}
	b := flatbuffers.NewBuilder(0)
	b.Prep(8, 0)
	var columnTables []flatbuffers.UOffsetT
	for _, column := range collection.columns {
		name := b.CreateString(column.name)
		b.StartObject(11)
		b.PrependUOffsetTSlot(0, name, 0)
		b.PrependByteSlot(1, column.columnType, 0)
		columnTables = append(columnTables, b.EndObject())
	}
	var crs flatbuffers.UOffsetT
	if collection.srid != sridUnknown {
		b.StartObject(6)
		b.PrependInt32Slot(1, collection.srid, 0)
		crs = b.EndObject()
	}
	var envelope flatbuffers.UOffsetT
	if indexed {
		b.StartVector(8, 4, 8)
		for _, v := range []float64{extent.maxY, extent.maxX, extent.minY, extent.minX} {
			b.PrependFloat64(v)
		}
		envelope = b.EndVector(4)
	}
	var columns flatbuffers.UOffsetT
	if len(columnTables) > 0 {
		b.StartVector(4, len(columnTables), 4)
		for i := len(columnTables) - 1; i >= 0; i-- {
			b.PrependUOffsetT(columnTables[i])
		}
		columns = b.EndVector(len(columnTables))
	}
	b.StartObject(14)
	if indexed {
		b.PrependUint64Slot(8, uint64(len(features)), 0)
	}
	if crs != 0 {
		b.PrependUOffsetTSlot(10, crs, 0)
	}
	if columns != 0 {
		b.PrependUOffsetTSlot(7, columns, 0)
	}
	if envelope != 0 {
		b.PrependUOffsetTSlot(1, envelope, 0)
	}
	if !indexed {
		b.PrependUint16Slot(9, 0, 16)
	}
	b.PrependBoolSlot(4, collection.hasM, false)
	b.PrependBoolSlot(3, collection.hasZ, false)
	b.PrependByteSlot(2, collection.geometryType, 0)
	b.FinishSizePrefixed(b.EndObject())
	file := append(append([]byte(nil), fgbMagic...), b.FinishedBytes()...)
	encodedFeatures := make([][]byte, len(features))
	for i, feature := range features {
		encodedFeatures[i] = fgbFeatureBytes(feature)
	}
	if indexed {
		file = append(file, fgbIndex(boxes, encodedFeatures)...)
	}
	for _, encoded := range encodedFeatures {
		file = append(file, encoded...)
	}
	return file, nil
}

// fgbLevelSizes returns the number of nodes on each level of a packed R-tree of 16-entry nodes over the given number
// of items, from the leaves up to the root.
func fgbLevelSizes(count int) []int {
	sizes := []int{count}
	for n := count; ; {
		n = (n + 15) / 16
		sizes = append(sizes, n)
		if n == 1 {
			return sizes
		}
	}
}

// fgbIndex returns the packed R-tree over the boxes of the encoded features, in their order. Each node holds a box
// and, for a leaf, the byte offset of its feature, or for any other node, the index of its first child. The root's
// level comes first.
func fgbIndex(boxes []bbox, features [][]byte) []byte {
	sizes := fgbLevelSizes(len(boxes))
	starts := make([]int, len(sizes))
	total := 0
	for level := len(sizes) - 1; level >= 0; level-- {
		starts[level] = total
		total += sizes[level]
	}
	nodes := make([]bbox, total)
	offsets := make([]uint64, total)
	offset := uint64(0)
	for i, box := range boxes {
		nodes[starts[0]+i], offsets[starts[0]+i] = box, offset
		offset += uint64(len(features[i]))
	}
	for level := 1; level < len(sizes); level++ {
		for i := 0; i < sizes[level]; i++ {
			first := starts[level-1] + i*16
			box := bbox{minX: math.Inf(1), minY: math.Inf(1), maxX: math.Inf(-1), maxY: math.Inf(-1)}
			for child := first; child < first+16 && child < starts[level-1]+sizes[level-1]; child++ {
				box.minX, box.minY = math.Min(box.minX, nodes[child].minX), math.Min(box.minY, nodes[child].minY)
				box.maxX, box.maxY = math.Max(box.maxX, nodes[child].maxX), math.Max(box.maxY, nodes[child].maxY)
			}
			nodes[starts[level]+i], offsets[starts[level]+i] = box, uint64(first)
		}
	}
	var index []byte
	for i, node := range nodes {
		for _, v := range []float64{node.minX, node.minY, node.maxX, node.maxY} {
			index = binary.LittleEndian.AppendUint64(index, math.Float64bits(v))
		}
		index = binary.LittleEndian.AppendUint64(index, offsets[i])
	}
	return index
}

// fgbReader reads a FlatGeobuf file.
type fgbReader struct {
	data         []byte
	columns      []fgbColumn
	geometryType byte
	hasZ, hasM   bool
	srid         int32
	position     int
}

// newFGBReader reads the header of a FlatGeobuf file and skips its index, leaving the reader at the first feature.
func newFGBReader(data []byte) (*fgbReader, error) {
	if len(data) < len(fgbMagic)+4 || !bytes.Equal(data[:3], fgbMagic[:3]) {
		return nil, pgerror.New(pgcode.Internal, "Data is not FlatGeobuf")
	}
	size := int(binary.LittleEndian.Uint32(data[8:]))
	if 12+size > len(data) {
		return nil, pgerror.New(pgcode.Internal, "Data is not FlatGeobuf")
	}
	header := data[12 : 12+size]
	r := &fgbReader{data: data, position: 12 + size}
	t := flatbuffers.Table{Bytes: header, Pos: flatbuffers.GetUOffsetT(header)}
	if o := flatbuffers.UOffsetT(t.Offset(8)); o != 0 {
		r.geometryType = t.GetByte(o + t.Pos)
	}
	if o := flatbuffers.UOffsetT(t.Offset(10)); o != 0 {
		r.hasZ = t.GetBool(o + t.Pos)
	}
	if o := flatbuffers.UOffsetT(t.Offset(12)); o != 0 {
		r.hasM = t.GetBool(o + t.Pos)
	}
	if o := flatbuffers.UOffsetT(t.Offset(18)); o != 0 {
		vector := t.Vector(o)
		for i := 0; i < t.VectorLen(o); i++ {
			column := flatbuffers.Table{Bytes: header, Pos: t.Indirect(vector + flatbuffers.UOffsetT(i*4))}
			var c fgbColumn
			if co := flatbuffers.UOffsetT(column.Offset(4)); co != 0 {
				c.name = column.String(co + column.Pos)
			}
			if co := flatbuffers.UOffsetT(column.Offset(6)); co != 0 {
				c.columnType = column.GetByte(co + column.Pos)
			}
			r.columns = append(r.columns, c)
		}
	}
	featureCount := uint64(0)
	if o := flatbuffers.UOffsetT(t.Offset(20)); o != 0 {
		featureCount = t.GetUint64(o + t.Pos)
	}
	nodeSize := uint16(16)
	if o := flatbuffers.UOffsetT(t.Offset(22)); o != 0 {
		nodeSize = t.GetUint16(o + t.Pos)
	}
	if o := flatbuffers.UOffsetT(t.Offset(24)); o != 0 {
		crs := flatbuffers.Table{Bytes: header, Pos: t.Indirect(o + t.Pos)}
		if co := flatbuffers.UOffsetT(crs.Offset(6)); co != 0 {
			r.srid = crs.GetInt32(co + crs.Pos)
		}
	}
	if nodeSize > 0 && featureCount > 0 {
		nodes := 0
		for _, size := range fgbLevelSizesFor(int(featureCount), int(nodeSize)) {
			nodes += size
		}
		r.position += nodes * 40
	}
	return r, nil
}

// fgbLevelSizesFor returns the number of nodes on each level of a packed R-tree with nodes of the given size.
func fgbLevelSizesFor(count int, nodeSize int) []int {
	sizes := []int{count}
	for n := count; ; {
		n = (n + nodeSize - 1) / nodeSize
		sizes = append(sizes, n)
		if n == 1 {
			return sizes
		}
	}
}

// next returns the geometry and the property values of the next feature, or false when none remain. Values that a
// feature leaves out are nil.
func (r *fgbReader) next(ctx *sql.Context) (Geometry, []any, bool, error) {
	if r.position+4 > len(r.data) {
		return Geometry{}, nil, false, nil
	}
	size := int(binary.LittleEndian.Uint32(r.data[r.position:]))
	if r.position+4+size > len(r.data) {
		return Geometry{}, nil, false, pgerror.New(pgcode.Internal, "Data is not FlatGeobuf")
	}
	buf := r.data[r.position+4 : r.position+4+size]
	r.position += 4 + size
	t := flatbuffers.Table{Bytes: buf, Pos: flatbuffers.GetUOffsetT(buf)}
	ctype := dimXY
	if r.hasZ && r.hasM {
		ctype = dimXYZM
	} else if r.hasZ {
		ctype = dimXYZ
	} else if r.hasM {
		ctype = dimXYM
	}
	g := Geometry{SRID: r.srid, Shape: shape{typeNum: map[byte]int{1: pointType, 2: lineType, 3: polygonType, 4: multiPointType, 5: multiLineType, 6: multiPolygonType, 7: collectionType}[r.geometryType], ctype: ctype}}
	if o := flatbuffers.UOffsetT(t.Offset(4)); o != 0 {
		g.Shape = fgbReadGeometry(flatbuffers.Table{Bytes: buf, Pos: t.Indirect(o + t.Pos)}, r.geometryType, ctype)
	}
	values := make([]any, len(r.columns))
	if o := flatbuffers.UOffsetT(t.Offset(6)); o != 0 {
		properties := t.ByteVector(o + t.Pos)
		for p := 0; p+2 <= len(properties); {
			column := int(binary.LittleEndian.Uint16(properties[p:]))
			p += 2
			if column >= len(r.columns) {
				break
			}
			var value any
			switch r.columns[column].columnType {
			case fgbBool:
				value, p = properties[p] != 0, p+1
			case fgbShort:
				value, p = int16(binary.LittleEndian.Uint16(properties[p:])), p+2
			case fgbInt:
				value, p = int32(binary.LittleEndian.Uint32(properties[p:])), p+4
			case fgbLong:
				value, p = int64(binary.LittleEndian.Uint64(properties[p:])), p+8
			case fgbFloat:
				value, p = math.Float32frombits(binary.LittleEndian.Uint32(properties[p:])), p+4
			case fgbDouble:
				value, p = math.Float64frombits(binary.LittleEndian.Uint64(properties[p:])), p+8
			default:
				length := int(binary.LittleEndian.Uint32(properties[p:]))
				value, p = properties[p+4:p+4+length], p+4+length
			}
			values[column] = value
		}
	}
	return g, values, true, nil
}

// fgbReadGeometry reads a FlatGeobuf geometry table of the given geometry type.
func fgbReadGeometry(t flatbuffers.Table, geometryType byte, ctype coordinatesType) shape {
	if o := flatbuffers.UOffsetT(t.Offset(16)); o != 0 {
		geometryType = t.GetByte(o + t.Pos)
	}
	doubles := func(slot flatbuffers.VOffsetT) []float64 {
		o := flatbuffers.UOffsetT(t.Offset(slot))
		if o == 0 {
			return nil
		}
		vector := t.Vector(o)
		vals := make([]float64, t.VectorLen(o))
		for i := range vals {
			vals[i] = t.GetFloat64(vector + flatbuffers.UOffsetT(i*8))
		}
		return vals
	}
	var ends []int
	if o := flatbuffers.UOffsetT(t.Offset(4)); o != 0 {
		vector := t.Vector(o)
		for i := 0; i < t.VectorLen(o); i++ {
			ends = append(ends, int(t.GetUint32(vector+flatbuffers.UOffsetT(i*4))))
		}
	}
	xy, z, m := doubles(6), doubles(8), doubles(10)
	sequence := func(from int, to int) []float64 {
		var seq []float64
		for i := from; i < to; i++ {
			seq = append(seq, xy[2*i], xy[2*i+1])
			if ctype.Is3D() && i < len(z) {
				seq = append(seq, z[i])
			}
			if ctype.IsMeasured() && i < len(m) {
				seq = append(seq, m[i])
			}
		}
		return seq
	}
	count := len(xy) / 2
	if len(ends) == 0 {
		ends = []int{count}
	}
	s := shape{typeNum: map[byte]int{1: pointType, 2: lineType, 3: polygonType, 4: multiPointType, 5: multiLineType, 6: multiPolygonType, 7: collectionType}[geometryType], ctype: ctype}
	switch s.typeNum {
	case pointType, lineType:
		s.coords = sequence(0, count)
	case multiPointType:
		for i := 0; i < count; i++ {
			s.members = append(s.members, shape{typeNum: pointType, ctype: ctype, coords: sequence(i, i+1)})
		}
	case polygonType, multiLineType:
		start := 0
		for _, end := range ends {
			if s.typeNum == polygonType {
				s.rings = append(s.rings, sequence(start, end))
			} else {
				s.members = append(s.members, shape{typeNum: lineType, ctype: ctype, coords: sequence(start, end)})
			}
			start = end
		}
	case multiPolygonType, collectionType:
		if o := flatbuffers.UOffsetT(t.Offset(18)); o != 0 {
			vector := t.Vector(o)
			for i := 0; i < t.VectorLen(o); i++ {
				part := flatbuffers.Table{Bytes: t.Bytes, Pos: t.Indirect(vector + flatbuffers.UOffsetT(i*4))}
				s.members = append(s.members, fgbReadGeometry(part, map[int]byte{multiPolygonType: 3, collectionType: 0}[s.typeNum], ctype))
			}
		}
	}
	return s
}

// fgbCompatible reports whether a column of a FlatGeobuf column type can be read into a column of a PostgreSQL type.
func fgbCompatible(columnType byte, t *pgtypes.DoltgresType) bool {
	switch columnType {
	case fgbBool:
		return t.ID == pgtypes.Bool.ID
	case fgbShort, fgbInt, fgbLong:
		return t.ID == pgtypes.Int16.ID || t.ID == pgtypes.Int32.ID || t.ID == pgtypes.Int64.ID || t.ID == pgtypes.Numeric.ID
	case fgbFloat, fgbDouble:
		return t.ID == pgtypes.Float32.ID || t.ID == pgtypes.Float64.ID || t.ID == pgtypes.Numeric.ID
	case fgbBinary:
		return t.ID == pgtypes.Bytea.ID
	default:
		return t.ID == pgtypes.Text.ID || t.ID == pgtypes.VarChar.ID || t.ID == pgtypes.JsonB.ID || t.ID == pgtypes.Date.ID ||
			t.ID == pgtypes.Timestamp.ID || t.ID == pgtypes.TimestampTZ.ID
	}
}

// stFromFlatGeobuf implements ST_FromFlatGeobuf, which returns a row of the type of its first argument for each
// feature of a FlatGeobuf file: the feature's position from zero, its geometry, and then its properties, which must
// match the type's remaining columns in number and kind.
func stFromFlatGeobuf(ctx *sql.Context, types []*pgtypes.DoltgresType, args ...any) (any, error) {
	rowType := types[len(types)-1]
	data, _ := args[1].([]byte)
	if len(data) == 0 {
		return sql.RowsToRowIter(), nil
	}
	reader, err := newFGBReader(data)
	if err != nil {
		return nil, err
	}
	attrs := rowType.CompositeAttrs
	if len(attrs)-2 != len(reader.columns) {
		return nil, pgerror.Newf(pgcode.DatatypeMismatch, "flatgeobuf: column count mismatch: file has %d columns, target type has %d", len(reader.columns), len(attrs)-2)
	}
	attrTypes := make([]*pgtypes.DoltgresType, len(attrs))
	for i, attr := range attrs {
		attrTypes[i] = attr.Type
	}
	for i, column := range reader.columns {
		if !fgbCompatible(column.columnType, attrTypes[i+2]) {
			return nil, pgerror.Newf(pgcode.DatatypeMismatch, "flatgeobuf: column \"%s\" type mismatch: file type \"%s\" is not compatible with PostgreSQL type %s",
				column.name, fgbColumnTypeNames[column.columnType], attrTypes[i+2].String())
		}
	}
	var rows []sql.Row
	for position := int32(0); ; position++ {
		g, values, ok, err := reader.next(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		record := make([]pgtypes.RecordValue, len(attrs))
		record[0] = pgtypes.RecordValue{Value: position, Type: attrTypes[0]}
		record[1] = pgtypes.RecordValue{Type: attrTypes[1]}
		if !g.Shape.isEmpty() {
			record[1].Value = g
		}
		for i, value := range values {
			converted, err := fgbConvert(ctx, value, attrTypes[i+2])
			if err != nil {
				return nil, err
			}
			record[i+2] = pgtypes.RecordValue{Value: converted, Type: attrTypes[i+2]}
		}
		rows = append(rows, sql.Row{record})
	}
	return sql.RowsToRowIter(rows...), nil
}

// fgbConvert converts a property value read from a FlatGeobuf file to the value of a column of the given type.
func fgbConvert(ctx *sql.Context, value any, t *pgtypes.DoltgresType) (any, error) {
	switch value := value.(type) {
	case nil:
		return nil, nil
	case []byte:
		if t.ID == pgtypes.Bytea.ID {
			return value, nil
		}
		return t.IoInput(ctx, string(value))
	case bool:
		return value, nil
	default:
		return t.IoInput(ctx, fmt.Sprint(value))
	}
}

// stFromFlatGeobufToTable implements ST_FromFlatGeobufToTable, which creates a table with an integer id, a geometry,
// and a column for each column of a FlatGeobuf file, without adding any rows.
func stFromFlatGeobufToTable(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	reader, err := newFGBReader(args[2].([]byte))
	if err != nil {
		return nil, err
	}
	definitions := []string{"id int", "geom geometry"}
	for _, column := range reader.columns {
		definitions = append(definitions, hook.QuoteIdentifier(column.name)+" "+fgbTableTypes[column.columnType])
	}
	statement := fmt.Sprintf("CREATE TABLE %s.%s (%s)", hook.QuoteIdentifier(args[0].(string)), hook.QuoteIdentifier(args[1].(string)), strings.Join(definitions, ", "))
	if err = hook.RunStatement(ctx, runner, statement); err != nil {
		return nil, err
	}
	return "", nil
}

// fgbRoutines returns the routines that write and read FlatGeobuf files.
func fgbRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "pgis_asflatgeobuf_transfn", Symbol: "pgis_asflatgeobuf_transfn_internal_anyelement", Parameters: params("internal", "anyelement"), Returns: "internal", TypedImpl: fgbTransition},
		{Name: "pgis_asflatgeobuf_transfn", Symbol: "pgis_asflatgeobuf_transfn_internal_anyelement_bool", Parameters: params("internal", "anyelement", "bool"), Returns: "internal", TypedImpl: fgbTransition},
		{Name: "pgis_asflatgeobuf_transfn", Symbol: "pgis_asflatgeobuf_transfn", Parameters: params("internal", "anyelement", "bool", "text"), Returns: "internal", TypedImpl: fgbTransition},
		{Name: "pgis_asflatgeobuf_finalfn", Symbol: "pgis_asflatgeobuf_finalfn", Parameters: params("internal"), Returns: "bytea", Impl: fgbFinal},
		{Name: "st_fromflatgeobuf", Symbol: "pgis_fromflatgeobuf", Parameters: params("anyelement", "bytea"), Returns: "anyelement", SetOf: true, TypedImpl: stFromFlatGeobuf},
		{Name: "st_fromflatgeobuftotable", Symbol: "pgis_tablefromflatgeobuf", Parameters: params("text", "text", "bytea"), Returns: "void", Strict: true, QueryImpl: stFromFlatGeobufToTable},
	}
}

// fgbAggregates returns the aggregate that writes FlatGeobuf files.
func fgbAggregates() []extdef.Aggregate {
	return []extdef.Aggregate{
		{Name: "st_asflatgeobuf", Parameters: params("anyelement"), Returns: "bytea", StateType: "internal", Transition: "pgis_asflatgeobuf_transfn_internal_anyelement", Final: "pgis_asflatgeobuf_finalfn"},
		{Name: "st_asflatgeobuf", Parameters: params("anyelement", "bool"), Returns: "bytea", StateType: "internal", Transition: "pgis_asflatgeobuf_transfn_internal_anyelement_bool", Final: "pgis_asflatgeobuf_finalfn"},
		{Name: "st_asflatgeobuf", Parameters: params("anyelement", "bool", "text"), Returns: "bytea", StateType: "internal", Transition: "pgis_asflatgeobuf_transfn", Final: "pgis_asflatgeobuf_finalfn"},
	}
}
