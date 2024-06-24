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
var _ DoltgresValidType = UuidType{}

// Alignment implements the DoltgresType interface.
func (b UuidType) Alignment() TypeAlignment {
	return TypeAlignment_Char
}

// BaseID implements the DoltgresType interface.
func (b UuidType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Uuid
}

// BaseName implements the DoltgresType interface.
func (b UuidType) BaseName() string {
	return "uuid"
}

// Category implements the DoltgresType interface.
func (b UuidType) Category() TypeCategory {
	return TypeCategory_UserDefinedTypes
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
	switch val := val.(type) {
	case uuid.UUID:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
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
	return b.IoOutput(val)
}

// GetSerializationID implements the DoltgresType interface.
func (b UuidType) GetSerializationID() SerializationID {
	return SerializationID_Uuid
}

// IoInput implements the DoltgresType interface.
func (b UuidType) IoInput(input string) (any, error) {
	return uuid.FromString(input)
}

// IoOutput implements the DoltgresType interface.
func (b UuidType) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return converted.(uuid.UUID).String(), nil
}

func (b UuidType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b UuidType) IsUnbounded() bool {
	return false
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

// SerializeType implements the DoltgresType interface.
func (b UuidType) SerializeType() ([]byte, error) {
	return SerializationID_Uuid.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b UuidType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Uuid, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
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
