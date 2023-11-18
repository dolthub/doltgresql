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

package connection

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// decodeBuffer provides a way to track how much of a buffer has been used, which is useful when decoding a message,
// determining the next buffer point after the current message, and resetting the buffer in the event of an error.
type decodeBuffer struct {
	data        []byte
	nextBuffer  []byte
	resetBuffer []byte
	skipHeader  bool
}

// advance moves the buffer forward by the given amount.
func (db *decodeBuffer) advance(n int32) {
	db.data = db.data[n:]
	db.nextBuffer = db.nextBuffer[n:]
}

// setDataLength sets the length of the data buffer to the given amount.
func (db *decodeBuffer) setDataLength(n int32) {
	if n > int32(len(db.data)) {
		n = int32(len(db.data))
	} else if n < 0 {
		n = 0
	}
	db.data = db.data[:n]
}

// next replaces the current data buffer with the nextBuffer buffer, while updating the resetBuffer buffer.
func (db *decodeBuffer) next() {
	db.data = db.nextBuffer
	db.resetBuffer = db.nextBuffer
}

// reset replaces the current data and nextBuffer buffers with the resetBuffer buffer.
func (db *decodeBuffer) reset() {
	db.data = db.resetBuffer
	db.nextBuffer = db.resetBuffer
}

// copy returns a copy of this decodeBuffer.
func (db *decodeBuffer) copy() *decodeBuffer {
	return &decodeBuffer{
		data:        db.data,
		nextBuffer:  db.nextBuffer,
		resetBuffer: db.resetBuffer,
	}
}

// newDecodeBuffer returns a new *decodeBuffer.
func newDecodeBuffer(buffer []byte) *decodeBuffer {
	return &decodeBuffer{
		data:        buffer,
		nextBuffer:  buffer,
		resetBuffer: buffer,
	}
}

// decode writes the contents of the buffer into the given fields. The iteration count determines how many times the
// fields will be looped over.
func decode(buffer *decodeBuffer, fields []FieldGroup, iterations int32) error {
	for iteration := int32(0); iteration < iterations; iteration++ {
		for i, field := range fields[iteration] {
			// Some calls to decode will have already processed the message header and length to determine the message type,
			// so skip those fields when decoding.
			if buffer.skipHeader &&
				(field.Flags&Header != 0 || field.Flags&MessageLengthInclusive != 0) {
				continue
			}

			if len(buffer.data) == 0 {
				return errors.New("buffer too small")
			}

			switch field.Type {
			case Byte1, Int8:
				data := int32(buffer.data[0])
				if field.Flags&StaticData != 0 && field.Data.(int32) != data {
					return errors.New("static data differs from the buffer data")
				}
				field.Data = data
				buffer.advance(1)
			case ByteN:
				if i > 0 && fields[iteration][i-1].Flags&ByteCount != 0 {
					byteCount := fields[iteration][i-1].Data.(int32)
					// -1 is a valid value for byte counts, which is used to signal a NULL value.
					// We don't need to care about the assumption, so we can just treat it equivalent to zero.
					if byteCount == -1 {
						byteCount = 0
					}
					data := make([]byte, byteCount)
					copy(data, buffer.data)
					if field.Flags&StaticData != 0 && bytes.Compare(field.Data.([]byte), data) != 0 {
						return errors.New("static data differs from the buffer data")
					}
					field.Data = data
					buffer.advance(byteCount)
				} else {
					data := make([]byte, len(buffer.data))
					copy(data, buffer.data)
					if field.Flags&StaticData != 0 && bytes.Compare(field.Data.([]byte), data) != 0 {
						return errors.New("static data differs from the buffer data")
					}
					field.Data = data
					buffer.advance(int32(len(buffer.data)))
				}
			case Int16:
				data := int32(binary.BigEndian.Uint16(buffer.data))
				if field.Flags&StaticData != 0 && field.Data.(int32) != data {
					return errors.New("static data differs from the buffer data")
				}
				field.Data = data
				buffer.advance(2)
			case Int32:
				data := int32(binary.BigEndian.Uint32(buffer.data))
				if field.Flags&StaticData != 0 && field.Data.(int32) != data {
					return errors.New("static data differs from the buffer data")
				}
				field.Data = data
				buffer.advance(4)
			case String:
				found := false
				for bufferIdx := range buffer.data {
					if buffer.data[bufferIdx] == 0 {
						data := string(buffer.data[:bufferIdx])
						if field.Flags&StaticData != 0 && field.Data.(string) != data {
							return errors.New("static data differs from the buffer data")
						}
						field.Data = data
						buffer.advance(int32(bufferIdx))
						if field.Flags&ExcludeTerminator == 0 {
							buffer.advance(1)
						}
						found = true
						break
					}
				}
				if !found {
					return errors.New("terminating zero not found for string")
				}
			case Repeated:
				// Track if we've decoded at least once, so that we only update the count if we've decoded something
				decodedAtLeastOnce := false
				originalChildren := field.Copy().Children[0]
				for i := 1; len(buffer.data) > 0; i++ {
					// If there is only a single byte left, then it may be the terminator, so we check.
					// Otherwise, we'll assume that we should pass it to the child.
					if len(buffer.data) == 1 && field.Flags&RepeatedTerminator != 0 {
						if buffer.data[0] == 0 {
							buffer.advance(1)
							break
						} else {
							return fmt.Errorf("Expected terminator after Repeated type, found invalid byte: %d", buffer.data[0])
						}
					}
					field.extend(i, originalChildren)
					if err := decode(buffer, field.Children[len(field.Children)-1:], 1); err != nil {
						return err
					}
					decodedAtLeastOnce = true
				}
				if decodedAtLeastOnce {
					field.Data = int32(len(field.Children))
				}
			default:
				panic("message type has not been defined")
			}

			if field.Flags&MessageLengthInclusive != 0 {
				messageLength := field.Data.(int32)
				switch field.Type {
				case Byte1, Int8:
					messageLength -= 1
				case Int16:
					messageLength -= 2
				case Int32:
					messageLength -= 4
				}
				if messageLength > int32(len(buffer.data)) {
					return errors.New("message length is greater than the buffer size")
				}
				buffer.setDataLength(messageLength)
			} else if field.Flags&MessageLengthExclusive != 0 {
				messageLength := field.Data.(int32)
				if messageLength > int32(len(buffer.data)) {
					return errors.New("message length is greater than the buffer size")
				}
				buffer.setDataLength(messageLength)
			}
			if len(field.Children) > 0 && field.Type != Repeated {
				count, ok := field.Data.(int32)
				if !ok {
					return errors.New("non-integer is being used as a count")
				}
				// Counts may be negative numbers, which have a special meaning depending on the message.
				// In all such cases, they'll never have children, so we can just check for cases where it's > 0.
				if count > 0 {
					field.extend(int(count), field.Children[0])
					if err := decode(buffer, field.Children, count); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// encode transforms the message into a byte slice, which may be sent to a connection.
func encode(ms MessageFormat) ([]byte, error) {
	buffer := bytes.Buffer{}
	encodeLoop(&buffer, []FieldGroup{ms.Fields}, 1)
	if ms.info.appendNullByte {
		buffer.WriteByte(0)
	}
	data := buffer.Bytes()

	// Find and write the message length
	byteOffset := int32(0)
	for i, field := range ms.Fields {
		if field.Flags&(MessageLengthInclusive|MessageLengthExclusive) != 0 {
			typeLength := int32(0)
			// Exclusive lengths must take their own type size into account and exclude them from the overall length
			if field.Flags&MessageLengthExclusive != 0 {
				switch field.Type {
				case Byte1, Int8:
					typeLength = 1
				case Int16:
					typeLength = 2
				case Int32:
					typeLength = 4
				}
			}
			messageLength := int32(len(data)) - byteOffset - typeLength
			switch field.Type {
			case Byte1, Int8:
				data[byteOffset] = byte(messageLength)
			case Int16:
				binary.BigEndian.PutUint16(data[byteOffset:], uint16(messageLength))
			case Int32:
				binary.BigEndian.PutUint32(data[byteOffset:], uint32(messageLength))
			default:
				panic("invalid type for message length")
			}
			break
		}

		// Advance the offset
		switch field.Type {
		case Byte1, Int8:
			byteOffset += 1
		case ByteN:
			if i > 0 && ms.Fields[i-1].Flags&ByteCount != 0 {
				byteOffset += ms.Fields[i-1].Data.(int32)
			} else {
				byteOffset = int32(len(data)) // Last field, so we can set it to the remaining data
			}
		case Int16:
			byteOffset += 2
		case Int32:
			byteOffset += 4
		case String:
			found := false
			for bufferIdx := range data[byteOffset:] {
				if data[bufferIdx] == 0 {
					found = true
					byteOffset += int32(bufferIdx)
					if field.Flags&ExcludeTerminator == 0 {
						byteOffset += 1
					}
					break
				}
			}
			if !found {
				return nil, errors.New("terminating zero not found for string")
			}
		case Repeated:
			byteOffset = int32(len(data)) // Last field, so we can set it to the remaining data
		default:
			panic("message type has not been defined")
		}
	}
	return data, nil
}

// encodeLoop is the inner recursive loop of encode, which writes the given fields into the buffer. The iteration
// count determines how many times the fields are looped over.
func encodeLoop(buffer *bytes.Buffer, fields []FieldGroup, iterations int32) {
	for iteration := int32(0); iteration < iterations; iteration++ {
		for _, field := range fields[iteration] {
			switch field.Type {
			case Byte1:
				_ = binary.Write(buffer, binary.BigEndian, byte(field.Data.(int32)))
			case ByteN:
				buffer.Write(field.Data.([]byte))
			case Int8:
				_ = binary.Write(buffer, binary.BigEndian, int8(field.Data.(int32)))
			case Int16:
				_ = binary.Write(buffer, binary.BigEndian, int16(field.Data.(int32)))
			case Int32:
				_ = binary.Write(buffer, binary.BigEndian, field.Data.(int32))
			case String:
				buffer.WriteString(field.Data.(string))
				if field.Flags&ExcludeTerminator == 0 {
					buffer.WriteByte(0)
				}
			case Repeated:
				// We don't write anything for repeated fields, since they repeat their children until the end
			default:
				panic("message type has not been defined")
			}

			if len(field.Children) > 0 {
				encodeLoop(buffer, field.Children, field.Data.(int32))
			}
		}
	}
}
