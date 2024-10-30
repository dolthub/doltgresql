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
	Typ          tree.ResolvableTypeReference
	ResolvedType DoltgresType
	IsArray      bool
}

var _ types.ExtendedType = ResolvableType{}

// CollationCoercibility implements the types.ExtendedType interface.
func (b ResolvableType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	panic("ResolvableType is a placeholder type, but CollationCoercibility() was called")
}

// Compare implements the types.ExtendedType interface.
func (b ResolvableType) Compare(v1 any, v2 any) (int, error) {
	panic("ResolvableType is a placeholder type, but Compare() was called")
}

// Convert implements the types.ExtendedType interface.
func (b ResolvableType) Convert(val any) (any, sql.ConvertInRange, error) {
	panic("ResolvableType is a placeholder type, but Convert() was called")
}

// Equals implements the types.ExtendedType interface.
func (b ResolvableType) Equals(otherType sql.Type) bool {
	panic("ResolvableType is a placeholder type, but Equals() was called")
}

// FormatValue implements the types.ExtendedType interface.
func (b ResolvableType) FormatValue(val any) (string, error) {
	panic("ResolvableType is a placeholder type, but FormatValue() was called")
}

// MaxSerializedWidth implements the types.ExtendedType interface.
func (b ResolvableType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	panic("ResolvableType is a placeholder type, but MaxSerializedWidth() was called")
}

// MaxTextResponseByteLength implements the types.ExtendedType interface.
func (b ResolvableType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	panic("ResolvableType is a placeholder type, but MaxTextResponseByteLength() was called")
}

// Promote implements the types.ExtendedType interface.
func (b ResolvableType) Promote() sql.Type {
	panic("ResolvableType is a placeholder type, but Promote() was called")
}

// SerializedCompare implements the types.ExtendedType interface.
func (b ResolvableType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	panic("ResolvableType is a placeholder type, but SerializedCompare() was called")
}

// SQL implements the types.ExtendedType interface.
func (b ResolvableType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	panic("ResolvableType is a placeholder type, but SQL() was called")
}

// String implements the types.ExtendedType interface.
func (b ResolvableType) String() string {
	return fmt.Sprintf("ResolvableType(%s)", b.Typ.SQLString())
}

// Type implements the types.ExtendedType interface.
func (b ResolvableType) Type() query.Type {
	panic("ResolvableType is a placeholder type, but Type() was called")
}

// ValueType implements the types.ExtendedType interface.
func (b ResolvableType) ValueType() reflect.Type {
	panic("ResolvableType is a placeholder type, but ValueType() was called")
}

// Zero implements the types.ExtendedType interface.
func (b ResolvableType) Zero() any {
	panic("ResolvableType is a placeholder type, but Zero() was called")
}

// SerializeValue implements the types.ExtendedType interface.
func (b ResolvableType) SerializeValue(val any) ([]byte, error) {
	panic("ResolvableType is a placeholder type, but SerializeValue() was called")
}

// DeserializeValue implements the types.ExtendedType interface.
func (b ResolvableType) DeserializeValue(val []byte) (any, error) {
	panic("ResolvableType is a placeholder type, but DeserializeValue() was called")
}
