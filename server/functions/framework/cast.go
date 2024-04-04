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

package framework

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TODO: Right now, all casts are global. We should decide how to handle this in the presence of branches, sessions, etc.

// TypeCastFunction is a function that takes a value of a particular kind of type, and returns it as another kind of type.
type TypeCastFunction func(ctx Context, val any) (any, error)

// TypeCast is used to cast from one type to another.
type TypeCast struct {
	FromType pgtypes.DoltgresType
	ToType   pgtypes.DoltgresType
	Function TypeCastFunction
}

// typeCastMutex is used to lock the type cast map and array when writing.
var typeCastMutex = &sync.RWMutex{}

// typeCastsMap is a map that maps: from -> to -> function.
var typeCastsMap = map[pgtypes.DoltgresTypeBaseID]map[pgtypes.DoltgresTypeBaseID]TypeCastFunction{}

// typeCastsArray is a slice that holds all registered casts from the given type.
var typeCastsArray = map[pgtypes.DoltgresTypeBaseID][]pgtypes.DoltgresType{}

// AddTypeCast registers the given type cast.
func AddTypeCast(cast TypeCast) error {
	typeCastMutex.Lock()
	defer typeCastMutex.Unlock()

	toMap, ok := typeCastsMap[cast.FromType.BaseID()]
	if !ok {
		toMap = map[pgtypes.DoltgresTypeBaseID]TypeCastFunction{}
		typeCastsMap[cast.FromType.BaseID()] = toMap
		typeCastsArray[cast.FromType.BaseID()] = nil
	}
	if _, ok := toMap[cast.ToType.BaseID()]; ok {
		// TODO: return the actual Postgres error
		return fmt.Errorf("cast from `%s` to `%s` already exists", cast.FromType.String(), cast.ToType.String())
	}
	toMap[cast.ToType.BaseID()] = cast.Function
	typeCastsArray[cast.FromType.BaseID()] = append(typeCastsArray[cast.FromType.BaseID()], cast.ToType)
	return nil
}

// MustAddTypeCast registers the given type cast. Panics if an error occurs.
func MustAddTypeCast(cast TypeCast) {
	if err := AddTypeCast(cast); err != nil {
		panic(err)
	}
}

// GetPotentialCasts returns all registered type casts from the given type.
func GetPotentialCasts(fromType pgtypes.DoltgresTypeBaseID) []pgtypes.DoltgresType {
	typeCastMutex.RLock()
	defer typeCastMutex.RUnlock()

	return typeCastsArray[fromType]
}

// GetCast returns the type cast function that will cast the "from" type to the "to" type. Returns nil if such a cast is
// not valid.
func GetCast(fromType pgtypes.DoltgresTypeBaseID, toType pgtypes.DoltgresTypeBaseID) TypeCastFunction {
	typeCastMutex.RLock()
	defer typeCastMutex.RUnlock()

	if fromType == toType {
		return identityCast
	}
	if toMap, ok := typeCastsMap[fromType]; ok {
		if f, ok := toMap[toType]; ok {
			return f
		}
	}
	// If there isn't a direct mapping, then we need to check if the types are array variants.
	// As long as the base types are convertable, the array variants are also convertable.
	if fromArrayType, ok := pgtypes.IsBaseIDArrayType(fromType); ok {
		if toArrayType, ok := pgtypes.IsBaseIDArrayType(toType); ok {
			if toMap, ok := typeCastsMap[fromArrayType.BaseType().BaseID()]; ok {
				if f, ok := toMap[toArrayType.BaseType().BaseID()]; ok {
					// We use a closure that can unwrap the slice, since conversion functions expect a singular non-nil value
					return func(ctx Context, vals any) (any, error) {
						var err error
						oldVals := vals.([]any)
						newVals := make([]any, len(oldVals))
						for i, oldVal := range oldVals {
							if oldVal == nil {
								continue
							}
							newVals[i], err = f(ctx, oldVal)
							if err != nil {
								return nil, err
							}
						}
						return newVals, nil
					}
				}
			}
		}
	}
	return nil
}

// GetIdentityCast returns the identity cast function.
func GetIdentityCast() TypeCastFunction {
	return identityCast
}

func init() {
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float32,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			return float64(val.(float32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float32,
		ToType:   pgtypes.Int16,
		Function: func(ctx Context, val any) (any, error) {
			if val.(float32) > 32767 || val.(float32) < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val.(float32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float32,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) {
			if val.(float32) > 2147483647 || val.(float32) < -2147483648 {
				return nil, fmt.Errorf("integer out of range")
			}
			return int32(val.(float32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float32,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			if val.(float32) > 9223372036854775807 || val.(float32) < -9223372036854775808 {
				return nil, fmt.Errorf("bigint out of range")
			}
			return int64(val.(float32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float32,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromFloat(float64(val.(float32))), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float32,
		ToType:   pgtypes.VarChar,
		Function: func(ctx Context, val any) (any, error) {
			return strconv.FormatFloat(float64(val.(float32)), 'g', -1, 32), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			return float32(val.(float64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Int16,
		Function: func(ctx Context, val any) (any, error) {
			if val.(float64) > 32767 || val.(float64) < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val.(float64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) {
			if val.(float64) > 2147483647 || val.(float64) < -2147483648 {
				return nil, fmt.Errorf("integer out of range")
			}
			return int32(val.(float64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			if val.(float64) > 9223372036854775807 || val.(float64) < -9223372036854775808 {
				return nil, fmt.Errorf("bigint out of range")
			}
			return int64(val.(float64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromFloat(val.(float64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.VarChar,
		Function: func(ctx Context, val any) (any, error) {
			return strconv.FormatFloat(val.(float64), 'g', -1, 64), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			return float32(val.(int16)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			return float64(val.(int16)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) {
			return int32(val.(int16)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			return int64(val.(int16)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromInt(int64(val.(int16))), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.VarChar,
		Function: func(ctx Context, val any) (any, error) {
			return strconv.FormatInt(int64(val.(int16)), 10), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			return float32(val.(int32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			return float64(val.(int32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int16,
		Function: func(ctx Context, val any) (any, error) {
			if val.(int32) > 32767 || val.(int32) < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val.(int32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			return int64(val.(int32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromInt(int64(val.(int32))), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.VarChar,
		Function: func(ctx Context, val any) (any, error) {
			return strconv.FormatInt(int64(val.(int32)), 10), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			return float32(val.(int64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			return float64(val.(int64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int16,
		Function: func(ctx Context, val any) (any, error) {
			if val.(int64) > 32767 || val.(int64) < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val.(int64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) {
			if val.(int64) > 2147483647 || val.(int64) < -2147483648 {
				return nil, fmt.Errorf("integer out of range")
			}
			return int32(val.(int64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromInt(val.(int64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.VarChar,
		Function: func(ctx Context, val any) (any, error) {
			return strconv.FormatInt(val.(int64), 10), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			f, _ := val.(decimal.Decimal).Float64()
			return float32(f), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			f, _ := val.(decimal.Decimal).Float64()
			return f, nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Int16,
		Function: func(ctx Context, val any) (any, error) {
			d := val.(decimal.Decimal)
			if d.LessThan(pgtypes.NumericValueMinInt16) || d.GreaterThan(pgtypes.NumericValueMaxInt16) {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(d.IntPart()), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) {
			d := val.(decimal.Decimal)
			if d.LessThan(pgtypes.NumericValueMinInt32) || d.GreaterThan(pgtypes.NumericValueMaxInt32) {
				return nil, fmt.Errorf("integer out of range")
			}
			return int32(d.IntPart()), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			d := val.(decimal.Decimal)
			if d.LessThan(pgtypes.NumericValueMinInt64) || d.GreaterThan(pgtypes.NumericValueMaxInt64) {
				return nil, fmt.Errorf("bigint out of range")
			}
			return int64(d.IntPart()), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.VarChar,
		Function: func(ctx Context, val any) (any, error) {
			return val.(decimal.Decimal).String(), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Bool,
		Function: func(ctx Context, val any) (any, error) {
			lowerVal := strings.TrimSpace(strings.ToLower(val.(string)))
			if lowerVal == "true" || lowerVal == "yes" || lowerVal == "on" || lowerVal == "1" {
				return true, nil
			} else if lowerVal == "false" || lowerVal == "no" || lowerVal == "off" || lowerVal == "0" {
				return false, nil
			} else {
				return nil, fmt.Errorf("invalid string value for boolean: %q", val)
			}
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseFloat(val.(string), 32)
			return float32(out), err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseFloat(val.(string), 64)
			return out, err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int16,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseInt(val.(string), 10, 16)
			return int16(out), err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseInt(val.(string), 10, 32)
			return int32(out), err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseInt(val.(string), 10, 64)
			return out, err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromString(val.(string))
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Uuid,
		Function: func(ctx Context, val any) (any, error) {
			return uuid.FromString(val.(string))
		},
	})
}

// identityCast simply returns the input.
func identityCast(ctx Context, val any) (any, error) {
	return val, nil
}
