// Copyright 2024 Dolthub, Inc.
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

package utils

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/dolthub/doltgresql/core/id"
)

// Writer handles type-safe writing into a byte buffer, which may later be read from using Reader. The Writer will
// automatically grow as it is written to. The serialized forms of booleans, ints, uints, and floats are
// byte-comparable, meaning it is valid to use bytes.Compare() without needing to deserialize them. Variable-length
// encoded values, strings, and slices are not byte-comparable. This is not safe for concurrent use.
type Writer struct {
	buf      *bytes.Buffer
	numSlice []byte
}

// NewWriter creates a new Writer with the given starting capacity. A larger starting capacity reduces reallocations at
// the cost of potentially wasted memory.
func NewWriter(capacity uint64) *Writer {
	// If capacity is zero, then we'll set it to something arbitrary to try and minimize reallocations
	if capacity == 0 {
		capacity = 32
	}
	return &Writer{
		buf:      bytes.NewBuffer(make([]byte, 0, capacity)),
		numSlice: make([]byte, 10), // 10 bytes will cover all integers and variable-length integers
	}
}

// Bool writes a bool.
func (writer *Writer) Bool(val bool) {
	if val {
		writer.buf.WriteByte(1)
	} else {
		writer.buf.WriteByte(0)
	}
}

// Int8 writes an int8.
func (writer *Writer) Int8(val int8) {
	writer.buf.WriteByte(byte(val) + (1 << 7))
}

// Int16 writes an int16.
func (writer *Writer) Int16(val int16) {
	writer.Uint16(uint16(val) + (1 << 15))
}

// Int32 writes an int32.
func (writer *Writer) Int32(val int32) {
	writer.Uint32(uint32(val) + (1 << 31))
}

// Int64 writes an int64.
func (writer *Writer) Int64(val int64) {
	writer.Uint64(uint64(val) + (1 << 63))
}

// Uint8 writes a uint8.
func (writer *Writer) Uint8(val uint8) {
	writer.buf.WriteByte(val)
}

// Uint16 writes a uint16.
func (writer *Writer) Uint16(val uint16) {
	binary.BigEndian.PutUint16(writer.numSlice, val)
	writer.buf.Write(writer.numSlice[:2])
}

// Uint32 writes a uint32.
func (writer *Writer) Uint32(val uint32) {
	binary.BigEndian.PutUint32(writer.numSlice, val)
	writer.buf.Write(writer.numSlice[:4])
}

// Uint64 writes a uint64.
func (writer *Writer) Uint64(val uint64) {
	binary.BigEndian.PutUint64(writer.numSlice, val)
	writer.buf.Write(writer.numSlice[:8])
}

// Byte writes a byte. This is equivalent to Uint8, but is included since it is more common to refer to a byte rather
// than a uint8.
func (writer *Writer) Byte(val byte) {
	writer.buf.WriteByte(val)
}

// Float32 writes a float32.
func (writer *Writer) Float32(val float32) {
	// Float encoding produces byte-comparable serialized values when looking at the exponent and mantissa. This means
	// that we just have to flip to exponent and mantissa for negative values, and flip the sign bit so that negatives
	// sort before positives. To do this, we take advantage of arithmetic shifting by casting to a signed integer to
	// create a mask that only exists for negative values.
	uval := math.Float32bits(val)
	uval ^= uint32(int32(uval)>>31) & 0x7FFFFFFF
	uval ^= 0x80000000
	writer.Uint32(uval)
}

// Float64 writes a float64.
func (writer *Writer) Float64(val float64) {
	// Float encoding produces byte-comparable serialized values when looking at the exponent and mantissa. This means
	// that we just have to flip to exponent and mantissa for negative values, and flip the sign bit so that negatives
	// sort before positives. To do this, we take advantage of arithmetic shifting by casting to a signed integer to
	// create a mask that only exists for negative values.
	uval := math.Float64bits(val)
	uval ^= uint64(int64(uval)>>63) & 0x7FFFFFFFFFFFFFFF
	uval ^= 0x8000000000000000
	writer.Uint64(uval)
}

// VariableInt writes an int64 using variable-length encoding. Smaller values use less space at the cost of larger
// values using more bytes, but this is generally more space-efficient. This does carry a small computational hit when
// reading.
func (writer *Writer) VariableInt(val int64) {
	count := binary.PutVarint(writer.numSlice, val)
	writer.buf.Write(writer.numSlice[:count])
}

// VariableUint writes a uint64 using variable-length encoding. Smaller values use less space at the cost of larger
// values using more bytes, but this is generally more space-efficient. This does carry a small computational hit when
// reading.
func (writer *Writer) VariableUint(val uint64) {
	count := binary.PutUvarint(writer.numSlice, val)
	writer.buf.Write(writer.numSlice[:count])
}

// String writes a string.
func (writer *Writer) String(val string) {
	writer.VariableUint(uint64(len(val)))
	writer.buf.WriteString(val)
}

// Id writes an internal ID.
func (writer *Writer) Id(val id.Id) {
	writer.String(string(val))
}

// BoolSlice writes a bool slice.
func (writer *Writer) BoolSlice(vals []bool) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Bool(vals[i])
	}
}

// Int8Slice writes an int8 slice.
func (writer *Writer) Int8Slice(vals []int8) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Int8(vals[i])
	}
}

// Int16Slice writes an int16 slice.
func (writer *Writer) Int16Slice(vals []int16) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Int16(vals[i])
	}
}

// Int32Slice writes an int32 slice.
func (writer *Writer) Int32Slice(vals []int32) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Int32(vals[i])
	}
}

// Int64Slice writes an int64 slice.
func (writer *Writer) Int64Slice(vals []int64) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Int64(vals[i])
	}
}

// Uint8Slice writes a uint8 slice.
func (writer *Writer) Uint8Slice(vals []uint8) {
	writer.VariableUint(uint64(len(vals)))
	writer.buf.Write(vals)
}

// Uint16Slice writes a uint16 slice.
func (writer *Writer) Uint16Slice(vals []uint16) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Uint16(vals[i])
	}
}

// Uint32Slice writes a uint32 slice.
func (writer *Writer) Uint32Slice(vals []uint32) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Uint32(vals[i])
	}
}

// Uint64Slice writes a uint64 slice.
func (writer *Writer) Uint64Slice(vals []uint64) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Uint64(vals[i])
	}
}

// ByteSlice writes a byte slice. This is equivalent to Uint8Slice, but is included since it is more common to refer to
// byte slices than uint8 slices.
func (writer *Writer) ByteSlice(vals []byte) {
	writer.Uint8Slice(vals)
}

// Float32Slice writes a float32 slice.
func (writer *Writer) Float32Slice(vals []float32) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Float32(vals[i])
	}
}

// Float64Slice writes a float64 slice.
func (writer *Writer) Float64Slice(vals []float64) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.Float64(vals[i])
	}
}

// VariableIntSlice writes an int64 slice using variable-length encoding. Smaller values use less space at the cost of
// larger values using more space, but this is generally more space-efficient. This does carry a computational hit when
// reading.
func (writer *Writer) VariableIntSlice(vals []int64) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.VariableInt(vals[i])
	}
}

// VariableUintSlice writes a uint64 slice using variable-length encoding. Smaller values use less space at the cost of
// larger values using more space, but this is generally more space-efficient. This does carry a computational hit when
// reading.
func (writer *Writer) VariableUintSlice(vals []uint64) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.VariableUint(vals[i])
	}
}

// StringSlice writes a string slice.
func (writer *Writer) StringSlice(vals []string) {
	writer.VariableUint(uint64(len(vals)))
	for i := range vals {
		writer.String(vals[i])
	}
}

// Data returns the data written to the Writer.
func (writer *Writer) Data() []byte {
	return writer.buf.Bytes()
}

// Reset resets the Writer to be empty, but it retains the underlying storage for use by future writes.
func (writer *Writer) Reset() {
	writer.buf.Reset()
}
