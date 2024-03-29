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
)

// Null is the null type
var Null = NullType{}

// NullType is the extended type implementation of the PostgreSQL null.
type NullType struct{}

var _ DoltgresType = NullType{}

// BaseID implements the DoltgresType interface.
func (b NullType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID(SerializationID_Null)
}

// CollationCoercibility implements the DoltgresType interface.
func (b NullType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b NullType) Compare(v1 any, v2 any) (int, error) {
	return 0, nil
}

// Convert implements the DoltgresType interface.
func (b NullType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val.(type) {
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, sql.ErrInvalidType.New(b)
	}
}

// Equals implements the DoltgresType interface.
func (b NullType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b NullType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b NullType) FormatValue(val any) (string, error) {
	return "NULL", nil
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b NullType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b NullType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 1
}

// OID implements the DoltgresType interface.
func (b NullType) OID() uint32 {
	return 0
}

// Promote implements the DoltgresType interface.
func (b NullType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b NullType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return 0, nil
}

// SQL implements the DoltgresType interface.
func (b NullType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	return sqltypes.NULL, nil
}

// String implements the DoltgresType interface.
func (b NullType) String() string {
	return "null"
}

// Type implements the DoltgresType interface.
func (b NullType) Type() query.Type {
	return sqltypes.Null
}

// ValueType implements the DoltgresType interface.
func (b NullType) ValueType() reflect.Type {
	return reflect.TypeOf(nil)
}

// Zero implements the DoltgresType interface.
func (b NullType) Zero() any {
	return nil
}

// SerializeValue implements the DoltgresType interface.
func (b NullType) SerializeValue(val any) ([]byte, error) {
	return nil, nil
}

// DeserializeValue implements the DoltgresType interface.
func (b NullType) DeserializeValue(val []byte) (any, error) {
	return nil, nil
}
