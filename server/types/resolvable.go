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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// ResolvableType represents any non-built-in type
// that needs resolution at analyzer stage.
// It is used for domain types, and it can be used
// for other user-defined types we don't support yet.
type ResolvableType struct {
	Typ tree.ResolvableTypeReference
}

var _ DoltgresType = ResolvableType{}

// Alignment implements the DoltgresType interface.
func (b ResolvableType) Alignment() TypeAlignment {
	panic("ResolvableType is a placeholder type, but Alignment() was called")
}

// BaseID implements the DoltgresType interface.
func (b ResolvableType) BaseID() DoltgresTypeBaseID {
	panic("ResolvableType is a placeholder type, but BaseID() was called")
}

// BaseName implements the DoltgresType interface.
func (b ResolvableType) BaseName() string {
	return fmt.Sprintf("ResolvableType(%s)", b.Typ.SQLString())
}

// Category implements the DoltgresType interface.
func (b ResolvableType) Category() TypeCategory {
	panic("ResolvableType is a placeholder type, but Category() was called")
}

// CollationCoercibility implements the DoltgresType interface.
func (b ResolvableType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	panic("ResolvableType is a placeholder type, but CollationCoercibility() was called")
}

// Compare implements the DoltgresType interface.
func (b ResolvableType) Compare(v1 any, v2 any) (int, error) {
	panic("ResolvableType is a placeholder type, but Compare() was called")
}

// Convert implements the DoltgresType interface.
func (b ResolvableType) Convert(val any) (any, sql.ConvertInRange, error) {
	panic("ResolvableType is a placeholder type, but Convert() was called")
}

// Equals implements the DoltgresType interface.
func (b ResolvableType) Equals(otherType sql.Type) bool {
	panic("ResolvableType is a placeholder type, but Equals() was called")
}

// FormatValue implements the DoltgresType interface.
func (b ResolvableType) FormatValue(val any) (string, error) {
	panic("ResolvableType is a placeholder type, but FormatValue() was called")
}

// GetSerializationID implements the DoltgresType interface.
func (b ResolvableType) GetSerializationID() SerializationID {
	panic("ResolvableType is a placeholder type, but GetSerializationID() was called")
}

// IoInput implements the DoltgresType interface.
func (b ResolvableType) IoInput(ctx *sql.Context, input string) (any, error) {
	panic("ResolvableType is a placeholder type, but IoInput() was called")
}

// IoOutput implements the DoltgresType interface.
func (b ResolvableType) IoOutput(ctx *sql.Context, output any) (string, error) {
	panic("ResolvableType is a placeholder type, but IoOutput() was called")
}

// IsPreferredType implements the DoltgresType interface.
func (b ResolvableType) IsPreferredType() bool {
	panic("ResolvableType is a placeholder type, but IsPreferredType() was called")
}

// IsUnbounded implements the DoltgresType interface.
func (b ResolvableType) IsUnbounded() bool {
	panic("ResolvableType is a placeholder type, but IsUnbounded() was called")
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b ResolvableType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	panic("ResolvableType is a placeholder type, but MaxSerializedWidth() was called")
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b ResolvableType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	panic("ResolvableType is a placeholder type, but MaxTextResponseByteLength() was called")
}

// OID implements the DoltgresType interface.
func (b ResolvableType) OID() uint32 {
	panic("ResolvableType is a placeholder type, but OID() was called")
}

// Promote implements the DoltgresType interface.
func (b ResolvableType) Promote() sql.Type {
	panic("ResolvableType is a placeholder type, but Promote() was called")
}

// SerializedCompare implements the DoltgresType interface.
func (b ResolvableType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	panic("ResolvableType is a placeholder type, but SerializedCompare() was called")
}

// SQL implements the DoltgresType interface.
func (b ResolvableType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	panic("ResolvableType is a placeholder type, but SQL() was called")
}

// String implements the DoltgresType interface.
func (b ResolvableType) String() string {
	return fmt.Sprintf("ResolvableType(%s)", b.Typ.SQLString())
}

// ToArrayType implements the DoltgresType interface.
func (b ResolvableType) ToArrayType() DoltgresArrayType {
	panic("ResolvableType is a placeholder type, but ToArrayType() was called")
}

// Type implements the DoltgresType interface.
func (b ResolvableType) Type() query.Type {
	panic("ResolvableType is a placeholder type, but Type() was called")
}

// ValueType implements the DoltgresType interface.
func (b ResolvableType) ValueType() reflect.Type {
	panic("ResolvableType is a placeholder type, but ValueType() was called")
}

// Zero implements the DoltgresType interface.
func (b ResolvableType) Zero() any {
	panic("ResolvableType is a placeholder type, but Zero() was called")
}

// SerializeType implements the DoltgresType interface.
func (b ResolvableType) SerializeType() ([]byte, error) {
	panic("ResolvableType is a placeholder type, but SerializeType() was called")
}

// deserializeType implements the DoltgresType interface.
func (b ResolvableType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	panic("ResolvableType is a placeholder type, but deserializeType() was called")
}

// SerializeValue implements the DoltgresType interface.
func (b ResolvableType) SerializeValue(val any) ([]byte, error) {
	panic("ResolvableType is a placeholder type, but SerializeValue() was called")
}

// DeserializeValue implements the DoltgresType interface.
func (b ResolvableType) DeserializeValue(val []byte) (any, error) {
	panic("ResolvableType is a placeholder type, but DeserializeValue() was called")
}
