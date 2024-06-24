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
	"encoding/hex"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dolthub/doltgresql/utils"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// Bytea is the byte string type.
var Bytea = ByteaType{}

// ByteaType is the extended type implementation of the PostgreSQL bytea.
type ByteaType struct{}

var _ DoltgresType = ByteaType{}
var _ DoltgresValidType = ByteaType{}

// Alignment implements the DoltgresType interface.
func (b ByteaType) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresType interface.
func (b ByteaType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Bytea
}

// BaseName implements the DoltgresType interface.
func (b ByteaType) BaseName() string {
	return "bytea"
}

// Category implements the DoltgresType interface.
func (b ByteaType) Category() TypeCategory {
	return TypeCategory_UserDefinedTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b ByteaType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b ByteaType) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.([]byte)
	bb := bc.([]byte)
	return bytes.Compare(ab, bb), nil
}

// Convert implements the DoltgresType interface.
func (b ByteaType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case []byte:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b ByteaType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b ByteaType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b ByteaType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(val)
}

// GetSerializationID implements the DoltgresType interface.
func (b ByteaType) GetSerializationID() SerializationID {
	return SerializationID_Bytea
}

// IoInput implements the DoltgresType interface.
func (b ByteaType) IoInput(input string) (any, error) {
	if strings.HasPrefix(input, `\x`) {
		return hex.DecodeString(input[2:])
	} else {
		return []byte(input), nil
	}
}

// IoOutput implements the DoltgresType interface.
func (b ByteaType) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return `\x` + hex.EncodeToString(converted.([]byte)), nil
}

func (b ByteaType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b ByteaType) IsUnbounded() bool {
	return true
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b ByteaType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b ByteaType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresType interface.
func (b ByteaType) OID() uint32 {
	return uint32(oid.T_bytea)
}

// Promote implements the DoltgresType interface.
func (b ByteaType) Promote() sql.Type {
	return Bytea
}

// SerializedCompare implements the DoltgresType interface.
func (b ByteaType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}
	return serializedStringCompare(v1, v2), nil
}

// SQL implements the DoltgresType interface.
func (b ByteaType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.FormatValue(v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Blob, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresType interface.
func (b ByteaType) String() string {
	return "bytea"
}

// ToArrayType implements the DoltgresType interface.
func (b ByteaType) ToArrayType() DoltgresArrayType {
	return ByteaArray
}

// Type implements the DoltgresType interface.
func (b ByteaType) Type() query.Type {
	return sqltypes.Blob
}

// ValueType implements the DoltgresType interface.
func (b ByteaType) ValueType() reflect.Type {
	return reflect.TypeOf([]byte{})
}

// Zero implements the DoltgresType interface.
func (b ByteaType) Zero() any {
	return []byte{}
}

// SerializeType implements the DoltgresType interface.
func (b ByteaType) SerializeType() ([]byte, error) {
	return SerializationID_Bytea.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b ByteaType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Bytea, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b ByteaType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	str := converted.([]byte)
	writer := utils.NewWriter(uint64(len(str) + 4))
	writer.ByteSlice(str)
	return writer.Data(), nil
}

// DeserializeValue implements the DoltgresType interface.
func (b ByteaType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(val)
	return reader.ByteSlice(), nil
}
