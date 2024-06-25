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

// Int16 is an int16.
var Int16 = Int16Type{}

// Int16Type is the extended type implementation of the PostgreSQL smallint.
type Int16Type struct{}

var _ DoltgresType = Int16Type{}

// Alignment implements the DoltgresType interface.
func (b Int16Type) Alignment() TypeAlignment {
	return TypeAlignment_Short
}

// BaseID implements the DoltgresType interface.
func (b Int16Type) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Int16
}

// BaseName implements the DoltgresType interface.
func (b Int16Type) BaseName() string {
	return "int2"
}

// Category implements the DoltgresType interface.
func (b Int16Type) Category() TypeCategory {
	return TypeCategory_NumericTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b Int16Type) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b Int16Type) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(int16)
	bb := bc.(int16)
	if ab == bb {
		return 0, nil
	} else if ab < bb {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (b Int16Type) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case int16:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b Int16Type) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b Int16Type) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b Int16Type) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(val)
}

// GetSerializationID implements the DoltgresType interface.
func (b Int16Type) GetSerializationID() SerializationID {
	return SerializationID_Int16
}

// IoInput implements the DoltgresType interface.
func (b Int16Type) IoInput(input string) (any, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(input), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid input syntax for type %s: %q", b.String(), input)
	}
	if val > 32767 || val < -32768 {
		return nil, fmt.Errorf("value %q is out of range for type %s", input, b.String())
	}
	return int16(val), nil
}

// IoOutput implements the DoltgresType interface.
func (b Int16Type) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(int64(converted.(int16)), 10), nil
}

// IsPreferredType implements the DoltgresType interface.
func (b Int16Type) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b Int16Type) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b Int16Type) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b Int16Type) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 2
}

// OID implements the DoltgresType interface.
func (b Int16Type) OID() uint32 {
	return uint32(oid.T_int2)
}

// Promote implements the DoltgresType interface.
func (b Int16Type) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b Int16Type) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b Int16Type) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.FormatValue(v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresType interface.
func (b Int16Type) String() string {
	return "smallint"
}

// ToArrayType implements the DoltgresType interface.
func (b Int16Type) ToArrayType() DoltgresArrayType {
	return Int16Array
}

// Type implements the DoltgresType interface.
func (b Int16Type) Type() query.Type {
	return sqltypes.Int16
}

// ValueType implements the DoltgresType interface.
func (b Int16Type) ValueType() reflect.Type {
	return reflect.TypeOf(int16(0))
}

// Zero implements the DoltgresType interface.
func (b Int16Type) Zero() any {
	return int16(0)
}

// SerializeType implements the DoltgresType interface.
func (b Int16Type) SerializeType() ([]byte, error) {
	return SerializationID_Int16.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b Int16Type) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Int16, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b Int16Type) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	retVal := make([]byte, 2)
	binary.BigEndian.PutUint16(retVal, uint16(converted.(int16))+(1<<15))
	return retVal, nil
}

// DeserializeValue implements the DoltgresType interface.
func (b Int16Type) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return int16(binary.BigEndian.Uint16(val) - (1 << 15)), nil
}
