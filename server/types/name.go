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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/utils"
)

// NameLength is the constant length of Name in Postgres 15.
const NameLength = 63

// Name is a 63-byte internal type for object names.
var Name = NameType{Length: NameLength}

// NameType is the extended type implementation of the PostgreSQL name.
type NameType struct {
	// Length represents the maximum number of characters that the type may hold.
	Length uint32
}

var _ DoltgresType = NameType{}
var _ DoltgresValidType = NameType{}

// Alignment implements the DoltgresType interface.
func (b NameType) Alignment() TypeAlignment {
	return TypeAlignment_Char
}

// BaseID implements the DoltgresType interface.
func (b NameType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Name
}

// BaseName implements the DoltgresType interface.
func (b NameType) BaseName() string {
	return "name"
}

// Category implements the DoltgresType interface.
func (b NameType) Category() TypeCategory {
	return TypeCategory_StringTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b NameType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b NameType) Compare(v1 any, v2 any) (int, error) {
	return compareVarChar(b, v1, v2)
}

// Convert implements the DoltgresType interface.
func (b NameType) Convert(val any) (any, sql.ConvertInRange, error) {
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
func (b NameType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b NameType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b NameType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(val)
}

// GetSerializationID implements the DoltgresType interface.
func (b NameType) GetSerializationID() SerializationID {
	return SerializationID_Name
}

// IoInput implements the DoltgresType interface.
func (b NameType) IoInput(input string) (any, error) {
	// Name seems to never throw an error, regardless of the context or how long the input is
	input, _ = truncateString(input, b.Length)
	return input, nil
}

// IoOutput implements the DoltgresType interface.
func (b NameType) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	str, _ := truncateString(converted.(string), b.Length)
	return str, nil
}

func (b NameType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b NameType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b NameType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b NameType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return b.Length * 4
}

// OID implements the DoltgresType interface.
func (b NameType) OID() uint32 {
	return uint32(oid.T_name)
}

// Promote implements the DoltgresType interface.
func (b NameType) Promote() sql.Type {
	return Name
}

// SerializedCompare implements the DoltgresType interface.
func (b NameType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b NameType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b NameType) String() string {
	return "name"
}

// ToArrayType implements the DoltgresType interface.
func (b NameType) ToArrayType() DoltgresArrayType {
	return NameArray
}

// Type implements the DoltgresType interface.
func (b NameType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (b NameType) ValueType() reflect.Type {
	return reflect.TypeOf("")
}

// Zero implements the DoltgresType interface.
func (b NameType) Zero() any {
	return ""
}

// SerializeType implements the DoltgresType interface.
func (b NameType) SerializeType() ([]byte, error) {
	t := make([]byte, serializationIDHeaderSize+4)
	copy(t, SerializationID_Name.ToByteSlice(0))
	binary.LittleEndian.PutUint32(t[serializationIDHeaderSize:], b.Length)
	return t, nil
}

// deserializeType implements the DoltgresType interface.
func (b NameType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return NameType{
			Length: binary.LittleEndian.Uint32(metadata),
		}, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b NameType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	str := converted.(string)
	writer := utils.NewWriter(uint64(len(str) + 1))
	writer.String(str)
	return writer.Data(), nil
}

// DeserializeValue implements the DoltgresType interface.
func (b NameType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(val)
	return reader.String(), nil
}
