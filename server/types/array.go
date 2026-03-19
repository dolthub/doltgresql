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

package types

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// CreateArrayTypeFromBaseType create array type from given type.
func CreateArrayTypeFromBaseType(baseType *DoltgresType) *DoltgresType {
	align := TypeAlignment_Int
	if baseType.Align == TypeAlignment_Double {
		align = TypeAlignment_Double
	}
	return &DoltgresType{
		ID:                  baseType.Array,
		TypLength:           int16(-1),
		PassedByVal:         false,
		TypType:             TypeType_Base,
		TypCategory:         TypeCategory_ArrayTypes,
		IsPreferred:         false,
		IsDefined:           true,
		Delimiter:           ",",
		RelID:               id.Null,
		SubscriptFunc:       toFuncID("array_subscript_handler", toInternal("internal")),
		Elem:                baseType.ID,
		Array:               id.NullType,
		InputFunc:           toFuncID("array_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
		OutputFunc:          toFuncID("array_out", toInternal("anyarray")),
		ReceiveFunc:         toFuncID("array_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
		SendFunc:            toFuncID("array_send", toInternal("anyarray")),
		ModInFunc:           baseType.ModInFunc,
		ModOutFunc:          baseType.ModOutFunc,
		AnalyzeFunc:         toFuncID("array_typanalyze", toInternal("internal")),
		Align:               align,
		Storage:             TypeStorage_Extended,
		NotNull:             false,
		BaseTypeID:          id.NullType,
		TypMod:              -1,
		NDims:               0,
		TypCollation:        baseType.TypCollation,
		DefaulBin:           "",
		Default:             "",
		Acl:                 nil,
		Checks:              nil,
		InternalName:        fmt.Sprintf("%s[]", baseType.Name()), // This will be set to the proper name in ToArrayType
		attTypMod:           baseType.attTypMod,                   // TODO: check
		CompareFunc:         toFuncID("btarraycmp", toInternal("anyarray"), toInternal("anyarray")),
		SerializationFunc:   serializeTypeArray,
		DeserializationFunc: deserializeTypeArray,
	}
}

// LogicalArrayElementTypes is a map of array element types for particular array types where the logical type varies
// from the declared type, as needed. Some types that have a NULL element for pg_catalog compatibility have a logical
// type that we need during analysis for function calls.
var LogicalArrayElementTypes = map[id.Type]*DoltgresType{
	toInternal("anyarray"): AnyElement,
}

// serializeTypeArray handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeArray(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	return serializeArray(ctx, val.([]any), t.ArrayBaseType())
}

// deserializeTypeArray handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeArray(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	return deserializeArray(ctx, data, t.ArrayBaseType())
}

// deserializeArray serializes an array of given base type.
func serializeArray(ctx *sql.Context, vals []any, baseType *DoltgresType) ([]byte, error) {
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
}

// deserializeArray deserializes an array of given base type.
func deserializeArray(ctx *sql.Context, data []byte, baseType *DoltgresType) ([]any, error) {
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
}
