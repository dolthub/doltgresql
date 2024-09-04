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
	"reflect"
	"strings"

	"github.com/lib/pq/oid"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/shopspring/decimal"
)

const (
	MaxUint32 = 4294967295  // MaxUint32 is the largest possible value of Uint32
	MinInt32  = -2147483648 // MinInt32 is the smallest possible value of Int32
)

var (
	NumericValueMaxInt16  = decimal.NewFromInt(32767)                // NumericValueMaxInt16 is the max Int16 value for NUMERIC types
	NumericValueMaxInt32  = decimal.NewFromInt(2147483647)           // NumericValueMaxInt32 is the max Int32 value for NUMERIC types
	NumericValueMaxInt64  = decimal.NewFromInt(9223372036854775807)  // NumericValueMaxInt64 is the max Int64 value for NUMERIC types
	NumericValueMinInt16  = decimal.NewFromInt(-32768)               // NumericValueMinInt16 is the min Int16 value for NUMERIC types
	NumericValueMinInt32  = decimal.NewFromInt(MinInt32)             // NumericValueMinInt32 is the min Int32 value for NUMERIC types
	NumericValueMinInt64  = decimal.NewFromInt(-9223372036854775808) // NumericValueMinInt64 is the min Int64 value for NUMERIC types
	NumericValueMaxUint32 = decimal.NewFromInt(MaxUint32)            // NumericValueMaxUint32 is the max Uint32 value for NUMERIC types
)

// Numeric is a precise and unbounded decimal value.
var Numeric = NumericType{-1, -1}

// NumericType is the extended type implementation of the PostgreSQL numeric.
type NumericType struct {
	// TODO: implement precision and scale
	Precision int32
	Scale     int32
}

var _ DoltgresType = NumericType{}

// Alignment implements the DoltgresType interface.
func (b NumericType) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresType interface.
func (b NumericType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Numeric
}

// BaseName implements the DoltgresType interface.
func (b NumericType) BaseName() string {
	return "numeric"
}

// Category implements the DoltgresType interface.
func (b NumericType) Category() TypeCategory {
	return TypeCategory_NumericTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b NumericType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b NumericType) Compare(v1 any, v2 any) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	ac, _, err := b.Convert(v1)
	if err != nil {
		return 0, err
	}
	bc, _, err := b.Convert(v2)
	if err != nil {
		return 0, err
	}

	ab := ac.(decimal.Decimal)
	bb := bc.(decimal.Decimal)
	return ab.Cmp(bb), nil
}

// Convert implements the DoltgresType interface.
func (b NumericType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case decimal.Decimal:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b NumericType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresType interface.
func (b NumericType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresType interface.
func (b NumericType) GetSerializationID() SerializationID {
	return SerializationID_Numeric
}

// IoInput implements the DoltgresType interface.
func (b NumericType) IoInput(ctx *sql.Context, input string) (any, error) {
	val, err := decimal.NewFromString(strings.TrimSpace(input))
	if err != nil {
		return nil, fmt.Errorf("invalid input syntax for type %s: %q", b.String(), input)
	}
	return val, nil
}

// IoOutput implements the DoltgresType interface.
func (b NumericType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return converted.(decimal.Decimal).String(), nil
}

// IsPreferredType implements the DoltgresType interface.
func (b NumericType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b NumericType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b NumericType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b NumericType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 65535
}

// OID implements the DoltgresType interface.
func (b NumericType) OID() uint32 {
	return uint32(oid.T_numeric)
}

// Promote implements the DoltgresType interface.
func (b NumericType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b NumericType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	ac, err := b.DeserializeValue(v1)
	if err != nil {
		return 0, err
	}
	bc, err := b.DeserializeValue(v2)
	if err != nil {
		return 0, err
	}
	ab := ac.(decimal.Decimal)
	bb := bc.(decimal.Decimal)
	return ab.Cmp(bb), nil
}

// SQL implements the DoltgresType interface.
func (b NumericType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.IoOutput(ctx, v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.VarChar, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresType interface.
func (b NumericType) String() string {
	return "numeric"
}

// ToArrayType implements the DoltgresType interface.
func (b NumericType) ToArrayType() DoltgresArrayType {
	return NumericArray
}

// Type implements the DoltgresType interface.
func (b NumericType) Type() query.Type {
	return sqltypes.Decimal
}

// ValToByteArray implements the DoltgresType interface.
func (b NumericType) ValToByteArray(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	return []byte(converted.(decimal.Decimal).String()), nil
}

// ValueType implements the DoltgresType interface.
func (b NumericType) ValueType() reflect.Type {
	return reflect.TypeOf(decimal.Zero)
}

// Zero implements the DoltgresType interface.
func (b NumericType) Zero() any {
	return decimal.Zero
}

// SerializeType implements the DoltgresType interface.
func (b NumericType) SerializeType() ([]byte, error) {
	t := make([]byte, serializationIDHeaderSize+8)
	copy(t, SerializationID_Numeric.ToByteSlice(0))
	binary.LittleEndian.PutUint32(t[serializationIDHeaderSize:], uint32(b.Precision))
	binary.LittleEndian.PutUint32(t[serializationIDHeaderSize+4:], uint32(b.Scale))
	return t, nil
}

// deserializeType implements the DoltgresType interface.
func (b NumericType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return NumericType{
			Precision: int32(binary.LittleEndian.Uint32(metadata)),
			Scale:     int32(binary.LittleEndian.Uint32(metadata[4:])),
		}, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b NumericType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	return converted.(decimal.Decimal).MarshalBinary()
}

// DeserializeValue implements the DoltgresType interface.
func (b NumericType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	retVal := decimal.NewFromInt(0)
	err := retVal.UnmarshalBinary(val)
	return retVal, err
}
