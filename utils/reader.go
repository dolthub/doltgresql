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
	"encoding/binary"
	"math"

	"github.com/dolthub/doltgresql/core/id"
)

// Reader handles type-safe reading from a byte slice, which was created by a Writer. This is not safe for concurrent
// use.
type Reader struct {
	buf    []byte
	offset uint64
}

// NewReader creates a new Reader that will read from the given data.
func NewReader(data []byte) *Reader {
	return &Reader{
		buf:    data,
		offset: 0,
	}
}

// Bool reads a bool.
func (reader *Reader) Bool() bool {
	reader.offset += 1
	if reader.buf[reader.offset-1] == 1 {
		return true
	} else {
		return false
	}
}

// Int8 reads an int8.
func (reader *Reader) Int8() int8 {
	return int8(reader.Uint8() - (1 << 7))
}

// Int16 reads an int16.
func (reader *Reader) Int16() int16 {
	return int16(reader.Uint16() - (1 << 15))
}

// Int32 reads an int32.
func (reader *Reader) Int32() int32 {
	return int32(reader.Uint32() - (1 << 31))
}

// Int64 reads an int64.
func (reader *Reader) Int64() int64 {
	return int64(reader.Uint64() - (1 << 63))
}

// Uint8 reads a uint8.
func (reader *Reader) Uint8() uint8 {
	reader.offset += 1
	return reader.buf[reader.offset-1]
}

// Uint16 reads a uint16.
func (reader *Reader) Uint16() uint16 {
	reader.offset += 2
	return binary.BigEndian.Uint16(reader.buf[reader.offset-2:])
}

// Uint32 reads a uint32.
func (reader *Reader) Uint32() uint32 {
	reader.offset += 4
	return binary.BigEndian.Uint32(reader.buf[reader.offset-4:])
}

// Uint64 reads a uint64.
func (reader *Reader) Uint64() uint64 {
	reader.offset += 8
	return binary.BigEndian.Uint64(reader.buf[reader.offset-8:])
}

// Byte reads a byte. This is equivalent to Uint8, but is included since it is more common to refer to a byte rather
// than a uint8.
func (reader *Reader) Byte() byte {
	return reader.Uint8()
}

// Float32 reads a float32.
func (reader *Reader) Float32() float32 {
	// For more details, look at Writer.Float32
	uval := reader.Uint32()
	uval ^= 0x80000000
	uval ^= uint32(int32(uval)>>31) & 0x7FFFFFFF
	return math.Float32frombits(uval)
}

// Float64 reads a float64.
func (reader *Reader) Float64() float64 {
	// For more details, look at Writer.Float64
	uval := reader.Uint64()
	uval ^= 0x8000000000000000
	uval ^= uint64(int64(uval)>>63) & 0x7FFFFFFFFFFFFFFF
	return math.Float64frombits(uval)
}

// VariableInt reads an int64 that was written using variable-length encoding.
func (reader *Reader) VariableInt() int64 {
	uval := reader.VariableUint()
	// binary.PutVarint performs the inverse of this and writes a uint64, so we're undoing it to get the original int64
	val := int64(uval >> 1)
	if uval&1 != 0 {
		val = ^val
	}
	return val
}

// VariableUint reads a uint64 that was written using variable-length encoding.
func (reader *Reader) VariableUint() uint64 {
	// This has been adapted from one of our blogs:
	// https://www.dolthub.com/blog/2021-01-08-optimizing-varint-decoding/
	b := uint64(reader.buf[reader.offset])
	if b < 0x80 {
		reader.offset += 1
		return b
	}
	x := b & 0x7f
	b = uint64(reader.buf[reader.offset+1])
	if b < 0x80 {
		reader.offset += 2
		return x | (b << 7)
	}
	x |= (b & 0x7f) << 7
	b = uint64(reader.buf[reader.offset+2])
	if b < 0x80 {
		reader.offset += 3
		return x | (b << 14)
	}
	x |= (b & 0x7f) << 14
	b = uint64(reader.buf[reader.offset+3])
	if b < 0x80 {
		reader.offset += 4
		return x | (b << 21)
	}
	x |= (b & 0x7f) << 21
	b = uint64(reader.buf[reader.offset+4])
	if b < 0x80 {
		reader.offset += 5
		return x | (b << 28)
	}
	x |= (b & 0x7f) << 28
	b = uint64(reader.buf[reader.offset+5])
	if b < 0x80 {
		reader.offset += 6
		return x | (b << 35)
	}
	x |= (b & 0x7f) << 35
	b = uint64(reader.buf[reader.offset+6])
	if b < 0x80 {
		reader.offset += 7
		return x | (b << 42)
	}
	x |= (b & 0x7f) << 42
	b = uint64(reader.buf[reader.offset+7])
	if b < 0x80 {
		reader.offset += 8
		return x | (b << 49)
	}
	x |= (b & 0x7f) << 49
	b = uint64(reader.buf[reader.offset+8])
	if b < 0x80 {
		reader.offset += 9
		return x | (b << 56)
	}
	x |= (b & 0x7f) << 56
	b = uint64(reader.buf[reader.offset+9])
	if b < 0x80 {
		reader.offset += 10
		return x | (b << 63)
	}
	reader.offset += 10
	return 0xffffffffffffffff
}

// String reads a string.
func (reader *Reader) String() string {
	length := reader.VariableUint()
	reader.offset += length
	return string(reader.buf[reader.offset-length : reader.offset])
}

// Id reads an internal ID.
func (reader *Reader) Id() id.Id {
	return id.Id(reader.String())
}

// BoolSlice reads a bool slice.
func (reader *Reader) BoolSlice() []bool {
	count := reader.VariableUint()
	vals := make([]bool, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Bool()
	}
	return vals
}

// Int8Slice reads an int8 slice.
func (reader *Reader) Int8Slice() []int8 {
	count := reader.VariableUint()
	vals := make([]int8, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Int8()
	}
	return vals
}

// Int16Slice reads an int16 slice.
func (reader *Reader) Int16Slice() []int16 {
	count := reader.VariableUint()
	vals := make([]int16, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Int16()
	}
	return vals
}

// Int32Slice reads an int32 slice.
func (reader *Reader) Int32Slice() []int32 {
	count := reader.VariableUint()
	vals := make([]int32, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Int32()
	}
	return vals
}

// Int64Slice reads an int64 slice.
func (reader *Reader) Int64Slice() []int64 {
	count := reader.VariableUint()
	vals := make([]int64, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Int64()
	}
	return vals
}

// Uint8Slice reads a uint8 slice.
func (reader *Reader) Uint8Slice() []uint8 {
	count := reader.VariableUint()
	vals := make([]uint8, count)
	copy(vals, reader.buf[reader.offset:reader.offset+count])
	reader.offset += count
	return vals
}

// Uint16Slice reads a uint16 slice.
func (reader *Reader) Uint16Slice() []uint16 {
	count := reader.VariableUint()
	vals := make([]uint16, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Uint16()
	}
	return vals
}

// Uint32Slice reads a uint32 slice.
func (reader *Reader) Uint32Slice() []uint32 {
	count := reader.VariableUint()
	vals := make([]uint32, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Uint32()
	}
	return vals
}

// Uint64Slice reads a uint64 slice.
func (reader *Reader) Uint64Slice() []uint64 {
	count := reader.VariableUint()
	vals := make([]uint64, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Uint64()
	}
	return vals
}

// ByteSlice reads a byte slice. This is equivalent to Uint8Slice, but is included since it is more common to refer to
// byte slices than uint8 slices.
func (reader *Reader) ByteSlice() []byte {
	return reader.Uint8Slice()
}

// Float32Slice reads a float32 slice.
func (reader *Reader) Float32Slice() []float32 {
	count := reader.VariableUint()
	vals := make([]float32, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Float32()
	}
	return vals
}

// Float64Slice reads a float64 slice.
func (reader *Reader) Float64Slice() []float64 {
	count := reader.VariableUint()
	vals := make([]float64, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.Float64()
	}
	return vals
}

// VariableIntSlice reads an int64 slice that was written using variable-length encoding.
func (reader *Reader) VariableIntSlice() []int64 {
	count := reader.VariableUint()
	vals := make([]int64, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.VariableInt()
	}
	return vals
}

// VariableUintSlice reads a uint64 slice that was written using variable-length encoding.
func (reader *Reader) VariableUintSlice() []uint64 {
	count := reader.VariableUint()
	vals := make([]uint64, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.VariableUint()
	}
	return vals
}

// StringSlice reads a string slice.
func (reader *Reader) StringSlice() []string {
	count := reader.VariableUint()
	vals := make([]string, count)
	for i := uint64(0); i < count; i++ {
		vals[i] = reader.String()
	}
	return vals
}

// IsEmpty returns true when all of the data has been read.
func (reader *Reader) IsEmpty() bool {
	return reader.offset >= uint64(len(reader.buf))
}

// RemainingBytes returns the number of bytes that have not yet been read.
func (reader *Reader) RemainingBytes() uint64 {
	if reader.IsEmpty() {
		return 0
	}
	return uint64(len(reader.buf)) - reader.offset
}

// BytesRead returns the number of bytes that have been read.
func (reader *Reader) BytesRead() uint64 {
	return reader.offset
}

// AdvanceReader reads the next N bytes from the given Reader. This is only available for specific, performance-oriented
// circumstances, and should never be used otherwise. This is a standalone function to discourage its use, as it will
// not show up as a function of the Reader object in most IDEs. This uses a branchless comparison to limit the size of n
// to the end of the reader, and it does not allocate a new byte slice (returns a portion from the original byte slice).
func AdvanceReader(reader *Reader, n uint64) []byte {
	// This branchless code makes an assumption that the reader contains less than 9223372036854775808 bytes.
	// With this assumption, it is equivalent to the following:
	// if reader.offset + n > uint64(len(reader.buf)) || n >= 0x8000000000000000 {
	//     n = uint64(len(reader.buf)) - reader.offset
	// }
	maxN := uint64(len(reader.buf)) - reader.offset
	delta := int64(n - maxN)
	mask := (delta >> 63) ^ (int64(n&0x8000000000000000) >> 63)
	n = (n & uint64(mask)) | (maxN & ^uint64(mask))

	reader.offset += n
	return reader.buf[reader.offset-n : reader.offset]
}
