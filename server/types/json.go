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
	"unsafe"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/goccy/go-json"
	"github.com/lib/pq/oid"
)

// Json is the standard JSON type.
var Json = JsonType{}

// JsonType is the extended type implementation of the PostgreSQL json.
type JsonType struct{}

var _ DoltgresType = JsonType{}

// Alignment implements the DoltgresType interface.
func (b JsonType) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresType interface.
func (b JsonType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Json
}

// BaseName implements the DoltgresType interface.
func (b JsonType) BaseName() string {
	return "json"
}

// Category implements the DoltgresType interface.
func (b JsonType) Category() TypeCategory {
	return TypeCategory_UserDefinedTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b JsonType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b JsonType) Compare(v1 any, v2 any) (int, error) {
	// JSON does not have any default ordering operators (ORDER BY does not work, etc.), so this is strictly for GMS/Dolt
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
func (b JsonType) Convert(val any) (any, sql.ConvertInRange, error) {
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
func (b JsonType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresType interface.
func (b JsonType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresType interface.
func (b JsonType) GetSerializationID() SerializationID {
	return SerializationID_Json
}

// IoInput implements the DoltgresType interface.
func (b JsonType) IoInput(ctx *sql.Context, input string) (any, error) {
	if json.Valid(unsafe.Slice(unsafe.StringData(input), len(input))) {
		return input, nil
	}
	return nil, fmt.Errorf("invalid input syntax for type json")
}

// IoOutput implements the DoltgresType interface.
func (b JsonType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return converted.(string), nil
}

// IsPreferredType implements the DoltgresType interface.
func (b JsonType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b JsonType) IsUnbounded() bool {
	return true
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b JsonType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b JsonType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresType interface.
func (b JsonType) OID() uint32 {
	return uint32(oid.T_json)
}

// Promote implements the DoltgresType interface.
func (b JsonType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b JsonType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
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
func (b JsonType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b JsonType) String() string {
	return "json"
}

// ToArrayType implements the DoltgresType interface.
func (b JsonType) ToArrayType() DoltgresArrayType {
	return JsonArray
}

// Type implements the DoltgresType interface.
func (b JsonType) Type() query.Type {
	return sqltypes.Text
}

// ValToByteArray implements the DoltgresType interface.
func (b JsonType) ValToByteArray(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	value, err := b.IoOutput(nil, val)
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

// ValueType implements the DoltgresType interface.
func (b JsonType) ValueType() reflect.Type {
	return reflect.TypeOf("")
}

// Zero implements the DoltgresType interface.
func (b JsonType) Zero() any {
	return ""
}

// SerializeType implements the DoltgresType interface.
func (b JsonType) SerializeType() ([]byte, error) {
	return SerializationID_Json.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b JsonType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Json, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b JsonType) SerializeValue(val any) ([]byte, error) {
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
func (b JsonType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return string(val), nil
}
