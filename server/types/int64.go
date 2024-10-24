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
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// Int64 is an int64.
var Int64 = DoltgresType{
	Oid:           uint32(oid.T_int8),
	Name:          "int8",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(8),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_NumericTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__int8),
	InputFunc:     "int8in",
	OutputFunc:    "int8out",
	ReceiveFunc:   "int8recv",
	SendFunc:      "int8send",
	ModInFunc:     "-",
	ModOutFunc:    "-",
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

// Int64Type is the extended type implementation of the PostgreSQL bigint.
type Int64Type struct{}

var _ DoltgresTypeInterface = Int64Type{}

// Alignment implements the DoltgresTypeInterface interface.
func (b Int64Type) Alignment() TypeAlignment {
	return TypeAlignment_Double
}

// BaseID implements the DoltgresTypeInterface interface.
func (b Int64Type) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Int64
}

// BaseName implements the DoltgresTypeInterface interface.
func (b Int64Type) BaseName() string {
	return "int8"
}

// Category implements the DoltgresTypeInterface interface.
func (b Int64Type) Category() TypeCategory {
	return TypeCategory_NumericTypes
}

// CollationCoercibility implements the DoltgresTypeInterface interface.
func (b Int64Type) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresTypeInterface interface.
func (b Int64Type) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(int64)
	bb := bc.(int64)
	if ab == bb {
		return 0, nil
	} else if ab < bb {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresTypeInterface interface.
func (b Int64Type) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case int64:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresTypeInterface interface.
func (b Int64Type) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresTypeInterface interface.
func (b Int64Type) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresTypeInterface interface.
func (b Int64Type) GetSerializationID() SerializationID {
	return SerializationID_Int64
}

// IoInput implements the DoltgresTypeInterface interface.
func (b Int64Type) IoInput(ctx *sql.Context, input string) (any, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid input syntax for type %s: %q", b.String(), input)
	}
	return val, nil
}

// IoOutput implements the DoltgresTypeInterface interface.
func (b Int64Type) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(converted.(int64), 10), nil
}

// IsPreferredType implements the DoltgresTypeInterface interface.
func (b Int64Type) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresTypeInterface interface.
func (b Int64Type) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresTypeInterface interface.
func (b Int64Type) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresTypeInterface interface.
func (b Int64Type) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 8
}

// OID implements the DoltgresTypeInterface interface.
func (b Int64Type) OID() uint32 {
	return uint32(oid.T_int8)
}

// Promote implements the DoltgresTypeInterface interface.
func (b Int64Type) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresTypeInterface interface.
func (b Int64Type) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	return bytes.Compare(v1, v2), nil
}

// SQL implements the DoltgresTypeInterface interface.
func (b Int64Type) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b Int64Type) String() string {
	return "bigint"
}

// ToArrayType implements the DoltgresTypeInterface interface.
func (b Int64Type) ToArrayType() DoltgresArrayType {
	return Int64Array
}

// DoltgresType implements the DoltgresTypeInterface interface.
func (b Int64Type) Type() query.Type {
	return sqltypes.Int64
}

// ValueType implements the DoltgresTypeInterface interface.
func (b Int64Type) ValueType() reflect.Type {
	return reflect.TypeOf(int64(0))
}

// Zero implements the DoltgresTypeInterface interface.
func (b Int64Type) Zero() any {
	return int64(0)
}

// SerializeType implements the DoltgresTypeInterface interface.
func (b Int64Type) SerializeType() ([]byte, error) {
	return SerializationID_Int64.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresTypeInterface interface.
func (b Int64Type) deserializeType(version uint16, metadata []byte) (DoltgresTypeInterface, error) {
	switch version {
	case 0:
		return Int64, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresTypeInterface interface.
func (b Int64Type) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	retVal := make([]byte, 8)
	binary.BigEndian.PutUint64(retVal, uint64(converted.(int64))+(1<<63))
	return retVal, nil
}

// DeserializeValue implements the DoltgresTypeInterface interface.
func (b Int64Type) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return int64(binary.BigEndian.Uint64(val) - (1 << 63)), nil
}
