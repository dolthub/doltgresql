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
	"reflect"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
)

var _ DoltgresType = CompositeType{}

type CompositeType struct {
	elements []DoltgresType
}

func (c CompositeType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return 0, 0
}

func (c CompositeType) Compare(i interface{}, i2 interface{}) (int, error) {
	return 0, nil
}

func (c CompositeType) Convert(i interface{}) (interface{}, sql.ConvertInRange, error) {
	return nil, sql.OutOfRange, nil
}

func (c CompositeType) Equals(otherType sql.Type) bool {
	return false
}

func (c CompositeType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 0
}

func (c CompositeType) Promote() sql.Type {
	return c
}

func (c CompositeType) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	// TODO
	return sqltypes.NULL, nil
}

func (c CompositeType) Type() query.Type {
	return query.Type_TUPLE
}

func (c CompositeType) ValueType() reflect.Type {
	// TODO implement me
	return nil
}

func (c CompositeType) Zero() interface{} {
	return nil
}

func (c CompositeType) String() string {
	return ""
}

func (c CompositeType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return 0, nil
}

func (c CompositeType) SerializeValue(val any) ([]byte, error) {
	return nil, nil
}

func (c CompositeType) DeserializeValue(val []byte) (any, error) {
	panic("implement me")
}

func (c CompositeType) FormatValue(val any) (string, error) {
	panic("implement me")
}

func (c CompositeType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

func (c CompositeType) Alignment() TypeAlignment {
	return TypeAlignment_Char
}

func (c CompositeType) BaseID() DoltgresTypeBaseID {
	// TODO
	return DoltgresTypeBaseID_Composite
}

func (c CompositeType) BaseName() string {
	return "composite"
}

func (c CompositeType) Category() TypeCategory {
	return TypeCategory_CompositeTypes
}

func (c CompositeType) GetSerializationID() SerializationID {
	return SerializationID_Composite
}

func (c CompositeType) IoInput(ctx *sql.Context, input string) (any, error) {
	panic("implement me")
}

func (c CompositeType) IoOutput(ctx *sql.Context, output any) (string, error) {
	panic("implement me")
}

func (c CompositeType) IsPreferredType() bool {
	return false
}

func (c CompositeType) IsUnbounded() bool {
	return true
}

func (c CompositeType) OID() uint32 {
	return 0
}

func (c CompositeType) SerializeType() ([]byte, error) {
	panic("implement me")
}

func (c CompositeType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	panic("implement me")
}

func (c CompositeType) ToArrayType() DoltgresArrayType {
	panic("implement me")
}
