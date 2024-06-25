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
	"reflect"
	"strings"

	"github.com/lib/pq/oid"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
)

// Bool is the standard boolean.
var Bool = BoolType{}

// BoolType is the extended type implementation of the PostgreSQL boolean.
type BoolType struct{}

var _ DoltgresType = BoolType{}

// Alignment implements the DoltgresType interface.
func (b BoolType) Alignment() TypeAlignment {
	return TypeAlignment_Char
}

// BaseID implements the DoltgresType interface.
func (b BoolType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Bool
}

// BaseName implements the DoltgresType interface.
func (b BoolType) BaseName() string {
	return "bool"
}

// Category implements the DoltgresType interface.
func (b BoolType) Category() TypeCategory {
	return TypeCategory_BooleanTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b BoolType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b BoolType) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(bool)
	bb := bc.(bool)
	if ab == bb {
		return 0, nil
	} else if !ab {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the DoltgresType interface.
func (b BoolType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case bool:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b BoolType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b BoolType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b BoolType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(val)
}

// GetSerializationID implements the DoltgresType interface.
func (b BoolType) GetSerializationID() SerializationID {
	return SerializationID_Bool
}

// IoInput implements the DoltgresType interface.
func (b BoolType) IoInput(input string) (any, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "true" || input == "t" || input == "yes" || input == "on" || input == "1" {
		return true, nil
	} else if input == "false" || input == "f" || input == "no" || input == "off" || input == "0" {
		return false, nil
	} else {
		return nil, fmt.Errorf("invalid input syntax for type %s: %q", b.String(), input)
	}
}

// IoOutput implements the DoltgresType interface.
func (b BoolType) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	if converted.(bool) {
		return "true", nil
	} else {
		return "false", nil
	}
}

// IsPreferredType implements the DoltgresType interface.
func (b BoolType) IsPreferredType() bool {
	return true
}

// IsUnbounded implements the DoltgresType interface.
func (b BoolType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b BoolType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b BoolType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 1
}

// OID implements the DoltgresType interface.
func (b BoolType) OID() uint32 {
	return uint32(oid.T_bool)
}

// Promote implements the DoltgresType interface.
func (b BoolType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b BoolType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	if v1[0] == v2[0] {
		return 0, nil
	} else if v1[0] == 0 {
		return -1, nil
	} else {
		return 1, nil
	}
}

// SQL implements the DoltgresType interface.
func (b BoolType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, _, err := b.Convert(v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	var valBytes []byte
	if value.(bool) {
		//TODO: use Wireshark and check whether we're returning these strings or something else
		valBytes = types.AppendAndSliceBytes(dest, []byte{'t'})
	} else {
		valBytes = types.AppendAndSliceBytes(dest, []byte{'f'})
	}
	return sqltypes.MakeTrusted(sqltypes.Text, valBytes), nil
}

// String implements the DoltgresType interface.
func (b BoolType) String() string {
	return "boolean"
}

// ToArrayType implements the DoltgresType interface.
func (b BoolType) ToArrayType() DoltgresArrayType {
	return BoolArray
}

// Type implements the DoltgresType interface.
func (b BoolType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (b BoolType) ValueType() reflect.Type {
	return reflect.TypeOf(bool(false))
}

// Zero implements the DoltgresType interface.
func (b BoolType) Zero() any {
	return false
}

// SerializeType implements the DoltgresType interface.
func (b BoolType) SerializeType() ([]byte, error) {
	return SerializationID_Bool.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b BoolType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Bool, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b BoolType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	if converted.(bool) {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

// DeserializeValue implements the DoltgresType interface.
func (b BoolType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return val[0] != 0, nil
}
