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
	"github.com/dolthub/doltgresql/utils"
	"reflect"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
)

type DomainType struct {
	SchemaName string
	Name       string
	DataType   DoltgresType
	//
	DefaultExpr sql.Expression
	NotNull     bool
	Checks      sql.CheckConstraints
}

func (d DomainType) ResolveType(def sql.Expression, notNull bool, checks sql.CheckConstraints) {
	d.DefaultExpr = def
	d.NotNull = notNull
	d.Checks = checks
}

var _ DoltgresType = DomainType{}

func (d DomainType) Alignment() TypeAlignment {
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
	return d.DataType.Category()
}

func (d DomainType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return d.DataType.CollationCoercibility(ctx)
}

func (d DomainType) Compare(i interface{}, i2 interface{}) (int, error) {
	return d.DataType.Compare(i, i2)
}

func (d DomainType) Convert(i interface{}) (interface{}, sql.ConvertInRange, error) {
	// TODO: get ctx?
	val, err := d.evaluateChecks(nil, i)
	if err != nil {
		return nil, false, err
	}
	return d.DataType.Convert(val)
}

func (d DomainType) Equals(otherType sql.Type) bool {
	return d.DataType.Equals(otherType)
}

func (d DomainType) FormatValue(val any) (string, error) {
	return d.DataType.FormatValue(val)
}

func (d DomainType) GetSerializationID() SerializationID {
	return SerializationId_Domain
}

func (d DomainType) IoInput(ctx *sql.Context, input string) (any, error) {
	return d.DataType.IoInput(ctx, input)
}

func (d DomainType) IoOutput(ctx *sql.Context, output any) (string, error) {
	return d.DataType.IoOutput(ctx, output)
}

func (d DomainType) IsPreferredType() bool {
	return d.DataType.IsPreferredType()
}

func (d DomainType) IsUnbounded() bool {
	return d.DataType.IsUnbounded()
}

func (d DomainType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return d.DataType.MaxSerializedWidth()
}

func (d DomainType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return d.DataType.MaxTextResponseByteLength(ctx)
}

func (d DomainType) OID() uint32 {
	//TODO : different oid than the AsType one
	return d.DataType.OID()
}

func (d DomainType) Promote() sql.Type {
	return d.DataType.Promote()
}

func (d DomainType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return d.DataType.SerializedCompare(v1, v2)
}

func (d DomainType) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	return d.DataType.SQL(ctx, dest, v)
}

func (d DomainType) String() string {
	return d.Name
}

func (d DomainType) ToArrayType() DoltgresArrayType {
	// TODO: allowed?
	return d.DataType.ToArrayType()
}

func (d DomainType) Type() query.Type {
	return d.DataType.Type()
}

func (d DomainType) ValueType() reflect.Type {
	return d.DataType.ValueType()
}

func (d DomainType) Zero() interface{} {
	return d.DataType.Zero()
}

func (d DomainType) SerializeType() ([]byte, error) {
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
	//return , nil
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
		// TODO: retrieve other information here?
		return d, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, d.String())
	}
}

func (d DomainType) SerializeValue(val any) ([]byte, error) {
	return d.DataType.SerializeValue(val)
}

func (d DomainType) DeserializeValue(val []byte) (any, error) {
	return d.DataType.DeserializeValue(val)
}

func (d DomainType) evaluateChecks(ctx *sql.Context, val any) (any, error) {
	if val == nil && d.NotNull {
		return nil, fmt.Errorf("is null but it's not null column")
	}
	if val == nil && d.DefaultExpr != nil {
		return d.DefaultExpr.Eval(ctx, sql.Row{val})
	}
	for _, check := range d.Checks {
		res, err := sql.EvaluateCondition(ctx, check.Expr, sql.Row{val})
		if err != nil {
			return nil, err
		}
		if sql.IsFalse(res) {
			return nil, sql.ErrCheckConstraintViolated.New(check.Name)
		}
	}
	return val, nil
}

func (d DomainType) GetBaseType() DoltgresType {
	switch t := d.DataType.(type) {
	case DomainType:
		return t.GetBaseType()
	default:
		// TODO: how to make sure this is an built-in type?
		return t
	}
}
