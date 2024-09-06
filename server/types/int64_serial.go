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

// Int64Serial is an int16 serial type.
var Int64Serial = Int64TypeSerial{}

// Int64TypeSerial is the extended type implementation of the PostgreSQL bigserial.
type Int64TypeSerial struct{}

var _ DoltgresType = Int64TypeSerial{}

// Alignment implements the DoltgresType interface.
func (b Int64TypeSerial) Alignment() TypeAlignment {
	return TypeAlignment_Double
}

// BaseID implements the DoltgresType interface.
func (b Int64TypeSerial) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Int64Serial
}

// BaseName implements the DoltgresType interface.
func (b Int64TypeSerial) BaseName() string {
	return "bigserial"
}

// Category implements the DoltgresType interface.
func (b Int64TypeSerial) Category() TypeCategory {
	return TypeCategory_UnknownTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b Int64TypeSerial) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b Int64TypeSerial) Compare(v1 any, v2 any) (int, error) {
	return 0, fmt.Errorf("SERIAL types are not comparable")
}

// Convert implements the DoltgresType interface.
func (b Int64TypeSerial) Convert(val any) (any, sql.ConvertInRange, error) {
	return nil, sql.OutOfRange, fmt.Errorf("SERIAL types are not convertable")
}

// Equals implements the DoltgresType interface.
func (b Int64TypeSerial) Equals(otherType sql.Type) bool {
	_, ok := otherType.(Int64TypeSerial)
	return ok
}

// FormatValue implements the DoltgresType interface.
func (b Int64TypeSerial) FormatValue(val any) (string, error) {
	return "", fmt.Errorf("SERIAL types are not formattable")
}

// GetSerializationID implements the DoltgresType interface.
func (b Int64TypeSerial) GetSerializationID() SerializationID {
	return SerializationID_Invalid
}

// IoInput implements the DoltgresType interface.
func (b Int64TypeSerial) IoInput(ctx *sql.Context, input string) (any, error) {
	return "", fmt.Errorf("SERIAL types cannot receive I/O input")
}

// IoOutput implements the DoltgresType interface.
func (b Int64TypeSerial) IoOutput(ctx *sql.Context, output any) (string, error) {
	return "", fmt.Errorf("SERIAL types cannot produce I/O output")
}

// IsPreferredType implements the DoltgresType interface.
func (b Int64TypeSerial) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b Int64TypeSerial) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b Int64TypeSerial) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b Int64TypeSerial) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 8
}

// OID implements the DoltgresType interface.
func (b Int64TypeSerial) OID() uint32 {
	return uint32(oid.T_int8)
}

// Promote implements the DoltgresType interface.
func (b Int64TypeSerial) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b Int64TypeSerial) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return 0, fmt.Errorf("SERIAL types are not comparable")
}

// SQL implements the DoltgresType interface.
func (b Int64TypeSerial) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	return sqltypes.Value{}, fmt.Errorf("SERIAL types may not be passed over the wire")
}

// String implements the DoltgresType interface.
func (b Int64TypeSerial) String() string {
	return "bigserial"
}

// ToArrayType implements the DoltgresType interface.
func (b Int64TypeSerial) ToArrayType() DoltgresArrayType {
	return Unknown
}

// Type implements the DoltgresType interface.
func (b Int64TypeSerial) Type() query.Type {
	return sqltypes.Int64
}

// ValueType implements the DoltgresType interface.
func (b Int64TypeSerial) ValueType() reflect.Type {
	return reflect.TypeOf(int64(0))
}

// Zero implements the DoltgresType interface.
func (b Int64TypeSerial) Zero() any {
	return int64(0)
}

// SerializeType implements the DoltgresType interface.
func (b Int64TypeSerial) SerializeType() ([]byte, error) {
	return nil, fmt.Errorf("SERIAL types are not serializable")
}

// deserializeType implements the DoltgresType interface.
func (b Int64TypeSerial) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	return nil, fmt.Errorf("SERIAL types are not deserializable")
}

// SerializeValue implements the DoltgresType interface.
func (b Int64TypeSerial) SerializeValue(val any) ([]byte, error) {
	return nil, fmt.Errorf("SERIAL types are not serializable")
}

// DeserializeValue implements the DoltgresType interface.
func (b Int64TypeSerial) DeserializeValue(val []byte) (any, error) {
	return nil, fmt.Errorf("SERIAL types are not deserializable")
}
