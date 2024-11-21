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

// Int32 is an int32.
var Int32 = Int32Type{}

// Int32Type is the extended type implementation of the PostgreSQL integer.
type Int32Type struct{}

var _ DoltgresType = Int32Type{}

// Alignment implements the DoltgresType interface.
func (b Int32Type) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresType interface.
func (b Int32Type) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Int32
}

// BaseName implements the DoltgresType interface.
func (b Int32Type) BaseName() string {
	return "int4"
}

// Category implements the DoltgresType interface.
func (b Int32Type) Category() TypeCategory {
	return TypeCategory_NumericTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b Int32Type) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b Int32Type) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(int32)
	bb := bc.(int32)
	if ab == bb {
		return 0, nil
	} else if ab < bb {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (b Int32Type) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case int32:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b Int32Type) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresType interface.
func (b Int32Type) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresType interface.
func (b Int32Type) GetSerializationID() SerializationID {
	return SerializationID_Int32
}

// IoInput implements the DoltgresType interface.
func (b Int32Type) IoInput(ctx *sql.Context, input string) (any, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(input), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid input syntax for type %s: %q", b.String(), input)
	}
	if val > 2147483647 || val < -2147483648 {
		return nil, fmt.Errorf("value %q is out of range for type %s", input, b.String())
	}
	return int32(val), nil
}

// IoOutput implements the DoltgresType interface.
func (b Int32Type) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(int64(converted.(int32)), 10), nil
}

// IsPreferredType implements the DoltgresType interface.
func (b Int32Type) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b Int32Type) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b Int32Type) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b Int32Type) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 4
}

// OID implements the DoltgresType interface.
func (b Int32Type) OID() uint32 {
	return uint32(oid.T_int4)
}

// Promote implements the DoltgresType interface.
func (b Int32Type) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b Int32Type) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b Int32Type) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b Int32Type) String() string {
	return "integer"
}

// ToArrayType implements the DoltgresType interface.
func (b Int32Type) ToArrayType() DoltgresArrayType {
	return Int32Array
}

// Type implements the DoltgresType interface.
func (b Int32Type) Type() query.Type {
	return sqltypes.Int32
}

// ValueType implements the DoltgresType interface.
func (b Int32Type) ValueType() reflect.Type {
	return reflect.TypeOf(int32(0))
}

// Zero implements the DoltgresType interface.
func (b Int32Type) Zero() any {
	return int32(0)
}

// SerializeType implements the DoltgresType interface.
func (b Int32Type) SerializeType() ([]byte, error) {
	return SerializationID_Int32.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b Int32Type) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Int32, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b Int32Type) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	retVal := make([]byte, 4)
	binary.BigEndian.PutUint32(retVal, uint32(converted.(int32))+(1<<31))
	return retVal, nil
}

// DeserializeValue implements the DoltgresType interface.
func (b Int32Type) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return int32(binary.BigEndian.Uint32(val) - (1 << 31)), nil
}
