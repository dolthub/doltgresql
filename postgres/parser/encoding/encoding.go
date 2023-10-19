// Copyright 2023 Dolthub, Inc.
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

// Copyright 2014 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/cockroachdb/apd/v2"
	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
)

const (
	encodedNull = 0x00
	// A marker greater than NULL but lower than any other value.
	// This value is not actually ever present in a stored key, but
	// it's used in keys used as span boundaries for index scans.
	encodedNotNull = 0x01

	floatNaN     = encodedNotNull + 1
	floatNeg     = floatNaN + 1
	floatZero    = floatNeg + 1
	floatPos     = floatZero + 1
	floatNaNDesc = floatPos + 1 // NaN encoded descendingly

	// The gap between floatNaNDesc and bytesMarker was left for
	// compatibility reasons.
	bytesMarker          byte = 0x12
	bytesDescMarker      byte = bytesMarker + 1
	timeMarker           byte = bytesDescMarker + 1
	durationBigNegMarker byte = timeMarker + 1 // Only used for durations < MinInt64 nanos.
	durationMarker       byte = durationBigNegMarker + 1
	durationBigPosMarker byte = durationMarker + 1 // Only used for durations > MaxInt64 nanos.

	decimalNaN              = durationBigPosMarker + 1 // 24
	decimalNegativeInfinity = decimalNaN + 1
	decimalNegLarge         = decimalNegativeInfinity + 1
	decimalNegMedium        = decimalNegLarge + 11
	decimalNegSmall         = decimalNegMedium + 1
	decimalZero             = decimalNegSmall + 1
	decimalPosSmall         = decimalZero + 1
	decimalPosMedium        = decimalPosSmall + 1
	decimalPosLarge         = decimalPosMedium + 11
	decimalInfinity         = decimalPosLarge + 1
	decimalNaNDesc          = decimalInfinity + 1 // NaN encoded descendingly
	decimalTerminator       = 0x00

	jsonInvertedIndex = decimalNaNDesc + 1
	jsonEmptyArray    = jsonInvertedIndex + 1
	jsonEmptyObject   = jsonEmptyArray + 1

	bitArrayMarker             = jsonEmptyObject + 1
	bitArrayDescMarker         = bitArrayMarker + 1
	bitArrayDataTerminator     = 0x00
	bitArrayDataDescTerminator = 0xff

	timeTZMarker  = bitArrayDescMarker + 1
	geoMarker     = timeTZMarker + 1
	geoDescMarker = geoMarker + 1

	// Markers and terminators for key encoding Datum arrays in sorted order.
	// For the arrayKeyMarker and other types like bytes and bit arrays, it
	// might be unclear why we have a separate marker for the ascending and
	// descending cases. This is necessary because the terminators for these
	// encodings are different depending on the direction the data is encoded
	// in. In order to safely decode a set of bytes without knowing the direction
	// of the encoding, we must store this information in the marker. Otherwise,
	// we would not know what terminator to look for when decoding this format.
	arrayKeyMarker           = geoDescMarker + 1
	arrayKeyDescendingMarker = arrayKeyMarker + 1

	box2DMarker = arrayKeyDescendingMarker + 1

	arrayKeyTerminator           byte = 0x00
	arrayKeyDescendingTerminator byte = 0xFF

	// IntMin is chosen such that the range of int tags does not overlap the
	// ascii character set that is frequently used in testing.
	IntMin      = 0x80 // 128
	intMaxWidth = 8
	intZero     = IntMin + intMaxWidth           // 136
	intSmall    = IntMax - intZero - intMaxWidth // 109
	// IntMax is the maximum int tag value.
	IntMax = 0xfd // 253

	// Nulls come last when encoded descendingly.
	// This value is not actually ever present in a stored key, but
	// it's used in keys used as span boundaries for index scans.
	encodedNotNullDesc = 0xfe
	encodedNullDesc    = 0xff
)

// Direction for ordering results.
type Direction int

// Direction values.
const (
	_ Direction = iota
	Ascending
	Descending
)

const escapeLength = 2

// EncodeUint32Ascending encodes the uint32 value using a big-endian 4 byte
// representation. The bytes are appended to the supplied buffer and
// the final buffer is returned.
func EncodeUint32Ascending(b []byte, v uint32) []byte {
	return append(b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// PutUint32Ascending encodes the uint32 value using a big-endian 4 byte
// representation at the specified index, lengthening the input slice if
// necessary.
func PutUint32Ascending(b []byte, v uint32, idx int) []byte {
	for len(b) < idx+4 {
		b = append(b, 0)
	}
	b[idx] = byte(v >> 24)
	b[idx+1] = byte(v >> 16)
	b[idx+2] = byte(v >> 8)
	b[idx+3] = byte(v)
	return b
}

// DecodeUint32Ascending decodes a uint32 from the input buffer, treating
// the input as a big-endian 4 byte uint32 representation. The remainder
// of the input buffer and the decoded uint32 are returned.
func DecodeUint32Ascending(b []byte) ([]byte, uint32, error) {
	if len(b) < 4 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uint32 int value")
	}
	v := binary.BigEndian.Uint32(b)
	return b[4:], v, nil
}

const uint64AscendingEncodedLength = 8

// EncodeUint64Ascending encodes the uint64 value using a big-endian 8 byte
// representation. The bytes are appended to the supplied buffer and
// the final buffer is returned.
func EncodeUint64Ascending(b []byte, v uint64) []byte {
	return append(b,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// DecodeUint64Ascending decodes a uint64 from the input buffer, treating
// the input as a big-endian 8 byte uint64 representation. The remainder
// of the input buffer and the decoded uint64 are returned.
func DecodeUint64Ascending(b []byte) ([]byte, uint64, error) {
	if len(b) < 8 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uint64 int value")
	}
	v := binary.BigEndian.Uint64(b)
	return b[8:], v, nil
}

// MaxVarintLen is the maximum length of a value encoded using any of:
// - EncodeVarintAscending
// - EncodeVarintDescending
// - EncodeUvarintAscending
// - EncodeUvarintDescending
const MaxVarintLen = 9

// getVarintLen returns the encoded length of an encoded varint. Assumes the
// slice has at least one byte.
func getVarintLen(b []byte) (int, error) {
	length := int(b[0]) - intZero
	if length >= 0 {
		if length <= intSmall {
			// just the tag
			return 1, nil
		}
		// tag and length-intSmall bytes
		length = 1 + length - intSmall
	} else {
		// tag and -length bytes
		length = 1 - length
	}

	if length > len(b) {
		return 0, errors.Errorf("varint length %d exceeds slice length %d", length, len(b))
	}
	return length, nil
}

// EncodeUvarintAscending encodes the uint64 value using a variable length
// (length-prefixed) representation. The length is encoded as a single
// byte indicating the number of encoded bytes (-8) to follow. See
// EncodeVarintAscending for rationale. The encoded bytes are appended to the
// supplied buffer and the final buffer is returned.
func EncodeUvarintAscending(b []byte, v uint64) []byte {
	switch {
	case v <= intSmall:
		return append(b, intZero+byte(v))
	case v <= 0xff:
		return append(b, IntMax-7, byte(v))
	case v <= 0xffff:
		return append(b, IntMax-6, byte(v>>8), byte(v))
	case v <= 0xffffff:
		return append(b, IntMax-5, byte(v>>16), byte(v>>8), byte(v))
	case v <= 0xffffffff:
		return append(b, IntMax-4, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	case v <= 0xffffffffff:
		return append(b, IntMax-3, byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8),
			byte(v))
	case v <= 0xffffffffffff:
		return append(b, IntMax-2, byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16),
			byte(v>>8), byte(v))
	case v <= 0xffffffffffffff:
		return append(b, IntMax-1, byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24),
			byte(v>>16), byte(v>>8), byte(v))
	default:
		return append(b, IntMax, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
			byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	}
}

// EncodeUvarintDescending encodes the uint64 value so that it sorts in
// reverse order, from largest to smallest.
func EncodeUvarintDescending(b []byte, v uint64) []byte {
	switch {
	case v == 0:
		return append(b, IntMin+8)
	case v <= 0xff:
		v = ^v
		return append(b, IntMin+7, byte(v))
	case v <= 0xffff:
		v = ^v
		return append(b, IntMin+6, byte(v>>8), byte(v))
	case v <= 0xffffff:
		v = ^v
		return append(b, IntMin+5, byte(v>>16), byte(v>>8), byte(v))
	case v <= 0xffffffff:
		v = ^v
		return append(b, IntMin+4, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	case v <= 0xffffffffff:
		v = ^v
		return append(b, IntMin+3, byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8),
			byte(v))
	case v <= 0xffffffffffff:
		v = ^v
		return append(b, IntMin+2, byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16),
			byte(v>>8), byte(v))
	case v <= 0xffffffffffffff:
		v = ^v
		return append(b, IntMin+1, byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24),
			byte(v>>16), byte(v>>8), byte(v))
	default:
		v = ^v
		return append(b, IntMin, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
			byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	}
}

// highestByteIndex returns the index (0 to 7) of the highest nonzero byte in v.
func highestByteIndex(v uint64) int {
	l := 0
	if v > 0xffffffff {
		v >>= 32
		l += 4
	}
	if v > 0xffff {
		v >>= 16
		l += 2
	}
	if v > 0xff {
		l++
	}
	return l
}

// EncLenUvarintAscending returns the encoding length for EncodeUvarintAscending
// without actually encoding.
func EncLenUvarintAscending(v uint64) int {
	if v <= intSmall {
		return 1
	}
	return 2 + highestByteIndex(v)
}

// EncLenUvarintDescending returns the encoding length for
// EncodeUvarintDescending without actually encoding.
func EncLenUvarintDescending(v uint64) int {
	if v == 0 {
		return 1
	}
	return 2 + highestByteIndex(v)
}

// DecodeUvarintAscending decodes a varint encoded uint64 from the input
// buffer. The remainder of the input buffer and the decoded uint64
// are returned.
func DecodeUvarintAscending(b []byte) ([]byte, uint64, error) {
	if len(b) == 0 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value")
	}
	length := int(b[0]) - intZero
	b = b[1:] // skip length byte
	if length <= intSmall {
		return b, uint64(length), nil
	}
	length -= intSmall
	if length < 0 || length > 8 {
		return nil, 0, errors.Errorf("invalid uvarint length of %d", length)
	} else if len(b) < length {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value: %q", b)
	}
	var v uint64
	// It is faster to range over the elements in a slice than to index
	// into the slice on each loop iteration.
	for _, t := range b[:length] {
		v = (v << 8) | uint64(t)
	}
	return b[length:], v, nil
}

// DecodeUvarintDescending decodes a uint64 value which was encoded
// using EncodeUvarintDescending.
func DecodeUvarintDescending(b []byte) ([]byte, uint64, error) {
	if len(b) == 0 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value")
	}
	length := intZero - int(b[0])
	b = b[1:] // skip length byte
	if length < 0 || length > 8 {
		return nil, 0, errors.Errorf("invalid uvarint length of %d", length)
	} else if len(b) < length {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value: %q", b)
	}
	var x uint64
	for _, t := range b[:length] {
		x = (x << 8) | uint64(^t)
	}
	return b[length:], x, nil
}

const (
	// <term>     -> \x00\x01
	// \x00       -> \x00\xff
	escape                   byte = 0x00
	escapedTerm              byte = 0x01
	escapedJSONObjectKeyTerm byte = 0x02
	escapedJSONArray         byte = 0x03
	escaped00                byte = 0xff
	escapedFF                byte = 0x00
)

type escapes struct {
	escape      byte
	escapedTerm byte
	escaped00   byte
	escapedFF   byte
	marker      byte
}

var (
	ascendingBytesEscapes  = escapes{escape, escapedTerm, escaped00, escapedFF, bytesMarker}
	descendingBytesEscapes = escapes{^escape, ^escapedTerm, ^escaped00, ^escapedFF, bytesDescMarker}

	ascendingGeoEscapes  = escapes{escape, escapedTerm, escaped00, escapedFF, geoMarker}
	descendingGeoEscapes = escapes{^escape, ^escapedTerm, ^escaped00, ^escapedFF, geoDescMarker}
)

// encodeBytesAscendingWithTerminatorAndPrefix encodes the []byte value using an escape-based
// encoding. The encoded value is terminated with the sequence
// "\x00\terminator". The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned. The terminator allows us to pass
// different terminators for things such as JSON key encoding.
func encodeBytesAscendingWithTerminatorAndPrefix(
	b []byte, data []byte, terminator byte, prefix byte,
) []byte {
	b = append(b, prefix)
	return encodeBytesAscendingWithTerminator(b, data, terminator)
}

// encodeBytesAscendingWithTerminator encodes the []byte value using an escape-based
// encoding. The encoded value is terminated with the sequence
// "\x00\terminator". The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned. The terminator allows us to pass
// different terminators for things such as JSON key encoding.
func encodeBytesAscendingWithTerminator(b []byte, data []byte, terminator byte) []byte {
	bs := encodeBytesAscendingWithoutTerminatorOrPrefix(b, data)
	return append(bs, escape, terminator)
}

// encodeBytesAscendingWithoutTerminatorOrPrefix encodes the []byte value using an escape-based
// encoding.
func encodeBytesAscendingWithoutTerminatorOrPrefix(b []byte, data []byte) []byte {
	for {
		// IndexByte is implemented by the go runtime in assembly and is
		// much faster than looping over the bytes in the slice.
		i := bytes.IndexByte(data, escape)
		if i == -1 {
			break
		}
		b = append(b, data[:i]...)
		b = append(b, escape, escaped00)
		data = data[i+1:]
	}
	return append(b, data...)
}

// getBytesLength finds the length of a bytes encoding.
func getBytesLength(b []byte, e escapes) (int, error) {
	// Skip the tag.
	skipped := 1
	for {
		i := bytes.IndexByte(b[skipped:], e.escape)
		if i == -1 {
			return 0, errors.Errorf("did not find terminator %#x in buffer %#x", e.escape, b)
		}
		if i+1 >= len(b) {
			return 0, errors.Errorf("malformed escape in buffer %#x", b)
		}
		skipped += i + escapeLength
		if b[skipped-1] == e.escapedTerm {
			return skipped, nil
		}
	}
}

// UnsafeConvertStringToBytes converts a string to a byte array to be used with
// string encoding functions. Note that the output byte array should not be
// modified if the input string is expected to be used again - doing so could
// violate Go semantics.
func UnsafeConvertStringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	// We unsafely convert the string to a []byte to avoid the
	// usual allocation when converting to a []byte. This is
	// kosher because we know that EncodeBytes{,Descending} does
	// not keep a reference to the value it encodes. The first
	// step is getting access to the string internals.
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	// Next we treat the string data as a maximally sized array which we
	// slice. This usage is safe because the pointer value remains in the string.
	return (*[0x7fffffff]byte)(unsafe.Pointer(hdr.Data))[:len(s):len(s)]
}

// EncodeStringAscending encodes the string value using an escape-based encoding. See
// EncodeBytes for details. The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned.
func EncodeStringAscending(b []byte, s string) []byte {
	return encodeStringAscendingWithTerminatorAndPrefix(b, s, ascendingBytesEscapes.escapedTerm, bytesMarker)
}

// encodeStringAscendingWithTerminatorAndPrefix encodes the string value using an escape-based encoding. See
// EncodeBytes for details. The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned. We can also pass a terminator byte to be used with
// JSON key encoding.
func encodeStringAscendingWithTerminatorAndPrefix(
	b []byte, s string, terminator byte, prefix byte,
) []byte {
	unsafeString := UnsafeConvertStringToBytes(s)
	return encodeBytesAscendingWithTerminatorAndPrefix(b, unsafeString, terminator, prefix)
}

// EncodeJSONKeyStringAscending encodes the JSON key string value with a JSON specific escaped
// terminator. This allows us to encode keys in the same number of bytes as a string,
// while at the same time giving us a sentinel to identify JSON keys. The end parameter is used
// to determine if this is the last key in a a JSON path. If it is we don't add a separator after it.
func EncodeJSONKeyStringAscending(b []byte, s string, end bool) []byte {
	str := UnsafeConvertStringToBytes(s)

	if end {
		return encodeBytesAscendingWithoutTerminatorOrPrefix(b, str)
	}
	return encodeBytesAscendingWithTerminator(b, str, escapedJSONObjectKeyTerm)
}

// EncodeJSONEmptyArray returns a byte array b with a byte to signify an empty JSON array.
func EncodeJSONEmptyArray(b []byte) []byte {
	return append(b, escape, escapedTerm, jsonEmptyArray)
}

// AddJSONPathTerminator adds a json path terminator to a byte array.
func AddJSONPathTerminator(b []byte) []byte {
	return append(b, escape, escapedTerm)
}

// EncodeJSONEmptyObject returns a byte array b with a byte to signify an empty JSON object.
func EncodeJSONEmptyObject(b []byte) []byte {
	return append(b, escape, escapedTerm, jsonEmptyObject)
}

// EncodeNullAscending encodes a NULL value. The encodes bytes are appended to the
// supplied buffer and the final buffer is returned. The encoded value for a
// NULL is guaranteed to not be a prefix for the EncodeVarint, EncodeFloat,
// EncodeBytes and EncodeString encodings.
func EncodeNullAscending(b []byte) []byte {
	return append(b, encodedNull)
}

// EncodeJSONAscending encodes a JSON Type. The encoded bytes are appended to the
// supplied buffer and the final buffer is returned.
func EncodeJSONAscending(b []byte) []byte {
	return append(b, jsonInvertedIndex)
}

// EncodeArrayAscending encodes a value used to signify membership of an array for JSON objects.
func EncodeArrayAscending(b []byte) []byte {
	return append(b, escape, escapedJSONArray)
}

// EncodeTrueAscending encodes the boolean value true for use with JSON inverted indexes.
func EncodeTrueAscending(b []byte) []byte {
	return append(b, byte(True))
}

// EncodeFalseAscending encodes the boolean value false for use with JSON inverted indexes.
func EncodeFalseAscending(b []byte) []byte {
	return append(b, byte(False))
}

// getBitArrayWordsLen returns the number of bit array words in the
// encoded bytes and the size in bytes of the encoded word array
// (excluding the terminator byte).
func getBitArrayWordsLen(b []byte, term byte) (int, int, error) {
	bSearch := b
	numWords := 0
	sz := 0
	for {
		if len(bSearch) == 0 {
			return 0, 0, errors.Errorf("slice too short for bit array (%d)", len(b))
		}
		if bSearch[0] == term {
			break
		}
		vLen, err := getVarintLen(bSearch)
		if err != nil {
			return 0, 0, err
		}
		bSearch = bSearch[vLen:]
		numWords++
		sz += vLen
	}
	return numWords, sz, nil
}

// Type represents the type of a value encoded by
// Encode{Null,NotNull,Varint,Uvarint,Float,Bytes}.
//go:generate stringer -type=Type
type Type int

// Type values.
// TODO(dan, arjun): Make this into a proto enum.
// The 'Type' annotations are necessary for producing stringer-generated values.
const (
	Unknown   Type = 0
	Null      Type = 1
	NotNull   Type = 2
	Int       Type = 3
	Float     Type = 4
	Decimal   Type = 5
	Bytes     Type = 6
	BytesDesc Type = 7 // Bytes encoded descendingly
	Time      Type = 8
	Duration  Type = 9
	True      Type = 10
	False     Type = 11
	UUID      Type = 12
	Array     Type = 13
	IPAddr    Type = 14
	// SentinelType is used for bit manipulation to check if the encoded type
	// value requires more than 4 bits, and thus will be encoded in two bytes. It
	// is not used as a type value, and thus intentionally overlaps with the
	// subsequent type value. The 'Type' annotation is intentionally omitted here.
	SentinelType      = 15
	JSON         Type = 15
	Tuple        Type = 16
	BitArray     Type = 17
	BitArrayDesc Type = 18 // BitArray encoded descendingly
	TimeTZ       Type = 19
	Geo          Type = 20
	GeoDesc      Type = 21
	ArrayKeyAsc  Type = 22 // Array key encoding
	ArrayKeyDesc Type = 23 // Array key encoded descendingly
	Box2D        Type = 24
)

// typMap maps an encoded type byte to a decoded Type. It's got 256 slots, one
// for every possible byte value.
var typMap [256]Type

func init() {
	buf := []byte{0}
	for i := range typMap {
		buf[0] = byte(i)
		typMap[i] = slowPeekType(buf)
	}
}

// PeekType peeks at the type of the value encoded at the start of b.
func PeekType(b []byte) Type {
	if len(b) >= 1 {
		return typMap[b[0]]
	}
	return Unknown
}

// slowPeekType is the old implementation of PeekType. It's used to generate
// the lookup table for PeekType.
func slowPeekType(b []byte) Type {
	if len(b) >= 1 {
		m := b[0]
		switch {
		case m == encodedNull, m == encodedNullDesc:
			return Null
		case m == encodedNotNull, m == encodedNotNullDesc:
			return NotNull
		case m == arrayKeyMarker:
			return ArrayKeyAsc
		case m == arrayKeyDescendingMarker:
			return ArrayKeyDesc
		case m == bytesMarker:
			return Bytes
		case m == bytesDescMarker:
			return BytesDesc
		case m == bitArrayMarker:
			return BitArray
		case m == bitArrayDescMarker:
			return BitArrayDesc
		case m == timeMarker:
			return Time
		case m == timeTZMarker:
			return TimeTZ
		case m == geoMarker:
			return Geo
		case m == box2DMarker:
			return Box2D
		case m == geoDescMarker:
			return GeoDesc
		case m == byte(Array):
			return Array
		case m == byte(True):
			return True
		case m == byte(False):
			return False
		case m == durationBigNegMarker, m == durationMarker, m == durationBigPosMarker:
			return Duration
		case m >= IntMin && m <= IntMax:
			return Int
		case m >= floatNaN && m <= floatNaNDesc:
			return Float
		case m >= decimalNaN && m <= decimalNaNDesc:
			return Decimal
		}
	}
	return Unknown
}

// GetMultiVarintLen find the length of <num> encoded varints that follow a
// 1-byte tag.
func GetMultiVarintLen(b []byte, num int) (int, error) {
	p := 1
	for i := 0; i < num && p < len(b); i++ {
		len, err := getVarintLen(b[p:])
		if err != nil {
			return 0, err
		}
		p += len
	}
	return p, nil
}

// getMultiNonsortingVarintLen finds the length of <num> encoded nonsorting varints.
func getMultiNonsortingVarintLen(b []byte, num int) (int, error) {
	p := 0
	for i := 0; i < num && p < len(b); i++ {
		_, len, _, err := DecodeNonsortingStdlibVarint(b[p:])
		if err != nil {
			return 0, err
		}
		p += len
	}
	return p, nil
}

// getArrayLength returns the length of a key encoded array. The input
// must have had the array type marker stripped from the front.
func getArrayLength(buf []byte, dir Direction) (int, error) {
	result := 0
	for {
		if len(buf) == 0 {
			return 0, errors.AssertionFailedf("invalid array encoding (unterminated)")
		}
		if IsArrayKeyDone(buf, dir) {
			// Increment to include the terminator byte.
			result++
			break
		}
		next, err := PeekLength(buf)
		if err != nil {
			return 0, err
		}
		// Shift buf over by the encoded data amount.
		buf = buf[next:]
		result += next
	}
	return result, nil
}

// peekBox2DLength peeks to look at the length of a box2d encoding.
func peekBox2DLength(b []byte) (int, error) {
	length := 0
	curr := b
	for i := 0; i < 4; i++ {
		if len(curr) == 0 {
			return 0, errors.Newf("slice too short for box2d")
		}
		switch curr[0] {
		case floatNaN, floatNaNDesc, floatZero:
			length++
			curr = curr[1:]
		case floatNeg, floatPos:
			length += 9
			curr = curr[9:]
		default:
			return 0, errors.Newf("unexpected marker for box2d: %x", curr[0])
		}
	}
	return length, nil
}

// PeekLength returns the length of the encoded value at the start of b.  Note:
// if this function succeeds, it's not a guarantee that decoding the value will
// succeed. PeekLength is meant to be used on key encoded data only.
func PeekLength(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.Errorf("empty slice")
	}
	m := b[0]
	switch m {
	case encodedNull, encodedNullDesc, encodedNotNull, encodedNotNullDesc,
		floatNaN, floatNaNDesc, floatZero, decimalZero, byte(True), byte(False):
		// interleavedSentinel also falls into this path. Since it
		// contains the same byte value as encodedNotNullDesc, it
		// cannot be included explicitly in the case statement.
		// ascendingNullWithinArrayKey and descendingNullWithinArrayKey also
		// contain the same byte values as encodedNotNull and encodedNotNullDesc
		// respectively.
		return 1, nil
	case bitArrayMarker, bitArrayDescMarker:
		terminator := byte(bitArrayDataTerminator)
		if m == bitArrayDescMarker {
			terminator = bitArrayDataDescTerminator
		}
		_, n, err := getBitArrayWordsLen(b[1:], terminator)
		if err != nil {
			return 1 + n, err
		}
		m, err := getVarintLen(b[n+2:])
		if err != nil {
			return 1 + n + m + 1, err
		}
		return 1 + n + m + 1, nil
	case arrayKeyMarker, arrayKeyDescendingMarker:
		dir := Ascending
		if m == arrayKeyDescendingMarker {
			dir = Descending
		}
		length, err := getArrayLength(b[1:], dir)
		return 1 + length, err
	case bytesMarker:
		return getBytesLength(b, ascendingBytesEscapes)
	case box2DMarker:
		if len(b) == 0 {
			return 0, errors.Newf("slice too short for box2d")
		}
		length, err := peekBox2DLength(b[1:])
		if err != nil {
			return 0, err
		}
		return 1 + length, nil
	case geoMarker:
		// Expect to reserve at least 8 bytes for int64.
		if len(b) < 8 {
			return 0, errors.Errorf("slice too short for spatial object (%d)", len(b))
		}
		ret, err := getBytesLength(b[8:], ascendingGeoEscapes)
		if err != nil {
			return 0, err
		}
		return 8 + ret, nil
	case jsonInvertedIndex:
		return getJSONInvertedIndexKeyLength(b)
	case bytesDescMarker:
		return getBytesLength(b, descendingBytesEscapes)
	case geoDescMarker:
		// Expect to reserve at least 8 bytes for int64.
		if len(b) < 8 {
			return 0, errors.Errorf("slice too short for spatial object (%d)", len(b))
		}
		ret, err := getBytesLength(b[8:], descendingGeoEscapes)
		if err != nil {
			return 0, err
		}
		return 8 + ret, nil
	case timeMarker, timeTZMarker:
		return GetMultiVarintLen(b, 2)
	case durationBigNegMarker, durationMarker, durationBigPosMarker:
		return GetMultiVarintLen(b, 3)
	case floatNeg, floatPos:
		// the marker is followed by 8 bytes
		if len(b) < 9 {
			return 0, errors.Errorf("slice too short for float (%d)", len(b))
		}
		return 9, nil
	}
	if m >= IntMin && m <= IntMax {
		return getVarintLen(b)
	}
	if m >= decimalNaN && m <= decimalNaNDesc {
		return getDecimalLen(b)
	}
	return 0, errors.Errorf("unknown tag %d", m)
}

// DecodeNonsortingStdlibVarint decodes a value encoded by EncodeNonsortingVarint. It
// returns the length of the encoded varint and value.
func DecodeNonsortingStdlibVarint(b []byte) (remaining []byte, length int, value int64, err error) {
	value, length = binary.Varint(b)
	if length <= 0 {
		return nil, 0, 0, fmt.Errorf("int64 varint decoding failed: %d", length)
	}
	return b[length:], length, value, nil
}

// DecodeNonsortingUvarint decodes a value encoded by EncodeNonsortingUvarint. It
// returns the length of the encoded varint and value.
func DecodeNonsortingUvarint(buf []byte) (remaining []byte, length int, value uint64, err error) {
	// TODO(dan): Handle overflow.
	for i, b := range buf {
		value += uint64(b & 0x7f)
		if b < 0x80 {
			return buf[i+1:], i + 1, value, nil
		}
		value <<= 7
	}
	return buf, 0, 0, nil
}

// DecodeNonsortingStdlibUvarint decodes a value encoded with binary.PutUvarint. It
// returns the length of the encoded varint and value.
func DecodeNonsortingStdlibUvarint(
	buf []byte,
) (remaining []byte, length int, value uint64, err error) {
	i, n := binary.Uvarint(buf)
	if n <= 0 {
		return buf, 0, 0, errors.New("buffer too small")
	}
	return buf[n:], n, i, nil
}

const floatValueEncodedLength = uint64AscendingEncodedLength

// EncodeUntaggedDecimalValue encodes an apd.Decimal value, appends it to the supplied
// buffer, and returns the final buffer.
func EncodeUntaggedDecimalValue(appendTo []byte, d *apd.Decimal) []byte {
	// To avoid the allocation, leave space for the varint, encode the decimal,
	// encode the varint, and shift the encoded decimal to the end of the
	// varint.
	varintPos := len(appendTo)
	// Manually append 10 (binary.MaxVarintLen64) 0s to avoid the allocation.
	appendTo = append(appendTo, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	decOffset := len(appendTo)
	appendTo = EncodeNonsortingDecimal(appendTo, d)
	decLen := len(appendTo) - decOffset
	varintLen := binary.PutUvarint(appendTo[varintPos:decOffset], uint64(decLen))
	copy(appendTo[varintPos+varintLen:varintPos+varintLen+decLen], appendTo[decOffset:decOffset+decLen])
	return appendTo[:varintPos+varintLen+decLen]
}

// DecodeValueTag decodes a value encoded by EncodeValueTag, used as a prefix in
// each of the other EncodeFooValue methods.
//
// The tag is structured such that the encoded column id can be dropped from the
// front by removing the first `typeOffset` bytes. DecodeValueTag,
// PeekValueLength and each of the DecodeFooValue methods will still work as
// expected with `b[typeOffset:]`. (Except, obviously, the column id is no
// longer encoded so if this suffix is passed back to DecodeValueTag, the
// returned colID should be discarded.)
//
// Concretely:
//     b := ...
//     typeOffset, _, colID, typ, err := DecodeValueTag(b)
//     _, _, _, typ, err := DecodeValueTag(b[typeOffset:])
// will return the same typ and err and
//     DecodeFooValue(b)
//     DecodeFooValue(b[typeOffset:])
// will return the same thing. PeekValueLength works as expected with either of
// `b` or `b[typeOffset:]`.
func DecodeValueTag(b []byte) (typeOffset int, dataOffset int, colID uint32, typ Type, err error) {
	// TODO(dan): This can be made faster by special casing the single byte
	// version and skipping the column id extraction when it's not needed.
	if len(b) == 0 {
		return 0, 0, 0, Unknown, fmt.Errorf("empty array")
	}
	var n int
	var tag uint64
	b, n, tag, err = DecodeNonsortingUvarint(b)
	if err != nil {
		return 0, 0, 0, Unknown, err
	}
	colID = uint32(tag >> 4)

	typ = Type(tag & 0xf)
	typeOffset = n - 1
	dataOffset = n
	if typ == SentinelType {
		_, n, tag, err = DecodeNonsortingUvarint(b)
		if err != nil {
			return 0, 0, 0, Unknown, err
		}
		typ = Type(tag)
		dataOffset += n
	}
	return typeOffset, dataOffset, colID, typ, nil
}

// DecodeUntaggedDecimalValue decodes a value encoded by EncodeUntaggedDecimalValue.
func DecodeUntaggedDecimalValue(b []byte) (remaining []byte, d apd.Decimal, err error) {
	var i uint64
	b, _, i, err = DecodeNonsortingStdlibUvarint(b)
	if err != nil {
		return b, apd.Decimal{}, err
	}
	d, err = DecodeNonsortingDecimal(b[:int(i)], nil)
	return b[int(i):], d, err
}

const uuidValueEncodedLength = 16

var _ [uuidValueEncodedLength]byte = uuid.UUID{} // Assert that uuid.UUID is length 16.

// getInvertedIndexKeyLength finds the length of an inverted index key
// encoded as a byte array.
func getInvertedIndexKeyLength(b []byte) (int, error) {
	skipped := 0
	for {
		i := bytes.IndexByte(b[skipped:], escape)
		if i == -1 {
			return 0, errors.Errorf("malformed inverted index key in buffer %#x", b)
		}
		skipped += i + escapeLength
		switch b[skipped-1] {
		case escapedTerm, jsonEmptyObject, jsonEmptyArray:
			return skipped, nil
		}
	}
}

// getJSONInvertedIndexKeyLength returns the length of encoded JSON inverted index
// key at the start of b.
func getJSONInvertedIndexKeyLength(buf []byte) (int, error) {
	len, err := getInvertedIndexKeyLength(buf)
	if err != nil {
		return 0, err
	}

	switch buf[len] {
	case jsonEmptyArray, jsonEmptyObject:
		return len + 1, nil

	default:
		valLen, err := PeekLength(buf[len:])
		if err != nil {
			return 0, err
		}

		return len + valLen, nil
	}
}

// IsArrayKeyDone returns if the first byte in the input is the array
// terminator for the input direction.
func IsArrayKeyDone(buf []byte, dir Direction) bool {
	expected := arrayKeyTerminator
	if dir == Descending {
		expected = arrayKeyDescendingTerminator
	}
	return buf[0] == expected
}
