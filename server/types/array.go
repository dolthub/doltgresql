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
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// CreateArrayTypeFromBaseType create array type from given type. This also sets the `Array` on the given base type if
// it has not already been set.
func CreateArrayTypeFromBaseType(baseType *DoltgresType) *DoltgresType {
	align := TypeAlignment_Int
	if baseType.Align == TypeAlignment_Double {
		align = TypeAlignment_Double
	}
	var arrayID id.Type
	if baseType.Array == nil || baseType.Array == internalNullType || baseType.Array.ID == id.NullType {
		arrayID = id.NewType(baseType.ID.SchemaName(), "_"+baseType.ID.TypeName())
	} else {
		arrayID = baseType.Array.ID
	}
	arrayType := &DoltgresType{
		ID:                  arrayID,
		TypLength:           int16(-1),
		PassedByVal:         false,
		TypType:             TypeType_Base,
		TypCategory:         TypeCategory_ArrayTypes,
		IsPreferred:         false,
		IsDefined:           true,
		Delimiter:           ",",
		RelID:               id.Null,
		SubscriptFunc:       toFuncID("array_subscript_handler", toInternal("internal")),
		Elem:                baseType,
		Array:               internalNullType,
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
		BaseTypeType:        internalNullType,
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
		serializedVersion:   arrayTypeVersion,
	}
	if baseType.Array == nil || baseType.Array == internalNullType || baseType.Array.ID == id.NullType {
		baseType.Array = arrayType
	}
	return arrayType
}

// serializeTypeArray handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeArray(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	vals := val.([]any)
	dims := ArrayDims(vals, t.ArrayBaseType())
	if t.serializedVersion == 0 && len(dims) > 1 {
		return nil, pgerror.WithCandidateCode(errors.New("multidimensional arrays are not supported by the column's type version, alter the column's type to upgrade it"), pgcode.FeatureNotSupported)
	}
	return serializeArray(ctx, vals, t.ArrayBaseType())
}

// deserializeTypeArray handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeArray(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	return deserializeArray(ctx, data, t.ArrayBaseType())
}

// serializeArray serializes an array of given base type. Multidimensional arrays are flattened, and the length of each
// dimension trails the elements.
func serializeArray(ctx *sql.Context, vals []any, baseType *DoltgresType) ([]byte, error) {
	dims := ArrayDims(vals, baseType)
	vals = FlattenArray(vals, baseType)
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
	// Write the final offset, which marks the end of the element data
	binary.LittleEndian.PutUint32(offsets[len(offsets)-4:], currentOffset)
	if len(dims) > 1 {
		for _, dim := range dims {
			var dimBytes [4]byte
			binary.LittleEndian.PutUint32(dimBytes[:], uint32(dim))
			bb.Write(dimBytes[:])
		}
	}
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
	// Any data after the final offset is the length of each dimension
	dimsOffset := binary.LittleEndian.Uint32(data[(elementCount+1)*4:])
	dims := make([]int32, (uint32(len(data))-dimsOffset)/4)
	for i := range dims {
		dims[i] = int32(binary.LittleEndian.Uint32(data[dimsOffset+uint32(i)*4:]))
	}
	return InflateArray(output, dims), nil
}

// ArrayDims returns the length of each dimension of the given array value, which is empty for an empty array.
// Elements of an array-like base type, such as a vector, are never treated as sub-arrays.
func ArrayDims(vals []any, baseType *DoltgresType) []int32 {
	if !baseType.IsArrayCategory() {
		return arrayDims(vals)
	}
	if len(vals) == 0 {
		return nil
	}
	return []int32{int32(len(vals))}
}

// arrayDims returns the length of each dimension of the given array value, treating every nested slice as a
// sub-array.
func arrayDims(vals []any) []int32 {
	var dims []int32
	for len(vals) > 0 {
		dims = append(dims, int32(len(vals)))
		subArray, ok := vals[0].([]any)
		if !ok {
			break
		}
		vals = subArray
	}
	return dims
}

// FlattenArray returns the elements of the given array value in row-major order. Elements of an array-like base type,
// such as a vector, are never treated as sub-arrays.
func FlattenArray(vals []any, baseType *DoltgresType) []any {
	if baseType.IsArrayCategory() {
		return vals
	}
	return flattenArray(vals)
}

// flattenArray returns the elements of the given array value in row-major order, treating every nested slice as a
// sub-array.
func flattenArray(vals []any) []any {
	if len(vals) == 0 {
		return vals
	}
	if _, ok := vals[0].([]any); !ok {
		return vals
	}
	elementCount := 1
	for _, dim := range arrayDims(vals) {
		elementCount *= int(dim)
	}
	return appendFlattened(make([]any, 0, elementCount), vals)
}

// appendFlattened appends the elements of the given array value to `flattened` in row-major order, treating every
// nested slice as a sub-array.
func appendFlattened(flattened []any, vals []any) []any {
	if len(vals) > 0 {
		if _, ok := vals[0].([]any); ok {
			for _, subArray := range vals {
				flattened = appendFlattened(flattened, subArray.([]any))
			}
			return flattened
		}
	}
	return append(flattened, vals...)
}

// InflateArray nests the given row-major elements into an array value with the given dimensions.
func InflateArray(flattened []any, dims []int32) []any {
	if len(dims) <= 1 || len(flattened) == 0 {
		return flattened
	}
	subArrayLength := len(flattened) / int(dims[0])
	vals := make([]any, dims[0])
	for i := range vals {
		vals[i] = InflateArray(flattened[i*subArrayLength:(i+1)*subArrayLength], dims[1:])
	}
	return vals
}

// ValidateAccumulatedArrays returns an error when the given array values cannot be accumulated as the sub-arrays of a
// new array.
func ValidateAccumulatedArrays(vals []any) error {
	var firstDims []int32
	for i, val := range vals {
		if val == nil {
			return pgerror.WithCandidateCode(errors.New("cannot accumulate null arrays"), pgcode.NullValueNotAllowed)
		}
		dims := arrayDims(val.([]any))
		if i == 0 {
			if len(dims) == 0 {
				return pgerror.WithCandidateCode(errors.New("cannot accumulate empty arrays"), pgcode.ArraySubscript)
			}
			firstDims = dims
		} else if !slices.Equal(dims, firstDims) {
			return pgerror.WithCandidateCode(errors.New("cannot accumulate arrays of different dimensionality"), pgcode.ArraySubscript)
		}
	}
	return nil
}

// SameArrayDims returns whether the elements of the given array value are either all scalars, or all sub-arrays with
// the same dimensions.
func SameArrayDims(vals []any) bool {
	if len(vals) == 0 {
		return true
	}
	first, firstIsSubArray := vals[0].([]any)
	firstDims := arrayDims(first)
	for _, val := range vals[1:] {
		subArray, isSubArray := val.([]any)
		if isSubArray != firstIsSubArray || !slices.Equal(arrayDims(subArray), firstDims) {
			return false
		}
	}
	return true
}
