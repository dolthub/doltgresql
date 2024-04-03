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
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
	"github.com/shopspring/decimal"
)

// Int16 is an int16.
var Int16 = Int16Type{}

// Int16Type is the extended type implementation of the PostgreSQL smallint.
type Int16Type struct{}

var _ DoltgresType = Int16Type{}

// BaseID implements the DoltgresType interface.
func (b Int16Type) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID(SerializationID_Int16)
}

// CollationCoercibility implements the DoltgresType interface.
func (b Int16Type) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b Int16Type) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(int16)
	bb := bc.(int16)
	if ab == bb {
		return 0, nil
	} else if ab < bb {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (b Int16Type) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case bool:
		if val {
			return int16(1), sql.InRange, nil
		}
		return int16(0), sql.InRange, nil
	case int:
		return int16(val), sql.InRange, nil
	case uint:
		return int16(val), sql.InRange, nil
	case int8:
		return int16(val), sql.InRange, nil
	case uint8:
		return int16(val), sql.InRange, nil
	case int16:
		return int16(val), sql.InRange, nil
	case uint16:
		return int16(val), sql.InRange, nil
	case int32:
		return int16(val), sql.InRange, nil
	case uint32:
		return int16(val), sql.InRange, nil
	case int64:
		return int16(val), sql.InRange, nil
	case uint64:
		return int16(val), sql.InRange, nil
	case float32:
		return int16(val), sql.InRange, nil
	case float64:
		return int16(val), sql.InRange, nil
	case decimal.NullDecimal:
		if !val.Valid {
			return nil, sql.InRange, nil
		}
		return b.Convert(val.Decimal)
	case decimal.Decimal:
		v, _ := val.Float64()
		return int16(v), sql.InRange, nil
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, sql.OutOfRange, err
		}
		return int16(i), sql.InRange, nil
	case []byte:
		return b.Convert(string(val))
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b Int16Type) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b Int16Type) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b Int16Type) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(int64(converted.(int16)), 10), nil
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b Int16Type) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b Int16Type) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 2
}

// OID implements the DoltgresType interface.
func (b Int16Type) OID() uint32 {
	return uint32(oid.T_int2)
}

// Promote implements the DoltgresType interface.
func (b Int16Type) Promote() sql.Type {
	return Int32
}

// SerializedCompare implements the DoltgresType interface.
func (b Int16Type) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	return bytes.Compare(v1, v2), nil
}

// SerializeType implements the DoltgresType interface.
func (b Int16Type) SerializeType() ([]byte, error) {
	return SerializationID_Int16.ToByteSlice(), nil
}

// SQL implements the DoltgresType interface.
func (b Int16Type) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.FormatValue(v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresType interface.
func (b Int16Type) String() string {
	return "smallint"
}

// ToArrayType implements the DoltgresType interface.
func (b Int16Type) ToArrayType() DoltgresArrayType {
	return Int16Array
}

// Type implements the DoltgresType interface.
func (b Int16Type) Type() query.Type {
	return sqltypes.Int16
}

// ValueType implements the DoltgresType interface.
func (b Int16Type) ValueType() reflect.Type {
	return reflect.TypeOf(int16(0))
}

// Zero implements the DoltgresType interface.
func (b Int16Type) Zero() any {
	return int16(0)
}

// SerializeValue implements the DoltgresType interface.
func (b Int16Type) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	retVal := make([]byte, 2)
	binary.BigEndian.PutUint16(retVal, uint16(converted.(int16))+(1<<15))
	return retVal, nil
}

// DeserializeValue implements the DoltgresType interface.
func (b Int16Type) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return int16(binary.BigEndian.Uint16(val) - (1 << 15)), nil
}
