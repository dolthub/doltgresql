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

// AnyNonArray is a pseudo-type that can represent any type that isn't an array type.
var AnyNonArray = AnyNonArrayType{}

// AnyNonArrayType is the extended type implementation of the PostgreSQL anynonarray.
type AnyNonArrayType struct{}

var _ DoltgresType = AnyNonArrayType{}
var _ DoltgresPolymorphicType = AnyNonArrayType{}

// Alignment implements the DoltgresType interface.
func (ana AnyNonArrayType) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresType interface.
func (ana AnyNonArrayType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_AnyNonArray
}

// BaseName implements the DoltgresType interface.
func (ana AnyNonArrayType) BaseName() string {
	return "anynonarray"
}

// Category implements the DoltgresType interface.
func (ana AnyNonArrayType) Category() TypeCategory {
	return TypeCategory_PseudoTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (ana AnyNonArrayType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (ana AnyNonArrayType) Compare(v1 any, v2 any) (int, error) {
	return 0, fmt.Errorf("%s cannot compare values", ana.String())
}

// Convert implements the DoltgresType interface.
func (ana AnyNonArrayType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case []any:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", ana.String(), val)
	default:
		return val, sql.InRange, nil
	}
}

// Equals implements the DoltgresType interface.
func (ana AnyNonArrayType) Equals(otherType sql.Type) bool {
	_, ok := otherType.(AnyNonArrayType)
	return ok
}

// FormatValue implements the DoltgresType interface.
func (ana AnyNonArrayType) FormatValue(val any) (string, error) {
	return "", fmt.Errorf("%s cannot format values", ana.String())
}

// GetSerializationID implements the DoltgresType interface.
func (ana AnyNonArrayType) GetSerializationID() SerializationID {
	return SerializationID_Invalid
}

// IoInput implements the DoltgresType interface.
func (ana AnyNonArrayType) IoInput(ctx *sql.Context, input string) (any, error) {
	return "", fmt.Errorf("%s cannot receive I/O input", ana.String())
}

// IoOutput implements the DoltgresType interface.
func (ana AnyNonArrayType) IoOutput(ctx *sql.Context, output any) (string, error) {
	return "", fmt.Errorf("%s cannot produce I/O output", ana.String())
}

// IsPreferredType implements the DoltgresType interface.
func (ana AnyNonArrayType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (ana AnyNonArrayType) IsUnbounded() bool {
	return true
}

// IsValid implements the DoltgresPolymorphicType interface.
func (ana AnyNonArrayType) IsValid(target DoltgresType) bool {
	_, ok := target.(DoltgresArrayType)
	return !ok
}

// MaxSerializedWidth implements the DoltgresType interface.
func (ana AnyNonArrayType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (ana AnyNonArrayType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresType interface.
func (ana AnyNonArrayType) OID() uint32 {
	return uint32(oid.T_anynonarray)
}

// Promote implements the DoltgresType interface.
func (ana AnyNonArrayType) Promote() sql.Type {
	return ana
}

// SerializedCompare implements the DoltgresType interface.
func (ana AnyNonArrayType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return 0, fmt.Errorf("%s cannot compare serialized values", ana.String())
}

// SQL implements the DoltgresType interface.
func (ana AnyNonArrayType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	return sqltypes.Value{}, fmt.Errorf("%s cannot output values in the wire format", ana.String())
}

// String implements the DoltgresType interface.
func (ana AnyNonArrayType) String() string {
	return "anynonarray"
}

// ToArrayType implements the DoltgresType interface.
func (ana AnyNonArrayType) ToArrayType() DoltgresArrayType {
	return Unknown
}

// Type implements the DoltgresType interface.
func (ana AnyNonArrayType) Type() query.Type {
	return sqltypes.Text
}

// ValToByteArray implements the DoltgresType interface.
func (ana AnyNonArrayType) ValToByteArray(val any) ([]byte, error) {
	return nil, fmt.Errorf("%s cannot output values in the wire format", ana.String())
}

// ValueType implements the DoltgresType interface.
func (ana AnyNonArrayType) ValueType() reflect.Type {
	var val any
	return reflect.TypeOf(val)
}

// Zero implements the DoltgresType interface.
func (ana AnyNonArrayType) Zero() any {
	var val any
	return val
}

// SerializeType implements the DoltgresType interface.
func (ana AnyNonArrayType) SerializeType() ([]byte, error) {
	return nil, fmt.Errorf("%s cannot be serialized", ana.String())
}

// deserializeType implements the DoltgresType interface.
func (ana AnyNonArrayType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	return nil, fmt.Errorf("%s cannot be deserialized", ana.String())
}

// SerializeValue implements the DoltgresType interface.
func (ana AnyNonArrayType) SerializeValue(val any) ([]byte, error) {
	return nil, fmt.Errorf("%s cannot serialize values", ana.String())
}

// DeserializeValue implements the DoltgresType interface.
func (ana AnyNonArrayType) DeserializeValue(val []byte) (any, error) {
	return nil, fmt.Errorf("%s cannot deserialize values", ana.String())
}
