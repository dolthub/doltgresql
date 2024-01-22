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
	"fmt"
	"reflect"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/shopspring/decimal"
)

//TODO: document everything

var Bool types.Custom

type BoolStructure struct{}

var _ types.CustomStructure = BoolStructure{}

func (b BoolStructure) SerializeValue(c types.Custom, val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := BoolConvert(c, val)
	if err != nil {
		return nil, err
	}
	if converted.(int8) == 0 {
		return []byte{0}, nil
	} else {
		return []byte{1}, nil
	}
}

func (b BoolStructure) DeserializeValue(c types.Custom, val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return int8(val[0]), nil
}

func (b BoolStructure) FormatValue(c types.Custom, val any) (string, error) {
	if val == nil {
		return "", nil
	}
	converted, _, err := BoolConvert(c, val)
	if err != nil {
		return "", err
	}
	if converted.(int8) == 0 {
		return "false", nil
	} else {
		return "true", nil
	}
}

func BoolCompare(c types.Custom, v1 any, v2 any) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}
	if hasNulls, res := types.CompareNulls(v1, v2); hasNulls {
		return res, nil
	}

	ac, _, err := BoolConvert(c, v1)
	if err != nil {
		return 0, err
	}
	bc, _, err := BoolConvert(c, v2)
	if err != nil {
		return 0, err
	}

	ab := ac.(int8)
	bb := bc.(int8)
	if ab == bb {
		return 0, nil
	} else if ab == 0 {
		return -1, nil
	} else {
		return 1, nil
	}
}

func BoolConvert(c types.Custom, val any) (any, sql.ConvertInRange, error) {
	if val == nil {
		return nil, sql.InRange, nil
	}

	switch val := val.(type) {
	case bool:
		if val {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case int:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case uint:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case int8:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case uint8:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case int16:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case uint16:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case int32:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case uint32:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case int64:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case uint64:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case float32:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case float64:
		if val != 0 {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case decimal.NullDecimal:
		if !val.Valid {
			return nil, sql.InRange, nil
		}
		return BoolConvert(c, val.Decimal)
	case decimal.Decimal:
		if !val.Equal(decimal.NewFromInt(0)) {
			return int8(1), sql.InRange, nil
		} else {
			return int8(0), sql.InRange, nil
		}
	case string:
		val = strings.TrimSpace(strings.ToLower(val))
		if val == "true" || val == "yes" || val == "on" || val == "1" {
			return int8(1), sql.InRange, nil
		} else if val == "false" || val == "no" || val == "off" || val == "0" {
			return int8(0), sql.InRange, nil
		} else {
			return int8(0), sql.OutOfRange, fmt.Errorf("invalid string value for boolean")
		}
	case []byte:
		return BoolConvert(c, string(val))
	default:
		return nil, sql.OutOfRange, sql.ErrInvalidType.New(c)
	}
}

func BoolEquals(c types.Custom, otherType sql.Type) bool {
	if _, ok := otherType.(types.Custom); ok {
		return true
	}
	return false
}

func BoolMaxTextResponseByteLength(ctx *sql.Context, c types.Custom) uint32 {
	return 1
}

func BoolPromote(c types.Custom) types.CustomStructure {
	return c.GetStructure()
}

func BoolSQL(ctx *sql.Context, c types.Custom, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, _, err := BoolConvert(c, v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	var valBytes []byte
	if value.(int8) != 0 {
		valBytes = types.AppendAndSliceBytes(dest, []byte{'1'})
	} else {
		valBytes = types.AppendAndSliceBytes(dest, []byte{'0'})
	}
	return sqltypes.MakeTrusted(sqltypes.Int8, valBytes), nil
}

func BoolType(c types.Custom) query.Type {
	return sqltypes.Int8
}

func BoolValueType(c types.Custom) reflect.Type {
	return reflect.TypeOf(int8(0))
}

func BoolZero(c types.Custom) any {
	return int8(0)
}

func BoolString(c types.Custom) string {
	return "boolean"
}

func init() {
	Bool = types.RegisterCustomType(BoolStructure{}, types.CustomFunctions{
		Compare:                   BoolCompare,
		Convert:                   BoolConvert,
		Equals:                    BoolEquals,
		MaxTextResponseByteLength: BoolMaxTextResponseByteLength,
		Promote:                   BoolPromote,
		SQL:                       BoolSQL,
		Type:                      BoolType,
		ValueType:                 BoolValueType,
		Zero:                      BoolZero,
		String:                    BoolString,
	})
}
