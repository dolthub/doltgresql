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

	"github.com/dolthub/doltgresql/utils"
)

type DomainType struct {
	Schema      string
	Name        string
	AsType      DoltgresType
	DefaultExpr string
	NotNull     bool
	Checks      []*sql.CheckDefinition
}

var _ DoltgresType = DomainType{}

// Alignment implements the DoltgresType interface.
func (d DomainType) Alignment() TypeAlignment {
	return d.AsType.Alignment()
}

// BaseID implements the DoltgresType interface.
func (d DomainType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseId_Domain
}

// BaseName implements the DoltgresType interface.
func (d DomainType) BaseName() string {
	return d.Name
}

// Category implements the DoltgresType interface.
func (d DomainType) Category() TypeCategory {
	return d.AsType.Category()
}

// CollationCoercibility implements the DoltgresType interface.
func (d DomainType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return d.AsType.CollationCoercibility(ctx)
}

// Compare implements the DoltgresType interface.
func (d DomainType) Compare(i interface{}, i2 interface{}) (int, error) {
	return d.AsType.Compare(i, i2)
}

// Convert implements the DoltgresType interface.
func (d DomainType) Convert(i interface{}) (interface{}, sql.ConvertInRange, error) {
	return d.AsType.Convert(i)
}

// Equals implements the DoltgresType interface.
func (d DomainType) Equals(otherType sql.Type) bool {
	return d.AsType.Equals(otherType)
}

// FormatValue implements the types.ExtendedType interface.
func (d DomainType) FormatValue(val any) (string, error) {
	return d.AsType.FormatValue(val)
}

// GetSerializationID implements the DoltgresType interface.
func (d DomainType) GetSerializationID() SerializationID {
	return SerializationId_Domain
}

// IoInput implements the DoltgresType interface.
func (d DomainType) IoInput(ctx *sql.Context, input string) (any, error) {
	return d.AsType.IoInput(ctx, input)
}

// IoOutput implements the DoltgresType interface.
func (d DomainType) IoOutput(ctx *sql.Context, output any) (string, error) {
	return d.AsType.IoOutput(ctx, output)
}

// IsPreferredType implements the DoltgresType interface.
func (d DomainType) IsPreferredType() bool {
	return d.AsType.IsPreferredType()
}

// IsUnbounded implements the DoltgresType interface.
func (d DomainType) IsUnbounded() bool {
	return d.AsType.IsUnbounded()
}

// MaxSerializedWidth implements the types.ExtendedType interface.
func (d DomainType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return d.AsType.MaxSerializedWidth()
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (d DomainType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return d.AsType.MaxTextResponseByteLength(ctx)
}

// OID implements the DoltgresType interface.
func (d DomainType) OID() uint32 {
	//TODO: generate unique oid
	return d.AsType.OID()
}

// Promote implements the DoltgresType interface.
func (d DomainType) Promote() sql.Type {
	return d.AsType.Promote()
}

// SerializedCompare implements the DoltgresType interface.
func (d DomainType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return d.AsType.SerializedCompare(v1, v2)
}

// SQL implements the DoltgresType interface.
func (d DomainType) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	return d.AsType.SQL(ctx, dest, v)
}

// String implements the DoltgresType interface.
func (d DomainType) String() string {
	return d.Name
}

// ToArrayType implements the DoltgresType interface.
func (d DomainType) ToArrayType() DoltgresArrayType {
	return d.AsType.ToArrayType()
}

// Type implements the DoltgresType interface.
func (d DomainType) Type() query.Type {
	return d.AsType.Type()
}

// ValueType implements the DoltgresType interface.
func (d DomainType) ValueType() reflect.Type {
	return d.AsType.ValueType()
}

// Zero implements the DoltgresType interface.
func (d DomainType) Zero() interface{} {
	return d.AsType.Zero()
}

// SerializeType implements the types.ExtendedType interface.
func (d DomainType) SerializeType() ([]byte, error) {
	b := SerializationId_Domain.ToByteSlice(0)
	writer := utils.NewWriter(256)
	writer.String(d.Schema)
	writer.String(d.Name)
	writer.String(d.DefaultExpr)
	writer.Bool(d.NotNull)
	writer.VariableUint(uint64(len(d.Checks)))
	for _, check := range d.Checks {
		writer.String(check.Name)
		writer.String(check.CheckExpression)
	}
	asTyp, err := d.AsType.SerializeType()
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
		d.Schema = reader.String()
		d.Name = reader.String()
		d.DefaultExpr = reader.String()
		d.NotNull = reader.Bool()
		numOfChecks := reader.VariableUint()
		for k := uint64(0); k < numOfChecks; k++ {
			checkName := reader.String()
			checkExpr := reader.String()
			d.Checks = append(d.Checks, &sql.CheckDefinition{
				Name:            checkName,
				CheckExpression: checkExpr,
				Enforced:        true,
			})
		}
		t, err := DeserializeType(reader.RemainingBytes())
		if err != nil {
			return nil, err
		}
		d.AsType = t.(DoltgresType)
		return d, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, d.String())
	}
}

// SerializeValue implements the types.ExtendedType interface.
func (d DomainType) SerializeValue(val any) ([]byte, error) {
	return d.AsType.SerializeValue(val)
}

// DeserializeValue implements the types.ExtendedType interface.
func (d DomainType) DeserializeValue(val []byte) (any, error) {
	return d.AsType.DeserializeValue(val)
}

// UnderlyingBaseType implements the DoltgresDomainType interface.
func (d DomainType) UnderlyingBaseType() DoltgresType {
	switch t := d.AsType.(type) {
	case DomainType:
		return t.UnderlyingBaseType()
	default:
		return t
	}
}
