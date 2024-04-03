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
	"fmt"
	"math/big"
	"reflect"

	"github.com/lib/pq/oid"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/shopspring/decimal"
)

var (
	NumericValueMaxInt16 = decimal.NewFromInt(32767)                // NumericValueMaxInt16 is the max Int16 value for NUMERIC types
	NumericValueMaxInt32 = decimal.NewFromInt(2147483647)           // NumericValueMaxInt16 is the max Int32 value for NUMERIC types
	NumericValueMaxInt64 = decimal.NewFromInt(9223372036854775807)  // NumericValueMaxInt16 is the max Int64 value for NUMERIC types
	NumericValueMinInt16 = decimal.NewFromInt(-32768)               // NumericValueMaxInt16 is the min Int16 value for NUMERIC types
	NumericValueMinInt32 = decimal.NewFromInt(-2147483648)          // NumericValueMaxInt16 is the min Int32 value for NUMERIC types
	NumericValueMinInt64 = decimal.NewFromInt(-9223372036854775808) // NumericValueMaxInt16 is the min Int64 value for NUMERIC types
)

// Numeric is a precise and unbounded decimal value.
var Numeric = NumericType{}

// NumericType is the extended type implementation of the PostgreSQL numeric.
type NumericType struct{}

var _ DoltgresType = NumericType{}

// BaseID implements the DoltgresType interface.
func (b NumericType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID(SerializationID_Numeric)
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
	case bool:
		if val {
			return decimal.NewFromInt(1), sql.InRange, nil
		}
		return decimal.NewFromInt(0), sql.InRange, nil
	case int:
		return decimal.NewFromInt(int64(val)), sql.InRange, nil
	case uint:
		return decimal.NewFromInt(int64(val)), sql.InRange, nil
	case int8:
		return decimal.NewFromInt(int64(val)), sql.InRange, nil
	case uint8:
		return decimal.NewFromInt(int64(val)), sql.InRange, nil
	case int16:
		return decimal.NewFromInt(int64(val)), sql.InRange, nil
	case uint16:
		return decimal.NewFromInt(int64(val)), sql.InRange, nil
	case int32:
		return decimal.NewFromInt(int64(val)), sql.InRange, nil
	case uint32:
		return decimal.NewFromInt(int64(val)), sql.InRange, nil
	case int64:
		return decimal.NewFromInt(val), sql.InRange, nil
	case uint64:
		return decimal.NewFromBigInt(new(big.Int).SetUint64(val), 1), sql.InRange, nil
	case float32:
		return decimal.NewFromFloat(float64(val)), sql.InRange, nil
	case float64:
		return decimal.NewFromFloat(val), sql.InRange, nil
	case decimal.NullDecimal:
		if !val.Valid {
			return nil, sql.InRange, nil
		}
		return val.Decimal, sql.InRange, nil
	case decimal.Decimal:
		return val, sql.InRange, nil
	case string:
		d, err := decimal.NewFromString(val)
		return d, sql.InRange, err
	case []byte:
		return b.Convert(string(val))
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

// FormatSerializedValue implements the DoltgresType interface.
func (b NumericType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b NumericType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return "", err
	}
	return converted.(decimal.Decimal).String(), nil
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b NumericType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K //TODO: probably should have inline and ref versions
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

// SerializeType implements the DoltgresType interface.
func (b NumericType) SerializeType() ([]byte, error) {
	return SerializationID_Numeric.ToByteSlice(), nil
}

// SQL implements the DoltgresType interface.
func (b NumericType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.FormatValue(v)
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

// ValueType implements the DoltgresType interface.
func (b NumericType) ValueType() reflect.Type {
	return reflect.TypeOf(decimal.Zero)
}

// Zero implements the DoltgresType interface.
func (b NumericType) Zero() any {
	return decimal.Zero
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
