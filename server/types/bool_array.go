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
	"math"
	"reflect"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"

	"github.com/dolthub/doltgresql/utils"
)

// BoolArray is the standard boolean array.
var BoolArray = BoolArrayType{}

// BoolArrayType is the extended type implementation of the PostgreSQL boolean.
type BoolArrayType struct{}

var _ types.ExtendedType = BoolArrayType{}

// CollationCoercibility implements the types.ExtendedType interface.
func (b BoolArrayType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the types.ExtendedType interface.
func (b BoolArrayType) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.([]bool)
	bb := bc.([]bool)
	minLength := utils.Min(len(ab), len(bb))
	for i := 0; i < minLength; i++ {
		if ab[i] == bb[i] {
			continue
		} else if !ab[i] {
			return -1, nil
		} else {
			return 1, nil
		}
	}
	if len(ab) == len(bb) {
		return 0, nil
	} else if len(ab) < len(bb) {
		return -1, nil
	} else {
		return 1, nil
	}
}

// Convert implements the types.ExtendedType interface.
func (b BoolArrayType) Convert(val any) (any, sql.ConvertInRange, error) {
	if val == nil {
		return nil, sql.InRange, nil
	}

	switch val := val.(type) {
	case []bool:
		return val, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, sql.ErrInvalidType.New(b)
	}
}

// Equals implements the types.ExtendedType interface.
func (b BoolArrayType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the types.ExtendedType interface.
func (b BoolArrayType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the types.ExtendedType interface.
func (b BoolArrayType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	for i, v := range converted.([]bool) {
		if i > 0 {
			sb.WriteString(", ")
		}
		if v {
			return "true", nil
		} else {
			return "false", nil
		}
	}
	return sb.String(), nil
}

// MaxSerializedWidth implements the types.ExtendedType interface.
func (b BoolArrayType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the types.ExtendedType interface.
func (b BoolArrayType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// Promote implements the types.ExtendedType interface.
func (b BoolArrayType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the types.ExtendedType interface.
func (b BoolArrayType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	minLength := utils.Min(len(v1), len(v2))
	for i := 0; i < minLength; i++ {
		if v1[i] == v2[i] {
			continue
		} else if v1[i] == 0 {
			return -1, nil
		} else {
			return 1, nil
		}
	}
	if len(v1) == len(v2) {
		return 0, nil
	} else if len(v1) < len(v2) {
		return -1, nil
	} else {
		return 1, nil
	}
}

// SQL implements the types.ExtendedType interface.
func (b BoolArrayType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	valueAny, _, err := b.Convert(v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	value := valueAny.([]bool)
	if len(value) == 0 {
		return sqltypes.MakeTrusted(b.Type(), types.AppendAndSliceBytes(dest, []byte{'{', '}'})), nil
	}
	valBytes := make([]byte, 2+len(value)+(len(value)-1)) // {t,f,t} | we're including the brackets and commas
	valBytes[0] = '{'
	valBytes[len(valBytes)-1] = '}'
	valBytesIndex := 1
	for valueIndex := range value {
		if valueIndex > 0 {
			valBytes[valBytesIndex] = ','
			valBytesIndex++
		}
		if value[valueIndex] {
			valBytes[valBytesIndex] = 't'
		} else {
			valBytes[valBytesIndex] = 'f'
		}
		valBytesIndex++
	}
	return sqltypes.MakeTrusted(b.Type(), types.AppendAndSliceBytes(dest, valBytes)), nil
}

// String implements the types.ExtendedType interface.
func (b BoolArrayType) String() string {
	return "boolean[]"
}

// Type implements the types.ExtendedType interface.
func (b BoolArrayType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the types.ExtendedType interface.
func (b BoolArrayType) ValueType() reflect.Type {
	return reflect.TypeOf([]bool{})
}

// Zero implements the types.ExtendedType interface.
func (b BoolArrayType) Zero() any {
	return []bool{}
}

// SerializeValue implements the types.ExtendedType interface.
func (b BoolArrayType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	convertedAny, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	converted := convertedAny.([]bool)
	output := make([]byte, len(converted))
	for i := range converted {
		if converted[i] {
			output[i] = 1
		} else {
			output[i] = 0
		}
	}
	return output, nil
}

// DeserializeValue implements the types.ExtendedType interface.
func (b BoolArrayType) DeserializeValue(val []byte) (any, error) {
	if val == nil {
		return nil, nil
	}
	output := make([]bool, len(val))
	for i := range val {
		output[i] = val[i] != 0
	}
	return output, nil
}
