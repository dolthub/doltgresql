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
	"math"
	"reflect"

	"github.com/dolthub/doltgresql/utils"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// Text is the text type.
var Text = DoltgresType{
	Oid:           uint32(oid.T_text),
	Name:          "text",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_StringTypes,
	IsPreferred:   true,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__text),
	InputFunc:     "textin",
	OutputFunc:    "textout",
	ReceiveFunc:   "textrecv",
	SendFunc:      "textsend",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Int,
	Storage:       TypeStorage_Extended,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	Collation:     100,
	DefaulBin:     "",
	Default:       "",
	Acl:           "",
	Checks:        nil,
}

// TextType is the extended type implementation of the PostgreSQL text.
type TextType struct{}

var _ DoltgresTypeInterface = TextType{}

// Alignment implements the DoltgresTypeInterface interface.
func (b TextType) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresTypeInterface interface.
func (b TextType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Text
}

// BaseName implements the DoltgresTypeInterface interface.
func (b TextType) BaseName() string {
	return "text"
}

// Category implements the DoltgresTypeInterface interface.
func (b TextType) Category() TypeCategory {
	return TypeCategory_StringTypes
}

// CollationCoercibility implements the DoltgresTypeInterface interface.
func (b TextType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresTypeInterface interface.
func (b TextType) Compare(v1 any, v2 any) (int, error) {
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

// Convert implements the DoltgresTypeInterface interface.
func (b TextType) Convert(val any) (any, sql.ConvertInRange, error) {
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
func (b TextType) Equals(otherType sql.Type) bool {
	if _, ok := otherType.(TextType); !ok {
		return false
	}

	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresTypeInterface interface.
func (b TextType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresTypeInterface interface.
func (b TextType) GetSerializationID() SerializationID {
	return SerializationID_Text
}

// IoInput implements the DoltgresTypeInterface interface.
func (b TextType) IoInput(ctx *sql.Context, input string) (any, error) {
	return input, nil
}

// IoOutput implements the DoltgresTypeInterface interface.
func (b TextType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return converted.(string), nil
}

// IsPreferredType implements the DoltgresTypeInterface interface.
func (b TextType) IsPreferredType() bool {
	return true
}

// IsUnbounded implements the DoltgresTypeInterface interface.
func (b TextType) IsUnbounded() bool {
	return true
}

// MaxSerializedWidth implements the DoltgresTypeInterface interface.
func (b TextType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresTypeInterface interface.
func (b TextType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresTypeInterface interface.
func (b TextType) OID() uint32 {
	return uint32(oid.T_text)
}

// Promote implements the DoltgresTypeInterface interface.
func (b TextType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresTypeInterface interface.
func (b TextType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b TextType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b TextType) String() string {
	return "text"
}

// ToArrayType implements the DoltgresTypeInterface interface.
func (b TextType) ToArrayType() DoltgresArrayType {
	return TextArray
}

// DoltgresType implements the DoltgresTypeInterface interface.
func (b TextType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresTypeInterface interface.
func (b TextType) ValueType() reflect.Type {
	return reflect.TypeOf("")
}

// Zero implements the DoltgresTypeInterface interface.
func (b TextType) Zero() any {
	return ""
}

// SerializeType implements the DoltgresTypeInterface interface.
func (b TextType) SerializeType() ([]byte, error) {
	return SerializationID_Text.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresTypeInterface interface.
func (b TextType) deserializeType(version uint16, metadata []byte) (DoltgresTypeInterface, error) {
	switch version {
	case 0:
		return Text, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresTypeInterface interface.
func (b TextType) SerializeValue(val any) ([]byte, error) {
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
func (b TextType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(val)
	return reader.String(), nil
}
