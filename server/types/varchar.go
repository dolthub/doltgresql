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

	"github.com/lib/pq/oid"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
)

const (
	// varCharMax is the maximum number of characters (not bytes) that a VarChar may contain.
	varCharMax = 10485760
	// varCharInline is the maximum number of characters (not bytes) that are "guaranteed" to fit inline.
	varCharInline = 16383
)

// VarCharInline is a varchar that has the max inline length automatically set.
var VarCharInline = VarCharType{Length: varCharInline}

// VarCharMax is a varchar that has the max length.
var VarCharMax = VarCharType{Length: varCharMax}

// VarCharType is the extended type implementation of the PostgreSQL varchar.
type VarCharType struct {
	Length uint32
}

var _ DoltgresType = VarCharType{}

// BaseID implements the DoltgresType interface.
func (b VarCharType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID(SerializationID_VarChar)
}

// CollationCoercibility implements the DoltgresType interface.
func (b VarCharType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b VarCharType) Compare(v1 any, v2 any) (int, error) {
	if hasNulls, res := types.CompareNulls(v1, v2); hasNulls {
		return res, nil
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
	case []byte:
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

// MaxSerializedWidth implements the DoltgresType interface.
func (b VarCharType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	if b.Length <= varCharInline {
		return types.ExtendedTypeSerializedWidth_64K
	} else {
		return types.ExtendedTypeSerializedWidth_Unbounded
	}
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b VarCharType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return b.Length * 4
}

// OID implements the DoltgresType interface.
func (b VarCharType) OID() uint32 {
	return uint32(oid.T_varchar)
}

// Promote implements the DoltgresType interface.
func (b VarCharType) Promote() sql.Type {
	return VarCharMax
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

	//TODO: can be byte-compare unicode strings like this?
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
	return fmt.Sprintf("varchar(%d)", b.Length)
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
