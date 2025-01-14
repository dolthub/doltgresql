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

package functions

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initArray registers the functions to the catalog.
func initArray() {
	framework.RegisterFunction(array_in)
	framework.RegisterFunction(array_out)
	framework.RegisterFunction(array_recv)
	framework.RegisterFunction(array_send)
	framework.RegisterFunction(btarraycmp)
	framework.RegisterFunction(array_subscript_handler)
}

// array_in represents the PostgreSQL function of array type IO input.
var array_in = framework.Function3{
	Name:       "array_in",
	Return:     pgtypes.AnyArray,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		baseTypeOid := val2.(id.Id)
		baseType := pgtypes.IDToBuiltInDoltgresType[id.Type(baseTypeOid)]
		typmod := val3.(int32)
		baseType = baseType.WithAttTypMod(typmod)
		if len(input) < 2 || input[0] != '{' || input[len(input)-1] != '}' {
			// This error is regarded as a critical error, and thus we immediately return the error alongside a nil
			// value. Returning a nil value is a signal to not ignore the error.
			return nil, fmt.Errorf(`malformed array literal: "%s"`, input)
		}
		// We'll remove the surrounding braces since we've already verified that they're there
		input = input[1 : len(input)-1]
		var values []any
		var err error
		sb := strings.Builder{}
		quoteStartCount := 0
		quoteEndCount := 0
		escaped := false
		// Iterate over each rune in the input to collect and process the rune elements
		for _, r := range input {
			if escaped {
				sb.WriteRune(r)
				escaped = false
			} else if quoteStartCount > quoteEndCount {
				switch r {
				case '\\':
					escaped = true
				case '"':
					quoteEndCount++
				default:
					sb.WriteRune(r)
				}
			} else {
				switch r {
				case ' ', '\t', '\n', '\r':
					continue
				case '\\':
					escaped = true
				case '"':
					quoteStartCount++
				case ',':
					if quoteStartCount >= 2 {
						// This is a malformed string, thus we treat it as a critical error.
						return nil, fmt.Errorf(`malformed array literal: "%s"`, input)
					}
					str := sb.String()
					var innerValue any
					if quoteStartCount == 0 && strings.EqualFold(str, "null") {
						// An unquoted case-insensitive NULL is treated as an actual null value
						innerValue = nil
					} else {
						var nErr error
						innerValue, nErr = baseType.IoInput(ctx, str)
						if nErr != nil && err == nil {
							// This is a non-critical error, therefore the error may be ignored at a higher layer (such as
							// an explicit cast) and the inner type will still return a valid result, so we must allow the
							// values to propagate.
							err = nErr
						}
					}
					values = append(values, innerValue)
					sb.Reset()
					quoteStartCount = 0
					quoteEndCount = 0
				default:
					sb.WriteRune(r)
				}
			}
		}
		// Use anything remaining in the buffer as the last element
		if sb.Len() > 0 {
			if escaped || quoteStartCount > quoteEndCount || quoteStartCount >= 2 {
				// These errors are regarded as critical errors, and thus we immediately return the error alongside a nil
				// value. Returning a nil value is a signal to not ignore the error.
				return nil, fmt.Errorf(`malformed array literal: "%s"`, input)
			} else {
				str := sb.String()
				var innerValue any
				if quoteStartCount == 0 && strings.EqualFold(str, "NULL") {
					// An unquoted case-insensitive NULL is treated as an actual null value
					innerValue = nil
				} else {
					var nErr error
					innerValue, nErr = baseType.IoInput(ctx, str)
					if nErr != nil && err == nil {
						// This is a non-critical error, therefore the error may be ignored at a higher layer (such as
						// an explicit cast) and the inner type will still return a valid result, so we must allow the
						// values to propagate.
						err = nErr
					}
				}
				values = append(values, innerValue)
			}
		}

		return values, err
	},
}

// array_out represents the PostgreSQL function of array type IO output.
var array_out = framework.Function1{
	Name:       "array_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		arrType := t[0]
		baseType := arrType.ArrayBaseType()
		return pgtypes.ArrToString(ctx, val.([]any), baseType, false)
	},
}

// array_recv represents the PostgreSQL function of array type IO receive.
var array_recv = framework.Function3{
	Name:       "array_recv",
	Return:     pgtypes.AnyArray,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		baseTypeOid := val2.(id.Id)
		baseType := pgtypes.IDToBuiltInDoltgresType[id.Type(baseTypeOid)]
		typmod := val3.(int32)
		baseType = baseType.WithAttTypMod(typmod)
		// Check for the nil value, then ensure the minimum length of the slice
		if len(data) == 0 {
			return nil, nil
		}
		if len(data) < 4 {
			return nil, fmt.Errorf("deserializing non-nil array value has invalid length of %d", len(data))
		}
		// Grab the number of elements and construct an output slice of the appropriate size
		elementCount := binary.LittleEndian.Uint32(data)
		output := make([]any, elementCount)
		// Read all elements
		for i := uint32(0); i < elementCount; i++ {
			// We read from i+1 to account for the element count at the beginning
			offset := binary.LittleEndian.Uint32(data[(i+1)*4:])
			// If the value is null, then we can skip it, since the output slice default initializes all values to nil
			if data[offset] == 1 {
				continue
			}
			// The element data is everything from the offset to the next offset, excluding the null determinant
			nextOffset := binary.LittleEndian.Uint32(data[(i+2)*4:])
			o, err := baseType.DeserializeValue(data[offset+1 : nextOffset])
			if err != nil {
				return nil, err
			}
			output[i] = o
		}
		// Returns all read elements
		return output, nil
	},
}

// array_send represents the PostgreSQL function of array type IO send.
var array_send = framework.Function1{
	Name:       "array_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		arrType := t[0]
		baseType := arrType.ArrayBaseType()
		vals := val.([]any)

		bb := bytes.Buffer{}
		// Write the element count to a buffer. We're using an array since it's stack-allocated, so no need for pooling.
		var elementCount [4]byte
		binary.LittleEndian.PutUint32(elementCount[:], uint32(len(vals)))
		bb.Write(elementCount[:])
		// Create an array that contains the offsets for each value. Since we can't update the offset portion of the buffer
		// as we determine the offsets, we have to track them outside the buffer. We'll overwrite the buffer later with the
		// correct offsets. The last offset represents the end of the slice, which simplifies the logic for reading elements
		// using the "current offset to next offset" strategy. We use a byte slice since the buffer only works with byte
		// slices.
		offsets := make([]byte, (len(vals)+1)*4)
		bb.Write(offsets)
		// The starting offset for the first element is Count(uint32) + (NumberOfElementOffsets * sizeof(uint32))
		currentOffset := uint32(4 + (len(vals)+1)*4)
		for i := range vals {
			// Write the current offset
			binary.LittleEndian.PutUint32(offsets[i*4:], currentOffset)
			// Handle serialization of the value
			// TODO: ARRAYs may be multidimensional, such as ARRAY[[4,2],[6,3]], which isn't accounted for here
			serializedVal, err := baseType.SerializeValue(vals[i])
			if err != nil {
				return nil, err
			}
			// Handle the nil case and non-nil case
			if serializedVal == nil {
				bb.WriteByte(1)
				currentOffset += 1
			} else {
				bb.WriteByte(0)
				bb.Write(serializedVal)
				currentOffset += 1 + uint32(len(serializedVal))
			}
		}
		// Write the final offset, which will equal the length of the serialized slice
		binary.LittleEndian.PutUint32(offsets[len(offsets)-4:], currentOffset)
		// Get the final output, and write the updated offsets to it
		outputBytes := bb.Bytes()
		copy(outputBytes[4:], offsets)
		return outputBytes, nil
	},
}

// btarraycmp represents the PostgreSQL function of array type byte compare.
var btarraycmp = framework.Function2{
	Name:       "btarraycmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		at := t[0]
		bt := t[1]
		if !at.Equals(bt) {
			// TODO: currently, types should match.
			// Technically, does not have to e.g.: float4 vs float8
			return nil, fmt.Errorf("different type comparison is not supported yet")
		}

		ab := val1.([]any)
		bb := val2.([]any)
		minLength := utils.Min(len(ab), len(bb))
		for i := 0; i < minLength; i++ {
			res, err := at.ArrayBaseType().Compare(ab[i], bb[i])
			if err != nil {
				return 0, err
			}
			if res != 0 {
				return res, nil
			}
		}
		if len(ab) == len(bb) {
			return int32(0), nil
		} else if len(ab) < len(bb) {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// array_subscript_handler represents the PostgreSQL function of array type subscript handler.
var array_subscript_handler = framework.Function1{
	Name:       "array_subscript_handler",
	Return:     pgtypes.Internal,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return []byte{}, nil
	},
}
