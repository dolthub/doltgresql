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
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// Float64 is an float64.
var Float64 = Float64Type{}

// Float64Type is the extended type implementation of the PostgreSQL double precision.
type Float64Type struct{}

var _ DoltgresType = Float64Type{}

// Alignment implements the DoltgresType interface.
func (b Float64Type) Alignment() TypeAlignment {
	return TypeAlignment_Double
}

// BaseID implements the DoltgresType interface.
func (b Float64Type) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Float64
}

// BaseName implements the DoltgresType interface.
func (b Float64Type) BaseName() string {
	return "float8"
}

// Category implements the DoltgresType interface.
func (b Float64Type) Category() TypeCategory {
	return TypeCategory_NumericTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b Float64Type) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b Float64Type) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(float64)
	bb := bc.(float64)
	if ab == bb {
		return 0, nil
	} else if ab < bb {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (b Float64Type) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case float64:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b Float64Type) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b Float64Type) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b Float64Type) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(converted.(float64), 'g', -1, 64), nil
}

// GetSerializationID implements the DoltgresType interface.
func (b Float64Type) GetSerializationID() SerializationID {
	return SerializationID_Float64
}

// IoInput implements the DoltgresType interface.
func (b Float64Type) IoInput(input string) (any, error) {
	val, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid input syntax for type %s: %q", b.String(), input)
	}
	return val, nil
}

// IoOutput implements the DoltgresType interface.
func (b Float64Type) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(converted.(float64), 'f', -1, 64), nil
}

// IsPreferredType implements the DoltgresType interface.
func (b Float64Type) IsPreferredType() bool {
	return true
}

// IsUnbounded implements the DoltgresType interface.
func (b Float64Type) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b Float64Type) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b Float64Type) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 8
}

// OID implements the DoltgresType interface.
func (b Float64Type) OID() uint32 {
	return uint32(oid.T_float8)
}

// Promote implements the DoltgresType interface.
func (b Float64Type) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b Float64Type) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b Float64Type) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b Float64Type) String() string {
	return "double precision"
}

// ToArrayType implements the DoltgresType interface.
func (b Float64Type) ToArrayType() DoltgresArrayType {
	return Float64Array
}

// Type implements the DoltgresType interface.
func (b Float64Type) Type() query.Type {
	return sqltypes.Float64
}

// ValueType implements the DoltgresType interface.
func (b Float64Type) ValueType() reflect.Type {
	return reflect.TypeOf(float64(0))
}

// Zero implements the DoltgresType interface.
func (b Float64Type) Zero() any {
	return float64(0)
}

// SerializeType implements the DoltgresType interface.
func (b Float64Type) SerializeType() ([]byte, error) {
	return SerializationID_Float64.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b Float64Type) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Float64, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b Float64Type) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	retVal := make([]byte, 8)
	// Make the serialized form trivially comparable using bytes.Compare: https://stackoverflow.com/a/54557561
	unsignedBits := math.Float64bits(converted.(float64))
	if converted.(float64) >= 0 {
		unsignedBits ^= 1 << 63
	} else {
		unsignedBits = ^unsignedBits
	}
	binary.BigEndian.PutUint64(retVal, unsignedBits)
	return retVal, nil
}

// DeserializeValue implements the DoltgresType interface.
func (b Float64Type) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	unsignedBits := binary.BigEndian.Uint64(val)
	if unsignedBits&(1<<63) != 0 {
		unsignedBits ^= 1 << 63
	} else {
		unsignedBits = ^unsignedBits
	}
	return math.Float64frombits(unsignedBits), nil
}
