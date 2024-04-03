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

// BaseID implements the DoltgresType interface.
func (ac arrayContainer) BaseID() DoltgresTypeBaseID {
	// The serializationID might be enough, but it's technically possible for us to use the same serialization ID with
	// different inner types, so this ensures uniqueness. It is safe to change base IDs in the future (unlike
	// serialization IDs, which must never be changed, only added to), so we can change this at any time if we feel it
	// is necessary to.
	return (DoltgresTypeBaseID(ac.serializationID) << 16) | ac.innerType.BaseID()
}

// BaseType implements the DoltgresArrayType interface.
func (ac arrayContainer) BaseType() DoltgresType {
	return ac.innerType
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
		return 0, sql.ErrInvalidType.New(ac)
	}
	bb, ok := v2.([]any)
	if !ok {
		return 0, sql.ErrInvalidType.New(ac)
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
	if val == nil {
		return nil, sql.InRange, nil
	}
	valSlice, ok := val.([]any)
	if !ok {
		return nil, sql.OutOfRange, sql.ErrInvalidType.New(ac)
	}
	// TODO: should we create a new slice or update the current slice? New slice every time seems wasteful
	newSlice := make([]any, len(valSlice))
	for i := range valSlice {
		var err error
		newSlice[i], _, err = ac.innerType.Convert(valSlice[i])
		if err != nil {
			return nil, sql.OutOfRange, err
		}
	}
	return newSlice, sql.InRange, nil
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
	converted, _, err := ac.Convert(val)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	for i, v := range converted.([]any) {
		if i > 0 {
			sb.WriteString(", ")
		}
		if v != nil {
			str, err := ac.innerType.FormatValue(v)
			if err != nil {
				return "", err
			}
			sb.WriteString(str)
		} else {
			sb.WriteString("NULL")
		}
	}
	return sb.String(), nil
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

// SerializeType implements the DoltgresType interface.
func (ac arrayContainer) SerializeType() ([]byte, error) {
	innerSerialized, err := ac.innerType.SerializeType()
	if err != nil {
		return nil, err
	}
	serialized := make([]byte, len(innerSerialized)+2)
	binary.LittleEndian.PutUint16(serialized, uint16(ac.serializationID))
	copy(serialized[2:], innerSerialized)
	return serialized, nil
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

// SerializeValue implements the DoltgresType interface.
func (ac arrayContainer) SerializeValue(valInterface any) ([]byte, error) {
	// Check for a nil value and convert to the expected type
	if valInterface == nil {
		return nil, nil
	}
	converted, _, err := ac.Convert(valInterface)
	if err != nil {
		return nil, err
	}
	vals := converted.([]any)

	// Create the buffer that we'll write the array's contents to
	arrayBuffer := bytes.Buffer{}
	innerSerializedWidth := ac.innerType.MaxSerializedWidth()
	// Write the total length to a buffer. We'll reuse this buffer for all uint-to-bytes operations.
	lengthBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(lengthBuffer, uint64(len(vals)))
	arrayBuffer.Write(lengthBuffer)

	// Each value is serialized as the following: IsNull, Size, Data
	// IsNull is one byte that represents whether the value is null. If this is true/1, then Size and Data are absent.
	// Size is 2 or 8 bytes and states the size of the Data. It is valid for this to equal zero for some types (like strings).
	// Data contains the actual data representing a value.
	for i := range vals {
		val, err := ac.innerType.SerializeValue(vals[i])
		if err != nil {
			return nil, err
		}

		if val == nil {
			arrayBuffer.WriteByte(1)
		} else {
			arrayBuffer.WriteByte(0)
			switch innerSerializedWidth {
			case types.ExtendedTypeSerializedWidth_64K:
				binary.BigEndian.PutUint16(lengthBuffer, uint16(len(val)))
				arrayBuffer.Write(lengthBuffer[:2])
			case types.ExtendedTypeSerializedWidth_Unbounded:
				binary.BigEndian.PutUint64(lengthBuffer, uint64(len(val)))
				arrayBuffer.Write(lengthBuffer)
			default:
				return nil, fmt.Errorf("array type encountered unexpected serializable width")
			}
			arrayBuffer.Write(val)
		}
	}
	return arrayBuffer.Bytes(), nil
}

// DeserializeValue implements the DoltgresType interface.
func (ac arrayContainer) DeserializeValue(val []byte) (any, error) {
	// Check for the nil value, then ensure the minimum length of the slice
	if val == nil {
		return nil, nil
	}
	if len(val) < 8 {
		return nil, fmt.Errorf("deserializing non-nil array value has invalid length of %d", len(val))
	}
	// Read the length and construct the output slice
	length := binary.BigEndian.Uint64(val)
	val = val[8:]
	output := make([]any, length)
	innerSerializedWidth := ac.innerType.MaxSerializedWidth()
	// TODO: faster/better to remove length checks and defer-recover for panics?
	for i := range output {
		// Check if the value is null
		if len(val) < 1 {
			return nil, fmt.Errorf("deserializing array encountered missing null check for value")
		}
		if val[0] == 1 {
			// output is filled with nil values on initialization, so we only need to move the val slice forward
			val = val[1:]
			continue
		}
		val = val[1:]

		// Read the length of the data for this value
		var dataLength uint64
		switch innerSerializedWidth {
		case types.ExtendedTypeSerializedWidth_64K:
			if len(val) < 2 {
				return nil, fmt.Errorf("deserializing array encountered missing size for value")
			}
			dataLength = uint64(binary.BigEndian.Uint16(val))
			val = val[2:]
		case types.ExtendedTypeSerializedWidth_Unbounded:
			if len(val) < 8 {
				return nil, fmt.Errorf("deserializing array encountered missing size for value")
			}
			dataLength = binary.BigEndian.Uint64(val)
			val = val[8:]
		default:
			return nil, fmt.Errorf("array type encountered unexpected serializable width")
		}
		if uint64(len(val)) < dataLength {
			return nil, fmt.Errorf("deserializing array encountered size too large for data")
		}

		// Read the data using the length from the previous step
		deserializedValue, err := ac.innerType.DeserializeValue(val[:dataLength])
		if err != nil {
			return nil, err
		}
		val = val[dataLength:]
		output[i] = deserializedValue
	}

	// Make sure that we read everything
	if len(val) > 0 {
		return nil, fmt.Errorf("deserialized array has extra data at the end")
	}
	return output, nil
}

// withInnerDeserialization implements the DoltgresArrayType interface.
func (ac arrayContainer) withInnerDeserialization(innerSerializedType []byte) (types.ExtendedType, error) {
	innerType, err := DeserializeType(innerSerializedType[2:])
	if err != nil {
		return nil, err
	}
	return arrayContainer{
		innerType:       innerType.(DoltgresType),
		serializationID: ac.serializationID,
		oid:             ac.oid,
		funcs:           ac.funcs,
	}, nil
}

// arrayContainerSQL implements the default SQL function for arrayContainer.
func arrayContainerSQL(ctx *sql.Context, ac arrayContainer, dest []byte, valInterface any) (sqltypes.Value, error) {
	if valInterface == nil {
		return sqltypes.NULL, nil
	}
	converted, _, err := ac.Convert(valInterface)
	if err != nil {
		return sqltypes.Value{}, err
	}
	vals := converted.([]any)
	if len(vals) == 0 {
		return sqltypes.MakeTrusted(ac.Type(), types.AppendAndSliceBytes(dest, []byte{'{', '}'})), nil
	}
	bb := bytes.Buffer{}
	bb.WriteRune('{')
	for i := range vals {
		if i > 0 {
			bb.WriteRune(',')
		}
		if vals[i] == nil {
			bb.WriteString("NULL")
			continue
		}
		valBytes, err := ac.innerType.SQL(ctx, nil, vals[i])
		if err != nil {
			return sqltypes.Value{}, err
		}
		bb.Write(valBytes.Raw())
	}
	bb.WriteRune('}')
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, bb.Bytes())), nil
}
