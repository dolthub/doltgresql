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

package utils

import (
	"bytes"
	"encoding/binary"
	"math"
)

// WireRW handles all read and write operations for the Postgres binary wire format. This should only be used for
// interacting with the wire format, and not for internal use (utils.Reader and utils.Writer exist for internal use).
type WireRW struct {
	buf      *bytes.Buffer
	readIdx  uint32
	numSlice []byte
}

// NewWireWriter creates a new WireRW for writing to. This is intended strictly for writing to the binary wire format
// that Postgres expects, and should not be used for internal write operations (utils.Writer is for internal use).
func NewWireWriter() *WireRW {
	return &WireRW{
		buf:      bytes.NewBuffer(make([]byte, 0, 8)),
		readIdx:  0,
		numSlice: make([]byte, 8), // 8 bytes will cover all floats and integers
	}
}

// NewWireReader creates a new WireRW for reading from data previously written by either a valid Postgres server or
// a WireRW instance. This is intended strictly for reading the binary wire format that Postgres expects, and should not
// be used for internal read operations (utils.Reader is for internal use).
func NewWireReader(data []byte) *WireRW {
	return &WireRW{
		buf:      bytes.NewBuffer(data),
		readIdx:  0,
		numSlice: make([]byte, 8), // 8 bytes will cover all floats and integers
	}
}

// ReadBool reads a 1-byte bool.
func (rw *WireRW) ReadBool() bool {
	return rw.ReadInt8() != 0
}

// ReadInt8 reads an int8.
func (rw *WireRW) ReadInt8() int8 {
	return int8(rw.ReadUint8())
}

// ReadInt16 reads an int16.
func (rw *WireRW) ReadInt16() int16 {
	return int16(rw.ReadUint16())
}

// ReadInt32 reads an int32.
func (rw *WireRW) ReadInt32() int32 {
	return int32(rw.ReadUint32())
}

// ReadInt64 reads an int64.
func (rw *WireRW) ReadInt64() int64 {
	return int64(rw.ReadUint64())
}

// ReadUint8 reads a uint8.
func (rw *WireRW) ReadUint8() uint8 {
	data := rw.buf.Bytes()[rw.readIdx]
	rw.readIdx += 1
	return data
}

// ReadUint16 reads a uint16.
func (rw *WireRW) ReadUint16() uint16 {
	data := rw.buf.Bytes()[rw.readIdx : rw.readIdx+2]
	rw.readIdx += 2
	return binary.BigEndian.Uint16(data)
}

// ReadUint32 reads a uint32.
func (rw *WireRW) ReadUint32() uint32 {
	data := rw.buf.Bytes()[rw.readIdx : rw.readIdx+4]
	rw.readIdx += 4
	return binary.BigEndian.Uint32(data)
}

// ReadUint64 reads a uint64.
func (rw *WireRW) ReadUint64() uint64 {
	data := rw.buf.Bytes()[rw.readIdx : rw.readIdx+8]
	rw.readIdx += 8
	return binary.BigEndian.Uint64(data)
}

// ReadFloat32 reads a float32.
func (rw *WireRW) ReadFloat32() float32 {
	return math.Float32frombits(rw.ReadUint32())
}

// ReadFloat64 reads a float64.
func (rw *WireRW) ReadFloat64() float64 {
	return math.Float64frombits(rw.ReadUint64())
}

// ReadString reads the next N bytes as a string.
func (rw *WireRW) ReadString(n uint32) string {
	return string(rw.ReadBytes(n))
}

// ReadBytes reads the next N bytes.
func (rw *WireRW) ReadBytes(n uint32) []byte {
	data := rw.buf.Bytes()[rw.readIdx : rw.readIdx+n]
	rw.readIdx += n
	return data
}

// WriteBool writes a 1-byte bool.
func (rw *WireRW) WriteBool(val bool) *WireRW {
	if val {
		rw.buf.WriteByte(1)
	} else {
		rw.buf.WriteByte(0)
	}
	return rw
}

// WriteInt8 writes an int8.
func (rw *WireRW) WriteInt8(val int8) *WireRW {
	rw.WriteUint8(byte(val))
	return rw
}

// WriteInt16 writes an int16.
func (rw *WireRW) WriteInt16(val int16) *WireRW {
	rw.WriteUint16(uint16(val))
	return rw
}

// WriteInt32 writes an int32.
func (rw *WireRW) WriteInt32(val int32) *WireRW {
	rw.WriteUint32(uint32(val))
	return rw
}

// WriteInt64 writes an int64.
func (rw *WireRW) WriteInt64(val int64) *WireRW {
	rw.WriteUint64(uint64(val))
	return rw
}

// WriteUint8 writes a uint8.
func (rw *WireRW) WriteUint8(val uint8) *WireRW {
	rw.buf.WriteByte(val)
	return rw
}

// WriteUint16 writes a uint16.
func (rw *WireRW) WriteUint16(val uint16) *WireRW {
	binary.BigEndian.PutUint16(rw.numSlice, val)
	rw.buf.Write(rw.numSlice[:2])
	return rw
}

// WriteUint32 writes a uint32.
func (rw *WireRW) WriteUint32(val uint32) *WireRW {
	binary.BigEndian.PutUint32(rw.numSlice, val)
	rw.buf.Write(rw.numSlice[:4])
	return rw
}

// WriteUint64 writes a uint64.
func (rw *WireRW) WriteUint64(val uint64) *WireRW {
	binary.BigEndian.PutUint64(rw.numSlice, val)
	rw.buf.Write(rw.numSlice[:8])
	return rw
}

// WriteFloat32 writes a float32.
func (rw *WireRW) WriteFloat32(val float32) *WireRW {
	rw.WriteUint32(math.Float32bits(val))
	return rw
}

// WriteFloat64 writes a float64.
func (rw *WireRW) WriteFloat64(val float64) *WireRW {
	rw.WriteUint64(math.Float64bits(val))
	return rw
}

// WriteString writes the raw string bytes.
func (rw *WireRW) WriteString(val string) *WireRW {
	_, _ = rw.buf.WriteString(val)
	return rw
}

// WriteBytes writes the bytes given.
func (rw *WireRW) WriteBytes(val []byte) *WireRW {
	_, _ = rw.buf.Write(val)
	return rw
}

// Reserve ensures that there are at least N bytes in the buffer, which can prevent reallocations if the space needed is
// known upfront.
func (rw *WireRW) Reserve(n uint64) {
	rw.buf.Grow(int(n))
}

// BufferSize returns the number of bytes in the buffer.
func (rw *WireRW) BufferSize() uint64 {
	return uint64(rw.buf.Len())
}

// BufferData returns the data in the buffer. This does not make a copy.
func (rw *WireRW) BufferData() []byte {
	return rw.buf.Bytes()
}
