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
	"reflect"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// Int16Serial is an int16 serial type.
var Int16Serial = Int16TypeSerial{}

// Int16TypeSerial is the extended type implementation of the PostgreSQL smallserial.
type Int16TypeSerial struct{}

var _ DoltgresType = Int16TypeSerial{}

// Alignment implements the DoltgresType interface.
func (b Int16TypeSerial) Alignment() TypeAlignment {
	return TypeAlignment_Short
}

// BaseID implements the DoltgresType interface.
func (b Int16TypeSerial) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Int16Serial
}

// BaseName implements the DoltgresType interface.
func (b Int16TypeSerial) BaseName() string {
	return "smallserial"
}

// Category implements the DoltgresType interface.
func (b Int16TypeSerial) Category() TypeCategory {
	return TypeCategory_UnknownTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b Int16TypeSerial) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b Int16TypeSerial) Compare(v1 any, v2 any) (int, error) {
	return 0, fmt.Errorf("SERIAL types are not comparable")
}

// Convert implements the DoltgresType interface.
func (b Int16TypeSerial) Convert(val any) (any, sql.ConvertInRange, error) {
	return nil, sql.OutOfRange, fmt.Errorf("SERIAL types are not convertable")
}

// Equals implements the DoltgresType interface.
func (b Int16TypeSerial) Equals(otherType sql.Type) bool {
	_, ok := otherType.(Int16TypeSerial)
	return ok
}

// FormatSerializedValue implements the DoltgresType interface.
func (b Int16TypeSerial) FormatSerializedValue(val []byte) (string, error) {
	return "", fmt.Errorf("SERIAL types are not formattable")
}

// FormatValue implements the DoltgresType interface.
func (b Int16TypeSerial) FormatValue(val any) (string, error) {
	return "", fmt.Errorf("SERIAL types are not formattable")
}

// GetSerializationID implements the DoltgresType interface.
func (b Int16TypeSerial) GetSerializationID() SerializationID {
	return SerializationID_Invalid
}

// IoInput implements the DoltgresType interface.
func (b Int16TypeSerial) IoInput(input string) (any, error) {
	return "", fmt.Errorf("SERIAL types cannot receive I/O input")
}

// IoOutput implements the DoltgresType interface.
func (b Int16TypeSerial) IoOutput(output any) (string, error) {
	return "", fmt.Errorf("SERIAL types cannot produce I/O output")
}

// IsPreferredType implements the DoltgresType interface.
func (b Int16TypeSerial) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b Int16TypeSerial) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b Int16TypeSerial) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b Int16TypeSerial) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 2
}

// OID implements the DoltgresType interface.
func (b Int16TypeSerial) OID() uint32 {
	return uint32(oid.T_int2)
}

// Promote implements the DoltgresType interface.
func (b Int16TypeSerial) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b Int16TypeSerial) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return 0, fmt.Errorf("SERIAL types are not comparable")
}

// SQL implements the DoltgresType interface.
func (b Int16TypeSerial) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	return sqltypes.Value{}, fmt.Errorf("SERIAL types may not be passed over the wire")
}

// String implements the DoltgresType interface.
func (b Int16TypeSerial) String() string {
	return "smallserial"
}

// ToArrayType implements the DoltgresType interface.
func (b Int16TypeSerial) ToArrayType() DoltgresArrayType {
	return Unknown
}

// Type implements the DoltgresType interface.
func (b Int16TypeSerial) Type() query.Type {
	return sqltypes.Int16
}

// ValueType implements the DoltgresType interface.
func (b Int16TypeSerial) ValueType() reflect.Type {
	return reflect.TypeOf(int16(0))
}

// Zero implements the DoltgresType interface.
func (b Int16TypeSerial) Zero() any {
	return int16(0)
}

// SerializeType implements the DoltgresType interface.
func (b Int16TypeSerial) SerializeType() ([]byte, error) {
	return nil, fmt.Errorf("SERIAL types are not serializable")
}

// deserializeType implements the DoltgresType interface.
func (b Int16TypeSerial) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	return nil, fmt.Errorf("SERIAL types are not deserializable")
}

// SerializeValue implements the DoltgresType interface.
func (b Int16TypeSerial) SerializeValue(val any) ([]byte, error) {
	return nil, fmt.Errorf("SERIAL types are not serializable")
}

// DeserializeValue implements the DoltgresType interface.
func (b Int16TypeSerial) DeserializeValue(val []byte) (any, error) {
	return nil, fmt.Errorf("SERIAL types are not deserializable")
}
