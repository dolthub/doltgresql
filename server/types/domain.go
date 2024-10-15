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
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/utils"
)

var ErrTypeAlreadyExists = errors.NewKind(`type "%s" already exists`)
var ErrTypeDoesNotExist = errors.NewKind(`type "%s" does not exist`)

// TODO: use maybe separate unresolved domain type?

type DomainType struct {
	SchemaName string
	Name       string
	DataType   DoltgresType
}

var _ DoltgresType = DomainType{}

func (d DomainType) Alignment() TypeAlignment {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.Alignment()
}

func (d DomainType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseId_Domain
}

func (d DomainType) BaseName() string {
	// TODO
	return d.Name
}

func (d DomainType) Category() TypeCategory {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.Category()
}

func (d DomainType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.CollationCoercibility(ctx)
}

func (d DomainType) Compare(i interface{}, i2 interface{}) (int, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.Compare(i, i2)
}

func (d DomainType) Convert(i interface{}) (interface{}, sql.ConvertInRange, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.Convert(i)
}

func (d DomainType) Equals(otherType sql.Type) bool {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.Equals(otherType)
}

func (d DomainType) FormatValue(val any) (string, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.FormatValue(val)
}

func (d DomainType) GetSerializationID() SerializationID {
	return SerializationId_Domain
}

func (d DomainType) IoInput(ctx *sql.Context, input string) (any, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.IoInput(ctx, input)
}

func (d DomainType) IoOutput(ctx *sql.Context, output any) (string, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.IoOutput(ctx, output)
}

func (d DomainType) IsPreferredType() bool {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.IsPreferredType()
}

func (d DomainType) IsUnbounded() bool {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.IsUnbounded()
}

func (d DomainType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.MaxSerializedWidth()
}

func (d DomainType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.MaxTextResponseByteLength(ctx)
}

func (d DomainType) OID() uint32 {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	//TODO : different oid than the AsType one
	return d.DataType.OID()
}

func (d DomainType) Promote() sql.Type {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.Promote()
}

func (d DomainType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.SerializedCompare(v1, v2)
}

func (d DomainType) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.SQL(ctx, dest, v)
}

func (d DomainType) String() string {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.Name
}

func (d DomainType) ToArrayType() DoltgresArrayType {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	// TODO: allowed?
	return d.DataType.ToArrayType()
}

func (d DomainType) Type() query.Type {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.Type()
}

func (d DomainType) ValueType() reflect.Type {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.ValueType()
}

func (d DomainType) Zero() interface{} {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.Zero()
}

func (d DomainType) SerializeType() ([]byte, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	b := SerializationId_Domain.ToByteSlice(0)
	writer := utils.NewWriter(256)
	writer.String(d.SchemaName)
	writer.String(d.Name)
	asTyp, err := d.DataType.SerializeType()
	if err != nil {
		return nil, err
	}
	b = append(b, writer.Data()...)
	return append(b, asTyp...), nil
}

func (d DomainType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		reader := utils.NewReader(metadata)
		d.SchemaName = reader.String()
		d.Name = reader.String()
		t, err := DeserializeType(reader.RemainingBytes())
		if err != nil {
			return nil, err
		}
		d.DataType = t.(DoltgresType)
		return d, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, d.String())
	}
}

func (d DomainType) SerializeValue(val any) ([]byte, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.SerializeValue(val)
}

func (d DomainType) DeserializeValue(val []byte) (any, error) {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	return d.DataType.DeserializeValue(val)
}

func (d DomainType) GetBaseType() DoltgresType {
	if d.DataType == nil {
		panic("unresolved domain")
	}
	switch t := d.DataType.(type) {
	case DomainType:
		return t.GetBaseType()
	default:
		// TODO: how to make sure this is an built-in type?
		return t
	}
}
