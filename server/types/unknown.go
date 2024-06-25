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
	"fmt"
	"math"
	"reflect"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// Unknown represents an invalid or indeterminate type. This is primarily used internally.
var Unknown = UnknownType{}

// UnknownType is the extended type implementation of the PostgreSQL unknown type.
type UnknownType struct{}

var _ DoltgresType = UnknownType{}
var _ DoltgresArrayType = UnknownType{}

// Alignment implements the DoltgresType interface.
func (u UnknownType) Alignment() TypeAlignment {
	return TypeAlignment_Char
}

// BaseID implements the DoltgresType interface.
func (u UnknownType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Unknown
}

// BaseName implements the DoltgresType interface.
func (u UnknownType) BaseName() string {
	return "unknown"
}

// Category implements the DoltgresType interface.
func (u UnknownType) Category() TypeCategory {
	return TypeCategory_UnknownTypes
}

// BaseType implements the DoltgresArrayType interface.
func (u UnknownType) BaseType() DoltgresType {
	return Unknown
}

// CollationCoercibility implements the DoltgresType interface.
func (u UnknownType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (u UnknownType) Compare(v1 any, v2 any) (int, error) {
	return 0, fmt.Errorf("%s cannot compare values", u.String())
}

// Convert implements the DoltgresType interface.
func (u UnknownType) Convert(val any) (any, sql.ConvertInRange, error) {
	return nil, sql.OutOfRange, fmt.Errorf("%s cannot convert values", u.String())
}

// Equals implements the DoltgresType interface.
func (u UnknownType) Equals(otherType sql.Type) bool {
	_, ok := otherType.(UnknownType)
	return ok
}

// FormatSerializedValue implements the DoltgresType interface.
func (u UnknownType) FormatSerializedValue(val []byte) (string, error) {
	return "", fmt.Errorf("%s cannot format serialized values", u.String())
}

// FormatValue implements the DoltgresType interface.
func (u UnknownType) FormatValue(val any) (string, error) {
	return "", fmt.Errorf("%s cannot format values", u.String())
}

// GetSerializationID implements the DoltgresType interface.
func (u UnknownType) GetSerializationID() SerializationID {
	return SerializationID_Invalid
}

// IoInput implements the DoltgresType interface.
func (u UnknownType) IoInput(input string) (any, error) {
	return "", fmt.Errorf("%s cannot receive I/O input", u.String())
}

// IoOutput implements the DoltgresType interface.
func (u UnknownType) IoOutput(output any) (string, error) {
	return "", fmt.Errorf("%s cannot produce I/O output", u.String())
}

// IsPreferredType implements the DoltgresType interface.
func (b UnknownType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (u UnknownType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (u UnknownType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (u UnknownType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresType interface.
func (u UnknownType) OID() uint32 {
	return uint32(oid.T_unknown)
}

// Promote implements the DoltgresType interface.
func (u UnknownType) Promote() sql.Type {
	return u
}

// SerializedCompare implements the DoltgresType interface.
func (u UnknownType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return 0, fmt.Errorf("%s cannot compare serialized values", u.String())
}

// SQL implements the DoltgresType interface.
func (u UnknownType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	return sqltypes.Value{}, fmt.Errorf("%s cannot output values in the wire format", u.String())
}

// String implements the DoltgresType interface.
func (u UnknownType) String() string {
	return "unknown"
}

// ToArrayType implements the DoltgresType interface.
func (u UnknownType) ToArrayType() DoltgresArrayType {
	return u
}

// Type implements the DoltgresType interface.
func (u UnknownType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (u UnknownType) ValueType() reflect.Type {
	return reflect.TypeOf(any(nil))
}

// Zero implements the DoltgresType interface.
func (u UnknownType) Zero() any {
	return any(nil)
}

// SerializeType implements the DoltgresType interface.
func (u UnknownType) SerializeType() ([]byte, error) {
	return nil, fmt.Errorf("%s cannot be serialized", u.String())
}

// deserializeType implements the DoltgresType interface.
func (u UnknownType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	return nil, fmt.Errorf("%s cannot be deserialized", u.String())
}

// SerializeValue implements the DoltgresType interface.
func (u UnknownType) SerializeValue(val any) ([]byte, error) {
	return nil, fmt.Errorf("%s cannot serialize values", u.String())
}

// DeserializeValue implements the DoltgresType interface.
func (u UnknownType) DeserializeValue(val []byte) (any, error) {
	return nil, fmt.Errorf("%s cannot deserialize values", u.String())
}
