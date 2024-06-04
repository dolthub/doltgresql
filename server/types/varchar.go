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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/utils"
)

const (
	// StringMaxLength is the maximum number of characters (not bytes) that a Char, VarChar, or BpChar may contain.
	StringMaxLength = 10485760
	// stringInline is the maximum number of characters (not bytes) that are "guaranteed" to fit inline.
	stringInline = 16383
	// stringUnbounded is used to represent that a type does not define a limit on the strings that it accepts. Values
	// are still limited by the field size limit, but it won't be enforced by the type.
	stringUnbounded = 0
)

// VarChar is a varchar that has an unbounded length.
var VarChar = VarCharType{Length: stringUnbounded}

// VarCharType is the extended type implementation of the PostgreSQL varchar.
type VarCharType struct {
	// Length represents the maximum number of characters that the type may hold.
	// When this is zero, we treat it as completely unbounded (which is still limited by the field size limit).
	Length uint32
}

var _ DoltgresType = VarCharType{}

// BaseID implements the DoltgresType interface.
func (b VarCharType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_VarChar
}

// CollationCoercibility implements the DoltgresType interface.
func (b VarCharType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b VarCharType) Compare(v1 any, v2 any) (int, error) {
	return compareVarChar(b, v1, v2)
}

func compareVarChar(b DoltgresType, v1 any, v2 any) (int, error) {
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

	ab := ac.(string)
	bb := bc.(string)
	if ab == bb {
		return 0, nil
	} else if ab < bb {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (b VarCharType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case string:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b VarCharType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b VarCharType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b VarCharType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(val)
}

// GetSerializationID implements the DoltgresType interface.
func (b VarCharType) GetSerializationID() SerializationID {
	return SerializationID_VarChar
}

// IoInput implements the DoltgresType interface.
func (b VarCharType) IoInput(input string) (any, error) {
	if b.IsUnbounded() {
		return input, nil
	}
	input, runeLength := truncateString(input, b.Length)
	if runeLength > b.Length {
		return input, fmt.Errorf("value too long for type %s", b.String())
	} else {
		return input, nil
	}
}

// IoOutput implements the DoltgresType interface.
func (b VarCharType) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	if b.IsUnbounded() {
		return converted.(string), nil
	}
	str, _ := truncateString(converted.(string), b.Length)
	return str, nil
}

// IsUnbounded implements the DoltgresType interface.
func (b VarCharType) IsUnbounded() bool {
	return b.Length == stringUnbounded
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b VarCharType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	if b.Length != stringUnbounded && b.Length <= stringInline {
		return types.ExtendedTypeSerializedWidth_64K
	} else {
		return types.ExtendedTypeSerializedWidth_Unbounded
	}
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b VarCharType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	if b.Length == stringUnbounded {
		return math.MaxUint32
	} else {
		return b.Length * 4
	}
}

// OID implements the DoltgresType interface.
func (b VarCharType) OID() uint32 {
	return uint32(oid.T_varchar)
}

// Promote implements the DoltgresType interface.
func (b VarCharType) Promote() sql.Type {
	return VarChar
}

// SerializedCompare implements the DoltgresType interface.
func (b VarCharType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b VarCharType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b VarCharType) String() string {
	if b.Length == stringUnbounded {
		return "varchar"
	}
	return fmt.Sprintf("varchar(%d)", b.Length)
}

// ToArrayType implements the DoltgresType interface.
func (b VarCharType) ToArrayType() DoltgresArrayType {
	return createArrayTypeWithFuncs(b, SerializationID_VarCharArray, oid.T__varchar, arrayContainerFunctions{
		SQL: stringArraySQL,
	})
}

// Type implements the DoltgresType interface.
func (b VarCharType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (b VarCharType) ValueType() reflect.Type {
	return reflect.TypeOf("")
}

// Zero implements the DoltgresType interface.
func (b VarCharType) Zero() any {
	return ""
}

// SerializeType implements the DoltgresType interface.
func (b VarCharType) SerializeType() ([]byte, error) {
	t := make([]byte, serializationIDHeaderSize+4)
	copy(t, SerializationID_VarChar.ToByteSlice(0))
	binary.LittleEndian.PutUint32(t[serializationIDHeaderSize:], b.Length)
	return t, nil
}

// deserializeType implements the DoltgresType interface.
func (b VarCharType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return VarCharType{
			Length: binary.LittleEndian.Uint32(metadata),
		}, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b VarCharType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	str := converted.(string)
	writer := utils.NewWriter(uint64(len(str) + 4))
	writer.String(str)
	return writer.Data(), nil
}

// DeserializeValue implements the DoltgresType interface.
func (b VarCharType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(val)
	return reader.String(), nil
}
