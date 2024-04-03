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

// Int32 is an int32.
var Int32 = Int32Type{}

// Int32Type is the extended type implementation of the PostgreSQL integer.
type Int32Type struct{}

var _ DoltgresType = Int32Type{}

// BaseID implements the DoltgresType interface.
func (b Int32Type) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID(SerializationID_Int32)
}

// CollationCoercibility implements the DoltgresType interface.
func (b Int32Type) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b Int32Type) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(int32)
	bb := bc.(int32)
	if ab == bb {
		return 0, nil
	} else if ab < bb {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (b Int32Type) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case bool:
		if val {
			return int32(1), sql.InRange, nil
		}
		return int32(0), sql.InRange, nil
	case int:
		return int32(val), sql.InRange, nil
	case uint:
		return int32(val), sql.InRange, nil
	case int8:
		return int32(val), sql.InRange, nil
	case uint8:
		return int32(val), sql.InRange, nil
	case int16:
		return int32(val), sql.InRange, nil
	case uint16:
		return int32(val), sql.InRange, nil
	case int32:
		return int32(val), sql.InRange, nil
	case uint32:
		return int32(val), sql.InRange, nil
	case int64:
		return int32(val), sql.InRange, nil
	case uint64:
		return int32(val), sql.InRange, nil
	case float32:
		return int32(val), sql.InRange, nil
	case float64:
		return int32(val), sql.InRange, nil
	case decimal.NullDecimal:
		if !val.Valid {
			return nil, sql.InRange, nil
		}
		return b.Convert(val.Decimal)
	case decimal.Decimal:
		return int32(val.IntPart()), sql.InRange, nil
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, sql.OutOfRange, err
		}
		return int32(i), sql.InRange, nil
	case []byte:
		return b.Convert(string(val))
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b Int32Type) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b Int32Type) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b Int32Type) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(int64(converted.(int32)), 10), nil
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b Int32Type) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b Int32Type) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 4
}

// OID implements the DoltgresType interface.
func (b Int32Type) OID() uint32 {
	return uint32(oid.T_int4)
}

// Promote implements the DoltgresType interface.
func (b Int32Type) Promote() sql.Type {
	return Int64
}

// SerializedCompare implements the DoltgresType interface.
func (b Int32Type) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	return bytes.Compare(v1, v2), nil
}

// SQL implements the DoltgresType interface.
func (b Int32Type) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b Int32Type) String() string {
	return "integer"
}

// Type implements the DoltgresType interface.
func (b Int32Type) Type() query.Type {
	return sqltypes.Int32
}

// ValueType implements the DoltgresType interface.
func (b Int32Type) ValueType() reflect.Type {
	return reflect.TypeOf(int32(0))
}

// Zero implements the DoltgresType interface.
func (b Int32Type) Zero() any {
	return int32(0)
}

// SerializeValue implements the DoltgresType interface.
func (b Int32Type) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	retVal := make([]byte, 4)
	binary.BigEndian.PutUint32(retVal, uint32(converted.(int32))+(1<<31))
	return retVal, nil
}

// DeserializeValue implements the DoltgresType interface.
func (b Int32Type) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return int32(binary.BigEndian.Uint32(val) - (1 << 31)), nil
}
