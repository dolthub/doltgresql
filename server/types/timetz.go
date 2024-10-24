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
	"time"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/timetz"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// TimeTZ is the time with a time zone. Precision is unbounded.
var TimeTZ = DoltgresType{
	Oid:           uint32(oid.T_timetz),
	Name:          "timetz",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(12),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_DateTimeTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__timetz),
	InputFunc:     "timetz_in",
	OutputFunc:    "timetz_out",
	ReceiveFunc:   "timetz_recv",
	SendFunc:      "timetz_send",
	ModInFunc:     "timetztypmodin",
	ModOutFunc:    "timetztypmodout",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Double,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	Collation:     0,
	DefaulBin:     "",
	Default:       "",
	Acl:           "",
	Checks:        nil,
}

// TimeTZType is the extended type implementation of the PostgreSQL time with time zone.
type TimeTZType struct {
	// TODO: implement precision
	Precision int8
}

var _ DoltgresTypeInterface = TimeTZType{}

// Alignment implements the DoltgresTypeInterface interface.
func (b TimeTZType) Alignment() TypeAlignment {
	return TypeAlignment_Double
}

// BaseID implements the DoltgresTypeInterface interface.
func (b TimeTZType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_TimeTZ
}

// BaseName implements the DoltgresTypeInterface interface.
func (b TimeTZType) BaseName() string {
	return "timetz"
}

// Category implements the DoltgresTypeInterface interface.
func (b TimeTZType) Category() TypeCategory {
	return TypeCategory_DateTimeTypes
}

// CollationCoercibility implements the DoltgresTypeInterface interface.
func (b TimeTZType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresTypeInterface interface.
func (b TimeTZType) Compare(v1 any, v2 any) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	ac, _, err := b.Convert(v1)
	if err != nil {
		return 0, err
	}
	bc, _, err := b.Convert(v2)
	if err != nil {
		return 0, err
	}

	ab := ac.(time.Time)
	bb := bc.(time.Time)
	return ab.Compare(bb), nil
}

// Convert implements the DoltgresTypeInterface interface.
func (b TimeTZType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case time.Time:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresTypeInterface interface.
func (b TimeTZType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresTypeInterface interface.
func (b TimeTZType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresTypeInterface interface.
func (b TimeTZType) GetSerializationID() SerializationID {
	return SerializationID_TimeTZ
}

// IoInput implements the DoltgresTypeInterface interface.
func (b TimeTZType) IoInput(ctx *sql.Context, input string) (any, error) {
	p := b.Precision
	if p == -1 {
		p = 6
	}
	loc, err := GetServerLocation(ctx)
	if err != nil {
		return nil, err
	}
	t, _, err := timetz.ParseTimeTZ(time.Now().In(loc), input, tree.TimeFamilyPrecisionToRoundDuration(int32(p)))
	if err != nil {
		return nil, err
	}
	return t.ToTime(), nil
}

// IoOutput implements the DoltgresTypeInterface interface.
func (b TimeTZType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	// TODO: this always displays the time with an offset relevant to the server location
	t := converted.(time.Time)
	return timetz.MakeTimeTZFromTime(t).String(), nil
}

// IsPreferredType implements the DoltgresTypeInterface interface.
func (b TimeTZType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresTypeInterface interface.
func (b TimeTZType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresTypeInterface interface.
func (b TimeTZType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresTypeInterface interface.
func (b TimeTZType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 12
}

// OID implements the DoltgresTypeInterface interface.
func (b TimeTZType) OID() uint32 {
	return uint32(oid.T_timetz)
}

// Promote implements the DoltgresTypeInterface interface.
func (b TimeTZType) Promote() sql.Type {
	return TimeTZ
}

// SerializedCompare implements the DoltgresTypeInterface interface.
func (b TimeTZType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	// The marshalled time format is byte-comparable
	return bytes.Compare(v1, v2), nil
}

// SQL implements the DoltgresTypeInterface interface.
func (b TimeTZType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.IoOutput(ctx, v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresTypeInterface interface.
func (b TimeTZType) String() string {
	if b.Precision == -1 {
		return "timetz"
	}
	return fmt.Sprintf("timetz(%d)", b.Precision)
}

// ToArrayType implements the DoltgresTypeInterface interface.
func (b TimeTZType) ToArrayType() DoltgresArrayType {
	return createArrayType(b, SerializationID_TimeTZArray, oid.T__timetz)
}

// DoltgresType implements the DoltgresTypeInterface interface.
func (b TimeTZType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresTypeInterface interface.
func (b TimeTZType) ValueType() reflect.Type {
	return reflect.TypeOf(time.Time{})
}

// Zero implements the DoltgresTypeInterface interface.
func (b TimeTZType) Zero() any {
	return time.Time{}
}

// SerializeType implements the DoltgresTypeInterface interface.
func (b TimeTZType) SerializeType() ([]byte, error) {
	t := make([]byte, serializationIDHeaderSize+1)
	copy(t, SerializationID_TimeTZ.ToByteSlice(0))
	t[serializationIDHeaderSize] = byte(b.Precision)
	return t, nil
}

// deserializeType implements the DoltgresTypeInterface interface.
func (b TimeTZType) deserializeType(version uint16, metadata []byte) (DoltgresTypeInterface, error) {
	switch version {
	case 0:
		return TimeTZType{
			Precision: int8(metadata[0]),
		}, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresTypeInterface interface.
func (b TimeTZType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	return converted.(time.Time).MarshalBinary()
}

// DeserializeValue implements the DoltgresTypeInterface interface.
func (b TimeTZType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	t := time.Time{}
	if err := t.UnmarshalBinary(val); err != nil {
		return nil, err
	}
	return t, nil
}
