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

package server

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// writeBinaryWireData writes the data matching the given id.Type to the given writer. Technically, these should all be
// within the type's "send" function, however we incorrectly use those for internal serialization.
func writeBinaryWireData(ctx *sql.Context, t *pgtypes.DoltgresType, writer *WireRW, v interface{}) error {
	// TODO: move all of this to the "send" functions and change the existing "send" and "receive" functions to simply be
	//  our internal serialization logic
	if v == nil {
		return nil
	}
	// If the type is unresolved, then we need to resolve it first
	if t.IsUnresolved {
		typs, err := core.GetTypesCollectionFromContext(ctx)
		if err != nil {
			return err
		}
		resolvedType, err := typs.GetType(ctx, t.ID)
		if err != nil {
			return err
		}
		t = resolvedType
	}
	// Then check if the value is wrapped, so we can unwrap it first (if needed)
	if wrapper, ok := v.(sql.AnyWrapper); ok {
		var err error
		v, err = wrapper.UnwrapAny(ctx)
		if err != nil {
			return err
		}
		if v == nil {
			return nil
		}
	}
	// We handle array types separately, since they all follow the same core writing scheme
	if t.IsArrayType() {
		vals := v.([]any)
		// Check for nulls first
		hasNull := false
		for _, val := range vals {
			if val == nil {
				hasNull = true
				break
			}
		}
		// Count the number of dimensions
		dimensions := int32(0)
		innerVals := vals
		for len(innerVals) > 0 {
			dimensions++
			slice, ok := innerVals[0].([]any)
			if !ok {
				break
			}
			innerVals = slice
		}
		if dimensions > 1 {
			return errors.Errorf("arrays with %d dimensions are not yet supported using the binary format", dimensions)
		}
		writer.WriteInt32(dimensions) // Write the number of dimensions
		if hasNull {
			writer.WriteInt32(1)
		} else {
			writer.WriteInt32(0)
		}
		writer.WriteUint32(id.Cache().ToOID(t.ArrayBaseType().ID.AsId())) // Element OID
		for i := int32(0); i < dimensions; i++ {
			writer.WriteInt32(int32(len(vals))) // Elements in this dimension
			writer.WriteInt32(1)                // Lower bound, or what index number we start at (seems to always be 1?)
			for _, val := range vals {
				if val == nil {
					writer.WriteInt32(-1)
				} else {
					valWriter := NewWireWriter()
					if err := writeBinaryWireData(ctx, t.ArrayBaseType(), valWriter, val); err != nil {
						return err
					}
					valBytes := valWriter.BufferData()
					writer.WriteInt32(int32(len(valBytes)))
					writer.WriteBytes(valBytes)
				}
			}
		}
		return nil
	}
	// We also handle record and composite types separately
	if t.IsCompositeType() {
		recordVals := v.([]pgtypes.RecordValue)
		writer.WriteInt32(int32(len(recordVals)))
		for _, recordVal := range recordVals {
			switch recordType := recordVal.Type.(type) {
			case *pgtypes.DoltgresType:
				writer.WriteUint32(id.Cache().ToOID(recordType.ID.AsId()))
				if recordVal.Value != nil {
					valWriter := NewWireWriter()
					if err := writeBinaryWireData(ctx, recordType, valWriter, recordVal.Value); err != nil {
						return err
					}
					valBytes := valWriter.BufferData()
					writer.WriteInt32(int32(len(valBytes)))
					writer.WriteBytes(valBytes)
				} else {
					writer.WriteInt32(-1)
				}
			default:
				cast := pgexprs.NewGMSCast(expression.NewLiteral(recordVal.Value, recordType))
				writer.WriteUint32(id.Cache().ToOID(cast.DoltgresType().ID.AsId()))
				if recordVal.Value != nil {
					castVal, err := cast.Eval(ctx, nil)
					if err != nil {
						return err
					}
					valWriter := NewWireWriter()
					if err := writeBinaryWireData(ctx, cast.DoltgresType(), valWriter, castVal); err != nil {
						return err
					}
					valBytes := valWriter.BufferData()
					writer.WriteInt32(int32(len(valBytes)))
					writer.WriteBytes(valBytes)
				} else {
					writer.WriteInt32(-1)
				}
			}
		}
		return nil
	}
	switch t.TypType {
	// Finally, we handle the remaining user-defined types that aren't composite or table types
	case pgtypes.TypeType_Domain:
		// Domain types use their underlying type for serialization
		return writeBinaryWireData(ctx, t.DomainUnderlyingBaseType(), writer, v)
	case pgtypes.TypeType_Enum:
		// Enum types write their underlying text values directly
		writer.WriteString(v.(string))
		return nil
	}
	switch t.ID {
	case pgtypes.Bit.ID:
		// We process bits in chunks of 8, so we append zeroes until our string is evenly divisible by 8
		bitString := v.(string)
		if len(bitString)%8 != 0 {
			bitString += strings.Repeat("0", 8-(len(bitString)%8))
		}
		writer.Reserve(uint64(4 + (len(bitString) / 8)))
		writer.WriteInt32(t.GetAttTypMod())
		for bufIdx := 0; bufIdx < len(bitString); bufIdx += 8 {
			parsedByte, err := strconv.ParseUint(bitString[bufIdx:bufIdx+8], 2, 8)
			if err != nil {
				return errors.Errorf(`error encountered while converting "BIT" to binary wire format:\n%s`, err.Error())
			}
			writer.WriteUint8(byte(parsedByte))
		}
		return nil
	case pgtypes.Bool.ID:
		writer.WriteBool(v.(bool))
		return nil
	case pgtypes.BpChar.ID:
		str, err := t.IoOutput(ctx, v)
		if err != nil {
			return err
		}
		writer.WriteString(str)
		return nil
	case pgtypes.Bytea.ID:
		writer.WriteBytes(v.([]byte))
		return nil
	case pgtypes.Date.ID:
		postgresEpoch := time.UnixMilli(946684800000).UTC() // Jan 1, 2000 @ Midnight
		deltaInMilliseconds := v.(time.Time).UTC().UnixMilli() - postgresEpoch.UnixMilli()
		const millisecondsPerDay = 86400000
		days := uint32(deltaInMilliseconds / millisecondsPerDay)
		writer.WriteUint32(days)
		return nil
	case pgtypes.Float32.ID:
		writer.WriteFloat32(v.(float32))
		return nil
	case pgtypes.Float64.ID:
		writer.WriteFloat64(v.(float64))
		return nil
	case pgtypes.Int16.ID:
		writer.WriteInt16(v.(int16))
		return nil
	case pgtypes.Int32.ID:
		writer.WriteInt32(v.(int32))
		return nil
	case pgtypes.Int64.ID:
		writer.WriteInt64(v.(int64))
		return nil
	case pgtypes.InternalChar.ID:
		str := v.(string)
		if len(str) == 1 {
			writer.WriteUint8(str[0])
		} else if len(str) == 0 {
			writer.WriteUint8(0)
		} else {
			return errors.New(`"char" found multiple characters during binary wire formatting`)
		}
		return nil
	case pgtypes.Interval.ID:
		dur := v.(duration.Duration)
		writer.WriteInt64(dur.Nanos() / 1000)
		writer.WriteInt32(int32(dur.Days))
		writer.WriteInt32(int32(dur.Months))
		return nil
	case pgtypes.Json.ID:
		writer.WriteString(v.(string))
		return nil
	case pgtypes.JsonB.ID:
		textVal, err := t.SQL(ctx, nil, v)
		if err != nil {
			return err
		}
		writer.WriteUint8(1)
		writer.WriteBytes(textVal.ToBytes())
		return nil
	case pgtypes.Name.ID:
		writer.WriteString(v.(string))
		return nil
	case pgtypes.Numeric.ID:
		dec := v.(decimal.Decimal)
		// Short-circuit if this is the zero value
		if dec.IsZero() {
			writer.WriteBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0})
			return nil
		}
		// There's a way to do this more efficiently, but we can do that work once this becomes a performance issue.
		// This is based on the terminology used in Postgres' `numeric.c` file
		decStr := dec.String()
		isNegative := false
		if strings.HasPrefix(decStr, "-") {
			isNegative = true
			decStr = decStr[1:]
		}
		// Split the integer and fractional parts
		var intPart string
		var fractPart string
		if idx := strings.Index(decStr, "."); idx != -1 {
			intPart = decStr[:idx]
			fractPart = decStr[idx+1:]
		} else {
			intPart = decStr
		}
		// Find the "dscale", which is the number of digits in the fractional part
		typmod := t.GetAttTypMod()
		var dscale int16
		if typmod != -1 {
			_, dscale32 := pgtypes.GetPrecisionAndScaleFromTypmod(typmod)
			dscale = int16(dscale32)
		} else {
			dscale = int16(len(fractPart))
		}
		// Pad the integer and fractional parts so that we can take groups of 4 numbers
		if intPart == "0" {
			intPart = ""
		} else if len(intPart)%4 != 0 {
			intPart = strings.Repeat("0", 4-(len(intPart)%4)) + intPart
		}
		if len(fractPart)%4 != 0 {
			fractPart = fractPart + strings.Repeat("0", 4-(len(fractPart)%4))
		}
		// Write the "ndigits" first, or the number of base-10000 digits
		writer.WriteInt16(int16((len(intPart) / 4) + (len(fractPart) / 4)))
		// Write the "weight", which is the number of base-10000 digits in the integer part subtracted by 1
		writer.WriteInt16(int16((len(intPart) / 4) - 1))
		// Write the "sign"
		if isNegative {
			writer.WriteInt16(16384)
		} else {
			writer.WriteInt16(0)
		}
		// Write the "dscale"
		writer.WriteInt16(dscale)
		// Write all of the digits
		fullPart := intPart + fractPart
		for i := 0; i < len(fullPart); i += 4 {
			part, err := strconv.Atoi(fullPart[i : i+4])
			if err != nil {
				return err
			}
			writer.WriteInt16(int16(part))
		}
		return nil
	case pgtypes.Oid.ID:
		writer.WriteUint32(id.Cache().ToOID(v.(id.Id)))
		return nil
	case pgtypes.Regclass.ID:
		writer.WriteUint32(id.Cache().ToOID(v.(id.Id)))
		return nil
	case pgtypes.Regproc.ID:
		writer.WriteUint32(id.Cache().ToOID(v.(id.Id)))
		return nil
	case pgtypes.Regtype.ID:
		writer.WriteUint32(id.Cache().ToOID(v.(id.Id)))
		return nil
	case pgtypes.Text.ID:
		writer.WriteString(v.(string))
		return nil
	case pgtypes.Time.ID:
		writer.WriteInt64(v.(time.Time).UnixMicro())
		return nil
	case pgtypes.Timestamp.ID, pgtypes.TimestampTZ.ID:
		postgresEpoch := time.UnixMilli(946684800000).UTC() // Jan 1, 2000 @ Midnight
		deltaInMicroseconds := v.(time.Time).UTC().UnixMicro() - postgresEpoch.UnixMicro()
		writer.WriteInt64(deltaInMicroseconds)
		return nil
	case pgtypes.TimeTZ.ID:
		// We have to isolate the UTC time from the timezone, so we subtract the timezone delta from the original time
		tim := v.(time.Time)
		timezone, _ := strconv.Atoi(tim.Format("-070000"))
		isNegative := false
		if timezone < 0 {
			isNegative = true
			timezone = -timezone
		}
		seconds := timezone % 100
		minutes := (timezone / 100) % 100
		hours := (timezone / 10000) % 100
		totalSeconds := int32(seconds + (60 * minutes) + (3600 * hours))
		if !isNegative {
			totalSeconds = -totalSeconds // The sign is inverted when writing the integer
		}
		timeOffset := time.Duration(-totalSeconds) * time.Second // Adding a negative is the same as subtracting
		tim = tim.UTC().Add(timeOffset)
		writer.WriteInt64(tim.UnixMicro())
		writer.WriteInt32(totalSeconds)
		return nil
	case pgtypes.Uuid.ID:
		buf, err := v.(uuid.UUID).MarshalBinary()
		if err != nil {
			return err
		}
		writer.WriteBytes(buf)
		return nil
	case pgtypes.VarBit.ID:
		bitString := v.(string)
		originalLength := int32(len(bitString))
		// We process bits in chunks of 8, so we append zeroes until our string is evenly divisible by 8
		if len(bitString)%8 != 0 {
			bitString += strings.Repeat("0", 8-(len(bitString)%8))
		}
		writer.Reserve(uint64(4 + (len(bitString) / 8)))
		writer.WriteInt32(originalLength)
		for bufIdx := 0; bufIdx < len(bitString); bufIdx += 8 {
			parsedByte, err := strconv.ParseUint(bitString[bufIdx:bufIdx+8], 2, 8)
			if err != nil {
				return errors.Errorf(`error encountered while converting "VARBIT" to binary wire format:\n%s`, err.Error())
			}
			writer.WriteUint8(byte(parsedByte))
		}
		return nil
	case pgtypes.VarChar.ID:
		str, err := t.IoOutput(ctx, v)
		if err != nil {
			return err
		}
		writer.WriteString(str)
		return nil
	case pgtypes.Unknown.ID:
		str, ok := v.(string)
		if !ok {
			return errors.Errorf(`non-string value encountered while converting "UNKNOWN" to binary wire format:\n%T`, v)
		}
		writer.WriteString(str)
		return nil
	case pgtypes.Xid.ID:
		writer.WriteUint32(v.(uint32))
		return nil
	default:
		return errors.Errorf(`type "%s" does not implement a binary wire format, please open an issue: https://github.com/dolthub/doltgresql/issues`, t.Name())
	}
}

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
