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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/utils"
)

// InternalCharLength will always be 1.
const InternalCharLength = 1

// InternalChar is a single-byte internal type. In Postgres, it's displayed as "char".
var InternalChar = DoltgresType{
	Oid:           uint32(oid.T_char),
	Name:          "char",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(1),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_InternalUseTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__char),
	InputFunc:     "charin",
	OutputFunc:    "charout",
	ReceiveFunc:   "charrecv",
	SendFunc:      "charsend",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Char,
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

// InternalCharType is the type implementation of the internal PostgreSQL "char" type.
type InternalCharType struct{}

var _ DoltgresTypeInterface = InternalCharType{}

// Alignment implements the DoltgresTypeInterface interface.
func (b InternalCharType) Alignment() TypeAlignment {
	return TypeAlignment_Char
}

// BaseID implements the DoltgresTypeInterface interface.
func (b InternalCharType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_InternalChar
}

// BaseName implements the DoltgresTypeInterface interface.
func (b InternalCharType) BaseName() string {
	return `"char"`
}

// Category implements the DoltgresTypeInterface interface.
func (b InternalCharType) Category() TypeCategory {
	return TypeCategory_InternalUseTypes
}

// CollationCoercibility implements the DoltgresTypeInterface interface.
func (b InternalCharType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresTypeInterface interface.
func (b InternalCharType) Compare(v1 any, v2 any) (int, error) {
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

// Convert implements the DoltgresTypeInterface interface.
func (b InternalCharType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case string:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresTypeInterface interface.
func (b InternalCharType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresTypeInterface interface.
func (b InternalCharType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresTypeInterface interface.
func (b InternalCharType) GetSerializationID() SerializationID {
	return SerializationID_InternalChar
}

// IoInput implements the DoltgresTypeInterface interface.
func (b InternalCharType) IoInput(ctx *sql.Context, input string) (any, error) {
	c := []byte(input)
	if uint32(len(c)) > InternalCharLength {
		return input[:InternalCharLength], nil
	}
	return input, nil
}

// IoOutput implements the DoltgresTypeInterface interface.
func (b InternalCharType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	str := converted.(string)
	if uint32(len(str)) > InternalCharLength {
		return str[:InternalCharLength], nil
	}
	return str, nil
}

// IsPreferredType implements the DoltgresTypeInterface interface.
func (b InternalCharType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresTypeInterface interface.
func (b InternalCharType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresTypeInterface interface.
func (b InternalCharType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresTypeInterface interface.
func (b InternalCharType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return InternalCharLength
}

// OID implements the DoltgresTypeInterface interface.
func (b InternalCharType) OID() uint32 {
	return uint32(oid.T_char)
}

// Promote implements the DoltgresTypeInterface interface.
func (b InternalCharType) Promote() sql.Type {
	return InternalChar
}

// SerializedCompare implements the DoltgresTypeInterface interface.
func (b InternalCharType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}
	return serializedStringCompare(v1, v2), nil
}

// SQL implements the DoltgresTypeInterface interface.
func (b InternalCharType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b InternalCharType) String() string {
	return `"char"`
}

// ToArrayType implements the DoltgresTypeInterface interface.
func (b InternalCharType) ToArrayType() DoltgresArrayType {
	return InternalCharArray
}

// DoltgresType implements the DoltgresTypeInterface interface.
func (b InternalCharType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresTypeInterface interface.
func (b InternalCharType) ValueType() reflect.Type {
	return reflect.TypeOf("")
}

// Zero implements the DoltgresTypeInterface interface.
func (b InternalCharType) Zero() any {
	return ""
}

// SerializeType implements the DoltgresTypeInterface interface.
func (b InternalCharType) SerializeType() ([]byte, error) {
	t := make([]byte, serializationIDHeaderSize+4)
	copy(t, SerializationID_InternalChar.ToByteSlice(0))
	binary.LittleEndian.PutUint32(t[serializationIDHeaderSize:], InternalCharLength)
	return t, nil
}

// deserializeType implements the DoltgresTypeInterface interface.
func (b InternalCharType) deserializeType(version uint16, metadata []byte) (DoltgresTypeInterface, error) {
	switch version {
	case 0:
		return InternalCharType{}, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresTypeInterface interface.
func (b InternalCharType) SerializeValue(val any) ([]byte, error) {
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

// DeserializeValue implements the DoltgresTypeInterface interface.
func (b InternalCharType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(val)
	return reader.String(), nil
}
