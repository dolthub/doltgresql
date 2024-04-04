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
	"reflect"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
)

// Uuid is the UUID type.
var Uuid = UuidType{}

// UuidType is the extended type implementation of the PostgreSQL UUID.
type UuidType struct{}

var _ DoltgresType = UuidType{}

// BaseID implements the DoltgresType interface.
func (b UuidType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID(SerializationID_Uuid)
}

// CollationCoercibility implements the DoltgresType interface.
func (b UuidType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b UuidType) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(uuid.UUID)
	bb := bc.(uuid.UUID)
	return bytes.Compare(ab.GetBytesMut(), bb.GetBytesMut()), nil
}

// Convert implements the DoltgresType interface.
func (b UuidType) Convert(val any) (any, sql.ConvertInRange, error) {
	if val == nil {
		return nil, sql.InRange, nil
	}

	switch val := val.(type) {
	case string:
		uuidVal, err := uuid.FromString(val)
		if err != nil {
			return nil, sql.OutOfRange, err
		}
		return uuidVal, sql.InRange, nil
	case uuid.UUID:
		return val, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, sql.ErrInvalidType.New(b)
	}
}

// Equals implements the DoltgresType interface.
func (b UuidType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b UuidType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b UuidType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return "", err
	}
	return converted.(uuid.UUID).String(), nil
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b UuidType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b UuidType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 16
}

// OID implements the DoltgresType interface.
func (b UuidType) OID() uint32 {
	return uint32(oid.T_uuid)
}

// Promote implements the DoltgresType interface.
func (b UuidType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b UuidType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b UuidType) SerializeType() ([]byte, error) {
	return SerializationID_Uuid.ToByteSlice(), nil
}

// SQL implements the DoltgresType interface.
func (b UuidType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, _, err := b.Convert(v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value.(uuid.UUID).String()))), nil
}

// String implements the DoltgresType interface.
func (b UuidType) String() string {
	return "uuid"
}

// ToArrayType implements the DoltgresType interface.
func (b UuidType) ToArrayType() DoltgresArrayType {
	return UuidArray
}

// Type implements the DoltgresType interface.
func (b UuidType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (b UuidType) ValueType() reflect.Type {
	return reflect.TypeOf(uuid.UUID{})
}

// Zero implements the DoltgresType interface.
func (b UuidType) Zero() any {
	return uuid.UUID{}
}

// SerializeValue implements the DoltgresType interface.
func (b UuidType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	return converted.(uuid.UUID).GetBytes(), nil
}

// DeserializeValue implements the DoltgresType interface.
func (b UuidType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return uuid.FromBytes(val)
}
