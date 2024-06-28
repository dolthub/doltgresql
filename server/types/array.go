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
	"math"
	"reflect"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/utils"
)

// arrayContainer is a type that wraps non-array types, giving them array functionality without requiring a bespoke
// implementation.
type arrayContainer struct {
	innerType       DoltgresType
	serializationID SerializationID
	oid             oid.Oid
	funcs           arrayContainerFunctions
}

// arrayContainerFunctions are overrides for the default array implementations of specific functions. If they are left
// nil, then it uses the default implementation.
type arrayContainerFunctions struct {
	// SQL is similar to the function with the same name that is found on sql.Type. This just takes an additional
	// arrayContainer parameter.
	SQL func(ctx *sql.Context, ac arrayContainer, dest []byte, valInterface any) (sqltypes.Value, error)
}

var _ DoltgresType = arrayContainer{}
var _ DoltgresArrayType = arrayContainer{}

// createArrayType creates an array variant of the given type. Uses the default array implementations for all possible
// overrides.
func createArrayType(innerType DoltgresType, serializationID SerializationID, arrayOid oid.Oid) DoltgresArrayType {
	return createArrayTypeWithFuncs(innerType, serializationID, arrayOid, arrayContainerFunctions{})
}

// createArrayTypeWithFuncs creates an array variant of the given type. Uses the provided function overrides if they're
// not nil. If any are nil, then they use the default array implementations.
func createArrayTypeWithFuncs(innerType DoltgresType, serializationID SerializationID, arrayOid oid.Oid, funcs arrayContainerFunctions) DoltgresArrayType {
	if funcs.SQL == nil {
		funcs.SQL = arrayContainerSQL
	}
	return arrayContainer{
		innerType:       innerType,
		serializationID: serializationID,
		oid:             arrayOid,
		funcs:           funcs,
	}
}

// Alignment implements the DoltgresType interface.
func (ac arrayContainer) Alignment() TypeAlignment {
	return ac.innerType.Alignment()
}

// BaseID implements the DoltgresType interface.
func (ac arrayContainer) BaseID() DoltgresTypeBaseID {
	// The serializationID might be enough, but it's technically possible for us to use the same serialization ID with
	// different inner types, so this ensures uniqueness. It is safe to change base IDs in the future (unlike
	// serialization IDs, which must never be changed, only added to), so we can change this at any time if we feel it
	// is necessary to.
	return (DoltgresTypeBaseID(ac.serializationID) << 16) | ac.innerType.BaseID()
}

// BaseName implements the DoltgresType interface.
func (ac arrayContainer) BaseName() string {
	return ac.innerType.BaseName()
}

// BaseType implements the DoltgresArrayType interface.
func (ac arrayContainer) BaseType() DoltgresType {
	return ac.innerType
}

// Category implements the DoltgresType interface.
func (ac arrayContainer) Category() TypeCategory {
	return ac.innerType.Category()
}

// CollationCoercibility implements the DoltgresType interface.
func (ac arrayContainer) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (ac arrayContainer) Compare(v1 any, v2 any) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	ab, ok := v1.([]any)
	if !ok {
		return 0, fmt.Errorf("%s: unhandled type: %T", ac.String(), v1)
	}
	bb, ok := v2.([]any)
	if !ok {
		return 0, fmt.Errorf("%s: unhandled type: %T", ac.String(), v2)
	}

	minLength := utils.Min(len(ab), len(bb))
	for i := 0; i < minLength; i++ {
		res, err := ac.innerType.Compare(ab[i], bb[i])
		if err != nil {
			return 0, err
		}
		if res != 0 {
			return res, nil
		}
	}
	if len(ab) == len(bb) {
		return 0, nil
	} else if len(ab) < len(bb) {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (ac arrayContainer) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case []any:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	case []string:
		anyArray := make([]any, len(val))
		for i, s := range val {
			anyArray[i] = s
		}
		return anyArray, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", ac.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (ac arrayContainer) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(ac), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (ac arrayContainer) FormatSerializedValue(val []byte) (string, error) {
	//TODO: write a far more optimized version of this that does not deserialize the entire array at once
	deserialized, err := ac.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return ac.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (ac arrayContainer) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return ac.IoOutput(val)
}

// GetSerializationID implements the DoltgresType interface.
func (ac arrayContainer) GetSerializationID() SerializationID {
	return ac.serializationID
}

// IoInput implements the DoltgresType interface.
func (ac arrayContainer) IoInput(input string) (any, error) {
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
					innerValue, nErr = ac.innerType.IoInput(str)
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
				innerValue, nErr = ac.innerType.IoInput(str)
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
}

// IoOutput implements the DoltgresType interface.
func (ac arrayContainer) IoOutput(output any) (string, error) {
	converted, _, err := ac.Convert(output)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	sb.WriteRune('{')
	for i, v := range converted.([]any) {
		if i > 0 {
			sb.WriteString(",")
		}
		if v != nil {
			str, err := ac.innerType.IoOutput(v)
			if err != nil {
				return "", err
			}
			shouldQuote := false
			for _, r := range str {
				switch r {
				case ' ', ',', '{', '}', '\\', '"':
					shouldQuote = true
				}
			}
			if shouldQuote || strings.EqualFold(str, "NULL") {
				sb.WriteRune('"')
				sb.WriteString(strings.ReplaceAll(str, `"`, `\"`))
				sb.WriteRune('"')
			} else {
				sb.WriteString(str)
			}
		} else {
			sb.WriteString("NULL")
		}
	}
	sb.WriteRune('}')
	return sb.String(), nil
}

// IsPreferredType implements the DoltgresType interface.
func (ac arrayContainer) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (ac arrayContainer) IsUnbounded() bool {
	return true
}

// MaxSerializedWidth implements the DoltgresType interface.
func (ac arrayContainer) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (ac arrayContainer) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresType interface.
func (ac arrayContainer) OID() uint32 {
	return uint32(ac.oid)
}

// Promote implements the DoltgresType interface.
func (ac arrayContainer) Promote() sql.Type {
	return ac
}

// SerializedCompare implements the DoltgresType interface.
func (ac arrayContainer) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	//TODO: write a far more optimized version of this that does not deserialize the entire arrays at once
	dv1, err := ac.DeserializeValue(v1)
	if err != nil {
		return 0, err
	}
	dv2, err := ac.DeserializeValue(v2)
	if err != nil {
		return 0, err
	}
	return ac.Compare(dv1, dv2)
}

// SQL implements the DoltgresType interface.
func (ac arrayContainer) SQL(ctx *sql.Context, dest []byte, valInterface any) (sqltypes.Value, error) {
	return ac.funcs.SQL(ctx, ac, dest, valInterface)
}

// String implements the DoltgresType interface.
func (ac arrayContainer) String() string {
	return ac.innerType.String() + "[]"
}

// ToArrayType implements the DoltgresType interface.
func (ac arrayContainer) ToArrayType() DoltgresArrayType {
	return ac
}

// Type implements the DoltgresType interface.
func (ac arrayContainer) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (ac arrayContainer) ValueType() reflect.Type {
	return reflect.TypeOf([]any{})
}

// Zero implements the DoltgresType interface.
func (ac arrayContainer) Zero() any {
	return []any{}
}

// SerializeType implements the DoltgresType interface.
func (ac arrayContainer) SerializeType() ([]byte, error) {
	innerSerialized, err := ac.innerType.SerializeType()
	if err != nil {
		return nil, err
	}
	serialized := make([]byte, serializationIDHeaderSize+len(innerSerialized))
	copy(serialized, ac.serializationID.ToByteSlice(0))
	copy(serialized[serializationIDHeaderSize:], innerSerialized)
	return serialized, nil
}

// deserializeType implements the DoltgresType interface.
func (ac arrayContainer) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		innerType, err := DeserializeType(metadata)
		if err != nil {
			return nil, err
		}
		return innerType.(DoltgresType).ToArrayType(), nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for arrays", version)
	}
}

// SerializeValue implements the DoltgresType interface.
func (ac arrayContainer) SerializeValue(valInterface any) ([]byte, error) {
	// The binary format is as follows:
	// The first value is always the number of serialized elements (uint32).
	// The next section contains offsets to the start of each element (uint32). There are N+1 offsets to elements.
	// The last offset contains the length of the slice.
	// The last section is the data section, where all elements store their data.
	// Each element comprises two values: a single byte stating if it's null, and the data itself.
	// You may determine the length of the data by using the following offset, as the data occupies all bytes up to the next offset.
	// The last element is a special case, as its data simply occupies all bytes up to the end of the slice.
	// The data may have a length of zero, which is distinct from null for some types.
	// In addition, a null value will always have a data length of zero.
	// This format allows for O(1) point lookups.

	// Check for a nil value and convert to the expected type
	if valInterface == nil {
		return nil, nil
	}
	converted, _, err := ac.Convert(valInterface)
	if err != nil {
		return nil, err
	}
	vals := converted.([]any)

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
		serializedVal, err := ac.innerType.SerializeValue(vals[i])
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

// DeserializeValue implements the DoltgresType interface.
func (ac arrayContainer) DeserializeValue(serializedVals []byte) (_ any, err error) {
	// Check for the nil value, then ensure the minimum length of the slice
	if serializedVals == nil {
		return nil, nil
	}
	if len(serializedVals) < 4 {
		return nil, fmt.Errorf("deserializing non-nil array value has invalid length of %d", len(serializedVals))
	}
	// Grab the number of elements and construct an output slice of the appropriate size
	elementCount := binary.LittleEndian.Uint32(serializedVals)
	output := make([]any, elementCount)
	// Read all elements
	for i := uint32(0); i < elementCount; i++ {
		// We read from i+1 to account for the element count at the beginning
		offset := binary.LittleEndian.Uint32(serializedVals[(i+1)*4:])
		// If the value is null, then we can skip it, since the output slice default initializes all values to nil
		if serializedVals[offset] == 1 {
			continue
		}
		// The element data is everything from the offset to the next offset, excluding the null determinant
		nextOffset := binary.LittleEndian.Uint32(serializedVals[(i+2)*4:])
		output[i], err = ac.innerType.DeserializeValue(serializedVals[offset+1 : nextOffset])
		if err != nil {
			return nil, err
		}
	}
	// Returns all of the read elements
	return output, nil
}

// arrayContainerSQL implements the default SQL function for arrayContainer.
func arrayContainerSQL(ctx *sql.Context, ac arrayContainer, dest []byte, value any) (sqltypes.Value, error) {
	str, err := ac.FormatValue(value)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(str))), nil
}
