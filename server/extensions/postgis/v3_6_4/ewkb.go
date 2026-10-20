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
)

// The flags that extended WKB sets on a geometry's type number.
const (
	ewkbZFlag    = 0x80000000
	ewkbMFlag    = 0x40000000
	ewkbSRIDFlag = 0x20000000
)

// quietNaNBits is the NaN that marks the ordinates of an empty point.
const quietNaNBits = 0x7FF8000000000000

// wkbVariant selects how a geometry's type number, dimensions, and SRID are written to WKB.
type wkbVariant uint8

const (
	// wkbExtended writes PostGIS' extended WKB, which flags dimensions and carries the SRID.
	wkbExtended wkbVariant = iota
	// wkbISO writes ISO WKB, which offsets the type number by the dimensions and omits the SRID.
	wkbISO
)

// wkbTypeNumbers holds the WKB type number of each type whose number differs from PostGIS' own.
var wkbTypeNumbers = map[int]int{
	polyhedralSurfaceType: 15,
	triangleType:          17,
	tinType:               16,
}

// fromWKBTypeNumbers holds PostGIS' type number for each WKB type number that differs from it.
var fromWKBTypeNumbers = map[int]int{
	15: polyhedralSurfaceType,
	16: tinType,
	17: triangleType,
}

// allowedMembers holds the member types that each collection type accepts. A geometry collection accepts any type.
var allowedMembers = map[int][]int{
	multiPointType:        {pointType},
	multiLineType:         {lineType},
	multiPolygonType:      {polygonType},
	compoundType:          {lineType, circStringType},
	curvePolyType:         {lineType, circStringType, compoundType},
	multiCurveType:        {lineType, circStringType, compoundType},
	multiSurfaceType:      {polygonType, curvePolyType},
	polyhedralSurfaceType: {polygonType},
	tinType:               {triangleType},
}

// wkbReader decodes WKB, keeping track of the current position.
type wkbReader struct {
	data []byte
	pos  int
}

// errWKBTruncated is returned when WKB ends before the geometry does.
var errWKBTruncated = pgerror.New(pgcode.Internal, "WKB structure does not match expected size!")

// decodeWKB decodes extended or ISO WKB into a geometry.
func decodeWKB(ctx *sql.Context, data []byte) (Geometry, error) {
	r := &wkbReader{data: data}
	s, srid, err := r.readGeometry()
	if err != nil {
		return Geometry{}, err
	}
	return Geometry{SRID: clampSRID(ctx, srid), Shape: s}, nil
}

// readGeometry reads one geometry, along with the SRID found in its header.
func (r *wkbReader) readGeometry() (shape, int32, error) {
	if r.pos >= len(r.data) {
		return shape{}, 0, errWKBTruncated
	}
	var order binary.ByteOrder
	switch r.data[r.pos] {
	case 0:
		order = binary.BigEndian
	case 1:
		order = binary.LittleEndian
	default:
		return shape{}, 0, pgerror.New(pgcode.Internal, "Invalid endian flag value encountered.")
	}
	r.pos++
	typeWord, err := r.readUint32(order)
	if err != nil {
		return shape{}, 0, err
	}
	ctype := dimXY
	if typeWord&ewkbZFlag != 0 {
		ctype |= dimXYZ
	}
	if typeWord&ewkbMFlag != 0 {
		ctype |= dimXYM
	}
	var srid int32
	if typeWord&ewkbSRIDFlag != 0 {
		sridWord, err := r.readUint32(order)
		if err != nil {
			return shape{}, 0, err
		}
		srid = int32(sridWord)
	}
	wkbType := int(typeWord & 0x0FFFFFFF)
	switch wkbType / 1000 {
	case 1:
		ctype |= dimXYZ
	case 2:
		ctype |= dimXYM
	case 3:
		ctype |= dimXYZM
	}
	typeNum := wkbType % 1000
	if internal, ok := fromWKBTypeNumbers[typeNum]; ok {
		typeNum = internal
	} else if typeNum < pointType || typeNum > multiSurfaceType {
		return shape{}, 0, pgerror.Newf(pgcode.Internal, "Unknown WKB type (%d)! Full WKB type number was (%d).", typeNum, typeWord)
	}
	s, err := r.readBody(order, typeNum, ctype)
	return s, srid, err
}

// readBody reads the body of a geometry whose header has already been read.
func (r *wkbReader) readBody(order binary.ByteOrder, typeNum int, ctype coordinatesType) (shape, error) {
	s := shape{typeNum: typeNum, ctype: ctype}
	var err error
	switch typeNum {
	case pointType:
		if s.coords, err = r.readFloats(order, ctype.Dimension()); err != nil {
			return shape{}, err
		}
		if math.IsNaN(s.coords[0]) && math.IsNaN(s.coords[1]) {
			s.coords = nil
		}
	case lineType, circStringType:
		s.coords, err = r.readSequence(order, ctype)
	case triangleType:
		var rings [][]float64
		if rings, err = r.readRings(order, ctype); err == nil && len(rings) > 0 {
			s.coords = rings[0]
		}
	case polygonType:
		s.rings, err = r.readRings(order, ctype)
	default:
		var count uint32
		if count, err = r.readUint32(order); err != nil {
			return shape{}, err
		}
		s.members = make([]shape, 0, min(count, 1024))
		for i := uint32(0); i < count; i++ {
			member, _, err := r.readGeometry()
			if err != nil {
				return shape{}, err
			}
			if !memberAllowed(typeNum, member.typeNum) {
				return shape{}, pgerror.Newf(pgcode.Internal, "Invalid subtype (%s) for collection type (%s)",
					typeNameForNumber(member.typeNum), typeNameForNumber(typeNum))
			}
			s.members = append(s.members, member.withDims(ctype))
		}
	}
	return s, err
}

// memberAllowed returns whether a collection of the given type may hold a member of the given type.
func memberAllowed(collectionTypeNum int, memberTypeNum int) bool {
	allowed, ok := allowedMembers[collectionTypeNum]
	if !ok {
		return true
	}
	for _, typeNum := range allowed {
		if typeNum == memberTypeNum {
			return true
		}
	}
	return false
}

// readRings reads a count of rings followed by each ring's sequence.
func (r *wkbReader) readRings(order binary.ByteOrder, ctype coordinatesType) ([][]float64, error) {
	count, err := r.readUint32(order)
	if err != nil {
		return nil, err
	}
	rings := make([][]float64, 0, min(count, 1024))
	for i := uint32(0); i < count; i++ {
		ring, err := r.readSequence(order, ctype)
		if err != nil {
			return nil, err
		}
		rings = append(rings, ring)
	}
	return rings, nil
}

// readSequence reads a count of points followed by their ordinates.
func (r *wkbReader) readSequence(order binary.ByteOrder, ctype coordinatesType) ([]float64, error) {
	count, err := r.readUint32(order)
	if err != nil {
		return nil, err
	}
	if int(count) > (len(r.data)-r.pos)/8 {
		return nil, errWKBTruncated
	}
	return r.readFloats(order, int(count)*ctype.Dimension())
}

// readUint32 reads a single unsigned 32-bit integer.
func (r *wkbReader) readUint32(order binary.ByteOrder) (uint32, error) {
	if r.pos+4 > len(r.data) {
		return 0, errWKBTruncated
	}
	val := order.Uint32(r.data[r.pos:])
	r.pos += 4
	return val, nil
}

// readFloats reads the given number of 64-bit floats.
func (r *wkbReader) readFloats(order binary.ByteOrder, count int) ([]float64, error) {
	if r.pos+count*8 > len(r.data) {
		return nil, errWKBTruncated
	}
	vals := make([]float64, count)
	for i := range vals {
		vals[i] = math.Float64frombits(order.Uint64(r.data[r.pos:]))
		r.pos += 8
	}
	return vals, nil
}

// encodeWKB encodes the geometry as WKB of the given variant and byte order.
func encodeWKB(g Geometry, order binary.AppendByteOrder, variant wkbVariant) []byte {
	return appendWKB(nil, g.Shape, g.SRID, order, variant)
}

// appendWKB appends the WKB of the shape to `dst`. The SRID is written only when it is known and the variant is
// extended, and members are always written without one.
func appendWKB(dst []byte, s shape, srid int32, order binary.AppendByteOrder, variant wkbVariant) []byte {
	typeWord := uint32(s.typeNum)
	if wkbType, ok := wkbTypeNumbers[s.typeNum]; ok {
		typeWord = uint32(wkbType)
	}
	if variant == wkbExtended {
		if s.hasZ() {
			typeWord |= ewkbZFlag
		}
		if s.hasM() {
			typeWord |= ewkbMFlag
		}
		if srid != sridUnknown {
			typeWord |= ewkbSRIDFlag
		}
	} else {
		typeWord += uint32(s.ctype) * 1000
	}
	if order == binary.BigEndian {
		dst = append(dst, 0)
	} else {
		dst = append(dst, 1)
	}
	dst = order.AppendUint32(dst, typeWord)
	if variant == wkbExtended && srid != sridUnknown {
		dst = order.AppendUint32(dst, uint32(srid))
	}
	switch s.typeNum {
	case pointType:
		if len(s.coords) == 0 {
			for i := 0; i < s.ctype.Dimension(); i++ {
				dst = order.AppendUint64(dst, quietNaNBits)
			}
			return dst
		}
		return appendFloats(dst, s.coords, order)
	case lineType, circStringType:
		dst = order.AppendUint32(dst, uint32(len(s.coords)/s.ctype.Dimension()))
		return appendFloats(dst, s.coords, order)
	case triangleType:
		if len(s.coords) == 0 {
			return order.AppendUint32(dst, 0)
		}
		dst = order.AppendUint32(dst, 1)
		dst = order.AppendUint32(dst, uint32(len(s.coords)/s.ctype.Dimension()))
		return appendFloats(dst, s.coords, order)
	case polygonType:
		dst = order.AppendUint32(dst, uint32(len(s.rings)))
		for _, ring := range s.rings {
			dst = order.AppendUint32(dst, uint32(len(ring)/s.ctype.Dimension()))
			dst = appendFloats(dst, ring, order)
		}
		return dst
	default:
		dst = order.AppendUint32(dst, uint32(len(s.members)))
		for _, member := range s.members {
			dst = appendWKB(dst, member, sridUnknown, order, variant)
		}
		return dst
	}
}

// appendFloats appends each value as a 64-bit float.
func appendFloats(dst []byte, vals []float64, order binary.AppendByteOrder) []byte {
	for _, val := range vals {
		dst = order.AppendUint64(dst, math.Float64bits(val))
	}
	return dst
}
