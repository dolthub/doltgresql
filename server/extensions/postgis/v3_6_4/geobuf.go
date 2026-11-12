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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// geobufFeature is a row that ST_AsGeobuf gathers: its geometry, and its properties as the indexes of their keys and
// their encoded values.
type geobufFeature struct {
	geometry shape
	keys     []int
	values   [][]byte
}

// geobufCollection is the state of ST_AsGeobuf: the name of the geometry column, the keys of the other columns, and
// the features gathered so far.
type geobufCollection struct {
	geomName   string
	geomColumn int
	resolved   bool
	keys       []string
	features   []geobufFeature
}

// geobufTransition implements the transition functions of ST_AsGeobuf, which gather each row that has a geometry.
// The row's other columns that are not null become the feature's properties.
func geobufTransition(ctx *sql.Context, types []*pgtypes.DoltgresType, args ...any) (any, error) {
	collection, _ := args[0].(*geobufCollection)
	if collection == nil {
		collection = &geobufCollection{}
		if len(args) > 2 && args[2] != nil {
			collection.geomName = args[2].(string)
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
	if !collection.resolved {
		collection.geomColumn = -1
		for i, value := range values {
			if t, ok := value.Type.(*pgtypes.DoltgresType); ok && t.Name() == "geometry" && collection.geomColumn < 0 && (collection.geomName == "" || names[i] == collection.geomName) {
				collection.geomColumn = i
			}
		}
		if collection.geomColumn < 0 {
			return nil, pgerror.New(pgcode.Internal, "encode_keys: no geometry column found")
		}
		for i := range values {
			if i != collection.geomColumn {
				collection.keys = append(collection.keys, names[i])
			}
		}
		collection.resolved = true
	}
	geometry, ok := values[collection.geomColumn].Value.(Geometry)
	if !ok {
		return collection, nil
	}
	feature := geobufFeature{geometry: geometry.Shape}
	key := 0
	for i, field := range values {
		if i == collection.geomColumn {
			continue
		}
		key++
		if field.Value == nil {
			continue
		}
		var encoded []byte
		t, _ := field.Type.(*pgtypes.DoltgresType)
		if integer, ok := integerValue(field.Value); ok && t != nil && isIntegerType(t) {
			feature.keys = append(feature.keys, key-1)
			feature.values = append(feature.values, geobufIntegerValue(integer))
			continue
		}
		switch value := field.Value.(type) {
		case float32:
			encoded = binary.LittleEndian.AppendUint64(appendField(nil, 2, 1), math.Float64bits(float64(value)))
		case float64:
			encoded = binary.LittleEndian.AppendUint64(appendField(nil, 2, 1), math.Float64bits(value))
		case bool:
			encoded = appendBytesField(nil, 1, []byte(map[bool]string{false: "f", true: "t"}[value]))
		default:
			if t == nil {
				continue
			}
			text, err := t.IoOutput(ctx, value)
			if err != nil {
				return nil, err
			}
			encoded = appendBytesField(nil, 1, []byte(text))
		}
		feature.keys = append(feature.keys, key-1)
		feature.values = append(feature.values, encoded)
	}
	collection.features = append(collection.features, feature)
	return collection, nil
}

// geobufIntegerValue encodes an integer as a Geobuf value, as its magnitude with a field for its sign.
func geobufIntegerValue(v int64) []byte {
	if v >= 0 {
		return appendVarint(appendField(nil, 3, 0), uint64(v))
	}
	return appendVarint(appendField(nil, 4, 0), uint64(-v))
}

// geobufPrecision returns the number of decimal places that the coordinates of the features are written with. Each
// ordinate in turn raises it by one place, up to six, when rounding it to the places reached before it changes its
// first six significant digits.
func geobufPrecision(features []geobufFeature) int {
	precision := 0
	for _, feature := range features {
		feature.geometry.mapCoords(func(point []float64) []float64 {
			for _, ordinate := range point {
				scale := math.Pow(10, float64(precision))
				if precision < 6 && formatG(math.Round(ordinate*scale)/scale, 6) != formatG(ordinate, 6) {
					precision++
				}
			}
			return point
		})
	}
	return precision
}

// geobufWriter writes Geobuf geometries with a fixed number of dimensions and decimal places.
type geobufWriter struct {
	dims  int
	scale float64
}

// sequence appends the ordinates of the points of a sequence to `coords`, each scaled up and rounded up to an integer
// and written as its difference from the point before it, leaving out the closing point of a ring. It returns the
// number of points written.
func (w geobufWriter) sequence(coords []uint64, seq []float64, sourceDims int, closed bool) ([]uint64, int) {
	count := len(seq) / sourceDims
	if closed {
		count--
	}
	last := make([]int64, w.dims)
	for i := 0; i < count; i++ {
		for d := 0; d < w.dims; d++ {
			value := int64(0)
			if d < sourceDims {
				value = int64(math.Ceil(seq[i*sourceDims+d] * w.scale))
			}
			coords = append(coords, zigzag(value-last[d]))
			last[d] = value
		}
	}
	return coords, count
}

// geometry encodes a shape as a Geobuf geometry message.
func (w geobufWriter) geometry(s shape) []byte {
	sourceDims := s.ctype.Dimension()
	var lengths, coords []uint64
	var children [][]byte
	geomType := map[int]uint64{pointType: 0, multiPointType: 1, lineType: 2, multiLineType: 3, polygonType: 4, multiPolygonType: 5, collectionType: 6}[s.typeNum]
	switch s.typeNum {
	case pointType, lineType:
		coords, _ = w.sequence(coords, s.coords, sourceDims, false)
	case multiPointType:
		var points []float64
		for _, member := range s.members {
			points = append(points, member.coords...)
		}
		coords, _ = w.sequence(coords, points, sourceDims, false)
	case multiLineType:
		for _, member := range s.members {
			var count int
			coords, count = w.sequence(coords, member.coords, sourceDims, false)
			lengths = append(lengths, uint64(count))
		}
	case polygonType:
		for _, ring := range s.rings {
			var count int
			coords, count = w.sequence(coords, ring, sourceDims, true)
			lengths = append(lengths, uint64(count))
		}
		if len(s.rings) == 1 {
			lengths = nil
		}
	case multiPolygonType:
		lengths = append(lengths, uint64(len(s.members)))
		for _, member := range s.members {
			lengths = append(lengths, uint64(len(member.rings)))
			for _, ring := range member.rings {
				var count int
				coords, count = w.sequence(coords, ring, sourceDims, true)
				lengths = append(lengths, uint64(count))
			}
		}
	case collectionType:
		for _, member := range s.members {
			children = append(children, w.geometry(member))
		}
	}
	encoded := appendVarint(appendField(nil, 1, 0), geomType)
	packed := func(field int, vals []uint64) {
		if len(vals) == 0 {
			return
		}
		var buf []byte
		for _, v := range vals {
			buf = appendVarint(buf, v)
		}
		encoded = appendBytesField(encoded, field, buf)
	}
	packed(2, lengths)
	packed(3, coords)
	for _, child := range children {
		encoded = appendBytesField(encoded, 4, child)
	}
	return encoded
}

// geobufFinal implements the final function of ST_AsGeobuf, which encodes the gathered features as a Geobuf feature
// collection, or returns null when there are none.
func geobufFinal(ctx *sql.Context, args ...any) (any, error) {
	collection, _ := args[0].(*geobufCollection)
	if collection == nil {
		return nil, nil
	}
	dims := 2
	for _, feature := range collection.features {
		if feature.geometry.hasZ() {
			dims = 3
		}
	}
	precision := geobufPrecision(collection.features)
	writer := geobufWriter{dims: dims, scale: math.Pow(10, float64(precision))}
	var features []byte
	for _, feature := range collection.features {
		encoded := appendBytesField(nil, 1, writer.geometry(feature.geometry))
		for _, value := range feature.values {
			encoded = appendBytesField(encoded, 13, value)
		}
		if len(feature.values) > 0 {
			var properties []byte
			for i, key := range feature.keys {
				properties = appendVarint(appendVarint(properties, uint64(key)), uint64(i))
			}
			encoded = appendBytesField(encoded, 14, properties)
		}
		features = appendBytesField(features, 1, encoded)
	}
	var data []byte
	for _, key := range collection.keys {
		data = appendBytesField(data, 1, []byte(key))
	}
	if dims != 2 {
		data = appendVarint(appendField(data, 2, 0), uint64(dims))
	}
	if precision != 6 {
		data = appendVarint(appendField(data, 3, 0), uint64(precision))
	}
	return appendBytesField(data, 4, features), nil
}

// geobufRoutines returns the routines that encode Geobuf documents.
func geobufRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "pgis_asgeobuf_transfn", Symbol: "pgis_asgeobuf_transfn_internal_anyelement", Parameters: params("internal", "anyelement"), Returns: "internal", TypedImpl: geobufTransition},
		{Name: "pgis_asgeobuf_transfn", Symbol: "pgis_asgeobuf_transfn", Parameters: params("internal", "anyelement", "text"), Returns: "internal", TypedImpl: geobufTransition},
		{Name: "pgis_asgeobuf_finalfn", Symbol: "pgis_asgeobuf_finalfn", Parameters: params("internal"), Returns: "bytea", Impl: geobufFinal},
	}
}

// geobufAggregates returns the aggregate that encodes Geobuf documents.
func geobufAggregates() []extdef.Aggregate {
	return []extdef.Aggregate{
		{Name: "st_asgeobuf", Parameters: params("anyelement"), Returns: "bytea", StateType: "internal", Transition: "pgis_asgeobuf_transfn_internal_anyelement", Final: "pgis_asgeobuf_finalfn"},
		{Name: "st_asgeobuf", Parameters: params("anyelement", "text"), Returns: "bytea", StateType: "internal", Transition: "pgis_asgeobuf_transfn", Final: "pgis_asgeobuf_finalfn"},
	}
}
