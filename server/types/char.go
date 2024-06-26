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
	"strings"

	"github.com/dolthub/doltgresql/utils"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// BpChar is a char that has an unbounded length. "bpchar" and "char" are the same type, distinguished by the length
// being bounded or unbounded.
var BpChar = CharType{Length: stringUnbounded}

// InternalChar is a single-byte internal type. In Postgres, it's displayed as "char".
var InternalChar = CharType{Length: 1}

// CharType is the extended type implementation of the PostgreSQL char and bpchar, which are the same type internally.
type CharType struct {
	// Length represents the maximum number of characters that the type may hold.
	// When this is set to unbounded, then it becomes recognized as bpchar.
	Length uint32
}

var _ DoltgresType = CharType{}

// Alignment implements the DoltgresType interface.
func (b CharType) Alignment() TypeAlignment {
	if b.Length == stringUnbounded {
		return TypeAlignment_Int
	} else {
		return TypeAlignment_Char
	}
}

// BaseID implements the DoltgresType interface.
func (b CharType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Char
}

// BaseName implements the DoltgresType interface.
func (b CharType) BaseName() string {
	if b.Length == stringUnbounded {
		return "bpchar"
	} else {
		return "char"
	}
}

// Category implements the DoltgresType interface.
func (b CharType) Category() TypeCategory {
	if b.Length == stringUnbounded {
		return TypeCategory_StringTypes
	} else {
		// TODO: check if it only applies when Length == 1
		return TypeCategory_InternalUseTypes
	}
}

// CollationCoercibility implements the DoltgresType interface.
func (b CharType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b CharType) Compare(v1 any, v2 any) (int, error) {
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

	ab := strings.TrimRight(ac.(string), " ")
	bb := strings.TrimRight(bc.(string), " ")
	if ab == bb {
		return 0, nil
	} else if ab < bb {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (b CharType) Convert(val any) (any, sql.ConvertInRange, error) {
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
func (b CharType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b CharType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b CharType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(val)
}

// GetSerializationID implements the DoltgresType interface.
func (b CharType) GetSerializationID() SerializationID {
	return SerializationID_Char
}

// IoInput implements the DoltgresType interface.
func (b CharType) IoInput(input string) (any, error) {
	if b.IsUnbounded() {
		return input, nil
	} else {
		input, runeLength := truncateString(input, b.Length)
		if runeLength > b.Length {
			return input, fmt.Errorf("value too long for type %s", b.String())
		} else if runeLength < b.Length {
			return input + strings.Repeat(" ", int(b.Length-runeLength)), nil
		} else {
			return input, nil
		}
	}
}

// IoOutput implements the DoltgresType interface.
func (b CharType) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	if b.IsUnbounded() {
		return converted.(string), nil
	} else {
		str, runeCount := truncateString(converted.(string), b.Length)
		if runeCount < b.Length {
			return str + strings.Repeat(" ", int(b.Length-runeCount)), nil
		}
		return str, nil
	}
}

// IsPreferredType implements the DoltgresType interface.
func (b CharType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b CharType) IsUnbounded() bool {
	return b.Length == stringUnbounded
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b CharType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	if b.Length != stringUnbounded && b.Length <= stringInline {
		return types.ExtendedTypeSerializedWidth_64K
	} else {
		return types.ExtendedTypeSerializedWidth_Unbounded
	}
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b CharType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	if b.Length == stringUnbounded {
		return math.MaxUint32
	} else {
		return b.Length * 4
	}
}

// OID implements the DoltgresType interface.
func (b CharType) OID() uint32 {
	if b.Length == stringUnbounded {
		return uint32(oid.T_bpchar)
	} else {
		return uint32(oid.T_char)
	}
}

// Promote implements the DoltgresType interface.
func (b CharType) Promote() sql.Type {
	return BpChar
}

// SerializedCompare implements the DoltgresType interface.
func (b CharType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b CharType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b CharType) String() string {
	return fmt.Sprintf("character(%d)", b.Length)
}

// ToArrayType implements the DoltgresType interface.
func (b CharType) ToArrayType() DoltgresArrayType {
	if b.Length == stringUnbounded {
		return createArrayType(b, SerializationID_CharArray, oid.T__bpchar)
	} else {
		return createArrayType(b, SerializationID_CharArray, oid.T__char)
	}
}

// Type implements the DoltgresType interface.
func (b CharType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (b CharType) ValueType() reflect.Type {
	return reflect.TypeOf("")
}

// Zero implements the DoltgresType interface.
func (b CharType) Zero() any {
	return ""
}

// SerializeType implements the DoltgresType interface.
func (b CharType) SerializeType() ([]byte, error) {
	t := make([]byte, serializationIDHeaderSize+4)
	copy(t, SerializationID_Char.ToByteSlice(0))
	binary.LittleEndian.PutUint32(t[serializationIDHeaderSize:], b.Length)
	return t, nil
}

// deserializeType implements the DoltgresType interface.
func (b CharType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return CharType{
			Length: binary.LittleEndian.Uint32(metadata),
		}, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b CharType) SerializeValue(val any) ([]byte, error) {
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
func (b CharType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(val)
	return reader.String(), nil
}
