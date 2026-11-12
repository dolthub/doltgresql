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

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// twkbWriter writes Tiny WKB, which stores each ordinate as a scaled integer that differs from the one before it.
type twkbWriter struct {
	precisions [4]int
	withSizes  bool
	withBoxes  bool
}

// scales returns the factor that scales each ordinate of a shape with the given coordinates type to an integer.
func (w twkbWriter) scales(ctype coordinatesType) []float64 {
	scales := []float64{math.Pow(10, float64(w.precisions[0])), math.Pow(10, float64(w.precisions[1]))}
	if ctype.Is3D() {
		scales = append(scales, math.Pow(10, float64(w.precisions[2])))
	}
	if ctype.IsMeasured() {
		scales = append(scales, math.Pow(10, float64(w.precisions[3])))
	}
	return scales
}

// write returns the Tiny WKB of a shape that is not empty, with an identifier for each member when `ids` is given.
func (w twkbWriter) write(s shape, ids []int64) []byte {
	scales := w.scales(s.ctype)
	dims := len(scales)
	var body []byte
	last := make([]int64, dims)
	points := func(seq []float64, withCount bool) {
		count := len(seq) / dims
		if withCount {
			body = binary.AppendUvarint(body, uint64(count))
		}
		for i := 0; i < count; i++ {
			for d := 0; d < dims; d++ {
				value := int64(math.Round(seq[i*dims+d] * scales[d]))
				body = binary.AppendUvarint(body, zigzag(value-last[d]))
				last[d] = value
			}
		}
	}
	idList := func() {
		if ids == nil {
			return
		}
		for _, id := range ids {
			body = binary.AppendUvarint(body, zigzag(id))
		}
	}
	switch s.typeNum {
	case pointType:
		points(s.coords, false)
	case lineType:
		points(s.coords, true)
	case polygonType:
		body = binary.AppendUvarint(body, uint64(len(s.rings)))
		for _, ring := range s.rings {
			points(ring, true)
		}
	case multiPointType, multiLineType, multiPolygonType:
		body = binary.AppendUvarint(body, uint64(len(s.members)))
		idList()
		for _, member := range s.members {
			switch member.typeNum {
			case pointType:
				points(member.coords, false)
			case lineType:
				points(member.coords, true)
			case polygonType:
				body = binary.AppendUvarint(body, uint64(len(member.rings)))
				for _, ring := range member.rings {
					points(ring, true)
				}
			}
		}
	case collectionType:
		body = binary.AppendUvarint(body, uint64(len(s.members)))
		idList()
		for _, member := range s.members {
			body = append(body, w.write(member, nil)...)
		}
	}
	header := []byte{byte(s.typeNum) | byte(zigzag(int64(w.precisions[0])))<<4, 0}
	if w.withBoxes {
		header[1] |= 0x01
	}
	if w.withSizes {
		header[1] |= 0x02
	}
	if ids != nil {
		header[1] |= 0x04
	}
	if s.hasZ() || s.hasM() {
		header[1] |= 0x08
		extended := byte(0)
		if s.hasZ() {
			extended |= 0x01 | byte(w.precisions[2])<<2
		}
		if s.hasM() {
			extended |= 0x02 | byte(w.precisions[3])<<5
		}
		header = append(header, extended)
	}
	var box []byte
	if w.withBoxes {
		mins, maxs := make([]int64, dims), make([]int64, dims)
		for d := range mins {
			mins[d], maxs[d] = math.MaxInt64, math.MinInt64
		}
		s.mapCoords(func(point []float64) []float64 {
			for d := 0; d < dims && d < len(point); d++ {
				value := int64(math.Round(point[d] * scales[d]))
				mins[d], maxs[d] = min(mins[d], value), max(maxs[d], value)
			}
			return point
		})
		for d := 0; d < dims; d++ {
			box = binary.AppendUvarint(box, zigzag(mins[d]))
			box = binary.AppendUvarint(box, zigzag(maxs[d]-mins[d]))
		}
	}
	if w.withSizes {
		header = binary.AppendUvarint(header, uint64(len(box)+len(body)))
	}
	return append(append(header, box...), body...)
}

// twkbReader reads Tiny WKB.
type twkbReader struct {
	data     []byte
	position int
}

// errTWKBSize is returned when Tiny WKB ends before its structure does.
var errTWKBSize = pgerror.New(pgcode.Internal, "twkb_parse_state_advance: TWKB structure does not match expected size!")

// uvarint reads an unsigned varint.
func (r *twkbReader) uvarint() (uint64, error) {
	v, n := binary.Uvarint(r.data[r.position:])
	if n == 0 {
		return 0, errTWKBSize
	}
	if n < 0 {
		return 0, pgerror.New(pgcode.Internal, "twkb_parse_state_varint: varint overflow")
	}
	r.position += n
	return v, nil
}

// varint reads a zigzag-encoded signed varint.
func (r *twkbReader) varint() (int64, error) {
	v, err := r.uvarint()
	return int64(v>>1) ^ -int64(v&1), err
}

// read returns the shape that the Tiny WKB at the reader's position describes.
func (r *twkbReader) read() (shape, error) {
	if r.position+2 > len(r.data) {
		return shape{}, errTWKBSize
	}
	typeAndPrecision, metadata := r.data[r.position], r.data[r.position+1]
	r.position += 2
	encodedPrecision := int64(typeAndPrecision >> 4)
	precision := int(encodedPrecision>>1 ^ -(encodedPrecision & 1))
	s := shape{typeNum: int(typeAndPrecision & 0x0F)}
	scales := []float64{math.Pow(10, float64(precision)), math.Pow(10, float64(precision))}
	if metadata&0x08 != 0 {
		if r.position >= len(r.data) {
			return shape{}, errTWKBSize
		}
		extended := r.data[r.position]
		r.position++
		hasZ, hasM := extended&0x01 != 0, extended&0x02 != 0
		if hasZ {
			s.ctype |= dimXYZ
		}
		if hasM {
			s.ctype |= dimXYM
		}
		if hasZ {
			scales = append(scales, math.Pow(10, float64(extended>>2&0x07)))
		}
		if hasM {
			scales = append(scales, math.Pow(10, float64(extended>>5&0x07)))
		}
	}
	if metadata&0x02 != 0 {
		if _, err := r.uvarint(); err != nil {
			return shape{}, err
		}
	}
	if metadata&0x10 != 0 {
		return s, nil
	}
	dims := len(scales)
	if metadata&0x01 != 0 {
		for i := 0; i < 2*dims; i++ {
			if _, err := r.varint(); err != nil {
				return shape{}, err
			}
		}
	}
	last := make([]int64, dims)
	points := func(count int) ([]float64, error) {
		seq := make([]float64, 0, count*dims)
		for i := 0; i < count; i++ {
			for d := 0; d < dims; d++ {
				delta, err := r.varint()
				if err != nil {
					return nil, err
				}
				last[d] += delta
				seq = append(seq, float64(last[d])/scales[d])
			}
		}
		return seq, nil
	}
	countedPoints := func() ([]float64, error) {
		count, err := r.uvarint()
		if err != nil {
			return nil, err
		}
		return points(int(count))
	}
	polygon := func() (shape, error) {
		count, err := r.uvarint()
		if err != nil {
			return shape{}, err
		}
		p := shape{typeNum: polygonType, ctype: s.ctype}
		for i := 0; i < int(count); i++ {
			ring, err := countedPoints()
			if err != nil {
				return shape{}, err
			}
			p.rings = append(p.rings, ring)
		}
		return p, nil
	}
	var err error
	switch s.typeNum {
	case pointType:
		s.coords, err = points(1)
	case lineType:
		s.coords, err = countedPoints()
	case polygonType:
		return polygon()
	case multiPointType, multiLineType, multiPolygonType, collectionType:
		var count uint64
		if count, err = r.uvarint(); err != nil {
			return shape{}, err
		}
		if metadata&0x04 != 0 {
			for i := 0; i < int(count); i++ {
				if _, err = r.varint(); err != nil {
					return shape{}, err
				}
			}
		}
		for i := 0; i < int(count); i++ {
			var member shape
			switch s.typeNum {
			case multiPointType:
				member = shape{typeNum: pointType, ctype: s.ctype}
				member.coords, err = points(1)
			case multiLineType:
				member = shape{typeNum: lineType, ctype: s.ctype}
				member.coords, err = countedPoints()
			case multiPolygonType:
				member, err = polygon()
			default:
				member, err = r.read()
			}
			if err != nil {
				return shape{}, err
			}
			s.members = append(s.members, member)
		}
	default:
		return shape{}, pgerror.Newf(pgcode.Internal, "Unknown WKB type (%d)!", s.typeNum)
	}
	return s, err
}
