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

// AnyElement is a pseudo-type that can represent any type.
var AnyElement = AnyElementType{}

// AnyElementType is the extended type implementation of the PostgreSQL anyelement.
type AnyElementType struct{}

var _ DoltgresType = AnyElementType{}
var _ DoltgresPolymorphicType = AnyElementType{}

// Alignment implements the DoltgresType interface.
func (ae AnyElementType) Alignment() TypeAlignment {
	return TypeAlignment_Double
}

// BaseID implements the DoltgresType interface.
func (ae AnyElementType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_AnyElement
}

// BaseName implements the DoltgresType interface.
func (ae AnyElementType) BaseName() string {
	return "anyelement"
}

// Category implements the DoltgresType interface.
func (ae AnyElementType) Category() TypeCategory {
	return TypeCategory_PseudoTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (ae AnyElementType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (ae AnyElementType) Compare(v1 any, v2 any) (int, error) {
	return 0, fmt.Errorf("%s cannot compare values", ae.String())
}

// Convert implements the DoltgresType interface.
func (ae AnyElementType) Convert(val any) (any, sql.ConvertInRange, error) {
	return val, sql.InRange, nil
}

// Equals implements the DoltgresType interface.
func (ae AnyElementType) Equals(otherType sql.Type) bool {
	_, ok := otherType.(AnyElementType)
	return ok
}

// FormatSerializedValue implements the DoltgresType interface.
func (ae AnyElementType) FormatSerializedValue(val []byte) (string, error) {
	return "", fmt.Errorf("%s cannot format serialized values", ae.String())
}

// FormatValue implements the DoltgresType interface.
func (ae AnyElementType) FormatValue(val any) (string, error) {
	return "", fmt.Errorf("%s cannot format values", ae.String())
}

// GetSerializationID implements the DoltgresType interface.
func (ae AnyElementType) GetSerializationID() SerializationID {
	return SerializationID_Invalid
}

// IoInput implements the DoltgresType interface.
func (ae AnyElementType) IoInput(input string) (any, error) {
	return "", fmt.Errorf("%s cannot receive I/O input", ae.String())
}

// IoOutput implements the DoltgresType interface.
func (ae AnyElementType) IoOutput(output any) (string, error) {
	return "", fmt.Errorf("%s cannot produce I/O output", ae.String())
}

// IsPreferredType implements the DoltgresType interface.
func (ae AnyElementType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (ae AnyElementType) IsUnbounded() bool {
	return true
}

// IsValid implements the DoltgresPolymorphicType interface.
func (ae AnyElementType) IsValid(target DoltgresType) bool {
	return true
}

// MaxSerializedWidth implements the DoltgresType interface.
func (ae AnyElementType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (ae AnyElementType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresType interface.
func (ae AnyElementType) OID() uint32 {
	return uint32(oid.T_anyelement)
}

// Promote implements the DoltgresType interface.
func (ae AnyElementType) Promote() sql.Type {
	return ae
}

// SerializedCompare implements the DoltgresType interface.
func (ae AnyElementType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return 0, fmt.Errorf("%s cannot compare serialized values", ae.String())
}

// SQL implements the DoltgresType interface.
func (ae AnyElementType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	return sqltypes.Value{}, fmt.Errorf("%s cannot output values in the wire format", ae.String())
}

// String implements the DoltgresType interface.
func (ae AnyElementType) String() string {
	return "anyelement"
}

// ToArrayType implements the DoltgresType interface.
func (ae AnyElementType) ToArrayType() DoltgresArrayType {
	return Unknown
}

// Type implements the DoltgresType interface.
func (ae AnyElementType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (ae AnyElementType) ValueType() reflect.Type {
	var val any
	return reflect.TypeOf(val)
}

// Zero implements the DoltgresType interface.
func (ae AnyElementType) Zero() any {
	var val any
	return val
}

// SerializeType implements the DoltgresType interface.
func (ae AnyElementType) SerializeType() ([]byte, error) {
	return nil, fmt.Errorf("%s cannot be serialized", ae.String())
}

// deserializeType implements the DoltgresType interface.
func (ae AnyElementType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	return nil, fmt.Errorf("%s cannot be deserialized", ae.String())
}

// SerializeValue implements the DoltgresType interface.
func (ae AnyElementType) SerializeValue(val any) ([]byte, error) {
	return nil, fmt.Errorf("%s cannot serialize values", ae.String())
}

// DeserializeValue implements the DoltgresType interface.
func (ae AnyElementType) DeserializeValue(val []byte) (any, error) {
	return nil, fmt.Errorf("%s cannot deserialize values", ae.String())
}
