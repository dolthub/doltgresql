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

	"github.com/jackc/pgio"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// Regproc is the OID type for finding function names.
var Regproc = RegprocType{}

// RegprocType is the extended type implementation of the PostgreSQL regproc.
type RegprocType struct{}

var _ DoltgresType = RegprocType{}

// Alignment implements the DoltgresType interface.
func (b RegprocType) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresType interface.
func (b RegprocType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Regproc
}

// BaseName implements the DoltgresType interface.
func (b RegprocType) BaseName() string {
	return "regproc"
}

// Category implements the DoltgresType interface.
func (b RegprocType) Category() TypeCategory {
	return TypeCategory_NumericTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b RegprocType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b RegprocType) Compare(v1 any, v2 any) (int, error) {
	return OidType{}.Compare(v1, v2)
}

// Convert implements the DoltgresType interface.
func (b RegprocType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case uint32:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b RegprocType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresType interface.
func (b RegprocType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresType interface.
func (b RegprocType) GetSerializationID() SerializationID {
	return SerializationID_Invalid
}

// Regproc_IoInput is the implementation for IoInput that is being set from another package to avoid circular dependencies.
var Regproc_IoInput func(ctx *sql.Context, input string) (uint32, error)

// IoInput implements the DoltgresType interface.
func (b RegprocType) IoInput(ctx *sql.Context, input string) (any, error) {
	return Regproc_IoInput(ctx, input)
}

// Regproc_IoOutput is the implementation for IoOutput that is being set from another package to avoid circular dependencies.
var Regproc_IoOutput func(ctx *sql.Context, oid uint32) (string, error)

// IoOutput implements the DoltgresType interface.
func (b RegprocType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return Regproc_IoOutput(ctx, converted.(uint32))
}

// IsPreferredType implements the DoltgresType interface.
func (b RegprocType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b RegprocType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b RegprocType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b RegprocType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 4
}

// OID implements the DoltgresType interface.
func (b RegprocType) OID() uint32 {
	return uint32(oid.T_regproc)
}

// Promote implements the DoltgresType interface.
func (b RegprocType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b RegprocType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	return bytes.Compare(v1, v2), nil
}

// SQL implements the DoltgresType interface.
func (b RegprocType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.IoOutput(ctx, v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresType interface.
func (b RegprocType) String() string {
	return "regproc"
}

// ToArrayType implements the DoltgresType interface.
func (b RegprocType) ToArrayType() DoltgresArrayType {
	return RegprocArray
}

// Type implements the DoltgresType interface.
func (b RegprocType) Type() query.Type {
	return sqltypes.Text
}

// ValToByteArray implements the DoltgresType interface.
func (b RegprocType) ValToByteArray(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	return pgio.AppendUint32(nil, converted.(uint32)), nil
}

// ValueType implements the DoltgresType interface.
func (b RegprocType) ValueType() reflect.Type {
	return reflect.TypeOf(uint32(0))
}

// Zero implements the DoltgresType interface.
func (b RegprocType) Zero() any {
	return uint32(0)
}

// SerializeType implements the DoltgresType interface.
func (b RegprocType) SerializeType() ([]byte, error) {
	return nil, fmt.Errorf("%s cannot be serialized", b.String())
}

// deserializeType implements the DoltgresType interface.
func (b RegprocType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	return nil, fmt.Errorf("%s cannot be deserialized", b.String())
}

// SerializeValue implements the DoltgresType interface.
func (b RegprocType) SerializeValue(val any) ([]byte, error) {
	return nil, fmt.Errorf("%s cannot serialize values", b.String())
}

// DeserializeValue implements the DoltgresType interface.
func (b RegprocType) DeserializeValue(val []byte) (any, error) {
	return nil, fmt.Errorf("%s cannot deserialize values", b.String())
}
