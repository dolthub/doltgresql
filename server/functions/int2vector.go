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

package functions

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInt2vector registers the functions to the catalog.
func initInt2vector() {
	framework.RegisterFunction(int2vectorin)
	framework.RegisterFunction(int2vectorout)
	framework.RegisterFunction(int2vectorrecv)
	framework.RegisterFunction(int2vectorsend)
}

// int2vectorin represents the PostgreSQL function of int2vector type IO input.
var int2vectorin = framework.Function1{
	Name:       "int2vectorin",
	Return:     pgtypes.Int16vector,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		strValues := strings.Split(input, " ")
		var values = make([]any, len(strValues))
		for i, strValue := range strValues {
			innerValue, err := pgtypes.Int16.IoInput(ctx, strValue)
			if err != nil {
				return nil, err
			}
			values[i] = innerValue.(int16)
		}
		return values, nil
	},
}

// int2vectorout represents the PostgreSQL function of int2vector type IO output.
var int2vectorout = framework.Function1{
	Name:       "int2vectorout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int16vector},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.VectorToString(ctx, val.([]any), pgtypes.Int16, false)
	},
}

// int2vectorrecv represents the PostgreSQL function of int2vector type IO receive.
var int2vectorrecv = framework.Function1{
	Name:       "int2vectorrecv",
	Return:     pgtypes.Int16vector,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		baseType := pgtypes.Int16
		// Check for the nil value, then ensure the minimum length of the slice
		if len(data) == 0 {
			return nil, nil
		}
		if len(data) < 4 {
			return nil, errors.Errorf("deserializing non-nil array value has invalid length of %d", len(data))
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
			o, err := baseType.DeserializeValue(ctx, data[offset+1:nextOffset])
			if err != nil {
				return nil, err
			}
			output[i] = o
		}
		// Returns all read elements
		return output, nil
	},
}

// int2vectorsend represents the PostgreSQL function of int2vector type IO send.
var int2vectorsend = framework.Function1{
	Name:       "int2vectorsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int16vector},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		baseType := pgtypes.Int16
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
			serializedVal, err := baseType.SerializeValue(ctx, vals[i])
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
