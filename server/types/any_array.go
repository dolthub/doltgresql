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

// AnyArray is an array that may contain elements of any type.
var AnyArray = AnyArrayType{}

// AnyArrayType is the extended type implementation of the PostgreSQL anyarray.
type AnyArrayType struct{}

var _ DoltgresType = AnyArrayType{}
var _ DoltgresArrayType = AnyArrayType{}

// BaseID implements the DoltgresType interface.
func (aa AnyArrayType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_AnyArray
}

// BaseType implements the DoltgresArrayType interface.
func (aa AnyArrayType) BaseType() DoltgresType {
	return Unknown
}

// CollationCoercibility implements the DoltgresType interface.
func (aa AnyArrayType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (aa AnyArrayType) Compare(v1 any, v2 any) (int, error) {
	return 0, fmt.Errorf("%s cannot compare values", aa.String())
}

// Convert implements the DoltgresType interface.
func (aa AnyArrayType) Convert(val any) (any, sql.ConvertInRange, error) {
	return nil, sql.OutOfRange, fmt.Errorf("%s cannot convert values", aa.String())
}

// Equals implements the DoltgresType interface.
func (aa AnyArrayType) Equals(otherType sql.Type) bool {
	_, ok := otherType.(AnyArrayType)
	return ok
}

// FormatSerializedValue implements the DoltgresType interface.
func (aa AnyArrayType) FormatSerializedValue(val []byte) (string, error) {
	return "", fmt.Errorf("%s cannot format serialized values", aa.String())
}

// FormatValue implements the DoltgresType interface.
func (aa AnyArrayType) FormatValue(val any) (string, error) {
	return "", fmt.Errorf("%s cannot format values", aa.String())
}

// GetSerializationID implements the DoltgresType interface.
func (aa AnyArrayType) GetSerializationID() SerializationID {
	return SerializationID_Invalid
}

// IoInput implements the DoltgresType interface.
func (aa AnyArrayType) IoInput(input string) (any, error) {
	return "", fmt.Errorf("%s cannot receive I/O input", aa.String())
}

// IoOutput implements the DoltgresType interface.
func (aa AnyArrayType) IoOutput(output any) (string, error) {
	return "", fmt.Errorf("%s cannot produce I/O output", aa.String())
}

// IsUnbounded implements the DoltgresType interface.
func (aa AnyArrayType) IsUnbounded() bool {
	return true
}

// MaxSerializedWidth implements the DoltgresType interface.
func (aa AnyArrayType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (aa AnyArrayType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresType interface.
func (aa AnyArrayType) OID() uint32 {
	return uint32(oid.T_anyarray)
}

// Promote implements the DoltgresType interface.
func (aa AnyArrayType) Promote() sql.Type {
	return aa
}

// SerializedCompare implements the DoltgresType interface.
func (aa AnyArrayType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return 0, fmt.Errorf("%s cannot compare serialized values", aa.String())
}

// SQL implements the DoltgresType interface.
func (aa AnyArrayType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	return sqltypes.Value{}, fmt.Errorf("%s cannot output values in the wire format", aa.String())
}

// String implements the DoltgresType interface.
func (aa AnyArrayType) String() string {
	return "anyarray"
}

// ToArrayType implements the DoltgresType interface.
func (aa AnyArrayType) ToArrayType() DoltgresArrayType {
	return aa
}

// Type implements the DoltgresType interface.
func (aa AnyArrayType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (aa AnyArrayType) ValueType() reflect.Type {
	return reflect.TypeOf([]any{})
}

// Zero implements the DoltgresType interface.
func (aa AnyArrayType) Zero() any {
	return []any{}
}

// SerializeType implements the DoltgresType interface.
func (aa AnyArrayType) SerializeType() ([]byte, error) {
	return nil, fmt.Errorf("%s cannot be serialized", aa.String())
}

// deserializeType implements the DoltgresType interface.
func (aa AnyArrayType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	return nil, fmt.Errorf("%s cannot be deserialized", aa.String())
}

// SerializeValue implements the DoltgresType interface.
func (aa AnyArrayType) SerializeValue(val any) ([]byte, error) {
	return nil, fmt.Errorf("%s cannot serialize values", aa.String())
}

// DeserializeValue implements the DoltgresType interface.
func (aa AnyArrayType) DeserializeValue(val []byte) (any, error) {
	return nil, fmt.Errorf("%s cannot deserialize values", aa.String())
}
