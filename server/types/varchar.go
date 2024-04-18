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
	"unicode/utf8"

	"github.com/lib/pq/oid"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
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
	// TODO: need to check if this always truncates for values that are too large, or if it's just the default behavior
	switch val := val.(type) {
	case string:
		// First we'll do a byte-length check since it's always >= the rune-count check, and it's far faster
		if b.Length != stringUnbounded && uint32(len(val)) > b.Length {
			// The byte-length is greater, so now we'll do a rune-count
			if uint32(utf8.RuneCountInString(val)) > b.Length {
				// TODO: figure out if there's a faster way to truncate based on rune count
				startString := val
				for i := uint32(0); i < b.Length; i++ {
					_, size := utf8.DecodeRuneInString(val)
					val = val[size:]
				}
				return startString[:len(startString)-len(val)], sql.InRange, nil
			}
		}
		return val, sql.InRange, nil
	case []byte:
		// First we'll do a byte-length check since it's always >= the rune-count check, and it's far faster
		if b.Length != stringUnbounded && uint32(len(val)) > b.Length {
			// The byte-length is greater, so now we'll do a rune-count
			if uint32(utf8.RuneCount(val)) > b.Length {
				// TODO: figure out if there's a faster way to truncate based on rune count
				startBytes := val
				for i := uint32(0); i < b.Length; i++ {
					_, size := utf8.DecodeRune(val)
					val = val[size:]
				}
				return string(startBytes[:len(startBytes)-len(val)]), sql.InRange, nil
			}
		}
		return string(val), sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, sql.ErrInvalidType.New(b)
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
	converted, _, err := b.Convert(val)
	if err != nil {
		return "", err
	}
	return converted.(string), nil
}

// GetSerializationID implements the DoltgresType interface.
func (b VarCharType) GetSerializationID() SerializationID {
	return SerializationID_VarChar
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
	return bytes.Compare(v1, v2), nil
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
	return []byte(converted.(string)), nil
}

// DeserializeValue implements the DoltgresType interface.
func (b VarCharType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return string(val), nil
}
