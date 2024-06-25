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
)

// Null is the null type
var Null = NullType{}

// NullType is the extended type implementation of the PostgreSQL null.
type NullType struct{}

var _ DoltgresType = NullType{}

// Alignment implements the DoltgresType interface.
func (b NullType) Alignment() TypeAlignment {
	return TypeAlignment_Char
}

// BaseID implements the DoltgresType interface.
func (b NullType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Null
}

// BaseName implements the DoltgresType interface.
func (b NullType) BaseName() string {
	return "null"
}

// Category implements the DoltgresType interface.
func (b NullType) Category() TypeCategory {
	return TypeCategory_UnknownTypes
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
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
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

// GetSerializationID implements the DoltgresType interface.
func (b NullType) GetSerializationID() SerializationID {
	return SerializationID_Null
}

// IoInput implements the DoltgresType interface.
func (b NullType) IoInput(input string) (any, error) {
	return "", fmt.Errorf("%s cannot receive I/O input", b.String())
}

// IoOutput implements the DoltgresType interface.
func (b NullType) IoOutput(output any) (string, error) {
	return "", fmt.Errorf("%s cannot produce I/O output", b.String())
}

// IsPreferredType implements the DoltgresType interface.
func (b NullType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b NullType) IsUnbounded() bool {
	return false
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

// ToArrayType implements the DoltgresType interface.
func (b NullType) ToArrayType() DoltgresArrayType {
	return Unknown
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

// SerializeType implements the DoltgresType interface.
func (b NullType) SerializeType() ([]byte, error) {
	return SerializationID_Null.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b NullType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Null, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b NullType) SerializeValue(val any) ([]byte, error) {
	return nil, nil
}

// DeserializeValue implements the DoltgresType interface.
func (b NullType) DeserializeValue(val []byte) (any, error) {
	return nil, nil
}
